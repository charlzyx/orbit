package workspace

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const DefaultWorktreeDirectory = ".wt"
const DefaultSetupScript = ".orbit/worktree-setup.sh"

type Worktree struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Head        string `json:"head"`
	Branch      string `json:"branch,omitempty"`
	Detached    bool   `json:"detached"`
	Main        bool   `json:"main"`
	Bare        bool   `json:"bare"`
	Locked      bool   `json:"locked"`
	LockReason  string `json:"lockReason,omitempty"`
	Prunable    bool   `json:"prunable"`
	PruneReason string `json:"pruneReason,omitempty"`
}
type WorktreeError struct {
	Kind   string
	Path   string
	Branch string
	Cause  error
}

func (e *WorktreeError) Error() string {
	detail := e.Path
	if e.Branch != "" {
		detail += " (" + e.Branch + ")"
	}
	if e.Cause != nil {
		detail += ": " + e.Cause.Error()
	}
	return "worktree " + e.Kind + ": " + detail
}
func (e *WorktreeError) Unwrap() error { return e.Cause }

type WorktreeConfig struct {
	initialized       bool
	Directory         string   `json:"directory"`
	SetupScript       string   `json:"setup_script"`
	RunSetupAuto      bool     `json:"run_setup_auto"`
	OpenAfterCreate   bool     `json:"open_after_create"`
	AllowedSetupRepos []string `json:"allowed_setup_repos"`
}

func DefaultWorktreeConfig() WorktreeConfig {
	return WorktreeConfig{initialized: true, Directory: DefaultWorktreeDirectory, SetupScript: DefaultSetupScript, RunSetupAuto: true, OpenAfterCreate: true}
}
func LoadWorktreeConfig(path string) WorktreeConfig {
	c := DefaultWorktreeConfig()
	raw, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	if json.Unmarshal(raw, &c) != nil {
		return DefaultWorktreeConfig()
	}
	return c
}
func SaveWorktreeConfig(path string, c WorktreeConfig) error {
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".worktrees-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
func canonicalPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		p = filepath.Clean(path)
	}
	if real, e := filepath.EvalSymlinks(p); e == nil {
		return real
	}
	var suffix []string
	cur := p
	for {
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		suffix = append(suffix, filepath.Base(cur))
		if real, e := filepath.EvalSymlinks(parent); e == nil {
			for i := len(suffix) - 1; i >= 0; i-- {
				real = filepath.Join(real, suffix[i])
			}
			return real
		}
		cur = parent
	}
	return filepath.Clean(p)
}
func SamePath(a, b string) bool { return canonicalPath(a) == canonicalPath(b) }
func ExpandPath(base, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "~" || strings.HasPrefix(raw, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			if raw == "~" {
				return home
			}
			return filepath.Join(home, raw[2:])
		}
	}
	if filepath.IsAbs(raw) {
		return filepath.Clean(raw)
	}
	return filepath.Join(base, raw)
}
func (c WorktreeConfig) Effective(root string) (directory, script string) {
	var repo struct {
		Directory   *string `json:"directory"`
		SetupScript *string `json:"setup_script"`
	}
	if raw, e := os.ReadFile(filepath.Join(root, ".orbit", "worktree.json")); e == nil {
		if e = json.Unmarshal(raw, &repo); e != nil {
			repo.Directory = nil
			repo.SetupScript = nil
		}
	}
	dir, setup := c.Directory, c.SetupScript
	// A literal zero-value manager is convenient; loaded/default configs keep
	// intentional empty strings (an empty setup path disables the script).
	if !c.initialized && dir == "" && setup == "" && !c.RunSetupAuto && !c.OpenAfterCreate && len(c.AllowedSetupRepos) == 0 {
		dir = DefaultWorktreeDirectory
		setup = DefaultSetupScript
	}
	if repo.Directory != nil {
		dir = *repo.Directory
	}
	if repo.SetupScript != nil {
		setup = *repo.SetupScript
	}
	return ExpandPath(root, dir), ExpandPath(root, setup)
}
func (c WorktreeConfig) SetupAllowed(root string) bool {
	key := canonicalPath(root)
	for _, p := range c.AllowedSetupRepos {
		if canonicalPath(p) == key {
			return true
		}
	}
	return false
}
func (c *WorktreeConfig) AllowSetup(root string) {
	if !c.SetupAllowed(root) {
		c.AllowedSetupRepos = append(c.AllowedSetupRepos, canonicalPath(root))
	}
}

// ParseWorktrees parses -z porcelain without splitting paths on whitespace/LF.
func ParseWorktrees(raw []byte) []Worktree {
	var out []Worktree
	var cur *Worktree
	flush := func() {
		if cur != nil {
			cur.Main = len(out) == 0
			cur.Name = filepath.Base(cur.Path)
			out = append(out, *cur)
			cur = nil
		}
	}
	for _, field := range strings.Split(string(raw), "\x00") {
		if field == "" {
			flush()
			continue
		}
		if strings.HasPrefix(field, "worktree ") {
			flush()
			cur = &Worktree{Path: strings.TrimPrefix(field, "worktree ")}
			continue
		}
		if cur == nil {
			continue
		}
		switch {
		case strings.HasPrefix(field, "HEAD "):
			cur.Head = strings.TrimPrefix(field, "HEAD ")
		case strings.HasPrefix(field, "branch "):
			cur.Branch = strings.TrimPrefix(strings.TrimPrefix(field, "branch "), "refs/heads/")
		case field == "detached":
			cur.Detached = true
		case field == "bare":
			cur.Bare = true
		case field == "locked":
			cur.Locked = true
		case strings.HasPrefix(field, "locked "):
			cur.Locked = true
			cur.LockReason = strings.TrimPrefix(field, "locked ")
		case field == "prunable":
			cur.Prunable = true
		case strings.HasPrefix(field, "prunable "):
			cur.Prunable = true
			cur.PruneReason = strings.TrimPrefix(field, "prunable ")
		}
	}
	flush()
	return out
}

type WorktreeManager struct {
	Dir    string
	Binary string
	Config WorktreeConfig
}

var worktreeMutationLocks sync.Map

// Serialize even separately constructed managers for the same common Git dir.
// Waiting respects cancellation, so an abandoned job never mutates later.
func (m *WorktreeManager) lockMutation(ctx context.Context) (func(), error) {
	common, e := m.git().text(ctx, "rev-parse", "--git-common-dir")
	if e != nil {
		return nil, e
	}
	if !filepath.IsAbs(common) {
		common = filepath.Join(m.Dir, common)
	}
	key := canonicalPath(common)
	v, _ := worktreeMutationLocks.LoadOrStore(key, make(chan struct{}, 1))
	lock := v.(chan struct{})
	select {
	case lock <- struct{}{}:
		var once sync.Once
		return func() { once.Do(func() { <-lock }) }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func (m *WorktreeManager) git() Git { return Git{Dir: m.Dir, Binary: m.Binary} }
func (m *WorktreeManager) List(ctx context.Context) ([]Worktree, error) {
	raw, err := m.git().Run(ctx, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		raw, err = m.git().Run(ctx, "-c", "core.quotePath=false", "worktree", "list", "--porcelain")
		if err != nil {
			return nil, &WorktreeError{Kind: "not_repository", Path: m.Dir, Cause: err}
		}
		raw = []byte(strings.ReplaceAll(string(raw), "\n", "\x00"))
	}
	return ParseWorktrees(raw), nil
}
func (m *WorktreeManager) RepositoryRoot(ctx context.Context) (string, error) {
	ws, err := m.List(ctx)
	if err != nil {
		return "", err
	}
	if len(ws) == 0 {
		return "", &WorktreeError{Kind: "not_repository", Path: m.Dir}
	}
	return ws[0].Path, nil
}
func (m *WorktreeManager) ResolvePath(ctx context.Context, name string) (string, error) {
	n, err := ValidateName(name)
	if err != nil {
		return "", &WorktreeError{Kind: "invalid_name", Path: name, Cause: err}
	}
	root, err := m.RepositoryRoot(ctx)
	if err != nil {
		return "", err
	}
	dir, _ := m.Config.Effective(root)
	p := filepath.Join(dir, n)
	if !pathUnder(canonicalPath(p), canonicalPath(dir)) {
		return "", &WorktreeError{Kind: "invalid_name", Path: name}
	}
	return p, nil
}
func (m *WorktreeManager) FindByPath(ctx context.Context, path string) (Worktree, error) {
	ws, err := m.List(ctx)
	if err != nil {
		return Worktree{}, err
	}
	for _, w := range ws {
		if SamePath(w.Path, path) {
			return w, nil
		}
	}
	return Worktree{}, &WorktreeError{Kind: "not_worktree", Path: path}
}
func (m *WorktreeManager) Dirty(ctx context.Context, path string) (bool, error) {
	raw, err := (Git{Dir: path, Binary: m.Binary}).Run(ctx, "status", "--porcelain=v1", "-z", "--untracked-files=normal")
	return len(raw) > 0, err
}

type CreateWorktreeOptions struct {
	Name         string
	Branch       string
	NewBranch    bool
	StartPoint   string
	RunSetup     bool
	ConfirmSetup bool
}
type WorktreeCreation struct {
	Worktree Worktree
	Setup    *SetupOutcome
}

func (m *WorktreeManager) Create(ctx context.Context, o CreateWorktreeOptions) (WorktreeCreation, error) {
	var result WorktreeCreation
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return result, err
	}
	defer unlock()
	if err := ctx.Err(); err != nil {
		return result, err
	}
	path, err := m.ResolvePath(ctx, o.Name)
	if err != nil {
		return result, err
	}
	if _, err = os.Lstat(path); err == nil {
		return result, &WorktreeError{Kind: "path_exists", Path: path}
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	if err = validRevision(o.Branch); err != nil {
		return result, &WorktreeError{Kind: "invalid_branch", Branch: o.Branch, Cause: err}
	}
	if _, err = m.git().Run(ctx, "check-ref-format", "--branch", o.Branch); err != nil {
		return result, &WorktreeError{Kind: "invalid_branch", Branch: o.Branch, Cause: err}
	}
	ws, err := m.List(ctx)
	if err != nil {
		return result, err
	}
	for _, w := range ws {
		if w.Branch == o.Branch {
			return result, &WorktreeError{Kind: "branch_checked_out", Path: w.Path, Branch: o.Branch}
		}
	}
	root := ws[0].Path
	_, script := m.Config.Effective(root)
	runSetup := o.RunSetup || m.Config.RunSetupAuto
	fi, scriptErr := os.Stat(script)
	hasScript := scriptErr == nil && fi.Mode().IsRegular()
	if runSetup && hasScript && !o.ConfirmSetup && !m.Config.SetupAllowed(root) {
		return result, &WorktreeError{Kind: "setup_confirmation_required", Path: script, Cause: ErrConfirmationRequired}
	}
	if o.StartPoint != "" {
		if err = validRevision(o.StartPoint); err != nil {
			return result, err
		}
	}
	if err = os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return result, err
	}
	args := []string{"worktree", "add"}
	if o.NewBranch {
		args = append(args, "-b", o.Branch, path)
		if o.StartPoint != "" {
			args = append(args, "--", o.StartPoint)
		}
	} else {
		args = append(args, path, "--", o.Branch)
	}
	if _, err = m.git().Run(ctx, args...); err != nil {
		return result, err
	}
	result.Worktree, err = m.FindByPath(ctx, path)
	if err != nil {
		return result, err
	}
	unlock()
	if runSetup && hasScript {
		outcome := RunSetupScript(ctx, script, root, path, true)
		result.Setup = &outcome
	}
	return result, nil
}
func (m *WorktreeManager) Remove(ctx context.Context, path string, force bool) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	w, err := m.FindByPath(ctx, path)
	if err != nil {
		return err
	}
	if w.Main {
		return &WorktreeError{Kind: "main_worktree", Path: path}
	}
	if w.Locked {
		return &WorktreeError{Kind: "locked", Path: path}
	}
	if !force {
		if _, e := os.Stat(w.Path); e == nil {
			dirty, e := m.Dirty(ctx, w.Path)
			if e != nil {
				return e
			}
			if dirty {
				return &WorktreeError{Kind: "dirty", Path: path}
			}
		}
	}
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--", w.Path)
	_, err = m.git().Run(ctx, args...)
	return err
}
func (m *WorktreeManager) Move(ctx context.Context, path, destination string) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	w, err := m.FindByPath(ctx, path)
	if err != nil {
		return err
	}
	if w.Main {
		return &WorktreeError{Kind: "main_worktree", Path: path}
	}
	if w.Locked {
		return &WorktreeError{Kind: "locked", Path: path}
	}
	destination, err = m.ResolveDestination(ctx, destination)
	if err != nil {
		return err
	}
	if SamePath(path, destination) {
		return nil
	}
	if _, e := os.Lstat(destination); e == nil {
		return ErrDestinationExists
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	_, err = m.git().Run(ctx, "worktree", "move", "--", w.Path, destination)
	return err
}
func (m *WorktreeManager) Rename(ctx context.Context, path, name string) (string, error) {
	dst, err := m.ResolvePath(ctx, name)
	if err != nil {
		return "", err
	}
	return dst, m.Move(ctx, path, dst)
}
func (m *WorktreeManager) Lock(ctx context.Context, path, reason string) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	w, err := m.FindByPath(ctx, path)
	if err != nil {
		return err
	}
	if w.Locked {
		return nil
	}
	args := []string{"worktree", "lock"}
	if strings.TrimSpace(reason) != "" {
		args = append(args, "--reason", reason)
	}
	args = append(args, "--", w.Path)
	_, err = m.git().Run(ctx, args...)
	return err
}
func (m *WorktreeManager) Unlock(ctx context.Context, path string) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	w, err := m.FindByPath(ctx, path)
	if err != nil {
		return err
	}
	if !w.Locked {
		return nil
	}
	_, err = m.git().Run(ctx, "worktree", "unlock", "--", w.Path)
	return err
}
func (m *WorktreeManager) Prune(ctx context.Context) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	_, err = m.git().Run(ctx, "worktree", "prune")
	return err
}
func (m *WorktreeManager) Repair(ctx context.Context, paths ...string) error {
	unlock, err := m.lockMutation(ctx)
	if err != nil {
		return err
	}
	defer unlock()
	args := []string{"worktree", "repair", "--"}
	for _, p := range paths {
		if p == "" || strings.ContainsRune(p, 0) {
			return fmt.Errorf("invalid repair path")
		}
		args = append(args, p)
	}
	_, err = m.git().Run(ctx, args...)
	return err
}

// WorktreeSlug matches Orbit's ASCII-only suggestion policy. It is used only
// for defaults and never normalizes a name or branch explicitly entered by a user.
func WorktreeSlug(raw string) string {
	var out strings.Builder
	dash := false
	for _, r := range raw {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' {
			if dash && out.Len() != 0 {
				out.WriteByte('-')
			}
			dash = false
			if r >= 'A' && r <= 'Z' {
				r += 'a' - 'A'
			}
			out.WriteRune(r)
		} else {
			dash = true
		}
	}
	return strings.Trim(strings.Trim(out.String(), "-"), ".")
}

// ResolveDestination matches the worktree controller's move semantics: relative
// paths are based on the main repository and ~/ is expanded, never process cwd.
func (m *WorktreeManager) ResolveDestination(ctx context.Context, raw string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return "", &WorktreeError{Kind: "invalid_destination", Path: raw}
	}
	root, err := m.RepositoryRoot(ctx)
	if err != nil {
		return "", err
	}
	return canonicalPath(ExpandPath(root, raw)), nil
}
