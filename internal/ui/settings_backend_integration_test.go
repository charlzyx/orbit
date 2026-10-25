package orbitui

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// Pinned Rust considers requests and tools, not standalone error rows, when
// deciding whether the dashboard is empty (model.rs:614, aggregate.rs:1150).
func TestSettingsBackendNativeUnbilledErrorsRespectSourceEmptyGate(t *testing.T) {
	for _, mixed := range []bool{false, true} {
		t.Run(fmt.Sprint("mixed=", mixed), func(t *testing.T) {
			a := fixtureUsageApp(t)
			s := &a.usageState
			at := s.fixtureNow
			s.index = workbench.UsageIndex{Errors: []workbench.UsageError{{At: at, Session: "fixture", SessionPath: "/fixture/session.jsonl", Kind: "provider", Message: "Unbilled fixture failure"}}}
			if mixed {
				s.index.Records = []workbench.UsageRecord{{At: at, Session: "fixture", SessionPath: "/fixture/session.jsonl", Tokens: workbench.Tokens{Input: 5, Output: 2, Total: 7}}}
			}
			s.projection = buildUsageProjection(s.index, s.filter, "day", time.UTC)
			s.mode = "details"
			s.detail = "failures"
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Frame()
			if len(s.index.Errors) != 1 || len(s.projection.Errors) != 1 {
				t.Fatal("empty gate deleted actual error data")
			}
			visible := false
			for _, text := range tt.Texts() {
				if strings.Contains(text, "Unbilled fixture failure") {
					visible = true
				}
			}
			if mixed {
				if !visible || tt.HasText("No usage data yet") {
					t.Fatal("mixed real-request/unbilled-error data was hidden")
				}
			} else {
				if visible || !tt.HasText("No usage data yet") {
					t.Fatal("error-only store does not match source empty gate")
				}
			}
			if s.projection.Summary.PricedRequests != 0 || s.projection.Summary.Cost != 0 {
				t.Fatal("unpriced activity fabricated cost")
			}
			var exported bytes.Buffer
			if err := s.index.ExportJSON(&exported, s.filter, workbench.UsageExportOptions{Now: at, Location: time.UTC}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(exported.String(), "Unbilled fixture failure") {
				t.Fatal("empty visible gate lost exportable error")
			}
		})
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
