package mcpui

import (
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func (s *State) field(c *ui.Context, key, label, hint string, value *string, lines int, password, readOnly bool) {
	a := s.appearance(c)
	field := ui.Field(c.Key(key), label, func() {
		var input ui.Element
		if lines > 1 {
			input = ui.TextArea(c.Key("input"), value).Height(a.Px(float32(lines*18 + 12))).MinHeight(a.Px(48))
		} else {
			input = ui.TextInput(c.Key("input"), value).Height(a.Px(30))
		}
		input.Label(label).AlignSelf(ui.Stretch).FontSize(a.Px(13)).ReadOnly(readOnly).Disabled(s.Busy)
		if key == "name" {
			input.AutoFocus()
		}
		if password {
			input.Password()
		}
	}).Gap(a.Space(4)).FontSize(a.Px(12))
	if hint != "" {
		field.Description(hint)
	}
}
func (s *State) editorView(c *ui.Context) {
	if s.Editor == nil {
		return
	}
	d := s.Editor
	a := s.appearance(c)
	w, h := c.Size()
	title := s.label("mcp.add_server", "Add server")
	if d.original != nil {
		title = s.label("mcp.edit_server", "Edit server")
	}
	ui.DialogBase(c.Key("editor"), &s.editorOpen, func(back, panel ui.Element) {
		alpha := float32(.42)
		if !a.native.Dark {
			alpha = .22
		}
		back.Background(ui.RGBA(0, 0, 0, alpha))
		panel.Width(min(560, max(0, w-2*a.Space(24)))).MaxHeight(max(0, h-2*a.Space(24))).Padding(0).Gap(0).Background(a.Elevated).Radius(a.Px(12)).Border(1, a.native.Border).Children(func() {
			ui.Column(c.Key("header")).Padding(a.Space(8), a.Space(12), a.Space(4)).Gap(a.Space(8)).Shrink(0).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).Children(func() {
				ui.Text(c, title).FontSize(a.Px(16)).FontWeight(500)
				ui.Text(c, s.label("mcp.editor_subtitle", "The agent connects and calls the tools; Orbit manages the configuration.")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			})
			ui.Scroll(c.Key("fields")).Height(max(100, h-a.Px(154))).Shrink(1).Children(func() {
				ui.Column(c).Padding(a.Space(16), a.Space(12)).Gap(a.Space(16)).Children(func() {
					s.transportChoices(c, d)
					s.field(c, "name", s.label("mcp.name", "Name"), s.label("mcp.name_hint", "Letters, numbers, hyphens and underscores."), &d.Name, 1, false, d.original != nil)
					s.field(c, "description", s.label("mcp.description", "Description"), s.label("mcp.description_hint", "An optional note shown only in Orbit."), &d.Description, 1, false, false)
					if d.Transport == "stdio" {
						s.field(c, "command", s.label("mcp.command", "Command"), s.label("mcp.command_hint", "Executable to launch, for example npx or uvx. No shell is used."), &d.Command, 1, false, false)
						s.field(c, "args", s.label("mcp.arguments", "Arguments"), s.label("mcp.args_hint", "One argument per line."), &d.Args, 4, false, false)
						s.field(c, "env", s.label("mcp.environment", "Environment variables"), s.label("mcp.env_hint", "One NAME=value per line. Use ${NAME} references or enter a secret to store privately."), &d.Env, 5, false, false)
						s.field(c, "cwd", s.label("mcp.cwd", "Working directory"), s.label("mcp.cwd_hint", "Leave empty to use Pi's default working directory."), &d.Cwd, 1, false, false)
					} else {
						s.field(c, "url", s.label("mcp.url", "Server URL"), s.label("mcp.url_hint", "An absolute http:// or https:// streamable HTTP endpoint."), &d.URL, 1, false, false)
						s.field(c, "headers", s.label("mcp.headers", "Headers"), s.label("mcp.headers_hint", "One Header: value per line. Authorization: Bearer ${TOKEN} keeps secrets out of configuration."), &d.Headers, 5, false, false)
						if s.button(c, "oauth-toggle", s.label("mcp.oauth_settings", "OAuth client settings"), false).Clicked() {
							d.OAuthOpen = !d.OAuthOpen
						}
						ui.Text(c, s.label("mcp.oauth_settings_hint", "Optional. Most servers register Pi automatically and need none of these fields.")).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
						if d.OAuthOpen {
							ui.Column(c.Key("oauth")).Padding(a.Space(12)).Gap(a.Space(12)).Border(1, a.native.Border).Radius(a.Px(8)).Children(func() {
								s.field(c, "client-id", s.label("mcp.oauth_client_id", "Client ID"), "", &d.OAuthClientID, 1, false, false)
								s.field(c, "client-secret", s.label("mcp.oauth_client_secret", "Client secret"), s.label("mcp.oauth_client_secret_hint", "Use ${NAME}, or store a literal privately when saving."), &d.OAuthClientSecret, 1, true, false)
								s.field(c, "callback-port", s.label("mcp.oauth_callback_port", "Callback port"), s.label("mcp.oauth_callback_port_hint", "Optional port from 1 to 65535."), &d.OAuthCallbackPort, 1, false, false)
								s.field(c, "callback-url", s.label("mcp.oauth_callback_url", "Callback URL"), s.label("mcp.oauth_callback_url_hint", "Optional loopback HTTP URL, for example http://localhost:8080/callback."), &d.OAuthCallbackURL, 1, false, false)
								s.field(c, "oauth-scope", s.label("mcp.oauth_scope", "OAuth scope"), "", &d.OAuthScope, 1, false, false)
							})
						}
					}
					ui.Column(c.Key("exposure")).Gap(a.Space(4)).Children(func() {
						ui.Text(c, s.label("mcp.exposure", "Tool exposure")).FontSize(a.Px(12)).FontWeight(500)
						for _, mode := range workbench.MCPExposures {
							ui.Column(c.Key(mode)).Padding(a.Space(6), a.Space(8)).Radius(a.Px(6)).Background(func() ui.Color {
								if d.Exposure == mode {
									return a.native.Selection
								}
								return ui.Transparent
							}()).Children(func() {
								ui.Radio(c, &d.Exposure, mode, s.exposureLabel(mode)).Disabled(s.Busy)
								ui.Text(c, s.exposureHint(mode)).PaddingStart(a.Space(24)).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
							})
						}
					})
					ui.Column(c.Key("scope")).Gap(a.Space(4)).Children(func() {
						ui.Text(c, s.label("mcp.scope", "Scope")).FontSize(a.Px(12)).FontWeight(500)
						ui.Radio(c.Key("global"), &d.Scope, "global", s.scopeLabel("global")).Disabled(s.Busy)
						ui.Text(c, s.label("mcp.scope_global_hint", "Available to all workspaces through Pi's global configuration.")).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
						ui.Radio(c.Key("project"), &d.Scope, "project", s.scopeLabel("project")).Disabled(s.Busy || s.Workspace == "")
						ui.Text(c, s.label("mcp.scope_project_hint", "Saved in the current workspace's .pi/mcp.json. Project entries override global entries with the same name.")).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
					})
					ui.Checkbox(c.Key("enabled"), &d.Enabled, s.label("mcp.enabled", "Enabled")).Disabled(s.Busy)
					s.field(c, "timeout", s.label("mcp.timeout", "Request timeout (seconds)"), s.label("mcp.timeout_hint", "Leave empty to use Pi's default timeout."), &d.Timeout, 1, false, false)
				})
			})
			if d.Error != "" {
				message := d.Error
				if d.validationError {
					message = s.displayValidation(message)
				}
				ui.Text(c, message).FontSize(a.Px(12)).TextColor(a.native.Danger).MaxLines(3).Shrink(0)
			}
			ui.Row(c.Key("footer")).Padding(a.Space(8)).Gap(a.Space(4)).Shrink(0).Justify(ui.End).BorderWidth(1, 0, 0, 0).BorderColor(a.native.Border).Children(func() {
				if s.button(c, "cancel", s.label("mcp.cancel", "Cancel"), false).Clicked() {
					s.editorOpen = false
				}
				if s.button(c, "save", s.label("mcp.save", "Save"), true).Disabled(s.Busy).Clicked() {
					s.saveEditor()
				}
			})
		})
	})
	if !s.editorOpen {
		s.Editor = nil
	}
}
func (s *State) exposureHint(mode string) string {
	switch mode {
	case "codemode":
		return s.label("mcp.exposure_codemode_hint", "Callable from codemode scripts, not declared to the model.")
	case "codemode-deferred":
		return s.label("mcp.exposure_codemode_deferred_hint", "Like codemode, but not listed in the codemode description.")
	case "deferred":
		return s.label("mcp.exposure_deferred_hint", "The model loads the tools on demand with tool_search.")
	case "direct":
		return s.label("mcp.exposure_direct_hint", "Declared to the model like built-in tools.")
	case "hidden":
		return s.label("mcp.exposure_hidden_hint", "Registered but not callable.")
	}
	return ""
}
func (s *State) secretView(c *ui.Context) {
	if !s.secret.open {
		return
	}
	a := s.appearance(c)
	w, h := c.Size()
	ui.DialogBase(c.Key("secret-dialog"), &s.secret.open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(0, 0, 0, .45))
		panel.Width(min(a.Px(500), max(0, w-32))).MaxHeight(max(0, h-32)).Padding(a.Space(20)).Gap(a.Space(12)).Background(a.Elevated).Radius(a.Px(12)).Border(1, a.native.Border).Children(func() {
			ui.Text(c, s.label("mcp.set_secret", "Set secret…")).FontSize(a.Px(17)).FontWeight(600)
			ui.Text(c, fmt.Sprintf("%s · %s", s.scopeLabel(s.secret.scope), s.secret.server)).TextColor(a.native.TextMuted).FontSize(a.Px(12))
			s.field(c, "variable", s.label("mcp.secret_variable", "Variable name"), s.label("mcp.secret_reference_hint", "Use this exact ${NAME} reference in the server's environment, headers or OAuth secret."), &s.secret.name, 1, false, false)
			s.field(c, "value", s.label("mcp.secret_value", "Secret value"), "", &s.secret.value, 1, true, false)
			if len(s.secretNames) > 0 {
				ui.Text(c, s.label("mcp.saved_secret_names", "Saved variable names")+": "+strings.Join(s.secretNames, ", ")).MaxLines(3).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
			}
			if s.secret.error != "" {
				ui.Text(c, s.secret.error).TextColor(a.native.Danger).FontSize(a.Px(12))
			}
			ui.Row(c).Justify(ui.End).Gap(a.Space(8)).Children(func() {
				if s.button(c, "cancel-secret", s.label("mcp.cancel", "Cancel"), false).Clicked() {
					s.secret.open = false
				}
				if s.button(c, "store-secret", s.label("mcp.store_secret", "Store secret"), true).Disabled(s.Busy).Clicked() {
					s.saveSecret()
				}
			})
		})
	})
	if !s.secret.open {
		s.secret = secretPrompt{}
	}
}

func (s *State) transportChoices(c *ui.Context, d *Draft) {
	a := s.appearance(c)
	ui.Column(c.Key("transport")).Gap(a.Space(8)).Children(func() {
		ui.Text(c, s.label("mcp.transport", "Transport")).FontSize(a.Px(12)).FontWeight(500)
		ui.Row(c).Gap(a.Space(6)).Children(func() {
			for _, kind := range []string{"stdio", "http"} {
				label := s.label("mcp.transport_stdio", "Stdio")
				if kind == "http" {
					label = s.label("mcp.transport_http", "Streamable HTTP")
				}
				button := s.button(c, kind, label, false).Selected(d.Transport == kind).Disabled(s.Busy)
				if d.Transport == kind {
					button.Background(a.native.Accent.Alpha(.12)).BorderColor(a.native.Accent.Alpha(.5))
				}
				if button.Clicked() {
					d.Transport = kind
				}
			}
		})
		hint := s.label("mcp.stdio_hint", "This MCP server runs a local process on this computer.")
		if d.Transport == "http" {
			hint = s.label("mcp.http_hint", "This MCP server communicates with an external service.")
		}
		ui.Text(c, hint).FontSize(a.Px(11)).TextColor(a.native.TextMuted)
	})
}
func (s *State) exposureLabel(mode string) string {
	fallback := map[string]string{"codemode": "Codemode", "codemode-deferred": "Codemode (deferred)", "deferred": "Deferred", "direct": "Direct", "hidden": "Hidden"}[mode]
	return s.label("mcp.exposure_"+strings.ReplaceAll(mode, "-", "_"), fallback)
}
