package agent

// ResumeSetupFor retries only an intended launch that has never owned a child.
// Admission and target checks are atomic; unlike Start/Restart it cannot select
// a replacement task after an asynchronous dependency refresh completes.
func (m *Manager) ResumeSetupFor(target WorkflowTarget, binary string) bool {
	m.mu.Lock()
	entry := m.sessions[target.key]
	if m.closed || m.active != target.key || entry == nil || entry.workflowGeneration != target.generation || m.PiBinary() != binary || entry.client != nil || entry.stopped || entry.restartClosing || entry.busy() || entry.state.snapshot.ConnectionError == "" {
		m.mu.Unlock()
		return false
	}
	row := entry.state.snapshot.ActiveSession
	entry.loading = true
	entry.state.snapshot.Status = "Connecting"
	entry.state.snapshot.setConnectionError("")
	m.workers.Add(1)
	m.mu.Unlock()
	m.changed()
	go func() {
		defer m.workers.Done()
		m.start(target.key, entry, row.Workspace, row.Path)
		m.mu.Lock()
		if m.closed && m.sessions[target.key] == entry {
			entry.loading = false
		}
		m.mu.Unlock()
	}()
	return true
}
