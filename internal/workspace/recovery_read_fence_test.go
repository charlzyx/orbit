package workspace

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestRecoveryHeldReadCannotResurrectDeletedNamespace(t *testing.T) {
	e := NewEditor()
	defer e.CloseAll()
	p := filepath.Join(t.TempDir(), "a")
	if err := os.WriteFile(p, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	e.readFile = func(path string) FileContent { c := LoadFile(path); close(entered); <-release; return c }
	done := make(chan error, 1)
	go func() { _, err := e.OpenInactive(p); done <- err }()
	<-entered
	if err := os.Remove(p); err != nil {
		t.Fatal(err)
	}
	e.CloseUnder(p)
	close(release)
	err := <-done
	if err == nil || len(e.Documents()) != 0 {
		t.Fatalf("stale deleted namespace registered: err=%v docs=%+v", err, e.Documents())
	}
}
func TestRecoveryHeldReadCannotUndoNewerClose(t *testing.T) {
	e := NewEditor()
	defer e.CloseAll()
	p := filepath.Join(t.TempDir(), "a")
	if err := os.WriteFile(p, []byte("old"), 0600); err != nil {
		t.Fatal(err)
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
	done := make(chan error, 1)
	go func() { _, err := e.OpenInactive(p); done <- err }()
	<-entered
	if _, err := e.OpenInactive(p); err != nil {
		t.Fatal(err)
	}
	if err := e.Close(p); err != nil {
		t.Fatal(err)
	}
	close(release)
	err := <-done
	if err == nil || len(e.Documents()) != 0 {
		t.Fatalf("held read undid close: err=%v docs=%+v", err, e.Documents())
	}
}
