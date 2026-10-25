package orbitui

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
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
	ui.Column(c.Key("settings-nav")).Width(SettingsNavWidth).FillHeight().LineHeight(sourceDefaultLineHeight).Background(t.Surface).BorderWidth(0, 1, 0, 0).BorderColor(t.Border).Children(func() {
		ui.Box(c).Height(HeaderHeight).AlignSelf(ui.Stretch).DragWindow()
		ui.Column(c).AlignSelf(ui.Stretch).Padding(t.Space(4), t.Space(8), t.Space(12)).Children(func() {
			if a.button(c, a.text("settings.back", "Back"), "icons/arrow-left.svg", false).Height(t.Px(28)).AlignSelf(ui.Stretch).Clicked() {
				a.Navigate("chat")
			}
		})
		ui.Column(c).AlignSelf(ui.Stretch).Padding(0, t.Space(8), t.Space(12)).Children(func() {
			ui.Row(c).Height(t.Px(36)).PaddingX(t.Px(10)).Gap(t.Space(6)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
				a.icon(c, "icons/search.svg", 14, t.TextMuted)
				ui.TextInputBase(c, &a.settingsSearch).Label("Search settings").Placeholder(a.text("settings.search_settings", "Search settings…")).Grow(1)
			})
		})
		ui.Text(c, a.text("settings.settings", "SETTINGS")).FontSize(t.Px(10)).FontWeight(600).TextColor(t.TextMuted).Padding(0, 0, t.Space(6), t.Space(16))
		ui.Scroll(c).Grow(1).MinHeight(0).AlignSelf(ui.Stretch).PaddingX(t.Space(8)).Gap(t.Space(2)).Children(func() {
			for _, cat := range settingsCategories {
				if cat.ID == "runtime" || cat.ID == "skills" || cat.ID == "about" || cat.ID == "bug" {
					ui.Box(c).AlignSelf(ui.Stretch).Height(1).Background(t.Border).Margin(t.Space(6), 0, t.Space(6))
				}
				selected := cat.ID == a.settingsPage
				b := ui.ButtonBase(c.Key("settings-nav:" + cat.ID)).AlignSelf(ui.Stretch).Height(t.Px(28)).PaddingX(t.Space(8)).Gap(t.Space(8)).Radius(t.Px(6)).Justify(ui.Start).Label(a.settingsTitle(cat.ID, cat.Title))
				color, icon := t.TextSecondary, t.TextMuted
				if selected {
					b.Background(t.Selection)
					color = t.SelectionFg
					icon = t.AccentText
				} else if b.Hovered() {
					b.Background(t.Hover)
				}
				b.Children(func() { a.icon(c, cat.Icon, 14, icon); ui.Text(c, a.settingsTitle(cat.ID, cat.Title)).TextColor(color) })
				if b.Clicked() {
					a.selectSettingsPage(cat.ID)
				}
			}
		})
		a.line(c)
		ui.Row(c).Height(HeaderHeight).PaddingX(t.Space(16)).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.text("app.name", "Orbit")).FontSize(t.Px(12)).TextColor(t.TextMuted)
			if a.Options.Version != "" {
				ui.Text(c, "v"+a.Options.Version).Font(a.Fonts.CodeFamily).FontSize(t.Px(12)).TextColor(t.TextMuted)
			}
		})
	})
	if strings.TrimSpace(a.settingsSearch) != "" {
		a.settingsSearchResults(c)
		return
	}
	if a.settingsPage == "bug" && a.Options.BugReport != nil {
		a.Options.BugReport(c)
		return
	}
	if a.settingsPage == "mcp" && a.Options.MCP != nil {
		a.Options.MCP(c)
		return
	}
	if (a.settingsPage == "skills" || a.settingsPage == "plugins") && a.Options.Resources != nil {
		a.Options.Resources(c, a.settingsPage)
		return
	}
	scroll := a.settingsScroll[a.settingsPage]
	if scroll == nil {
		scroll = &ui.ScrollState{}
		a.settingsScroll[a.settingsPage] = scroll
	}
	ui.Column(c.Key("settings-main")).Grow(1).FillHeight().MinWidth(0).LineHeight(sourceDefaultLineHeight).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).AlignSelf(ui.Center).Children(func() {
			ui.Column(c).AlignSelf(ui.Stretch).Padding(t.Space(20), t.Space(24), t.Space(12)).Gap(t.Space(4)).Children(func() {
				if a.settingsPage == "about" {
					ui.Image(c, logos["app-icon.png"]).Size(48, 48)
				}
				for _, cat := range settingsCategories {
					if cat.ID == a.settingsPage {
						ui.Text(c, a.settingsTitle(cat.ID, cat.Title)).Label("Settings page title").FontSize(t.Px(16)).LineHeight(1.3).Height(sourceTextHeight(t.Px(16), 1.3)).FontWeight(500)
						ui.Text(c, a.settingsDescription(cat.ID, cat.Description)).FontSize(t.Px(12)).Height(sourceTextHeight(t.Px(12), sourceDefaultLineHeight)).TextColor(t.TextMuted)
						break
					}
				}
				if a.settingsPage == "providers" {
					ui.Box(c).Height(a.Theme.Space(8))
					a.providerToolbar(c)
				}
				if a.settingsPage == "models" && a.Models != nil {
					ui.Box(c).Height(a.Theme.Space(8))
					a.Models.Toolbar(c)
				}
			})
			ui.Box(c).AlignSelf(ui.Stretch).Height(1).MarginX(t.Space(24)).Background(t.Border)
		})
		ui.Scroll(c.Key("settings-body:" + a.settingsPage)).TrackScroll(scroll).Grow(1).FillHeight().MinHeight(0).AlignSelf(ui.Stretch).Children(func() {
			ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).AlignSelf(ui.Center).Padding(t.Space(20), t.Space(24), t.Space(24)).Gap(t.Space(20)).Children(func() {
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
				case "models":
					a.modelsSettings(c)
				case "providers":
					a.providersSettings(c)
				case "privacy":
					a.privacySettings(c)
				case "about":
					a.aboutSettings(c)
				case "bug":
					a.bugSettings(c)
				default:
					a.Empty(c, "Service unavailable", "The native service has not been mounted.")
				}
			})
		})
	})
}
func (a *App) settingsDescription(id, fallback string) string {
	key := "settings." + id + "_description"
	if id == "worktrees" {
		key = "worktree.settings.description"
	}
	if id == "bug" {
		key = "settings.report_bug_description"
	}
	return a.text(key, fallback)
}
func (a *App) selectSettingsPage(id string) {
	a.settingsPage = id
	a.nativeSettings.selectMenu.id = ""
	a.providerState.editor = nil
	a.secretOpen = false
	a.secretValue = ""
	a.loadSettings(id)
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
func (a *App) modelsSettings(c *ui.Context) {
	if a.Models != nil {
		a.Models.View(c)
	}
}
func (a *App) loadSettings(category string) {
	if category == "mcp" && a.Options.RefreshMCP != nil {
		a.Options.RefreshMCP()
		return
	}
	if (category == "skills" || category == "plugins") && a.Options.RefreshResources != nil {
		a.Options.RefreshResources(category)
		return
	}
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
