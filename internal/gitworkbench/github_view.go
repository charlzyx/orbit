package gitworkbench

import (
	"context"
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func (s *State) githubReady(c *ui.Context) bool {
	if s.Busy && !s.Data.GitHub.Installed {
		muted(c, "Checking GitHub CLI…").Padding(20)
		return false
	}
	if !s.Data.GitHub.Installed || !s.Data.GitHub.Authenticated {
		ui.Column(c).Padding(30).Gap(12).Children(func() {
			title := "Install GitHub CLI"
			detail := "Orbit uses gh for GitHub issues and pull requests. Credentials stay with gh."
			if s.Data.GitHub.Installed {
				title = "Sign in to GitHub"
				detail = "Run gh auth login in a terminal, then refresh this page."
			}
			ui.Text(c, title).FontSize(20).Bold()
			ui.Text(c, detail)
			if s.Data.GitHub.Detail != "" {
				muted(c, s.Data.GitHub.Detail).Selectable()
			}
			ui.Row(c).Gap(8).Children(func() {
				if s.Options.OpenURL != nil && button(c, "GitHub CLI setup", false).Clicked() {
					s.Options.OpenURL("https://cli.github.com/")
				}
				if s.Options.Authenticate != nil && button(c, "Authenticate", false).Clicked() {
					s.Options.Authenticate()
				}
				if button(c, "Refresh", s.Busy).Clicked() {
					s.Refresh()
				}
			})
		})
		return false
	}
	if !s.Data.GitHub.Supported {
		ui.Text(c, s.Data.GitHub.Detail).TextColor(c.Theme().Warning).PaddingX(14)
	}
	return true
}
func stateIndex(value string, options []string) int {
	for i, s := range options {
		if value == s {
			return i
		}
	}
	return 0
}
func (s *State) issues(c *ui.Context) {
	if !s.githubReady(c) {
		return
	}
	if s.IssueDetail != nil {
		s.issueView(c)
		return
	}
	ui.Row(c).PaddingX(14).Gap(6).Children(func() {
		states := []string{"open", "closed", "all"}
		index := stateIndex(s.IssueFilter.State, states)
		if selectIndex(c, &index, states).Width(100).Label("Issue state").Changed() {
			s.IssueFilter.State = states[index]
			s.Refresh()
		}
		ui.TextInput(c, &s.IssueFilter.Search).Placeholder("Search issues").Label("Issue search").Grow(1)
		ui.TextInput(c, &s.IssueFilter.Label).Placeholder("Label").Label("Issue label filter").Width(110)
		ui.TextInput(c, &s.IssueFilter.Assignee).Placeholder("Assignee").Label("Issue assignee filter").Width(110)
		if button(c, "Apply", s.Busy).Clicked() {
			s.Refresh()
		}
		if button(c, "Clear filters", s.Busy).Clicked() {
			s.IssueFilter = workspace.IssueFilter{}
			s.Refresh()
		}
		if button(c, "New issue", s.Busy).Clicked() {
			s.openDialog("new-issue")
		}
	})
	ui.Row(c).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			s.list.Key = func(i int) any { return s.Data.Issues[i].Number }
			ui.List(c.Key("issues-list"), &s.list, len(s.Data.Issues), func(i int) {
				issue := s.Data.Issues[i]
				b := ui.ButtonBase(c).Height(70).AlignSelf(ui.Stretch).Padding(8, 14).Label(issue.Title).Children(func() {
					if issue.Author != nil {
						s.avatar(c, issue.Author.Login, "")
					}
					ui.Column(c).Grow(1).Gap(4).Children(func() {
						ui.Text(c, issue.Title).SingleLine()
						author := ""
						if issue.Author != nil {
							author = issue.Author.Login
						}
						muted(c, fmt.Sprintf("#%d · %s · %s", issue.Number, author, issue.State))
						ui.Row(c).Gap(4).Children(func() {
							for _, label := range issue.Labels {
								muted(c, label.Name).PaddingX(4).Background(c.Theme().Accent.Alpha(.12)).Radius(3)
							}
						})
					})
					muted(c, fmt.Sprintf("%d comments", len(issue.Comments)))
				})
				if b.Clicked() {
					s.SelectIssue(issue.Number)
				}
				b.ContextMenu(func(m *ui.Menu) {
					if m.Item("Copy URL").Chosen() {
						c.WriteClipboard(issue.URL)
					}
					if s.Options.OpenURL != nil && m.Item("Open on GitHub").Chosen() {
						s.Options.OpenURL(issue.URL)
					}
				})
			}).Grow(1).MinHeight(0)
			if len(s.Data.Issues) == 0 && !s.Busy {
				muted(c, "No issues match these filters.").Padding(20)
			}
		})
	})
}
func (s *State) issueView(c *ui.Context) {
	d := s.IssueDetail
	ui.Scroll(c).Grow(1).MinHeight(0).AlignSelf(ui.Stretch).Padding(14).Gap(10).Border(1, c.Theme().Border).Children(func() {
		ui.Row(c).Children(func() {
			muted(c, fmt.Sprintf("Issue #%d · %s", d.Number, d.State))
			ui.Spacer(c)
			if button(c, "All issues", false).Clicked() {
				s.IssueDetail = nil
			}
		})
		ui.Text(c, d.Title).FontSize(18).Bold()
		if d.Author != nil {
			s.authorMeta(c, d.Author.Login, d.CreatedAt)
		}
		s.markdown(c, d.Body)
		ui.Row(c).Gap(6).Children(func() {
			if s.Options.OpenURL != nil && button(c, "Open on GitHub", false).Clicked() {
				s.Options.OpenURL(d.URL)
			}
			if button(c, "Copy URL", false).Clicked() {
				c.WriteClipboard(d.URL)
			}
		})
		ui.Text(c, "Labels").Bold()
		ui.Row(c).Gap(4).Children(func() {
			for _, l := range d.Labels {
				button(c, l.Name, s.Busy).Menu(func(m *ui.Menu) {
					if m.Item("Remove label").Chosen() {
						s.RunGitHub("Remove label", func(ctx context.Context, h workspace.GitHub) (string, error) {
							return h.EditIssueLabels(ctx, d.Number, nil, []string{l.Name})
						})
					}
				})
			}
			if button(c, "Edit labels…", s.Busy).Clicked() {
				s.openDialog("issue-labels")
			}
		})
		if len(d.Assignees) > 0 {
			names := []string{}
			for _, u := range d.Assignees {
				names = append(names, u.Login)
			}
			muted(c, "Assigned to "+strings.Join(names, ", "))
		}
		ui.Row(c).Gap(6).Children(func() {
			if button(c, "Comment…", s.Busy).Clicked() {
				s.openDialog("issue-comment")
			}
			if d.IsOpen() {
				if button(c, "Close issue…", s.Busy).Clicked() {
					s.openDialog("close-issue")
				}
			} else {
				if button(c, "Reopen issue", s.Busy).Clicked() {
					s.RunGitHub("Reopen issue", func(ctx context.Context, h workspace.GitHub) (string, error) { return h.ReopenIssue(ctx, d.Number) })
				}
			}
		})
		separator(c)
		for i, comment := range d.Comments {
			ui.Column(c.Key(i)).Gap(4).Children(func() {
				author := "Deleted user"
				if comment.Author != nil {
					author = comment.Author.Login
				}
				s.authorMeta(c, author, comment.CreatedAt)
				s.markdown(c, comment.Body)
				separator(c)
			})
		}
	})
}
func (s *State) pulls(c *ui.Context) {
	if !s.githubReady(c) {
		return
	}
	if s.PullDetail != nil {
		s.pullView(c)
		return
	}
	ui.Row(c).PaddingX(14).Gap(6).Children(func() {
		states := []string{"open", "closed", "merged", "all"}
		index := stateIndex(s.PullFilter.State, states)
		if selectIndex(c, &index, states).Width(100).Label("Pull request state").Changed() {
			s.PullFilter.State = states[index]
			s.Refresh()
		}
		ui.TextInput(c, &s.PullFilter.Search).Placeholder("Search pull requests").Label("Pull request search").Grow(1)
		ui.TextInput(c, &s.PullFilter.Base).Placeholder("Base branch").Label("Pull request base filter").Width(120)
		ui.TextInput(c, &s.PullFilter.Author).Placeholder("Author").Label("Pull request author filter").Width(100)
		if button(c, "Apply", s.Busy).Clicked() {
			s.Refresh()
		}
		if button(c, "Clear filters", s.Busy).Clicked() {
			s.PullFilter = workspace.PullFilter{}
			s.Refresh()
		}
		if button(c, "New pull request", s.Busy).Clicked() {
			s.openDialog("new-pull")
		}
	})
	ui.Row(c).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			s.list.Key = func(i int) any { return s.Data.Pulls[i].Number }
			ui.List(c.Key("pull-list"), &s.list, len(s.Data.Pulls), func(i int) {
				pr := s.Data.Pulls[i]
				b := ui.ButtonBase(c).Height(72).AlignSelf(ui.Stretch).Padding(8, 14).Gap(8).Label(pr.Title).Children(func() {
					if pr.Author != nil {
						s.avatar(c, pr.Author.Login, "")
					}
					ui.Column(c).Grow(1).Gap(4).Children(func() {
						ui.Text(c, pr.Title).SingleLine()
						author := ""
						if pr.Author != nil {
							author = pr.Author.Login
						}
						state := pr.State
						if pr.IsDraft {
							state = "DRAFT"
						}
						muted(c, fmt.Sprintf("#%d · %s · %s", pr.Number, author, state))
						muted(c, pr.HeadRef+" → "+pr.BaseRef)
					})
					if bucket, has := pr.ChecksSummary(); has {
						ui.Text(c, string(bucket)).FontSize(12).TextColor(checkColor(c, bucket))
					}
					muted(c, fmt.Sprintf("+%d −%d", pr.Additions, pr.Deletions))
				})
				if b.Clicked() {
					s.SelectPull(pr.Number)
				}
				b.ContextMenu(func(m *ui.Menu) {
					if m.Item("Copy URL").Chosen() {
						c.WriteClipboard(pr.URL)
					}
					if s.Options.OpenURL != nil && m.Item("Open on GitHub").Chosen() {
						s.Options.OpenURL(pr.URL)
					}
				})
			}).Grow(1).MinHeight(0)
			if len(s.Data.Pulls) == 0 && !s.Busy {
				muted(c, "No pull requests match these filters.").Padding(20)
			}
		})
	})
}
func checkColor(c *ui.Context, b workspace.CheckBucket) ui.Color {
	switch b {
	case workspace.CheckFailed:
		return c.Theme().Danger
	case workspace.CheckPassed:
		return c.Theme().Success
	case workspace.CheckPending:
		return c.Theme().Warning
	}
	return c.Theme().TextMuted
}
func (s *State) pullView(c *ui.Context) {
	d := s.PullDetail
	ui.Scroll(c).Grow(1).MinHeight(0).AlignSelf(ui.Stretch).Padding(14).Gap(10).Border(1, c.Theme().Border).Children(func() {
		ui.Row(c).Children(func() {
			muted(c, fmt.Sprintf("Pull request #%d · %s", d.Number, d.State))
			ui.Spacer(c)
			if button(c, "All pull requests", false).Clicked() {
				s.PullDetail = nil
			}
		})
		ui.Text(c, d.Title).FontSize(18).Bold()
		muted(c, d.HeadRef+" → "+d.BaseRef)
		if d.IsDraft {
			ui.Text(c, "Draft").TextColor(c.Theme().Warning)
		}
		s.markdown(c, d.Body)
		muted(c, fmt.Sprintf("%d files · +%d −%d", d.ChangedFiles, d.Additions, d.Deletions))
		muted(c, "Merge status: "+d.Mergeable+" / "+d.MergeStateStatus)
		if d.ReviewDecision != "" {
			muted(c, "Review: "+d.ReviewDecision)
		}
		ui.Row(c).Gap(6).Children(func() {
			if s.Options.OpenURL != nil && button(c, "Open on GitHub", false).Clicked() {
				s.Options.OpenURL(d.URL)
			}
			if button(c, "Copy URL", false).Clicked() {
				c.WriteClipboard(d.URL)
			}
			if button(c, "Checkout…", s.Busy).Clicked() {
				s.Confirm("Check out pull request?", "Switch this workspace to "+d.HeadRef+". The checkout is refused if the workspace is dirty or another Git operation is in progress.", func() {
					s.RunGitHub("Checkout pull request", func(ctx context.Context, h workspace.GitHub) (string, error) { return h.CheckoutPull(ctx, d.Number) })
				})
			}
		})
		ui.Row(c).Gap(6).Children(func() {
			if button(c, "Comment…", s.Busy).Clicked() {
				s.openDialog("pull-comment")
			}
			if d.IsOpen() {
				if button(c, "Review…", s.Busy).Clicked() {
					s.openDialog("pull-review")
				}
				if button(c, "Merge…", s.Busy || d.IsDraft).Clicked() {
					s.openDialog("merge-pull")
				}
				if button(c, "Close…", s.Busy).Clicked() {
					s.Confirm("Close pull request?", d.Title, func() {
						s.RunGitHub("Close pull request", func(ctx context.Context, h workspace.GitHub) (string, error) { return h.ClosePull(ctx, d.Number) })
					})
				}
			} else if !d.IsMerged() {
				if button(c, "Reopen", s.Busy).Clicked() {
					s.RunGitHub("Reopen pull request", func(ctx context.Context, h workspace.GitHub) (string, error) { return h.ReopenPull(ctx, d.Number) })
				}
			}
		})
		separator(c)
		ui.Text(c, "Checks").Bold()
		if len(d.Checks) == 0 {
			muted(c, "No checks reported")
		}
		for i, check := range d.Checks {
			ui.Row(c.Key(i)).Gap(6).Children(func() {
				ui.Text(c, string(check.BucketKind())).FontSize(11).TextColor(checkColor(c, check.BucketKind()))
				muted(c, check.Label()).Grow(1)
				if check.URL() != "" && s.Options.OpenURL != nil && button(c, "Details", false).Clicked() {
					s.Options.OpenURL(check.URL())
				}
			})
		}
		separator(c)
		ui.Text(c, "Files").Bold()
		for _, file := range d.Files[:min(300, len(d.Files))] {
			ui.Row(c.Key(file.Path)).Gap(4).Children(func() {
				muted(c, file.Path).Grow(1)
				muted(c, fmt.Sprintf("+%d −%d", file.Additions, file.Deletions))
			})
		}
		if len(d.Files) > 300 {
			muted(c, fmt.Sprintf("%d more files", len(d.Files)-300))
		}
		ui.Text(c, "Commits").Bold()
		for _, commit := range d.Commits {
			muted(c, commit.OID[:min(8, len(commit.OID))]+" "+commit.Headline)
		}
		separator(c)
		for i, review := range d.Reviews {
			ui.Column(c.Key(fmt.Sprintf("review:%d", i))).Gap(4).Children(func() {
				author := "Deleted user"
				if review.Author != nil {
					author = review.Author.Login
				}
				s.authorMeta(c, author, review.State+" · "+review.SubmittedAt)
				s.markdown(c, review.Body)
				separator(c)
			})
		}
		for i, comment := range d.Comments {
			ui.Column(c.Key(fmt.Sprintf("comment:%d", i))).Gap(4).Children(func() {
				author := "Deleted user"
				if comment.Author != nil {
					author = comment.Author.Login
				}
				s.authorMeta(c, author, comment.CreatedAt)
				s.markdown(c, comment.Body)
				separator(c)
			})
		}
	})
}
func (s *State) markdown(c *ui.Context, body string) {
	if body == "" {
		return
	}
	markdown.View(c, body, &markdown.Options{Theme: markdown.DefaultTheme(c.Theme()), OnLink: func(url string) {
		if s.Options.OpenURL != nil {
			s.Options.OpenURL(url)
		}
	}})
}

func (s *State) authorMeta(c *ui.Context, author, suffix string) {
	ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
		s.avatar(c, author, "")
		muted(c, author+" · "+suffix)
	})
}
