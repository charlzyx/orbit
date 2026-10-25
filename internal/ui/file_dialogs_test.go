package orbitui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func holdFileDialog(t *testing.T, x *Controller, action Action, paths []string, err error) func() {
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
	wait := func() { close(started); <-release }
	x.fileDialogs.attachments = func(*mygo.Window, string) ([]string, error) { wait(); return paths, err }
	x.fileDialogs.image = x.fileDialogs.attachments
	x.fileDialogs.save = func(*mygo.Window, string, string, []mygo.FileFilter) (string, error) {
		wait()
		if len(paths) == 0 {
			return "", err
		}
		return paths[0], err
	}
	returned := make(chan struct{})
	go func() { x.Action(action); close(returned) }()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("file dialog not started")
	}
	select {
	case <-returned:
	case <-time.After(2 * time.Second):
		unblock()
		<-returned
		t.Fatal("file dialog blocked action caller")
	}
	return unblock
}
func TestNativeFileDialogAsyncSingletonAndAttachmentOwnership(t *testing.T) {
	x := workspaceDialogFixture(t)
	path := filepath.Join(t.TempDir(), "文档 space.txt")
	release := holdFileDialog(t, x, Action{Kind: "BrowseAttachments"}, []string{path}, nil)
	x.Action(Action{Kind: "BrowseAttachments"})
	x.Action(Action{Kind: "BrowseWorkspace"}) // no second native sheet of another kind
	if x.workspaceBrowse.pending {
		t.Fatal("stacked workspace picker behind file dialog")
	}
	if x.UI.Draft != "" {
		t.Fatal("attachment applied before answer")
	}
	release()
	drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
	ui.NewTester(x.UI.View, 1200, 800).Frame()
	if x.UI.Draft != path+" " {
		t.Fatal("path not applied on UI queue", x.UI.Draft)
	}
}
func TestNativeFileDialogStaleTaskAndWindowDoNotAttachOrExport(t *testing.T) {
	for _, kind := range []string{"BrowseAttachments"} {
		for _, stale := range []string{"task", "window"} {
			t.Run(kind+"/"+stale, func(t *testing.T) {
				x := workspaceDialogFixture(t)
				release := holdFileDialog(t, x, Action{Kind: kind}, []string{filepath.Join(t.TempDir(), "chosen.txt")}, nil)
				if stale == "task" {
					x.Action(Action{Kind: "NewSession"})
				} else {
					x.UI.Window = new(mygo.Window)
				}
				release()
				drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
				if x.UI.Draft != "" {
					t.Fatal("late result affected newer task/window")
				}
				x.UI.Window = nil
			})
		}
	}
}
func TestNativeUsageExportCapturesDataBeforeDialog(t *testing.T) {
	x := workspaceDialogFixture(t)
	index := workbench.UsageIndex{Records: []workbench.UsageRecord{{Provider: "first", Model: "original", Cost: 1, HasCost: true}, {Provider: "other", Model: "excluded", Cost: 2, HasCost: true}}}
	filter := workbench.UsageFilter{Providers: []string{"first"}}
	action := Action{Kind: "ExportUsage", Value: "csv", Data: map[string]any{"index": index, "filter": filter}}
	path := filepath.Join(t.TempDir(), "usage.csv")
	release := holdFileDialog(t, x, action, []string{path}, nil)
	index.Records[0].Model = "mutated"
	filter.Providers[0] = "other"
	// Export belongs to the captured Usage view, not whatever workspace is now active.
	x.Action(Action{Kind: "NewSession"})
	release()
	drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
	deadline := time.Now().Add(3 * time.Second)
	var data []byte
	for time.Now().Before(deadline) {
		data, _ = os.ReadFile(path)
		if len(data) > 0 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !strings.Contains(string(data), "original") || strings.Contains(string(data), "mutated") || strings.Contains(string(data), "excluded") {
		t.Fatal("did not export captured filtered data", string(data))
	}
}
func TestNativeUsageExportCancelOrErrorLeavesExistingFile(t *testing.T) {
	for _, outcome := range []string{"cancel", "error", "closed-window"} {
		t.Run(outcome, func(t *testing.T) {
			x := workspaceDialogFixture(t)
			path := filepath.Join(t.TempDir(), "existing.csv")
			os.WriteFile(path, []byte("untouched"), 0600)
			chosen := path
			var err error
			if outcome == "cancel" {
				chosen = ""
			}
			if outcome == "error" {
				err = errors.New("save failed")
			}
			release := holdFileDialog(t, x, Action{Kind: "ExportUsage", Value: "csv", Data: map[string]any{"index": workbench.UsageIndex{}}}, []string{chosen}, err)
			if outcome == "closed-window" {
				x.UI.Window = new(mygo.Window)
			}
			release()
			drainControllerDialogUpdates(t, x, func() bool { return !x.fileDialogs.pending })
			data, _ := os.ReadFile(path)
			if string(data) != "untouched" {
				t.Fatal("unconfirmed export modified file")
			}
			x.UI.Window = nil
		})
	}
}
func TestCloneUsageExportOwnsMutableCollections(t *testing.T) {
	index := workbench.UsageIndex{Warnings: []string{"old"}, Records: []workbench.UsageRecord{{Model: "old"}}}
	filter := workbench.UsageFilter{Models: []string{"old"}}
	options := workbench.UsageExportOptions{Insights: []string{"old"}, SessionTitles: map[string]string{"session": "old"}}
	i, f, o := cloneUsageExport(index, filter, options)
	index.Warnings[0] = "new"
	index.Records[0].Model = "new"
	filter.Models[0] = "new"
	options.Insights[0] = "new"
	options.SessionTitles["session"] = "new"
	if i.Warnings[0] != "old" || i.Records[0].Model != "old" || f.Models[0] != "old" || o.Insights[0] != "old" || o.SessionTitles["session"] != "old" {
		t.Fatal("export retained caller-owned data")
	}
}

func TestAttachmentReadCannotAdoptNewEmptyIDTask(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "image", true: "error"}[fail], func(t *testing.T) {
			a := &App{updates: make(chan func(), 64)}
			a.Model.Workspace = "fixture"
			a.Model.RuntimeID = "runtime"
			var generation uint64
			a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, true }
			started, release := make(chan struct{}), make(chan struct{})
			a.addPaths([]string{"synthetic.png"}, func(string) ([]byte, string, error) {
				close(started)
				<-release
				if fail {
					return nil, "", errors.New("obsolete read error")
				}
				return []byte("not an actual image"), "image/png", nil
			})
			<-started
			generation++
			close(release)
			select {
			case apply := <-a.updates:
				apply()
			case <-time.After(2 * time.Second):
				t.Fatal("read did not finish")
			}
			if len(a.attachments) > 0 || len(a.toastStack.Items()) > 0 {
				t.Fatal("stale read adopted newer task or reported obsolete error")
			}
			select {
			case <-a.updates:
				t.Fatal("stale read started another decode job")
			default:
			}
		})
	}
}
func TestAttachmentDecodeRejectsNewOwner(t *testing.T) {
	a := &App{updates: make(chan func(), 64)}
	a.Model.Workspace = "one"
	owner := a.attachmentOwner()
	a.addImageFor(owner, []byte("x"), "image/png", "synthetic")
	a.Model.Workspace = "two"
	select {
	case apply := <-a.updates:
		apply()
	case <-time.After(2 * time.Second):
		t.Fatal("decode did not finish")
	}
	if len(a.attachments) != 0 {
		t.Fatal("decode attached into another workspace")
	}
}

func TestNativeFileDialogDropsCompletionQueuedBeforeShutdown(t *testing.T) {
	x := &Controller{UI: &App{updates: make(chan func(), 1)}, stop: make(chan struct{})}
	applied := false
	x.fileDialog(nativeDialogTarget{}, func() ([]string, error) { return []string{"selected"}, nil }, func([]string, error) { applied = true })
	apply := takeControllerDialogUpdate(t, x)
	close(x.stop)
	apply()
	if applied || x.fileDialogs.pending {
		t.Fatal("closed controller consumed file-dialog result")
	}
}

func TestAttachmentManagedOwnerSurvivesOrdinaryIdentityHydration(t *testing.T) {
	a := &App{updates: make(chan func(), 64)}
	a.Model.Workspace = "same"
	a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, 1, true }
	owner := a.attachmentOwner()
	a.addImageFor(owner, []byte("image fixture"), "image/png", "fixture")
	a.Model.RuntimeID = "hydrated-runtime"
	a.Model.SessionID = "hydrated-session"
	a.Model.SessionPath = "hydrated-path"
	select {
	case apply := <-a.updates:
		apply()
	case <-time.After(2 * time.Second):
		t.Fatal("decode did not finish")
	}
	if len(a.attachments) != 1 {
		t.Fatal("same managed task lost attachment on ID hydration")
	}
}
