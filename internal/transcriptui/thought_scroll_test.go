package transcriptui

import (
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
)

func TestThoughtScrollInputKindsAndOwnerAdmission(t *testing.T) {
	s := ui.ScrollState{Y: 95, MaxY: 100}
	calls := 0
	accept := false
	residual := float32(0)
	fn := thoughtScrollInput(&s, func(d float32) bool { calls++; residual = d; return accept })
	if !fn(ui.InputEvent{Kind: ui.InputScroll, DY: 10}) || s.Y != 95 {
		t.Fatal("rejected owner mutated inner")
	}
	accept = true
	for _, kind := range []ui.InputKind{ui.InputKeyDown, ui.InputText, ui.InputCompose, ui.InputPointerDown, ui.InputPointerUp, ui.InputPointerMove, ui.InputCommand, ui.InputSelect, ui.InputZoom} {
		before := calls
		if fn(ui.InputEvent{Kind: kind, DY: 10}) || calls != before || s.Y != 95 {
			t.Fatalf("consumed unrelated kind%d", kind)
		}
	}
	if !fn(ui.InputEvent{Kind: ui.InputScroll, DY: 10, Mods: ui.Shift, Precise: true}) || s.Y != 100 || residual != 5 {
		t.Fatal("callback-level Shift source contract", s, residual)
	}
	before := calls
	if fn(ui.InputEvent{Kind: ui.InputScroll, DY: float32(math.NaN())}) || calls != before {
		t.Fatal("nonfinite accepted")
	}
	if thoughtScrollInput(&s, nil)(ui.InputEvent{Kind: ui.InputScroll, DY: -10}) {
		t.Fatal("nil opt-in changed default")
	}
	s.Y = 50
	if !fn(ui.InputEvent{Kind: ui.InputScroll, DY: 10}) || residual != 0 || s.Y != 60 {
		t.Fatal("zero-residual owner validation omitted")
	}
}
func TestThoughtOptInStreamFollowDetach(t *testing.T) {
	var state State
	state.Session("optin-follow")
	row := Row{ID: "thought", Role: "assistant", Streaming: true, Steps: []Step{{ID: "step", Thinking: strings.Repeat("Controlled live thought\n", 60), Streaming: true}}}
	o := nativeOptions()
	o.ThoughtScroll = func(float32) bool { return true }
	tt := ui.NewTester(nativeView(func(c *ui.Context) { Turn(c, &state, row, o) }), 850, 900)
	tt.Frame()
	h, ok := tt.Find("Thinking")
	if !ok {
		t.Fatal("header absent")
	}
	var s *ui.ScrollState
	for _, v := range state.thoughtScrolls {
		s = v
	}
	if s == nil || s.Y != s.MaxY {
		t.Fatal("initial follow")
	}
	tt.Scroll(h.X+50, h.Y+h.H+40, 0, -100)
	at := s.Y
	if at >= s.MaxY {
		t.Fatal("did not detach")
	}
	row.Steps[0].Thinking += strings.Repeat("stream growth\n", 20)
	tt.Frame()
	if s.Y != at || s.Y >= s.MaxY {
		t.Fatal("growth stole detached position")
	}
	tt.Scroll(h.X+50, h.Y+h.H+40, 0, 100000)
	row.Steps[0].Thinking += "newest"
	tt.Frame()
	if s.Y != s.MaxY {
		t.Fatal("edge did not rearm follow")
	}
	tt.Scroll(h.X+50, h.Y+h.H+40, 0, -5)
	at = s.Y
	row.Steps[0].Thinking += "\nnext newest"
	tt.Frame()
	if s.Y != at || s.Y >= s.MaxY {
		t.Fatal("reverse did not detach again")
	}
	// Real pointer drag/key Copy through the opted-in body still uses native selection.
	b, ok := tt.Find(row.Steps[0].Thinking)
	if !ok || b.H < 40 {
		t.Fatal("body not visible for selection", b)
	}
	tt.Press(b.X+2, b.Y+5)
	tt.Move(b.X+180, b.Y+40)
	tt.Release(b.X+180, b.Y+40)
	tt.Key(ui.Cmd, ui.KeyC)
	copied := tt.Clipboard()
	if !strings.Contains(copied, "stream growth") || !strings.Contains(copied, "\n") {
		t.Fatalf("Thought selection/copy swallowed: %q", copied)
	}

}
