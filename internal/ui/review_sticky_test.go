package orbitui

import (
	"math"
	"strings"
	"testing"
)

// Docked Review is ungrouped (one working-tree snapshot). Source
// sidepane.rs:1485–1550 pins its current file header and lets the next push it.
func TestMountedReviewStickyFileHeaderAfterVerticalAndHorizontalWheel(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(map[bool]string{false: "unified", true: "split"}[split], func(t *testing.T) {
			a, tt := mountedReview(t)
			a.reviewState.data.Lines[1].Content = strings.Repeat("W", 500)
			if err := tt.Click("Unwrap lines"); err != nil {
				t.Fatal(err)
			}
			if split {
				if err := tt.Click("Side-by-side view"); err != nil {
					t.Fatal(err)
				}
			}
			tt.Frame()
			viewport, _ := tt.Find("Review horizontal viewport")
			tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 450)
			first, _ := a.reviewState.list.Visible()
			if first < 5 {
				t.Fatal("fixture did not vertically scroll", first)
			}
			header, ok := tt.Find("Review sticky file header 0")
			if !ok || math.Abs(float64(header.Y-viewport.Y)) > 1 {
				t.Fatal("file header disappeared instead of sticking", header, viewport, ok)
			}
			tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 300, 0)
			header, ok = tt.Find("Review sticky file header 0")
			if !ok || math.Abs(float64(header.X-viewport.X)) > 1 || math.Abs(float64(header.Y-viewport.Y)) > 1 {
				t.Fatal("sticky header panned away", header, viewport)
			}
			// The actual pinned toggle remains hit-testable after both axes move.
			tt.ClickAt(header.X+24, header.Y+17)
			if !a.reviewState.collapsed[0] {
				t.Fatal("pinned header action did not collapse its file")
			}
			if err := tt.Click("Review file two.go"); err != nil {
				t.Fatal(err)
			}
			tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 150)
			header, ok = tt.Find("Review sticky file header 1")
			if !ok || math.Abs(float64(header.Y-viewport.Y)) > 1 {
				t.Fatal("next file did not become sticky", header, viewport, ok)
			}
		})
	}
}

func TestMountedReviewNextHeaderPushesPinnedHeader(t *testing.T) {
	a, tt := mountedReview(t)
	viewport, _ := tt.Find("Review horizontal viewport")
	sawPush := false
	for i := 0; i < 100; i++ {
		tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 20)
		next, exists := tt.Find("Review file header 1")
		old, oldExists := tt.Find("Review sticky file header 0")
		// Find returns clipped visible bounds: the outgoing header height
		// shrinks during the push. Compare against its original 34px height.
		if exists && oldExists && next.Y > viewport.Y && next.Y < viewport.Y+34 {
			if math.Abs(float64(old.Y+old.H-next.Y)) > 1 {
				t.Fatal("incoming header overlaps instead of pushing", old, next)
			}
			sawPush = true
			break
		}
	}
	if !sawPush {
		t.Fatal("never observed next-file header pushing pinned header")
	}
	if a.reviewState.selected != 0 {
		t.Fatal("scrolling changed selected-file navigation state")
	}
}

func TestMountedReviewStickyPressCannotActivateDifferentFileAfterScroll(t *testing.T) {
	a, tt := mountedReview(t)
	viewport, _ := tt.Find("Review horizontal viewport")
	tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 300)
	old, ok := tt.Find("Review sticky file header 0")
	if !ok {
		t.Fatal("fixture missing sticky first file")
	}
	x, y := old.X+24, old.Y+17
	tt.Press(x, y)
	tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 1600)
	if _, ok := tt.Find("Review sticky file header 1"); !ok {
		t.Fatal("fixture did not switch sticky file")
	}
	tt.Release(x, y)
	tt.Frame()
	if a.reviewState.collapsed[1] {
		t.Fatal("press on first file collapsed a different file after scrolling")
	}
}

func TestMountedReviewWheelOverStickyHeaderStillScrollsDiff(t *testing.T) {
	a, tt := mountedReview(t)
	viewport, _ := tt.Find("Review horizontal viewport")
	tt.Scroll(viewport.X+100, viewport.Y+100, 0, 400)
	header, ok := tt.Find("Review sticky file header 0")
	if !ok {
		t.Fatal("fixture missing sticky header")
	}
	first, _ := a.reviewState.list.Visible()
	tt.Scroll(header.X+header.W/2, header.Y+header.H/2, 0, 120)
	if got, _ := a.reviewState.list.Visible(); got <= first {
		t.Fatal("sticky header swallowed native wheel", first, got)
	}
}

func TestMountedReviewStickyHeaderWheelAxesAndResize(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(map[bool]string{false: "unified", true: "split"}[split], func(t *testing.T) {
			a, tt := mountedReview(t)
			a.reviewState.data.Lines[1].Content = strings.Repeat("W", 500)
			if err := tt.Click("Unwrap lines"); err != nil {
				t.Fatal(err)
			}
			if split {
				if err := tt.Click("Side-by-side view"); err != nil {
					t.Fatal(err)
				}
			}
			tt.Frame()
			viewport, _ := tt.Find("Review horizontal viewport")
			tt.Scroll(viewport.X+100, viewport.Y+100, 0, 400)
			header, ok := tt.Find("Review sticky file header 0")
			if !ok {
				t.Fatal("missing sticky header")
			}
			first, _ := a.reviewState.list.Visible()
			tt.Scroll(header.X+header.W/2, header.Y+17, 100000, 0)
			if a.reviewState.horizontal.X < 1000 {
				t.Fatal("header swallowed horizontal wheel", a.reviewState.horizontal)
			}
			if got, _ := a.reviewState.list.Visible(); got != first {
				t.Fatal("header horizontal wheel changed vertical anchor", first, got)
			}
			tt.SetSize(950, 820)
			tt.Frame()
			viewport, _ = tt.Find("Review horizontal viewport")
			header, ok = tt.Find("Review sticky file header 0")
			if !ok || math.Abs(float64(header.X-viewport.X)) > 1 || math.Abs(float64(header.W-viewport.W)) > 1 {
				t.Fatal("resized/panned sticky bounds", header, viewport)
			}
			if a.reviewState.horizontal.X > a.reviewState.horizontal.MaxX+1 {
				t.Fatal("resize kept out-of-range pan")
			}
			tt.Scroll(header.X+header.W/2, header.Y+17, -200, 0)
			before := a.reviewState.horizontal.X
			if before <= 0 {
				t.Fatal("fixture lost horizontal pan")
			}
			first, _ = a.reviewState.list.Visible()
			tt.Scroll(header.X+header.W/2, header.Y+17, 0, 120)
			if got, _ := a.reviewState.list.Visible(); got <= first {
				t.Fatal("resized header swallowed vertical wheel", first, got)
			}
			if a.reviewState.horizontal.X != before {
				t.Fatal("vertical wheel moved horizontal pan")
			}
		})
	}
}
