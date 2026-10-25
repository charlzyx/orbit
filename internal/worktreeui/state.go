// Package worktreeui is the native worktree page. State belongs to the UI
// thread; workers exchange immutable results tagged with workspace generation.
package worktreeui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
)

type Options struct {
	Translate      func(string) string
	OnChange       func()
	OpenWorkspace  func(string)
	OpenTerminal   func(string)
	RevealPath     func(string)
	Config         *workspace.WorktreeConfig
	ConfigPath     string
	HideHowTo      bool
	OnDismissHowTo func()
}
type State struct {
	Workspace                                            string
	Options                                              Options
	Config                                               workspace.WorktreeConfig
	configDraft                                          workspace.WorktreeConfig
	Rows                                                 []workspace.Worktree
	Branches                                             []string
	Root                                                 string
	Dirty                                                map[string]bool
	Busy                                                 bool
	Activity, Status                                     string
	Error                                                error
	Setup                                                *workspace.SetupOutcome
	SetupWorktree                                        string
	Filter                                               string
	ShowHowTo                                            bool
	ShowOutput                                           bool
	generation                                           uint64
	results                                              chan result
	cancel                                               context.CancelFunc
	initialized, closed                                  bool
	refreshed                                            time.Time
	dialog                                               string
	dialogOpen                                           bool
	name, branch, startPoint, value, existingBranch      string
	newBranch, runSetup, openAfter, force, branchTouched bool
	target                                               workspace.Worktree
	confirmOpen                                          bool
	confirmTitle, confirmDetail, confirmLabel            string
	confirmAction                                        func()
	pending                                              *workspace.CreateWorktreeOptions
	lastAction                                           func()
}
type result struct {
	generation    uint64
	rows          []workspace.Worktree
	branches      []string
	dirty         map[string]bool
	root          string
	apply         func(*State)
	err           error
	status        string
	preserveError bool
}

func New(dir string, o Options) *State {
	cfg := workspace.DefaultWorktreeConfig()
	if o.ConfigPath != "" {
		cfg = workspace.LoadWorktreeConfig(o.ConfigPath)
	}
	if o.Config != nil {
		cfg = *o.Config
	}
	cfg.AllowedSetupRepos = append([]string(nil), cfg.AllowedSetupRepos...)
	return &State{Workspace: dir, Options: o, Config: cfg, Dirty: map[string]bool{}, results: make(chan result, 16), ShowHowTo: !o.HideHowTo}
}
func (s *State) SetWorkspace(dir string) {
	if workspace.SamePath(dir, s.Workspace) {
		return
	}
	s.Cancel()
	s.Workspace = dir
	s.Rows = nil
	s.Branches = nil
	s.Root = ""
	s.Dirty = map[string]bool{}
	s.Error = nil
	s.Setup = nil
	s.Status = ""
	s.dialogOpen = false
	s.confirmOpen = false
	s.pending = nil
	s.lastAction = nil
	s.Filter = ""
	s.ShowOutput = false
	s.initialized = false
}
func (s *State) Cancel() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.generation++
	s.Busy = false
	s.Activity = ""
	s.initialized = false
}
func (s *State) Close() { s.Cancel(); s.closed = true }
func (s *State) start(label string, job func(context.Context, *workspace.WorktreeManager) result) bool {
	if s.Busy || s.closed {
		return false
	}
	s.Busy = true
	s.Activity = label
	if label != "Refreshing worktrees" {
		s.Error = nil
		s.Status = ""
	}
	s.initialized = true
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	cfg := s.Config
	cfg.AllowedSetupRepos = append([]string(nil), cfg.AllowedSetupRepos...)
	m := &workspace.WorktreeManager{Dir: s.Workspace, Config: cfg}
	channel, wake := s.results, s.Options.OnChange
	go func() {
		defer cancel()
		r := job(ctx, m)
		r.generation = generation
		if r.err == nil {
			rows, e := m.List(ctx)
			if e != nil {
				r.err = e
			} else {
				r.rows = rows
				r.dirty = map[string]bool{}
				for _, w := range rows {
					if !w.Bare {
						dirty, e := m.Dirty(ctx, w.Path)
						if e == nil {
							r.dirty[w.Path] = dirty
						}
					}
				}
				if len(rows) > 0 {
					r.root = rows[0].Path
				}
				g := workspace.Git{Dir: m.Dir}
				r.branches, _ = g.Branches(ctx)
				if refs, e := g.Refs(ctx); e == nil {
					for _, ref := range refs {
						if ref.Kind == workspace.RefRemote {
							r.branches = append(r.branches, ref.Name)
						}
					}
				}
			}
		}
		select {
		case channel <- r:
			if wake != nil {
				wake()
			}
		case <-ctx.Done():
		}
	}()
	return true
}
func (s *State) Tick() {
	for {
		select {
		case r := <-s.results:
			if r.generation != s.generation || s.closed {
				continue
			}
			s.Busy = false
			s.Activity = ""
			s.cancel = nil
			s.refreshed = time.Now()
			if !r.preserveError || r.err != nil {
				s.Error = r.err
			}
			if r.rows != nil {
				s.Rows = r.rows
				s.Dirty = r.dirty
				s.Root = r.root
				s.Branches = r.branches
			}
			if r.status != "" && r.err == nil {
				s.Status = r.status
			}
			if r.apply != nil {
				r.apply(s)
			}
			var we *workspace.WorktreeError
			if errors.As(r.err, &we) && we.Kind == "branch_checked_out" {
				s.pending = nil
				s.target = workspace.Worktree{Path: we.Path, Branch: we.Branch}
				s.dialog = "branch_in_use"
				s.dialogOpen = true
			}
			if errors.As(r.err, &we) && we.Kind == "setup_confirmation_required" && s.pending != nil {
				pending := *s.pending
				s.confirm(s.text("worktree.allow_setup.title", "Run this repository's setup script?"), s.text("worktree.allow_setup.body", "This repository contains %{script}.\n\nIt will execute commands on your machine after %{name} is created. Orbit remembers this choice for the repository.", "script", we.Path, "name", pending.Name), func() { s.allowSetupAndCreate(pending) })
				s.confirmLabel = s.text("worktree.allow_setup.allow", "Allow & Create")
			}
		default:
			return
		}
	}
}
func (s *State) Refresh() bool {
	return s.start("Refreshing worktrees", func(context.Context, *workspace.WorktreeManager) result { return result{preserveError: true} })
}
func (s *State) WatchChanged() {
	if !s.Busy && time.Since(s.refreshed) >= 2*time.Second {
		s.Refresh()
	}
}

// applyConfig commits only after durable storage succeeds. A failed trust or
// settings write must stay visible and must not accidentally start a mutation.
func (s *State) applyConfig(next workspace.WorktreeConfig) bool {
	next.AllowedSetupRepos = append([]string(nil), next.AllowedSetupRepos...)
	if s.Options.ConfigPath != "" {
		if err := workspace.SaveWorktreeConfig(s.Options.ConfigPath, next); err != nil {
			s.Error = err
			s.lastAction = func() { s.applyConfig(next) }
			return false
		}
	}
	s.Config = next
	s.Error = nil
	return true
}
func (s *State) Create(o workspace.CreateWorktreeOptions, openAfter bool) bool {
	if s.Busy || s.closed {
		return false
	}
	s.pending = &o
	s.openAfter = openAfter
	s.lastAction = func() { s.Create(o, openAfter) }
	return s.start("Creating worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		m.Config.RunSetupAuto = o.RunSetup
		created, e := m.Create(ctx, o)
		r := result{err: e}
		if e == nil {
			r.status = "Created " + created.Worktree.Path
			r.apply = func(s *State) {
				s.pending = nil
				s.dialogOpen = false
				s.Setup = created.Setup
				s.SetupWorktree = created.Worktree.Path
				if openAfter && (created.Setup == nil || created.Setup.Success) && s.Options.OpenWorkspace != nil {
					s.Options.OpenWorkspace(created.Worktree.Path)
				}
			}
		}
		return r
	})
}
func (s *State) Remove(path string, force, confirmed bool) bool {
	if s.Busy || s.closed {
		return false
	}
	if !confirmed {
		s.Error = workspace.ErrConfirmationRequired
		return false
	}
	s.lastAction = func() { s.Remove(path, force, true) }
	active := workspace.SamePath(s.Workspace, path)
	root := s.Root
	return s.start("Removing worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		e := m.Remove(ctx, path, force)
		r := result{err: e, status: "Worktree removed; branch preserved"}
		if e == nil && active {
			m.Dir = root
			r.apply = func(s *State) {
				if s.Options.OpenWorkspace != nil {
					s.Options.OpenWorkspace(root)
				}
			}
		}
		return r
	})
}
func (s *State) Move(path, dest string) bool {
	if s.Busy || s.closed {
		return false
	}
	s.lastAction = func() { s.Move(path, dest) }
	active := workspace.SamePath(s.Workspace, path)
	return s.start("Moving worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		resolved, e := m.ResolveDestination(ctx, dest)
		if e == nil {
			e = m.Move(ctx, path, resolved)
		}
		r := result{err: e, status: "Worktree moved"}
		if e == nil && active {
			m.Dir = resolved
			r.apply = func(s *State) {
				if s.Options.OpenWorkspace != nil {
					s.Options.OpenWorkspace(resolved)
				}
			}
		}
		return r
	})
}
func (s *State) Rename(path, name string) bool {
	if s.Busy || s.closed {
		return false
	}
	s.lastAction = func() { s.Rename(path, name) }
	active := workspace.SamePath(s.Workspace, path)
	return s.start("Renaming worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		dest, e := m.Rename(ctx, path, name)
		r := result{err: e, status: "Worktree renamed"}
		if e == nil && active {
			m.Dir = dest
			r.apply = func(s *State) {
				if s.Options.OpenWorkspace != nil {
					s.Options.OpenWorkspace(dest)
				}
			}
		}
		return r
	})
}
func (s *State) Lock(path, reason string) bool {
	if s.Busy || s.closed {
		return false
	}
	s.lastAction = func() { s.Lock(path, reason) }
	return s.start("Locking worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		return result{err: m.Lock(ctx, path, reason), status: "Worktree locked"}
	})
}
func (s *State) Unlock(path string) bool {
	if s.Busy || s.closed {
		return false
	}
	s.lastAction = func() { s.Unlock(path) }
	return s.start("Unlocking worktree", func(ctx context.Context, m *workspace.WorktreeManager) result {
		return result{err: m.Unlock(ctx, path), status: "Worktree unlocked"}
	})
}
func (s *State) Prune(confirmed bool) bool {
	if s.Busy || s.closed {
		return false
	}
	if !confirmed {
		s.Error = workspace.ErrConfirmationRequired
		return false
	}
	s.lastAction = func() { s.Prune(true) }
	return s.start("Pruning stale worktree records", func(ctx context.Context, m *workspace.WorktreeManager) result {
		return result{err: m.Prune(ctx), status: "Stale worktree records pruned"}
	})
}
func (s *State) Repair() bool {
	if s.Busy || s.closed {
		return false
	}
	s.lastAction = func() { s.Repair() }
	return s.start("Repairing worktree records", func(ctx context.Context, m *workspace.WorktreeManager) result {
		return result{err: m.Repair(ctx), status: "Worktree records repaired"}
	})
}
func (s *State) RetrySetup(confirmed bool) bool {
	if s.Busy || s.closed {
		return false
	}
	if !confirmed {
		s.Error = workspace.ErrConfirmationRequired
		return false
	}
	if s.Setup == nil {
		return false
	}
	s.lastAction = func() { s.RetrySetup(true) }
	script, root, tree, open := s.Setup.Script, s.Root, s.SetupWorktree, s.openAfter
	return s.start("Running setup script", func(ctx context.Context, m *workspace.WorktreeManager) result {
		out := workspace.RunSetupScript(ctx, script, root, tree, true)
		return result{apply: func(s *State) {
			s.Setup = &out
			if out.Success && open && s.Options.OpenWorkspace != nil {
				s.Options.OpenWorkspace(tree)
			}
		}}
	})
}
func (s *State) confirm(title, detail string, f func()) {
	s.dialogOpen = false
	s.confirmLabel = ""
	s.confirmTitle = title
	s.confirmDetail = detail
	s.confirmAction = f
	s.confirmOpen = true
}
func (s *State) createDialog() {
	s.dialog = "create"
	s.dialogOpen = true
	s.name = ""
	s.branch = ""
	s.branchTouched = false
	s.existingBranch = ""
	active := ""
	for _, w := range s.Rows {
		if workspace.SamePath(w.Path, s.Workspace) {
			active = w.Branch
		}
	}
	for _, branch := range s.Branches {
		if s.existingBranch == "" {
			s.existingBranch = branch
		}
		if branch != active {
			s.existingBranch = branch
			break
		}
	}
	s.startPoint = ""
	s.newBranch = true
	s.runSetup = s.Config.RunSetupAuto
	s.openAfter = s.Config.OpenAfterCreate
}
func (s *State) editDialog(kind string, w workspace.Worktree) {
	s.dialog = kind
	s.dialogOpen = true
	s.target = w
	s.force = false
	s.value = ""
	switch kind {
	case "rename":
		s.value = filepath.Base(w.Path)
	case "move":
		s.value = w.Path
	case "lock":
		s.value = w.LockReason
	}
}

func (s *State) deriveBranch() {
	if !s.branchTouched {
		if slug := workspace.WorktreeSlug(s.name); slug != "" {
			s.branch = "orbit/" + slug
		}
	}
}
func (s *State) createOptions() workspace.CreateWorktreeOptions {
	branch := s.branch
	if !s.newBranch {
		branch = s.existingBranch
	}
	return workspace.CreateWorktreeOptions{Name: strings.TrimSpace(s.name), Branch: strings.TrimSpace(branch), NewBranch: s.newBranch, StartPoint: strings.TrimSpace(s.startPoint), RunSetup: s.runSetup}
}
func (s *State) quickCreateDialog() {
	s.createDialog()
	base := ""
	for _, w := range s.Rows {
		if workspace.SamePath(w.Path, s.Workspace) {
			base = workspace.WorktreeSlug(w.Branch)
			break
		}
	}
	if base == "" {
		base = "worktree"
	}
	candidate := base
	for n := 2; ; n++ {
		used := false
		for _, w := range s.Rows {
			used = used || w.Name == candidate
		}
		// Matches source's quick-create collision check without running Git on the UI thread.
		if s.Root != "" {
			if _, err := os.Stat(filepath.Join(s.Root, candidate)); err == nil {
				used = true
			}
		}
		if !used {
			break
		}
		candidate = fmt.Sprintf("%s-%d", base, n)
	}
	s.name = candidate
	s.deriveBranch()
}

func (s *State) allowSetupAndCreate(o workspace.CreateWorktreeOptions) {
	o.ConfirmSetup = true
	next := s.Config
	next.AllowedSetupRepos = append([]string(nil), s.Config.AllowedSetupRepos...)
	next.AllowSetup(s.Root)
	if !s.applyConfig(next) {
		// A retry resumes this exact approved create, not a previous operation.
		s.lastAction = func() { s.allowSetupAndCreate(o) }
		return
	}
	s.Create(o, s.openAfter)
}
