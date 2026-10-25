package worktreeui

import (
	"testing"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// W50 source regression, initially observed failing with native MenuButton.
// Now exercises real key dispatch through the native popup; OS-specific focus
// and platform event dispatch still require their separate host acceptance.
func TestWorkInDismissalRestoresComposerFocus(t *testing.T) {
	root := t.TempDir()
	s := New(root, Options{OpenWorkspace: func(string) {}})
	defer s.Close()
	s.initialized = true
	s.Rows = []workspace.Worktree{{Name: "root", Path: root, Main: true}}
	draft := ""
	var composer ui.Handle
	s.Options.FocusComposer = func() { composer.Focus() }
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Children(func() { ui.TextArea(c, &draft).Label("Composer").Bind(&composer); WorkInPicker(c, s, nil) })
	}, 700, 500)
	if err := tt.Click("Composer"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Composer") {
		t.Fatal("fixture composer was not focused before opening picker")
	}
	for i := 0; i < 3; i++ {
		if err := tt.Click("Work in: Local"); err != nil {
			t.Fatal(err)
		}
		tt.Key(0, ui.KeyEscape)
		tt.Frame()
		if !tt.Focused("Composer") {
			t.Errorf("W50 FAIL: dismissing Work In did not restore composer focus; Rust app/worktrees.rs::close_work_in_menu explicitly focuses self.input")
		}
	}
}
