package orbitui

import (
	"github.com/egoist/mygo/ui"
	"math"
)

// GPUI elements/text.rs::request_layout resolves phi() directly without
// using the rounded TextStyle helper. Fractional line boxes are preserved.
// Button frames override this with 1.0, and card titles use explicit 20px.
func usageLineHeight(size float32) float32 { return size * 1.618034 }

func (a *App) usageColumnWidth(c *ui.Context) float32 {
	width, _ := c.Size()
	if a.SidebarVisible {
		width -= a.SidebarWidth
	}
	// Host feature_card owns the leading inset and its two border edges.
	return min(float32(1180), width-a.Theme.Space(12)-2)
}
func (a *App) setUsageMode(mode string) {
	s := &a.usageState
	if s.mode == mode {
		return
	}
	s.mode = mode
	s.scroll.X, s.scroll.Y = 0, 0
	a.saveUsagePrefs()
}

// raised_fill in theme/tokens.rs changes HSL lightness only, preserving hue
// and saturation, then adjusts alpha for translucent source colors.
func usageRaisedColors(color ui.Color) (ui.Color, ui.Color) {
	shift := func(delta float64) ui.Color {
		r, g, b := float64(color.R)/255, float64(color.G)/255, float64(color.B)/255
		lo, hi := min(r, g, b), max(r, g, b)
		light := (lo + hi) / 2
		target := max(0.0, min(1.0, light+delta))
		sat := 0.0
		if hi > lo {
			sat = (hi - lo) / (1 - math.Abs(2*light-1))
		}
		chroma := (1 - math.Abs(2*target-1)) * sat
		component := func(v float64) uint8 {
			out := target
			if hi > lo {
				out = target - chroma/2 + (v-lo)/(hi-lo)*chroma
			}
			return uint8(math.Round(max(0.0, min(1.0, out)) * 255))
		}
		return ui.Color{R: component(r), G: component(g), B: component(b), A: color.A}
	}
	top, bottom := shift(.02), shift(-.012)
	if float32(color.A)/255 < .95 {
		top = top.Alpha(1.12)
		bottom = bottom.Alpha(.94)
	}
	return top, bottom
}
func (a *App) usageRaised(button ui.Element, fill ui.Color, shadow bool) {
	top, bottom := usageRaisedColors(fill)
	button.Gradient(top, bottom, 180)
	if shadow {
		contact := a.Theme.composerContact()
		button.Shadow(0, 1, 1, 0, contact.Alpha(.45)).Shadow(0, 1, 3, -1, contact.Alpha(.25))
	}
}
func (a *App) usageButtonFrame(c *ui.Context, label string) ui.Element {
	t := a.Theme
	return ui.ButtonBase(c).Label(label).Height(t.Px(28)).PaddingX(t.Space(8)).Gap(t.Space(4)).Radius(t.Px(8)).AlignItems(ui.Center).NoWrap()
}
func (a *App) usageTextButton(c *ui.Context, label, icon string, enabled bool) ui.Element {
	t := a.Theme
	b := a.usageButtonFrame(c, label).Disabled(!enabled)
	fg := t.TextSecondary
	if !enabled {
		fg = t.TextMuted
	} else if b.Hovered() {
		b.Background(t.Hover)
		fg = t.Text
	}
	b.Children(func() {
		if icon != "" {
			a.icon(c, icon, t.Px(14), fg)
		}
		ui.Text(c, label).FontSize(t.Px(14)).FixedLineHeight(t.Px(14)).TextColor(fg).NoWrap()
	})
	return b
}
func (a *App) usageChip(c *ui.Context, label, icon string, active, inset bool) ui.Element {
	t := a.Theme
	b := a.usageButtonFrame(c, label).Border(1, t.Border)
	fg, leading, fill := t.TextSecondary, t.TextMuted, t.Elevated
	if inset {
		fill = t.Background
	}
	if active {
		fg, leading, fill = t.Text, t.Text, t.Active
		b.BorderColor(t.BorderStrong())
	} else if b.Hovered() {
		fill = t.Hover
	}
	a.usageRaised(b, fill, true)
	b.Children(func() {
		if icon != "" {
			a.icon(c, icon, t.Px(14), leading)
		}
		ui.Text(c, label).FontSize(t.Px(14)).FixedLineHeight(t.Px(14)).TextColor(fg).MaxWidth(180).MaxLines(1)
		a.icon(c, "icons/chevron-down.svg", t.Px(12), fg)
	})
	return b
}
func (a *App) usageSegments(c *ui.Context, key string, options []string, active string, inset bool, label func(string) string, enabled func(string) bool, pick func(string)) {
	t := a.Theme
	ui.Row(c.Key("usage-segments:" + key)).Label("Usage " + key + " segments").Gap(0).AlignItems(ui.Center).Children(func() {
		for i, option := range options {
			text := label(option)
			on := enabled == nil || enabled(option)
			selected := option == active
			b := a.usageButtonFrame(c.Key(option), text).Border(1, t.Border).Radius(0).Disabled(!on && !selected)
			if len(options) == 1 {
				b.Radius(t.Px(8))
			} else if i == 0 {
				b.Radius(t.Px(8), 0, 0, t.Px(8))
			} else if i == len(options)-1 {
				b.Radius(0, t.Px(8), t.Px(8), 0)
			}
			if i > 0 {
				b.BorderWidth(1, 1, 1, 0)
			}
			fill, fg := t.Elevated, t.TextSecondary
			if inset {
				fill = t.Background
			}
			if selected {
				fill, fg = t.Active, t.Text
			} else if !on {
				fg = t.TextDisabled
			} else if b.Hovered() {
				fill = t.Hover
			}
			if !on && !selected {
				b.Background(fill)
			} else {
				a.usageRaised(b, fill, false)
			}
			b.Children(func() {
				text := ui.Text(c, text).FontSize(t.Px(14)).FixedLineHeight(t.Px(14)).TextColor(fg).NoWrap()
				if selected {
					text.FontWeight(500)
				}
			})
			// Selecting an already-active segment is a source no-op. On mouse-up the
			// selected branch prevents a second mutation after pointer-down selection.
			if on && !selected && (b.Pressed() || b.Clicked()) {
				pick(option)
			}
		}
	})
}
