package orbitui

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestUsageSourceTableCacheAndEndedDateRegisters(t *testing.T) {
	a := fixtureUsageApp(t)
	now := a.usageState.now()
	index := workbench.UsageIndex{ScannedAt: now}
	tokens := []workbench.Tokens{{Input: 100, Output: 10, Total: 110}, {Input: 100, Output: 10, CacheWrite: 50, Total: 160}, {Input: 100, Output: 10, CacheRead: 20, Total: 130}}
	for i, id := range []string{"absent", "write-only", "read"} {
		path := "/fixture/" + id + ".jsonl"
		index.Records = append(index.Records, workbench.UsageRecord{At: now, Session: id, SessionPath: path, SessionTitle: id, Provider: "provider", Model: "model", Tokens: tokens[i]})
		index.Sessions = append(index.Sessions, workbench.UsageSession{ID: id, Path: path, Title: id, Started: now.Add(-time.Duration(3-i) * 24 * time.Hour), Ended: now.Add(-time.Duration(i) * time.Hour)})
	}
	a.usageState.projection = buildUsageProjection(index, workbench.UsageFilter{}, "day", time.UTC)
	rows := a.usageTableRows("sessions")
	wants := map[string]string{"absent": "—", "write-only": "0", "read": "20"}
	for _, r := range rows {
		if got := a.usageCell(r, "cache", "sessions"); got != wants[r.ID] {
			t.Fatalf("%s cache %q want %q", r.ID, got, wants[r.ID])
		}
		if a.usageCell(r, "model", "sessions") != "model · provider" {
			t.Fatal("provider-scoped model display lost")
		}
		if got := a.usageCell(r, "started", "sessions"); got != r.Ended.Format("Jan 2") {
			t.Fatal("source Started must show Ended date", got, r)
		}
	}
	tab := &usageTableState{Sort: ui.SortOrder{Column: "cache", Descending: true}, Size: 10, Page: 1}
	result := queryUsageTable(rows, tab, "sessions", "tokens", "average")
	if result.Rows[0].ID != "read" {
		t.Fatal("table cache sorted writes instead of reads", result.Rows)
	}
	tab.Sort = ui.SortOrder{Column: "started", Descending: true}
	result = queryUsageTable(rows, tab, "sessions", "tokens", "average")
	if result.Rows[0].ID != "absent" {
		t.Fatal("Started sort not source Ended timestamp")
	}
	v, valid := usageMetricValue(usageSummaryAggregate(a.usageState.projection.Summary), "cache")
	if !valid || v != 70 {
		t.Fatal("table register changed chart read+write accounting", v, valid)
	}
	if a.usageCell(usageTableRow{usageAggregate: usageAggregate{Tokens: workbench.Tokens{CacheWrite: 50}}}, "cache", "breakdown:models") != "0" {
		t.Fatal("breakdown cache includes writes")
	}
	if a.usageCell(usageTableRow{}, "errors", "breakdown:tools") != "—" {
		t.Fatal("zero tool failures should be source dash")
	}
}
func TestUsageSourceTableTotalsScopeAndCostCoverage(t *testing.T) {
	a := fixtureUsageApp(t)
	for _, kind := range []string{"sessions", "series", "daily", "breakdown:models"} {
		tab := a.usageState.table(kind)
		tab.Search = "impossible query"
		result := queryUsageTable(a.usageTableRows(kind), tab, kind, "tokens", "average")
		filtered, page := a.usageTableTotals(kind, result)
		if page != "this page: 0 requests · 0 tokens" {
			t.Fatal(kind, page)
		}
		if kind == "sessions" || kind == "series" {
			if !strings.Contains(filtered, "14 requests") || !strings.Contains(filtered, "$1.26") {
				t.Fatal("snapshot total narrowed by table search", kind, filtered)
			}
		} else if filtered != "Filtered totals: 0 requests · 0 tokens" {
			t.Fatal("searched total did not match zero result or fabricated cost", kind, filtered)
		}
	}
	a.usageState.projection.Summary.PricedRequests = 0
	result := usageTableResult{Total: 3, Rows: make([]usageTableRow, 2), Totals: usageAggregate{Requests: 5, ToolFailures: 2}, PageTotals: usageAggregate{Requests: 3, ToolFailures: 1}}
	filtered, _ := a.usageTableTotals("sessions", result)
	if strings.Contains(filtered, "$") {
		t.Fatal("unknown price made visible", filtered)
	}
	filtered, page := a.usageTableTotals("breakdown:tools", result)
	if filtered != "Filtered totals: 5 calls · 2 failed" || page != "this page: 3 calls · 1 failed" {
		t.Fatal(filtered, page)
	}
	filtered, page = a.usageTableTotals("failures", result)
	if filtered != "Filtered totals: 3 events" || page != "this page: 2 events" {
		t.Fatal(filtered, page)
	}
}
func TestUsageSourceTableEmptyFrameAndFooter(t *testing.T) {
	a := fixtureUsageApp(t)
	a.usageState.projection.Sessions = a.usageState.projection.Sessions[:1]
	kind := "sessions"
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Padding(10).Gap(8).Children(func() { a.usageTable(c, kind) }) }, 1100, 650)
	tt.Frame()
	one, _ := tt.Find("Usage sessions table")
	if one.H != 82 {
		t.Fatal("one populated row must not retain 3-row blank space", one)
	}
	a.usageState.table(kind).Search = "impossible query"
	tt.Frame()
	table, _ := tt.Find("Usage sessions table")
	message, ok := tt.Find("No sessions match this search.")
	if !ok || table.H != 162 || message.Y < table.Y+40 || message.Y+message.H > table.Y+table.H {
		t.Fatal("no-match must sit inside three-row empty frame", table, message, ok)
	}
	if math.Abs(float64(message.X+message.W/2-table.X-table.W/2)) > 1 {
		t.Fatal("empty message not centered", table, message)
	}
	kind = "failures"
	tt.Frame()
	if !tt.HasText("Filtered totals: 1 events") || !tt.HasText("Retries are unavailable: pi records automatic retries as live events, not in session files.") {
		t.Fatal("failure totals or exact retry limitation missing")
	}
}
