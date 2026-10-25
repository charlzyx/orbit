package orbitui

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

// Frozen usage/chart.rs owns these dimensions, not the label font's measured
// width or MyGo's automatic tick spacing. All marks remain native elements.
const (
	usageChartHeight     float32 = 168
	usageChartAxisWidth  float32 = 56
	usageChartLabelStrip float32 = 18
	usageChartPlotHeight         = usageChartHeight - usageChartLabelStrip
	usageChartMaxLabels          = 8
	usageChartLabelWidth float32 = 64
)

// Follow chart.rs::point_value, including its actual average fallback for a
// sparse percentile bucket (the comment on LatencyMetric::value is stale).
// Source latency units are milliseconds, important for an all-zero 0..1 axis.
// Availability/readout semantics remain separate from the plotted baseline.
func usageChartValue(point usagePoint, metric, latency string) float64 {
	value, _ := usageMetricValue(point.usageAggregate, metric)
	if metric == "latency" {
		if measured, ok := point.Latency.value(latency); ok {
			value = measured
		}
		value *= 1000
	}
	// Scanner data is finite. Keep malformed synthetic/restored data from
	// emitting non-finite native path coordinates as well.
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func usageChartAxisLabel(value float64, metric string) string {
	switch metric {
	case "cost":
		return usageCost(value)
	case "latency":
		return usageDuration(value / 1000)
	default:
		// Rust float-to-u64 casts saturate, including the rounded float64
		// representation of MaxUint64; Go conversions do not guarantee that.
		if value >= float64(math.MaxUint64) {
			return compactCount(math.MaxUint64)
		}
		if math.IsNaN(value) || value <= 0 {
			return "0"
		}
		return compactCount(uint64(value))
	}
}

func usageChartQuarterTicks(maximum float64) []chart.Tick {
	ticks := make([]chart.Tick, 5)
	for i := range ticks {
		ticks[i] = chart.Tick{At: float32(i) / 4, Value: maximum * (float64(i) / 4)}
	}
	return ticks
}

func usageChartLabelIndices(count int) []int {
	step := max(1, (count+usageChartMaxLabels-1)/usageChartMaxLabels)
	var indices []int
	for i := step - 1; i < count; i += step {
		indices = append(indices, i)
	}
	return indices
}

func usageChartScale(maximum float64) chart.Linear {
	// chart.rs guards the denominator with f64::EPSILON, not the smallest
	// nonzero f64. Preserve the source behavior for tiny cost values.
	scale := chart.NewLinear(0, max(maximum, math.Nextafter(1, 2)-1))
	scale.Clamp = true
	return scale
}

// Native hit targets span the full 168px surface, including its label strip,
// like the source overlay. Marks and paths use only its top 150px plot.
func usageChartMarkerY(fraction float32) float32 {
	return 1 - (1-fraction)*usageChartPlotHeight/usageChartHeight
}

func (a *App) usageTimeline(c *ui.Context, points []usagePoint) {
	s, t := &a.usageState, a.Theme
	labels := make([]string, len(points))
	values := make([]float64, len(points))
	maximum := 0.0
	for i, point := range points {
		labels[i] = usageBucketLabel(point.At, s.granularity, s.now())
		values[i] = usageChartValue(point, s.metric, s.latencyMetric)
		maximum = max(maximum, values[i])
	}
	maximum = usageNiceMaximum(maximum)
	x := chart.NewBand(labels...).Padding(0, 0)
	y := usageChartScale(maximum)
	xs, ys, base := make([]float32, len(points)), make([]float32, len(points)), make([]float32, len(points))
	for i, value := range values {
		xs[i], ys[i] = x.At(i), y.At(value)
	}

	ui.Row(c.Key("usage-timeline")).Label("Usage timeline").Height(usageChartHeight).Shrink(0).AlignSelf(ui.Stretch).Children(func() {
		ui.Box(c).Label("Usage y axis").Width(usageChartAxisWidth).Height(usageChartHeight).Shrink(0).Children(func() {
			ui.Box(c).Height(usageChartPlotHeight).FillWidth().Children(func() {
				for i, tick := range usageChartQuarterTicks(maximum) {
					ui.Box(c).Label(fmt.Sprintf("Usage y tick %d", i)).Absolute().
						TopPercent((1-tick.At)*100).Right(8).Margin(-t.Px(10)*1.4/2, 0, 0, 0).Children(func() {
						ui.Text(c, usageChartAxisLabel(tick.Value, s.metric)).FontSize(t.Px(10)).
							FixedLineHeight(usageLineHeight(t.Px(10))).TextColor(t.TextMuted).NoWrap()
					})
				}
			})
		})
		area := ui.Box(c.Key("usage-trend")).Label("Usage trend").
			Description("Left and right arrow keys inspect buckets. Enter focuses the selected bucket.").
			Grow(1).MinWidth(0).Height(usageChartHeight)
		area.Draw(func(p *ui.Painter, r ui.Rect) {
			r.H = usageChartPlotHeight
			for quarter := 0; quarter < 4; quarter++ {
				yy := r.Y + r.H*float32(quarter)/4
				for xx := r.X; xx < r.X+r.W; xx += 6 {
					p.Fill(ui.Rect{X: xx, Y: yy, W: min(float32(4), r.X+r.W-xx), H: 1}, t.Border, 0)
				}
			}
			p.Fill(ui.Rect{X: r.X, Y: r.Y + r.H, W: r.W, H: 1}, t.Border, 0)
			p.FillPath(chart.Area(r, xs, base, ys, chart.CurveLinear), t.Accent.Alpha(.14))
			p.StrokePath(chart.Line(r, xs, ys, chart.CurveLinear), 1, t.Accent)
		})
		area.Children(func() {
			ui.Box(c).Label("Usage x axis").Absolute().Left(0).Right(0).Top(usageChartPlotHeight).Height(usageChartLabelStrip).Children(func() {
				for _, i := range usageChartLabelIndices(len(points)) {
					label := ui.Box(c).Label(fmt.Sprintf("Usage x tick %d", i)).Absolute().Top(4)
					switch {
					case i == 0:
						label.Left(0)
					case i+1 == len(points):
						label.Right(0)
					default:
						label.LeftPercent(xs[i]*100).Width(usageChartLabelWidth).Margin(0, 0, 0, -usageChartLabelWidth/2).TextAlign(ui.Center)
					}
					label.Children(func() {
						ui.Text(c, labels[i]).FontSize(t.Px(10)).FixedLineHeight(usageLineHeight(t.Px(10))).TextColor(t.TextMuted).NoWrap()
					})
				}
			})
			for i := range points {
				if s.filter.FocusStart.Equal(points[i].At) && s.filter.FocusEnd.Equal(points[i].End) {
					chart.Rect(c, x.Band(i), chart.Span{Lo: usageChartMarkerY(0), Hi: 1}).Label("Usage selected bucket").Background(t.Accent.Alpha(.10))
				}
				cell := chart.Rect(c, x.Band(i), chart.Span{Lo: 0, Hi: 1}).Tooltip(usagePointReadout(points[i], s.metric, s.latencyMetric)).Label("Usage bucket " + points[i].At.Format(time.RFC3339))
				if cell.Clicked() {
					a.focusUsageBucket(points[i].At, points[i].End)
				}
			}
			chart.Select(c, area, &s.selected, xs, func(i int) string {
				// A baseline or average fallback is plot geometry, not evidence
				// that a cost or the chosen latency percentile was measured.
				value, measured := usageMetricValue(points[i].usageAggregate, s.metric)
				if s.metric == "latency" {
					value, measured = points[i].Latency.value(s.latencyMetric)
				}
				return labels[i] + ": " + usageMetricText(value, measured, s.metric)
			})
			if i := s.selected; i >= 0 && i < len(points) {
				chart.RuleX(c, xs[i], 1).Bottom(usageChartLabelStrip).Background(t.TextMuted)
				chart.Point(c, xs[i], usageChartMarkerY(ys[i]), 5, 5).Label("Usage chart marker").Radius(3).Background(t.Accent)
				align := ui.Center
				if xs[i] < .2 {
					align = ui.Start
				} else if xs[i] > .8 {
					align = ui.End
				}
				chart.Anchor(c, xs[i], 1, align, ui.Start, func() {
					ui.Text(c, usagePointReadout(points[i], s.metric, s.latencyMetric)).FontSize(11).Padding(5, 8).Radius(5).Background(t.Surface).Border(1, t.Border).NoWrap()
				})
				if area.Shortcut(0, ui.KeyEnter) {
					a.focusUsageBucket(points[i].At, points[i].End)
				}
			}
		})
	})
}
