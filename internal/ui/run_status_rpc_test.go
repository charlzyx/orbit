package orbitui

import (
	"bufio"
	"encoding/json"
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// Actual JSONL subprocess: synthetic provider events traverse Client -> Manager
// -> adaptSnapshot -> the mounted chat button -> Controller.Action -> RPC.
func TestRunStatusRPCChild(t *testing.T) {
	if os.Getenv("ORBIT_STATUS_CHILD") != "1" {
		return
	}
	write := func(v any) { _ = json.NewEncoder(os.Stdout).Encode(v) }
	scanner := bufio.NewScanner(os.Stdin)
	reject := false
	statsRequests, compactRequests := 0, 0
	var heldCompact map[string]any
	compacted := false
	for scanner.Scan() {
		var cmd map[string]any
		if json.Unmarshal(scanner.Bytes(), &cmd) != nil {
			continue
		}
		kind, _ := cmd["type"].(string)
		reply := map[string]any{"type": "response", "id": cmd["id"], "command": kind, "success": true}
		switch kind {
		case "get_state":
			reply["data"] = map[string]any{"sessionId": fmt.Sprintf("status-%d", os.Getpid())}
		case "get_session_stats":
			statsRequests++
			contextUsage := map[string]any{"tokens": 800, "contextWindow": 1000, "percent": 80}
			if compacted {
				contextUsage["tokens"] = nil
				contextUsage["percent"] = nil
			}
			reply["data"] = map[string]any{"tokens": map[string]any{"input": 10, "output": 20, "cacheRead": 30, "cacheWrite": 40, "total": 100}, "cost": .001, "contextUsage": contextUsage, "statsRequests": statsRequests, "compactRequests": compactRequests}
		case "compact":
			compactRequests++
			heldCompact = reply
			write(map[string]any{"type": "compaction_start"})
			continue
		case "release_context_compact":
			if heldCompact != nil {
				if reject {
					heldCompact["success"], heldCompact["error"] = false, "fixture compact rejected"
				} else {
					compacted = true
				}
				write(map[string]any{"type": "compaction_end", "aborted": reject, "result": map[string]any{"summary": "fixture summary"}})
				write(heldCompact)
				heldCompact = nil
			}
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
		case "status_test_event":
			reject, _ = cmd["reject"].(bool)
			write(cmd["event"])
		case "abort_retry":
			if reject {
				reply["success"], reply["error"] = false, "fixture cancel rejected"
			} else {
				write(map[string]any{"type": "auto_retry_end", "success": true})
			}
		}
		write(reply)
	}
	os.Exit(0)
}

func TestRunStatusRPCMountedCancellationAndParking(t *testing.T) {
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
		t.Fatalf("status RPC timeout: %+v", x.Agent.Snapshot())
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !x.Agent.AnyBusy() })
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	emit := func(event map[string]any, reject bool) {
		t.Helper()
		if err := x.Agent.Send("status_test_event", map[string]any{"event": event, "reject": reject}); err != nil {
			t.Fatal(err)
		}
	}

	x.UI.catalog, err = orbitassets.Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	for i, kind := range []string{"missing", "null", "literal"} {
		event := map[string]any{"type": "auto_retry_start", "attempt": 10 + i}
		if kind == "null" {
			event["errorMessage"] = nil
		}
		if kind == "literal" {
			event["errorMessage"] = "transient error"
		}
		emit(event, false)
		wait(func(s agent.Snapshot) bool { return s.RetryDetail.Attempt == uint64(10+i) })
		tt.Frame()
		text := strings.Join(tt.Texts(), "|")
		if kind == "literal" {
			if !strings.Contains(text, "transient error") || strings.Contains(text, x.UI.Tr("events.transient_error")) {
				t.Fatal("provider literal translated", text)
			}
		} else if !strings.Contains(text, x.UI.Tr("events.transient_error")) || strings.Contains(text, "transient error") {
			t.Fatal("fallback not localized", kind, text)
		}
	}
	x.UI.catalog = x.UI.english
	emit(map[string]any{"type": "auto_retry_start", "attempt": 2, "maxAttempts": 4, "errorMessage": "provider unavailable"}, true)
	wait(func(s agent.Snapshot) bool { return s.RetryDetail.Active && s.RetryDetail.Attempt == 2 })
	tt.Frame()
	if _, ok := tt.Find("Retrying — attempt 2 of 4 · provider unavailable"); !ok {
		t.Fatal(tt.Texts())
	}
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.ConnectionError == "fixture cancel rejected" })
	tt.Frame()
	if _, ok := tt.Find("Cancel retry"); ok {
		t.Fatal("failed RPC resurrected cancelled strip")
	}
	if !x.Agent.Snapshot().Retrying || !x.Agent.AnyBusy() {
		t.Fatal("optimistic UI faked idle runtime")
	}
	if !tt.HasText("fixture cancel rejected") {
		t.Fatal("cancel error hidden", tt.Texts())
	}
	previous := x.Agent.Snapshot().RetryDetail.Revision
	emit(map[string]any{"type": "auto_retry_start", "attempt": 2, "maxAttempts": 4, "errorMessage": "provider unavailable"}, true)
	wait(func(s agent.Snapshot) bool { return s.RetryDetail.Revision > previous })
	tt.Frame()
	if _, ok := tt.Find("Cancel retry"); !ok {
		t.Fatal("identical new retry event remained cancelled")
	}
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Cancel retry"); ok {
		t.Fatal("second cancel did not hide")
	}
	emit(map[string]any{"type": "auto_retry_start", "attempt": 3, "errorMessage": "try again"}, false)
	wait(func(s agent.Snapshot) bool { return s.RetryDetail.Attempt == 3 })
	tt.Frame()
	if err := tt.Click("Cancel retry"); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return !s.Retrying && !s.RetryDetail.Active })
	tt.Frame()
	if _, ok := tt.Find("Cancel retry"); ok {
		t.Fatal("observed end retained cancel")
	}
	emit(map[string]any{"type": "auto_retry_start", "attempt": 4}, false)
	wait(func(s agent.Snapshot) bool { return s.RetryDetail.Active })
	old := x.Agent.Snapshot().RuntimeID
	if err := x.Agent.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.RuntimeID != old && !s.Retrying })
	tt.Frame()
	if _, ok := tt.Find("Run status"); ok {
		t.Fatal("parked retry leaked into active session")
	}
	emit(map[string]any{"type": "compaction_start"}, false)
	wait(func(s agent.Snapshot) bool { return s.Compacting })
	tt.Frame()
	if _, ok := tt.Find(x.UI.Tr("view.preparing_conversation_context")); !ok {
		t.Fatal("RPC compaction missing", tt.Texts())
	}
	emit(map[string]any{"type": "auto_retry_start"}, false)
	wait(func(s agent.Snapshot) bool { return s.RetryDetail.Active })
	tt.Frame()
	if err := x.Agent.Stop(); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return !s.Connected })
	tt.Frame()
	if _, ok := tt.Find("Run status"); ok {
		t.Fatal("stopped runtime retained strip")
	}
}
