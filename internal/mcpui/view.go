package mcpui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type appearance struct {
	native *ui.Theme
	Theme
}

func (s *State) appearance(c *ui.Context) appearance {
	a := appearance{native: c.Theme()}
	if s.Options.Theme != nil {
		a.Theme = s.Options.Theme()
	}
	if a.Px == nil {
		a.Px = func(n float32) float32 { return n * a.native.FontSize / 14 }
	}
	if a.Space == nil {
		a.Space = func(n int) float32 { return a.Px(float32(n)) }
	}
	if a.Elevated == (ui.Color{}) {
		a.Elevated = a.native.Surface
	}
	if a.Secondary == (ui.Color{}) {
		a.Secondary = a.native.Text.Alpha(.82)
	}
	if a.Action == (ui.Color{}) {
		a.Action = a.native.Accent
	}
	if a.ActionText == (ui.Color{}) {
		a.ActionText = a.native.AccentText
	}
	return a
}

var iconOnce sync.Once
var mcpIcons map[string]*ui.SVG

func icon(c *ui.Context, name string, size float32, color ui.Color) ui.Element {
	iconOnce.Do(func() {
		mcpIcons = map[string]*ui.SVG{}
		for _, name := range []string{"search", "refresh", "plus", "tools/mcp"} {
			if data, err := orbitassets.Asset("icons/" + name + ".svg"); err == nil {
				if svg, err := ui.ParseSVG(data); err == nil {
					mcpIcons[name] = svg
				}
			}
		}
	})
	if svg := mcpIcons[name]; svg != nil {
		return ui.Icon(c, svg).Size(size, size).Shrink(0).TextColor(color)
	}
	return ui.Element{}
}

func (s *State) button(c *ui.Context, key, label string, primary bool) ui.Element {
	a := s.appearance(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(a.Px(28)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Px(4)).Radius(a.Px(8)).AlignItems(ui.Center).Background(a.Elevated).Border(1, a.native.Border)
	if primary {
		b.Background(a.Action).Border(1, a.Action)
	}
	b.Children(func() {
		color := a.Secondary
		if primary {
			color = a.ActionText
		}
		if key == "add" {
			icon(c, "plus", a.Px(16), color)
		}
		ui.Text(c, label).FontSize(a.Px(14)).LineHeight(1).FontWeight(500).TextColor(color).SingleLine()
	})
	return b
}
func (s *State) iconButton(c *ui.Context, key, glyph, label string) ui.Element {
	a := s.appearance(c)
	b := ui.ButtonBase(c.Key(key)).Size(a.Px(26), a.Px(26)).Center().Radius(a.Px(5)).Label(label).Tooltip(label)
	b.Children(func() {
		if glyph == "↻" {
			icon(c, "refresh", a.Px(16), a.Secondary)
		} else {
			ui.Text(c, glyph).FontSize(a.Px(14)).TextColor(a.native.TextMuted)
		}
	})
	return b
}
func (s *State) chip(c *ui.Context, key, label string, color ui.Color) {
	a := s.appearance(c)
	ui.Row(c.Key(key)).Padding(a.Space(2), a.Space(6)).Radius(a.Px(4)).Background(color.Alpha(.09)).Children(func() { ui.Text(c, label).FontSize(a.Px(10)).TextColor(color).SingleLine() })
}
func (s *State) notice(c *ui.Context, key, title, detail string, color ui.Color, action func()) {
	a := s.appearance(c)
	ui.Column(c.Key(key)).Padding(a.Space(12)).Gap(a.Space(4)).Radius(a.Px(8)).Background(color.Alpha(.06)).Border(1, color.Alpha(.25)).Children(func() {
		ui.Text(c, title).FontSize(a.Px(13)).FontWeight(600).TextColor(color)
		ui.Text(c, detail).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Selectable()
		if action != nil {
			action()
		}
	})
}

// View mounts directly in the Settings content slot, outside another Scroll.
func (s *State) View(c *ui.Context) ui.Element {
	s.lastViewed = time.Now()
	if !s.loaded && !s.Busy && !s.closed {
		s.Refresh()
	}
	c = c.Key("mcp-settings")
	a := s.appearance(c)
	root := ui.Column(c).Grow(1).MinHeight(0).MinWidth(0).AlignSelf(ui.Stretch).LineHeight(1.618034)
	root.Children(func() {
		ui.Column(c.Key("pinned-header")).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(960).Shrink(0).Padding(a.Space(20), a.Space(24), a.Space(12)).Gap(a.Space(12)).Children(func() {
			ui.Column(c.Key("title")).Gap(a.Space(4)).Children(func() {
				ui.Text(c, s.label("settings.mcp", "MCP")).FontSize(a.Px(16)).LineHeight(1.3).FontWeight(500)
				ui.Text(c, s.label("settings.mcp_description", "Model Context Protocol servers Orbit manages for the agent.")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			})
			ui.Column(c.Key("toolbar-and-status")).Gap(a.Space(8)).Children(func() {
				ui.Row(c.Key("toolbar")).Shrink(0).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
					ui.Row(c.Key("search-row")).Grow(1).MinWidth(40).Height(a.Px(36)).Shrink(0).Padding(0, a.Px(10)).Gap(a.Px(4)).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).AlignItems(ui.Center).Children(func() {
						icon(c, "search", a.Px(14), a.native.TextMuted)
						ui.TextInput(c.Key("filter"), &s.Query).Placeholder(s.label("mcp.search_servers", "Search MCP servers…")).Label(s.label("mcp.search_servers", "Search MCP servers")).Grow(1).MinWidth(0).Height(a.Px(36)).Padding(0).Border(0, ui.Color{}).Background(ui.Color{}).FocusRing(false)
					})
					if s.iconButton(c, "refresh", "↻", s.label("mcp.refresh", "Refresh")).Size(a.Px(32), a.Px(32)).Background(a.Elevated).Border(1, a.native.Border).Disabled(s.Busy).Clicked() {
						s.Refresh()
					}
					if s.button(c, "add", s.label("mcp.add_server", "Add server"), true).Height(a.Px(32)).Disabled(s.Busy && !s.Probing).Clicked() {
						s.BeginAdd()
					}
				})
				s.statusLine(c)
			})
		})
		ui.Box(c).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(912).Height(1).Shrink(0).Background(a.native.Border)
		ui.Scroll(c.Key("body")).Grow(1).MinHeight(0).Children(func() {
			ui.Column(c).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(960).Padding(a.Space(20), a.Space(24), a.Space(24)).Gap(a.Space(20)).Children(func() {
				for i, err := range s.Errors {
					s.notice(c, fmt.Sprintf("config-error-%d", i), s.label("mcp.config_error", "Configuration error"), err, a.native.Danger, nil)
				}
				for i, err := range s.probeErrors {
					s.notice(c, fmt.Sprintf("probe-error-%d", i), s.label("mcp.probe_error", "Connection error"), err, a.native.Danger, nil)
				}
				if s.Error != "" {
					s.notice(c, "action-error", s.label("mcp.action_failed", "MCP action failed"), s.Error, a.native.Danger, nil)
				}
				if s.probeNote != "" {
					s.notice(c, "project-note", s.label("mcp.project_untrusted", "Project configuration"), s.probeNote, a.native.Warning, nil)
				}
				if s.codemode != nil && !*s.codemode && s.hasCodemode() {
					s.notice(c, "codemode-disabled", s.label("mcp.codemode_disabled_title", "Codemode is disabled"), s.label("mcp.codemode_disabled_body", "Some servers use codemode exposure, but autoEnableCodemode is disabled in Pi's configuration. Their tools may not be available to the agent."), a.native.Warning, nil)
				}
				if s.externalChange {
					s.notice(c, "external-change", s.label("mcp.external_change_title", "Configuration changed outside Orbit"), s.label("mcp.external_change_body", "The list now reflects the files on disk. Reload Pi to apply these changes to the current session."), a.native.Warning, func() {
						if s.button(c, "reload", s.label("mcp.reload", "Reload Pi"), false).Disabled(s.Options.OnApply == nil || s.Busy).Clicked() {
							s.apply()
						}
					})
				} else if s.applyPending {
					s.notice(c, "apply-pending", s.label("mcp.restart_title", "Apply MCP changes"), s.label("mcp.restart_body", "The current Pi session must reload to use the saved configuration. An active turn must finish before it can restart."), a.native.Warning, func() {
						if s.button(c, "apply", s.label("mcp.apply_now", "Apply now"), false).Disabled(s.Options.OnApply == nil || s.Busy).Clicked() {
							s.apply()
						}
					})
				}
				visible := s.visibleServers()
				if len(s.Servers) == 0 {
					s.emptyState(c)
				} else if len(visible) == 0 {
					ui.Column(c.Key("no-matches")).Padding(a.Space(12), a.Space(16)).Gap(a.Space(3)).Children(func() {
						ui.Text(c, s.label("mcp.no_matches", "No matching servers")).FontSize(a.Px(14)).FontWeight(500)
						ui.Text(c, s.label("mcp.no_matches_hint", "Try a server name, a command, or a URL.")).FontSize(a.Px(12)).TextColor(a.Secondary)
					})
				} else {
					for _, scope := range []string{"global", "project"} {
						var scoped []workbench.MCPServer
						for _, server := range visible {
							if server.Scope == scope {
								scoped = append(scoped, server)
							}
						}
						if len(scoped) == 0 {
							continue
						}
						ui.Column(c.Key("scope-group-" + scope)).Gap(a.Space(8)).Children(func() {
							ui.Text(c, s.scopeLabel(scope)).FontSize(a.Px(12)).FontWeight(500).TextColor(a.native.TextMuted)
							ui.Column(c.Key("servers")).Radius(a.Px(12)).Border(1, a.native.Border).Children(func() {
								for i, server := range scoped {
									s.serverRow(c.Key(server.Name), server, i > 0)
								}
							})
						})
					}
				}
				if s.Status != "" {
					ui.Text(c, s.Status).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				}
			})
		})
	})
	if !s.HasModal() {
		if root.Shortcut(0, ui.KeyDown) {
			s.moveSelection(1)
		}
		if root.Shortcut(0, ui.KeyUp) {
			s.moveSelection(-1)
		}
		if root.Shortcut(0, ui.KeyEnter) && s.Selected != "" {
			s.BeginEdit(s.Selected)
		}
		if root.Shortcut(0, ui.KeySpace) && s.Selected != "" {
			if server := s.server(s.Selected); server != nil {
				s.SetEnabled(server.Name, !enabled(server.Definition))
			}
		}
	}
	s.editorView(c)
	s.secretView(c)
	if s.confirm.open {
		if ui.AlertDialog(c.Key("confirm"), &s.confirm.open, s.confirm.title, s.confirm.detail, s.label("mcp.cancel", "Cancel"), s.label("mcp.confirm", "Confirm")) == 1 {
			accept := s.confirm.accept
			s.confirm = confirmation{}
			if accept != nil {
				accept()
			}
		} else if !s.confirm.open {
			s.confirm = confirmation{}
		}
	}
	return root
}
func (s *State) hasCodemode() bool {
	for _, server := range s.Servers {
		if enabled(server.Definition) && strings.HasPrefix(exposure(server.Definition), "codemode") {
			return true
		}
	}
	return false
}
func (s *State) visibleServers() []workbench.MCPServer {
	q := strings.ToLower(strings.TrimSpace(s.Query))
	out := []workbench.MCPServer{}
	for _, server := range s.Servers {
		haystack := server.Name + " " + server.Description + " " + str(server.Definition["command"]) + " " + str(server.Definition["url"]) + " " + server.Scope
		if q == "" || strings.Contains(strings.ToLower(haystack), q) {
			out = append(out, server)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Name < out[j].Name
	})
	return out
}
func (s *State) moveSelection(delta int) {
	visible := s.visibleServers()
	if len(visible) == 0 {
		s.Selected = ""
		return
	}
	i := -1
	for n, server := range visible {
		if server.Name == s.Selected {
			i = n
			break
		}
	}
	i = max(0, min(len(visible)-1, i+delta))
	s.Selected = visible[i].Name
}

// Source mcp_empty_card is deliberately unboxed; unlike Models and Plugins,
// it has no filled card or 64-DIP tile. Loading omits the empty claim and CTA.
func (s *State) emptyState(c *ui.Context) {
	a := s.appearance(c)
	loading := s.Probing || (!s.loaded && s.Busy)
	ui.Column(c.Key("empty")).Label("MCP empty state").AlignSelf(ui.Stretch).Padding(a.Space(24), a.Space(16)).Gap(a.Space(6)).AlignItems(ui.Center).Children(func() {
		if loading {
			ui.Spinner(c).Size(a.Px(48), a.Px(48))
			ui.Text(c, s.label("mcp.checking", "Checking connections…")).FontSize(a.Px(14)).FontWeight(500)
			return
		}
		icon(c, "tools/mcp", a.Px(48), a.native.TextMuted).Label("MCP empty glyph")
		ui.Text(c, s.label("mcp.empty_title", "No MCP servers configured.")).FontSize(a.Px(14)).FontWeight(500)
		ui.Text(c, s.label("mcp.empty_hint", "Connect an MCP server to give your agent additional tools.")).FontSize(a.Px(12)).TextColor(a.Secondary)
		if s.button(c, "empty-add", s.label("mcp.add_server", "Add server"), true).Disabled(s.Busy && !s.Probing).Clicked() {
			s.BeginAdd()
		}
	})
}
func (s *State) statusLine(c *ui.Context) {
	a := s.appearance(c)
	status := s.label("mcp.not_checked", "Not checked yet")
	if s.Probing || (!s.loaded && s.Busy) {
		status = s.label("mcp.checking", "Checking connections…")
	} else if !s.checkedAt.IsZero() {
		status = s.label("mcp.checked_ago", "Checked %{seconds}s ago", "seconds", fmt.Sprint(max(1, int(time.Since(s.checkedAt).Seconds()))))
	} else if len(s.Servers) == 0 {
		status = s.label("mcp.no_servers", "No MCP servers configured.")
	}
	ui.Row(c.Key("status-line")).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
		ui.Text(c, status).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine().Shrink(0)
		if len(s.Servers) > 0 {
			ui.Text(c, s.label("mcp.server_count", "%{count} server(s)", "count", fmt.Sprint(len(s.Servers)))).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Grow(1).MinWidth(0).SingleLine()
			ui.Text(c, s.label("mcp.keys_hint", "↑↓ Navigate · ⏎ Edit · Space Enable/Disable")).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine().MinWidth(0)
		}
		path := s.service.Preferences.PiDir
		ui.Text(c, path).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine().MinWidth(0).Shrink(1).Tooltip(filepath.Join(path, "mcp.json"))
		if s.Workspace != "" {
			path := filepath.Join(s.Workspace, ".pi")
			ui.Text(c, s.label("mcp.project_dir_label", "· project: %{path}", "path", path)).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine().MinWidth(0).Shrink(1).Tooltip(filepath.Join(path, "mcp.json"))
		}
	})
}
func (s *State) serverRow(c *ui.Context, server workbench.MCPServer, divider bool) {
	a := s.appearance(c)
	runtime := s.Runtime[server.Name]
	selected := s.Selected == server.Name
	row := ui.Column(c).Padding(a.Space(12), a.Space(16)).Gap(a.Space(8))
	if divider {
		row.BorderWidth(1, 0, 0, 0).BorderColor(a.native.Border)
	}
	if selected {
		row.Background(a.native.Selection.Alpha(.35))
	}
	row.Children(func() {
		ui.Row(c).Gap(a.Space(12)).AlignItems(ui.Start).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(a.Space(4)).Children(func() {
				ui.Row(c).Wrap().Gap(a.Space(6)).AlignItems(ui.Center).Children(func() {
					if ui.ButtonBase(c.Key("select")).Label(s.label("mcp.details", "Server details") + ": " + server.Name).Children(func() { ui.Text(c, server.Name).FontSize(a.Px(13)).FontWeight(600).TextColor(a.native.Text) }).Clicked() {
						if selected {
							s.Selected = ""
						} else {
							s.Selected = server.Name
						}
					}
					s.chip(c, "scope", s.scopeLabel(server.Scope), a.native.TextMuted)
					s.chip(c, "exposure", exposure(server.Definition), a.native.TextMuted)
				})
				if server.Description != "" {
					ui.Text(c, server.Description).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				}
				ui.Text(c, transportSummary(server)).FontSize(a.Px(12)).TextColor(a.native.TextMuted).MaxLines(2)
				if runtime.State == "connected" {
					ui.Text(c, fmt.Sprintf("%d %s · %d %s", len(runtime.Tools), s.label("mcp.tools", "tools"), runtime.Resources, s.label("mcp.resources", "resources"))).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
				}
				if runtime.Error != "" {
					ui.Text(c, runtime.Error).FontSize(a.Px(12)).TextColor(a.native.Danger).Selectable()
					if hint := s.errorHint(runtime.Error); hint != "" {
						ui.Text(c, hint).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
					}
				}
			})
			ui.Column(c).Shrink(0).Gap(a.Space(8)).AlignItems(ui.End).Children(func() {
				label, color := s.runtimeLabel(runtime.State, a)
				s.chip(c, "status", label, color)
				if s.removeTarget == server.Name {
					ui.Row(c).Gap(a.Space(4)).Children(func() {
						if s.button(c, "remove-cancel", s.label("mcp.cancel", "Cancel"), false).Clicked() {
							s.removeTarget = ""
						}
						if s.button(c, "remove-confirm", s.label("mcp.remove", "Remove"), true).Disabled(s.Busy && !s.Probing).Clicked() {
							s.remove(server.Name)
						}
					})
				} else if s.authTarget == server.Name {
					ui.Text(c, s.label("mcp.waiting_for_browser", "Waiting for browser…")).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
					if s.button(c, "cancel-auth", s.label("mcp.cancel", "Cancel"), false).Clicked() {
						if s.cancel != nil {
							s.cancel()
						}
					}
				} else {
					ui.Row(c).Gap(a.Space(4)).AlignItems(ui.Center).Children(func() {
						on := enabled(server.Definition)
						if ui.Switch(c.Key("enabled"), &on).Label(s.label("mcp.enabled", "Enabled") + ": " + server.Name).Tooltip(s.label("mcp.enabled", "Enabled")).Disabled(s.Busy && !s.Probing).Changed() {
							s.SetEnabled(server.Name, on)
						}
						if s.iconButton(c, "test", "↻", s.label("mcp.test_connection", "Test connection")).Disabled(s.Busy).Clicked() {
							s.Probe()
						}
						if s.iconButton(c, "edit", "✎", s.label("mcp.edit", "Edit")).Disabled(s.Busy && !s.Probing).Clicked() {
							s.BeginEdit(server.Name)
						}
						if s.iconButton(c, "delete", "×", s.label("mcp.delete", "Delete")).Disabled(s.Busy && !s.Probing).Clicked() {
							s.removeTarget = server.Name
							s.Selected = server.Name
						}
					})
					if runtime.State == "needs-auth" && oauthCandidate(server) {
						if s.button(c, "sign-in", s.label("mcp.sign_in", "Sign in"), true).Disabled(s.Busy && !s.Probing).Clicked() {
							s.auth(server.Name, true)
						}
					}
				}
			})
		})
		if s.removeTarget == server.Name {
			ui.Text(c, s.label("mcp.remove_warning", "Remove this server from its configuration file? Unreferenced Orbit-generated secrets for this server will also be removed.")).FontSize(a.Px(12)).TextColor(a.native.Warning)
		}
		if selected {
			s.serverDetail(c, server, runtime)
		}
	})
}
func (s *State) serverDetail(c *ui.Context, server workbench.MCPServer, runtime Runtime) {
	a := s.appearance(c)
	ui.Column(c.Key("detail")).Gap(a.Space(8)).Children(func() {
		ui.Text(c, server.Source).FontSize(a.Px(11)).TextColor(a.native.TextMuted).Selectable()
		if len(runtime.Tools) > 0 {
			ui.Text(c, s.label("mcp.tools", "Tools")).FontSize(a.Px(12)).FontWeight(600)
			for _, tool := range runtime.Tools {
				ui.Column(c.Key(tool.Name)).Padding(a.Space(6)).Gap(a.Space(4)).Radius(a.Px(5)).Background(a.native.Surface).Children(func() {
					ui.Row(c).Wrap().Gap(a.Space(6)).Children(func() {
						ui.Text(c, tool.Name).Font(a.CodeFont).FontSize(a.Px(11)).Selectable()
						mode := tool.Exposure
						if mode == "" {
							mode = exposure(server.Definition)
						}
						s.chip(c, "mode", mode, a.native.TextMuted)
					})
					if tool.Description != "" {
						ui.Text(c, tool.Description).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
					}
					if len(tool.Schema) > 0 {
						ui.Text(c, string(tool.Schema)).Font(a.CodeFont).FontSize(a.Px(11)).TextColor(a.native.TextMuted).Selectable()
					}
				})
			}
		} else if runtime.State != "connected" {
			ui.Text(c, s.label("mcp.tools_not_probed", "Tool details appear after Pi reports a successful connection.")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		}
		if runtime.ResourceTemplates > 0 {
			ui.Text(c, fmt.Sprintf("%d %s", runtime.ResourceTemplates, s.label("mcp.resource_templates", "resource templates"))).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		}
		ui.Row(c).Wrap().Gap(a.Space(6)).Children(func() {
			if s.button(c, "secret", s.label("mcp.set_secret", "Set secret…"), false).Disabled(s.Busy && !s.Probing).Clicked() {
				s.openSecret(server)
			}
			if oauthCandidate(server) {
				if s.button(c, "oauth-login", s.label("mcp.sign_in", "Sign in"), false).Disabled(s.Busy && !s.Probing).Clicked() {
					s.auth(server.Name, true)
				}
				if s.button(c, "oauth-logout", s.label("mcp.sign_out", "Sign out"), false).Disabled(s.Busy && !s.Probing).Clicked() {
					name := server.Name
					s.confirm = confirmation{title: s.label("mcp.sign_out", "Sign out") + " " + name + "?", detail: s.label("mcp.sign_out_detail", "Pi will delete this server's saved OAuth credentials. Reload the current session afterward to drop its existing access."), open: true, accept: func() { s.auth(name, false) }}
				}
			}
		})
	})
}
func transportSummary(server workbench.MCPServer) string {
	if url := str(server.Definition["url"]); url != "" {
		return "HTTP · " + url
	}
	parts := []string{str(server.Definition["command"])}
	args := array(server.Definition["args"])
	for i, arg := range args {
		if i == 3 {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, str(arg))
	}
	return "stdio · " + strings.Join(parts, " ")
}
func (s *State) scopeLabel(scope string) string {
	if scope == "project" {
		return s.label("mcp.scope_project", "Project")
	}
	return s.label("mcp.scope_global", "Global")
}
func (s *State) runtimeLabel(state string, a appearance) (string, ui.Color) {
	switch state {
	case "connected":
		return "● " + s.label("mcp.status_connected", "Connected"), a.native.Success
	case "failed":
		return "! " + s.label("mcp.status_failed", "Failed"), a.native.Danger
	case "needs-auth":
		return "! " + s.label("mcp.status_needs_auth", "Sign-in required"), a.native.Warning
	case "connecting":
		return "◌ " + s.label("mcp.status_connecting", "Connecting"), a.native.TextMuted
	case "disabled":
		return "○ " + s.label("mcp.disabled", "Disabled"), a.native.TextMuted
	case "disconnected":
		return "○ " + s.label("mcp.status_disconnected", "Disconnected"), a.native.TextMuted
	default:
		return "○ " + s.label("mcp.status_unknown", "Not checked"), a.native.TextMuted
	}
}
func (s *State) errorHint(text string) string {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "not found") || strings.Contains(lower, "enoent"):
		return s.label("mcp.hint_command_missing", "Check that the server command is installed and available on PATH.")
	case strings.Contains(lower, "401") || strings.Contains(lower, "unauthorized") || strings.Contains(lower, "authentication"):
		return s.label("mcp.hint_auth", "Check the server's credentials or sign in again.")
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline"):
		return s.label("mcp.hint_timeout", "The server did not respond in time. Check its command, URL and timeout.")
	case strings.Contains(lower, "connect") || strings.Contains(lower, "dns") || strings.Contains(lower, "network"):
		return s.label("mcp.hint_network", "Check the server URL and your network connection.")
	}
	return ""
}
