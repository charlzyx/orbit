package orbitui

import (
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
	"strings"
)

// Decode the installed applications' already-loaded icons once on the existing
// discovery worker. Painting only reads this immutable result; no file IO or decoding.
func decodeOpenAppIcons(apps []native.ExternalApp) map[string]ui.ImageSource {
	out := make(map[string]ui.ImageSource)
	for _, app := range apps {
		if len(app.Icon) == 0 {
			continue
		}
		if icon, err := ui.DecodeBitmap(app.Icon); err == nil {
			out[app.ID] = icon
		}
	}
	return out
}
func (a *App) headerOpenIn(c *ui.Context) {
	if len(a.OpenApps) == 0 {
		return
	}
	t := a.Theme
	preferred := a.OpenPrefs.Preferred(a.Model.Workspace, a.OpenApps)
	ui.Row(c.Key("header-open-in")).Height(t.Px(28)).Gap(0).AlignItems(ui.Center).Background(t.Elevated).Border(1, t.Border).Radius(t.Px(8)).Shadow(0, 1, 1, 0, t.composerContact()).Children(func() {
		if preferred != nil {
			b := ui.ButtonBase(c.Key("preferred")).Label(preferred.Label).Tooltip(strings.ReplaceAll(a.text("open_in.open_in_app", "Open in %{app}"), "%{app}", preferred.Label)).Height(t.Px(28)).Padding(0, t.Space(8)).Center()
			b.Children(func() {
				if icon := a.openAppIcons[preferred.ID]; icon != nil {
					ui.Image(c, icon).Size(t.Px(18), t.Px(18))
				} else {
					a.icon(c, "icons/folder.svg", t.Px(18), t.TextSecondary)
				}
			})
			if b.Hovered() {
				b.Background(t.Hover)
			}
			if b.Clicked() {
				a.Dispatch(Action{Kind: "OpenInApp", Value: preferred.ID})
			}
			ui.Box(c).Width(1).Height(t.Px(20)).Background(t.Border)
		}
		a.iconButton(c, "Open in…", "icons/chevron-down.svg").Size(t.Px(22), t.Px(28)).Menu(func(m *ui.Menu) {
			for _, app := range a.OpenApps {
				if m.Item(app.Label).Checked(preferred != nil && app.ID == preferred.ID).Chosen() {
					a.Dispatch(Action{Kind: "OpenInApp", Value: app.ID})
				}
			}
			for _, entry := range []struct{ label, kind string }{{"File manager", "RevealWorkspace"}, {"Terminal", "ToggleTerminal"}, {"Files", "OpenFiles"}, {"Git", "OpenGit"}} {
				if m.Item(entry.label).Chosen() {
					a.command(entry.kind, "")
				}
			}
		})
	})
}
