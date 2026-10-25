package orbitui

import (
	"math"
	"reflect"
	"testing"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func TestReviewTreeSourceOrderingFilterAndExpansion(t *testing.T) {
	files := []workspace.ReviewFile{{Path: "src/z.go"}, {Path: "README.md"}, {Path: "src/nested/B.go"}, {Path: "src/a.go"}, {Path: "Á.md"}}
	expanded := reviewDirectories(files)
	rows := reviewTreeRows(files, expanded, "")
	var paths []string
	for _, r := range rows {
		paths = append(paths, r.path)
	}
	if want := []string{"README.md", "src", "src/a.go", "src/nested", "src/nested/B.go", "src/z.go", "Á.md"}; !reflect.DeepEqual(paths, want) {
		t.Fatal(paths)
	}
	delete(expanded, "src")
	rows = reviewTreeRows(files, expanded, "")
	if len(rows) != 3 || rows[1].expanded {
		t.Fatal("collapsed parent leaked descendants", rows)
	}
	rows = reviewTreeRows(files, expanded, "  NESTED/b  ")
	if len(rows) != 3 || !rows[0].expanded || !rows[1].expanded || rows[2].file != 2 {
		t.Fatal("filter did not reveal ancestors", rows)
	}
	if expanded["src"] {
		t.Fatal("filter mutated saved expansion")
	}
	if got := reviewTreeRows(files, expanded, "á.md"); len(got) != 0 {
		t.Fatal("source uses ASCII folding, not Unicode case folding", got)
	}
}

func TestReviewTreeRefreshKeepsClosedDirsAndExpandsNewOnes(t *testing.T) {
	s := reviewState{data: workspace.ReviewSnapshot{Files: []workspace.ReviewFile{{Path: "src/a.go"}, {Path: "docs/a.md"}}}}
	s.syncTreeRows()
	s.toggleTreeDirectory("src")
	s.data.Files = append(s.data.Files, workspace.ReviewFile{Path: "new/dir/file.txt"}, workspace.ReviewFile{Path: "src/b.go"})
	s.syncTreeRows()
	if s.expandedDirs["src"] || !s.expandedDirs["new"] || !s.expandedDirs["new/dir"] {
		t.Fatal("refresh reset known dir or failed to expand new dirs", s.expandedDirs)
	}
}

func TestMountedReviewTreeHierarchyFilterKeyboardAndResponsiveBounds(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	s.data.Files[0].Path = "src/a.go"
	s.data.Files[1].Path = "src/nested/b.go"
	s.treeDirty = true
	tt.Frame()
	r, ok := tt.Find("Review file tree")
	if !ok || math.Abs(float64(r.W-193.2)) > .05 {
		t.Fatal("source default460 pane/42% tree width", r, ok)
	}
	if err := tt.Click("Review directory src"); err != nil {
		t.Fatal(err)
	}
	if len(s.treeRows) != 1 || s.treeRows[0].expanded {
		t.Fatal("pointer did not collapse directory", s.treeRows)
	}
	tt.Key(0, ui.KeyEnter)
	if len(s.treeRows) != 4 || !s.treeRows[0].expanded {
		t.Fatal("Enter did not expand directory", s.treeRows)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown) // nested directory
	tt.Key(0, ui.KeySpace)
	if s.expandedDirs["src/nested"] {
		t.Fatal("Space did not collapse nested directory")
	}
	if err := tt.Click("Filter review files"); err != nil {
		t.Fatal(err)
	}
	tt.Type("B.GO")
	if len(s.treeRows) != 3 || s.treeRows[2].file != 1 {
		t.Fatal("native filter missing path/ancestors", s.treeRows)
	}
	tt.Key(0, ui.KeyEscape)
	if s.treeQuery != "" || !a.ReviewVisible || s.expandedDirs["src/nested"] {
		t.Fatal("filter Escape closed pane or lost collapse state")
	}
	a.ReviewWidth = 439
	tt.Frame()
	if _, ok := tt.Find("Review file tree"); ok {
		t.Fatal("tree visible below source440 threshold")
	}
	a.ReviewWidth = 440
	tt.Frame()
	if r, ok := tt.Find("Review file tree"); !ok || math.Abs(float64(r.W-184.8)) > .05 {
		t.Fatal("tree not restored at threshold", r, ok)
	}
	a.ReviewWidth = 900
	tt.Frame()
	if r, ok := tt.Find("Review file tree"); !ok || r.W != 240 {
		t.Fatal("tree exceeded source240 cap", r, ok)
	}
}

func TestMountedReviewCardOccludesNewTaskBackdrop(t *testing.T) {
	a, tt := mountedReview(t)
	r, ok := tt.Find("Review card")
	if !ok || r.X != 1022 || r.W != 448 {
		t.Fatal("source12px gutter and460px dock", r, ok)
	}
	tree, ok := tt.Find("Review file tree")
	if !ok {
		t.Fatal("missing tree")
	}
	im := tt.Image()
	want := a.Theme.Background
	for y := int(tree.Y + 300); y < int(tree.Y+340); y++ {
		for x := int(tree.X + 20); x < int(tree.X+tree.W-20); x++ {
			r, g, b, a := im.At(x, y).RGBA()
			if uint8(r>>8) != want.R || uint8(g>>8) != want.G || uint8(b>>8) != want.B || a != 65535 {
				t.Fatalf("backdrop leaked through review at %d,%d: %d/%d/%d", x, y, r>>8, g>>8, b>>8)
			}
		}
	}
}
