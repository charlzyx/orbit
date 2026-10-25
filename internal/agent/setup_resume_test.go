package agent

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func failedSetupManager(t *testing.T) (*Manager, string, WorkflowTarget) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "not-yet-installed-pi")
	m := testManager(t, func(o *Options) { o.PiBinary = path })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active != nil && s.Active.State == "failed" })
	return m, path, m.CaptureWorkflowTarget()
}
func TestSetupResumeStartsExactNeverLiveEntryOnlyOnce(t *testing.T) {
	m, path, target := failedSetupManager(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(exe, path); err != nil {
		t.Fatal(err)
	}
	before := m.RuntimeSnapshot().Active.Key
	if !m.ResumeSetupFor(target, path) {
		t.Fatal("ready dependency did not retry failed launch")
	}
	if m.ResumeSetupFor(target, path) {
		t.Fatal("duplicate resume admitted")
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	after := m.RuntimeSnapshot().Active
	if after.Key != before || after.SessionID != "s-1" || after.PID == 0 {
		t.Fatal("resume created/replaced task", after)
	}
	if m.ResumeSetupFor(target, path) {
		t.Fatal("live child restarted by refresh")
	}
}
func TestSetupResumeRejectsChangedIntentBinaryStoppedAndClosed(t *testing.T) {
	for _, kind := range []string{"task", "generation", "binary", "stop", "close"} {
		t.Run(kind, func(t *testing.T) {
			m, path, target := failedSetupManager(t)
			switch kind {
			case "task":
				if err := m.Open(t.TempDir(), ""); err != nil {
					t.Fatal(err)
				}
			case "generation":
				m.mu.Lock()
				m.sessions[m.active].workflowGeneration++
				m.mu.Unlock()
			case "binary":
				m.configuredBinary.Store("another-selected-pi")
			case "stop":
				if err := m.Stop(); err != nil {
					t.Fatal(err)
				}
			case "close":
				m.Close()
			}
			key := m.RuntimeSnapshot().Active.Key
			if m.ResumeSetupFor(target, path) {
				t.Fatal("obsolete setup result admitted")
			}
			if m.RuntimeSnapshot().Active.Key != key {
				t.Fatal("rejected result replaced task")
			}
		})
	}
}
func TestSetupResumeNeverRestartsPreviouslyLiveStoppedChild(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	target := m.CaptureWorkflowTarget()
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if m.ResumeSetupFor(target, m.PiBinary()) {
		t.Fatal("read-only setup resumed intentionally stopped live task")
	}
}

func TestSetupResumeHeldLaunchCannotReviveAfterStopOrClose(t *testing.T) {
	for _, kind := range []string{"stop", "close"} {
		t.Run(kind, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "late-pi")
			started, release := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			m := testManager(t, func(o *Options) {
				o.PiBinary = path
				o.ResolveChildEnv = func(string) ([]string, error) {
					if calls.Add(1) > 1 {
						close(started)
						<-release
					}
					return nil, nil
				}
			})
			t.Cleanup(unblock)
			if err := m.Open(t.TempDir(), ""); err != nil {
				t.Fatal(err)
			}
			waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "failed" })
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Symlink(exe, path); err != nil {
				t.Fatal(err)
			}
			if !m.ResumeSetupFor(m.CaptureWorkflowTarget(), path) {
				t.Fatal("resume not admitted")
			}
			<-started
			done := make(chan struct{})
			if kind == "close" {
				go func() { m.Close(); close(done) }()
				deadline := time.Now().Add(time.Second)
				for !m.RuntimeSnapshot().Closed && time.Now().Before(deadline) {
					time.Sleep(time.Millisecond)
				}
			} else {
				if err := m.Stop(); err != nil {
					t.Fatal(err)
				}
				close(done)
			}
			unblock()
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("lifecycle leaked resume worker")
			}
			waitRuntime(t, m, func(s RuntimeSnapshot) bool { return !s.Active.Loading })
			got := m.RuntimeSnapshot()
			if got.Active.Alive || got.Active.PID != 0 || m.Snapshot().Connected {
				t.Fatal("obsolete setup launch resurrected child", got.Active)
			}
		})
	}
}
