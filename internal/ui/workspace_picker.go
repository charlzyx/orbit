package orbitui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

type workspacePickerEntry struct {
	Name, Path string
	Updated    time.Time
}
type workspacePickerState struct {
	entries          []workspacePickerEntry
	query, lastQuery string
	cursor           int
	scroll           ui.ScrollState
	dismissedAt      time.Time
	pointer          ui.Point
	pointerKnown     bool
}

func workspaceFieldWidth(main float32) float32 { return max(240, min(400, main-48)) }
func (a *App) recentWorkspaces() []workspacePickerEntry {
	rows := []workspacePickerEntry{}
	seen := map[string]bool{}
	add := func(path string, at time.Time) {
		if path != "" && !seen[path] && len(rows) < 8 {
			seen[path] = true
			rows = append(rows, workspacePickerEntry{filepath.Base(filepath.Clean(path)), path, at})
		}
	}
	add(a.Model.Workspace, time.Time{})
	sessions := append([]SessionRow(nil), a.Model.Sessions...)
	sort.SliceStable(sessions, func(i, j int) bool { return sessions[i].Updated.After(sessions[j].Updated) })
	for _, s := range sessions {
		add(s.Workspace, s.Updated)
	}
	return rows
}
func (s *workspacePickerState) filtered() []workspacePickerEntry {
	q := strings.ToLower(s.query)
	var rows []workspacePickerEntry
	for _, r := range s.entries {
		if strings.Contains(strings.ToLower(r.Name), q) || strings.Contains(strings.ToLower(r.Path), q) {
			rows = append(rows, r)
		}
	}
	return rows
}
func (a *App) toggleWorkspacePicker(now time.Time) {
	s := &a.workspacePicker
	if a.workspaceOpen {
		a.closeWorkspacePicker(false, now)
		return
	}
	if !s.dismissedAt.IsZero() && now.Sub(s.dismissedAt) < 200*time.Millisecond {
		return
	}
	if a.Models != nil {
		a.Models.DismissPicker(false)
	}
	a.modelOpen = false
	a.thinkingOpen = false
	a.autocomplete.Dismiss()
	s.entries = a.recentWorkspaces()
	s.query = ""
	s.lastQuery = ""
	s.cursor = 0
	s.scroll = ui.ScrollState{}
	s.pointerKnown = false
	a.workspaceOpen = true
}
func (a *App) closeWorkspacePicker(mouse bool, now time.Time) {
	a.workspaceOpen = false
	if mouse {
		a.workspacePicker.dismissedAt = now
	}
	a.composer.Focus()
}
func (a *App) chooseWorkspaceRow() {
	rows := a.workspacePicker.filtered()
	index := a.workspacePicker.cursor
	a.workspaceOpen = false
	if index >= 0 && index < len(rows) {
		if rows[index].Path != a.Model.Workspace {
			a.Dispatch(Action{Kind: "SetWorkspace", Value: rows[index].Path})
		}
		a.composer.Focus()
	} else {
		a.Dispatch(Action{Kind: "BrowseWorkspace"})
		a.composer.Focus()
	}
}
func (a *App) workspacePickerKey(k ui.TextKey) bool {
	if !a.workspaceOpen {
		return false
	}
	if q, ok := any(k).(interface{ Composing() bool }); ok && q.Composing() {
		return k.Key == ui.KeyEnter
	}
	if k.Mods != 0 {
		return false
	}
	s := &a.workspacePicker
	switch k.Key {
	case ui.KeyDown, ui.KeyUp:
		count := len(s.filtered()) + 1
		delta := 1
		if k.Key == ui.KeyUp {
			delta = -1
		}
		s.cursor = (s.cursor + delta + count) % count
		rowH := a.Theme.Px(23) + a.Theme.Px(8)
		top := float32(s.cursor) * rowH
		if top < s.scroll.Y {
			s.scroll.Y = top
		} else if top+rowH > s.scroll.Y+4*rowH {
			s.scroll.Y = top - 3*rowH
		}
		return true
	case ui.KeyEnter:
		a.chooseWorkspaceRow()
		return true
	}
	return false
}
func workspaceAge(at, now time.Time) string {
	seconds := max(int64(0), int64(now.Sub(at).Seconds()))
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
func (a *App) workspacePickerPopup(c *ui.Context, anchor ui.Element, width float32) {
	if !a.workspaceOpen {
		return
	}
	t := a.Theme
	s := &a.workspacePicker
	popup := ui.PopoverBase(c.Key("workspace-picker-popup"), anchor, &a.workspaceOpen, func(panel ui.Element) {
		panel.Column().Width(width).Margin(5, 0, 0, 0).Radius(t.Px(12)).Border(1, t.Border).Background(t.Elevated).Font(a.Fonts.UIFamily).FontSize(t.FontSize).LineHeight(sourceDefaultLineHeight).Clip().Label("Recent workspace picker").Shadow(0, 6, 12, 0, ui.RGBA(0, 0, 0, .12))
		pointerMoved := false
		if x, y, ok := panel.PointerPosition(); ok {
			p := ui.Point{X: x, Y: y}
			pointerMoved = s.pointerKnown && p != s.pointer
			s.pointer = p
			s.pointerKnown = true
		}
		ui.Row(c).Height(t.Px(36)).AlignSelf(ui.Stretch).AlignItems(ui.Center).PaddingX(t.Px(10)).Gap(t.Space(6)).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			a.icon(c, "icons/search.svg", t.Px(14), t.TextMuted)
			in := ui.TextInputBase(c.Key("workspace-filter"), &s.query).Label("Filter recent folders").Placeholder(a.text("workspace_picker.filter_folders", "Filter folders…")).Grow(1).AutoFocus().OnTextKey(a.workspacePickerKey)
			if in.Changed() || s.query != s.lastQuery {
				s.lastQuery = s.query
				s.cursor = 0
				s.scroll.Y = 0
			}
		})
		rows := s.filtered()
		s.cursor = max(0, min(s.cursor, len(rows)))
		ui.Row(c).AlignSelf(ui.Stretch).Height(t.Px(28)).Padding(t.Space(4), t.Px(10), t.Space(4)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.text("workspace_picker.recent_folders", "Recent folders")).FontSize(t.Px(12)).TextColor(t.TextMuted)
			ui.Spacer(c)
			if len(rows) > 0 {
				ui.Text(c, fmt.Sprint(len(rows))).FontSize(t.Px(12)).TextColor(t.TextMuted)
			}
		})
		rowHeight := t.Px(23) + t.Px(8)
		if len(rows) == 0 {
			ui.Row(c).AlignSelf(ui.Stretch).Padding(t.Space(8), t.Px(10)).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
				a.icon(c, "icons/search.svg", 14, t.TextMuted)
				ui.Column(c).Grow(1).Children(func() {
					ui.Text(c, a.text("workspace_picker.no_folders_match", "No folders match")).TextColor(t.TextMuted)
					ui.Text(c, a.text("workspace_picker.try_another_name_or_choose_a_folder_below", "Try another name, or choose a folder below.")).FontSize(t.Px(12)).TextColor(t.TextMuted)
				})
			})
		} else {
			ui.Scroll(c.Key("workspace-recents")).TrackScroll(&s.scroll).AlignSelf(ui.Stretch).Height(min(4*rowHeight, float32(len(rows))*rowHeight+2*t.Space(4))).PaddingY(t.Space(4)).Children(func() {
				for i, r := range rows {
					current := r.Path == a.Model.Workspace
					b := ui.ButtonBase(c.Key("workspace-row:" + r.Path)).AlignSelf(ui.Stretch).Height(rowHeight).MarginX(t.Space(4)).PaddingX(t.Space(6)).Gap(t.Space(6)).Radius(t.Px(6)).Justify(ui.Start).Label("Workspace: " + r.Path)
					if pointerMoved && b.Hovered() {
						s.cursor = i
					}
					if i == s.cursor || current {
						b.Background(t.Selection)
					} else if b.Hovered() {
						b.Background(t.Hover)
					}
					b.Children(func() {
						a.icon(c, "icons/folder.svg", t.Px(14), t.TextMuted)
						weight := 400
						if i == s.cursor || current {
							weight = 500
						}
						ui.Text(c, r.Name).MaxWidth(170).MaxLines(1).FontWeight(weight)
						ui.Text(c, r.Path).Grow(1).MinWidth(0).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextMuted)
						if current {
							a.icon(c, "icons/check.svg", t.Px(14), t.AccentText)
						} else if !r.Updated.IsZero() {
							ui.Text(c, workspaceAge(r.Updated, c.Now())).FontSize(t.Px(12)).TextColor(t.TextMuted)
						}
					})
					if b.Clicked() {
						s.cursor = i
						a.chooseWorkspaceRow()
					}
				}
			})
		}
		ui.Column(c).AlignSelf(ui.Stretch).PaddingY(t.Space(4)).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Children(func() {
			b := ui.ButtonBase(c.Key("workspace-browse")).AlignSelf(ui.Stretch).Height(rowHeight).MarginX(t.Space(4)).PaddingX(t.Space(6)).Gap(t.Space(6)).Radius(t.Px(6)).Justify(ui.Start).Label(a.text("workspace_picker.choose_folder", "Choose folder…"))
			if pointerMoved && b.Hovered() {
				s.cursor = len(rows)
			}
			if s.cursor == len(rows) {
				b.Background(t.Selection)
			} else if b.Hovered() {
				b.Background(t.Hover)
			}
			b.Children(func() {
				a.icon(c, "icons/arrow-up-right.svg", t.Px(14), t.TextSecondary)
				ui.Text(c, a.text("workspace_picker.choose_folder", "Choose folder…")).TextColor(t.TextSecondary)
			})
			if b.Clicked() {
				s.cursor = len(rows)
				a.chooseWorkspaceRow()
			}
		})
	})
	if popup.Dismissed() {
		a.closeWorkspacePicker(popup.PressedOutside(), c.Now())
	}
}
