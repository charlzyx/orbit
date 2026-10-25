package orbitui

import (
	"github.com/egoist/mygo/ui"
	"math"
	"testing"
)

// Source settings.rs::typography_preview stretches both equal-width columns;
// only the terminal content sets the shared panel height.
func TestSourceTypographyPreviewColumnsStretchToSharedHeight(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "appearance"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	left, ok := tt.Find("Interface typography preview")
	if !ok {
		t.Fatal("interface preview missing")
	}
	right, ok := tt.Find("Terminal typography preview")
	if !ok {
		t.Fatal("terminal preview missing")
	}
	if math.Abs(float64(left.Y-right.Y)) > .1 || math.Abs(float64(left.H-right.H)) > .1 || math.Abs(float64(left.W-right.W)) > .1 {
		t.Fatal("source equal preview geometry lost", left, right)
	}
}
