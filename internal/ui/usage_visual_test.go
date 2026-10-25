package orbitui

import (
	"context"
	"encoding/json"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestUsageSourceNumericRegisters(t *testing.T) {
	for _, tc := range []struct {
		n    uint64
		want string
	}{{0, "0"}, {842, "842"}, {1000, "1K"}, {1250, "1.25K"}, {12400, "12.4K"}, {842100, "842K"}, {1284000, "1.28M"}, {18421842, "18.4M"}, {2100000000, "2.1B"}} {
		if got := compactCount(tc.n); got != tc.want {
			t.Errorf("compact %d: %s != %s", tc.n, got, tc.want)
		}
	}
	if usageExact(18421842) != "18,421,842" || usageCount(1250) != "1,250" {
		t.Fatal("exact register lost")
	}
	for _, tc := range []struct {
		n    float64
		want string
	}{{0, "0%"}, {.004, "<0.01%"}, {.42, "0.42%"}, {52, "52%"}, {99.999, "99.9%"}, {100, "100%"}, {math.NaN(), "—"}} {
		if got := usagePercent(tc.n); got != tc.want {
			t.Errorf("percent %f: %s != %s", tc.n, got, tc.want)
		}
	}
	for _, tc := range []struct {
		n    float64
		want string
	}{{0, "$0.00"}, {.004, "<$0.01"}, {.84, "$0.84"}, {1284.2, "$1,284.20"}, {math.NaN(), "—"}} {
		if got := usageCost(tc.n); got != tc.want {
			t.Errorf("cost %f: %s != %s", tc.n, got, tc.want)
		}
	}
	for _, tc := range []struct {
		n    float64
		want string
	}{{0, "0ms"}, {.42, "420ms"}, {1.8, "1.8s"}, {42, "42s"}, {134, "2m 14s"}, {8040, "2h 14m"}, {-1, "—"}} {
		if got := usageDuration(tc.n); got != tc.want {
			t.Errorf("duration %f: %s != %s", tc.n, got, tc.want)
		}
	}
	if usageSpan(76*time.Hour) != "3d 4h" || usageSpan(0) != "—" {
		t.Fatal("session span register")
	}
}
func TestUsageNativeKPISelectsMetricAndUnknownsDoNot(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageSummary(c) }) }, 1100, 450)
	tt.Frame()
	for _, tc := range []struct{ label, metric string }{{"Requests", "requests"}, {"Total tokens", "tokens"}, {"Cache hit rate", "cache"}, {"Avg response", "latency"}} {
		if err := tt.Click(tc.label); err != nil {
			t.Fatal(err)
		}
		if a.usageState.metric != tc.metric {
			t.Fatalf("KPI %s did not select %s", tc.label, tc.metric)
		}
	}
	if tt.HasText("View failures") {
		t.Fatal("zero-error fixture shows failure action")
	}
	p := &a.usageState.projection
	p.Summary.TimedRequests = 0
	p.Summary.Tokens.CacheRead = 0
	p.Summary.Tokens.CacheWrite = 0
	a.usageState.metric = "tokens"
	tt.Frame()
	tt.Click("Avg response")
	tt.Click("Cache hit rate")
	if a.usageState.metric != "tokens" || !tt.HasText("Unavailable") {
		t.Fatal("unavailable KPI became fake zero or interactive")
	}
}
func TestUsageChartNativeKeyboardExactReadoutAndQuartiles(t *testing.T) {
	levels := usageHeatmapThresholds([]float64{0, math.NaN(), 1, 2, 3, 4, 5, 6, 7, 8})
	if levels != [3]float64{2, 4, 6} {
		t.Fatal(levels)
	}
	for _, tc := range []struct {
		n     float64
		level int
	}{{0, 0}, {1, 1}, {3, 2}, {5, 3}, {8, 4}} {
		if got := usageHeatmapLevel(tc.n, levels, 8); got != tc.level {
			t.Fatal(tc, got)
		}
	}
	if usageHeatmapLevel(7, [3]float64{7, 7, 7}, 7) != 4 {
		t.Fatal("single busiest did not use strongest shade")
	}
	a := fixtureUsageApp(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(12).Children(func() { a.usageChart(c) }) }, 1100, 400)
	tt.Frame()
	if err := tt.Click("Usage trend"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyHome)
	if a.usageState.selected != 0 {
		t.Fatal("chart Home did not inspect first bucket")
	}
	tt.Key(0, ui.KeyRight)
	if a.usageState.selected != 1 {
		t.Fatal("chart Right did not advance")
	}
	tt.Key(0, ui.KeyEnter)
	want := a.usageState.projection.Series[1]
	if !a.usageState.filter.FocusStart.Equal(want.At) {
		t.Fatal("chart Enter did not focus inspected bucket")
	}
	tt.Key(0, ui.KeyEnter)
	if !a.usageState.filter.FocusStart.IsZero() {
		t.Fatal("chart repeated Enter did not clear")
	}
	tt.Key(0, ui.KeyEscape)
	if a.usageState.selected != -1 {
		t.Fatal("chart Escape did not end inspection")
	}
	text := usagePointReadout(usagePoint{usageAggregate: usageAggregate{Tokens: workbench.Tokens{Total: 18421842, Input: 18421800, Output: 42}}}, "tokens", "average")
	if !strings.Contains(text, "18,421,842") || !strings.Contains(text, "18,421,800") {
		t.Fatal("readout rounded away exact count", text)
	}
}
func TestUsageEmptyStoreSyncNoticesFirstSession(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	s.lastRequest = time.Now().Add(-time.Minute)
	a.Model.Sessions = nil
	a.syncUsageStore()
	if !s.sourceKnown {
		t.Fatal("empty baseline not recorded")
	}
	var scans atomic.Int32
	s.scan = func(context.Context) (workbench.UsageIndex, error) { scans.Add(1); return usageFixtureIndex(), nil }
	a.Model.Sessions = []SessionRow{{Path: "/new.jsonl", Updated: time.Now()}}
	a.syncUsageStore()
	for s.loading {
		drainUsageUpdate(t, a)
	}
	if scans.Load() != 1 {
		t.Fatal("first session was mistaken for baseline")
	}
	a.syncUsageStore()
	if s.loading {
		t.Fatal("unchanged store rescheduled")
	}
}
func TestUsageNativeSourcePanelsAndUnavailableSemantics(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageHealthDetails(c) }) }, 1100, 700)
	tt.Frame()
	for _, label := range []string{"Token health", "Token composition", "Cache performance", "Hit rate over time", a.text("usage.show_cached_only", "Show cached only")} {
		if !tt.HasText(label) {
			t.Fatal("missing source health panel", label)
		}
	}
	if tt.HasText("Response time") {
		t.Fatal("model response timing misplaced in health")
	}
	a.usageState.projection.Summary.Tokens.CacheRead = 0
	a.usageState.projection.Summary.Tokens.CacheWrite = 0
	tt.Frame()
	if tt.HasText("0%") || tt.HasText(a.text("usage.show_cached_only", "Show cached only")) {
		t.Fatal("missing cache became measured zero")
	}
	a.usageState.projection.Summary.Tokens.CacheWrite = 100
	tt.Frame()
	if !tt.HasText("0%") {
		t.Fatal("write-only measured zero disappeared")
	}
}
func TestNativeUsageComponentCaptures(t *testing.T) {
	out := os.Getenv("ORBIT_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_CAPTURE_DIR for labelled CPU fixtures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	for _, state := range []string{"trend-tokens", "trend-cost", "trend-latency", "trend-keyboard-focus", "health-wide", "health-narrow", "health-unavailable", "sessions-columns", "filters-models", "filters-models-search", "loading", "refresh-error"} {
		t.Run(state, func(t *testing.T) {
			a := fixtureUsageApp(t)
			width, height := 1100, 700
			view := func(c *ui.Context) {
				ui.Column(c).Fill().Padding(20).Gap(12).Background(a.Theme.Background).Children(func() {
					ui.Text(c, "Explicit Usage fixture · "+state).FontSize(11).TextColor(a.Theme.TextMuted)
					switch {
					case strings.HasPrefix(state, "trend-"):
						a.usageChart(c)
					case strings.HasPrefix(state, "health-"):
						a.usageHealthDetails(c)
					case state == "sessions-columns":
						a.usageTable(c, "sessions")
					case strings.HasPrefix(state, "filters-"):
						a.usageDimensionFilter(c, "models")
					case state == "loading":
						a.usageSkeleton(c)
					case state == "refresh-error":
						a.usagePage(c)
					}
				})
			}
			switch state {
			case "trend-cost":
				a.usageState.metric = "cost"
			case "trend-latency":
				a.usageState.metric = "latency"
			case "health-narrow":
				width, height = 650, 920
			case "health-unavailable":
				for i := range a.usageState.index.Records {
					r := &a.usageState.index.Records[i]
					r.Tokens.CacheRead = 0
					r.Tokens.CacheWrite = 0
					r.Tokens.Total = usageAdd(r.Tokens.Input, r.Tokens.Output)
				}
				a.usageState.projection = buildUsageProjection(a.usageState.index, a.usageState.filter, "day", time.UTC)
			case "refresh-error":
				a.usageState.err = "Explicit fixture: session store permission denied"
			}
			tt := ui.NewTester(view, width, height)
			tt.SetDark(true)
			tt.Frame()
			if strings.HasPrefix(state, "trend-") {
				if r, ok := tt.Find("Response time"); !ok || r.X < 0 || r.W < 30 {
					t.Fatalf("timing label bounds: %+v %t", r, ok)
				}
			}
			switch state {
			case "trend-keyboard-focus":
				tt.Click("Usage trend")
				tt.Key(0, ui.KeyHome)
				tt.Key(0, ui.KeyEnter)
			case "sessions-columns":
				tt.Click("Columns")
				if len(tt.Menu()) == 0 {
					t.Fatal("columns menu did not open")
				}
				b, _ := json.MarshalIndent(map[string]any{"note": "CPU Tester does not paint OS-native menus", "native_menu": tt.Menu()}, "", "  ")
				if err := os.WriteFile(filepath.Join(out, "usage-sessions-columns-native-menu.json"), b, 0644); err != nil {
					t.Fatal(err)
				}
			case "filters-models", "filters-models-search":
				tt.Click("All models")
				if state == "filters-models-search" {
					tt.Type("fixture-large")
					if a.usageState.filterViews["models"].Search != "fixture-large" {
						t.Fatal("filter search did not own focus")
					}
				}
			}
			f, err := os.Create(filepath.Join(out, "usage-"+state+"-fixture-dark.png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, tt.Image())
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUsageNativeNarrowHealthActionAndCalendarRange(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(20).Children(func() { a.usageHealthDetails(c) }) }, 650, 920)
	tt.Frame()
	if err := tt.Click(a.text("usage.show_cached_only", "Show cached only")); err != nil {
		t.Fatal(err)
	}
	if !a.usageState.filter.CachedOnly {
		t.Fatal("narrow health action clipped or inactive")
	}
	s := &a.usageState
	s.customOpen = true
	s.customMonth = time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	old := s.filter
	tt = ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(20).Children(func() { a.usageCalendarPicker(c) }) }, 480, 500)
	tt.Frame()
	tt.Click("Choose date 2026-10-20")
	if !s.customPicking || s.customStart != "2026-10-20" || s.customEnd != "" || !s.filter.Start.Equal(old.Start) {
		t.Fatal("first date should stage, not apply")
	}
	tt.Click("Choose date 2026-10-18")
	if s.customOpen || s.customPicking || s.preset != "custom" || s.filter.Start.Format("2006-01-02") != "2026-10-18" || s.filter.End.Format("2006-01-02") != "2026-10-21" {
		t.Fatal("reversed inclusive range not normalized", s.filter)
	}
	at := time.Date(2026, 10, 25, 0, 0, 0, 0, time.Local)
	a.scopeUsageDay(at)
	if s.granularity != "hour" {
		t.Fatal("day drilldown did not adapt to hourly source buckets")
	}
	a.scopeUsageDay(at)
	if s.granularity != "day" {
		t.Fatal("restored date range retained hourly buckets")
	}
}

func TestUsageResponsiveSessionPlanAndTemporalTies(t *testing.T) {
	columns := usageTableColumns("sessions", "tokens")
	visible := usageSessionColumns(columns, nil, 1116)
	has := func(cols []ui.TableColumn, id string) bool {
		for _, c := range cols {
			if c.ID == id {
				return true
			}
		}
		return false
	}
	if !has(visible, "started") || has(visible, "input") || has(visible, "output") || has(visible, "cache") || has(visible, "tools") || !has(visible, "open") {
		t.Fatal("source optional plan incorrect", visible)
	}
	small := usageSessionColumns(columns, map[string]float32{"title": 240}, 700)
	if has(small, "started") || small[0].Width != 240 || !has(small, "duration") {
		t.Fatal("narrow plan lost required columns or dragged title")
	}
	rows := []usageTableRow{{usageAggregate: usageAggregate{Key: "later", Name: "A", ID: "a", At: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Requests: 1}}, {usageAggregate: usageAggregate{Key: "earlier", Name: "Z", ID: "z", At: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), Requests: 1}}}
	for _, kind := range []string{"series", "daily", "failures"} {
		q := queryUsageTable(rows, &usageTableState{Sort: ui.SortOrder{Column: "requests", Descending: true}, Size: 25, Page: 1}, kind, "tokens", "average")
		if q.Rows[0].Key != "earlier" {
			t.Fatal("tie ordered by title/date text, not timestamp", kind)
		}
	}
	a, tt := usageTableFixture(t, "sessions")
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	if err := tt.Click("Open Task 079"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "SwitchSession" || actions[0].Value != "/sessions/79.jsonl" {
		t.Fatal("trailing open used wrong path", actions)
	}
}

func TestUsageNativeLatencySegmentsAndUnavailableReason(t *testing.T) {
	a := fixtureUsageApp(t)
	a.usageState.metric = "latency"
	a.usageState.latencyMetric = "average"
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(12).Children(func() { a.usageChart(c) }) }, 1100, 500)
	tt.Frame()
	for _, label := range []string{"Average", "P50", "P95", "P99"} {
		if _, ok := tt.Find(label); !ok {
			t.Fatalf("inline latency segment missing: %s", label)
		}
	}
	tt.Click("P95")
	if a.usageState.latencyMetric != "average" || len(tt.Menu()) != 0 {
		t.Fatal("undersampled percentile activated or dropdown opened")
	}
	samples := make([]time.Duration, 20)
	for i := range samples {
		samples[i] = time.Duration(i+1) * time.Second
	}
	a.usageState.projection.Latency = usageLatencyFrom(samples)
	tt.Frame()
	for _, tc := range []struct{ label, metric string }{{"P95", "p95"}, {"P50", "p50"}, {"P99", "p99"}, {"Average", "average"}, {"P95", "p95"}} {
		if err := tt.Click(tc.label); err != nil {
			t.Fatal(err)
		}
		if a.usageState.latencyMetric != tc.metric {
			t.Fatalf("segment %s did not persist selection", tc.label)
		}
	}
	saved, err := os.ReadFile(filepath.Join(a.Options.Store.RootDir, "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	var prefs map[string]json.RawMessage
	if err := json.Unmarshal(saved, &prefs); err != nil {
		t.Fatal(err)
	}
	raw := prefs["latency_metric"]
	if string(raw) != `"p95"` {
		t.Fatalf("latency enum not persisted: %s", raw)
	}
	summary := a.usageState.projection.Summary
	summary.PricedRequests, summary.TimedRequests = 0, 0
	got := a.usageUnavailableReason(summary)
	if !strings.Contains(got, "cost, latency not shown") || !strings.Contains(got, "price table") || !strings.Contains(got, "request gaps") {
		t.Fatal(got)
	}
	summary.PricedRequests, summary.TimedRequests = 1, 1
	if got := a.usageUnavailableReason(summary); got != "" {
		t.Fatalf("partial pricing invented warning: %s", got)
	}
}
