package agent

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

// mcpConfigStamp mirrors Rust mcp::config::fingerprint's domain: paths and raw
// bytes of global plus TARGET project files, including absence/read failure.
// It deliberately does not use the settings UI's parsed/merged server snapshot.
// Called only in manager workers, never Open, Snapshot, Tick or the UI loop.
func mcpConfigStamp(home, workspace string) [32]byte {
	h := sha256.New()
	for _, p := range []string{filepath.Join(home, ".pi", "agent", "mcp.json"), filepath.Join(workspace, ".pi", "mcp.json")} {
		_ = binary.Write(h, binary.LittleEndian, uint64(len(p)))
		_, _ = h.Write([]byte(p))
		raw, err := os.ReadFile(p)
		if err != nil {
			_, _ = h.Write([]byte{0})
			continue
		}
		_, _ = h.Write([]byte{1})
		_ = binary.Write(h, binary.LittleEndian, uint64(len(raw)))
		_, _ = h.Write(raw)
	}
	var stamp [32]byte
	copy(stamp[:], h.Sum(nil))
	return stamp
}
func (m *Manager) readMCPStamp(workspace string) [32]byte {
	if m.mcpFingerprint != nil {
		return m.mcpFingerprint(m.options.Home, workspace)
	}
	return mcpConfigStamp(m.options.Home, workspace)
}

// newEntryLocked atomically binds a new runtime to the selected session. Keeping
// this under the validation lock prevents a late completion stealing selection.
func (m *Manager) newEntryLocked(workspace, path, workflow string) (string, *liveSession) {
	m.sequence++
	key := fmt.Sprintf("live-%d", m.sequence)
	entry := &liveSession{state: newState(workspace, path), loading: true, launcher: m.PiBinary()}
	if validWorkflow(workflow) {
		entry.workflowPending = workflow
		entry.state.snapshot.WorkflowMode = workflow
	}
	entry.state.snapshot.Status = "Connecting"
	m.sessions[key] = entry
	m.active = key
	return key, entry
}

func (m *Manager) validateParkedMCP(key string, entry *liveSession, workspace, path string, selection uint64) {
	stamp := m.readMCPStamp(workspace)
	m.mu.Lock()
	if m.closed || m.sessions[key] != entry || entry.stopped || entry.mcpChecking != selection {
		m.mu.Unlock()
		return
	}
	entry.mcpChecking = 0
	entry.loading = false
	// A newer navigation wins. The now-background client is left untouched;
	// its old loaded stamp ensures the next idle reopen validates it again.
	if m.active != key || m.mcpSelection != selection || entry.busy() ||
		entry.mcpStampKnown && entry.mcpStamp == stamp && entry.state.snapshot.Connected && entry.client != nil && entry.client.Alive() {
		m.mu.Unlock()
		m.changed()
		return
	}
	// No prompt can start between the busy check and detachment. Only the stale
	// idle target is removed; running and unrelated parked clients are preserved.
	delete(m.sessions, key)
	nextKey, next := m.newEntryLocked(workspace, path, entry.workflowPending)
	m.mu.Unlock()
	m.changed()
	if entry.client != nil {
		_ = entry.client.Close()
	}
	m.start(nextKey, next, workspace, path)
}
