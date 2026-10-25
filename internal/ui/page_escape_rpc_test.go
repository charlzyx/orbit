package orbitui

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// This is a credential-free JSONL child process. It observes the production
// Controller -> Manager -> RPC path, not just a replacement Abort callback.
func TestPageEscapeRPCChild(t *testing.T) {
	if os.Getenv("ORBIT_PAGE_ESCAPE_CHILD") != "1" {
		return
	}
	log, err := os.OpenFile(os.Getenv("ORBIT_PAGE_ESCAPE_LOG"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		os.Exit(2)
	}
	defer log.Close()
	write := func(v any) { b, _ := json.Marshal(v); fmt.Fprintln(os.Stdout, string(b)) }
	running := false
	queued := false
	var heldClear map[string]any
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var cmd map[string]any
		if json.Unmarshal(scanner.Bytes(), &cmd) != nil {
			continue
		}
		kind, _ := cmd["type"].(string)
		fmt.Fprintln(log, kind)
		response := map[string]any{"type": "response", "id": cmd["id"], "command": kind, "success": true}
		switch kind {
		case "get_state":
			response["data"] = map[string]any{"sessionId": "escape-fixture", "isStreaming": running, "model": map[string]any{"id": "test", "provider": "fixture", "name": "Fixture"}}
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
		case "auth.list", "quota.list":
			response["success"] = false
			response["error"] = "Unknown command: " + kind
		case "prompt":
			running = true
			response["data"] = map[string]any{"disposition": "started"}
			write(response)
			write(map[string]any{"type": "agent_start"})
			continue
		case "emit_queue":
			queued = true
			write(map[string]any{"type": "queue_update", "steering": []string{"steer fixture"}, "followUp": []string{"follow fixture"}})
		case "clear_queue":
			if queued {
				response["data"] = map[string]any{"steering": []string{"steer fixture"}, "followUp": []string{"follow fixture"}}
			} else {
				response["data"] = map[string]any{"steering": []string{}, "followUp": []string{}}
			}
			queued = false
			if os.Getenv("ORBIT_PAGE_ESCAPE_DEFER_CLEAR") == "1" {
				heldClear = response
				continue
			}
		case "release_clear":
			// Deliberately release only after abort has arrived. A client that
			// awaits clear_queue before sending abort cannot reach this state.
			if !running && heldClear != nil {
				write(heldClear)
				heldClear = nil
			}
		case "abort":
			running = false
			write(response)
			write(map[string]any{"type": "agent_settled", "aborted": true})
			continue
		}
		write(response)
	}
	os.Exit(0)
}

func newPageEscapeRPCFixture(t *testing.T, deferredClear ...bool) (*Controller, string, string, func(func(agent.Snapshot) bool)) {
	t.Helper()
	x := maintenanceFixture(t)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(t.TempDir(), "rpc-types.txt")
	env := map[string]string{"ORBIT_PAGE_ESCAPE_CHILD": "1", "ORBIT_PAGE_ESCAPE_LOG": log}
	commandTimeout := 2 * time.Second
	if len(deferredClear) > 0 && deferredClear[0] {
		env["ORBIT_PAGE_ESCAPE_DEFER_CLEAR"] = "1"
		commandTimeout = 15 * time.Second // Longer than wait: waiting for clear must fail the test.
	}
	m := agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestPageEscapeRPCChild$", "--"}, Home: t.TempDir(), Environment: []string{}, EnvOverrides: env, DisableExtensions: true, DisablePatches: true, CommandTimeout: commandTimeout})
	t.Cleanup(func() { m.Close() })
	x.Agent = m
	workspace := t.TempDir()
	if err := m.Open(workspace, ""); err != nil {
		t.Fatal(err)
	}
	wait := func(pred func(agent.Snapshot) bool) {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			m.Tick()
			if pred(m.Snapshot()) {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatal("Pi fixture state timeout", m.Snapshot())
	}
	wait(func(s agent.Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
	return x, workspace, log, wait
}

func TestPageEscapeProductionRoutingKeepsRunUntilChatAbort(t *testing.T) {
	for _, page := range []string{"settings", "worktrees"} {
		for _, queued := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/queued=%v", page, queued), func(t *testing.T) {
				x, workspace, log, wait := newPageEscapeRPCFixture(t)
				m := x.Agent
				if err := m.Submit("hold", nil, agent.SendPrompt); err != nil {
					t.Fatal(err)
				}
				wait(func(s agent.Snapshot) bool { return s.Running })
				if queued {
					if err := m.Send("emit_queue", nil); err != nil {
						t.Fatal(err)
					}
					wait(func(s agent.Snapshot) bool { return len(s.Queue.Steering) == 1 && len(s.Queue.FollowUp) == 1 })
				}
				x.UI.SetSnapshot(adaptSnapshot(m.Snapshot(), workspace, nil, nil))
				x.UI.page = page
				x.UI.Draft = "preserve draft"
				var actions []Action
				x.UI.Options.OnAction = func(a Action) { actions = append(actions, a); x.Action(a) }
				tt := ui.NewTester(x.UI.View, 1470, 922)
				tt.Frame()
				tt.Key(0, ui.KeyEscape)
				if x.UI.page != "chat" || len(actions) != 0 || !m.Snapshot().Running || x.UI.Draft != "preserve draft" {
					t.Fatal("page dismissal reached agent", x.UI.page, actions, m.Snapshot().Running)
				}
				beforeRevision := m.Snapshot().ComposerRevision
				beforeEpoch := x.UI.extensionUI.composerEpoch
				// The immediately following Escape now has chat's genuine abort meaning.
				tt.Key(0, ui.KeyEscape)
				wait(func(s agent.Snapshot) bool { return !s.Running })
				if len(actions) != 1 || actions[0].Kind != "AbortRun" {
					t.Fatal(actions)
				}
				b, err := os.ReadFile(log)
				if err != nil {
					t.Fatal(err)
				}
				commands := strings.Fields(string(b))
				var stops []string
				for _, v := range commands {
					if v == "clear_queue" || v == "abort" {
						stops = append(stops, v)
					}
				}
				expected := "abort"
				wantDraft, wantRevision, wantEpoch := "preserve draft", beforeRevision, beforeEpoch
				if queued {
					expected = "clear_queue,abort"
					wantDraft, wantRevision, wantEpoch = "steer fixture\n\nfollow fixture", beforeRevision+1, beforeEpoch+1
					if m.Snapshot().ComposerText != "steer fixture\n\nfollow fixture" {
						t.Fatal("queued text was not made available for composer restoration", m.Snapshot().ComposerText)
					}
				}
				if strings.Join(stops, ",") != expected {
					t.Fatal("expected exactly one source abort sequence", stops)
				}
				snapshot := m.Snapshot()
				if snapshot.ComposerRevision != wantRevision {
					t.Fatalf("abort composer revision = %d, want %d", snapshot.ComposerRevision, wantRevision)
				}
				// Exercise the production projection and host application, rather than
				// stopping at the manager's available restoration text.
				x.UI.SetSnapshot(adaptSnapshot(snapshot, workspace, nil, nil))
				tt.Frame()
				if x.UI.Draft != wantDraft || x.UI.extensionUI.composerEpoch != wantEpoch {
					t.Fatalf("host restoration = %q at epoch %d, want %q at %d", x.UI.Draft, x.UI.extensionUI.composerEpoch, wantDraft, wantEpoch)
				}
				if err := tt.Click("Message"); err != nil {
					t.Fatal(err)
				}
				tt.Key(ui.Cmd, ui.KeyA)
				tt.Key(ui.Cmd, ui.KeyC)
				if tt.Clipboard() != wantDraft {
					t.Fatalf("native editor text = %q, want restored draft %q", tt.Clipboard(), wantDraft)
				}
				tt.Type("later local edit")
				if x.UI.Draft != "later local edit" {
					t.Fatal("native composer edit did not reach host", x.UI.Draft)
				}
				for i := 0; i < 3; i++ {
					x.UI.SetSnapshot(adaptSnapshot(m.Snapshot(), workspace, nil, nil))
					tt.Frame()
					if x.UI.Draft != "later local edit" || x.UI.extensionUI.composerEpoch != wantEpoch {
						t.Fatal("repeated snapshot overwrote newer local edit", x.UI.Draft, x.UI.extensionUI.composerEpoch)
					}
				}
			})
		}
	}
}

func TestAbortPipelineStopsBeforeClearResponseAndRestoresLater(t *testing.T) {
	x, workspace, log, wait := newPageEscapeRPCFixture(t, true)
	m := x.Agent
	if err := m.Submit("hold", nil, agent.SendPrompt); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.Running })
	if err := m.Send("emit_queue", nil); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return len(s.Queue.Steering)+len(s.Queue.FollowUp) == 2 })
	before := m.Snapshot()
	x.UI.SetSnapshot(adaptSnapshot(before, workspace, nil, nil))
	x.UI.Draft = "local draft before clear reply"
	var actions []Action
	x.UI.Options.OnAction = func(a Action) { actions = append(actions, a); x.Action(a) }
	tt := ui.NewTester(x.UI.View, 1470, 922)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	// The child cannot answer clear_queue yet. This proves abort is delivered
	// without waiting for that response or the 15-second command timeout.
	wait(func(s agent.Snapshot) bool { return !s.Running })
	stopped := m.Snapshot()
	if stopped.ComposerRevision != before.ComposerRevision || stopped.ComposerText != before.ComposerText {
		t.Fatal("restoration happened before dropped text arrived", stopped.ComposerRevision, stopped.ComposerText)
	}
	x.UI.SetSnapshot(adaptSnapshot(stopped, workspace, nil, nil))
	tt.Frame()
	if x.UI.Draft != "local draft before clear reply" {
		t.Fatal("pre-response draft changed", x.UI.Draft)
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var stops []string
	for _, cmd := range strings.Fields(string(b)) {
		if cmd == "clear_queue" || cmd == "abort" {
			stops = append(stops, cmd)
		}
	}
	if strings.Join(stops, ",") != "clear_queue,abort" || len(actions) != 1 || actions[0].Kind != "AbortRun" {
		t.Fatal("source ordered pipeline not sent exactly once", stops, actions)
	}
	if err := m.Send("release_clear", nil); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.ComposerRevision == before.ComposerRevision+1 })
	snapshot := m.Snapshot()
	x.UI.SetSnapshot(adaptSnapshot(snapshot, workspace, nil, nil))
	tt.Frame()
	want := "steer fixture\n\nfollow fixture"
	if snapshot.Running || snapshot.ComposerText != want || x.UI.Draft != want {
		t.Fatal("late clear response did not restore exactly the dropped text", snapshot.Running, snapshot.ComposerText, x.UI.Draft)
	}
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != want {
		t.Fatal("late restoration did not reach native editor", tt.Clipboard())
	}
	tt.Type("later local edit")
	for i := 0; i < 3; i++ {
		x.UI.SetSnapshot(adaptSnapshot(m.Snapshot(), workspace, nil, nil))
		tt.Frame()
		if x.UI.Draft != "later local edit" || m.Snapshot().ComposerRevision != before.ComposerRevision+1 {
			t.Fatal("late restoration replayed", x.UI.Draft)
		}
	}
}

func TestQueueClearProductionRoutingDoesNotRestoreOrAbort(t *testing.T) {
	x, workspace, log, wait := newPageEscapeRPCFixture(t)
	m := x.Agent
	if err := m.Submit("hold", nil, agent.SendPrompt); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return s.Running })
	if err := m.Send("emit_queue", nil); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return len(s.Queue.Steering) == 1 && len(s.Queue.FollowUp) == 1 })
	before := m.Snapshot()
	x.UI.SetSnapshot(adaptSnapshot(before, workspace, nil, nil))
	x.UI.Draft = "preserve local draft"
	beforeEpoch := x.UI.extensionUI.composerEpoch
	var actions []Action
	x.UI.Options.OnAction = func(a Action) { actions = append(actions, a); x.Action(a) }
	tt := ui.NewTester(x.UI.View, 1470, 922)
	tt.Frame()
	if err := tt.Click(x.UI.Tr("view.clear_queue")); err != nil {
		t.Fatal(err)
	}
	wait(func(s agent.Snapshot) bool { return len(s.Queue.Steering)+len(s.Queue.FollowUp) == 0 })
	after := m.Snapshot()
	if !after.Running || after.ComposerRevision != before.ComposerRevision || after.ComposerText != before.ComposerText {
		t.Fatal("Clear queue restored text or stopped the run", after.Running, after.ComposerText, after.ComposerRevision)
	}
	if len(actions) != 1 || actions[0].Kind != "ClearQueue" {
		t.Fatal("queue header dispatched more than its single action", actions)
	}
	for i := 0; i < 3; i++ {
		x.UI.SetSnapshot(adaptSnapshot(m.Snapshot(), workspace, nil, nil))
		tt.Frame()
		if x.UI.Draft != "preserve local draft" || x.UI.extensionUI.composerEpoch != beforeEpoch {
			t.Fatal("Clear queue disturbed host composer", x.UI.Draft, x.UI.extensionUI.composerEpoch)
		}
		if _, visible := tt.Find("Queue bar"); visible {
			t.Fatal("cleared queue is still visible")
		}
	}
	b, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var stops []string
	for _, cmd := range strings.Fields(string(b)) {
		if cmd == "clear_queue" || cmd == "abort" {
			stops = append(stops, cmd)
		}
	}
	if strings.Join(stops, ",") != "clear_queue" {
		t.Fatal("Clear queue emitted an abort or duplicate clear", stops)
	}
}

func TestSourceNativeAbortCommandHonorsModalThenPagePrecedence(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.Model.Running = true
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	a.confirmOpen = true
	a.nativeCommand("AbortRun")
	if a.page != "settings" || len(actions) != 0 {
		t.Fatal("native command escaped modal")
	}
	a.confirmOpen = false
	a.nativeCommand("AbortRun")
	if a.page != "chat" || len(actions) != 0 {
		t.Fatal("native command did not dismiss Settings")
	}
	a.nativeCommand("AbortRun")
	if len(actions) != 1 || actions[0].Kind != "AbortRun" {
		t.Fatal("chat abort was swallowed", actions)
	}
}
