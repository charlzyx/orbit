package orbitui

import (
	"errors"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func holdBackgroundDialog(t *testing.T, x *Controller, paths []string, err error) func() {
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
	x.fileDialogs.background = func(o mygo.OpenDialogOptions) ([]string, error) {
		if o.Title == "" || len(o.Filters) != 1 {
			t.Error("lost native options")
		}
		close(started)
		<-release
		return paths, err
	}
	returned := make(chan struct{})
	go func() { x.Action(Action{Kind: "BrowseBackground"}); close(returned) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("dialog did not start")
	}
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		unblock()
		<-returned
		t.Fatal("background dialog blocked caller")
	}
	return unblock
}

func TestBackgroundDialogAsyncAppliesGlobalChoiceOnOwnerQueue(t *testing.T) {
	x := workspaceDialogFixture(t)
	drainControllerDialogUpdates(t, x, func() bool { return !x.UI.backdrop.loading })
	path := filepath.Join(t.TempDir(), "背景 image.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, image.NewNRGBA(image.Rect(0, 0, 16, 16)))
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	release := holdBackgroundDialog(t, x, []string{path}, nil)
	x.Action(Action{Kind: "BrowseBackground"})
	x.Action(Action{Kind: "BrowseWorkspace"})
	x.Action(Action{Kind: "BrowseAttachments"})
	if x.workspaceBrowse.pending {
		t.Fatal("stacked native picker")
	}
	ui.NewTester(x.View, 1470, 922).Frame()
	x.UI.Navigate("chat") // background is global, not tied to settings page
	release()
	drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending && !x.UI.backdrop.loading })
	if x.UI.backdrop.config.Source == nil || x.UI.backdrop.bitmap == nil || x.UI.backdrop.err != "" {
		t.Fatalf("background not loaded: %+v", x.UI.backdrop.config)
	}
}
func TestBackgroundDialogCancelErrorAndStaleOwner(t *testing.T) {
	for _, kind := range []string{"cancel", "error", "reset", "window", "close"} {
		t.Run(kind, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			drainControllerDialogUpdates(t, x, func() bool { return !x.UI.backdrop.loading })
			paths := []string{filepath.Join(t.TempDir(), "must-not-be-read.png")}
			var err error
			if kind == "cancel" {
				paths = nil
			}
			if kind == "error" {
				err = errors.New("synthetic picker error")
			}
			release := holdBackgroundDialog(t, x, paths, err)
			switch kind {
			case "reset":
				x.UI.clearBackground()
			case "window":
				x.UI.Window = new(mygo.Window)
			case "close":
				x.Close()
			}
			release()
			// Close drops completion; its queue must not be required to drain.
			if kind == "close" {
				return
			}
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending && !x.UI.backdrop.loading })
			if x.UI.backdrop.config.Source != nil || x.UI.backdrop.err != "" {
				t.Fatal("cancel/stale result reached photo worker")
			}
			x.UI.Window = nil
		})
	}
}
