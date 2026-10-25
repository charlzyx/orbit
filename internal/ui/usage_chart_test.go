package orbitui

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// Behavioral reference: 299a32b, crates/orbit-pi/src/usage/chart.rs, whose
// SHA-256 is 96efbbb05ecafdba4b2703d86a071af1307d184c1bc4e7a089a6706f1c4d9e13.
// These source-contract/layout tests are not native OS pixel-parity proof.
func TestUsageSourceChartQuarterTicksAndFiniteDomain(t *testing.T) {
	for _, tc := range []struct{ value, want float64 }{
		{0, 1}, {-1, 1}, {math.NaN(), 1}, {math.Inf(1), 1}, {math.Inf(-1), 1},
		{1, 1}, {1.1, 2}, {2.1, 2.5}, {2.5, 2.5}, {2.6, 5}, {5.1, 10},
		{842, 1000}, {2500, 2500}, {18_400_000, 20_000_000}, {.0042, .005},
	} {
		if got := usageNiceMaximum(tc.value); got != tc.want {
			t.Fatalf("nice maximum %g = %g, want %g", tc.value, got, tc.want)
		}
	}
	for _, value := range []float64{0, 1, 842, 2500, 10000, .0042, math.SmallestNonzeroFloat64, math.MaxFloat64} {
		maximum := usageNiceMaximum(value)
		if maximum <= 0 || math.IsInf(maximum, 0) || math.IsNaN(maximum) || maximum < value {
			t.Fatalf("non-finite/undersized domain for %g: %g", value, maximum)
		}
		ticks := usageChartQuarterTicks(maximum)
		if len(ticks) != 5 {
			t.Fatalf("%g produced %d ticks", maximum, len(ticks))
		}
		for i, tick := range ticks {
			if tick.At != float32(i)/4 || tick.Value != maximum*(float64(i)/4) || math.IsInf(tick.Value, 0) || math.IsNaN(tick.Value) {
				t.Fatalf("non-quarter/non-finite tick: %+v", tick)
			}
		}
	}
	var labels []string
	for _, tick := range usageChartQuarterTicks(10000) {
		labels = append(labels, usageChartAxisLabel(tick.Value, "tokens"))
	}
	if !reflect.DeepEqual(labels, []string{"0", "2.5K", "5K", "7.5K", "10K"}) {
		t.Fatal("automatic D3 ticks returned instead of source quarters", labels)
	}
	if got := usageChartAxisLabel(float64(math.MaxUint64), "tokens"); got != compactCount(math.MaxUint64) {
		t.Fatal("maximum unsigned token label did not saturate", got)
	}
}

func TestUsageSourceChartSparseLabels(t *testing.T) {
	for _, tc := range []struct {
		count int
		want  []int
	}{
		{0, nil}, {1, []int{0}}, {7, []int{0, 1, 2, 3, 4, 5, 6}},
		{8, []int{0, 1, 2, 3, 4, 5, 6, 7}}, {9, []int{1, 3, 5, 7}},
		{14, []int{1, 3, 5, 7, 9, 11, 13}}, {17, []int{2, 5, 8, 11, 14}},
		{70, []int{8, 17, 26, 35, 44, 53, 62}},
	} {
		if got := usageChartLabelIndices(tc.count); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%d labels: %v, want %v", tc.count, got, tc.want)
		}
	}
}

func TestUsageSourceChartMissingValuesAndLatencyFallback(t *testing.T) {
	for _, metric := range []string{"tokens", "requests", "input", "output", "cache", "cost", "latency", "errors"} {
		if got := usageChartValue(usagePoint{}, metric, "p95"); got != 0 {
			t.Fatalf("source missing %s point is the baseline, got %g", metric, got)
		}
	}
	point := usagePoint{usageAggregate: usageAggregate{Timed: 2, Duration: 4 * time.Second}, Latency: usageLatencyFrom([]time.Duration{time.Second, 3 * time.Second})}
	if _, measured := point.Latency.value("p95"); measured {
		t.Fatal("sparse p95 must not be reported as a measured percentile")
	}
	if got := usageChartValue(point, "latency", "p95"); got != 2000 {
		t.Fatal("source sparse p95 must plot average in milliseconds", got)
	}
	point.Latency = usageLatency{Samples: 20, Total: 40 * time.Second, P95: 3 * time.Second}
	if got := usageChartValue(point, "latency", "p95"); got != 3000 {
		t.Fatal("measured p95 lost", got)
	}
	if usageChartAxisLabel(1, "latency") != "1ms" || usageChartAxisLabel(1800, "latency") != "1.8s" {
		t.Fatal("latency unit or all-zero domain changed")
	}
	for _, malformed := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		point.Cost, point.Priced = malformed, 1
		if got := usageChartValue(point, "cost", "average"); got != 0 {
			t.Fatal("non-finite input reached native path", got)
		}
	}
	for _, maximum := range []float64{0, math.SmallestNonzeroFloat64, .001, 10000, math.MaxFloat64} {
		scale := usageChartScale(usageNiceMaximum(maximum))
		if got := scale.At(maximum); math.IsNaN(float64(got)) || math.IsInf(float64(got), 0) || got < 0 || got > 1 {
			t.Fatal("source scale emitted non-finite/out-of-range coordinate", maximum, got)
		}
	}
	tiny := usageChartScale(1e-20)
	if tiny.Hi != math.Nextafter(1, 2)-1 || tiny.At(1e-20) >= 1 {
		t.Fatal("source epsilon denominator not preserved", tiny)
	}
	point.Cost, point.Priced = -1, 1
	y := usageChartScale(1)
	if y.At(usageChartValue(point, "cost", "average")) != 0 {
		t.Fatal("source clamps negative plot values to the baseline")
	}
}

func TestUsageSourceChartNativeFixedGeometry(t *testing.T) {
	for _, tc := range []struct {
		width int
		font  float32
	}{{800, 14}, {1470, 14}, {800, 21}} {
		t.Run(fmt.Sprintf("width%d-font%g", tc.width, tc.font), func(t *testing.T) {
			a := usagePairedFixtureApp(t)
			a.Theme.FontSize = tc.font
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Padding(12).Children(func() { a.usageTimeline(c, a.usageState.projection.Series) })
			}, tc.width, 400)
			tt.Frame()
			find := func(label string) ui.Rect {
				t.Helper()
				r, ok := tt.Find(label)
				if !ok {
					t.Fatal("missing", label)
				}
				return r
			}
			outer, axis, area, strip := find("Usage timeline"), find("Usage y axis"), find("Usage trend"), find("Usage x axis")
			if axis.W != 56 || axis.H != 168 || outer.H != 168 || area.X != axis.X+56 || area.H != 168 || area.W != outer.W-56 || strip.Y != area.Y+150 || strip.H != 18 {
				t.Fatalf("source fixed geometry lost: outer%+v axis%+v area%+v strip%+v", outer, axis, area, strip)
			}
			for i := 0; i <= 4; i++ {
				tick := find(fmt.Sprintf("Usage y tick %d", i))
				wantY := area.Y + 150*(1-float32(i)/4) - a.Theme.Px(10)*1.4/2
				if math.Abs(float64(tick.Y-wantY)) > .001 || math.Abs(float64(tick.X+tick.W-(area.X-8))) > .001 {
					t.Fatalf("source quarter tick %d misaligned %+v vs y%g", i, tick, wantY)
				}
			}
			for _, i := range usageChartLabelIndices(len(a.usageState.projection.Series)) {
				tick := find(fmt.Sprintf("Usage x tick %d", i))
				if tick.Y != strip.Y+4 {
					t.Fatal("x-axis strip padding", tick, strip)
				}
				if i == 13 {
					if math.Abs(float64(tick.X+tick.W-(area.X+area.W))) > .001 {
						t.Fatal("last label must hug right edge", tick, area)
					}
				} else if math.Abs(float64(tick.W-64)) > .001 || math.Abs(float64(tick.X+32-(area.X+(float32(i)+.5)/14*area.W))) > .001 {
					t.Fatal("interior source label center/width", i, tick, area)
				}
			}
			if _, ok := tt.Find("Usage x tick 0"); ok {
				t.Fatal("first label erroneously included in sparse axis")
			}
		})
	}
}

func TestUsageSourceChartNativeMarksKeyboardAndScope(t *testing.T) {
	a := usagePairedFixtureApp(t)
	s := &a.usageState
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(12).Children(func() { a.usageTimeline(c, s.projection.Series) })
	}, 1100, 400)
	tt.Frame()
	area, _ := tt.Find("Usage trend")
	for i, point := range s.projection.Series {
		cell, ok := tt.Find("Usage bucket " + point.At.Format(time.RFC3339))
		if !ok || math.Abs(float64(cell.W-area.W/14)) > .001 || cell.H != 168 || math.Abs(float64(cell.X-(area.X+float32(i)/14*area.W))) > .001 {
			t.Fatalf("native bucket hit target %d: %+v", i, cell)
		}
		// The label strip remains part of the source's bucket hit target.
		tt.Move(cell.X+cell.W/2, cell.Y+160)
		tt.Frame()
		if s.selected != i {
			t.Fatalf("hover selected %d, want %d", s.selected, i)
		}
	}
	last := s.projection.Series[len(s.projection.Series)-1]
	lastCell, _ := tt.Find("Usage bucket " + last.At.Format(time.RFC3339))
	tt.ClickAt(lastCell.X+lastCell.W/2, lastCell.Y+160)
	if !s.filter.FocusStart.Equal(last.At) {
		t.Fatal("clicking label strip did not focus its native bucket")
	}
	tt.ClickAt(lastCell.X+lastCell.W/2, lastCell.Y+160)
	if !s.filter.FocusStart.IsZero() {
		t.Fatal("repeated bucket click did not clear scope")
	}
	tt.Move(2, 2)
	tt.Frame()
	if s.selected != -1 {
		t.Fatal("leaving native chart retained hover")
	}
	if err := tt.Click("Usage trend"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyHome)
	if s.selected != 0 {
		t.Fatal("Home did not select first bucket")
	}
	tt.Key(0, ui.KeyRight)
	if s.selected != 1 {
		t.Fatal("Right did not advance")
	}
	tt.Key(0, ui.KeyEnter)
	if !s.filter.FocusStart.Equal(s.projection.Series[1].At) {
		t.Fatal("Enter failed to scope exact bucket")
	}
	tt.Frame()
	band, ok := tt.Find("Usage selected bucket")
	if !ok || math.Abs(float64(band.H-150)) > .001 {
		t.Fatal("selection band must exclude label strip", band)
	}
	tt.Key(0, ui.KeyEnter)
	if !s.filter.FocusStart.IsZero() {
		t.Fatal("repeated Enter did not clear scope")
	}
	tt.Key(0, ui.KeyEnd)
	if s.selected != 13 {
		t.Fatal("End did not select final bucket")
	}
	tt.Key(0, ui.KeyLeft)
	if s.selected != 12 {
		t.Fatal("Left did not move back")
	}
	tt.Key(0, ui.KeyEscape)
	if s.selected != -1 {
		t.Fatal("Escape did not dismiss readout")
	}

	// A zero/unmeasured point still has a finite baseline marker and remains
	// operable; it is not silently dropped by NaN-based path segmentation.
	s.metric = "latency"
	s.projection.Series = []usagePoint{{At: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC), usageAggregate: usageAggregate{Tokens: workbench.Tokens{}}}}
	tt.Frame()
	tt.Key(0, ui.KeyHome)
	tt.Frame()
	firstTick, ok := tt.Find("Usage x tick 0")
	if !ok || firstTick.X != area.X {
		t.Fatal("single-bucket first label must hug left edge", firstTick, area)
	}
	if !tt.HasText("1ms") {
		t.Fatal("all-zero source latency axis must top out at 1ms")
	}
	marker, ok := tt.Find("Usage chart marker")
	if !ok || math.Abs(float64(marker.Y+marker.H/2-(area.Y+150))) > .001 {
		t.Fatal("zero point marker not on baseline", marker, area)
	}
}

func TestUsageSourceChartNativeAnnouncementsRespectMeasurements(t *testing.T) {
	for _, tc := range []struct {
		name, metric, latency, want string
		point                       usagePoint
		plotted                     float64
	}{
		{name: "missing cost", metric: "cost", want: "Not measured"},
		{name: "measured cost", metric: "cost", point: usagePoint{usageAggregate: usageAggregate{Cost: .25, Priced: 1}}, plotted: .25, want: "$0.25"},
		{name: "measured zero cost", metric: "cost", point: usagePoint{usageAggregate: usageAggregate{Priced: 1}}, want: "$0.00"},
		{name: "missing latency", metric: "latency", latency: "average", want: "Not measured"},
		{name: "sparse p95", metric: "latency", latency: "p95", point: usagePoint{
			usageAggregate: usageAggregate{Timed: 2, Duration: 4 * time.Second},
			Latency:        usageLatencyFrom([]time.Duration{time.Second, 3 * time.Second}),
		}, plotted: 2000, want: "Not measured"},
		{name: "measured p95", metric: "latency", latency: "p95", point: usagePoint{Latency: usageLatency{Samples: 20, P95: 3 * time.Second}}, plotted: 3000, want: "3.0s"},
		{name: "measured zero p95", metric: "latency", latency: "p95", point: usagePoint{Latency: usageLatency{Samples: 20}}, want: "0ms"},
		{name: "measured zero average", metric: "latency", latency: "average", point: usagePoint{Latency: usageLatency{Samples: 1}}, want: "0ms"},
		{name: "zero tokens", metric: "tokens", want: "0 tokens"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := usagePairedFixtureApp(t)
			s := &a.usageState
			s.metric, s.latencyMetric, s.selected = tc.metric, tc.latency, -1
			point := tc.point
			point.At = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
			point.End = point.At.AddDate(0, 0, 1)
			s.projection.Series = []usagePoint{point}
			if got := usageChartValue(point, tc.metric, tc.latency); got != tc.plotted {
				t.Fatalf("announcement fix changed source plotting: got %g, want %g", got, tc.plotted)
			}
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Column(c).Fill().Padding(12).Children(func() { a.usageTimeline(c, s.projection.Series) })
			}, 800, 300)
			tt.Frame()
			tt.Key(0, ui.KeyTab)
			if !tt.Focused("Usage trend") {
				t.Fatal("native chart did not receive keyboard focus")
			}
			tt.Announcements()
			tt.Key(0, ui.KeyRight)
			want := []string{"Oct 1: " + tc.want}
			if got := tt.Announcements(); !reflect.DeepEqual(got, want) {
				t.Fatalf("mounted native chart announced %q, want truthful measurement %q", got, want)
			}
			// Re-inspection after dismissal must use the same availability,
			// rather than retaining a prior synthesized numeric announcement.
			tt.Key(0, ui.KeyEscape)
			tt.Key(0, ui.KeyHome)
			if got := tt.Announcements(); !reflect.DeepEqual(got, want) {
				t.Fatalf("repeated mounted announcement %q, want %q", got, want)
			}
		})
	}
}
