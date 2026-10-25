package gitworkbench

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"testing"
)

func embeddedGapFixture(t *testing.T, staged, split bool, position string, count int) (*State, *ui.Tester) {
	s, tt := embeddedReviewFixture(t, split)
	hidden := make([]workspace.ReviewLine, count)
	for i := range hidden {
		hidden[i] = workspace.ReviewLine{File: 0, Kind: "context", OldLine: i + 1, NewLine: i + 1, Content: fmt.Sprintf("retained %d", i+1)}
	}
	snapshot := s.diffSnapshot(staged)
	snapshot.Lines = []workspace.ReviewLine{{File: 0, Kind: "file_header", Content: "long.go"}, {File: 0, Kind: "gap", Gap: &workspace.ReviewGap{ID: 77, Position: position, Count: count, Hidden: hidden}}, {File: 0, Kind: "addition", NewLine: count + 1, Content: "changed"}}
	// Only this source remains expanded so both staged states mount near the top.
	s.diffExpanded[fmt.Sprintf("%t:long.go", !staged)] = false
	s.list = ui.ListState{}
	tt.Frame()
	return s, tt
}
func TestEmbeddedGapDirectionalSourceControls(t *testing.T) {
	for _, staged := range []bool{false, true} {
		for _, split := range []bool{false, true} {
			for _, tc := range []struct {
				position, direction string
				first               int
			}{{"leading", "end", 151}, {"trailing", "start", 1}, {"between", "start", 1}, {"between", "end", 151}} {
				t.Run(fmt.Sprintf("%t/%t/%s/%s", staged, split, tc.position, tc.direction), func(t *testing.T) {
					s, tt := embeddedGapFixture(t, staged, split, tc.position, 250)
					if err := tt.Click("Expand context " + tc.direction); err != nil {
						t.Fatal(err)
					}
					var lines []int
					for _, line := range s.diffSnapshot(staged).Lines {
						if line.Kind == "context" {
							lines = append(lines, line.NewLine)
						}
					}
					if len(lines) != workspace.DefaultExpansionLines || lines[0] != tc.first {
						t.Fatal("wrong source chunk", lines)
					}
				})
			}
		}
	}
}
func TestEmbeddedGapLabelShiftAndUnavailable(t *testing.T) {
	for _, count := range []int{70, 250} {
		for _, shift := range []bool{false, true} {
			t.Run(fmt.Sprintf("%d/%t", count, shift), func(t *testing.T) {
				s, tt := embeddedGapFixture(t, false, true, "between", count)
				if count == 70 {
					if _, ok := tt.Find("Expand context both"); !ok {
						t.Fatal("small gap missing both target")
					}
					if _, ok := tt.Find("Expand context start"); ok {
						t.Fatal("small gap exposes split controls")
					}
				}
				mods := ui.Modifiers(0)
				if shift {
					mods = ui.Shift
				}
				if err := tt.ClickWith(mods, fmt.Sprintf("Show %d unchanged lines", count)); err != nil {
					t.Fatal(err)
				}
				got := 0
				for _, line := range s.Data.UnstagedReview.Lines {
					if line.Kind == "context" {
						got++
					}
				}
				want := min(count, workspace.DefaultExpansionLines*2)
				if shift {
					want = count
				}
				if got != want {
					t.Fatal("label expansion", got, want)
				}
			})
		}
	}
	s, tt := embeddedGapFixture(t, true, false, "leading", 250)
	s.Data.StagedReview.Lines[1].Gap.Hidden = nil
	tt.Frame()
	if _, ok := tt.Find("Expand context end"); ok {
		t.Fatal("unavailable retained context offers action")
	}
	if err := tt.Click("Show 250 unchanged lines"); err != nil {
		t.Fatal(err)
	}
	if len(s.Data.StagedReview.Lines) != 3 {
		t.Fatal("unavailable context mutated snapshot")
	}
}

func TestEmbeddedGapGutterGeometryRepeatedAndShift(t *testing.T) {
	for _, staged := range []bool{false, true} {
		for _, split := range []bool{false, true} {
			t.Run(fmt.Sprintf("%t/%t", staged, split), func(t *testing.T) {
				s, tt := embeddedGapFixture(t, staged, split, "leading", 250)
				row, _ := tt.Find(fmt.Sprintf("Git context gap %t:77", staged))
				button, _ := tt.Find("Expand context end")
				gutter, _ := tt.Find("Git context gutter")
				if row.H != 32 || gutter.W != 46 || button.W != 45 || button.H != 32 {
					t.Fatal("source gutter geometry", row, button)
				}
				for _, remaining := range []int{150, 50} {
					if err := tt.Click("Expand context end"); err != nil {
						t.Fatal(err)
					}
					got := 0
					for _, line := range s.diffSnapshot(staged).Lines {
						if line.Gap != nil {
							got += line.Gap.Count
						}
					}
					if got != remaining {
						t.Fatal("repeated expansion used wrong gap", got, remaining)
					}
				}
				if err := tt.ClickWith(ui.Shift, "Expand context end"); err != nil {
					t.Fatal(err)
				}
				if len(s.diffSnapshot(staged).Lines) != 252 {
					t.Fatal("Shift gutter did not reveal all retained context", len(s.diffSnapshot(staged).Lines))
				}
			})
		}
	}
	_, tt := embeddedGapFixture(t, false, false, "between", 250)
	top, _ := tt.Find("Expand context start")
	bottom, _ := tt.Find("Expand context end")
	if top.W != 45 || top.H != 16 || bottom.H != 16 || top.Y+top.H != bottom.Y {
		t.Fatal("split directional gutter geometry", top, bottom)
	}
}

func TestEmbeddedGapHeldPressCannotExpandReplacement(t *testing.T) {
	s, tt := embeddedGapFixture(t, false, false, "leading", 250)
	b, _ := tt.Find("Expand context end")
	x, y := b.X+b.W/2, b.Y+b.H/2
	tt.Press(x, y)
	s.Data.UnstagedReview.Lines = append([]workspace.ReviewLine(nil), s.Data.UnstagedReview.Lines...)
	tt.Frame()
	tt.Release(x, y)
	tt.Frame()
	if len(s.Data.UnstagedReview.Lines) != 3 {
		t.Fatal("held press expanded a replacement snapshot")
	}
	if err := tt.Click("Expand context end"); err != nil {
		t.Fatal(err)
	}
	if len(s.Data.UnstagedReview.Lines) <= 3 {
		t.Fatal("fresh replacement control did not work")
	}
}
