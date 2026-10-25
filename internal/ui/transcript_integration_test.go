package orbitui

import (
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestLogicalTranscriptKeepsAssistantProseAndToolResult(t *testing.T) {
	a := fixture(t)
	messages := []agent.Message{
		{ID: "u", Role: "user", Text: "Question 用户"},
		{ID: "a", Role: "assistant", ToolName: "read", Blocks: []agent.ContentBlock{{Type: "text", Text: "Before tool."}, {Type: "toolCall", ToolCallID: "call", ToolName: "read", Arguments: json.RawMessage(`{"path":"example.go"}`)}}},
		{ID: "result", Role: "toolResult", ToolCallID: "call", ToolName: "read", ToolOutput: "unique file contents", Text: "unique file contents"},
		{ID: "b", Role: "assistant", Text: "Answer after tool."},
	}
	a.SetSnapshot(Snapshot{SessionID: "s", RawMessages: messages})
	if len(a.Model.Messages) != 2 || a.Model.Messages[1].Logical == nil {
		t.Fatalf("not grouped: %+v", a.Model.Messages)
	}
	if a.Model.Messages[1].Text != "Before tool.\n\nAnswer after tool." {
		t.Fatal("assistant prose discarded", a.Model.Messages[1].Text)
	}
	a.Search = "unique file contents"
	if len(a.searchMatches()) != 1 {
		t.Fatal("attached tool result lost in search")
	}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	text := strings.Join(tt.Texts(), "\n")
	if !strings.Contains(text, "Before tool.") || !strings.Contains(text, "Answer after tool.") {
		t.Fatal("native turn omitted prose", text)
	}
	if _, ok := tt.Find("Copy Message"); !ok {
		t.Fatal("logical row copy footer missing")
	}
}
func TestHostCrossMessageMarkdownAndAlertSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{SessionID: "s", Messages: []Message{{ID: "u", Role: "user", Text: "First 用户 message."}, {ID: "a", Role: "assistant", Text: "Second **styled** message.\n\n> [!NOTE]\n> Last alert line."}}})
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	first, ok := tt.Find("First 用户 message.")
	if !ok {
		t.Fatal("user missing")
	}
	last, ok := tt.Find("Last alert line.")
	if !ok {
		t.Fatal("alert missing")
	}
	tt.Press(first.X, first.Y+first.H/2)
	tt.Move(last.X+last.W-1, last.Y+last.H/2)
	tt.Release(last.X+last.W-1, last.Y+last.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	text := tt.Clipboard()
	for _, want := range []string{"First 用户 message.", "Second styled message.", "Last alert line."} {
		if !strings.Contains(text, want) {
			t.Fatalf("cross-message alert selection missing %q: %q", want, text)
		}
	}
}

func TestHostRetainedTranscriptSelectionUnmountAndSessionReset(t *testing.T) {
	a := fixture(t)
	var messages []Message
	for i := 0; i < 90; i++ {
		messages = append(messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Row %02d 用户文字", i)})
	}
	a.SetSnapshot(Snapshot{SessionID: "retained-a", Messages: messages})
	tt := ui.NewTester(a.View, 1200, 800)
	a.transcript.ScrollTo(0, ui.Start)
	tt.Frame()
	tt.Frame()
	first, ok := tt.Find(messages[0].Text)
	if !ok {
		t.Fatal("first row not mounted")
	}
	last, ok := tt.Find(messages[2].Text)
	if !ok {
		t.Fatal("third row not mounted")
	}
	tt.Press(first.X+1, first.Y+first.H/2)
	tt.Move(last.X+last.W-1, last.Y+last.H/2)
	tt.Release(last.X+last.W-1, last.Y+last.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	copied := tt.Clipboard()
	for _, want := range []string{messages[0].Text, messages[1].Text, messages[2].Text} {
		if !strings.Contains(copied, want) {
			t.Fatal("initial selection incomplete", copied)
		}
	}
	a.transcript.ScrollToEnd()
	for i := 0; i < 5; i++ {
		tt.Frame()
	}
	tt.SetClipboard("retention sentinel")
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != copied {
		t.Fatal("virtualized selection was lost", tt.Clipboard())
	}
	a.SetSnapshot(Snapshot{SessionID: "retained-b", Messages: []Message{{ID: "other", Role: "user", Text: "Other session"}}})
	tt.Frame()
	tt.SetClipboard("reset sentinel")
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != "reset sentinel" {
		t.Fatal("selection leaked into another session", tt.Clipboard())
	}
}
