package orbitui

import (
	"math"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// The source derives the header's next state from the current active key. In
// particular an already-default descending key resets to itself, not ascending.
func usageNextSort(kind string, current ui.SortOrder, column string) ui.SortOrder {
	if current.Column != column {
		return ui.SortOrder{Column: column}
	}
	if !current.Descending {
		return ui.SortOrder{Column: column, Descending: true}
	}
	return usageDefaultSort(kind)
}
func (a *App) setUsageTableSort(kind, column string) {
	s := &a.usageState
	t := s.table(kind)
	next := usageNextSort(kind, t.Sort, column)
	if next == t.Sort {
		return
	}
	t.Sort, t.observedSort = next, next
	if !strings.HasPrefix(kind, "breakdown:") {
		t.Page = 1
	}
	t.Selected = -1
	if kind == "sessions" {
		s.sort, s.descending = next.Column, next.Descending
		a.saveUsagePrefs()
	}
	if strings.HasPrefix(kind, "breakdown:") && kind != "breakdown:tools" {
		s.breakdownSort, s.breakdownDescending = next.Column, next.Descending
		for key, other := range s.tables {
			if strings.HasPrefix(key, "breakdown:") && key != "breakdown:tools" {
				other.Sort, other.observedSort = next, next
			}
		}
		a.saveUsagePrefs()
	}
}
func usageClipCell(text string, width, fontSize float32) string {
	budget := max(3, int(math.Floor(float64(max(24, width-32)/(fontSize*.52)))))
	runes := []rune(text)
	if len(runes) <= budget {
		return text
	}
	return string(runes[:budget-1]) + "…"
}
func (a *App) usageCellColor(kind string, r usageTableRow, column string) ui.Color {
	if kind == "series" {
		switch column {
		case "value":
			return a.Theme.Text
		case "time":
			return a.Theme.TextSecondary
		default:
			return a.Theme.TextMuted
		}
	}
	switch column {
	case "title", "name", "tool", "tokens", "date":
		return a.Theme.Text
	case "workspace", "provider", "model", "requests", "calls", "session":
		return a.Theme.TextSecondary
	case "errors":
		if r.Errors+r.ToolFailures > 0 {
			return a.Theme.Error
		}
	case "kind":
		if r.Failure != nil && r.Failure.Kind == "provider" {
			return a.Theme.Error
		}
		return a.Theme.Warning
	}
	return a.Theme.TextMuted
}

// Keep virtualization/keyboard in MyGo's List but own the source-specific
// chrome. Generic Table adds reordering, autofit, different padding and caret
// semantics, none of which the frozen Usage table exposes.
func (a *App) usageNativeTable(c *ui.Context, kind string, columns []ui.TableColumn, result usageTableResult) (ui.Element, ui.Element) {
	s, t := &a.usageState, a.usageState.table(kind)
	columns = append([]ui.TableColumn(nil), columns...)
	total := float32(0)
	for i := range columns {
		if w, ok := t.List.Columns.Widths[columns[i].ID]; ok && !columns[i].Fixed {
			columns[i].Width = max(48, w)
		}
		total += columns[i].Width
	}
	n := len(result.Rows)
	frameRows := n
	if n == 0 {
		frameRows = 3
	}
	frame := ui.Column(c.Key("usage-table:"+kind)).Height(float32(frameRows*40+42)).Border(1, a.Theme.Border).Radius(a.Theme.Px(12)).Clip().AlignSelf(ui.Stretch).Role(ui.RoleTable).Label("Usage " + kind + " table")
	var body ui.Element
	rowBounds := map[int]*ui.Handle{}
	if t.rowHandles == nil {
		t.rowHandles = map[int]*ui.Handle{}
	}
	base := c.Theme()
	theme := *base
	theme.Radius = 0
	// Our rows paint only source scope selection, not a generic List's cursor.
	theme.Accent = ui.RGBA(0, 0, 0, 0)
	theme.AccentText = a.Theme.Text
	c.SetTheme(&theme)
	frame.Children(func() {
		head := ui.Row(c).Height(40).Shrink(0).Clip().BorderWidth(0, 0, 1, 0).BorderColor(a.Theme.Border).Role(ui.RoleRow)
		head.Children(func() {
			for _, col := range columns {
				h := ui.Row(c.Key("header:"+col.ID)).Width(col.Width).Height(40).Shrink(0).Padding(0, 16).Gap(a.Theme.Space(6)).AlignItems(ui.Center).Role(ui.RoleColumnHeader).Label(col.Title)
				if col.Align == ui.End {
					h.Justify(ui.End)
				}
				if col.Sortable {
					h.Cursor(ui.CursorPointer)
					if h.Hovered() {
						h.Background(a.Theme.Hover)
					}
					if h.Clicked() {
						a.setUsageTableSort(kind, col.ID)
					}
				}
				h.Children(func() {
					// GPUI's header is a clipped container around a nowrap label.
					// Keep that boundary instead of shortening the text itself:
					// SingleLine reserves an ellipsis and loses extra characters.
					ui.Box(c).MinWidth(0).Clip().Label("Usage header label " + kind + " " + col.ID).Children(func() {
						ui.Text(c, col.Title).Label("Usage header text " + kind + " " + col.ID).NoWrap().Shrink(0).FontSize(a.Theme.Px(14)).FixedLineHeight(usageLineHeight(a.Theme.Px(14))).FontWeight(500).TextColor(a.Theme.Text)
					})
					if col.Sortable {
						caret := ui.Column(c).Shrink(0).Gap(0).AlignItems(ui.Center).Label("Sort carets " + kind + " " + col.ID)
						if t.Sort.Column != col.ID {
							caret.Opacity(0)
						}
						caret.Children(func() {
							a.icon(c, "icons/chevron-up.svg", a.Theme.Px(10), a.Theme.TextSecondary)
							a.icon(c, "icons/chevron-down.svg", a.Theme.Px(10), a.Theme.TextSecondary)
						})
					}
					if !col.Fixed {
						grip := ui.Box(c.Key("resize:" + col.ID)).Absolute().Top(0).Bottom(0).Right(-2.5).Width(5).Cursor(ui.CursorResizeEW).Label("Resize " + kind + " " + col.ID)
						grip.Clicked() // consume pointer presses without sorting the header
						if dx, _, held := grip.Dragged(); held && dx != 0 {
							if t.List.Columns.Widths == nil {
								t.List.Columns.Widths = map[string]float32{}
							}
							t.List.Columns.Widths[col.ID] = max(48, col.Width+dx)
						}
						if grip.Hovered() {
							grip.Children(func() { ui.Box(c).Absolute().Top(0).Bottom(0).Left(2).Width(1).Background(a.Theme.BorderStrong()) })
						}
					}
				})
			}
		})
		body = ui.List(c.Key("usage-rows:"+kind), &t.List, n, func(i int) {
			r := result.Rows[i]
			row := ui.Row(c.Key("row")).Height(40).MinWidth(total).AlignSelf(ui.Stretch).AlignItems(ui.Center).Role(ui.RoleRow)
			if i < n-1 {
				row.BorderWidth(0, 0, 1, 0).BorderColor(a.Theme.Border)
			}
			if row.Hovered() {
				row.Background(a.Theme.Hover)
			}
			if t.Selected == i && (kind == "sessions" || kind == "series") {
				row.Background(a.Theme.Selection)
			}
			handle := t.rowHandles[i]
			if handle == nil {
				handle = &ui.Handle{}
				t.rowHandles[i] = handle
			}
			row.Bind(handle)
			rowBounds[i] = handle
			row.Children(func() {
				for _, col := range columns {
					cell := ui.Row(c.Key(col.ID)).Width(col.Width).Height(40).Shrink(0).Padding(0, 16).AlignItems(ui.Center).Role(ui.RoleCell)
					if col.Align == ui.End {
						cell.Justify(ui.End)
					}
					cell.Children(func() {
						if col.ID == "open" {
							cell.Padding(0).Justify(ui.Center)
							button := ui.Box(c).Size(a.Theme.Px(22), a.Theme.Px(22)).Center().Radius(a.Theme.Px(6)).KeepFocus().Cursor(ui.CursorPointer).Label("Open " + r.Name).Tooltip(a.text("usage.ctx_open_session", "Open session")).Disabled(r.Path == "")
							if !row.Hovered() {
								button.Opacity(0)
							}
							if button.Hovered() {
								button.Background(a.Theme.OverlayStrong())
							}
							button.Children(func() { a.icon(c, "icons/arrow-up-right.svg", a.Theme.Px(14), a.Theme.TextSecondary) })
							if button.Clicked() {
								a.openUsageSession(r)
							}
							return
						}
						if (kind == "daily" && col.ID == "date") || (strings.HasPrefix(kind, "breakdown:") && (col.ID == "name" || col.ID == "tool")) {
							cell.Gap(a.Theme.Space(8))
							label, sub, size := r.Name, r.Sub, a.Theme.Px(12)
							if kind == "daily" {
								sub = a.text("usage.granularity_"+r.Granularity, r.Granularity)
							}
							if col.ID == "tool" {
								class := workbench.UsageToolClass(r.Name)
								sub = a.text("usage.class_"+class, titleCase(class))
								size = a.Theme.Px(14)
							}
							labelText := ui.Text(c, label).MinWidth(0).SingleLine().FontSize(size).TextColor(a.Theme.Text)
							if col.ID == "tool" {
								labelText.MaxWidth(max(0, col.Width-96))
							} else {
								labelText.MaxWidth(max(0, col.Width-24))
							}
							if sub != "" {
								ui.Text(c, sub).MinWidth(0).Grow(1).SingleLine().FontSize(a.Theme.Px(10)).TextColor(a.Theme.TextMuted)
							}
							return
						}
						raw := a.usageCell(r, col.ID, kind)
						if col.ID == "share" {
							raw = usagePercent(a.usageBreakdownFraction(r, result.Totals, false) * 100)
						}
						text := ui.Text(c, usageClipCell(raw, col.Width, a.Theme.Px(13))).MinWidth(0).SingleLine().FontSize(a.Theme.Px(14)).TextColor(a.usageCellColor(kind, r, col.ID))
						if col.Align == ui.End {
							text.Font(a.Fonts.CodeFamily)
						}
					})
				}
				if kind == "sessions" && s.menu("context").Open && s.contextRow == i {
					a.usageSessionContextPopup(c, row, r)
				}
			})
		}).Grow(1).RowWidth(total)
		head.ScrollWith(body)
	})
	c.SetTheme(base)
	if kind == "sessions" {
		frame.Bind(&t.contextHandle)
		a.usageSessionContext(c, frame, &t.contextHandle, result.Rows, rowBounds)
	}
	return frame, body
}
