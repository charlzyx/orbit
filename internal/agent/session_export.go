package agent

import (
	"errors"
	"fmt"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// SessionExportTarget binds a native Save dialog to the task and child captured
// before the dialog opens. A hydrated session ID does not change this identity.
type SessionExportTarget struct {
	task                   WorkflowTarget
	client                 *pirpc.Client
	sessionID, sessionPath string
}
type sessionCommandLease struct {
	kind     string
	client   *pirpc.Client
	notified bool          // protected by Manager.mu
	done     chan struct{} // closed once after actual completion and owner cleanup
}

func sessionIdentityCommand(kind string) bool {
	return kind == pirpc.NewSession || kind == pirpc.SwitchSession || kind == pirpc.Fork || kind == pirpc.Clone
}

// Metadata writes target the captured conversation just like identity/export
// commands, but renaming does not itself advance task identity generation.
func sessionTargetedCommand(kind string) bool {
	return sessionIdentityCommand(kind) || kind == pirpc.SetSessionName
}
func (m *Manager) CaptureSessionExportTarget() (SessionExportTarget, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.sessions[m.active]
	if m.closed || e == nil || e.client == nil || e.loading || e.stopped || e.sessionCommand != nil || !e.client.Alive() || !e.state.snapshot.Connected {
		return SessionExportTarget{}, errors.New("Pi session is not ready to export")
	}
	return SessionExportTarget{WorkflowTarget{m.active, e.workflowGeneration}, e.client, e.state.snapshot.ActiveSession.ID, e.state.snapshot.ActiveSession.Path}, nil
}
func (m *Manager) reserveSessionCommandLocked(e *liveSession, kind string) (*sessionCommandLease, error) {
	if e.sessionCommand != nil {
		return nil, errors.New("A session change or export is still pending; wait or restart Pi")
	}
	count := 0
	for _, entry := range m.sessions {
		if entry.sessionCommand != nil {
			count++
		}
	}
	if count >= m.options.MaxParked+1 {
		return nil, errors.New("Too many pending session operations; wait or stop a pending Pi session")
	}
	lease := &sessionCommandLease{kind: kind, client: e.client, done: make(chan struct{})}
	e.sessionCommand = lease
	return lease, nil
}
func (m *Manager) releaseSessionCommand(e *liveSession, lease *sessionCommandLease) {
	m.mu.Lock()
	if e.sessionCommand == lease {
		e.sessionCommand = nil
		if e.stopped || m.closed {
			e.creatingSession = false
			e.loading = false
			e.state.snapshot.Connected = false
			if e.stopped {
				e.state.snapshot.Status = "Stopped"
			} else {
				e.state.snapshot.Status = "Disconnected"
			}
		}
		if lease.done != nil {
			close(lease.done)
		}
	}
	m.mu.Unlock()
	m.changed()
}

// A deadline only notifies. It cannot cancel a Pi command already sent, so the
// reservation survives until a real response or child/manager termination.
func (m *Manager) callSessionCommand(e *liveSession, lease *sessionCommandLease, command pirpc.Command) (pirpc.Response, error) {
	timer := time.AfterFunc(m.options.CommandTimeout, func() {
		m.mu.Lock()
		notify := !m.closed && e.sessionCommand == lease && !lease.notified
		if notify {
			lease.notified = true
			e.localDiagnostics = append(e.localDiagnostics, fmt.Sprintf("%s is still pending in Pi; wait or restart the session. A timeout does not cancel the remote operation.", lease.kind))
		}
		m.mu.Unlock()
		if notify {
			m.changed()
		}
	})
	defer timer.Stop()
	response, err := lease.client.Call(m.ctx, command)
	// A protocol mismatch is not proof that our command completed. Client.Call
	// has dropped that correlation waiter, so keep this entry quarantined until
	// its child exits (or the manager closes); Restart/Stop remain available.
	if response.Command != command.Type && m.ctx.Err() == nil {
		select {
		case <-lease.client.Done():
		case <-m.ctx.Done():
		}
	}
	return response, err
}
func (m *Manager) ExportSessionFor(target SessionExportTarget, outputPath string) error {
	return m.exportSessionFor(target, outputPath, false)
}
func (m *Manager) exportSessionFor(target SessionExportTarget, outputPath string, allowDefault bool) error {
	if outputPath == "" && !allowDefault {
		return errors.New("export output path is required")
	}
	m.mu.Lock()
	e := m.sessions[target.task.key]
	if m.closed || m.active != target.task.key || e == nil || e.workflowGeneration != target.task.generation || (target.sessionID != "" && e.state.snapshot.ActiveSession.ID != target.sessionID) || (target.sessionPath != "" && e.state.snapshot.ActiveSession.Path != target.sessionPath) || e.client != target.client || e.client == nil || e.loading || e.stopped || !e.client.Alive() || !e.state.snapshot.Connected {
		m.mu.Unlock()
		return errors.New("The export's original Pi task is no longer active")
	}
	lease, err := m.reserveSessionCommandLocked(e, pirpc.ExportHTML)
	m.mu.Unlock()
	if err != nil {
		return err
	}
	if !m.launch(func() {
		defer m.releaseSessionCommand(e, lease)
		params := map[string]any{}
		if outputPath != "" {
			params["outputPath"] = outputPath
		}
		_, err := m.callSessionCommand(e, lease, pirpc.NewCommand(pirpc.ExportHTML, params))
		if err != nil {
			m.mu.Lock()
			if !m.closed && m.sessions[target.task.key] == e && !e.stopped {
				e.localDiagnostics = append(e.localDiagnostics, "Session export failed: "+err.Error())
			}
			m.mu.Unlock()
		}
		m.Tick()
	}) {
		m.releaseSessionCommand(e, lease)
		return pirpc.ErrClosed
	}
	m.changed()
	return nil
}

// Check under Manager.mu before applying an identity worker's result. A lease
// also distinguishes repeated intents on one entry; active selection may change
// while a parked operation legitimately finishes.
func (m *Manager) sessionCommandCurrentLocked(key string, e *liveSession, lease *sessionCommandLease) bool {
	return !m.closed && m.sessions[key] == e && !e.stopped && e.client == lease.client && e.sessionCommand == lease
}
