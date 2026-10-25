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
			if point.Tokens.Total > peak.Tokens.Total {
				peak = point
			}
		}
	}
	if nonzero >= 5 {
		mean /= float64(nonzero)
		if float64(peak.Tokens.Total) >= mean*2 {
			add("usage.insight_peak_bucket", fmt.Sprintf("%s is %.1f× the average bucket volume.", peak.At.Format("Jan 2"), float64(peak.Tokens.Total)/mean), "warning", "bucket", peak.At.Format("Jan 2"), "factor", fmt.Sprintf("%.1f", float64(peak.Tokens.Total)/mean), "metric", p.Granularity)
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
	ui.Column(c).Gap(6).Children(func() {
		ui.Text(c, label).FontWeight(500)
		if l.Samples == 0 {
			ui.Text(c, "Not measured").TextColor(a.Theme.TextMuted)
			return
		}
		avg, _ := l.value("average")
		line := fmt.Sprintf("avg %.2fs · max %s", avg, l.Max.Round(time.Millisecond))
		if l.Samples >= 20 {
			line += fmt.Sprintf(" · p50 %s · p95 %s · p99 %s", l.P50.Round(time.Millisecond), l.P95.Round(time.Millisecond), l.P99.Round(time.Millisecond))
		}
		ui.Text(c, line).Font(a.Fonts.CodeFamily)
		ui.Text(c, usageLatencyHint(l.P95, l.Samples)).TextColor(a.Theme.TextMuted).FontSize(11)
	})
}
func (a *App) usageHealthDetails(c *ui.Context) {
	s := &a.usageState
	p := s.projection
	a.usageCard(c, a.text("usage.cache_performance", "Cache performance"), "", func() {
		rate, ok := usageCacheRate(p.Summary.Tokens)
		if !ok {
			ui.Text(c, a.text("usage.cache_unavailable_range", "Cache data unavailable: no cache tokens were reported in this range.")).TextColor(a.Theme.TextMuted)
			return
		}
		ui.Text(c, fmt.Sprintf("%.1f%% of prompt tokens served from cache", rate)).FontSize(20)
		for _, item := range []struct{ label, value string }{{"Cache reads", compactCount(p.Summary.Tokens.CacheRead)}, {"Cache writes", compactCount(p.Summary.Tokens.CacheWrite)}, {"Uncached input", compactCount(p.Summary.Tokens.Input)}, {"Requests served from cache", fmt.Sprintf("%d of %d", p.CachedRequests, p.Summary.Requests)}} {
			ui.Row(c).Children(func() { ui.Text(c, item.label).Grow(1); ui.Text(c, item.value).Font(a.Fonts.CodeFamily) })
		}
		ui.Text(c, "Hit rate = cache read / (input + cache read + cache write)").FontSize(11).TextColor(a.Theme.TextMuted)
		if a.button(c, map[bool]string{true: "Show all requests", false: "Show cached only"}[s.filter.CachedOnly], "", false).Clicked() {
			s.filter.CachedOnly = !s.filter.CachedOnly
			if s.filter.CachedOnly {
				s.filter.ErrorsOnly = false
			}
		}
		ui.Row(c).Height(50).Gap(2).Children(func() {
			for _, point := range p.Series {
				rate, ok := usageCacheRate(point.Tokens)
				height := float32(3)
				tip := point.At.Format("Jan 2 15:04") + " · no cache traffic"
				if ok {
					height = max(3, float32(rate)/100*50)
					tip = fmt.Sprintf("%s · %.1f%% hit rate", point.At.Format("Jan 2 15:04"), rate)
				}
				ui.Box(c).Grow(1).Height(height).AlignSelf(ui.End).Background(a.Theme.Accent.Alpha(.65)).Tooltip(tip)
			}
		})
	})
	a.usageCard(c, "Response time", "Only measured response and tool durations.", func() {
		a.usageLatencyLine(c, p.Latency, "Model responses")
		a.line(c)
		a.usageLatencyLine(c, p.ToolLatency, "Tools")
	})
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
		cost = fmt.Sprintf("$%.4f · $%.4f / req", s.Cost, s.Cost/float64(s.PricedRequests))
	}
	stats := []struct{ key, label, value string }{{"input_tokens", "Input tokens", compactCount(s.Tokens.Input)}, {"output_tokens", "Output tokens", compactCount(s.Tokens.Output)}, {"cost", "Cost", cost}, {"failed", "Failed", fmt.Sprint(s.Errors)}, {"sessions", "Sessions", fmt.Sprint(len(p.Sessions))}, {"turns", "Turns", fmt.Sprint(s.Turns)}, {"tool_calls", "Tool calls", fmt.Sprint(s.ToolRuns)}, {"bash_commands", "Bash commands", fmt.Sprint(s.BashRuns)}}
	if p.ToolFailures > 0 {
		stats = append(stats, struct{ key, label, value string }{"tool_failures", "Tool failures", fmt.Sprint(p.ToolFailures)})
	}
	if s.TimedRequests > 0 {
		stats = append(stats, struct{ key, label, value string }{"generation_time", "Generation time", s.Duration.Round(time.Second).String()})
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
		stats = append(stats, struct{ key, label, value string }{"cost_per_mtok", "Cost / M tokens", fmt.Sprintf("$%.4f", s.Cost*1e6/float64(s.Tokens.Total))})
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
