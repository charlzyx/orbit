package gitworkbench

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func button(c *ui.Context, label string, disabled bool) ui.Element {
	return ui.Button(c, label).Disabled(disabled)
}
func muted(c *ui.Context, text string) ui.Element {
	return ui.Text(c, text).FontSize(12).TextColor(c.Theme().TextMuted)
}
func separator(c *ui.Context) { ui.Box(c).Height(1).AlignSelf(ui.Stretch).Background(c.Theme().Border) }
func View(c *ui.Context, s *State) ui.Element {
	s.Tick()
	if !s.initialized {
		s.Refresh()
	}
	if !s.Busy && s.initialized && time.Since(s.refreshed) > 4*time.Second && s.Tab < 4 {
		s.Refresh()
	}
	root := ui.Column(c.Key("git-workbench")).Fill().MinWidth(0).Background(c.Theme().Background).Gap(8)
	root.Children(func() {
		tab := s.Tab
		if ui.Tabs(c, &tab, "Changes", "History", "Graph", "Stashes", "Issues", "Pull requests").Changed() {
			s.SetTab(tab)
		}
		ui.Row(c).Height(42).PaddingX(14).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "Git").FontSize(18).Bold()
			if s.Data.GitHubUser != nil {
				s.avatar(c, s.Data.GitHubUser.Login, "")
				muted(c, "@"+s.Data.GitHubUser.Login)
			}
			branch := s.Data.Branch
			if branch == "" {
				branch = "Detached HEAD"
			}
			if button(c, branch+" ▾", s.Busy).Clicked() {
				s.openDialog("branches")
			}
			if s.Data.Tracking.HasUpstream {
				muted(c, fmt.Sprintf("↑%d ↓%d", s.Data.Tracking.Ahead, s.Data.Tracking.Behind))
			}
			ui.Spacer(c)
			if s.Busy {
				ui.Spinner(c).Size(14, 14)
				muted(c, s.Activity)
				if button(c, "Cancel", false).Clicked() {
					s.Cancel()
					s.Status = "Cancelled. Refresh to inspect any changes already applied."
				}
			} else {
				if button(c, "Refresh", false).Clicked() {
					s.Refresh()
				}
				if button(c, "Fetch", false).Clicked() {
					s.RunGit("Fetch", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.Fetch(ctx) })
				}
				if button(c, "Pull", false).Clicked() {
					s.RunGit("Pull", func(ctx context.Context, g workspace.Git) (string, error) { return g.Pull(ctx) })
				}
				if button(c, "Push", false).Clicked() {
					s.RunGit("Push", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, false) })
				}
				if button(c, "Sync", false).Clicked() {
					s.RunGit("Sync", func(ctx context.Context, g workspace.Git) (string, error) { return g.Sync(ctx) })
				}
			}
			button(c, "More", s.Busy).Menu(func(m *ui.Menu) {
				if m.Item("Merge branch…").Chosen() {
					s.openDialog("merge")
				}
				if m.Item("Rebase onto…").Chosen() {
					s.openDialog("rebase")
				}
				if m.Item("Stash changes…").Chosen() {
					s.openDialog("stash")
				}
				if m.Item("Force push with lease…").Chosen() {
					s.Confirm("Force push with lease?", "This can replace commits on the remote branch. The push will fail if the remote changed since your last fetch.", func() {
						s.RunGit("Force push with lease", func(ctx context.Context, g workspace.Git) (string, error) { return g.Push(ctx, true) })
					})
				}
			})
		})
		separator(c)
		s.failure(c)
		s.operation(c)
		if s.Status != "" {
			ui.Row(c).PaddingX(14).Gap(8).Children(func() {
				muted(c, s.Status).MaxLines(3).Grow(1)
				if button(c, "Dismiss", false).Clicked() {
					s.Status = ""
				}
			})
		}
		switch s.Tab {
		case 0:
			s.changes(c)
		case 1:
			s.history(c, false)
		case 2:
			s.history(c, true)
		case 3:
			s.stashes(c)
		case 4:
			s.issues(c)
		case 5:
			s.pulls(c)
		}
	})
	if !s.dialogOpen && !s.confirmOpen {
		for i := 0; i < 6; i++ {
			if root.Shortcut(ui.Cmd, ui.Key1+ui.Key(i)) {
				s.SetTab(i)
			}
		}
		if (s.IssueDetail != nil || s.PullDetail != nil) && root.Shortcut(0, ui.KeyEscape) {
			s.CloseTransient()
		}
	}
	s.dialogView(c)
	s.confirmView(c)
	return root
}
func (s *State) operation(c *ui.Context) {
	op := s.Data.Operation
	if op == workspace.OperationNone {
		return
	}
	ui.Row(c).MarginX(14).Padding(10).Gap(8).Background(c.Theme().Warning.Alpha(.12)).Children(func() {
		ui.Text(c, fmt.Sprintf("%s in progress · %d conflicts", op, len(s.Data.Conflicts))).Grow(1)
		if op != workspace.OperationBisect && button(c, "Continue", s.Busy || len(s.Data.Conflicts) > 0).Clicked() {
			s.RunGit("Continue "+string(op), func(ctx context.Context, g workspace.Git) (string, error) { return g.Continue(ctx, op) })
		}
		if op.CanSkip() && button(c, "Skip…", s.Busy).Clicked() {
			s.Confirm("Skip current patch?", "The current patch will not be applied.", func() {
				s.RunGit("Skip "+string(op), func(ctx context.Context, g workspace.Git) (string, error) { return g.Skip(ctx, op) })
			})
		}
		if button(c, "Abort…", s.Busy).Clicked() {
			s.Confirm("Abort "+string(op)+"?", "Git will restore the pre-operation state. Conflict-resolution edits may be discarded.", func() {
				s.RunGit("Abort "+string(op), func(ctx context.Context, g workspace.Git) (string, error) { return g.Abort(ctx, op) })
			})
		}
	})
	if len(s.Data.Conflicts) > 0 {
		ui.Row(c).MarginX(14).Gap(6).Children(func() {
			for _, path := range s.Data.Conflicts[:min(8, len(s.Data.Conflicts))] {
				if button(c, path, s.Options.OpenFile == nil).Clicked() {
					s.Options.OpenFile(filepath.Join(s.Git.Dir, path))
				}
			}
		})
	}
}

type changeItem struct {
	line           *workspace.ReviewLine
	split          *workspace.ReviewSplitRow
	diffIndex      int
	header, path   string
	row            workspace.StatusRow
	staged, folder bool
	count          int
}

func (s *State) changeItems() []changeItem {
	var result []changeItem
	for _, staged := range []bool{true, false} {
		title := "Unstaged changes"
		if staged {
			title = "Staged changes"
		}
		var rows []workspace.StatusRow
		for _, r := range s.Data.Changes {
			if ((staged && r.Staged()) || (!staged && r.Unstaged())) && (s.ChangesFilter == "" || strings.Contains(strings.ToLower(r.Path), strings.ToLower(s.ChangesFilter))) {
				rows = append(rows, r)
			}
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
		result = append(result, changeItem{header: title, staged: staged, count: len(rows)})
		lastDir := ""
		for _, r := range rows {
			dir := filepath.ToSlash(filepath.Dir(r.Path))
			if s.Tree && dir != "." {
				key := title + ":" + dir
				if dir != lastDir {
					result = append(result, changeItem{path: key, header: dir, folder: true, staged: staged})
					lastDir = dir
				}
				if s.collapsed[key] {
					continue
				}
			}
			result = append(result, changeItem{path: r.Path, row: r, staged: staged})
			if s.diffExpanded[fmt.Sprintf("%t:%s", staged, r.Path)] {
				result = append(result, s.fileDiff(staged, r.Path)...)
			}
		}
	}
	return result
}
func (s *State) changes(c *ui.Context) {
	ui.Row(c).PaddingX(14).Gap(8).Children(func() {
		ui.TextInput(c, &s.ChangesFilter).Placeholder("Filter changed files…").Label("Filter changed files").Grow(1)
		ui.Checkbox(c, &s.Tree, "Tree")
		ui.Checkbox(c, &s.DiffWrap, "Wrap")
		ui.Checkbox(c, &s.DiffSplit, "Split")
		button(c, "Expand", false).Menu(func(m *ui.Menu) {
			if m.Item("Expand all files").Chosen() {
				for _, r := range s.Data.Changes {
					s.diffExpanded["true:"+r.Path] = true
					s.diffExpanded["false:"+r.Path] = true
				}
			}
			if m.Item("Collapse all files").Chosen() {
				s.diffExpanded = map[string]bool{}
			}
		})
		if button(c, "Stage all", s.Busy || len(s.Data.Changes) == 0).Clicked() {
			s.RunGit("Stage all", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.StageAll(ctx) })
		}
		if button(c, "Unstage all", s.Busy).Clicked() {
			s.RunGit("Unstage all", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.UnstageAll(ctx) })
		}
		if button(c, "Discard unstaged…", s.Busy || len(s.Data.Changes) == 0).Clicked() {
			s.Confirm("Discard all unstaged changes?", "This restores tracked files to the index and permanently removes untracked files. Staged changes are kept.", func() {
				s.RunGit("Discard unstaged", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.DiscardAll(ctx) })
			})
		}
	})
	items := s.changeItems()
	s.list.Key = func(i int) any {
		r := items[i]
		return fmt.Sprintf("%t:%s:%s:%d", r.staged, r.header, r.path, r.diffIndex)
	}
	ui.List(c.Key("changes-list"), &s.list, len(items), func(i int) {
		item := items[i]
		if item.line != nil || item.split != nil {
			s.embeddedDiffLine(c, item)
			return
		}
		if item.folder {
			if button(c, "▸ "+item.header, false).Clicked() {
				s.collapsed[item.path] = !s.collapsed[item.path]
			}
			return
		}
		if item.header != "" {
			ui.Row(c).Padding(8, 14).Children(func() {
				ui.Text(c, fmt.Sprintf("%s (%d)", item.header, item.count)).Bold()
				ui.Spacer(c)
				if item.count > 0 && button(c, "Review", false).Clicked() {
					source := "unstaged"
					if item.staged {
						source = "staged"
					}
					s.review(source, "", "")
				}
			})
			if item.count == 0 {
				muted(c, "No changes").Padding(8, 14)
			}
			return
		}
		r := item.row
		ui.Row(c).Padding(4, 14).Gap(8).AlignItems(ui.Center).Children(func() {
			code := r.Worktree
			if item.staged {
				code = r.Index
			}
			color := c.Theme().TextMuted
			if r.Conflict() {
				color = c.Theme().Danger
			}
			ui.Text(c, string(code)).Width(16).Font("monospace").TextColor(color)
			label := r.Path
			if s.Tree {
				label = filepath.Base(label)
			}
			b := ui.ButtonBase(c).Grow(1).Label(r.Path).Padding(5).Children(func() { ui.Text(c, label).SingleLine() })
			if b.Clicked() {
				s.diffExpanded[fmt.Sprintf("%t:%s", item.staged, r.Path)] = !s.diffExpanded[fmt.Sprintf("%t:%s", item.staged, r.Path)]
			}
			b.ContextMenu(func(m *ui.Menu) {
				if m.Item("Open in Review").Chosen() {
					source := "unstaged"
					if item.staged {
						source = "staged"
					}
					s.review(source, "", r.Path)
				}
				if m.Item("Copy relative path").Chosen() {
					c.WriteClipboard(r.Path)
				}
				if m.Item("Copy absolute path").Chosen() {
					c.WriteClipboard(filepath.Join(s.Git.Dir, r.Path))
				}
				if s.Options.OpenFile != nil && m.Item("Open file").Chosen() {
					s.Options.OpenFile(filepath.Join(s.Git.Dir, r.Path))
				}
				if s.Options.RevealFile != nil && m.Item("Reveal in file manager").Chosen() {
					s.Options.RevealFile(filepath.Join(s.Git.Dir, r.Path))
				}
				if m.Item("File history").Chosen() {
					s.HistoryOptions.Path = r.Path
					s.SetTab(1)
				}
			})
			adds, dels := r.UnstagedAdditions, r.UnstagedDeletions
			if item.staged {
				adds, dels = r.StagedAdditions, r.StagedDeletions
			}
			ui.Text(c, fmt.Sprintf("+%d", adds)).FontSize(12).TextColor(c.Theme().Success)
			ui.Text(c, fmt.Sprintf("−%d", dels)).FontSize(12).TextColor(c.Theme().Danger)
			if item.staged {
				if button(c, "Unstage", s.Busy).Clicked() {
					s.RunGit("Unstage", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.Unstage(ctx, r.Path) })
				}
			} else {
				if button(c, "Stage", s.Busy).Clicked() {
					paths := []string{r.Path}
					if r.OrigPath != "" {
						paths = append(paths, r.OrigPath)
					}
					s.RunGit("Stage", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.Stage(ctx, paths...) })
				}
				if button(c, "Discard…", s.Busy).Clicked() {
					s.Confirm("Discard "+r.Path+"?", "Unstaged edits will be lost. Untracked files are permanently removed.", func() {
						s.RunGit("Discard", func(ctx context.Context, g workspace.Git) (string, error) { return "", g.Discard(ctx, r.Path) })
					})
				}
			}
		})
	}).Grow(1).MinHeight(0)
	separator(c)
	ui.Column(c).Padding(10, 14).Gap(8).Children(func() {
		if s.commitBarMode() != "commit" {
			s.cleanCommitBar(c)
			return
		}
		field := ui.TextArea(c, &s.CommitMessage).Placeholder("Commit message (leave blank to generate)…").Label("Commit message").Height(72).AlignSelf(ui.Stretch)
		if !s.dialogOpen && !s.confirmOpen && field.Shortcut(ui.Cmd, ui.KeyEnter) {
			s.Commit(false)
		}
		ui.Row(c).Gap(8).Children(func() {
			ui.Checkbox(c, &s.IncludeUnstaged, "Include unstaged changes")
			ui.Spacer(c)
			if s.HasGenerator() && button(c, "Generate message", s.Busy || len(s.Data.Changes) == 0).Clicked() {
				s.RequestGenerate(false, false)
			}
			if button(c, "Commit", !s.CanCommit()).Clicked() {
				s.Commit(false)
			}
			if button(c, "Commit & push", !s.CanCommit()).Clicked() {
				s.Commit(true)
			}
		})
	})
}
func (s *State) history(c *ui.Context, graph bool) {
	ui.Row(c).PaddingX(14).Gap(6).Children(func() {
		if graph {
			if ui.Checkbox(c, &s.GraphAllRefs, "All refs").Changed() {
				s.Refresh()
			}
			muted(c, "Local branches"+map[bool]string{true: " + remote branches and tags"}[s.GraphAllRefs])
		} else {
			ui.TextInput(c, &s.HistoryOptions.Grep).Placeholder("Search commit messages").Label("Commit search").Grow(1)
			ui.TextInput(c, &s.HistoryOptions.Author).Placeholder("Author").Label("Author filter").Width(140)
			ui.TextInput(c, &s.HistoryOptions.Path).Placeholder("File path").Label("History path filter").Width(160)
			ui.Checkbox(c, &s.HistoryOptions.AllBranches, "All branches")
			ui.Checkbox(c, &s.HistoryOptions.Follow, "Follow renames")
			if button(c, "Clear filters", s.Busy).Clicked() {
				s.HistoryOptions = workspace.HistoryOptions{Limit: 80}
				s.Refresh()
			}
			if button(c, "Apply", s.Busy).Clicked() {
				s.Refresh()
			}
		}
	})
	ui.Row(c).Grow(1).MinHeight(0).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			count := len(s.Data.History)
			if graph {
				count = len(s.Data.Graph)
			}
			s.list.Key = func(i int) any {
				if graph {
					return s.Data.Graph[i].Commit.Hash
				}
				return s.Data.History[i].Hash
			}
			ui.List(c.Key("commits"), &s.list, count, func(i int) {
				var entry workspace.CommitEntry
				if graph {
					entry = s.Data.Graph[i].Commit.CommitEntry
				} else {
					entry = s.Data.History[i]
				}
				ui.Column(c).AlignSelf(ui.Stretch).Children(func() {
					b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Height(56).Padding(4, 14).Gap(8).Label(entry.Subject)
					if s.CommitDetail != nil && s.CommitDetail.Hash == entry.Hash {
						b.Background(c.Theme().Selection)
					}
					b.Children(func() {
						if graph {
							drawGraph(c, s.Data.Graph[i])
						}
						s.avatar(c, entry.Author, entry.AuthorEmail)
						ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
							ui.Text(c, entry.Subject).SingleLine()
							author := ui.ButtonBase(c).Label("Filter author " + entry.Author).Children(func() { muted(c, entry.Author+" · "+entry.Relative).SingleLine() })
							if author.Clicked() {
								s.HistoryOptions.Author = entry.Author
								s.Tab = 1
								s.Refresh()
							}
						})
						for _, r := range entry.Refs {
							ui.Text(c, r.Name).FontSize(11).Padding(2, 4).Radius(3).Background(c.Theme().Accent.Alpha(.14))
						}
						muted(c, entry.Short).Font("monospace")
					})
					if b.Clicked() {
						if s.CommitDetail != nil && s.CommitDetail.Hash == entry.Hash {
							s.CommitDetail = nil
						} else {
							s.SelectCommit(entry.Hash)
						}
					}
					b.ContextMenu(func(m *ui.Menu) {
						if m.Item("Copy commit hash").Chosen() {
							c.WriteClipboard(entry.Hash)
						}
						if m.Item("Review commit").Chosen() {
							s.review("committed", entry.Hash, "")
						}
						if s.Options.OpenURL != nil && m.Item("Open commit on remote").Chosen() {
							open := s.Options.OpenURL
							s.start("Resolving remote", func(ctx context.Context, g workspace.Git, _ workspace.GitHub) (func(*State), string, error) {
								web, e := g.RemoteWeb(ctx)
								return func(*State) {
									if e == nil {
										open(web.CommitURL(entry.Hash))
									}
								}, "", e
							})
						}
					})
					if s.CommitDetail != nil && s.CommitDetail.Hash == entry.Hash {
						s.commitView(c)
					}
				})
			}).Grow(1).MinHeight(0)
			if count == 0 && !s.Busy {
				muted(c, "No commits match these filters.").Padding(20)
			}
			if button(c, "Load more", s.Busy).Clicked() {
				if graph {
					s.GraphLimit += 40
				} else {
					s.HistoryOptions.Limit += 40
				}
				s.Refresh()
			}
		})
	})
}
func drawGraph(c *ui.Context, row workspace.GraphRow) {
	palette := []ui.Color{c.Theme().Accent, c.Theme().Success, c.Theme().Warning, ui.RGB(179, 131, 227), ui.RGB(60, 190, 190)}
	ui.Box(c).Size(float32(row.LaneCount*14+10), 56).Draw(func(p *ui.Painter, r ui.Rect) {
		x := func(lane int) float32 { return r.X + 8 + float32(lane)*14 }
		mid := r.Y + r.H/2
		for _, lane := range row.Verticals {
			p.Line(x(lane), r.Y, x(lane), r.Y+r.H, 1.5, palette[lane%len(palette)])
		}
		color := palette[row.NodeLane%len(palette)]
		if row.NodeHasIncoming {
			p.Line(x(row.NodeLane), r.Y, x(row.NodeLane), mid, 1.5, color)
		}
		for _, lane := range row.Parents {
			p.Line(x(row.NodeLane), mid, x(lane), r.Y+r.H, 1.5, palette[lane%len(palette)])
		}
		p.Fill(ui.Rect{X: x(row.NodeLane) - 3.5, Y: mid - 3.5, W: 7, H: 7}, color, 3.5)
	})
}
func (s *State) commitView(c *ui.Context) {
	d := s.CommitDetail
	ui.Column(c).AlignSelf(ui.Stretch).Padding(14).Gap(10).Border(1, c.Theme().Border).Children(func() {
		ui.Row(c).Children(func() {
			ui.Text(c, "Commit").Bold()
			ui.Spacer(c)
			if button(c, "Close", false).Clicked() {
				s.CommitDetail = nil
			}
		})
		ui.Text(c, d.Subject).FontSize(17).Bold()
		ui.Text(c, d.Hash).Font("monospace").FontSize(11).Selectable()
		s.avatar(c, d.Author, d.AuthorEmail)
		muted(c, d.Author+" <"+d.AuthorEmail+">")
		muted(c, d.AuthorDate)
		muted(c, "Committed by "+d.Committer+" · "+d.CommitterDate)
		if d.Body != "" {
			ui.Text(c, d.Body).Selectable()
		}
		ui.Row(c).Gap(6).Children(func() {
			if button(c, "Copy hash", false).Clicked() {
				c.WriteClipboard(d.Hash)
			}
			if button(c, "Review", false).Clicked() {
				s.review("committed", d.Hash, "")
			}
		})
		for _, parent := range d.Parents {
			if button(c, "Parent "+parent[:min(8, len(parent))], s.Busy).Clicked() {
				s.SelectCommit(parent)
			}
		}
		separator(c)
		for _, f := range d.Files[:min(300, len(d.Files))] {
			ui.Row(c.Key(f.Path)).Gap(6).Children(func() {
				b := button(c, f.Status+" "+f.Path, false).Grow(1)
				if b.Clicked() {
					s.review("committed", d.Hash, f.Path)
				}
				b.ContextMenu(func(m *ui.Menu) {
					if m.Item("Review this commit file").Chosen() {
						s.review("committed", d.Hash, f.Path)
					}
					if m.Item("Review uncommitted changes").Chosen() {
						s.review("uncommitted", "", f.Path)
					}
					if s.Options.OpenFile != nil && m.Item("Open current file").Chosen() {
						s.Options.OpenFile(filepath.Join(s.Git.Dir, f.Path))
					}
					if s.Options.RevealFile != nil && m.Item("Reveal in file manager").Chosen() {
						s.Options.RevealFile(filepath.Join(s.Git.Dir, f.Path))
					}
					if m.Item("Copy relative path").Chosen() {
						c.WriteClipboard(f.Path)
					}
					if m.Item("Copy absolute path").Chosen() {
						c.WriteClipboard(filepath.Join(s.Git.Dir, f.Path))
					}
					if m.Item("File history (follow renames)").Chosen() {
						s.HistoryOptions.Path, s.HistoryOptions.Follow = f.Path, true
						s.HistoryOptions.Limit, s.Tab, s.CommitDetail = 80, 1, nil
						s.Refresh()
					}
				})
				muted(c, fmt.Sprintf("+%d −%d", f.Additions, f.Deletions))
			})
		}
		if len(d.Files) > 300 {
			muted(c, fmt.Sprintf("%d more files", len(d.Files)-300))
		}
	})
}
