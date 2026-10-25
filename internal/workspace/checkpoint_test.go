package workspace

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cpWrite(t *testing.T, root, path, text string) {
	t.Helper()
	if e := os.WriteFile(filepath.Join(root, path), []byte(text), 0600); e != nil {
		t.Fatal(e)
	}
}
func cpRun(t *testing.T, g Git, args ...string) string {
	t.Helper()
	b, e := g.Run(context.Background(), args...)
	if e != nil {
		t.Fatal(e)
	}
	return strings.TrimSpace(string(b))
}
func TestCheckpointDoesNotChangeIndexHeadOrWorktree(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	cpWrite(t, root, "seed.txt", "staged\n")
	cpRun(t, g, "add", "seed.txt")
	cpWrite(t, root, "seed.txt", "unstaged\n")
	cpWrite(t, root, "new\tfile\n.txt", "untracked\n")
	indexPath := filepath.Join(root, ".git", "index")
	index, _ := os.ReadFile(indexPath)
	head := cpRun(t, g, "rev-parse", "HEAD")
	status := cpRun(t, g, "status", "--porcelain=v1", "-z")
	before, _ := os.Stat(indexPath)
	for i := 0; i < 2; i++ {
		snap, e := g.CaptureWorktreeCommit(ctx)
		if e != nil {
			t.Fatal(e)
		}
		if text := cpRun(t, g, "show", snap+":seed.txt"); text != "unstaged" {
			t.Fatal(text)
		}
		if _, e = g.IndexTree(ctx); e != nil {
			t.Fatal(e)
		}
	}
	after, _ := os.ReadFile(indexPath)
	info, _ := os.Stat(indexPath)
	if !bytes.Equal(index, after) || !before.ModTime().Equal(info.ModTime()) {
		t.Fatal("index rewritten")
	}
	if cpRun(t, g, "rev-parse", "HEAD") != head || cpRun(t, g, "status", "--porcelain=v1", "-z") != status {
		t.Fatal("user state changed")
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".git", "orbit-checkpoint-index-*"))
	if len(matches) > 0 {
		t.Fatal("temporary indexes leaked", matches)
	}
}
func TestCheckpointTurnFreezesDiffAndRestoresCounter(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	cpWrite(t, root, "seed.txt", "before turn\n")
	if e := g.CaptureTurnStart(ctx, "session/a:b", 1); e != nil {
		t.Fatal(e)
	}
	cpWrite(t, root, "seed.txt", "during turn\n")
	cpWrite(t, root, "new.txt", "agent\n")
	if e := g.CaptureTurn(ctx, "session/a:b", 1); e != nil {
		t.Fatal(e)
	}
	cpWrite(t, root, "seed.txt", "after turn\n")
	d, e := g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "session/a:b", 1})
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(d.Patch, "-before turn") || !strings.Contains(d.Patch, "+during turn") || strings.Contains(d.Patch, "after turn") {
		t.Fatal(d.Patch)
	}
	n, e := (Git{Dir: root}).LatestTurn(ctx, "session/a:b")
	if e != nil || n != 1 {
		t.Fatal(n, e)
	}
	if SessionKey("a /é") != "a___" {
		t.Fatal(SessionKey("a /é"))
	}
	if _, e = g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "session/a:b", 0}); e == nil {
		t.Fatal("baseline as turn")
	}
	if _, e = g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "missing", 1}); e == nil {
		t.Fatal("missing checkpoint")
	}
}
func TestCheckpointBranchSwitchExcludesExistingTargetChanges(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	cpRun(t, g, "switch", "-c", "target")
	cpWrite(t, root, "target.txt", "existing target\n")
	cpRun(t, g, "add", ".")
	cpRun(t, g, "commit", "-m", "existing target")
	cpRun(t, g, "switch", "main")
	cpWrite(t, root, "local.txt", "pre-turn dirty\n")
	if e := g.CaptureTurnStart(ctx, "switch", 1); e != nil {
		t.Fatal(e)
	}
	cpRun(t, g, "switch", "target")
	cpWrite(t, root, "agent.txt", "agent edit\n")
	if e := g.CaptureTurn(ctx, "switch", 1); e != nil {
		t.Fatal(e)
	}
	d, e := g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "switch", 1})
	if e != nil {
		t.Fatal(e)
	}
	s := ParseReview(d)
	if len(s.Files) != 1 || s.Files[0].Path != "agent.txt" {
		t.Fatal(s.Files, d.Patch)
	}
}
func TestCheckpointNewBranchIncludesCommittedAgentWork(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	if e := g.CaptureTurnStart(ctx, "newbranch", 1); e != nil {
		t.Fatal(e)
	}
	cpRun(t, g, "switch", "-c", "new-task")
	cpWrite(t, root, "agent.txt", "created\n")
	cpRun(t, g, "add", ".")
	cpRun(t, g, "commit", "-m", "agent change")
	if e := g.CaptureTurn(ctx, "newbranch", 1); e != nil {
		t.Fatal(e)
	}
	d, e := g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "newbranch", 1})
	if e != nil || !strings.Contains(d.Patch, "+created") {
		t.Fatal(d, e)
	}
}
func TestCheckpointUnbornAndCancellation(t *testing.T) {
	root := t.TempDir()
	g := Git{Dir: root}
	cpRun(t, g, "init", "-b", "main")
	cpWrite(t, root, "first.txt", "first\n")
	ctx := context.Background()
	if e := g.CaptureTurnStart(ctx, "new", 1); e != nil {
		t.Fatal(e)
	}
	cpWrite(t, root, "first.txt", "second\n")
	if e := g.CaptureTurn(ctx, "new", 1); e != nil {
		t.Fatal(e)
	}
	d, e := g.CollectReview(ctx, ReviewRequest{ReviewStaged, "", 0})
	if e != nil || d.Patch != "" {
		t.Fatal(d, e)
	}
	c, cancel := context.WithCancel(ctx)
	cancel()
	if e = g.CaptureTurnStart(c, "cancel", 1); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	if _, e = g.ResolveCommit(ctx, TurnStartRef("cancel", 1)); e == nil {
		t.Fatal("cancel wrote ref")
	}
}

func TestCheckpointPlainFolderAndLegacyBaselineAreNonErrors(t *testing.T) {
	ctx := context.Background()
	plain := Git{Dir: t.TempDir()}
	if e := plain.CaptureTurnStart(ctx, "plain", 1); e != nil {
		t.Fatal(e)
	}
	if e := plain.CaptureTurn(ctx, "plain", 1); e != nil {
		t.Fatal(e)
	}
	root := wtTestRepo(t)
	g := Git{Dir: root}
	if e := g.CaptureTurn(ctx, "legacy", 0); e != nil {
		t.Fatal(e)
	}
	if _, e := g.ResolveCommit(ctx, CheckpointRef("legacy", 0)); e != nil {
		t.Fatal(e)
	}
	cpWrite(t, root, "agent", "legacy turn\n")
	if e := g.CaptureTurn(ctx, "legacy", 1); e != nil {
		t.Fatal(e)
	}
	d, e := g.CollectReview(ctx, ReviewRequest{ReviewLastTurn, "legacy", 1})
	if e != nil || !strings.Contains(d.Patch, "legacy turn") {
		t.Fatal(d, e)
	}
}
