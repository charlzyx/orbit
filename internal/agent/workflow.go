package agent

import (
	"errors"
	"github.com/charlzyx/orbit/internal/preferences"
)

// WorkflowTarget captures task identity before a host queues an I/O worker.
// Fields are deliberately opaque: an empty-ID draft is identified by its live
// entry and generation, never merely by the currently selected workspace.
type WorkflowTarget struct {
	key        string
	generation uint64
}

func (m *Manager) CaptureWorkflowTarget() WorkflowTarget {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.workflowTargetLocked(m.active)
}
func (m *Manager) workflowTargetLocked(key string) WorkflowTarget {
	if e := m.sessions[key]; e != nil {
		return WorkflowTarget{key, e.workflowGeneration}
	}
	return WorkflowTarget{}
}

// CaptureWorkflowTargetFor binds a displayed composer to its current task;
// matching unhydrated tasks remain valid (the existing setter holds pending mode).
// This is identity capture only, with no filesystem work or Pi readiness RPC.
func (m *Manager) CaptureWorkflowTargetFor(runtimeID, sessionID, path, workspace string) (WorkflowTarget, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e := m.sessions[m.active]
	if m.closed || e == nil || e.stopped || e.stopping || e.restartClosing || runtimeID == "" || m.active != runtimeID || e.state.snapshot.ActiveSession.ID != sessionID || e.state.snapshot.ActiveSession.Path != path || e.state.snapshot.ActiveSession.Workspace != workspace {
		return WorkflowTarget{}, errors.New("The displayed task changed; reopen the mode picker after it refreshes")
	}
	return m.workflowTargetLocked(m.active), nil
}

// setWorkflowMode follows composer_ops.rs::choose_workflow_mode: a choice
// made before Pi supplies an ID belongs to that new task, not the parked one.
func (m *Manager) setWorkflowMode(mode string) error {
	return m.SetWorkflowModeFor(m.CaptureWorkflowTarget(), mode)
}

func (m *Manager) SetWorkflowModeFor(target WorkflowTarget, mode string) error {
	if !validWorkflow(mode) {
		return errors.New("unknown workflow mode")
	}
	m.mu.Lock()
	entry := m.sessions[target.key]
	if m.closed || entry == nil || entry.workflowGeneration != target.generation {
		m.mu.Unlock()
		return errors.New("workflow target task is no longer available")
	}
	if m.workflowStore == nil {
		m.workflowStore = preferences.New(m.options.Home)
	}
	entry.workflowVersion++
	version := entry.workflowVersion
	entry.state.snapshot.WorkflowMode = mode
	id := entry.state.snapshot.ActiveSession.ID
	if id == "" {
		entry.workflowPending = mode
		m.mu.Unlock()
		m.changed()
		return nil
	}
	entry.workflowPending = ""
	m.mu.Unlock()
	m.changed()
	return m.persistWorkflow(entry, id, mode, version)
}

func (m *Manager) persistWorkflow(entry *liveSession, id, mode string, version uint64) error {
	// Serializing across sessions avoids read/modify/write lost updates. Never
	// hold the UI/state mutex over disk I/O.
	m.workflowMu.Lock()
	defer m.workflowMu.Unlock()
	m.mu.Lock()
	current := !m.closed && entry.workflowVersion == version && entry.state.snapshot.ActiveSession.ID == id
	m.mu.Unlock()
	if !current {
		return nil
	}
	err := m.workflowStore.SaveWorkflow(id, mode)
	m.mu.Lock()
	if entry.workflowVersion == version && entry.state.snapshot.ActiveSession.ID == id {
		if err == nil {
			entry.workflowPending = ""
		} else {
			entry.localDiagnostics = append(entry.localDiagnostics, "Could not save workflow: "+err.Error())
		}
	}
	m.mu.Unlock()
	m.changed()
	return err
}

// workflowIdentityLocked returns an I/O worker, never executes it on Tick.
// A late result cannot overwrite a newer explicit selection or another ID.
func (m *Manager) workflowIdentityLocked(entry *liveSession) func() {
	id := entry.state.snapshot.ActiveSession.ID
	if id == "" || id == entry.workflowID {
		return nil
	}
	entry.workflowID = id
	entry.workflowVersion++
	version := entry.workflowVersion
	pending := entry.workflowPending
	if pending != "" {
		entry.state.snapshot.WorkflowMode = pending
		return func() { _ = m.persistWorkflow(entry, id, pending, version) }
	}
	return func() {
		mode := m.workflowStore.Workflow(id)
		m.mu.Lock()
		if !m.closed && entry.workflowVersion == version && entry.state.snapshot.ActiveSession.ID == id {
			entry.state.snapshot.WorkflowMode = mode
		}
		m.mu.Unlock()
		m.changed()
	}
}
