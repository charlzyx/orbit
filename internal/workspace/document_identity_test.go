package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentIdentityRenameReopenAndStaleEdit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "old")
	if err := os.WriteFile(p, []byte("old"), 0600); err != nil {
		t.Fatal(err)
	}
	e := NewEditor()
	defer e.CloseAll()
	original, err := e.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	newpath, err := RenamePath(p, "new")
	if err != nil {
		t.Fatal(err)
	}
	e.RenameUnder(p, newpath)
	renamed, _ := e.Active()
	if original.ID == 0 || renamed.ID != original.ID {
		t.Fatal("rename changed document identity")
	}
	if err := os.WriteFile(p, []byte("replacement"), 0600); err != nil {
		t.Fatal(err)
	}
	reopened, err := e.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.ID == original.ID {
		t.Fatal("new path occupant reused ID")
	}
	if err := e.SetTextID(original.ID, "original late edit"); err != nil {
		t.Fatal(err)
	}
	active, _ := e.Active()
	if active.Text != "replacement" {
		t.Fatal("ID edit targeted new oldpath", active)
	}
	if err := e.Close(newpath); err != nil {
		t.Fatal(err)
	}
	if err := e.SetTextID(original.ID, "stale"); err != ErrDocumentNotOpen {
		t.Fatal("closed ID accepted edit", err)
	}
	if err := e.Close(p); err != nil {
		t.Fatal(err)
	}
	again, err := e.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID == reopened.ID {
		t.Fatal("close reopen reused ID")
	}
}
