package gitworkbench

import (
	"context"
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
	"time"
)

func embeddedReviewFixture(t *testing.T, split bool) (*State, *ui.Tester) {
	return embeddedReviewFixtureTheme(t, split, nil)
}
func embeddedReviewFixtureTheme(t *testing.T, split bool, theme *ui.Theme) (*State, *ui.Tester) {
	t.Helper()
	s := New(t.TempDir(), Options{})
	t.Cleanup(s.Close)
	// Static synthetic geometry must not start the production 4s Git refresh.
	// Real refresh/dataset replacement is covered separately below.
	s.initialized, s.refreshed = true, time.Now().Add(time.Hour)
	s.DiffSplit = split
	s.Data.Branch = "fixture"
	s.Data.Changes = []workspace.StatusRow{{Path: "long.go", Index: 'M', Worktree: 'M'}}
	for _, staged := range []bool{true, false} {
		snapshot := s.diffSnapshot(staged)
		snapshot.Files = []workspace.ReviewFile{{Path: "long.go", DiffLine: 0}}
		snapshot.Lines = []workspace.ReviewLine{{File: 0, Kind: "file_header", Content: "long.go"}}
		for i := 1; i <= 150; i++ {
			text := fmt.Sprintf("context %d", i)
			if i == 1 {
				text = strings.Repeat("long_", 100) + "TAIL_SENTINEL"
			}
			snapshot.Lines = append(snapshot.Lines, workspace.ReviewLine{File: 0, Kind: "context", OldLine: i, NewLine: i, Content: text})
		}
		s.diffExpanded[fmt.Sprintf("%t:long.go", staged)] = true
	}
	tt := ui.NewTester(func(c *ui.Context) {
		if theme != nil {
			c.SetTheme(theme)
		}
		View(c, s)
	}, 1222, 878)
	tt.Frame()
	return s, tt
}

func TestEmbeddedReviewUnwrappedLongLineHasHorizontalViewport(t *testing.T) {
	_, tt := embeddedReviewFixture(t, false)
	if r, ok := tt.Find("Git diff horizontal viewport"); !ok || r.W <= 0 || r.H <= 0 {
		t.Fatal("expanded Git diff still clips long-line tail", r, ok)
	}
}

func TestEmbeddedReviewPointerPanCopyAndVerticalDrag(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(fmt.Sprint(split), func(t *testing.T) {
			s, tt := embeddedReviewFixture(t, split)
			viewport, _ := tt.Find("Git diff horizontal viewport")
			if s.diffView.horizontal.MaxX < 1000 {
				t.Fatal("intrinsic width not exposed", s.diffView.horizontal)
			}
			first, last := s.list.Visible()
			if last >= len(s.changeItems())-1 {
				t.Fatal("lost vertical virtualization")
			}
			label := s.Data.StagedReview.Lines[1].Content
			before, ok := tt.Find(label)
			if !ok {
				t.Fatal("long context not mounted")
			}
			tt.Scroll(viewport.X+400, viewport.Y+100, 100000, 0)
			after, ok := tt.Find(label)
			if !split && (!ok || after.X >= before.X || after.X+after.W > viewport.X+viewport.W) {
				t.Fatal("pan cannot reach visible tail", before, after, viewport)
			}
			if got, _ := s.list.Visible(); got != first {
				t.Fatal("pan changed vertical anchor")
			}
			if !split {
				tt.ClickAt(after.X+after.W-30, after.Y+after.H/2)
				tt.Key(ui.Ctrl, ui.KeyA)
				tt.Key(ui.Ctrl, ui.KeyC)
				if !strings.Contains(tt.Clipboard(), "TAIL_SENTINEL") {
					t.Fatal("native selectable rich text lost hidden tail on copy", tt.Clipboard())
				}
			}
			for _, pan := range []float32{0, 300, 100000} {
				s.list.ScrollTo(0, ui.Start)
				s.diffView.horizontal.X = 0
				tt.Frame()
				if pan > 0 {
					tt.Scroll(viewport.X+100, viewport.Y+100, pan, 0)
				}
				first, _ = s.list.Visible()
				x, y := viewport.X+viewport.W-5, viewport.Y+80
				tt.Move(x, y)
				tt.Frame()
				tt.Press(x, y)
				tt.Move(x, y+150)
				tt.Release(x, y+150)
				tt.Frame()
				if got, _ := s.list.Visible(); got <= first+3 {
					t.Fatal("visible edge scrollbar not draggable", pan, first, got)
				}
			}
			// Public checkbox action, not a private-state assertion-only simulation.
			if err := tt.Click("Wrap"); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			if s.diffView.horizontal.MaxX != 0 || s.diffView.horizontal.X != 0 {
				t.Fatal("wrap retained hidden horizontal scroll", s.diffView.horizontal)
			}
		})
	}
}

func TestEmbeddedReviewMetricsRefreshLayoutSwitchAndResize(t *testing.T) {
	theme := *ui.LightTheme()
	s, tt := embeddedReviewFixtureTheme(t, false, &theme)
	before := s.diffView.horizontal.MaxX
	s.Data.StagedReview.Lines = append([]workspace.ReviewLine(nil), s.Data.StagedReview.Lines...)
	s.Data.StagedReview.Lines[1].Content = strings.Repeat("W", 1000) + "LONGER_TAIL"
	tt.Frame()
	if s.diffView.horizontal.MaxX <= before {
		t.Fatal("same-count new dataset kept stale width")
	}
	before = s.diffView.horizontal.MaxX
	oldMetrics := s.diffView.width.metrics
	theme.FontSize = 22
	theme.Spacing = 5
	theme.Font = "serif"
	tt.Frame()
	if s.diffView.width.metrics == oldMetrics {
		t.Fatal("font metrics cache not invalidated")
	}
	// Embedded code remains explicitly monospace12: do not pretend UI font
	// settings change its font. The actual renderer and measured extent agree.
	viewport, _ := tt.Find("Git diff horizontal viewport")
	tt.Scroll(viewport.X+300, viewport.Y+150, 100000, 0)
	if err := tt.Click("Split"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if s.diffView.horizontal.MaxX <= before {
		t.Fatal("split did not expand both halves")
	}
	if err := tt.Click("Split"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Scroll(viewport.X+300, viewport.Y+150, 100000, 0)
	tt.SetSize(950, 820)
	tt.Frame()
	viewport, _ = tt.Find("Git diff horizontal viewport")
	// Resizing must keep the scrollbar on the new visible edge, not the old
	// width cached before the horizontal viewport shrank.
	s.list.ScrollTo(0, ui.Start)
	tt.Frame()
	x, y := viewport.X+viewport.W-5, viewport.Y+100
	tt.Move(x, y)
	tt.Frame()
	tt.Press(x, y)
	tt.Move(x, y+130)
	tt.Release(x, y+130)
	tt.Frame()
	if first, _ := s.list.Visible(); first < 3 {
		t.Fatal("resize left scrollbar outside viewport")
	}
	section, ok := tt.Find("Git sticky section true")
	if !ok {
		t.Fatal("resized section missing")
	}
	file, ok := tt.Find("Git sticky file true:long.go")
	if !ok {
		t.Fatal("resized file missing")
	}
	if math.Abs(float64(file.Y-section.Y-section.H)) > 1 {
		t.Fatal("scaled header metrics overlap", section, file)
	}
}

func TestEmbeddedReviewHeaderBindingsPrunedOnReplacement(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	if len(s.diffView.headers) == 0 {
		t.Fatal("fixture did not bind headers")
	}
	s.Data.Changes = []workspace.StatusRow{{Path: "new.go", Index: 'M', Worktree: ' '}}
	s.Data.StagedReview = workspace.ReviewSnapshot{Files: []workspace.ReviewFile{{Path: "new.go"}}, Lines: []workspace.ReviewLine{{File: 0, Kind: "file_header", Content: "new.go"}, {File: 0, Kind: "addition", NewLine: 1, Content: "replacement"}}}
	s.Data.UnstagedReview = workspace.ReviewSnapshot{}
	tt.Frame()
	for key := range s.diffView.headers {
		if key.path == "long.go" {
			t.Fatal("old dataset header binding retained", key)
		}
	}
}

func TestEmbeddedReviewRealRefreshReplacesWidthsAndHeaders(t *testing.T) {
	g := testRepo(t)
	writeTest(t, g, "old.go", "base\n")
	for _, args := range [][]string{{"add", "old.go"}, {"commit", "-m", "fixture"}} {
		if _, err := g.Run(context.Background(), args...); err != nil {
			t.Fatal(err)
		}
	}
	writeTest(t, g, "old.go", strings.Repeat("wide_", 100)+"\n")
	s := New(g.Dir, Options{})
	defer s.Close()
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	s.diffExpanded["false:old.go"] = true
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1222, 878)
	tt.Frame()
	oldWidth := s.diffView.horizontal.MaxX
	if oldWidth < 1000 {
		t.Fatal("real diff fixture lacks long-line extent", oldWidth)
	}
	oldKey := diffHeaderKey{path: "old.go"}
	if s.diffView.headers[oldKey] == nil {
		t.Fatal("real old header not bound")
	}
	writeTest(t, g, "old.go", "base\n")
	writeTest(t, g, "new.go", strings.Repeat("wider_", 200)+"\n")
	if err := g.Stage(context.Background(), "new.go"); err != nil {
		t.Fatal(err)
	}
	s.Refresh()
	awaitIdle(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	s.diffExpanded["true:new.go"] = true
	tt.Frame()
	if s.diffView.horizontal.MaxX <= oldWidth {
		t.Fatal("real refresh left stale width", oldWidth, s.diffView.horizontal)
	}
	if s.diffView.headers[oldKey] != nil {
		t.Fatal("real refresh retained removed path header")
	}
	if s.diffView.headers[diffHeaderKey{staged: true, path: "new.go"}] == nil {
		t.Fatal("real staged replacement header missing")
	}
}
