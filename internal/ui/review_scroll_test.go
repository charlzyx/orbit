package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
)

// Rust sidepane.rs:1413–1483 keeps the virtual list inside a horizontal
// viewport when wrapping is disabled. A clipped single-line label is not
// an equivalent implementation: the tail must be reachable by pointer pan.
func TestMountedReviewUnwrappedLongLineHasHorizontalViewport(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.data.Lines[1].Content = strings.Repeat("wide_", 100) + "TAIL_SENTINEL"
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if r, ok := tt.Find("Review horizontal viewport"); !ok || r.W <= 0 || r.H <= 0 {
		t.Fatal("unwrapped native diff has no horizontal viewport; tail remains clipped", r, ok)
	}
}

func TestMountedReviewPointerPanReachesTailAndKeepsVerticalVirtualization(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(map[bool]string{false: "unified", true: "split"}[split], func(t *testing.T) {
			a, tt := mountedReview(t)
			a.reviewState.data.Lines[1].Content = strings.Repeat("wide_", 100) + "TAIL_SENTINEL"
			if err := tt.Click("Unwrap lines"); err != nil {
				t.Fatal(err)
			}
			if split {
				if err := tt.Click("Side-by-side view"); err != nil {
					t.Fatal(err)
				}
			}
			tt.Frame()
			viewport, ok := tt.Find("Review horizontal viewport")
			if !ok {
				t.Fatal("missing viewport")
			}
			text, ok := tt.Find(a.reviewState.data.Lines[1].Content)
			if !ok {
				t.Fatal("missing first line")
			}
			if a.reviewState.horizontal.MaxX < 1000 {
				t.Fatal("long line not laid out at intrinsic width", a.reviewState.horizontal)
			}
			header, _ := tt.Find("Review file header 0")
			first, last := a.reviewState.list.Visible()
			if last >= len(a.reviewState.data.Lines)-1 {
				t.Fatal("vertical list stopped virtualizing", first, last)
			}
			tt.Scroll(viewport.X+viewport.W/2, viewport.Y+80, 100000, 0)
			moved, ok := tt.Find(a.reviewState.data.Lines[1].Content)
			if !ok {
				t.Fatal("pan unmounted first line")
			}
			if a.reviewState.horizontal.X <= 0 || moved.X >= text.X {
				t.Fatal("real pointer pan did not move text", text, moved, a.reviewState.horizontal)
			}
			pinned, _ := tt.Find("Review file header 0")
			if math.Abs(float64(pinned.X-header.X)) > 1 || math.Abs(float64(pinned.W-viewport.W)) > 1 {
				t.Fatal("file header panned with code", header, pinned, viewport)
			}
			if got, _ := a.reviewState.list.Visible(); got != first {
				t.Fatal("horizontal pan changed vertical anchor", got, first)
			}
			if !split && math.Abs(float64(moved.X+moved.W-(viewport.X+viewport.W-10))) > 2 {
				t.Fatal("tail edge not reachable with source 10px body end padding", moved, viewport)
			}
			tt.Scroll(viewport.X+viewport.W/2, viewport.Y+80, 0, 400)
			if got, _ := a.reviewState.list.Visible(); got <= first {
				t.Fatal("vertical pointer wheel no longer reaches virtual list", got, first)
			}
			if err := tt.Click("Wrap lines"); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			if a.reviewState.horizontal.X != 0 || a.reviewState.horizontal.MaxX > 1 {
				t.Fatal("wrap left horizontal offset/overflow", a.reviewState.horizontal)
			}
		})
	}
}

func TestMountedReviewSnapshotSwitchResetsHorizontalPan(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.data.Lines[1].Content = strings.Repeat("W", 500)
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	viewport, _ := tt.Find("Review horizontal viewport")
	tt.Scroll(viewport.X+30, viewport.Y+90, 300, 0)
	if a.reviewState.horizontal.X <= 0 {
		t.Fatal("fixture did not pan")
	}
	a.reviewState.clearReviewSnapshot()
	if a.reviewState.horizontal.X != 0 {
		t.Fatal("new workspace/source inherited pan")
	}
	// ListState resets with the same snapshot boundary, not on ordinary pan.
	a.reviewState.list.ScrollTo(0, ui.Start)
}

func TestMountedReviewHiddenContextWidthAndFontChanges(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	hidden := workspace.ReviewLine{File: 0, Kind: "context", NewLine: 90, Content: strings.Repeat("界", 180) + "hidden_tail"}
	s.data.Lines[2] = workspace.ReviewLine{File: 0, Kind: "gap", Gap: &workspace.ReviewGap{ID: 987, Count: 1, Hidden: []workspace.ReviewLine{hidden}, Position: "between"}}
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	before := s.horizontal.MaxX
	if before < 1000 {
		t.Fatal("hidden long context omitted from width", before)
	}
	s.data.ExpandGap(2, "all")
	tt.Frame()
	if math.Abs(float64(s.horizontal.MaxX-before)) > 1 {
		t.Fatal("context reveal changed scroll extent", before, s.horizontal.MaxX)
	}
	a.Prefs.EditorFontSize *= 1.5
	tt.Frame()
	if s.horizontal.MaxX <= before {
		t.Fatal("font change kept stale width", before, s.horizontal.MaxX)
	}
}

func TestMountedReviewFileNavigationKeepsPanAndHeaderActionsReachable(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.data.Lines[1].Content = strings.Repeat("W", 500)
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	viewport, _ := tt.Find("Review horizontal viewport")
	tt.Scroll(viewport.X+30, viewport.Y+90, 300, 0)
	pan := a.reviewState.horizontal.X
	if err := tt.Click("Review file two.go"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.reviewState.selected != 1 || a.reviewState.horizontal.X != pan {
		t.Fatal("file navigation lost selection/pan")
	}
	header, ok := tt.Find("Review file header 1")
	if !ok || math.Abs(float64(header.X-viewport.X)) > 1 || math.Abs(float64(header.W-viewport.W)) > 1 {
		t.Fatal("new file header not pinned to viewport", header, viewport)
	}
}

func TestMountedReviewVerticalScrollbarDragAtVisibleRightEdge(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.data.Lines[1].Content = strings.Repeat("W", 500)
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	viewport, _ := tt.Find("Review horizontal viewport")
	for _, pan := range []float32{0, 300, 100000} {
		a.reviewState.list.ScrollTo(0, ui.Start)
		tt.Frame()
		if pan > 0 {
			tt.Scroll(viewport.X+100, viewport.Y+90, pan, 0)
		}
		first, _ := a.reviewState.list.Visible()
		x, y := viewport.X+viewport.W-5, viewport.Y+80
		tt.Move(x, y)
		tt.Frame() // native overlay bars only hit-test once painted
		tt.Press(x, y)
		tt.Move(x, y+230)
		tt.Release(x, y+230)
		tt.Frame()
		if got, _ := a.reviewState.list.Visible(); got <= first+3 {
			t.Fatal("vertical scrollbar unreachable at visible edge", pan, first, got, viewport)
		}
	}
}

func TestMountedReviewDiffTextKeepsReadableAccessibleName(t *testing.T) {
	a, tt := mountedReview(t)
	text := a.reviewState.data.Lines[1].Content
	if !tt.HasText(text) {
		t.Fatal("diff text was replaced by an internal geometry label", tt.Texts())
	}
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText(text) {
		t.Fatal("unwrapping hid accessible diff content")
	}
}
