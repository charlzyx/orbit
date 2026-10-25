package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestAppearanceSingleTabStop(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "appearance"
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Search settings"); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 40 && !tt.Focused("Appearance mode"); i++ {
		tt.Key(0, ui.KeyTab)
		tt.Frame()
	}
	if !tt.Focused("Appearance mode") {
		t.Fatal("group not keyboard reachable")
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Theme selector") {
		t.Fatal("Tab visits an individual appearance segment instead of next setting")
	}
	tt.Key(ui.Shift, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Appearance mode") {
		t.Fatal("ShiftTab visits an individual segment instead of group")
	}
	for _, step := range []struct {
		key  ui.Key
		want string
	}{{ui.KeyHome, "light"}, {ui.KeyLeft, "system"}, {ui.KeyRight, "light"}, {ui.KeyDown, "dark"}, {ui.KeyUp, "light"}, {ui.KeyEnd, "system"}} {
		tt.Key(0, step.key)
		tt.Frame()
		if a.Appearance.Mode != step.want || !tt.Focused("Appearance mode") {
			t.Fatalf("key %v value=%s expected=%s", step.key, a.Appearance.Mode, step.want)
		}
	}
	if e := tt.Click("Search settings"); e != nil {
		t.Fatal(e)
	}
	if e := tt.Click(a.Tr("settings.appearance_dark")); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if a.Appearance.Mode != "dark" || !tt.Focused("Search settings") {
		t.Fatal("mouse segment selection stole original input focus", a.Appearance.Mode)
	}
}

func TestAppearanceSegmentsDisabledStayInert(t *testing.T) {
	a := fixture(t)
	disabled := true
	value := "light"
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Children(func() {
			ui.Button(c, "Before")
			a.sourceSettingsSegments(c, "Fixture appearance", value, []settingsChoice{{"light", "Light"}, {"dark", "Dark"}, {"system", "System"}}, disabled, func(v string) { value = v; calls++ })
			ui.Button(c, "After")
		})
	}, 500, 300)
	if e := tt.Click("Before"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("After") {
		t.Fatal("disabled appearance retained Tab stop")
	}
	if e := tt.Click("Dark"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	if calls != 0 || value != "light" {
		t.Fatal("disabled segment changed", calls, value)
	}
	disabled = false
	tt.Frame()
	if e := tt.Click("Before"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("Fixture appearance") {
		t.Fatal("reenabled group unreachable")
	}
	tt.Key(0, ui.KeyRight)
	tt.Frame()
	if value != "dark" || calls != 1 {
		t.Fatal("reenabled group did not select exactly once", value, calls)
	}
	tt.Key(0, ui.KeyTab)
	tt.Frame()
	if !tt.Focused("After") {
		t.Fatal("reenabled child reentered Tab order")
	}
}
