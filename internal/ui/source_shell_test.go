package orbitui

import (
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
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
