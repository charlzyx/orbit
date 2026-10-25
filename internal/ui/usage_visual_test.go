package orbitui

import (
	"context"
	"encoding/json"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
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
	for _, state := range []string{"trend-tokens", "trend-cost", "trend-latency", "trend-keyboard-focus", "health-wide", "health-narrow", "health-unavailable", "sessions-columns", "filters-models", "filters-models-search", "loading", "refresh-error", "range-presets", "range-start", "range-complete", "filters-no-matches", "trend-latency-p95", "sessions-context", "export-menu", "page-size"} {
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
					case state == "sessions-columns" || state == "sessions-context" || state == "page-size":
						a.usageTable(c, "sessions")
					case state == "export-menu":
						a.usageExportControl(c)
					case strings.HasPrefix(state, "filters-"):
						a.usageDimensionFilter(c, "models")
					case strings.HasPrefix(state, "range-"):
						a.usageRangeControl(c)
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
			case "trend-latency-p95":
				a.usageState.metric = "latency"
				a.usageState.latencyMetric = "p95"
				base := append([]workbench.UsageRecord(nil), a.usageState.index.Records...)
				a.usageState.index.Records = nil
				for _, record := range base {
					for sample := 1; sample <= 20; sample++ {
						r := record
						r.At = r.At.Add(time.Duration(sample) * time.Millisecond)
						r.Duration = time.Duration(sample) * time.Second
						r.HasDuration = true
						a.usageState.index.Records = append(a.usageState.index.Records, r)
					}
				}
				a.usageState.projection = buildUsageProjection(a.usageState.index, a.usageState.filter, "day", time.UTC)
			case "range-presets", "range-start", "range-complete":
				width, height = 600, 922
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
			case "sessions-context":
				if err := tt.RightClick("Fixture task fixture-1"); err != nil {
					t.Fatal(err)
				}
				if !tt.HasText("Copy session ID") || tt.Menu() != nil {
					t.Fatal("painted context unavailable")
				}
			case "export-menu":
				tt.Click("Export")
			case "page-size":
				tt.Click("Rows per page sessions")
			case "sessions-columns":
				tt.Click("Columns")
				if !a.usageState.menu("columns:sessions").Open || !tt.HasText("Show all") {
					t.Fatal("painted source columns panel did not open")
				}
				bounds, ok := tt.Find("Usage columns:sessions menu")
				if !ok || bounds.W < 100 || bounds.H < 200 {
					t.Fatal("column popup lacks painted geometry")
				}
				b, _ := json.MarshalIndent(map[string]any{"note": "Current source columns popup is painted by MyGo; supersedes prior OS-menu metadata-only capture", "panel_bounds": bounds}, "", "  ")
				if err := os.WriteFile(filepath.Join(out, "usage-sessions-columns-native-menu.json"), b, 0644); err != nil {
					t.Fatal(err)
				}
			case "filters-models", "filters-models-search", "filters-no-matches":
				tt.Click("All models")
				if state == "filters-no-matches" {
					tt.Type("zz-no-match")
				}
				if state == "filters-models-search" {
					tt.Type("fixture-large")
					if a.usageState.filterViews["models"].Search != "fixture-large" {
						t.Fatal("filter search did not own focus")
					}
				}
			case "range-presets", "range-start", "range-complete":
				tt.Click("All time")
				if state != "range-presets" {
					tt.Click("Custom…")
					tt.Click("Choose date 2026-10-18")
				}
				if state == "range-complete" {
					tt.Click("Choose date 2026-10-20")
					tt.Click("Oct 18 – Oct 20, 2026")
					tt.Click("Custom…")
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

func TestUsageNativeRangePopoverCalendarSourceFlow(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(20).Children(func() { a.usageRangeControl(c) }) }, 600, 922)
	tt.Frame()
	if err := tt.Click("All time"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Custom…"); err != nil {
		t.Fatal(err)
	}
	if !s.rangeOpen || !s.customOpen || s.customMonth.Year() != 2026 || s.customMonth.Month() != 10 {
		t.Fatal("custom calendar did not seed from actual all-time data")
	}
	if tt.HasText("Apply date range") || tt.HasText("Start date (YYYY-MM-DD)") {
		t.Fatal("invented modal controls")
	}
	if !tt.HasText("Today") || !tt.HasText("Last 7 days") {
		t.Fatal("preset rows disappeared when calendar expanded")
	}
	if _, ok := tt.Find("Choose date 2026-09-30"); ok {
		t.Fatal("out-of-month day is an interactive date instead of padding")
	}
	before := s.filter
	tt.Click("Choose date 2026-10-20")
	if !s.customPicking || !tt.HasText("Pick the end date") || !s.filter.Start.Equal(before.Start) {
		t.Fatal("first click applied instead of staging")
	}
	tt.Key(0, ui.KeyEnter)
	if s.rangeOpen || s.customOpen || !s.filter.Start.Equal(before.Start) {
		t.Fatalf("Enter should dismiss staged calendar without changing range: open=%v custom=%v start=%v before=%v", s.rangeOpen, s.customOpen, s.filter.Start, before.Start)
	}
	tt.Click("All time")
	tt.Click("Custom…")
	tt.Click("Choose date 2026-10-20")
	tt.Click("Choose date 2026-10-18")
	if s.rangeOpen || s.customOpen || s.filterMenu != "" || s.customStart != "2026-10-18" || s.customEnd != "2026-10-20" {
		t.Fatal("second click did not normalize/apply/close")
	}
	if s.filter.End.Format("2006-01-02") != "2026-10-21" {
		t.Fatal("inclusive end not converted to exclusive next civil day")
	}
	tt.Click("Oct 18 – Oct 20, 2026")
	tt.Click("Custom…")
	if !s.customOpen || s.customPicking || !tt.HasText("Oct 18 – Oct 20, 2026") {
		t.Fatal("completed range did not reopen as selected range")
	}
	tt.Click("Calendar previous month")
	if s.customMonth.Month() != 9 {
		t.Fatal("previous month did not advance calendar")
	}
	tt.Click("Calendar next month")
	if s.customMonth.Month() != 10 {
		t.Fatal("next month did not return")
	}
	tt.Key(0, ui.KeyEscape)
	if a.modalOpen() {
		t.Fatal("calendar left modal guard set")
	}
}

func TestUsageSourceSummaryAndCacheActionsAreExclusive(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	s.index.Records[0].Outcome = "error"
	s.index.Records[0].Error = "Fixture provider failure"
	s.projection = buildUsageProjection(s.index, s.filter, "day", time.UTC)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageSummary(c) }) }, 1100, 500)
	tt.Frame()
	if err := tt.Click(a.text("usage.view_failures", "View failures")); err != nil {
		t.Fatal(err)
	}
	if !s.filter.ErrorsOnly || s.filter.CachedOnly {
		t.Fatal("failed action did not set exclusive source filter")
	}
	tt = ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageHealthDetails(c) }) }, 1100, 700)
	tt.Frame()
	if err := tt.Click(a.text("usage.show_cached_only", "Show cached only")); err != nil {
		t.Fatal(err)
	}
	if !s.filter.CachedOnly || s.filter.ErrorsOnly {
		t.Fatal("cache action did not clear failure filter")
	}
	if err := tt.Click(a.text("usage.show_all_requests", "Show all requests")); err != nil {
		t.Fatal(err)
	}
	if s.filter.CachedOnly || s.filter.ErrorsOnly {
		t.Fatal("cache toggle did not restore all requests")
	}
}

func TestUsageSourceEmptyUnreadableFilteredStatesAndHeader(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	for _, tc := range []struct {
		name   string
		index  workbench.UsageIndex
		filter workbench.UsageFilter
		want   string
	}{
		{"empty", workbench.UsageIndex{}, s.filter, "empty"},
		{"unreadable", workbench.UsageIndex{UnreadableFiles: 2}, s.filter, "unreadable"},
		{"warning-not-unreadable", workbench.UsageIndex{Warnings: []string{"enumeration warning"}}, s.filter, "empty"},
		{"errors-alone", workbench.UsageIndex{Errors: s.index.Errors}, s.filter, "empty"},
		{"scope", s.index, workbench.UsageFilter{Start: s.filter.Start, End: s.filter.End, Provider: "missing"}, "filtered"},
		{"outside-range", s.index, workbench.UsageFilter{Start: s.fixtureNow.AddDate(1, 0, 0), End: s.fixtureNow.AddDate(1, 0, 1)}, "empty"},
	} {
		p := buildUsageProjection(tc.index, tc.filter, "day", time.UTC)
		if got := usageEmptyState(tc.index, p); got != tc.want {
			t.Fatalf("%s: %s != %s", tc.name, got, tc.want)
		}
	}
	for _, tc := range []struct {
		age  time.Duration
		want string
	}{{0, "just now"}, {44 * time.Second, "just now"}, {45 * time.Second, "0m ago"}, {90 * time.Minute, "1h ago"}, {49 * time.Hour, "2d ago"}} {
		if got := a.usageAge(tc.age); got != tc.want {
			t.Fatalf("age %v: %s", tc.age, got)
		}
	}
	s.index = workbench.UsageIndex{UnreadableFiles: 2}
	s.projection = buildUsageProjection(s.index, s.filter, "day", time.UTC)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if !tt.HasText("Unable to read usage data") || tt.HasText("Try again") {
		t.Fatal("all-unreadable branch mixed with retry state")
	}
	tt.Click("Export")
	if len(tt.Menu()) > 0 {
		t.Fatal("empty data export enabled")
	}
	s.index = workbench.UsageIndex{}
	s.projection = buildUsageProjection(s.index, s.filter, "day", time.UTC)
	tt.Frame()
	if !tt.HasText("No usage data yet") {
		t.Fatal("empty store message missing")
	}
	s.loaded = false
	s.err = "Explicit fixture: could not scan"
	tt.Frame()
	if !tt.HasText("Unable to load usage data") {
		t.Fatal("first error missing")
	}
	if _, ok := tt.Find("Try again"); !ok {
		t.Fatal("first error retry missing")
	}
}

func TestUsageStaticSourceRowsHaveNoInventedActions(t *testing.T) {
	for _, kind := range []string{"daily", "failures", "breakdown:tools"} {
		t.Run(kind, func(t *testing.T) {
			a := fixtureUsageApp(t)
			s := &a.usageState
			tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(12).Children(func() { a.usageTable(c, kind) }) }, 1100, 700)
			tt.Frame()
			if s.table(kind).List.Selected != nil || s.table(kind).List.CellMenu != nil {
				t.Fatal("static source rows gained selection/context menu")
			}
			tt.Click("Usage " + kind + " table")
			tt.Key(0, ui.KeyEnd)
			tt.Key(0, ui.KeyEnter)
			if s.detailError != nil || s.detailSession != nil || !s.filter.FocusStart.IsZero() {
				t.Fatal("static row created details or focus")
			}
		})
	}
}

func TestUsagePaintedTableAndExportPopovers(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(20).Children(func() { a.usageTable(c, "sessions") }) }, 1100, 800)
	tt.Frame()
	tt.Click("Columns")
	if !s.menu("columns:sessions").Open || len(tt.Menu()) > 0 {
		t.Fatal("Columns should use painted popup, not OS menu")
	}
	if _, ok := tt.Find("Toggle sessions column title"); ok {
		t.Fatal("fixed title included in column menu")
	}
	if _, ok := tt.Find("Toggle sessions column tools"); ok {
		t.Fatal("unused source Tools enum included in menu")
	}
	if tt.HasText("Reset column widths") {
		t.Fatal("invented reset-widths menu action")
	}
	tt.Click("Toggle sessions column workspace")
	if !slices.Contains(s.table("sessions").Hidden, "workspace") || !s.menu("columns:sessions").Open {
		t.Fatal("column toggle did not apply and stay open")
	}
	tt.Click("Show all")
	if len(s.table("sessions").Hidden) > 0 || !s.menu("columns:sessions").Open {
		t.Fatal("show all should restore and stay open")
	}
	tt.Key(0, ui.KeyEnter)
	if a.modalOpen() {
		t.Fatal("Enter did not dismiss table popover")
	}
	tt.Click("Rows per page sessions")
	if !s.menu("page-size:sessions").Open {
		t.Fatal("footer rows popup missing")
	}
	tt.Click("Rows per page 25")
	if !s.menu("page-size:sessions").Open {
		t.Fatal("choosing existing size should leave source menu open")
	}
	tt.Click("Rows per page 10")
	if s.table("sessions").Size != 10 || s.table("sessions").Page != 1 || s.menu("page-size:sessions").Open {
		t.Fatal("new page size did not apply/reset/close")
	}
	if s.table("daily").Size != 25 {
		t.Fatal("page size leaked to independent table")
	}
	tt.Click("Columns")
	tt.Key(0, ui.KeyEscape)
	if a.modalOpen() {
		t.Fatal("Escape left table menu guard set")
	}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt = ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(20).Children(func() { a.usageExportControl(c) }) }, 600, 400)
	tt.Frame()
	tt.Click("Export")
	if !s.menu("export").Open || len(tt.Menu()) > 0 {
		t.Fatal("export should be painted popup")
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 0 || s.menu("export").Open {
		t.Fatal("Enter exported instead of dismissing")
	}
	tt.Click("Export")
	tt.Click(a.text("usage.export_csv", "Filtered requests (CSV)"))
	if len(actions) != 1 || actions[0].Kind != "ExportUsage" || actions[0].Value != "csv" || s.menu("export").Open {
		t.Fatal("CSV action/close mismatch", actions)
	}
}
