package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func (a *App) worktreeSettings(c *ui.Context) {
	path := filepath.Join(a.Options.Store.RootDir, "worktrees.json")
	save := func(config workspace.WorktreeConfig) {
		if err := workspace.SaveWorktreeConfig(path, config); err != nil {
			a.save(err)
		} else {
			a.worktreePrefs = config
			a.Dispatch(Action{Kind: "WorktreePreferencesChanged"})
		}
	}
	s := &a.nativeSettings
	if !s.worktreeInputsLoaded {
		s.worktreeDirectory = a.worktreePrefs.Directory
		s.worktreeScript = a.worktreePrefs.SetupScript
		s.worktreeInputsLoaded = true
	}
	a.sourceSettingsSection(c, a.Tr("worktree.settings.manage_section"), "", func() {
		a.sourceSettingsRow(c, a.Tr("worktree.settings.manage"), a.Tr("worktree.settings.manage_hint"), "", func() {
			if a.sourceSettingsButton(c, a.Tr("worktree.settings.manage_open")).TextColor(a.Theme.AccentText).BorderColor(a.Theme.Accent).Clicked() {
				a.command("OpenWorktrees", "")
			}
		})
	})
	a.sourceSettingsSection(c, a.Tr("worktree.settings.location_section"), "", func() {
		a.sourceSettingsRow(c, a.Tr("worktree.settings.directory"), a.Tr("worktree.settings.directory_hint"), "", func() {
			if ui.TextInput(c.Key("worktree-directory"), &s.worktreeDirectory).Label("Worktree directory").Placeholder(a.Tr("worktree.settings.directory_placeholder")).Width(240).Height(36).Background(a.Theme.Background).Changed() {
				v := strings.TrimSpace(s.worktreeDirectory)
				if v != "" && v != a.worktreePrefs.Directory {
					next := a.worktreePrefs
					next.Directory = v
					save(next)
				}
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("worktree.settings.setup_script"), a.Tr("worktree.settings.setup_script_hint"), "", func() {
			if ui.TextInput(c.Key("worktree-script"), &s.worktreeScript).Label("Worktree setup script").Placeholder(a.Tr("worktree.settings.script_placeholder")).Width(240).Height(36).Background(a.Theme.Background).Changed() {
				v := strings.TrimSpace(s.worktreeScript)
				if v != "" && v != a.worktreePrefs.SetupScript {
					next := a.worktreePrefs
					next.SetupScript = v
					save(next)
				}
			}
		})
	})
	a.sourceSettingsSection(c, a.Tr("worktree.settings.creation_section"), "", func() {
		a.sourceSettingsRow(c, a.Tr("worktree.settings.run_setup_auto"), a.Tr("worktree.settings.run_setup_auto_hint"), "", func() {
			next := a.worktreePrefs
			if a.sourceSettingsToggle(c, &next.RunSetupAuto, "Run setup automatically").Changed() {
				save(next)
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("worktree.settings.open_after_create"), a.Tr("worktree.settings.open_after_create_hint"), "", func() {
			next := a.worktreePrefs
			if a.sourceSettingsToggle(c, &next.OpenAfterCreate, "Open after creation").Changed() {
				save(next)
			}
		})
	})
}
func (a *App) privacySettings(c *ui.Context) {
	a.sourceSettingsSection(c, a.Tr("settings.session_retention"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.session_retention_enabled"), a.Tr("settings.session_retention_enabled_description"), "", func() {
			if a.sourceSettingsToggle(c, &a.retention.Enabled, "Automatic session cleanup").Changed() {
				a.save(a.Options.Store.SaveRetention(a.retention))
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.session_retention_action"), a.Tr("settings.session_retention_action_description"), "", func() {
			a.sourceSettingsSelect(c, "Retention action", a.retention.Mode, []settingsChoice{{"archive", a.Tr("settings.session_retention_mode_archive")}, {"delete", a.Tr("settings.session_retention_mode_delete")}}, func(v string) { a.retention.Mode = v; a.save(a.Options.Store.SaveRetention(a.retention)) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.session_retention_age"), a.Tr("settings.session_retention_age_description"), "", func() {
			choices := []settingsChoice{}
			for _, n := range []int{7, 14, 30, 60, 90, 180, 365} {
				v := strconv.Itoa(n)
				choices = append(choices, settingsChoice{v, strings.ReplaceAll(a.Tr("settings.session_retention_days"), "%{days}", v)})
			}
			a.sourceSettingsSelect(c, "Retention idle age", strconv.Itoa(a.retention.Days), choices, func(v string) {
				a.retention.Days, _ = strconv.Atoi(v)
				a.save(a.Options.Store.SaveRetention(a.retention))
			})
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.session_retention_run"), a.Tr("settings.session_retention_run_description"), "", func() {
			if a.sourceSettingsButton(c, a.Tr("settings.session_retention_run")).Label("Run session cleanup").Height(a.Theme.Px(32)).Disabled(a.retentionBusy).Clicked() {
				if a.retention.Mode == "delete" {
					a.Confirm("Permanently delete old tasks?", fmt.Sprintf("Delete saved sessions idle for more than %d days. Open tasks are excluded. This cannot be undone.", a.retention.Days), func() { a.Dispatch(Action{Kind: "RunRetention"}) })
				} else {
					a.Dispatch(Action{Kind: "RunRetention"})
				}
			}
		})
	})
	a.privacyNative(c)
}
func (a *App) aboutSettings(c *ui.Context) {
	t := a.Theme
	b := ui.ButtonBase(c.Key("about-star-banner")).Label("Star Orbit on GitHub").AlignSelf(ui.Stretch).Padding(t.Space(6), t.Space(8)).Gap(t.Space(8)).Radius(t.Px(8)).Border(1, t.Border).Background(t.Elevated).Justify(ui.Start)
	if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() {
		a.icon(c, "icons/star.svg", 14, t.AccentText)
		ui.Column(c).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, a.Tr("sidebar.star_orbit")).FontSize(t.Px(12)).FontWeight(500)
			ui.Text(c, a.Tr("sidebar.star_orbit_detail")).FontSize(t.Px(10)).TextColor(t.TextMuted)
		})
	})
	if b.Clicked() {
		a.Dispatch(Action{Kind: "OpenLink", Value: "https://github.com/charlzyx/orbit"})
	}
	a.sourceSettingsGroup(c, "About versions", func() {
		a.sourceSettingsRow(c, a.Tr("settings.orbit_pi"), a.Tr("settings.native_workbench_for_the_pi_coding_agent"), "", func() { ui.Text(c, "v"+a.Options.Version).FontSize(t.Px(12)).TextColor(t.TextSecondary) })
		a.sourceSettingsDivider(c)
		// The migrated implementation must never claim the reference's GPUI engine.
		a.sourceSettingsRow(c, "MyGo", "Native GPU UI framework · Go implementation ("+runtime.Version()+").", "", func() { ui.Text(c, "9c69cf829ece").FontSize(t.Px(12)).TextColor(t.TextSecondary) })
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.pi_cli"), a.Tr("settings.the_only_agent_runtime_pi_speaks_its_own_rpc_pro"), "", func() { a.sourceConnectionStatus(c) })
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.source"), a.Tr("settings.open_source_under_apache_2_0_code_issues_and_rel"), "", func() {
			if a.sourceSettingsButton(c, "↗ "+a.Tr("settings.github")).Clicked() {
				a.Dispatch(Action{Kind: "OpenLink", Value: "https://github.com/charlzyx/orbit"})
			}
		})
	})
	if a.Options.Updater != nil && a.Options.Updater.Snapshot().Enabled {
		a.updateSettings(c)
	}
}
func (a *App) sourceConnectionStatus(c *ui.Context) {
	connected := a.Model.Connected || (a.Runtime.Active != nil && a.Runtime.Active.Alive)
	label, color := a.Tr("settings.not_running"), a.Theme.Error
	if connected {
		label = a.Tr("status.connected")
		color = a.Theme.Success
	}
	ui.Row(c).Gap(a.Theme.Space(6)).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(7, 7).Radius(3.5).Background(color)
		ui.Text(c, label).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextSecondary)
	})
}
