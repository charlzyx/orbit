package orbitui

import (
	"github.com/egoist/mygo/ui"
	"math"
	"path/filepath"
	"strings"
)

// GPUI 0.2.2 TextStyle::default uses phi(); preserve it for source-native settings rows.
const sourceDefaultLineHeight float32 = 1.618034

func (a *App) sourceSettingsSection(c *ui.Context, title, description string, body func()) {
	t := a.Theme
	ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(8)).Children(func() {
		ui.Column(c).AlignSelf(ui.Stretch).PaddingX(t.Space(4)).Gap(t.Space(2)).Children(func() {
			ui.Text(c, strings.ToUpper(title)).FontSize(t.Px(10)).Height(sourceTextHeight(t.Px(10), sourceDefaultLineHeight)).FontWeight(500).TextColor(t.TextMuted)
			if description != "" {
				ui.Text(c, description).FontSize(t.Px(12)).TextColor(t.TextMuted)
			}
		})
		ui.Column(c).Label("Settings group "+title).AlignSelf(ui.Stretch).Background(t.SurfaceSecondary).Border(1, t.Border).Radius(t.Px(12)).Children(body)
	})
}
func (a *App) sourceSettingsRow(c *ui.Context, title, description, meta string, control func()) {
	t := a.Theme
	ui.Row(c).AlignSelf(ui.Stretch).Padding(t.Space(12), t.Space(16)).Gap(t.Space(16)).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Grow(1).MinWidth(0).Gap(t.Space(3)).Children(func() {
			ui.Text(c, title).FontWeight(500).TextColor(t.Text)
			if description != "" {
				ui.Text(c, description).FontSize(t.Px(12)).TextColor(t.TextSecondary)
			}
			if meta != "" {
				ui.Text(c, meta).FontSize(t.Px(12)).TextColor(t.TextMuted).MaxLines(1).Tooltip(meta)
			}
		})
		if control != nil {
			control()
		}
	})
}
func (a *App) sourceSettingsDivider(c *ui.Context) {
	ui.Box(c).AlignSelf(ui.Stretch).Height(1).MarginX(a.Theme.Space(16)).Background(a.Theme.Border)
}
func (a *App) generalSettings(c *ui.Context) {
	a.sourceSettingsSection(c, a.text("settings.this_machine", "This machine"), "", func() {
		a.sourceSettingsRow(c, a.text("settings.pi_agent", "pi agent"), a.text("settings.spawned_as_a_child_process_newline_delimited_jso", "Spawned as a child process — newline-delimited JSON over stdio."), "", func() {
			connected := a.Model.Connected || (a.Runtime.Active != nil && a.Runtime.Active.Alive)
			label, color := a.text("settings.not_running", "Not running"), a.Theme.Error
			if connected {
				label = a.text("status.connected", "Connected")
				color = a.Theme.Success
			}
			ui.Row(c).Gap(a.Theme.Space(6)).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(7, 7).Radius(3.5).Background(color)
				ui.Text(c, label).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextSecondary)
			})
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.text("settings.local_by_default", "Local by default"), a.text("settings.sessions_live_in_pi_s_own_store_on_this_computer", "Sessions live in pi's own store on this computer — no daemon, no cloud."), filepath.Join(a.Options.Store.PiDir, "sessions"), nil)
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.text("settings.workspace", "Workspace"), a.text("settings.new_tasks_start_in_this_directory", "New tasks start in this directory."), a.Model.Workspace, nil)
	})
	// Source updater_ui.rs::updater_section is absent when updates are unavailable.
	if a.Options.Updater != nil && a.Options.Updater.Snapshot().Enabled {
		a.sourceSettingsSection(c, a.text("updater_ui.updates", "Updates"), "", func() { ui.Column(c).Padding(12, 16).Gap(12).Children(func() { a.updateSettings(c) }) })
	}
	a.sourceSettingsSection(c, a.text("settings.notifications", "Notifications"), a.text("settings.notifications_description", "When a run finishes or pi needs an answer."), func() {
		n := a.notificationPrefs
		rows := []struct {
			key, description, fallback string
			value                      *bool
		}{
			{"settings.desktop_notifications", "settings.show_a_system_banner_when_a_run_finishes_while_o", "Desktop notifications", &n.Desktop},
			{"settings.in_app_toasts", "settings.show_a_toast_in_the_window_when_a_run_finishes_o", "In-app toasts", &n.Toasts},
			{"settings.notification_sound", "settings.play_the_system_alert_sound_when_a_background_ru", "Notification sound", &n.Sound},
		}
		for i, r := range rows {
			label := a.text(r.key, r.fallback)
			a.sourceSettingsRow(c, label, a.Tr(r.description), "", func() {
				if a.sourceSettingsToggle(c, r.value, label).Changed() {
					a.notificationPrefs = n
					a.save(a.Options.Store.SaveNotifications(n))
					a.Dispatch(Action{Kind: "NotificationsChanged"})
				}
			})
			if i < len(rows)-1 {
				a.sourceSettingsDivider(c)
			}
		}
	})
}

func (a *App) sourceSettingsToggle(c *ui.Context, value *bool, label string) ui.Element {
	t := a.Theme
	sw := ui.SwitchBase(c, value).Size(36, t.Space(20)).Radius(t.Space(20)/2).Border(1, t.Border).Label(label + " toggle")
	if *value {
		sw.Background(t.Accent)
	} else {
		sw.Background(t.Elevated)
	}
	if sw.Hovered() {
		opacity := float32(.16)
		if t.Dark {
			opacity = .14
		}
		sw.BorderColor(t.Text.Alpha(opacity))
	}
	on := *value
	sw.Draw(func(p *ui.Painter, r ui.Rect) {
		size := float32(14)
		x := r.X + 1 + t.Space(2)
		if on {
			x = r.X + r.W - 1 - t.Space(2) - size
		}
		knob := ui.RGB(255, 255, 255)
		if t.Dark {
			knob = t.Text
		}
		p.Fill(ui.Rect{X: x, Y: r.Y + (r.H-size)/2, W: size, H: size}, knob, size/2)
	})
	return sw
}

// GPUI resolves line boxes to whole logical pixels; MyGo otherwise retains fractions.
func sourceTextHeight(size, line float32) float32 { return float32(math.Round(float64(size * line))) }
