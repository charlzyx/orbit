package workspace

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type InProgress string

const (
	OperationNone       InProgress = ""
	OperationMerge      InProgress = "merge"
	OperationRebase     InProgress = "rebase"
	OperationCherryPick InProgress = "cherry-pick"
	OperationRevert     InProgress = "revert"
	OperationBisect     InProgress = "bisect"
)

func (o InProgress) CanSkip() bool { return o == OperationRebase || o == OperationCherryPick }
func (g Git) Operation(ctx context.Context) (InProgress, error) {
	dir, e := g.text(ctx, "rev-parse", "--git-dir")
	if e != nil {
		return OperationNone, e
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(g.Dir, dir)
	}
	for _, m := range []struct {
		name      string
		op        InProgress
		directory bool
	}{{"rebase-merge", OperationRebase, true}, {"rebase-apply", OperationRebase, true}, {"MERGE_HEAD", OperationMerge, false}, {"CHERRY_PICK_HEAD", OperationCherryPick, false}, {"REVERT_HEAD", OperationRevert, false}, {"BISECT_LOG", OperationBisect, false}} {
		s, e := os.Stat(filepath.Join(dir, m.name))
		if e == nil && s.IsDir() == m.directory {
			return m.op, nil
		}
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return OperationNone, e
		}
	}
	return OperationNone, nil
}
func (g Git) Conflicts(ctx context.Context) ([]string, error) {
	b, e := g.Run(ctx, "diff", "--name-only", "--diff-filter=U", "-z")
	if e != nil {
		return nil, e
	}
	var files []string
	for _, p := range strings.Split(string(b), "\x00") {
		if p != "" {
			files = append(files, p)
		}
	}
	return files, nil
}
func (g Git) checkOperation(ctx context.Context, expected InProgress) error {
	actual, e := g.Operation(ctx)
	if e != nil {
		return e
	}
	if actual == OperationNone {
		return errors.New("no Git operation is in progress")
	}
	if actual != expected {
		return fmt.Errorf("operation changed: expected %s, found %s", expected, actual)
	}
	return nil
}
func (g Git) Continue(ctx context.Context, op InProgress) (string, error) {
	if e := g.checkOperation(ctx, op); e != nil {
		return "", e
	}
	if op == OperationBisect {
		return "", errors.New("bisect requires an explicit good/bad decision")
	}
	return g.text(ctx, string(op), "--continue")
}
func (g Git) Abort(ctx context.Context, op InProgress) (string, error) {
	if e := g.checkOperation(ctx, op); e != nil {
		return "", e
	}
	if op == OperationBisect {
		return g.text(ctx, "bisect", "reset")
	}
	return g.text(ctx, string(op), "--abort")
}
func (g Git) Skip(ctx context.Context, op InProgress) (string, error) {
	if !op.CanSkip() {
		return "", errors.New("this operation does not support skip")
	}
	if e := g.checkOperation(ctx, op); e != nil {
		return "", e
	}
	return g.text(ctx, string(op), "--skip")
}

type RefEntry struct {
	Name string
	Kind RefKind
}

func (g Git) Refs(ctx context.Context) ([]RefEntry, error) {
	s, e := g.text(ctx, "for-each-ref", "--sort=-committerdate", "--format=%(refname)", "refs/heads/", "refs/remotes/", "refs/tags/")
	if e != nil {
		return nil, e
	}
	var refs []RefEntry
	for _, r := range nonemptyLines(s) {
		for _, kind := range []struct {
			prefix string
			kind   RefKind
		}{{"refs/heads/", RefBranch}, {"refs/remotes/", RefRemote}, {"refs/tags/", RefTag}} {
			if strings.HasPrefix(r, kind.prefix) {
				name := strings.TrimPrefix(r, kind.prefix)
				if kind.kind != RefRemote || !strings.HasSuffix(name, "/HEAD") {
					refs = append(refs, RefEntry{name, kind.kind})
				}
				break
			}
		}
	}
	return refs, nil
}

type MergeMode string

const (
	MergeRegular MergeMode = "merge"
	MergeNoFF    MergeMode = "no-ff"
	MergeSquash  MergeMode = "squash"
)

func (g Git) Merge(ctx context.Context, ref string, mode MergeMode) (string, error) {
	if e := validRevision(ref); e != nil {
		return "", e
	}
	args := []string{"merge", "--no-edit"}
	switch mode {
	case "", MergeRegular:
	case MergeNoFF:
		args = append(args, "--no-ff")
	case MergeSquash:
		args = append(args, "--squash")
	default:
		return "", fmt.Errorf("invalid merge mode %q", mode)
	}
	return g.text(ctx, append(args, ref)...)
}
func (g Git) Rebase(ctx context.Context, ref string, autostash bool) (string, error) {
	if e := validRevision(ref); e != nil {
		return "", e
	}
	args := []string{"rebase"}
	if autostash {
		args = append(args, "--autostash")
	}
	return g.text(ctx, append(args, ref)...)
}
func (g Git) RenameBranch(ctx context.Context, from, to string) error {
	if e := validRevision(from); e != nil {
		return e
	}
	if e := validRevision(to); e != nil {
		return e
	}
	if e := g.do(ctx, "check-ref-format", "--branch", to); e != nil {
		return e
	}
	return g.do(ctx, "branch", "-m", from, to)
}
func (g Git) DeleteBranch(ctx context.Context, branch string, force bool) error {
	if e := validRevision(branch); e != nil {
		return e
	}
	flag := "-d"
	if force {
		flag = "-D"
	}
	return g.do(ctx, "branch", flag, branch)
}

type StashEntry struct {
	Index   int
	Message string
}

func (g Git) Stashes(ctx context.Context) ([]StashEntry, error) {
	s, e := g.text(ctx, "stash", "list", "--format=%gd%x1f%gs")
	if e != nil {
		return nil, e
	}
	var result []StashEntry
	for _, line := range nonemptyLines(s) {
		p := strings.SplitN(line, "\x1f", 2)
		if len(p) != 2 {
			return nil, errors.New("malformed stash entry")
		}
		n, e := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(p[0], "stash@{"), "}"))
		if e != nil {
			return nil, e
		}
		result = append(result, StashEntry{n, p[1]})
	}
	return result, nil
}
func (g Git) PushStash(ctx context.Context, message string, includeUntracked bool) (string, error) {
	args := []string{"stash", "push"}
	if includeUntracked {
		args = append(args, "--include-untracked")
	}
	if strings.TrimSpace(message) != "" {
		args = append(args, "-m", message)
	}
	return g.text(ctx, args...)
}
func stashRef(index int) (string, error) {
	if index < 0 {
		return "", errors.New("stash index cannot be negative")
	}
	return fmt.Sprintf("stash@{%d}", index), nil
}
func (g Git) ApplyStash(ctx context.Context, index int, pop bool) (string, error) {
	ref, e := stashRef(index)
	if e != nil {
		return "", e
	}
	action := "apply"
	if pop {
		action = "pop"
	}
	return g.text(ctx, "stash", action, ref)
}
func (g Git) DropStash(ctx context.Context, index int) (string, error) {
	ref, e := stashRef(index)
	if e != nil {
		return "", e
	}
	return g.text(ctx, "stash", "drop", ref)
}

type FailureKind string

const (
	FailurePushRejected     FailureKind = "push_rejected"
	FailureDiverged         FailureKind = "diverged"
	FailureMergeConflict    FailureKind = "merge_conflicts"
	FailureRebaseConflict   FailureKind = "rebase_conflicts"
	FailureLocalOverwritten FailureKind = "local_overwritten"
	FailureDirtyWorktree    FailureKind = "dirty_worktree"
	FailureAuth             FailureKind = "auth"
	FailureSecretScan       FailureKind = "secret_scan"
	FailureLargeFile        FailureKind = "large_file"
	FailureLFS              FailureKind = "lfs"
	FailureProtected        FailureKind = "protected"
	FailurePreReceiveHook   FailureKind = "pre_receive_hook"
	FailureRemoteMissing    FailureKind = "remote_not_found"
	FailureNetwork          FailureKind = "network"
	FailureNoUpstream       FailureKind = "no_upstream"
	FailureNotRepo          FailureKind = "not_a_repo"
	FailureGeneric          FailureKind = "generic"
)

type RecoveryAction string

const (
	RecoveryPull           RecoveryAction = "pull"
	RecoveryMerge          RecoveryAction = "merge"
	RecoveryRebase         RecoveryAction = "rebase"
	RecoveryForceWithLease RecoveryAction = "force_with_lease"
	RecoveryAbort          RecoveryAction = "abort"
	RecoveryContinue       RecoveryAction = "continue"
	RecoverySkip           RecoveryAction = "skip"
	RecoveryOpenConflicts  RecoveryAction = "open_conflicts"
	RecoveryAuthenticate   RecoveryAction = "reauthenticate"
	RecoveryRetry          RecoveryAction = "retry"
	RecoveryPublish        RecoveryAction = "publish"
)

type ActionError struct {
	Kind             FailureKind
	TitleKey, Detail string
}

func ClassifyGitFailure(raw string) ActionError {
	h := strings.ToLower(raw)
	has := func(s string) bool { return strings.Contains(h, s) }
	k := FailureGeneric
	switch {
	case has("push protection") || has("gh013") || has("secret scanning") || has("push declined due to"):
		k = FailureSecretScan
	case has("exceeds github's file size limit") || has("gh001") || (has("file is") && has("mb; this exceeds")):
		k = FailureLargeFile
	case has("git-lfs") || has("lfs filter") || has("this repository is over its data quota"):
		k = FailureLFS
	case has("protected branch") || has("gh006"):
		k = FailureProtected
	case has("pre-receive hook declined"):
		k = FailurePreReceiveHook
	case has("[rejected]") || has("non-fast-forward") || has("failed to push some refs") || has("fetch first"):
		k = FailurePushRejected
	case has("not possible to fast-forward") || has("divergent branches") || has("need to specify how to reconcile"):
		k = FailureDiverged
	case has("could not apply") || (has("rebase") && has("conflict")):
		k = FailureRebaseConflict
	case has("automatic merge failed") || has("merge conflict") || has("conflict"):
		k = FailureMergeConflict
	case has("would be overwritten") || has("your local changes"):
		k = FailureLocalOverwritten
	case has("cannot rebase") || has("cannot pull with rebase") || has("unstaged changes"):
		k = FailureDirtyWorktree
	case has("authentication failed") || has("could not read username") || has("could not read password") || has("permission denied") || has("403 forbidden") || has("401 unauthorized"):
		k = FailureAuth
	case has("repository not found") || has("does not appear to be a git repository") || has("couldn't find remote ref") || has("no such remote"):
		k = FailureRemoteMissing
	case has("could not resolve host") || has("unable to access") || has("network is unreachable") || has("timed out"):
		k = FailureNetwork
	case has("no upstream branch") || has("no branch to push"):
		k = FailureNoUpstream
	case has("not a git repository"):
		k = FailureNotRepo
	}
	var lines []string
	prevBlank := false
	for _, l := range strings.Split(raw, "\n") {
		l = strings.TrimRight(l, " \t\r")
		blank := l == ""
		if !(blank && prevBlank) {
			lines = append(lines, l)
		}
		prevBlank = blank
	}
	detail := strings.TrimSpace(strings.Join(lines, "\n"))
	if detail == "" {
		detail = "Git did not provide an error message"
	}
	return ActionError{k, "git_panel.fail_" + string(k), detail}
}
func (e ActionError) RecoveryActions() []RecoveryAction {
	switch e.Kind {
	case FailurePushRejected:
		return []RecoveryAction{RecoveryPull, RecoveryMerge, RecoveryForceWithLease}
	case FailureDiverged:
		return []RecoveryAction{RecoveryMerge, RecoveryRebase}
	case FailureMergeConflict:
		return []RecoveryAction{RecoveryOpenConflicts, RecoveryAbort, RecoveryContinue}
	case FailureRebaseConflict:
		return []RecoveryAction{RecoveryOpenConflicts, RecoveryAbort, RecoveryContinue, RecoverySkip}
	case FailureLocalOverwritten:
		return []RecoveryAction{RecoveryAbort}
	case FailureAuth:
		return []RecoveryAction{RecoveryAuthenticate, RecoveryRetry}
	case FailureNetwork:
		return []RecoveryAction{RecoveryRetry}
	case FailureNoUpstream:
		return []RecoveryAction{RecoveryPublish}
	}
	return nil
}
