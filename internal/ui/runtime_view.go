package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"strings"
	"time"
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
	snap := a.Runtime
	p := snap.Active
	t := a.Theme
	state := "stopped"
	pid := "—"
	executable := snap.ConfiguredExecutable
	if executable == "" {
		executable = "pi (resolved on start)"
	}
	if p != nil {
		state = p.State
		if p.Alive {
			state = "running"
		}
		if p.Alive && p.PID > 0 {
			pid = fmt.Sprint(p.PID)
		}
		if p.Executable != "" {
			executable = p.Executable
		} else if p.Launcher != "" {
			executable = p.Launcher
		}
	}
	description := a.Tr("settings.runtime_" + state)
	if state == "running" {
		description = a.Tr("settings.spawned_child_process")
	}
	if description == "settings.runtime_"+state {
		description = "pi --mode rpc · direct child process"
	}
	runtimeValue := func(value string, path bool) {
		e := ui.Text(c, value).FontSize(t.Px(12)).TextColor(t.Text).MaxLines(1).MaxWidth(650).Selectable()
		if path {
			e.Font(a.Fonts.CodeFamily).TextColor(t.TextSecondary).Tooltip(value)
		}
	}
	a.sourceSettingsSection(c, a.Tr("settings.process"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.status"), description, "", func() {
			color := t.TextMuted
			if state == "running" {
				color = t.Success
			}
			if state == "failed" || state == "exited" {
				color = t.Error
			}
			label := a.Tr("runtime.state_" + state)
			if strings.HasPrefix(label, "runtime.state_") {
				label = state
			}
			ui.Row(c).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
				ui.Row(c).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() { ui.Box(c).Size(7, 7).Radius(3.5).Background(color); ui.Text(c, label).FontSize(t.Px(12)) })
				if p != nil && (p.Alive || p.Loading) {
					if a.sourceSettingsButton(c, "Restart").Disabled(state == "stopping").Clicked() {
						a.Dispatch(Action{Kind: "RestartRuntime"})
					}
					if a.sourceSettingsButton(c, "Stop").Disabled(state == "stopping").Clicked() {
						a.Dispatch(Action{Kind: "StopRuntime"})
					}
				} else {
					if a.sourceSettingsButton(c, "Start").Disabled(snap.Closed).Clicked() {
						a.Dispatch(Action{Kind: "StartRuntime"})
					}
				}
			})
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.process_id"), "", "", func() { runtimeValue(pid, false) })
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.binary"), "", "", func() { runtimeValue(executable, true) })
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.uptime"), "", "", func() { runtimeValue(runtimeUptime(p), false) })
		a.sourceSettingsDivider(c)
		transport := snap.Transport
		if transport == "" || transport == "stdio JSONL" {
			transport = a.Tr("settings.transport_stdio")
		}
		a.sourceSettingsRow(c, a.Tr("settings.transport"), "", "", func() { runtimeValue(transport, false) })
		a.sourceSettingsDivider(c)
		patches := a.Model.Extensions.Patches
		summary := "Disabled — pi keeps the current patches until it updates"
		if patches.Enabled {
			summary = "Enabled — patches apply before the next agent launch"
			if len(patches.Applied) > 0 {
				summary = "Applied on this launch: " + strings.NewReplacer("custom-ui-rpc", "custom UI", "quota-rpc", "quota", "auth-rpc", "auth").Replace(strings.Join(patches.Applied, ", "))
			}
		}
		a.sourceSettingsRow(c, "RPC patches", summary, "", func() {
			enabled := patches.Enabled
			if a.sourceSettingsToggle(c, &enabled, "RPC patches").Disabled(a.rpcPatchSaving).Changed() {
				a.Dispatch(Action{Kind: "SetRPCPatches", Data: map[string]any{"enabled": enabled}})
			}
		})
		if patches.PackageRoot != "" {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, "Package root", "", "", func() { runtimeValue(patches.PackageRoot, true) })
		}
		if patches.Error != "" {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, "RPC patches error", patches.Error, "", nil)
		}
		if p != nil && p.Error != "" {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, a.Tr("settings.last_error"), "", "", func() { ui.Text(c, p.Error).FontSize(t.Px(12)).TextColor(t.Error).MaxWidth(650).Selectable() })
		}
		if p != nil && p.ExitCode != nil {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, "Exit code", "", "", func() { runtimeValue(fmt.Sprint(*p.ExitCode), false) })
		}
	})
	if len(snap.Parked) > 0 {
		a.sourceSettingsSection(c, a.Tr("settings.background"), "", func() {
			desc := strings.ReplaceAll(a.Tr("settings.background_sessions"), "%{count}", fmt.Sprint(len(snap.Parked)))
			a.sourceSettingsRow(c, a.Tr("settings.parked_processes"), desc, "", nil)
			for _, other := range snap.Parked {
				name := other.SessionName
				if name == "" {
					name = filepath.Base(other.SessionPath)
				}
				value := other.State
				if other.PID > 0 {
					value = fmt.Sprintf("PID %d · %s", other.PID, other.State)
				}
				a.sourceSettingsDivider(c)
				a.sourceSettingsRow(c, name, "", "", func() { runtimeValue(value, false) })
			}
		})
	}
	ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(8)).Children(func() {
		ui.Text(c, a.Tr("settings.recent_stderr")).FontSize(t.Px(10)).FontWeight(500).TextColor(t.TextMuted).PaddingX(t.Space(4))
		if p == nil || len(p.RecentStderr) == 0 {
			ui.Text(c, a.Tr("settings.no_output_from_the_pi_process")).FontSize(t.Px(12)).TextColor(t.TextMuted).Padding(t.Space(12), t.Space(16))
			return
		}
		ui.Column(c).AlignSelf(ui.Stretch).Padding(t.Space(8), t.Space(12)).Gap(t.Space(2)).Background(t.CodeBg).Radius(t.Px(12)).Children(func() {
			lines := p.RecentStderr
			if len(lines) > 8 {
				lines = lines[len(lines)-8:]
			}
			for _, line := range lines {
				ui.Text(c, line).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 11.5 / 13).TextColor(t.CodeText).MaxLines(1).Selectable()
			}
		})
	})
}
