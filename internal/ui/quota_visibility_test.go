package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestQuotaPopoverOffersRecoverableHideControl(t *testing.T) {
	a := fixture(t)
	reports := []agent.QuotaReport{{Provider: "alpha", Windows: []agent.QuotaWindow{{ID: "5h", UsedPercent: quotaPtr(20)}}}, {Provider: "beta", Note: "fixture reported usage"}}
	open := true
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, reports, "supported", "alpha") }, 800, 700)
	if _, ok := tt.Find(a.text("session.hide_provider", "Hide provider") + " beta"); !ok {
		t.Fatal("quota popover lacks source hide provider action")
	}
}

func waitQuotaPreference(t *testing.T, a *App) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case f := <-a.updates:
			if f != nil {
				f()
			}
		default:
		}
		if a.quotaVisibility.persisted == a.quotaVisibility.revision.Load() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("quota preference did not persist", a.quotaVisibility.persisted, a.quotaVisibility.revision.Load())
}
func quotaFixtureReports() []agent.QuotaReport {
	return []agent.QuotaReport{{Provider: "alpha", Windows: []agent.QuotaWindow{{ID: "5h", UsedPercent: quotaPtr(20)}}}, {Provider: "beta", Note: "fixture reported usage"}}
}
func TestQuotaMountedHideCollapsedRecoveryShowAndRestart(t *testing.T) {
	a := fixture(t)
	reports := quotaFixtureReports()
	open := true
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, reports, "supported", "alpha") }, 800, 700)
	hide := a.text("session.hide_provider", "Hide provider") + " beta"
	if err := tt.Click(hide); err != nil {
		t.Fatal(err)
	}
	if !a.quotaVisibility.hidden["beta"] {
		t.Fatal("mounted hide action did not update state")
	}
	if _, ok := tt.Find(hide); ok {
		t.Fatal("hidden provider still in visible list")
	}
	if _, ok := tt.Find("Hidden providers"); !ok {
		t.Fatal("hidden provider not recoverable")
	}
	show := a.text("session.show_provider", "Show provider") + " beta"
	if _, ok := tt.Find(show); ok {
		t.Fatal("hidden group was not initially collapsed")
	}
	if err := tt.Click("Hidden providers"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find(show); !ok {
		t.Fatal("expanded hidden group lacks restore action")
	}
	waitQuotaPreference(t, a)
	reloaded := New(Options{Store: a.Options.Store})
	if !reloaded.quotaVisibility.hidden["beta"] {
		t.Fatal("hidden preference lost on shell recreation")
	}
	if err := tt.Click(show); err != nil {
		t.Fatal(err)
	}
	if a.quotaVisibility.hidden["beta"] {
		t.Fatal("mounted restore action did not unhide")
	}
	waitQuotaPreference(t, a)
	if len(a.Options.Store.HiddenQuotaProviders()) != 0 {
		t.Fatal("restored provider remained persisted hidden")
	}
}
func TestQuotaHidingDoesNotFilterHeadlineOrSettingsCards(t *testing.T) {
	a := fixture(t)
	reports := quotaFixtureReports()
	before := quotaSummary(reports, "alpha")
	a.toggleQuotaProvider("alpha")
	a.toggleQuotaProvider("beta")
	visible, hidden := a.partitionQuotaReports(quotaReports(reports))
	if len(visible) != 0 || len(hidden) != 2 {
		t.Fatal("hide-all partition incorrect")
	}
	if after := quotaSummary(reports, "alpha"); after != before {
		t.Fatal("hide preference changed usage headline")
	}
	tt := ui.NewTester(func(c *ui.Context) {
		for _, report := range reports {
			a.runtimeQuotaCard(c.Key(report.Provider), report, time.Now())
		}
	}, 800, 700)
	if !tt.HasText("alpha") || !tt.HasText("beta") {
		t.Fatal("settings card list was filtered")
	}
	if _, ok := tt.Find(a.text("session.show_provider", "Show provider") + " alpha"); ok {
		t.Fatal("popup-only visibility control leaked into settings")
	}
	waitQuotaPreference(t, a)
}
func TestQuotaRapidTogglesPersistLatestStateIncludingAbsentProvider(t *testing.T) {
	a := fixture(t)
	for range 101 {
		a.toggleQuotaProvider("not-currently-reported")
	}
	for range 100 {
		a.toggleQuotaProvider("alpha")
	}
	waitQuotaPreference(t, a)
	ids := a.Options.Store.HiddenQuotaProviders()
	if len(ids) != 1 || ids[0] != "not-currently-reported" {
		t.Fatal("older worker overwrote latest state", ids)
	}
}
func TestQuotaWriteFailureLeavesCurrentStateAndReportsError(t *testing.T) {
	a := fixture(t)
	if err := os.MkdirAll(filepath.Join(a.Options.Store.RootDir, preferences.HiddenQuotaProvidersFile), 0700); err != nil {
		t.Fatal(err)
	}
	a.toggleQuotaProvider("alpha")
	deadline := time.Now().Add(3 * time.Second)
	reported := false
	for time.Now().Before(deadline) && !reported {
		select {
		case f := <-a.updates:
			if f != nil {
				f()
			}
		default:
		}
		for _, toast := range a.toastStack.Items() {
			if strings.Contains(toast.Body, preferences.HiddenQuotaProvidersFile) {
				reported = true
			}
		}
		time.Sleep(time.Millisecond)
	}
	if !a.quotaVisibility.hidden["alpha"] || a.quotaVisibility.persisted != 0 {
		t.Fatal("failed write fabricated persistence or rolled back current UI")
	}
	if !reported {
		t.Fatal("failed preference write was silent")
	}
}

func TestQuotaVisibilityChineseCountAndHiddenGroup(t *testing.T) {
	a := fixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	a.quotaVisibility.hidden["beta"] = true
	reports := quotaFixtureReports()
	open := true
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, reports, "supported", "alpha") }, 800, 700)
	if !tt.HasText("Provider usage · 1 个服务商") {
		t.Fatal("provider count did not use source Chinese translation")
	}
	if !tt.HasText("已隐藏（1）") {
		t.Fatal("hidden group count did not use source Chinese translation")
	}
	if err := tt.Click("Hidden providers"); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find(a.text("session.show_provider", "Show provider") + " beta"); !ok {
		t.Fatal("localized restore action missing")
	}
}
