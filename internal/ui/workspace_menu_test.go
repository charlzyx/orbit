package orbitui

import (
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkspaceHeaderSourceActions(t *testing.T) {
	a := sourceSidebarFixture(t)
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy path"); err != nil {
		t.Fatal(err)
	}
	if got := tt.Clipboard(); got != "/work/alpha" {
		t.Fatal("wrong workspace path", got)
	}
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Icon & color"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find("Workspace appearance"); !ok {
		t.Fatal("missing appearance picker")
	}
}

func awaitWorkspaceEdit(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for a.workspaceActions.busy && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
		tt.Frame()
	}
	if a.workspaceActions.busy {
		t.Fatal("workspace edit did not complete")
	}
}
func TestWorkspaceAppearancePersistsAndReset(t *testing.T) {
	a := sourceSidebarFixture(t)
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick("alpha"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Icon & color"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Workspace icon rocket-01"); err != nil {
		t.Fatal(err)
	}
	awaitWorkspaceEdit(t, a, tt)
	if err := tt.Click("Workspace tint blue"); err != nil {
		t.Fatal(err)
	}
	awaitWorkspaceEdit(t, a, tt)
	w := a.workspaceEntry("/work/alpha")
	if w.Icon != "rocket-01" || w.Tint != "blue" {
		t.Fatal(*w)
	}
	saved := a.Options.Store.Workspaces()
	if saved.Workspaces[1].Icon != w.Icon || saved.Workspaces[1].Tint != w.Tint {
		t.Fatal(saved)
	}
	if err := tt.Click("Reset"); err != nil {
		t.Fatal(err)
	}
	awaitWorkspaceEdit(t, a, tt)
	w = a.workspaceEntry("/work/alpha")
	if w.Icon != "" || w.Tint != "" {
		t.Fatal("reset", *w)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.workspaceActions.open {
		t.Fatal("Escape did not close picker")
	}
}
func TestWorkspaceRemoveOnlyOwnListAndCanReadd(t *testing.T) {
	a := sourceSidebarFixture(t)
	root := t.TempDir()
	path := filepath.Join(root, "saved.jsonl")
	if err := os.WriteFile(path, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	a.Workspaces.Workspaces = []preferences.Workspace{{Path: root, AddedAt: 123}}
	a.Model.Workspace = root
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	label := filepath.Base(root)
	a.sidebarState.collapsedWorkspaces = map[string]bool{label: true}
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.RightClick(label); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Remove from sidebar"); err != nil {
		t.Fatal(err)
	}
	awaitWorkspaceEdit(t, a, tt)
	if len(a.Workspaces.Workspaces) != 0 || len(a.Options.Store.Workspaces().Workspaces) != 0 {
		t.Fatal("project still listed")
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != "untouched" {
		t.Fatal("remove touched session", err)
	}
	if a.Model.Workspace != root || a.sidebarState.collapsedWorkspaces[label] {
		t.Fatal("remove changed active task or retained group state")
	}
	if err := a.Options.Store.AddWorkspace(root); err != nil {
		t.Fatal(err)
	}
	if len(a.Options.Store.Workspaces().Workspaces) != 1 {
		t.Fatal("re-add failed")
	}
}
func TestWorkspaceMenuStaleNativeChoice(t *testing.T) {
	for _, mode := range []string{"roundtrip", "services-roundtrip", "replace", "same-basename"} {
		t.Run(mode, func(t *testing.T) {
			a := sourceSidebarFixture(t)
			tt := ui.NewTester(a.View, 1222, 878)
			tt.Frame()
			if err := tt.RightClick("alpha"); err != nil {
				t.Fatal(err)
			}
			switch mode {
			case "roundtrip":
				s := a.Model
				s.Workspace = "/work/beta"
				a.SetSnapshot(s)
				s.Workspace = "/work/alpha"
				a.SetSnapshot(s)
			case "services-roundtrip":
				a.invalidateWorkspaceMenu("/work/beta")
				a.Model.Workspace = "/work/beta"
				a.invalidateWorkspaceMenu("/work/alpha")
				a.Model.Workspace = "/work/alpha"
			case "replace":
				a.Workspaces.Workspaces = append([]preferences.Workspace(nil), a.Workspaces.Workspaces...)
			case "same-basename":
				a.Workspaces.Workspaces[1].Path = "/other/alpha"
			}
			_ = tt.ChooseMenuItem("Copy path")
			if tt.Clipboard() != "" {
				t.Fatal("stale choice retargeted", tt.Clipboard())
			}
			tt.CloseMenu()
			tt.Frame()
			if err := tt.RightClick("alpha"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("Copy path"); err != nil {
				t.Fatal(err)
			}
			want := "/work/alpha"
			if mode == "same-basename" {
				want = "/other/alpha"
			}
			if tt.Clipboard() != want {
				t.Fatal("fresh choice", tt.Clipboard())
			}
		})
	}
}
func TestWorkspaceEditCompletionDoesNotMutateReplacement(t *testing.T) {
	a := sourceSidebarFixture(t)
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	old := a.workspaceEntry("/work/alpha")
	a.editWorkspace(old, false, "rocket-01", "blue")

	a.Workspaces.Workspaces = append([]preferences.Workspace(nil), a.Workspaces.Workspaces...)
	a.workspaceEntry("/work/alpha").Icon = "star"
	deadline := time.Now().Add(3 * time.Second)
	for a.workspaceActions.busy && time.Now().Before(deadline) {
		a.drainUpdates()
		time.Sleep(time.Millisecond)
	}
	if a.workspaceActions.busy {
		t.Fatal("own workspace edit did not complete")
	}

	if a.workspaceEntry("/work/alpha").Icon != "star" {
		t.Fatal("stale completion replaced new UI owner")
	}
}
func TestWorkspaceStatusMarks(t *testing.T) {
	a := sourceSidebarFixture(t)
	for _, tc := range []struct {
		w    preferences.Workspace
		icon string
	}{{preferences.Workspace{Path: "/work/alpha"}, "folder-open"}, {preferences.Workspace{Path: "/work/empty"}, "folder"}, {preferences.Workspace{Path: "/work/alpha", Icon: "star"}, "star"}, {preferences.Workspace{Path: "/work/alpha", Icon: "../../escape", Tint: "invalid"}, "folder-open"}} {
		icon, _ := a.workspaceMark(tc.w)
		if icon != "icons/"+tc.icon+".svg" {
			t.Fatal(icon, tc)
		}
	}
	a.Model.Workspace = "elsewhere"
	w := preferences.Workspace{Path: "/work/alpha"}
	icon, _ := a.workspaceMark(w)
	if icon != "icons/folder-sync.svg" {
		t.Fatal(icon)
	}
	a.Model.Sessions = []SessionRow{{Workspace: w.Path, Archived: true}}
	icon, _ = a.workspaceMark(w)
	if icon != "icons/folder-archive.svg" {
		t.Fatal(icon)
	}
}

func TestWorkspaceActionsTriggerAndHiddenPickerDismissal(t *testing.T) {
	for _, mode := range []string{"sidebar", "settings", "grouping"} {
		t.Run(mode, func(t *testing.T) {
			a := sourceSidebarFixture(t)
			tt := ui.NewTester(a.View, 1222, 878)
			tt.Frame()
			if err := tt.Click("Workspace actions /work/alpha"); err != nil {
				t.Fatal(err)
			}
			if err := tt.ChooseMenuItem("Icon & color"); err != nil {
				t.Fatal(err)
			}
			if !a.workspaceActions.open {
				t.Fatal("clickable menu did not open appearance")
			}
			switch mode {
			case "sidebar":
				a.SidebarVisible = false
			case "settings":
				a.Navigate("settings")
			case "grouping":
				a.Workspaces.GroupBy = "one_list"
			}
			tt.Frame()
			if a.workspaceActions.open {
				t.Fatal("hidden picker kept modal shortcut guard")
			}
		})
	}
}
func TestWorkspaceEditFailureKeepsProject(t *testing.T) {
	a := sourceSidebarFixture(t)
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	// A later same-process preferences write invalidates this snapshot. Failure
	// refreshes the authoritative list but does not remove the selected project.
	if err := a.Options.Store.SaveWorkspaces(a.Options.Store.Workspaces()); err != nil {
		t.Fatal(err)
	}
	w := a.workspaceEntry("/work/alpha")
	a.editWorkspace(w, true, "", "")
	awaitWorkspaceEdit(t, a, tt)
	if a.workspaceEntry(w.Path) == nil {
		t.Fatal("failed write removed UI project")
	}
}

func TestWorkspaceFlatRowUsesOwnStatus(t *testing.T) {
	a := sourceSidebarFixture(t)
	for _, tc := range []struct {
		s    SessionRow
		icon string
	}{{SessionRow{Workspace: "/work/alpha"}, "folder"}, {SessionRow{Workspace: "/work/alpha", Running: true}, "folder-sync"}, {SessionRow{Workspace: "/work/alpha", Archived: true}, "folder-archive"}, {SessionRow{Workspace: "/work/alpha", Path: a.Model.SessionPath}, "folder-open"}} {
		icon, _ := a.sessionWorkspaceMark(tc.s)
		if icon != "icons/"+tc.icon+".svg" {
			t.Fatal(icon, tc)
		}
	}
	a.workspaceEntry("/work/alpha").Icon = "star"
	icon, _ := a.sessionWorkspaceMark(SessionRow{Workspace: "/work/alpha", Running: true})
	if icon != "icons/star.svg" {
		t.Fatal("custom mark lost", icon)
	}
}

func TestWorkspaceMalformedStoreFailureKeepsVisibleProjects(t *testing.T) {
	a := sourceSidebarFixture(t)
	if err := a.Options.Store.SaveWorkspaces(a.Workspaces); err != nil {
		t.Fatal(err)
	}
	a.Workspaces = a.Options.Store.Workspaces()
	if err := os.WriteFile(filepath.Join(a.Options.Store.RootDir, "workspaces.json"), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	old := a.workspaceEntry("/work/alpha")
	a.editWorkspace(old, true, "", "")
	awaitWorkspaceEdit(t, a, tt)
	if a.workspaceEntry(old.Path) != old || len(a.Workspaces.Workspaces) != 3 {
		t.Fatal("unreadable store replaced UI with empty project list")
	}
}
