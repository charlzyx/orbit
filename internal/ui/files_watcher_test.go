package orbitui

import (
	"github.com/egoist/mygo/ui/editor"
	"os"
	"testing"
	"time"
)

func watcherCompletion(t *testing.T, a *App) func() {
	t.Helper()
	select {
	case apply := <-a.updates:
		return apply
	case <-time.After(5 * time.Second):
		t.Fatal("watcher worker did not return")
		return nil
	}
}

func TestWatcherLateNativeEditPreserved(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	if err := os.WriteFile(path, []byte("external\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a) // Real background load completed; UI callback held.
	observed := a.files.versions[path]
	st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "late native\n"})
	pending := st.Buffer.Version()
	apply()
	d, _ := a.files.editor.Active()
	t.Logf("domain=%q native=%q ack=%d pending=%d", d.Text, st.Buffer.String(), a.files.versions[path], pending)
	if st.Buffer.String() != "late native\n" || !st.Buffer.CanUndo() {
		t.Error("watcher completion replaced pending native edit")
	}
	if a.files.versions[path] != observed {
		t.Error("watcher completion acknowledged pending edit")
	}
	if d.Text != "external\n" {
		t.Fatal("fixture did not actually reload external bytes", d)
	}
}

func TestWatcherCleanReloadAndSelfEcho(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	v := st.Buffer.Version()
	a.reloadWatchedFiles()
	watcherCompletion(t, a)()
	if st.Buffer.Version() != v {
		t.Fatal("self echo changed native buffer")
	}
	if err := os.WriteFile(path, []byte("external\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a.reloadWatchedFiles()
	watcherCompletion(t, a)()
	if st.Buffer.String() != "external\n" || a.files.versions[path] != st.Buffer.Version() {
		t.Fatal("clean external reload not applied")
	}
}

func TestWatcherDispatchSyncsPendingEdit(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "pending\n"})
	a.reloadWatchedFiles()
	d, _ := a.files.editor.Active()
	if d.Text != "pending\n" || a.files.versions[path] != st.Buffer.Version() {
		t.Fatal("dispatch failed to sync native pending", d)
	}
	watcherCompletion(t, a)()
	if st.Buffer.String() != "pending\n" {
		t.Fatal("reload replaced synchronized draft")
	}
}

func TestWatcherLatePendingAcrossReadonlyAndReadError(t *testing.T) {
	for _, mode := range []string{"lossy", "missing"} {
		t.Run(mode, func(t *testing.T) {
			a, _, path := mountedFile(t, "original\n")
			st := a.files.states[path]
			var err error
			if mode == "lossy" {
				err = os.WriteFile(path, []byte{0xff, '\n'}, 0600)
			} else {
				err = os.Remove(path)
			}
			if err != nil {
				t.Fatal(err)
			}
			a.reloadWatchedFiles()
			apply := watcherCompletion(t, a)
			ack := a.files.versions[path]
			st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "late pending\n"})
			apply()
			if st.Buffer.String() != "late pending\n" || a.files.versions[path] != ack || !st.Buffer.CanUndo() {
				t.Fatal("readonly/error reload replaced native draft")
			}
			d, _ := a.files.editor.Active()
			if mode == "lossy" && !d.Content.Lossy || mode == "missing" && d.Content.Error == "" {
				t.Fatal("fixture did not reach external mode", d)
			}
			a.reloadWatchedFiles() // Keep disk recovery possible, without acknowledging pending text.
			watcherCompletion(t, a)()
			if st.Buffer.String() != "late pending\n" || a.files.versions[path] != ack {
				t.Fatal("dispatch acknowledged readonly pending")
			}
		})
	}
}

func TestWatcherCompletionExcludesNewTab(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a)
	other := path + ".other"
	if err := os.WriteFile(other, []byte("other"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := a.files.editor.Open(other)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	st := a.files.states[other]
	ack := a.files.versions[other]
	st.Buffer.Edit(editor.Edit{Start: 0, End: 5, Text: "new tab pending"})
	apply()
	if st.Buffer.String() != "new tab pending" || a.files.versions[other] != ack {
		t.Fatal("old watcher applied to a new tab")
	}
}

func TestWatcherCompletionRejectsReopenedTarget(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	old := a.files.states[path]
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a)
	a.closeFile(path)
	drainRegressionUpdates(t, a, func() bool { return a.files.states[path] == nil })
	if err := os.WriteFile(path, []byte("reopened"), 0600); err != nil {
		t.Fatal(err)
	}
	a.OpenFile(path)
	drainRegressionUpdates(t, a, func() bool { return a.files.states[path] != nil })
	st := a.files.states[path]
	if st == nil || st == old {
		t.Fatal("real close/reopen did not replace native target")
	}
	// No pending version: rejection must be original target identity, not only pending guard.
	if err := os.WriteFile(path, []byte("new domain"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := a.files.editor.Reload(path); err != nil {
		t.Fatal(err)
	}
	apply()
	if st.Buffer.String() != "reopened" {
		t.Fatal("old watcher completion updated reopened target")
	}
}

func TestWatcherDirtyDomainDoesNotReplaceLaterNativeEdit(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "first draft\n"})
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a)
	ack := a.files.versions[path]
	st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "later draft\n"})
	apply()
	if st.Buffer.String() != "later draft\n" || a.files.versions[path] != ack {
		t.Fatal("domain dirty did not protect later native draft")
	}
}

func TestWatcherReadonlyPendingCanRecoverWithoutReplacement(t *testing.T) {
	a, tt, path := mountedFile(t, "original\n")
	st := a.files.states[path]
	if err := os.WriteFile(path, []byte{0xff, '\n'}, 0600); err != nil {
		t.Fatal(err)
	}
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a)
	st.Buffer.Edit(editor.Edit{Start: 0, End: len(st.Buffer.String()), Text: "pending\n"})
	apply()
	ack := a.files.versions[path]
	if err := os.WriteFile(path, []byte("recovered disk\n"), 0600); err != nil {
		t.Fatal(err)
	}
	a.reloadWatchedFiles()
	watcherCompletion(t, a)()
	d, _ := a.files.editor.Active()
	if !d.Content.Editable() || st.Buffer.String() != "pending\n" || a.files.versions[path] != ack {
		t.Fatal("recovery replaced or acknowledged pending", d)
	}
	tt.Frame()
	d, _ = a.files.editor.Active()
	if d.Text != "pending\n" || a.files.versions[path] != st.Buffer.Version() {
		t.Fatal("normal next render did not synchronize recovered draft", d)
	}
}

func TestWatcherCompletionRejectsReplacedFilesOwner(t *testing.T) {
	a, _, path := mountedFile(t, "original\n")
	original := a.files
	a.reloadWatchedFiles()
	apply := watcherCompletion(t, a)
	// Defensive owner invariant only: production workspace changes currently
	// retain filesState, so this is not a claimed workspace-switch reproduction.
	a.files = nil
	a.initFiles()
	replacement := a.files
	defer func() { _ = replacement.editor.CloseAll(); a.files = original }()
	d, err := replacement.editor.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	a.syncDocument(d)
	st := replacement.states[path]
	if err := original.editor.SetText(path, "old owner text"); err != nil {
		t.Fatal(err)
	}
	apply()
	if st.Buffer.String() != "original\n" {
		t.Fatal("old owner completion touched replacement")
	}
}
