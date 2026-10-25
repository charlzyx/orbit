package orbitui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestNativeProviderEditorValidationSecretAndMinimumSize(t *testing.T) {
	a := fixture(t)
	a.Options.PiBinary = "/nonexistent-fixture-pi"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: workbench.BuiltinProviders()}
	a.page = "settings"
	a.settingsPage = "providers"
	a.editProvider(workbench.Provider{})
	tt := ui.NewTester(a.View, 800, 560)
	tt.Frame()
	r, ok := tt.Find("Save provider")
	if !ok || r.Y+r.H > 560.5 {
		t.Fatalf("provider save unreachable %+v", r)
	}
	if !a.modalOpen() {
		t.Fatal("provider editor failed modal interception")
	}
	if e := tt.Click("Save provider"); e != nil {
		t.Fatal(e)
	}
	if a.providerState.editor == nil || a.providerState.editor.error == "" {
		t.Fatal("invalid provider discarded instead of inline validation")
	}
	d := a.providerState.editor
	d.edit.ID = "fixture-provider"
	d.edit.BaseURL = "https://example.invalid/v1"
	d.models = "fixture-model"
	d.key = "fixture-secret-never-painted"
	tt.Frame()
	if strings.Contains(strings.Join(tt.Texts(), " "), d.key) {
		t.Fatal("key painted as plain text")
	}
	if e := tt.Click("Save provider"); e != nil {
		t.Fatal(e)
	}
	drainRegressionUpdates(t, a, func() bool { return a.providerState.editor == nil && !a.settingsLoading })
	rows, e := a.Service().Providers()
	if e != nil || len(rows) != 1 || rows[0].ID != "fixture-provider" || !rows[0].HasAPIKey {
		t.Fatalf("native save didn't persist metadata: %+v %v", rows, e)
	}
}
func TestNativeProviderSaveFailureKeepsDraft(t *testing.T) {
	a := fixture(t)
	a.Options.Store.PiDir = filepath.Join(t.TempDir(), "blocked")
	if e := os.WriteFile(a.Options.Store.PiDir, []byte("not a directory"), 0600); e != nil {
		t.Fatal(e)
	}
	a.editProvider(workbench.Provider{})
	d := a.providerState.editor
	d.edit.ID = "fixture"
	d.edit.BaseURL = "https://example.invalid"
	d.models = "m"
	d.key = "fixture-only"
	a.saveProviderDraft(d)
	drainRegressionUpdates(t, a, func() bool { return !d.busy })
	if a.providerState.editor != d || d.error == "" || d.key != "fixture-only" {
		t.Fatal("failed save lost editable draft")
	}
}
func TestProviderCredentialCapabilities(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "providers"
	a.providerState.query = "openai-codex"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: workbench.BuiltinProviders()}
	tt := ui.NewTester(a.View, 1200, 800)
	tt.Frame()
	if e := tt.Click("Configure"); e != nil {
		t.Fatal(e)
	}
	for _, item := range tt.Menu() {
		if item == "API key…" {
			t.Fatal("OAuth-only provider offered API key")
		}
	}
	tt.CloseMenu()
	a.openProviderSecret("test", false)
	tt.Frame()
	tt.Type("fixture-key")
	tt.Key(0, ui.KeyEscape)
	if a.secretOpen || a.secretValue != "" {
		t.Fatal("cancel retained secret in UI state")
	}
}
func TestSettingsPageScrollRetained(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "shortcuts"
	tt := ui.NewTester(a.View, 1000, 650)
	tt.Frame()
	tt.Move(700, 450)
	tt.Scroll(700, 450, 0, 900)
	before := a.settingsScroll["shortcuts"].Y
	if before <= 0 {
		t.Fatalf("shortcuts did not scroll: %+v", a.settingsScroll["shortcuts"])
	}
	a.settingsPage = "appearance"
	for i := 0; i < 8; i++ {
		tt.Frame()
	}
	a.settingsPage = "shortcuts"
	tt.Frame()
	after := a.settingsScroll["shortcuts"].Y
	if after != before {
		t.Fatalf("scroll lost across page eviction: %g -> %g", before, after)
	}
}
