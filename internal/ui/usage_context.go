package orbitui

import (
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// Handle only secondary presses at the table boundary. Primary clicks,
// double-clicks, selection and keyboard events remain native Table behavior.
func (a *App) usageSessionContext(c *ui.Context, table ui.Element, tableHandle *ui.Handle, rows []usageTableRow, bounds map[int]*ui.Handle) {
	s := &a.usageState
	m := s.menu("context")
	rowRect := func(i int) ui.Rect {
		if bounds[i] == nil {
			return ui.Rect{}
		}
		r := bounds[i].Bounds(c)
		tableRect := tableHandle.Bounds(c)
		return ui.Rect{X: tableRect.X, Y: r.Y - (40-r.H)/2, W: tableRect.W, H: 40}
	}
	table.HandleInput(func(ev ui.InputEvent) bool {
		if ev.Kind == ui.InputPointerUp && ev.Button == 1 && s.contextFresh {
			s.contextFresh = false
			return true
		}
		if ev.Kind != ui.InputPointerDown || ev.Button != 1 {
			return false
		}
		y := tableHandle.Bounds(c).Y + ev.Y
		for i := range rows {
			r := rowRect(i)
			if r.H > 0 && y >= r.Y && y < r.Y+r.H {
				s.closeUsagePopovers("context")
				s.contextRow, s.contextFresh = i, true
				m.Open = true
				s.filterMenu = "context"
				return true
			}
		}
		return false
	})
}

func (a *App) usageSessionContextPopup(c *ui.Context, anchor ui.Element, row usageTableRow) {
	s := &a.usageState
	m := s.menu("context")
	closeMenu := func() {
		m.Open = false
		if s.filterMenu == "context" {
			s.filterMenu = ""
		}
	}
	fresh := s.contextFresh
	ui.Overlay(c.Key("usage-session-context-overlay"), func() {
		panel := ui.Column(c).AttachTo(anchor, ui.AnchorTopStart, ui.AnchorTopStart).Margin(0).Padding(a.Theme.Space(4), 0).Gap(a.Theme.Space(1)).Background(a.Theme.Elevated).Border(1, a.Theme.BorderStrong()).Radius(a.Theme.Px(12)).Label("Usage session context menu")
		a.usagePopupShadow(panel)
		// Plain pointer rows match the source: they do not create keyboard picker
		// semantics or move focus away from the underlying native table.
		panel.Clicked()
		panel.Children(func() {
			item := func(key, fallback string, vars map[string]string, enabled bool, run func()) {
				label := a.usageText(key, fallback, vars)
				b := ui.Row(c.Key(key)).Height(a.Theme.Px(23)+2*a.Theme.Space(4)).Padding(a.Theme.Space(4), a.Theme.Space(6)).Margin(0, a.Theme.Space(4)).Radius(a.Theme.Px(6)).AlignItems(ui.Center).KeepFocus().Label(label).Disabled(!enabled)
				if b.Hovered() {
					b.Background(a.Theme.Hover)
				}
				color := a.Theme.TextSecondary
				if b.Hovered() {
					color = a.Theme.Text
				}
				b.Children(func() { ui.Text(c, label).FontSize(a.Theme.Px(14)).NoWrap().TextColor(color) })
				if b.Clicked() {
					run()
				}
			}
			item("usage.ctx_open_session", "Open session", nil, row.Path != "", func() { closeMenu(); a.openUsageSession(row) })
			item("usage.ctx_scope_to_session", "Scope to this session", nil, true, func() {
				closeMenu()
				s.filter.SessionPath = row.Path
				s.filter.Session = ""
				if row.Path == "" {
					s.filter.Session = row.ID
				}
			})
			ui.Box(c).Height(1).AlignSelf(ui.Stretch).Margin(a.Theme.Space(6), 0).Background(a.Theme.Border)
			item("usage.ctx_copy_session_id", "Copy session ID", nil, true, func() { a.Dispatch(Action{Kind: "CopyText", Value: row.ID}) })
			ui.Box(c).Height(1).AlignSelf(ui.Stretch).Margin(a.Theme.Space(6), 0).Background(a.Theme.Border)
			item("usage.ctx_filter_by_provider", "Filter by provider: %{name}", map[string]string{"name": usageProviderLabel(row.Provider)}, true, func() { s.toggleFilter("providers", usageFilterOption{Key: row.Provider}) })
			if row.Model != "" {
				item("usage.ctx_filter_by_model", "Filter by model: %{name}", map[string]string{"name": usageModelPair(row.Provider, row.Model)}, true, func() {
					s.toggleFilter("models", usageFilterOption{Model: workbench.UsageModelRef{Provider: row.Provider, Model: row.Model}})
				})
			}
			item("usage.ctx_filter_by_workspace", "Filter by workspace", nil, true, func() { s.toggleFilter("workspaces", usageFilterOption{Key: row.Workspace}) })
		})
		if !fresh && panel.PressedOutside() {
			closeMenu()
		}
	})
}

// Frozen ModalSurface elevation; the last layer is the contact line.
func (a *App) usagePopupShadow(panel ui.Element) {
	alpha1, alpha2, contact := float32(.12), float32(.08), float32(.12)
	if !a.Theme.Dark {
		alpha1, alpha2, contact = .06, .06, .04
	}
	panel.Shadow(0, 2, 3, 0, ui.RGBA(0, 0, 0, alpha1)).Shadow(0, 3, 6, 0, ui.RGBA(0, 0, 0, alpha2)).Shadow(0, 6, 12, 0, ui.RGBA(0, 0, 0, .04)).Shadow(0, 1, 0, 0, ui.RGBA(0, 0, 0, contact))
}

func (a *App) usageNativePopover(c *ui.Context, anchor ui.Element, open *bool, width float32, body func()) {
	ui.PopoverBase(c, anchor, open, func(panel ui.Element) {
		panel.Margin(4, 0, 0, 0).Background(a.Theme.Elevated).Border(1, a.Theme.BorderStrong()).Radius(a.Theme.Px(12))
		if width > 0 {
			panel.Width(width)
		}
		a.usagePopupShadow(panel)
		body()
	})
}
