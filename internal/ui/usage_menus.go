package orbitui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type usageMenuState struct{ Open bool }

func (s *usageState) menu(key string) *usageMenuState {
	if s.menus == nil {
		s.menus = map[string]*usageMenuState{}
	}
	if s.menus[key] == nil {
		s.menus[key] = &usageMenuState{}
	}
	return s.menus[key]
}
func (s *usageState) closeUsagePopovers(except string) {
	for key, m := range s.menus {
		if key != except {
			m.Open = false
		}
	}
	if except != "range" {
		s.rangeOpen = false
		s.customOpen = false
	}
	for kind, v := range s.filterViews {
		if kind != except {
			v.Open = false
		}
	}
}
func (a *App) usagePopover(c *ui.Context, key, label, icon string, enabled bool, width float32, body func(func())) {
	s := &a.usageState
	m := s.menu(key)
	anchor := a.button(c, label, icon, false).AlignSelf(ui.Start).Border(1, a.Theme.Border).Disabled(!enabled)
	if strings.HasPrefix(key, "page-size:") {
		anchor.Label("Rows per page " + strings.TrimPrefix(key, "page-size:"))
	}
	if key != "export" {
		anchor.Children(func() { a.icon(c, "icons/chevron-down.svg", 12, a.Theme.TextMuted) })
	}
	if anchor.Clicked() {
		m.Open = !m.Open
		s.closeUsagePopovers(key)
		if m.Open {
			s.filterMenu = key
		} else {
			s.filterMenu = ""
		}
	}
	if !enabled {
		m.Open = false
		if s.filterMenu == key {
			s.filterMenu = ""
		}
	}
	wasOpen := m.Open
	closeMenu := func() {
		m.Open = false
		if s.filterMenu == key {
			s.filterMenu = ""
		}
	}
	a.usageNativePopover(c.Key("usage-popup:"+key), anchor, &m.Open, width, func() {
		root := ui.Column(c).AlignSelf(ui.Stretch).Gap(0).Label("Usage " + key + " menu")
		confirm := root.Shortcut(0, ui.KeyEnter)
		confirm = root.OverlayShortcut(0, ui.KeyEnter) || confirm
		if confirm {
			closeMenu()
		}
		root.Children(func() { body(closeMenu) })
	})
	if wasOpen && !m.Open && s.filterMenu == key {
		s.filterMenu = ""
	}
}
func (a *App) usageExportControl(c *ui.Context) {
	s := &a.usageState
	enabled := s.loaded && !s.projecting && s.computed != "" && (s.projection.Summary.Requests > 0 || len(s.projection.Tools) > 0)
	a.usagePopover(c, "export", a.text("usage.export", "Export"), "icons/upload.svg", enabled, 268, func(closeMenu func()) {
		for _, format := range []string{"csv", "json"} {
			title := a.text("usage.export_"+format, format)
			hint := a.text("usage.export_"+format+"_hint", "")
			if a.usagePickerRow(c.Key(format), title, hint, false).Label(title).Clicked() {
				closeMenu()
				a.Dispatch(Action{Kind: "ExportUsage", Value: format, Data: map[string]any{"index": s.index, "filter": cloneUsageFilter(s.filter), "options": workbench.UsageExportOptions{Preset: s.preset, Granularity: s.granularity, SessionSearch: s.table("sessions").Search, Location: time.Local, Now: s.now(), Insights: a.usageInsightTexts(usageInsights(s.projection, s.preset == "all"))}}})
			}
		}
	})
}
func (a *App) usageColumnControl(c *ui.Context, kind string, all []ui.TableColumn, required string) {
	s := &a.usageState
	t := s.table(kind)
	persist := func() {
		if kind == "sessions" {
			s.hiddenColumns = append([]string(nil), t.Hidden...)
			a.saveUsagePrefs()
		}
	}
	a.usagePopover(c, "columns:"+kind, a.text("usage.columns", "Columns"), "", true, 0, func(closeMenu func()) {
		a.usageMenuTitle(c, a.text("usage.columns", "Columns"))
		for _, col := range all {
			if col.ID == required || col.ID == "open" || (kind == "sessions" && col.ID == "tools") {
				continue
			}
			if a.usagePickerRow(c.Key(col.ID), col.Title, "", !slices.Contains(t.Hidden, col.ID)).Label("Toggle " + kind + " column " + col.ID).Clicked() {
				t.toggleColumn(col.ID, required)
				persist()
			}
		}
		a.usageMenuFooter(c, func() {
			if a.usageFooterButton(c, a.text("usage.show_all", "Show all")).Clicked() {
				t.Hidden = nil
				persist()
			}
		})
	})
}
func (a *App) usagePageSizeControl(c *ui.Context, kind string) {
	s := &a.usageState
	t := s.table(kind)
	a.usagePopover(c, "page-size:"+kind, fmt.Sprint(t.Size), "", true, 0, func(closeMenu func()) {
		a.usageMenuTitle(c, a.text("usage.rows_per_page", "Rows per page"))
		for _, size := range []int{10, 25, 50, 100} {
			if a.usagePickerRow(c.Key(size), fmt.Sprint(size), "", t.Size == size).Label(fmt.Sprintf("Rows per page %d", size)).Clicked() && t.Size != size {
				t.setPageSize(size)
				if kind == "sessions" {
					s.pageSize = size
					a.saveUsagePrefs()
				}
				closeMenu()
			}
		}
	})
}

func (a *App) usageMenuTitle(c *ui.Context, label string) {
	ui.Column(c).AlignSelf(ui.Stretch).Padding(0, a.Theme.Space(2), a.Theme.Space(4)).Children(func() {
		ui.Row(c).Height(a.Theme.Px(20)).Padding(0, a.Theme.Px(8)).AlignItems(ui.Center).Children(func() { ui.Text(c, label).SingleLine().FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted) })
	})
}
func (a *App) usageMenuFooter(c *ui.Context, body func()) {
	ui.Row(c).AlignSelf(ui.Stretch).Padding(a.Theme.Space(4)).Gap(a.Theme.Space(4)).AlignItems(ui.Center).BorderWidth(1, 0, 0, 0).BorderColor(a.Theme.Border).Children(body)
}
func (a *App) usageFooterButton(c *ui.Context, label string) ui.Element {
	button := ui.ButtonBase(c).Height(a.Theme.Px(28)).Grow(1).Basis(0).Padding(0, a.Theme.Space(8)).Radius(a.Theme.Px(8)).Center().Label(label)
	color := a.Theme.TextMuted
	if button.Hovered() {
		button.Background(a.Theme.Hover)
		color = a.Theme.TextSecondary
	}
	button.Children(func() { ui.Text(c, label).SingleLine().FontSize(a.Theme.Px(14)).TextColor(color) })
	return button
}
