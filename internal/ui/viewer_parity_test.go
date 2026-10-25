package orbitui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
)

func TestReviewNavigationUsesRenderedCollapsedSplitRows(t *testing.T) {
	a := fixture(t)
	s := &a.reviewState
	s.data = workspace.ReviewSnapshot{Files: []workspace.ReviewFile{{Path: "a", DiffLine: 0}, {Path: "b", DiffLine: 5}}, Lines: []workspace.ReviewLine{
		{File: 0, Kind: "file_header"}, {File: 0, Kind: "hunk_header"}, {File: 0, Kind: "deletion"}, {File: 0, Kind: "addition"}, {File: 0, Kind: "context"},
		{File: 1, Kind: "file_header"}, {File: 1, Kind: "hunk_header"}, {File: 1, Kind: "deletion"}, {File: 1, Kind: "addition"},
	}}
	for _, split := range []bool{false, true} {
		s.split = split
		s.selected = 0
		a.reviewCommand("ReviewCollapseAll")
		a.reviewFile(1)
		if s.collapsed[1] || !s.collapsed[0] {
			t.Fatal("file navigation did not expand only the destination")
		}
		markers := s.markers("file_header")
		if len(markers) != 2 || markers[1].row != 1 {
			t.Fatalf("split=%v wrong file row: %+v", split, markers)
		}
		hunks := s.markers("hunk_header")
		if len(hunks) != 1 || hunks[0].row != 2 {
			t.Fatalf("split=%v wrong hunk row: %+v", split, hunks)
		}
		a.reviewCommand("ReviewTreeToggle")
		if !s.collapsed[1] {
			t.Fatal("tree toggle did not collapse")
		}
	}
}

func TestInactiveFileLoadCannotStealSelectionAfterWorkspaceSwitch(t *testing.T) {
	a := fixture(t)
	a.initFiles()
	root := t.TempDir()
	a.Model.Workspace = root
	first, second := filepath.Join(root, "first.txt"), filepath.Join(root, "second.txt")
	for _, path := range []string{first, second} {
		if e := os.WriteFile(path, []byte("content"), 0600); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := a.files.editor.Open(first); e != nil {
		t.Fatal(e)
	}
	a.OpenFile(second)
	var apply func()
	select {
	case apply = <-a.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("file load timed out")
	}
	a.Model.Workspace = t.TempDir()
	apply()
	d, ok := a.files.editor.Active()
	if !ok || d.Content.Path != first {
		t.Fatal("stale worker changed active tab", d.Content.Path)
	}
	_ = a.files.editor.CloseAll()
}

func TestDirtyPathsBlockRenameAndDelete(t *testing.T) {
	a := fixture(t)
	a.initFiles()
	p := filepath.Join(t.TempDir(), "unsaved.txt")
	os.WriteFile(p, []byte("saved"), 0600)
	a.files.editor.Open(p)
	a.files.editor.SetText(p, "unsaved")
	a.renameFile(p)
	a.deleteFile(p)
	if a.confirmOpen || a.promptOpen {
		t.Fatal("dirty path offered destructive override")
	}
	b, e := os.ReadFile(p)
	if e != nil || string(b) != "saved" {
		t.Fatal("path changed", e, string(b))
	}
	a.files.editor.CloseAll()
}

func TestTruncatedViewerBufferStaysBounded(t *testing.T) {
	a := fixture(t)
	a.initFiles()
	p := filepath.Join(t.TempDir(), "many-lines.txt")
	d := workspace.Document{Text: strings.Repeat("full\n", 1000), Content: workspace.FileContent{Path: p, Truncated: true, Lines: []string{"first", "second"}, Mode: workspace.FileText}}
	a.syncDocument(d)
	if got := a.files.states[p].Buffer.String(); got != "first\nsecond" {
		t.Fatalf("viewer loaded uncapped source: %d bytes", len(got))
	}
}

func TestReviewKeepsSameSourceSnapshotOnRefreshFailure(t *testing.T) {
	a := fixture(t)
	root := regressionRepo(t)
	a.Model.Workspace = root
	os.WriteFile(filepath.Join(root, "changed"), []byte("new"), 0600)
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if len(a.reviewState.data.Lines) == 0 {
		t.Fatal("no initial review")
	}
	if e := os.Rename(filepath.Join(root, ".git"), filepath.Join(root, ".git-away")); e != nil {
		t.Fatal(e)
	}
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if a.reviewState.err == "" || len(a.reviewState.data.Lines) == 0 {
		t.Fatal("failed refresh discarded same-source snapshot")
	}
	a.reviewState.source = workspace.ReviewStaged
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if len(a.reviewState.data.Lines) != 0 {
		t.Fatal("failure displayed a snapshot belonging to another source")
	}
}
