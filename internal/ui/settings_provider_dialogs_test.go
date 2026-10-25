package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/background"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
)

func TestSourceProviderEditorModalChipsKeyboardAndPersistence(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.Options.PiBinary = "/nonexistent-fixture-pi"
	a.editProvider(workbench.Provider{})
	tt := ui.NewTester(a.View, 800, 560)
	tt.Frame()
	dialog, ok := tt.Find("Provider editor dialog")
	if !ok || dialog.W != 520 || dialog.H > 512 || dialog.Y < 24 || dialog.Y+dialog.H > 536 {
		t.Fatal("source provider editor escaped viewport", dialog)
	}
	save, _ := tt.Find("Save provider")
	if save.Y+save.H > dialog.Y+dialog.H {
		t.Fatal("footer not reachable", save, dialog)
	}
	// Identity initially owns native text focus. Enter submits validation rather
	// than inserting a newline or being routed into the hidden composer.
	tt.Type("future-endpoint")
	tt.Key(0, ui.KeyEnter)
	d := a.providerState.editor
	if d == nil || d.edit.ID != "future-endpoint" || d.error == "" {
		t.Fatal("identity focus/Enter validation failed", d)
	}
	body, _ := tt.Find("Provider editor body")
	tt.Scroll(body.X+body.W/2, body.Y+body.H/2, 0, 200)
	tt.Frame()
	if err := tt.Click("API family anthropic-messages"); err != nil {
		t.Fatal(err)
	}
	if d.edit.API != "anthropic-messages" {
		t.Fatal("source chip did not change API family")
	}
	d.edit.BaseURL = "https://example.invalid/v1"
	d.models = "one, two"
	d.key = "fixture-secret-must-not-paint"
	tt.Frame()
	if strings.Contains(strings.Join(tt.Texts(), "\n"), d.key) {
		t.Fatal("credential painted")
	}
	if err := tt.Click("Save provider"); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, a, func() bool { return a.providerState.editor == nil && !a.settingsLoading })
	providers, err := a.Service().Providers()
	if err != nil || len(providers) != 1 || providers[0].API != "anthropic-messages" || len(providers[0].ModelIDs) != 2 {
		t.Fatal(providers, err)
	}
}
func TestSourceProviderCredentialModalSecretAndAlternativeMethod(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "p", Name: "Provider Name", APIKey: true, OAuth: true}}}
	a.Model.AuthProviders = []agent.AuthProvider{{ID: "p", Methods: []agent.AuthMethod{{ID: "browser", Label: "Sign in"}, {ID: "api_key"}}}}
	var action Action
	a.Options.OnAction = func(v Action) { action = v }
	a.openProviderSecret("p", false)
	tt := ui.NewTester(a.View, 800, 560)
	tt.Frame()
	panel, ok := tt.Find("Provider credential dialog")
	if !ok || panel.W != 460 || panel.Y < 24 || panel.Y+panel.H > 536 {
		t.Fatal(panel)
	}
	if !tt.HasText("API key — Provider Name") {
		t.Fatal("source key header lost display name")
	}
	tt.Type("fixture-secret")
	if a.secretValue != "fixture-secret" || strings.Contains(strings.Join(tt.Texts(), "\n"), "fixture-secret") {
		t.Fatal("secret focus or masking failed")
	}
	if err := tt.Click("Sign in instead"); err != nil {
		t.Fatal(err)
	}
	if a.secretOpen || a.secretValue != "" || action.Kind != "Login" || action.Value != "p" || action.Data["method"] != "browser" {
		t.Fatal("alternative sign-in retained secret or invented method", action)
	}
	a.openProviderSecret("p", true)
	tt.Frame()
	tt.Type("__Secure-session=fixture")
	tt.Key(0, ui.KeyEscape)
	if a.secretOpen || a.secretValue != "" {
		t.Fatal("Escape retained session cookie")
	}
}
func TestSourceBackgroundBlurGeometryPercentAndSteppedKeyboard(t *testing.T) {
	a := fixture(t)
	drainRegressionUpdates(t, a, func() bool { return !a.backdrop.loading })
	source := "fixture-only.png"
	a.backdrop.config.Source = &source
	a.backdrop.config.Tuning = background.DefaultTuning()
	a.backdrop.config.Blur = 3
	a.backdrop.loading = true // Changes must coalesce, not run a fake image job.
	tt := ui.NewTester(func(c *ui.Context) { a.backgroundSettings(c) }, 800, 400)
	tt.Frame()
	r, ok := tt.Find("Background blur")
	if !ok || r.W != 184 || r.H != 20 || !tt.HasText("50%") {
		t.Fatal("source blur geometry/readout", r, tt.Texts())
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	tt.Key(0, ui.KeyRight)
	if a.backdrop.pending == nil || a.backdrop.pending.Blur != 3.5 {
		t.Fatal("native blur keyboard lost half-sigma step", a.backdrop.pending)
	}
	tt.Key(0, ui.KeyEnd)
	if a.backdrop.pending.Blur != background.BlurMax || !tt.HasText("100%") {
		t.Fatal("native blur End did not reach100%", a.backdrop.pending, tt.Texts())
	}
}

func TestSourceCredentialSaveDoesNotDiscardReopenedDraft(t *testing.T) {
	a := fixture(t)
	a.Options.PiBinary = "/nonexistent-fixture-pi"
	a.openProviderSecret("p", false)
	a.secretValue = "first-fixture-key"
	a.saveProviderSecret()
	// The asynchronous completion is queued until the UI drains updates.
	a.secretOpen = false
	a.openProviderSecret("p", false)
	a.secretValue = "second-unsaved-fixture-key"
	drainRegressionUpdates(t, a, func() bool { return !a.providerState.secretBusy && !a.settingsLoading })
	if !a.secretOpen || a.secretValue != "second-unsaved-fixture-key" {
		t.Fatal("an earlier credential save consumed a newer draft")
	}
	credentials, err := a.Service().Credentials()
	if err != nil || !credentials["p"].Configured {
		t.Fatal("completed save lost its real effect", credentials, err)
	}
}

func TestSourceProviderAdvertisedEmptyMethodsDoNotFallBackToStaticCapability(t *testing.T) {
	a := fixture(t)
	a.Navigate("settings")
	a.settingsPage = "providers"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "p", Name: "Provider", APIKey: true, OAuth: true}}}
	a.Model.AuthProviders = []agent.AuthProvider{{ID: "p", Name: "Provider"}}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if _, ok := tt.Find("API key p"); ok {
		t.Fatal("explicit empty runtime methods were replaced by a guessed static auth control")
	}
	if !tt.HasText(a.Tr("settings.no_sign_in_methods_available")) {
		t.Fatal("empty runtime capability was not explained")
	}
}
