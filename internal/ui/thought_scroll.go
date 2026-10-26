package orbitui

import "math"

func (a *App) thoughtScrollChain(owner inlineImageToken) func(float32) bool {
	return func(delta float32) bool {
		if !a.ownsInlineImage(owner) || math.IsNaN(float64(delta)) || math.IsInf(float64(delta), 0) {
			return false
		}
		if delta == 0 {
			return true
		}
		s := &a.transcriptScroll
		if math.IsNaN(float64(s.Y)) || math.IsInf(float64(s.Y), 0) || math.IsNaN(float64(s.MaxY)) || math.IsInf(float64(s.MaxY), 0) || s.MaxY < 0 {
			return false
		}
		target := max(float32(0), min(s.MaxY, s.Y+delta))
		if target == s.MaxY {
			// Use the virtual List sentinel: MaxY can still be an estimated extent.
			a.transcriptActivity.unread = false
			a.transcript.ScrollToEnd()
		} else {
			s.Y = target
		}
		return true
	}
}
