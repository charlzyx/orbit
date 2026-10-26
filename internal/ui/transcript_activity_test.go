package orbitui

import (
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"testing"
)

func activityRawFixture(t *testing.T) (*App, *ui.Tester, Snapshot) {
	t.Helper()
	a := fixture(t)
	s := Snapshot{RuntimeID: "r", SessionID: "s", Running: true}
	for i := 0; i < 100; i++ {
		s.RawMessages = append(s.RawMessages, agent.Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Raw history row %d", i)})
	}
	s.RawMessages = append(s.RawMessages, agent.Message{ID: "assistant", Role: "assistant", Streaming: true, Blocks: []agent.ContentBlock{{Type: "toolCall", ToolCallID: "call", ToolName: "bash", Arguments: json.RawMessage(`{"command":"test"}`)}}}, agent.Message{ID: "tool:call", Role: "tool", ToolCallID: "call", ToolName: "bash", Streaming: true, ToolOutput: "first"})
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	return a, tt, s
}
func activityLeaveTail(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	for i := 0; i < 10; i++ {
		tt.Scroll(700, 400, 0, -30)
	}
	if a.transcript.AtEnd() {
		t.Fatal("wheel failed")
	}
}
func TestTranscriptActivityRawToolPartialAndCompletion(t *testing.T) {
	for _, open := range []bool{false, true} {
		t.Run(fmt.Sprint(open), func(t *testing.T) {
			a, tt, s := activityRawFixture(t)
			if open {
				a.transcriptUI.Toggle("tool:call", false)
				tt.Frame()
			}
			activityLeaveTail(t, a, tt)
			s.RawMessages = append([]agent.Message(nil), s.RawMessages...)
			s.RawMessages[len(s.RawMessages)-1].ToolOutput = "partial changed"
			a.SetSnapshot(s)
			tt.Frame()
			if a.transcriptActivity.unread != open {
				t.Fatalf("partial: unread=%t open=%t", a.transcriptActivity.unread, open)
			}
			a.transcriptActivity.unread = false
			s.RawMessages = append([]agent.Message(nil), s.RawMessages...)
			s.RawMessages[len(s.RawMessages)-1].Streaming = false
			a.SetSnapshot(s)
			tt.Frame()
			if !a.transcriptActivity.unread {
				t.Fatal("completion must report even when collapsed")
			}
		})
	}
}
func TestTranscriptActivityTailInitializationResetAndReorder(t *testing.T) {
	a := fixture(t)
	s := Snapshot{RuntimeID: "r", SessionID: "s"}
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	s.Messages = []Message{{ID: "first", Role: "user", Text: "First message"}}
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("empty to first message flashed activity")
	}
	for i := 1; i < 100; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("Row %d", i)})
	}
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread || !a.transcript.AtEnd() {
		t.Fatal("following tail signaled activity")
	}
	activityLeaveTail(t, a, tt)
	s.Messages = append([]Message(nil), s.Messages...)
	s.Messages[99].Text += " update"
	a.SetSnapshot(s)
	tt.Frame()
	if !a.transcriptActivity.unread {
		t.Fatal("missing activity")
	}
	tt.Scroll(700, 400, 0, 100000)
	tt.Frame()
	if !a.transcript.AtEnd() || !a.transcriptActivity.unread {
		t.Fatal("ordinary wheel must hide, not erase sticky activity")
	}
	a.SetSnapshot(s)
	tt.Frame()
	if !a.transcriptActivity.unread {
		t.Fatal("unchanged snapshot cleared source sticky activity")
	}
	activityLeaveTail(t, a, tt)
	if _, ok := tt.Find(a.Tr("message_scroller.new_activity")); !ok {
		t.Fatal("sticky activity did not reappear")
	}
	if err := tt.Click("Jump to latest"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("explicit jump did not clear")
	}
	activityLeaveTail(t, a, tt)
	s.Messages = append([]Message(nil), s.Messages...)
	s.Messages[99].Text += " tail cycle"
	a.SetSnapshot(s)
	tt.Frame()
	tt.Scroll(700, 400, 0, 100000)
	tt.Frame()
	if !a.transcriptActivity.unread || !a.transcript.AtEnd() {
		t.Fatal("tail cycle fixture failed")
	}
	s.Messages = append([]Message(nil), s.Messages...)
	s.Messages[99].Text += " following delta"
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("content remeasure while following did not clear sticky activity")
	}
	activityLeaveTail(t, a, tt)
	s.Messages = append([]Message(nil), s.Messages...)
	s.Messages[99].Text += " again"
	a.SetSnapshot(s)
	tt.Frame()
	s.Messages = append([]Message(nil), s.Messages...)
	s.Messages[1], s.Messages[2] = s.Messages[2], s.Messages[1]
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("history reorder treated as new activity")
	}
	s.Messages = s.Messages[:50]
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("history deletion treated as activity")
	}
	s.RuntimeID = "other-runtime"
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("same session ID runtime replacement leaked activity")
	}
}
func TestTranscriptActivityIgnoresClockUsageAndDetectsImage(t *testing.T) {
	a, tt, s := activityRawFixture(t)
	activityLeaveTail(t, a, tt)
	s.RawMessages = append([]agent.Message(nil), s.RawMessages...)
	s.RawMessages[100].Usage = json.RawMessage(`{"input":42,"output":17}`)
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("usage alone signaled activity")
	}
	// Exercise content comparison without decoding any image or inventing a large fixture.
	x := Message{Images: []agent.Image{{MIMEType: "image/png", Data: "old"}}}
	y := Message{Images: []agent.Image{{MIMEType: "image/png", Data: "new"}}}
	if !a.transcriptMessageChanged(x, y) {
		t.Fatal("same-size image replacement omitted")
	}
	if a.transcriptMessageChanged(x, x) {
		t.Fatal("unchanged image reported")
	}
}

func TestTranscriptActivityOpaqueOwnerAndHiddenRoundtrip(t *testing.T) {
	a, tt, s := activityRawFixture(t)
	var generation uint64 = 1
	a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, true }
	a.SetSnapshot(s)
	tt.Frame()
	activityLeaveTail(t, a, tt)
	s.RawMessages = append([]agent.Message(nil), s.RawMessages...)
	s.RawMessages[100].Text = "new response"
	a.SetSnapshot(s)
	tt.Frame()
	if !a.transcriptActivity.unread {
		t.Fatal("new owner activity missing")
	}
	// Identical visible IDs are not enough: a new task generation owns the view.
	generation++
	a.SetSnapshot(s)
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("opaque task generation leaked unread")
	}
	a.Navigate("settings")
	old := a.inlineImageToken()
	generation++
	a.SetSnapshot(s)
	if a.transcriptActivity.owner == old {
		t.Fatal("hidden owner not adopted")
	}
	a.Navigate("chat")
	tt.Frame()
	if a.transcriptActivity.unread {
		t.Fatal("hidden owner roundtrip leaked unread")
	}
}

func TestTranscriptActivityExplicitSubmitClears(t *testing.T) {
	a, tt, s := activityRawFixture(t)
	activityLeaveTail(t, a, tt)
	s.RawMessages = append([]agent.Message(nil), s.RawMessages...)
	s.RawMessages[100].Text = "new delta"
	a.SetSnapshot(s)
	tt.Frame()
	if !a.transcriptActivity.unread {
		t.Fatal("missing unread before submit")
	}
	submitted := 0
	a.Options.OnAction = func(x Action) {
		if x.Kind == "Submit" {
			submitted++
		}
	}
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("next question")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if submitted != 1 || a.transcriptActivity.unread || !a.transcript.AtEnd() {
		t.Fatalf("submit route count=%d unread=%t tail=%t", submitted, a.transcriptActivity.unread, a.transcript.AtEnd())
	}
}
