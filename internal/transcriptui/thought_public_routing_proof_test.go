package transcriptui

import (
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

// Test-only app-level mechanism, NOT installed on production Thought.
func TestAuditPublicThoughtInputRoutingMechanism(t *testing.T) {
	cases := []struct {
		name                 string
		inner, outer, dx, dy float32
		wheel, shift         bool
	}{
		{"bottom-overshoot", 795, 0, 0, 10, false, false}, {"top-overshoot", 5, 20, 0, -10, false, false},
		{"mid-card", 50, 20, 0, 10, false, false}, {"horizontal-fallback", 795, 0, 10, 0, false, false},
		{"no-outer-range", 795, 0, 0, 10, false, false}, {"short-inner", 0, 0, 0, 10, false, false},
		{"notch-same-dip", 795, 0, 0, 10, true, false},
		{"beyond-both-ends", 795, 0, 0, 100000, false, false}, {"outer-start-clamp", 5, 0, 0, -100000, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var inner, outer ui.ScrollState
			var list ui.ListState
			var bounds ui.Rect
			var seen ui.InputEvent
			calls := 0
			live := true
			count, contentHeight := 20, float32(1000)
			if tc.name == "no-outer-range" {
				count = 1
			}
			if tc.name == "short-inner" {
				contentHeight = 100
			}
			var captured func(ui.InputEvent) bool
			record := func(fn func(ui.InputEvent) bool) func(ui.InputEvent) bool { captured = fn; return fn }
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Children(func() {
					ui.List(c, &list, count, func(i int) {
						if i == 0 {
							ui.Scroll(c).Height(200).TrackScroll(&inner).HandleInput(record(func(ev ui.InputEvent) bool {
								if ev.Kind != ui.InputScroll || !live {
									return false
								}
								d := ev.DY
								if d == 0 {
									d = ev.DX
								}
								if d == 0 || math.IsNaN(float64(d)) || math.IsInf(float64(d), 0) {
									return false
								}
								seen = ev
								calls++
								before := inner.Y
								inner.Y = max(float32(0), min(inner.MaxY, before+d))
								residual := d - (inner.Y - before)
								if residual != 0 {
									outer.Y = max(float32(0), min(outer.MaxY, outer.Y+residual))
									if outer.Y == outer.MaxY {
										list.ScrollToEnd()
									}
								}
								return true
							})).Children(func() { ui.Text(c, "Mechanism controlled content").Height(contentHeight) })
						} else {
							ui.Text(c, "outer").Height(100)
						}
					}).Grow(1).TrackScroll(&outer)
				})
				bounds = inner.Handle.Bounds(c)
			}, 500, 400)
			tt.Frame()
			inner.Y = tc.inner
			outer.Y = tc.outer
			tt.Frame()
			tt.Frame()
			beforeI, beforeO, maxI, maxO := inner.Y, outer.Y, inner.MaxY, outer.MaxY
			d := tc.dy
			if d == 0 {
				d = tc.dx
			}
			wantI := max(float32(0), min(maxI, beforeI+d))
			wantO := max(float32(0), min(maxO, beforeO+d-(wantI-beforeI)))
			if tc.shift {
				tt.HoldModifiers(ui.Shift)
			}
			if tc.wheel {
				tt.Wheel(bounds.X+30, max(float32(30), bounds.Y+30), tc.dx, tc.dy)
			} else {
				tt.Scroll(bounds.X+30, max(float32(30), bounds.Y+30), tc.dx, tc.dy)
			}
			tt.Frame()
			if wantO == maxO {
				if !list.AtEnd() {
					t.Fatal("saturated residual did not reach virtual tail")
				}
				wantO = outer.MaxY
			} // List measurement refines estimates after scrolling; validate its actual end.
			if calls != 1 || math.Abs(float64(inner.Y-wantI)) > .01 || math.Abs(float64(outer.Y-wantO)) > .01 {
				t.Fatalf("calls%d inner%v want%v outer%v want%v maxima%v/%v", calls, inner.Y, wantI, outer.Y, wantO, maxI, maxO)
			}
			if seen.Precise == tc.wheel || seen.DX != tc.dx || seen.DY != tc.dy {
				t.Fatal("public DIP/precise contract changed", seen)
			}
			if tc.shift && seen.Mods&ui.Shift == 0 {
				t.Fatal("shift absent")
			}
			// Stale callback refuses BEFORE any direct tracked-state mutation. Native
			// defaults may still scroll; this check concerns owner-bound callback only.
			live = false
			oldCalls := calls
			oldI, oldO := inner.Y, outer.Y
			if captured(ui.InputEvent{Kind: ui.InputScroll, DY: 10, Precise: true}) || inner.Y != oldI || outer.Y != oldO {
				t.Fatal("stale captured owner modified tracked state")
			}
			tt.Scroll(30, 30, 0, 1)
			if calls != oldCalls {
				t.Fatal("stale owner callback executed")
			}
		})
	}
}
