package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func menuAuditSnapshot() Snapshot {
	return Snapshot{RuntimeID: "runtime-a", SessionID: "session-a", SessionPath: "session-a.jsonl", Workspace: "/work/a", Messages: []Message{{ID: "first", Role: "assistant", Text: "FIRST_MESSAGE"}, {ID: "second", Role: "assistant", Text: "SECOND_MESSAGE"}}}
}
func TestMessageMenuAuditCurrentHit(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(menuAuditSnapshot())
	tt := ui.NewTester(a.View, 1470, 922)
	for _, want := range []string{"FIRST_MESSAGE", "SECOND_MESSAGE", "FIRST_MESSAGE"} {
		if err := tt.RightClick(want); err != nil {
			t.Fatal(err)
		}
		if err := tt.ChooseMenuItem("Copy Message"); err != nil {
			t.Fatal(err)
		}
		if tt.Clipboard() != want {
			t.Fatalf("copy=%q want=%q", tt.Clipboard(), want)
		}
	}
}
func TestMessageMenuAuditOldOwner(t *testing.T) {
	for _, kind := range []string{"runtime", "runtime ABA", "session", "workspace", "close"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			s := menuAuditSnapshot()
			a.SetSnapshot(s)
			tt := ui.NewTester(a.View, 1470, 922)
			if err := tt.RightClick("FIRST_MESSAGE"); err != nil {
				t.Fatal(err)
			}
			tt.SetClipboard("sentinel")
			switch kind {
			case "runtime":
				s.RuntimeID = "runtime-b"
			case "runtime ABA":
				s.RuntimeID = "runtime-b"
				a.SetSnapshot(s)
				s.RuntimeID = "runtime-a"
			case "session":
				s.SessionID = "session-b"
			case "workspace":
				s.Workspace = "/work/b"
			case "close":
				a.closeInlineImages()
			}
			s.Messages = []Message{{ID: "first", Role: "assistant", Text: "NEW_MESSAGE"}}
			a.SetSnapshot(s)
			tt.Frame()
			_ = tt.ChooseMenuItem("Copy Message")
			if tt.Clipboard() != "sentinel" {
				t.Fatalf("old menu retargeted %q", tt.Clipboard())
			}
			if kind != "close" {
				if err := tt.RightClick("NEW_MESSAGE"); err != nil {
					t.Fatal(err)
				}
				if err := tt.ChooseMenuItem("Copy Message"); err != nil {
					t.Fatal(err)
				}
				if tt.Clipboard() != "NEW_MESSAGE" {
					t.Fatal("fresh owner menu broken", tt.Clipboard())
				}
			}
		})
	}
}
func TestMessageMenuAuditPointerMovesIntoNativeMenu(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(menuAuditSnapshot())
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.RightClick("FIRST_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	tt.Move(1460, 10)
	tt.Frame()
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "FIRST_MESSAGE" {
		t.Fatal(tt.Clipboard())
	}
}
func TestMessageMenuAuditActualServiceTaskAndClose(t *testing.T) {
	for _, kind := range []string{"service ABA", "new task", "close"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			a := x.UI
			s := menuAuditSnapshot()
			s.Workspace = x.workspace
			a.SetSnapshot(s)
			tt := ui.NewTester(a.View, 1470, 922)
			if err := tt.RightClick("FIRST_MESSAGE"); err != nil {
				t.Fatal(err)
			}
			tt.SetClipboard("sentinel")
			switch kind {
			case "service ABA":
				p := x.workspace
				x.changeWorkspaceServices(t.TempDir())
				x.changeWorkspaceServices(p)
			case "new task":
				x.Action(Action{Kind: "NewSession"})
			case "close":
				x.Close()
			}
			tt.Frame()
			_ = tt.ChooseMenuItem("Copy Message")
			if tt.Clipboard() != "sentinel" {
				t.Fatal("retired actual owner copied", tt.Clipboard())
			}
		})
	}
}
