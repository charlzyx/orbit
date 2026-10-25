package agent

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func validateSessionFile(path string) (SessionRow, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return SessionRow{}, err
	}
	if !info.Mode().IsRegular() || filepath.Ext(path) != ".jsonl" {
		return SessionRow{}, errors.New("select a regular Pi session JSONL file")
	}
	return ReadSession(path)
}
func (m *Manager) operationError(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	m.lastError = err.Error()
	m.mu.Unlock()
	m.changed()
}

// DeleteSession is for a UI-confirmed deletion. Live runtime entries are
// protected; inactive files retain a recoverable copy before source unlink.
func (m *Manager) DeleteSession(path string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("session path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !m.launch(func() {
		if _, err := validateSessionFile(absolute); err != nil {
			m.operationError(err)
			return
		}
		canonical, err := canonicalSessionPath(absolute)
		if err != nil {
			m.operationError(err)
			return
		}
		absolute = canonical
		matches, err := m.reserveSessionDeletion(absolute)
		if err != nil {
			m.operationError(err)
			return
		}
		m.mu.Lock()
		expectedIdentity := m.deletingSessionInfo[absolute]
		m.mu.Unlock()
		deleted := false
		defer func() {
			m.mu.Lock()
			if deleted {
				if m.deletedSessionSequence == nil {
					m.deletedSessionSequence = map[string]uint64{}
				}
				m.deletedSessionSequence[absolute] = m.sequence
				if m.deletedSessionInfo == nil {
					m.deletedSessionInfo = map[string]os.FileInfo{}
				}
				m.deletedSessionInfo[absolute] = m.deletingSessionInfo[absolute]
			}
			delete(m.deletingSessions, absolute)
			delete(m.deletingSessionInfo, absolute)
			m.mu.Unlock()
		}()
		directory := filepath.Join(m.options.Home, ".orbit-pi", "deleted-sessions")
		if err := os.MkdirAll(directory, 0700); err != nil {
			m.operationError(err)
			return
		}
		destination := filepath.Join(directory, fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(absolute)))
		if err := verifyReservedSessionIdentity(absolute, expectedIdentity); err != nil {
			m.operationError(err)
			return
		}
		if err := copyExclusive(absolute, destination); err != nil {
			m.operationError(err)
			return
		}
		if err := verifyReservedSessionIdentity(absolute, expectedIdentity); err != nil {
			m.operationError(err)
			return
		}
		if err := os.Remove(absolute); err != nil {
			m.operationError(err)
			return
		}
		// The final defer records the completed-delete fence and releases the
		// reservation together, including Opens queued during this cleanup.
		deleted = true
		m.mu.Lock()
		for key, entry := range matches {
			if current := m.sessions[key]; current == entry && key != m.active && !entry.busy() && !entry.restartClosing && (entry.client == nil || !entry.client.Alive()) {
				delete(m.sessions, key)
			}
		}
		m.mu.Unlock()
		m.ReloadSessions()
		m.changed()
	}) {
		return pirpc.ErrClosed
	}
	return nil
}
func copyExclusive(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	successful := false
	defer func() {
		output.Close()
		if !successful {
			os.Remove(target)
		}
	}()
	if _, err = io.Copy(output, input); err != nil {
		return err
	}
	if err = output.Sync(); err != nil {
		return err
	}
	if err = output.Close(); err != nil {
		return err
	}
	successful = true
	return nil
}

// CloneSessionFile matches Orbit's offline clone semantics: retain every
// complete entry/branch verbatim, replace only header identity/provenance, and
// drop a partial trailing write. The source is never modified.
func CloneSessionFile(source string) (string, error) {
	if _, err := validateSessionFile(source); err != nil {
		return "", err
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return "", err
	}
	complete := bytes.LastIndexByte(data, '\n')
	if complete < 0 {
		return "", errors.New("session has no complete header")
	}
	data = data[:complete+1]
	first := bytes.IndexByte(data, '\n')
	var header map[string]any
	if json.Unmarshal(data[:first], &header) != nil || header["type"] != "session" {
		return "", errors.New("invalid Pi session header")
	}
	id, err := sessionID()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	header["id"] = id
	header["timestamp"] = now
	header["parentSession"] = source
	encoded, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	target := filepath.Join(filepath.Dir(source), strings.NewReplacer(":", "-", ".", "-").Replace(now)+"_"+id+".jsonl")
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	success := false
	defer func() {
		file.Close()
		if !success {
			os.Remove(target)
		}
	}()
	if _, err = file.Write(append(encoded, '\n')); err != nil {
		return "", err
	}
	if _, err = file.Write(data[first+1:]); err != nil {
		return "", err
	}
	if err = file.Sync(); err != nil {
		return "", err
	}
	if err = file.Close(); err != nil {
		return "", err
	}
	success = true
	return target, nil
}
func sessionID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	stamp := uint64(time.Now().UnixMilli())
	for i := 5; i >= 0; i-- {
		value[i] = byte(stamp)
		stamp >>= 8
	}
	value[6] = (value[6] & 0x0f) | 0x70
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
func (m *Manager) CloneSavedSession(path string) error {
	if path == "" {
		return errors.New("session path is required")
	}
	if !m.launch(func() {
		row, err := validateSessionFile(path)
		if err != nil {
			m.operationError(err)
			return
		}
		clone, err := CloneSessionFile(path)
		if err != nil {
			m.operationError(err)
			return
		}
		m.ReloadSessions()
		m.operationError(m.Open(row.Workspace, clone))
	}) {
		return pirpc.ErrClosed
	}
	return nil
}

// RenameSavedSession targets a particular file, never whichever sidebar row
// happens to be active when an asynchronous command eventually runs.
func (m *Manager) RenameSavedSession(path, name string) error {
	if path == "" {
		return errors.New("session path is required")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if !m.launch(func() {
		// Preserve the regular-JSONL admission rule, including leaf-symlink rejection.
		info, statErr := os.Lstat(absolute)
		if statErr != nil {
			m.operationError(statErr)
			return
		}
		if !info.Mode().IsRegular() || filepath.Ext(absolute) != ".jsonl" {
			m.operationError(errors.New("select a regular Pi session JSONL file"))
			return
		}
		canonical, targetInfo, canonicalErr := readSessionFileIdentity(absolute)
		if canonicalErr != nil {
			m.operationError(canonicalErr)
			return
		}
		absolute = canonical

		m.mu.Lock()
		if reservedSessionFile(absolute, targetInfo, m.deletingSessions, m.deletingSessionInfo) || reservedSessionFile(absolute, targetInfo, m.metadataSessions, m.metadataSessionInfo) {
			m.mu.Unlock()
			m.operationError(errors.New("session file operation is in progress; try again when it finishes"))
			return
		}
		if m.metadataSessions == nil {
			m.metadataSessions = map[string]bool{}
		}
		m.metadataSessions[absolute] = true
		if m.metadataSessionInfo == nil {
			m.metadataSessionInfo = map[string]os.FileInfo{}
		}
		m.metadataSessionInfo[absolute] = targetInfo
		m.mu.Unlock()
		var target *metadataSessionTarget
		defer func() {
			m.mu.Lock()
			delete(m.metadataSessions, absolute)
			delete(m.metadataSessionInfo, absolute)
			m.mu.Unlock()
			// Cleanup is complete before restart's lease waiter is released.
			if target != nil {
				m.releaseSessionCommand(target.entry, target.lease)
			}
		}()
		row, err := validateSessionFile(absolute)
		if err != nil {
			m.operationError(err)
			return
		}
		target, err = m.reserveMetadataSessionTarget(absolute, targetInfo)
		if err != nil {
			m.operationError(err)
			return
		}
		temporary := target == nil
		var client *pirpc.Client
		if target != nil {
			client = target.lease.client
		}
		if temporary {
			// Metadata-only process: no project resources, model calls, network refresh,
			// tools, auto-title, quota bridge, or persistent credential changes.
			client, err = pirpc.Start(pirpc.Config{Executable: m.PiBinary(), PrefixArgs: m.options.PiPrefixArgs, WorkspaceDir: row.Workspace, Environment: m.options.Environment, EnvOverrides: m.piEnvOverrides(), NoSession: true, Offline: true, Args: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools"}})
			if err != nil {
				m.operationError(err)
				return
			}
			defer client.Close()
			response, err := m.callClient(client, pirpc.NewCommand(pirpc.SwitchSession, map[string]any{"sessionPath": absolute}))
			if err != nil {
				m.operationError(err)
				return
			}
			var value struct{ Cancelled bool }
			_ = response.DecodeData(&value)
			if value.Cancelled {
				m.operationError(errors.New("Pi cancelled opening the selected session"))
				return
			}
		}
		command := pirpc.NewCommand(pirpc.SetSessionName, map[string]any{"name": name})
		if target != nil {
			_, err = m.callSessionCommand(target.entry, target.lease, command)
		} else {
			_, err = m.callClient(client, command)
		}
		if err != nil {
			if target == nil {
				m.operationError(err)
			} else {
				m.mu.Lock()
				if m.sessionCommandCurrentLocked(target.key, target.entry, target.lease) {
					target.entry.localDiagnostics = append(target.entry.localDiagnostics, "Session rename failed: "+err.Error())
				}
				m.mu.Unlock()
				m.changed()
			}
			return
		}
		if target != nil {
			m.mu.Lock()
			current := m.sessionCommandCurrentLocked(target.key, target.entry, target.lease)
			m.mu.Unlock()
			if current {
				_, _ = m.callSessionCommand(target.entry, target.lease, pirpc.NewCommand(pirpc.GetState, nil))
			}
		}
		m.ReloadSessions()
		m.Tick()
		m.changed()
	}) {
		return pirpc.ErrClosed
	}
	return nil
}

// WaitForIdle is useful for explicit workflows that need a completed run. It
// does not send a prompt or infer completion merely from a command response.
func (m *Manager) WaitForIdle(ctx context.Context) error {
	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()
	for {
		m.Tick()
		m.mu.Lock()
		entry := m.sessions[m.active]
		idle := entry != nil && !entry.loading && !entry.state.snapshot.Running && !entry.state.awaiting
		closed := m.closed
		m.mu.Unlock()
		if closed {
			return pirpc.ErrClosed
		}
		if idle {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
