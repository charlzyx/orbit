package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestThoughtStaleAppHitTreeDoesNotScrollNewRuntime(t *testing.T) {
	a := fixture(t)
	s := Snapshot{RuntimeID: "old-runtime", SessionID: "same-session", Running: true}
	for i := 0; i < 30; i++ {
		s.Messages = append(s.Messages, Message{ID: fmt.Sprint(i), Role: "user", Text: fmt.Sprintf("History %d", i)})
	}
	thought := transcriptui.Row{ID: "thought", Role: "assistant", Streaming: true, Steps: []transcriptui.Step{{ID: "step", Thinking: strings.Repeat("Owned actual Thought\n", 60), Streaming: true}}}
	s.Messages = append(s.Messages, Message{ID: "thought", Role: "assistant", Streaming: true, Logical: &thought})
	a.SetSnapshot(s)
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	tt.Scroll(270, 400, 0, -50)
	b, ok := tt.Find(thought.Steps[0].Thinking)
	if !ok || b.H <= 0 {
		t.Fatal("body absent")
	}
	before := a.transcriptScroll.Y
	s.RuntimeID = "new-runtime"
	a.SetSnapshot(s) // No frame: event still targets previous rendered owner.
	unread := a.transcriptActivity.unread
	tt.Scroll(b.X+30, b.Y+b.H/2, 0, 100000)
	if a.transcriptScroll.Y != before || a.transcriptActivity.unread != unread {
		t.Fatalf("old committed target mutated new runtime: outer %v -> %v unread%t -> %t", before, a.transcriptScroll.Y, unread, a.transcriptActivity.unread)
	}
}
