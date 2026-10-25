package orbitui

import (
	"errors"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"reflect"
	"strings"
	"testing"
)

func TestProviderTerminalLoginUsesExplicitUnsupportedCapabilityAndLiteralArguments(t *testing.T) {
	x := maintenanceFixture(t)
	x.UI.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "fixture-provider", Name: "Fixture", OAuth: true}, {ID: "api-only", APIKey: true}}}
	calls := 0
	var argv []string
	open := func(bin string, args ...string) error { calls++; argv = append([]string{bin}, args...); return nil }
	for _, support := range []string{"", "supported"} {
		x.UI.Model.Auth.Support = support
		if x.loginProviderTerminal("fixture-provider", open) == nil {
			t.Fatal("unknown/supported RPC used fallback")
		}
	}
	x.UI.Model.Auth.Support = "unsupported"
	for _, id := range []string{"", "fixture;echo nope", "../bad", "api-only", "unknown"} {
		if x.loginProviderTerminal(id, open) == nil {
			t.Fatalf("accepted unsupported id %q", id)
		}
	}
	if calls != 0 {
		t.Fatal("validation started terminal")
	}
	if err := x.loginProviderTerminal("fixture-provider", open); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(argv, []string{"pi", "/login", "fixture-provider"}) || calls != 1 {
		t.Fatal(argv, calls)
	}
	if err := x.loginProviderTerminal("fixture-provider", func(string, ...string) error { return errors.New("fixture terminal unavailable") }); err == nil {
		t.Fatal("terminal launch failure hidden")
	}
}
func TestProviderLegacyOAuthIsExplicitAndSearchPlaceholderIsLocalized(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "providers"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "fixture", Name: "Fixture", OAuth: true}}}
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	_, searchPresent := tt.Find("Search providers")
	if !searchPresent || a.Tr("app.search_providers") != "Search providers…" || strings.Contains(strings.Join(tt.Texts(), " "), "settings.providers_search") {
		t.Fatal("provider search leaked untranslated key")
	}
	if _, ok := tt.Find("Sign in terminal fixture"); ok {
		t.Fatal("pending support offered legacy login")
	}
	a.Model.Auth.Support = "unsupported"
	tt.Frame()
	if err := tt.Click("Sign in terminal fixture"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "LoginTerminal" || actions[0].Value != "fixture" {
		t.Fatal(actions)
	}
	a.Model.Auth.Support = "supported"
	tt.Frame()
	if _, ok := tt.Find("Sign in terminal fixture"); ok {
		t.Fatal("supported RPC retained legacy login")
	}
}
