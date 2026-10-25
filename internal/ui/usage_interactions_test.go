package orbitui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func usageTableFixture(t *testing.T, kind string) (*App, *ui.Tester) {
	t.Helper()
	a := fixtureUsageApp(t)
	for i := 0; i < 80; i++ {
		a.usageState.projection.Sessions = append(a.usageState.projection.Sessions, usageAggregate{Key: fmt.Sprint(i), ID: fmt.Sprint(i), Name: fmt.Sprintf("Task %03d", i), Path: fmt.Sprintf("/sessions/%d.jsonl", i), Requests: i + 1, Tokens: workbench.Tokens{Total: uint64(i) * 100}, Provider: "provider", Model: "model", Workspace: "/work"})
	}
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(10).Gap(8).Children(func() { a.usageTable(c, kind) }) }, 1000, 720)
	tt.Frame()
	return a, tt
}
func TestUsageIndependentTablesQuerySearchSortPageAndColumns(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	for _, kind := range []string{"sessions", "daily", "series", "failures", "breakdown:models", "breakdown:providers", "breakdown:workspaces", "breakdown:tools"} {
		tab := s.table(kind)
		tab.Search = kind
		tab.Page = 3
		tab.setPageSize(10)
		tab.Page = 2
		tab.toggleColumn("requests", "title")
	}
	sessions := s.table("sessions")
	sessions.Search = "fixture"
	sessions.Sort = ui.SortOrder{Column: "title"}
	sessions.Page = 90
	result := queryUsageTable(a.usageTableRows("sessions"), sessions, "sessions", "tokens", "average")
	if result.Total != 4 || result.Page != 1 || result.Pages != 1 {
		t.Fatal(result)
	}
	for _, kind := range []string{"daily", "series", "failures", "breakdown:models", "breakdown:providers", "breakdown:workspaces", "breakdown:tools"} {
		tab := s.table(kind)
		if tab.Search != kind || tab.Page != 2 || !slices.Contains(tab.Hidden, "requests") {
			t.Fatalf("session table mutated %s: %+v", kind, tab)
		}
	}
	before := s.projection.Summary
	sessions.Search = "impossible"
	if got := queryUsageTable(a.usageTableRows("sessions"), sessions, "sessions", "tokens", "average"); got.Total != 0 || got.Page != 1 {
		t.Fatal(got)
	}
	if s.projection.Summary != before || s.filter.Search != "" {
		t.Fatal("table search narrowed entire analytics")
	}
	sessions.toggleColumn("title", "title")
	if slices.Contains(sessions.Hidden, "title") {
		t.Fatal("identity column hidden")
	}
	sessions.setPageSize(7)
	if sessions.Size != 10 {
		t.Fatal("invalid page size accepted")
	}
}
func TestUsageNativeTableSortCyclePagingResizeAndKeyboard(t *testing.T) {
	a, tt := usageTableFixture(t, "sessions")
	tab := a.usageState.table("sessions")
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tab.Page = 2
	tt.Frame()
	if err := tt.Click("Session"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tab.Sort.Column != "title" || tab.Sort.Descending || tab.Page != 1 {
		t.Fatalf("first sort %+v", tab)
	}
	if err := tt.Click("Session"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tab.Sort.Descending {
		t.Fatal("second click not descending")
	}
	tt.Click("Session")
	tt.Frame()
	if tab.Sort != usageDefaultSort("sessions") {
		t.Fatalf("third click not source default: %+v", tab.Sort)
	}
	if err := tt.Click("Next"); err != nil {
		t.Fatal(err)
	}
	if tab.Page != 2 {
		t.Fatal("next page failed")
	}
	tt.Click("Search sessions")
	tt.Type("Task 079")
	tt.Frame()
	if tab.Page != 1 || !tt.HasText("Task 079") {
		t.Fatalf("search reset failed %v", tt.Texts())
	}
	r, ok := tt.Find("Task 079")
	if !ok {
		t.Fatal("row missing")
	}
	tt.ClickAt(float32(r.X+4), float32(r.Y+r.H/2))
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "SwitchSession" || actions[0].Value != "/sessions/79.jsonl" {
		t.Fatal("keyboard opened wrong session", actions)
	}
	if a.usageState.filter.SessionPath != "/sessions/79.jsonl" {
		t.Fatal("single click did not scope exact session")
	}
	a.usageState.detailSession = nil
	tab.Search = ""
	tt.Frame()
	h, ok := tt.Find("Session")
	if !ok {
		t.Fatal("header missing")
	}
	edge := float32(h.X+h.W) - 2
	y := float32(h.Y + h.H/2)
	tt.Press(edge, y)
	tt.Move(edge+40, y)
	tt.Release(edge+40, y)
	if len(tab.List.Columns.Widths) == 0 {
		t.Fatal("column resize did not persist in table state")
	}
	for _, width := range tab.List.Columns.Widths {
		if width < 48 {
			t.Fatal("column narrower than source guard")
		}
	}
}
func TestUsageComparisonCacheLatencyThresholdsAndFocusedPreviousWindow(t *testing.T) {
	start := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)
	index := workbench.UsageIndex{ScannedAt: start.Add(24 * time.Hour)}
	for i := 0; i < 20; i++ {
		for _, previous := range []bool{false, true} {
			at := start.Add(time.Duration(i) * time.Minute)
			tokens := uint64(200)
			duration := time.Duration(i+1) * time.Second
			cacheRead := uint64(100)
			if previous {
				at = at.Add(-24 * time.Hour)
				tokens = 100
				duration = time.Second
				cacheRead = 10
			}
			index.Records = append(index.Records, workbench.UsageRecord{At: at, Session: "s", Provider: "a", Model: "same", Tokens: workbench.Tokens{Input: 10, CacheRead: cacheRead, Total: tokens}, HasDuration: true, Duration: duration})
		}
	}
	f := workbench.UsageFilter{Start: start, End: start.Add(24 * time.Hour), FocusStart: start, FocusEnd: start.Add(time.Hour)}
	p := buildUsageProjection(index, f, "hour", time.UTC)
	if p.Previous == nil || p.Previous.Requests != 20 || p.Summary.Requests != 20 {
		t.Fatal("previous inherited current focus", p.Previous)
	}
	if p.Latency.P50 != 10*time.Second || p.Latency.P95 != 19*time.Second || p.Latency.P99 != 20*time.Second {
		t.Fatal(p.Latency)
	}
	if v, ok := p.Series[0].Latency.value("p95"); !ok || v != 19 {
		t.Fatalf("per bucket p95 %v %v", v, ok)
	}
	if got := usageComparison(p, "tokens"); got != "+100.0% vs previous period" {
		t.Fatal(got)
	}
	insights := usageInsights(p, false)
	if len(insights) < 2 || insights[0].Key != "usage.insight_token_delta" || insights[1].Key != "usage.insight_cache_hit" {
		t.Fatal(insights)
	}
	sparse := usageLatencyFrom([]time.Duration{0, 2 * time.Second})
	if avg, ok := sparse.value("average"); !ok || avg != 1 {
		t.Fatal("measured zero excluded")
	}
	for _, metric := range []string{"p50", "p95", "p99"} {
		if _, ok := sparse.value(metric); ok {
			t.Fatal("insufficient percentile fabricated")
		}
	}
	if _, ok := (usageLatency{}).value("average"); ok {
		t.Fatal("missing average fabricated")
	}
	if rate, ok := usageCacheRate(workbench.Tokens{Input: 100, CacheWrite: 10}); !ok || rate != 0 {
		t.Fatal("write-only traffic is a measured zero")
	}
	if _, ok := usageCacheRate(workbench.Tokens{Input: 100}); ok {
		t.Fatal("no cache semantics fabricated")
	}
	if _, ok := usageDelta(2, 0); ok {
		t.Fatal("zero baseline gave percentage")
	}
	all := buildUsageProjection(index, workbench.UsageFilter{Start: time.UnixMilli(0), End: start.Add(24 * time.Hour)}, "day", time.UTC)
	if all.Previous != nil || len(all.Series) != 20 || all.Series[0].At.Day() != 1 {
		t.Fatal("all-time epoch or comparison error", len(all.Series))
	}
}
func TestUsageSessionLifetimeTopModelToolsAndCoarseBuckets(t *testing.T) {
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	index := workbench.UsageIndex{Sessions: []workbench.UsageSession{{ID: "same", Path: "/a", Title: "A", Started: start, Ended: start.Add(48 * time.Hour)}}, Records: []workbench.UsageRecord{{At: start, Session: "same", SessionPath: "/a", SessionTitle: "A", Provider: "p", Model: "small", Tokens: workbench.Tokens{Total: 1}}, {At: start.Add(time.Hour), Session: "same", SessionPath: "/a", SessionTitle: "A", Provider: "q", Model: "large", Tokens: workbench.Tokens{Total: 100}, HasDuration: true, Duration: time.Second}}, Tools: []workbench.ToolRun{{At: start, Session: "same", SessionPath: "/a", Name: "read", HasResult: true, Success: false, HasDuration: true, Duration: 2 * time.Second}}}
	p := buildUsageProjection(index, workbench.UsageFilter{}, "week", time.UTC)
	r := p.Sessions[0]
	if r.Model != "large" || r.Provider != "q" || r.ID != "same" || r.Ended.Sub(r.Started) != 48*time.Hour || r.Duration != time.Second || r.ToolFailures != 1 || r.ToolRuns != 1 {
		t.Fatal(r)
	}
	if len(p.Daily) != 1 || p.Daily[0].At.Weekday() != time.Monday || p.ToolLatency.Max != 2*time.Second {
		t.Fatal("coarse buckets/tool timing mismatch")
	}
}
func TestUsageProviderScopedMultiselectSearchAllClearAndNativeEscape(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	s.index.Records = append(s.index.Records, workbench.UsageRecord{At: s.fixtureNow, Provider: "second", Model: "fixture-fast", Tokens: workbench.Tokens{Total: 10}})
	s.projection = buildUsageProjection(s.index, s.filter, "day", time.UTC)
	opts := s.filterOptions("models", "fixture-fast")
	if len(opts) != 2 || opts[0].Key == opts[1].Key {
		t.Fatal(opts)
	}
	s.toggleFilter("models", opts[0])
	if len(s.filter.ModelRefs) != 1 {
		t.Fatal("no paired scope")
	}
	s.selectFilterOptions("models", opts)
	if len(s.filter.ModelRefs) != 2 {
		t.Fatal("select all omitted pair")
	}
	s.clearDimension("models")
	if len(s.filter.ModelRefs) > 0 {
		t.Fatal("clear failed")
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() { a.usageDimensionFilter(c, "models") })
	}, 800, 560)
	tt.Frame()
	if err := tt.Click("All models"); err != nil {
		t.Fatal(err)
	}
	if s.filterMenu != "models" || !a.modalOpen() {
		t.Fatal("picker did not block app menus")
	}
	if err := tt.Click("Filter models"); err != nil {
		t.Fatal(err)
	}
	tt.Type("second")
	tt.Frame()
	if err := tt.Click("Select all"); err != nil {
		t.Fatal(err)
	}
	if len(s.filter.ModelRefs) != len(s.projection.ModelRefOptions) {
		t.Fatal(s.filter.ModelRefs)
	}
	tt.Key(0, ui.KeyEscape)
	if s.filterMenu != "" || s.filterViews["models"].Open {
		t.Fatal("Escape did not close filter")
	}
	if len(s.filter.ModelRefs) != len(s.projection.ModelRefOptions) {
		t.Fatal("Escape erased applied scope")
	}
}
func TestUsagePreferencesSourceSchemaAndTransientIsolation(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	start := time.Date(2026, 10, 1, 0, 0, 0, 0, time.Local)
	s.preset = "custom"
	s.filter.Start, s.filter.End = start, start.AddDate(0, 0, 3)
	s.latencyMetric = "p95"
	s.hiddenColumns = []string{"input", "output"}
	s.breakdownSort = "requests"
	s.breakdownDescending = false
	s.filter.ModelRefs = []workbench.UsageModelRef{{Provider: "secret-provider", Model: "selected"}}
	s.table("sessions").Search = "private search"
	s.table("daily").Size = 100
	s.table("daily").List.Columns.Widths = map[string]float32{"date": 333}
	a.saveUsagePrefs()
	raw, err := os.ReadFile(filepath.Join(a.Options.Store.RootDir, "usage.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "private search") || strings.Contains(string(raw), "secret-provider") || strings.Contains(string(raw), "333") {
		t.Fatal("transient state persisted", string(raw))
	}
	var saved map[string]json.RawMessage
	json.Unmarshal(raw, &saved)
	var end int64
	json.Unmarshal(saved["custom_end"], &end)
	if end != start.AddDate(0, 0, 2).UnixMilli() {
		t.Fatal("custom date persisted exclusive rather than inclusive", end)
	}
	b := New(Options{Store: a.Options.Store})
	if b.usageState.latencyMetric != "p95" || !b.usageState.filter.End.Equal(s.filter.End) || len(b.usageState.hiddenColumns) != 2 || b.usageState.table("daily").Size != 25 || b.usageState.table("sessions").Search != "" {
		t.Fatal("source view prefs failed roundtrip", b.usageState)
	}
	if err := a.Options.Store.Write("usage.json", map[string]any{"preset": "wrong", "metric": "bad", "detail_tab": "unknown", "session_sort": "invented"}); err != nil {
		t.Fatal(err)
	}
	b.loadUsagePrefs()
	if b.usageState.preset != "last7" || b.usageState.metric != "tokens" || b.usageState.detail != "sessions" || b.usageState.sort != "tokens" {
		t.Fatal("invalid enum mislabeled actual range")
	}
}
func TestUsageCalendarScanAnchorDSTToggleAndKeyboard(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	anchor := time.Date(2026, 3, 10, 12, 0, 0, 0, loc)
	index := workbench.UsageIndex{ScannedAt: anchor}
	grid, first, last := usageCalendarBounds(index, loc)
	if !last.Equal(time.Date(2026, 3, 10, 0, 0, 0, 0, loc)) || !first.Equal(last.Add(-365*24*time.Hour)) || grid.Weekday() != time.Monday {
		t.Fatal(grid, first, last)
	}
	a := fixtureUsageApp(t)
	original := a.usageState.filter
	originalPreset := a.usageState.preset
	at := bucketStart(a.usageState.fixtureNow, "day", time.Local)
	a.scopeUsageDay(at)
	if a.usageState.preset != "custom" || !a.usageState.filter.Start.Equal(at) {
		t.Fatal("day did not scope")
	}
	a.scopeUsageDay(at)
	if a.usageState.preset != originalPreset || !a.usageState.filter.Start.Equal(original.Start) || !a.usageState.filter.End.Equal(original.End) {
		t.Fatal("second day click did not restore range")
	}
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageHeatmap(c) }) }, 900, 400)
	tt.Frame()
	r, ok := tt.Find("Usage activity calendar")
	if !ok {
		t.Fatal("calendar accessibility target missing")
	}
	tt.ClickAt(float32(r.X+2), float32(r.Y+2))
	tt.Key(0, ui.KeyEnd)
	if a.usageState.calendarSelected < 0 {
		t.Fatal("keyboard inspection absent")
	}
	before := a.usageState.calendarSelected
	tt.Key(0, ui.KeyLeft)
	if a.usageState.calendarSelected != before-7 {
		t.Fatal("calendar arrow geometry incorrect")
	}
	tt.Key(0, ui.KeyEnter)
	if a.usageState.preset != "custom" {
		t.Fatal("keyboard day did not activate")
	}
	tt.Key(0, ui.KeyEscape)
	if a.usageState.calendarSelected != -1 {
		t.Fatal("inspection Escape failed")
	}
}
func drainUsageUpdate(t *testing.T, a *App) {
	t.Helper()
	select {
	case apply := <-a.updates:
		apply()
	case <-time.After(3 * time.Second):
		t.Fatal("usage worker timed out")
	}
}
func TestUsageRefreshCoalescesAndFailureKeepsLastSnapshot(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	before := s.index
	updated := s.updated
	var calls atomic.Int32
	release := make(chan struct{})
	s.scan = func(context.Context) (workbench.UsageIndex, error) {
		calls.Add(1)
		<-release
		return workbench.UsageIndex{}, errors.New("fixture scan denied")
	}
	a.refreshUsage()
	a.refreshUsage()
	a.refreshUsage()
	if !s.loading || !s.pending {
		t.Fatal("refresh not coalesced")
	}
	close(release)
	for s.loading {
		drainUsageUpdate(t, a)
	}
	if calls.Load() != 2 || s.loading || s.pending || s.err != "fixture scan denied" || !s.updated.Equal(updated) || len(s.index.Records) != len(before.Records) {
		t.Fatalf("failed refresh state incorrect calls=%d loading=%t pending=%t error=%q", calls.Load(), s.loading, s.pending, s.err)
	}
	s.scan = func(context.Context) (workbench.UsageIndex, error) { return usageFixtureIndex(), nil }
	a.refreshUsage()
	for s.loading {
		drainUsageUpdate(t, a)
	}
	if s.err != "" || !s.loaded {
		t.Fatal("retry did not recover")
	}
}

func TestUsageProjectionRejectsLateResultAndClonesFilterSlices(t *testing.T) {
	a := &App{updates: make(chan func(), 4)}
	s := &a.usageState
	s.index = usageFixtureIndex()
	s.loaded = true
	s.granularity = "day"
	first, second := make(chan struct{}), make(chan struct{})
	seen := make(chan workbench.UsageFilter, 2)
	s.project = func(index workbench.UsageIndex, f workbench.UsageFilter, g string, l *time.Location) usageProjection {
		if f.Provider == "old" {
			<-first
		} else {
			<-second
		}
		seen <- f
		return buildUsageProjection(index, f, g, l)
	}
	s.filter.Provider = "old"
	s.filter.Models = []string{"fixture-fast"}
	a.projectUsage()
	s.filter.Models[0] = "fixture-large"
	s.filter.Provider = "fixture-provider"
	a.projectUsage()
	close(second)
	for s.projecting {
		drainUsageUpdate(t, a)
	}
	latest := s.computed
	if s.projection.Summary.Requests == 0 {
		t.Fatal("new filter not applied")
	}
	close(first)
	// No fixture initialization jobs: this is necessarily the stale apply.
	drainUsageUpdate(t, a)
	if s.computed != latest || s.projecting || s.projection.Summary.Requests == 0 {
		t.Fatal("late projection overwrote current filter")
	}
	for n := 0; n < 2; n++ {
		f := <-seen
		if f.Provider == "old" && f.Models[0] != "fixture-fast" {
			t.Fatal("projection captured mutable filter slice")
		}
	}
}
func TestUsageNativeFilterConfirmDismissesWithoutChangingScope(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() { a.usageDimensionFilter(c, "models") })
	}, 800, 560)
	tt.Click("All models")
	tt.Type("fixture")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if len(s.filter.ModelRefs) != 0 || s.filterMenu != "" || a.modalOpen() {
		t.Fatal("source Enter should only dismiss")
	}
	tt.Click("All models")
	tt.Type("zz-no-match")
	if !tt.HasText("No match for “zz-no-match”") {
		t.Fatal("no-match text missing")
	}
	if _, ok := tt.Find("All"); !ok {
		t.Fatal("All row disappeared on no matches")
	}
	if err := tt.Click("Select all"); err != nil {
		t.Fatal(err)
	}
	if len(s.filter.ModelRefs) != len(s.projection.ModelRefOptions) {
		t.Fatal("Select all filtered by query")
	}
	if err := tt.Click("All"); err != nil {
		t.Fatal(err)
	}
	if len(s.filter.ModelRefs) != 0 {
		t.Fatal("All did not clear dimension")
	}
	tt.Key(0, ui.KeyEscape)
	if a.modalOpen() {
		t.Fatal("closed popup blocks app")
	}
}
func TestUsageNativeContextMenuExportAndColumnPreferences(t *testing.T) {
	a, tt := usageTableFixture(t, "sessions")
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tab := a.usageState.table("sessions")
	tab.Search = "Task 079"
	a.usageState.contextRow = -1
	tt.Frame()
	if err := tt.RightClick("Task 079"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Copy session ID"); err != nil {
		t.Fatalf("%v; context=%+v fresh=%v row=%d texts=%v", err, a.usageState.menu("context"), a.usageState.contextFresh, a.usageState.contextRow, tt.Texts())
	}
	if len(actions) != 1 || actions[0].Kind != "CopyText" || actions[0].Value != "79" {
		t.Fatal(actions)
	}
	tt.Click("Scope to this session")
	if a.usageState.filter.SessionPath != "/sessions/79.jsonl" {
		t.Fatal("context scope lost path identity")
	}
	tt.Click("Columns")
	if err := tt.Click("Toggle sessions column input"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tab.Hidden, "input") {
		t.Fatal("native column chooser did not hide")
	}
	var saved map[string]json.RawMessage
	a.Options.Store.Read("usage.json", &saved)
	if !strings.Contains(string(saved["hidden_columns"]), "input") {
		t.Fatal("session visibility not persisted")
	}
	if !a.usageState.menu("columns:sessions").Open {
		t.Fatal("source column toggle should keep popup open")
	}
	tt.Click("Show all")
	if len(tab.Hidden) > 0 {
		t.Fatal("reset did not show columns")
	}
	a = fixtureUsageApp(t)
	a.usageState.table("sessions").Search = "private task search"
	actions = nil
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt = ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Click("Export")
	if err := tt.Click(a.text("usage.export_json", "Current view (JSON)")); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "ExportUsage" || actions[0].Value != "json" {
		t.Fatal(actions)
	}
	opts := actions[0].Data["options"].(workbench.UsageExportOptions)
	filter := actions[0].Data["filter"].(workbench.UsageFilter)
	if opts.SessionSearch != "private task search" || filter.Search != "" || len(opts.Insights) == 0 {
		t.Fatal("export conflated page scope with table search or dropped actual insights", opts)
	}
	a.usageState.projecting = true
	tt.Frame()
	tt.Click("Export")
	if a.usageState.menu("export").Open {
		t.Fatal("stale projection remained exportable")
	}
}
func TestUsagePrunesRemovedScopeAndPageLifecycle(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	s.filter.SessionPath = "/removed"
	s.filter.ModelRefs = []workbench.UsageModelRef{{Provider: "missing", Model: "removed"}, {Provider: "fixture-provider", Model: "fixture-fast"}}
	if !pruneUsageFilter(s.index, &s.filter) || s.filter.SessionPath != "" || len(s.filter.ModelRefs) != 1 {
		t.Fatal("removed source scope stayed selected")
	}
	s.filterMenu = "models"
	s.filterViews = map[string]*usageFilterView{"models": {Open: true}}
	s.customOpen = true
	s.calendarSelected = 3
	a.closeUsage()
	if a.modalOpen() || s.filterViews["models"].Open || s.calendarSelected != -1 {
		t.Fatal("closed Usage leaked picker state")
	}
	s.updated = time.Now()
	s.loaded = true
	s.scroll.Y = 200
	var calls atomic.Int32
	s.scan = func(context.Context) (workbench.UsageIndex, error) { calls.Add(1); return usageFixtureIndex(), nil }
	a.openUsage()
	if calls.Load() != 0 || s.loading || s.scroll.Y != 0 {
		t.Fatal("fresh open reloaded or retained old scroll")
	}
	s.updated = time.Now().Add(-61 * time.Second)
	a.openUsage()
	for s.loading {
		drainUsageUpdate(t, a)
	}
	if calls.Load() != 1 {
		t.Fatal("stale open skipped scan")
	}
}
