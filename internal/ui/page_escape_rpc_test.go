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

func TestPageEscapeProductionRoutingKeepsRunUntilChatAbort(t *testing.T) {
	for _, page := range []string{"settings", "worktrees"} {
		for _, queued := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/queued=%v", page, queued), func(t *testing.T) {
				x := maintenanceFixture(t)
				exe, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				log := filepath.Join(t.TempDir(), "rpc-types.txt")
				m := agent.NewManager(agent.Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestPageEscapeRPCChild$", "--"}, Home: t.TempDir(), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_PAGE_ESCAPE_CHILD": "1", "ORBIT_PAGE_ESCAPE_LOG": log}, DisableExtensions: true, DisablePatches: true, CommandTimeout: 2 * time.Second})
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
				if queued {
					expected = "clear_queue,abort"
					if m.Snapshot().ComposerText != "steer fixture\n\nfollow fixture" {
						t.Fatal("queued text was not made available for composer restoration", m.Snapshot().ComposerText)
					}
				}
				if strings.Join(stops, ",") != expected {
					t.Fatal("expected exactly one source abort sequence", stops)
				}
			})
		}
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
