package orbitui

import (
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func renameAuditFixture(t *testing.T) (*App, *ui.Tester, *[]Action) {
	t.Helper()
	a := sourceSidebarFixture(t)
	a.Model.SessionPath = "/synthetic/first.jsonl"
	a.Model.SessionID = "first"
	a.Model.RuntimeID = "runtime-first"
	a.Model.SessionName = "Displayed first task"
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	a.paletteOpen = true
	a.paletteQuery = "rename title name session"
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if err := tt.Click(a.Tr("command_palette.rename_session")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.promptOpen || a.promptValue != "Displayed first task" {
		t.Fatal("wrong displayed target", a.promptValue)
	}
	return a, tt, &actions
}
func TestPaletteRenameAuditCapturedTarget(t *testing.T) {
	a, tt, got := renameAuditFixture(t)
	a.promptValue = "Renamed first task"
	tt.Frame()
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(*got) != 1 || (*got)[0].Data["path"] != "/synthetic/first.jsonl" || (*got)[0].SessionID != "first" {
		t.Fatal("rename lacks displayed target identity", *got)
	}
}
func TestPaletteRenameAuditOwnerSwitchAndABA(t *testing.T) {
	for _, dimension := range []string{"session", "path", "runtime", "workspace"} {
		for _, aba := range []bool{false, true} {
			t.Run(dimension+map[bool]string{false: " switch", true: " ABA"}[aba], func(t *testing.T) {
				a, tt, got := renameAuditFixture(t)
				original := a.Model
				next := original
				switch dimension {
				case "session":
					next.SessionID = "second"
				case "path":
					next.SessionPath = "/synthetic/second.jsonl"
				case "runtime":
					next.RuntimeID = "runtime-second"
				case "workspace":
					next.Workspace = "/synthetic/second-workspace"
				}
				a.promptValue = "Rename intended for first"
				r, ok := tt.Find("Save")
				if !ok {
					t.Fatal("missing Save")
				}
				x, y := r.X+r.W/2, r.Y+r.H/2
				tt.Press(x, y)
				a.SetSnapshot(next)
				if aba {
					a.SetSnapshot(original)
				}
				tt.Frame()
				tt.Release(x, y)
				tt.Frame()
				if len(*got) != 0 {
					t.Fatal("held old prompt submitted after owner replacement", *got)
				}
			})
		}
	}
}

func TestPaletteRenameAuditDismissReplacementAndOneShot(t *testing.T) {
	for _, mode := range []string{"cancel", "escape", "replacement", "enter"} {
		t.Run(mode, func(t *testing.T) {
			a, tt, got := renameAuditFixture(t)
			commit := a.promptCommit
			a.promptValue = "Intended new name"
			tt.Frame()
			switch mode {
			case "cancel":
				if err := tt.Click("Cancel"); err != nil {
					t.Fatal(err)
				}
			case "escape":
				tt.Key(0, ui.KeyEscape)
			case "replacement":
				a.Prompt("Other prompt", "other", func(string) {})
			case "enter":
				next := a.Model
				next.SessionName = "Model title changed only"
				a.SetSnapshot(next)
				tt.Key(0, ui.KeyEnter)
			}
			tt.Frame()
			commit("Late replay")
			want := 0
			if mode == "enter" {
				want = 1
			}
			if len(*got) != want {
				t.Fatal("stale/repeated prompt callback", mode, *got)
			}
			if want == 1 && ((*got)[0].Value != "Intended new name" || (*got)[0].Data["path"] != "/synthetic/first.jsonl") {
				t.Fatal(*got)
			}
		})
	}
}
func TestPaletteRenameAuditHeldSaveCannotActivateReplacement(t *testing.T) {
	a, tt, got := renameAuditFixture(t)
	r, ok := tt.Find("Save")
	if !ok {
		t.Fatal("missing Save")
	}
	x, y := r.X+r.W/2, r.Y+r.H/2
	tt.Press(x, y)
	next := a.Model
	next.SessionID = "second"
	next.SessionPath = "/synthetic/second.jsonl"
	next.SessionName = "Displayed second task"
	a.SetSnapshot(next)
	a.command("RenameSession", "")
	tt.Frame()
	tt.Release(x, y)
	tt.Frame()
	if len(*got) != 0 {
		t.Fatal("old press activated replacement", *got)
	}
	if !a.promptOpen {
		t.Fatal("replacement should remain open")
	}
	a.promptValue = "Renamed second task"
	tt.Frame()
	if err := tt.Click("Save"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(*got) != 1 || (*got)[0].Data["path"] != "/synthetic/second.jsonl" {
		t.Fatal(*got)
	}
}
func TestPaletteRenameAuditServiceABAAndPathless(t *testing.T) {
	x := workspaceDialogFixture(t)
	x.UI.Options.CaptureSessionRenameTarget = nil // UI-only emitted-action fixture.
	a := x.UI
	original := x.workspace
	a.Model.SessionPath = "/synthetic/first.jsonl"
	a.Model.SessionID = "first"
	a.Model.RuntimeID = "runtime-first"
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v) }
	a.command("RenameSession", "")
	commit := a.promptCommit
	x.changeWorkspaceServices(t.TempDir())
	x.changeWorkspaceServices(original)
	commit("stale")
	if len(got) != 0 || a.promptOpen {
		t.Fatal("service ABA survived", got)
	}
	a.command("RenameSession", "")
	commit = a.promptCommit
	x.changeWorkspaceServices(original)
	commit("valid")
	if len(got) != 1 || got[0].Data["path"] != "/synthetic/first.jsonl" {
		t.Fatal("same-workspace refresh lost valid target", got)
	}
	a.Model.SessionPath = ""
	a.promptOpen = false
	a.command("RenameSession", "")
	if a.promptOpen || a.paletteRename != nil {
		t.Fatal("pathless rename prompt armed")
	}
}

func TestPaletteRenameAuditStaleUISnapshotUsesCapturedSavedPath(t *testing.T) {
	x, _, nextWorkspace, _ := workspaceMenuController(t)
	a := x.UI
	a.Options.CaptureSessionRenameTarget = nil // Cold saved-path route, not live target capture.
	// This is a synthetic JSONL child, not Pi. Missing original file makes
	// metadata routing observable before any saved-session Rename RPC.
	original := filepath.Join(t.TempDir(), "displayed-old-task.jsonl")
	a.Model.SessionPath = original
	a.Model.SessionID = "displayed-old"
	a.Model.RuntimeID = "displayed-runtime"
	a.Model.SessionName = "Old displayed name"
	a.command("RenameSession", "")
	commit := a.promptCommit
	if err := x.Agent.Open(nextWorkspace, ""); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		s := x.Agent.Snapshot()
		if s.Connected && s.ActiveSession.Workspace == nextWorkspace {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if s := x.Agent.Snapshot(); !s.Connected || s.ActiveSession.Workspace != nextWorkspace || a.Model.SessionPath != original {
		t.Fatal("fixture did not separate ready manager from stale UI")
	}
	commit("Rename old displayed task")
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		x.Agent.Tick()
		s := x.Agent.Snapshot()
		if strings.Contains(s.ConnectionError, original) {
			if s.ActiveSession.Workspace != nextWorkspace || s.ActiveSession.Name == "Rename old displayed task" {
				t.Fatal("replacement task changed", s.ActiveSession)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("captured file admission was bypassed for manager's current task", x.Agent.Snapshot().ConnectionError)
}

func TestPaletteRenameAuditControllerRejectsMissingCapturedPath(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	a.Model.SessionPath = "/synthetic/current-must-not-be-inferred.jsonl"
	manager := x.Agent
	x.Agent = nil
	defer func() { x.Agent = manager }()
	// RenameSavedSession's empty-path guard returns before using a manager;
	// any fallback to current-runtime Rename would dereference the nil manager.
	x.Action(Action{Kind: "RenameSession", Value: "must not target current"})
	x.Agent = manager // Restore services before mounting unrelated composer UI.
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if !tt.HasText("session path is required") {
		t.Fatal("missing explicit path error")
	}
}
