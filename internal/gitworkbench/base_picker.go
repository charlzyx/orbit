package gitworkbench

import (
	"sort"
	"strings"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// defaultBase preserves GitPanel::default_base, including its kind-agnostic
// preferred-name pass. Only the fallback skips the current/local non-branch ref.
func (s *State) defaultBase() string {
	for _, candidate := range []string{"main", "master"} {
		for _, ref := range s.Data.Refs {
			if ref.Name == candidate {
				return candidate
			}
		}
	}
	for _, ref := range s.Data.Refs {
		if ref.Kind == workspace.RefBranch && ref.Name != s.Data.Branch {
			return ref.Name
		}
	}
	return "main"
}

func (s *State) baseBranches() []string {
	// Rust str::to_lowercase includes contextual/fuller Unicode mappings,
	// unlike Go strings.ToLower (Greek final sigma, dotted capital I).
	lower := cases.Lower(language.Und)
	needle := lower.String(strings.TrimSpace(s.baseQuery))
	var names []string
	for _, ref := range s.Data.Refs {
		if ref.Kind == workspace.RefBranch && strings.Contains(lower.String(ref.Name), needle) {
			names = append(names, ref.Name)
		}
	}
	// Rust sort_by_key is stable: preserve original ordering of case-fold ties.
	sort.SliceStable(names, func(i, j int) bool { return lower.String(names[i]) < lower.String(names[j]) })
	return names
}

func (s *State) closeBasePicker() {
	s.basePickerOpen = false
	s.baseDismissedAt = time.Now()
}

var pickerIconsOnce sync.Once
var pickerIcons map[string]*ui.SVG

func pickerIcon(c *ui.Context, name string, size float32, color ui.Color) {
	pickerIconsOnce.Do(func() {
		pickerIcons = make(map[string]*ui.SVG)
		for _, name := range []string{"branch", "search", "check", "chevron-down"} {
			if data, err := orbitassets.Asset("icons/" + name + ".svg"); err == nil {
				if svg, err := ui.ParseSVG(data); err == nil {
					pickerIcons[name] = svg
				}
			}
		}
	})
	if svg := pickerIcons[name]; svg != nil {
		ui.Icon(c, svg).Size(size, size).Shrink(0).TextColor(color)
	}
}

// Source pr_base_popup is an anchored local-branch search, not a free-form
// editable ref or OS menu. Selecting is pointer-only in the source handler;
// PickerConfirm/Prev/Next are registered but have no GitPanel listener.
func (s *State) baseBranchSelector(c *ui.Context) {
	t := c.Theme()
	label := s.base
	if label == "" {
		label = "Choose base"
	}
	trigger := ui.ButtonBase(c.Key("chip")).Label("Base branch").Height(28).Padding(0, 8).Gap(4).Radius(6).Border(1, t.Border).Background(t.Surface)
	if trigger.Hovered() || s.basePickerOpen {
		trigger.Background(t.SurfaceHover)
	}
	trigger.Children(func() {
		pickerIcon(c, "branch", 14, t.TextMuted)
		ui.Text(c, label).FontSize(14).MaxWidth(180).SingleLine()
		pickerIcon(c, "chevron-down", 12, t.TextMuted)
	})
	if trigger.Clicked() {
		if dismissed := s.baseDismissedAt; !dismissed.IsZero() && time.Since(dismissed) < 200*time.Millisecond {
			s.baseDismissedAt = time.Time{}
		} else {
			s.baseDismissedAt = time.Time{}
			s.basePickerOpen = !s.basePickerOpen
			if s.basePickerOpen {
				s.baseQuery = ""
			}
		}
	}
	if !s.basePickerOpen {
		return
	}
	panel := ui.PopoverBase(c.Key("popup"), trigger, &s.basePickerOpen, func(panel ui.Element) {
		panel.Label("PR base branches").Width(320).AttachTo(trigger, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 2, 0).Background(t.Surface).Border(1, t.Border).Radius(6).Gap(0).Children(func() {
			ui.Row(c.Key("search-row")).Height(36).PaddingX(10).Gap(8).AlignItems(ui.Center).Children(func() {
				pickerIcon(c, "search", 16, t.TextMuted)
				ui.TextInput(c.Key("search"), &s.baseQuery).Label("Search base branches").Placeholder("Search branches…").AutoFocus().Grow(1).MinWidth(0).Border(0, ui.Color{}).FocusRing(false).Background(ui.Color{}).OnTextKey(func(k ui.TextKey) bool {
					// Never turn search Enter into a PR submission or pretend the
					// source has an arrow-selected branch cursor.
					return k.Key == ui.KeyEnter || k.Key == ui.KeyUp || k.Key == ui.KeyDown
				})
			})
			separator(c)
			ui.Scroll(c.Key("branches")).MaxHeight(320).AlignSelf(ui.Stretch).Children(func() {
				ui.Column(c).PaddingY(4).Gap(0).AlignSelf(ui.Stretch).Children(func() {
					names := s.baseBranches()
					if len(names) == 0 {
						ui.Text(c, "No matching branches").Padding(6, 8).TextColor(t.TextMuted)
						return
					}
					muted(c, "Local branches").Padding(4, 8)
					for i, name := range names {
						selected := name == s.base
						row := ui.ButtonBase(c.Key(i)).Label(name).AlignSelf(ui.Stretch).Height(14*1.618+8).Shrink(0).MarginX(4).Padding(4, 6).Radius(6).Gap(6)
						if selected {
							row.Background(t.Selection)
						} else if row.Hovered() {
							row.Background(t.SurfaceHover)
						}
						row.Children(func() {
							pickerIcon(c, "branch", 14, t.TextMuted)
							ui.Text(c, name).SingleLine().MinWidth(0).Grow(1)
							if selected {
								pickerIcon(c, "check", 14, t.Text)
							}
						})
						if row.Clicked() {
							s.base = name
							s.closeBasePicker()
						}
					}
				})
			})
		})
	})
	if panel.Dismissed() {
		s.closeBasePicker()
	}
}
