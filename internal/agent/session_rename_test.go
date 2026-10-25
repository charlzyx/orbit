package agent

import (
	"github.com/charlzyx/orbit/internal/pirpc"
	"os"
	"path/filepath"
	"testing"
)

func renameTestManager(t *testing.T, hold string) (*Manager, string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "not-yet-persisted.jsonl")
	log := filepath.Join(t.TempDir(), "rename.log")
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_FILE"] = path
		o.EnvOverrides["ORBIT_SESSION_RENAME_LOG"] = log
		o.EnvOverrides["ORBIT_SESSION_HOLD"] = hold
	})
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.ID == "s-1" && s.ActiveSession.Path == path
	})
	return m, w, log
}
func renameTestTarget(t *testing.T, m *Manager) SessionRenameTarget {
	t.Helper()
	s := m.Snapshot()
	v, err := m.CaptureSessionRenameTarget(s.RuntimeID, s.ActiveSession.ID, s.ActiveSession.Path)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestSessionRenameCaptureRejectsStaleDisplayedIdentity(t *testing.T) {
	m, _, _ := renameTestManager(t, "")
	s := m.Snapshot()
	for _, fields := range [][3]string{{"other-runtime", s.ActiveSession.ID, s.ActiveSession.Path}, {s.RuntimeID, "other-id", s.ActiveSession.Path}, {s.RuntimeID, s.ActiveSession.ID, "/other/path"}, {s.RuntimeID, s.ActiveSession.ID, ""}} {
		if _, err := m.CaptureSessionRenameTarget(fields[0], fields[1], fields[2]); err == nil {
			t.Fatal("stale display captured", fields)
		}
	}
	if err := m.RenameSessionFor(SessionRenameTarget{}, "invalid zero target"); err == nil {
		t.Fatal("zero target accepted")
	}
}
func TestSessionRenameRejectsSameClientNewSessionAndGenerationABA(t *testing.T) {
	m, w, log := renameTestManager(t, "")
	target := renameTestTarget(t, m)
	if err := m.NewSession(w); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-2" })
	m.mu.Lock()
	sameClient := m.sessions[m.active].client == target.client
	m.mu.Unlock()
	if !sameClient {
		t.Fatal("fixture did not reuse child")
	}
	if err := m.RenameSessionFor(target, "stale"); err == nil {
		t.Fatal("same-child new task accepted old rename")
	}
	// Even if IDs/paths return to identical values, generation must reject ABA.
	m.mu.Lock()
	m.sessions[m.active].state.snapshot.ActiveSession.ID = target.sessionID
	m.sessions[m.active].state.snapshot.ActiveSession.Path = target.sessionPath
	m.mu.Unlock()
	if err := m.RenameSessionFor(target, "stale ABA"); err == nil {
		t.Fatal("generation ABA accepted")
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("stale rename reached child", err)
	}
}
func TestSessionRenameStopCloseAndReplacementRejectCapturedTarget(t *testing.T) {
	for _, mode := range []string{"stop", "close", "replace"} {
		t.Run(mode, func(t *testing.T) {
			m, _, log := renameTestManager(t, "")
			target := renameTestTarget(t, m)
			switch mode {
			case "stop":
				if err := m.Stop(); err != nil {
					t.Fatal(err)
				}
			case "close":
				if err := m.Close(); err != nil {
					t.Fatal(err)
				}
			case "replace":
				if err := m.Open(t.TempDir(), ""); err != nil {
					t.Fatal(err)
				}
				waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != target.task.key })
			}
			if err := m.RenameSessionFor(target, "stale"); err == nil {
				t.Fatal("unavailable task accepted rename")
			}
			if _, err := os.Stat(log); !os.IsNotExist(err) {
				t.Fatal("stale rename reached child", err)
			}
		})
	}
}
func TestSessionRenameRetainsExactLeaseAcrossPendingAndNewTask(t *testing.T) {
	m, w, log := renameTestManager(t, pirpc.SetSessionName)
	target := renameTestTarget(t, m)
	if err := m.RenameSessionFor(target, "exact old task"); err != nil {
		t.Fatal(err)
	}
	waitRemoteSessionHeld(t, target.client)
	waitFor(t, m, func(Snapshot) bool {
		m.mu.Lock()
		defer m.mu.Unlock()
		return target.entry.sessionCommand != nil && target.entry.sessionCommand.notified
	}) // A diagnostic timeout is not command completion; the lease stays held.
	if err := m.RenameSessionFor(target, "duplicate"); err == nil {
		t.Fatal("duplicate lease accepted")
	}
	if _, err := m.CaptureSessionRenameTarget(target.task.key, target.sessionID, target.sessionPath); err == nil {
		t.Fatal("leased child captured again")
	}
	if err := m.Send(pirpc.Fork, map[string]any{"entryId": "x"}); err == nil {
		t.Fatal("identity RPC bypassed lease")
	}
	if err := m.NewSession(w); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != target.task.key })
	releaseExportTest(t, target.client)
	waitExportLeaseGone(t, m, target.entry)
	raw, err := os.ReadFile(log)
	if err != nil || string(raw) != "s-1:exact old task" {
		t.Fatal("lease renamed wrong child", string(raw), err)
	}
	if s := m.Snapshot(); s.RuntimeID == target.task.key {
		t.Fatal("completion selected old task")
	}
}
func TestSessionRenameCloseDuringHeldCallCompletes(t *testing.T) {
	m, _, _ := renameTestManager(t, pirpc.SetSessionName)
	target := renameTestTarget(t, m)
	if err := m.RenameSessionFor(target, "held"); err != nil {
		t.Fatal(err)
	}
	waitRemoteSessionHeld(t, target.client)
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	waitExportLeaseGone(t, m, target.entry)
}
