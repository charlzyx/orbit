package transcriptui

import (
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestThoughtStaleMountedInputMustNotFallThrough(t *testing.T) {
	var state State
	state.Session("old")
	live := true
	o := nativeOptions()
	o.ThoughtScroll = func(float32) bool { return live }
	row := Row{ID: "thought", Role: "assistant", Streaming: true, Steps: []Step{{ID: "step", Thinking: strings.Repeat("Old rendered Thought\n", 60), Streaming: true}}}
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Turn(c, &state, row, o) }), 850, 900)
	tt.Frame()
	b, ok := tt.Find(row.Steps[0].Thinking)
	if !ok || b.H <= 0 {
		t.Fatal("old body absent")
	}
	var old *ui.ScrollState
	for _, s := range state.thoughtScrolls {
		old = s
	}
	before := old.Y
	live = false
	state.Session("new") // Owner changes before rebuilding the old committed hit tree.
	tt.Scroll(b.X+30, b.Y+b.H/2, 0, -10)
	if old.Y != before {
		t.Fatalf("rejected custom event fell through into old native state: %v -> %v", before, old.Y)
	}
}
