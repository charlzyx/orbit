package orbitui

import (
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"runtime"
	"testing"
)

func TestEmptySidebarDoesNotInventCwdProject(t *testing.T) {
	a := fixture(t)
	a.Model.Workspace = t.TempDir()
	if len(a.sidebarRows()) != 0 {
		t.Fatal("startup cwd invented saved project")
	}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if _, ok := tt.Find("No projects yet"); !ok {
		t.Fatal("fresh store missing source empty state")
	}
	a.Workspaces.Workspaces = append(a.Workspaces.Workspaces, preferences.Workspace{Path: a.Model.Workspace})
	if len(a.sidebarRows()) != 1 {
		t.Fatal("explicit saved workspace missing")
	}
}
func TestSourceNativeMenuDoesNotAllocatePersistentWindowStrip(t *testing.T) {
	o := sourceWindowOptions(1440, 920)
	if o.HideWindowButtons != (runtime.GOOS == "linux") {
		t.Fatal("native window buttons do not match source OS", o)
	}
	if !o.AutoHideMenuBar || o.MinWidth != 960 || o.MinHeight != 640 || !o.Maximized {
		t.Fatalf("source window options: %+v", o)
	}
}
func TestNativeMenuRespectsExternalPageModals(t *testing.T) {
	for _, page := range []string{"bug", "mcp", "skills", "plugins", "git", "worktrees"} {
		a := fixture(t)
		a.Options.HasModal = func() bool { return true }
		a.Draft = "preserved"
		calls := 0
		a.Options.OnAction = func(Action) { calls++ }
		a.nativeCommand("NewSession")
		a.nativeCommand("OpenSettings")
		a.nativeCommand("ToggleTerminal")
		if calls != 0 || a.Draft != "preserved" || a.page != "chat" {
			t.Fatal("native action escaped page modal", page)
		}
	}
}

func TestSourceHeaderNoOpenInWithoutAppsAndInitialComposerFocus(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("Open in…"); ok {
		t.Fatal("invented Open-in control without discovered apps")
	}
	if !tt.Focused("Message") {
		t.Fatal("initial composer is not focused")
	}
	if e := tt.Click("Review changes"); e != nil {
		t.Fatal(e)
	}
	if !a.ReviewVisible || a.ExplorerVisible {
		t.Fatal("source header dock button did not open Review")
	}
}

func TestSourceMainHeaderUsesTwentyPixelInset(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	r, ok := tt.Find(a.text("settings.new_task", "New task"))
	main, found := tt.Find("New task backdrop")
	if !ok || !found || r.X-main.X != 20 {
		t.Fatal("source header title inset", r, main)
	}
}

func TestSourceSidebarIncludesDividerAndPersistsResize(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	main, ok := tt.Find("New task backdrop")
	if !ok || main.X != 248 {
		t.Fatal("source main origin includes extra divider", main)
	}
	div, ok := tt.Find("Divider")
	if !ok || div.X != 247 || div.W != 1 {
		t.Fatal("source divider origin", div)
	}
	if e := tt.Click("Divider"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	tt.Frame()
	if a.SidebarWidth <= 248 || a.Options.Store.Layout().SidebarWidth != float64(a.SidebarWidth) {
		t.Fatal("native split keyboard resize did not persist source outer width", a.SidebarWidth, a.Options.Store.Layout())
	}
	main, _ = tt.Find("New task backdrop")
	if main.X != a.SidebarWidth {
		t.Fatal("resized divider counted twice", main, a.SidebarWidth)
	}
	tt.Key(0, ui.KeyLeft)
	tt.Frame()
	tt.Frame()
	if a.SidebarWidth != 248 {
		t.Fatal("reverse resize drift", a.SidebarWidth)
	}
}
