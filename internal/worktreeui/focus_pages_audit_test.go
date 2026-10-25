package worktreeui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestAuditWorktreeCreateKeyboardCancel(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.initialized = true
	s.ShowHowTo = false
	s.Root = s.Workspace
	tt := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1100, 800)
	if e := tt.Click("Filter worktrees"); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 20 && !tt.Focused("New worktree"); i++ {
		tt.Key(0, ui.KeyTab)
		tt.Frame()
	}
	if !tt.Focused("New worktree") {
		t.Fatal("New worktree unreachable")
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !tt.Focused("Directory name") {
		t.Fatal("create did not focus name")
	}
	tt.Type("synthetic-not-created")
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Create a new branch") {
		t.Fatal("branch mode checkbox not reachable")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Branch name") {
		t.Fatal("branch unreachable from name")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.dialogOpen || s.Busy {
		t.Fatal("Escape failed or started operation")
	}
}
