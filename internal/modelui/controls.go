package modelui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

type choice struct {
	label string
	value int
}

func (s *State) modelChoices(sentinel string) []choice {
	var out []choice
	if sentinel != "" {
		out = append(out, choice{sentinel, -1})
	}
	for i, m := range s.Snapshot.Models {
		out = append(out, choice{ProviderName(m.Provider) + " · " + modelLabel(m.Name, m.ID), i})
	}
	return out
}
func (s *State) ActiveModelControl(c *ui.Context) ui.Element {
	options := s.modelChoices("")
	selected := 0
	for i, m := range s.Snapshot.Models {
		if s.IsActive(m) {
			selected = i
			break
		}
	}
	return s.control(c, "active-model", s.label("settings.active_model", "Active model"), options, selected, 28, func(i int) bool { return s.SelectModel(i, false) })
}
func (s *State) ProviderControl(c *ui.Context) ui.Element {
	providers := s.Providers()
	options := []choice{{s.label("settings.all_providers", "All providers"), -1}}
	selected := 0
	for i, p := range providers {
		options = append(options, choice{ProviderName(p), i})
		if s.ProviderFilter == p {
			selected = i + 1
		}
	}
	return s.control(c, "provider-filter", s.label("settings.all_providers", "All providers"), options, selected, 32, func(i int) bool {
		if i < 0 {
			s.ProviderFilter = ""
		} else if i < len(providers) {
			s.ProviderFilter = providers[i]
		}
		return true
	})
}
func (s *State) DefaultModelControl(c *ui.Context) ui.Element {
	return s.control(c, "default-model", s.label("settings.default_model", "Default model"), s.modelChoices(s.label("settings.pi_default", "pi default")), s.DefaultModelIndex()+1, 28, s.ChooseDefaultModel)
}
func (s *State) DefaultThinkingControl(c *ui.Context) ui.Element {
	label := s.label("settings.pi_default", "pi default")
	levels := s.DefaultThinkingLevels()
	if levels == nil {
		return s.button(c, "default-thinking-disabled", label, "", 28, false).Label("Default thinking").Disabled(true)
	}
	options := []choice{{label, -1}}
	selected := 0
	for i, level := range levels {
		options = append(options, choice{displayLevel(level), i})
		if s.Defaults.Thinking == level {
			selected = i + 1
		}
	}
	return s.control(c, "default-thinking", s.label("settings.default_thinking", "Default thinking"), options, selected, 28, func(i int) bool {
		if i < 0 {
			return s.ChooseDefaultThinking("")
		}
		if i >= len(levels) {
			return false
		}
		return s.ChooseDefaultThinking(levels[i])
	})
}
func (s *State) TitleModelControl(c *ui.Context) ui.Element {
	selected := 0
	if s.Title.Model != nil {
		for i, m := range s.Snapshot.Models {
			if key(m) == *s.Title.Model {
				selected = i + 1
				break
			}
		}
	}
	return s.control(c, "title-model", s.label("settings.title_model", "Title model"), s.modelChoices(s.label("settings.active_session_model", "Active session model")), selected, 28, s.ChooseTitleModel)
}
func visibleChoices(options []choice, query string) []int {
	var out []int
	q := strings.ToLower(query)
	for i, o := range options {
		if q == "" || strings.Contains(strings.ToLower(o.label), q) {
			out = append(out, i)
		}
	}
	return out
}
func stepVisible(visible []int, current, direction int) int {
	if len(visible) == 0 {
		return -1
	}
	position := slices.Index(visible, current)
	if position < 0 {
		if direction < 0 {
			return visible[len(visible)-1]
		}
		return visible[0]
	}
	if direction < 0 {
		position = max(0, position-1)
	} else {
		position = min(len(visible)-1, position+1)
	}
	return visible[position]
}
func (s *State) control(c *ui.Context, kind, label string, options []choice, selected int, height float32, apply func(int) bool) ui.Element {
	c = c.Key("select:" + kind)
	a := s.appearance(c)
	if len(options) == 0 {
		return s.button(c, kind+":disabled", s.label("settings.no_models_reported", "No models reported"), "", height, false).Label(label).Disabled(true)
	}
	selected = max(0, min(selected, len(options)-1))
	open := s.selectMenu.kind == kind
	trigger := ui.ButtonBase(c.Key(kind+":trigger")).Label(label).Height(a.Px(height)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(4)).Radius(a.Px(8)).Border(1, a.native.Border).Background(a.Elevated).AlignItems(ui.Center)
	if open {
		trigger.Background(a.Overlay).Border(1, a.BorderStrong)
	} else if trigger.Hovered() {
		trigger.Background(a.Overlay)
	}
	trigger.Children(func() {
		ui.Text(c, options[selected].label).FontSize(a.Px(14)).LineHeight(1).TextColor(a.Secondary).SingleLine()
		icon(c, "icons/chevron-down.svg", a.Px(12), a.native.TextMuted)
	})
	if trigger.Clicked() {
		if !s.selectMenu.dismissedAt.IsZero() && time.Since(s.selectMenu.dismissedAt) < 200*time.Millisecond {
			s.selectMenu.dismissedAt = time.Time{}
		} else if open {
			s.selectMenu.kind = ""
			open = false
		} else {
			s.DismissPicker(false)
			s.selectMenu = selectState{kind: kind, cursor: selected, reveal: true}
			open = true
		}
	}
	if !open {
		return trigger
	}
	panel := ui.PopoverBase(c.Key(kind+":popover"), trigger, &open, func(panel ui.Element) {
		w, _ := c.Size()
		panel.Width(min(float32(360), max(float32(0), w-16))).MaxWidth(max(float32(0), w-16)).AttachTo(trigger, ui.AnchorBottomRight, ui.AnchorTopRight).Margin(4, 0, 0, 0).Background(a.Elevated).Border(1, a.BorderStrong).Radius(a.Px(8)).Gap(0).Children(func() {
			visible := visibleChoices(options, s.selectMenu.query)
			ui.Row(c.Key("search-row")).Height(a.Px(36)).Padding(0, a.Px(10)).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
				icon(c, "icons/search.svg", a.Px(16), a.native.TextMuted)
				in := ui.TextInput(c.Key("search"), &s.selectMenu.query).Border(0, ui.Color{}).Background(ui.Color{}).FocusRing(false).Padding(0).Label(label + " filter").Placeholder(s.label("app.filter", "Filter…")).AutoFocus().Grow(1).OnTextKey(func(k ui.TextKey) bool {
					visible = visibleChoices(options, s.selectMenu.query)
					switch k.Key {
					case ui.KeyUp, ui.KeyDown:
						direction := 1
						if k.Key == ui.KeyUp {
							direction = -1
						}
						s.selectMenu.cursor = stepVisible(visible, s.selectMenu.cursor, direction)
						s.selectMenu.reveal = true
						return true
					case ui.KeyEnter:
						index := s.selectMenu.cursor
						if !slices.Contains(visible, index) {
							if len(visible) == 0 {
								return true
							}
							index = visible[0]
						}
						if index >= 0 && index < len(options) && apply(options[index].value) {
							s.selectMenu.kind = ""
							open = false
						}
						return true
					case ui.KeyEscape:
						s.selectMenu.kind = ""
						open = false
						return true
					}
					return false
				})
				if in.Changed() {
					s.selectMenu.reveal = true
				}
			})
			ui.Box(c).Height(1).Shrink(0).Background(a.native.Border)
			visible = visibleChoices(options, s.selectMenu.query)
			if len(visible) == 0 {
				ui.Text(c, s.label("settings.no_matches", "No matches")).Padding(a.Space(8)).FontSize(a.Px(14)).TextColor(a.native.TextMuted)
				return
			}
			rowH := a.Px(14*1.618 + 8)
			listH := min(float32(220), float32(len(visible))*rowH+a.Space(4)*2)
			if s.selectMenu.reveal {
				position := slices.Index(visible, s.selectMenu.cursor)
				if position >= 0 {
					s.selectMenu.list.ScrollTo(position, ui.Center)
				}
				s.selectMenu.reveal = false
			}
			s.selectMenu.list.Key = func(i int) any { return visible[i] }
			s.selectMenu.list.Label = func(i int) string { return options[visible[i]].label }
			ui.List(c.Key("choices"), &s.selectMenu.list, len(visible), func(row int) {
				original := visible[row]
				o := options[original]
				b := ui.ButtonBase(c.Key(fmt.Sprint(original))).AlignSelf(ui.Stretch).Height(rowH).Shrink(0).Padding(a.Px(4), a.Space(6)).Gap(a.Space(8)).Label(o.label)
				if original == selected || original == s.selectMenu.cursor {
					b.Background(a.Active)
				} else if b.Hovered() {
					b.Background(a.Hover)
				}
				b.Children(func() {
					ui.Text(c, o.label).FontSize(a.Px(14)).SingleLine().Grow(1).MinWidth(0)
					if original == selected {
						icon(c, "icons/check.svg", a.Px(14), a.AccentText)
					}
				})
				if b.Clicked() && apply(o.value) {
					s.selectMenu.kind = ""
					open = false
				}
			}).Height(listH).AlignSelf(ui.Stretch)
		})
	})
	if panel.Dismissed() {
		s.selectMenu.kind = ""
		s.selectMenu.dismissedAt = time.Now()
	}
	return trigger
}
