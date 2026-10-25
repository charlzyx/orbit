package worktreeui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"testing"
)

func TestWorkInRowsAndNativeMenuSelection(t *testing.T) {
	root := t.TempDir()
	linked := filepath.Join(root, ".wt", "task")
	var opened string
	s := New(root, Options{OpenWorkspace: func(p string) { opened = p }})
	s.initialized = true
	s.Rows = []workspace.Worktree{{Name: "root", Path: root, Main: true}, {Name: "task", Path: linked}}
	rows := s.WorkInRows()
	if len(rows) != 3 || !rows[0].Selected || !rows[2].Create {
		t.Fatal(rows)
	}
	s.Workspace = linked
	rows = s.WorkInRows()
	if !rows[1].Selected || rows[0].Selected {
		t.Fatal(rows)
	}
	tester := ui.NewTester(func(c *ui.Context) { WorkInPicker(c, s, nil) }, 600, 400)
	tester.Frame()
	if e := tester.Click("Work in: task"); e != nil {
		t.Fatal(e)
	}
	tester.CloseMenu()
	if opened != "" {
		t.Fatal("Escape selected worktree")
	}
	if e := tester.Click("Work in: task"); e != nil {
		t.Fatal(e)
	}
	if e := tester.ChooseMenuItem("Local"); e != nil {
		t.Fatal(e)
	}
	if opened != root {
		t.Fatal(opened)
	}
}
