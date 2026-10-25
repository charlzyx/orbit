package orbitui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// A credential-free JSONL subprocess; every phase reaches the real Manager
// reducer and Controller.View/tickWindowless. No replacement Tick bool is used.
func TestHeartbeatRPCChild(t *testing.T) {
	if os.Getenv("ORBIT_HEARTBEAT_CHILD") != "1" {
		return
	}
	write := func(v any) { _ = json.NewEncoder(os.Stdout).Encode(v) }
	running, compacting := false, false
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var cmd map[string]any
		if json.Unmarshal(scanner.Bytes(), &cmd) != nil {
			continue
		}
		kind, _ := cmd["type"].(string)
		reply := map[string]any{"type": "response", "id": cmd["id"], "command": kind, "success": true}
		switch kind {
		case "get_state":
			reply["data"] = map[string]any{"sessionId": fmt.Sprintf("heartbeat-%d", os.Getpid()), "isStreaming": running, "isCompacting": compacting}
		case "get_messages":
			reply["data"] = map[string]any{"messages": []any{}}
		case "get_available_models":
			reply["data"] = map[string]any{"models": []any{}}
		case "get_available_thinking_levels":
			reply["data"] = map[string]any{"levels": []string{"off"}}
		case "get_commands":
			reply["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			reply["data"] = map[string]any{"entries": []any{}}
		case "auth.list", "quota.list":
			reply["success"], reply["error"] = false, "Unknown command: "+kind
		case "heartbeat_phase":
			switch cmd["phase"] {
			case "busy":
				running = true
				write(map[string]any{"type": "agent_start"})
			case "retry":
				write(map[string]any{"type": "auto_retry_start"})
			case "compaction":
				compacting = true
				write(map[string]any{"type": "compaction_start"})
			case "idle":
				running, compacting = false, false
				write(map[string]any{"type": "agent_settled"})
				write(map[string]any{"type": "compaction_end", "aborted": true})
			}
		}
		write(reply)
	}
	os.Exit(0)
}

func newHeartbeatRPCController(t *testing.T) (*Controller, func(func(agent.Snapshot) bool)) {
	t.Helper()
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
	x.Agent = agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestHeartbeatRPCChild$", "--"}, Home: x.Store.Home, SessionDir: filepath.Join(x.Store.PiDir, "sessions"), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_HEARTBEAT_CHILD": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
	if err := x.Agent.Open(x.workspace, ""); err != nil {
		t.Fatal(err)
	}
	x.UI.Prefs.ReduceMotion = true
	x.UI.page, x.UI.settingsPage = "settings", "appearance"
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
		t.Fatalf("windowless RPC drain did not reach phase: %+v", x.Agent.Snapshot())
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !x.Agent.AnyBusy() })
	return x, wait
}

// Retry and compaction are explicit source is_active cases even when Running is
// false. Parked busy is a documented Go extension of that source predicate.
func TestHeartbeatViewUsesManagerActivityNotForegroundRunning(t *testing.T) {
	for _, phase := range []string{"idle", "busy", "retry", "compaction", "parked"} {
		t.Run(phase, func(t *testing.T) {
			x, wait := newHeartbeatRPCController(t)
			if phase != "idle" {
				emit := phase
				if phase == "parked" {
					emit = "busy"
				}
				if err := x.Agent.Send("heartbeat_phase", map[string]any{"phase": emit}); err != nil {
					t.Fatal(err)
				}
				wait(func(s agent.Snapshot) bool {
					return emit == "busy" && s.Running || emit == "retry" && s.Retrying || emit == "compaction" && s.Compacting
				})
			}
			if phase == "parked" {
				old := x.Agent.Snapshot().RuntimeID
				if err := x.Agent.Open(x.workspace, ""); err != nil {
					t.Fatal(err)
				}
				wait(func(s agent.Snapshot) bool { return s.Connected && s.RuntimeID != old && !s.Running })
				parked := x.Agent.RuntimeSnapshot().Parked
				if len(parked) != 1 || !parked[0].Busy {
					t.Fatalf("fixture lost running parked runtime: %+v", parked)
				}
			}
			want := heartbeatDelay(phase != "idle")
			if active := x.tickWindowless(); active != (phase != "idle") {
				t.Fatalf("hidden route returned active=%v for %s", active, phase)
			}
			var frames []time.Time
			tt := ui.NewTester(func(c *ui.Context) {
				frames = append(frames, c.Now())
				x.View(c)
			}, 1000, 700)
			if phase != "busy" && x.UI.Model.Running {
				t.Fatal("fixture did not distinguish activity from foreground Running")
			}
			before, at := len(frames), tt.Now()
			tt.Advance(want - time.Millisecond)
			if len(frames) != before {
				t.Fatalf("foreground heartbeat fired early: %v", frames)
			}
			tt.Advance(time.Millisecond)
			if len(frames) != before+1 || frames[len(frames)-1].Sub(at) != want {
				t.Fatalf("Controller.View did not schedule %v: before=%d frames=%v", want, before, frames)
			}
			if phase == "retry" || phase == "compaction" || phase == "busy" {
				if err := x.Agent.Send("heartbeat_phase", map[string]any{"phase": "idle"}); err != nil {
					t.Fatal(err)
				}
				wait(func(s agent.Snapshot) bool { return !s.Running && !s.Retrying && !s.Compacting })
				// Manager retains its pre-drain activity for the settle tick;
				// the next drain must return to the 90 ms idle interval.
				if x.tickWindowless() {
					t.Fatal("settled runtime never returned to idle cadence")
				}
			}
		})
	}
}
