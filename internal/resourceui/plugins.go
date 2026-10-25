package resourceui

import (
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func (s *State) pluginsView(c *ui.Context) ui.Element {
	a := s.appearance(c)
	root := ui.Column(c).Grow(1).MinWidth(0).MinHeight(0).AlignSelf(ui.Stretch).Background(a.native.Background)
	root.Children(func() {
		ui.Column(c.Key("pinned-header")).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(960).Shrink(0).Padding(a.Space(20), a.Space(24), a.Space(12)).Gap(a.Space(12)).Children(func() {
			ui.Column(c).Gap(a.Space(4)).Children(func() {
				ui.Text(c, s.label("settings.plugins", "Plugins")).FontSize(a.Px(16)).FontWeight(500)
				ui.Text(c, s.label("settings.plugins_description", "pi packages for this machine and this project. pi loads them at startup. Restart pi to apply changes.")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			})
			ui.Row(c.Key("filter-row")).Gap(a.Space(8)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).Children(func() {
				icon(c, "search", a.Px(14), a.native.TextMuted)
				ui.TextInput(c.Key("filter"), &s.PluginQuery).Placeholder(s.label("app.search_installed_plugins", "Search installed plugins…")).Label(s.label("resources.search_plugins", "Search installed plugins")).Grow(1).Height(a.Px(32)).Border(0, a.native.Border).Background(a.native.Background)
			})
			ui.Row(c.Key("install-row")).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
				ui.TextInput(c.Key("source"), &s.InstallSource).Placeholder(s.label("app.plugin_source_placeholder", "npm:@scope/pkg, git:github.com/owner/repo, or ./path")).Label(s.label("resources.package_source", "Package source")).Grow(1).MinWidth(32).Height(a.Px(36)).Disabled(s.Busy)
				ui.Row(c.Key("scope")).Gap(a.Space(4)).Shrink(0).Children(func() {
					for _, scope := range []string{"global", "project"} {
						label := s.label("settings.global", "Global")
						if scope == "project" {
							label = s.label("settings.project", "Project")
						}
						b := s.button(c, "scope-"+scope, label, "", false).Height(a.Px(32)).Disabled(s.Busy || (scope == "project" && s.Workspace == ""))
						if s.InstallScope == scope {
							b.Background(a.native.SurfacePressed)
						}
						if scope == "project" && s.Workspace == "" {
							b.Tooltip(s.label("resources.select_workspace", "Select a workspace to install project plugins"))
						}
						if b.Clicked() {
							s.InstallScope = scope
						}
					}
				})
				if s.button(c, "install", s.label("settings.install", "Install"), "plus", true).Height(a.Px(32)).Disabled(s.Busy || strings.TrimSpace(s.InstallSource) == "").Clicked() {
					s.Install()
				}
			})
			ui.Row(c.Key("footer")).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					s.statusLine(c, s.label("settings.n_packages", "%{count} package(s) configured", "count", fmt.Sprint(len(s.Plugins))))
				})
				if s.button(c, "refresh", s.label("settings.refresh", "Refresh"), "refresh", false).Height(a.Px(32)).Disabled(s.Busy).Clicked() {
					s.RefreshPlugins()
				}
			})
		})
		ui.Box(c).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(912).Height(1).Shrink(0).Background(a.native.Border)
		ui.Scroll(c.Key("plugin-scroll")).TrackScroll(&s.pluginScroll).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c.Key("plugin-content")).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(960).Padding(a.Space(20), a.Space(24), a.Space(24)).Gap(a.Space(16)).Children(func() {
				for i, e := range s.PluginErrors {
					s.notice(c, fmt.Sprintf("read-error-%d", i), s.label("settings.could_not_read_settings", "Could not read settings"), e, a.native.Danger)
				}
				if s.Error != "" {
					s.notice(c, "action-error", s.label("resources.plugin_failed", "Plugin action failed"), s.Error, a.native.Danger)
				}
				for i, e := range s.UpdateErrors {
					s.notice(c, fmt.Sprintf("update-error-%d", i), s.label("resources.update_check_failed", "Could not check a plugin"), e, a.native.Warning)
				}
				visible := s.visiblePlugins()
				if len(s.Plugins) == 0 {
					s.pluginEmpty(c, "extensions", s.label("settings.no_plugins_installed", "No plugins installed"), s.label("settings.no_plugins_installed_hint", "Install a pi package using an npm, Git, or local source above."))
				} else if len(visible) == 0 {
					s.pluginEmpty(c, "search", s.label("settings.no_plugins_match", "No plugins match"), s.label("settings.no_plugins_match_hint", "Try a different search."))
				} else {
					if strings.TrimSpace(s.PluginQuery) != "" {
						installed := 0
						for _, p := range s.Plugins {
							if p.Installed {
								installed++
							}
						}
						ui.Text(c, s.label("settings.plugins_shown_installed", "%{shown} shown · %{total} configured · %{installed} installed", "shown", fmt.Sprint(len(visible)), "total", fmt.Sprint(len(s.Plugins)), "installed", fmt.Sprint(installed))).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
					}
					ui.Column(c.Key("plugin-group")).Radius(a.Px(12)).Border(1, a.native.Border).Children(func() {
						for i, p := range visible {
							if i > 0 {
								ui.Box(c).Height(1).Background(a.native.Border).Margin(0, a.Space(16))
							}
							s.pluginRow(c, p)
						}
					})
				}
				if s.Status != "" && s.Options.Notify == nil {
					ui.Text(c, s.Status).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Selectable()
				}
			})
		})
	})
	return root
}

func (s *State) pluginEmpty(c *ui.Context, glyph, title, hint string) {
	a := s.appearance(c)
	ui.Column(c.Key("empty")).Padding(a.Space(32)).Gap(a.Space(8)).Radius(a.Px(12)).Border(1, a.native.Border).AlignItems(ui.Center).Children(func() {
		if s.Busy {
			ui.Spinner(c).Size(a.Px(22), a.Px(22))
		} else {
			icon(c, glyph, a.Px(26), a.native.TextMuted)
		}
		ui.Text(c, title).FontSize(a.Px(14)).FontWeight(500)
		ui.Text(c, hint).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
	})
}
func (s *State) pluginRow(c *ui.Context, p workbench.Plugin) {
	c = c.Key(pluginKey(p))
	a := s.appearance(c)
	ui.Column(c).Padding(a.Space(12), a.Space(16)).Gap(a.Space(8)).Children(func() {
		ui.Row(c.Key("main")).Gap(a.Space(12)).AlignItems(ui.Center).Wrap().Children(func() {
			ui.Box(c).Size(28, 28).Shrink(0).Radius(a.Px(8)).Background(a.Elevated).Center().Children(func() { icon(c, "extensions", a.Px(14), a.native.TextMuted) })
			ui.Column(c).Grow(1).MinWidth(120).Gap(a.Space(2)).Children(func() {
				ui.Text(c, p.Name).FontSize(a.Px(14)).FontWeight(500).SingleLine().Tooltip(p.Name)
				ui.Text(c, p.Source).Font(a.CodeFont).FontSize(a.Px(10.5)).TextColor(a.native.TextMuted).SingleLine().Tooltip(p.Source)
			})
			ui.Row(c.Key("actions")).Gap(a.Space(8)).Shrink(0).Wrap().Children(func() {
				if s.pluginRemove == pluginKey(p) {
					ui.Text(c, s.label("settings.remove_this_plugin", "Remove this plugin?")).FontSize(a.Px(12)).TextColor(a.native.Danger)
					if s.button(c, "confirm-remove", s.label("resources.confirm_remove", "Confirm remove"), "trash", false).Disabled(s.Busy).Clicked() {
						s.ConfirmRemovePlugin()
					}
					if s.button(c, "cancel-remove", s.label("git_panel.cancel", "Cancel"), "", false).Disabled(s.Busy).Clicked() {
						s.Escape()
					}
				} else {
					if p.Installed {
						if s.button(c, "update", s.label("settings.update", "Update"), "refresh", false).Disabled(s.Busy).Clicked() {
							s.UpdatePlugin(pluginKey(p))
						}
					}
					if s.button(c, "remove", s.label("settings.remove", "Remove"), "trash", false).Disabled(s.Busy).Clicked() {
						s.AskRemovePlugin(pluginKey(p))
					}
				}
			})
		})
		ui.Row(c.Key("badges")).Padding(0, 0, 0, a.Space(40)).Gap(a.Space(6)).Wrap().Children(func() {
			scope := s.label("settings.global", "Global")
			if p.Scope == "project" {
				scope = s.label("settings.project", "Project")
			}
			s.chip(c, "scope", scope, a.native.TextMuted)
			s.chip(c, "kind", p.Kind, a.native.TextMuted)
			if p.Version != "" {
				s.chip(c, "version", "v"+p.Version, a.native.TextMuted)
			}
			if !p.Installed {
				s.chip(c, "not-installed", s.label("settings.not_installed", "Not installed"), a.native.Danger)
			}
		})
		if p.InstallPath != "" {
			ui.Text(c, s.displayPath(p.InstallPath)).Padding(0, 0, 0, a.Space(40)).Font(a.CodeFont).FontSize(a.Px(10.5)).TextColor(a.native.TextMuted).SingleLine().Tooltip(p.InstallPath)
		}
		if version := s.Skipped[p.Source]; version != "" {
			ui.Row(c.Key("skipped")).Padding(0, 0, 0, a.Space(40)).Gap(a.Space(8)).Wrap().AlignItems(ui.Center).Children(func() {
				ui.Text(c, s.label("resources.skipped_version", "Skipped %{version}", "version", version)).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				if s.button(c, "reset-skip", s.label("resources.reset_skip", "Reset skip"), "", false).Disabled(s.Busy).Clicked() {
					s.SkipUpdate(p.Source, "")
				}
			})
		}
	})
}
