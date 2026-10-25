package orbitui

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func usageFixtureIndex() workbench.UsageIndex {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	index := workbench.UsageIndex{ScannedAt: now}
	for i := 0; i < 14; i++ {
		session := fmt.Sprintf("fixture-%d", i%4)
		model := "fixture-fast"
		provider := "fixture-provider"
		if i%3 == 0 {
			model = "fixture-large"
		}
		r := workbench.UsageRecord{At: now.AddDate(0, 0, -i), Session: session, SessionTitle: "Fixture task " + session, SessionPath: "/fixture/sessions/" + session + ".jsonl", Workspace: "/fixture/workspace", Provider: provider, Model: model, Tokens: workbench.Tokens{Input: uint64(1000 + i*130), Output: uint64(500 + i*50), CacheRead: uint64(3000 + i*200), Total: uint64(4500 + i*380)}, Cost: float64(i+1) * .012, HasCost: true, ModelPriced: true, HasDuration: true, Duration: time.Duration(i+2) * time.Second}
		index.Records = append(index.Records, r)
		index.Turns = append(index.Turns, workbench.UsageTurn{At: r.At, Session: r.Session, Workspace: r.Workspace, SessionPath: r.SessionPath, SessionTitle: r.SessionTitle})
		index.Tools = append(index.Tools, workbench.ToolRun{At: r.At, Session: r.Session, Workspace: r.Workspace, SessionPath: r.SessionPath, SessionTitle: r.SessionTitle, Provider: r.Provider, Model: r.Model, Name: "read", HasResult: true, Success: true})
	}
	index.Errors = []workbench.UsageError{{At: now, Session: "fixture-0", SessionTitle: "Fixture failed call", SessionPath: "/fixture/sessions/fixture-0.jsonl", Workspace: "/fixture/workspace", Provider: "fixture-provider", Model: "fixture-large", Kind: "provider", Message: "Explicit test fixture: provider unavailable"}}
	return index
}
func fixtureUsageApp(t *testing.T) *App {
	t.Helper()
	a := fixture(t)
	a.page = "usage"
	s := &a.usageState
	s.loaded = true
	s.index = usageFixtureIndex()
	s.preset = "all"
	s.filter = workbench.UsageFilter{}
	s.updated = time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	s.fixtureNow = s.updated
	s.revision = 1
	s.projection = buildUsageProjection(s.index, s.filter, s.granularity, time.UTC)
	s.requested = fmt.Sprintf("%d|%s|%+v", s.revision, s.granularity, s.filter)
	s.computed = s.requested
	return a
}
func TestUsageProjectionPreservesScopeCalendarAndMissingMetrics(t *testing.T) {
	index := usageFixtureIndex()
	from := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	index.Tools = append(index.Tools, workbench.ToolRun{At: from, Name: "pending", Provider: "fixture-provider", Model: "fixture-large"})
	p := buildUsageProjection(index, workbench.UsageFilter{Start: from, Provider: "fixture-provider"}, "day", time.UTC)
	if len(p.Series) != 2 || len(p.Calendar) != 14 {
		t.Fatalf("calendar must ignore range while trend obeys it: %d/%d", len(p.Series), len(p.Calendar))
	}
	if len(p.Errors) != 1 || p.Errors[0].Message == "" {
		t.Fatal("unbilled provider failure lost")
	}
	for _, g := range p.ToolGroups {
		if g.Name == "pending" && g.ToolFailures != 0 {
			t.Fatal("unfinished tool counted as failure")
		}
	}
	if v, ok := usageMetricValue(usageAggregate{}, "latency"); ok || v != 0 {
		t.Fatal("missing latency invented")
	}
	if _, ok := usageMetricValue(usageAggregate{}, "cost"); ok {
		t.Fatal("missing price invented")
	}
	if usageLatencyHint(0, 0) != "No timed responses" {
		t.Fatal("missing p95 invented")
	}
	for _, g := range p.Sessions {
		if !strings.HasPrefix(g.Name, "Fixture task") {
			t.Fatal("actual session title lost")
		}
	}
}
func TestUsageSortTiesAreStable(t *testing.T) {
	for _, desc := range []bool{false, true} {
		rows := []usageAggregate{{Key: "c", Requests: 2}, {Key: "a", Requests: 2}, {Key: "b", Requests: 1}}
		sortUsageRows(rows, "requests", desc)
		if desc {
			if rows[0].Key != "a" || rows[1].Key != "c" {
				t.Fatal(rows)
			}
		} else if rows[1].Key != "a" || rows[2].Key != "c" {
			t.Fatal(rows)
		}
	}
}
func TestUsageNativeLoadedEmptyErrorAndModal(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.SetDark(true)
	tt.Frame()
	if !tt.HasText("Daily activity") || !tt.HasText("Summary") {
		t.Fatalf("missing loaded native cards: %v", tt.Texts())
	}
	a.usageState.mode = "details"
	tt.Frame()
	if !tt.HasText("Breakdown") || !tt.HasText("Records") {
		t.Fatal("details missing")
	}
	a.usageState.detail = "failures"
	tt.Frame()
	if !strings.Contains(strings.Join(tt.Texts(), " "), "provider unavailable") {
		t.Fatal("unbilled failure hidden")
	}
	a.usageState.detailError = &a.usageState.projection.Errors[0]
	tt.Frame()
	if !a.modalOpen() {
		t.Fatal("usage detail didn't block native menus")
	}
	tt.Key(0, ui.KeyEscape)
	if a.usageState.detailError != nil {
		t.Fatal("failure detail Escape failed")
	}
	a.usageState.index = workbench.UsageIndex{}
	a.usageState.projection = buildUsageProjection(a.usageState.index, a.usageState.filter, "day", time.UTC)
	tt.Frame()
	if !tt.HasText("No usage in this range") {
		t.Fatal("empty data not honest")
	}
	a.usageState.loaded = false
	a.usageState.err = "Test fixture: permission denied"
	tt.Frame()
	if !tt.HasText("Usage unavailable") {
		t.Fatal("error disguised as zero usage")
	}
}
func TestUsageNativeMinimumWindowControls(t *testing.T) {
	a := fixtureUsageApp(t)
	tt := ui.NewTester(a.View, 800, 560)
	tt.Frame()
	for _, label := range []string{"All workspaces", "All providers", "All models", "Details", "Export", "Refresh"} {
		r, ok := tt.Find(label)
		if !ok || r.X < 0 || r.X+r.W > 800.5 || r.Y+r.H > 560.5 {
			t.Fatalf("%s clipped at minimum window: %+v", label, r)
		}
	}
	a.usageState.customOpen = true
	a.usageState.customStart = "2026-10-01"
	a.usageState.customEnd = "2026-10-25"
	tt.Frame()
	r, ok := tt.Find("Apply date range")
	if !ok || r.Y+r.H > 560.5 {
		t.Fatal("custom date apply unreachable")
	}
	tt.Key(0, ui.KeyEscape)
	if a.usageState.customOpen {
		t.Fatal("date Escape failed")
	}
}
func TestNativeUsageCapture(t *testing.T) {
	out := os.Getenv("ORBIT_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_CAPTURE_DIR for explicitly labelled fixture screenshots")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	for _, mode := range []string{"simple", "details"} {
		a := fixtureUsageApp(t)
		a.usageState.mode = mode
		tt := ui.NewTester(a.View, 1470, 922)
		tt.SetDark(true)
		tt.Frame()
		f, e := os.Create(filepath.Join(out, "usage-"+mode+"-fixture-dark.png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, tt.Image())
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
}
