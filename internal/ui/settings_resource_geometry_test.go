package orbitui

import (
	"math"
	"testing"

	"github.com/charlzyx/orbit/internal/modelui"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// CPU layout assertions, not an OS execution or pixel-parity claim. The
// source header has an inner 4-DIP title gap and outer 12-DIP toolbar gap.
func TestSourceResourceSettingsHeaderToolbarGeometry(t *testing.T) {
	for _, page := range []string{"models", "providers"} {
		t.Run(page, func(t *testing.T) {
			a := fixture(t)
			a.Navigate("settings")
			a.settingsPage = page
			a.Models.SetSnapshot(modelui.Snapshot{Connected: true})
			a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "source-fixture", Name: "Source fixture", APIKey: true}}}
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			title, _ := tt.Find("Settings page title")
			actionLabel, cardLabel := "Update model list", "Model empty state"
			if page == "providers" {
				actionLabel = "Refresh providers"
				cardLabel = "Provider card source-fixture"
			}
			action, ok := tt.Find(actionLabel)
			card, found := tt.Find(cardLabel)
			if title.X != 399 || title.Y != 20 || title.H != 21 || !ok || math.Abs(float64(action.Y-124)) > .1 || action.H != 32 || !found || card.X != 399 || math.Abs(float64(card.Y-189)) > .1 {
				t.Fatal("source resource header/body bounds", title, action, card)
			}
			t.Logf("%s title=%+v action=%+v card=%+v", page, title, action, card)
		})
	}
}
