package workspace

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const turnStartMetadataPrefix = "Orbit-Turn-Start: "

func SessionKey(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, s)
}
func CheckpointRef(session string, turn int) string {
	return fmt.Sprintf("refs/orbit/session-%s-turn-%d", SessionKey(session), turn)
}
func TurnStartRef(session string, turn int) string {
	return fmt.Sprintf("refs/orbit/session-%s-turn-start-%d", SessionKey(session), turn)
}
func TurnDiffBaseRef(session string, turn int) string {
	return fmt.Sprintf("refs/orbit/session-%s-turn-diff-%d", SessionKey(session), turn)
}

type turnStartMetadata struct {
	Head   string            `json:"head"`
	Branch string            `json:"branch"`
	Refs   map[string]string `json:"refs"`
}

// runInput is private plumbing for commit/ref transactions. It inherits Git's
// noninteractive policy, but private indexes and identities stay child-local.
func (g Git) runInput(ctx context.Context, input []byte, env []string, args ...string) ([]byte, error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	bin := g.Binary
	if bin == "" {
		bin = "git"
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	hideGitConsole(cmd)
	cmd.Dir = g.Dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdin = bytes.NewReader(input)
	cmd.WaitDelay = 2 * time.Second
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if e := cmd.Run(); e != nil {
		code := -1
		var ee *exec.ExitError
		if errors.As(e, &ee) {
			code = ee.ExitCode()
		}
		if ctx.Err() != nil {
			e = ctx.Err()
		}
		return nil, &CommandError{Binary: bin, Args: args, Stdout: out.String(), Stderr: stderr.String(), ExitCode: code, Cause: e}
	}
	return out.Bytes(), nil
}
func (g Git) ResolveCommit(ctx context.Context, ref string) (string, error) {
	if e := validRevision(ref); e != nil {
		return "", e
	}
	return g.text(ctx, "rev-parse", "--verify", ref+"^{commit}")
}
func (g Git) emptyTree(ctx context.Context) (string, error) {
	out, e := g.runInput(ctx, nil, nil, "mktree")
	return strings.TrimSpace(string(out)), e
}
func (g Git) headOrEmpty(ctx context.Context) (string, error) {
	head, e := g.ResolveCommit(ctx, "HEAD")
	if e == nil {
		return head, nil
	}
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return g.emptyTree(ctx)
}
func (g Git) privateIndex(ctx context.Context, copyIndex bool) (string, func(), error) {
	dir, e := g.text(ctx, "rev-parse", "--absolute-git-dir")
	if e != nil {
		return "", nil, e
	}
	f, e := os.CreateTemp(dir, "orbit-checkpoint-index-*")
	if e != nil {
		return "", nil, e
	}
	p := f.Name()
	f.Close()
	os.Remove(p)
	cleanup := func() { os.Remove(p); os.Remove(p + ".lock") }
	if copyIndex {
		raw, e := os.ReadFile(filepath.Join(dir, "index"))
		if e == nil {
			if e = os.WriteFile(p, raw, 0600); e != nil {
				cleanup()
				return "", nil, e
			}
		} else if !errors.Is(e, os.ErrNotExist) {
			cleanup()
			return "", nil, e
		}
	}
	return p, cleanup, nil
}

var snapshotIdentity = []string{"GIT_AUTHOR_NAME=Orbit", "GIT_AUTHOR_EMAIL=orbit@localhost", "GIT_COMMITTER_NAME=Orbit", "GIT_COMMITTER_EMAIL=orbit@localhost"}

func (g Git) commitTree(ctx context.Context, tree, message string, parents []string) (string, error) {
	args := []string{"commit-tree", tree, "-m", message}
	for _, p := range parents {
		args = append(args, "-p", p)
	}
	out, e := g.runInput(ctx, nil, snapshotIdentity, args...)
	return strings.TrimSpace(string(out)), e
}
func (g Git) captureWorktree(ctx context.Context, head, message string, parents []string) (string, error) {
	index, cleanup, e := g.privateIndex(ctx, false)
	if e != nil {
		return "", e
	}
	defer cleanup()
	env := []string{"GIT_INDEX_FILE=" + index}
	if head != "" {
		if _, e = g.runInput(ctx, nil, env, "read-tree", head); e != nil {
			return "", e
		}
	}
	if _, e = g.runInput(ctx, nil, env, "add", "-A", "--", "."); e != nil {
		return "", e
	}
	out, e := g.runInput(ctx, nil, env, "write-tree")
	if e != nil {
		return "", e
	}
	tree := strings.TrimSpace(string(out))
	if tree == "" {
		return "", errors.New("git returned empty tree id")
	}
	return g.commitTree(ctx, tree, message, parents)
}

// CaptureWorktreeCommit snapshots tracked and untracked files without changing
// the user's index, HEAD, branches, or worktree. It writes dangling objects.
func (g Git) CaptureWorktreeCommit(ctx context.Context) (string, error) {
	ok, e := g.IsRepo(ctx)
	if e != nil {
		return "", fmt.Errorf("worktree snapshot requires repository: %w", e)
	}
	if !ok {
		return "", errors.New("worktree snapshot requires repository")
	}
	head, _ := g.ResolveCommit(ctx, "HEAD")
	return g.captureWorktree(ctx, head, "Orbit worktree snapshot", nil)
}
func (g Git) IndexTree(ctx context.Context) (string, error) {
	index, cleanup, e := g.privateIndex(ctx, true)
	if e != nil {
		return "", e
	}
	defer cleanup()
	out, e := g.runInput(ctx, nil, []string{"GIT_INDEX_FILE=" + index}, "write-tree")
	return strings.TrimSpace(string(out)), e
}
func (g Git) repositoryRefs(ctx context.Context) (map[string]string, error) {
	out, e := g.text(ctx, "for-each-ref", "--format=%(refname)%09%(objecttype)%09%(objectname)%09%(*objecttype)%09%(*objectname)", "refs/heads", "refs/remotes", "refs/tags")
	if e != nil {
		return nil, e
	}
	refs := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(line, "\t")
		if len(f) >= 3 && f[1] == "commit" {
			refs[f[0]] = f[2]
		} else if len(f) >= 5 && f[3] == "commit" {
			refs[f[0]] = f[4]
		}
	}
	return refs, nil
}
func (g Git) CaptureTurnStart(ctx context.Context, session string, turn int) error {
	if turn < 1 || session == "" {
		return errors.New("turn and session required")
	}
	ok, e := g.IsRepo(ctx)
	if e != nil || !ok {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return nil // Plain folders support Pi sessions without Git checkpoints.
	}
	head, _ := g.ResolveCommit(ctx, "HEAD")
	branch, _ := g.text(ctx, "symbolic-ref", "--quiet", "HEAD")
	refs, e := g.repositoryRefs(ctx)
	if e != nil {
		return e
	}
	meta := turnStartMetadata{head, branch, refs}
	raw, e := json.Marshal(meta)
	if e != nil {
		return e
	}
	seen := map[string]bool{}
	var parents []string
	appendParent := func(p string) {
		if p != "" && !seen[p] {
			seen[p] = true
			parents = append(parents, p)
		}
	}
	appendParent(head)
	var keys []string
	for k := range refs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		appendParent(refs[k])
	}
	commit, e := g.captureWorktree(ctx, head, "Orbit turn start snapshot\n\n"+turnStartMetadataPrefix+string(raw), parents)
	if e != nil {
		return e
	}
	commands := "update " + TurnStartRef(session, turn) + " " + commit + "\n"
	if _, e = g.ResolveCommit(ctx, CheckpointRef(session, turn-1)); e != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		commands += "update " + CheckpointRef(session, turn-1) + " " + commit + "\n"
	}
	_, e = g.runInput(ctx, []byte(commands), nil, "update-ref", "--stdin")
	return e
}
func (g Git) CaptureTurn(ctx context.Context, session string, turn int) error {
	if turn < 0 || session == "" {
		return errors.New("turn and session required")
	}
	ok, e := g.IsRepo(ctx)
	if e != nil || !ok {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return nil // Plain folders support Pi sessions without Git checkpoints.
	}
	head, _ := g.ResolveCommit(ctx, "HEAD")
	branch, _ := g.text(ctx, "symbolic-ref", "--quiet", "HEAD")
	end, e := g.captureWorktree(ctx, head, "Orbit worktree snapshot", nil)
	if e != nil {
		return e
	}
	if _, e = g.Run(ctx, "update-ref", CheckpointRef(session, turn), end); e != nil {
		return e
	}
	start := TurnStartRef(session, turn)
	if turn == 0 {
		return nil
	}
	if _, e = g.ResolveCommit(ctx, start); e != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return nil
	}
	base, e := g.turnDiffBase(ctx, start, head, branch)
	if e != nil {
		return e
	}
	_, e = g.Run(ctx, "update-ref", TurnDiffBaseRef(session, turn), base)
	return e
}
func (g Git) turnDiffBase(ctx context.Context, start, endHead, endBranch string) (string, error) {
	message, e := g.text(ctx, "show", "-s", "--format=%B", start)
	if e != nil {
		return "", e
	}
	var meta turnStartMetadata
	found := false
	for _, line := range strings.Split(message, "\n") {
		if strings.HasPrefix(line, turnStartMetadataPrefix) {
			e = json.Unmarshal([]byte(strings.TrimPrefix(line, turnStartMetadataPrefix)), &meta)
			found = true
			break
		}
	}
	if !found {
		return "", errors.New("turn start metadata unavailable")
	}
	if e != nil {
		return "", e
	}
	same := meta.Branch == endBranch
	if meta.Branch == "" && endBranch == "" {
		same = meta.Head == endHead
	}
	if same {
		return g.ResolveCommit(ctx, start)
	}
	target := meta.Refs[endBranch]
	if target == "" {
		if endHead == "" {
			tree, e := g.emptyTree(ctx)
			if e != nil {
				return "", e
			}
			target, e = g.commitTree(ctx, tree, "Orbit empty turn diff base", nil)
			if e != nil {
				return "", e
			}
		} else {
			newCommits, e := g.text(ctx, "rev-list", "--first-parent", "--reverse", endHead, "--not", start)
			if e != nil {
				return "", e
			}
			first := strings.Split(newCommits, "\n")[0]
			target = endHead
			if first != "" {
				target, e = g.ResolveCommit(ctx, first+"^1")
				if e != nil {
					tree, e := g.emptyTree(ctx)
					if e != nil {
						return "", e
					}
					target, e = g.commitTree(ctx, tree, "Orbit empty turn diff base", nil)
					if e != nil {
						return "", e
					}
				}
			}
		}
	}
	if meta.Head == "" {
		return target, nil
	}
	if meta.Head == target {
		return g.ResolveCommit(ctx, start)
	}
	dirty, e := g.text(ctx, "diff", "--name-only", meta.Head, start)
	if e != nil {
		return "", e
	}
	if dirty == "" {
		return target, nil
	}
	out, e := g.text(ctx, "merge-tree", "--write-tree", "--merge-base", meta.Head, target, start)
	if e != nil {
		return "", e
	}
	tree := strings.Split(out, "\n")[0]
	if tree == "" {
		return "", errors.New("merge-tree returned no tree")
	}
	return g.commitTree(ctx, tree, "Orbit turn diff base", nil)
}
func (g Git) LatestTurn(ctx context.Context, session string) (int, error) {
	prefix := "refs/orbit/session-" + SessionKey(session) + "-turn-"
	out, e := g.text(ctx, "for-each-ref", "--format=%(refname)", prefix+"*")
	if e != nil {
		return 0, e
	}
	max := 0
	for _, line := range strings.Split(out, "\n") {
		suffix := strings.TrimPrefix(line, prefix)
		n, e := strconv.Atoi(suffix)
		if e == nil && n > max {
			max = n
		}
	}
	return max, nil
}
