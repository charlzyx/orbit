package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestNewTaskMountedMaintenanceKeepsDraftAndAttachments(t *testing.T) {
	for _, entry := range []string{"sidebar", "shortcut", "native", "palette"} {
		t.Run(entry, func(t *testing.T) {
			x := maintenanceFixture(t)
			a := x.UI
			a.Options.OnAction = x.Action
			a.Draft = "retained task text"
			a.attachments = []attachment{{ID: "kept-image", Name: "image.png"}}
			release, err := x.beginRuntimeOperation()
			if err != nil {
				t.Fatal(err)
			}
			defer release()
			tt := ui.NewTester(a.View, 1200, 800)
			switch entry {
			case "sidebar":
				if err := tt.Click(a.text("menu.new_task", "New Task")); err != nil {
					t.Fatal(err)
				}
			case "shortcut":
				tt.Key(ui.Cmd, ui.KeyN)
			case "native":
				a.nativeCommand("NewSession")
				tt.Frame()
			case "palette":
				a.paletteOpen = true
				a.paletteQuery = "new session"
				tt.Frame()
				tt.Key(0, ui.KeyEnter)
			}
			if a.Draft != "retained task text" || len(a.attachments) != 1 || a.attachments[0].ID != "kept-image" || a.pendingNew {
				t.Fatal("rejected New Task discarded local intent", a.Draft, a.attachments, a.pendingNew)
			}
		})
	}
}

func TestNewTaskDefaultNoopAndControllerErrorKeepLocalIntent(t *testing.T) {
	for _, kind := range []string{"absent", "noop", "closed"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			switch kind {
			case "absent":
				a.Options.OnAction = nil
			case "noop":
				a.Options.OnAction = func(Action) {}
			case "closed":
				x := maintenanceFixture(t)
				a = x.UI
				a.Options.OnAction = x.Action
				x.Agent.Close()
			}
			a.Draft = "keep"
			a.attachments = []attachment{{ID: "kept"}}
			a.Navigate("settings")
			a.nativeCommand("NewSession")
			if a.Draft != "keep" || len(a.attachments) != 1 || a.attachments[0].ID != "kept" || a.pendingNew || a.page != "settings" {
				t.Fatal("unadmitted intent changed local state", kind, a.Draft, a.pendingNew, a.page)
			}
		})
	}
}
func TestNewTaskPositiveHostAdmissionClearsOnlyOriginalDraft(t *testing.T) {
	for _, kind := range []string{"accepted", "new-task", "new-text", "new-attachment"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			a.Draft = "old draft"
			a.attachments = []attachment{{ID: "old"}}
			oldID := a.Model.SessionID
			a.Options.OnAction = func(action Action) {
				switch kind {
				case "new-task":
					next := a.Model
					next.RuntimeID = "other-runtime"
					next.SessionID = "other-session"
					a.SetSnapshot(next)
					a.Draft = "new draft"
				case "new-text":
					a.Draft = "new draft"
				case "new-attachment":
					a.attachments = []attachment{{ID: "new"}}
				}
				action.acceptNewTask()
			}
			a.command("NewSession", "")
			if kind == "accepted" {
				if a.Draft != "" || len(a.attachments) != 0 || !a.pendingNew || a.drafts[oldID].Text != "old draft" || len(a.drafts[oldID].Attachments) != 1 {
					t.Fatal("accepted task did not preserve old draft before clearing")
				}
			} else {
				if a.pendingNew {
					t.Fatal("old admission overwrote newer task presentation")
				}
				if (kind == "new-task" || kind == "new-text") && a.Draft != "new draft" {
					t.Fatal("accepted old intent cleared newer draft", a.Draft)
				}
				if kind == "new-attachment" && (len(a.attachments) != 1 || a.attachments[0].ID != "new") {
					t.Fatal("accepted old intent cleared newer attachment")
				}
			}
		})
	}
}
func TestNewTaskActualControllerAdmissionIsNotRemoteCompletion(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	a.Draft = "keep in prior draft"
	a.attachments = []attachment{{ID: "kept-before-new"}}
	before := x.Agent.CaptureWorkflowTarget()
	oldID := a.Model.SessionID
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Key(ui.Cmd, ui.KeyN)
	if before == x.Agent.CaptureWorkflowTarget() {
		t.Fatal("actual host did not accept fresh task")
	}
	if a.Draft != "" || len(a.attachments) != 0 || !a.pendingNew || a.drafts[oldID].Text != "keep in prior draft" {
		t.Fatal("positive actual admission not applied", a.Draft, a.pendingNew)
	}
	// The fixture launcher is deliberately nonexistent: acceptance is task intent,
	// never a claim that this process/RPC subsequently connected successfully.
}

func TestNewTaskLateAcknowledgementCannotClearLaterDraft(t *testing.T) {
	a := fixture(t)
	a.Draft = "old draft"
	ready, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	a.Options.OnAction = func(action Action) { go func() { close(ready); <-release; action.acceptNewTask(); close(done) }() }
	a.command("NewSession", "")
	<-ready
	if a.Draft != "old draft" || a.pendingNew {
		t.Fatal("unacknowledged asynchronous callback cleared draft")
	}
	a.Draft = "newer local edit"
	close(release)
	<-done
	if a.Draft != "newer local edit" || a.pendingNew {
		t.Fatal("late admission changed unrelated later draft")
	}
}
