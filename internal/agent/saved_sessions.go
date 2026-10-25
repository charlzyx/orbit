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

// DeleteSession is for a UI-confirmed deletion. It stops the matching process,
// then preserves a complete recoverable copy before unlinking the source.
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
		m.mu.Lock()
		var clients []*pirpc.Client
		for key, entry := range m.sessions {
			if entry.state.snapshot.ActiveSession.Path == absolute {
				if entry.client != nil {
					clients = append(clients, entry.client)
				}
				delete(m.sessions, key)
				if m.active == key {
					m.active = ""
				}
			}
		}
		m.mu.Unlock()
		for _, client := range clients {
			_ = client.Close()
		}
		directory := filepath.Join(m.options.Home, ".orbit-pi", "deleted-sessions")
		if err := os.MkdirAll(directory, 0700); err != nil {
			m.operationError(err)
			return
		}
		destination := filepath.Join(directory, fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(absolute)))
		if err := copyExclusive(absolute, destination); err != nil {
			m.operationError(err)
			return
		}
		if err := os.Remove(absolute); err != nil {
			m.operationError(err)
			return
		}
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
		row, err := validateSessionFile(absolute)
		if err != nil {
			m.operationError(err)
			return
		}
		m.mu.Lock()
		var client *pirpc.Client
		for _, entry := range m.sessions {
			if entry.state.snapshot.ActiveSession.Path == absolute && entry.client != nil && entry.client.Alive() && !entry.loading {
				client = entry.client
				break
			}
		}
		m.mu.Unlock()
		temporary := client == nil
		if temporary {
			// Metadata-only process: no project resources, model calls, network refresh,
			// tools, auto-title, quota bridge, or persistent credential changes.
			client, err = pirpc.Start(pirpc.Config{Executable: m.PiBinary(), PrefixArgs: m.options.PiPrefixArgs, WorkspaceDir: row.Workspace, Environment: m.options.Environment, EnvOverrides: map[string]string{"HOME": m.options.Home, "USERPROFILE": m.options.Home, "PI_CODING_AGENT_DIR": filepath.Join(m.options.Home, ".pi", "agent")}, NoSession: true, Offline: true, Args: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools"}})
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
		_, err = m.callClient(client, pirpc.NewCommand(pirpc.SetSessionName, map[string]any{"name": name}))
		if err != nil {
			m.operationError(err)
			return
		}
		if !temporary {
			_, _ = m.callClient(client, pirpc.NewCommand(pirpc.GetState, nil))
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
