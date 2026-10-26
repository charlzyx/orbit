package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// A real child speaks JSONL; it never authenticates or executes a provider/tool.
func TestUpstreamP1Process(t *testing.T) {
	if os.Getenv("ORBIT_P1_HELPER") != "1" {
		return
	}
	write := func(v any) { b, _ := json.Marshal(v); fmt.Println(string(b)) }
	scan := bufio.NewScanner(os.Stdin)
	var composerID, firstComposerID any
	composerCount := 0
	sessionCount := 1
	sessionID := "p1"
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case "emit_message_shape":
			if raw, ok := c["raw"].(string); ok {
				fmt.Println(raw)
			} else {
				write(c["event"])
			}
		case pirpc.GetState:
			r["data"] = map[string]any{"sessionId": sessionID, "sessionName": "fixture"}
		case pirpc.NewSession:
			sessionCount++
			sessionID = "p1"
			if sessionCount%2 == 0 {
				sessionID = "p2"
			}
			r["data"] = map[string]any{"cancelled": false}
		case "empty_prompt_success", "empty_prompt_failure":
			write(map[string]any{"type": "response", "command": pirpc.Prompt, "success": kind == "empty_prompt_success", "error": "untracked failure", "data": map[string]any{"disposition": "handled"}})
		case "finish_first_composer":
			write(map[string]any{"type": "response", "id": firstComposerID, "command": pirpc.Prompt, "success": true, "data": map[string]any{"disposition": "handled"}})
		case pirpc.Prompt:
			if c["message"] == "composer" || c["message"] == "composer-pending" {
				composerID = c["id"]
				composerCount++
				if firstComposerID == nil {
					firstComposerID = composerID
				}
				if c["message"] == "composer" {
					write(map[string]any{"type": "auto_retry_start", "attempt": 1})
				}
				status := "composer received"
				if composerCount > 1 {
					status = fmt.Sprintf("composer received %d", composerCount)
				}
				write(map[string]any{"type": "extension_ui_request", "method": "setStatus", "statusText": status})
				continue
			}
			r["data"] = map[string]any{"disposition": "handled"}
			if c["message"] == "foreign-fail" {
				r["success"] = false
				r["error"] = "foreign rejected"
			}
		case "fail_first_composer", "fail_current_composer":
			id := composerID
			if kind == "fail_first_composer" {
				id = firstComposerID
			}
			write(map[string]any{"type": "response", "id": id, "command": pirpc.Prompt, "success": false, "error": "old rejected"})
		case "finish_composer":
			write(map[string]any{"type": "response", "id": composerID, "command": pirpc.Prompt, "success": true, "data": map[string]any{"disposition": "handled"}})
		case pirpc.Abort:
			write(map[string]any{"type": "auto_retry_end", "success": false})
			write(map[string]any{"type": "agent_settled", "aborted": true})
		}
		write(r)
	}
	os.Exit(0)
}
func p1Manager(t *testing.T) *Manager {
	return testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestUpstreamP1Process$", "--"}
		o.EnvOverrides = map[string]string{"ORBIT_P1_HELPER": "1"}
		o.CommandTimeout = 10 * time.Second
	})
}
func p1Composer(t *testing.T, m *Manager, retry bool) *pirpc.Client {
	t.Helper()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	message := "composer-pending"
	if retry {
		message = "composer"
	}
	if err := m.Submit(message, nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return s.ExtensionStatusText == "composer received" && s.Awaiting && s.Retrying == retry
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[m.active].client
}
func p1Call(t *testing.T, m *Manager, c *pirpc.Client, kind string, params map[string]any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	if _, err := c.Call(ctx, pirpc.NewCommand(kind, params)); err != nil {
		t.Fatal(err)
	}
	m.Tick()
}
func TestUpstreamP1CancelledRetryIsNotFailure(t *testing.T) {
	for _, send := range []bool{false, true} {
		name := "interactive"
		if send {
			name = "raw_abort"
		}
		t.Run(name, func(t *testing.T) {
			m := p1Manager(t)
			c := p1Composer(t, m, true)
			var err error
			if send {
				err = m.Send(pirpc.Abort, nil)
			} else {
				err = m.Abort()
			}
			if err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return !s.Retrying && !s.Awaiting })
			p1Call(t, m, c, "barrier", nil)
			if got := m.Snapshot(); got.ConnectionError != "" || got.RetryFailure.Active {
				t.Fatalf("cancel reported exhaustion: %q %+v", got.ConnectionError, got.RetryFailure)
			}
		})
	}
}
func TestUpstreamP1ForeignHandledPreservesComposer(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	// A separate real RPC request is handled without starting an agent turn.
	p1Call(t, m, c, pirpc.Prompt, map[string]any{"message": "extension-command"})
	m.mu.Lock()
	awaiting, turn := m.sessions[m.active].state.awaiting, m.sessions[m.active].turnOpen
	m.mu.Unlock()
	if !awaiting || !turn {
		t.Fatalf("foreign handled response closed composer: awaiting=%v turn=%v", awaiting, turn)
	}
	p1Call(t, m, c, "finish_composer", nil)
	m.mu.Lock()
	awaiting, turn = m.sessions[m.active].state.awaiting, m.sessions[m.active].turnOpen
	m.mu.Unlock()
	if awaiting || turn || m.AnyBusy() {
		t.Fatalf("matching handled response did not finish: awaiting=%v turn=%v", awaiting, turn)
	}
}

func TestUpstreamP1CommandErrorsClearOnlySameSource(t *testing.T) {
	s := newState("/fixture", "")
	apply(t, &s, `{"type":"response","id":"a","command":"set_model","success":false,"error":"模型不可用"}`)
	apply(t, &s, `{"type":"response","id":"b","command":"get_messages","success":true,"data":{"messages":[]}}`)
	if s.snapshot.ConnectionError != "模型不可用" {
		t.Fatal("unrelated success erased error", s.snapshot.ConnectionError)
	}
	apply(t, &s, `{"type":"response","id":"c","command":"set_model","success":true}`)
	if s.snapshot.ConnectionError != "" {
		t.Fatal("successful retry retained same-source error", s.snapshot.ConnectionError)
	}
}
func TestUpstreamP1NewErrorReplacesSource(t *testing.T) {
	s := newState("/fixture", "")
	apply(t, &s, `{"type":"response","id":"a","command":"set_model","success":false,"error":"same literal"}`)
	apply(t, &s, `{"type":"response","id":"b","command":"get_messages","success":false,"error":"same literal"}`)
	apply(t, &s, `{"type":"response","id":"c","command":"set_model","success":true}`)
	if s.snapshot.ConnectionError != "same literal" {
		t.Fatal("text-matching cleared foreign source")
	}
	apply(t, &s, `{"type":"response","id":"d","command":"get_messages","success":true,"data":{"messages":[]}}`)
	if s.snapshot.ConnectionError != "" {
		t.Fatal("new owner success retained stale error", s.snapshot.ConnectionError)
	}
}

func TestUpstreamP1AbortIntentLifecycle(t *testing.T) {
	for _, tc := range []struct {
		name            string
		busy, lateStart bool
		terminal        string
		wantFailure     bool
	}{
		{"idle_abort_cannot_mask", false, false, "", true},
		{"awaiting_late_start", true, true, "", false},
		{"settled_retires_intent", true, false, `{"type":"agent_settled"}`, true},
		{"exit_retires_intent", true, false, `{"type":"process_exited"}`, true},
		{"transport_retires_intent", true, false, `{"type":"transport_error","error":"closed"}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := newState("/fixture", "")
			s.awaiting = tc.busy
			s.requestAbort()
			if tc.lateStart {
				apply(t, &s, `{"type":"agent_start"}`)
			}
			if tc.terminal != "" {
				apply(t, &s, tc.terminal)
			}
			apply(t, &s, `{"type":"auto_retry_end","success":false,"finalError":"Retry cancelled"}`)
			if s.snapshot.RetryFailure.Active != tc.wantFailure {
				t.Fatal("wrong intent lifecycle", s.snapshot.RetryFailure)
			}
			if s.abortRequested {
				t.Fatal("end did not consume abort intent")
			}
			apply(t, &s, `{"type":"auto_retry_end","success":false,"finalError":"real later failure"}`)
			if !s.snapshot.RetryFailure.Active || s.snapshot.RetryFailure.Message != "real later failure" {
				t.Fatal("consumed intent masked later failure")
			}
		})
	}
}
func TestUpstreamP1ErrorOwnershipNonCommandAndAgent(t *testing.T) {
	s := newState("/fixture", "")
	s.snapshot.setCommandError("set_model", "same")
	s.snapshot.setConnectionError("same")
	apply(t, &s, `{"type":"response","command":"set_model","success":true}`)
	if s.snapshot.ConnectionError != "same" {
		t.Fatal("generic writer retained obsolete ownership")
	}
	apply(t, &s, `{"type":"message_end","message":{"role":"assistant","stopReason":"error","errorMessage":" provider rejected "}}`)
	if s.snapshot.ConnectionError != "provider rejected" {
		t.Fatal("agent error not surfaced", s.snapshot.ConnectionError)
	}
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{}}`)
	if s.snapshot.ConnectionError != "provider rejected" {
		t.Fatal("poll cleared agent error")
	}
	apply(t, &s, `{"type":"message_end","message":{"role":"assistant","stopReason":"aborted","errorMessage":"Request aborted"}}`)
	if s.snapshot.ConnectionError != "" {
		t.Fatal("abort surfaced as failure or stale agent error survived")
	}
	s.snapshot.setCommandError("set_model", "model failed")
	apply(t, &s, `{"type":"message_end","message":{"role":"assistant","content":[]}}`)
	if s.snapshot.ConnectionError != "model failed" {
		t.Fatal("successful message cleared command error")
	}
}
func TestUpstreamP1AssistantErrorSourceContract(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"role":"assistant","stopReason":"stop","errorMessage":" odd but real "}`, "odd but real"},
		{`{"role":"assistant","stopReason":"aborted","errorMessage":"aborted"}`, ""},
		{`{"role":"toolResult","errorMessage":"tool"}`, ""},
		{`{"message":{"role":"assistant","errorMessage":"nested"}}`, "nested"},
		{`{"role":"assistant","errorMessage":" \n "}`, ""},
		{`{"role":"assistant","errorMessage":4}`, ""},
	} {
		if got := assistantMessageError(json.RawMessage(tc.raw)); got != tc.want {
			t.Errorf("%s got %q want %q", tc.raw, got, tc.want)
		}
	}
}
func TestUpstreamP1DedicatedResponsesCannotClearBanner(t *testing.T) {
	for _, command := range []string{pirpc.AuthList, pirpc.AuthStatus, pirpc.AuthLogin, pirpc.QuotaList, pirpc.GetEntries} {
		t.Run(command, func(t *testing.T) {
			s := newState("/fixture", "")
			s.snapshot.setConnectionError("unrelated")
			raw, _ := json.Marshal(map[string]any{"type": "response", "command": command, "success": true, "data": map[string]any{}})
			apply(t, &s, string(raw))
			if s.snapshot.ConnectionError != "unrelated" {
				t.Fatal("dedicated response cleared generic banner")
			}
		})
	}
}

func TestUpstreamP1LatePromptAndResetIsolation(t *testing.T) {
	s := newState("/fixture", "")
	s.snapshot.ActiveSession.ID = "old"
	s.promptInFlight = "old-prompt"
	s.awaiting = true
	s.requestAbort()
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"new"}}`)
	if s.abortRequested || s.promptInFlight != "" {
		t.Fatal("session reset retained request ownership")
	}
	s.promptInFlight = "new-prompt"
	s.awaiting = true
	apply(t, &s, `{"type":"response","id":"old-prompt","command":"prompt","success":true,"data":{"disposition":"handled"}}`)
	if !s.awaiting || s.promptInFlight != "new-prompt" {
		t.Fatal("old handled response consumed new composer")
	}
	apply(t, &s, `{"type":"response","id":"new-prompt","command":"prompt","success":true,"data":{"disposition":"handled"}}`)
	if s.awaiting || s.promptInFlight != "" {
		t.Fatal("matching response did not consume request")
	}
	apply(t, &s, `{"type":"auto_retry_end","success":false}`)
	if !s.snapshot.RetryFailure.Active {
		t.Fatal("old session cancel masked new failure")
	}
	s.promptInFlight = "settled"
	apply(t, &s, `{"type":"agent_settled"}`)
	if s.promptInFlight != "" {
		t.Fatal("settled request retained")
	}
}
