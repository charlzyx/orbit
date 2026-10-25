package orbitui

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

func runtimeUptime(p *agent.RuntimeProcess) string {
	if p == nil || !p.Alive || p.StartedAt.IsZero() {
		return "—"
	}
	d := p.Uptime.Round(time.Second)
	if d < 0 {
		d = 0
	}
	return d.String()
}
func (a *App) runtimeSettings(c *ui.Context) {
	defer a.runtimeNative(c)
	snap := a.Runtime
	p := snap.Active
	state := "stopped"
	pid := "—"
	executable := snap.ConfiguredExecutable
	transport := snap.Transport
	if transport == "" {
		transport = "stdio JSONL"
	}
	if executable == "" {
		executable = "pi (resolved on start)"
	}
	if p != nil {
		state = p.State
		if p.PID > 0 {
			pid = fmt.Sprint(p.PID)
		}
		if p.Executable != "" {
			executable = p.Executable
		} else if p.Launcher != "" {
			executable = p.Launcher
		}
	}
	a.section(c, a.text("settings.process", "Process"), func() {
		a.setting(c, a.text("settings.status", "Status"), "pi --mode rpc · direct child process", func() {
			color := a.Theme.TextMuted
			if p != nil && p.Alive {
				color = a.Theme.Success
			}
			if state == "failed" || state == "exited" {
				color = a.Theme.Error
			}
			ui.Text(c, state).TextColor(color)
			if p != nil && (p.Alive || p.Loading) {
				if a.button(c, "Restart", "icons/refresh.svg", false).Disabled(state == "stopping").Clicked() {
					a.Dispatch(Action{Kind: "RestartRuntime"})
				}
				if a.button(c, "Stop", "icons/stop.svg", false).Disabled(state == "stopping").Clicked() {
					a.Dispatch(Action{Kind: "StopRuntime"})
				}
			} else {
				if a.button(c, "Start", "icons/start-up.svg", true).Disabled(snap.Closed).Clicked() {
					a.Dispatch(Action{Kind: "StartRuntime"})
				}
			}
		})
		a.setting(c, a.text("settings.process_id", "Process ID"), "", func() { ui.Text(c, pid).Font(a.Fonts.CodeFamily).Selectable() })
		a.setting(c, a.text("settings.binary", "Binary"), "", func() { ui.Text(c, executable).MaxWidth(500).MaxLines(2).Font(a.Fonts.CodeFamily).Selectable() })
		a.setting(c, a.text("settings.uptime", "Uptime"), "", func() { ui.Text(c, runtimeUptime(p)).Font(a.Fonts.CodeFamily) })
		a.setting(c, a.text("settings.transport", "Transport"), "", func() { ui.Text(c, transport) })
		patches := a.Model.Extensions.Patches
		summary := "Disabled"
		if patches.Enabled {
			summary = fmt.Sprintf("Enabled · %d applied", len(patches.Applied))
		}
		a.setting(c, "RPC patches", summary, nil)
		if patches.PackageRoot != "" {
			a.setting(c, "Package root", patches.PackageRoot, nil)
		}
		if patches.Error != "" {
			ui.Text(c, patches.Error).TextColor(a.Theme.Error).Selectable()
		}
		if p != nil && p.Error != "" {
			a.setting(c, a.text("settings.last_error", "Last error"), p.Error, nil)
		}
		if p != nil && p.ExitCode != nil {
			a.setting(c, "Exit code", fmt.Sprint(*p.ExitCode), nil)
		}
		ui.Row(c).Wrap().Gap(8).Children(func() {
			if a.button(c, "Choose pi executable…", "icons/folder.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "ChoosePiBinary"})
			}
			if a.button(c, "Check pi update", "icons/refresh.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "CheckPiUpdate"})
			}
		})
	})
	if len(snap.Parked) > 0 {
		a.section(c, a.text("settings.background", "Background"), func() {
			a.setting(c, a.text("settings.parked_processes", "Parked processes"), fmt.Sprintf("%d background sessions", len(snap.Parked)), nil)
			for _, other := range snap.Parked {
				name := other.SessionName
				if name == "" {
					name = filepath.Base(other.SessionPath)
				}
				value := other.State
				if other.PID > 0 {
					value = fmt.Sprintf("PID %d · %s", other.PID, other.State)
				}
				a.setting(c, name, value, nil)
			}
		})
	}
	a.section(c, a.text("settings.stderr", "Recent stderr"), func() {
		if p == nil || len(p.RecentStderr) == 0 {
			ui.Text(c, a.text("settings.no_output_from_the_pi_process", "No output from the pi process.")).TextColor(a.Theme.TextMuted)
			return
		}
		ui.Scroll(c).MaxHeight(180).AlignSelf(ui.Stretch).Padding(12).Background(a.Theme.CodeBg).Radius(8).Children(func() {
			for _, line := range p.RecentStderr {
				ui.Text(c, line).Font(a.Fonts.CodeFamily).FontSize(a.Theme.CodeSize * 11.5 / 13).TextColor(a.Theme.CodeText).Selectable()
			}
		})
	})
}
func (a *App) agentRuntimeSettings(c *ui.Context) {
	a.section(c, "Active session behavior", func() {
		for _, item := range []struct{ label, value, kind string }{{"Steering delivery", a.Model.SteeringMode, "SetSteeringMode"}, {"Follow-up delivery", a.Model.FollowUpMode, "SetFollowUpMode"}} {
			label := item.value
			if label == "" {
				label = "Unknown"
			}
			a.setting(c, item.label, "One queued message at a time, or all available messages.", func() {
				ui.MenuButton(c, label, func(m *ui.Menu) {
					for _, mode := range []string{"one-at-a-time", "all"} {
						if m.Item(mode).Checked(mode == item.value).Chosen() {
							a.Dispatch(Action{Kind: item.kind, Value: mode})
						}
					}
				}).Disabled(!a.Model.Connected || item.value == "")
			})
		}
		a.setting(c, "Automatic compaction", "Compact the active session when context reaches pi's threshold.", func() {
			if !a.Model.AutoCompactionKnown {
				ui.Text(c, "Unknown").TextColor(a.Theme.TextMuted)
				return
			}
			value := a.Model.AutoCompactionEnabled
			if ui.Switch(c, &value).Label("Automatic compaction").Disabled(!a.Model.Connected).Changed() {
				a.Dispatch(Action{Kind: "SetAutoCompaction", Data: map[string]any{"enabled": value}})
			}
		})
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
		a.setting(c, "Automatic retry", status, func() {
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
			}).Disabled(!a.Model.Connected || retry.Pending || retry.Support == "unsupported")
		})
		if retry.Error != "" {
			ui.Text(c, retry.Error).TextColor(a.Theme.Error)
		}
	})
	a.section(c, "Automatic title model", func() {
		title := a.autoTitlePrefs
		label := a.text("settings.active_session_model", "Active session model")
		if title.Model != nil {
			label = title.Model.Provider + " / " + title.Model.ID
		}
		ui.MenuButton(c, label, func(m *ui.Menu) {
			if m.Item(a.text("settings.active_session_model", "Active session model")).Checked(title.Model == nil).Chosen() {
				title.Model = nil
				a.autoTitlePrefs = title
				a.save(a.Options.Store.SaveAutoTitle(title))
				a.Dispatch(Action{Kind: "AutoTitleChanged"})
			}
			for _, model := range a.Model.Models {
				if m.Item(model.Provider + " / " + model.Name).Checked(title.Model != nil && title.Model.ID == model.ID && title.Model.Provider == model.Provider).Chosen() {
					title.Model = &preferences.Model{Provider: model.Provider, ID: model.ID}
					a.autoTitlePrefs = title
					a.save(a.Options.Store.SaveAutoTitle(title))
					a.Dispatch(Action{Kind: "AutoTitleChanged"})
				}
			}
		}).Disabled(!title.Enabled)
	})
}
