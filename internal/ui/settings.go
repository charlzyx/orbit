package orbitui

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

var settingsCategories = []struct{ ID, Title, Icon, Description string }{
	{"general", "General", "icons/settings.svg", "App behavior, local data, and notifications."},
	{"appearance", "Appearance", "icons/contrast.svg", "Theme, typography, density, and motion."},
	{"shortcuts", "Shortcuts", "icons/keyboard.svg", "Keyboard commands for every part of the workbench."},
	{"privacy", "Privacy", "icons/lock.svg", "Local data and privacy preferences."},
	{"runtime", "Runtime", "icons/server-stack.svg", "The pi CLI, its process, and protocol."},
	{"agent", "Agent", "icons/spark.svg", "Default model, task behavior, and access policy."},
	{"worktrees", "Worktrees", "icons/branch.svg", "Worktree location and setup policy."},
	{"providers", "Providers", "icons/cloud.svg", "Authentication and custom provider configuration."},
	{"models", "Models", "icons/tag-01.svg", "Models reported by pi's catalog."},
	{"skills", "Skills", "icons/magic-wand.svg", "Instructions available to the agent."},
	{"plugins", "Plugins", "icons/extensions.svg", "Packages installed in pi."},
	{"mcp", "MCP", "icons/tools/mcp.svg", "Model Context Protocol servers and their configuration."},
	{"about", "About", "icons/info.svg", "Orbit, rebuilt in Go with native MyGo UI."},
	{"bug", "Report a bug", "icons/github.svg", "Diagnostics and issue reporting."},
}

func (a *App) settings(c *ui.Context) {
	t := a.Theme
	ui.Column(c).Width(SettingsNavWidth).FillHeight().Background(t.Surface).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
		ui.Box(c).Height(HeaderHeight)
		if a.button(c, "Back", "icons/arrow-left.svg", false).Margin(8, 12).Clicked() {
			a.Navigate("chat")
		}
		ui.TextInput(c, &a.settingsSearch).Placeholder("Search settings…").Label("Search settings").Margin(10, 12).AlignSelf(ui.Stretch)
		a.line(c)
		ui.Text(c, "SETTINGS").FontSize(t.Px(10)).TextColor(t.TextMuted).Padding(12, 16, 4)
		ui.Scroll(c).Grow(1).MinHeight(0).PaddingX(8).Gap(2).Children(func() {
			for _, cat := range settingsCategories {
				title := a.settingsTitle(cat.ID, cat.Title)
				if a.settingsSearch != "" && !strings.Contains(strings.ToLower(title+" "+cat.Title+" "+cat.Description), strings.ToLower(a.settingsSearch)) {
					continue
				}
				if a.button(c, title, cat.Icon, cat.ID == a.settingsPage).AlignSelf(ui.Stretch).Height(32).Clicked() {
					a.settingsPage = cat.ID
					a.loadSettings(cat.ID)
				}
			}
		})
		a.line(c)
		ui.Text(c, "Orbit · Go native").Height(HeaderHeight).Padding(12, 16).FontSize(t.Px(12)).TextColor(t.TextMuted)
	})
	if a.settingsPage == "bug" && a.Options.BugReport != nil {
		a.Options.BugReport(c)
		return
	}
	ui.Scroll(c.Key("settings-body:"+a.settingsPage)).Grow(1).MinWidth(0).Padding(24, 32).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).AlignSelf(ui.Center).Gap(t.Space(20)).Children(func() {
			for _, cat := range settingsCategories {
				if cat.ID == a.settingsPage {
					ui.Column(c).Gap(6).Children(func() {
						ui.Text(c, a.settingsTitle(cat.ID, cat.Title)).FontSize(t.Px(20)).FontWeight(500)
						ui.Text(c, cat.Description).TextColor(t.TextMuted)
					})
					break
				}
			}
			a.line(c)
			if a.settingsLoading {
				ui.Spinner(c).Size(18, 18)
			}
			for _, err := range a.settingsErrors {
				ui.Text(c, err).TextColor(t.Error).Selectable()
			}
			switch a.settingsPage {
			case "general":
				a.generalSettings(c)
			case "appearance":
				a.appearanceSettings(c)
			case "shortcuts":
				a.shortcutsSettings(c)
			case "runtime":
				a.runtimeSettings(c)
			case "agent":
				a.agentSettings(c)
			case "worktrees":
				a.worktreeSettings(c)
			case "bug":
				a.bugSettings(c)
			case "models":
				a.modelsSettings(c)
			case "skills":
				a.skillsSettings(c)
			case "plugins":
				a.pluginsSettings(c)
			case "mcp":
				a.mcpSettings(c)
			case "providers":
				a.providersSettings(c)
			case "privacy":
				a.privacySettings(c)
			case "about":
				a.aboutSettings(c)
			}
		})
	})
}
func (a *App) settingsTitle(id, fallback string) string {
	key := "settings." + id
	if id == "bug" {
		key = "settings.report_bug"
	}
	if id == "worktrees" {
		key = "worktree.settings.title"
	}
	return a.text(key, fallback)
}
func (a *App) section(c *ui.Context, title string, fn func()) {
	t := a.Theme
	ui.Column(c).AlignSelf(ui.Stretch).Gap(10).Children(func() {
		ui.Text(c, strings.ToUpper(title)).FontSize(t.Px(10)).TextColor(t.TextMuted).PaddingX(4)
		ui.Column(c).AlignSelf(ui.Stretch).Padding(16).Gap(12).Background(t.Surface).Border(1, t.Border).Radius(12).Children(fn)
	})
}
func (a *App) setting(c *ui.Context, title, description string, control func()) {
	ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(16).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
			ui.Text(c, title).FontWeight(500)
			if description != "" {
				ui.Text(c, description).TextColor(a.Theme.TextMuted).FontSize(a.Theme.Px(12))
			}
		})
		if control != nil {
			control()
		}
	})
}
func (a *App) selectSetting(c *ui.Context, title, description, value string, choices []string, apply func(string)) {
	a.setting(c, title, description, func() {
		ui.MenuButton(c, value, func(m *ui.Menu) {
			for _, s := range choices {
				if m.Item(s).Checked(s == value).Chosen() {
					apply(s)
				}
			}
		}).MinWidth(120)
	})
}
func (a *App) generalSettings(c *ui.Context) {
	a.section(c, "This machine", func() {
		status := "Disconnected"
		if a.Model.Connected {
			status = "Connected"
		}
		a.setting(c, "pi agent", "Spawned as a child process — newline-delimited JSON over stdio.", func() {
			color := a.Theme.TextMuted
			if a.Model.Connected {
				color = a.Theme.Success
			}
			ui.Text(c, status).TextColor(color)
		})
		a.line(c)
		a.setting(c, "Local by default", a.Options.Store.PiDir+"/sessions", nil)
		a.line(c)
		a.setting(c, "Workspace", a.Model.Workspace, func() {
			if a.button(c, "Browse…", "icons/folder.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "BrowseWorkspace"})
			}
		})
	})
	a.section(c, "Notifications", func() {
		n := a.Options.Store.Notifications()
		a.setting(c, "Desktop notifications", "Show a system banner when a run finishes in the background.", func() {
			if ui.Switch(c, &n.Desktop).Label("Desktop notifications").Changed() {
				a.save(a.Options.Store.SaveNotifications(n))
				a.Dispatch(Action{Kind: "NotificationsChanged"})
			}
		})
		a.line(c)
		a.setting(c, "In-app toasts", "Show results and requests while Orbit is in front.", func() {
			if ui.Switch(c, &n.Toasts).Label("In-app toasts").Changed() {
				a.save(a.Options.Store.SaveNotifications(n))
				a.Dispatch(Action{Kind: "NotificationsChanged"})
			}
		})
		a.line(c)
		a.setting(c, "Notification sound", "Play an alert when a background task needs attention.", func() {
			if ui.Switch(c, &n.Sound).Label("Notification sound").Changed() {
				a.save(a.Options.Store.SaveNotifications(n))
				a.Dispatch(Action{Kind: "NotificationsChanged"})
			}
		})
	})
	a.section(c, "Updates", func() {
		a.updateSettings(c)
		a.setting(c, "Check for Updates…", "Check for a compatible, signed native Go release.", func() {
			if a.button(c, "Check for Updates…", "icons/refresh.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "CheckForUpdates"})
			}
		})
	})
}
func (a *App) appearanceSettings(c *ui.Context) {
	a.section(c, "Appearance", func() {
		a.selectSetting(c, "Appearance", "Follow the desktop or choose an appearance.", a.Appearance.Mode, []string{"system", "light", "dark"}, func(v string) { a.Appearance.Mode = v; a.save(a.Options.Store.SaveAppearance(a.Appearance)) })
		a.line(c)
		a.selectSetting(c, "Dark theme", "Original Orbit semantic palettes", a.Appearance.DarkTheme, preferences.DarkThemes, func(v string) { a.Appearance.DarkTheme = v; a.save(a.Options.Store.SaveAppearance(a.Appearance)) })
		a.selectSetting(c, "Light theme", "Original Orbit semantic palettes", a.Appearance.LightTheme, preferences.LightThemes, func(v string) { a.Appearance.LightTheme = v; a.save(a.Options.Store.SaveAppearance(a.Appearance)) })
	})
	a.section(c, "Typography", func() {
		a.selectSetting(c, "UI font size", "Interface text and design tokens scale together.", fmt.Sprint(a.Prefs.UIFontSize), []string{"11", "12", "13", "14", "15", "16", "18", "20"}, func(v string) {
			a.Prefs.UIFontSize, _ = strconv.ParseFloat(v, 64)
			a.save(a.Options.Store.SaveUI(a.Prefs))
		})
		a.selectSetting(c, "Code font size", "Code and file editor text", fmt.Sprint(a.Prefs.EditorFontSize), []string{"11", "12", "13", "14", "15", "16", "18", "20"}, func(v string) {
			a.Prefs.EditorFontSize, _ = strconv.ParseFloat(v, 64)
			a.save(a.Options.Store.SaveUI(a.Prefs))
		})
		a.selectSetting(c, "Terminal font size", "Terminal text", fmt.Sprint(a.Prefs.TerminalFontSize), []string{"11", "12", "13", "14", "15", "16", "18", "20"}, func(v string) {
			a.Prefs.TerminalFontSize, _ = strconv.ParseFloat(v, 64)
			a.save(a.Options.Store.SaveUI(a.Prefs))
			a.Dispatch(Action{Kind: "TerminalFontChanged"})
		})
		families := ui.FontFamilies()
		var names, mono []string
		for _, f := range families {
			if f.Registered {
				names = append(names, f.Name)
				if f.Monospaced {
					mono = append(mono, f.Name)
				}
			}
		}
		a.selectSetting(c, "Interface font", "Bundled fonts are available without a network connection.", a.Fonts.UIFamily, names, func(v string) { a.Fonts.UIFamily = v; a.save(a.Options.Store.SaveFonts(a.Fonts)) })
		a.selectSetting(c, "Code font", "Monospaced font for code and terminal", a.Fonts.CodeFamily, mono, func(v string) {
			a.Fonts.CodeFamily = v
			a.save(a.Options.Store.SaveFonts(a.Fonts))
			a.Dispatch(Action{Kind: "TerminalFontChanged"})
		})
	})
	a.section(c, "Layout and motion", func() {
		a.selectSetting(c, "Spacing density", "Compact 80%, default 100%, comfortable 120%.", strconv.Itoa(a.Prefs.SpacingDensity), []string{"80", "85", "90", "95", "100", "105", "110", "115", "120"}, func(v string) { a.Prefs.SpacingDensity, _ = strconv.Atoi(v); a.save(a.Options.Store.SaveUI(a.Prefs)) })
		a.setting(c, "Reduce motion", "Disable nonessential animations.", func() {
			if ui.Switch(c, &a.Prefs.ReduceMotion).Label("Reduce motion").Changed() {
				a.save(a.Options.Store.SaveUI(a.Prefs))
			}
		})
		a.selectSetting(c, "Language", "All bundled Orbit translations", a.Prefs.Language, preferences.Languages, func(v string) {
			a.Prefs.Language = v
			a.reloadLocale()
			a.save(a.Options.Store.SaveUI(a.Prefs))
			a.Dispatch(Action{Kind: "LanguageChanged"})
		})
	})
}
func (a *App) shortcutsSettings(c *ui.Context) {
	for _, category := range []string{"Session", "Agent", "Navigation", "Panels", "Review", "Git", "Application"} {
		category := category
		a.section(c, category, func() {
			for _, cmd := range Commands {
				if cmd.Category != category {
					continue
				}
				chords := []string{}
				for _, b := range Bindings {
					if b.Command == cmd.ID {
						chords = append(chords, shortcutLabel(b.Chord))
					}
				}
				if len(chords) == 0 {
					chords = []string{"Command palette"}
				}
				a.setting(c, a.Tr(cmd.TitleKey), "", func() {
					ui.Text(c, strings.Join(chords, "  ")).FontSize(12).Padding(4, 8).Background(a.Theme.Elevated).Radius(4)
				})
			}
		})
	}
}
func (a *App) runtimeSettings(c *ui.Context) {
	defer a.runtimeNative(c)
	a.section(c, "Process", func() {
		a.setting(c, "Runtime", "pi --mode rpc · direct stdio JSONL", nil)
		a.setting(c, "Status", a.Model.Status, nil)
		if a.Model.Error != "" {
			ui.Text(c, a.Model.Error).TextColor(a.Theme.Error)
		}
		ui.Row(c).Gap(8).Children(func() {
			if a.button(c, "Restart runtime", "icons/refresh.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "RestartRuntime"})
			}
			if a.button(c, "Choose pi executable…", "icons/folder.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "ChoosePiBinary"})
			}
			if a.button(c, "Check pi update", "icons/refresh.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "CheckPiUpdate"})
			}
		})
	})
}
func (a *App) agentSettings(c *ui.Context) {
	a.section(c, "New sessions", func() {
		d := a.Options.Store.SessionDefault()
		label := d.ModelID
		if label == "" {
			label = "Use pi default"
		}
		a.setting(c, "Default model", "Applies only to new sessions; existing sessions keep their model.", func() {
			ui.MenuButton(c, label, func(m *ui.Menu) {
				if m.Item("Use pi default").Chosen() {
					a.save(a.Options.Store.SaveSessionDefault(preferences.SessionDefault{}))
				}
				for _, model := range a.Model.Models {
					if m.Item(model.Provider + " / " + model.ID).Chosen() {
						d.Provider = model.Provider
						d.ModelID = model.ID
						a.save(a.Options.Store.SaveSessionDefault(d))
					}
				}
			})
		})
		a.selectSetting(c, "Default thinking", "Applied when a new session begins.", d.Thinking, a.Model.ThinkingLevels, func(v string) { d.Thinking = v; a.save(a.Options.Store.SaveSessionDefault(d)) })
		a.selectSetting(c, "Send while running", "Enter queues a follow-up or steers the running agent.", a.Prefs.ComposerSendMode, []string{"follow_up", "steer"}, func(v string) { a.Prefs.ComposerSendMode = v; a.save(a.Options.Store.SaveUI(a.Prefs)) })
		title := a.Options.Store.AutoTitle()
		a.setting(c, "Automatic task titles", "Generate task titles using the bundled pi extension.", func() {
			if ui.Switch(c, &title.Enabled).Label("Automatic task titles").Changed() {
				a.save(a.Options.Store.SaveAutoTitle(title))
				a.Dispatch(Action{Kind: "AutoTitleChanged"})
			}
		})
	})
	a.section(c, "Access", func() {
		a.selectSetting(c, "Access policy", "Guarded by Orbit's pi extension, not an OS sandbox.", a.access, preferences.AccessModes, func(v string) {
			a.access = v
			a.save(a.Options.Store.SaveAccess(v))
			a.Dispatch(Action{Kind: "SetAccess", Value: v})
		})
	})
}
func (a *App) modelsSettings(c *ui.Context) {
	ui.Row(c).Gap(8).Children(func() {
		ui.TextInput(c, &a.modelQuery).Placeholder("Filter models…").Label("Filter models").Grow(1)
		if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
			a.Dispatch(Action{Kind: "RefreshModels"})
		}
	})
	for _, m := range a.Model.Models {
		if a.modelQuery != "" && !strings.Contains(strings.ToLower(m.Provider+" "+m.ID+" "+m.Name), strings.ToLower(a.modelQuery)) {
			continue
		}
		m := m
		a.section(c, m.Provider, func() {
			a.setting(c, m.Name, m.ID, func() {
				if a.button(c, "Select", "", m.ID == a.Model.ModelID).Clicked() {
					a.Dispatch(Action{Kind: "SetModel", Value: m.ID, Data: map[string]any{"provider": m.Provider}})
				}
			})
		})
	}
	if len(a.Model.Models) == 0 {
		ui.Text(c, "No models have been reported by pi.").TextColor(a.Theme.TextMuted)
	}
}
func (a *App) privacySettings(c *ui.Context) {
	defer a.privacyNative(c)
	a.section(c, a.text("settings.session_retention", "Session retention"), func() {
		ret := a.retention
		a.setting(c, a.text("settings.session_retention_enabled", "Automatic session cleanup"), "Apply this policy once per launch, or with Clean up now. Open tasks are kept.", func() {
			if ui.Switch(c, &ret.Enabled).Label("Automatic session cleanup").Changed() {
				a.retention = ret
				a.save(a.Options.Store.SaveRetention(ret))
			}
		})
		a.selectSetting(c, "Action", "Deletion permanently removes matching saved session files.", ret.Mode, []string{"archive", "delete"}, func(v string) { a.retention.Mode = v; a.save(a.Options.Store.SaveRetention(a.retention)) })
		a.selectSetting(c, "Keep recent sessions", "Idle age in days, based on message activity", strconv.Itoa(ret.Days), []string{"7", "14", "30", "60", "90", "180", "365"}, func(v string) {
			a.retention.Days, _ = strconv.Atoi(v)
			a.save(a.Options.Store.SaveRetention(a.retention))
		})
		if a.button(c, a.text("settings.session_retention_run", "Clean up now"), "icons/refresh.svg", false).Disabled(!ret.Enabled || a.retentionBusy).Clicked() {
			if ret.Mode == "delete" {
				a.Confirm("Permanently delete old tasks?", fmt.Sprintf("Delete saved sessions idle for more than %d days. Open tasks are excluded. This cannot be undone.", ret.Days), func() { a.Dispatch(Action{Kind: "RunRetention"}) })
			} else {
				a.Dispatch(Action{Kind: "RunRetention"})
			}
		}
	})

	a.section(c, "Local data", func() {
		a.setting(c, "Session storage", a.Options.Store.PiDir+"/sessions", func() {
			if a.button(c, "Open folder", "icons/folder.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "RevealPath", Value: a.Options.Store.PiDir + "/sessions"})
			}
		})
		a.setting(c, "Orbit preferences", a.Options.Store.RootDir, func() {
			if a.button(c, "Open folder", "icons/folder.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "RevealPath", Value: a.Options.Store.RootDir})
			}
		})
		ui.Text(c, "Messages are sent to the model provider selected in pi. Orbit does not substitute another agent runtime or cloud store.").TextColor(a.Theme.TextMuted)
	})
}
func (a *App) aboutSettings(c *ui.Context) {
	defer a.updateSettings(c)
	a.section(c, "Orbit", func() {
		ui.Text(c, "Orbit "+a.Options.Version+" · Go native").FontSize(24).FontWeight(600)
		ui.Text(c, "Native desktop workbench for the pi coding agent.")
		ui.Text(c, "MyGo 9c69cf829ece · "+runtime.Version()+" · Pi RPC").TextColor(a.Theme.TextMuted)
		ui.Text(c, "Copyright 2026 Orbit contributors. Apache License 2.0. MyGo and bundled third-party assets retain their respective licenses.").TextColor(a.Theme.TextMuted)
		if a.button(c, "Project repository", "icons/external-link.svg", false).Clicked() {
			a.Dispatch(Action{Kind: "OpenLink", Value: "https://github.com/charlzyx/orbit"})
		}
	})
}
func (a *App) loadSettings(category string) {
	if category == "providers" {
		a.refreshProviders()
		return
	}
	a.settingsGeneration++
	generation := a.settingsGeneration
	root := a.Model.Workspace
	a.settingsErrors = nil
	a.settingsLoading = false
	if category != "skills" && category != "plugins" && category != "mcp" && category != "providers" {
		return
	}
	a.settingsLoading = true
	svc := a.Service()
	apply := func(fn func()) func() {
		return func() {
			if generation != a.settingsGeneration || root != a.Model.Workspace {
				return
			}
			fn()
		}
	}
	a.Async(func() func() {
		var errs []error
		switch category {
		case "skills":
			v, e := svc.Skills()
			return apply(func() { a.skills = v; a.settingsLoading = false; a.settingsErrors = errorStrings(e) })
		case "plugins":
			v, e := svc.Plugins()
			return apply(func() { a.plugins = v; a.settingsLoading = false; a.settingsErrors = errorStrings(e) })
		case "mcp":
			v, e := svc.MCP()
			return apply(func() { a.servers = v; a.settingsLoading = false; a.settingsErrors = errorStrings(e) })
		}
		return apply(func() {
			a.settingsLoading = false
			a.settingsErrors = errorStrings(errs)
			a.Dispatch(Action{Kind: "RefreshProviders"})
		})
	})
}
func errorStrings(errs []error) []string {
	var out []string
	for _, e := range errs {
		if e != nil {
			out = append(out, e.Error())
		}
	}
	return out
}
func (a *App) skillsSettings(c *ui.Context) {
	ui.Row(c).Gap(8).Children(func() {
		if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
			a.loadSettings("skills")
		}
		if a.button(c, "Open skills folder", "icons/folder.svg", false).Clicked() {
			a.Dispatch(Action{Kind: "RevealPath", Value: a.Options.Store.PiDir + "/skills"})
		}
	})
	for i, s := range a.skills {
		i, s := i, s
		ui.Column(c).Border(1, a.Theme.Border).Radius(8).Padding(12).Gap(8).Children(func() {
			a.setting(c, s.Name, s.Description, func() {
				enabled := s.Enabled
				if ui.Switch(c, &enabled).Label("Enable " + s.Name).Changed() {
					svc := a.Service()
					a.Async(func() func() {
						e := svc.SetSkillEnabled(s, enabled)
						return func() { a.save(e); a.loadSettings("skills") }
					})
				}
			})
			ui.Row(c).Gap(6).Children(func() {
				ui.Text(c, s.Scope+" · "+s.Path).FontSize(11).TextColor(a.Theme.TextMuted).Grow(1)
				if a.button(c, "View", "", a.selectedSkill == i).Clicked() {
					if a.selectedSkill == i {
						a.selectedSkill = -1
					} else {
						a.selectedSkill = i
					}
				}
				if a.iconButton(c, "Delete skill", "icons/trash.svg").Clicked() {
					a.Confirm("Delete skill?", s.Path, func() {
						svc := a.Service()
						a.Async(func() func() { e := svc.DeleteSkill(s.Path); return func() { a.save(e); a.loadSettings("skills") } })
					})
				}
			})
			if a.selectedSkill == i {
				markdown.View(c, s.Content, &a.md)
			}
		})
	}
	if len(a.skills) == 0 && !a.settingsLoading {
		ui.Text(c, "No skills found in the current pi configuration.").TextColor(a.Theme.TextMuted)
	}
}
func (a *App) pluginsSettings(c *ui.Context) {
	ui.Row(c).Gap(8).Children(func() {
		if a.button(c, "Install package…", "icons/plus.svg", false).Clicked() {
			a.Prompt("Install pi package", "", func(source string) { a.pluginAction("install", source, "user") })
		}
		if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
			a.loadSettings("plugins")
		}
	})
	for _, p := range a.plugins {
		p := p
		a.section(c, p.Scope, func() {
			a.setting(c, p.Name, p.Source, func() {
				a.button(c, "Actions", "icons/more.svg", false).Menu(func(m *ui.Menu) {
					if m.Item("Update").Chosen() {
						a.pluginAction("update", p.Source, p.Scope)
					}
					if m.Item("Remove…").Chosen() {
						a.Confirm("Remove package?", p.Source, func() { a.pluginAction("remove", p.Source, p.Scope) })
					}
				})
			})
		})
	}
	if len(a.plugins) == 0 && !a.settingsLoading {
		ui.Text(c, "No installed packages found.").TextColor(a.Theme.TextMuted)
	}
}
func (a *App) pluginAction(action, source, scope string) {
	svc := a.Service()
	a.settingsLoading = true
	a.Async(func() func() {
		out, e := svc.PluginAction(context.Background(), action, source, scope)
		return func() {
			a.settingsLoading = false
			if e != nil {
				a.Notify(e.Error(), true)
			} else {
				a.Notify(out, false)
			}
			a.loadSettings("plugins")
		}
	})
}
func (a *App) mcpSettings(c *ui.Context) {
	ui.Row(c).Gap(8).Children(func() {
		if a.button(c, "Add server…", "icons/plus.svg", false).Clicked() {
			a.editMCP("", "user", "", `{"command":"","args":[]}`)
		}
		if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
			a.loadSettings("mcp")
		}
		if a.button(c, "Probe servers", "icons/refresh.svg", false).Clicked() {
			svc := a.Service()
			a.Async(func() func() {
				raw, e := svc.ProbeMCP(context.Background())
				return func() {
					if e != nil {
						a.Notify(e.Error(), true)
					} else {
						a.Notify(string(raw), false)
					}
				}
			})
		}
	})
	for _, server := range a.servers {
		s := server
		a.section(c, s.Scope, func() {
			a.setting(c, s.Name, s.Description, func() {
				a.button(c, "Actions", "icons/more.svg", false).Menu(func(m *ui.Menu) {
					if m.Item("Edit…").Chosen() {
						b, _ := json.MarshalIndent(s.Definition, "", "  ")
						a.editMCP(s.Name, s.Scope, s.Name, string(b))
					}
					if m.Item("Remove…").Chosen() {
						a.Confirm("Remove MCP server?", s.Name, func() {
							svc := a.Service()
							a.Async(func() func() { e := svc.RemoveMCP(s.Scope, s.Name); return func() { a.save(e); a.loadSettings("mcp") } })
						})
					}
				})
			})
		})
	}
	if len(a.servers) == 0 && !a.settingsLoading {
		ui.Text(c, "No MCP servers configured.").TextColor(a.Theme.TextMuted)
	}
}
func (a *App) editMCP(oldName, scope, name, value string) {
	a.jsonTitle = "MCP server"
	a.jsonName = name
	a.jsonScope = scope
	a.jsonValue = value
	a.jsonOpen = true
	svc := a.Service()
	a.jsonCommit = func(name, scope, value string) error {
		var d map[string]any
		if e := json.Unmarshal([]byte(value), &d); e != nil {
			return e
		}
		return svc.SaveMCP(scope, oldName, name, d)
	}
}
func (a *App) jsonDialog(c *ui.Context) {
	if !a.jsonOpen {
		return
	}
	a.modal(c.Key("configuration-editor"), &a.jsonOpen, 680, func() {
		ui.Text(c, a.jsonTitle).FontSize(18).Bold()
		ui.TextInput(c, &a.jsonName).Label("Name").Placeholder("Name").AlignSelf(ui.Stretch)
		ui.MenuButton(c, a.jsonScope, func(m *ui.Menu) {
			for _, s := range []string{"user", "project"} {
				if m.Item(s).Chosen() {
					a.jsonScope = s
				}
			}
		})
		_, height := c.Size()
		ui.TextArea(c, &a.jsonValue).Font(a.Fonts.CodeFamily).Label("JSON configuration").Height(min(360, max(100, height-250))).AlignSelf(ui.Stretch)
		ui.Row(c).Justify(ui.End).Gap(8).Children(func() {
			if a.button(c, "Cancel", "", false).Clicked() {
				a.jsonOpen = false
			}
			if a.button(c, "Save", "", true).Clicked() {
				fn, name, scope, value := a.jsonCommit, a.jsonName, a.jsonScope, a.jsonValue
				if fn != nil {
					a.Async(func() func() {
						e := fn(name, scope, value)
						return func() {
							if e != nil {
								a.Notify(e.Error(), true)
							} else {
								a.jsonOpen = false
								a.loadSettings(a.settingsPage)
							}
						}
					})
				}
			}
		})
	})
}
func (a *App) providersSettings(c *ui.Context) {
	a.section(c, "Providers", func() {
		ui.Text(c, "Provider authentication is handled by pi over RPC. Credentials remain in pi's configuration.").TextColor(a.Theme.TextMuted)
		if a.button(c, "Refresh providers", "icons/refresh.svg", false).Clicked() {
			a.refreshProviders()
		}
		a.providerRows(c)
	})
}

func (a *App) worktreeSettings(c *ui.Context) {
	path := filepath.Join(a.Options.Store.RootDir, "worktrees.json")
	config := workspace.LoadWorktreeConfig(path)
	a.section(c, "Worktree defaults", func() {
		a.setting(c, "Directory", config.Directory, func() {
			if a.button(c, "Edit…", "", false).Clicked() {
				a.Prompt("Worktree directory", config.Directory, func(v string) { config.Directory = v; a.save(workspace.SaveWorktreeConfig(path, config)) })
			}
		})
		a.setting(c, "Setup script", config.SetupScript, func() {
			if a.button(c, "Edit…", "", false).Clicked() {
				a.Prompt("Setup script", config.SetupScript, func(v string) { config.SetupScript = v; a.save(workspace.SaveWorktreeConfig(path, config)) })
			}
		})
		a.setting(c, "Run setup automatically", "Untrusted repository scripts still require confirmation.", func() {
			if ui.Switch(c, &config.RunSetupAuto).Changed() {
				a.save(workspace.SaveWorktreeConfig(path, config))
			}
		})
		a.setting(c, "Open after creation", "Switch to a newly created worktree.", func() {
			if ui.Switch(c, &config.OpenAfterCreate).Changed() {
				a.save(workspace.SaveWorktreeConfig(path, config))
			}
		})
	})
}
func (a *App) bugSettings(c *ui.Context) {
	a.section(c, "Report a bug", func() {
		ui.Text(c, "Review diagnostics before sharing them. Do not include API keys or private session content.").TextColor(a.Theme.TextMuted)
		if a.button(c, "Copy diagnostics", "icons/copy.svg", false).Clicked() {
			c.Services().WriteClipboard("Orbit Go native\nMyGo 0.6.3\nRuntime connected: " + fmt.Sprint(a.Model.Connected))
		}
		if a.button(c, "Open issue tracker", "icons/external-link.svg", false).Clicked() {
			a.Dispatch(Action{Kind: "OpenLink", Value: "https://github.com/charlzyx/orbit/issues"})
		}
	})
}
