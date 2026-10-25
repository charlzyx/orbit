package orbitui

import (
	"encoding/json"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"os"
	"testing"
	"time"
)

func TestContextMeterRPCMountedRefreshCompactAndFailure(t *testing.T) {
	if err := LoadAssets(); err != nil {
		t.Fatal(err)
	}
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	_ = x.Agent.Close()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestRunStatusRPCChild$", "--"}, Home: x.Store.Home, Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_STATUS_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	wait := func(pred func(agent.Snapshot) bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			x.tickWindowless()
			if pred(x.Agent.Snapshot()) {
				return
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatalf("context RPC timeout %+v", x.Agent.Snapshot())
	}
	count := func(s agent.Snapshot, key string) int {
		var stats map[string]json.RawMessage
		_ = json.Unmarshal(s.Stats, &stats)
		var n int
		_ = json.Unmarshal(stats[key], &n)
		return n
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && !x.Agent.AnyBusy() && count(s, "statsRequests") > 0 })
	before := count(x.Agent.Snapshot(), "statsRequests")
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Context usage"); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return count(s, "statsRequests") > before })
	tt.Frame()
	if _, ok := tt.Find("80% Full"); !ok {
		t.Fatal("stats did not reach mounted detail", tt.Texts())
	}
	// Fail the first compaction after observing the real pending command.
	if err := x.Agent.Send("status_test_event", map[string]any{"event": map[string]any{"type": "session_info_changed", "name": "ready"}, "reject": true}); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.ActiveSession.Name == "ready" })
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	if !x.Agent.Snapshot().Compacting || !x.Agent.AnyBusy() {
		t.Fatal("compact button did not reserve runtime")
	}
	tt.Frame()
	if _, ok := tt.Find("Compacting…"); !ok {
		t.Fatal(tt.Texts())
	}
	for i := 0; i < 3; i++ {
		if err := tt.Click("Compact context"); err != nil {
			t.Fatal(err)
		}
	}
	x.Action(Action{Kind: "RefreshContextStats"})
	wait(func(s agent.Snapshot) bool { return count(s, "compactRequests") == 1 })
	if err := x.Agent.Send("release_context_compact", nil); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return !s.Compacting && s.ConnectionError == "fixture compact rejected" })
	tt.Frame()
	if _, ok := tt.Find("Compact now"); !ok {
		t.Fatal("failure did not allow retry", tt.Texts())
	}
	if err := x.Agent.Send("status_test_event", map[string]any{"event": map[string]any{"type": "session_info_changed", "name": "retry"}}); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.ActiveSession.Name == "retry" })
	if err := tt.Click("Compact context"); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "RefreshContextStats"})
	wait(func(s agent.Snapshot) bool { return count(s, "compactRequests") == 2 })
	if err := x.Agent.Send("release_context_compact", nil); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool {
		return !s.Compacting && count(s, "statsRequests") > before+2 && !runtimeUsage(s).ContextAvailable && runtimeUsage(s).ContextWindow == 1000
	})
	tt.Frame()
	if _, ok := tt.Find("Unknown"); !ok {
		t.Fatal("post-compaction estimate lied", tt.Texts())
	}
	if _, ok := tt.Find("of 1K tokens"); !ok {
		t.Fatal("known capacity lost", tt.Texts())
	}
}
