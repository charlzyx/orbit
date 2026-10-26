package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func recoveryFile(t *testing.T) (*Editor, string, Document) {
	t.Helper()
	e := NewEditor()
	t.Cleanup(func() { e.CloseAll() })
	p := filepath.Join(t.TempDir(), "old")
	if err := os.WriteFile(p, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	d, err := e.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	return e, p, d
}
func recoveryRequest(e *Editor, old, name string) RenameRequest {
	r := RenameRequest{Old: old, Name: name}
	for _, d := range e.Documents() {
		if pathUnder(d.Content.Path, old) {
			r.Targets = append(r.Targets, RenameTarget{d.ID, d.Content.Path})
		}
	}
	return r
}
func TestRecoveryRenameHeldIOKeepsLateEditAndSaveIdentity(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			e, p, d := recoveryFile(t)
			e.delay = time.Hour
			req := recoveryRequest(e, p, "new")
			entered, release := make(chan struct{}), make(chan struct{})
			denied := errors.New("fixture rename denied")
			e.renamePath = func(path, name string) (string, error) {
				close(entered)
				<-release
				if fail {
					return "", denied
				}
				return RenamePath(path, name)
			}
			done := make(chan error, 1)
			go func() { _, err := e.RenameDocuments(req); done <- err }()
			<-entered
			responsive := make(chan error, 1)
			go func() { _ = e.Documents(); responsive <- e.SetTextID(d.ID, "late draft") }()
			select {
			case err := <-responsive:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(2 * time.Second):
				close(release)
				<-done
				t.Fatal("state mutex held across filesystem IO")
			}
			if _, err := e.OpenInactive(p); !errors.Is(err, ErrNamespaceBusy) {
				t.Error("reserved Open", err)
			}
			if err := e.ReloadID(d.ID); !errors.Is(err, ErrNamespaceBusy) {
				t.Error("reserved Reload", err)
			}
			saved := make(chan error, 1)
			go func() { saved <- e.SaveID(d.ID) }()
			close(release)
			err := <-done
			if fail && !errors.Is(err, denied) || !fail && err != nil {
				t.Fatal(err)
			}
			if err := <-saved; err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(filepath.Dir(p), "new")
			if fail {
				want = p
			}
			raw, err := os.ReadFile(want)
			if err != nil || string(raw) != "late draft" {
				t.Fatal(string(raw), err)
			}
			live, _ := e.Active()
			if live.ID != d.ID || live.Content.Path != want || live.Text != "late draft" || live.Dirty {
				t.Fatal("lost identity/edit", live)
			}
			if !fail {
				if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
					t.Fatal("old path recreated", err)
				}
			}
			if e.reservation != nil {
				t.Fatal("leaked reservation")
			}
		})
	}
}
func TestRecoveryAutosaveAfterFSRenameUsesNewPath(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e, p, d := recoveryFile(t)
		entered, release := make(chan struct{}), make(chan struct{})
		written := make(chan string, 1)
		e.writeFile = func(path, text string) error {
			err := SaveExisting(path, text)
			if err == nil {
				written <- path
			}
			return err
		}
		e.renamePath = func(path, name string) (string, error) {
			next, err := RenamePath(path, name)
			if err != nil {
				return next, err
			}
			close(entered)
			<-release
			return next, nil
		}
		req := recoveryRequest(e, p, "new")
		done := make(chan error, 1)
		go func() { _, err := e.RenameDocuments(req); done <- err }()
		<-entered
		if err := e.SetTextID(d.ID, "timer draft"); err != nil {
			t.Fatal(err)
		}
		time.Sleep(AutosaveDelay)
		close(release)
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		path := <-written
		synctest.Wait()
		want := filepath.Join(filepath.Dir(p), "new")
		if path != want {
			t.Fatal("timer wrote old path", path)
		}
		raw, err := os.ReadFile(want)
		if err != nil || string(raw) != "timer draft" {
			t.Fatal(string(raw), err)
		}
		if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("timer recreated old path", err)
		}
	})
}
func TestRecoveryAdmissionRejectsChangedDirtyClosedAndDestination(t *testing.T) {
	for _, mode := range []string{"changed_id", "dirty", "closed", "destination"} {
		t.Run(mode, func(t *testing.T) {
			e, p, d := recoveryFile(t)
			e.delay = time.Hour
			r := recoveryRequest(e, p, "new")
			want := ErrRenameChanged
			switch mode {
			case "changed_id":
				e.CloseID(d.ID)
				e.Open(p)
			case "dirty":
				e.SetTextID(d.ID, "draft")
				want = ErrRenameDirty
			case "closed":
				e.CloseAll()
				want = ErrEditorClosed
			case "destination":
				dest := filepath.Join(filepath.Dir(p), "new")
				os.WriteFile(dest, []byte("dest"), 0600)
				e.OpenInactive(dest)
				want = ErrDestinationExists
			}
			_, err := e.RenameDocuments(r)
			if !errors.Is(err, want) {
				t.Fatal(err, want)
			}
			if _, err := os.Stat(p); err != nil {
				t.Fatal("refused rename changed filesystem", err)
			}
		})
	}
}
func TestRecoveryReloadSequenceAndRenameEpoch(t *testing.T) {
	for _, mode := range []string{"newer_read", "away_back", "failed_rename"} {
		t.Run(mode, func(t *testing.T) {
			e, p, d := recoveryFile(t)
			entered, release := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			e.readFile = func(path string) FileContent {
				c := LoadFile(path)
				if calls.Add(1) == 1 {
					close(entered)
					<-release
				}
				return c
			}
			done := make(chan error, 1)
			go func() { done <- e.ReloadID(d.ID) }()
			<-entered
			want := ErrNamespaceChanged
			switch mode {
			case "newer_read":
				os.WriteFile(p, []byte("new content"), 0600)
				if err := e.ReloadID(d.ID); err != nil {
					t.Fatal(err)
				}
				want = ErrReadSuperseded
			case "away_back":
				r, err := e.RenameDocuments(recoveryRequest(e, p, "new"))
				if err != nil {
					t.Fatal(err)
				}
				if _, err = e.RenameDocuments(recoveryRequest(e, r.Path, "old")); err != nil {
					t.Fatal(err)
				}
			case "failed_rename":
				e.renamePath = func(string, string) (string, error) { return "", os.ErrPermission }
				if _, err := e.RenameDocuments(recoveryRequest(e, p, "new")); !errors.Is(err, os.ErrPermission) {
					t.Fatal(err)
				}
			}
			close(release)
			if err := <-done; !errors.Is(err, want) {
				t.Fatal("stale read applied", err, want)
			}
			e.readFile = nil
			if mode == "newer_read" {
				live, _ := e.Active()
				if live.Text != "new content" {
					t.Fatal("old read replaced new", live.Text)
				}
			}
		})
	}
}
func TestRecoveryCloseEpochControls(t *testing.T) {
	for _, mode := range []string{"close", "delete_empty", "failed_close", "already_closed"} {
		t.Run(mode, func(t *testing.T) {
			e, p, d := recoveryFile(t)
			other := filepath.Join(filepath.Dir(p), "other")
			os.WriteFile(other, []byte("other"), 0600)
			if mode == "failed_close" {
				e.delay = time.Hour
				e.SetTextID(d.ID, "draft")
				e.writeFile = func(string, string) error { return os.ErrPermission }
			}
			if mode == "already_closed" {
				e.CloseID(d.ID)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			e.readFile = func(path string) FileContent { c := LoadFile(path); close(entered); <-release; return c }
			done := make(chan error, 1)
			go func() { _, err := e.OpenInactive(other); done <- err }()
			<-entered
			epoch := e.namespaceEpoch
			var err error
			if mode == "delete_empty" {
				e.CloseUnder(filepath.Join(filepath.Dir(p), "missing"))
			} else {
				err = e.CloseID(d.ID)
			}
			close(release)
			readErr := <-done
			if mode == "failed_close" {
				if !errors.Is(err, os.ErrPermission) || readErr != nil || e.namespaceEpoch != epoch {
					t.Fatal(err, readErr)
				}
				live, _ := e.Active()
				if live.ID != d.ID || !live.Dirty || live.Text != "draft" {
					t.Fatal("failed close lost draft")
				}
				e.writeFile = nil
			} else if mode == "already_closed" {
				if err != nil || readErr != nil || e.namespaceEpoch != epoch {
					t.Fatal(err, readErr)
				}
			} else if err != nil || !errors.Is(readErr, ErrNamespaceChanged) || epoch == e.namespaceEpoch {
				t.Fatal(err, readErr)
			}
			e.readFile = nil
			if _, err := e.OpenInactive(other); err != nil {
				t.Fatal("explicit retry", err)
			}
		})
	}
}
func TestRecoveryCloseReceiptAndIdentityActivation(t *testing.T) {
	e, p, d := recoveryFile(t)
	e.delay = time.Hour
	e.SetTextID(d.ID, "draft")
	e.writeFile = func(string, string) error { return os.ErrPermission }
	r, err := e.CloseIDReceipt(d.ID)
	if !errors.Is(err, os.ErrPermission) || r != (CloseReceipt{}) {
		t.Fatal(r, err)
	}
	e.writeFile = nil
	if err = e.SaveID(d.ID); err != nil {
		t.Fatal(err)
	}
	renamed, err := e.RenameDocuments(recoveryRequest(e, p, "new"))
	if err != nil {
		t.Fatal(err)
	}
	r, err = e.CloseIDReceipt(d.ID)
	if err != nil || r != (CloseReceipt{d.ID, renamed.Path}) {
		t.Fatal(r, err)
	}
	again, err := e.CloseIDReceipt(d.ID)
	if err != nil || again != (CloseReceipt{}) {
		t.Fatal(again, err)
	}
	os.WriteFile(p, []byte("replacement"), 0600)
	replacement, _ := e.OpenInactive(p)
	if _, err := e.ActivateID(d.ID); !errors.Is(err, ErrDocumentNotOpen) {
		t.Fatal("activated replacement", err)
	}
	if got, err := e.ActivateID(replacement.ID); err != nil || got.ID != replacement.ID {
		t.Fatal(got, err)
	}
}
func TestRecoveryReopenVacancyReadFences(t *testing.T) {
	for _, mode := range []string{"normal", "occupied_before", "occupied_during", "rename_during", "owner_closed"} {
		t.Run(mode, func(t *testing.T) {
			e, p, original := recoveryFile(t)
			receipt, err := e.CloseIDReceipt(original.ID)
			if err != nil {
				t.Fatal(err)
			}
			if mode == "normal" {
				d, err := e.ReopenClosed(receipt)
				if err != nil || d.ID == original.ID || d.ID == 0 || d.Text != "original" {
					t.Fatal(d, err)
				}
				return
			}
			if mode == "occupied_before" {
				e.OpenInactive(p)
				d, err := e.ReopenClosed(receipt)
				if !errors.Is(err, ErrReopenOccupied) || d.ID != 0 {
					t.Fatal(d, err)
				}
				return
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var calls atomic.Int32
			e.readFile = func(path string) FileContent {
				c := LoadFile(path)
				if calls.Add(1) == 1 {
					close(entered)
					<-release
				}
				return c
			}
			type result struct {
				d   Document
				err error
			}
			done := make(chan result, 1)
			go func() { d, err := e.ReopenClosed(receipt); done <- result{d, err} }()
			<-entered
			want := ErrReopenOccupied
			switch mode {
			case "occupied_during":
				if _, err := e.OpenInactive(p); err != nil {
					t.Fatal(err)
				}
			case "rename_during":
				if _, err := e.RenameDocuments(recoveryRequest(e, p, "new")); err != nil {
					t.Fatal(err)
				}
				want = ErrNamespaceChanged
			case "owner_closed":
				if err := e.CloseAll(); err != nil {
					t.Fatal(err)
				}
				want = ErrEditorClosed
			}
			close(release)
			got := <-done
			if !errors.Is(got.err, want) || got.d.ID != 0 {
				t.Fatal("adopted/stale reopen", got, want)
			}
			e.readFile = nil
		})
	}
}
func TestRecoveryRenameAdmissionRejectsNewMember(t *testing.T) {
	e := NewEditor()
	defer e.CloseAll()
	dir := t.TempDir()
	p := filepath.Join(dir, "first")
	os.WriteFile(p, []byte("one"), 0600)
	e.Open(p)
	r := recoveryRequest(e, dir, "renamed")
	other := filepath.Join(dir, "second")
	os.WriteFile(other, []byte("two"), 0600)
	e.OpenInactive(other)
	if _, err := e.RenameDocuments(r); !errors.Is(err, ErrRenameChanged) {
		t.Fatal("new member accepted", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("filesystem modified", err)
	}
}
func TestRecoveryCloseAllFailureAndClosedIdentity(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[fail], func(t *testing.T) {
			e, _, d := recoveryFile(t)
			e.delay = time.Hour
			e.SetTextID(d.ID, "draft")
			var writes int
			e.writeFile = func(p, text string) error {
				writes++
				if fail {
					return os.ErrPermission
				}
				return SaveExisting(p, text)
			}
			err := e.CloseAll()
			if fail {
				if !errors.Is(err, os.ErrPermission) {
					t.Fatal(err)
				}
				live, _ := e.Active()
				if live.ID != d.ID || !live.Dirty || live.Text != "draft" || e.closed {
					t.Fatal("failed CloseAll unrecoverable")
				}
				fail = false
				if err := e.CloseID(d.ID); err != nil {
					t.Fatal("repair close", err)
				}
			} else {
				if err != nil || !e.closed {
					t.Fatal(err)
				}
				before := writes
				if err := e.CloseID(d.ID); err != nil || writes != before {
					t.Fatal("repeat close saved again", err)
				}
				if err := e.SaveID(d.ID); !errors.Is(err, ErrEditorClosed) {
					t.Fatal("SaveID after closed", err)
				}
			}
		})
	}
}
func TestRecoverySaveIDNeverWritesReplacement(t *testing.T) {
	e, p, d := recoveryFile(t)
	e.delay = time.Hour
	r, err := e.RenameDocuments(recoveryRequest(e, p, "new"))
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(p, []byte("replacement"), 0600)
	replacement, _ := e.OpenInactive(p)
	e.SetTextID(d.ID, "original draft")
	e.SetTextID(replacement.ID, "replacement draft")
	if err := e.SaveID(d.ID); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(p)
	moved, _ := os.ReadFile(r.Path)
	if string(raw) != "replacement" || string(moved) != "original draft" {
		t.Fatal("identity save retargeted", string(raw), string(moved))
	}
}
