package agent

import (
	"errors"
	"strings"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// SessionRenameTarget binds the displayed task to its exact live child before
// an editable prompt opens. A new Pi session can have a path but no file yet.
type SessionRenameTarget struct {
	task                   WorkflowTarget
	entry                  *liveSession
	client                 *pirpc.Client
	sessionID, sessionPath string
}

// CaptureSessionRenameTarget does no filesystem work. All displayed identity
// fields must match under the same lock; a stale UI cannot capture a new task.
func (m *Manager) CaptureSessionRenameTarget(runtimeID, sessionID, sessionPath string) (SessionRenameTarget, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.sessions[m.active]
	if runtimeID == "" || sessionPath == "" || m.active != runtimeID || !renameEntryReady(e) || m.closed || e.sessionCommand != nil || e.state.snapshot.ActiveSession.ID != sessionID || e.state.snapshot.ActiveSession.Path != sessionPath {
		return SessionRenameTarget{}, errors.New("The displayed Pi task is no longer ready to rename; reopen Rename after it refreshes")
	}
	return SessionRenameTarget{WorkflowTarget{m.active, e.workflowGeneration}, e, e.client, sessionID, sessionPath}, nil
}
func renameEntryReady(e *liveSession) bool {
	return e != nil && e.client != nil && !e.loading && !e.creatingSession && !e.restartClosing && !e.stopping && !e.stopped && e.client.Alive() && e.state.snapshot.Connected
}

// RenameSessionFor revalidates and leases the captured child atomically. It
// never infers a target from the current selection or the filesystem at Save.
func (m *Manager) RenameSessionFor(target SessionRenameTarget, name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("session name cannot be empty")
	}
	m.mu.Lock()
	e := m.sessions[target.task.key]
	if m.closed || m.active != target.task.key || e != target.entry || !renameEntryReady(e) || e.workflowGeneration != target.task.generation || e.client != target.client || e.state.snapshot.ActiveSession.ID != target.sessionID || e.state.snapshot.ActiveSession.Path != target.sessionPath {
		m.mu.Unlock()
		return errors.New("The rename's original Pi task is no longer active; reopen Rename")
	}
	lease, err := m.reserveSessionCommandLocked(e, pirpc.SetSessionName)
	m.mu.Unlock()
	if err != nil {
		return err
	}
	if !m.launch(func() {
		defer m.releaseSessionCommand(e, lease)
		_, err := m.callSessionCommand(e, lease, pirpc.NewCommand(pirpc.SetSessionName, map[string]any{"name": name}))
		if err == nil {
			_, err = m.callSessionCommand(e, lease, pirpc.NewCommand(pirpc.GetState, nil))
		}
		if err != nil {
			m.mu.Lock()
			if m.sessionCommandCurrentLocked(target.task.key, e, lease) {
				e.localDiagnostics = append(e.localDiagnostics, "Session rename failed: "+err.Error())
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
