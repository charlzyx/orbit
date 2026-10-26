package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestFileNamesAndNoClobber(t *testing.T) {
	for _, name := range []string{"", " ", ".", "..", "../x", "a/b", "a\\b", "a\x00b", "a\nb", strings.Repeat("a", 256)} {
		if _, err := ValidateName(name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	dir := t.TempDir()
	p, err := CreateFile(dir, " example.txt ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(p, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = CreateFile(dir, "example.txt"); err == nil {
		t.Fatal("clobbered")
	}
	raw, _ := os.ReadFile(p)
	if string(raw) != "keep" {
		t.Fatal(string(raw))
	}
	other, _ := CreateFile(dir, "other")
	if _, err = RenamePath(p, "other"); !errors.Is(err, ErrDestinationExists) {
		t.Fatal(err)
	}
	if _, err = RenamePath(p, "example.txt"); err != nil {
		t.Fatal(err)
	}
	if _, err = RenamePath(p, "renamed"); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(other); err != nil {
		t.Fatal(err)
	}
}
func TestFileLoadingModesAndGuards(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name     string
		body     []byte
		mode     FileMode
		editable bool
	}{{"code.go", []byte("package hi\n"), FileCode, true}, {"readme.md", []byte("# hi\n"), FileMarkdown, true}, {"plain", []byte("hi"), FileText, true}, {"binary", []byte{'x', 0, 'y'}, FileBinary, false}, {"bad.txt", []byte{0xff, 'a'}, FileText, false}, {"picture.png", []byte("not decoded here"), FileImage, false}}
	for _, tc := range cases {
		p := filepath.Join(dir, tc.name)
		os.WriteFile(p, tc.body, 0600)
		c := LoadFile(p)
		if c.Mode != tc.mode || c.Editable() != tc.editable {
			t.Fatalf("%s: %+v", tc.name, c)
		}
	}
	p := filepath.Join(dir, "large")
	f, _ := os.Create(p)
	f.Truncate(MaxFileBytes + 1)
	f.Close()
	if LoadFile(p).Mode != FileTooLarge {
		t.Fatal("large")
	}
	if LoadFile(dir).Error == "" || LoadFile(filepath.Join(dir, "missing")).Error == "" {
		t.Fatal("error state missing")
	}
	p = filepath.Join(dir, "lines")
	os.WriteFile(p, []byte(strings.Repeat("\n", MaxFileLines+1)), 0600)
	c := LoadFile(p)
	if !c.Truncated || c.Editable() {
		t.Fatal("line guard")
	}
}
func TestSaveExistingDoesNotRecreate(t *testing.T) {
	p := filepath.Join(t.TempDir(), "gone")
	if err := SaveExisting(p, "no"); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}
func TestDeleteConfirmationAndCancellation(t *testing.T) {
	p, _ := CreateFile(t.TempDir(), "file")
	if !errors.Is(DeletePath(context.Background(), p, false), ErrConfirmationRequired) {
		t.Fatal("confirmation bypass")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(DeletePath(ctx, p, true), context.Canceled) {
		t.Fatal("cancel")
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("deleted")
	}
	if runtime.GOOS != "darwin" {
		if err := DeletePath(context.Background(), p, true); err != nil {
			t.Fatal(err)
		}
		if err := DeletePath(context.Background(), p, true); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("repeat must report missing", err)
		}
	}
}
func TestEditorAutosaveUndoRenameAndClose(t *testing.T) {
	p, _ := CreateFile(t.TempDir(), "a.go")
	e := NewEditor()
	e.delay = 20 * time.Millisecond
	defer e.CloseAll()
	if _, err := e.Open(p); err != nil {
		t.Fatal(err)
	}
	if err := e.SetText(p, "one"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetText(p, "two"); err != nil {
		t.Fatal(err)
	}
	e.Undo(p)
	d, _ := e.Active()
	if d.Text != "one" {
		t.Fatal(d)
	}
	e.Redo(p)
	if err := e.Save(p); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(p)
	if string(raw) != "two" {
		t.Fatal(string(raw))
	}
	new, err := RenamePath(p, "b.go")
	if err != nil {
		t.Fatal(err)
	}
	e.RenameUnder(p, new)
	e.SetText(new, "renamed")
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		raw, _ = os.ReadFile(new)
		if string(raw) == "renamed" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if string(raw) != "renamed" {
		t.Fatal("autosave")
	}
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("old path recreated")
	}
	if err := e.Close(new); err != nil {
		t.Fatal(err)
	}
	if err := e.Close(new); err != nil {
		t.Fatal(err)
	}
	if _, ok := e.Active(); ok {
		t.Fatal("closed active")
	}
}
func TestEditorDirtyReloadAndFailedSaveKeepBuffer(t *testing.T) {
	p, _ := CreateFile(t.TempDir(), "a.txt")
	e := NewEditor()
	e.delay = time.Hour
	defer func() { e.CloseUnder(p); e.CloseAll() }()
	e.Open(p)
	e.SetText(p, "draft")
	os.WriteFile(p, []byte("external"), 0600)
	e.Reload(p)
	d, _ := e.Active()
	if d.Text != "draft" {
		t.Fatal("overwrote dirty")
	}
	os.Remove(p)
	if err := e.Close(p); err == nil {
		t.Fatal("close lost unsaved buffer")
	}
	d, ok := e.Active()
	if !ok || !d.Dirty || d.Text != "draft" || d.LastError == "" {
		t.Fatalf("%+v %t", d, ok)
	}
}
func TestEditorReadonly(t *testing.T) {
	p := filepath.Join(t.TempDir(), "binary")
	os.WriteFile(p, []byte{0, 1}, 0600)
	e := NewEditor()
	defer e.CloseAll()
	e.Open(p)
	if !errors.Is(e.SetText(p, "oops"), ErrReadOnly) {
		t.Fatal("editable binary")
	}
}

func TestEditorDebounceAcrossTabsAndClosingLastEdited(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		root := t.TempDir()
		a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
		if err := os.WriteFile(a, []byte("a"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(b, []byte("b"), 0600); err != nil {
			t.Fatal(err)
		}
		e := NewEditor()
		e.delay = 100 * time.Millisecond
		defer e.CloseAll()
		entered, release, completed := make(chan struct{}), make(chan struct{}), make(chan error, 1)
		released := false
		defer func() {
			if !released {
				close(release)
			}
		}()
		e.writeFile = func(path, text string) error {
			if path == a {
				close(entered)
				<-release
			}
			err := SaveExisting(path, text)
			if path == a {
				completed <- err
			}
			return err
		}
		if _, err := e.Open(a); err != nil {
			t.Fatal(err)
		}
		if _, err := e.Open(b); err != nil {
			t.Fatal(err)
		}
		if err := e.SetText(a, "changed a"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(60 * time.Millisecond)
		if err := e.SetText(b, "changed b"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(60 * time.Millisecond)
		synctest.Wait()
		// The second edit reset the global deadline: no timer write may start yet.
		select {
		case <-entered:
			t.Fatal("autosave fired before global idle window")
		default:
		}
		raw, err := os.ReadFile(a)
		if err != nil || string(raw) != "a" {
			t.Fatal("autosave fired before global idle window", string(raw), err)
		}
		if err := e.Close(b); err != nil {
			t.Fatal(err)
		}
		time.Sleep(70 * time.Millisecond)
		synctest.Wait()
		select {
		case <-entered:
		default:
			t.Fatal("timer did not save dirty neighbor after closing last-edited tab")
		}
		// Timer expiry is not I/O completion: prove the old observation can still
		// see original bytes, then release and wait for the actual disk write.
		raw, err = os.ReadFile(a)
		if err != nil || string(raw) != "a" {
			t.Fatal("held write changed disk", string(raw), err)
		}
		close(release)
		released = true
		if err := <-completed; err != nil {
			t.Fatal(err)
		}
		synctest.Wait() // Settle saveDocument bookkeeping after the write seam.
		raw, err = os.ReadFile(a)
		if err != nil || string(raw) != "changed a" {
			t.Fatal("closing last-edited tab abandoned dirty neighbor", string(raw), err)
		}
		raw, err = os.ReadFile(b)
		if err != nil || string(raw) != "changed b" {
			t.Fatal("Close did not save its own target", string(raw), err)
		}
		d, ok := e.Active()
		if !ok || d.Content.Path != a || d.Dirty || d.Saving {
			t.Fatal("neighbor save did not settle", d)
		}
	})
}

func TestEditorAutosaveEpochRejectsOldAndSavesAllDirty(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		root := t.TempDir()
		e := NewEditor()
		defer e.CloseAll()
		paths := []string{filepath.Join(root, "a"), filepath.Join(root, "b")}
		for _, p := range paths {
			if err := os.WriteFile(p, []byte("original"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := e.Open(p); err != nil {
				t.Fatal(err)
			}
		}
		if err := e.SetText(paths[0], "first"); err != nil {
			t.Fatal(err)
		}
		old := e.saveEpoch
		if err := e.SetText(paths[1], "second"); err != nil {
			t.Fatal(err)
		}
		current := e.saveEpoch
		if old == current {
			t.Fatal("edit did not advance autosave epoch")
		}
		if err := e.flushAll(old, true); err != nil {
			t.Fatal(err)
		}
		for _, p := range paths {
			raw, err := os.ReadFile(p)
			if err != nil || string(raw) != "original" {
				t.Fatal("stale epoch wrote disk", string(raw), err)
			}
		}
		if err := e.flushAll(current, true); err != nil {
			t.Fatal(err)
		}
		for i, p := range paths {
			raw, err := os.ReadFile(p)
			if err != nil || string(raw) != []string{"first", "second"}[i] {
				t.Fatal("current epoch missed dirty tab", string(raw), err)
			}
		}
	})
}

func TestSaveExistingAtomicPreservesModeAndConflict(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "run.sh")
	os.WriteFile(path, []byte("old\n"), 0750)
	if e := SaveExisting(path, "new\n"); e != nil {
		t.Fatal(e)
	}
	fi, _ := os.Stat(path)
	if runtime.GOOS != "windows" && fi.Mode().Perm() != 0750 {
		t.Fatal(fi.Mode())
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".*.orbit-save-*"))
	if len(matches) != 0 {
		t.Fatal(matches)
	}
	editor := NewEditor()
	editor.delay = time.Hour
	defer editor.CloseAll()
	editor.Open(path)
	editor.SetText(path, "unsaved\n")
	os.WriteFile(path, []byte("external\n"), 0750)
	editor.Reload(path)
	d, _ := editor.Active()
	if !d.Conflict || d.Text != "unsaved\n" {
		t.Fatal(d)
	}
	editor.SaveAll()
	d, _ = editor.Active()
	if d.Conflict || d.Dirty {
		t.Fatal(d)
	}
}

func TestSlowSaveDoesNotBlockSnapshotsOrLoseConcurrentEdits(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a")
	os.WriteFile(path, []byte("original"), 0600)
	e := NewEditor()
	e.delay = time.Hour
	defer e.CloseAll()
	e.Open(path)
	e.SetText(path, "first edit")
	started, release := make(chan struct{}), make(chan struct{})
	e.writeFile = func(path, text string) error { close(started); <-release; return SaveExisting(path, text) }
	saved := make(chan error, 1)
	go func() { saved <- e.Save(path) }()
	<-started
	responsive := make(chan Document, 1)
	go func() { d, _ := e.Active(); e.SetText(path, "second edit"); responsive <- d }()
	select {
	case d := <-responsive:
		if !d.Saving {
			t.Fatal("in-flight save not visible")
		}
	case <-time.After(200 * time.Millisecond):
		close(release)
		t.Fatal("disk write blocked UI snapshot/edit")
	}
	close(release)
	if err := <-saved; err != nil {
		t.Fatal(err)
	}
	d, _ := e.Active()
	if d.Text != "second edit" || !d.Dirty || d.Saving {
		t.Fatal("save completion lost newer edit", d)
	}
	raw, _ := os.ReadFile(path)
	if string(raw) != "first edit" {
		t.Fatal(string(raw))
	}
	e.writeFile = SaveExisting
	if err := e.SaveAll(); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(path)
	if string(raw) != "second edit" {
		t.Fatal(string(raw))
	}
}
func TestCloseKeepsTabIfUserEditsDuringSlowSave(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a")
	os.WriteFile(path, []byte("original"), 0600)
	e := NewEditor()
	e.delay = time.Hour
	defer e.CloseAll()
	e.Open(path)
	e.SetText(path, "first")
	started, release := make(chan struct{}), make(chan struct{})
	e.writeFile = func(path, text string) error { close(started); <-release; return SaveExisting(path, text) }
	closed := make(chan error, 1)
	go func() { closed <- e.Close(path) }()
	<-started
	e.SetText(path, "newer")
	close(release)
	if err := <-closed; !errors.Is(err, ErrDocumentChanged) {
		t.Fatal(err)
	}
	d, ok := e.Active()
	if !ok || !d.Dirty || d.Text != "newer" {
		t.Fatal("close dropped new input", d)
	}
	e.writeFile = SaveExisting
}

func TestEditorInactiveLoadNeverStealsSelection(t *testing.T) {
	root := t.TempDir()
	first := filepath.Join(root, "first.txt")
	other := filepath.Join(root, "other.txt")
	os.WriteFile(first, []byte("first"), 0600)
	os.WriteFile(other, []byte("other"), 0600)
	e := NewEditor()
	defer e.CloseAll()
	if _, err := e.OpenInactive(first); err != nil {
		t.Fatal(err)
	}
	if _, ok := e.Active(); ok {
		t.Fatal("inactive first load activated")
	}
	if err := e.Activate(first); err != nil {
		t.Fatal(err)
	}
	if _, err := e.OpenInactive(other); err != nil {
		t.Fatal(err)
	}
	if d, ok := e.Active(); !ok || d.Content.Path != first {
		t.Fatal("background open stole selection", d)
	}
	if err := e.Activate(other); err != nil {
		t.Fatal(err)
	}
	if _, err := e.OpenInactive(first); err != nil {
		t.Fatal(err)
	}
	if d, ok := e.Active(); !ok || d.Content.Path != other {
		t.Fatal("cached inactive open stole selection", d)
	}
	if _, err := e.Open(first); err != nil {
		t.Fatal(err)
	}
	if d, ok := e.Active(); !ok || d.Content.Path != first {
		t.Fatal("legacy Open no longer selects", d)
	}
}

func TestEditorSelfSaveReloadPreservesUndoHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.txt")
	os.WriteFile(path, []byte("before"), 0600)
	e := NewEditor()
	defer e.CloseAll()
	e.Open(path)
	e.SetText(path, "after")
	if err := e.Save(path); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(path); err != nil {
		t.Fatal(err)
	}
	if err := e.Undo(path); err != nil {
		t.Fatal(err)
	}
	d, _ := e.Active()
	if d.Text != "before" || !d.Dirty {
		t.Fatal("self-save echo discarded undo", d)
	}
	// A matching external save acknowledges the buffer rather than reporting a conflict.
	os.WriteFile(path, []byte("before"), 0600)
	e.Reload(path)
	d, _ = e.Active()
	if d.Dirty || d.Conflict {
		t.Fatal("matching external save not reconciled", d)
	}
}
