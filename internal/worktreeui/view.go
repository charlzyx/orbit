package worktreeui

import (
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func button(c *ui.Context, label string, disabled bool) ui.Element {
	return ui.Button(c, label).Disabled(disabled)
}
func View(c *ui.Context, s *State) ui.Element {
	s.Tick()
	if !s.initialized && !s.Busy && !s.closed {
		s.Refresh()
	}
	t := c.Theme()
	return ui.Column(c).FillWidth().FillHeight().MinHeight(0).Background(t.Background).Children(func() {
		ui.Row(c).Height(44).PaddingX(16).Gap(8).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			ui.Text(c, s.text("worktree.page.title", "Worktrees")).FontWeight(600).Grow(1)
			if s.Busy {
				ui.Spinner(c).Size(16, 16)
				ui.Text(c, s.Activity).TextColor(t.TextMuted)
				if button(c, s.text("settings.cancel", "Cancel"), false).Clicked() {
					s.Cancel()
				}
			}
			if button(c, s.text("common.refresh", "Refresh"), s.Busy).Clicked() {
				s.Refresh()
			}
			if button(c, s.text("worktree.work_in.new", "New worktree"), s.Busy || s.Root == "").Clicked() {
				s.createDialog()
			}
			ui.MenuButton(c, s.text("worktree.page.advanced", "Advanced"), func(m *ui.Menu) {
				if m.Item(s.text("worktree.advanced.repair", "Repair worktree records")).Disabled(s.Busy).Chosen() {
					s.Repair()
				}
				if m.Item(s.text("worktree.advanced.prune", "Prune stale records…")).Disabled(s.Busy).Chosen() {
					s.confirm("Prune stale worktree records?", "Remove Git administrative records for worktree directories that no longer exist. Locked worktrees remain protected.", func() { s.Prune(true) })
				}
				if m.Item(s.text("worktree.settings.title", "Worktree settings…")).Chosen() {
					s.configDraft = s.Config
					s.configDraft.AllowedSetupRepos = append([]string(nil), s.Config.AllowedSetupRepos...)
					s.dialog = "settings"
					s.dialogOpen = true
				}
			})
		})
		ui.Scroll(c).Grow(1).MinHeight(0).Padding(24, 32).Children(func() {
			ui.Column(c).MaxWidth(720).FillWidth().AlignSelf(ui.Center).Gap(14).Children(func() {
				if s.ShowHowTo {
					ui.Column(c).Padding(16).Gap(8).Background(t.Surface).Border(1, t.Border).Radius(8).Children(func() {
						ui.Row(c).Children(func() {
							ui.Text(c, s.text("worktree.howto.title", "How worktrees work")).FontWeight(600).Grow(1)
							if button(c, s.text("common.dismiss", "Dismiss"), false).Clicked() {
								s.ShowHowTo = false
								if s.Options.OnDismissHowTo != nil {
									s.Options.OnDismissHowTo()
								}
							}
						})
						ui.Text(c, s.text("worktree.howto.intro", "Work on several branches in separate folders while sharing one Git repository. Each folder opens as an Orbit workspace with its own files, agent, terminal and review.")).TextColor(t.TextMuted)
					})
				}
				if s.Error != nil {
					ui.Column(c).Gap(8).Padding(12).Border(1, t.Danger).Radius(6).Children(func() {
						ui.Text(c, s.Error.Error()).TextColor(t.Danger).Selectable()
						if button(c, s.text("git_panel.recover_retry", "Retry"), s.Busy).Clicked() {
							if s.lastAction != nil {
								s.lastAction()
							} else {
								s.Refresh()
							}
						}
					})
				}
				if s.Status != "" {
					ui.Text(c, s.Status).TextColor(t.Success).Selectable()
				}
				if s.Root != "" {
					ui.Text(c, s.Root).TextColor(t.TextMuted).Selectable()
				}
				ui.TextInput(c, &s.Filter).Placeholder("Filter worktrees and branches…").Label("Filter worktrees").AlignSelf(ui.Stretch)
				if len(s.Rows) == 0 && !s.Busy {
					ui.Text(c, "Open a Git repository to manage worktrees.").TextColor(t.TextMuted)
				}
				for _, w := range s.Rows {
					if s.Filter != "" && !strings.Contains(strings.ToLower(w.Name+" "+w.Path+" "+w.Branch), strings.ToLower(s.Filter)) {
						continue
					}
					row(c, s, w)
				}
				if s.Setup != nil {
					setup(c, s)
				}
			})
		})
		dialogs(c, s)
	})
}
func row(c *ui.Context, s *State, w workspace.Worktree) {
	t := c.Theme()
	ui.Column(c.Key(w.Path)).Padding(14).Gap(8).Background(t.Surface).Border(1, t.Border).Radius(8).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, w.Name).FontWeight(600)
			if w.Main {
				ui.Text(c, "Main").TextColor(t.TextMuted)
			}
			if workspace.SamePath(w.Path, s.Workspace) {
				ui.Text(c, s.text("worktree.row.active", "Active")).TextColor(t.Accent)
			}
			if w.Locked {
				ui.Text(c, "Locked").TextColor(t.Warning).Tooltip(w.LockReason)
			}
			if w.Prunable {
				ui.Text(c, "Prunable").TextColor(t.Warning).Tooltip(w.PruneReason)
			}
			if s.Dirty[w.Path] {
				ui.Text(c, "Modified").TextColor(t.Warning)
			}
			ui.Box(c).Grow(1)
			if button(c, s.text("worktree.menu.open", "Open"), s.Options.OpenWorkspace == nil || s.Busy || w.Bare).Clicked() {
				s.Options.OpenWorkspace(w.Path)
			}
			ui.MenuButton(c, "Actions", func(m *ui.Menu) {
				if m.Item("Open terminal").Disabled(s.Options.OpenTerminal == nil || w.Bare).Chosen() {
					s.Options.OpenTerminal(w.Path)
				}
				if m.Item(s.text("explorer.reveal", "Reveal in file manager")).Disabled(s.Options.RevealPath == nil).Chosen() {
					s.Options.RevealPath(w.Path)
				}
				if m.Item(s.text("explorer.copy_path", "Copy path")).Chosen() {
					c.Services().WriteClipboard(w.Path)
				}
				m.Separator()
				if m.Item(s.text("worktree.menu.rename", "Rename…")).Disabled(s.Busy || w.Main || w.Locked).Chosen() {
					s.editDialog("rename", w)
				}
				if m.Item(s.text("worktree.menu.move", "Move…")).Disabled(s.Busy || w.Main || w.Locked).Chosen() {
					s.editDialog("move", w)
				}
				if w.Locked {
					if m.Item(s.text("worktree.menu.unlock", "Unlock")).Disabled(s.Busy || w.Main).Chosen() {
						s.Unlock(w.Path)
					}
				} else {
					if m.Item(s.text("worktree.menu.lock", "Lock…")).Disabled(s.Busy || w.Main).Chosen() {
						s.editDialog("lock", w)
					}
				}
				m.Separator()
				if m.Item(s.text("worktree.menu.remove", "Remove…")).Disabled(s.Busy || w.Main || w.Locked).Chosen() {
					s.editDialog("remove", w)
				}
			})
		})
		branch := w.Branch
		if w.Detached {
			branch = "Detached HEAD · " + w.Head
		}
		if branch == "" {
			branch = w.Head
		}
		ui.Text(c, branch).TextColor(t.Accent)
		ui.Text(c, w.Path).FontSize(t.Rem(.85)).TextColor(t.TextMuted).Selectable()
	})
}
func modal(c *ui.Context, open *bool, fn func()) {
	t := c.Theme()
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(560).MaxWidth(760).MaxHeight(680).Padding(20).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(fn)
	})
}
func dialogs(c *ui.Context, s *State) {
	if s.confirmOpen {
		modal(c, &s.confirmOpen, func() {
			ui.Text(c, s.confirmTitle).FontSize(18).FontWeight(600)
			ui.Text(c, s.confirmDetail).Selectable()
			ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
				if button(c, s.text("settings.cancel", "Cancel"), false).Clicked() {
					s.confirmOpen = false
				}
				label := s.confirmLabel
				if label == "" {
					label = s.text("git_panel.recover_continue", "Continue")
				}
				if button(c, label, s.Busy).Clicked() {
					f := s.confirmAction
					s.confirmOpen = false
					if f != nil {
						f()
					}
				}
			})
		})
	}
	if !s.dialogOpen {
		return
	}
	modal(c, &s.dialogOpen, func() {
		submit := false
		switch s.dialog {
		case "create":
			ui.Text(c, s.text("worktree.dialog.create_title", "Create worktree")).FontSize(18).FontWeight(600)
			nameInput := ui.TextInput(c, &s.name).Label(s.text("worktree.field.name", "Directory name")).Placeholder(s.text("worktree.field.name", "Directory name")).AutoFocus().AlignSelf(ui.Stretch)
			if nameInput.Changed() {
				s.deriveBranch()
			}
			submit = nameInput.Submitted()
			ui.Checkbox(c, &s.newBranch, s.text("worktree.field.branch_new", "Create a new branch"))
			if s.newBranch {
				branchInput := ui.TextInput(c, &s.branch).Label(s.text("worktree.field.branch", "Branch name")).Placeholder(s.text("worktree.field.branch", "Branch name")).AlignSelf(ui.Stretch)
				if branchInput.Changed() {
					s.branchTouched = true
				}
				submit = submit || branchInput.Submitted()
				startInput := ui.TextInput(c, &s.startPoint).Label("Start point").Placeholder(s.text("worktree.field.start_placeholder", "Start from HEAD (or a branch / commit)")).AlignSelf(ui.Stretch)
				submit = submit || startInput.Submitted()
			} else {
				label := s.existingBranch
				if label == "" {
					label = s.text("worktree.field.branch_existing", "Choose existing branch")
				}
				ui.MenuButton(c, label, func(m *ui.Menu) {
					for _, branch := range s.Branches {
						used := false
						for _, w := range s.Rows {
							used = used || w.Branch == branch
						}
						title := branch
						if used {
							title += " · " + s.text("worktree.row.active", "Active")
						}
						if m.Item(title).Chosen() {
							s.existingBranch = branch
						}
					}
				}).AlignSelf(ui.Stretch)
			}
			ui.Checkbox(c, &s.runSetup, s.text("worktree.field.run_setup", "Run repository setup script"))
			ui.Checkbox(c, &s.openAfter, s.text("worktree.settings.open_after_create", "Open after creation and successful setup"))
		case "branch_in_use":
			ui.Text(c, s.text("worktree.branch_in_use.title", "Branch already has a worktree")).FontSize(18).FontWeight(600)
			ui.Text(c, s.text("worktree.branch_in_use.body", "Git checks out a branch in one worktree at a time."))
			ui.Text(c, s.target.Branch).TextColor(c.Theme().Accent)
			ui.Text(c, s.target.Path).Selectable()
		case "remove":
			ui.Text(c, s.text("worktree.remove.title", "Remove worktree?")).FontSize(18).FontWeight(600)
			ui.Text(c, s.target.Path).Selectable()
			ui.Text(c, s.text("worktree.remove.clean_body", "The directory and its files will be removed. The Git branch is kept. Commit or copy any work you want to retain before continuing.")).TextColor(c.Theme().Warning)
			if s.Dirty[s.target.Path] {
				ui.Text(c, s.text("worktree.remove.dirty_body", "This worktree has uncommitted changes.")).TextColor(c.Theme().Danger)
			}
			ui.Checkbox(c, &s.force, "Force removal, including uncommitted files")
		case "settings":
			ui.Text(c, s.text("worktree.settings.title", "Worktree settings")).FontSize(18).FontWeight(600)
			ui.TextInput(c, &s.configDraft.Directory).Label(s.text("worktree.settings.directory", "Worktree directory")).Placeholder(".wt").AlignSelf(ui.Stretch)
			ui.TextInput(c, &s.configDraft.SetupScript).Label(s.text("worktree.settings.setup_script", "Setup script")).Placeholder(".orbit/worktree-setup.sh").AlignSelf(ui.Stretch)
			ui.Checkbox(c, &s.configDraft.RunSetupAuto, s.text("worktree.settings.run_setup_auto", "Run setup by default"))
			ui.Checkbox(c, &s.configDraft.OpenAfterCreate, s.text("worktree.settings.open_after_create", "Open after creation"))
			if len(s.Config.AllowedSetupRepos) > 0 {
				ui.Text(c, fmt.Sprintf("Trusted setup repositories: %d", len(s.Config.AllowedSetupRepos)))
				if button(c, "Forget setup trust", false).Clicked() {
					s.confirm("Forget trusted setup repositories?", "Setup scripts will ask for confirmation again.", func() { next := s.Config; next.AllowedSetupRepos = nil; s.applyConfig(next) })
				}
			}
		default:
			ui.Text(c, strings.ToUpper(s.dialog[:1])+s.dialog[1:]+" worktree").FontSize(18).FontWeight(600)
			ui.Text(c, s.target.Path).TextColor(c.Theme().TextMuted)
			in := ui.TextInput(c, &s.value).Label(s.dialog + " value").AutoFocus().AlignSelf(ui.Stretch)
			submit = in.Submitted()
		}
		if s.Error != nil {
			ui.Text(c, s.Error.Error()).TextColor(c.Theme().Danger).Selectable()
		}
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if button(c, s.text("settings.cancel", "Cancel"), false).Clicked() {
				s.dialogOpen = false
			}
			label := s.text("mcp.save", "Save")
			if s.dialog == "create" {
				label = s.text("worktree.dialog.create", "Create")
			}
			if s.dialog == "branch_in_use" {
				label = s.text("worktree.branch_in_use.open", "Open Existing Worktree")
			}
			if s.dialog == "remove" {
				label = s.text("worktree.remove.remove", "Remove")
			}
			disabled := s.Busy || (s.dialog == "create" && (strings.TrimSpace(s.name) == "" || s.createOptions().Branch == "")) || (s.dialog == "remove" && s.Dirty[s.target.Path] && !s.force) || (s.dialog == "branch_in_use" && s.Options.OpenWorkspace == nil)
			if button(c, label, disabled).Clicked() {
				submit = true
			}
		})
		if submit && !s.Busy && (s.dialog != "create" || (strings.TrimSpace(s.name) != "" && s.createOptions().Branch != "")) {
			switch s.dialog {
			case "create":
				s.Create(s.createOptions(), s.openAfter)
			case "branch_in_use":
				if s.Options.OpenWorkspace != nil {
					s.Options.OpenWorkspace(s.target.Path)
					s.dialogOpen = false
				}
			case "remove":
				s.Remove(s.target.Path, s.force, true)
				s.dialogOpen = false
			case "rename":
				s.Rename(s.target.Path, s.value)
				s.dialogOpen = false
			case "move":
				s.Move(s.target.Path, s.value)
				s.dialogOpen = false
			case "lock":
				s.Lock(s.target.Path, s.value)
				s.dialogOpen = false
			case "settings":
				if s.applyConfig(s.configDraft) {
					s.dialogOpen = false
				}
			}
		}
	})
}
func setup(c *ui.Context, s *State) {
	o := s.Setup
	t := c.Theme()
	ui.Column(c).Padding(14).Gap(8).Border(1, t.Border).Radius(8).Children(func() {
		title := s.text("worktree.setup.failed", "Setup failed; worktree was preserved")
		color := t.Danger
		if o.Success {
			title = s.text("worktree.setup.succeeded", "Setup completed")
			color = t.Success
		}
		if o.Cancelled {
			title = "Setup cancelled; worktree was preserved"
		}
		ui.Text(c, title).FontWeight(600).TextColor(color)
		ui.Text(c, o.Script).TextColor(t.TextMuted).Selectable()
		ui.Row(c).Gap(8).Children(func() {
			if button(c, "Show / hide output", false).Clicked() {
				s.ShowOutput = !s.ShowOutput
			}
			if button(c, s.text("worktree.setup.retry", "Retry setup"), s.Busy).Clicked() {
				s.confirm("Run setup script again?", o.Script+"\nThe script may change files or install dependencies.", func() { s.RetrySetup(true) })
			}
			if button(c, s.text("worktree.setup.open_worktree", "Open worktree"), s.Options.OpenWorkspace == nil).Clicked() {
				s.Options.OpenWorkspace(s.SetupWorktree)
			}
			if button(c, s.text("common.dismiss", "Dismiss"), false).Clicked() {
				s.Setup = nil
			}
		})
		if s.ShowOutput {
			ui.Scroll(c).Height(220).Children(func() { ui.Text(c, o.Output()).Selectable().FontSize(12) })
		}
	})
}
