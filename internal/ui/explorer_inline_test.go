package orbitui

import (
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMountedExplorerCreateIsInlineNotModal(t *testing.T) {
	a, tt, _ := mountedExplorer(t)
	if err := tt.Click("New file"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.promptOpen {
		t.Fatal("Explorer create still uses global modal instead of inline row")
	}
	if r, ok := tt.Find("Explorer entry name"); !ok || r.W < 40 {
		t.Fatal("missing usable inline entry", r, ok)
	}
}

func TestMountedExplorerInlineInvalidNameEditAndEscape(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	if err := tt.Click("New file"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Type("../outside")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.files.entry == nil || a.files.entry.value != "../outside" || a.files.entry.err == "" || !tt.Focused("Explorer entry name") {
		t.Fatal("invalid name lost entry/focus", a.files.entry)
	}
	if _, err := os.Stat(filepath.Join(root, "outside")); err == nil {
		t.Fatal("invalid create wrote a file")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("valid.txt")
	tt.Frame()
	if a.files.entry.err != "" {
		t.Fatal("editing did not clear error")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.files.entry != nil || !tt.Focused("Explorer files") {
		t.Fatal("Escape did not cancel/refocus tree")
	}
	if _, err := os.Stat(filepath.Join(root, "valid.txt")); !os.IsNotExist(err) {
		t.Fatal("cancel wrote file", err)
	}
}
func TestMountedExplorerInlineCreateDuplicateAndRetry(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	if err := tt.Click("New file"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Type("alpha.go")
	tt.Key(0, ui.KeyEnter)
	drainRegressionUpdates(t, a, func() bool { return !a.files.mutating })
	tt.Frame()
	if a.files.entry == nil || a.files.entry.err == "" || a.files.entry.value != "alpha.go" {
		t.Fatal("collision did not preserve retryable input", a.files.entry)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("created.go")
	tt.Key(0, ui.KeyEnter)
	tt.Key(0, ui.KeyEnter)
	drainRegressionUpdates(t, a, func() bool { return !a.files.mutating && a.files.entry == nil })
	tt.Frame()
	if _, err := os.Stat(filepath.Join(root, "created.go")); err != nil {
		t.Fatal(err)
	}
	if a.promptOpen {
		t.Fatal("inline opened modal")
	}
}
func TestMountedExplorerInlineRenameAtRowAndCancelClick(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Rename…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if r, ok := tt.Find("Explorer rename row alpha.go"); !ok || r.W < 40 {
		t.Fatal("missing rename row", r, ok)
	}
	if a.promptOpen || a.files.entry.value != "alpha.go" {
		t.Fatal("wrong rename editor")
	}
	if err := tt.Click("Explorer row bravo.go"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.files.entry != nil || a.page == "files" {
		t.Fatal("click-away activated instead of cancelling")
	}
	if _, err := os.Stat(filepath.Join(root, "alpha.go")); err != nil {
		t.Fatal(err)
	}
	if err := tt.RightClick("Explorer row alpha.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Rename…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("renamed.go")
	tt.Key(0, ui.KeyEnter)
	drainRegressionUpdates(t, a, func() bool { return !a.files.mutating && a.files.entry == nil })
	tt.Frame()
	if _, err := os.Stat(filepath.Join(root, "renamed.go")); err != nil {
		t.Fatal(err)
	}
}
func TestMountedExplorerInlineWorkspaceRoundTripCancelsIntent(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	if err := tt.Click("New folder"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	tt.Type("not-created")
	old := a.files.entry
	a.Model.Workspace = t.TempDir()
	a.RefreshFiles()
	a.Model.Workspace = root
	a.RefreshFiles()
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading && a.files.loaded == root })
	tt.Frame()
	a.commitExplorerEntry(old)
	if a.files.entry != nil || a.files.mutating {
		t.Fatal("old entry accepted after workspace roundtrip")
	}
	if _, err := os.Stat(filepath.Join(root, "not-created")); !os.IsNotExist(err) {
		t.Fatal("stale intent wrote", err)
	}
}

func TestMountedExplorerInlineFolderParentAndNarrowWidth(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	a.ExplorerWidth = explorerMinWidth
	tt.Frame()
	if err := tt.RightClick("Explorer row docs"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("New folder…"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if r, ok := tt.Find("Explorer entry name"); !ok || r.W < 40 {
		t.Fatal("narrow inline input collapsed", r, ok)
	}
	tt.Type("nested")
	tt.Key(0, ui.KeyEnter)
	drainRegressionUpdates(t, a, func() bool { return !a.files.mutating && a.files.entry == nil })
	if info, err := os.Stat(filepath.Join(root, "docs", "nested")); err != nil || !info.IsDir() {
		t.Fatal("wrong selected parent", err)
	}
}
func TestExplorerInlineDelayedCompletionCannotStealNewWorkspace(t *testing.T) {
	a, _, root := mountedExplorer(t)
	a.beginExplorerEntry(root, "", false)
	e := a.files.entry
	e.value = "accepted-before-switch.txt"
	a.commitExplorerEntry(e)
	a.commitExplorerEntry(e)
	var apply func()
	select {
	case apply = <-a.updates:
	case <-time.After(2 * time.Second):
		t.Fatal("completion did not arrive")
	}
	if !e.pending || !a.files.mutating {
		t.Fatal("operation not held for owner apply")
	}
	other := t.TempDir()
	a.Model.Workspace = other
	a.RefreshFiles()
	a.beginExplorerEntry(other, "", true) // busy operation must not accept another submit
	apply()
	if a.page == "files" || a.files.entry != nil || a.files.mutating {
		t.Fatal("stale completion stole navigation/state")
	}
	if _, err := os.Stat(filepath.Join(root, "accepted-before-switch.txt")); err != nil {
		t.Fatal("accepted operation lost original target", err)
	}
	drainRegressionUpdates(t, a, func() bool { return !a.files.loading })
}
func TestExplorerInlineCancelPendingCompletionDoesNotReopen(t *testing.T) {
	a, _, root := mountedExplorer(t)
	a.beginExplorerEntry(root, "", false)
	e := a.files.entry
	e.value = "accepted.txt"
	a.commitExplorerEntry(e)
	var apply func()
	select {
	case apply = <-a.updates:
	case <-time.After(2 * time.Second):
		t.Fatal("completion missing")
	}
	a.cancelExplorerEntry()
	apply()
	if a.page == "files" || a.files.entry != nil {
		t.Fatal("cancelled owner completion reopened UI")
	}
	// Cancel after commit does not promise to undo an already accepted disk operation.
	if _, err := os.Stat(filepath.Join(root, "accepted.txt")); err != nil {
		t.Fatal(err)
	}
}
func TestMountedExplorerInlineRenameInvalidDirtyAndFilterDisappearance(t *testing.T) {
	a, tt, root := mountedExplorer(t)
	path := filepath.Join(root, "alpha.go")
	if _, err := a.files.editor.Open(path); err != nil {
		t.Fatal(err)
	}
	a.files.editor.SetText(path, "unsaved")
	a.beginExplorerEntry(root, path, false)
	if a.files.entry != nil {
		t.Fatal("dirty rename offered")
	}
	if err := a.files.editor.Save(path); err != nil {
		t.Fatal(err)
	}
	a.beginExplorerEntry(root, path, false)
	tt.Frame()
	e := a.files.entry
	e.value = "../bad"
	a.commitExplorerEntry(e)
	if e.err == "" || a.files.entry != e {
		t.Fatal("invalid rename lost input")
	}
	e.value = "new.go"
	a.files.editor.SetText(path, "new unsaved")
	a.commitExplorerEntry(e)
	if e.pending || e.err != a.text("explorer.err_unsaved", "Save unsaved edits before renaming or deleting this path.") {
		t.Fatal("submit ignored newly dirty editor", e.err)
	}
	a.files.query = "bravo"
	tt.Frame()
	if a.files.entry != nil {
		t.Fatal("filtered-away rename stranded editor")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
