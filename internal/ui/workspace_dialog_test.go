package orbitui

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func workspaceDialogFixture(t *testing.T) *Controller {
	t.Helper()
	home := t.TempDir()
	x, err := NewController(RunOptions{Home: home, Workspace: t.TempDir(), PiBinary: filepath.Join(home, "missing-synthetic-pi"), Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	return x
}

func holdWorkspaceDialog(t *testing.T, x *Controller, path string, err error) func() {
	t.Helper()
	started, release := make(chan struct{}), make(chan struct{})
	released := false
	unblock := func() {
		if !released {
			released = true
			close(release)
		}
	}
	t.Cleanup(unblock)
	x.workspaceBrowse.open = func(parent *mygo.Window, current string) (string, error) {
		close(started)
		<-release
		return path, err
	}
	returned := make(chan struct{})
	go func() { x.Action(Action{Kind: "BrowseWorkspace"}); close(returned) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("native picker never started")
	}
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		unblock()
		<-returned
		t.Fatal("BrowseWorkspace blocked its action/view caller while native sheet was open")
	}
	return unblock
}

func TestWorkspaceDialogReturnsBeforeAnswerAndAppliesOnUIQueue(t *testing.T) {
	x := workspaceDialogFixture(t)
	target := filepath.Join(t.TempDir(), "含 空格的项目")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	initial := x.workspace
	release := holdWorkspaceDialog(t, x, target, nil)
	// Production UI can render again while the native answer is outstanding.
	tt := ui.NewTester(x.View, 1470, 922)
	tt.Frame()
	// Repeated click while the first sheet is pending must not invoke open again.
	x.Action(Action{Kind: "BrowseWorkspace"})
	if !x.workspaceBrowse.pending || x.workspace != initial {
		t.Fatal("pending sheet changed workspace")
	}
	release()
	// Native completion only queues an update, never mutates the controller.
	if x.workspace != initial {
		t.Fatal("worker changed workspace outside UI queue")
	}
	drainControllerDialogUpdates(t, x, func() bool { return !x.workspaceBrowse.pending })
	if x.workspace != target || x.workspaceBrowse.pending {
		t.Fatal("selected path did not apply on UI queue", x.workspace)
	}
}

func TestWorkspaceDialogCancelAndErrorPermitReopen(t *testing.T) {
	for _, err := range []error{nil, errors.New("synthetic picker failure")} {
		t.Run(map[bool]string{false: "cancel", true: "error"}[err != nil], func(t *testing.T) {
			x := workspaceDialogFixture(t)
			initial := x.workspace
			answer := ""
			if err != nil {
				answer = t.TempDir()
			}
			release := holdWorkspaceDialog(t, x, answer, err)
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.workspaceBrowse.pending })
			if x.workspace != initial {
				t.Fatal("cancel/error changed workspace")
			}
			release = holdWorkspaceDialog(t, x, "", nil)
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.workspaceBrowse.pending })
		})
	}
}

func TestWorkspaceDialogRejectsLateAnswerAfterNewIntent(t *testing.T) {
	for _, kind := range []string{"workspace", "workspace-roundtrip", "new-session", "session-open", "window-replaced"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			stale := t.TempDir()
			release := holdWorkspaceDialog(t, x, stale, nil)
			switch kind {
			case "workspace":
				x.SetWorkspace(t.TempDir())
			case "workspace-roundtrip":
				initial := x.workspace
				x.SetWorkspace(t.TempDir())
				x.SetWorkspace(initial)
			case "new-session":
				x.Action(Action{Kind: "NewSession"})
			case "session-open":
				x.OpenSession(filepath.Join(t.TempDir(), "absent-session.jsonl"))
			case "window-replaced":
				x.UI.Window = new(mygo.Window) // compare only; not a native window
			}
			want := x.workspace
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.workspaceBrowse.pending })
			if x.workspace != want {
				t.Fatal("obsolete picker answer overrode newer intent", x.workspace, want)
			}
			x.UI.Window = nil
		})
	}
}

func TestWorkspaceDialogShutdownDiscardsQueuedAndLaterAnswers(t *testing.T) {
	for _, queued := range []bool{false, true} {
		t.Run(map[bool]string{false: "pending", true: "queued"}[queued], func(t *testing.T) {
			x := &Controller{UI: &App{updates: make(chan func(), 64)}, workspace: t.TempDir(), stop: make(chan struct{})}
			initial := x.workspace
			release := holdWorkspaceDialog(t, x, t.TempDir(), nil)
			var apply func()
			if queued {
				release()
				apply = takeControllerDialogUpdate(t, x)
			}
			close(x.stop) // Close shuts this channel before closing any owned service.
			release()
			if apply != nil {
				apply()
			}
			if x.workspace != initial {
				t.Fatal("answer mutated closed controller")
			}
			x.Action(Action{Kind: "BrowseWorkspace"}) // closed app must not open another sheet
		})
	}
}
