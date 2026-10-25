package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"testing"
)

func mountedReviewGap(t *testing.T, position string, count int) (*App, *ui.Tester) {
	t.Helper()
	a, tt := mountedReview(t)
	hidden := make([]workspace.ReviewLine, count)
	for i := range hidden {
		hidden[i] = workspace.ReviewLine{File: 0, Kind: "context", OldLine: i + 1, NewLine: i + 1, Content: fmt.Sprintf("retained context %d", i+1)}
	}
	a.reviewState.data = workspace.ReviewSnapshot{Files: []workspace.ReviewFile{{Path: "one.go", DiffLine: 0}}, Lines: []workspace.ReviewLine{
		{File: 0, Kind: "file_header", Content: "one.go"},
		{File: 0, Kind: "gap", Gap: &workspace.ReviewGap{ID: 77, Position: position, Count: count, Hidden: hidden}},
		{File: 0, Kind: "addition", NewLine: count + 1, Content: "changed"},
	}}
	a.reviewState.list = ui.ListState{}
	tt.Frame()
	return a, tt
}

// Source sidepane.rs:1758–1764,1803–1808: edge-specific chunk expansion;
// Shift reveals all retained context, without rereading a changing disk.
func TestMountedReviewGapDirectionAndShiftAll(t *testing.T) {
	for _, tc := range []struct {
		position, direction string
		first               int
	}{
		{"leading", "end", 151}, {"trailing", "start", 1}, {"between", "start", 1}, {"between", "end", 151},
	} {
		t.Run(tc.position+"-"+tc.direction, func(t *testing.T) {
			a, tt := mountedReviewGap(t, tc.position, 250)
			if err := tt.Click("Expand context " + tc.direction); err != nil {
				t.Fatal(err)
			}
			var got []int
			for _, line := range a.reviewState.data.Lines {
				if line.Kind == "context" {
					got = append(got, line.NewLine)
				}
			}
			if len(got) != workspace.DefaultExpansionLines || got[0] != tc.first {
				t.Fatal("wrong directional chunk", got)
			}
		})
	}
	a, tt := mountedReviewGap(t, "leading", 250)
	if err := tt.ClickWith(ui.Shift, "Expand context end"); err != nil {
		t.Fatal(err)
	}
	for _, line := range a.reviewState.data.Lines {
		if line.Gap != nil {
			t.Fatal("Shift did not reveal all retained lines")
		}
	}
	if len(a.reviewState.data.Lines) != 252 {
		t.Fatal("Shift expansion lost lines", len(a.reviewState.data.Lines))
	}
}

func TestMountedReviewGapLabelSmallBetweenDisabledAndRepeatedSplit(t *testing.T) {
	for _, count := range []int{70, 250} {
		a, tt := mountedReviewGap(t, "between", count)
		if count <= workspace.DefaultExpansionLines {
			if _, ok := tt.Find("Expand context both"); !ok {
				t.Fatal("small between gap lacks single both-edge button")
			}
			if _, ok := tt.Find("Expand context start"); ok {
				t.Fatal("small between gap incorrectly has two controls")
			}
		}
		if err := tt.Click(fmt.Sprintf("Show %d unchanged lines", count)); err != nil {
			t.Fatal(err)
		}
		retained := 0
		for _, line := range a.reviewState.data.Lines {
			if line.Kind == "context" {
				retained++
			}
		}
		if retained != min(count, workspace.DefaultExpansionLines*2) {
			t.Fatal("label not both-edge expansion", retained)
		}
	}
	a, tt := mountedReviewGap(t, "leading", 250)
	if err := tt.Click("Side-by-side view"); err != nil {
		t.Fatal(err)
	}
	for _, remaining := range []int{150, 50, 0} {
		if err := tt.Click("Expand context end"); err != nil {
			t.Fatal(err)
		}
		got := 0
		for _, line := range a.reviewState.data.Lines {
			if line.Gap != nil {
				got += line.Gap.Count
			}
		}
		if got != remaining {
			t.Fatal("split/repeated expansion resolved wrong snapshot index", got, remaining)
		}
	}
	a, tt = mountedReviewGap(t, "leading", 250)
	a.reviewState.data.Lines[1].Gap.Hidden = nil
	tt.Frame()
	if _, ok := tt.Find("Expand context end"); ok {
		t.Fatal("abbreviated patch must not offer fake context")
	}
	if err := tt.Click("Show 250 unchanged lines"); err != nil {
		t.Fatal(err)
	}
	if len(a.reviewState.data.Lines) != 3 {
		t.Fatal("disabled gap mutated snapshot")
	}
}

func TestReviewGapIdentityUsesStableID(t *testing.T) {
	one := workspace.ReviewLine{File: 0, Kind: "gap", Gap: &workspace.ReviewGap{ID: 1}}
	two := one
	two.Gap = &workspace.ReviewGap{ID: 2}
	if reviewLineKey(one) == reviewLineKey(two) {
		t.Fatal("different gaps in same file share a virtual-row key")
	}
	one.Gap.Count = 100
	key := reviewLineKey(one)
	one.Gap.Count = 50
	if reviewLineKey(one) != key {
		t.Fatal("partial expansion changes gap identity")
	}
}

func TestMountedReviewGapTwoControlsUseSourceFixedHalfHeight(t *testing.T) {
	_, tt := mountedReviewGap(t, "between", 250)
	top, ok := tt.Find("Expand context start")
	if !ok {
		t.Fatal("missing start")
	}
	bottom, ok := tt.Find("Expand context end")
	if !ok {
		t.Fatal("missing end")
	}
	row, ok := tt.Find("Review context gap 77")
	if !ok {
		t.Fatal("missing gap")
	}
	if row.H != 32 || top.H != 16 || bottom.H != 16 || top.Y+top.H != bottom.Y {
		t.Fatal("source fixes first gutter control to16px, not flex-distributed border heights", row, top, bottom)
	}
}
