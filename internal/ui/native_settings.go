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
	selectMenu                        settingsSelectState
	worktreeInputsLoaded              bool
	worktreeDirectory, worktreeScript string
	temp                              native.TempFiles
	privacyLoaded                     bool
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
	a.Confirm(a.Tr("go.privacy.confirm"), a.Tr("go.privacy.confirm_detail"), func() {
		a.Async(func() func() {
			out := native.CleanTempFiles(snapshot, ai)
			return func() {
				a.Notify(strings.NewReplacer("%{count}", fmt.Sprint(out.Removed), "%{bytes}", fmt.Sprint(out.Bytes), "%{active}", fmt.Sprint(len(out.SkippedActive)), "%{failed}", fmt.Sprint(len(out.Failed))).Replace(a.Tr("go.privacy.removed")), len(out.Failed) > 0)
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
		ui.Text(c, a.Tr("go.update.unconfigured")).TextColor(a.Theme.TextMuted)
		return
	}
	ui.Text(c, strings.ReplaceAll(a.Tr("go.update.status"), "%{status}", a.text("go.update.state_"+string(s.Status), string(s.Status)))).TextColor(a.Theme.TextMuted)
	if s.Error != "" {
		ui.Text(c, s.Error).TextColor(a.Theme.Error)
	}
	if s.Available != nil {
		ui.Text(c, "GoPi "+s.Available.Version).FontWeight(500)
		ui.Text(c, s.Available.Notes).MaxLines(8)
	}
	if s.Status == native.UpdateDownloading {
		ui.Text(c, strings.NewReplacer("%{downloaded}", fmt.Sprint(s.Downloaded), "%{total}", fmt.Sprint(s.Total)).Replace(a.Tr("go.update.bytes")))
	}
	if s.Status == native.UpdateStaged && s.CanInstall {
		if a.button(c, a.Tr("go.update.install"), "icons/refresh.svg", true).Clicked() {
			a.Confirm(a.Tr("go.update.confirm"), a.Tr("go.update.confirm_detail"), func() {
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
