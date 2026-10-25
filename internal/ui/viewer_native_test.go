package orbitui

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func mountedFile(t *testing.T, text string) (*App, *ui.Tester, string) {
	t.Helper()
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	root := t.TempDir()
	path := filepath.Join(root, "fixture.txt")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	a.Model.Workspace = root
	a.initFiles()
	d, err := a.files.editor.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	a.page = "files"
	t.Cleanup(func() { _ = a.files.editor.CloseAll() })
	// Mount the actual editor and tab controls, but apply worker completions
	// explicitly so close/reopen races can be scheduled deterministically.
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Background(a.Theme.Background).Children(func() { a.fileEditor(c) })
	}, 1000, 700)
	tt.Frame()
	if err := tt.Click("File editor"); err != nil {
		t.Fatal(err)
	}
	return a, tt, path
}

func TestMountedFileSelectionUndoRedoAndTabCaret(t *testing.T) {
	a, tt, path := mountedFile(t, "first 中文\nsecond line\n")
	st := a.files.states[path]
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != "first 中文\nsecond line\n" {
		t.Fatal("mounted editor did not copy full Unicode selection", tt.Clipboard())
	}
	tt.Type("replacement 中文")
	if st.Buffer.String() != "replacement 中文" {
		t.Fatal("mounted selection replacement", st.Buffer.String())
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if st.Buffer.String() != "first 中文\nsecond line\n" || st.Selection().Anchor == st.Selection().Head {
		t.Fatal("undo did not restore selected text", st.Buffer.String(), st.Selection())
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
	if st.Buffer.String() != "replacement 中文" || st.Cursor() != len("replacement 中文") {
		t.Fatal("redo did not restore text and UTF-8 caret", st.Buffer.String(), st.Selection())
	}
	tt.Key(0, ui.KeyLeft)
	want := st.Selection()
	other := filepath.Join(filepath.Dir(path), "other.txt")
	if err := os.WriteFile(other, []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := a.files.editor.Open(other)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	tt.Frame()
	if err := tt.Click("fixture.txt •"); err != nil {
		t.Fatal(err)
	}
	if st.Selection() != want {
		t.Fatal("switching tabs lost caret", st.Selection(), want)
	}
	tt.Type("!")
	if st.Buffer.String() != "replacement 中!文" {
		t.Fatal("tab activation did not restore editor keyboard focus", st.Buffer.String())
	}
}

func TestMountedFileFailedCloseKeepsDraftAndCanRetry(t *testing.T) {
	a, tt, path := mountedFile(t, "saved")
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("unsaved 中文")
	backup := path + ".original"
	if err := os.Rename(path, backup); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Close file"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { d, ok := a.files.editor.Active(); return ok && d.LastError != "" })
	tt.Frame()
	d, ok := a.files.editor.Active()
	if !ok || !d.Dirty || d.Text != "unsaved 中文" || a.files.states[path].Buffer.String() != d.Text || a.page != "files" {
		t.Fatal("failed close discarded draft", d, a.page)
	}
	if err := os.Remove(path); err != nil { // Empty directory created by this fixture.
		t.Fatal(err)
	}
	if err := os.Rename(backup, path); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { d, _ := a.files.editor.Active(); return !d.Dirty && d.LastError == "" })
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "unsaved 中文" {
		t.Fatal("retry did not save retained draft", string(got), err)
	}
}

func TestMountedFileStaleCloseCompletionKeepsReopenedState(t *testing.T) {
	a, tt, path := mountedFile(t, "first\nsecond\n")
	if err := tt.Click("Close file"); err != nil {
		t.Fatal(err)
	}
	var finish func()
	select {
	case finish = <-a.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("close did not complete")
	}
	d, err := a.files.editor.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	st := a.files.states[path]
	st.Select(2, 2)
	finish()
	tt.Frame()
	if a.files.states[path] != st || a.files.states[path].Cursor() != 2 {
		t.Fatal("old close completion deleted a reopened editor's state")
	}
}

func TestMountedFilePrimaryWAndShiftWHaveDistinctScopes(t *testing.T) {
	a, tt, path := mountedFile(t, "keep this tab")
	tt.Key(ui.Cmd|ui.Shift, ui.KeyW)
	if a.page != "chat" || len(a.files.editor.Documents()) != 1 {
		t.Fatal("shift-W did not hide Files while retaining its tab")
	}
	a.page = "files"
	tt.Frame()
	if err := tt.Click("File editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyW)
	drainRegressionUpdates(t, a, func() bool { return len(a.files.editor.Documents()) == 0 && a.page == "chat" })
	if _, ok := a.files.states[path]; ok {
		t.Fatal("last-tab close retained stale editor state")
	}
}

func TestMountedFileReopenQueuedBehindCloseCannotStealLaterSelection(t *testing.T) {
	a, tt, path := mountedFile(t, "first")
	other := filepath.Join(filepath.Dir(path), "other.txt")
	if err := os.WriteFile(other, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := a.files.editor.Open(other)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	a.activateFile(path)
	tt.Frame()
	if err := tt.Click("Close file"); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(path)      // Queued until the pending close is reaped.
	a.activateFile(other) // A newer explicit choice cancels that queued reveal.
	drainRegressionUpdates(t, a, func() bool { return len(a.files.closing) == 0 })
	tt.Frame()
	active, _ := a.files.editor.Active()
	if active.Content.Path != other {
		t.Fatal("old close/reopen stole newer active tab", active.Content.Path)
	}
}

func mountedReview(t *testing.T) (*App, *ui.Tester) {
	t.Helper()
	a := fixture(t)
	a.ReviewVisible = true
	s := &a.reviewState
	s.tree = true
	s.root = a.Model.Workspace
	for file, path := range []string{"one.go", "two.go"} {
		s.data.Files = append(s.data.Files, workspace.ReviewFile{Path: path, DiffLine: len(s.data.Lines)})
		s.data.Lines = append(s.data.Lines, workspace.ReviewLine{File: file, Kind: "file_header", Content: path})
		for line := 1; line <= 65; line++ {
			kind := "context"
			if line == 8 || line == 35 {
				kind = "addition"
				s.data.Lines = append(s.data.Lines, workspace.ReviewLine{File: file, Kind: "deletion", OldLine: line, Content: fmt.Sprintf("old file%d line%d", file, line)})
			}
			s.data.Lines = append(s.data.Lines, workspace.ReviewLine{File: file, Kind: kind, NewLine: line, Content: fmt.Sprintf("file%d line%d", file, line)})
		}
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Review file one.go"); err != nil {
		t.Fatal(err)
	}
	return a, tt
}

func TestMountedReviewFileKeysClampAtEnds(t *testing.T) {
	a, tt := mountedReview(t)
	for _, selected := range []int{1, 1} {
		tt.Key(0, ui.KeyN)
		if a.reviewState.selected != selected {
			t.Fatal("source next-file must clamp, not wrap", a.reviewState.selected)
		}
	}
	for _, selected := range []int{0, 0} {
		tt.Key(0, ui.KeyP)
		if a.reviewState.selected != selected {
			t.Fatal("source previous-file must clamp, not wrap", a.reviewState.selected)
		}
	}
}

func TestMountedReviewTreeCursorWaitsForEnter(t *testing.T) {
	a, tt := mountedReview(t)
	tt.Key(0, ui.KeyDown)
	if a.reviewState.selected != 0 {
		t.Fatal("tree cursor movement selected a file before Enter")
	}
	a.reviewState.collapsed = map[int]bool{1: true}
	tt.Key(0, ui.KeyEnter)
	if a.reviewState.selected != 1 || a.reviewState.collapsed[1] {
		t.Fatal("tree Enter did not select and expand its file")
	}
	for range 3 {
		tt.Key(0, ui.KeyDown)
	}
	if a.reviewState.treeCursor == nil || *a.reviewState.treeCursor != 1 {
		t.Fatal("tree cursor wrapped at last row")
	}
}

func TestMountedReviewSplitPreservesLogicalContextRow(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.list.ScrollTo(50, ui.Start)
	tt.Frame()
	a.toggleReviewSplit()
	tt.Frame()
	first, _ := a.reviewState.list.Visible()
	if first != 48 {
		t.Fatal("split toggle lost context row after two paired changes", first)
	}
	a.toggleReviewSplit()
	tt.Frame()
	first, _ = a.reviewState.list.Visible()
	if first != 50 {
		t.Fatal("unified toggle failed to restore same context row", first)
	}
}

func TestMountedReviewHunkKeysNavigateFullContextRuns(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(fmt.Sprint(split), func(t *testing.T) {
			a, tt := mountedReview(t)
			a.reviewState.split = split
			tt.Frame()
			targets := []int{8, 36, 76}
			if split {
				targets = []int{8, 35, 74}
			}
			for _, want := range targets {
				tt.Key(0, ui.KeyBracketRight)
				tt.Frame()
				first, _ := a.reviewState.list.Visible()
				if first != want {
					t.Fatalf("full-context hunk target got%d want%d", first, want)
				}
			}
			if a.reviewState.selected != 1 {
				t.Fatal("hunk selection did not follow the file")
			}
		})
	}
}

func TestMountedReviewRefreshRetainsSelectedPathAfterReorder(t *testing.T) {
	a := fixture(t)
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	for _, name := range []string{"b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("before\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := g.Run(context.Background(), "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run(context.Background(), "commit", "-m", "fixtures"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("after\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a.Model.Workspace = root
	a.ReviewVisible = true
	a.reviewState.tree = true
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Review file c.txt"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("new\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run(context.Background(), "add", "a.txt"); err != nil {
		t.Fatal(err)
	}
	a.RefreshReview()
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	tt.Frame()
	s := &a.reviewState
	if s.selected < 0 || s.selected >= len(s.data.Files) || s.data.Files[s.selected].Path != "c.txt" {
		t.Fatal("refresh moved selected file by ordinal", s.selected, s.data.Files)
	}
}

func TestMountedReviewOpenFileKeepsWholeSnapshotAndCommitFilter(t *testing.T) {
	a := fixture(t)
	root := regressionRepo(t)
	g := workspace.Git{Dir: root}
	for _, name := range []string{"one.txt", "two.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("before\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-m", "review files"}} {
		if _, err := g.Run(context.Background(), args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"one.txt", "two.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("after\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a.Model.Workspace = root
	// Mount the callback seam used by Git's changed-file action. No fake diff
	// is injected: both file contents come from the disposable Git repository.
	tt := ui.NewTester(func(c *ui.Context) {
		if ui.Button(c, "Open file diff").Clicked() {
			a.OpenReviewDiff("unstaged", "", "two.txt")
		}
	}, 600, 400)
	tt.Frame()
	if err := tt.Click("Open file diff"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	s := &a.reviewState
	if len(s.data.Files) != 2 || s.data.Files[s.selected].Path != "two.txt" || s.source != workspace.ReviewUncommitted || s.path != "" {
		t.Fatal("working-tree file action filtered instead of selecting", s.source, s.path, s.selected, s.data.Files)
	}
	// The distinct explicit-ref API still isolates that historical file.
	a.OpenReviewDiff("committed", "HEAD", "two.txt")
	drainRegressionUpdates(t, a, func() bool { return !a.reviewState.loading })
	if s.err != "" || len(s.data.Files) != 1 || s.data.Files[0].Path != "two.txt" || s.ref != "HEAD" {
		t.Fatal("explicit committed-file comparison lost scope", s.err, s.data.Files)
	}
}

func TestMountedReviewTreeKeysDoNotHijackToolbar(t *testing.T) {
	a, tt := mountedReview(t)
	if err := tt.Click("Review options"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	selected := a.reviewState.selected
	tt.Key(0, ui.KeyN)
	if a.reviewState.selected != selected {
		t.Fatal("tree navigation escaped its native focus scope")
	}
}

func TestMountedFileMarkdownPreviewRestoresEditorFocusAndCaret(t *testing.T) {
	a, tt, path := mountedFile(t, "other")
	md := filepath.Join(filepath.Dir(path), "notes.md")
	if err := os.WriteFile(md, []byte("# Notes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := a.files.editor.Open(md)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	tt.Frame()
	if err := tt.Click("Preview"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.Focused("File editor") {
		t.Fatal("preview-to-source did not focus the native editor")
	}
	tt.Key(ui.Cmd, ui.KeyEnd)
	tt.Type("中文")
	st := a.files.states[md]
	selection := st.Selection()
	if err := tt.Click("Preview"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	if d, _ := a.files.editor.Active(); !d.Preview {
		t.Fatal("source-to-preview left Enter trapped on the toggle")
	}
	if err := tt.Click("Preview"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.Focused("File editor") || st.Selection() != selection {
		t.Fatal("preview roundtrip lost editor focus or caret")
	}
	tt.Type("!")
	if st.Buffer.String() != "# Notes\n中文!" {
		t.Fatal("preview roundtrip lost buffer", st.Buffer.String())
	}
}

// These are explicitly fixture CPU captures, not matched desktop baselines.
func TestNativeViewerCapture(t *testing.T) {
	out := os.Getenv("ORBIT_VIEWER_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_VIEWER_CAPTURE_DIR to capture native viewer fixtures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	save := func(name string, tt *ui.Tester) {
		t.Helper()
		tt.Frame()
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
	a, tt := mountedReview(t)
	a.reviewState.data.Files[0].Path = "src/editor.go"
	a.reviewState.data.Files[1].Path = "src/viewer/review.go"
	a.reviewState.treeDirty = true
	save("review-default-fixture", tt)
	a.ReviewWidth = 900
	save("review-wide-fixture", tt)
	a, tt, _ = mountedFile(t, "first 中文\nsecond line\n")
	tt.Key(ui.Cmd, ui.KeyA)
	save("editor-selection-fixture", tt)
}

func TestMountedFileOpenCachedTabIsImmediateAndKeepsNativeBuffer(t *testing.T) {
	a, tt, path := mountedFile(t, "saved")
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("current 中文")
	st := a.files.states[path]
	selection := st.Selection()
	a.OpenFile(path)
	if a.files.focusPending != path {
		t.Fatal("cached tab open unnecessarily waited for an asynchronous snapshot")
	}
	tt.Frame()
	tt.Type("!")
	if st.Buffer.String() != "current 中文!" || selection.Head != len("current 中文") {
		t.Fatal("cached open changed native text/caret", st.Buffer.String(), selection)
	}
}

func TestMountedFileOpenCompletionUsesCurrentDocumentAfterReload(t *testing.T) {
	a, tt, first := mountedFile(t, "first")
	path := filepath.Join(filepath.Dir(first), "loaded.txt")
	if err := os.WriteFile(path, []byte("old disk"), 0600); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(path)
	var finish func()
	select {
	case finish = <-a.updates:
	case <-time.After(5 * time.Second):
		t.Fatal("open did not complete")
	}
	if err := os.WriteFile(path, []byte("new disk"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.files.editor.Reload(path); err != nil {
		t.Fatal(err)
	}
	finish()
	tt.Frame()
	if st := a.files.states[path]; st == nil {
		t.Fatal("loaded document has no native buffer")
	} else if st.Buffer.String() != "new disk" {
		t.Fatal("old open snapshot replaced newer reloaded document", st.Buffer.String())
	}
}
