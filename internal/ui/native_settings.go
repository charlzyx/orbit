package orbitui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type nativeSettingsState struct {
	selectMenu                         settingsSelectState
	worktreeInputsLoaded               bool
	worktreeDirectory, worktreeScript  string
	dependencies                       []native.Dependency
	installation                       native.PiInstallation
	temp                               native.TempFiles
	runtimeLoaded, privacyLoaded, busy bool
	latest, output                     string
}

func (a *App) loadRuntime() {
	s := &a.nativeSettings
	if s.busy {
		return
	}
	s.busy = true
	a.Async(func() func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		deps := native.CheckDependencies(ctx)
		p := native.DetectPiInstallation(ctx)
		return func() { s.dependencies = deps; s.installation = p; s.busy = false; s.runtimeLoaded = true }
	})
}
func (a *App) runtimeNative(c *ui.Context) {
	s := &a.nativeSettings
	if !s.runtimeLoaded && !s.busy {
		a.loadRuntime()
	}
	a.section(c, "Dependencies", func() {
		if s.busy {
			ui.Spinner(c).Size(18, 18)
		}
		for _, d := range s.dependencies {
			d := d
			detail := d.Version + " · " + d.Path
			if !d.Installed {
				detail = "Not installed"
			}
			if d.Error != "" {
				detail = d.Error
			}
			a.setting(c, d.Name, detail, func() {
				if !d.Installed {
					if a.button(c, "Copy install command", "icons/copy.svg", false).Clicked() {
						c.Services().WriteClipboard(d.InstallHint)
					}
				}
			})
		}
		if a.button(c, "Recheck dependencies", "icons/refresh.svg", false).Clicked() {
			a.loadRuntime()
		}
	})
	a.section(c, "Pi installation", func() {
		a.setting(c, "Managed by", string(s.installation.Kind), nil)
		if s.latest != "" {
			a.setting(c, "Latest available", s.latest, nil)
		}
		ui.Row(c).Gap(8).Children(func() {
			if a.button(c, "Check update", "icons/refresh.svg", false).Disabled(s.busy).Clicked() {
				a.checkPiUpdate()
			}
			if !s.installation.Installed() {
				if a.button(c, "Install pi…", "icons/download.svg", false).Disabled(s.busy).Clicked() {
					a.confirmPiInstall(false)
				}
			} else if native.PlanPiUpdate(s.installation, s.latest) == native.PiMigrateNPM {
				if a.button(c, "Migrate npm installation…", "", false).Disabled(s.busy).Clicked() {
					a.confirmPiInstall(true)
				}
			} else if native.PlanPiUpdate(s.installation, s.latest) == native.PiUpdateManaged {
				if a.button(c, "Update pi…", "", false).Disabled(s.busy).Clicked() {
					a.Confirm("Update pi?", "Update the official managed pi installation. Running tasks should finish first.", func() { a.runPiOperation(false, false) })
				}
			}
			if a.button(c, "Refresh model catalog", "icons/refresh.svg", false).Disabled(a.Models != nil && a.Models.Refreshing).Clicked() {
				a.Dispatch(Action{Kind: "RefreshModels"})
			}

		})
		if s.output != "" {
			ui.Scroll(c).MaxHeight(240).Children(func() { ui.Text(c, s.output).Font(a.Fonts.CodeFamily).Selectable() })
		}
	})
}
func (a *App) checkPiUpdate() {
	s := &a.nativeSettings
	if s.busy {
		return
	}
	s.busy = true
	a.Async(func() func() {
		v, e := native.LatestPiVersion(context.Background(), nil)
		return func() {
			s.busy = false
			s.latest = v
			a.save(e)
			if e == nil && native.PlanPiUpdate(s.installation, v) == native.PiUpToDate {
				a.Notify("Pi is up to date", false)
			}
		}
	})
}
func (a *App) confirmPiInstall(migrate bool) {
	detail := "Download and run the official pi.dev installer. This installs pi for your user."
	if migrate {
		detail += " The existing global npm installation will be removed and replaced; your sessions and credentials are preserved."
	}
	a.Confirm("Install pi?", detail, func() { a.runPiOperation(true, migrate) })
}
func (a *App) runPiOperation(install, migrate bool) {
	release := func() {}
	if a.Options.BeginRuntimeOperation != nil {
		var e error
		release, e = a.Options.BeginRuntimeOperation()
		if e != nil {
			a.save(e)
			return
		}
	}
	s := &a.nativeSettings
	s.busy = true
	a.Async(func() func() {
		defer release()
		var r native.CommandOutput
		var e error
		if install {
			r, e = native.InstallPi(context.Background(), migrate)
		} else {
			r, e = native.RunPiUpdate(context.Background())
		}
		return func() {
			s.busy = false
			s.output = strings.TrimSpace(string(r.Stdout) + "\n" + string(r.Stderr))
			a.save(e)
			a.loadRuntime()
			if e == nil {
				a.Dispatch(Action{Kind: "RuntimeChanged"})
			}
		}
	})
}
func (a *App) loadPrivacy() {
	s := &a.nativeSettings
	a.Async(func() func() { v := native.ScanTempFiles(); return func() { s.temp = v; s.privacyLoaded = true } })
}
func (a *App) privacyNative(c *ui.Context) {
	s := &a.nativeSettings
	if !s.privacyLoaded {
		s.privacyLoaded = true
		a.loadPrivacy()
	}
	a.sourceTempSection(c, false)
	a.sourceTempSection(c, true)
}
func (a *App) sourceTempSection(c *ui.Context, ai bool) {
	s := &a.nativeSettings
	prefix := "settings.temporary_files"
	dir := s.temp.Directory
	entries := s.temp.Entries
	bytes := s.temp.TotalBytes()
	if dir == "" {
		dir = native.TempDirectory()
	}
	if ai {
		prefix = "settings.ai_scratch_files"
		dir = native.TempRoot()
		entries = s.temp.AIEntries
		bytes = s.temp.AITotalBytes()
	}
	a.sourceSettingsSection(c, a.Tr(prefix), a.Tr(prefix+"_description"), func() {
		a.sourceSettingsRow(c, a.Tr(prefix+"_location"), a.Tr(prefix+"_location_description"), dir, func() {
			if a.sourceSettingsButton(c, a.Tr("settings.temporary_files_open")).Label(map[bool]string{false: "Open Orbit temporary folder", true: "Open AI scratch folder"}[ai]).Height(a.Theme.Px(32)).Clicked() {
				a.Dispatch(Action{Kind: "RevealPath", Value: dir})
			}
		})
		a.sourceSettingsDivider(c)
		if len(entries) == 0 {
			a.sourceSettingsRow(c, a.Tr(prefix+"_empty"), "", "", nil)
			return
		}
		for i, entry := range entries {
			if i >= 20 {
				break
			}
			if i > 0 {
				a.sourceSettingsDivider(c)
			}
			kind := a.Tr("settings.temporary_files_kind_file")
			if entry.IsDir {
				kind = a.Tr("settings.temporary_files_kind_folder")
			}
			meta := kind + " · " + sourceTempSize(entry.Bytes) + " · " + sourceTempAge(entry.Modified, time.Now())
			if entry.Active {
				meta += " · " + a.Tr("settings.temporary_files_in_use")
			}
			a.sourceSettingsRow(c, entry.Name, "", meta, nil)
		}
		if len(entries) > 20 {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, strings.ReplaceAll(a.Tr("settings.temporary_files_more"), "%{count}", fmt.Sprint(len(entries)-20)), "", "", nil)
		}
		a.sourceSettingsDivider(c)
		summary := strings.NewReplacer("%{count}", fmt.Sprint(len(entries)), "%{size}", sourceTempSize(bytes)).Replace(a.Tr("settings.temporary_files_summary"))
		a.sourceSettingsRow(c, a.Tr(prefix+"_clean"), a.Tr(prefix+"_clean_description"), summary, func() {
			if a.sourceSettingsButton(c, a.Tr(prefix+"_clean")).Label(map[bool]string{false: "Clean Orbit temporary files", true: "Clean AI scratch files"}[ai]).Height(a.Theme.Px(32)).Clicked() {
				a.cleanTemp(ai)
			}
		})
	})
}
func sourceTempSize(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
func sourceTempAge(at, now time.Time) string {
	seconds := int64(now.Sub(at).Seconds())
	switch {
	case seconds < 60:
		return "now"
	case seconds < 3600:
		return fmt.Sprintf("%dm", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%dh", seconds/3600)
	case seconds < 2592000:
		return fmt.Sprintf("%dd", seconds/86400)
	case seconds < 31536000:
		return fmt.Sprintf("%dmo", seconds/2592000)
	default:
		return fmt.Sprintf("%dy", seconds/31536000)
	}
}
func (a *App) cleanTemp(ai bool) {
	snapshot := a.nativeSettings.temp
	a.Confirm("Permanently delete temporary files?", "Eligible temporary files will be removed permanently. Active and recent files are retained.", func() {
		a.Async(func() func() {
			out := native.CleanTempFiles(snapshot, ai)
			return func() {
				a.Notify(fmt.Sprintf("Removed %d files (%d bytes); %d active retained; %d failures", out.Removed, out.Bytes, len(out.SkippedActive), len(out.Failed)), len(out.Failed) > 0)
				for _, f := range out.Failed {
					a.Notify(f.Path+": "+f.Error, true)
				}
				a.loadPrivacy()
			}
		})
	})
}

func (a *App) updateSettings(c *ui.Context) {
	u := a.Options.Updater
	if u == nil {
		return
	}
	s := u.Snapshot()
	if !s.Enabled {
		ui.Text(c, "Signed updates are not configured for this build.").TextColor(a.Theme.TextMuted)
		return
	}
	ui.Text(c, "Status: "+string(s.Status)).TextColor(a.Theme.TextMuted)
	if s.Error != "" {
		ui.Text(c, s.Error).TextColor(a.Theme.Error)
	}
	if s.Available != nil {
		ui.Text(c, "Orbit "+s.Available.Version).FontWeight(500)
		ui.Text(c, s.Available.Notes).MaxLines(8)
	}
	if s.Status == native.UpdateDownloading {
		ui.Text(c, fmt.Sprintf("%d / %d bytes", s.Downloaded, s.Total))
	}
	if s.Status == native.UpdateStaged && s.CanInstall {
		if a.button(c, "Install and restart…", "icons/refresh.svg", true).Clicked() {
			a.Confirm("Install verified update and restart?", "Running tasks will be stopped after the verified installer is launched.", func() {
				a.Async(func() func() {
					e := u.Install(context.Background())
					return func() {
						if e != nil {
							a.save(e)
						} else {
							mygo.App.Quit()
						}
					}
				})
			})
		}
	}
}
