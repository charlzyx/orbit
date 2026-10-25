package orbitui

import (
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
	"strconv"
	"strings"
)

type onboardingState struct {
	installation                native.PiInstallation
	migrating                   bool
	enabled, loaded, refreshing bool
	dependencies                []native.Dependency
	facts                       []native.HostFact
	error                       string
}

func (a *App) setupBlocked() bool {
	s := &a.setup
	return s.enabled && (!s.loaded || s.error != "" || !native.AllRequiredInstalled(s.dependencies))
}
func (a *App) onboardingView(c *ui.Context) {
	s := &a.setup
	t := a.Theme
	ui.Scroll(c.Key("setup-page")).Fill().MinWidth(0).Children(func() {
		ui.Column(c).AlignSelf(ui.Center).WidthPercent(100).MaxWidth(560).Padding(32, 24).Gap(18).Children(func() {
			ui.Text(c, a.Tr("view.set_up_orbit")).FontSize(t.Px(26)).FontWeight(600)
			ui.Text(c, a.Tr("view.setup_missing_hint")).TextColor(t.TextSecondary)
			if !s.loaded {
				ui.Row(c).Gap(8).Children(func() { ui.Spinner(c).Size(16, 16); ui.Text(c, a.Tr("common.checking")) })
			}
			for _, d := range s.dependencies {
				d := d
				ui.Column(c.Key("setup-dependency:"+d.Bin)).AlignSelf(ui.Stretch).Padding(12, 14).Gap(8).Background(t.Elevated).Border(1, t.Border).Radius(10).Children(func() {
					ui.Row(c).AlignSelf(ui.Stretch).Gap(8).Children(func() {
						status := t.Error
						if d.Installed {
							status = t.Success
						}
						ui.Box(c).Size(9, 9).Radius(4.5).Background(status)
						ui.Text(c, d.Name).FontWeight(600)
						ui.Spacer(c)
						label := a.Tr("setup.optional")
						if d.Required {
							label = a.Tr("setup.required")
						}
						ui.Text(c, label).FontSize(t.Px(11)).TextColor(t.TextMuted)
					})
					if d.DetailKey != "" {
						ui.Text(c, a.Tr(d.DetailKey)).TextColor(t.TextSecondary)
					}
					if d.Installed {
						label := d.Version
						if label == "" {
							label = a.Tr("setup.installed")
						}
						ui.Text(c, label).TextColor(t.Success)
					} else if d.InstallHint != "" {
						ui.Row(c).AlignSelf(ui.Stretch).Gap(8).Children(func() {
							ui.Text(c, d.InstallHint).Font(a.Fonts.CodeFamily).Grow(1).Selectable()
							if a.iconButton(c, "Copy install command "+d.Bin, "icons/copy.svg").Clicked() {
								c.Services().WriteClipboard(d.InstallHint)
							}
						})
					}
					if d.Error != "" {
						ui.Text(c, d.Error).TextColor(t.Error).Selectable()
					}
				})
			}
			ui.Column(c.Key("setup-host-facts")).AlignSelf(ui.Stretch).Border(1, t.Border).Radius(10).Children(func() {
				for i, f := range s.facts {
					f := f
					row := ui.Row(c.Key(struct {
						Kind  string
						Index int
					}{"setup-fact", i})).AlignSelf(ui.Stretch).Padding(8, 14).Gap(12).AlignItems(ui.Center)
					row.Children(func() {
						label := f.Label
						if f.LabelKey != "" {
							label = a.Tr(f.LabelKey)
						}
						ui.Text(c, label).Width(88).Shrink(0).FontSize(t.Px(12)).TextColor(t.TextMuted)
						value := f.Value
						if f.ValueKey != "" {
							value = a.Tr(f.ValueKey)
						}
						ui.Text(c, value).Grow(1).MinWidth(0).MaxLines(1).Ellipsis("…").Tooltip(value).FontSize(t.Px(12)).TextColor(t.TextSecondary).Selectable()
						note := f.Note
						if f.NoteKey != "" {
							note = a.Tr(f.NoteKey)
						}
						if note != "" {
							color := t.TextMuted
							if f.Alert {
								color = t.Error
							}
							ui.Text(c, note).MaxWidth(150).MaxLines(2).FontSize(t.Px(12)).TextColor(color)
						}
					})
					if i < len(s.facts)-1 {
						ui.Box(c).AlignSelf(ui.Stretch).Height(1).Background(t.Border)
					}
				}
			})
			if s.installation.Kind == native.PiNPM {
				a.setupMigrationCard(c)
			}
			if s.error != "" {
				ui.Text(c, s.error).TextColor(t.Error).Selectable()
			}
			ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
				if s.loaded && s.error == "" {
					ui.Text(c, strings.ReplaceAll(a.Tr("setup.missing_count"), "%{count}", strconv.Itoa(native.MissingRequiredCount(s.dependencies)))).TextColor(t.TextMuted)
				}
				label := a.Tr("common.refresh")
				if s.refreshing {
					label = a.Tr("common.checking")
				}
				if a.button(c, label, "icons/refresh.svg", false).Label("Refresh setup").Disabled(s.refreshing || s.migrating).Clicked() {
					a.Dispatch(Action{Kind: "RefreshSetup"})
				}
			})
		})
	})
}
