package orbitui

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestWorkspacePickerSourceRecentOrderingAndFilter(t *testing.T) {
	a := fixture(t)
	a.Model.Workspace = "/projects/current"
	now := time.Now()
	for i := 0; i < 12; i++ {
		a.Model.Sessions = append(a.Model.Sessions, SessionRow{ID: fmt.Sprint(i), Workspace: fmt.Sprintf("/projects/folder-%02d", i), Updated: now.Add(time.Duration(i) * time.Minute)})
	}
	a.Model.Sessions = append(a.Model.Sessions, SessionRow{Workspace: "/projects/folder-11", Updated: now.Add(time.Hour)})
	rows := a.recentWorkspaces()
	if len(rows) != 8 || rows[0].Path != a.Model.Workspace || rows[1].Path != "/projects/folder-11" || !rows[0].Updated.IsZero() {
		t.Fatalf("source recents: %+v", rows)
	}
	a.workspacePicker.entries = rows
	a.workspacePicker.query = "FOLDER-10"
	if got := a.workspacePicker.filtered(); len(got) != 1 || got[0].Name != "folder-10" {
		t.Fatal(got)
	}
	for _, v := range []struct{ w, want float32 }{{1000, 400}, {448, 400}, {300, 252}, {100, 240}} {
		if got := workspaceFieldWidth(v.w); got != v.want {
			t.Fatal(v, got)
		}
	}
}
func TestNativeWorkspacePickerKeyboardFilteringNoopAndFocus(t *testing.T) {
	a := fixture(t)
	a.Model.Workspace = "/projects/current"
	a.Model.Sessions = []SessionRow{{Workspace: "/projects/another", Updated: time.Now()}}
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Click("Message")
	for i := 0; i < 3; i++ {
		if e := tt.Click("Choose workspace"); e != nil {
			t.Fatal(e)
		}
		tt.Frame()
		if !a.workspaceOpen {
			t.Fatal("workspace popup did not open")
		}
		tt.Key(0, ui.KeyEscape)
		tt.Frame()
		if a.workspaceOpen || !tt.Focused("Message") {
			t.Fatal("Escape did not restore composer")
		}
	}
	tt.Click("Choose workspace")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(actions) != 0 || a.workspaceOpen {
		t.Fatal("current workspace must dismiss without creating a session", actions)
	}
	tt.Click("Choose workspace")
	tt.Type("another")
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "SetWorkspace" || actions[0].Value != "/projects/another" {
		t.Fatal("filtered workspace was not selected", actions)
	}
	tt.Click("Choose workspace")
	tt.Type("no-match")
	tt.Frame()
	if !tt.HasText(a.Tr("workspace_picker.no_folders_match")) {
		t.Fatal("source empty result missing")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(actions) != 2 || actions[1].Kind != "BrowseWorkspace" {
		t.Fatal("zero-result Enter must browse", actions)
	}
}
func TestNativeWorkspacePickerWrapModalGuardAndBounds(t *testing.T) {
	a := fixture(t)
	a.Model.Workspace = "/projects/current"
	a.Model.Sessions = []SessionRow{{Workspace: "/projects/other", Updated: time.Now()}}
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 960, 640)
	tt.Click("Choose workspace")
	tt.Frame()
	r, ok := tt.Find("Recent workspace picker")
	if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 960 || r.Y+r.H > 640 {
		t.Fatalf("popup out of source minimum bounds: %+v", r)
	}
	a.nativeCommand("NewSession")
	if len(actions) > 0 {
		t.Fatal("native command acted behind workspace picker")
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "BrowseWorkspace" {
		t.Fatal("Up must wrap to browse", actions)
	}
	tt.Click("Choose workspace")
	tt.Frame()
	tt.ClickAt(300, 80)
	tt.Frame()
	if a.workspaceOpen {
		t.Fatal("outside press did not dismiss")
	}
	if a.workspacePicker.dismissedAt.IsZero() {
		t.Fatal("pointer dismissal did not establish source200ms guard")
	}
	a.toggleWorkspacePicker(a.workspacePicker.dismissedAt.Add(199 * time.Millisecond))
	if a.workspaceOpen {
		t.Fatal("same-gesture guard reopened popup")
	}
	a.toggleWorkspacePicker(a.workspacePicker.dismissedAt.Add(200 * time.Millisecond))
	if !a.workspaceOpen {
		t.Fatal("guard prevented later opening")
	}
}
func TestSourceSidebarWordmarkAndNewTaskGeometry(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	r, ok := tt.Find("Sidebar wordmark")
	if !ok || r.W != 76 || r.Y != 46 || r.X != 86 {
		t.Fatal("source wordmark width/top", r)
	}
	b, ok := tt.Find(a.text("sidebar.new_session", "New Task"))
	if !ok || b.H != 32 || b.X != 12 {
		t.Fatal("source primary button metrics", b)
	}
	if !tt.HasText(shortcutLabel("secondary-n")) {
		t.Fatal("source New Task shortcut absent")
	}
}
