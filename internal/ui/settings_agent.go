package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
)

func (a *App) agentSettings(c *ui.Context) {
	if a.Model.Retrying {
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
			a.sourceSettingsSegments(c, "Enter while running", a.Prefs.ComposerSendMode, []settingsChoice{{"follow_up", a.Tr("composer.queue_follow_up")}, {"steer", a.Tr("composer.steer_task")}}, false, func(v string) { a.Prefs.ComposerSendMode = v; a.save(a.Options.Store.SaveUI(a.Prefs)) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.follow_up_messages"), a.Tr("settings.messages_sent_while_the_agent_is_running_wait_in"), "", func() {
			if a.Model.FollowUpMode == "" {
				ui.Text(c, "Unknown").FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
			}
			a.sourceSettingsSegments(c, "Follow-up delivery", a.Model.FollowUpMode, []settingsChoice{{"one-at-a-time", a.Tr("settings.one_at_a_time")}, {"all", a.Tr("settings.all")}}, !a.Model.Connected || a.Model.FollowUpMode == "", func(v string) { a.Dispatch(Action{Kind: "SetFollowUpMode", Value: v}) })
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
		// Do not reproduce Rust's invented initial enabled state. A successful
		// setter acknowledgement remains the only evidence for this value.
		a.sourceSettingsRow(c, a.Tr("settings.auto_retry"), a.Tr("settings.retry_automatically_on_transient_errors_overload"), status, func() {
			ui.MenuButton(c, "Set retry policy", func(m *ui.Menu) {
				for _, value := range []bool{true, false} {
					label := "Enable"
					if !value {
						label = "Disable"
					}
					if m.Item(label).Checked(retry.Known && retry.Enabled == value).Chosen() {
						a.Dispatch(Action{Kind: "SetAutoRetry", Data: map[string]any{"enabled": value}})
					}
				}
			}).Height(a.Theme.Px(28)).Padding(0, a.Theme.Space(8)).Disabled(!a.Model.Connected || retry.Pending || retry.Support == "unsupported")
		})
		if retry.Error != "" {
			ui.Text(c, retry.Error).TextColor(a.Theme.Error).Padding(8, 16)
		}
	})
}
