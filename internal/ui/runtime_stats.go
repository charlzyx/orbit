package orbitui

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// runtimeUsage preserves the distinct meanings of Pi's cumulative session
// counters and its current contextUsage estimate. Source: Pi 1.4.0
// docs/rpc-commands.md#get_session_stats and orbit-rpc ContextUsage/SessionUsage.
// Never substitute visible transcript totals: Pi also accounts for tools,
// compaction and branch summaries which may not be in that transcript.
func runtimeUsage(s agent.Snapshot) Usage {
	var stats map[string]json.RawMessage
	if json.Unmarshal(s.Stats, &stats) != nil {
		return Usage{}
	}
	var out Usage
	var tokens map[string]json.RawMessage
	if json.Unmarshal(stats["tokens"], &tokens) == nil && tokens != nil {
		var a, b, c, d bool
		out.Input, a = usageInteger(tokens["input"])
		out.Output, b = usageInteger(tokens["output"])
		out.CacheRead, c = usageInteger(tokens["cacheRead"])
		out.CacheWrite, d = usageInteger(tokens["cacheWrite"])
		// The pinned SessionStats schema supplies every bucket. A partial or
		// malformed object must not make omitted buckets look like real zeros.
		out.TokensAvailable = a && b && c && d
	}
	out.Cost, out.CostAvailable = usageNumber(stats["cost"])
	var context map[string]json.RawMessage
	if json.Unmarshal(stats["contextUsage"], &context) == nil && context != nil {
		out.ContextWindow, _ = usageInteger(context["contextWindow"])
		var haveTokens bool
		out.ContextTokens, haveTokens = usageInteger(context["tokens"])
		percent, havePercent := usageNumber(context["percent"])
		if !havePercent && haveTokens && out.ContextWindow > 0 {
			percent, havePercent = float64(out.ContextTokens)/float64(out.ContextWindow)*100, true
		}
		if out.ContextWindow > 0 && havePercent && !s.Compacting {
			out.ContextPercent = math.Max(0, math.Min(100, percent))
			out.ContextAvailable = true
		}
	}
	out.Available = out.TokensAvailable || out.CostAvailable || out.ContextAvailable
	return out
}

// latestAssistantUsage is per-message usage, not cumulative session usage or a
// context estimate. The newest assistant message owns the answer even if it
// has no reported usage yet; falling back further would show stale output.
func latestAssistantUsage(s agent.Snapshot) Usage {
	for i := len(s.Messages) - 1; i >= 0; i-- {
		message := s.Messages[i]
		if message.Role != "assistant" {
			continue
		}
		var value map[string]json.RawMessage
		if json.Unmarshal(message.Usage, &value) != nil || value == nil {
			return Usage{}
		}
		out := Usage{}
		out.Input, _ = usageInteger(value["input"])
		out.Output, _ = usageInteger(value["output"])
		out.CacheRead, _ = usageInteger(value["cacheRead"])
		out.CacheWrite, _ = usageInteger(value["cacheWrite"])
		total, ok := usageInteger(value["totalTokens"])
		nonzero := out.Input > 0 || out.Output > 0 || out.CacheRead > 0 || out.CacheWrite > 0
		out.TokensAvailable = nonzero || ok && total > 0
		out.Cost, out.CostAvailable = usageNumber(value["cost"])
		if !out.CostAvailable {
			var cost map[string]json.RawMessage
			if json.Unmarshal(value["cost"], &cost) == nil {
				out.Cost, out.CostAvailable = usageNumber(cost["total"])
			}
		}
		// Pi streaming placeholders report all-zero usage; Orbit suppresses them.
		if !out.TokensAvailable && (!out.CostAvailable || out.Cost == 0) {
			return Usage{}
		}
		out.Available = true
		return out
	}
	return Usage{}
}
func usageNumber(raw json.RawMessage) (float64, bool) {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "null" {
		return 0, false
	}
	var value float64
	if json.Unmarshal(raw, &value) != nil {
		return 0, false
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, false
	}
	return value, true
}
func usageInteger(raw json.RawMessage) (int64, bool) {
	n, ok := usageNumber(raw)
	if !ok || n >= float64(math.MaxInt64) || math.Trunc(n) != n {
		return 0, false
	}
	return int64(n), true
}
func contextPercentLabel(usage Usage) string {
	if !usage.ContextAvailable {
		return ""
	}
	if usage.ContextTokens > 0 && math.Round(usage.ContextPercent) < 1 {
		return "<1%"
	}
	return fmt.Sprintf("%.0f%%", usage.ContextPercent)
}
func cacheReadPercent(usage Usage) (float64, bool) {
	total := float64(usage.Input) + float64(usage.CacheRead) + float64(usage.CacheWrite)
	if !usage.TokensAvailable || total <= 0 {
		return 0, false
	}
	return math.Max(0, math.Min(100, float64(usage.CacheRead)/total*100)), true
}

type quotaHeadline struct {
	Provider, Window, Text string
	Fraction               float64
	Meter                  bool
}

func quotaFraction(window agent.QuotaWindow) (float64, bool) {
	if p := window.UsedPercent; p != nil && !math.IsNaN(*p) && !math.IsInf(*p, 0) {
		return math.Max(0, math.Min(1, *p/100)), true
	}
	if window.Used != nil && window.Limit != nil && *window.Limit > 0 && !math.IsNaN(*window.Used) && !math.IsInf(*window.Used, 0) && !math.IsInf(*window.Limit, 0) {
		return math.Max(0, math.Min(1, *window.Used / *window.Limit)), true
	}
	return 0, false
}
func fiveHourWindow(window agent.QuotaWindow) bool {
	for _, value := range []string{window.ID, window.Label} {
		value = strings.NewReplacer("five", "5", "-", " ", "_", " ").Replace(strings.ToLower(value))
		tokens := strings.Fields(value)
		for i, token := range tokens {
			if token == "5h" || token == "5" && i+1 < len(tokens) && tokens[i+1] == "hour" {
				return true
			}
		}
	}
	return false
}
func quotaReports(reports []agent.QuotaReport) []agent.QuotaReport {
	out := make([]agent.QuotaReport, 0, len(reports))
	for _, report := range reports {
		if len(report.Windows) > 0 || len(report.Balances) > 0 || report.Note != "" || report.Error != "" {
			out = append(out, report)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Provider < out[j].Provider })
	return out
}

// quotaSummary mirrors Orbit QuotaManager::headline: active 5h, active peak,
// active balance, global 5h peak, global peak. Errored reports never meter.
func quotaSummary(reports []agent.QuotaReport, provider string) quotaHeadline {
	reports = quotaReports(reports)
	makeWindow := func(report agent.QuotaReport, window agent.QuotaWindow) quotaHeadline {
		fraction, _ := quotaFraction(window)
		label := window.Label
		if fiveHourWindow(window) {
			label = "5h"
		}
		return quotaHeadline{Provider: report.Provider, Window: label, Text: fmt.Sprintf("%.0f%%", fraction*100), Fraction: fraction, Meter: true}
	}
	for _, report := range reports {
		if report.Provider != provider || report.Error != "" {
			continue
		}
		best := -1
		peak := -1.0
		for i, window := range report.Windows {
			if fraction, ok := quotaFraction(window); ok {
				if fiveHourWindow(window) {
					return makeWindow(report, window)
				}
				if fraction > peak {
					best, peak = i, fraction
				}
			}
		}
		if best >= 0 {
			return makeWindow(report, report.Windows[best])
		}
		if len(report.Balances) > 0 {
			balance := report.Balances[0]
			return quotaHeadline{Provider: report.Provider, Text: strings.TrimSpace(quotaAmount(balance.Amount) + " " + balance.Currency)}
		}
	}
	for _, onlyFive := range []bool{true, false} {
		peak := -1.0
		var best quotaHeadline
		for _, report := range reports {
			if report.Error != "" {
				continue
			}
			for _, window := range report.Windows {
				if onlyFive && !fiveHourWindow(window) {
					continue
				}
				if fraction, ok := quotaFraction(window); ok && fraction > peak {
					peak = fraction
					best = makeWindow(report, window)
				}
			}
		}
		if peak >= 0 {
			return best
		}
	}
	return quotaHeadline{Text: "Usage"}
}
func quotaAmount(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "Unavailable"
	}
	if math.Trunc(value) == value {
		return fmt.Sprintf("%.0f", value)
	}
	return fmt.Sprintf("%.2f", value)
}
func quotaResetHint(epoch int64, now time.Time) string {
	reset := time.UnixMilli(epoch)
	remaining := reset.Sub(now)
	if remaining <= 0 || remaining >= 7*24*time.Hour {
		return "resets at " + reset.In(now.Location()).Format("Jan 2, 15:04")
	}
	if remaining < time.Minute {
		return "resets in <1m"
	}
	if remaining < time.Hour {
		return fmt.Sprintf("resets in %dm", int(remaining/time.Minute))
	}
	if remaining < 24*time.Hour {
		h, m := int(remaining/time.Hour), int(remaining/time.Minute)%60
		if m == 0 {
			return fmt.Sprintf("resets in %dh", h)
		}
		return fmt.Sprintf("resets in %dh %dm", h, m)
	}
	d, h := int(remaining/(24*time.Hour)), int(remaining/time.Hour)%24
	if h == 0 {
		return fmt.Sprintf("resets in %dd", d)
	}
	return fmt.Sprintf("resets in %dd %dh", d, h)
}

// runtimeQuotaControl has no process/network access. Mount it in the header,
// passing persistent UI-owned open state; RefreshQuota must be handled by the
// controller. Hidden-provider preferences can filter only the popup reports,
// never the summary or the provider settings page.
func (a *App) runtimeQuotaControl(c *ui.Context, open *bool, reports []agent.QuotaReport, support, provider string) {
	reports = quotaReports(reports)
	if len(reports) == 0 {
		return
	}
	headline := quotaSummary(reports, provider)
	label := strings.TrimSpace(strings.Join([]string{headline.Provider, headline.Window, headline.Text}, " "))
	anchor := a.button(c, label, "icons/spark.svg", *open).Height(28)
	if anchor.Clicked() {
		*open = !*open
	}
	ui.Popover(c, anchor, open, func() {
		ui.Column(c).Width(340).MaxHeight(540).Padding(12).Gap(10).Children(func() {
			ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
				ui.Text(c, "Provider usage").Grow(1).FontWeight(500)
				if a.iconButton(c, "Refresh provider usage", "icons/refresh.svg").Clicked() {
					a.Dispatch(Action{Kind: "RefreshQuota"})
				}
			})
			ui.Scroll(c).MaxHeight(460).Gap(10).Children(func() {
				for _, report := range reports {
					a.runtimeQuotaCard(c.Key(report.Provider), report, time.Now())
				}
				if support == "unsupported" {
					ui.Text(c, "Showing extension-reported quota; direct quota RPC is unavailable.").FontSize(11).TextColor(a.Theme.TextMuted)
				}
			})
		})
	})
}
func (a *App) runtimeQuotaCard(c *ui.Context, report agent.QuotaReport, now time.Time) {
	ui.Column(c).Padding(10).Gap(6).Border(1, a.Theme.Border).Radius(8).Children(func() {
		ui.Row(c).Gap(8).Children(func() {
			ui.Text(c, report.Provider).Grow(1).FontWeight(500)
			if report.Plan != "" {
				ui.Text(c, report.Plan).TextColor(a.Theme.TextMuted)
			}
		})
		for _, window := range report.Windows {
			value := ""
			switch {
			case window.UsedPercent != nil:
				value = quotaAmount(*window.UsedPercent) + "% used"
			case window.Used != nil && window.Limit != nil:
				value = quotaAmount(*window.Used) + " / " + quotaAmount(*window.Limit)
			case window.Used != nil:
				value = strings.TrimSpace(quotaAmount(*window.Used) + " " + window.Unit)
			}
			if value == "" {
				continue
			}
			ui.Row(c).Gap(8).Children(func() { ui.Text(c, window.Label).Grow(1).TextColor(a.Theme.TextMuted); ui.Text(c, value) })
			if fraction, ok := quotaFraction(window); ok && report.Error == "" {
				ui.Progress(c, fraction).Height(5)
			}
			if window.ResetsAt != nil {
				ui.Text(c, quotaResetHint(*window.ResetsAt, now)).FontSize(11).TextColor(a.Theme.TextMuted)
			}
		}
		for _, balance := range report.Balances {
			ui.Row(c).Gap(8).Children(func() {
				ui.Text(c, balance.Label).Grow(1).TextColor(a.Theme.TextMuted)
				ui.Text(c, strings.TrimSpace(quotaAmount(balance.Amount)+" "+balance.Currency))
			})
		}
		if report.Error != "" {
			ui.Text(c, report.Error).TextColor(a.Theme.Error)
		} else if report.Note != "" {
			hasReset := false
			for _, window := range report.Windows {
				hasReset = hasReset || window.ResetsAt != nil
			}
			if !hasReset || len(report.Windows) == 0 && len(report.Balances) == 0 {
				ui.Text(c, report.Note).TextColor(a.Theme.TextMuted).FontSize(11)
			}
		}
	})
}
