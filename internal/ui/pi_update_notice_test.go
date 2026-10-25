package orbitui

import (
	"strings"
	"testing"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func TestPiUpdateStatusSourceNotifications(t *testing.T) {
	for _, tc := range []struct{ state, key string }{
		{"available", "pi_update.available_background"},
		{string(native.PiUnsupportedManager), "pi_update.unsupported_package_manager"},
		{string(native.PiUnknownVersion), "pi_update.unsupported_unknown"},
	} {
		t.Run(tc.state, func(t *testing.T) {
			a := fixture(t)
			actions := 0
			a.Options.OnAction = func(Action) { actions++ }
			page := a.page
			event := native.PiUpdateEvent{State: tc.state, Latest: "0.60.2", Detail: "secret-output-that-must-not-be-rendered"}
			a.notifyPiUpdateStatus(event)
			want := orbitassets.Translate(a.catalog, a.english, tc.key, map[string]string{"version": event.Latest})
			items := a.toastStack.Items()
			if len(items) != 1 || items[0].Body != want || items[0].Kind != native.ToastInfo {
				t.Fatalf("event %s: %+v; want %q info", tc.state, items, want)
			}
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			texts := strings.Join(tt.Texts(), "\n")
			if !strings.Contains(texts, want) || strings.Contains(texts, event.Detail) {
				t.Fatalf("incorrect mounted notification: %s", texts)
			}
			if actions != 0 || a.page != page {
				t.Fatal("informational event dispatched an action or navigated")
			}
		})
	}
}

func TestPiUpdateStatusCatalogAndUnknownRemainSilent(t *testing.T) {
	a := fixture(t)
	a.catalog = map[string]string{"pi_update.available_background": "版本 %{version} 可用"}
	a.notifyPiUpdateStatus(native.PiUpdateEvent{State: "available", Latest: "0.60.2"})
	items := a.toastStack.Items()
	if len(items) != 1 || items[0].Body != "版本 0.60.2 可用" {
		t.Fatalf("localized substitution: %+v", items)
	}
	for _, state := range []string{"deferred", "updating", "unchanged", "unexpected", "failed", "updated", "migration_required"} {
		a.notifyPiUpdateStatus(native.PiUpdateEvent{State: state, Detail: "do-not-render"})
	}
	if len(a.toastStack.Items()) != 1 {
		t.Fatal("unknown or separately owned state emitted a notice")
	}
	a.catalog = nil
	a.notifyPiUpdateStatus(native.PiUpdateEvent{State: string(native.PiUnknownVersion)})
	items = a.toastStack.Items()
	if len(items) != 2 || items[1].Body != a.english["pi_update.unsupported_unknown"] {
		t.Fatalf("English fallback: %+v", items)
	}
}
