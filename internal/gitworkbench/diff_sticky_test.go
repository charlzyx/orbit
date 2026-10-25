package gitworkbench

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"math"
	"strings"
	"testing"
)

func TestEmbeddedReviewStackedSectionAndFileHeadersStayActionable(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	viewport, _ := tt.Find("Git diff horizontal viewport")
	tt.Scroll(viewport.X+300, viewport.Y+150, 0, 400)
	section, ok := tt.Find("Git sticky section true")
	if !ok {
		t.Fatal("staged section did not pin")
	}
	file, ok := tt.Find("Git sticky file true:long.go")
	if !ok {
		t.Fatal("file header did not pin beneath section")
	}
	if math.Abs(float64(section.Y-viewport.Y)) > 1 || math.Abs(float64(file.Y-section.Y-section.H)) > 1 {
		t.Fatal("sticky layers overlap or leave viewport", section, file, viewport)
	}
	tt.Scroll(viewport.X+300, viewport.Y+150, 300, 0)
	section, _ = tt.Find("Git sticky section true")
	file, _ = tt.Find("Git sticky file true:long.go")
	if math.Abs(float64(section.X-viewport.X)) > 1 || math.Abs(float64(file.X-viewport.X)) > 1 {
		t.Fatal("sticky controls pan with diff", section, file, viewport)
	}
	// Clicking the header's name collapses this staged diff, without invoking
	// any Git operation or changing the unstaged copy of the same path.
	tt.ClickAt(file.X+file.W/2, file.Y+file.H/2)
	if s.diffExpanded["true:long.go"] || !s.diffExpanded["false:long.go"] {
		t.Fatal("sticky action changed wrong stage/path", s.diffExpanded)
	}
	tt.Scroll(viewport.X+300, viewport.Y+150, 0, 300)
	section, ok = tt.Find("Git sticky section false")
	if !ok {
		t.Fatal("unstaged section did not replace staged")
	}
	file, ok = tt.Find("Git sticky file false:long.go")
	if !ok {
		t.Fatal("unstaged file did not replace staged")
	}
	if math.Abs(float64(section.Y-viewport.Y)) > 1 || math.Abs(float64(file.Y-section.Y-section.H)) > 1 {
		t.Fatal("second group did not keep layer order", section, file)
	}
}

func TestEmbeddedReviewIncomingSectionPushesWholeStickyPair(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	items := s.changeItems()
	next := -1
	for i, item := range items {
		if isChangeSection(item) && !item.staged {
			next = i
			break
		}
	}
	if next < 3 {
		t.Fatal("fixture missing second section")
	}
	s.list.ScrollTo(next-2, ui.Start)
	tt.Frame()
	viewport, _ := tt.Find("Git diff horizontal viewport")
	sawPush := false
	for i := 0; i < 12; i++ {
		tt.Scroll(viewport.X+300, viewport.Y+150, 0, 5)
		incoming, ok := tt.Find("Git section false")
		outgoing, old := tt.Find("Git sticky section true")
		if ok && old && incoming.Y > viewport.Y && incoming.Y < viewport.Y+50 {
			if math.Abs(float64(outgoing.Y+outgoing.H-incoming.Y)) > 1 {
				t.Fatal("section not pushed by incoming group", outgoing, incoming)
			}
			if _, exists := tt.Find("Git sticky file true:long.go"); exists {
				t.Fatal("file header floats over incoming section")
			}
			sawPush = true
			break
		}
	}
	if !sawPush {
		t.Fatal("did not observe staged/unstaged push-off")
	}
}

func TestEmbeddedReviewWorkspaceResetAndHiddenContextWidth(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	hidden := workspace.ReviewLine{File: 0, Kind: "context", Content: strings.Repeat("界", 300)}
	s.Data.StagedReview.Lines = append(s.Data.StagedReview.Lines, workspace.ReviewLine{File: 0, Kind: "gap", Gap: &workspace.ReviewGap{ID: 7, Count: 1, Hidden: []workspace.ReviewLine{hidden}}})
	tt.Frame()
	before := s.diffView.horizontal.MaxX
	s.Data.StagedReview.ExpandGap(len(s.Data.StagedReview.Lines)-1, "all")
	tt.Frame()
	if math.Abs(float64(s.diffView.horizontal.MaxX-before)) > 1 {
		t.Fatal("hidden-context reveal changed extent")
	}
	viewport, _ := tt.Find("Git diff horizontal viewport")
	tt.Scroll(viewport.X+300, viewport.Y+150, 400, 0)
	s.SetWorkspace(t.TempDir())
	if s.diffView.horizontal.X != 0 || s.diffView.horizontal.MaxX != 0 || len(s.diffView.headers) != 0 {
		t.Fatal("new repo inherited pan or header handles")
	}
}

func TestEmbeddedReviewStickyDiscardStillRequiresConfirmation(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	items := s.changeItems()
	target := -1
	for i, item := range items {
		if isChangeFile(item) && !item.staged {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("fixture missing unstaged file")
	}
	s.list.ScrollTo(target+4, ui.Start)
	tt.Frame()
	if _, ok := tt.Find("Git sticky file false:long.go"); !ok {
		t.Fatal("unstaged sticky header missing")
	}
	if err := tt.Click("Discard…"); err != nil {
		t.Fatal(err)
	}
	if !s.confirmOpen || s.Busy {
		t.Fatal("sticky discard bypassed confirmation", s.confirmOpen, s.Busy)
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if s.confirmOpen || s.Busy {
		t.Fatal("cancel failed or started operation")
	}
	if !s.diffExpanded["false:long.go"] {
		t.Fatal("cancel changed expanded diff")
	}
}

func TestEmbeddedReviewNextFilePushesOnlyFileLayer(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	s.Data.Changes = append(s.Data.Changes, workspace.StatusRow{Path: "other.go", Index: 'M', Worktree: ' '})
	diff := &s.Data.StagedReview
	diff.Files = append(diff.Files, workspace.ReviewFile{Path: "other.go", DiffLine: len(diff.Lines)})
	diff.Lines = append(diff.Lines, workspace.ReviewLine{File: 1, Kind: "file_header", Content: "other.go"})
	for i := 1; i <= 100; i++ {
		diff.Lines = append(diff.Lines, workspace.ReviewLine{File: 1, Kind: "context", OldLine: i, NewLine: i, Content: "another file context"})
	}
	s.diffExpanded["true:other.go"] = true
	tt.Frame()
	items := s.changeItems()
	target := -1
	for i, item := range items {
		if isChangeFile(item) && item.staged && item.path == "other.go" {
			target = i
			break
		}
	}
	if target < 2 {
		t.Fatal("fixture missing second file")
	}
	s.list.ScrollTo(target-3, ui.Start)
	tt.Frame()
	viewport, _ := tt.Find("Git diff horizontal viewport")
	sawPush := false
	for i := 0; i < 20; i++ {
		tt.Scroll(viewport.X+300, viewport.Y+180, 0, 5)
		section, secOK := tt.Find("Git sticky section true")
		old, oldOK := tt.Find("Git sticky file true:long.go")
		next, nextOK := tt.Find("Git file true:other.go")
		if secOK && oldOK && nextOK && next.Y > section.Y+section.H && next.Y < section.Y+section.H+42 {
			if math.Abs(float64(section.Y-viewport.Y)) > 1 {
				t.Fatal("file push moved section", section, viewport)
			}
			if math.Abs(float64(old.Y+old.H-next.Y)) > 1 {
				t.Fatal("next file overlaps current file header", old, next)
			}
			sawPush = true
			break
		}
	}
	if !sawPush {
		t.Fatal("did not observe independent file-layer push")
	}
}

func TestEmbeddedReviewWheelOverStickyHeaderStillScrollsDiff(t *testing.T) {
	s, tt := embeddedReviewFixture(t, false)
	viewport, _ := tt.Find("Git diff horizontal viewport")
	tt.Scroll(viewport.X+300, viewport.Y+150, 0, 400)
	header, ok := tt.Find("Git sticky section true")
	if !ok {
		t.Fatal("fixture has no sticky header")
	}
	first, _ := s.list.Visible()
	tt.Scroll(header.X+header.W/2, header.Y+header.H/2, 0, 120)
	if got, _ := s.list.Visible(); got <= first {
		t.Fatal("sticky overlay swallowed vertical wheel", first, got)
	}
}
