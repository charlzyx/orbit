package workspace

// Git plumbing for the native workbench. All subprocesses are cancellable and
// return real exit failures. Confirmation policy belongs to the calling UI;
// destructive alternatives are never selected automatically here.
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Git struct{ Dir, Binary string }

type CommandError struct {
	Binary         string
	Args           []string
	Stdout, Stderr string
	ExitCode       int
	Cause          error
}

func (e *CommandError) Error() string {
	detail := strings.TrimSpace(e.Stderr)
	if detail == "" {
		detail = strings.TrimSpace(e.Stdout)
	}
	if detail == "" {
		detail = e.Cause.Error()
	}
	return fmt.Sprintf("%s: %s", filepath.Base(e.Binary), detail)
}
func (e *CommandError) Unwrap() error { return e.Cause }

func commandOutput(ctx context.Context, dir, binary string, env, args []string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	configureSetupCancellation(cmd)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		code := -1
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return nil, &CommandError{Binary: binary, Args: append([]string(nil), args...), Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: code, Cause: err}
	}
	return stdout.Bytes(), nil
}
func (g Git) Run(ctx context.Context, args ...string) ([]byte, error) {
	bin := g.Binary
	if bin == "" {
		bin = "git"
	}
	return commandOutput(ctx, g.Dir, bin, []string{"GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0", "GIT_EDITOR=true", "GIT_SEQUENCE_EDITOR=true", "LC_ALL=C"}, args)
}
func (g Git) text(ctx context.Context, args ...string) (string, error) {
	b, e := g.Run(ctx, args...)
	return strings.TrimSpace(string(b)), e
}
func (g Git) do(ctx context.Context, args ...string) error { _, e := g.Run(ctx, args...); return e }
func validRevision(ref string) error {
	if strings.TrimSpace(ref) == "" || strings.HasPrefix(ref, "-") || strings.ContainsAny(ref, "\x00\r\n") {
		return fmt.Errorf("invalid Git revision %q", ref)
	}
	return nil
}
func literalPaths(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, errors.New("at least one path is required")
	}
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		clean := filepath.Clean(p)
		if p == "" || strings.ContainsRune(p, 0) || filepath.IsAbs(p) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("path is outside workspace: %q", p)
		}
		out = append(out, ":(literal)"+filepath.ToSlash(p))
	}
	return out, nil
}
func (g Git) IsRepo(ctx context.Context) (bool, error) {
	s, e := g.text(ctx, "rev-parse", "--is-inside-work-tree")
	if e != nil {
		return false, e
	}
	return s == "true", nil
}
func (g Git) HasCommits(ctx context.Context) (bool, error) {
	_, e := g.Run(ctx, "rev-parse", "--verify", "HEAD")
	if e != nil {
		var ce *CommandError
		if errors.As(e, &ce) && ce.ExitCode == 128 {
			if _, err := g.IsRepo(ctx); err == nil {
				return false, nil
			}
		}
		return false, e
	}
	return true, nil
}
func (g Git) CurrentBranch(ctx context.Context) (string, error) {
	return g.text(ctx, "branch", "--show-current")
}
func (g Git) UserName(ctx context.Context) (string, error) { return g.text(ctx, "config", "user.name") }
func (g Git) Branches(ctx context.Context) ([]string, error) {
	s, e := g.text(ctx, "for-each-ref", "--sort=-committerdate", "--format=%(refname:short)", "refs/heads/")
	if e != nil {
		return nil, e
	}
	return nonemptyLines(s), nil
}
func nonemptyLines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}
func (g Git) RecentBranches(ctx context.Context) ([]string, error) {
	branches, e := g.Branches(ctx)
	if e != nil {
		return nil, e
	}
	known := map[string]bool{}
	for _, b := range branches {
		known[b] = true
	}
	raw, e := g.text(ctx, "reflog", "show", "HEAD", "--format=%gs", "-n", "200")
	if e != nil {
		ok, err := g.HasCommits(ctx)
		if err == nil && !ok {
			return nil, nil
		}
		return nil, e
	}
	var result []string
	seen := map[string]bool{}
	for _, line := range nonemptyLines(raw) {
		if strings.HasPrefix(line, "checkout: moving from ") {
			parts := strings.SplitN(strings.TrimPrefix(line, "checkout: moving from "), " to ", 2)
			if len(parts) == 2 {
				for _, b := range []string{parts[1], parts[0]} {
					if known[b] && !seen[b] {
						result = append(result, b)
						seen[b] = true
					}
				}
			}
		}
	}
	return result, nil
}
func (g Git) Checkout(ctx context.Context, ref string) error {
	if e := validRevision(ref); e != nil {
		return e
	}
	return g.do(ctx, "checkout", ref, "--")
}
func (g Git) CreateBranch(ctx context.Context, name string) error {
	if e := validRevision(name); e != nil {
		return e
	}
	if e := g.do(ctx, "check-ref-format", "--branch", name); e != nil {
		return e
	}
	return g.do(ctx, "checkout", "-b", name)
}
func (g Git) CheckoutRemote(ctx context.Context, ref string) error {
	if e := validRevision(ref); e != nil {
		return e
	}
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return errors.New("remote reference must include remote/branch")
	}
	bs, e := g.Branches(ctx)
	if e != nil {
		return e
	}
	for _, b := range bs {
		if b == parts[1] {
			return g.Checkout(ctx, b)
		}
	}
	return g.do(ctx, "checkout", "-b", parts[1], "--track", ref)
}

type StatusRow struct {
	Path, OrigPath                                                         string
	Index, Worktree                                                        byte
	StagedAdditions, StagedDeletions, UnstagedAdditions, UnstagedDeletions int
}

func (s StatusRow) Staged() bool   { return s.Index != ' ' && s.Index != '?' && s.Index != '!' }
func (s StatusRow) Unstaged() bool { return s.Worktree != ' ' && s.Worktree != '!' }
func (s StatusRow) Conflict() bool {
	return s.Index == 'U' || s.Worktree == 'U' || (s.Index == 'A' && s.Worktree == 'A') || (s.Index == 'D' && s.Worktree == 'D')
}
func (g Git) Status(ctx context.Context) ([]StatusRow, error) {
	raw, e := g.Run(ctx, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if e != nil {
		return nil, e
	}
	parts := bytes.Split(raw, []byte{0})
	var rows []StatusRow
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		if len(p) == 0 {
			continue
		}
		if len(p) < 4 || p[2] != ' ' {
			return nil, fmt.Errorf("invalid git status record %q", p)
		}
		r := StatusRow{Index: p[0], Worktree: p[1], Path: string(p[3:])}
		if r.Index == 'R' || r.Index == 'C' || r.Worktree == 'R' || r.Worktree == 'C' {
			i++
			if i >= len(parts) || len(parts[i]) == 0 {
				return nil, errors.New("missing rename source")
			}
			r.OrigPath = string(parts[i])
		}
		rows = append(rows, r)
	}
	staged, e := g.numstat(ctx, true)
	if e != nil {
		return nil, e
	}
	unstaged, e := g.numstat(ctx, false)
	if e != nil {
		return nil, e
	}
	for i := range rows {
		r := &rows[i]
		a := staged[r.Path]
		r.StagedAdditions = a[0]
		r.StagedDeletions = a[1]
		a = unstaged[r.Path]
		r.UnstagedAdditions = a[0]
		r.UnstagedDeletions = a[1]
		if r.Index == '?' {
			info, e := os.Lstat(filepath.Join(g.Dir, r.Path))
			if e == nil && info.Mode().IsRegular() && info.Size() <= 2*1024*1024 {
				b, e := os.ReadFile(filepath.Join(g.Dir, r.Path))
				if e == nil && !bytes.ContainsRune(b, 0) {
					r.UnstagedAdditions = bytes.Count(b, []byte{'\n'})
					if len(b) > 0 && b[len(b)-1] != '\n' {
						r.UnstagedAdditions++
					}
				}
			}
		}
	}
	return rows, nil
}
func (g Git) numstat(ctx context.Context, staged bool) (map[string][2]int, error) {
	args := []string{"diff", "--numstat", "-z", "--no-renames"}
	if staged {
		args = append(args, "--cached")
	}
	raw, e := g.Run(ctx, args...)
	if e != nil {
		return nil, e
	}
	return parseNumstat(raw), nil
}
func parseNumstat(raw []byte) map[string][2]int {
	m := map[string][2]int{}
	for _, p := range bytes.Split(raw, []byte{0}) {
		v := bytes.SplitN(p, []byte{'\t'}, 3)
		if len(v) != 3 {
			continue
		}
		a, _ := strconv.Atoi(string(v[0]))
		d, _ := strconv.Atoi(string(v[1]))
		m[string(v[2])] = [2]int{a, d}
	}
	return m
}
func (g Git) Stage(ctx context.Context, paths ...string) error {
	ps, e := literalPaths(paths)
	if e != nil {
		return e
	}
	return g.do(ctx, append([]string{"add", "--"}, ps...)...)
}
func (g Git) StageAll(ctx context.Context) error { return g.do(ctx, "add", "-A", "--", ".") }
func (g Git) Unstage(ctx context.Context, paths ...string) error {
	ps, e := literalPaths(paths)
	if e != nil {
		return e
	}
	ok, e := g.HasCommits(ctx)
	if e != nil {
		return e
	}
	args := []string{"reset", "-q", "HEAD", "--"}
	if !ok {
		args = []string{"rm", "--cached", "-r", "--ignore-unmatch", "--"}
	}
	return g.do(ctx, append(args, ps...)...)
}
func (g Git) UnstageAll(ctx context.Context) error { return g.Unstage(ctx, ".") }
func (g Git) Discard(ctx context.Context, paths ...string) error {
	ps, e := literalPaths(paths)
	if e != nil {
		return e
	}
	raw, e := g.Run(ctx, append([]string{"ls-files", "--cached", "-z", "--"}, ps...)...)
	if e != nil {
		return e
	}
	var tracked []string
	for _, p := range bytes.Split(raw, []byte{0}) {
		if len(p) > 0 {
			tracked = append(tracked, ":(literal)"+string(p))
		}
	}
	if len(tracked) > 0 {
		if e := g.do(ctx, append([]string{"restore", "--worktree", "--"}, tracked...)...); e != nil {
			return e
		}
	}
	return g.do(ctx, append([]string{"clean", "-fdq", "--"}, ps...)...)
}
func (g Git) DiscardAll(ctx context.Context) error { return g.Discard(ctx, ".") }
func (g Git) Commit(ctx context.Context, message string, includeUnstaged bool) (string, error) {
	if strings.TrimSpace(message) == "" {
		return "", errors.New("commit message cannot be empty")
	}
	if includeUnstaged {
		if e := g.StageAll(ctx); e != nil {
			return "", e
		}
	}
	return g.text(ctx, "commit", "-m", message)
}
func (g Git) StagedPatch(ctx context.Context) (string, error) {
	b, e := g.Run(ctx, "diff", "--cached", "--no-ext-diff", "--no-color")
	return string(b), e
}
func (g Git) UnstagedPatch(ctx context.Context) (string, error) {
	b, e := g.Run(ctx, "diff", "--no-ext-diff", "--no-color")
	return string(b), e
}
func (g Git) FileAtRevision(ctx context.Context, rev, path string) ([]byte, error) {
	if e := validRevision(rev); e != nil {
		return nil, e
	}
	if _, e := literalPaths([]string{path}); e != nil {
		return nil, e
	}
	return g.Run(ctx, "show", rev+":"+filepath.ToSlash(path))
}

type AheadBehind struct {
	Ahead, Behind int
	HasUpstream   bool
	Upstream      string
}

func (g Git) AheadBehind(ctx context.Context) (AheadBehind, error) {
	var result AheadBehind
	up, e := g.text(ctx, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if e != nil {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if _, err := g.IsRepo(ctx); err != nil {
			return result, err
		}
		var ce *CommandError
		if errors.As(e, &ce) && (strings.Contains(ce.Stderr, "no upstream") || strings.Contains(ce.Stderr, "no such branch") || strings.Contains(ce.Stderr, "HEAD does not point to a branch")) {
			return result, nil
		}
		return result, e
	}
	result.Upstream = up
	result.HasUpstream = true
	s, e := g.text(ctx, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if e != nil {
		return result, e
	}
	_, e = fmt.Sscanf(s, "%d %d", &result.Ahead, &result.Behind)
	return result, e
}
func (g Git) Fetch(ctx context.Context) error          { return g.do(ctx, "fetch", "--quiet") }
func (g Git) Pull(ctx context.Context) (string, error) { return g.text(ctx, "pull", "--ff-only") }
func (g Git) MergeUpstream(ctx context.Context) (string, error) {
	return g.text(ctx, "pull", "--no-rebase", "--ff", "--no-edit")
}
func (g Git) RebaseUpstream(ctx context.Context) (string, error) {
	return g.text(ctx, "pull", "--rebase", "--autostash")
}
func (g Git) Push(ctx context.Context, forceWithLease bool) (string, error) {
	s, e := g.AheadBehind(ctx)
	if e != nil {
		return "", e
	}
	args := []string{"push"}
	if forceWithLease {
		args = append(args, "--force-with-lease")
	}
	if !s.HasUpstream {
		branch, e := g.CurrentBranch(ctx)
		if e != nil {
			return "", e
		}
		if branch == "" {
			return "", errors.New("cannot publish detached HEAD")
		}
		args = append(args, "-u", "origin", branch)
	}
	return g.text(ctx, args...)
}
func (g Git) Sync(ctx context.Context) (string, error) {
	if e := g.Fetch(ctx); e != nil {
		return "", e
	}
	s, e := g.AheadBehind(ctx)
	if e != nil {
		return "", e
	}
	if !s.HasUpstream {
		return "Fetched; branch has no upstream", nil
	}
	var messages []string
	if s.Behind > 0 {
		var out string
		if s.Ahead > 0 {
			out, e = g.MergeUpstream(ctx)
		} else {
			out, e = g.Pull(ctx)
		}
		if e != nil {
			return "", e
		}
		messages = append(messages, out)
	}
	s, e = g.AheadBehind(ctx)
	if e != nil {
		return "", e
	}
	if s.Ahead > 0 {
		out, e := g.Push(ctx, false)
		if e != nil {
			return "", e
		}
		messages = append(messages, out)
	}
	return strings.Join(messages, "\n"), nil
}

type RefKind string

const (
	RefHead   RefKind = "head"
	RefBranch RefKind = "branch"
	RefRemote RefKind = "remote"
	RefTag    RefKind = "tag"
)

type RefLabel struct {
	Name string
	Kind RefKind
}
type CommitEntry struct {
	Hash, Short, Author, AuthorEmail, Relative, Subject string
	Refs                                                []RefLabel
}
type HistoryOptions struct {
	Limit, Skip        int
	AllBranches        bool
	Author, Path, Grep string
	Follow             bool
}

func (g Git) History(ctx context.Context, o HistoryOptions) ([]CommitEntry, error) {
	if o.Limit <= 0 {
		o.Limit = 100
	}
	if o.Skip < 0 {
		return nil, errors.New("negative history offset")
	}
	args := []string{"log", "--decorate=full", "--date-order", "-z", fmt.Sprintf("--max-count=%d", o.Limit), fmt.Sprintf("--skip=%d", o.Skip), "--format=%H%x1f%h%x1f%an%x1f%ae%x1f%ar%x1f%s%x1f%D"}
	if o.AllBranches {
		args = append(args, "--all")
	}
	if o.Author != "" {
		args = append(args, "--author="+o.Author)
	}
	if o.Grep != "" {
		args = append(args, "-i", "--grep="+o.Grep)
	}
	if o.Path != "" {
		ps, e := literalPaths([]string{o.Path})
		if e != nil {
			return nil, e
		}
		if o.Follow {
			args = append(args, "--follow")
		}
		args = append(args, "--", ps[0])
	}
	raw, e := g.Run(ctx, args...)
	if e != nil {
		ok, err := g.HasCommits(ctx)
		if err == nil && !ok {
			return nil, nil
		}
		return nil, e
	}
	var out []CommitEntry
	for _, r := range bytes.Split(raw, []byte{0}) {
		if len(r) == 0 {
			continue
		}
		p := strings.SplitN(string(r), "\x1f", 7)
		if len(p) != 7 {
			return nil, errors.New("malformed Git history")
		}
		out = append(out, CommitEntry{p[0], p[1], p[2], p[3], p[4], p[5], parseRefs(p[6])})
	}
	return out, nil
}
func parseRefs(raw string) []RefLabel {
	var out []RefLabel
	for _, r := range strings.Split(raw, ", ") {
		if r == "" {
			continue
		}
		kind := RefBranch
		if strings.HasPrefix(r, "HEAD -> ") {
			r = strings.TrimPrefix(r, "HEAD -> ")
			kind = RefHead
		} else if r == "HEAD" {
			kind = RefHead
		} else if strings.HasPrefix(r, "tag: ") {
			r = strings.TrimPrefix(r, "tag: ")
			kind = RefTag
		} else if strings.HasPrefix(r, "refs/remotes/") {
			kind = RefRemote
		}
		for _, prefix := range []string{"refs/heads/", "refs/remotes/", "refs/tags/"} {
			r = strings.TrimPrefix(r, prefix)
		}
		out = append(out, RefLabel{Name: r, Kind: kind})
	}
	return out
}

type CommitFile struct {
	Path, Status         string
	Additions, Deletions int
}
type CommitDetail struct {
	Hash, Short, Subject, Author, AuthorEmail, AuthorDate, Committer, CommitterEmail, CommitterDate, Body string
	Parents                                                                                               []string
	Refs                                                                                                  []RefLabel
	Files                                                                                                 []CommitFile
}

func (g Git) CommitDetail(ctx context.Context, rev string) (CommitDetail, error) {
	var d CommitDetail
	if e := validRevision(rev); e != nil {
		return d, e
	}
	raw, e := g.Run(ctx, "show", "-s", "--decorate=full", "--format=%H%x00%h%x00%s%x00%an%x00%ae%x00%aI%x00%cn%x00%ce%x00%cI%x00%P%x00%D%x00%b", rev, "--")
	if e != nil {
		return d, e
	}
	p := bytes.SplitN(raw, []byte{0}, 12)
	if len(p) != 12 {
		return d, errors.New("malformed commit detail")
	}
	d = CommitDetail{Hash: string(p[0]), Short: string(p[1]), Subject: string(p[2]), Author: string(p[3]), AuthorEmail: string(p[4]), AuthorDate: string(p[5]), Committer: string(p[6]), CommitterEmail: string(p[7]), CommitterDate: string(p[8]), Parents: strings.Fields(string(p[9])), Refs: parseRefs(string(p[10])), Body: strings.TrimSpace(string(p[11]))}
	raw, e = g.Run(ctx, "diff-tree", "--root", "--no-commit-id", "-r", "--name-status", "-z", "--no-renames", rev, "--")
	if e != nil {
		return d, e
	}
	parts := bytes.Split(raw, []byte{0})
	for i := 0; i+1 < len(parts) && len(parts[i]) > 0; i += 2 {
		d.Files = append(d.Files, CommitFile{Path: string(parts[i+1]), Status: string(parts[i])})
	}
	raw, e = g.Run(ctx, "diff-tree", "--root", "--no-commit-id", "-r", "--numstat", "-z", "--no-renames", rev, "--")
	if e != nil {
		return d, e
	}
	ns := parseNumstat(raw)
	for i := range d.Files {
		n := ns[d.Files[i].Path]
		d.Files[i].Additions = n[0]
		d.Files[i].Deletions = n[1]
	}
	return d, nil
}

type GraphCommit struct {
	CommitEntry
	Parents []string
}
type GraphRow struct {
	Commit              GraphCommit
	NodeLane, LaneCount int
	Verticals           []int
	NodeHasIncoming     bool
	Parents             []int
}

func (g Git) GraphHistory(ctx context.Context, limit int, allRefs bool) ([]GraphCommit, error) {
	if limit <= 0 {
		limit = 100
	}
	args := []string{"log", "--decorate=full", "--branches", "HEAD", "--date-order", "-z", fmt.Sprintf("--max-count=%d", limit), "--format=%H%x1f%h%x1f%an%x1f%ae%x1f%ar%x1f%s%x1f%P%x1f%D"}
	if allRefs {
		args = append(args, "--remotes", "--tags")
	}
	raw, e := g.Run(ctx, args...)
	if e != nil {
		ok, err := g.HasCommits(ctx)
		if err == nil && !ok {
			return nil, nil
		}
		return nil, e
	}
	var out []GraphCommit
	for _, r := range bytes.Split(raw, []byte{0}) {
		if len(r) == 0 {
			continue
		}
		p := strings.SplitN(string(r), "\x1f", 8)
		if len(p) != 8 {
			return nil, errors.New("malformed Git graph")
		}
		out = append(out, GraphCommit{CommitEntry{p[0], p[1], p[2], p[3], p[4], p[5], parseRefs(p[7])}, strings.Fields(p[6])})
	}
	return out, nil
}
func LayoutGraph(commits []GraphCommit) []GraphRow {
	known := map[string]bool{}
	for _, c := range commits {
		known[c.Hash] = true
	}
	var lanes []string
	var rows []GraphRow
	for _, c := range commits {
		node := -1
		for i, h := range lanes {
			if h == c.Hash {
				node = i
				break
			}
		}
		if node < 0 {
			node = len(lanes)
			lanes = append(lanes, "")
		}
		top := append([]string(nil), lanes...)
		row := GraphRow{Commit: c, NodeLane: node, NodeHasIncoming: top[node] != ""}
		lanes[node] = ""
		visible := 0
		for _, parent := range c.Parents {
			if !known[parent] {
				continue
			}
			lane := -1
			for i, h := range lanes {
				if h == parent {
					lane = i
					break
				}
			}
			if lane < 0 {
				if visible == 0 && lanes[node] == "" {
					lane = node
				} else {
					for i := node + 1; i < len(lanes); i++ {
						if lanes[i] == "" {
							lane = i
							break
						}
					}
					if lane < 0 {
						lane = len(lanes)
						lanes = append(lanes, "")
					}
				}
				lanes[lane] = parent
			}
			visible++
			row.Parents = append(row.Parents, lane)
		}
		for i, h := range top {
			if i != node && h != "" {
				row.Verticals = append(row.Verticals, i)
			}
		}
		for len(lanes) > node+1 && lanes[len(lanes)-1] == "" {
			lanes = lanes[:len(lanes)-1]
		}
		row.LaneCount = len(lanes)
		for _, p := range row.Parents {
			if p+1 > row.LaneCount {
				row.LaneCount = p + 1
			}
		}
		rows = append(rows, row)
	}
	return rows
}

type RemoteWeb struct{ Base, Forge, Owner, Name string }

func ParseRemoteWeb(raw string) (RemoteWeb, error) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") && strings.Contains(raw, ":") {
		raw = raw[strings.LastIndex(raw, "@")+1:]
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return RemoteWeb{}, errors.New("invalid SSH remote")
		}
		raw = "https://" + parts[0] + "/" + parts[1]
	}
	u, e := url.Parse(raw)
	if e != nil || u.Host == "" {
		return RemoteWeb{}, errors.New("remote has no web address")
	}
	host := strings.ToLower(u.Hostname())
	forge := ""
	switch {
	case host == "github.com" || host == "www.github.com":
		forge = "github"
	case host == "gitlab.com" || host == "www.gitlab.com":
		forge = "gitlab"
	case host == "bitbucket.org" || host == "www.bitbucket.org":
		forge = "bitbucket"
	default:
		return RemoteWeb{}, errors.New("unsupported remote forge")
	}
	u.Scheme = "https"
	u.Host = host
	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	u.Path = strings.TrimSuffix(strings.TrimSuffix(u.Path, "/"), ".git")
	path := strings.Trim(u.Path, "/")
	last := strings.LastIndex(path, "/")
	if last < 1 || last == len(path)-1 {
		return RemoteWeb{}, errors.New("remote has no owner/repository path")
	}
	return RemoteWeb{Base: strings.TrimSuffix(u.String(), "/"), Forge: forge, Owner: path[:last], Name: path[last+1:]}, nil
}
func (r RemoteWeb) CommitURL(hash string) string {
	switch r.Forge {
	case "gitlab":
		return r.Base + "/-/commit/" + url.PathEscape(hash)
	case "bitbucket":
		return r.Base + "/commits/" + url.PathEscape(hash)
	default:
		return r.Base + "/commit/" + url.PathEscape(hash)
	}
}
func (g Git) RemoteWeb(ctx context.Context) (RemoteWeb, error) {
	raw, e := g.text(ctx, "remote", "get-url", "origin")
	if e != nil {
		return RemoteWeb{}, e
	}
	return ParseRemoteWeb(raw)
}
