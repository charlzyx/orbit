package resourceui

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// fixedNavTester models the source's 240-DIP reserved settings navigation slot.
// It deliberately does not masquerade as the full production host.
func fixedNavTester(s *State, page string, w, h int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		th := ui.DarkTheme()
		th.Font = ".ZedSans"
		th.FontSize = 14
		th.Background = ui.Hex("#191919")
		th.Surface = ui.Hex("#171717")
		th.Border = ui.Hex("#2b2b2b")
		th.Text = ui.Hex("#d4d4d4")
		th.TextMuted = ui.Hex("#858585")
		th.Accent = ui.Hex("#df7758")
		th.Selection = ui.Hex("#37271f")
		c.SetTheme(th)
		s.Options.Theme = func() Theme { return Theme{Elevated: ui.Hex("#242424"), CodeFont: ".ZedMono"} }
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Background(th.Background).Children(func() {
			ui.Column(c).Width(240).Shrink(0).Background(th.Surface).BorderWidth(0, 1, 0, 0).BorderColor(th.Border).Children(func() {
				ui.Box(c).Height(44)
				ui.Text(c, "Settings navigation fixture").Padding(12, 16).FontSize(12).TextColor(th.TextMuted)
				ui.Text(c, "Native page under test: "+page).Padding(12, 16).FontSize(12).TextColor(th.TextMuted)
				ui.Box(c).Grow(1)
				ui.Text(c, "Explicit test fixture · not a live account").Padding(12, 16).FontSize(10).TextColor(th.TextMuted)
			})
			s.View(c, page)
		})
		s.Overlay(c)
	}, w, h)
}
func TestNativeKeyboardSkillsAndFilterIsolation(t *testing.T) {
	s := fixture(t)
	visible := s.visibleSkills()
	if visible[0].Scope != "project" {
		t.Fatal("visual group order")
	}
	tt := tester(s, "skills", 1230, 922)
	click(t, tt, "Skill: "+visible[0].Name)
	tt.Key(0, ui.KeyDown)
	if s.Selected != visible[1].Path {
		t.Fatal("arrow did not follow visual group order", s.Selected)
	}
	tt.Key(0, ui.KeyEnd)
	if s.Selected != visible[len(visible)-1].Path {
		t.Fatal("End")
	}
	tt.Key(0, ui.KeyHome)
	if s.Selected != visible[0].Path {
		t.Fatal("Home")
	}
	click(t, tt, "Search skills")
	tt.Type("ui")
	before := s.Selected
	tt.Key(0, ui.KeyDown)
	if s.Selected != before {
		t.Fatal("list swallowed text-input navigation")
	}
}

func TestKeyboardEndFocusesAndRevealsLastSkill(t *testing.T) {
	s := fixture(t)
	for i := 0; i < 40; i++ {
		name := fmt.Sprintf("project-%02d", i)
		put(t, filepath.Join(s.Workspace, ".pi", "skills", name, "SKILL.md"), "---\nname: "+name+"\ndescription: Long-list fixture.\n---\n# Fixture\n")
	}
	s.Refresh()
	waitState(t, s)
	visible := s.visibleSkills()
	tt := fixedNavTester(s, "skills", 800, 560)
	click(t, tt, "Skill: "+visible[0].Name)
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	tt.Frame()
	label := "Skill: " + visible[len(visible)-1].Name
	r, ok := tt.Find(label)
	if !ok || r.Y < 0 || r.Y+r.H > 530 || !tt.Focused(label) {
		t.Fatal("End did not focus/reveal final row", r, tt.Focused(label))
	}
}

func TestNativePluginEnterDoesNotCommitButTabButtonDoes(t *testing.T) {
	s := fixture(t)
	s.SetPiBinary(fakePi(t, `echo explicit-install > "$HOME/install-marker"; echo installed`))
	tt := fixedNavTester(s, "plugins", 800, 560)
	click(t, tt, "Package source")
	tt.Type("npm:keyboard-fixture")
	tt.Key(0, ui.KeyEnter)
	if s.Busy {
		t.Fatal("source field Enter unexpectedly installed")
	}
	if _, err := os.Stat(filepath.Join(s.service.Preferences.Home, "install-marker")); !os.IsNotExist(err) {
		t.Fatal("source Enter ran process")
	}
	for _, label := range []string{"Global", "Project", "Install"} {
		tt.Key(0, ui.KeyTab)
		if !tt.Focused(label) {
			t.Fatal("Tab focus did not reach", label)
		}
	}
	tt.Key(0, ui.KeyEnter)
	waitState(t, s)
	if _, err := os.Stat(filepath.Join(s.service.Preferences.Home, "install-marker")); err != nil {
		t.Fatal("explicit keyboard button did not install", err)
	}
}
func TestMinimumWindowSourceGeometryAndControlReachability(t *testing.T) {
	for _, page := range []string{"skills", "plugins"} {
		t.Run(page, func(t *testing.T) {
			s := fixture(t)
			s.Options.OpenFile = func(string) {}
			s.Options.RevealPath = func(string) {}
			s.Options.CopyText = func(string) {}
			for _, p := range s.Skills {
				if p.Name == "native-review" {
					s.SelectSkill(p.Path)
				}
			}
			tt := fixedNavTester(s, page, 800, 560)
			labels := []string{"Search installed plugins", "Package source", "Global", "Project", "Install", "Refresh"}
			if page == "skills" {
				labels = []string{"Search skills", "Enable skill", "Open SKILL.md", "Show in File Manager", "Copy path", "Delete"}
			}
			for _, label := range labels {
				r, ok := tt.Find(label)
				if !ok || r.X < 240 || r.X+r.W > 800.5 || r.Y < 0 || r.Y+r.H > 560.5 {
					t.Fatalf("control %q not reachable at minimum window: %+v (%v)", label, r, ok)
				}
			}
			if page == "skills" {
				row, _ := tt.Find("Skill: native-review")
				if row.X < 240 || row.X+row.W > 560 {
					t.Fatal("master escaped fixed 320-DIP slot", row)
				}
				click(t, tt, "Delete")
				if !s.HasModal() {
					t.Fatal("confirm absent")
				}
				r, ok := tt.Find("Cancel")
				if !ok || r.X+r.W > 800.5 || r.Y+r.H > 560.5 {
					t.Fatal("confirmation unreachable", r)
				}
				tt.Key(0, ui.KeyEscape)
				if s.HasModal() {
					t.Fatal("minimum Escape")
				}
			}
		})
	}
}
func TestNativeUpdateNoticeRequiresActionAndReservesGeometry(t *testing.T) {
	s := fixture(t)
	s.Updates = []workbench.PluginUpdate{{Source: "npm:fixture-plugin", Name: "fixture-plugin", Scope: "global", Current: "1.2.3", Latest: "2.0.0"}}
	tt := fixedNavTester(s, "plugins", 800, 560)
	if s.Busy || !tt.HasText("Update available") {
		t.Fatal("notice started operation or absent")
	}
	for _, label := range []string{"Close update notice", "Skip", "Update now"} {
		r, ok := tt.Find(label)
		if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 800 || r.Y+r.H > 560 {
			t.Fatal("notice button offscreen", label, r)
		}
	}
	click(t, tt, "Close update notice")
	waitState(t, s)
	tt.Frame()
	if len(s.Updates) != 0 || s.Skipped["npm:fixture-plugin"] != "2.0.0" || tt.HasText("Update available") {
		t.Fatal("close did not persist exact skip")
	}
}
func TestNativeDeletedSelectionDoesNotReuseStaleTarget(t *testing.T) {
	s := fixture(t)
	first, second := s.Skills[0], s.Skills[1]
	s.SelectSkill(first.Path)
	tt := tester(s, "skills", 1230, 922)
	click(t, tt, "Delete")
	click(t, tt, "Skill: "+second.Name)
	if s.HasModal() {
		t.Fatal("old confirmation survived selection change")
	}
	s.ConfirmDeleteSkill()
	if s.Busy {
		t.Fatal("stale confirm operated")
	}
	if _, err := os.Stat(first.Path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(second.Path); err != nil {
		t.Fatal(err)
	}
}
func TestLongErrorAndBusyRemainReachable(t *testing.T) {
	s := fixture(t)
	s.Error = strings.Repeat("Fixture error with a long path /not-a-real-user/file. ", 20)
	tt := fixedNavTester(s, "plugins", 800, 560)
	if !tt.HasText("Plugin action failed") {
		t.Fatal("error absent")
	}
	for _, label := range []string{"Package source", "Refresh"} {
		r, _ := tt.Find(label)
		if r.Y < 0 || r.Y+r.H > 560 {
			t.Fatal("error displaced pinned toolbar")
		}
	}
}
func TestNativeCaptureFullResourceSlots(t *testing.T) {
	out := os.Getenv("ORBIT_RESOURCE_CAPTURE_DIR")
	if out == "" {
		t.Skip("explicit screenshot fixture opt-in")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{1470, 922}, {800, 560}} {
		for _, page := range []string{"skills", "plugins"} {
			s := fixture(t)
			for _, p := range s.Skills {
				if p.Name == "native-review" {
					s.SelectSkill(p.Path)
				}
			}
			s.Options.OpenFile = func(string) {}
			s.Options.RevealPath = func(string) {}
			s.Options.CopyText = func(string) {}
			tt := fixedNavTester(s, page, size[0], size[1])
			tt.Frame()
			path := filepath.Join(out, fmt.Sprintf("%s-source-slot-%dx%d.png", page, size[0], size[1]))
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err = png.Encode(f, tt.Image()); err != nil {
				t.Fatal(err)
			}
			f.Close()
		}
	}
}
