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
	a.section(c, "Temporary files", func() {
		a.setting(c, "Orbit temporary data", fmt.Sprintf("%d entries · %d bytes", len(s.temp.Entries), s.temp.TotalBytes()), func() {
			if a.button(c, "Clean…", "icons/trash.svg", false).Clicked() {
				a.cleanTemp(false)
			}
		})
		a.setting(c, "Pi AI scratch files", fmt.Sprintf("%d entries · %d bytes", len(s.temp.AIEntries), s.temp.AITotalBytes()), func() {
			if a.button(c, "Clean…", "icons/trash.svg", false).Clicked() {
				a.cleanTemp(true)
			}
		})
		ui.Text(c, "Active process files and recently modified AI files are protected. Cleanup does not touch credentials, projects, or saved sessions.").TextColor(a.Theme.TextMuted)
		if a.button(c, "Refresh", "icons/refresh.svg", false).Clicked() {
			a.loadPrivacy()
		}
	})
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
