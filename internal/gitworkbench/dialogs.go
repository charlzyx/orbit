package gitworkbench

import (
	"context"
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func modal(c *ui.Context, open *bool, width float32, body func()) {
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(width).MaxWidth(780).MaxHeight(720).Padding(18).Gap(12).Background(c.Theme().Surface).Border(1, c.Theme().Border).Radius(10).Children(body)
	})
}
func (s *State) confirmView(c *ui.Context) {
	if !s.confirmOpen {
		return
	}
	modal(c, &s.confirmOpen, 520, func() {
		ui.Text(c, s.confirmTitle).FontSize(18).Bold()
		ui.Text(c, s.confirmDetail).Selectable()
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if button(c, "Cancel", false).Clicked() {
				s.confirmOpen = false
				s.confirmAction = nil
			}
			if button(c, "Confirm", s.Busy).Clicked() {
				action := s.confirmAction
				s.confirmAction = nil
				s.confirmOpen = false
				if action != nil {
					action()
				}
			}
		})
	})
}
func splitValues(text string) []string {
	var values []string
	seen := map[string]bool{}
	for _, s := range strings.Split(text, ",") {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			values = append(values, s)
			seen[s] = true
		}
	}
	return values
}
func (s *State) dialogView(c *ui.Context) {
	if !s.dialogOpen {
		return
	}
	modal(c, &s.dialogOpen, 620, func() {
		switch s.dialog {
		case "stage-generate":
			ui.Text(c, "Include unstaged changes?").FontSize(18).Bold()
			ui.Text(c, "Generate from the current index, or stage all changes first. Generated text stays editable unless you requested Commit.")
			staged := false
			for _, r := range s.Data.Changes {
				if r.Staged() {
					staged = true
				}
			}
			ui.Row(c).Gap(8).Children(func() {
				if button(c, "Cancel", false).Clicked() {
					s.dialogOpen = false
				}
				if button(c, "Use staged only", s.Busy || !staged).Clicked() {
					s.dialogOpen = false
					s.generateCommitFlow(false, s.pendingCommit, s.pendingPush)
				}
				if button(c, "Stage all and generate", s.Busy).Clicked() {
					s.dialogOpen = false
					s.generateCommitFlow(true, s.pendingCommit, s.pendingPush)
				}
			})

		case "branches":
			s.branchesDialog(c)
		case "rename-branch":
			ui.Text(c, "Rename branch").FontSize(18).Bold()
			muted(c, s.base)
			ui.TextInput(c, &s.title).Label("New branch name").AutoFocus().AlignSelf(ui.Stretch)
			s.dialogActions(c, "Rename", s.title == "", func() {
				from, to := s.base, s.title
				s.RunGit("Rename branch", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.RenameBranch(ctx, from, to) })
			})
		case "merge", "rebase":
			s.mergeDialog(c)
		case "stash":
			ui.Text(c, "Stash changes").FontSize(18).Bold()
			ui.TextInput(c, &s.title).Label("Stash message").Placeholder("Optional message").AutoFocus().AlignSelf(ui.Stretch)
			ui.Checkbox(c, &s.includeUntracked, "Include untracked files")
			s.dialogActions(c, "Stash", false, func() {
				message, untracked := s.title, s.includeUntracked
				s.RunGit("Stash", func(ctx context.Context, g workspace.Git) (string, error) {
					return g.PushStash(ctx, message, untracked)
				})
			})
		case "new-issue", "new-pull":
			s.createDialog(c)
		case "issue-comment", "pull-comment":
			ui.Text(c, "Add comment").FontSize(18).Bold()
			ui.TextArea(c, &s.body).Label("Comment").Height(220).AlignSelf(ui.Stretch).AutoFocus()
			s.dialogActions(c, "Post comment", strings.TrimSpace(s.body) == "", func() {
				body := s.body
				if s.dialog == "issue-comment" && s.IssueDetail != nil {
					number := s.IssueDetail.Number
					s.RunGitHub("Comment on issue", func(ctx context.Context, h workspace.GitHub) (string, error) {
						return h.CommentIssue(ctx, number, body)
					})
				} else if s.PullDetail != nil {
					number := s.PullDetail.Number
					s.RunGitHub("Comment on pull request", func(ctx context.Context, h workspace.GitHub) (string, error) { return h.CommentPull(ctx, number, body) })
				}
			})
		case "issue-labels":
			ui.Text(c, "Edit issue labels").FontSize(18).Bold()
			ui.TextInput(c, &s.labels).Label("Labels to add").Placeholder("Add labels, comma separated").AlignSelf(ui.Stretch)
			ui.TextInput(c, &s.title).Label("Labels to remove").Placeholder("Remove labels, comma separated").AlignSelf(ui.Stretch)
			s.labelSuggestions(c)
			s.dialogActions(c, "Apply labels", s.labels == "" && s.title == "", func() {
				if s.IssueDetail == nil {
					return
				}
				number, add, remove := s.IssueDetail.Number, splitValues(s.labels), splitValues(s.title)
				s.RunGitHub("Edit issue labels", func(ctx context.Context, h workspace.GitHub) (string, error) {
					return h.EditIssueLabels(ctx, number, add, remove)
				})
			})
		case "close-issue":
			ui.Text(c, "Close issue?").FontSize(18).Bold()
			if s.IssueDetail != nil {
				ui.Text(c, s.IssueDetail.Title)
			}
			selectIndex(c, &s.mode, []string{"Completed", "Not planned"}).Label("Close reason")
			s.dialogActions(c, "Close issue", false, func() {
				if s.IssueDetail == nil {
					return
				}
				number := s.IssueDetail.Number
				reason := []string{"completed", "not planned"}[s.mode]
				s.RunGitHub("Close issue", func(ctx context.Context, h workspace.GitHub) (string, error) {
					return h.CloseIssue(ctx, number, reason)
				})
			})
		case "pull-review":
			ui.Text(c, "Review pull request").FontSize(18).Bold()
			selectIndex(c, &s.mode, []string{"Approve", "Request changes", "Comment"}).Label("Review action")
			ui.TextArea(c, &s.body).Label("Review body").Placeholder("Review feedback").Height(200).AlignSelf(ui.Stretch)
			s.dialogActions(c, "Submit review", s.mode != 0 && strings.TrimSpace(s.body) == "", func() {
				if s.PullDetail == nil {
					return
				}
				number, body := s.PullDetail.Number, s.body
				action := []workspace.PullReviewAction{workspace.ReviewApprove, workspace.ReviewRequestChanges, workspace.ReviewComment}[s.mode]
				s.RunGitHub("Submit review", func(ctx context.Context, h workspace.GitHub) (string, error) {
					return h.ReviewPull(ctx, number, action, body)
				})
			})
		case "merge-pull":
			ui.Text(c, "Merge pull request?").FontSize(18).Bold()
			if s.PullDetail != nil {
				ui.Text(c, s.PullDetail.Title)
				muted(c, s.PullDetail.HeadRef+" → "+s.PullDetail.BaseRef)
			}
			selectIndex(c, &s.mode, []string{"Merge commit", "Squash and merge", "Rebase and merge"}).Label("Merge method")
			ui.Checkbox(c, &s.deleteBranch, "Delete head branch after merging")
			ui.Text(c, "This changes the remote repository.").TextColor(c.Theme().Warning)
			s.dialogActions(c, "Merge pull request", false, func() {
				if s.PullDetail == nil {
					return
				}
				number, remove := s.PullDetail.Number, s.deleteBranch
				method := []workspace.PullMergeMethod{workspace.PullMerge, workspace.PullSquash, workspace.PullRebase}[s.mode]
				s.RunGitHub("Merge pull request", func(ctx context.Context, h workspace.GitHub) (string, error) {
					return h.MergePull(ctx, number, method, remove)
				})
			})
		}
	})
}
func (s *State) dialogActions(c *ui.Context, label string, disabled bool, submit func()) {
	ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
		if button(c, "Cancel", false).Clicked() {
			s.dialogOpen = false
		}
		if button(c, label, s.Busy || disabled).Clicked() {
			if s.dialog != "new-issue" && s.dialog != "new-pull" {
				s.dialogOpen = false
			}
			submit()
		}
	})
}
func (s *State) branchesDialog(c *ui.Context) {
	ui.Text(c, "Switch branch").FontSize(18).Bold()
	ui.TextInput(c, &s.branchQuery).Placeholder("Filter branches, remotes and tags").Label("Branch filter").AutoFocus().AlignSelf(ui.Stretch)
	var refs []workspace.RefEntry
	seen := map[string]bool{}
	for _, name := range s.Data.RecentBranches {
		refs = append(refs, workspace.RefEntry{Name: name, Kind: workspace.RefBranch})
		seen[name] = true
	}
	for _, ref := range s.Data.Refs {
		if ref.Kind != workspace.RefBranch || !seen[ref.Name] {
			refs = append(refs, ref)
		}
	}
	filtered := refs[:0]
	for _, ref := range refs {
		if s.branchQuery == "" || strings.Contains(strings.ToLower(ref.Name), strings.ToLower(s.branchQuery)) {
			filtered = append(filtered, ref)
		}
	}
	ui.List(c, nil, len(filtered), func(i int) {
		ref := filtered[i]
		ui.Row(c.Key(string(ref.Kind) + ref.Name)).Gap(6).Children(func() {
			label := ref.Name
			if ref.Name == s.Data.Branch && ref.Kind == workspace.RefBranch {
				label += " ✓"
			}
			b := button(c, label, s.Busy).Grow(1)
			if b.Clicked() {
				s.dialogOpen = false
				s.RunGit("Checkout "+ref.Name, func(ctx context.Context, g workspace.Git) (string, error) {
					if ref.Kind == workspace.RefRemote {
						return "", g.CheckoutRemote(ctx, ref.Name)
					}
					return "", g.Checkout(ctx, ref.Name)
				})
			}
			muted(c, string(ref.Kind))
			if ref.Kind == workspace.RefBranch {
				button(c, "…", s.Busy).Menu(func(m *ui.Menu) {
					if m.Item("Rename…").Chosen() {
						s.openDialog("rename-branch")
						s.base = ref.Name
						s.title = ref.Name
					}
					if m.Item("Delete…").Chosen() {
						s.dialogOpen = false
						s.Confirm("Delete branch "+ref.Name+"?", "Git will refuse to delete an unmerged branch. The current branch cannot be deleted.", func() {
							s.RunGit("Delete branch", func(ctx context.Context, g workspace.Git) (string, error) {
								return "", g.DeleteBranch(ctx, ref.Name, false)
							})
						})
					}
					if m.Item("Force delete…").Chosen() {
						s.dialogOpen = false
						s.Confirm("Force delete branch "+ref.Name+"?", "Unmerged commits can become unreachable. This does not delete a remote branch.", func() {
							s.RunGit("Force delete branch", func(ctx context.Context, g workspace.Git) (string, error) {
								return "", g.DeleteBranch(ctx, ref.Name, true)
							})
						})
					}
				})
			}
		})
	}).Height(340)
	separator(c)
	ui.Row(c).Gap(8).Children(func() {
		ui.TextInput(c, &s.title).Label("New branch name").Placeholder("New branch name").Grow(1)
		if button(c, "Create branch", s.Busy || strings.TrimSpace(s.title) == "").Clicked() {
			name := s.title
			s.dialogOpen = false
			s.RunGit("Create branch", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.CreateBranch(ctx, name) })
		}
	})
}
func (s *State) mergeDialog(c *ui.Context) {
	label := "Merge branch"
	if s.dialog == "rebase" {
		label = "Rebase onto branch"
	}
	ui.Text(c, label).FontSize(18).Bold()
	ui.Row(c).Gap(6).Children(func() {
		ui.TextInput(c, &s.base).Label("Target reference").Placeholder("Branch, remote branch, tag or commit").Grow(1)
		button(c, "Choose reference", false).Menu(func(m *ui.Menu) {
			for _, ref := range s.Data.Refs {
				if m.Item(ref.Name).Chosen() {
					s.base = ref.Name
				}
			}
		})
	})
	if s.dialog == "merge" {
		selectIndex(c, &s.mode, []string{"Regular merge", "Always create merge commit", "Squash into index"}).Label("Merge mode")
	} else {
		ui.Checkbox(c, &s.includeUntracked, "Autostash uncommitted changes")
		ui.Text(c, "Rebase rewrites local commit history and may stop on conflicts.").TextColor(c.Theme().Warning)
	}
	s.dialogActions(c, label, strings.TrimSpace(s.base) == "", func() {
		ref := s.base
		if s.dialog == "rebase" {
			autostash := s.includeUntracked
			s.RunGit("Rebase", func(ctx context.Context, g workspace.Git) (string, error) { return g.Rebase(ctx, ref, autostash) })
		} else {
			mode := []workspace.MergeMode{workspace.MergeRegular, workspace.MergeNoFF, workspace.MergeSquash}[s.mode]
			s.RunGit("Merge", func(ctx context.Context, g workspace.Git) (string, error) { return g.Merge(ctx, ref, mode) })
		}
	})
}
func (s *State) labelSuggestions(c *ui.Context) {
	if len(s.Data.Labels) == 0 {
		return
	}
	button(c, "Repository labels", false).Menu(func(m *ui.Menu) {
		for _, label := range s.Data.Labels {
			if m.Item(label.Name).Chosen() {
				if s.labels != "" {
					s.labels += ", "
				}
				s.labels += label.Name
			}
		}
	})
}
func (s *State) createDialog(c *ui.Context) {
	pull := s.dialog == "new-pull"
	title := "New issue"
	if pull {
		title = "New pull request"
	}
	ui.Text(c, title).FontSize(18).Bold()
	if s.Error != nil {
		ui.Text(c, s.Error.Error()).Selectable()
	}
	s.templatePicker(c, pull)
	ui.TextInput(c, &s.title).Label("Title").Placeholder("Title").AutoFocus().AlignSelf(ui.Stretch)
	ui.TextArea(c, &s.body).Label("Description").Placeholder("Markdown description").Height(200).AlignSelf(ui.Stretch)
	ui.TextInput(c, &s.labels).Label("Labels").Placeholder("Labels, comma separated").AlignSelf(ui.Stretch)
	s.labelSuggestions(c)
	ui.TextArea(c, &s.notes).Label("Draft generation notes").Placeholder("Notes for draft generation (optional)").Height(64).AlignSelf(ui.Stretch)
	if s.Options.GenerateDraft != nil && button(c, "Generate draft", s.Busy).Clicked() {
		s.generateDraft(pull)
	}
	if pull {
		ui.TextInput(c, &s.base).Label("Base branch").Placeholder("Base branch (repository default if empty)").AlignSelf(ui.Stretch)
		ui.Checkbox(c, &s.draft, "Create as draft")
	} else {
		ui.TextInput(c, &s.assignees).Label("Assignees").Placeholder("Assignees, comma separated").AlignSelf(ui.Stretch)
		if s.Options.PickAttachments != nil && button(c, "Add attachments…", s.Busy).Clicked() {
			pick := s.Options.PickAttachments
			s.start("Choosing attachments", func(ctx context.Context, _ workspace.Git, _ workspace.GitHub) (func(*State), string, error) {
				paths, e := pick(ctx)
				return func(s *State) {
					if e == nil {
						s.attachments = append(s.attachments, paths...)
					}
				}, "", e
			})
		}
		for i, path := range s.attachments {
			ui.Row(c.Key(i)).Gap(6).Children(func() {
				muted(c, path).Grow(1)
				if button(c, "Remove", false).Clicked() {
					s.attachments = append(s.attachments[:i], s.attachments[i+1:]...)
				}
			})
		}
	}
	muted(c, fmt.Sprintf("This will publish to the current repository on GitHub."))
	s.dialogActions(c, "Create", strings.TrimSpace(s.title) == "", func() {
		if pull {
			o := workspace.CreatePullOptions{Title: strings.TrimSpace(s.title), Body: s.body, Base: s.base, Draft: s.draft, Labels: splitValues(s.labels)}
			s.CreatePull(o)
		} else {
			o := workspace.CreateIssueOptions{Title: strings.TrimSpace(s.title), Body: s.body, Labels: splitValues(s.labels), Assignees: splitValues(s.assignees), Attachments: append([]string(nil), s.attachments...)}
			s.CreateIssue(o)
		}
	})
}

func selectIndex(c *ui.Context, index *int, options []string) ui.Element {
	if *index < 0 || *index >= len(options) {
		*index = 0
	}
	value := options[*index]
	el := ui.Select(c, &value, options)
	if el.Changed() {
		*index = stateIndex(value, options)
	}
	return el
}
