package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Current domain API feasibility, not an instrumented UI/OS event trace.
// A channel holds the existing boundary between real FS rename and RenameUnder.
func TestRenameGapSnapshotLateEditAndRecoveryProbe(t *testing.T) {
	root := t.TempDir()
	old := filepath.Join(root, "old.txt")
	if err := os.WriteFile(old, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	e := NewEditor()
	e.delay = time.Hour
	defer e.CloseAll()
	if _, err := e.Open(old); err != nil {
		t.Fatal(err)
	}
	fsDone := make(chan string, 1)
	allowDomain := make(chan struct{})
	done := make(chan struct{})
	go func() {
		p, err := RenamePath(old, "new.txt")
		if err != nil {
			fsDone <- ""
			close(done)
			return
		}
		fsDone <- p
		<-allowDomain
		e.RenameUnder(old, p)
		close(done)
	}()
	renamed := <-fsDone
	if renamed == "" {
		t.Fatal("real filesystem rename failed")
	}
	released := false
	defer func() {
		if !released {
			close(allowDomain)
		}
		<-done
	}()
	snapshot := make(chan []Document, 1)
	go func() { snapshot <- e.Documents() }()
	select {
	case docs := <-snapshot:
		if len(docs) != 1 || docs[0].Content.Path != old {
			t.Fatal("wrong gap snapshot", docs)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("snapshot blocked while domain remap held")
	}
	if err := e.SetText(old, "late retained draft"); err != nil {
		t.Fatal(err)
	}
	err := e.Save(old) // Represents a save that enters the currently ungated gap.
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal("gap save did not reject vanished old path", err)
	}
	d, _ := e.Active()
	if d.Text != "late retained draft" || !d.Dirty || d.LastError == "" {
		t.Fatal("failed gap save lost domain draft", d)
	}
	t.Logf("gap: save=%v text=%q dirty=%v", err, d.Text, d.Dirty)
	close(allowDomain)
	released = true
	<-done
	d, _ = e.Active()
	if d.Content.Path != renamed || d.Text != "late retained draft" || !d.Dirty {
		t.Fatal("domain remap discarded late edit", d)
	}
	if err := e.SetText(old, "native edit bound to obsolete path"); !errors.Is(err, ErrDocumentNotOpen) {
		t.Fatal("oldpath late binding unexpectedly resolved", err)
	}
	if err := e.Save(renamed); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(renamed)
	if err != nil || string(raw) != "late retained draft" {
		t.Fatal("newpath explicit recovery failed", string(raw), err)
	}
	if _, err := os.Stat(old); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("old path recreated", err)
	}
	t.Log("snapshot nonblocking; domain late draft preserved and explicit newpath save recovered; oldpath binding rejected")
}
