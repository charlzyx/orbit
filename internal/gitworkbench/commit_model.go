package gitworkbench

import (
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/modelui"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

type commitModelPicker struct {
	scroll             ui.ScrollState
	reveal             bool
	selected           *preferences.Model
	loaded             bool
	open, focusMessage bool
	query              string
	cursor             int
	dismissed          time.Time
}

// ResolveCommitModel never substitutes another model for an explicit identity
// missing from the catalog. The catalog supplies capabilities, not permission.
func ResolveCommitModel(selected *preferences.Model, current agent.ModelChoice, catalog []agent.ModelChoice) agent.ModelChoice {
	if selected == nil {
		if current.ID == "" {
			return agent.ModelChoice{}
		}
		return current
	}
	for _, m := range catalog {
		if m.Provider == selected.Provider && m.ID == selected.ID {
			return m
		}
	}
	return agent.ModelChoice{Provider: selected.Provider, ID: selected.ID}
}
func (s *State) loadCommitModel() {
	if s.commitModel.loaded || s.Options.CommitModelStore == nil {
		return
	}
	s.commitModel.loaded = true
	model, err := s.Options.CommitModelStore.CommitModel()
	s.commitModel.selected = model
	if err != nil {
		s.Error = fmt.Errorf("Commit model preference could not be read; following active session: %w", err)
	}
}
func (s *State) commitCatalog() ([]agent.ModelChoice, agent.ModelChoice) {
	if s.Options.CommitModelCatalog == nil {
		return nil, agent.ModelChoice{}
	}
	return s.Options.CommitModelCatalog()
}
func (s *State) resolvedCommitModel() agent.ModelChoice {
	s.loadCommitModel()
	catalog, current := s.commitCatalog()
	return ResolveCommitModel(s.commitModel.selected, current, catalog)
}
func (s *State) selectCommitModel(model *preferences.Model) {
	if s.Busy {
		return
	}
	if model != nil {
		copy := *model
		model = &copy
	}
	s.commitModel.selected = model
	if s.Options.CommitModelStore != nil {
		if err := s.Options.CommitModelStore.SaveCommitModel(model); err != nil {
			s.Error = fmt.Errorf("Commit model applies for this run only; preference was not saved: %w", err)
		}
	}
	s.dismissCommitModel(false)
}
func (s *State) dismissCommitModel(mouse bool) {
	s.commitModel.open = false
	s.commitModel.focusMessage = true
	if mouse {
		s.commitModel.dismissed = time.Now()
	}
}
func (s *State) commitModelRows() []agent.ModelChoice {
	models, _ := s.commitCatalog()
	q := strings.ToLower(strings.TrimSpace(s.commitModel.query))
	rows := []agent.ModelChoice{{Name: "Follow active session"}}
	for _, m := range models {
		if m.ID != "" && m.Provider != "" && strings.Contains(strings.ToLower(m.Name+" "+m.ID+" "+m.Provider+" "+modelui.ProviderName(m.Provider)), q) {
			rows = append(rows, m)
		}
	}
	return rows
}
func (s *State) confirmCommitModel() {
	rows := s.commitModelRows()
	i := min(max(s.commitModel.cursor, 0), len(rows)-1)
	if i == 0 {
		s.selectCommitModel(nil)
	} else {
		s.selectCommitModel(&preferences.Model{Provider: rows[i].Provider, ID: rows[i].ID})
	}
}

// Reuse the workbench's native anchored picker primitives (baseBranchSelector),
// not the chat model action sink. Follow remains available even for no matches.
func (s *State) commitModelSelector(c *ui.Context) {
	if s.Options.CommitModelStore == nil {
		return
	}
	s.loadCommitModel()
	t := c.Theme()
	label := "Follow active session"
	if m := s.commitModel.selected; m != nil {
		label = m.Provider + " / " + m.ID
	}
	trigger := ui.ButtonBase(c.Key("commit-model-chip")).Label("Commit model").Disabled(s.Busy).Height(28).Padding(0, 8).Gap(4).Radius(6).Border(1, t.Border).Background(t.Surface)
	trigger.Children(func() { ui.Text(c, label).MaxWidth(190).SingleLine(); pickerIcon(c, "chevron-down", 12, t.TextMuted) })
	if trigger.Clicked() && !s.Busy {
		if !s.commitModel.dismissed.IsZero() && time.Since(s.commitModel.dismissed) < 200*time.Millisecond {
			s.commitModel.dismissed = time.Time{}
			s.commitModel.focusMessage = true
		} else if s.commitModel.open {
			s.dismissCommitModel(false)
		} else {
			s.commitModel.open = !s.commitModel.open
			s.commitModel.query = ""
			s.commitModel.cursor = 0
			for i, m := range s.commitModelRows() {
				if pick := s.commitModel.selected; pick != nil && pick.Provider == m.Provider && pick.ID == m.ID {
					s.commitModel.cursor = i
					break
				}
			}
			s.commitModel.reveal = true
		}
	}
	if s.Busy {
		s.commitModel.open = false
	}
	if !s.commitModel.open {
		return
	}
	panel := ui.PopoverBase(c.Key("commit-model-popup"), trigger, &s.commitModel.open, func(panel ui.Element) {
		panel.Label("Commit model picker").Width(360).AttachTo(trigger, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 2, 0).Background(t.Surface).Border(1, t.Border).Radius(6).Column().Children(func() {
			search := ui.TextInput(c.Key("commit-model-search"), &s.commitModel.query).Label("Search commit models").Placeholder("Search models…").AutoFocus().AlignSelf(ui.Stretch).OnTextKey(func(k ui.TextKey) bool {
				switch k.Key {
				case ui.KeyUp:
					s.commitModel.cursor = max(0, s.commitModel.cursor-1)
					s.commitModel.reveal = true
					return true
				case ui.KeyDown:
					s.commitModel.cursor = min(len(s.commitModelRows())-1, s.commitModel.cursor+1)
					s.commitModel.reveal = true
					return true
				case ui.KeyEnter:
					s.confirmCommitModel()
					return true
				case ui.KeyEscape:
					s.dismissCommitModel(false)
					return true
				}
				return false
			})
			if search.Changed() {
				s.commitModel.cursor = 0
				s.commitModel.reveal = true
			}
			ui.Scroll(c.Key("commit-model-rows")).KeepFocus().TrackScroll(&s.commitModel.scroll).MaxHeight(300).AlignSelf(ui.Stretch).Children(func() {
				ui.Column(c).AlignSelf(ui.Stretch).Children(func() {
					for i, m := range s.commitModelRows() {
						label := "Follow active session"
						if i > 0 {
							name := m.Name
							if name == "" {
								name = m.ID
							}
							label = name + " · " + m.Provider
						}
						row := ui.ButtonBase(c.Key(preferences.Model{Provider: m.Provider, ID: m.ID})).Label(label).Height(32).AlignSelf(ui.Stretch).Padding(4, 8)
						if i == s.commitModel.cursor {
							row.Background(t.Selection)
						} else if row.Hovered() {
							row.Background(t.SurfaceHover)
						}
						if s.commitModel.reveal && i == s.commitModel.cursor {
							row.ScrollIntoView()
						}
						row.Children(func() {
							ui.Text(c, label).SingleLine().Grow(1)
							if pick := s.commitModel.selected; (i == 0 && pick == nil) || (pick != nil && pick.Provider == m.Provider && pick.ID == m.ID) {
								pickerIcon(c, "check", 14, t.Text)
							}
						})
						if row.Clicked() {
							if i == 0 {
								s.selectCommitModel(nil)
							} else {
								s.selectCommitModel(&preferences.Model{Provider: m.Provider, ID: m.ID})
							}
						}
					}
				})
			})
		})
	})
	s.commitModel.reveal = false
	if panel.Dismissed() {
		s.dismissCommitModel(panel.PressedOutside())
	}
}
