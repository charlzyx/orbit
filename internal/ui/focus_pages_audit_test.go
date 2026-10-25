package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func tabToAudit(t *testing.T, tt *ui.Tester, label string, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if tt.Focused(label) {
			return
		}
		tt.Key(0, ui.KeyTab)
		tt.Frame()
	}
	t.Fatalf("keyboard could not reach %q in %d Tab stops", label, n)
}
func TestAuditSettingsKeyboardReachAndReturn(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Search settings"); e != nil {
		t.Fatal(e)
	}
	tabToAudit(t, tt, "Appearance", 30)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.settingsPage != "appearance" {
		t.Fatal("category activation failed", a.settingsPage)
	}
	tabToAudit(t, tt, "Appearance mode", 70)
	tt.Key(0, ui.KeyEnd)
	tt.Frame()
	if a.Appearance.Mode != "system" {
		t.Fatal("End did not choose source System mode", a.Appearance.Mode)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.page != "chat" || !tt.Focused("Message") {
		t.Fatal("settings escape did not return to composer", a.page)
	}
}
func TestAuditUsageFilterKeyboardReachDismissAndReopen(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("All models"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	if !tt.Focused("Filter models") {
		t.Fatal("filter search not focused")
	}
	tabToAudit(t, tt, "Select all", 30)
	tt.Key(0, ui.KeySpace)
	tt.Frame()
	if len(a.usageState.filter.ModelRefs) == 0 {
		t.Fatal("Space did not activate focused Select all")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.usageState.filterMenu != "" {
		t.Fatal("filter did not dismiss")
	}
	// The trigger label changes after selection; tab out and back must remain usable.
	tabToAudit(t, tt, "Back to task", 90)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.page != "chat" {
		t.Fatal("keyboard Back did not leave usage")
	}
}

func TestAuditSettingsDropdownKeyboardOwnsEscape(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "appearance"
	tt := ui.NewTester(a.View, 1470, 922)
	if e := tt.Click("Search settings"); e != nil {
		t.Fatal(e)
	}
	tabToAudit(t, tt, "Language selector", 70)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if !tt.Focused("Language filter") {
		t.Fatal("language picker search not focused")
	}
	tt.Type("no-language-can-match-this")
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.nativeSettings.selectMenu.id != "Language" {
		t.Fatal("empty picker selected phantom row")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.page != "settings" || a.nativeSettings.selectMenu.id != "" {
		t.Fatal("Escape left settings rather than picker")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.page != "chat" || !tt.Focused("Message") {
		t.Fatal("second Escape failed to return composer")
	}
}
