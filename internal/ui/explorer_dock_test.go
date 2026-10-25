package orbitui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func mountedExplorer(t *testing.T) (*App, *ui.Tester, string) {
	t.Helper()
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	root := t.TempDir()
	for _, name := range []string{"docs/readme.md", "src/main.go", "alpha.go", "bravo.go", "notes.txt"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	a.Model.Workspace, a.ExplorerVisible = root, true
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	t.Cleanup(func() { _ = a.files.editor.CloseAll() })
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	return a, tt, root
}

func explorerBounds(t *testing.T, tt *ui.Tester, label string) ui.Rect {
	t.Helper()
	r, ok := tt.Find(label)
	if !ok {
		t.Fatal("missing Explorer element", label)
	}
	return r
}

// Source 299a32b panel.rs:40,1637: fixed 248 outer, 12 gutter, 6 handle
// at left9, 40 header, 36 filter, 28 rows, 26 footer. No duplicated tree.
func TestMountedExplorerSourceGeometryAndDragPersistence(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	dock := explorerBounds(t, tt, "Explorer dock")
	card := explorerBounds(t, tt, "Explorer card")
	handle := explorerBounds(t, tt, "Resize explorer")
	if dock.X != 1222 || dock.W != 248 || card.X != 1234 || card.W != 236 || card.Y != 44 || card.H != 878 {
		t.Fatal("source outer/card geometry", dock, card)
	}
	if handle.X != dock.X+9 || handle.W != 6 || handle.H != dock.H {
		t.Fatal("source visible-border grab strip", handle, dock)
	}
	for label, height := range map[string]float32{"Explorer toolbar": 40, "Explorer filter row": 36, "Explorer footer": 26, "Explorer row alpha.go": 28} {
		if r := explorerBounds(t, tt, label); r.H != height {
			t.Fatal("source element height", label, r)
		}
	}
	if !tt.HasText("5 files") {
		t.Fatal("footer must count all files, not visible directories/rows")
	}
	dragTo := func(x float32, want float32) {
		t.Helper()
		r := explorerBounds(t, tt, "Resize explorer")
		y := r.Y + 200
		tt.Press(r.X+r.W/2, y)
		tt.Move(x, y)
		tt.Release(x, y)
		tt.Frame()
		if got := a.explorerPaneWidth(); got != want {
			t.Fatal("source pointer-distance drag", got, want)
		}
		if got := a.Options.Store.Layout().ExplorerWidth; got != float64(want) {
			t.Fatal("drag was not persisted", got, want)
		}
	}
	// Source uses the pointer's absolute position, not initial-width minus dx.
	dragTo(1134, 336)
	dragTo(400, 460)
	dragTo(1460, 172)
	dragTo(1120, 350)
	if err := tt.Click("Hide explorer"); err != nil {
		t.Fatal(err)
	}
	if _, found := tt.Find("Explorer card"); found {
		t.Fatal("close left a second tree")
	}
	a.command("ToggleProjectPanel", "")
	tt.Frame()
	if got := explorerBounds(t, tt, "Explorer dock").W; got != 350 {
		t.Fatal("reopen lost dock width", got)
	}
	for range 2 {
		a.command("ToggleSidebar", "")
		tt.Frame()
		if explorerBounds(t, tt, "Explorer dock").W != 350 {
			t.Fatal("sidebar altered Explorer width")
		}
	}
	// Repeated moves use the live pointer, including reversing after a clamp.
	r := explorerBounds(t, tt, "Resize explorer")
	tt.Press(r.X+r.W/2, r.Y+200)
	tt.Move(400, r.Y+200)
	tt.Move(1120, r.Y+200)
	tt.Release(1120, r.Y+200)
	if a.explorerPaneWidth() != 350 {
		t.Fatal("drag reversal accumulated width deltas")
	}
	// Restart loads the persisted source width, not the old 200px fallback.
	reopened := New(Options{Store: a.Options.Store, Workspace: root})
	drainRegressionUpdates(t, reopened, func() bool { return !reopened.backdrop.loading })
	if reopened.explorerPaneWidth() != 350 {
		t.Fatal("new app did not load Explorer width")
	}
	for value, want := range map[float32]float32{0: 248, 80: 172, 3000: 460} {
		reopened.ExplorerWidth = value
		if reopened.explorerPaneWidth() != want {
			t.Fatal("source restored-width clamp", value, reopened.explorerPaneWidth())
		}
	}
}

func TestMountedExplorerActiveSelectionAndKeyboardCursor(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	if a.files.selected != -1 || a.files.active != "" {
		t.Fatal("source initially has no active row or keyboard cursor")
	}
	assertActivePaint := func(path string, selected bool) {
		t.Helper()
		r := explorerBounds(t, tt, "Explorer row "+path)
		p := tt.Image().RGBAAt(int(r.X+r.W-28), int(r.Y+r.H/2))
		want := a.Theme.Selection
		isBlue := p.R == want.R && p.G == want.G && p.B == want.B
		if isBlue != selected {
			t.Fatal("active-file selection paint", path, p, selected)
		}
	}
	assertActivePaint("docs", false)
	if err := tt.Click("Explorer row docs"); err != nil {
		t.Fatal(err)
	}
	if !a.files.expanded["docs"] || a.files.selected != 0 {
		t.Fatal("pointer did not select/expand directory")
	}
	tt.Key(0, ui.KeyLeft)
	if a.files.expanded["docs"] {
		t.Fatal("Left did not collapse selected directory")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyDown)
	if a.files.selected != 2 {
		t.Fatal("keyboard did not continue from pointer cursor", a.files.selected)
	}
	tt.Key(0, ui.KeyEnter)
	drainRegressionUpdates(t, a, func() bool {
		d, ok := a.files.editor.Active()
		return ok && d.Content.Path == filepath.Join(root, "alpha.go") && a.files.states[d.Content.Path] != nil
	})
	tt.Frame()
	if a.files.active != "alpha.go" {
		t.Fatal("opened file not active in tree", a.files.active)
	}
	assertActivePaint("alpha.go", true)
	assertActivePaint("docs", false)
	// Cursor movement is a quiet wash and must not move the open-file blue row.
	tree := explorerBounds(t, tt, "Explorer files")
	tt.ClickAt(tree.X+20, tree.Y+tree.H-40)
	tt.Key(0, ui.KeyDown)
	if a.files.selected != 3 || a.files.active != "alpha.go" {
		t.Fatal("cursor incorrectly activated a file", a.files.selected, a.files.active)
	}
	assertActivePaint("alpha.go", true)
	assertActivePaint("bravo.go", false)
	tt.Key(0, ui.KeySpace)
	drainRegressionUpdates(t, a, func() bool {
		d, ok := a.files.editor.Active()
		return ok && d.Content.Path == filepath.Join(root, "bravo.go") && a.files.states[d.Content.Path] != nil
	})
	tt.Frame()
	assertActivePaint("alpha.go", false)
	assertActivePaint("bravo.go", true)
	// Tab activation and programmatic opens reveal the active file, rather
	// than copying it into the independent keyboard cursor.
	a.OpenFile(filepath.Join(root, "docs", "readme.md"))
	drainRegressionUpdates(t, a, func() bool {
		d, ok := a.files.editor.Active()
		return ok && d.Content.Path == filepath.Join(root, "docs", "readme.md") && a.files.states[d.Content.Path] != nil
	})
	tt.Frame()
	if !a.files.expanded["docs"] || a.files.active != "docs/readme.md" {
		t.Fatal("active nested file was not revealed")
	}
	assertActivePaint("docs/readme.md", true)
	if err := tt.Click("alpha.go"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.files.active != "alpha.go" {
		t.Fatal("tab activation did not update Explorer")
	}
	assertActivePaint("alpha.go", true)
	assertActivePaint("docs/readme.md", false)
	// Source Left on a file does not move the cursor to the parent.
	if err := tt.RightClick("Explorer row docs/readme.md"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	selected := a.files.selected
	tt.Key(0, ui.KeyLeft)
	if a.files.selected != selected {
		t.Fatal("Left unexpectedly selected parent")
	}
	if err := tt.Click("Filter files"); err != nil {
		t.Fatal(err)
	}
	tt.Type("bravo")
	if !tt.HasText("5 files") {
		t.Fatal("filter changed total file count")
	}
	tt.Key(0, ui.KeyEscape)
	if a.files.query != "" {
		t.Fatal("Escape did not clear filter")
	}
}

func TestMountedExplorerWorkspaceChangeClearsCursorAndExpansion(t *testing.T) {
	a, tt, _ := mountedExplorer(t)
	if err := tt.Click("Explorer row docs"); err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("new workspace"), 0600); err != nil {
		t.Fatal(err)
	}
	a.Model.Workspace = root
	tt.Frame()
	if a.files.selected != -1 || len(a.files.expanded) != 0 || a.files.active != "" {
		t.Fatal("new workspace inherited tree navigation")
	}
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	tt.Frame()
	if _, ok := tt.Find("Explorer row new.txt"); !ok {
		t.Fatal("new workspace tree did not load")
	}
	if _, ok := tt.Find("Explorer row docs"); ok {
		t.Fatal("old workspace row survived")
	}
}
