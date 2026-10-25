package workspace

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func wtTestRepo(t *testing.T) string {
	t.Helper()
	if _, e := exec.LookPath("git"); e != nil {
		t.Skip("git unavailable")
	}
	root := t.TempDir()
	g := Git{Dir: root}
	ctx := context.Background()
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Orbit Test"}, {"config", "user.email", "orbit@example.invalid"}, {"config", "commit.gpgsign", "false"}, {"config", "core.hooksPath", filepath.Join(root, "no-hooks")}} {
		if _, e := g.Run(ctx, args...); e != nil {
			t.Fatal(e)
		}
	}
	os.WriteFile(filepath.Join(root, "seed.txt"), []byte("seed\n"), 0600)
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte(".wt/\n"), 0600)
	g.Run(ctx, "add", ".")
	if _, e := g.Run(ctx, "commit", "-m", "seed"); e != nil {
		t.Fatal(e)
	}
	return root
}
func TestWorktreeLifecyclePreservesBranches(t *testing.T) {
	root := wtTestRepo(t)
	m := &WorktreeManager{Dir: root, Config: DefaultWorktreeConfig()}
	ctx := context.Background()
	r, e := m.Create(ctx, CreateWorktreeOptions{Name: "task-1", Branch: "feature/one", NewBranch: true})
	if e != nil {
		t.Fatal(e)
	}
	if r.Worktree.Name != "task-1" || r.Worktree.Branch != "feature/one" || r.Worktree.Main {
		t.Fatal(r)
	}
	if _, e = m.Create(ctx, CreateWorktreeOptions{Name: "task-1", Branch: "feature/two", NewBranch: true}); e == nil {
		t.Fatal("overwrite")
	}
	if e = m.Lock(ctx, r.Worktree.Path, "work in progress"); e != nil {
		t.Fatal(e)
	}
	if e = m.Lock(ctx, r.Worktree.Path, "again"); e != nil {
		t.Fatal(e)
	}
	if e = m.Remove(ctx, r.Worktree.Path, true); e == nil {
		t.Fatal("locked force removal")
	}
	if e = m.Unlock(ctx, r.Worktree.Path); e != nil {
		t.Fatal(e)
	}
	if e = m.Unlock(ctx, r.Worktree.Path); e != nil {
		t.Fatal(e)
	}
	dst, e := m.Rename(ctx, r.Worktree.Path, "renamed")
	if e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(dst, "untracked"), []byte("dirty"), 0600)
	if e = m.Remove(ctx, dst, false); e == nil {
		t.Fatal("dirty removal")
	}
	if e = m.Remove(ctx, dst, true); e != nil {
		t.Fatal(e)
	}
	if _, e = (Git{Dir: root}).Run(ctx, "show-ref", "--verify", "refs/heads/feature/one"); e != nil {
		t.Fatal("branch deleted", e)
	}
	if e = m.Remove(ctx, root, true); e == nil {
		t.Fatal("main deleted")
	}
	if e = m.Prune(ctx); e != nil {
		t.Fatal(e)
	}
	if e = m.Repair(ctx); e != nil {
		t.Fatal(e)
	}
}
func TestWorktreeBranchCollisionAndCancelledCreate(t *testing.T) {
	root := wtTestRepo(t)
	m := &WorktreeManager{Dir: root}
	ctx := context.Background()
	if _, e := m.Create(ctx, CreateWorktreeOptions{Name: "main-copy", Branch: "main"}); e == nil {
		t.Fatal("checked-out branch")
	}
	for _, n := range []string{"../escape", "", `.\bad`} {
		if _, e := m.ResolvePath(ctx, n); e == nil {
			t.Fatal("bad name", n)
		}
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if _, e := m.Create(c, CreateWorktreeOptions{Name: "cancelled", Branch: "cancelled", NewBranch: true}); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(root, ".wt", "cancelled")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("cancel created")
	}
}
func TestWorktreeSetupConsentAndFailureKeepsTree(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX script")
	}
	root := wtTestRepo(t)
	os.MkdirAll(filepath.Join(root, ".orbit"), 0700)
	script := filepath.Join(root, DefaultSetupScript)
	os.WriteFile(script, []byte("printf '%s' \"$ORBIT_ROOT_PATH\" > root-path\nprintf '%s' \"$ORBIT_WORKTREE_PATH\" > worktree-path\necho failed >&2\nexit 7\n"), 0600)
	m := &WorktreeManager{Dir: root, Config: DefaultWorktreeConfig()}
	ctx := context.Background()
	opts := CreateWorktreeOptions{Name: "setup", Branch: "setup", NewBranch: true}
	if _, e := m.Create(ctx, opts); !errors.Is(e, ErrConfirmationRequired) {
		t.Fatal(e)
	}
	if _, e := os.Stat(filepath.Join(root, ".wt", "setup")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("created before consent")
	}
	opts.ConfirmSetup = true
	r, e := m.Create(ctx, opts)
	if e != nil {
		t.Fatal(e)
	}
	if r.Setup == nil || r.Setup.Success || r.Setup.ExitCode == nil || *r.Setup.ExitCode != 7 {
		t.Fatal(r)
	}
	if _, e = os.Stat(r.Worktree.Path); e != nil {
		t.Fatal("tree lost", e)
	}
	b, _ := os.ReadFile(filepath.Join(r.Worktree.Path, "root-path"))
	if string(b) != root {
		t.Fatal(string(b))
	}
	b, _ = os.ReadFile(filepath.Join(r.Worktree.Path, "worktree-path"))
	if string(b) != r.Worktree.Path {
		t.Fatal(string(b))
	}
}
func TestWorktreeConfigAndNULParser(t *testing.T) {
	root := t.TempDir()
	c := DefaultWorktreeConfig()
	c.AllowSetup(root)
	c.AllowSetup(filepath.Join(root, "."))
	if len(c.AllowedSetupRepos) != 1 {
		t.Fatal(c)
	}
	p := filepath.Join(root, "config.json")
	if e := SaveWorktreeConfig(p, c); e != nil {
		t.Fatal(e)
	}
	if !LoadWorktreeConfig(p).SetupAllowed(root) {
		t.Fatal("lost trust")
	}
	os.MkdirAll(filepath.Join(root, ".orbit"), 0700)
	os.WriteFile(filepath.Join(root, ".orbit", "worktree.json"), []byte(`{"directory":"custom","setup_script":"setup.sh"}`), 0600)
	dir, script := c.Effective(root)
	if dir != filepath.Join(root, "custom") || script != filepath.Join(root, "setup.sh") {
		t.Fatal(dir, script)
	}
	raw := []byte("worktree /tmp/main\x00HEAD abc\x00branch refs/heads/main\x00\x00worktree /tmp/space\nname\x00detached\x00locked reason\x00prunable gone\x00\x00")
	ws := ParseWorktrees(raw)
	if len(ws) != 2 || !ws[0].Main || ws[1].Main || !ws[1].Detached || ws[1].LockReason != "reason" || !strings.Contains(ws[1].Path, "\n") {
		t.Fatal(ws)
	}
}
func TestSetupCancellationAndConfirmation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX script")
	}
	root := t.TempDir()
	script := filepath.Join(root, "setup.sh")
	os.WriteFile(script, []byte("sleep 30\n"), 0600)
	o := RunSetupScript(context.Background(), script, root, root, false)
	if o.Success || o.Stderr != ErrConfirmationRequired.Error() {
		t.Fatal(o)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	start := time.Now()
	o = RunSetupScript(ctx, script, root, root, true)
	if !o.Cancelled || o.Success || time.Since(start) > 4*time.Second {
		t.Fatal(o, time.Since(start))
	}
}

func TestWorktreeConfigEmptyAndMalformedRepoOverrides(t *testing.T) {
	root := t.TempDir()
	cfg := DefaultWorktreeConfig()
	cfg.SetupScript = ""
	_, script := cfg.Effective(root)
	if script != root {
		t.Fatal("empty script reset to default", script)
	}
	path := filepath.Join(root, "worktrees.json")
	if e := SaveWorktreeConfig(path, cfg); e != nil {
		t.Fatal(e)
	}
	cfg = LoadWorktreeConfig(path)
	_, script = cfg.Effective(root)
	if script != root {
		t.Fatal(script)
	}
	os.MkdirAll(filepath.Join(root, ".orbit"), 0700)
	os.WriteFile(filepath.Join(root, ".orbit", "worktree.json"), []byte(`{"directory":"unexpected","setup_script":3}`), 0600)
	dir, _ := cfg.Effective(root)
	if dir != filepath.Join(root, ".wt") {
		t.Fatal("partially applied malformed config", dir)
	}
}

func TestWorktreeSlugMatchesSourceSuggestions(t *testing.T) {
	for input, want := range map[string]string{"feat/ui-make-minimal": "feat-ui-make-minimal", "113": "113", "Fix: Login Timeout!": "fix-login-timeout", "a__b.c": "a__b.c", "--": "", "...": "", "": "", "日本語": ""} {
		if got := WorktreeSlug(input); got != want {
			t.Errorf("%q: got %q, want %q", input, got, want)
		}
	}
}

func TestWorktreeMoveResolvesMainRootAndRenameConfiguredRoot(t *testing.T) {
	root := wtTestRepo(t)
	m := &WorktreeManager{Dir: root, Config: DefaultWorktreeConfig()}
	ctx := context.Background()
	created, err := m.Create(ctx, CreateWorktreeOptions{Name: "original", Branch: "feature", NewBranch: true})
	if err != nil {
		t.Fatal(err)
	}
	linked := &WorktreeManager{Dir: created.Worktree.Path, Config: DefaultWorktreeConfig()}
	if dest, err := linked.ResolveDestination(ctx, "relative-tree"); err != nil || dest != filepath.Join(root, "relative-tree") {
		t.Fatal(dest, err)
	}
	if err := m.Move(ctx, created.Worktree.Path, "external-tree"); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(root, "external-tree")
	cfg := DefaultWorktreeConfig()
	cfg.Directory = "custom-root"
	if err := os.Mkdir(filepath.Join(root, "custom-root"), 0700); err != nil {
		t.Fatal(err)
	}
	m.Config = cfg
	renamed, err := m.Rename(ctx, external, "renamed")
	if err != nil || renamed != filepath.Join(root, "custom-root", "renamed") {
		t.Fatal(renamed, err)
	}
	if _, err := os.Stat(external); !os.IsNotExist(err) {
		t.Fatal("external path not moved", err)
	}
	if w, err := m.FindByPath(ctx, renamed); err != nil || w.Branch != "feature" {
		t.Fatal(w, err)
	}
	if err := m.Move(ctx, renamed, ""); err == nil {
		t.Fatal("empty destination accepted")
	}
}
