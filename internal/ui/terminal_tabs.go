package orbitui

import (
	"fmt"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/egoist/mygo/ui"
)

// Terminal structure is changed synchronously on the owner UI. Resolve stable
// tab IDs at activation, not indices captured before another tab was removed.
func terminalTabIndex(p *orbiterm.Panel, id uint64) int {
	for i, tab := range p.Snapshot().Tabs {
		if tab.ID == id {
			return i
		}
	}
	return -1
}
func closeTerminalTab(p *orbiterm.Panel, id uint64) error {
	if i := terminalTabIndex(p, id); i >= 0 {
		return p.CloseTab(i)
	}
	return nil
}
func (a *App) terminalTab(c *ui.Context, p *orbiterm.Panel, tab orbiterm.TabInfo, active bool) {
	t := a.Theme
	chip := ui.Row(c).Label(fmt.Sprintf("Terminal tab %d", tab.ID)).Height(t.Px(28)).MaxWidth(orbiterm.TabMaximumWidth).Shrink(0).PaddingX(t.Space(6)).Gap(t.Space(4)).Radius(t.Px(6)).AlignItems(ui.Center).Tooltip(tab.Tooltip)
	if active {
		chip.Background(t.Selection)
	} else if chip.Hovered() {
		chip.Background(t.Hover)
	}
	chip.Children(func() {
		// Sibling controls deliberately avoid the source's close→select bubbling trap.
		selectTab := ui.ButtonBase(c.Key("select")).Label(tab.Label).KeepFocus().Grow(1).MinWidth(0).Height(t.Px(28)).Gap(t.Space(4)).Justify(ui.Start)
		selectTab.Children(func() {
			a.icon(c, tab.Icon, t.Px(12), t.TextSecondary)
			ui.Text(c, tab.Label).FontSize(t.Px(12)).SingleLine().MinWidth(0).Grow(1)
		})
		if selectTab.Clicked() {
			if i := terminalTabIndex(p, tab.ID); i >= 0 {
				p.Select(i)
			}
		}
		close := a.iconButton(c.Key("close"), fmt.Sprintf("%s %d", a.text("terminal.close_tab", "Close tab"), tab.ID), "icons/x.svg").Size(t.Px(20), t.Px(20)).KeepFocus().Tooltip(a.text("terminal.close_tab", "Close tab"))
		if close.Clicked() {
			a.save(closeTerminalTab(p, tab.ID))
		}
	})
}
