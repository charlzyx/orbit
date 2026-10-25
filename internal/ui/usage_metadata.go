package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
)

// Metadata uses the same projection as the card. A local table search only
// changes that table's match count, never the page's accounting or scopes.
func (a *App) usageSummaryMeta() string {
	p := a.usageState.projection
	return a.usageText("usage.summary_meta", "%{requests} requests · %{tokens} tokens · %{sessions} sessions", map[string]string{"requests": usageCount(uint64(p.Summary.Requests)), "tokens": compactCount(p.Summary.Tokens.Total), "sessions": usageCount(uint64(len(p.Sessions)))})
}
func (a *App) usageTrendMeta() string {
	s := &a.usageState
	by := a.text("usage.granularity_"+s.projection.Granularity, s.projection.Granularity)
	if s.metric == "latency" {
		label := s.latencyMetric
		if label == "average" || label == "" {
			label = a.text("usage.latency_average", "Average")
		}
		return a.usageText("usage.per_bucket", "%{value} per bucket · by %{by}", map[string]string{"value": strings.ToLower(label), "by": by})
	}
	value, valid := usageMetricValue(usageSummaryAggregate(s.projection.Summary), s.metric)
	text := usageMetricText(value, valid, s.metric)
	if !valid {
		text = a.text("usage.unavailable", "Unavailable")
	}
	return a.usageText("usage.total_by", "%{value} total · by %{by}", map[string]string{"value": text, "by": by})
}
func (a *App) usageTableMeta(kind string) string {
	s := &a.usageState
	rows := a.usageTableRows(kind)
	result := queryUsageTable(rows, s.table(kind), kind, s.metric, s.latencyMetric)
	granularity := s.projection.Granularity
	if kind == "daily" && granularity == "hour" {
		granularity = "day"
	}
	count, total := usageCount(uint64(result.Total)), usageCount(uint64(len(rows)))
	vars := map[string]string{"count": count, "buckets": count, "total": total, "by": a.text("usage.granularity_"+granularity, granularity)}
	switch kind {
	case "sessions":
		if result.Total == len(rows) {
			return a.usageText("usage.n_sessions", "%{count} sessions", vars)
		}
		return a.usageText("usage.n_of_n_sessions", "%{count} of %{total} sessions", vars)
	case "failures":
		provider, tool := 0, 0
		for _, e := range s.projection.Errors {
			if e.Kind == "provider" {
				provider++
			}
			if e.Kind == "tool" {
				tool++
			}
		}
		vars["provider"], vars["tool"], vars["stopped"], vars["events"] = usageCount(uint64(provider)), usageCount(uint64(tool)), usageCount(uint64(s.projection.Summary.Aborted)), count
		if result.Total == len(rows) {
			return a.usageText("usage.failures_meta", "%{provider} provider errors recorded · %{tool} tool failures · %{stopped} stopped", vars)
		}
		return a.usageText("usage.failures_meta_filtered", "%{events} of %{total} events · %{provider} provider · %{tool} tool · %{stopped} stopped", vars)
	default:
		if result.Total == len(rows) {
			return a.usageText("usage.n_buckets_by", "%{buckets} buckets · by %{by}", vars)
		}
		return a.usageText("usage.n_of_n_buckets_by", "%{buckets} of %{total} buckets · by %{by}", vars)
	}
}
func (a *App) usageBreakdownMeta() string {
	s := &a.usageState
	p := s.projection
	vars := map[string]string{"requests": usageCount(uint64(p.Summary.Requests)), "tokens": compactCount(p.Summary.Tokens.Total)}
	switch s.breakdown {
	case "providers":
		vars["providers"] = usageCount(uint64(len(p.Providers)))
		return a.usageText("usage.meta_providers_requests", "%{providers} providers · %{requests} requests", vars)
	case "workspaces":
		vars["workspaces"] = usageCount(uint64(len(p.Workspaces)))
		return a.usageText("usage.meta_workspaces_tokens", "%{workspaces} workspaces · %{tokens} tokens", vars)
	case "tools":
		calls := p.Summary.ToolRuns
		if calls == 0 {
			return a.text("usage.no_tool_calls", "no tool calls")
		}
		vars["calls"], vars["failed"], vars["rate"] = usageCount(uint64(calls)), usageCount(uint64(p.ToolFailures)), usagePercent(100*float64(p.ToolFailures)/float64(calls))
		if p.ToolFailures > 0 {
			return a.usageText("usage.meta_calls_failed_pct", "%{calls} calls · %{failed} failed (%{rate})", vars)
		}
		return a.usageText("usage.meta_calls_none_failed", "%{calls} calls · none failed", vars)
	default:
		vars["models"] = usageCount(uint64(len(p.Models)))
		return a.usageText("usage.meta_models_requests", "%{models} models · %{requests} requests", vars)
	}
}

func (a *App) usageMessageState(c *ui.Context, icon, title, body, action string, run func()) {
	ui.Column(c).AlignSelf(ui.Stretch).Height(400).AlignItems(ui.Center).Justify(ui.Center).Gap(8).Label("Usage message state").Children(func() {
		ui.Box(c).Size(64, 64).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(12).Center().Label("Usage message glyph").Children(func() { a.icon(c, icon, 32, a.Theme.TextMuted) })
		ui.Text(c, title).FontSize(a.Theme.Px(16)).Height(a.Theme.Px(20)).FontWeight(500)
		ui.Text(c, body).MaxWidth(440).FontSize(a.Theme.Px(12)).TextAlign(ui.Center).TextColor(a.Theme.TextMuted)
		if action != "" && run != nil {
			if a.button(c, action, "", false).AlignSelf(ui.Center).Margin(2, 0, 0, 0).Border(1, a.Theme.Border).Clicked() {
				run()
			}
		}
	})
}

func (a *App) usageTableTotals(kind string, result usageTableResult) (string, string) {
	count := func(n int) string { return usageCount(uint64(n)) }
	if kind == "failures" {
		return a.usageText("usage.filtered_totals_events", "Filtered totals: %{events} events", map[string]string{"events": count(result.Total)}), a.usageText("usage.page_events", "this page: %{events} events", map[string]string{"events": count(len(result.Rows))})
	}
	if kind == "breakdown:tools" {
		return a.usageText("usage.filtered_totals_calls", "Filtered totals: %{calls} calls · %{failed} failed", map[string]string{"calls": count(result.Totals.Requests), "failed": count(result.Totals.ToolFailures)}), a.usageText("usage.page_calls", "this page: %{calls} calls · %{failed} failed", map[string]string{"calls": count(result.PageTotals.Requests), "failed": count(result.PageTotals.ToolFailures)})
	}
	totals := result.Totals
	if kind == "sessions" || kind == "series" {
		totals = usageSummaryAggregate(a.usageState.projection.Summary)
	}
	filtered := a.usageText("usage.filtered_totals_requests", "Filtered totals: %{requests} requests · %{tokens} tokens", map[string]string{"requests": count(totals.Requests), "tokens": compactCount(totals.Tokens.Total)})
	if totals.Priced > 0 {
		filtered += " · " + usageCost(totals.Cost)
	}
	page := a.usageText("usage.page_requests", "this page: %{requests} requests · %{tokens} tokens", map[string]string{"requests": count(result.PageTotals.Requests), "tokens": compactCount(result.PageTotals.Tokens.Total)})
	return filtered, page
}
func (a *App) usageTableEmpty(kind string, footer bool) string {
	key, fallback := "no_rows_match_search", "No rows match this search."
	switch kind {
	case "sessions":
		key, fallback = "no_sessions_match_search", "No sessions match this search."
	case "daily", "series":
		key, fallback = "no_buckets_match_search", "No buckets match this search."
	case "failures":
		key, fallback = "no_failures_match_search", "No failures match this search."
	case "breakdown:tools":
		key, fallback = "no_tools_match_search", "No tools match this search."
	}
	if footer {
		if strings.HasPrefix(kind, "breakdown:") {
			key, fallback = "no_rows", "No rows"
		} else {
			key = strings.TrimSuffix(key, "_search")
			fallback = strings.TrimSuffix(fallback, " this search.")
		}
	}
	return a.text("usage."+key, fallback)
}
