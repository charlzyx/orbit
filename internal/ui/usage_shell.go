package orbitui

import "github.com/egoist/mygo/ui"

// app/view.rs::feature_card places the full-page Usage surface below the
// shared top bar, with only a leading inset. The page owns its inner padding.
func (a *App) usageShell(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("usage-card-inset")).Grow(1).MinHeight(0).FillWidth().Padding(0, 0, 0, t.Space(12)).Children(func() {
		ui.Column(c.Key("usage-card")).Label("Usage pane").Fill().MinHeight(0).Border(1, t.Border).Radius(t.Px(12)).Clip().Children(func() {
			a.renderFeaturePage(c)
		})
	})
}

// usage/view.rs::top_bar_leading keeps the interactive Back control outside
// the window-drag region, followed by the page glyph and larger title.
func (a *App) usageHeaderLeading(c *ui.Context) {
	t := a.Theme
	ui.Row(c.Key("usage-top-leading")).Grow(1).MinWidth(0).FillHeight().Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
		back := ui.ButtonBase(c.Key("usage-top-back")).KeepFocus().Label("Back to task").Tooltip(a.text("view.back", "Back")).Height(t.Px(28)).PaddingX(t.Space(8)).Gap(t.Space(4)).Radius(t.Px(8)).AlignItems(ui.Center)
		if back.Hovered() {
			back.Background(t.Hover)
		}
		back.Children(func() {
			a.icon(c, "icons/arrow-left.svg", t.Px(14), t.TextSecondary)
			ui.Text(c, a.text("view.back", "Back")).FontSize(t.Px(14)).TextColor(t.TextSecondary)
		})
		// Source closes on pointer-down rather than waiting for mouse-up.
		if back.Pressed() || back.Clicked() {
			a.Navigate("chat")
		}
		ui.Row(c.Key("usage-top-title")).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
			a.icon(c, "icons/chart-analysis.svg", t.Px(16), t.TextSecondary)
			ui.Text(c, a.text("view.usage", "Usage")).FontSize(t.Px(16)).FontWeight(500).TextColor(t.Text).MaxLines(1)
		})
		ui.Box(c.Key("usage-top-drag")).Grow(1).MinWidth(0).FillHeight().DragWindow().Label("Usage window drag region")
	})
}
