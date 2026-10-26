package orbitui

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func headerInteractionFixture(t *testing.T, sticky bool) (*App, *ui.Tester, ui.Rect) {
	t.Helper()
	a := fixture(t)
	a.Model.Workspace = t.TempDir()
	a.ReviewVisible = true
	s := &a.reviewState
	s.tree, s.root = true, a.Model.Workspace
	for file, path := range []string{"alpha.go", "bravo.go", "charlie.go"} {
		s.data.Files = append(s.data.Files, workspace.ReviewFile{Path: path, DiffLine: len(s.data.Lines), Additions: 17, Deletions: 9})
		s.data.Lines = append(s.data.Lines, workspace.ReviewLine{File: file, Kind: "file_header", Content: path})
		for line := 1; line <= 65; line++ {
			kind := "context"
			if line == 8 {
				kind = "addition"
			}
			s.data.Lines = append(s.data.Lines, workspace.ReviewLine{File: file, Kind: kind, NewLine: line, Content: fmt.Sprintf("file%d line%d", file, line)})
		}
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Review file alpha.go"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyBracketRight)
	if s.selected != 0 || s.diffCursor == nil {
		t.Fatal("initial real tree selection/hunk", s.selected, s.diffCursor)
	}
	viewport, ok := tt.Find("Review horizontal viewport")
	if !ok {
		t.Fatal("viewport absent")
	}
	label := "Review file header 1"
	if sticky {
		label = "Review sticky file header 1"
	}
	for i := 0; i < 180; i++ {
		r, ok := tt.Find(label)
		if ok && r.W > 0 && r.H > 0 && ((!sticky && r.Y > viewport.Y+50 && r.Y+r.H < viewport.Y+viewport.H-20) || (sticky && math.Abs(float64(r.Y-viewport.Y)) < 1)) {
			if s.selected != 0 {
				t.Fatal("wheel changed selection", s.selected)
			}
			t.Logf("header=%+v viewport=%+v selected=%d cursor=%d", r, viewport, s.selected, *s.diffCursor)
			return a, tt, r
		}
		tt.Scroll(viewport.X+viewport.W/2, viewport.Y+100, 0, 25)
	}
	t.Fatal("real wheel did not expose requested B header", label)
	return nil, nil, ui.Rect{}
}

func TestAuditReviewWholeHeaderToggle(t *testing.T) {
	for _, sticky := range []bool{false, true} {
		for _, region := range []string{"path", "counts"} {
			t.Run(fmt.Sprintf("sticky=%t/%s", sticky, region), func(t *testing.T) {
				a, tt, h := headerInteractionFixture(t, sticky)
				x := h.X + 110
				if region == "counts" {
					x = h.X + h.W - 75
				}
				tt.ClickAt(x, h.Y+h.H/2)
				if !a.reviewState.collapsed[1] {
					t.Fatalf("whole %s header did not toggle B; click=(%g,%g), header=%+v", region, x, h.Y+h.H/2, h)
				}
				if a.reviewState.selected != 0 {
					t.Fatal("header toggle changed selected A", a.reviewState.selected)
				}
			})
		}
	}
}
func TestAuditReviewChevronPreservesNavigationOrigin(t *testing.T) {
	for _, sticky := range []bool{false, true} {
		t.Run(fmt.Sprintf("sticky=%t", sticky), func(t *testing.T) {
			a, tt, h := headerInteractionFixture(t, sticky)
			before := *a.reviewState.diffCursor
			tt.ClickAt(h.X+24, h.Y+h.H/2)
			if !a.reviewState.collapsed[1] {
				t.Fatal("arrow did not toggle B")
			}
			if a.reviewState.selected != 0 {
				t.Errorf("toggle changed selected A to %d", a.reviewState.selected)
			}
			if a.reviewState.diffCursor == nil || *a.reviewState.diffCursor != before {
				t.Error("toggle changed hunk cursor")
			}
			a.reviewState.treeHandle.Focus()
			tt.Frame()
			tt.Key(0, ui.KeyN)
			if a.reviewState.selected != 1 {
				t.Errorf("n from original A should select B; selected=%d", a.reviewState.selected)
			}
		})
	}
}
func TestAuditReviewOpenChildDoesNotToggleHeader(t *testing.T) {
	for _, sticky := range []bool{false, true} {
		t.Run(fmt.Sprintf("sticky=%t", sticky), func(t *testing.T) {
			a, tt, h := headerInteractionFixture(t, sticky)
			path := filepath.Join(a.Model.Workspace, "bravo.go")
			if err := os.WriteFile(path, []byte("fixture only"), 0600); err != nil {
				t.Fatal(err)
			}
			tt.ClickAt(h.X+h.W-24, h.Y+h.H/2)
			if a.reviewState.collapsed[1] {
				t.Fatal("Open child also toggled parent")
			}
			if a.page != "files" || a.files == nil || a.files.desired != path {
				t.Fatal("Open child did not target B", a.page)
			}
		})
	}
}

func TestReviewHeaderArrowKeyboardKeepsFocusAndSelection(t *testing.T) {
	for _, sticky := range []bool{false, true} {
		t.Run(fmt.Sprintf("sticky=%t", sticky), func(t *testing.T) {
			a := fixture(t)
			a.reviewState.data.Files = []workspace.ReviewFile{{Path: "alpha.go"}, {Path: "bravo.go"}}
			a.reviewState.selected = 0
			cursor := 3
			a.reviewState.diffCursor = &cursor
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(a.Theme.Native())
				ui.Column(c).Width(360).Children(func() {
					a.reviewFileHeader(c, workspace.ReviewLine{File: 1, Kind: "file_header", Content: "bravo.go"}, sticky)
				})
			}, 360, 80)
			tt.Frame()
			tt.Key(0, ui.KeyTab)
			if !tt.Focused("Toggle file diff") {
				t.Fatal("arrow not keyboard focusable")
			}
			tt.Key(0, ui.KeySpace)
			if !a.reviewState.collapsed[1] || !tt.Focused("Toggle file diff") {
				t.Fatal("Space did not toggle exactly once/retain focus")
			}
			tt.Key(0, ui.KeyEnter)
			if a.reviewState.collapsed[1] || !tt.Focused("Toggle file diff") {
				t.Fatal("Enter did not expand exactly once/retain focus")
			}
			if a.reviewState.selected != 0 || *a.reviewState.diffCursor != 3 {
				t.Fatal("keyboard toggle changed navigation origin")
			}
		})
	}
}
func TestReviewHeaderBodyHoverAndCollapseRoundTripAnchor(t *testing.T) {
	a, tt, h := headerInteractionFixture(t, false)
	first, _ := a.reviewState.list.Visible()
	visible, _ := a.reviewState.visible()
	before := reviewLineKey(visible[first])
	y := a.reviewState.vertical.Y
	tt.Move(h.X+100, h.Y+h.H/2)
	tt.Frame()
	if tt.Cursor() != ui.CursorPointer {
		t.Fatal("header body lacks source pointer cursor", tt.Cursor())
	}
	hovered := tt.Image().RGBAAt(int(h.X+2), int(h.Y+h.H/2))
	tt.Move(10, 10)
	tt.Frame()
	idle := tt.Image().RGBAAt(int(h.X+2), int(h.Y+h.H/2))
	if idle == hovered {
		t.Fatal("header body lacks source hover paint", idle)
	}
	tt.ClickAt(h.X+100, h.Y+h.H/2)
	if !a.reviewState.collapsed[1] {
		t.Fatal("body collapse")
	}
	first, _ = a.reviewState.list.Visible()
	visible, _ = a.reviewState.visible()
	if reviewLineKey(visible[first]) != before || math.Abs(float64(a.reviewState.vertical.Y-y)) > 0.1 {
		t.Fatal("collapse below reader moved retained anchor", before, reviewLineKey(visible[first]), y, a.reviewState.vertical.Y)
	}
	collapsed, ok := tt.Find("Review file header 1")
	if !ok {
		t.Fatal("collapsed B disappeared")
	}
	tt.ClickAt(collapsed.X+100, collapsed.Y+collapsed.H/2)
	if a.reviewState.collapsed[1] {
		t.Fatal("body expand")
	}
	after, ok := tt.Find("Review file header 1")
	if !ok || math.Abs(float64(after.Y-h.Y)) > 0.1 {
		t.Fatal("roundtrip moved B header", h, after)
	}
	first, _ = a.reviewState.list.Visible()
	visible, _ = a.reviewState.visible()
	if reviewLineKey(visible[first]) != before {
		t.Fatal("expand changed semantic reader anchor")
	}
}
