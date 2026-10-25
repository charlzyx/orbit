package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func paletteDeleteFixture(t *testing.T) (*App, *ui.Tester, *[]Action) {
	t.Helper()
	a := sourceSidebarFixture(t)
	a.Model.SessionPath = "/store/first.jsonl"
	a.Model.SessionID = "first"
	a.Model.SessionName = "First fixture"
	a.Model.RuntimeID = "runtime-first"
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v) }
	a.paletteOpen = true
	a.paletteQuery = "delete remove session file disk"
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	return a, tt, &got
}
func armPaletteDelete(t *testing.T, a *App, tt *ui.Tester, keyboard bool) {
	t.Helper()
	if keyboard {
		tt.Key(0, ui.KeyEnter)
	} else if err := tt.Click(a.Tr("sidebar.delete_session_menu")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !a.confirmOpen {
		t.Fatal("missing explicit confirmation")
	}
}
func TestPaletteDeleteExactArmedTargetAndOneShot(t *testing.T) {
	for _, keyboard := range []bool{false, true} {
		t.Run(map[bool]string{false: "pointer", true: "keyboard"}[keyboard], func(t *testing.T) {
			a, tt, got := paletteDeleteFixture(t)
			armPaletteDelete(t, a, tt, keyboard)
			if !strings.Contains(a.confirmMessage, "First fixture") || !strings.Contains(a.confirmMessage, "/store/first.jsonl") {
				t.Fatal("confirmation must name captured victim", a.confirmMessage)
			}
			commit := a.confirmCommit
			s := a.Model
			s.SessionName = "Title changed only"
			a.SetSnapshot(s)
			tt.Frame()
			if err := tt.Click("Continue"); err != nil {
				t.Fatal(err)
			}
			tt.Frame()
			commit()
			if len(*got) != 1 || (*got)[0].Kind != "DeleteSession" || (*got)[0].Value != "/store/first.jsonl" || (*got)[0].SessionID != "first" {
				t.Fatal(*got)
			}
		})
	}
}
func TestPaletteDeleteSnapshotReplacementAndABACancelIntent(t *testing.T) {
	for _, dimension := range []string{"session", "path", "runtime", "workspace"} {
		t.Run(dimension, func(t *testing.T) {
			a, tt, got := paletteDeleteFixture(t)
			armPaletteDelete(t, a, tt, false)
			commit := a.confirmCommit
			original := a.Model
			s := original
			switch dimension {
			case "session":
				s.SessionID = "second"
			case "path":
				s.SessionPath = "/store/second.jsonl"
			case "runtime":
				s.RuntimeID = "runtime-second"
			case "workspace":
				s.Workspace = "/other/workspace"
			}
			a.SetSnapshot(s)
			a.SetSnapshot(original)
			tt.Frame()
			commit()
			if len(*got) != 0 || a.confirmOpen || a.paletteDelete != nil {
				t.Fatal("old confirmation survived owner ABA", *got)
			}
		})
	}
}
func TestPaletteDeleteCancelEscapeAndReplacement(t *testing.T) {
	for _, mode := range []string{"cancel", "escape", "replacement"} {
		t.Run(mode, func(t *testing.T) {
			a, tt, got := paletteDeleteFixture(t)
			armPaletteDelete(t, a, tt, false)
			commit := a.confirmCommit
			switch mode {
			case "cancel":
				if err := tt.Click("Cancel"); err != nil {
					t.Fatal(err)
				}
			case "escape":
				tt.Key(0, ui.KeyEscape)
			case "replacement":
				a.Confirm("Other confirmation", "No session deletion", func() {})
			}
			tt.Frame()
			commit()
			if len(*got) != 0 {
				t.Fatal("dismissed intent dispatched", mode, *got)
			}
		})
	}
}
func TestPaletteDeleteNoSavedPathDoesNotArm(t *testing.T) {
	a := fixture(t)
	a.command("DeleteSession", "")
	if a.confirmOpen || a.paletteDelete != nil {
		t.Fatal("empty path armed delete")
	}
}
func TestPaletteDeleteActualWorkspaceServicesInvalidateIntent(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	original := x.workspace
	a.Model.SessionPath = "/store/first.jsonl"
	a.Model.SessionID = "first"
	a.Model.RuntimeID = "runtime-first"
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v) }
	a.command("DeleteSession", "")
	commit := a.confirmCommit
	x.changeWorkspaceServices(t.TempDir())
	x.changeWorkspaceServices(original)
	commit()
	if len(got) != 0 || a.confirmOpen {
		t.Fatal("services ABA revived stale confirmation", got)
	}
	a.command("DeleteSession", "")
	commit = a.confirmCommit
	x.changeWorkspaceServices(original)
	commit()
	if len(got) != 1 || got[0].Value != "/store/first.jsonl" {
		t.Fatal("same workspace refresh lost valid intent", got)
	}
}

func TestPaletteDeleteHeldContinueCannotActivateReplacement(t *testing.T) {
	a, tt, got := paletteDeleteFixture(t)
	armPaletteDelete(t, a, tt, false)
	r, ok := tt.Find("Continue")
	if !ok {
		t.Fatal("missing continue")
	}
	px, py := r.X+r.W/2, r.Y+r.H/2
	tt.Press(px, py)
	next := a.Model
	next.SessionID = "second"
	next.SessionPath = "/store/second.jsonl"
	next.SessionName = "Second fixture"
	a.SetSnapshot(next)
	a.command("DeleteSession", "")
	tt.Frame()
	tt.Release(px, py)
	tt.Frame()
	if len(*got) != 0 {
		t.Fatal("old held press activated replacement target", *got)
	}
	if err := tt.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(*got) != 1 || (*got)[0].Value != "/store/second.jsonl" {
		t.Fatal("fresh replacement confirmation blocked", *got)
	}
}
