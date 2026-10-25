package agent

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func waitRuntime(t *testing.T, m *Manager, condition func(RuntimeSnapshot) bool) RuntimeSnapshot {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m.Tick()
		value := m.RuntimeSnapshot()
		if condition(value) {
			return value
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("runtime condition timed out: %+v", m.RuntimeSnapshot())
	return RuntimeSnapshot{}
}
func TestRuntimeSnapshotLifecycleParkSwitchReuseAndRestart(t *testing.T) {
	m := testManager(t, nil)
	if got := m.RuntimeSnapshot(); got.Active != nil || len(got.Parked) != 0 || got.Transport != "stdio JSONL" {
		t.Fatal(got)
	}
	one, two := t.TempDir(), t.TempDir()
	path := filepath.Join(one, "saved.jsonl")
	if err := m.Open(one, path); err != nil {
		t.Fatal(err)
	}
	first := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active != nil && s.Active.State == "idle" }).Active
	if first.PID <= 0 || first.StartedAt.IsZero() || first.Uptime <= 0 || first.Executable == "" || first.Launcher != m.PiBinary() || first.ExitCode != nil || !first.Alive {
		t.Fatal(first)
	}
	if first.AutoRetry.Known || first.AutoRetry.Support != "unknown" {
		t.Fatal("invented initial retry", first.AutoRetry)
	}
	if err := m.SetAutoRetry(false); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.AutoRetry.Known && !s.Active.AutoRetry.Pending })
	if err := m.Open(two, ""); err != nil {
		t.Fatal(err)
	}
	parked := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" && len(s.Parked) == 1 })
	if parked.Parked[0].PID != first.PID || !parked.Parked[0].StartedAt.Equal(first.StartedAt) || !parked.Parked[0].AutoRetry.Known {
		t.Fatal(parked)
	}
	if err := m.Open(one, path); err != nil {
		t.Fatal(err)
	}
	resumed := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" }).Active
	if resumed.PID != first.PID || !resumed.StartedAt.Equal(first.StartedAt) || !resumed.AutoRetry.Known {
		t.Fatal(resumed)
	}
	if err := m.NewSession(one); err != nil {
		t.Fatal(err)
	}
	reused := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" && s.Active.SessionID == "s-2" }).Active
	if reused.PID != first.PID || !reused.StartedAt.Equal(first.StartedAt) || !reused.AutoRetry.Known {
		t.Fatal(reused)
	}
	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	restarted := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" }).Active
	if restarted.PID == first.PID || !restarted.StartedAt.After(first.StartedAt) || restarted.AutoRetry.Known {
		t.Fatal(restarted)
	}
}
func TestRuntimeStopStartKeepsParkedAndManagerUsable(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	first := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" }).Active.PID
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	stopped := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "stopped" && !s.Active.Alive })
	if stopped.Closed || stopped.Active.Uptime != 0 || stopped.Active.ExitCode == nil || len(stopped.Parked) != 1 || stopped.Parked[0].PID != first || !stopped.Parked[0].Alive {
		t.Fatal(stopped)
	}
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	started := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if started.Active.PID == stopped.Active.PID || len(started.Parked) != 1 {
		t.Fatal(started)
	}
}
func TestRuntimeLoadingAndStopBeforeSpawn(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	m := testManager(t, func(o *Options) {
		o.ResolveChildEnv = func(string) ([]string, error) { close(entered); <-release; return nil, nil }
	})
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	<-entered
	loading := m.RuntimeSnapshot().Active
	if loading.State != "loading" || loading.PID != 0 || !loading.StartedAt.IsZero() || loading.Uptime != 0 {
		t.Fatal(loading)
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := m.WaitForIdle(ctx); err != nil {
		t.Fatal(err)
	}
	stopped := m.RuntimeSnapshot().Active
	if stopped.State != "stopped" || stopped.Alive || stopped.PID != 0 {
		t.Fatal(stopped)
	}
}
func TestRuntimeFailedSpawnAndUnexpectedExitAreObserved(t *testing.T) {
	m := testManager(t, func(o *Options) { o.PiBinary = filepath.Join(t.TempDir(), "missing-pi") })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	failed := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "failed" }).Active
	if failed.PID != 0 || failed.Alive || failed.Uptime != 0 || !failed.StartedAt.IsZero() || failed.Error == "" || failed.Launcher != m.PiBinary() {
		t.Fatal(failed)
	}
	live := testManager(t, nil)
	if err := live.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, live, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if err := live.Send("exit_process", nil); err != nil {
		t.Fatal(err)
	}
	exited := waitRuntime(t, live, func(s RuntimeSnapshot) bool { return s.Active.State == "exited" }).Active
	if exited.Alive || exited.ExitCode == nil || *exited.ExitCode != 7 || exited.Uptime != 0 {
		t.Fatal(exited)
	}
}
func TestRuntimeStderrIsBoundedRedactedAndDetached(t *testing.T) {
	m := testManager(t, func(o *Options) { o.EnvOverrides["ORBIT_DYNAMIC_SECRET"] = "sensitive-value" })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if err := m.Send("emit_stderr", nil); err != nil {
		t.Fatal(err)
	}
	value := waitRuntime(t, m, func(s RuntimeSnapshot) bool {
		return len(s.Active.RecentStderr) == 8 && strings.Contains(s.Active.RecentStderr[7], "line-11")
	})
	raw, _ := json.Marshal(value)
	if strings.Contains(string(raw), "sensitive-value") || !strings.Contains(string(raw), "[redacted]") {
		t.Fatal(string(raw))
	}
	value.Active.RecentStderr[0] = "modified"
	if m.RuntimeSnapshot().Active.RecentStderr[0] == "modified" {
		t.Fatal("snapshot aliases diagnostics")
	}
}
func TestRetrySetterWaitsForAckAndRejectsConcurrentRequests(t *testing.T) {
	m := testManager(t, func(o *Options) { o.EnvOverrides["ORBIT_RETRY_TEST_MODE"] = "delayed" })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if err := m.SetAutoRetry(true); err != nil {
		t.Fatal(err)
	}
	pending := m.RuntimeSnapshot().Active.AutoRetry
	if !pending.Pending || pending.Known {
		t.Fatal("optimistic value", pending)
	}
	if err := m.SetAutoRetry(false); err == nil {
		t.Fatal("concurrent setter accepted")
	}
	confirmed := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.AutoRetry.Known && !s.Active.AutoRetry.Pending }).Active.AutoRetry
	if confirmed.Support != "supported" || !confirmed.Enabled {
		t.Fatal(confirmed)
	}
	if err := m.Send(pirpc.SetAutoRetry, map[string]any{"enabled": false}); err != nil {
		t.Fatal(err)
	}
	if previous := m.RuntimeSnapshot().Active.AutoRetry; !previous.Pending || !previous.Known || !previous.Enabled {
		t.Fatal("pending request changed confirmed value", previous)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool {
		return s.Active.AutoRetry.Known && !s.Active.AutoRetry.Enabled && !s.Active.AutoRetry.Pending
	})
}

func TestRetryRejectedRequestPreservesKnownValueButTimeoutDoesNot(t *testing.T) {
	for _, mode := range []string{"fail_after_first", "timeout_after_first"} {
		t.Run(mode, func(t *testing.T) {
			m := testManager(t, func(o *Options) {
				o.EnvOverrides["ORBIT_RETRY_TEST_MODE"] = mode
				o.CommandTimeout = 100 * time.Millisecond
			})
			if err := m.Open(t.TempDir(), ""); err != nil {
				t.Fatal(err)
			}
			waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
			if err := m.SetAutoRetry(true); err != nil {
				t.Fatal(err)
			}
			waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.AutoRetry.Known && !s.Active.AutoRetry.Pending })
			if err := m.SetAutoRetry(false); err != nil {
				t.Fatal(err)
			}
			got := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return !s.Active.AutoRetry.Pending && s.Active.AutoRetry.Error != "" }).Active.AutoRetry
			if !got.Enabled || got.Support != "supported" || got.Known != (mode == "fail_after_first") {
				t.Fatal(got)
			}
		})
	}
}
func TestRetryFailureDoesNotInventValueOrSupport(t *testing.T) {
	for _, mode := range []string{"unknown", "error"} {
		t.Run(mode, func(t *testing.T) {
			m := testManager(t, func(o *Options) { o.EnvOverrides["ORBIT_RETRY_TEST_MODE"] = mode })
			if err := m.Open(t.TempDir(), ""); err != nil {
				t.Fatal(err)
			}
			waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
			if err := m.SetAutoRetry(true); err != nil {
				t.Fatal(err)
			}
			retry := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return !s.Active.AutoRetry.Pending && s.Active.AutoRetry.Error != "" }).Active.AutoRetry
			expected := "unknown"
			if mode == "unknown" {
				expected = "unsupported"
			}
			if retry.Known || retry.Enabled || retry.Support != expected {
				t.Fatal(retry)
			}
		})
	}
}
func TestRuntimeSnapshotConcurrentReaders(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	var readers sync.WaitGroup
	for range 5 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for range 100 {
				m.RuntimeSnapshot()
				m.Snapshot()
				m.Tick()
			}
		}()
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	readers.Wait()
}
func TestDiagnosticCredentialRedaction(t *testing.T) {
	for _, line := range []string{"Authorization: Bearer token-value", "api_key=key-value", `{"accessToken":"token-value"}`, `{"password":"password-value with spaces"}`, "PASSWORD=password-value", "known-secret-suffix"} {
		got := redactDiagnostic(line, []string{"known-secret", "known-secret-suffix"})
		if strings.Contains(got, "token-value") || strings.Contains(got, "key-value") || strings.Contains(got, "password-value") || strings.Contains(got, "suffix") {
			t.Fatal(line, got)
		}
	}
	if got := redactDiagnostic("private-material", []string{"HEADER\nprivate-material\nFOOTER"}); got != "[redacted]" {
		t.Fatal("multiline credential leaked", got)
	}
}

func TestRuntimeBusyProcessRemainsVisibleWhenParked(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" })
	if err := m.Submit("hold fixture", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	running := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "busy" }).Active
	if !running.Busy {
		t.Fatal(running)
	}
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	parked := waitRuntime(t, m, func(s RuntimeSnapshot) bool { return s.Active.State == "idle" && len(s.Parked) == 1 })
	if parked.Parked[0].PID != running.PID || !parked.Parked[0].Busy || parked.Parked[0].State != "busy" || !m.AnyBusy() {
		t.Fatal(parked)
	}
}
