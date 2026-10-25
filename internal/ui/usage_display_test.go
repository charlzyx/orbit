package orbitui

import (
	"math"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestUsageSourceDisplayLabelsKeepRawIdentities(t *testing.T) {
	for _, tc := range []struct{ id, want string }{{"", "unknown"}, {"chatgpt", "ChatGPT"}, {"open_ai-cloud", "Open ai cloud"}, {"éclair-api", "éclair api"}} {
		if got := usageProviderLabel(tc.id); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	for _, tc := range []struct{ id, want string }{{"namespace/model:cloud", "model:cloud"}, {"a/b/c", "b/c"}, {"a/", "a/"}, {"plain", "plain"}} {
		if got := usageModelLabel(tc.id); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	a := fixtureUsageApp(t)
	now := a.usageState.now()
	a.usageState.index = workbench.UsageIndex{Records: []workbench.UsageRecord{{At: now, Session: "fixture", Provider: "chatgpt", Model: "namespace/shared:cloud", Workspace: "/fixture/project", Tokens: workbench.Tokens{Input: 1, Total: 1}}}}
	a.usageState.projection = buildUsageProjection(a.usageState.index, workbench.UsageFilter{}, "day", time.UTC)
	opts := a.usageState.filterOptions("models", "ChatGPT")
	if len(opts) != 1 || opts[0].Label != "shared:cloud" || opts[0].Sub != "ChatGPT" || opts[0].Model.Model != "namespace/shared:cloud" {
		t.Fatal(opts)
	}
	a.usageState.toggleFilter("models", opts[0])
	if a.usageState.filter.ModelRefs[0].Model != "namespace/shared:cloud" {
		t.Fatal("display label became filter identity")
	}
	row := a.usageTableRows("sessions")[0]
	if a.usageCell(row, "workspace", "sessions") != "project" || a.usageCell(row, "provider", "sessions") != "ChatGPT" || a.usageCell(row, "model", "sessions") != "shared:cloud · ChatGPT" {
		t.Fatal(row)
	}
	model := a.usageTableRows("breakdown:models")[0]
	if model.Name != "shared:cloud" || model.Sub != "ChatGPT" {
		t.Fatal(model)
	}
	failure := usageTableRow{Failure: &workbench.UsageError{Kind: "tool"}}
	if a.usageCellColor("failures", failure, "kind") != a.Theme.Warning {
		t.Fatal("source tool failures use warning ink")
	}
}
func TestUsageSourceBucketLabelsAndSearch(t *testing.T) {
	a := fixtureUsageApp(t)
	// Source labels use the process local zone; construct 09:00 civil time,
	// not 09:00 UTC (which is 18:00 in the required Tokyo suite).
	at := time.Date(2026, 10, 5, 9, 0, 0, 0, time.Local)
	a.usageState.fixtureNow = at
	for _, tc := range []struct{ granularity, label, stamp, not string }{{"hour", "09:00", "Oct 5, 09:00", "Monday"}, {"day", "Oct 5", "Oct 5, 2026", "October"}, {"week", "Wk of Oct 5", "Week of Oct 5, 2026", "09:00"}, {"month", "Oct", "October 2026", "Oct 5"}} {
		a.usageState.projection = usageProjection{Granularity: tc.granularity, Series: []usagePoint{{At: at}}, Daily: []usagePoint{{At: at}}}
		rows := a.usageTableRows("series")
		if rows[0].Label != tc.label || rows[0].Stamp != tc.stamp || a.usageCell(rows[0], "time", "series") != tc.stamp {
			t.Fatal(tc, rows)
		}
		tab := &usageTableState{Page: 1, Size: 10, Search: tc.label}
		if queryUsageTable(rows, tab, "series", "tokens", "average").Total != 1 {
			t.Fatal("label not searchable", tc)
		}
		tab.Search = tc.not
		if queryUsageTable(rows, tab, "series", "tokens", "average").Total != 0 {
			t.Fatal("invented search term matched", tc)
		}
	}
	if usageBucketLabel(at.AddDate(-1, 0, 0), "day", at) != "Oct 5, 2025" || usageBucketLabel(at.AddDate(-1, 0, 0), "month", at) != "Oct 2025" {
		t.Fatal("year omitted on old bucket")
	}
}
func TestUsageSourceBreakdownBarsSearchTotalsAndInteraction(t *testing.T) {
	a := fixtureUsageApp(t)
	kind := "breakdown:models"
	a.usageState.projection.Models = []usageAggregate{{Key: "a", Name: "alpha", Sub: "Provider A", Provider: "a", Model: "m", Requests: 1, Tokens: workbench.Tokens{Total: 60}}, {Key: "b", Name: "beta", Sub: "Provider B", Provider: "b", Model: "m", Requests: 3, Tokens: workbench.Tokens{Total: 40}}}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(10).Children(func() { a.usageBreakdownChart(c, kind) })
	}, 1000, 400)
	tt.Frame()
	track, _ := tt.Find("Usage breakdown track a")
	fill, _ := tt.Find("Usage breakdown fill a")
	if track.W == 0 || math.Abs(float64(fill.W/track.W-.6)) > .002 || !tt.HasText("60%") {
		t.Fatal("bar normalized to largest row instead of searched total", track, fill, tt.Texts())
	}
	tab := a.usageState.table(kind)
	tab.Search = "Provider B"
	tt.Frame()
	if _, ok := tt.Find("Usage breakdown bar a"); ok || !tt.HasText("100%") {
		t.Fatal("chart search did not match table query", tt.Texts())
	}
	tab.Search = ""
	tab.Page = 3
	a.setUsageTableSort(kind, "requests")
	if tab.Page != 3 {
		t.Fatal("source breakdown sort does not reset page")
	}
	rows := a.usageBreakdownChartRows(kind)
	if rows.Rows[0].Key != "a" {
		t.Fatal("chart ignored source current sort")
	}
	a.setUsageTableSort(kind, "requests")
	rows = a.usageBreakdownChartRows(kind)
	if rows.Rows[0].Key != "b" {
		t.Fatal("chart kept stale token rank instead of source executable order")
	}
	tt.Frame()
	if err := tt.Click("Usage breakdown bar a"); err != nil {
		t.Fatal(err)
	}
	if len(a.usageState.filter.ModelRefs) != 1 || a.usageState.filter.ModelRefs[0].Provider != "a" {
		t.Fatal("bar click lost paired model identity")
	}
	a.usageState.projection.ToolGroups = []usageAggregate{{Key: "bash", Name: "bash", Requests: 2, ToolFailures: 1}, {Key: "custom", Name: "custom", Requests: 1}}
	kind = "breakdown:tools"
	a.usageState.filter.ModelRefs = nil
	tt.Frame()
	if !tt.HasText("Terminal") || !tt.HasText("66.7%") {
		t.Fatal("tool family or share missing", tt.Texts())
	}
	tt.Click("Usage breakdown bar bash")
	if len(a.usageState.filter.ModelRefs) != 0 || a.usageState.filter.Provider != "" {
		t.Fatal("static tool bar changed scope")
	}
	a.usageState.table(kind).Search = "terminal"
	tt.Frame()
	if _, ok := tt.Find("Usage breakdown bar custom"); ok {
		t.Fatal("tool family search not applied")
	}
}

func TestUsageSourceTableSearchDoesNotBridgeFieldsAndDisplayOnlyColumns(t *testing.T) {
	rows := []struct {
		kind  string
		row   usageTableRow
		query string
	}{{"breakdown:models", usageTableRow{usageAggregate: usageAggregate{Name: "alpha", Sub: "provider"}}, "alpha provider"}, {"sessions", usageTableRow{usageAggregate: usageAggregate{Name: "alpha", ID: "session"}}, "alpha session"}, {"failures", usageTableRow{Failure: &workbench.UsageError{Model: "alpha", SessionTitle: "session"}}, "alpha session"}, {"series", usageTableRow{Point: &usagePoint{}, Label: "Oct 5", Stamp: "Oct 5, 2026"}, "5 Oct"}}
	for _, tc := range rows {
		if usageRowMatches(tc.row, tc.kind, tc.query) {
			t.Fatal("search joined independent source fields", tc)
		}
	}
	for _, col := range usageTableColumns("breakdown:models", "tokens") {
		if col.ID == "share" && (!col.Fixed || col.Sortable) {
			t.Fatal("source Share is display-only", col)
		}
	}
	cols := usageSessionColumns(usageTableColumns("sessions", "tokens"), nil, 1100)
	open := cols[len(cols)-1]
	if open.ID != "open" || !open.Fixed {
		t.Fatal("source Open cannot be resized", open)
	}
}

func TestUsageSourcePickerSearchFooterAndRetainedPointerHighlight(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(20).Children(func() { a.usageDimensionFilter(c, "models") })
	}, 800, 500)
	if err := tt.Click("All models"); err != nil {
		t.Fatal(err)
	}
	head, ok := tt.Find("Usage filter search head")
	if !ok || head.H != 36 {
		t.Fatal("source search-head height", head, ok)
	}
	all, _ := tt.Find("Select all")
	clear, _ := tt.Find("Clear")
	if all.H != 28 || clear.H != 28 || math.Abs(float64(all.W-clear.W)) > .01 {
		t.Fatal("source footer has equal 28px buttons", all, clear)
	}
	options := s.filterOptions("models", "")
	second, ok := tt.Find(options[1].Label + " · " + options[1].Sub)
	if !ok {
		t.Fatal("second paired model absent")
	}
	tt.Move(second.X+8, second.Y+second.H/2)
	tt.Frame()
	if s.filterViews["models"].Selected != 2 {
		t.Fatal("pointer did not retain source row highlight")
	}
	tt.Move(4, 4)
	tt.Frame()
	if s.filterViews["models"].Selected != 2 {
		t.Fatal("mouse leave cleared source highlight")
	}
	if err := tt.Click("Filter models"); err != nil {
		t.Fatal(err)
	}
	tt.Type("zz-none")
	tt.Frame()
	if s.filterViews["models"].Selected != 0 || !tt.HasText("No match for “zz-none”") {
		t.Fatal("search did not reset highlight/preserve no-match row")
	}
	if err := tt.Click("Select all"); err != nil {
		t.Fatal(err)
	}
	if len(s.filter.ModelRefs) != len(options) {
		t.Fatal("paint correction changed global select-all behavior")
	}
	tt.Key(0, ui.KeyEnter)
	if s.filterMenu != "" {
		t.Fatal("Enter must dismiss, not select")
	}
}

func TestUsageSourceLocalizedColumnHeadersMenusAndSearch(t *testing.T) {
	a := fixtureUsageApp(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	for _, tc := range []struct {
		kind, id, metric, want string
		menu                   bool
	}{
		{"sessions", "title", "tokens", "会话", false}, {"sessions", "provider", "tokens", "服务商", false},
		{"daily", "cache-hit", "tokens", "缓存命中", false}, {"failures", "kind", "tokens", "类型", false},
		{"series", "value", "latency", "延迟", false}, {"series", "value", "latency", "指标", true},
		{"breakdown:tools", "tool", "tokens", "Tool", false}, {"breakdown:tools", "errors", "tokens", "Errors", false},
		{"breakdown:tools", "errors", "tokens", a.text("usage.metric_errors", "Errors"), true},
		{"breakdown:models", "share", "tokens", "Share", false}, {"breakdown:models", "share", "tokens", a.text("usage.col_share", "Share"), true},
	} {
		if got := a.usageColumnTitle(tc.kind, tc.id, tc.metric, tc.menu); got != tc.want {
			t.Fatal(tc, got)
		}
	}
	for _, tc := range []struct{ kind, want string }{{"sessions", "搜索会话…"}, {"breakdown:models", "搜索明细…"}, {"series", "搜索时段…"}, {"daily", "搜索时段…"}, {"failures", "搜索失败项…"}} {
		if a.usageSearchPlaceholder(tc.kind) != tc.want {
			t.Fatal(tc, a.usageSearchPlaceholder(tc.kind))
		}
	}
	a.usageState.metric = "latency"
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(a.Theme.Native())
		ui.Column(c).Fill().Padding(10).Children(func() { a.usageTable(c, "series") })
	}, 1000, 500)
	tt.Frame()
	if !tt.HasText("时间") || !tt.HasText("延迟") {
		t.Fatal("localized source labels not painted", tt.Texts())
	}
	if err := tt.Click(a.text("usage.columns", "Columns")); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("指标") {
		t.Fatal("Series menu copied metric header instead of source Metric label", tt.Texts())
	}
}
