package resourceui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

type appearance struct {
	native *ui.Theme
	Theme
}

func (s *State) appearance(c *ui.Context) appearance {
	a := appearance{native: c.Theme()}
	if s.Options.Theme != nil {
		a.Theme = s.Options.Theme()
	}
	if a.Px == nil {
		a.Px = func(n float32) float32 { return n * a.native.FontSize / 14 }
	}
	if a.Space == nil {
		a.Space = func(n int) float32 { return a.Px(float32(n)) }
	}
	if a.Elevated == (ui.Color{}) {
		a.Elevated = a.native.Surface
	}
	if a.SurfaceSecondary == (ui.Color{}) {
		a.SurfaceSecondary = a.native.Surface
	}
	if a.Secondary == (ui.Color{}) {
		a.Secondary = a.native.Text.Alpha(.82)
	}
	if a.Action == (ui.Color{}) {
		a.Action = a.native.Text
	}
	if a.ActionText == (ui.Color{}) {
		a.ActionText = a.native.Background
	}
	if a.SelectionText == (ui.Color{}) {
		a.SelectionText = a.native.Text
	}
	return a
}

var iconOnce sync.Once
var resourceIcons map[string]*ui.SVG

func icon(c *ui.Context, name string, size float32, color ui.Color) {
	iconOnce.Do(func() {
		resourceIcons = map[string]*ui.SVG{}
		for _, name := range []string{"magic-wand", "extensions", "search", "file", "folder", "copy", "trash", "refresh", "plus", "x"} {
			if data, err := orbitassets.Asset("icons/" + name + ".svg"); err == nil {
				if svg, err := ui.ParseSVG(data); err == nil {
					resourceIcons[name] = svg
				}
			}
		}
	})
	if svg := resourceIcons[name]; svg != nil {
		ui.Icon(c, svg).Size(size, size).Shrink(0).TextColor(color)
	}
}
func (s *State) button(c *ui.Context, key, label, glyph string, primary bool) ui.Element {
	a := s.appearance(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(a.Px(28)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(4)).Radius(a.Px(8)).Background(a.Elevated).Border(1, a.native.Border).AlignItems(ui.Center)
	color := a.Secondary
	if primary {
		b.Background(a.Action).Border(1, a.Action)
		color = a.ActionText
	}
	b.Children(func() {
		if glyph != "" {
			icon(c, glyph, a.Px(14), color)
		}
		ui.Text(c, label).FontSize(a.Px(14)).FontWeight(500).TextColor(color).SingleLine()
	})
	return b
}
func (s *State) iconButton(c *ui.Context, key, label, glyph string) ui.Element {
	a := s.appearance(c)
	return ui.ButtonBase(c.Key(key)).Label(label).Tooltip(label).Size(a.Px(26), a.Px(26)).Padding(0).Shrink(0).Center().Radius(a.Px(6)).Background(a.Elevated).Border(1, a.native.Border).Children(func() { icon(c, glyph, a.Px(14), a.native.TextMuted) })
}
func (s *State) chip(c *ui.Context, key, text string, color ui.Color) {
	a := s.appearance(c)
	ui.Row(c.Key(key)).Padding(a.Space(2), a.Space(6)).Radius(a.Px(4)).Background(color.Alpha(.10)).Children(func() { ui.Text(c, text).FontSize(a.Px(10)).TextColor(color).SingleLine() })
}
func (s *State) notice(c *ui.Context, key, title, detail string, color ui.Color) {
	a := s.appearance(c)
	ui.Column(c.Key(key)).Padding(a.Space(12)).Gap(a.Space(4)).Radius(a.Px(8)).Border(1, color.Alpha(.25)).Background(color.Alpha(.06)).Children(func() {
		ui.Text(c, title).FontSize(a.Px(13)).FontWeight(500).TextColor(color)
		ui.Text(c, detail).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Selectable()
	})
}
func (s *State) separator(c *ui.Context) { ui.Box(c).Height(1).Shrink(0).Background(c.Theme().Border) }

// View is mounted outside the host's generic settings scroll. Skills owns a
// master/detail pair; Plugins owns a pinned toolbar above a separate list scroll.
func (s *State) View(c *ui.Context, page string) ui.Element {
	if !s.loaded && !s.Busy && !s.closed {
		s.Refresh()
	}
	if s.page != page {
		if s.page != "" {
			s.Escape()
		}
		s.page = page
	}
	c = c.Key("resources-" + page)
	var root ui.Element
	if page == "skills" {
		root = s.skillsView(c)
	} else {
		root = s.pluginsView(c)
	}
	if root.Shortcut(0, ui.KeyEscape) || (s.HasModal() && c.Shortcut(0, ui.KeyEscape)) {
		s.Escape()
	}
	return root
}

func (s *State) statusLine(c *ui.Context, count string) {
	a := s.appearance(c)
	ui.Row(c.Key("status-line")).Gap(a.Space(8)).AlignItems(ui.Center).MinHeight(a.Px(28)).Children(func() {
		if s.Busy {
			ui.Spinner(c).Size(a.Px(14), a.Px(14)).Shrink(0)
			ui.Text(c, s.activity).Grow(1).MinWidth(0).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine()
			if s.button(c, "cancel-operation", s.label("git_panel.cancel", "Cancel"), "", false).Clicked() {
				s.Cancel()
			}
		} else {
			ui.Text(c, count).Grow(1).MinWidth(0).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine()
		}
	})
}

func (s *State) displayPath(path string) string {
	home := s.service.Preferences.UserHome
	if rel, err := filepath.Rel(home, path); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		if rel == "." {
			return "~"
		}
		return "~" + string(filepath.Separator) + rel
	}
	return path
}
func formatSize(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(n)/(1024*1024))
}
func (s *State) skillAge(modified, now time.Time) string {
	if modified.IsZero() {
		return s.label("skills_ui.unknown", "Unknown")
	}
	return s.label("skills_ui.n_ago", "%{time} ago", "time", relativeAge(modified, now))
}
func relativeAge(modified, now time.Time) string {
	d := now.Sub(modified)
	if d < time.Minute {
		return "now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	if d < 30*24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	if d < 365*24*time.Hour {
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
	return fmt.Sprintf("%dy", int(d.Hours()/(24*365)))
}
func (s *State) scopeLabel(scope string) string {
	switch scope {
	case "project":
		return s.label("skills_ui.project", "Project")
	case "user":
		return s.label("skills_ui.user", "User")
	default:
		return s.label("skills_ui.global", "Global")
	}
}
