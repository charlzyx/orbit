package agent

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// RetrySetting separates command support from value knowledge. Pi 1.4.0 has a
// setter but no getter in its RPC state. Known becomes true only on setter ACK.
type RetrySetting struct {
	Support                 string // unknown, supported, unsupported
	Enabled, Known, Pending bool
	Error                   string
}
type RuntimeProcess struct {
	Key, SessionID, SessionPath, SessionName, Workspace, Launcher, Executable, State, Error string
	PID                                                                                     int
	StartedAt                                                                               time.Time
	Uptime                                                                                  time.Duration // only meaningful when Alive; never fabricated for failure
	Alive, Loading, Busy                                                                    bool
	ExitCode                                                                                *int
	RecentStderr                                                                            []string
	AutoRetry                                                                               RetrySetting
}
type RuntimeSnapshot struct {
	ConfiguredExecutable, Transport string
	Active                          *RuntimeProcess
	Parked                          []RuntimeProcess
	Closed                          bool
}

// RuntimeSnapshot is an immutable, I/O-free projection of all manager-owned
// processes, including loading and failed entries. Times belong to the child,
// not the active session: parking or idle new_session cannot reset uptime.
func (m *Manager) RuntimeSnapshot() RuntimeSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := RuntimeSnapshot{ConfiguredExecutable: m.PiBinary(), Transport: "stdio JSONL", Closed: m.closed}
	now := time.Now()
	for key, entry := range m.sessions {
		row := entry.state.snapshot.ActiveSession
		process := RuntimeProcess{Key: key, SessionID: row.ID, SessionPath: row.Path, SessionName: row.Name, Workspace: row.Workspace, Launcher: entry.launcher, Loading: entry.loading, Busy: entry.busy(), AutoRetry: entry.autoRetry, Error: entry.state.snapshot.ConnectionError, State: "stopped"}
		if process.AutoRetry.Support == "" {
			process.AutoRetry.Support = "unknown"
		}
		if entry.loading {
			process.State = "loading"
		} else if process.Error != "" {
			process.State = "failed"
		}
		if entry.client != nil {
			child := entry.client.ProcessSnapshot()
			process.PID, process.Launcher, process.Executable = child.PID, child.Launcher, child.Executable
			process.StartedAt, process.Alive, process.ExitCode = child.StartedAt, child.Alive, child.ExitCode
			if child.Alive {
				process.Uptime = now.Sub(child.StartedAt)
				if process.Uptime < 0 {
					process.Uptime = 0
				}
				if !entry.loading && entry.state.snapshot.Connected {
					if process.Busy {
						process.State = "busy"
					} else {
						process.State = "idle"
					}
				}
			} else {
				process.State = "exited"
				process.Busy = false
				process.Loading = false
			}
			if process.Error == "" {
				if err := entry.client.Err(); err != nil && !errors.Is(err, pirpc.ErrClosed) {
					process.Error = err.Error()
				}
			}
			stderr := entry.client.RecentStderr()
			if len(stderr) > 8 {
				stderr = stderr[len(stderr)-8:]
			}
			process.RecentStderr = make([]string, len(stderr))
			for i, line := range stderr {
				process.RecentStderr[i] = redactDiagnostic(line, entry.redactions)
			}
		}
		if m.closed && !process.Alive {
			process.State = "stopped"
			process.Busy = false
			process.Loading = false
		}
		if entry.stopping {
			process.State = "stopping"
		} else if entry.stopped {
			process.State = "stopped"
		}
		process.Error = redactDiagnostic(process.Error, entry.redactions)
		process.AutoRetry.Error = redactDiagnostic(process.AutoRetry.Error, entry.redactions)
		if key == m.active {
			copy := process
			result.Active = &copy
		} else {
			result.Parked = append(result.Parked, process)
		}
	}
	sort.Slice(result.Parked, func(i, j int) bool {
		a, b := result.Parked[i], result.Parked[j]
		if a.SessionPath != b.SessionPath {
			return a.SessionPath < b.SessionPath
		}
		return a.Key < b.Key
	})
	return result
}

// Stop shuts down only the active child and keeps the manager, session path,
// diagnostics, and parked processes available. It is safe during bootstrap.
func (m *Manager) Stop() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	entry := m.sessions[m.active]
	if entry == nil {
		m.mu.Unlock()
		return nil
	}
	if entry.stopped {
		m.mu.Unlock()
		return nil
	}
	entry.stopped = true
	entry.loading = false
	if entry.promptCancel != nil {
		entry.promptCancel()
	}
	entry.state.awaiting = false
	entry.state.snapshot.Connected = false
	entry.state.snapshot.Running = false
	entry.state.snapshot.Compacting = false
	entry.state.snapshot.Retrying = false
	entry.state.snapshot.ConnectionError = ""
	entry.state.snapshot.Status = "Stopped"
	entry.state.snapshot.Dialogs = nil
	client := entry.client
	entry.stopping = client != nil && client.Alive()
	m.mu.Unlock()
	if client != nil {
		m.launch(func() {
			client.Close()
			m.mu.Lock()
			entry.stopping = false
			m.mu.Unlock()
			m.Tick()
			m.changed()
		})
	}
	m.changed()
	return nil
}

// Start resumes the active stopped/failed session in a new child. A manager
// with no active workspace must first be opened with Open(workspace, path).
func (m *Manager) Start() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	entry := m.sessions[m.active]
	if entry == nil {
		m.mu.Unlock()
		return errors.New("no active workspace; open a workspace first")
	}
	if entry.stopping {
		m.mu.Unlock()
		return errors.New("Pi process is still stopping")
	}
	if !entry.stopped && (entry.loading || entry.client != nil && entry.client.ProcessSnapshot().Alive && entry.state.snapshot.Connected) {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()
	return m.Restart()
}

// SetAutoRetry does not optimistically mutate a checkbox or probe on startup.
// At most one setter is in flight per child, avoiding out-of-order value ACKs.
func (m *Manager) SetAutoRetry(enabled bool) error {
	m.mu.Lock()
	key := m.active
	entry := m.sessions[key]
	if m.closed || entry == nil || entry.client == nil || entry.loading || !entry.client.Alive() || !entry.state.snapshot.Connected {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	if entry.autoRetry.Pending {
		m.mu.Unlock()
		return errors.New("auto-retry setting is already pending")
	}
	client := entry.client
	entry.autoRetry.Pending = true
	entry.autoRetry.Error = ""
	m.mu.Unlock()
	if !m.launch(func() {
		ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
		_, err := client.Call(ctx, pirpc.NewCommand(pirpc.SetAutoRetry, map[string]any{"enabled": enabled}))
		cancel()
		m.mu.Lock()
		if current := m.sessions[key]; current == entry && current.client == client {
			entry.autoRetry.Pending = false
			if err == nil {
				entry.autoRetry.Support = "supported"
				entry.autoRetry.Enabled = enabled
				entry.autoRetry.Known = true
			} else {
				entry.autoRetry.Error = redactDiagnostic(err.Error(), entry.redactions)
				var commandError *pirpc.CommandError
				if errors.As(err, &commandError) {
					if strings.Contains(strings.ToLower(err.Error()), "unknown command") || strings.Contains(strings.ToLower(err.Error()), "unsupported") {
						entry.autoRetry.Support = "unsupported"
					}
				} else {
					// A timeout/transport loss does not prove the setter failed to
					// execute. Preserve the last value, but mark current truth unknown.
					entry.autoRetry.Known = false
				}
			}
		}
		m.mu.Unlock()
		m.Tick()
		m.changed()
	}) {
		m.mu.Lock()
		entry.autoRetry.Pending = false
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	m.changed()
	return nil
}

func sensitiveEnvironmentName(name string) bool {
	name = strings.ToUpper(name)
	for _, part := range []string{"TOKEN", "SECRET", "PASSWORD", "PASSWD", "CREDENTIAL", "API_KEY", "APIKEY", "ACCESS_KEY", "PRIVATE_KEY", "AUTHORIZATION"} {
		if strings.Contains(name, part) {
			return true
		}
	}
	return false
}
func sensitiveEnvironmentValues(environment []string) []string {
	var values []string
	for _, variable := range environment {
		if name, value, ok := strings.Cut(variable, "="); ok && value != "" && sensitiveEnvironmentName(name) {
			values = append(values, value)
		}
	}
	return values
}

var credentialAssignment = regexp.MustCompile(`(?i)((?:[a-z_][a-z0-9_]*(?:token|secret|password|passwd|api_key|apikey|access_key|private_key|authorization)|token|secret|password|passwd|api_key|apikey|access_key|private_key|authorization)["']?\s*[:=]\s*)("[^"\r\n]*"|'[^'\r\n]*'|[^\s,;}]+)`)
var bearerCredential = regexp.MustCompile(`(?i)(\bBearer\s+)[A-Za-z0-9._~+/=-]+`)

func redactDiagnostic(value string, secrets []string) string {
	secrets = append([]string{}, secrets...)
	sort.Slice(secrets, func(i, j int) bool { return len(secrets[i]) > len(secrets[j]) })
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[redacted]")
			if strings.ContainsAny(secret, "\r\n") {
				for _, line := range strings.FieldsFunc(secret, func(r rune) bool { return r == '\r' || r == '\n' }) {
					value = strings.ReplaceAll(value, line, "[redacted]")
				}
			}
		}
	}
	value = bearerCredential.ReplaceAllString(value, "${1}[redacted]")
	return credentialAssignment.ReplaceAllString(value, "${1}[redacted]")
}
