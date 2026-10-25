package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func testGit(t *testing.T) Git {
	t.Helper()
	g := Git{Dir: t.TempDir()}
	testRun(t, g, "init", "-b", "main")
	testRun(t, g, "config", "user.name", "Orbit Test")
	testRun(t, g, "config", "user.email", "orbit@example.invalid")
	testRun(t, g, "config", "commit.gpgsign", "false")
	return g
}
func testRun(t *testing.T, g Git, args ...string) string {
	t.Helper()
	s, e := g.text(context.Background(), args...)
	if e != nil {
		t.Fatalf("git %v: %v", args, e)
	}
	return s
}
func writeGitFile(t *testing.T, g Git, path, body string) {
	t.Helper()
	p := filepath.Join(g.Dir, path)
	if e := os.MkdirAll(filepath.Dir(p), 0700); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(p, []byte(body), 0600); e != nil {
		t.Fatal(e)
	}
}
func testCommit(t *testing.T, g Git, path, body, message string) string {
	t.Helper()
	writeGitFile(t, g, path, body)
	if _, e := g.Commit(context.Background(), message, true); e != nil {
		t.Fatal(e)
	}
	return testRun(t, g, "rev-parse", "HEAD")
}
func requireNoError(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}

func TestGitUnbornStageCommitAndUnstage(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	ok, e := g.HasCommits(ctx)
	requireNoError(t, e)
	if ok {
		t.Fatal("unborn HEAD has commits")
	}
	history, e := g.History(ctx, HistoryOptions{})
	requireNoError(t, e)
	if len(history) != 0 {
		t.Fatal(history)
	}
	writeGitFile(t, g, "a.txt", "one\ntwo")
	writeGitFile(t, g, "binary", "a\x00b")
	rows, e := g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 2 || rows[0].UnstagedAdditions != 2 || rows[1].UnstagedAdditions != 0 {
		t.Fatalf("status: %+v", rows)
	}
	requireNoError(t, g.Stage(ctx, "a.txt"))
	requireNoError(t, g.Stage(ctx, "a.txt"))
	requireNoError(t, g.Unstage(ctx, "a.txt"))
	rows, e = g.Status(ctx)
	requireNoError(t, e)
	if rows[0].Staged() {
		t.Fatal("still staged")
	}
	requireNoError(t, g.StageAll(ctx))
	requireNoError(t, g.UnstageAll(ctx))
	if _, e = g.Commit(ctx, "", true); e == nil {
		t.Fatal("empty message accepted")
	}
	if _, e = g.Commit(ctx, "nothing staged", false); e == nil {
		t.Fatal("empty index committed")
	}
	_, e = g.Commit(ctx, "feat: initial\n\nFull body", true)
	requireNoError(t, e)
	if _, e = g.Commit(ctx, "repeat", false); e == nil {
		t.Fatal("repeated commit silently succeeded")
	}
	d, e := g.CommitDetail(ctx, "HEAD")
	requireNoError(t, e)
	if d.Subject != "feat: initial" || !strings.Contains(d.Body, "Full body") || len(d.Files) != 2 || len(d.Parents) != 0 {
		t.Fatalf("detail: %+v", d)
	}
	rows, e = g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 0 {
		t.Fatal(rows)
	}
}

func TestGitStatusPreservesWhitespaceRenameAndLiteralPaths(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	old := " old\tname\n.txt"
	next := " new\tname\n.txt"
	testCommit(t, g, old, "one\n", "initial")
	testRun(t, g, "mv", "--", old, next)
	rows, e := g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 1 || rows[0].Path != next || rows[0].OrigPath != old || rows[0].Index != 'R' {
		t.Fatalf("rename lost: %+v", rows)
	}
	_, e = g.Commit(ctx, "rename", false)
	requireNoError(t, e)
	writeGitFile(t, g, "[x].txt", "literal\n")
	writeGitFile(t, g, "x.txt", "not selected\n")
	requireNoError(t, g.Stage(ctx, "[x].txt"))
	rows, e = g.Status(ctx)
	requireNoError(t, e)
	for _, r := range rows {
		if r.Path == "x.txt" && r.Staged() {
			t.Fatal("pathspec glob escaped literal boundary")
		}
	}
	for _, p := range []string{"../outside", "/tmp/outside", "a\x00b"} {
		if e := g.Stage(ctx, p); e == nil {
			t.Fatalf("invalid path accepted %q", p)
		}
	}
	if e := g.Stage(ctx); e == nil {
		t.Fatal("empty paths stage all")
	}
	if e := g.Checkout(ctx, "--orphan=evil"); e == nil {
		t.Fatal("option injection accepted")
	}
	if _, e := g.FileAtRevision(ctx, "HEAD", "../outside"); e == nil {
		t.Fatal("outside revision path accepted")
	}
}

func TestGitDiscardPreservesIndex(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "first\n", "initial")
	writeGitFile(t, g, "a", "staged\n")
	requireNoError(t, g.Stage(ctx, "a"))
	writeGitFile(t, g, "a", "unstaged\nsecond\n")
	rows, e := g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 1 || !rows[0].Staged() || !rows[0].Unstaged() || rows[0].UnstagedAdditions != 2 {
		t.Fatalf("mixed: %+v", rows)
	}
	writeGitFile(t, g, "new.txt", "untracked")
	requireNoError(t, g.Discard(ctx, "a", "new.txt"))
	b, e := os.ReadFile(filepath.Join(g.Dir, "a"))
	requireNoError(t, e)
	if string(b) != "staged\n" {
		t.Fatalf("discard lost index: %q", b)
	}
	if _, e = os.Stat(filepath.Join(g.Dir, "new.txt")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("untracked not discarded", e)
	}
	requireNoError(t, g.UnstageAll(ctx))
	requireNoError(t, g.DiscardAll(ctx))
	rows, e = g.Status(ctx)
	requireNoError(t, e)
	if len(rows) != 0 {
		t.Fatalf("not clean %+v", rows)
	}
	requireNoError(t, g.DiscardAll(ctx))
}

func TestGitHistoryBranchesAndGraph(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "one\n", "alpha")
	requireNoError(t, g.CreateBranch(ctx, "feature/slash"))
	testCommit(t, g, "b", "two\n", "Beta feature")
	requireNoError(t, g.RenameBranch(ctx, "feature/slash", "feature/renamed"))
	requireNoError(t, g.Checkout(ctx, "main"))
	if e := g.DeleteBranch(ctx, "feature/renamed", false); e == nil {
		t.Fatal("unmerged branch deleted")
	}
	history, e := g.History(ctx, HistoryOptions{AllBranches: true, Grep: "beta", Author: "Orbit", Path: "b"})
	requireNoError(t, e)
	if len(history) != 1 || history[0].Subject != "Beta feature" {
		t.Fatal(history)
	}
	history, e = g.History(ctx, HistoryOptions{AllBranches: true, Limit: 1, Skip: 1})
	requireNoError(t, e)
	if len(history) != 1 {
		t.Fatal(history)
	}
	recent, e := g.RecentBranches(ctx)
	requireNoError(t, e)
	if len(recent) == 0 {
		t.Fatal("recent branches empty")
	}
	refs, e := g.Refs(ctx)
	requireNoError(t, e)
	if len(refs) != 2 {
		t.Fatal(refs)
	}
	graph, e := g.GraphHistory(ctx, 100, false)
	requireNoError(t, e)
	if len(graph) != 2 {
		t.Fatal(graph)
	}
	rows := LayoutGraph(graph)
	if len(rows) != 2 {
		t.Fatal(rows)
	}
	requireNoError(t, g.DeleteBranch(ctx, "feature/renamed", true))
	if e := g.DeleteBranch(ctx, "main", true); e == nil {
		t.Fatal("current branch delete succeeded")
	}
	if e := g.CreateBranch(ctx, "bad name"); e == nil {
		t.Fatal("bad branch name accepted")
	}
}

func TestGraphOctopusAndWindowBoundary(t *testing.T) {
	entry := func(hash string, parents ...string) GraphCommit {
		return GraphCommit{CommitEntry: CommitEntry{Hash: hash}, Parents: parents}
	}
	rows := LayoutGraph([]GraphCommit{entry("merge", "a", "b", "c"), entry("a", "root"), entry("b", "root"), entry("c", "root"), entry("root", "outside")})
	if len(rows[0].Parents) != 3 || rows[0].LaneCount != 3 {
		t.Fatal(rows[0])
	}
	if len(rows[4].Parents) != 0 {
		t.Fatal("offscreen parent creates lane")
	}
	for _, row := range rows {
		if row.NodeLane >= row.LaneCount {
			t.Fatal(row)
		}
	}
}

func TestGitStashesPreserveUntrackedChoice(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "one\n", "initial")
	writeGitFile(t, g, "a", "two\n")
	writeGitFile(t, g, "untracked", "keep")
	_, e := g.PushStash(ctx, "tracked only", false)
	requireNoError(t, e)
	if _, e = os.Stat(filepath.Join(g.Dir, "untracked")); e != nil {
		t.Fatal(e)
	}
	stashes, e := g.Stashes(ctx)
	requireNoError(t, e)
	if len(stashes) != 1 || !strings.Contains(stashes[0].Message, "tracked only") {
		t.Fatal(stashes)
	}
	_, e = g.ApplyStash(ctx, 0, true)
	requireNoError(t, e)
	_, e = g.PushStash(ctx, "all", true)
	requireNoError(t, e)
	if _, e = os.Stat(filepath.Join(g.Dir, "untracked")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal(e)
	}
	_, e = g.ApplyStash(ctx, 0, false)
	requireNoError(t, e)
	stashes, e = g.Stashes(ctx)
	requireNoError(t, e)
	if len(stashes) != 1 {
		t.Fatal(stashes)
	}
	_, e = g.DropStash(ctx, 0)
	requireNoError(t, e)
	if _, e = g.DropStash(ctx, 0); e == nil {
		t.Fatal("repeated drop accepted")
	}
	if _, e = g.ApplyStash(ctx, -1, false); e == nil {
		t.Fatal("negative stash index accepted")
	}
}

func conflictRepo(t *testing.T) (Git, string) {
	g := testGit(t)
	testCommit(t, g, "a", "base\n", "base")
	requireNoError(t, g.CreateBranch(context.Background(), "feature"))
	feature := testCommit(t, g, "a", "feature\n", "feature")
	requireNoError(t, g.Checkout(context.Background(), "main"))
	testCommit(t, g, "a", "main\n", "main")
	return g, feature
}
func TestGitMergeConflictRecovery(t *testing.T) {
	ctx := context.Background()
	g, _ := conflictRepo(t)
	if _, e := g.Merge(ctx, "feature", MergeRegular); e == nil {
		t.Fatal("merge should conflict")
	}
	op, e := g.Operation(ctx)
	requireNoError(t, e)
	if op != OperationMerge {
		t.Fatal(op)
	}
	files, e := g.Conflicts(ctx)
	requireNoError(t, e)
	if !reflect.DeepEqual(files, []string{"a"}) {
		t.Fatal(files)
	}
	if _, e = g.Continue(ctx, OperationRebase); e == nil {
		t.Fatal("stale operation accepted")
	}
	if _, e = g.Skip(ctx, op); e == nil {
		t.Fatal("merge skip accepted")
	}
	_, e = g.Abort(ctx, op)
	requireNoError(t, e)
	if _, e = g.Abort(ctx, op); e == nil {
		t.Fatal("repeated abort succeeded")
	}
	_, e = g.Merge(ctx, "feature", MergeNoFF)
	if e == nil {
		t.Fatal("expected conflict")
	}
	writeGitFile(t, g, "a", "resolved\n")
	requireNoError(t, g.Stage(ctx, "a"))
	_, e = g.Continue(ctx, OperationMerge)
	requireNoError(t, e)
	op, e = g.Operation(ctx)
	requireNoError(t, e)
	if op != OperationNone {
		t.Fatal(op)
	}
	d, e := g.CommitDetail(ctx, "HEAD")
	requireNoError(t, e)
	if len(d.Parents) != 2 {
		t.Fatal(d)
	}
}
func TestGitRebaseCherryPickAndBisectRecovery(t *testing.T) {
	ctx := context.Background()
	t.Run("rebase", func(t *testing.T) {
		g, _ := conflictRepo(t)
		if _, e := g.Rebase(ctx, "feature", false); e == nil {
			t.Fatal("expected conflict")
		}
		op, e := g.Operation(ctx)
		requireNoError(t, e)
		if op != OperationRebase {
			t.Fatal(op)
		}
		_, e = g.Skip(ctx, op)
		requireNoError(t, e)
		op, e = g.Operation(ctx)
		requireNoError(t, e)
		if op != OperationNone {
			t.Fatal(op)
		}
	})
	t.Run("cherry-pick", func(t *testing.T) {
		g, feature := conflictRepo(t)
		if _, e := g.Run(ctx, "cherry-pick", feature); e == nil {
			t.Fatal("expected conflict")
		}
		op, e := g.Operation(ctx)
		requireNoError(t, e)
		if op != OperationCherryPick {
			t.Fatal(op)
		}
		_, e = g.Skip(ctx, op)
		requireNoError(t, e)
	})
	t.Run("bisect", func(t *testing.T) {
		g := testGit(t)
		testCommit(t, g, "a", "one", "one")
		testCommit(t, g, "a", "two", "two")
		testRun(t, g, "bisect", "start")
		op, e := g.Operation(ctx)
		requireNoError(t, e)
		if op != OperationBisect {
			t.Fatal(op)
		}
		if _, e = g.Continue(ctx, op); e == nil {
			t.Fatal("bisect continued without good/bad")
		}
		_, e = g.Abort(ctx, op)
		requireNoError(t, e)
	})
}

func TestGitLocalRemoteSynchronization(t *testing.T) {
	ctx := context.Background()
	g := testGit(t)
	testCommit(t, g, "a", "one\n", "initial")
	bare := filepath.Join(t.TempDir(), "remote.git")
	testRun(t, g, "init", "--bare", bare)
	testRun(t, g, "remote", "add", "origin", bare)
	status, e := g.AheadBehind(ctx)
	requireNoError(t, e)
	if status.HasUpstream {
		t.Fatal(status)
	}
	_, e = g.Push(ctx, false)
	requireNoError(t, e)
	status, e = g.AheadBehind(ctx)
	requireNoError(t, e)
	if !status.HasUpstream || status.Ahead != 0 || status.Behind != 0 {
		t.Fatal(status)
	}
	cloneDir := filepath.Join(t.TempDir(), "clone")
	testRun(t, g, "clone", "-b", "main", bare, cloneDir)
	clone := Git{Dir: cloneDir}
	testRun(t, clone, "config", "user.name", "Orbit Clone")
	testRun(t, clone, "config", "user.email", "clone@example.invalid")
	testRun(t, clone, "config", "commit.gpgsign", "false")
	testCommit(t, clone, "b", "remote\n", "remote change")
	_, e = clone.Push(ctx, false)
	requireNoError(t, e)
	requireNoError(t, g.Fetch(ctx))
	status, e = g.AheadBehind(ctx)
	requireNoError(t, e)
	if status.Behind != 1 {
		t.Fatal(status)
	}
	_, e = g.Sync(ctx)
	requireNoError(t, e)
	testCommit(t, g, "c", "local\n", "local change")
	_, e = g.Sync(ctx)
	requireNoError(t, e)
	requireNoError(t, clone.Fetch(ctx))
	_, e = clone.Pull(ctx)
	requireNoError(t, e)
	if testRun(t, g, "rev-parse", "HEAD") != testRun(t, clone, "rev-parse", "HEAD") {
		t.Fatal("sync mismatch")
	}
	requireNoError(t, clone.CreateBranch(ctx, "other"))
	testCommit(t, clone, "d", "branch", "branch")
	_, e = clone.Push(ctx, false)
	requireNoError(t, e)
	requireNoError(t, g.Fetch(ctx))
	requireNoError(t, g.CheckoutRemote(ctx, "origin/other"))
	if branch, e := g.CurrentBranch(ctx); e != nil || branch != "other" {
		t.Fatal(branch, e)
	}
}

func TestGitFailureAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	g := testGit(t)
	if _, e := g.Status(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	bad := Git{Dir: t.TempDir()}
	if _, e := bad.Status(context.Background()); e == nil {
		t.Fatal("nonrepo succeeded")
	}
	missing := Git{Dir: g.Dir, Binary: filepath.Join(g.Dir, "missing-git")}
	if _, e := missing.Run(context.Background(), "status"); e == nil {
		t.Fatal("missing binary succeeded")
	}
	_, e := commandOutput(context.Background(), g.Dir, os.Args[0], []string{"ORBIT_COMMAND_HELPER=1"}, []string{"-test.run=TestGitCommandHelper", "--", "fail"})
	var ce *CommandError
	if !errors.As(e, &ce) || ce.ExitCode != 42 || !strings.Contains(ce.Stderr, "failure detail") || ce.Stdout != "partial output" {
		t.Fatalf("error details %+v", e)
	}
	deadline, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, e = commandOutput(deadline, g.Dir, os.Args[0], []string{"ORBIT_COMMAND_HELPER=1"}, []string{"-test.run=TestGitCommandHelper", "--", "sleep"})
	if !errors.Is(e, context.DeadlineExceeded) || time.Since(start) > 3*time.Second {
		t.Fatalf("cancel %v after %v", e, time.Since(start))
	}
}
func TestGitCommandHelper(t *testing.T) {
	if os.Getenv("ORBIT_COMMAND_HELPER") != "1" {
		return
	}
	if os.Args[len(os.Args)-1] == "sleep" {
		time.Sleep(30 * time.Second)
		os.Exit(0)
	}
	_, _ = os.Stdout.WriteString("partial output")
	_, _ = os.Stderr.WriteString("failure detail")
	os.Exit(42)
}
func TestGitRemoteURLsAndFailureClassification(t *testing.T) {
	for _, raw := range []string{"git@github.com:org/repo.git", "https://github.com/org/repo.git", "ssh://git@github.com/org/repo.git"} {
		web, e := ParseRemoteWeb(raw)
		requireNoError(t, e)
		if web.CommitURL("abc") != "https://github.com/org/repo/commit/abc" {
			t.Fatal(web)
		}
	}
	if _, e := ParseRemoteWeb("/local/repo"); e == nil {
		t.Fatal("local treated as website")
	}
	cases := map[string]FailureKind{"GH013 push protection; failed to push some refs": FailureSecretScan, "GH001 pre-receive hook declined": FailureLargeFile, "GH006 protected branch [rejected]": FailureProtected, "git-lfs filter failed": FailureLFS, "not possible to fast-forward": FailureDiverged, "rebase conflict could not apply": FailureRebaseConflict, "authentication failed": FailureAuth, "not a git repository": FailureNotRepo}
	for raw, want := range cases {
		got := ClassifyGitFailure(raw)
		if got.Kind != want {
			t.Fatalf("%s: %s != %s", raw, got.Kind, want)
		}
	}
	if len(ClassifyGitFailure("GH013").RecoveryActions()) != 0 {
		t.Fatal("secret failure offers unsafe retry")
	}
	if len(ClassifyGitFailure("[rejected]").RecoveryActions()) != 3 {
		t.Fatal("missing recovery actions")
	}
}
