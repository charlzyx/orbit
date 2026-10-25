package orbitui

import (
	"github.com/egoist/mygo/ui"
	"slices"
	"strings"
	"time"
)

// Values stay wire/storage identifiers; labels are the pinned source's human names.
type settingsChoice struct{ Value, Label string }

func (a *App) sourceSettingsButton(c *ui.Context, label string) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Height(t.Px(28)).Padding(0, t.Space(8)).Gap(t.Space(6)).Radius(t.Px(8)).Border(1, t.Border).Background(t.Elevated).TextColor(t.TextSecondary).FontSize(t.Px(14)).Children(func() { ui.Text(c, label) })
	if b.Hovered() {
		b.Background(t.Hover)
	}
	return b
}

func choiceLabel(value string, choices []settingsChoice) string {
	for _, choice := range choices {
		if choice.Value == value {
			return choice.Label
		}
	}
	return value
}

type settingsSelectState struct {
	id, query   string
	cursor      int
	reveal      bool
	list        ui.ListState
	dismissedAt time.Time
}

func visibleSettingsChoices(choices []settingsChoice, query string) []int {
	out := []int{}
	query = strings.ToLower(strings.TrimSpace(query))
	for i, v := range choices {
		if query == "" || strings.Contains(strings.ToLower(v.Label), query) {
			out = append(out, i)
		}
	}
	return out
}
func (a *App) sourceSettingsSelect(c *ui.Context, id, value string, choices []settingsChoice, apply func(string)) ui.Element {
	t := a.Theme
	s := &a.nativeSettings.selectMenu
	open := s.id == id
	selected := -1
	for i, v := range choices {
		if v.Value == value {
			selected = i
		}
	}
	trigger := ui.ButtonBase(c.Key(id+":trigger")).Label(id+" selector").Height(t.Px(28)).Shrink(0).Padding(0, t.Space(8)).Gap(t.Space(4)).Radius(t.Px(8)).Border(1, t.Border).Background(t.Elevated).AlignItems(ui.Center)
	if open || trigger.Hovered() {
		trigger.Background(t.overlay())
	}
	trigger.Children(func() {
		ui.Text(c, choiceLabel(value, choices)).FontSize(t.Px(14)).LineHeight(1).TextColor(t.TextSecondary).SingleLine()
		a.icon(c, "icons/chevron-down.svg", 12, t.TextMuted)
	})
	if trigger.Clicked() {
		if !s.dismissedAt.IsZero() && time.Since(s.dismissedAt) < 200*time.Millisecond {
			s.dismissedAt = time.Time{}
		} else if open {
			s.id = ""
			open = false
		} else {
			*s = settingsSelectState{id: id, cursor: max(0, selected), reveal: true}
			open = true
		}
	}
	if !open {
		return trigger
	}
	panel := ui.PopoverBase(c.Key(id+":popover"), trigger, &open, func(panel ui.Element) {
		width, _ := c.Size()
		panel.Width(min(float32(360), max(float32(0), width-16))).AttachTo(trigger, ui.AnchorBottomRight, ui.AnchorTopRight).Margin(4, 0, 0, 0).Background(t.Elevated).Border(1, t.Border).Radius(t.Px(8)).Gap(0).Children(func() {
			ui.Row(c).Height(t.Px(36)).Padding(0, t.Px(10)).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
				a.icon(c, "icons/search.svg", 16, t.TextMuted)
				input := ui.TextInput(c.Key("search"), &s.query).Label(id+" filter").Border(0, ui.Color{}).Background(ui.Color{}).FocusRing(false).Padding(0).Placeholder(a.text("app.filter", "Filter…")).AutoFocus().Grow(1).OnTextKey(func(k ui.TextKey) bool {
					visible := visibleSettingsChoices(choices, s.query)
					switch k.Key {
					case ui.KeyUp, ui.KeyDown:
						if len(visible) > 0 {
							position := slices.Index(visible, s.cursor)
							direction := 1
							if k.Key == ui.KeyUp {
								direction = -1
							}
							if position < 0 {
								if direction < 0 {
									position = 0
								} else {
									position = -1
								}
							}
							s.cursor = visible[(position+direction+len(visible))%len(visible)]
							s.reveal = true
						}
						return true
					case ui.KeyEnter:
						index := s.cursor
						if !slices.Contains(visible, index) {
							if len(visible) == 0 {
								return true
							}
							index = visible[0]
						}
						apply(choices[index].Value)
						s.id = ""
						open = false
						return true
					case ui.KeyEscape:
						s.id = ""
						open = false
						return true
					}
					return false
				})
				if input.Changed() {
					s.reveal = true
				}
			})
			ui.Box(c).Height(1).Shrink(0).Background(t.Border)
			visible := visibleSettingsChoices(choices, s.query)
			if len(visible) == 0 {
				ui.Text(c, a.Tr("settings.no_matches")).Padding(t.Space(8)).TextColor(t.TextMuted)
				return
			}
			rowH := t.Px(14*sourceDefaultLineHeight + 8)
			listH := min(float32(220), float32(len(visible))*rowH+t.Space(4)*2)
			if s.reveal {
				position := slices.Index(visible, s.cursor)
				if position >= 0 {
					s.list.ScrollTo(position, ui.Center)
				}
				s.reveal = false
			}
			s.list.Key = func(i int) any { return visible[i] }
			s.list.Label = func(i int) string { return choices[visible[i]].Label }
			ui.List(c.Key("choices"), &s.list, len(visible), func(row int) {
				original := visible[row]
				choice := choices[original]
				b := ui.ButtonBase(c.Key(choice.Value)).Label(choice.Label).AlignSelf(ui.Stretch).Height(rowH).Shrink(0).Padding(t.Px(4), t.Space(6)).Gap(t.Space(8))
				if original == selected || original == s.cursor {
					b.Background(t.Active)
				} else if b.Hovered() {
					b.Background(t.Hover)
				}
				b.Children(func() {
					ui.Text(c, choice.Label).FontSize(t.Px(14)).SingleLine().Grow(1).MinWidth(0)
					if original == selected {
						a.icon(c, "icons/check.svg", 14, t.AccentText)
					}
				})
				if b.Clicked() {
					apply(choice.Value)
					s.id = ""
					open = false
				}
			}).Height(listH).AlignSelf(ui.Stretch)
		})
	})
	if panel.Dismissed() && s.id == id {
		s.id = ""
		if panel.PressedOutside() {
			s.dismissedAt = time.Now()
		} else {
			s.dismissedAt = time.Time{}
		}
	}
	return trigger
}

func (a *App) sourceSettingsSegments(c *ui.Context, id, value string, choices []settingsChoice, disabled bool, apply func(string)) {
	t := a.Theme
	row := ui.Row(c.Key(id)).Label(id).Focusable().Disabled(disabled).AlignItems(ui.Center).Radius(t.Px(8)).Border(1, t.Border).Clip()
	row.Children(func() {
		for _, choice := range choices {
			b := ui.ButtonBase(c.Key(choice.Value)).Label(choice.Label).Height(t.Px(26)).Padding(0, t.Space(8)).Radius(0).FontSize(t.Px(14)).TextColor(t.TextSecondary).Background(t.Elevated)
			if value == choice.Value {
				b.Background(t.Selection).TextColor(t.Text)
			} else if b.Hovered() {
				b.Background(t.Hover)
			}
			b.Children(func() { ui.Text(c, choice.Label) })
			if b.Clicked() {
				apply(choice.Value)
			}
		}
	})
	index := 0
	for i, v := range choices {
		if v.Value == value {
			index = i
		}
	}
	next := index
	switch {
	case row.Shortcut(0, ui.KeyLeft), row.Shortcut(0, ui.KeyUp):
		next = (index + len(choices) - 1) % len(choices)
	case row.Shortcut(0, ui.KeyRight), row.Shortcut(0, ui.KeyDown):
		next = (index + 1) % len(choices)
	case row.Shortcut(0, ui.KeyHome):
		next = 0
	case row.Shortcut(0, ui.KeyEnd):
		next = len(choices) - 1
	}
	if next != index {
		apply(choices[next].Value)
	}
}
func (a *App) sourceSettingsGroup(c *ui.Context, label string, body func()) {
	t := a.Theme
	ui.Column(c).Label(label).AlignSelf(ui.Stretch).Background(t.SurfaceSecondary).Border(1, t.Border).Radius(t.Px(12)).Children(body)
}
func sourceShortcutLabel(chord string) string {
	// Native help is an advertised chord, not a dump of every key binding.
	parts := strings.Split(shortcutLabel(chord), "+")
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = strings.ToUpper(p)
		}
		switch p {
		case "escape":
			parts[i] = "Esc"
		case "enter":
			parts[i] = "Enter"
		case "tab":
			parts[i] = "Tab"
		case "space":
			parts[i] = "Space"
		case "up":
			parts[i] = "↑"
		case "down":
			parts[i] = "↓"
		}
	}
	label := strings.Join(parts, "+")
	if strings.ContainsAny(label, "⌘⌃⇧⌥") {
		runes := []rune(label)
		if n := len(runes); n > 0 && runes[n-1] >= 'a' && runes[n-1] <= 'z' {
			runes[n-1] -= 'a' - 'A'
			label = string(runes)
		}
	}
	return label
}
