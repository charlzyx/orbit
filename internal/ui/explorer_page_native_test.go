package orbitui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

// Baseline 299a32b app/view.rs renders FileViewer in the main column while
// ProjectPanel remains the sole right-docked tree. Mount the whole application:
// isolated editor tests cannot catch sharing one ListState across two trees.
func TestMountedExplorerOpensFilesWithoutDuplicatingTree(t *testing.T) {
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	root := t.TempDir()
	for name, text := range map[string]string{"alpha.go": "package alpha\n", "bravo.go": "package bravo\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a.Model.Workspace = root
	a.ExplorerVisible = true
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	t.Cleanup(func() { _ = a.files.editor.CloseAll() })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	for _, name := range []string{"alpha.go", "bravo.go"} {
		if err := tt.Click(name); err != nil {
			t.Fatal(err)
		}
		drainRegressionUpdates(t, a, func() bool {
			d, ok := a.files.editor.Active()
			return ok && d.Content.Path == filepath.Join(root, name) && a.files.states[d.Content.Path] != nil
		})
		tt.Frame()
		if a.page != "files" || !a.ExplorerVisible {
			t.Fatal("opening file must retain right Explorer", a.page, a.ExplorerVisible)
		}
		editorBounds, found := tt.Find("File editor")
		if !found {
			t.Fatal("editor not visible")
		}
		treeBounds, found := tt.Find("Filter files")
		if !found || treeBounds.X <= editorBounds.X+editorBounds.W/2 {
			t.Fatal("Explorer must remain right of main editor", editorBounds, treeBounds, found)
		}
	}
	if len(a.files.editor.Documents()) != 2 {
		t.Fatal("second open did not retain both tabs")
	}
	for range 2 {
		if err := tt.Click("Close file"); err != nil {
			t.Fatal(err)
		}
		drainRegressionUpdates(t, a, func() bool { return len(a.files.closing) == 0 })
		tt.Frame()
		if len(a.files.editor.Documents()) != 1 {
			t.Fatal("closing active file lost unrelated tab")
		}
		if err := tt.Click("bravo.go"); err != nil {
			t.Fatal(err)
		}
		drainRegressionUpdates(t, a, func() bool {
			d, ok := a.files.editor.Active()
			return ok && d.Content.Path == filepath.Join(root, "bravo.go") && a.files.states[d.Content.Path] != nil
		})
		tt.Frame()
	}
	for range 2 {
		a.command("ToggleSidebar", "")
		tt.Frame()
		if _, found := tt.Find("File editor"); !found {
			t.Fatal("sidebar toggle removed Files editor")
		}
	}
	a.Navigate("chat")
	tt.Frame()
	a.goHistory(-1)
	tt.Frame()
	if a.page != "files" || !a.ExplorerVisible {
		t.Fatal("history did not restore Files with its sole right tree")
	}
	a.ExplorerVisible = false
	tt.Frame()
	if _, found := tt.Find("Filter files"); found {
		t.Fatal("Files injected a replacement tree while Explorer is hidden")
	}
	if _, found := tt.Find("File editor"); !found {
		t.Fatal("hiding Explorer removed editor")
	}
	// Opening Review replaces the dock, not the editor. A full-width round
	// trip must retain its native buffer and focus/undo state.
	a.openReviewDock()
	tt.Frame()
	if err := tt.Click("File editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("package replacement\n")
	st := a.files.states[filepath.Join(root, "bravo.go")]
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	tt.Type("hidden")
	if st.Buffer.String() != "package replacement\n" {
		t.Fatal("hidden editor accepted input")
	}
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if st.Buffer.String() != "package bravo\n" {
		t.Fatal("Review round trip lost editor focus/undo", st.Buffer.String())
	}
	a.command("ToggleProjectPanel", "")
	tt.Frame()
	if a.ReviewVisible || !a.ExplorerVisible {
		t.Fatal("Explorer did not replace Review")
	}
}
