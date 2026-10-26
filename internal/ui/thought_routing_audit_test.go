package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestAuditThoughtResidualTailClearsUnread(t *testing.T) {
	for _, nested := range []bool{false, true} {
		t.Run(fmt.Sprint(nested), func(t *testing.T) {
			a := fixture(t)
			s := Snapshot{RuntimeID: "nested-r", SessionID: "nested-s", Running: true}
			for i := 0; i < 30; i++ {
				s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("History %d", i)})
			}
			thought := transcriptui.Row{ID: "thought", Role: "assistant", Streaming: true, Steps: []transcriptui.Step{{ID: "step", Thinking: strings.Repeat("Controlled nested thought\n", 60), Streaming: true}}}
			s.Messages = append(s.Messages, Message{ID: "thought", Role: "assistant", Streaming: true, Logical: &thought})
			a.SetSnapshot(s)
			tt := ui.NewTester(a.View, 1200, 800)
			tt.Frame()
			tt.Frame()
			// Outside the Thought content, use the transcript's left blank margin.
			tt.Scroll(270, 400, 0, -50)
			if a.transcript.AtEnd() {
				t.Fatal("failed to leave tail")
			}
			next := s
			next.Messages = append([]Message(nil), s.Messages...)
			next.Messages[0].Text += " changed"
			a.SetSnapshot(next)
			tt.Frame()
			if !a.transcriptActivity.unread {
				t.Fatal("actual snapshot change did not mark unread")
			}
			x, y := float32(270), float32(400)
			if nested {
				h, ok := tt.Find(thought.Steps[0].Thinking)
				if !ok {
					t.Fatal("visible Thought absent", tt.Texts())
				}
				if h.W <= 0 || h.H <= 0 {
					t.Fatal("thought text not visible", h)
				}
				x, y = h.X+min(float32(50), h.W/2), h.Y+h.H/2
				t.Logf("nested wheel point %v,%v header%+v", x, y, h)
			}
			tt.Scroll(x, y, 0, 100000)
			tt.Frame()
			if !a.transcript.AtEnd() {
				t.Fatal("did not reach tail")
			}
			tt.Scroll(270, 400, 0, -30)
			tt.Frame()
			if a.transcript.AtEnd() {
				t.Fatal("did not leave tail again")
			}
			if got := a.transcriptActivity.unread; got == nested {
				t.Errorf("nested=%t unread=%t; Rust nested scroll_to_end clears, ordinary List wheel retains", nested, got)
			}
		})
	}
}
