package orbitui

import (
	"context"
	"fmt"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func usagePairedFixtureApp(t *testing.T) *App {
	t.Helper()
	a := fixtureUsageApp(t)
	root := filepath.Join("..", "..", "testdata", "usage-native", "home", ".pi", "agent", "sessions")
	index, err := workbench.NewUsageScanner(root).Scan(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s := &a.usageState
	s.index = index
	s.index.ScannedAt = time.Date(2026, 10, 25, 8, 50, 0, 0, time.UTC)
	s.fixtureNow, s.updated = s.index.ScannedAt, s.index.ScannedAt
	s.filter = workbench.UsageFilter{Start: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)}
	s.mode, s.preset, s.metric = "simple", "custom", "tokens"
	s.projection = buildUsageProjection(s.index, s.filter, s.granularity, time.UTC)
	s.requested = fmt.Sprintf("%d|%s|%+v", s.revision, s.granularity, s.filter)
	s.computed = s.requested
	return a
}
func TestUsageSourceSimpleNativeFixtureGeometry(t *testing.T) {
	a := usagePairedFixtureApp(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	find := func(label string) ui.Rect {
		t.Helper()
		r, ok := tt.Find(label)
		if !ok {
			t.Fatal("missing", label)
		}
		return r
	}
	for _, label := range []string{"Usage card Daily activity", "Usage calendar months", "Usage activity calendar", "Usage calendar legend", "Usage card Summary", "Usage summary metadata", "Usage card Usage over time", "Requests", "33", "55.4K", "View failures", "Usage mode segments", "Simple", "Details", "Tokens", "Errors", "All workspaces"} {
		t.Log(label, find(label))
	}
	daily, summary, trend := find("Usage card Daily activity"), find("Usage card Summary"), find("Usage card Usage over time")
	if daily.X != 295 || daily.W != 1140 || daily.Y != 188 {
		t.Fatal("source pane/content/filter origin", daily)
	}
	if summary.X != daily.X || math.Abs(float64(summary.Y-daily.Y-daily.H-20)) > .01 || trend.X != daily.X || math.Abs(float64(trend.Y-summary.Y-summary.H-20)) > .01 {
		t.Fatal("source section spacing", daily, summary, trend)
	}
	months, grid, legend := find("Usage calendar months"), find("Usage activity calendar"), find("Usage calendar legend")
	if months.H != 16 || grid.Y != months.Y+months.H || legend.Y != grid.Y+grid.H || math.Abs(float64(legend.H-(8+usageLineHeight(10)))) > .001 {
		t.Fatal("source month/grid/legend nesting", months, grid, legend)
	}
	requests := find("Requests")
	if requests.X != summary.X+1 || requests.Y != summary.Y+46 || requests.W < 280 {
		t.Fatal("full-bleed source KPI", requests, summary)
	}
	if math.Abs(float64(find("33").H-usageLineHeight(20))) > .001 || math.Abs(float64(find("55.4K").H-usageLineHeight(20))) > .001 {
		t.Fatal("source phi numeric line box")
	}
	if !tt.HasText("1.68K avg/request") || !tt.HasText("p95 9.0s · 33 measured") {
		t.Fatal("source Summary hint formatting")
	}
	metadata, failure := find("Usage summary metadata"), find("View failures")
	if failure.H != 18 || failure.Y < metadata.Y || failure.Y+failure.H > metadata.Y+metadata.H || math.Abs(float64(metadata.H-(2*usageLineHeight(12)+6))) > .001 {
		t.Fatal("inline metadata failure control", metadata, failure)
	}
	simple, details := find("Simple"), find("Details")
	if simple.H != 28 || details.H != 28 || details.X != simple.X+simple.W {
		t.Fatal("joined source mode control", simple, details)
	}
	if find("All workspaces").H != 28 {
		t.Fatal("source medium filter height")
	}
	if a.usageState.projection.Summary.Requests != 33 || a.usageState.projection.Summary.Tokens.Total != 55360 {
		t.Fatal("real fixture accounting drift")
	}
	if out := os.Getenv("ORBIT_CAPTURE_DIR"); out != "" {
		if err := os.MkdirAll(out, 0755); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(out, "usage-simple-33-request-cpu.png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err = png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
	}
}
func TestUsageSourceSegmentsModePersistenceAndUnavailable(t *testing.T) {
	a := fixtureUsageApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Children(func() {
			a.usageSegments(c, "mode", []string{"simple", "details"}, s.mode, false, func(x string) string { return titleCase(x) }, nil, a.setUsageMode)
		})
	}, 400, 100)
	s.scroll.X, s.scroll.Y = 11, 400
	if err := tt.Click("Simple"); err != nil {
		t.Fatal(err)
	}
	if s.scroll.X != 11 || s.scroll.Y != 400 {
		t.Fatal("active mode reset source scroll")
	}
	r, _ := tt.Find("Details")
	tt.Press(r.X+4, r.Y+4)
	if s.mode != "details" || s.scroll.X != 0 || s.scroll.Y != 0 {
		t.Fatal("mode not switched on pointer-down", s.mode, s.scroll)
	}
	s.scroll.Y = 35
	tt.Release(r.X+4, r.Y+4)
	if s.scroll.Y != 35 {
		t.Fatal("mouse-up repeated mode mutation")
	}
	var saved struct {
		Mode string `json:"usage_mode"`
	}
	if err := a.Options.Store.Read("usage.json", &saved); err != nil || saved.Mode != "details" {
		t.Fatal("mode persistence", saved, err)
	}
	s.metric = "tokens"
	for i := 0; i < 2; i++ {
		tt = ui.NewTester(func(c *ui.Context) {
			a.usageSegments(c, "metric", []string{"tokens", "cost"}, s.metric, true, func(x string) string { return titleCase(x) }, func(x string) bool { return x != "cost" }, func(x string) { s.metric = x })
		}, 300, 100)
		if err := tt.Click("Cost"); err != nil {
			t.Fatal(err)
		}
		if s.metric != "tokens" || !tt.HasText("Cost") {
			t.Fatal("missing metric hidden or selected")
		}
	}
}
func TestUsageSourceRaisedFillAndLineMetrics(t *testing.T) {
	for _, tc := range []struct{ size, line float32 }{{10, 16.18034}, {12, 19.416408}, {14, 22.652476}, {20, 32.36068}, {15, 24.27051}} {
		if math.Abs(float64(usageLineHeight(tc.size)-tc.line)) > .0001 {
			t.Fatal(tc)
		}
	}
	top, bottom := usageRaisedColors(ui.Hex("#282828"))
	if top != ui.Hex("#2d2d2d") || bottom != ui.Hex("#252525") {
		t.Fatal("source HSL lift/sink", top, bottom)
	}
	top, bottom = usageRaisedColors(ui.RGBA(100, 150, 200, .5))
	if top.A != 143 || bottom.A != 120 || top.R <= 100 || top.G <= 150 || top.B <= 200 || bottom.R >= 100 || bottom.G >= 150 || bottom.B >= 200 {
		t.Fatal("source translucent ramp", top, bottom)
	}
}
