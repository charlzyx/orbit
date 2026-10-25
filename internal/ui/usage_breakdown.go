package orbitui

import (
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func (a *App) usageBreakdownChartRows(kind string) usageTableResult {
	rows := a.usageTableRows(kind)
	tab := *a.usageState.table(kind)
	tab.Page, tab.Size = 1, max(1, len(rows))
	return queryUsageTable(rows, &tab, kind, a.usageState.metric, a.usageState.latencyMetric)
}
func (a *App) usageBreakdownFraction(r usageTableRow, totals usageAggregate, tools bool) float64 {
	if tools {
		if totals.Requests > 0 {
			return float64(r.Requests) / float64(totals.Requests)
		}
		return 0
	}
	if totals.Tokens.Total > 0 {
		return float64(r.Tokens.Total) / float64(totals.Tokens.Total)
	}
	return usageShare(r.usageAggregate, a.usageState.projection.Summary)
}
func (a *App) usageBreakdownChart(c *ui.Context, kind string) {
	result := a.usageBreakdownChartRows(kind)
	tools := kind == "breakdown:tools"
	ui.Column(c).AlignSelf(ui.Stretch).Padding(0, 0, a.Theme.Space(12), 0).Gap(a.Theme.Space(2)).Label("Usage breakdown chart " + kind).Children(func() {
		if len(result.Rows) == 0 {
			key, fallback := "usage.no_usage_in_range", "No usage in this range."
			if tools {
				key, fallback = "usage.no_tool_calls_in_range", "No tool calls in this range."
			}
			ui.Text(c, a.text(key, fallback)).FontSize(a.Theme.Px(12)).FixedLineHeight(usageLineHeight(a.Theme.Px(12))).TextColor(a.Theme.TextMuted)
		}
		for _, r := range result.Rows[:min(8, len(result.Rows))] {
			fraction := a.usageBreakdownFraction(r, result.Totals, tools)
			label, sub := r.Name, r.Sub
			value := compactCount(r.Tokens.Total)
			if tools {
				class := workbench.UsageToolClass(r.Name)
				sub = a.text("usage.class_"+class, titleCase(class))
				value = usageCount(uint64(r.Requests))
			}
			line := ui.Row(c.Key(r.Key)).AlignSelf(ui.Stretch).Padding(a.Theme.Space(4), a.Theme.Space(8)).Radius(a.Theme.Px(6)).Gap(a.Theme.Space(8)).AlignItems(ui.Center).Label("Usage breakdown bar " + r.Key)
			if !tools {
				line.Cursor(ui.CursorPointer)
				if line.Hovered() {
					line.Background(a.Theme.Hover)
				}
				if line.Clicked() {
					a.activateUsageRow(kind, r)
				}
			} else {
				line.Tooltip(a.usageText("usage.tool_chart_tooltip", "%{name} · %{calls} calls · %{errors} failed · %{avg} avg", map[string]string{"name": label, "calls": value, "errors": usageCount(uint64(r.ToolFailures)), "avg": a.usageCell(r, "avg", kind)}))
			}
			line.Children(func() {
				// Source chart_label stacks Small/XSmall text with GPUI's phi
				// line boxes. Natural MyGo line metrics compress every row.
				ui.Column(c).Width(180).Shrink(0).Children(func() {
					ui.Text(c, label).SingleLine().FontSize(a.Theme.Px(12)).FixedLineHeight(usageLineHeight(a.Theme.Px(12))).TextColor(a.Theme.Text)
					if sub != "" {
						ui.Text(c, sub).SingleLine().FontSize(a.Theme.Px(10)).FixedLineHeight(usageLineHeight(a.Theme.Px(10))).TextColor(a.Theme.TextMuted)
					}
				})
				ui.Box(c).Grow(1).MinWidth(48).Height(8).Background(a.Theme.SurfaceSecondary).Radius(4).Clip().Label("Usage breakdown track " + r.Key).Children(func() {
					ui.Box(c).Height(8).WidthPercent(float32(max(0, min(1, fraction)) * 100)).Radius(4).Background(a.Theme.Accent).Label("Usage breakdown fill " + r.Key)
				})
				ui.Text(c, value).Width(72).Shrink(0).Font(a.Fonts.CodeFamily).FontSize(a.Theme.Px(12)).FixedLineHeight(usageLineHeight(a.Theme.Px(12))).TextColor(a.Theme.Text).TextAlign(ui.End)
				ui.Text(c, usagePercent(fraction*100)).Width(56).Shrink(0).SingleLine().Font(a.Fonts.CodeFamily).FontSize(a.Theme.Px(12)).FixedLineHeight(usageLineHeight(a.Theme.Px(12))).TextColor(a.Theme.TextMuted).TextAlign(ui.End)
			})
		}
	})
}
