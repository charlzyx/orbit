package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func chooseMention(t *testing.T, a *App, tt *ui.Tester) {
	t.Helper()
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(a.text("composer.mention_file", "Mention file")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
}
func TestComposerAddMentionCaretUndoAndCompletion(t *testing.T) {
	a := fixture(t)
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "hello.txt"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	a.Model.Workspace = root
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("你好 尾")
	tt.Key(0, ui.KeyLeft)
	chooseMention(t, a, tt)
	if a.Draft != "你好 @尾" {
		t.Fatal(a.Draft)
	}
	deadline := time.Now().Add(3 * time.Second)
	for !a.autocomplete.IsOpen() && time.Now().Before(deadline) {
		tt.Advance(50 * time.Millisecond)
		time.Sleep(time.Millisecond)
	}
	if !a.autocomplete.IsOpen() {
		t.Fatal("mention did not open file completion", a.autocomplete.Snapshot())
	}
	tt.Key(0, ui.KeyEscape)
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != "你好 尾" {
		t.Fatal("mention not undoable", a.Draft)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
	if a.Draft != "你好 @尾" {
		t.Fatal("mention redo", a.Draft)
	}
}
func TestComposerAddMentionOldNativeMenuCannotRetarget(t *testing.T) {
	a := fixture(t)
	generation := uint64(1)
	a.Options.AttachmentContext = func() (agent.WorkflowTarget, uint64, bool) { return agent.WorkflowTarget{}, generation, true }
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	generation++
	tt.Frame()
	if err := tt.ChooseMenuItem(a.text("composer.mention_file", "Mention file")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.Draft != "" || len(a.pendingAttachmentPaths) != 0 {
		t.Fatal("old menu inserted into replacement", a.Draft)
	}
	chooseMention(t, a, tt)
	if a.Draft != "@" {
		t.Fatal("fresh menu unusable", a.Draft)
	}
}
func TestComposerAddMentionCancelAndRepeated(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Add attachment"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Frame()
	if a.Draft != "" {
		t.Fatal("cancel inserted")
	}
	chooseMention(t, a, tt)
	chooseMention(t, a, tt)
	if a.Draft != "@@" {
		t.Fatal("repeated menu", a.Draft)
	}
}

func TestComposerAddMentionIMECommitBeforeInsert(t *testing.T) {
	a := fixture(t)
	submits := 0
	a.Options.OnAction = func(x Action) {
		if x.Kind == "Submit" {
			submits++
		}
	}
	tt := ui.NewTester(a.View, 1200, 800)
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Type("before ")
	tt.Compose("你", 1)
	chooseMention(t, a, tt)
	if submits != 0 {
		t.Fatal("menu submitted draft")
	}
	// Pinned MyGo commits composition when clicking away to the menu. The
	// insertion must follow that commit, without requiring another TextInput.
	if a.Draft != "before 你@" || len(a.pendingAttachmentPaths) != 0 {
		t.Fatalf("IME/menu ordering: %q", a.Draft)
	}
	tt.Type("!")
	if a.Draft != "before 你@!" {
		t.Fatal("menu did not restore composer focus", a.Draft)
	}
}
