package orbitui

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestMountedExplorerFolderMenuToggleAndDismiss(t *testing.T) {
	a, tt, _ := mountedExplorer(t)
	for _, label := range []string{"Expand", "Collapse", "Expand"} {
		if err := tt.RightClick("Explorer row docs"); err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(tt.Menu(), label) {
			t.Fatalf("missing folder action %s: %v", label, tt.Menu())
		}
		if err := tt.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if a.files.expanded["docs"] != (label == "Expand") {
			t.Fatal("folder toggle not applied", label)
		}
		_, visible := tt.Find("Explorer row docs/readme.md")
		if visible != (label == "Expand") {
			t.Fatal("children do not reflect toggle", label)
		}
		if tt.Menu() != nil {
			t.Fatal("menu remained after choice")
		}
	}
	if err := tt.RightClick("Explorer row docs"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Frame()
	if !a.files.expanded["docs"] {
		t.Fatal("cancel changed expansion")
	}
}

func TestMountedExplorerDefaultAppActionUsesClickedFileOrFolder(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	var got []Action
	a.Options.OnAction = func(action Action) { got = append(got, action) }
	for _, path := range []string{"alpha.go", "docs", "bravo.go"} {
		if err := tt.RightClick("Explorer row " + path); err != nil {
			t.Fatal(err)
		}
		if err := tt.ChooseMenuItem("Open in Default App"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if len(got) == 0 || got[len(got)-1].Kind != "OpenDefaultPath" || got[len(got)-1].Value != filepath.Join(root, path) {
			t.Fatal("wrong external target", got)
		}
		if tt.Menu() != nil {
			t.Fatal("menu did not dismiss")
		}
	}
	if len(got) != 3 {
		t.Fatal("duplicated external action", got)
	}
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Frame()
	if len(got) != 3 {
		t.Fatal("cancel launched external app")
	}
}

func TestExplorerMenuCannotRetargetAfterWorkspaceChange(t *testing.T) {
	a, tt, _ := mountedExplorer(t)
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(other, "alpha.go"), []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	a.Model.Workspace = other
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == other })
	tt.Frame()
	if tt.Menu() != nil {
		_ = tt.ChooseMenuItem("Open in Default App")
	}
	if len(actions) != 0 {
		t.Fatal("stale menu retargeted a different workspace", actions)
	}
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Open in Default App"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Value != filepath.Join(other, "alpha.go") {
		t.Fatal("new menu not bound to new root", actions)
	}
}

func TestDefaultAppLaunchSchedulesOutsideCallerAndReportsFailure(t *testing.T) {
	a := fixture(t)
	release := make(chan struct{})
	seen := make(chan string, 1)
	returned := make(chan struct{})
	path := filepath.Join(t.TempDir(), "chosen.txt")
	go func() {
		a.openPathDefault(path, func(p string) error { seen <- p; <-release; return errors.New("fixture launch failed") })
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(time.Second):
		t.Fatal("launch blocked caller")
	}
	select {
	case got := <-seen:
		if got != path {
			t.Fatal(got)
		}
	case <-time.After(time.Second):
		t.Fatal("launch not started")
	}
	a.Model.Workspace = t.TempDir()
	close(release)
	drainRegressionUpdates(t, a, func() bool { return len(a.toastStack.Items()) > 0 })
	if a.toastStack.Items()[0].Body != "fixture launch failed" {
		t.Fatal(a.toastStack.Items())
	}
}

func TestExplorerMenuDeleteStillRequiresConfirmation(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	path := filepath.Join(root, "alpha.go")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Delete…"); err != nil {
		t.Fatal(err)
	}
	if !a.confirmOpen {
		t.Fatal("delete skipped confirmation")
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatal("delete changed file before confirmation", err)
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	after, err = os.ReadFile(path)
	if err != nil || string(after) != string(before) || a.confirmOpen {
		t.Fatal("cancel changed fixture or left confirmation", err)
	}
}

func TestExplorerMenuRoundTripDoesNotReviveOldChoice(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	a.Model.Workspace = t.TempDir()
	a.RefreshFiles()
	a.Model.Workspace = root
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	tt.Frame()
	if tt.Menu() != nil {
		_ = tt.ChooseMenuItem("Open in Default App")
	}
	if len(actions) != 0 {
		t.Fatal("roundtrip revived prior workspace intent", actions)
	}
}
