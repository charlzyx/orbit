package worktreeui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestNativeWorktreeConfirmationCancelAndAccept(t *testing.T) {
	s := New("/test-only-fixture", Options{})
	s.initialized = true
	s.ShowHowTo = false
	calls := 0
	s.confirm("Run setup script?", "A test-only confirmation", func() { calls++ })
	tester := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1100, 800)
	tester.Frame()
	if calls != 0 {
		t.Fatal("action before confirmation")
	}
	tester.Key(0, ui.KeyEscape)
	if calls != 0 || s.confirmOpen {
		t.Fatal(calls, s.confirmOpen)
	}
	s.confirm("Run setup script?", "A test-only confirmation", func() { calls++ })
	tester.Frame()
	if e := tester.Click("Continue"); e != nil {
		t.Fatal(e)
	}
	if calls != 1 || s.confirmOpen {
		t.Fatal(calls, s.confirmOpen)
	}
}
func TestNativeWorktreeSettingsCancelPreservesConfig(t *testing.T) {
	s := New("/test-only-fixture", Options{})
	s.initialized = true
	s.ShowHowTo = false
	s.dialog = "settings"
	s.configDraft = s.Config
	s.configDraft.Directory = "changed-not-saved"
	s.dialogOpen = true
	tester := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1100, 800)
	tester.Frame()
	tester.Key(0, ui.KeyEscape)
	if s.Config.Directory == "changed-not-saved" || s.dialogOpen {
		t.Fatal("cancel committed draft", s.Config)
	}
}

func TestWorktreeTranslationFallbackAndInterpolation(t *testing.T) {
	s := New("fixture", Options{Translate: func(k string) string {
		if k == "worktree.operation.creating" {
			return "创建 %{name}…"
		}
		return k
	}})
	defer s.Close()
	if got := s.text("worktree.operation.creating", "Creating %{name}", "name", "任务"); got != "创建 任务…" {
		t.Fatal(got)
	}
	if got := s.text("absent.key", "fallback"); got != "fallback" {
		t.Fatal(got)
	}
}

func TestNativeCreateNameTypingDerivesBranchUntilManualEdit(t *testing.T) {
	s := New("/test-only-fixture", Options{})
	defer s.Close()
	s.initialized = true
	s.ShowHowTo = false
	s.createDialog()
	tester := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1100, 800)
	tester.Frame()
	if err := tester.Click("Directory name"); err != nil {
		t.Fatal(err)
	}
	tester.Type("Fix Login")
	if s.name != "Fix Login" || s.branch != "orbit/fix-login" || s.branchTouched {
		t.Fatal(s.name, s.branch, s.branchTouched)
	}
	if err := tester.Click("Branch name"); err != nil {
		t.Fatal(err)
	}
	tester.Type("-custom")
	if !s.branchTouched {
		t.Fatal("manual branch change not recorded")
	}
	manual := s.branch
	if err := tester.Click("Directory name"); err != nil {
		t.Fatal(err)
	}
	tester.Type(" Extra")
	if s.branch != manual {
		t.Fatal("name overwrote manual branch", s.branch, manual)
	}
	tester.Key(0, ui.KeyEscape)
	if s.dialogOpen || s.Busy {
		t.Fatal("cancel started creation")
	}
}

func TestNativeBranchInUseOpenRequiresExplicitClick(t *testing.T) {
	var opened string
	s := New("fixture", Options{OpenWorkspace: func(p string) { opened = p }})
	defer s.Close()
	s.initialized = true
	s.ShowHowTo = false
	s.dialog = "branch_in_use"
	s.dialogOpen = true
	s.target.Path = "existing-tree"
	s.target.Branch = "main"
	tester := ui.NewTester(func(c *ui.Context) { View(c, s) }, 1100, 800)
	tester.Frame()
	if opened != "" {
		t.Fatal("opened before click")
	}
	tester.Key(0, ui.KeyEscape)
	if opened != "" || s.dialogOpen {
		t.Fatal("escape navigated")
	}
	s.dialogOpen = true
	tester.Frame()
	if err := tester.Click("Open Existing Worktree"); err != nil {
		t.Fatal(err)
	}
	if opened != "existing-tree" || s.dialogOpen {
		t.Fatal(opened, s.dialogOpen)
	}
}
