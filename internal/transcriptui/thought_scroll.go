package transcriptui

import (
	"github.com/egoist/mygo/ui"
	"math"
)

// Only an explicitly opted-in Thought consumes wheel input. Other controls and
// event kinds retain MyGo's routing. Deltas already arrive in local DIPs; source
// Thought chaining is immediate for both precise and notch input.
func thoughtScrollInput(scroll *ui.ScrollState, chain func(float32) bool) func(ui.InputEvent) bool {
	return func(ev ui.InputEvent) bool {
		if chain == nil || ev.Kind != ui.InputScroll {
			return false
		}
		delta := ev.DY
		if delta == 0 {
			delta = ev.DX
		}
		finite := func(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
		if delta == 0 || !finite(delta) || !finite(scroll.Y) || !finite(scroll.MaxY) || scroll.MaxY < 0 {
			return false
		}
		before := max(float32(0), min(scroll.MaxY, scroll.Y))
		target := max(float32(0), min(scroll.MaxY, before+delta))
		residual := delta - (target - before)
		// Even residual0 validates the captured host owner before touching inner state.
		if !chain(residual) {
			// This committed Thought already owned the event. A rejected stale
			// owner must not fall through to native scrolling of the old hit tree.
			return true
		}
		scroll.Y = target
		return true
	}
}
