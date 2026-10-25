package orbitui

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func TestNativeExtensionDialogKeyboardInputAndFocus(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "s", Dialog: &DialogRequest{ID: "select", Method: "select", Title: "Pick", Options: []string{"One", "Two", "Three"}}})
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Value != "Three" {
		t.Fatalf("wrapping select response %v", actions)
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 {
		t.Fatal("duplicate response")
	}
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "s", Dialog: &DialogRequest{ID: "input", Method: "input", Title: "Input", Placeholder: "Expected placeholder"}})
	tt.Frame()
	if !tt.Focused("Input") || a.Model.Dialog.Placeholder != "Expected placeholder" {
		t.Fatal("input focus/placeholder missing", tt.Texts())
	}
	tt.Type("a typed answer")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 2 || actions[1].Value != "a typed answer" {
		t.Fatalf("input Enter %v", actions)
	}
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "s", Dialog: &DialogRequest{ID: "edit", Method: "editor", Title: "Edit", Default: "first"}})
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("replacement")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 3 || actions[2].Value != "replacement" {
		t.Fatalf("editor Enter %v", actions)
	}
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "s", Dialog: &DialogRequest{ID: "cancel", Method: "input", Title: "Cancel input"}})
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if len(actions) != 4 || actions[3].Data["cancelled"] != true {
		t.Fatalf("Escape %v", actions)
	}
	a.SetSnapshot(Snapshot{Connected: true, SessionID: "s"})
	tt.Frame()
	tt.Type("composer restored")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 5 || actions[4].Kind != "Submit" || actions[4].Value != "composer restored" {
		t.Fatalf("composer focus %v", actions)
	}
}
func TestNativeExtensionDialogIMEAndConfirmFalse(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "ime", Method: "input", Title: "IME answer"}})
	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	tt.Compose("你好", 2)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 0 {
		t.Fatal("IME candidate committed dialog")
	}
	tt.Compose("", 0)
	tt.Type("你好")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Value != "你好" {
		t.Fatal(actions)
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "confirm", Method: "confirm", Title: "Proceed?", Message: "Confirmation detail"}})
	tt.Frame()
	if e := tt.Click("No"); e != nil {
		t.Fatal(e)
	}
	if len(actions) != 2 || actions[1].Data["confirmed"] != false || actions[1].Data["cancelled"] != nil {
		t.Fatal(actions)
	}
}
func TestNativeInlineApprovalKeysDoNotAbortOrSubmit(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.Draft = "preserve draft"
	snapshot := Snapshot{SessionID: "s", Connected: true, Running: true, Approval: &agent.ApprovalRequest{ID: "guard", Tool: "bash", Detail: "echo fixture", Options: []string{"Allow once", "Always allow this tool", "Deny"}}}
	a.SetSnapshot(snapshot)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("Tool approval"); !ok {
		t.Fatal("approval not inline")
	}
	if a.modalOpen() {
		t.Fatal("inline approval became modal")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "RespondDialog" || actions[0].Value != "Always allow this tool" || a.Draft != "preserve draft" {
		t.Fatal(actions, a.Draft)
	}
	snapshot.Approval = &agent.ApprovalRequest{ID: "guard2", Tool: "edit", Detail: "file", Options: []string{"Allow once", "Always allow this tool", "Deny"}}
	a.SetSnapshot(snapshot)
	tt.Frame()
	tt.Key(0, ui.KeyEscape)
	if len(actions) != 2 || actions[1].Kind != "RespondDialog" || actions[1].Data["cancelled"] != true {
		t.Fatalf("Escape aborted instead of denied: %v", actions)
	}
}
func TestNativeQuestionnaireBackMultiCustomPreviewAndCommit(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	ask := &agent.AskPrompt{ID: "q", ToolCallID: "tool", Questions: []agent.AskQuestion{
		{Header: "Method", Question: "Which method?", Options: []agent.AskOption{{Label: "A", Description: "Alpha"}, {Label: "B", Description: "Beta", Preview: "Option preview content"}}},
		{Question: "Which checks?", MultiSelect: true, Options: []agent.AskOption{{Label: "X"}, {Label: "Y"}}},
		{Question: "What else?", Options: []agent.AskOption{{Label: "Nothing"}}},
	}}
	a.SetSnapshot(Snapshot{SessionID: "s", Connected: true, Running: true, Ask: ask})
	tt := ui.NewTester(a.View, 1470, 1000)
	tt.Frame()
	if a.modalOpen() {
		t.Fatal("questionnaire became modal")
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	if !tt.HasText("Option preview content") {
		t.Fatal("focused option preview missing")
	}
	if e := tt.Click("Next"); e != nil {
		t.Fatal(e)
	}
	if len(actions) != 0 {
		t.Fatal("partial answers were sent")
	}
	if e := tt.Click("X"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeySpace)
	if e := tt.Click("Previous question"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if a.extensionUI.askHighlight[0] != 1 {
		t.Fatal("back lost answer")
	}
	if e := tt.Click("A"); e != nil {
		t.Fatal(e, "cursor", a.extensionUI.askCursor, "texts", tt.Texts())
	}
	if e := tt.Click("Next"); e != nil {
		t.Fatal(e)
	}
	if !a.extensionUI.askAnswers[1].Checked[0] || !a.extensionUI.askAnswers[1].Checked[1] {
		t.Fatal("multi choice lost on back")
	}
	if e := tt.Click("Next"); e != nil {
		t.Fatal(e)
	}
	if e := tt.Click("Type something."); e != nil {
		t.Fatal(e)
	}
	tt.Type("custom result")
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "SubmitAsk" {
		t.Fatalf("commit %v", actions)
	}
	answers := actions[0].Data["answers"].([]agent.AskAnswer)
	if answers[0].Option != 0 || !answers[1].Checked[0] || !answers[1].Checked[1] || answers[2].Custom != "custom result" || answers[2].Kind != "custom" {
		t.Fatal(answers)
	}
	tt.Frame()
	if !tt.HasText("Sending your answers…") {
		t.Fatal("committed panel not read-only")
	}
	if e := tt.Click("Decline questionnaire"); e == nil && len(actions) != 1 {
		t.Fatal("submitted questionnaire declined")
	}
}
func TestNativeQuestionnaireEscapeAndIME(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	ask := &agent.AskPrompt{ID: "q", Questions: []agent.AskQuestion{{Question: "Write a reply", Options: nil}}}
	a.SetSnapshot(Snapshot{Connected: true, Running: true, Ask: ask})
	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	tt.Compose("答", 1)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 0 {
		t.Fatal("IME advanced questionnaire")
	}
	tt.Compose("", 0)
	tt.Key(0, ui.KeyEscape)
	if len(actions) != 1 || actions[0].Kind != "RespondDialog" || actions[0].Data["cancelled"] != true {
		t.Fatal(actions)
	}
}
func TestNativeWidgetPlacementAndNoticeDedup(t *testing.T) {
	a := fixture(t)
	s := Snapshot{SessionID: "s", Connected: true, Status: "quiet", ExtensionWidgets: []agent.ExtensionWidget{{Key: "above", Placement: "aboveEditor", Lines: []string{"Above widget"}}, {Key: "below", Placement: "belowEditor", Lines: []string{"Below widget"}}}, ExtensionNotices: []agent.ExtensionNotice{{ID: "n", Sequence: 1, Message: "Visible toast"}}}
	a.SetSnapshot(s)
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 1000)
	tt.Frame()
	above, ok := tt.Find("Above widget")
	if !ok {
		t.Fatal("above widget absent")
	}
	below, ok := tt.Find("Below widget")
	if !ok {
		t.Fatal("below widget absent")
	}
	composer, ok := tt.Find("Composer box")
	if !ok {
		t.Fatal("composer absent")
	}
	if above.Y >= composer.Y || below.Y < composer.Y+composer.H {
		t.Fatal("widget placements", above, composer, below)
	}
	if !tt.HasText("Visible toast") || a.Model.Status != "quiet" {
		t.Fatal("notify not separate from status")
	}
	// Let TTL elapse; replaying an unchanged snapshot must not revive the toast.
	a.toastStack.Expire(time.Now().Add(20*time.Second), true)
	a.SetSnapshot(s)
	tt.Frame()
	if tt.HasText("Visible toast") {
		t.Fatal("snapshot replay revived toast")
	}
	s.ExtensionWidgets = []agent.ExtensionWidget{{Key: "above", Placement: "belowEditor", Lines: []string{"Replacement"}}}
	a.SetSnapshot(s)
	tt.Frame()
	if tt.HasText("Above widget") || tt.HasText("Below widget") || !tt.HasText("Replacement") {
		t.Fatal("replace/clear projection stale")
	}
}
func TestWidgetStyledLinesSourceAttributesAndControls(t *testing.T) {
	a := fixture(t)
	spans := widgetSpans("\x1b[1;3;4;9;38;2;255;0;0m红\x1b[0m plain\x1b]8;;https://example.test\a link\x1b]8;;\a", a.Theme)
	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.Text)
	}
	if text.String() != "红 plain link" || len(spans) != 3 || spans[0].Weight != 700 || !spans[0].Italic || !spans[0].Underline || !spans[0].Strikethrough || spans[0].Color != ui.RGB(255, 0, 0) || spans[1].Weight != 400 {
		t.Fatal(spans, text.String())
	}
	if s := widgetSpans("\x1b[2K", a.Theme); len(s) != 1 || s[0].Text != " " {
		t.Fatal("empty line lost height", s)
	}
	if s := widgetSpans("\x1b[38;5;2mgreen", a.Theme); s[0].Color != a.Theme.Terminal().Palette[2] {
		t.Fatal(s)
	}
	if s := widgetSpans("\x1b[7minverse", a.Theme); s[0].Color != a.Theme.CodeBg || s[0].Background != a.Theme.CodeText {
		t.Fatal(s)
	}
}
func TestNativeQuestionnaireHistoryFromPiMessages(t *testing.T) {
	a := fixture(t)
	args := json.RawMessage(`{"questions":[{"question":"Which?","options":[{"label":"A"},{"label":"B"}]}]}`)
	snapshot := Snapshot{SessionID: "s", Connected: true, RawMessages: []agent.Message{
		{ID: "assistant", Role: "assistant", Blocks: []agent.ContentBlock{{Type: "toolCall", ToolCallID: "q", ToolName: "ask_user_question", Arguments: args}}},
		{ID: "tool:q", Role: "toolResult", ToolCallID: "q", ToolName: "ask_user_question", ToolOutput: `User has answered your questions: "Which?"="A".`},
	}}
	a.SetSnapshot(snapshot)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if e := tt.Click("Worked"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if e := tt.Click("Ran 1 tool"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if !tt.HasText("Which?") || !tt.HasText("Question") || tt.HasText("Waiting for an answer…") {
		t.Fatal("answered native history missing", tt.Texts(), a.Model.Messages)
	}
	snapshot.RawMessages[1].ToolOutput = "User declined to answer questions"
	a.SetSnapshot(snapshot)
	tt.Frame()
	if !tt.HasText("No answer") {
		t.Fatal("declined native history missing")
	}
	snapshot.RawMessages = snapshot.RawMessages[:1]
	snapshot.Running = true
	snapshot.RawMessages[0].Streaming = true
	a.SetSnapshot(snapshot)
	tt.Frame()
	if !tt.HasText("Waiting for an answer…") {
		t.Fatal("waiting native history missing")
	}
}

func TestNativeExtensionEditorDeliveryRevisionCaretUndoAndBlur(t *testing.T) {
	a := fixture(t)
	s := Snapshot{RuntimeID: "runtime-1", SessionID: "s", Connected: true}
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Type("original typed")
	tt.Key(ui.Cmd, ui.KeyA)
	s.ComposerText = "seed用户"
	s.ComposerRevision = 1
	a.SetSnapshot(s)
	tt.Frame()
	if a.Draft != "seed用户" || !tt.Focused("Message") {
		t.Fatal("replacement/focus", a.Draft, tt.Texts())
	}
	caret := a.reviewState.layout.composer
	if caret.anchor != 6 || caret.caret != 6 {
		t.Fatal("replacement caret not at rune end", caret)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != "seed用户" {
		t.Fatal("replacement retained pre-seed undo", a.Draft)
	}
	tt.Type("!")
	a.SetSnapshot(s)
	tt.Frame()
	if a.Draft != "seed用户!" {
		t.Fatal("same revision overwrote later user typing", a.Draft)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != "seed用户" {
		t.Fatal("unchanged snapshot lost native undo", a.Draft)
	}
	// Even an identical-value new event resets selection and undo.
	tt.Key(ui.Cmd, ui.KeyA)
	s.ComposerRevision = 2
	a.SetSnapshot(s)
	tt.Frame()
	tt.Type("?")
	if a.Draft != "seed用户?" {
		t.Fatal("same-value revision did not reset selection", a.Draft)
	}
	s.ComposerText = ""
	s.ComposerRevision = 3
	a.SetSnapshot(s)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != "" {
		t.Fatal("clear event not delivered/reset", a.Draft)
	}
	// set_editor_text changes content, not focus ownership.
	tt.Key(ui.Cmd, ui.KeyF)
	tt.Frame()
	if !tt.Focused("Find in transcript") {
		t.Fatal("find fixture focus")
	}
	s.ComposerText = "background seed"
	s.ComposerRevision = 4
	a.SetSnapshot(s)
	tt.Frame()
	if !tt.Focused("Find in transcript") || a.Draft != "background seed" {
		t.Fatal("replacement stole focus", a.Draft, tt.Texts())
	}
	// A restarted process may reuse session id and revision 1.
	s.RuntimeID = "runtime-2"
	s.ComposerRevision = 1
	s.ComposerText = "restart"
	a.SetSnapshot(s)
	tt.Frame()
	if a.Draft != "restart" {
		t.Fatal("restart revision suppressed")
	}
}
func TestNativeExtensionHiddenArrivalFocusAndReplacement(t *testing.T) {
	a, tt := mountedReview(t)
	if e := tt.Click("Message"); e != nil {
		t.Fatal(e)
	}
	tt.Type("before")
	tt.Key(ui.Cmd, ui.KeyA)
	if e := tt.Click("Full width review"); e != nil {
		t.Fatal(e)
	}
	s := a.Model
	s.RuntimeID = "runtime"
	s.ComposerRevision = 1
	s.ComposerText = "after"
	a.SetSnapshot(s)
	tt.Frame()
	if e := tt.Click("Restore review width"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if a.Draft != "after" || tt.Focused("Message") {
		t.Fatal("hidden replacement stole focus or lost content", a.Draft)
	}
	caret := a.reviewState.layout.composer
	if caret.anchor != 5 || caret.caret != 5 {
		t.Fatal("old retained selection replaced seed caret", caret)
	}
	if e := tt.Click("Full width review"); e != nil {
		t.Fatal(e)
	}
	s.Approval = &agent.ApprovalRequest{ID: "hidden", Tool: "bash", Options: []string{"Allow once", "Deny"}}
	a.SetSnapshot(s)
	tt.Frame()
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	// A hidden approval must not be keyboard-active under Review.
	tt.Key(0, ui.KeyEnter)
	for _, x := range actions {
		if x.Kind == "RespondDialog" {
			t.Fatal("hidden panel answered", actions)
		}
	}
	if e := tt.Click("Restore review width"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if !tt.Focused("Tool approval") {
		t.Fatal("revealed pending approval did not take focus")
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "RespondDialog" || actions[0].Value != "Allow once" {
		t.Fatal(actions)
	}
}
func TestNativeExtensionStatusIsQuietAndExpires(t *testing.T) {
	a := fixture(t)
	now := time.Now()
	s := Snapshot{ExtensionStatusText: "Quiet transient status", ExtensionStatusAt: now}
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if !tt.HasText("Quiet transient status") || len(a.toastStack.Items()) != 0 {
		t.Fatal("status not quiet/visible", tt.Texts())
	}
	if !extensionStatusVisible(now, now.Add(6*time.Second-time.Nanosecond)) || extensionStatusVisible(now, now.Add(6*time.Second)) {
		t.Fatal("source six-second expiry mismatch")
	}
	s.ExtensionStatusAt = now.Add(-7 * time.Second)
	a.SetSnapshot(s)
	tt.Frame()
	if tt.HasText("Quiet transient status") {
		t.Fatal("expired status retained")
	}
}

func TestNativeExtensionProjectionAndSourceLabels(t *testing.T) {
	now := time.Unix(123, 0)
	input := agent.Snapshot{RuntimeID: "runtime", ComposerRevision: 3, ComposerText: "", ExtensionStatusText: "quiet", ExtensionStatusAt: now, Approval: &agent.ApprovalRequest{ID: "guard"}, Ask: &agent.AskPrompt{ID: "ask"}, ExtensionWidgets: []agent.ExtensionWidget{{Key: "widget", Placement: "belowEditor", Lines: []string{"line"}}}, ExtensionNotices: []agent.ExtensionNotice{{ID: "notice", Sequence: 1, Message: "hello"}}, Dialogs: []agent.PendingDialog{{ID: "input", Method: "input", Placeholder: "placeholder"}}}
	got := adaptSnapshot(input, "", nil, nil)
	if got.RuntimeID != "runtime" || got.ComposerRevision != 3 || got.ComposerText != "" || got.ExtensionStatusText != "quiet" || !got.ExtensionStatusAt.Equal(now) || got.Approval.ID != "guard" || got.Ask.ID != "ask" || got.Dialog.Placeholder != "placeholder" || got.ExtensionWidgets[0].Key != "widget" || got.ExtensionNotices[0].ID != "notice" {
		t.Fatal(got)
	}
	a := fixture(t)
	a.SetSnapshot(Snapshot{Approval: &agent.ApprovalRequest{ID: "mcp", Tool: "mcp__repo__open_issue", Options: []string{"Allow once", "Always allow this tool", "Deny"}}})
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if !tt.HasText("Allow MCP · repo · open_issue?") {
		t.Fatal("MCP permission scope hidden", tt.Texts())
	}
	if box, ok := tt.Find("Tool approval"); !ok || box.H != 48 {
		t.Fatal("source approval bar geometry", box, ok)
	}
}

func TestNativeExtensionEmptySelectConfirmAndScrim(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "empty", Method: "select", Title: "No choices"}})
	tt := ui.NewTester(a.View, 1000, 700)
	tt.Frame()
	if !tt.HasText("No options were provided.") {
		t.Fatal("empty state missing")
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Data["cancelled"] != true {
		t.Fatal(actions)
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "yes", Method: "confirm", Title: "Yes by default?"}})
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 2 || actions[1].Data["confirmed"] != true {
		t.Fatal(actions)
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "scrim", Method: "input", Title: "Click outside"}})
	tt.Frame()
	tt.ClickAt(20, 680)
	if len(actions) != 3 || actions[2].Data["cancelled"] != true {
		t.Fatal(actions)
	}
}

// Measure source sibling spacing in the mounted composer, rather than merely
// asserting helper tokens. No generic parent gap may add space at placements.
func TestNativeExtensionSourceSiblingGeometry(t *testing.T) {
	for _, panel := range []string{"approval", "ask"} {
		t.Run(panel, func(t *testing.T) {
			a := fixture(t)
			s := Snapshot{Connected: true, SessionID: "spacing", ExtensionWidgets: []agent.ExtensionWidget{
				{Key: "one", Placement: "aboveEditor", Lines: []string{"First line"}},
				{Key: "two", Placement: "aboveEditor", Lines: []string{"Second line"}},
				{Key: "below", Placement: "belowEditor", Lines: []string{"Below line"}},
			}}
			label, bottom := "Tool approval", float32(8)
			if panel == "approval" {
				s.Approval = &agent.ApprovalRequest{ID: "guard", Tool: "bash", Options: []string{"Allow once", "Always allow this tool", "Deny"}}
			} else {
				s.Ask = &agent.AskPrompt{ID: "question", Questions: []agent.AskQuestion{{Question: "Choose?", Options: []agent.AskOption{{Label: "A"}}}}}
				label, bottom = "Questionnaire", a.Theme.Space(8)
			}
			a.SetSnapshot(s)
			tt := ui.NewTester(a.View, 1470, 1100)
			tt.Frame()
			p, okP := tt.Find(label)
			one, okOne := tt.Find("Extension widget one")
			two, okTwo := tt.Find("Extension widget two")
			below, okBelow := tt.Find("Extension widget below")
			composer, okComposer := tt.Find("Composer box")
			if !okP || !okOne || !okTwo || !okBelow || !okComposer {
				t.Fatal("missing mounted source siblings", tt.Texts())
			}
			if got := one.Y - (p.Y + p.H); got != bottom {
				t.Fatalf("%s bottom gap = %v, want %v", panel, got, bottom)
			}
			if got := two.Y - (one.Y + one.H); got != a.Theme.Space(6) {
				t.Fatalf("intra-widget gap = %v, want %v", got, a.Theme.Space(6))
			}
			if composer.Y != two.Y+two.H || below.Y != composer.Y+composer.H {
				t.Fatal("invented outer widget gap", two, composer, below)
			}
			before := one.H
			a.Prefs.EditorFontSize *= 2
			tt.Frame()
			one, _ = tt.Find("Extension widget one")
			if one.H-before != 18 {
				t.Fatalf("widget ignored editor typography: height %v -> %v", before, one.H)
			}
		})
	}
}

func TestNativeExtensionPointerThenKeyboardUsesExactChoice(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "select-pointer", Method: "select", Title: "Select by pointer", Options: []string{"First", "Second", "Third"}}})
	tt := ui.NewTester(a.View, 1000, 800)
	tt.Frame()
	second, ok := tt.Find("Second")
	if !ok {
		t.Fatal("missing option")
	}
	tt.Move(second.X+5, second.Y+5)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Data["id"] != "select-pointer" || actions[0].Value != "Second" {
		t.Fatal("pointer selection was not committed", actions)
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "select-keys", Method: "select", Title: "Pointer and keys", Options: []string{"First", "Second", "Third"}}})
	tt.Frame()
	tt.Move(0, 0)
	second, _ = tt.Find("Second")
	tt.Move(second.X+5, second.Y+5)
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 2 || actions[1].Data["id"] != "select-keys" || actions[1].Value != "Third" {
		t.Fatal("stationary pointer overrode keyboard selection", actions)
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "confirm-pointer", Method: "confirm", Title: "Confirm by pointer"}})
	tt.Frame()
	no, ok := tt.Find("No")
	if !ok {
		t.Fatal("missing No")
	}
	tt.Move(no.X+5, no.Y+5)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 3 || actions[2].Data["id"] != "confirm-pointer" || actions[2].Data["confirmed"] != false || actions[2].Data["cancelled"] == true {
		t.Fatal("hover No + Enter did not return confirmed:false", actions)
	}
}

func TestNativeQuestionnaireMultilineCustomAnswer(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	a.SetSnapshot(Snapshot{Ask: &agent.AskPrompt{ID: "multiline", Questions: []agent.AskQuestion{{Question: "Explain"}}}})
	tt := ui.NewTester(a.View, 1200, 1000)
	tt.Frame()
	input, ok := tt.Find("Custom answer")
	if !ok || !tt.Focused("Custom answer") {
		t.Fatal("custom input not focused")
	}
	firstHeight := input.H
	tt.Type("First")
	tt.Key(ui.Shift, ui.KeyEnter)
	tt.SetClipboard("Second\nThird\nFourth")
	tt.Key(ui.Cmd, ui.KeyV)
	tt.Frame()
	input, _ = tt.Find("Custom answer")
	if len(actions) != 0 || a.extensionUI.askAnswers[0].Custom != "First\nSecond\nThird\nFourth" {
		t.Fatal("multiline content was flattened or prematurely submitted", a.extensionUI.askAnswers[0], actions)
	}
	if input.H <= firstHeight || input.H-firstHeight > a.Theme.FontSize*sourceDefaultLineHeight*2+1 {
		t.Fatal("custom answer does not cap at three lines", firstHeight, input.H)
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "SubmitAsk" || actions[0].Data["answers"].([]agent.AskAnswer)[0].Custom != "First\nSecond\nThird\nFourth" {
		t.Fatal("plain Enter did not preserve the multiline answer", actions)
	}
}

func TestNativeExtensionKeyboardRevealIgnoresStationaryPointer(t *testing.T) {
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	var options []string
	for i := 0; i < 24; i++ {
		options = append(options, fmt.Sprintf("Option %02d", i))
	}
	a.SetSnapshot(Snapshot{Dialog: &DialogRequest{ID: "scroll-pointer", Method: "select", Title: "Long options", Options: options}})
	tt := ui.NewTester(a.View, 1000, 800)
	tt.Frame()
	option, ok := tt.Find("Option 02")
	if !ok {
		t.Fatal("missing visible option")
	}
	tt.Move(option.X+8, option.Y+8)
	if a.extensionUI.dialogCursor != 2 {
		t.Fatal("pointer did not select option2")
	}
	for i := 0; i < 15; i++ {
		tt.Key(0, ui.KeyDown)
		tt.Frame()
	}
	if a.extensionUI.dialogCursor != 17 {
		t.Fatalf("scroll moved a row under the stationary cursor: selected %d, want17", a.extensionUI.dialogCursor)
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Data["id"] != "scroll-pointer" || actions[0].Value != "Option 17" {
		t.Fatal("wrong keyboard-revealed response", actions)
	}
}
