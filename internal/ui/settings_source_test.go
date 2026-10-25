package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestSettingsSourceSearchIndexAndSectionNavigation(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	if len(settingsSearchIndex) != 55 {
		t.Fatal("source index row count changed", len(settingsSearchIndex))
	}
	hits := a.settingsSearchHits("pid")
	if len(hits) != 1 || hits[0].Section != "runtime" || hits[0].Label != "settings.process_id" {
		t.Fatalf("source keyword index: %+v", hits)
	}
	tt := ui.NewTester(a.View, 1470, 922)
	if err := tt.Click("Search settings"); err != nil {
		t.Fatal(err)
	}
	tt.Type("pid")
	tt.Frame()
	if _, ok := tt.Find("General"); !ok {
		t.Fatal("search incorrectly filtered category navigation")
	}
	if err := tt.Click(a.Tr("settings.process_id")); err != nil {
		t.Fatal(err)
	}
	if a.settingsPage != "runtime" || a.settingsSearch != "" {
		t.Fatal("result did not open section and clear source query")
	}
}
func TestSettingsSourceHeaderAndGeneralRows(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	r, ok := tt.Find("Settings page title")
	if !ok {
		t.Fatal("header missing")
	}
	if r.X != 399 || r.Y != 20 {
		t.Fatalf("source fixed header position: %+v", r)
	}
	text := strings.Join(tt.Texts(), "\n")
	for _, key := range []string{"settings.general_description", "settings.sessions_live_in_pi_s_own_store_on_this_computer", "settings.new_tasks_start_in_this_directory", "settings.notifications_description"} {
		if !strings.Contains(text, a.Tr(key)) {
			t.Fatal("missing source label", key, text)
		}
	}
	if strings.Contains(text, "Signed updates are not configured") || strings.Contains(text, "UPDATES") {
		t.Fatal("unavailable updater introduced extra source card")
	}
	if _, ok := tt.Find("Browse…"); ok {
		t.Fatal("General workspace is metadata, not a source browse control")
	}
	before := a.notificationPrefs.Sound
	if err := tt.Click(a.Tr("settings.notification_sound") + " toggle"); err != nil {
		t.Fatal(err)
	}
	if a.Options.Store.Notifications().Sound == before {
		t.Fatal("notification toggle did not persist")
	}
	tt.Key(0, ui.KeySpace)
	if a.Options.Store.Notifications().Sound != before {
		t.Fatal("focused switch did not toggle back by keyboard")
	}
}
