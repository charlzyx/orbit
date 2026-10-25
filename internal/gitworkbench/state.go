// Package gitworkbench implements the native Git page over the real workspace
// backend. Subprocess work is isolated from rendering; Tick applies only results
// for the current workspace generation. No remote mutation runs on refresh.
package gitworkbench

import (
	"context"
	"errors"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type Options struct {
	AvatarLoader         func(context.Context, string) ([]byte, error)
	AvatarHome           string
	GenerateCommitResult func(context.Context, bool) (workspace.GeneratedText, error)
	GenerateDraft        func(context.Context, workspace.DraftRequest) (workspace.Draft, error)
	OnChange             func()
	OpenWorktree         func(string)
	PickAttachments      func(context.Context) ([]string, error)
	OpenFile             func(path string)
	RevealFile           func(path string)
	OpenReview           func(source, ref, path string)
	OpenURL              func(url string)
	Authenticate         func()
	WorktreesView        func(*ui.Context)
	GenerateCommit       func(context.Context, bool) (string, error)
	Invalidate           func()
}
type Data struct {
	HasCommits                    bool
	StagedReview, UnstagedReview  workspace.ReviewSnapshot
	IssueTemplates, PullTemplates []workspace.RepositoryTemplate
	GitHubUser                    *workspace.GitHubUser
	Branch                        string
	Changes                       []workspace.StatusRow
	Branches                      []string
	Refs                          []workspace.RefEntry
	RecentBranches                []string
	Stashes                       []workspace.StashEntry
	Operation                     workspace.InProgress
	Conflicts                     []string
	Tracking                      workspace.AheadBehind
	History                       []workspace.CommitEntry
	Graph                         []workspace.GraphRow
	GitHub                        workspace.GitHubStatus
	Issues                        []workspace.GitHubIssue
	Pulls                         []workspace.GitHubPull
	Labels                        []workspace.GitHubLabel
}
type State struct {
	Git                                               workspace.Git
	GitHub                                            workspace.GitHub
	Options                                           Options
	Data                                              Data
	Tab                                               int
	Busy                                              bool
	Activity, Status                                  string
	Error                                             error
	CommitMessage                                     string
	IncludeUnstaged                                   bool
	ChangesFilter                                     string
	DiffWrap, DiffSplit                               bool
	diffExpanded                                      map[string]bool
	Tree                                              bool
	HistoryOptions                                    workspace.HistoryOptions
	GraphLimit                                        int
	GraphAllRefs                                      bool
	IssueFilter                                       workspace.IssueFilter
	PullFilter                                        workspace.PullFilter
	CommitDetail                                      *workspace.CommitDetail
	IssueDetail                                       *workspace.GitHubIssue
	PullDetail                                        *workspace.GitHubPull
	list                                              ui.ListState
	collapsed                                         map[string]bool
	avatarContext                                     context.Context
	avatarCancel                                      context.CancelFunc
	avatarResults                                     chan avatarResult
	avatarOverrides                                   map[string]string
	avatars                                           map[string]*ui.Bitmap
	avatarPending                                     map[string]bool
	results                                           chan result
	cancel                                            context.CancelFunc
	generation                                        uint64
	initialized                                       bool
	refreshed                                         time.Time
	refreshPending                                    bool
	lastAction                                        func()
	pendingCommit, pendingPush                        bool
	confirmOpen                                       bool
	confirmTitle, confirmDetail                       string
	confirmAction                                     func()
	dialog                                            string
	dialogOpen                                        bool
	title, body, labels, assignees, base, branchQuery string
	includeUntracked, draft, deleteBranch             bool
	mode                                              int
	notes, templateOriginal                           string
	templateIndex                                     int
	attachments                                       []string
}
type result struct {
	keepError  bool
	generation uint64
	apply      func(*State)
	status     string
	err        error
}
type loadSpec struct {
	tab         int
	history     workspace.HistoryOptions
	graphLimit  int
	graphAll    bool
	issues      workspace.IssueFilter
	pulls       workspace.PullFilter
	commit      string
	issue, pull int
}

func New(dir string, options Options) *State {
	avatarContext, avatarCancel := context.WithCancel(context.Background())
	return &State{avatarContext: avatarContext, avatarCancel: avatarCancel, avatarResults: make(chan avatarResult, 128), avatarOverrides: workspace.LoadAvatarOverrides(options.AvatarHome), avatars: map[string]*ui.Bitmap{}, avatarPending: map[string]bool{}, Git: workspace.Git{Dir: dir}, GitHub: workspace.GitHub{Dir: dir}, Options: options, HistoryOptions: workspace.HistoryOptions{Limit: 80}, GraphLimit: 80, Tree: true, collapsed: map[string]bool{}, diffExpanded: map[string]bool{}, results: make(chan result, 16)}
}
func (s *State) SetWorkspace(dir string) {
	if dir == s.Git.Dir {
		return
	}
	s.Cancel()
	s.generation++
	s.Git.Dir = dir
	s.GitHub.Dir = dir
	s.Data = Data{}
	s.Error = nil
	s.Status = ""
	s.CommitDetail = nil
	s.IssueDetail = nil
	s.PullDetail = nil
	s.CommitMessage = ""
	s.HistoryOptions = workspace.HistoryOptions{Limit: 80}
	s.GraphLimit, s.GraphAllRefs = 80, false
	s.IssueFilter, s.PullFilter = workspace.IssueFilter{}, workspace.PullFilter{}
	s.ChangesFilter, s.branchQuery = "", ""
	s.dialog, s.dialogOpen, s.confirmOpen = "", false, false
	s.confirmAction = nil
	s.confirmTitle, s.confirmDetail = "", ""
	s.pendingCommit, s.pendingPush, s.refreshPending = false, false, false
	s.title, s.body, s.labels, s.assignees, s.base, s.notes = "", "", "", "", "", ""
	s.templateIndex, s.templateOriginal, s.attachments = -1, "", nil
	s.list = ui.ListState{}
	s.initialized = false
	s.collapsed = map[string]bool{}
	s.diffExpanded = map[string]bool{}
}
func (s *State) Close() { s.Cancel(); s.avatarCancel(); s.generation++ }
func (s *State) Cancel() {
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.generation++
	s.Busy = false
	s.Activity = ""
	s.refreshed = time.Now()
	s.lastAction = nil
}
func (s *State) spec() loadSpec {
	o := loadSpec{s.Tab, s.HistoryOptions, s.GraphLimit, s.GraphAllRefs, s.IssueFilter, s.PullFilter, "", 0, 0}
	if s.CommitDetail != nil {
		o.commit = s.CommitDetail.Hash
	}
	if s.IssueDetail != nil {
		o.issue = s.IssueDetail.Number
	}
	if s.PullDetail != nil {
		o.pull = s.PullDetail.Number
	}
	return o
}
func (s *State) start(label string, job func(context.Context, workspace.Git, workspace.GitHub) (func(*State), string, error)) {
	if s.Busy {
		return
	}
	s.initialized = true
	s.Busy = true
	s.Activity = label
	keepError := label == "Refreshing"
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	g, h, channel, wake := s.Git, s.GitHub, s.results, s.Options.Invalidate
	if wake == nil {
		wake = s.Options.OnChange
	}
	go func() {
		defer cancel()
		apply, status, err := job(ctx, g, h)
		if ctx.Err() != nil {
			return
		}
		select {
		case channel <- result{keepError: keepError, generation: generation, apply: apply, status: status, err: err}:
			if wake != nil {
				wake()
			}
		case <-ctx.Done():
		}
	}()
}

// Tick must be called on the UI thread. It never blocks on I/O.
func (s *State) Tick() {
	s.drainAvatars()
	for {
		select {
		case r := <-s.results:
			if r.generation != s.generation {
				continue
			}
			s.Busy = false
			s.Activity = ""
			s.cancel = nil
			s.refreshed = time.Now()
			if r.apply != nil {
				r.apply(s)
			}
			if r.err != nil || !r.keepError {
				s.Error = r.err
			}
			if r.status != "" {
				s.Status = r.status
			}
			if s.refreshPending {
				s.refreshPending = false
				s.Refresh()
			}
		default:
			return
		}
	}
}
func (s *State) Refresh() {
	if s.Busy {
		return
	}
	spec := s.spec()
	s.start("Refreshing", func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		return load(ctx, g, h, spec)
	})
}
func load(ctx context.Context, g workspace.Git, h workspace.GitHub, o loadSpec) (func(*State), string, error) {
	var d Data
	d.IssueTemplates = workspace.IssueTemplates(g.Dir)
	d.PullTemplates = workspace.PullTemplates(g.Dir)
	var e error
	if d.Changes, e = g.Status(ctx); e != nil {
		return nil, "", e
	}
	if d.HasCommits, e = g.HasCommits(ctx); e != nil {
		return nil, "", e
	}
	if d.Branch, e = g.CurrentBranch(ctx); e != nil {
		return nil, "", e
	}
	if d.Branches, e = g.Branches(ctx); e != nil {
		return nil, "", e
	}
	if d.Refs, e = g.Refs(ctx); e != nil {
		return nil, "", e
	}
	if d.RecentBranches, e = g.RecentBranches(ctx); e != nil {
		return nil, "", e
	}
	if d.Stashes, e = g.Stashes(ctx); e != nil {
		return nil, "", e
	}
	if d.Operation, e = g.Operation(ctx); e != nil {
		return nil, "", e
	}
	if d.Conflicts, e = g.Conflicts(ctx); e != nil {
		return nil, "", e
	}
	if d.Tracking, e = g.AheadBehind(ctx); e != nil {
		return nil, "", e
	}
	var commit *workspace.CommitDetail
	var issue *workspace.GitHubIssue
	var pull *workspace.GitHubPull
	switch o.tab {
	case 0:
		if len(d.Conflicts) == 0 {
			var st, un workspace.ReviewDiff
			st, e = g.CollectReview(ctx, workspace.ReviewRequest{Source: workspace.ReviewStaged})
			if e == nil {
				un, e = g.CollectReview(ctx, workspace.ReviewRequest{Source: workspace.ReviewUnstaged})
			}
			if e == nil {
				d.StagedReview = workspace.ParseReview(st)
				d.UnstagedReview = workspace.ParseReview(un)
			}
		}
	case 1:
		d.History, e = g.History(ctx, o.history)
	case 2:
		var commits []workspace.GraphCommit
		commits, e = g.GraphHistory(ctx, o.graphLimit, o.graphAll)
		d.Graph = workspace.LayoutGraph(commits)
	case 4, 5:
		d.GitHub, e = h.Probe(ctx)
		if e == nil && d.GitHub.Authenticated {
			if user, err := h.CurrentUser(ctx); err == nil {
				d.GitHubUser = &user
			}
			if o.tab == 4 {
				d.Issues, e = h.Issues(ctx, o.issues)
			} else {
				d.Pulls, e = h.Pulls(ctx, o.pulls)
			}
			if e == nil {
				d.Labels, e = h.Labels(ctx)
			}
		}
	}
	if e != nil {
		return nil, "", e
	}
	if o.commit != "" && (o.tab == 1 || o.tab == 2) {
		v, err := g.CommitDetail(ctx, o.commit)
		if err != nil {
			return nil, "", err
		}
		commit = &v
	}
	if o.issue > 0 && o.tab == 4 && d.GitHub.Authenticated {
		v, err := h.Issue(ctx, o.issue)
		if err != nil {
			return nil, "", err
		}
		issue = &v
	}
	if o.pull > 0 && o.tab == 5 && d.GitHub.Authenticated {
		v, err := h.Pull(ctx, o.pull)
		if err != nil {
			return nil, "", err
		}
		pull = &v
	}
	return func(s *State) {
		if o.tab < 4 {
			d.GitHub = s.Data.GitHub
			d.GitHubUser = s.Data.GitHubUser
		}
		s.Data = d
		s.applyLateTemplate()
		if commit != nil {
			s.CommitDetail = commit
		}
		if issue != nil {
			s.IssueDetail = issue
		}
		if pull != nil {
			s.PullDetail = pull
		}
	}, "", nil
}
func (s *State) RunGit(label string, fn func(context.Context, workspace.Git) (string, error)) {
	if s.Busy {
		return
	}
	s.lastAction = func() { s.RunGit(label, fn) }
	spec := s.spec()
	s.start(label, func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		out, e := fn(ctx, g)
		if e != nil {
			if workspace.ClassifyGitFailure(e.Error()).Kind == workspace.FailurePushRejected {
				_ = g.Fetch(ctx)
			}
			apply, _, _ := load(ctx, g, h, spec)
			return apply, "", e
		}
		apply, _, e := load(ctx, g, h, spec)
		if out == "" {
			out = label + " completed"
		}
		return func(s *State) {
			s.lastAction = nil
			if apply != nil {
				apply(s)
			}
		}, out, e
	})
}
func (s *State) RunGitHub(label string, fn func(context.Context, workspace.GitHub) (string, error)) {
	if s.Busy {
		return
	}
	s.lastAction = func() { s.RunGitHub(label, fn) }
	spec := s.spec()
	s.start(label, func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		out, e := fn(ctx, h)
		if e != nil {
			if workspace.ClassifyGitFailure(e.Error()).Kind == workspace.FailurePushRejected {
				_ = g.Fetch(ctx)
			}
			apply, _, _ := load(ctx, g, h, spec)
			return apply, "", e
		}
		apply, _, e := load(ctx, g, h, spec)
		if out == "" {
			out = label + " completed"
		}
		return func(s *State) {
			s.lastAction = nil
			if apply != nil {
				apply(s)
			}
		}, out, e
	})
}
func (s *State) SelectCommit(hash string) {
	s.start("Loading commit", func(ctx context.Context, g workspace.Git, _ workspace.GitHub) (func(*State), string, error) {
		d, e := g.CommitDetail(ctx, hash)
		return func(s *State) {
			if e == nil {
				s.CommitDetail = &d
			}
		}, "", e
	})
}
func (s *State) SelectIssue(number int) {
	s.start("Loading issue", func(ctx context.Context, _ workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		d, e := h.Issue(ctx, number)
		return func(s *State) {
			if e == nil {
				s.IssueDetail = &d
			}
		}, "", e
	})
}
func (s *State) SelectPull(number int) {
	s.start("Loading pull request", func(ctx context.Context, _ workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		d, e := h.Pull(ctx, number)
		return func(s *State) {
			if e == nil {
				s.PullDetail = &d
			}
		}, "", e
	})
}
func (s *State) Confirm(title, detail string, action func()) {
	s.confirmTitle = title
	s.confirmDetail = detail
	s.confirmAction = action
	s.confirmOpen = true
}
func (s *State) openDialog(kind string) {
	s.dialog = kind
	s.dialogOpen = true
	s.title = ""
	s.body = ""
	s.labels = ""
	s.assignees = ""
	s.base = ""
	s.mode = 0
	s.draft = false
	s.includeUntracked = false
	s.deleteBranch = false
	s.attachments = nil
	s.notes = ""
	s.templateIndex = -1
	s.templateOriginal = ""
	if kind == "new-pull" {
		for _, name := range []string{"main", "master"} {
			for _, branch := range s.Data.Branches {
				if branch == name {
					s.base = branch
					break
				}
			}
			if s.base != "" {
				break
			}
		}
		if s.base == "" && len(s.Data.Branches) > 0 {
			s.base = s.Data.Branches[0]
		}
	}
	s.applyLateTemplate()
}
func (s *State) review(source, ref, path string) {
	if s.Options.OpenReview != nil {
		s.Options.OpenReview(source, ref, path)
	} else {
		s.Error = errors.New("Review is not connected to this page")
	}
}
func (s *State) SetTab(tab int) {
	if tab < 0 || tab > 5 || tab == s.Tab {
		return
	}
	s.Tab = tab
	s.list = ui.ListState{}
	if s.Busy {
		s.refreshPending = true
		return
	}
	s.Refresh()
}
