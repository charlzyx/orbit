package transcriptui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
)

func TestAuditThoughtSingleWheelResidual(t *testing.T) {
	var state State
	state.Session("nested-audit")
	row := Row{ID: "thought", Role: "assistant", Streaming: true, Steps: []Step{{ID: "step", Thinking: strings.Repeat("Controlled thought line\n", 60), Streaming: true}}}
	var list ui.ListState
	var outer ui.ScrollState
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Selectable().Children(func() {
			ui.List(c, &list, 20, func(i int) {
				if i == 0 {
					o := nativeOptions()
					o.ThoughtScroll = func(d float32) bool {
						outer.Y = max(float32(0), min(outer.MaxY, outer.Y+d))
						if d != 0 && outer.Y == outer.MaxY {
							list.ScrollToEnd()
						}
						return true
					}
					Turn(c, &state, row, o)
				} else {
					ui.Text(c, fmt.Sprintf("Outer row %d", i)).Height(60)
				}
			}).Grow(1).TrackScroll(&outer)
		})
	}, 850, 600)
	tt.Frame()
	tt.Frame()
	h, ok := tt.Find("Thinking")
	if !ok {
		t.Fatal("thought header absent", tt.Texts())
	}
	x, y := h.X+50, h.Y+h.H+40
	var inner *ui.ScrollState
	for _, s := range state.thoughtScrolls {
		inner = s
	}
	if inner == nil || inner.MaxY <= 0 {
		t.Fatal("inner scroll absent")
	}
	// Actual precise wheel detaches five DIPs from the inner live edge.
	tt.Scroll(x, y, 0, -5)
	beforeI, beforeO := inner.Y, outer.Y
	if math.Abs(float64(inner.MaxY-beforeI-5)) > .01 {
		t.Fatalf("setup inner=%v max=%v outer=%v", beforeI, inner.MaxY, beforeO)
	}
	tt.Scroll(x, y, 0, 10)
	t.Logf("single event inner %v -> %v (max %v); outer %v -> %v", beforeI, inner.Y, inner.MaxY, beforeO, outer.Y)
	if math.Abs(float64(inner.Y-beforeI-5)) > .01 {
		t.Errorf("inner must consume exactly5")
	}
	if math.Abs(float64(outer.Y-beforeO-5)) > .01 {
		t.Errorf("Rust residual contract: outer must consume remaining5, got %v", outer.Y-beforeO)
	}
	// Inner-only event must not move outer; exact-boundary then next event bubbles.
	outerStart := outer.Y
	tt.Scroll(x, y, 0, -10)
	if math.Abs(float64(inner.MaxY-inner.Y-10)) > .01 || outer.Y != outerStart {
		t.Fatal("mid-card negative control moved outer", inner.Y, outer.Y)
	}
	tt.Scroll(x, y, 0, 10)
	if inner.Y != inner.MaxY || outer.Y != outerStart {
		t.Fatal("exact-boundary negative control moved outer")
	}
	tt.Scroll(x, y, 0, 5)
	if math.Abs(float64(outer.Y-outerStart-5)) > .01 {
		t.Fatal("already-bottom next-event bubbling control failed", outer.Y, outerStart)
	}
	t.Log("mid-card/exact-edge negative controls and already-bottom full bubbling PASS")

}
