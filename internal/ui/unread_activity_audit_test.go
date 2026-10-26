package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

// Source: Rust299 transcript.rs insert_row/remeasure_row and
// message_scroller.rs note_activity/render_jump_button. State/label only;
// this fixture does not assert the still-different floating-button geometry.
func TestAuditScrolledTranscriptNewActivity(t *testing.T) {
	for _, change := range []string{"append", "stream text", "thinking", "tool output"} {
		t.Run(change, func(t *testing.T) {
			a := fixture(t)
			s := Snapshot{RuntimeID: "activity-runtime", SessionID: "activity-session"}
			for i := 0; i < 120; i++ {
				s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "assistant", Text: fmt.Sprintf("Activity row %03d. ", i) + strings.Repeat("Wrapped content. ", 10)})
			}
			s.Messages[119].Streaming = true
			a.SetSnapshot(s)
			tt := ui.NewTester(a.View, 1200, 800)
			tt.Frame()
			for i := 0; i < 12; i++ {
				tt.Scroll(700, 400, 0, -30)
			}
			if a.transcript.AtEnd() {
				t.Fatal("real wheel failed to leave tail")
			}
			first, _ := a.transcript.Visible()
			anchor, ok := tt.Find(strings.TrimSpace(s.Messages[first+1].Text))
			if !ok {
				t.Fatal("mounted anchor missing")
			}
			hasActivity := func() bool {
				return strings.Contains(strings.Join(tt.Texts(), "\n"), a.Tr("message_scroller.new_activity"))
			}
			if hasActivity() {
				t.Fatal("activity before content change")
			}
			a.SetSnapshot(s)
			tt.Frame()
			if hasActivity() {
				t.Fatal("duplicate snapshot reported activity")
			}
			a.reloadLocale()
			tt.Frame()
			if hasActivity() {
				t.Fatal("locale-only redraw reported activity")
			}
			s.Messages = append([]Message(nil), s.Messages...)
			switch change {
			case "append":
				s.Messages = append(s.Messages, Message{ID: "added", Role: "assistant", Text: "New appended response"})
			case "stream text":
				s.Messages[119].Text += " streamed delta"
			case "thinking":
				s.Messages[119].Thinking = "New thinking delta"
			case "tool output":
				s.Messages[119].ToolName = "read"
				s.Messages[119].ToolOutput = "New tool result"
			}
			a.SetSnapshot(s)
			tt.Frame()
			tt.Frame()
			after, _ := a.transcript.Visible()
			bound, ok := tt.Find(strings.TrimSpace(s.Messages[first+1].Text))
			if !ok || after != first || bound.Y != anchor.Y || a.transcript.AtEnd() {
				t.Fatalf("activity moved reader: first=%d after=%d y=%g->%g", first, after, anchor.Y, bound.Y)
			}
			if !hasActivity() {
				t.Error("source New activity label missing while reader is away")
			}
			if err := tt.Click("Jump to latest"); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			if !a.transcript.AtEnd() || hasActivity() {
				t.Fatal("jump did not clear activity/resume tail")
			}
			s.SessionID = "replacement"
			a.SetSnapshot(s)
			tt.Frame()
			if hasActivity() {
				t.Fatal("activity leaked to replacement session")
			}
		})
	}
}
