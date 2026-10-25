package orbitui

import (
	"fmt"
	orbitassets "github.com/charlzyx/orbit"
	"math"
	"path/filepath"
	"sort"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

type usageLatency struct {
	Samples                   int
	Total, P50, P95, P99, Max time.Duration
}

func usageLatencyFrom(samples []time.Duration) usageLatency {
	l := usageLatency{Samples: len(samples)}
	if len(samples) == 0 {
		return l
	}
	samples = append([]time.Duration(nil), samples...)
	sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
	for _, v := range samples {
		l.Total += v
	}
	pick := func(q float64) time.Duration {
		return samples[max(0, min(len(samples)-1, int(math.Ceil(q*float64(len(samples))))-1))]
	}
	l.P50, l.P95, l.P99, l.Max = pick(.5), pick(.95), pick(.99), samples[len(samples)-1]
	return l
}
func (l usageLatency) value(metric string) (float64, bool) {
	if l.Samples == 0 {
		return 0, false
	}
	if metric == "average" || metric == "" {
		return l.Total.Seconds() / float64(l.Samples), true
	}
	if l.Samples < 20 {
		return 0, false
	}
	switch metric {
	case "p50":
		return l.P50.Seconds(), true
	case "p95":
		return l.P95.Seconds(), true
	case "p99":
		return l.P99.Seconds(), true
	}
	return 0, false
}

type usageInsight struct {
	Key, Text, Tone string
	Variables       map[string]string
}

func usageSummaryAggregate(s workbench.UsageSummary) usageAggregate {
	return usageAggregate{Requests: s.Requests, Errors: s.Errors, Aborted: s.Aborted, Timed: s.TimedRequests, Priced: s.PricedRequests, Tokens: s.Tokens, Cost: s.Cost, Duration: s.Duration}
}
func usageDelta(current, previous float64) (float64, bool) {
	if math.Abs(previous) < math.SmallestNonzeroFloat64 {
		return 0, false
	}
	return (current - previous) / previous * 100, true
}
func usageCacheRate(t workbench.Tokens) (float64, bool) {
	if t.CacheRead == 0 && t.CacheWrite == 0 {
		return 0, false
	}
	return t.CacheRate(), true
}
func usageShare(row usageAggregate, total workbench.UsageSummary) float64 {
	if total.Tokens.Total > 0 {
		return float64(row.Tokens.Total) / float64(total.Tokens.Total)
	}
	if total.Requests > 0 {
		return float64(row.Requests) / float64(total.Requests)
	}
	return 0
}
func usageInsights(p usageProjection, all bool) []usageInsight {
	var out []usageInsight
	if p.Summary.Requests == 0 {
		return out
	}
	add := func(key, text, tone string, variables ...string) {
		vars := map[string]string{}
		for i := 0; i+1 < len(variables); i += 2 {
			vars[variables[i]] = variables[i+1]
		}
		out = append(out, usageInsight{Key: key, Text: text, Tone: tone, Variables: vars})
	}
	if p.Previous != nil {
		if delta, ok := usageDelta(float64(p.Summary.Tokens.Total), float64(p.Previous.Tokens.Total)); ok && math.Abs(delta) >= 10 {
			verb := "up"
			if delta < 0 {
				verb = "down"
			}
			add("usage.insight_token_delta", fmt.Sprintf("Token usage is %s %.0f%% versus the previous period (%s vs %s).", verb, math.Abs(delta), compactCount(p.Summary.Tokens.Total), compactCount(p.Previous.Tokens.Total)), "neutral", "verb", verb, "pct", fmt.Sprintf("%.0f", math.Abs(delta)), "now", compactCount(p.Summary.Tokens.Total), "before", compactCount(p.Previous.Tokens.Total))
		}
	}
	if p.Summary.Requests >= 20 && len(p.Models) > 1 {
		top := p.Models[0]
		share := usageShare(top, p.Summary)
		if share >= .2 {
			add("usage.insight_top_model", fmt.Sprintf("%s accounts for %.0f%% of token usage (%s across %d requests).", top.Model, share*100, compactCount(top.Tokens.Total), top.Requests), "neutral", "model", top.Model, "share", fmt.Sprintf("%.0f", share*100), "tokens", compactCount(top.Tokens.Total), "requests", fmt.Sprint(top.Requests))
		}
	}
	if p.Previous != nil {
		now, nok := usageCacheRate(p.Summary.Tokens)
		before, bok := usageCacheRate(p.Previous.Tokens)
		if nok && bok && math.Abs(now-before) >= 3 {
			verb, tone := "improved", "positive"
			if now < before {
				verb, tone = "fell", "warning"
			}
			add("usage.insight_cache_hit", fmt.Sprintf("Cache hit rate %s from %.0f%% to %.0f%%.", verb, before, now), tone, "verb", verb, "before", fmt.Sprintf("%.0f", before), "now", fmt.Sprintf("%.0f", now))
		}
	}
	if len(p.Workspaces) > 1 {
		top := p.Workspaces[0]
		share := usageShare(top, p.Summary)
		if share >= .2 {
			add("usage.insight_top_workspace", fmt.Sprintf("%s generated the most usage: %s tokens (%.0f%% of the period).", filepath.Base(top.Workspace), compactCount(top.Tokens.Total), share*100), "neutral", "workspace", filepath.Base(top.Workspace), "tokens", compactCount(top.Tokens.Total), "share", fmt.Sprintf("%.0f", share*100))
		}
	}
	var mean float64
	var peak usagePoint
	nonzero := 0
	for _, point := range p.Daily {
		if point.Tokens.Total > 0 {
			nonzero++
			mean += float64(point.Tokens.Total)
			if point.Tokens.Total >= peak.Tokens.Total {
				peak = point
			}
		}
	}
	if nonzero >= 5 {
		mean /= float64(nonzero)
		if float64(peak.Tokens.Total) >= mean*2 {
			add("usage.insight_peak_bucket", fmt.Sprintf("%s is %.1f× the average bucket volume.", peak.At.Format("Jan 2"), float64(peak.Tokens.Total)/mean), "warning", "bucket", peak.At.Format("Jan 2"), "factor", fmt.Sprintf("%.1f", float64(peak.Tokens.Total)/mean), "metric", map[bool]string{true: "day", false: p.Granularity}[p.Granularity == "hour"])
		}
	}
	if p.Summary.Errors > 0 {
		rate := float64(p.Summary.Errors) / float64(p.Summary.Requests) * 100
		if rate >= 1 {
			add("usage.insight_failures", fmt.Sprintf("%d of %d requests failed (%.1f%%).", p.Summary.Errors, p.Summary.Requests, rate), "warning", "failed", fmt.Sprint(p.Summary.Errors), "total", fmt.Sprint(p.Summary.Requests), "rate", fmt.Sprintf("%.1f", rate))
		}
	}
	if p.Latency.Samples >= 20 && p.Previous != nil && p.Previous.TimedRequests >= 20 {
		now, _ := p.Latency.value("average")
		before := p.Previous.Duration.Seconds() / float64(p.Previous.TimedRequests)
		if delta, ok := usageDelta(now, before); ok && math.Abs(delta) >= 10 {
			add("usage.insight_latency", fmt.Sprintf("Average response time moved from %.2fs to %.2fs (%d requests).", before, now, p.Latency.Samples), "neutral", "before", fmt.Sprintf("%.2fs", before), "now", fmt.Sprintf("%.2fs", now), "requests", fmt.Sprint(p.Latency.Samples))
		}
	}
	if all && len(p.Sessions) > 0 {
		add("usage.insight_all_time", fmt.Sprintf("%d sessions and %d requests recorded across all time.", len(p.Sessions), p.Summary.Requests), "neutral", "sessions", fmt.Sprint(len(p.Sessions)), "requests", fmt.Sprint(p.Summary.Requests))
	}
	return out[:min(4, len(out))]
}
func (a *App) usageSignals(c *ui.Context) {
	s := &a.usageState
	insights := usageInsights(s.projection, s.preset == "all")
	if len(insights) == 0 {
		return
	}
	a.usageCard(c, a.text("usage.signals", "Signals"), fmt.Sprintf("%d notes", len(insights)), func() {
		for _, insight := range insights {
			color := a.Theme.TextMuted
			if insight.Tone == "warning" {
				color = a.Theme.Warning
			}
			if insight.Tone == "positive" {
				color = a.Theme.Accent
			}
			ui.Text(c, a.usageInsightText(insight)).TextColor(color)
		}
	})
}
func (a *App) usageLatencyLine(c *ui.Context, l usageLatency, label string) {
	ui.Row(c.Key("usage-latency-readout:" + label)).Wrap().Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Text(c, label).NoWrap().FontWeight(500).FontSize(11).TextColor(a.Theme.TextMuted)
		if l.Samples == 0 {
			ui.Text(c, a.text("usage.unavailable", "Unavailable")).FontSize(11).TextColor(a.Theme.TextMuted)
			return
		}
		avg, _ := l.value("average")
		items := []struct{ label, value string }{{"avg", usageDuration(avg)}}
		if l.Samples >= 20 {
			for _, item := range []struct {
				label string
				value time.Duration
			}{{"p50", l.P50}, {"p95", l.P95}, {"p99", l.P99}} {
				items = append(items, struct{ label, value string }{item.label, usageDuration(item.value.Seconds())})
			}
		}
		items = append(items, struct{ label, value string }{"max", usageDuration(l.Max.Seconds())})
		for _, item := range items {
			ui.Row(c).Gap(4).Children(func() {
				ui.Text(c, item.label).TextColor(a.Theme.TextMuted).FontSize(11)
				ui.Text(c, item.value).Font(a.Fonts.CodeFamily).FontSize(11)
			})
		}
		ui.Text(c, usageLatencyHint(l.P95, l.Samples)).TextColor(a.Theme.TextMuted).FontSize(10)
	})
}
func (a *App) usageHealthDetails(c *ui.Context) {
	p := a.usageState.projection
	a.usageCard(c, a.text("usage.token_health", "Token health"), a.text("usage.composition_and_cache", "Composition and cache"), func() {
		width, _ := c.Size()
		if a.SidebarVisible {
			width -= a.SidebarWidth
		}
		width = min(width, 1180)
		panel := func(title string, body func()) {
			if width >= 820 {
				ui.Column(c).Grow(1).MinWidth(0).Children(func() { a.usageWell(c, title, body) })
			} else {
				a.usageWell(c, title, body)
			}
		}
		panels := func() {
			panel(a.text("usage.token_composition", "Token composition"), func() { a.usageComposition(c, p) })
			panel(a.text("usage.cache_performance", "Cache performance"), func() { a.usageCacheBody(c, p) })
		}
		if width >= 820 {
			ui.Row(c).AlignItems(ui.Start).Gap(16).Children(panels)
		} else {
			ui.Column(c).AlignSelf(ui.Stretch).Gap(16).Children(panels)
		}
		readable := 0
		for _, point := range p.Series {
			if point.Tokens.Input > 0 || point.Tokens.CacheRead > 0 {
				readable++
			}
		}
		if readable >= 2 {
			a.usageWell(c, a.text("view.hit_rate_over_time", "Hit rate over time"), func() {
				ui.Row(c).Height(56).Gap(2).Children(func() {
					for _, point := range p.Series {
						rate, ok := usageCacheRate(point.Tokens)
						height := float32(3)
						color := a.Theme.TextMuted.Alpha(.2)
						tip := point.At.Format("Jan 2 15:04") + " · " + a.text("usage.unavailable", "Unavailable")
						if ok {
							height = max(3, float32(rate)/100*56)
							color = a.Theme.Accent.Alpha(.65)
							tip = point.At.Format("Jan 2 15:04") + " · " + usagePercent(rate)
						}
						ui.Box(c).Grow(1).Height(height).AlignSelf(ui.End).Background(color).Tooltip(tip)
					}
				})
			})
		}
	})
}
func (a *App) usageWell(c *ui.Context, title string, body func()) {
	ui.Column(c).MinWidth(0).AlignSelf(ui.Stretch).Padding(12).Gap(12).Background(a.Theme.Background).Radius(8).Children(func() {
		ui.Text(c, title).FontSize(12).FontWeight(500)
		body()
	})
}
func (a *App) usageComposition(c *ui.Context, p usageProjection) {
	total := p.Summary.Tokens.Total
	for _, part := range []struct {
		key, label string
		value      uint64
	}{{"input", "Input", p.Summary.Tokens.Input}, {"output", "Output", p.Summary.Tokens.Output}, {"cache_read", "Cache read", p.Summary.Tokens.CacheRead}, {"cache_write", "Cache write", p.Summary.Tokens.CacheWrite}} {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, a.text("usage.composition_"+part.key, part.label)).Width(80).FontSize(11)
			ui.Box(c).Grow(1).Height(6).Radius(3).Background(a.Theme.Surface).Children(func() {
				if total > 0 {
					ui.Box(c).Height(6).WidthPercent(min(100, float32(part.value)*100/float32(total))).Radius(3).Background(a.Theme.Accent)
				}
			})
			ui.Text(c, compactCount(part.value)).Width(58).Font(a.Fonts.CodeFamily).FontSize(11).Tooltip(usageExact(part.value))
		})
	}
	ui.Text(c, a.text("usage.composition_formula", "Total = input + output + cache read + cache write. Reasoning tokens are a subset of output.")).FontSize(10).TextColor(a.Theme.TextMuted)
}
func (a *App) usageCacheBody(c *ui.Context, p usageProjection) {
	s := &a.usageState
	rate, ok := usageCacheRate(p.Summary.Tokens)
	if !ok {
		ui.Text(c, a.text("usage.cache_unavailable_range", "Cache data unavailable: no cache tokens were reported in this range.")).FontSize(11).TextColor(a.Theme.TextMuted)
		return
	}
	ui.Row(c).Wrap().Gap(8).AlignItems(ui.Center).Children(func() {
		ui.Text(c, usagePercent(rate)).Font(a.Fonts.CodeFamily).FontSize(20)
		ui.Text(c, a.text("view.of_prompt_tokens_served_from_cache", "of prompt tokens served from cache")).FontSize(11).TextColor(a.Theme.TextMuted)
	})
	ui.Box(c).AlignSelf(ui.Stretch).Height(6).Radius(3).Background(a.Theme.Surface).Children(func() { ui.Box(c).Height(6).WidthPercent(float32(rate)).Radius(3).Background(a.Theme.Accent) })
	for _, item := range []struct{ key, label, value string }{{"cache_reads", "Cache reads", compactCount(p.Summary.Tokens.CacheRead)}, {"cache_writes", "Cache writes", compactCount(p.Summary.Tokens.CacheWrite)}, {"uncached_input", "Uncached input", compactCount(p.Summary.Tokens.Input)}, {"requests_served_from_cache", "Requests served from cache", fmt.Sprintf("%s of %s", usageCount(uint64(p.CachedRequests)), usageCount(uint64(p.Summary.Requests)))}} {
		ui.Row(c).Gap(8).Children(func() {
			ui.Text(c, a.text("usage."+item.key, item.label)).Grow(1).FontSize(11)
			ui.Text(c, item.value).Font(a.Fonts.CodeFamily).FontSize(11)
		})
	}
	ui.Text(c, a.text("usage.hit_rate_formula", "Hit rate = cache read / (input + cache read + cache write)")).FontSize(10).TextColor(a.Theme.TextMuted)
	key, label := "usage.show_cached_only", "Show cached only"
	if s.filter.CachedOnly {
		key, label = "usage.show_all_requests", "Show all requests"
	}
	if a.button(c, a.text(key, label), "", false).Clicked() {
		s.filter.CachedOnly = !s.filter.CachedOnly
		if s.filter.CachedOnly {
			s.filter.ErrorsOnly = false
		}
	}
}

func usageInsightTexts(insights []usageInsight) []string {
	out := make([]string, len(insights))
	for i, v := range insights {
		out[i] = v.Text
	}
	return out
}

func usageComparison(p usageProjection, metric string) string {
	if p.Previous == nil {
		return ""
	}
	now, nok := usageMetricValue(usageSummaryAggregate(p.Summary), metric)
	before, bok := usageMetricValue(usageSummaryAggregate(*p.Previous), metric)
	if !nok || !bok {
		return ""
	}
	delta, ok := usageDelta(now, before)
	if !ok {
		if now > before {
			return "Up from zero in previous period"
		}
		return "No change from previous period"
	}
	return fmt.Sprintf("%+.1f%% vs previous period", delta)
}

func usageHeatmapThresholds(values []float64) [3]float64 {
	active := []float64{}
	for _, v := range values {
		if v > 0 && !math.IsNaN(v) {
			active = append(active, v)
		}
	}
	if len(active) == 0 {
		return [3]float64{}
	}
	sort.Float64s(active)
	pick := func(q float64) float64 {
		return active[max(0, min(len(active)-1, int(math.Ceil(q*float64(len(active))))-1))]
	}
	return [3]float64{pick(.25), pick(.5), pick(.75)}
}
func usageHeatmapLevel(v float64, thresholds [3]float64, maximum float64) int {
	if v <= 0 || math.IsNaN(v) {
		return 0
	}
	if v >= maximum {
		return 4
	}
	for i, threshold := range thresholds {
		if v <= threshold {
			return i + 1
		}
	}
	return 4
}

func (a *App) usageInsightText(insight usageInsight) string {
	vars := map[string]string{}
	for k, v := range insight.Variables {
		vars[k] = v
	}
	if verb := vars["verb"]; verb != "" {
		vars["verb"] = a.text("usage.insight_"+verb, verb)
	}
	if metric := vars["metric"]; metric != "" {
		vars["metric"] = a.text("usage.granularity_"+metric, metric)
	}
	text := orbitassets.Translate(a.catalog, a.english, insight.Key, vars)
	if text == insight.Key {
		return insight.Text
	}
	return text
}
func (a *App) usageInsightTexts(insights []usageInsight) []string {
	out := make([]string, len(insights))
	for i, v := range insights {
		out[i] = a.usageInsightText(v)
	}
	return out
}
func (a *App) usageSummaryDetails(c *ui.Context, p usageProjection) {
	s := p.Summary
	cost := a.text("usage.unavailable", "Unavailable")
	if s.PricedRequests > 0 {
		cost = usageCost(s.Cost) + " · " + usageCost(s.Cost/float64(s.PricedRequests)) + " / req"
	}
	failed := a.text("usage.none", "None")
	if s.Errors > 0 {
		failed = usageExact(uint64(s.Errors))
	}
	stats := []struct{ key, label, value string }{{"input_tokens", "Input tokens", compactCount(s.Tokens.Input)}, {"output_tokens", "Output tokens", compactCount(s.Tokens.Output)}, {"cost", "Cost", cost}, {"failed", "Failed", failed}, {"sessions", "Sessions", fmt.Sprint(len(p.Sessions))}, {"turns", "Turns", fmt.Sprint(s.Turns)}, {"tool_calls", "Tool calls", fmt.Sprint(s.ToolRuns)}, {"bash_commands", "Bash commands", fmt.Sprint(s.BashRuns)}}
	if p.ToolFailures > 0 {
		stats = append(stats, struct{ key, label, value string }{"tool_failures", "Tool failures", fmt.Sprint(p.ToolFailures)})
	}
	if s.TimedRequests > 0 {
		stats = append(stats, struct{ key, label, value string }{"generation_time", "Generation time", usageSpan(s.Duration)})
	}
	if s.Requests > 0 {
		stats = append(stats, struct{ key, label, value string }{"avg_prompt", "Avg prompt", compactCount(usageAdd(usageAdd(s.Tokens.Input, s.Tokens.CacheRead), s.Tokens.CacheWrite) / uint64(s.Requests))})
	}
	if s.Tokens.Input > 0 {
		ratio := float64(s.Tokens.Output) / float64(s.Tokens.Input)
		text := fmt.Sprintf("%.3f×", ratio)
		if ratio > 0 && ratio < .001 {
			text = "<0.001×"
		}
		stats = append(stats, struct{ key, label, value string }{"output_input", "Output / input", text})
	}
	if s.Tokens.Total > 0 && s.PricedRequests > 0 {
		stats = append(stats, struct{ key, label, value string }{"cost_per_mtok", "Cost / M tokens", usageCost(s.Cost * 1e6 / float64(s.Tokens.Total))})
	}
	if p.PeakPrompt > 0 {
		stats = append(stats, struct{ key, label, value string }{"peak_prompt", "Peak prompt", compactCount(p.PeakPrompt)})
	}
	if p.ReasoningReported {
		stats = append(stats, struct{ key, label, value string }{"reasoning", "Reasoning", compactCount(s.Tokens.Reasoning) + " of output"})
	}
	ui.Row(c).Wrap().Gap(16).Children(func() {
		for _, item := range stats {
			ui.Row(c).Gap(6).Children(func() {
				ui.Text(c, a.text("usage.stat_"+item.key, item.label)).TextColor(a.Theme.TextMuted).FontSize(11)
				ui.Text(c, item.value).Font(a.Fonts.CodeFamily).FontSize(11)
			})
		}
	})
}

func (a *App) usageSummary(c *ui.Context) {
	s := &a.usageState
	p := s.projection
	t := a.Theme
	a.usageCard(c, a.text("usage.summary", "Summary"), fmt.Sprintf("%s requests · %s sessions", usageCount(uint64(p.Summary.Requests)), usageCount(uint64(len(p.Sessions)))), func() {
		unavailable := a.text("usage.unavailable", "Unavailable")
		avg, avgHint := unavailable, a.text("usage.sub_no_request_gaps", "No request timing available")
		if p.Summary.TimedRequests > 0 {
			avg = usageDuration(p.Summary.Duration.Seconds() / float64(p.Summary.TimedRequests))
			avgHint = usageLatencyHint(p.P95, p.Timed)
		}
		cache, cacheHint := unavailable, a.text("usage.sub_no_cache_tokens", "No cache tokens reported")
		if rate, ok := usageCacheRate(p.Summary.Tokens); ok {
			cache = usagePercent(rate)
			cacheHint = compactCount(p.Summary.Tokens.CacheRead) + " read · " + compactCount(p.Summary.Tokens.CacheWrite) + " written"
		}
		tokenHint := a.text("usage.sub_no_requests", "No requests")
		if p.Summary.Requests > 0 {
			tokenHint = compactCount(p.Summary.Tokens.Total/uint64(p.Summary.Requests)) + " avg / request"
		}
		requestHint := usageComparison(p, "requests")
		if requestHint == "" {
			requestHint = a.text("usage.sub_requests_in_range", "Model calls in range")
		}
		stats := []struct {
			metric, key, label, value, hint string
			available                       bool
		}{
			{"requests", "requests", "Requests", usageCount(uint64(p.Summary.Requests)), requestHint, true},
			{"tokens", "total_tokens", "Total tokens", compactCount(p.Summary.Tokens.Total), tokenHint, true},
			{"cache", "cache_hit_rate", "Cache hit rate", cache, cacheHint, cache != unavailable},
			{"latency", "avg_response", "Avg response", avg, avgHint, avg != unavailable},
		}
		width, _ := c.Size()
		if a.SidebarVisible {
			width -= a.SidebarWidth
		}
		cols := 4
		if width < 880 {
			cols = 2
		}
		if width < 560 {
			cols = 1
		}
		for start := 0; start < len(stats); start += cols {
			ui.Row(c).Gap(0).Children(func() {
				for i, stat := range stats[start:min(start+cols, len(stats))] {
					cell := ui.ButtonBase(c.Key("usage-kpi:" + stat.metric)).Grow(1).MinWidth(0).Padding(12).Justify(ui.Start).Disabled(!stat.available).Label(a.text("usage.metric_"+stat.key, stat.label))
					if i+1 < cols {
						cell.BorderWidth(0, 1, 0, 0).BorderColor(t.Border)
					}
					if cell.Clicked() {
						s.metric = stat.metric
						a.saveUsagePrefs()
					}
					cell.Children(func() {
						ui.Column(c).Grow(1).AlignSelf(ui.Stretch).Gap(6).Children(func() {
							ui.Text(c, a.text("usage.metric_"+stat.key, stat.label)).TextColor(t.TextMuted).FontSize(11)
							ui.Text(c, stat.value).Font(a.Fonts.CodeFamily).FontSize(20)
							ui.Text(c, stat.hint).TextColor(t.TextMuted).FontSize(11)
						})
					})
				}
			})
		}
		a.line(c)
		a.usageSummaryDetails(c, p)
		if p.Summary.Errors > 0 && a.button(c, a.text("usage.view_failures", "View failures"), "", false).Clicked() {
			s.filter.ErrorsOnly = true
			s.filter.CachedOnly = false
		}
	})
}
