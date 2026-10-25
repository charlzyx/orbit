package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

const questionnaireArgs = `{"questions":[{"header":"Method","question":"Which method?","options":[{"label":"A","description":"Alpha"},{"label":"B","preview":"**Beta**"}]},{"question":"Which checks?","multiSelect":true,"options":[{"label":"X"},{"label":"Y"}]},{"question":"What else?","options":[{"label":"Nothing"}]}]}`

func uiEvent(t *testing.T, s *sessionState, raw string) {
	t.Helper()
	event, err := pirpc.ParseEvent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	event.ReceivedAt = time.Unix(123, 0)
	s.apply(event)
}
func responses(t *testing.T, s *sessionState, want ...pirpc.UIResponse) {
	t.Helper()
	got := s.takeUIResponses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UI replies got %#v want %#v", got, want)
	}
}
func strptr(s string) *string { return &s }
func TestExtensionSourceRoutingCollisionAndPlaceholder(t *testing.T) {
	s := newState("", "")
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"guard","method":"select","title":"[orbit-guard] bash\techo fixture","options":["Allow once","Always allow this tool","Deny"]}`)
	if s.snapshot.Approval == nil || s.snapshot.Approval.Tool != "bash" || s.snapshot.Approval.Detail != "echo fixture" || len(s.snapshot.Dialogs) != 0 {
		t.Fatal(s.snapshot)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"new","method":"confirm","title":"Second"}`)
	responses(t, &s, pirpc.UIResponse{ID: "new", Cancelled: true})
	if s.snapshot.Approval.ID != "guard" {
		t.Fatal("newcomer replaced guard")
	}
	s.removeDialog("guard")
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"input","method":"input","title":"Name","placeholder":"Your name","timeout":10}`)
	d := s.snapshot.Dialogs[0]
	if d.Placeholder != "Your name" || !d.ReceivedAt.Equal(time.Unix(123, 0)) {
		t.Fatal(d)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"guard2","method":"select","title":"[orbit-guard] malformed","options":["Deny"]}`)
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"next","method":"editor","prefill":"retained"}`)
	responses(t, &s, pirpc.UIResponse{ID: "guard2", Cancelled: true}, pirpc.UIResponse{ID: "next", Cancelled: true})
	if len(s.snapshot.Dialogs) != 1 || s.snapshot.Dialogs[0].ID != "input" {
		t.Fatal(s.snapshot.Dialogs)
	}
	s.removeDialog("input")
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"ordinary","method":"select","title":"Allow once?","options":["Deny"]}`)
	if s.snapshot.Approval != nil || s.snapshot.Dialogs[0].ID != "ordinary" {
		t.Fatal("unmarked select was classified as guard")
	}
}
func TestExtensionQuestionnaireBufferedReplayAndEndIdentity(t *testing.T) {
	s := newState("", "")
	uiEvent(t, &s, `{"type":"tool_execution_start","toolCallId":"ask-tool","toolName":"ask_user_question","args":`+questionnaireArgs+`}`)
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"q1","method":"select","options":["1. A","2. B","3. Type something."]}`)
	if s.snapshot.Ask == nil || len(s.snapshot.Ask.Questions) != 3 || s.snapshot.Dialogs != nil {
		t.Fatal(s.snapshot)
	}
	responses(t, &s)
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"stray","method":"input"}`)
	responses(t, &s, pirpc.UIResponse{ID: "stray", Cancelled: true})
	answers := []AskAnswer{{Kind: "option", Option: 1}, {Kind: "multi", Checked: []bool{false, true}}, {Kind: "custom", Custom: "ship it"}}
	if err := s.submitAsk("q1", answers); err != nil {
		t.Fatal(err)
	}
	responses(t, &s, pirpc.UIResponse{ID: "q1", Value: strptr("2. B")})
	answers[1].Checked[0] = true
	if !s.snapshot.Ask.Submitted {
		t.Fatal("submitted panel not read-only")
	}
	if err := s.submitAsk("q1", answers); err == nil {
		t.Fatal("double submission accepted")
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"q2","method":"input"}`)
	responses(t, &s, pirpc.UIResponse{ID: "q2", Value: strptr("2")})
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"q3","method":"select"}`)
	responses(t, &s, pirpc.UIResponse{ID: "q3", Value: strptr("2. Type something.")})
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"custom","method":"input"}`)
	responses(t, &s, pirpc.UIResponse{ID: "custom", Value: strptr("ship it")})
	if s.askReplay != nil {
		t.Fatal("completed replay still active")
	}
	uiEvent(t, &s, `{"type":"tool_execution_end","toolCallId":"unrelated","toolName":"bash"}`)
	if s.snapshot.Ask == nil {
		t.Fatal("unrelated end cleared questionnaire")
	}
	uiEvent(t, &s, `{"type":"tool_execution_end","toolCallId":"ask-tool","toolName":"ask_user_question","result":{"content":[{"type":"text","text":"User has answered your questions: \"Which method?\"=\"B\". \"Which checks?\"=\"Y\". \"What else?\"=\"ship it\"."}]}}`)
	if s.snapshot.Ask != nil || s.askToolID != "" {
		t.Fatal("tool finish left live panel")
	}
	if s.snapshot.Messages[0].ToolOutput == "" || len(s.snapshot.Messages[0].Blocks) != 1 {
		t.Fatal("answer transcript/arguments lost", s.snapshot.Messages)
	}
}
func TestExtensionQuestionnaireMalformedFallbackMismatchDecline(t *testing.T) {
	s := newState("", "")
	uiEvent(t, &s, `{"type":"tool_execution_start","toolCallId":"bad","toolName":"ask_user_question","args":{"questions":[{"header":"missing"}]}}`)
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"fallback","method":"select","options":["A"]}`)
	if s.snapshot.Ask != nil || len(s.snapshot.Dialogs) != 1 {
		t.Fatal("malformed questionnaire did not fall back")
	}
	s.removeDialog("fallback")
	uiEvent(t, &s, `{"type":"tool_execution_start","toolCallId":"ask","toolName":"ask_question","args":{"questions":[{"question":"Q","multiSelect":true,"options":[{"label":"A"}]}]}}`)
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"first","method":"input"}`)
	if err := s.submitAsk("first", []AskAnswer{{Kind: "multi", Checked: []bool{false}, Custom: "  typed  "}}); err != nil {
		t.Fatal(err)
	}
	responses(t, &s, pirpc.UIResponse{ID: "first", Value: strptr("typed")})
	// A source-incompatible request during an unfinished replay fails closed.
	s.askReplay = &askReplay{questions: []AskQuestion{{}}, answers: []AskAnswer{{Kind: "multi"}}}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"mismatch","method":"select"}`)
	responses(t, &s, pirpc.UIResponse{ID: "mismatch", Cancelled: true})
	if s.askReplay != nil {
		t.Fatal("mismatched replay retained")
	}
}
func TestExtensionWidgetsNotifyStatusAndTeardown(t *testing.T) {
	s := newState("", "")
	s.snapshot.Status = "Running"
	for _, record := range []string{
		`{"type":"extension_ui_request","id":"a","method":"setWidget","widgetKey":"tasks","widgetLines":["first",1],"widgetPlacement":"aboveEditor"}`,
		`{"type":"extension_ui_request","id":"b","method":"setWidget","widgetKey":"after","widgetLines":["below"],"widgetPlacement":"belowEditor"}`,
		`{"type":"extension_ui_request","id":"c","method":"setWidget","widgetKey":"tasks","widgetLines":["replaced"],"widgetPlacement":"belowEditor"}`,
		`{"type":"extension_ui_request","id":"n","method":"notify","message":"Saved"}`,
		`{"type":"extension_ui_request","id":"s","method":"setStatus","statusKey":"job","statusText":"Indexing"}`,
	} {
		uiEvent(t, &s, record)
	}
	if got := s.snapshot.ExtensionWidgets; len(got) != 2 || got[0].Key != "tasks" || got[0].Lines[0] != "replaced" || got[0].Placement != "belowEditor" {
		t.Fatal(got)
	}
	if s.snapshot.Status != "Running" || s.snapshot.StatusEntries["job"] != "Indexing" || len(s.snapshot.ExtensionNotices) != 1 || s.snapshot.ExtensionNotices[0].Message != "Saved" {
		t.Fatal(s.snapshot)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","method":"notify","message":""}`)
	uiEvent(t, &s, `{"type":"extension_ui_request","method":"notify","message":null}`)
	if len(s.snapshot.ExtensionNotices) != 2 || s.snapshot.ExtensionNotices[1].Message != "" {
		t.Fatal("notify string presence was not preserved", s.snapshot.ExtensionNotices)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","method":"setWidget","widgetKey":null,"widgetLines":["invalid"]}`)
	if len(s.snapshot.ExtensionWidgets) != 2 {
		t.Fatal("null widget key accepted")
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","method":"setWidget","widgetKey":"tasks"}`)
	if len(s.snapshot.ExtensionWidgets) != 1 || s.snapshot.ExtensionWidgets[0].Key != "after" {
		t.Fatal(s.snapshot.ExtensionWidgets)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","method":"setStatus","statusKey":"job"}`)
	if len(s.snapshot.StatusEntries) != 0 {
		t.Fatal(s.snapshot.StatusEntries)
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"q","method":"confirm"}`)
	s.cancelExtensionUI()
	responses(t, &s, pirpc.UIResponse{ID: "q", Cancelled: true})
	if len(s.snapshot.ExtensionWidgets) != 1 || len(s.snapshot.Dialogs) != 0 || s.snapshot.Approval != nil || s.snapshot.Ask != nil {
		t.Fatal("park must retain widgets, not blocking UI")
	}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"g","method":"select","title":"[orbit-guard] edit\tfile"}`)
	uiEvent(t, &s, `{"type":"process_exited","error":"closed"}`)
	if s.snapshot.Approval != nil || len(s.snapshot.ExtensionWidgets) != 0 {
		t.Fatal("exit left approval or widgets live")
	}
}

// A deterministic JSONL child verifies actual client/manager effects. It does
// not substitute for the separate pinned, real-Pi integration below.
func TestExtensionUITranscriptProcess(t *testing.T) {
	if os.Getenv("ORBIT_EXTENSION_UI_HELPER") != "1" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	write := func(v any) { b, _ := json.Marshal(v); fmt.Fprintln(os.Stdout, string(b)) }
	for scan.Scan() {
		var c map[string]json.RawMessage
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		var kind string
		_ = json.Unmarshal(c["type"], &kind)
		var id string
		_ = json.Unmarshal(c["id"], &id)
		if kind == "extension_ui_response" {
			c["type"] = json.RawMessage(`"ui_response_observed"`)
			write(c)
			continue
		}
		data := map[string]any{}
		switch kind {
		case "get_state":
			data = map[string]any{"sessionId": "fixture", "isStreaming": false, "thinkingLevel": "off"}
		case "get_messages":
			data["messages"] = []any{}
		case "get_entries":
			data["entries"] = []any{}
		case "emit_extensions":
			var events []json.RawMessage
			_ = json.Unmarshal(c["events"], &events)
			for _, e := range events {
				fmt.Fprintln(os.Stdout, string(e))
			}
		}
		write(map[string]any{"type": "response", "id": id, "command": kind, "success": true, "data": data})
	}
	os.Exit(0)
}
func extensionManager(t *testing.T) *Manager {
	t.Helper()
	exe, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	m := NewManager(Options{PiBinary: exe, PiPrefixArgs: []string{"-test.run=^TestExtensionUITranscriptProcess$", "--"}, Home: t.TempDir(), Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_EXTENSION_UI_HELPER": "1"}, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, CommandTimeout: 2 * time.Second})
	t.Cleanup(func() { m.Close() })
	if e = m.Open(t.TempDir(), ""); e != nil {
		t.Fatal(e)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	return m
}
func emitUI(t *testing.T, m *Manager, records ...string) {
	t.Helper()
	var events []json.RawMessage
	for _, r := range records {
		events = append(events, json.RawMessage(r))
	}
	if e := m.Send("emit_extensions", map[string]any{"events": events}); e != nil {
		t.Fatal(e)
	}
}
func observedUI(s Snapshot, id string) *pirpc.UIResponse {
	for _, raw := range s.UnknownEvents {
		var r pirpc.UIResponse
		if json.Unmarshal(raw, &r) == nil && r.Type == "ui_response_observed" && r.ID == id {
			return &r
		}
	}
	return nil
}

func TestManagerExtensionTranscriptResponsesAndParking(t *testing.T) {
	m := extensionManager(t)
	emitUI(t, m, `{"type":"extension_ui_request","id":"first","method":"input","placeholder":"Type here"}`, `{"type":"extension_ui_request","id":"second","method":"confirm"}`)
	s := waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "second") != nil })
	if !observedUI(s, "second").Cancelled || len(s.Dialogs) != 1 || s.Dialogs[0].ID != "first" {
		t.Fatal(s)
	}
	if e := m.RespondDialog("first", "", nil, false); e != nil {
		t.Fatal(e)
	}
	if e := m.RespondDialog("first", "duplicate", nil, false); e == nil {
		t.Fatal("duplicate click scheduled another answer")
	}
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "first") != nil })
	r := observedUI(s, "first")
	if r.Value == nil || *r.Value != "" || r.Cancelled {
		t.Fatal(r)
	}
	emitUI(t, m, `{"type":"extension_ui_request","id":"guard","method":"select","title":"[orbit-guard] bash\techo fixture","options":["Allow once","Always allow this tool","Deny"]}`)
	waitFor(t, m, func(s Snapshot) bool { return s.Approval != nil })
	if e := m.RespondDialog("guard", "Always allow this tool", nil, false); e != nil {
		t.Fatal(e)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "guard") != nil })
	if r = observedUI(s, "guard"); r.Value == nil || *r.Value != "Always allow this tool" {
		t.Fatal(r)
	}
	emitUI(t, m, `{"type":"tool_execution_start","toolCallId":"tool","toolName":"ask_user_question","args":{"questions":[{"question":"Q","options":[{"label":"A"}]}]}}`, `{"type":"extension_ui_request","id":"ask","method":"select"}`)
	waitFor(t, m, func(s Snapshot) bool { return s.Ask != nil })
	if e := m.SubmitAsk("ask", []AskAnswer{{Kind: "custom", Custom: "custom"}}); e != nil {
		t.Fatal(e)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "ask") != nil })
	if *observedUI(s, "ask").Value != "2. Type something." {
		t.Fatal(s)
	}
	emitUI(t, m, `{"type":"extension_ui_request","id":"text","method":"input"}`)
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "text") != nil })
	if *observedUI(s, "text").Value != "custom" {
		t.Fatal(s)
	}
	emitUI(t, m, `{"type":"tool_execution_end","toolCallId":"tool"}`, `{"type":"extension_ui_request","id":"park","method":"confirm"}`)
	waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) > 0 })
	m.mu.Lock()
	old := m.sessions[m.active]
	m.mu.Unlock()
	if e := m.Open(t.TempDir(), ""); e != nil {
		t.Fatal(e)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		m.Tick()
		m.mu.Lock()
		r := observedUI(old.state.snapshot, "park")
		count := len(old.state.snapshot.Dialogs)
		m.mu.Unlock()
		if r != nil && r.Cancelled && count == 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("already-open parked request was not cancelled over RPC")
}

func TestExtensionEditorRevisionAndSnapshotIsolation(t *testing.T) {
	s := newState("", "")
	for i, text := range []string{"same", "same", ""} {
		raw, _ := json.Marshal(map[string]any{"type": "extension_ui_request", "id": fmt.Sprint(i), "method": "set_editor_text", "text": text})
		uiEvent(t, &s, string(raw))
		if s.snapshot.ComposerRevision != uint64(i+1) || s.snapshot.ComposerText != text {
			t.Fatal(s.snapshot)
		}
	}
	qs := ParseAskQuestions(json.RawMessage(`{"questions":[null,{"question":null},{"question":"Q","header":{},"multiSelect":"bad","options":[null,{"label":null},{"label":"A","description":{},"preview":[]}]},{"question":"Q2","options":"bad"}]}`))
	if len(qs) != 2 || qs[0].Question != "Q" || len(qs[0].Options) != 1 || qs[0].Options[0].Label != "A" || qs[0].Options[0].Description != "" || qs[1].Question != "Q2" {
		t.Fatal(qs)
	}
}

func TestManagerExtensionNewSessionCancelsExistingIdleDialog(t *testing.T) {
	m := extensionManager(t)
	emitUI(t, m, `{"type":"extension_ui_request","id":"departing","method":"confirm"}`)
	s := waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 })
	if err := m.NewSession(s.ActiveSession.Workspace); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "departing") != nil })
	if !observedUI(s, "departing").Cancelled || len(s.Dialogs) != 0 {
		t.Fatal(s)
	}
}
func TestManagerExtensionQuestionnaireDeclineAndSnapshotIsolation(t *testing.T) {
	m := extensionManager(t)
	emitUI(t, m, `{"type":"tool_execution_start","toolCallId":"tool","toolName":"ask_user_question","args":{"questions":[{"question":"Q","options":[{"label":"A"}]}]}}`, `{"type":"extension_ui_request","id":"ask","method":"select"}`, `{"type":"extension_ui_request","method":"setWidget","widgetKey":"w","widgetLines":["line"]}`)
	s := waitFor(t, m, func(s Snapshot) bool { return s.Ask != nil && len(s.ExtensionWidgets) == 1 })
	s.Ask.Questions[0].Options[0].Label = "mutated"
	s.ExtensionWidgets[0].Lines[0] = "mutated"
	current := m.Snapshot()
	if current.Ask.Questions[0].Options[0].Label != "A" || current.ExtensionWidgets[0].Lines[0] != "line" {
		t.Fatal("snapshot leaked extension slice ownership")
	}
	if err := m.RespondDialog("ask", "", nil, true); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return observedUI(s, "ask") != nil })
	if !observedUI(s, "ask").Cancelled || s.Ask != nil {
		t.Fatal(s)
	}
	emitUI(t, m, `{"type":"tool_execution_end","toolCallId":"tool","toolName":"ask_user_question","result":{"content":[{"type":"text","text":"User declined to answer questions"}]}}`)
	s = waitFor(t, m, func(s Snapshot) bool {
		return len(s.Messages) > 0 && s.Messages[0].ToolOutput == "User declined to answer questions"
	})
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if s = m.Snapshot(); len(s.ExtensionWidgets) != 0 || s.Ask != nil || s.Approval != nil || len(s.Dialogs) != 0 {
		t.Fatal("stopped runtime kept live UI")
	}
}

func TestManagerExtensionDoesNotInventTimeoutCountdown(t *testing.T) {
	m := extensionManager(t)
	emitUI(t, m, `{"type":"extension_ui_request","id":"timed","method":"input","timeout":1}`)
	waitFor(t, m, func(s Snapshot) bool { return len(s.Dialogs) == 1 })
	time.Sleep(10 * time.Millisecond)
	m.Tick()
	if s := m.Snapshot(); len(s.Dialogs) != 1 || observedUI(s, "timed") != nil {
		t.Fatal("host invented local dialog expiration", s)
	}
	if err := m.RespondDialog("timed", "", nil, true); err != nil {
		t.Fatal(err)
	}
}
func TestExtensionEditorMalformedEventsAndLocalizedReplay(t *testing.T) {
	s := newState("", "")
	s.setComposerText("keep")
	for _, raw := range []string{`{"type":"extension_ui_request","method":"set_editor_text"}`, `{"type":"extension_ui_request","method":"set_editor_text","text":null}`, `{"type":"extension_ui_request","method":"set_editor_text","text":17}`} {
		uiEvent(t, &s, raw)
	}
	if s.snapshot.ComposerText != "keep" || s.snapshot.ComposerRevision != 1 {
		t.Fatal("malformed payload cleared draft")
	}
	s.askReplay = &askReplay{questions: []AskQuestion{{Options: []AskOption{{Label: "A"}}}}, answers: []AskAnswer{{Kind: "custom", Custom: "内容", CustomLabel: "输入内容。"}}}
	uiEvent(t, &s, `{"type":"extension_ui_request","id":"localized","method":"select"}`)
	responses(t, &s, pirpc.UIResponse{ID: "localized", Value: strptr("2. 输入内容。")})
}

func TestManagerExtensionWarmResumeRetainsWidgetsButCancelsParkedEvents(t *testing.T) {
	m := extensionManager(t)
	workspace := m.Snapshot().ActiveSession.Workspace
	path := filepath.Join(workspace, "fixture.jsonl")
	state, _ := json.Marshal(map[string]any{"type": "response", "command": "get_state", "success": true, "data": map[string]any{"sessionId": "fixture", "sessionFile": path}})
	emitUI(t, m, string(state), `{"type":"extension_ui_request","id":"widget","method":"setWidget","widgetKey":"retained","widgetLines":["Original"]}`, `{"type":"extension_ui_request","id":"pending","method":"input"}`)
	s := waitFor(t, m, func(s Snapshot) bool {
		return len(s.ExtensionWidgets) == 1 && len(s.Dialogs) == 1 && s.ActiveSession.Path == path
	})
	runtimeID := s.RuntimeID
	m.mu.Lock()
	old := m.sessions[m.active]
	m.mu.Unlock()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if len(m.Snapshot().ExtensionWidgets) != 0 {
		t.Fatal("widgets leaked into another session")
	}
	_, err := m.callClient(old.client, pirpc.NewCommand("emit_extensions", map[string]any{"events": []json.RawMessage{
		json.RawMessage(`{"type":"extension_ui_request","id":"parked-widget","method":"setWidget","widgetKey":"retained","widgetLines":["Not applied"]}`),
		json.RawMessage(`{"type":"extension_ui_request","id":"parked-notice","method":"notify","message":"Not surfaced"}`),
	}}))
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		m.Tick()
		m.mu.Lock()
		pending := observedUI(old.state.snapshot, "pending")
		widget := observedUI(old.state.snapshot, "parked-widget")
		notice := observedUI(old.state.snapshot, "parked-notice")
		done := pending != nil && pending.Cancelled && widget != nil && widget.Cancelled && notice != nil && notice.Cancelled
		m.mu.Unlock()
		if done {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("parked requests were not cancelled")
		}
		time.Sleep(time.Millisecond)
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	s = m.Snapshot()
	if s.RuntimeID != runtimeID || len(s.ExtensionWidgets) != 1 || s.ExtensionWidgets[0].Lines[0] != "Original" || len(s.Dialogs) != 0 || len(s.ExtensionNotices) != 0 {
		t.Fatal("warm resume lost parked widgets or surfaced parked UI", s)
	}
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	s = waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if len(s.ExtensionWidgets) != 0 {
		t.Fatal("new session retained previous widget")
	}
}
