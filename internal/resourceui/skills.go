package resourceui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func (s *State) skillsView(c *ui.Context) ui.Element {
	a := s.appearance(c)
	root := ui.Row(c).Grow(1).MinWidth(0).MinHeight(0).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Background(a.native.Background).LineHeight(1.618034)
	root.Children(func() {
		s.skillsList(c)
		selected := s.skill(s.Selected)
		if selected == nil {
			ui.Column(c.Key("select-empty")).Grow(1).MinWidth(0).Center().Gap(a.Space(8)).Children(func() {
				if s.Busy {
					ui.Spinner(c).Size(a.Px(22), a.Px(22))
				} else {
					icon(c, "magic-wand", a.Px(30), a.native.TextMuted)
				}
				ui.Text(c, s.label("skills_ui.select_a_skill", "Select a skill")).FontSize(a.Px(14)).FontWeight(500)
				ui.Text(c, s.label("skills_ui.pick_one_from_the_list_to_see_its_details_and_sk", "Pick one from the list to see its details and SKILL.md.")).FontSize(a.Px(12)).TextColor(a.Secondary)
				for i, e := range s.SkillErrors {
					s.notice(c, fmt.Sprintf("skill-error-%d", i), s.label("settings.could_not_read_settings", "Could not read settings"), e, a.native.Danger)
				}
				if s.Error != "" {
					s.notice(c, "action-error", s.label("resources.action_failed", "Action failed"), s.Error, a.native.Danger)
				}
			})
		} else {
			s.skillDetail(c, *selected)
		}
	})
	return root
}

func (s *State) skillsList(c *ui.Context) {
	a := s.appearance(c)
	visible := s.visibleSkills()
	ui.Column(c.Key("master")).Width(320).Shrink(0).MinHeight(0).Background(a.native.Surface).BorderWidth(0, 1, 0, 0).BorderColor(a.native.Border).Children(func() {
		ui.Row(c.Key("search-row")).Label("Skills search row").Height(a.Px(36)).Shrink(0).Margin(a.Space(12), a.Space(12), a.Space(8)).Padding(0, a.Px(10)).Gap(a.Px(4)).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).AlignItems(ui.Center).Children(func() {
			icon(c, "search", a.Px(14), a.native.TextMuted)
			ui.TextInput(c.Key("filter"), &s.SkillQuery).Placeholder(s.label("app.search_skills", "Search skills…")).Label(s.label("app.search_skills", "Search skills")).Grow(1).MinWidth(0).Height(a.Px(36)).Padding(0).Background(ui.Color{}).Border(0, ui.Color{}).FocusRing(false)
		})
		ui.Row(c.Key("all-row")).Margin(0, a.Space(12), a.Space(10)).Height(30).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(8)).Radius(a.Px(8)).Background(a.Elevated).Border(1, a.native.Border).AlignItems(ui.Center).Children(func() {
			icon(c, "magic-wand", a.Px(14), a.Secondary)
			ui.Text(c, s.label("skills_ui.all_skills", "All skills")).Grow(1).FontSize(a.Px(12)).TextColor(a.Secondary)
			ui.Text(c, fmt.Sprint(len(s.Skills))).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		})
		list := ui.Scroll(c.Key("skill-list")).TrackScroll(&s.listScroll).Grow(1).MinHeight(0).Focusable().Label(s.label("resources.skills_list", "Skills list"))
		list.Children(func() {
			for _, scope := range []string{"project", "global", "user"} {
				var group []workbench.Skill
				for _, p := range visible {
					if p.Scope == scope {
						group = append(group, p)
					}
				}
				if len(group) == 0 {
					continue
				}
				ui.Row(c.Key("group-"+scope)).Padding(10, a.Space(16), a.Space(4)).Gap(a.Space(8)).Children(func() {
					ui.Text(c, strings.ToUpper(s.scopeLabel(scope))).FontSize(a.Px(10)).FontWeight(500).TextColor(a.native.TextMuted)
					ui.Text(c, fmt.Sprint(len(group))).FontSize(a.Px(10)).TextColor(a.native.TextMuted)
				})
				for _, p := range group {
					s.skillRow(c, p)
				}
			}
			if len(visible) == 0 {
				ui.Column(c.Key("no-matches")).Label("Skills no matches state").Padding(28, a.Space(16)).AlignItems(ui.Center).Gap(a.Px(8)).Children(func() {
					if s.Busy {
						ui.Spinner(c).Size(a.Px(22), a.Px(22))
					} else {
						icon(c, "search", a.Px(22), a.native.TextMuted)
					}
					label := s.label("skills_ui.no_skills_match", "No skills match")
					if s.Busy && !s.loaded {
						label = s.activity
					}
					ui.Text(c, label).FontSize(a.Px(12)).TextColor(a.Secondary)
				})
			}
		})
		if list.Shortcut(0, ui.KeyDown) {
			s.moveSkill(1)
		}
		if list.Shortcut(0, ui.KeyUp) {
			s.moveSkill(-1)
		}
		if list.Shortcut(0, ui.KeyHome) {
			if len(visible) > 0 {
				s.SelectSkill(visible[0].Path)
				s.focusSkill = visible[0].Path
			}
		}
		if list.Shortcut(0, ui.KeyEnd) {
			if len(visible) > 0 {
				s.SelectSkill(visible[len(visible)-1].Path)
				s.focusSkill = visible[len(visible)-1].Path
			}
		}
		ui.Text(c, s.label("skills_ui.skill_count", "%{count} skill(s)", "count", fmt.Sprint(len(visible)))).Padding(a.Space(8), a.Px(14)).BorderWidth(1, 0, 0, 0).BorderColor(a.native.Border).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Shrink(0)
	})
}
func (s *State) skillRow(c *ui.Context, p workbench.Skill) {
	a := s.appearance(c)
	selected := p.Path == s.Selected
	row := ui.ButtonBase(c.Key(p.Path)).Label("Skill: "+p.Name).Margin(0, 10).Padding(7, a.Space(8)).Gap(10).MinHeight(a.Px(48)).Shrink(0).Radius(a.Px(6)).AlignItems(ui.Start)
	if selected {
		row.Background(a.native.Selection)
	} else if row.Hovered() {
		row.Background(a.native.SurfaceHover)
	}
	if !p.Enabled {
		row.Opacity(.55)
	}
	row.Children(func() {
		ui.Box(c).Size(26, 26).Shrink(0).Radius(a.Px(6)).Background(a.Elevated).Border(1, a.native.Border).Center().Children(func() { icon(c, "magic-wand", a.Px(14), a.native.TextMuted) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(a.Px(2)).Children(func() {
			textColor := a.native.Text
			if selected {
				textColor = a.SelectionText
			}
			ui.Text(c, p.Name).FontSize(a.Px(12)).FontWeight(500).TextColor(textColor).SingleLine()
			if p.Description != "" {
				ui.Text(c, p.Description).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine()
			}
		})
	})
	if row.Clicked() {
		s.SelectSkill(p.Path)
	}
	if s.focusSkill == p.Path {
		row.Focus().ScrollIntoView()
		s.focusSkill = ""
	}
	if selected && row.Focused() {
		row.ScrollIntoView()
	}
}

func (s *State) skillDetail(c *ui.Context, p workbench.Skill) {
	a := s.appearance(c)
	// Source keeps facts/actions fixed and only scrolls SKILL.md. At minimum
	// window widths the same controls wrap, with an outer scroll for reachability.
	width, _ := c.Size()
	compact := width < 1000
	var detail ui.Element
	if compact {
		detail = ui.Scroll(c.Key("detail-compact-" + p.Path)).TrackScroll(&s.detailScroll).Grow(1).MinWidth(0).MinHeight(0)
	} else {
		detail = ui.Column(c.Key("detail-" + p.Path)).Grow(1).MinWidth(0).MinHeight(0)
	}
	detail.Padding(18, a.Space(24), a.Space(12)).Gap(a.Space(12)).Children(func() {
		ui.Row(c.Key("header")).Gap(a.Space(12)).AlignItems(ui.Start).Children(func() {
			ui.Box(c).Size(40, 40).Shrink(0).Radius(a.Px(8)).Background(a.Elevated).Border(1, a.native.Border).Center().Children(func() { icon(c, "magic-wand", a.Px(20), a.native.TextMuted) })
			ui.Column(c).Grow(1).MinWidth(0).Gap(a.Px(2)).Children(func() {
				ui.Text(c, p.Name).FontSize(a.Px(17)).FontWeight(500).SingleLine().Tooltip(p.Name)
				subtitle := ""
				if p.Scope == "project" {
					subtitle = s.label("skills_ui.scope_project", "Project · %{path}", "path", s.displayPath(s.Workspace))
				} else if p.Scope == "user" {
					subtitle = s.label("skills_ui.scope_user", "User · agents")
				} else {
					subtitle = s.label("skills_ui.scope_agent", "Global · pi agent")
				}
				if p.ManualOnly {
					subtitle = s.label("skills_ui.scope_manual_only", "%{scope} · manual only", "scope", subtitle)
				}
				ui.Text(c, subtitle).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine().Tooltip(subtitle)
			})
			on := p.Enabled
			knob := a.native.Text
			if !a.native.Dark {
				knob = ui.RGB(255, 255, 255)
			}
			toggle := ui.SwitchBase(c.Key("enabled"), &on).Label(s.label("resources.enable_skill", "Enable skill")).Size(36, 20).Shrink(0).Padding(2).Radius(999).Border(1, a.native.Border).AlignItems(ui.Center).Background(a.Elevated).Disabled(s.Busy)
			if p.Enabled {
				toggle.Background(a.native.Accent).Justify(ui.End)
			} else {
				toggle.Justify(ui.Start)
			}
			toggle.Children(func() { ui.Box(c).Size(14, 14).Shrink(0).Radius(999).Background(knob) })
			if toggle.Changed() {
				s.SetSkillEnabled(p.Path, on)
			}
		})
		description, color := p.Description, a.Secondary
		if !p.Valid() {
			description = s.label("skills_ui.no_description", "No description. This skill is not valid for automatic discovery.")
			color = a.native.Danger
		}
		ui.Text(c, description).MaxWidth(760).FontSize(a.Px(12)).TextColor(color)
		ui.Column(c.Key("facts")).Gap(a.Px(9)).Children(func() {
			s.skillFact(c, "invoke", s.label("resources.invoke", "Invoke"), p.InvokeCommand())
			s.skillFact(c, "location", s.label("resources.location", "Location"), s.displayPath(filepath.Dir(p.Path)))
			s.skillFact(c, "size", s.label("resources.contents", "Contents"), formatSize(p.Size))
			s.skillFact(c, "updated", s.label("resources.updated", "Updated"), s.skillAge(p.Modified, time.Now()))
		})
		s.skillActions(c, p)
		if s.Busy {
			s.statusLine(c, "")
		}
		for i, e := range s.SkillErrors {
			s.notice(c, fmt.Sprintf("read-error-%d", i), s.label("settings.could_not_read_settings", "Could not read settings"), e, a.native.Danger)
		}
		if s.Error != "" {
			s.notice(c, "action-error", s.label("resources.action_failed", "Action failed"), s.Error, a.native.Danger)
		}
		if s.Status != "" && s.Options.Notify == nil {
			ui.Text(c, s.Status).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		}
		s.separator(c)
		ui.Text(c, s.label("skills_ui.skill_md", "SKILL.MD")).FontSize(a.Px(10)).FontWeight(500).TextColor(a.native.TextMuted).Shrink(0)
		body := func() {
			if strings.TrimSpace(p.Body()) == "" {
				ui.Text(c, s.label("skills_ui.skill_md_is_empty", "SKILL.md is empty")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				return
			}
			th := markdown.DefaultTheme(a.native)
			th.CodeFont = a.CodeFont
			th.FontSize = a.Px(14)
			th.LineHeight = 1.55
			th.Code = a.native.Accent
			th.CodeBackground = a.native.Accent.Alpha(.1)
			th.CodeBlockBackground = a.Elevated
			th.CodeBlockBorder = a.native.Border
			th.CodeBlockRadius = a.Px(8)
			s.md.Theme = th
			markdown.View(c.Key("markdown-"+p.Path), p.Body(), &s.md).MaxWidth(760).MinWidth(0)
		}
		if compact {
			body()
		} else {
			ui.Scroll(c.Key("body-scroll-" + p.Path)).TrackScroll(&s.detailScroll).Grow(1).MinHeight(0).Children(body)
		}
	})
}
func (s *State) skillFact(c *ui.Context, key, label, value string) {
	a := s.appearance(c)
	ui.Row(c.Key(key)).Gap(a.Space(12)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).Width(84).Shrink(0).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		ui.Text(c, value).Grow(1).MinWidth(0).FontSize(a.Px(12)).SingleLine().Tooltip(value).Selectable()
	})
}
func (s *State) skillActions(c *ui.Context, p workbench.Skill) {
	a := s.appearance(c)
	if s.skillDelete == p.Path {
		ui.Column(c.Key("delete-confirm")).Gap(a.Space(8)).Children(func() {
			ui.Text(c, s.label("skills_ui.delete_skill_hint", "Delete %{path} and all its contents? This cannot be undone.", "path", s.displayPath(filepath.Dir(p.Path)))).FontSize(a.Px(12)).TextColor(a.native.Danger)
			ui.Row(c).Gap(a.Space(8)).Wrap().Children(func() {
				if s.button(c, "confirm-delete", s.label("sidebar.delete", "Delete"), "trash", false).Disabled(s.Busy).Clicked() {
					s.ConfirmDeleteSkill()
				}
				if s.button(c, "cancel-delete", s.label("git_panel.cancel", "Cancel"), "", false).Disabled(s.Busy).Clicked() {
					s.Escape()
				}
			})
		})
		return
	}
	ui.Row(c.Key("actions")).Wrap().Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
		if s.button(c, "open", s.label("skills_ui.open_skill_md", "Open SKILL.md"), "file", false).Disabled(s.Options.OpenFile == nil).Clicked() {
			s.Options.OpenFile(p.Path)
		}
		if s.button(c, "reveal", s.label("skills_ui.show_in_file_manager", "Show in File Manager"), "folder", false).Disabled(s.Options.RevealPath == nil).Clicked() {
			s.Options.RevealPath(p.Path)
		}
		if s.button(c, "copy", s.label("sidebar.copy_path", "Copy path"), "copy", false).Disabled(s.Options.CopyText == nil).Clicked() {
			s.Options.CopyText(s.displayPath(p.Path))
			s.Status = s.label("skills_ui.copied_path", "Copied %{path}", "path", s.displayPath(p.Path))
			s.notify(s.Status, false)
		}
		ui.Box(c).Grow(1).MinWidth(0)
		if s.button(c, "delete", s.label("sidebar.delete", "Delete"), "trash", false).Disabled(s.Busy).Clicked() {
			s.AskDeleteSkill(p.Path)
		}
	})
}
