package orbitui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

// Rust retention_ui.rs gates only the launch pass on enabled. The explicit
// Privacy action must not quietly opt the user into future automatic cleanup.
func TestManualRetentionWithAutomaticCleanupOff(t *testing.T) {
	x := maintenanceFixture(t)
	path := filepath.Join(x.Store.PiDir, "sessions", "fixture", "old.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	data := []byte(`{"type":"session","version":3,"id":"old","timestamp":"2020-01-01T00:00:00Z","cwd":"/fixture"}
{"type":"message","id":"m1","parentId":null,"timestamp":"2020-01-01T00:00:01Z","message":{"role":"user","content":[{"type":"text","text":"old synthetic task"}],"timestamp":1577836801000}}
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := x.Store.SaveRetention(preferences.Retention{Enabled: false, Mode: "archive", Days: 30}); err != nil {
		t.Fatal(err)
	}
	x.runRetention(false)
	if x.UI.retentionBusy || len(x.Store.SessionSet(true)) != 0 {
		t.Fatal("disabled launch cleanup ran")
	}
	x.runRetention(true)
	if !x.UI.retentionBusy {
		t.Fatal("explicit cleanup was incorrectly gated by automatic cleanup")
	}
	drainRegressionUpdates(t, x.UI, func() bool { return !x.UI.retentionBusy })
	if !x.archived[path] {
		t.Fatal("manual cleanup did not archive the idle session")
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(data) {
		t.Fatal("archive modified the Pi session", err)
	}
	if x.Store.Retention().Enabled {
		t.Fatal("manual cleanup enabled future automatic cleanup")
	}
	if !x.sessionGate.TryLock() {
		t.Fatal("cleanup leaked maintenance lock")
	}
	x.sessionGate.Unlock()
}

func TestPrivacyManualCleanupOffStillConfirmsPermanentDeletion(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "privacy"
	a.nativeSettings.privacyLoaded = true
	a.retention = preferences.Retention{Enabled: false, Mode: "delete", Days: 30}
	if err := a.Options.Store.SaveRetention(a.retention); err != nil {
		t.Fatal(err)
	}
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Run session cleanup"); err != nil {
		t.Fatal(err)
	}
	if a.confirmTitle != "Permanently delete old tasks?" || !a.confirmOpen {
		t.Fatal("manual delete did not request confirmation")
	}
	if len(actions) != 0 {
		t.Fatal("deletion dispatched before confirmation", actions)
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 0 || a.confirmOpen {
		t.Fatal("cancel did not preserve sessions")
	}
	if err := tt.Click("Run session cleanup"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Continue"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "RunRetention" {
		t.Fatal("confirmed manual action not dispatched once", actions)
	}
	a.retentionBusy = true
	tt.Frame()
	if err := tt.Click("Run session cleanup"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || a.confirmOpen {
		t.Fatal("busy cleanup admitted repeated action")
	}
	if a.Options.Store.Retention().Enabled {
		t.Fatal("manual confirmation changed automatic cleanup preference")
	}
}
