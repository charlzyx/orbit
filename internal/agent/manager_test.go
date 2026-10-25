package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/pirpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAgentProcess(t *testing.T) {
	if os.Getenv("ORBIT_AGENT_HELPER") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	sessionID := "s-1"
	number := 1
	path := ""
	running := false
	compacted := false
	retryRequests := 0
	var heldAbortClear map[string]any
	var heldAbortResponse map[string]any
	queuePending, clearHeld := false, false
	var mu sync.Mutex
	write := func(value any) {
		mu.Lock()
		defer mu.Unlock()
		raw, _ := json.Marshal(value)
		fmt.Fprintln(os.Stdout, string(raw))
	}
	for scanner.Scan() {
		var c map[string]any
		if json.Unmarshal(scanner.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		response := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case "emit_abort_queue":
			queuePending = true
			write(map[string]any{"type": "queue_update", "steering": []string{"old steer"}, "followUp": []string{"old follow"}})
		case "emit_fresh_queue":
			write(map[string]any{"type": "queue_update", "steering": []string{}, "followUp": []string{"new queued text"}})
		case "clear_queue":
			response["data"] = map[string]any{"steering": []string{}, "followUp": []string{}}
			if queuePending {
				response["data"] = map[string]any{"steering": []string{"old steer"}, "followUp": []string{"old follow"}}
				queuePending = false
			}
			if os.Getenv("ORBIT_CLEAR_UNTIL_ABORT") == "1" && !clearHeld {
				heldAbortClear = response
				clearHeld = true
				continue
			}
		case "release_abort_clear":
			if !running && heldAbortClear != nil {
				write(heldAbortClear)
				heldAbortClear = nil
			}
			if !running && heldAbortResponse != nil {
				write(heldAbortResponse)
				heldAbortResponse = nil
			}
		case "emit_stderr":
			for i := 0; i < 12; i++ {
				fmt.Fprintf(os.Stderr, "line-%d %s\n", i, os.Getenv("ORBIT_DYNAMIC_SECRET"))
			}
		case "exit_process":
			os.Exit(7)
		case "set_auto_retry":
			retryRequests++
			switch os.Getenv("ORBIT_RETRY_TEST_MODE") {
			case "unknown":
				response["success"] = false
				response["error"] = "Unknown command: set_auto_retry"
			case "error":
				response["success"] = false
				response["error"] = "temporary settings write failed"
			case "delayed":
				time.Sleep(100 * time.Millisecond)
			case "fail_after_first":
				if retryRequests > 1 {
					response["success"] = false
					response["error"] = "temporary settings write failed"
				}
			case "timeout_after_first":
				if retryRequests > 1 {
					time.Sleep(300 * time.Millisecond)
				}
			}
		case "extension_ui_response":
			write(map[string]any{"type": "dialog_answered", "id": c["id"], "cancelled": c["cancelled"]})
			continue
		case "inspect_env":
			response["data"] = map[string]any{"value": os.Getenv("ORBIT_DYNAMIC_SECRET")}
		case "get_state":
			if delay, _ := strconv.Atoi(os.Getenv("ORBIT_STATE_DELAY_MS")); delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			response["data"] = map[string]any{"sessionId": sessionID, "sessionFile": path, "isStreaming": running, "thinkingLevel": "off", "autoCompactionEnabled": true, "steeringMode": "one-at-a-time", "followUpMode": "one-at-a-time", "model": map[string]any{"id": "mock", "provider": "test", "name": "Fixture"}}
		case "get_messages":
			response["data"] = map[string]any{"messages": []any{}}
		case "get_available_models":
			response["data"] = map[string]any{"models": []any{}}
		case "get_available_thinking_levels":
			response["data"] = map[string]any{"levels": []string{"off"}}
		case "get_commands":
			response["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			response["data"] = map[string]any{"entries": []any{}}
		case "get_session_stats":
			response["data"] = map[string]any{"tokens": map[string]any{"total": 0}, "cost": 0}
			if compacted {
				response["data"] = map[string]any{"tokens": map[string]any{"total": 37}, "cost": 0, "contextUsage": map[string]any{"tokens": nil, "percent": nil, "contextWindow": 200000}}
			}
		case "compact":
			compacted = true
			write(map[string]any{"type": "compaction_start", "reason": "manual"})
			write(map[string]any{"type": "compaction_end", "reason": "manual", "result": map[string]any{"summary": "Fixture"}, "aborted": false})
		case "auth.list", "quota.list":
			response["success"] = false
			response["error"] = "Unknown command: " + kind
		case "new_session":
			if delay, _ := strconv.Atoi(os.Getenv("ORBIT_NEW_DELAY_MS")); delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
			if os.Getenv("ORBIT_NEW_CANCEL") == "1" {
				response["data"] = map[string]any{"cancelled": true}
				break
			}
			number++
			sessionID = "s-" + strconv.Itoa(number)
			response["data"] = map[string]any{"cancelled": false}
		case "switch_session":
			path, _ = c["sessionPath"].(string)
			sessionID = "switched"
			response["data"] = map[string]any{"cancelled": false}
		case "prompt":
			running = true
			response["data"] = map[string]any{"disposition": "started"}
			write(response)
			write(map[string]any{"type": "agent_start"})
			if c["message"] == "dialog" {
				write(map[string]any{"type": "extension_ui_request", "id": "dialog-1", "method": "confirm", "title": "Proceed?"})
			}
			continue
		case "abort":
			running = false
			if os.Getenv("ORBIT_ABORT_REJECT") == "1" {
				response["success"] = false
				response["error"] = "original task abort rejection"
				heldAbortResponse = response
			} else {
				write(response)
			}
			write(map[string]any{"type": "agent_settled", "aborted": true})
			continue
		case "emit_dialog":
			write(response)
			write(map[string]any{"type": "extension_ui_request", "id": "parked-dialog", "method": "select", "options": []string{"Allow", "Deny"}})
			continue
		}
		write(response)
	}
	os.Exit(0)
}
func testManager(t *testing.T, change func(*Options)) *Manager {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	options := Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestAgentProcess$", "--"}, Home: t.TempDir(), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_AGENT_HELPER": "1"}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second}
	if change != nil {
		change(&options)
	}
	manager := NewManager(options)
	t.Cleanup(func() { manager.Close() })
	return manager
}
func waitFor(t *testing.T, m *Manager, condition func(Snapshot) bool) Snapshot {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m.Tick()
		snapshot := m.Snapshot()
		if condition(snapshot) {
			return snapshot
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition timed out: %+v", m.Snapshot())
	return Snapshot{}
}
func TestManagerBootstrapActionsAndIdleReuse(t *testing.T) {
	m := testManager(t, nil)
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	initial := waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	if initial.ActiveSession.ID != "s-1" {
		t.Fatal(initial)
	}
	m.mu.Lock()
	pid := m.sessions[m.active].client.PID()
	m.mu.Unlock()
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-2" })
	m.mu.Lock()
	nextPID := m.sessions[m.active].client.PID()
	m.mu.Unlock()
	if pid != nextPID {
		t.Fatal("idle process not reused")
	}
	if err := m.Submit("hold", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	if err := m.Abort(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Running })
}
func TestManagerParksRunningProcessAndReapsIdle(t *testing.T) {
	m := testManager(t, func(o *Options) { o.ParkedTTL = 40 * time.Millisecond })
	first, second := t.TempDir(), t.TempDir()
	if err := m.Open(first, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("hold", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	m.mu.Lock()
	oldKey := m.active
	old := m.sessions[oldKey]
	m.mu.Unlock()
	if err := m.Open(second, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.Workspace == second })
	time.Sleep(60 * time.Millisecond)
	m.Tick()
	m.mu.Lock()
	_, exists := m.sessions[oldKey]
	m.mu.Unlock()
	if !exists || !old.client.Alive() {
		t.Fatal("busy background process evicted")
	}
	_, err := m.callClient(old.client, pirpcCommand("abort"))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		m.Tick()
		m.mu.Lock()
		_, exists = m.sessions[oldKey]
		m.mu.Unlock()
		if !exists {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if exists {
		t.Fatal("idle background process not reaped")
	}
}
func TestManagerDialogResponseAndSnapshotIsolation(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("dialog", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	snapshot := waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) > 0 })
	snapshot.Dialogs[0].Title = "mutated"
	if m.Snapshot().Dialogs[0].Title == "mutated" {
		t.Fatal("snapshot is mutable")
	}
	yes := true
	if err := m.RespondDialog("dialog-1", "", &yes, false); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 0 })
	if err := m.RespondDialog("missing", "", nil, true); err == nil {
		t.Fatal("stale dialog accepted")
	}
}
func TestManagerConcurrentTicksAndSnapshots(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				m.Tick()
				_ = m.Snapshot()
				m.ReloadSessions()
			}
		}()
	}
	wg.Wait()
	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
}
func TestManagerModePersistence(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	if err := m.SetAccessMode("supervised"); err != nil {
		t.Fatal(err)
	}
	if err := m.SetWorkflowMode("plan"); err != nil {
		t.Fatal(err)
	}
	if readAccess(m.options.Home) != "supervised" {
		t.Fatal("access mode not saved")
	}
	var modes map[string]string
	raw, err := os.ReadFile(filepath.Join(m.options.Home, ".orbit-pi", "workflow.json"))
	if err != nil || json.Unmarshal(raw, &modes) != nil || modes["s-1"] != "plan" {
		t.Fatal(string(raw), err)
	}
	if err := m.SetAccessMode("invalid"); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func pirpcCommand(kind string) pirpc.Command { return pirpc.NewCommand(kind, nil) }

func TestDeleteIsRecoverableAndDoesNotTouchOtherSessions(t *testing.T) {
	m := testManager(t, nil)
	group := filepath.Join(m.options.SessionDir, "group")
	if err := os.MkdirAll(group, 0700); err != nil {
		t.Fatal(err)
	}
	raw := []byte(`{"type":"session","version":3,"id":"s","cwd":"/workspace","timestamp":"2026-01-01T00:00:00Z"}` + "\n" + `{"type":"message","message":{"role":"user","content":"hello","timestamp":1767225600000}}` + "\n")
	path := filepath.Join(group, "delete.jsonl")
	other := filepath.Join(group, "keep.jsonl")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(other, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.DeleteSession(path); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("source still present", err)
	}
	backups, err := os.ReadDir(filepath.Join(m.options.Home, ".orbit-pi", "deleted-sessions"))
	if err != nil || len(backups) != 1 {
		t.Fatal(backups, err)
	}
	backup, _ := os.ReadFile(filepath.Join(m.options.Home, ".orbit-pi", "deleted-sessions", backups[0].Name()))
	if string(backup) != string(raw) {
		t.Fatal("recoverable copy changed")
	}
	if _, err := os.Stat(other); err != nil {
		t.Fatal("unrelated session changed", err)
	}
}

func TestChildEnvironmentResolvedEverySpawnAndExplicitOverridesWin(t *testing.T) {
	var calls atomic.Int32
	m := testManager(t, func(o *Options) {
		o.ResolveChildEnv = func(workspace string) ([]string, error) {
			calls.Add(1)
			return []string{"ORBIT_DYNAMIC_SECRET=resolved-secret"}, nil
		}
		o.EnvOverrides["ORBIT_DYNAMIC_SECRET"] = "explicit-secret"
	})
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	m.mu.Lock()
	client := m.sessions[m.active].client
	m.mu.Unlock()
	response, err := m.callClient(client, pirpcCommand("inspect_env"))
	if err != nil {
		t.Fatal(err)
	}
	var data struct{ Value string }
	if response.DecodeData(&data) != nil || data.Value != "explicit-secret" {
		t.Fatal("explicit override did not win")
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	entry.state.snapshot.ConnectionError = "resolved-secret explicit-secret"
	entry.state.snapshot.Diagnostics = []string{"resolved-secret explicit-secret"}
	m.mu.Unlock()
	snapshot := m.Snapshot()
	if strings.Contains(snapshot.ConnectionError, "secret") || strings.Contains(strings.Join(snapshot.Diagnostics, " "), "secret") {
		t.Fatal("resolved or overridden secret leaked into diagnostics")
	}

	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if calls.Load() != 2 {
		t.Fatal(calls.Load())
	}
}

func TestSetPiBinaryValidatesBeforeReplacingConfiguration(t *testing.T) {
	m := testManager(t, nil)
	previous := m.PiBinary()
	if err := m.SetPiBinary(filepath.Join(t.TempDir(), "missing-pi")); err == nil {
		t.Fatal("missing executable accepted")
	}
	if m.PiBinary() != previous {
		t.Fatal("invalid executable replaced working configuration")
	}
	if err := m.SetPiBinary(previous); err != nil {
		t.Fatal(err)
	}
}

func TestLiveSessionPathsIncludesParkedLoadingAndReturnsDetachedSet(t *testing.T) {
	m := &Manager{active: "active", sessions: map[string]*liveSession{
		"active":  {state: newState("workspace", "/sessions/b.jsonl")},
		"parked":  {state: newState("workspace", "/sessions/a.jsonl")},
		"loading": {state: newState("workspace", "/sessions/c.jsonl"), loading: true},
		"same":    {state: newState("workspace", "/sessions/a.jsonl")},
		"draft":   {state: newState("workspace", "")},
	}}
	paths := m.LiveSessionPaths()
	if strings.Join(paths, ",") != "/sessions/a.jsonl,/sessions/b.jsonl,/sessions/c.jsonl" {
		t.Fatal(paths)
	}
	paths[0] = "changed"
	if m.LiveSessionPaths()[0] != "/sessions/a.jsonl" {
		t.Fatal("returned manager-owned slice")
	}
	m.closed = true
	if len(m.LiveSessionPaths()) != 0 {
		t.Fatal("closed manager retained live paths")
	}
}

func TestAnyBusyIncludesEveryBackgroundLifecyclePhase(t *testing.T) {
	phases := map[string]func(*liveSession){
		"loading":    func(s *liveSession) { s.loading = true },
		"awaiting":   func(s *liveSession) { s.state.awaiting = true },
		"running":    func(s *liveSession) { s.state.snapshot.Running = true },
		"compacting": func(s *liveSession) { s.state.snapshot.Compacting = true },
		"retrying":   func(s *liveSession) { s.state.snapshot.Retrying = true },
	}
	for phase, set := range phases {
		t.Run(phase, func(t *testing.T) {
			parked := &liveSession{state: newState("workspace", "saved")}
			m := &Manager{active: "active", sessions: map[string]*liveSession{
				"active": {state: newState("workspace", "active")}, "parked": parked,
			}}
			if m.AnyBusy() {
				t.Fatal("idle sessions marked busy")
			}
			set(parked)
			if !m.AnyBusy() {
				t.Fatal("background phase ignored")
			}
			m.closed = true
			if m.AnyBusy() {
				t.Fatal("closed manager marked busy")
			}
		})
	}
}

func TestProcessExitClearsCompactionAndRetryBusyFlags(t *testing.T) {
	state := newState("workspace", "session")
	state.snapshot.Compacting, state.snapshot.Retrying, state.snapshot.Running = true, true, true
	state.awaiting = true
	state.apply(pirpc.Event{Type: "process_exited"})
	if (&liveSession{state: state}).busy() {
		t.Fatal("exited process still busy")
	}
}

func TestManualCompactionRefreshesStatsWithoutAgentSettled(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Send(pirpc.Compact, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		var stats struct {
			Tokens       struct{ Total int }
			ContextUsage struct {
				Tokens        *int
				Percent       *float64
				ContextWindow int
			}
		}
		return json.Unmarshal(s.Stats, &stats) == nil && stats.Tokens.Total == 37 && stats.ContextUsage.ContextWindow == 200000 && stats.ContextUsage.Tokens == nil && stats.ContextUsage.Percent == nil && !s.Compacting
	})
}
