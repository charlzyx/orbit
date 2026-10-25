package worktreeui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"testing"
	"time"
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
	tester.Key(0, ui.KeyEscape)
	tester.Advance(250 * time.Millisecond)
	if opened != "" {
		t.Fatal("Escape selected worktree")
	}
	if e := tester.Click("Work in: task"); e != nil {
		t.Fatal(e)
	}
	if e := tester.Click("Local"); e != nil {
		t.Fatal(e)
	}
	if opened != root {
		t.Fatal(opened)
	}
}

func TestWorkInKeyboardWrapNoopAndDialogFocusGuard(t *testing.T) {
	root := t.TempDir()
	linked := filepath.Join(root, ".wt", "task")
	var opened []string
	focusCalls, openCalls, createCalls := 0, 0, 0
	s := New(root, Options{OpenWorkspace: func(p string) { opened = append(opened, p) }, FocusComposer: func() { focusCalls++ }, OnPickerOpen: func() { openCalls++ }})
	defer s.Close()
	s.initialized = true
	s.Rows = []workspace.Worktree{{Name: "root", Path: root, Main: true}, {Name: "task", Path: linked}}
	tt := ui.NewTester(func(c *ui.Context) { WorkInPicker(c, s, func() { createCalls++ }) }, 700, 500)
	open := func() {
		t.Helper()
		if err := tt.Click("Work in: Local"); err != nil {
			t.Fatal(err)
		}
		if !s.workInOpen {
			t.Fatal("picker did not open")
		}
	}
	open()
	tt.Key(0, ui.KeyEnter)
	if s.workInOpen || len(opened) != 0 {
		t.Fatal("selecting current worktree mutated workspace", opened)
	}
	open()
	tt.Key(0, ui.KeyUp)
	if s.workInCursor != 2 {
		t.Fatal("Up did not wrap to New", s.workInCursor)
	}
	tt.Key(0, ui.KeyDown)
	if s.workInCursor != 0 {
		t.Fatal("Down did not wrap", s.workInCursor)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if len(opened) != 1 || opened[0] != linked {
		t.Fatal("keyboard chose wrong worktree", opened)
	}
	open()
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	if !s.dialogOpen || createCalls != 1 || focusCalls != 0 {
		t.Fatal("New must retain dialog focus, not request composer focus", s.dialogOpen, createCalls, focusCalls)
	}
	s.workInOpen = true
	tt.Frame()
	if s.workInOpen || focusCalls != 0 {
		t.Fatal("existing modal did not suppress popup safely")
	}
	if openCalls != 3 {
		t.Fatal("open hook count", openCalls)
	}
}
