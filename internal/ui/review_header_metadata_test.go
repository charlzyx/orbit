package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMountedReviewHeaderSourceMetadata(t *testing.T) {
	for _, tc := range []struct {
		path, status   string
		added, deleted int
	}{
		{"alpha.go", "modified", 12, 3}, {"empty.txt", "added", 0, 0},
		{"image.png", "binary", 0, 0}, {"old => new.go", "modified", 0, 0},
		{strings.Repeat("directory/", 30) + "last.go", "deleted", 0, 12345},
	} {
		for _, sticky := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sticky=%t", tc.status, sticky), func(t *testing.T) {
				a := fixture(t)
				a.reviewState.data.Files = []workspace.ReviewFile{{Path: tc.path, Status: tc.status, Additions: tc.added, Deletions: tc.deleted}}
				tt := ui.NewTester(func(c *ui.Context) {
					c.SetTheme(a.Theme.Native())
					ui.Column(c).Width(320).Children(func() { a.reviewFileHeader(c, workspace.ReviewLine{File: 0, Kind: "file_header"}, sticky) })
				}, 320, 100)
				tt.Frame()
				icon, ok := tt.Find("Review file icon")
				if !ok {
					t.Fatal("source file glyph missing")
				}
				add, aok := tt.Find(fmt.Sprintf("+%d", tc.added))
				del, dok := tt.Find(fmt.Sprintf("-%d", tc.deleted))
				open, ook := tt.Find("Open file")
				if !aok || !dok || !ook {
					t.Fatal("source per-file counts/actions missing", aok, dok, ook)
				}
				path, pok := tt.Find(tc.path)
				if !pok {
					t.Fatal("path missing")
				}
				if icon.W <= 0 || icon.X+icon.W > path.X || path.X+path.W > add.X+.5 || add.X+add.W > del.X || del.X+del.W > open.X || open.X+open.W > 320 {
					t.Fatal("header icon/path/counts/action overlap or clip", icon, path, add, del, open)
				}
				if add.H > 34 || del.H > 34 {
					t.Fatal("counts wrapped", add, del)
				}
				if err := tt.Click("Toggle file diff"); err != nil {
					t.Fatal(err)
				}
				if !a.reviewState.collapsed[0] {
					t.Fatal("toggle ownership changed")
				}
			})
		}
	}
}

func TestMountedReviewHeaderMetadataFollowsStickySnapshotAndOpen(t *testing.T) {
	a, tt := mountedReview(t)
	s := &a.reviewState
	a.Model.Workspace = t.TempDir()
	s.root = a.Model.Workspace
	if err := os.WriteFile(filepath.Join(a.Model.Workspace, "one.go"), []byte("package fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	s.data.Files[0].Additions, s.data.Files[0].Deletions = 17, 9
	s.data.Lines[1].Content = strings.Repeat("wide_", 100)
	tt.Frame()
	if _, ok := tt.Find("+17"); !ok {
		t.Fatal("normal header metadata missing")
	}
	if err := tt.Click("Unwrap lines"); err != nil {
		t.Fatal(err)
	}
	viewport, _ := tt.Find("Review horizontal viewport")
	tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 0, 450)
	tt.Scroll(viewport.X+viewport.W/2, viewport.Y+90, 400, 0)
	header, ok := tt.Find("Review sticky file header 0")
	if !ok {
		t.Fatal("sticky header missing")
	}
	for _, label := range []string{"Review file icon", "+17", "-9"} {
		rect, ok := tt.Find(label)
		if !ok || rect.Y < header.Y || rect.Y+rect.H > header.Y+header.H+.5 {
			t.Fatal("sticky metadata missing or outside header", label, rect, header)
		}
	}
	s.data.Files = append([]workspace.ReviewFile(nil), s.data.Files...)
	s.data.Files[0].Additions, s.data.Files[0].Deletions = 21, 0
	s.generation++
	tt.Frame()
	if _, ok := tt.Find("+21"); !ok {
		t.Fatal("replacement snapshot metadata stale")
	}
	if _, ok := tt.Find("-0"); !ok {
		t.Fatal("zero deletion metadata missing")
	}
	// Use the real retained header Open action after both scroll axes moved.
	open, ok := tt.Find("Open file")
	if !ok {
		t.Fatal("Open action lost")
	}
	tt.ClickAt(open.X+open.W/2, open.Y+open.H/2)
	if a.files == nil || a.files.desired != filepath.Join(a.Model.Workspace, "one.go") {
		t.Fatal("Open action retargeted", a.files)
	}
}

func TestMountedReviewHeaderUsesParserZeroAndBinaryMetadata(t *testing.T) {
	for _, tc := range []struct{ numstat, patch, path string }{
		{"-\t-\tpicture.png\n", "diff --git a/picture.png b/picture.png\nBinary files a/picture.png and b/picture.png differ\n", "picture.png"},
		{"0\t0\tempty.txt\n", "diff --git a/empty.txt b/empty.txt\nnew file mode 100644\n", "empty.txt"},
		{"", "diff --git a/before.txt b/after.txt\nsimilarity index 100%\nrename from before.txt\nrename to after.txt\n", "after.txt"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			a := fixture(t)
			a.reviewState.data = workspace.ParseReview(workspace.ReviewDiff{Numstat: tc.numstat, Patch: tc.patch})
			if len(a.reviewState.data.Lines) == 0 {
				t.Fatal("parser fixture has no header")
			}
			tt := ui.NewTester(func(c *ui.Context) {
				c.SetTheme(a.Theme.Native())
				ui.Column(c).Width(320).Children(func() { a.reviewFileHeader(c, a.reviewState.data.Lines[0], false) })
			}, 320, 80)
			tt.Frame()
			for _, label := range []string{tc.path, "+0", "-0"} {
				if _, ok := tt.Find(label); !ok {
					t.Fatal("parsed source metadata missing", label)
				}
			}
		})
	}
}
