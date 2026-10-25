package orbitui

import (
	"strings"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func (a *App) agentSettings(c *ui.Context) {
	if a.retryPresentationActive() {
		a.sourceSettingsSection(c, a.Tr("settings.status"), "", func() {
			a.sourceSettingsRow(c, a.Tr("settings.retrying"), a.Tr("settings.pi_is_waiting_out_a_transient_provider_error_bef"), "", func() {
				if a.sourceSettingsButton(c, a.Tr("settings.abort_retry")).Disabled(!a.Model.Connected).Clicked() {
					a.Dispatch(Action{Kind: "AbortRetry"})
				}
			})
		})
	}
	if !a.Model.Connected {
		a.sourceSettingsSection(c, a.Tr("settings.status"), "", func() {
			a.sourceSettingsRow(c, a.Tr("settings.pi_is_not_connected"), a.Tr("settings.start_the_runtime_from_settings_runtime_to_chang"), "", nil)
		})
	}
	a.agentRuntimeSettings(c)
	a.sourceSettingsSection(c, a.Tr("settings.session_titles"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.auto_session_titles"), a.Tr("settings.name_each_session_from_its_first_exchange_so"), "", func() {
			value := a.autoTitlePrefs.Enabled
			if a.sourceSettingsToggle(c, &value, "Auto session titles").Changed() {
				a.autoTitlePrefs.Enabled = value
				a.save(a.Options.Store.SaveAutoTitle(a.autoTitlePrefs))
				a.Dispatch(Action{Kind: "AutoTitleChanged"})
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.title_model"), a.Tr("settings.model_that_writes_the_title_defaults_to_the_a"), "", func() {
			if a.Models != nil {
				a.Models.TitleModelControl(c)
			}
		})
	})
	a.sourceSettingsSection(c, a.Tr("settings.session_defaults"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.default_model"), a.Tr("settings.default_model_hint"), "", func() {
			if a.Models != nil {
				a.Models.DefaultModelControl(c)
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.default_thinking"), a.Tr("settings.default_thinking_hint"), "", func() {
			if a.Models != nil {
				a.Models.DefaultThinkingControl(c)
			}
		})
	})
}
func (a *App) agentRuntimeSettings(c *ui.Context) {
	a.sourceSettingsSection(c, a.Tr("settings.behavior"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.enter_while_running"), strings.ReplaceAll(a.Tr("settings.enter_while_running_hint"), "%{keys}", sourceShortcutLabel("alt-enter")), "", func() {
			a.agentSettingsSegments(c, "Enter while running", a.Prefs.ComposerSendMode, []settingsChoice{{"follow_up", a.Tr("composer.queue_follow_up")}, {"steer", a.Tr("composer.steer_task")}}, a.Theme.Active, false, func(v string) { a.Prefs.ComposerSendMode = v; a.save(a.Options.Store.SaveUI(a.Prefs)) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.follow_up_messages"), a.Tr("settings.messages_sent_while_the_agent_is_running_wait_in"), "", func() {
			if a.Model.FollowUpMode == "" {
				ui.Text(c, "Unknown").FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
			}
			a.agentSettingsSegments(c, "Follow-up delivery", a.Model.FollowUpMode, []settingsChoice{{"one-at-a-time", a.Tr("settings.one_at_a_time")}, {"all", a.Tr("settings.all")}}, a.Theme.Selection, !a.Model.Connected || a.Model.FollowUpMode == "", func(v string) { a.Dispatch(Action{Kind: "SetFollowUpMode", Value: v}) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.auto_compaction"), a.Tr("settings.compact_conversation_context_automatically_when_"), "", func() {
			if !a.Model.AutoCompactionKnown {
				ui.Text(c, "Unknown").FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
				return
			}
			value := a.Model.AutoCompactionEnabled
			if a.sourceSettingsToggle(c, &value, "Automatic compaction").Disabled(!a.Model.Connected).Changed() {
				a.Dispatch(Action{Kind: "SetAutoCompaction", Data: map[string]any{"enabled": value}})
			}
		})
		a.sourceSettingsDivider(c)
		retry := agent.RetrySetting{Support: "unknown"}
		if a.Runtime.Active != nil {
			retry = a.Runtime.Active.AutoRetry
		}
		status := "Unknown: this pi version has no getter"
		if retry.Known {
			status = map[bool]string{true: "Enabled", false: "Disabled"}[retry.Enabled]
		}
		if retry.Support == "unsupported" {
			status = "Unsupported by this pi runtime"
		}
		if retry.Pending {
			status = "Updating…"
		}
		a.agentRetryRow(c, retry, status)
		if retry.Error != "" {
			ui.Text(c, retry.Error).TextColor(a.Theme.Error).Padding(8, 16)
		}
	})
}

// D04 deliberately keeps the unobserved state unknown. Put its explanation on
// the title line and keep the control on the source's 36-DIP axis, so this
// truthfulness correction does not add a third text line or wrap the description
// merely because a long menu label replaced the source switch.
func (a *App) agentRetryRow(c *ui.Context, retry agent.RetrySetting, status string) {
	t := a.Theme
	disabled := !a.Model.Connected || retry.Pending || retry.Support == "unsupported"
	ui.Row(c).Label("Auto-retry setting row").AlignSelf(ui.Stretch).Padding(t.Space(12), t.Space(16)).Gap(t.Space(16)).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.Space(3)).Children(func() {
			ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(t.Space(8)).Children(func() {
				ui.Text(c, a.Tr("settings.auto_retry")).FontWeight(500).TextColor(t.Text).Shrink(0)
				ui.Text(c, status).FontSize(t.Px(12)).TextColor(t.TextMuted).Grow(1).MinWidth(0).MaxLines(1).Tooltip(status)
			})
			ui.Text(c, a.Tr("settings.retry_automatically_on_transient_errors_overload")).FontSize(t.Px(12)).TextColor(t.TextSecondary)
		})
		if retry.Known && retry.Support != "unsupported" {
			// Only an acknowledged value may be represented by an on/off switch.
			// Mutate a local copy, never the observed runtime snapshot optimistically.
			value := retry.Enabled
			if a.sourceSettingsToggle(c, &value, "Automatic retry").Disabled(disabled).Changed() {
				a.Dispatch(Action{Kind: "SetAutoRetry", Data: map[string]any{"enabled": value}})
			}
			return
		}
		ui.MenuButton(c, "", func(m *ui.Menu) {
			for _, value := range []bool{true, false} {
				label := "Enable"
				if !value {
					label = "Disable"
				}
				if m.Item(label).Disabled(disabled).Chosen() {
					a.Dispatch(Action{Kind: "SetAutoRetry", Data: map[string]any{"enabled": value}})
				}
			}
		}).Label("Set retry policy").Size(36, t.Space(20)).Padding(0).Radius(t.Space(10)).Tooltip(status).Disabled(disabled)
	})
}

// Agent's two source segmented controls intentionally use different selection
// roles: a neutral local composer preference, and blue runtime queue delivery.
// Keep this bounded to Agent instead of changing unrelated settings controls.
func (a *App) agentSettingsSegments(c *ui.Context, id, value string, choices []settingsChoice, selected ui.Color, disabled bool, apply func(string)) {
	t := a.Theme
	row := ui.Row(c.Key(id)).Label(id).Focusable().Disabled(disabled).AlignItems(ui.Center).Radius(t.Px(8)).Border(1, t.Border).Clip()
	row.Children(func() {
		for _, choice := range choices {
			b := ui.ButtonBase(c.Key(choice.Value)).Label(choice.Label).Height(t.Px(28)-2).Padding(0, t.Space(8)).Radius(0).FontSize(t.Px(14)).FontWeight(500).LineHeight(1).TextColor(t.TextSecondary)
			fill := t.Elevated
			if value == choice.Value {
				fill = selected
				b.TextColor(t.Text)
			} else if b.Hovered() {
				fill = t.Hover
			}
			// This shared pure color conversion implements Rust raised_fill's
			// HSL lift/sink, also used by the source-derived Usage controls.
			top, bottom := usageRaisedColors(fill)
			b.Gradient(top, bottom, 180)
			b.Children(func() { ui.Text(c, choice.Label) })
			if b.Clicked() && !disabled {
				apply(choice.Value)
			}
		}
	})
	if disabled || len(choices) == 0 {
		return
	}
	index := 0
	for i, choice := range choices {
		if choice.Value == value {
			index = i
		}
	}
	next := index
	switch {
	case row.Shortcut(0, ui.KeyLeft), row.Shortcut(0, ui.KeyUp):
		next = (index + len(choices) - 1) % len(choices)
	case row.Shortcut(0, ui.KeyRight), row.Shortcut(0, ui.KeyDown):
		next = (index + 1) % len(choices)
	case row.Shortcut(0, ui.KeyHome):
		next = 0
	case row.Shortcut(0, ui.KeyEnd):
		next = len(choices) - 1
	}
	if next != index {
		apply(choices[next].Value)
	}
}
