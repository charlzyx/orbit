package orbitui

import (
	"strings"
	"testing"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/egoist/mygo/ui"
)

func TestPiMigrationNoticeSourceCatalogAndDelivery(t *testing.T) {
	for _, lang := range []string{"en", "zh-CN", "missing-language", "fixture-localized"} {
		t.Run(lang, func(t *testing.T) {
			a := fixture(t)
			a.catalog, _ = orbitassets.Locale(lang)
			if lang == "fixture-localized" {
				a.catalog = map[string]string{"pi_update.migration_required": "迁移到官方安装以更新 pi v%{version}。"}
			}
			actions := 0
			a.Options.OnAction = func(Action) { actions++ }
			page := a.page
			a.notifyPiMigrationRequired("0.60.1")
			want := orbitassets.Translate(a.catalog, a.english, "pi_update.migration_required", map[string]string{"version": "0.60.1"})
			items := a.toastStack.Items()
			if len(items) != 1 || items[0].Body != want || items[0].Kind != native.ToastError {
				t.Fatalf("source migration notice: %+v, want %q error", items, want)
			}
			if !strings.Contains(want, "0.60.1") || strings.Contains(want, "%{version}") || strings.Contains(want, "Settings → Runtime") {
				t.Fatalf("incorrect version or legacy navigation: %q", want)
			}
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			if !strings.Contains(strings.Join(tt.Texts(), "\n"), want) {
				t.Fatal("notice was not rendered")
			}
			if actions != 0 || a.page != page {
				t.Fatal("notice unexpectedly dispatched or navigated")
			}
		})
	}
}
