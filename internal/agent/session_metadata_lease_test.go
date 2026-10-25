package agent

import (
	"github.com/charlzyx/orbit/internal/pirpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMetadataRenameRetainsOriginalTaskLeasePastTimeout(t *testing.T) {
	for _, generic := range []bool{false, true} {
		t.Run(map[bool]string{false: "saved-alias", true: "generic-active"}[generic], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "saved.jsonl")
			aliasDir := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(root, aliasDir); err != nil {
				t.Skip(err)
			}
			if err := os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"saved","cwd":`+strconv.Quote(root)+`}`+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			log := filepath.Join(t.TempDir(), "renamed-task")
			m := testManager(t, func(o *Options) {
				o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
				o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
				o.EnvOverrides["ORBIT_SESSION_HOLD"] = pirpc.SetSessionName
				o.EnvOverrides["ORBIT_SESSION_RENAME_LOG"] = log
			})
			if err := m.Open(root, filepath.Join(aliasDir, "saved.jsonl")); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.Connected })
			original := m.Snapshot()
			m.mu.Lock()
			entry := m.sessions[m.active]
			client := entry.client
			generation := entry.workflowGeneration
			m.mu.Unlock()
			var err error
			if generic {
				err = m.RenameSession("target-name")
			} else {
				err = m.RenameSavedSession(path, "target-name")
			}
			if err != nil {
				t.Fatal(err)
			}
			waitRemoteSessionHeld(t, client)
			// Use the ordinary bootstrap deadline and await the actual diagnostic,
			// not a 50ms startup budget or a timing-only sleep under race load.
			waitFor(t, m, func(Snapshot) bool {
				m.mu.Lock()
				defer m.mu.Unlock()
				return entry.sessionCommand != nil && entry.sessionCommand.notified
			})
			m.mu.Lock()
			leaseHeld := entry.sessionCommand != nil
			unchanged := entry.workflowGeneration == generation
			m.mu.Unlock()
			if !leaseHeld || !unchanged {
				t.Fatal("metadata timeout released lease or advanced identity", leaseHeld, unchanged)
			}
			for _, kind := range []string{pirpc.Fork, pirpc.Clone, pirpc.SwitchSession} {
				if err := m.Send(kind, map[string]any{"sessionPath": path}); err == nil {
					t.Fatal("identity command crossed held rename", kind)
				}
			}
			if err := m.NewSession(root); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != original.RuntimeID })
			selected := m.Snapshot().RuntimeID
			releaseExportTest(t, client)
			m.workers.Wait()
			m.Tick()
			got, err := os.ReadFile(log)
			if err != nil || string(got) != original.ActiveSession.ID+":target-name" {
				t.Fatal("rename landed on newer task", string(got), err)
			}
			if m.Snapshot().RuntimeID != selected {
				t.Fatal("parked completion stole selection")
			}
			m.mu.Lock()
			released := entry.sessionCommand == nil
			m.mu.Unlock()
			if !released {
				t.Fatal("real ACK did not release lease")
			}
		})
	}
}
func TestMetadataRenameIntentCannotCrossSameClientNewTaskDuringMarshal(t *testing.T) {
	m, w := exportManager(t, "")
	gate := exportMarshalGate{make(chan struct{}), make(chan struct{})}
	done := make(chan error, 1)
	go func() { done <- m.Send(pirpc.SetSessionName, map[string]any{"name": gate}) }()
	select {
	case <-gate.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("marshal did not enter")
	}
	if err := m.NewSession(w); err != nil {
		close(gate.release)
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-2" })
	close(gate.release)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("stale metadata intent accepted for new same-client task")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("metadata send stuck")
	}
}
func TestOpenWarmAliasFailsClosedDuringDeletionThenReuses(t *testing.T) {
	m := testManager(t, nil)
	root, path, alias := aliasSessionFixture(t)
	if err := m.Open(root, alias); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	original := m.RuntimeSnapshot().Active
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != original.Key })
	selected := m.Snapshot().RuntimeID
	m.mu.Lock()
	m.deletingSessions = map[string]bool{path: true}
	m.mu.Unlock()
	if err := m.Open(root, alias); err == nil || !strings.Contains(err.Error(), "deletion") {
		t.Fatal("warm alias bypassed reservation", err)
	}
	if m.Snapshot().RuntimeID != selected {
		t.Fatal("rejected alias reuse changed selection")
	}
	m.mu.Lock()
	delete(m.deletingSessions, path)
	m.mu.Unlock()
	if err := m.Open(root, alias); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID == original.Key })
	if m.RuntimeSnapshot().Active.PID != original.PID {
		t.Fatal("released reservation stopped warm reuse")
	}
}

func TestStoppedMetadataCompletionCannotPolluteReplacementTask(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "saved.jsonl")
	if err := os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"saved","cwd":`+strconv.Quote(root)+`}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_HOLD"] = pirpc.SetSessionName
	})
	if err := m.Open(root, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	old := m.Snapshot().RuntimeID
	m.mu.Lock()
	client := m.sessions[old].client
	m.mu.Unlock()
	if err := m.RenameSavedSession(path, "held"); err != nil {
		t.Fatal(err)
	}
	waitRemoteSessionHeld(t, client)
	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != old })
	m.workers.Wait()
	m.mu.Lock()
	remaining := len(m.metadataSessions) + len(m.metadataSessionInfo)
	m.mu.Unlock()
	if remaining != 0 {
		t.Fatal("metadata file reservation leaked after restart", remaining)
	}
	if got := m.Snapshot().ConnectionError; got != "" {
		t.Fatal("retired metadata error polluted replacement", got)
	}
}

// A cleanup owner may still be scheduled after child death. Shutdown must not
// deadlock waiting for that owner's done signal, and must not start a new child.
func TestCloseInterruptsRestartWaitingForSessionCleanup(t *testing.T) {
	m, _ := exportManager(t, "")
	m.mu.Lock()
	old := m.sessions[m.active]
	lease, err := m.reserveSessionCommandLocked(old, pirpc.SetSessionName)
	m.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	defer m.releaseSessionCommand(old, lease)
	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-old.client.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("old child did not close")
	}
	m.mu.Lock()
	replacement := m.sessions[m.active]
	waiting := replacement.restartClosing && replacement.client == nil
	m.mu.Unlock()
	if !waiting {
		t.Fatal("replacement bypassed pending owner cleanup")
	}
	done := make(chan error, 1)
	go func() { done <- m.Close() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Close blocked on owner cleanup")
	}
	m.mu.Lock()
	started := replacement.client != nil
	m.mu.Unlock()
	if started {
		t.Fatal("replacement started after shutdown")
	}
}
