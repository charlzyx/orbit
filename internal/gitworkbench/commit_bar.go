package gitworkbench

import (
	"context"
	"fmt"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (s *State) commitBarMode() string {
	if len(s.Data.Changes) > 0 {
		return "commit"
	}
	tracking := s.Data.Tracking
	if !tracking.HasUpstream && s.Data.HasCommits {
		return "publish"
	}
	if tracking.Ahead > 0 && tracking.Behind > 0 {
		return "merge"
	}
	if tracking.Ahead > 0 {
		return "push"
	}
	if tracking.Behind > 0 {
		return "pull"
	}
	return "up-to-date"
}
func (s *State) cleanCommitBar(c *ui.Context) {
	ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
		switch s.commitBarMode() {
		case "publish":
			muted(c, "This branch has no upstream")
			if button(c, "Publish branch", s.Busy).Clicked() {
				s.RunGit("Publish branch", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, false) })
			}
		case "push":
			muted(c, fmt.Sprintf("%d commits ahead", s.Data.Tracking.Ahead))
			if button(c, "Push", s.Busy).Clicked() {
				s.RunGit("Push", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, false) })
			}
		case "pull":
			muted(c, fmt.Sprintf("%d commits behind", s.Data.Tracking.Behind))
			if button(c, "Pull", s.Busy).Clicked() {
				s.RunGit("Pull", func(ctx context.Context, g workspace.Git) (string, error) { return g.Pull(ctx) })
			}
		case "merge":
			muted(c, "Local and upstream branches have diverged")
			if button(c, "Merge upstream", s.Busy).Clicked() {
				s.RunGit("Merge upstream", func(ctx context.Context, g workspace.Git) (string, error) { return g.MergeUpstream(ctx) })
			}
		default:
			muted(c, "Up to date")
		}
	})
}
