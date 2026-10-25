package gitworkbench

import (
	"context"
	"path/filepath"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (s *State) failure(c *ui.Context) {
	if s.Error == nil {
		return
	}
	failure := workspace.ClassifyGitFailure(s.Error.Error())
	titles := map[workspace.FailureKind]string{workspace.FailurePushRejected: "Push rejected", workspace.FailureDiverged: "Branches have diverged", workspace.FailureMergeConflict: "Merge conflicts need resolution", workspace.FailureRebaseConflict: "Rebase conflicts need resolution", workspace.FailureSecretScan: "Remote secret scanning blocked the push", workspace.FailureLargeFile: "A file exceeds the remote size limit", workspace.FailureProtected: "The branch is protected", workspace.FailureAuth: "Git authentication failed", workspace.FailureNetwork: "The remote could not be reached"}
	title := titles[failure.Kind]
	if title == "" {
		title = "Git action failed"
	}
	ui.Column(c).MarginX(14).Padding(10).Gap(6).Border(1, c.Theme().Danger).Radius(6).Children(func() {
		ui.Row(c).Children(func() {
			ui.Text(c, title).TextColor(c.Theme().Danger).Bold()
			ui.Spacer(c)
			if button(c, "Dismiss", false).Clicked() {
				s.Error = nil
			}
		})
		ui.Scroll(c).MaxHeight(180).AlignSelf(ui.Stretch).Children(func() { ui.Text(c, failure.Detail).Selectable() })
		ui.Row(c).Gap(6).Children(func() {
			if button(c, "Copy details", false).Clicked() {
				c.WriteClipboard(failure.Detail)
			}
			for _, action := range failure.RecoveryActions() {
				switch action {
				case workspace.RecoveryPull:
					if button(c, "Pull", s.Busy).Clicked() {
						s.RunGit("Pull", func(ctx context.Context, g workspace.Git) (string, error) { return g.Pull(ctx) })
					}
				case workspace.RecoveryMerge:
					if button(c, "Merge upstream", s.Busy).Clicked() {
						s.RunGit("Merge upstream", func(ctx context.Context, g workspace.Git) (string, error) { return g.MergeUpstream(ctx) })
					}
				case workspace.RecoveryRebase:
					if button(c, "Rebase upstream", s.Busy).Clicked() {
						s.RunGit("Rebase upstream", func(ctx context.Context, g workspace.Git) (string, error) { return g.RebaseUpstream(ctx) })
					}
				case workspace.RecoveryForceWithLease:
					if button(c, "Force with lease…", s.Busy).Clicked() {
						s.Confirm("Force push with lease?", "This can replace remote commits. Git refuses if the remote changed since the last fetch.", func() {
							s.RunGit("Force push with lease", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, true) })
						})
					}
				case workspace.RecoveryPublish:
					if button(c, "Publish branch", s.Busy).Clicked() {
						s.RunGit("Publish branch", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, false) })
					}
				case workspace.RecoveryRetry:
					if s.lastAction != nil && button(c, "Retry", s.Busy).Clicked() {
						s.lastAction()
					}
				case workspace.RecoveryAuthenticate:
					if button(c, "Copy authentication command", false).Clicked() {
						c.WriteClipboard("gh auth login")
						s.Status = "Copied gh auth login. Run it in a terminal, then refresh."
					}
				case workspace.RecoveryOpenConflicts:
					if button(c, "Open conflicts", s.Options.OpenFile == nil || len(s.Data.Conflicts) == 0).Clicked() {
						s.Options.OpenFile(filepath.Join(s.Git.Dir, s.Data.Conflicts[0]))
					}
				case workspace.RecoveryContinue:
					op := s.Data.Operation
					if op != workspace.OperationNone && button(c, "Continue", s.Busy || len(s.Data.Conflicts) > 0).Clicked() {
						s.RunGit("Continue", func(ctx context.Context, g workspace.Git) (string, error) { return g.Continue(ctx, op) })
					}
				case workspace.RecoverySkip:
					op := s.Data.Operation
					if op.CanSkip() && button(c, "Skip…", s.Busy).Clicked() {
						s.Confirm("Skip current patch?", "The current patch will not be applied.", func() {
							s.RunGit("Skip", func(ctx context.Context, g workspace.Git) (string, error) { return g.Skip(ctx, op) })
						})
					}
				case workspace.RecoveryAbort:
					op := s.Data.Operation
					if op != workspace.OperationNone && button(c, "Abort…", s.Busy).Clicked() {
						s.Confirm("Abort operation?", "Uncommitted conflict-resolution work may be discarded.", func() {
							s.RunGit("Abort", func(ctx context.Context, g workspace.Git) (string, error) { return g.Abort(ctx, op) })
						})
					}
				}
			}
		})
	})
}

// CloseTransient lets the host route Escape before leaving the Git page.
func (s *State) CloseTransient() bool {
	if s.basePickerOpen {
		s.closeBasePicker()
		return true
	}
	if s.confirmOpen {
		s.confirmOpen = false
		s.confirmAction = nil
		return true
	}
	if s.dialogOpen {
		s.dialogOpen = false
		return true
	}
	if s.IssueDetail != nil {
		s.IssueDetail = nil
		return true
	}
	if s.PullDetail != nil {
		s.PullDetail = nil
		return true
	}
	return false
}
