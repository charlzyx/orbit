package worktreeui

import (
	"context"
	"github.com/charlzyx/orbit/internal/workspace"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func repo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	g := workspace.Git{Dir: root}
	for _, a := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Test"}, {"config", "user.email", "test@example.invalid"}, {"config", "commit.gpgsign", "false"}, {"config", "core.hooksPath", filepath.Join(root, "no-hooks")}} {
		if _, e := g.Run(context.Background(), a...); e != nil {
			t.Fatal(e)
		}
	}
	os.WriteFile(filepath.Join(root, "file"), []byte("hello\n"), 0600)
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".wt/\n"), 0600)
	g.Run(context.Background(), "add", ".")
	if _, e := g.Run(context.Background(), "commit", "-m", "initial"); e != nil {
		t.Fatal(e)
	}
	return root
}
func wait(t *testing.T, s *State) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for s.Busy && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond * 5)
		s.Tick()
	}
	if s.Busy {
		t.Fatal("operation timed out", s.Activity)
	}
}
func TestRealWorktreePageLifecycle(t *testing.T) {
	root := repo(t)
	var opened string
	s := New(root, Options{OpenWorkspace: func(path string) { opened = path }})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	if s.Error != nil || len(s.Rows) != 1 || s.Root != root {
		t.Fatal(s.Error, s.Rows, s.Root)
	}
	o := workspace.CreateWorktreeOptions{Name: "task", Branch: "task", NewBranch: true}
	if !s.Create(o, true) || s.Create(o, true) {
		t.Fatal("busy guard")
	}
	wait(t, s)
	if s.Error != nil || len(s.Rows) != 2 || opened != filepath.Join(root, ".wt", "task") {
		t.Fatal(s.Error, s.Rows, opened)
	}
	path := opened
	s.Lock(path, "retain")
	wait(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	if s.Remove(path, true, false) {
		t.Fatal("missing confirmation")
	}
	s.Remove(path, true, true)
	wait(t, s)
	if s.Error == nil {
		t.Fatal("locked removal")
	}
	s.Unlock(path)
	wait(t, s)
	s.Rename(path, "renamed")
	wait(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	path = filepath.Join(root, ".wt", "renamed")
	s.Remove(path, false, true)
	wait(t, s)
	if s.Error != nil || len(s.Rows) != 1 {
		t.Fatal(s.Error, s.Rows)
	}
}
func TestSetupConsentFailureAndRetry(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX script fixture")
	}
	root := repo(t)
	os.MkdirAll(filepath.Join(root, ".orbit"), 0700)
	script := filepath.Join(root, ".orbit", "worktree-setup.sh")
	os.WriteFile(script, []byte("echo failed; exit 3\n"), 0600)
	s := New(root, Options{})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	o := workspace.CreateWorktreeOptions{Name: "task", Branch: "task", NewBranch: true, RunSetup: true}
	s.Create(o, false)
	wait(t, s)
	if !s.confirmOpen || s.pending == nil {
		t.Fatal("setup consent missing", s.Error)
	}
	if _, e := os.Stat(filepath.Join(root, ".wt", "task")); !os.IsNotExist(e) {
		t.Fatal("created before consent")
	}
	o.ConfirmSetup = true
	s.confirmOpen = false
	s.Create(o, false)
	wait(t, s)
	if s.Error != nil || s.Setup == nil || s.Setup.Success || !strings.Contains(s.Setup.Stdout, "failed") {
		t.Fatal(s.Error, s.Setup)
	}
	if _, e := os.Stat(filepath.Join(root, ".wt", "task")); e != nil {
		t.Fatal("failed setup removed tree")
	}
	os.WriteFile(script, []byte("echo success\n"), 0600)
	if s.RetrySetup(false) {
		t.Fatal("setup no confirmation")
	}
	s.RetrySetup(true)
	wait(t, s)
	if s.Setup == nil || !s.Setup.Success {
		t.Fatal(s.Setup)
	}
}
func TestOldWorkspaceResultAndCancelNeverApplied(t *testing.T) {
	s := New(t.TempDir(), Options{})
	old := s.generation
	s.SetWorkspace(t.TempDir())
	s.results <- result{generation: old, root: "wrong", rows: []workspace.Worktree{{Path: "wrong"}}, status: "stale"}
	s.Tick()
	if s.Root != "" || len(s.Rows) > 0 || s.Status == "stale" {
		t.Fatal("stale workspace result")
	}
	s.results <- result{generation: s.generation, status: "cancelled"}
	s.Cancel()
	s.Tick()
	if s.Status == "cancelled" {
		t.Fatal("cancel result applied")
	}
	s.Close()
	if s.Refresh() {
		t.Fatal("closed controller started job")
	}
}

func TestRefreshPreservesFailedActionAndNoFalseSuccess(t *testing.T) {
	root := repo(t)
	s := New(root, Options{})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	s.Remove(root, true, true)
	wait(t, s)
	if s.Error == nil || s.Status != "" {
		t.Fatal(s.Error, s.Status)
	}
	message := s.Error.Error()
	s.Refresh()
	wait(t, s)
	if s.Error == nil || s.Error.Error() != message {
		t.Fatal("refresh hid action error", s.Error)
	}
}

func TestExistingRemoteBranchAvailableWithoutNetwork(t *testing.T) {
	root := repo(t)
	g := workspace.Git{Dir: root}
	if _, e := g.Run(context.Background(), "update-ref", "refs/remotes/origin/task", "HEAD"); e != nil {
		t.Fatal(e)
	}
	s := New(root, Options{})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	found := false
	for _, branch := range s.Branches {
		found = found || branch == "origin/task"
	}
	if !found {
		t.Fatal(s.Branches)
	}
	s.Create(workspace.CreateWorktreeOptions{Name: "remote", Branch: "origin/task"}, false)
	wait(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	if len(s.Rows) != 2 || !s.Rows[1].Detached {
		t.Fatal(s.Rows)
	}
}

func TestFailedLockRetryDoesNotRepeatPreviousCreate(t *testing.T) {
	root := repo(t)
	s := New(root, Options{})
	defer s.Close()
	s.Create(workspace.CreateWorktreeOptions{Name: "task", Branch: "task", NewBranch: true}, false)
	wait(t, s)
	if s.Error != nil {
		t.Fatal(s.Error)
	}
	s.Lock(filepath.Join(root, "missing"), "test")
	wait(t, s)
	if s.Error == nil || s.lastAction == nil {
		t.Fatal("expected retryable lock failure")
	}
	failure := s.Error.Error()
	s.lastAction()
	wait(t, s)
	if s.Error == nil || s.Error.Error() != failure {
		t.Fatal("retry executed a different operation", failure, s.Error)
	}
}

func TestSettingsPersistenceFailureIsTransactionalAndRetryable(t *testing.T) {
	root := t.TempDir()
	blocker := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	s := New(root, Options{ConfigPath: filepath.Join(blocker, "worktrees.json")})
	defer s.Close()
	before := s.Config.Directory
	next := s.Config
	next.Directory = "must-not-commit-on-error"
	if s.applyConfig(next) || s.Error == nil || s.Config.Directory != before {
		t.Fatal(s.Config, s.Error)
	}
	if s.lastAction == nil {
		t.Fatal("missing retry")
	}
	if err := os.Remove(blocker); err != nil {
		t.Fatal(err)
	}
	s.lastAction()
	if s.Error != nil || s.Config.Directory != next.Directory {
		t.Fatal(s.Config, s.Error)
	}
	if got := workspace.LoadWorktreeConfig(s.Options.ConfigPath); got.Directory != next.Directory {
		t.Fatal(got)
	}
}

func TestWorkspaceChangeClearsRetryAndPresentationState(t *testing.T) {
	s := New(t.TempDir(), Options{})
	defer s.Close()
	s.lastAction = func() { t.Fatal("stale action") }
	s.Filter = "old-filter"
	s.ShowOutput = true
	s.SetWorkspace(t.TempDir())
	if s.lastAction != nil || s.Filter != "" || s.ShowOutput {
		t.Fatal("old workspace state retained")
	}
}

func TestCreateSuggestionsPreserveManualBranchAndModes(t *testing.T) {
	root := t.TempDir()
	s := New(root, Options{})
	defer s.Close()
	s.Root = root
	s.Rows = []workspace.Worktree{{Path: root, Name: "main", Branch: "feat/start", Main: true}, {Name: "feat-start", Path: filepath.Join(root, ".wt", "feat-start")}}
	s.Branches = []string{"feat/start", "spare", "origin/remote"}
	s.quickCreateDialog()
	if s.name != "feat-start-2" || s.branch != "orbit/feat-start-2" || s.existingBranch != "spare" {
		t.Fatal(s.name, s.branch, s.existingBranch)
	}
	s.name = "New Task"
	s.deriveBranch()
	if s.branch != "orbit/new-task" {
		t.Fatal(s.branch)
	}
	s.branch = "manual/task"
	s.branchTouched = true
	s.name = "Another"
	s.deriveBranch()
	if s.branch != "manual/task" {
		t.Fatal("manual branch overwritten", s.branch)
	}
	s.newBranch = false
	if o := s.createOptions(); o.Branch != "spare" || o.NewBranch {
		t.Fatal(o)
	}
	s.newBranch = true
	if o := s.createOptions(); o.Branch != "manual/task" || !o.NewBranch {
		t.Fatal(o)
	}
	s.createDialog()
	if s.branchTouched || s.name != "" || s.branch != "" {
		t.Fatal("creation state not reset")
	}
}

func TestBranchAlreadyUsedOpensExistingWorktreeDialog(t *testing.T) {
	root := repo(t)
	var opened string
	s := New(root, Options{OpenWorkspace: func(p string) { opened = p }})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	s.Create(workspace.CreateWorktreeOptions{Name: "cannot-use-main", Branch: "main"}, false)
	wait(t, s)
	if s.dialog != "branch_in_use" || !s.dialogOpen || s.target.Path != root || s.target.Branch != "main" {
		t.Fatal(s.dialog, s.target, s.Error)
	}
	if opened != "" {
		t.Fatal("navigation before confirmation", opened)
	}
	if _, err := os.Stat(filepath.Join(root, ".wt", "cannot-use-main")); !os.IsNotExist(err) {
		t.Fatal("created duplicate branch worktree", err)
	}
}

func TestSetupAllowRemembersCanonicalRepositoryBeforeCreation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	root := repo(t)
	os.Mkdir(filepath.Join(root, ".orbit"), 0700)
	os.WriteFile(filepath.Join(root, ".orbit", "worktree-setup.sh"), []byte("echo setup\n"), 0600)
	cfgPath := filepath.Join(t.TempDir(), "worktrees.json")
	s := New(root, Options{ConfigPath: cfgPath})
	defer s.Close()
	s.Refresh()
	wait(t, s)
	s.Create(workspace.CreateWorktreeOptions{Name: "task", Branch: "task", NewBranch: true, RunSetup: true}, false)
	wait(t, s)
	if !s.confirmOpen || s.confirmLabel != "Allow & Create" {
		t.Fatal(s.confirmOpen, s.confirmLabel, s.Error)
	}
	s.confirmOpen = false
	s.confirmAction()
	wait(t, s)
	if s.Error != nil || s.Setup == nil || !s.Setup.Success {
		t.Fatal(s.Error, s.Setup)
	}
	if !workspace.LoadWorktreeConfig(cfgPath).SetupAllowed(root) {
		t.Fatal("accepted setup trust not persisted")
	}
	s.Create(workspace.CreateWorktreeOptions{Name: "second", Branch: "second", NewBranch: true, RunSetup: true}, false)
	wait(t, s)
	if s.Error != nil || s.confirmOpen || s.Setup == nil || !s.Setup.Success {
		t.Fatal("trusted repo asked again", s.Error, s.Setup)
	}
}
