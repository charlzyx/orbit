package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
)

// Settings-specific shell follows settings.rs and theme::modal; it deliberately
// does not alter the generic dialog metrics used by unrelated feature modules.
func (a *App) sourceProviderModal(c *ui.Context, open *bool, width, height float32, label string, body func(ui.Element)) {
	t := a.Theme
	w, h := c.Size()
	ui.DialogBase(c, open, func(back, panel ui.Element) {
		opacity := float32(.22)
		if t.Dark {
			opacity = .42
		}
		back.Background(ui.RGBA(0, 0, 0, opacity))
		panel.Label(label).Width(min(width, max(0, w-48))).MaxHeight(max(0, h-48)).Padding(0).Gap(0).Background(t.Elevated).Border(1, t.Border).Radius(t.Px(12))
		if height > 0 {
			panel.Height(min(height, max(0, h-48)))
		}
		panel.Children(func() { body(panel) })
	})
}
func (a *App) sourceProviderModalHeader(c *ui.Context, title, subtitle string, mono bool) {
	t := a.Theme
	ui.Column(c).AlignSelf(ui.Stretch).Shrink(0).Padding(t.Space(8), t.Space(12), t.Space(4), t.Space(12)).Gap(t.Space(8)).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.Text(c, title).FontSize(t.Px(16)).LineHeight(1.6).FontWeight(500)
		text := ui.Text(c, subtitle).FontSize(t.Px(12)).TextColor(t.TextMuted)
		if mono {
			text.Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 11 / 13)
		}
	})
}
func (a *App) sourceProviderModalError(c *ui.Context, message string) {
	if message == "" {
		return
	}
	ui.Text(c, message).FontSize(a.Theme.Px(12)).TextColor(a.Theme.Error).Background(a.Theme.Error.Alpha(.1)).Padding(a.Theme.Space(4), a.Theme.Space(6)).Margin(0, a.Theme.Space(12), a.Theme.Space(4), a.Theme.Space(12)).Radius(a.Theme.Px(8)).Shrink(0)
}
func (a *App) sourceProviderModalFooter(c *ui.Context, saveLabel string, busy, saveDisabled bool, cancel, save func()) {
	t := a.Theme
	ui.Row(c).AlignSelf(ui.Stretch).Shrink(0).Padding(t.Space(8)).Gap(t.Px(4)).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Justify(ui.End).Children(func() {
		if a.sourceProviderButton(c, "Cancel", a.Tr("settings.cancel"), "", false).Height(t.Px(32)).Disabled(busy).Clicked() {
			cancel()
		}
		if a.sourceProviderButton(c, "Save provider", saveLabel, "", true).Height(t.Px(32)).Disabled(busy || saveDisabled).Clicked() {
			save()
		}
	})
}
func (a *App) providerDialog(c *ui.Context) {
	d := a.providerState.editor
	if d == nil {
		return
	}
	t := a.Theme
	open := true
	a.sourceProviderModal(c.Key("provider-editor:"+d.originalID), &open, 520, 560, "Provider editor dialog", func(panel ui.Element) {
		title := a.Tr("settings.add_provider")
		if d.originalID != "" {
			title = a.Tr("settings.configure_provider")
		}
		a.sourceProviderModalHeader(c, title, a.Tr("settings.saved_to_models_json"), false)
		ui.Scroll(c).Label("Provider editor body").Grow(1).MinHeight(0).AlignSelf(ui.Stretch).Padding(t.Space(16), t.Space(12)).Gap(t.Space(12)).Children(func() {
			field := func(id, label, hint, placeholder string, value *string, password, focus bool) {
				ui.Column(c.Key(id)).AlignSelf(ui.Stretch).Gap(t.Px(4)).Children(func() {
					ui.Text(c, label).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextSecondary)
					input := ui.TextInput(c, value).Label(id).Placeholder(placeholder).AlignSelf(ui.Stretch).Height(t.Px(36)).Padding(t.Px(6), t.Px(8)).Radius(t.Px(8)).Background(t.Background).Disabled(d.busy)
					if password {
						input.Password()
					}
					if focus {
						input.AutoFocus()
					}
					input.OnTextKey(func(k ui.TextKey) bool {
						if k.Key == ui.KeyEnter && !d.busy {
							a.saveProviderDraft(d)
							return true
						}
						return false
					})
					if hint != "" {
						ui.Text(c, hint).FontSize(t.Px(12)).TextColor(t.TextMuted)
					}
				})
			}
			if d.originalID == "" {
				field("Provider ID", a.Tr("settings.provider_id"), a.Tr("settings.provider_id_hint"), "my-gateway", &d.edit.ID, false, true)
			} else {
				ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Px(4)).Children(func() {
					ui.Text(c, a.Tr("settings.provider_id")).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextSecondary)
					ui.Text(c, d.originalID).Label("Provider ID").Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 12 / 13)
				})
			}
			field("Display name", a.Tr("settings.display_name"), a.Tr("settings.optional"), a.Tr("settings.optional_display_name"), &d.edit.Name, false, d.originalID != "")
			field("Base URL", a.Tr("settings.base_url"), a.Tr("settings.base_url_hint"), "https://api.example.com/v1", &d.edit.BaseURL, false, false)
			ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Px(4)).Children(func() {
				ui.Text(c, a.Tr("settings.api_type")).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextSecondary)
				ui.Row(c).AlignSelf(ui.Stretch).Wrap().Gap(t.Space(6)).Children(func() {
					for _, api := range providerAPIs {
						b := ui.ButtonBase(c.Key("api:"+api)).Label("API family "+api).Height(t.Px(22)).Padding(0, t.Space(6)).Radius(t.Px(6)).Border(1, t.Border).Background(t.Elevated).Disabled(d.busy)
						color := t.TextMuted
						if d.edit.API == api {
							b.BorderColor(t.Accent.Alpha(.5)).Background(t.Accent.Alpha(.12))
							color = t.AccentText
						}
						if b.Hovered() {
							b.Background(t.Hover)
						}
						b.Children(func() {
							ui.Text(c, api).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 12 / 13).TextColor(color).SingleLine()
						})
						if b.Clicked() {
							d.edit.API = api
						}
					}
				})
			})
			hint, placeholder := a.Tr("settings.api_key_optional_hint"), "sk-…"
			if d.hadAPIKey {
				hint, placeholder = a.Tr("settings.api_key_stored_hint"), "••••••••"
			}
			field("API key", a.Tr("settings.api_key"), hint, placeholder, &d.key, true, false)
			ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Px(4)).Children(func() {
				ui.Text(c, a.Tr("settings.models_field")).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextSecondary)
				ui.TextArea(c, &d.models).Label("Model IDs").Placeholder(a.Tr("settings.model_id_model_id_2")).MinHeight(t.Px(36)).MaxHeight(t.Px(102)).AlignSelf(ui.Stretch).Padding(t.Px(6), t.Px(8)).Radius(t.Px(8)).Background(t.Background).Disabled(d.busy)
				hint := a.Tr("settings.models_custom_hint")
				if d.inCatalog {
					hint = a.Tr("settings.models_builtin_hint")
				}
				ui.Text(c, hint).FontSize(t.Px(12)).TextColor(t.TextMuted)
			})
		})
		a.sourceProviderModalError(c, d.error)
		label := a.Tr("settings.add_provider")
		if d.originalID != "" {
			label = a.Tr("settings.save_changes")
		}
		a.sourceProviderModalFooter(c, label, d.busy, false, func() { open = false }, func() { a.saveProviderDraft(d) })
	})
	if !open {
		d.key = ""
		a.providerState.editor = nil
	}
}
func (a *App) secretDialog(c *ui.Context) {
	if !a.secretOpen {
		return
	}
	s, t := &a.providerState, a.Theme
	name, note := a.secretProvider, ""
	var provider sourceProviderView
	for _, v := range a.sourceProviderViews() {
		if v.meta.ID == a.secretProvider {
			name, note, provider = v.name, v.meta.Note, v
			break
		}
	}
	a.sourceProviderModal(c.Key("secret-input:"+a.secretProvider), &a.secretOpen, 460, 0, "Provider credential dialog", func(panel ui.Element) {
		a.sourceProviderModalHeader(c, strings.ReplaceAll(a.Tr("settings.api_key_for"), "%{name}", name), a.secretProvider, true)
		_, h := c.Size()
		ui.Scroll(c).Label("Provider credential body").MaxHeight(max(0, h-200)).MinHeight(0).AlignSelf(ui.Stretch).Padding(t.Space(16), t.Space(12)).Gap(t.Space(12)).Children(func() {
			label, hint, placeholder := a.Tr("settings.api_key"), a.Tr("settings.api_key_hint"), "sk-…"
			if s.secretCookie {
				label, hint, placeholder = a.Tr("settings.session_cookie"), a.Tr("settings.session_cookie_hint"), "__Secure-session=…"
			}
			ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Px(4)).Children(func() {
				ui.Text(c, label).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextSecondary)
				ui.TextInput(c, &a.secretValue).Password().Label(label).Placeholder(placeholder).AutoFocus().Height(t.Px(36)).Padding(t.Px(6), t.Px(8)).Radius(t.Px(8)).Background(t.Background).Disabled(s.secretBusy).AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
					if k.Key == ui.KeyEnter {
						a.saveProviderSecret()
						return true
					}
					return false
				})
				ui.Text(c, hint).FontSize(t.Px(12)).TextColor(t.TextMuted)
			})
			if note != "" {
				ui.Text(c, note).FontSize(t.Px(12)).TextColor(t.TextMuted)
			}
			if !s.secretCookie {
				for _, method := range provider.auth.Methods {
					if method.ID == "api_key" {
						continue
					}
					ui.Row(c).AlignSelf(ui.Stretch).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
						ui.Text(c, a.Tr("settings.prefer_a_subscription_sign_in_with_oauth_instead")).FontSize(t.Px(12)).TextColor(t.TextMuted).Grow(1)
						if a.sourceProviderButton(c, "Sign in instead", a.Tr("settings.sign_in"), "", false).Disabled(s.secretBusy).Clicked() {
							a.secretValue = ""
							a.secretOpen = false
							s.dismissedAuthSession = ""
							a.Dispatch(Action{Kind: "Login", Value: provider.meta.ID, Data: map[string]any{"method": method.ID}})
						}
					})
					break
				}
			}
			a.sourceProviderModalError(c, s.secretError)
		})
		ui.Row(c).AlignSelf(ui.Stretch).Shrink(0).Padding(t.Space(8)).Gap(t.Px(4)).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Justify(ui.End).Children(func() {
			if a.sourceProviderButton(c, "Cancel", a.Tr("settings.cancel"), "", false).Height(t.Px(32)).Disabled(s.secretBusy).Clicked() {
				a.secretOpen = false
			}
			if a.sourceProviderButton(c, "Save credential", a.Tr("settings.save_key"), "", true).Disabled(s.secretBusy || strings.TrimSpace(a.secretValue) == "").Clicked() {
				a.saveProviderSecret()
			}
		})
	})
	if !a.secretOpen {
		a.secretValue = ""
	}
}
