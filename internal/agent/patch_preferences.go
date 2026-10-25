package agent

import (
	"errors"
	"os"
	"strings"
)

// SetRPCPatchesEnabled changes the next-launch policy, never the running Pi
// installation. Save failure leaves the visible state and existing bytes intact.
func (m *Manager) SetRPCPatchesEnabled(enabled bool) error {
	m.patchPolicyMu.Lock()
	defer m.patchPolicyMu.Unlock()
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return errors.New("agent manager is closed")
	}
	if err := m.workflowStore.SetRPCPatchesEnabled(enabled); err != nil {
		return err
	}
	m.mu.Lock()
	m.patchPolicyEnabled = enabled
	m.patchPolicyVersion++
	for _, session := range m.sessions {
		session.state.snapshot.Extensions.Patches.Enabled = enabled
	}
	m.mu.Unlock()
	m.changed()
	return nil
}

func patchEnvironmentDisabled(options Options) bool {
	if value, exists := options.EnvOverrides["ORBIT_NO_RPC_PATCHES"]; exists {
		return value != ""
	}
	environment := options.Environment
	if environment == nil {
		environment = os.Environ()
	}
	value := ""
	for _, entry := range environment {
		if name, v, ok := strings.Cut(entry, "="); ok && name == "ORBIT_NO_RPC_PATCHES" {
			value = v
		}
	}
	return value != ""
}

// A slow launch reports the policy it prepared, but may not undo a later saved
// switch. Applied patches and diagnostics still describe the actual launch.
// Caller holds m.mu.
func (m *Manager) adoptPatchStatusLocked(status *ExtensionStatus, version uint64) {
	if version == m.patchPolicyVersion {
		m.patchPolicyEnabled = status.Patches.Enabled
	} else {
		status.Patches.Enabled = m.patchPolicyEnabled
	}
}
