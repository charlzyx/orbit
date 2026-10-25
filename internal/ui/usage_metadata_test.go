package orbitui

import (
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestUsageSourceHeaderMetadataAndLocalSearch(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	summary := a.usageSummaryMeta()
	want := "14 requests · " + compactCount(s.projection.Summary.Tokens.Total) + " tokens · 4 sessions"
	if summary != want {
		t.Fatalf("summary %q want %q", summary, want)
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() { a.usageCard(c, "Summary", a.usageSummaryMeta(), func() { ui.Text(c, "Card body sentinel") }) })
	}, 1100, 250)
	tt.Frame()
	title, ok := tt.Find("Summary")
	if !ok {
		t.Fatal("title missing")
	}
	meta, ok := tt.Find(summary)
	if !ok {
		t.Fatal("metadata missing")
	}
	if math.Abs(float64((title.Y+title.H/2)-(meta.Y+meta.H/2))) > .5 || meta.X <= title.X+title.W {
		t.Fatalf("not one centered header row: %#v %#v", title, meta)
	}
	body, _ := tt.Find("Card body sentinel")
	if body.Y < 44 || body.Y > 65 {
		t.Fatalf("invented second description row: %#v", body)
	}
	s.table("sessions").Search = "not a real session"
	if a.usageTableMeta("sessions") != "0 of 4 sessions" || a.usageSummaryMeta() != summary {
		t.Fatal("local search affected page counts")
	}
	before := a.usageTableMeta("series")
	s.table("daily").Search = "impossible"
	if a.usageTableMeta("series") != before {
		t.Fatal("daily search leaked to series")
	}
	if !strings.HasPrefix(a.usageTableMeta("daily"), "0 of ") {
		t.Fatal(a.usageTableMeta("daily"))
	}
	s.table("failures").Search = "not in error text"
	if a.usageTableMeta("failures") != "0 of 1 events · 1 provider · 0 tool · 0 stopped" {
		t.Fatal(a.usageTableMeta("failures"))
	}
	s.breakdown = "models"
	modelMeta := a.usageBreakdownMeta()
	s.table("breakdown:models").Search = "no results"
	if a.usageBreakdownMeta() != modelMeta {
		t.Fatal("breakdown metadata lost source scoped totals")
	}
	s.metric, s.latencyMetric = "latency", "p95"
	if a.usageTrendMeta() != "p95 per bucket · by day" {
		t.Fatal(a.usageTrendMeta())
	}
	s.metric = "cost"
	s.projection.Summary.PricedRequests = 0
	if strings.Contains(a.usageTrendMeta(), "$0") || !strings.Contains(a.usageTrendMeta(), "Unavailable") {
		t.Fatal("missing cost fabricated", a.usageTrendMeta())
	}
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	localized := a.usageSummaryMeta()
	if localized == summary || strings.Contains(localized, "%{") {
		t.Fatal("catalog variables not localized", localized)
	}
}
func TestUsageSourceNonSessionColumnPlans(t *testing.T) {
	cases := []struct {
		kind string
		want []float32
	}{
		{"series", []float32{480, 112, 104, 96}},
		{"daily", []float32{480, 104, 96, 104, 84}},
		{"failures", []float32{108, 96, 150, 200, 446}},
		{"breakdown:models", []float32{460, 92, 80, 80, 80, 88, 72}},
		{"breakdown:tools", []float32{460, 76, 80, 76, 76}},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			cols := usageDataColumns(usageTableColumns(tc.kind, "tokens"), nil, 1000, tc.kind)
			for i, col := range cols {
				if col.Width != tc.want[i] {
					t.Fatalf("%s=%v want %v", col.ID, col.Width, tc.want[i])
				}
				if tc.kind == "breakdown:tools" && (!col.Fixed || col.Sortable) {
					t.Fatal("tools invented reorder/resize/sort", col)
				}
				if col.ID == "message" && col.Sortable {
					t.Fatal("message invented sort")
				}
			}
		})
	}
	cols := usageTableColumns("failures", "tokens")
	cols = slices.DeleteFunc(cols, func(c ui.TableColumn) bool { return c.ID == "message" })
	cols = usageDataColumns(cols, map[string]float32{"model": 190}, 800, "failures")
	if cols[0].Width != 314 {
		t.Fatal("When did not absorb remainder with hidden Message and resized Model", cols[0])
	}
	a := fixtureUsageApp(t)
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.usageTable(c, "breakdown:tools") }) }, 1100, 500)
	tt.Frame()
	before := a.usageState.table("breakdown:tools").Sort
	for _, header := range []string{"Tool", "Calls", "Errors", "Avg", "Max"} {
		tt.Click(header)
		if a.usageState.table("breakdown:tools").Sort != before {
			t.Fatal("native tool header sorted", header)
		}
	}
}

func TestUsageSourceMessageGeometryAndRepeatedAction(t *testing.T) {
	a := fixtureUsageApp(t)
	calls := 0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			a.usageMessageState(c, "icons/filter.svg", "Scoped fixture title", "Exact domain fixture body with no fabricated records", "Clear filters", func() { calls++ })
		})
	}, 1000, 500)
	tt.Frame()
	panel, _ := tt.Find("Usage message state")
	glyph, _ := tt.Find("Usage message glyph")
	title, _ := tt.Find("Scoped fixture title")
	body, _ := tt.Find("Exact domain fixture body with no fabricated records")
	action, _ := tt.Find("Clear filters")
	if panel.H != 400 || glyph.W != 64 || glyph.H != 64 || body.W > 440 {
		t.Fatalf("source geometry wrong: %#v %#v %#v", panel, glyph, body)
	}
	center := panel.X + panel.W/2
	for _, r := range []ui.Rect{glyph, title, body, action} {
		if math.Abs(float64(r.X+r.W/2-center)) > .5 {
			t.Fatalf("message child not centered: %#v center=%v", r, center)
		}
	}
	if !(glyph.Y < title.Y && title.Y < body.Y && body.Y < action.Y && action.Y+action.H < panel.Y+panel.H) {
		t.Fatal("source message child order/action placement")
	}
	tt.Click("Clear filters")
	tt.Click("Clear filters")
	if calls != 2 {
		t.Fatal("repeated source action lost", calls)
	}
}

func TestUsagePaintedSessionContextRowScopeAndLifetime(t *testing.T) {
	a, tt := usageTableFixture(t, "sessions")
	s := &a.usageState
	s.table("sessions").Search = "Task 079"
	tt.Frame()
	var copied []Action
	a.Options.OnAction = func(v Action) { copied = append(copied, v) }
	title, _ := tt.Find("Task 079")
	table, _ := tt.Find("Usage sessions table")
	// Secondary press on the far-right row padding must behave like its text.
	tt.RightClickAt(table.X+table.W-3, title.Y+title.H/2)
	if tt.Menu() != nil || !tt.HasText("Usage session context menu") || s.filter.SessionPath != "" || s.table("sessions").Selected != -1 {
		t.Fatal("context is OS-menu substitute or scopes on right-click", tt.Menu(), s.filter)
	}
	panel, _ := tt.Find("Usage session context menu")
	if math.Abs(float64(panel.X-table.X)) > 1 || math.Abs(float64(panel.Y-(title.Y-(40-title.H)/2))) > 1 {
		t.Fatalf("context is not row-top-left anchored: panel=%#v table=%#v title=%#v", panel, table, title)
	}
	for range 2 {
		if err := tt.Click("Copy session ID"); err != nil {
			t.Fatal(err)
		}
	}
	if len(copied) != 2 || copied[0].Value != "79" || !s.menu("context").Open {
		t.Fatal("repeated copy should remain open", copied)
	}
	tt.Key(0, ui.KeyEscape)
	if !s.menu("context").Open {
		t.Fatal("invented source Escape handler on pointer-only context")
	}
	for range 2 {
		tt.Click("Filter by provider: Provider")
		if !s.menu("context").Open {
			t.Fatal("provider action prematurely dismissed context")
		}
	}
	if len(s.filter.Providers) != 0 {
		t.Fatal("provider toggles not reversible")
	}
	tt.Click("Filter by model: model · Provider")
	if len(s.filter.ModelRefs) != 1 || s.filter.ModelRefs[0].Provider != "provider" || s.filter.ModelRefs[0].Model != "model" || !s.menu("context").Open {
		t.Fatal("context model loses provider identity", s.filter)
	}
	tt.Click("Filter by model: model · Provider")
	tt.Click("Scope to this session")
	if s.menu("context").Open || s.filter.SessionPath != "/sessions/79.jsonl" {
		t.Fatal("scope must close and use exact session path")
	}
	s.filter.SessionPath = ""
	tt.Frame()
	tt.RightClick("Task 079")
	tt.Click("Search sessions")
	if s.menu("context").Open {
		t.Fatal("outside click did not dismiss")
	}
	for i := range s.projection.Sessions {
		if s.projection.Sessions[i].ID == "79" {
			s.projection.Sessions[i].Model = ""
			s.projection.Sessions[i].HasModel = false
		}
	}
	tt.Frame()
	tt.RightClick("Task 079")
	if tt.HasText("Filter by model: ") {
		t.Fatal("missing top model has fake action")
	}
	tt.Click("Open session")
	if len(copied) != 3 || copied[2].Kind != "SwitchSession" || copied[2].Value != "/sessions/79.jsonl" {
		t.Fatal("context open did not use path", copied)
	}
	a.page = "usage"
	tt.Frame()
	tt.RightClick("Task 079")
	a.closeUsage()
	tt.Frame()
	if s.menu("context").Open || tt.HasText("Usage session context menu") {
		t.Fatal("page lifecycle retained context")
	}
}
