package gitworkbench

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
)

func (s *State) CanCommit() bool {
	if s.Busy || s.Data.Operation != workspace.OperationNone {
		return false
	}
	for _, r := range s.Data.Changes {
		if r.Staged() || (s.IncludeUnstaged && r.Unstaged()) {
			return true
		}
	}
	return false
}
func (s *State) HasGenerator() bool {
	return s.Options.GenerateCommitResult != nil || s.Options.GenerateCommit != nil
}
func (s *State) RequestGenerate(commit, push bool) {
	if s.Busy {
		return
	}
	if !s.HasGenerator() {
		s.Error = errors.New("commit generation is not connected")
		return
	}
	for _, r := range s.Data.Changes {
		if r.Unstaged() && !s.IncludeUnstaged {
			s.openDialog("stage-generate")
			s.pendingCommit = commit
			s.pendingPush = push
			return
		}
	}
	s.generateCommitFlow(false, commit, push)
}
func (s *State) Commit(push bool) {
	if !s.CanCommit() {
		return
	}
	if strings.TrimSpace(s.CommitMessage) == "" {
		s.RequestGenerate(true, push)
		return
	}
	s.commitFlow(s.CommitMessage, s.IncludeUnstaged, push)
}
func (s *State) commitFlow(message string, include, push bool) {
	spec := s.spec()
	s.lastAction = nil
	s.start("Committing", func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		out, e := g.Commit(ctx, message, include)
		committed := e == nil
		if committed && push {
			out, e = g.Push(ctx, false)
			if e != nil {
				out = "Committed locally; push failed. Your commit was not rolled back."
			}
		}
		apply, _, refreshErr := load(ctx, g, h, spec)
		if e == nil {
			e = refreshErr
		}
		return func(s *State) {
			if apply != nil {
				apply(s)
			}
			if committed {
				s.CommitMessage = ""
			}
		}, out, e
	})
}
func (s *State) generateCommitFlow(stageAll, commit, push bool) {
	spec := s.spec()
	include := s.IncludeUnstaged
	generate, typed := s.Options.GenerateCommit, s.Options.GenerateCommitResult
	s.lastAction = nil
	s.start("Generating commit message", func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		if stageAll {
			if e := g.StageAll(ctx); e != nil {
				return nil, "", e
			}
		}
		var message, warning string
		var e error
		if typed != nil {
			var result workspace.GeneratedText
			result, e = typed(ctx, !include)
			message = result.Text
			if result.Fallback {
				warning = "Editable fallback: " + result.Warning
			}
		} else {
			message, e = generate(ctx, !include)
		}
		if e != nil {
			return nil, "", e
		}
		committed := false
		out := "Commit message ready"
		if warning != "" {
			out = warning
		}
		if commit {
			out, e = g.Commit(ctx, message, include)
			committed = e == nil
			if committed && push {
				out, e = g.Push(ctx, false)
				if e != nil {
					out = "Committed locally; push failed. Your commit was not rolled back."
				}
			}
		}
		apply, _, refreshErr := load(ctx, g, h, spec)
		if e == nil {
			e = refreshErr
		}
		return func(s *State) {
			if apply != nil {
				apply(s)
			}
			if committed {
				s.CommitMessage = ""
			} else {
				s.CommitMessage = message
			}
		}, out, e
	})
}
func createdNumber(raw string) int {
	u, e := url.Parse(strings.TrimSpace(raw))
	if e != nil || u.Host == "" {
		return 0
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) == 0 {
		return 0
	}
	n, _ := strconv.Atoi(parts[len(parts)-1])
	return n
}
func (s *State) CreateIssue(options workspace.CreateIssueOptions) {
	spec := s.spec()
	s.lastAction = nil
	s.start("Creating issue", func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		out, e := h.CreateIssue(ctx, options)
		if e != nil {
			return nil, "", e
		}
		number := createdNumber(out)
		spec.issue = number
		apply, _, e := load(ctx, g, h, spec)
		return func(s *State) {
			if apply != nil {
				apply(s)
			}
			s.title = ""
			s.body = ""
			s.notes = ""
			s.labels = ""
		}, fmt.Sprintf("Issue created: %s", out), e
	})
}
func (s *State) CreatePull(options workspace.CreatePullOptions) {
	spec := s.spec()
	s.lastAction = nil
	s.start("Creating pull request", func(ctx context.Context, g workspace.Git, h workspace.GitHub) (func(*State), string, error) {
		out, e := h.CreatePull(ctx, options)
		if e != nil {
			return nil, "", e
		}
		number := createdNumber(out)
		spec.pull = number
		apply, _, e := load(ctx, g, h, spec)
		return func(s *State) {
			if apply != nil {
				apply(s)
			}
			s.title = ""
			s.body = ""
			s.notes = ""
			s.labels = ""
		}, fmt.Sprintf("Pull request created: %s", out), e
	})
}
