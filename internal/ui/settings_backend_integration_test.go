package orbitui

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// These tests bridge source-backed Usage records into the real native page and
// its projection. They are not OS/platform or baseline-pixel acceptance tests.
func TestSettingsBackendProjectionUsesModelPricingCoverage(t *testing.T) {
	at := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	index := workbench.UsageIndex{Records: []workbench.UsageRecord{
		{At: at, Session: "priced", Provider: "p", Model: "m", Tokens: workbench.Tokens{Input: 1, Total: 1}, Cost: 1, HasCost: true, ModelPriced: true},
		{At: at, Session: "priced", Provider: "p", Model: "m", Tokens: workbench.Tokens{Input: 1, Total: 1}, ModelPriced: true},
		{At: at, Session: "unpriced", Provider: "q", Model: "m", Tokens: workbench.Tokens{Input: 1, Total: 1}, HasCost: true},
	}}
	p := buildUsageProjection(index, workbench.UsageFilter{}, "day", time.UTC)
	for _, g := range p.Models {
		want := 0
		if g.Provider == "p" {
			want = 2
		}
		if g.Priced != want {
			t.Errorf("provider %s pricing coverage=%d, want%d", g.Provider, g.Priced, want)
		}
	}
	if p.Summary.PricedRequests != 2 {
		t.Fatal(p.Summary)
	}
}
func TestSettingsBackendTrendDoesNotCompressMissingDays(t *testing.T) {
	from := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	index := workbench.UsageIndex{Records: []workbench.UsageRecord{{At: from, Tokens: workbench.Tokens{Total: 1}}, {At: from.AddDate(0, 0, 2), Tokens: workbench.Tokens{Total: 2}}}}
	p := buildUsageProjection(index, workbench.UsageFilter{Start: from, End: from.AddDate(0, 0, 3)}, "day", time.UTC)
	if len(p.Series) != 3 {
		t.Fatalf("missing day compressed out of trend: got%d buckets, want3", len(p.Series))
	}
	if p.Series[1].Requests != 0 || !p.Series[1].At.Equal(from.AddDate(0, 0, 1)) {
		t.Fatal("gap is not a zero-valued day")
	}
}
func TestSettingsBackendNativeUnbilledErrorsRemainVisible(t *testing.T) {
	a := fixtureUsageApp(t)
	at := a.usageState.fixtureNow
	a.usageState.index = workbench.UsageIndex{Errors: []workbench.UsageError{{At: at, Session: "fixture", SessionPath: "/fixture/session.jsonl", Kind: "provider", Message: "Unbilled fixture failure"}}}
	a.usageState.projection = buildUsageProjection(a.usageState.index, a.usageState.filter, "day", time.UTC)
	a.usageState.mode = "details"
	a.usageState.detail = "failures"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if tt.HasText("No usage in this range") {
		t.Fatal("unbilled failure was hidden by the no-request empty state")
	}
	if !tt.HasText("Failure details") {
		found := false
		for _, text := range tt.Texts() {
			if strings.Contains(text, "Unbilled fixture failure") {
				found = true
			}
		}
		if !found {
			t.Fatal("unbilled failure did not reach native records")
		}
	}
}
func TestSettingsBackendUsagePreferencesPreserveUnknownAndDefaultDescending(t *testing.T) {
	a := fixture(t)
	if !a.usageState.descending {
		t.Fatal("absent usage prefs must default to token-descending session order")
	}
	if e := a.Options.Store.Update("usage.json", func(m map[string]json.RawMessage) error { m["future"] = json.RawMessage(`{"keep":true}`); return nil }); e != nil {
		t.Fatal(e)
	}
	a.usageState.mode = "details"
	a.saveUsagePrefs()
	var m map[string]json.RawMessage
	if e := a.Options.Store.Read("usage.json", &m); e != nil {
		t.Fatal(e)
	}
	var future map[string]bool
	if err := json.Unmarshal(m["future"], &future); err != nil {
		t.Fatal(err)
	}
	if !future["keep"] || string(m["usage_mode"]) != "\"details\"" {
		t.Fatal(m)
	}
}
