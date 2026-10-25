package orbitui

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

func contextTokens(n int64) string {
	if n < 1000 {
		return fmt.Sprint(n)
	}
	unit, suffix := int64(1000), "K"
	if n >= 1000000 {
		unit, suffix = 1000000, "M"
	}
	whole, tenth := n/unit, (n%unit+unit/20)/(unit/10)
	if tenth == 10 {
		whole++
		tenth = 0
	}
	if tenth == 0 {
		return fmt.Sprintf("%d%s", whole, suffix)
	}
	return fmt.Sprintf("%d.%d%s", whole, tenth, suffix)
}
func contextTokenValue(n int64, known bool) string {
	if !known {
		return "—"
	}
	return contextTokens(n)
}

func contextCost(n float64) string {
	if math.IsNaN(n) || math.IsInf(n, 0) || n <= 0 {
		return "$0.00"
	}
	if n < .01 {
		return "<$0.01"
	}
	if n < 100 {
		return fmt.Sprintf("$%.2f", n)
	}
	return fmt.Sprintf("$%.0f", n)
}
func contextHitPercent(n float64) string {
	rounded := math.Round(math.Max(0, math.Min(100, n))*10) / 10
	if n < 100 && rounded >= 100 {
		rounded = 99.9
	}
	return strings.TrimSuffix(fmt.Sprintf("%.1f", rounded), ".0")
}
func (a *App) contextText(key string, vars map[string]string) string {
	return orbitassets.Translate(a.catalog, a.english, "context_meter."+key, vars)
}
func (a *App) conversationEstimate() int64 {
	var chars int64
	for _, m := range a.Model.Messages {
		if m.Role == "changedFiles" {
			continue
		}
		if m.Logical != nil {
			for _, s := range m.Logical.Steps {
				chars += int64(utf8.RuneCountInString(s.Text) + utf8.RuneCountInString(s.Thinking))
				for _, tool := range s.Tools {
					chars += int64(utf8.RuneCountInString(tool.Name) + utf8.RuneCountInString(tool.Summary))
				}
			}
		} else {
			chars += int64(utf8.RuneCountInString(m.Text) + utf8.RuneCountInString(m.Thinking))
		}
	}
	return (chars + 3) / 4
}
func (a *App) contextFill(u Usage) ui.Color {
	if u.ContextAvailable && u.ContextPercent >= 90 {
		return a.Theme.Error
	}
	if u.ContextAvailable && u.ContextPercent >= 70 {
		return a.Theme.Warning
	}
	return a.Theme.Accent
}
func (a *App) contextTotals(u Usage) string {
	if u.ContextWindow <= 0 {
		return a.Tr("context_meter.waiting_for_agent")
	}
	vars := map[string]string{"total": contextTokens(u.ContextWindow), "used": contextTokens(u.ContextTokens)}
	if u.ContextTokensAvailable {
		return a.contextText("tokens_of", vars)
	}
	return a.contextText("of_tokens", vars)
}
func (a *App) contextControl(c *ui.Context) {
	t, u := a.Theme, a.Model.Usage
	fill := a.contextFill(u)
	anchor := ui.ButtonBase(c.Key("context-meter")).Label("Context usage").Row().AlignItems(ui.Center).Gap(t.Space(6)).Shrink(0)
	var ring ui.Element
	anchor.Children(func() {
		if label := contextPercentLabel(u); label != "" {
			ui.Text(c, label).SingleLine().FontSize(t.Px(12)).TextColor(fill)
		}
		ring = ui.Box(c.Key("context-ring-box")).Size(22, 22).Center()
		ring.Children(func() {
			ui.Box(c.Key("context-ring")).Label("Context ring").Size(16, 16).Draw(func(p *ui.Painter, r ui.Rect) {
				cx, cy := r.X+r.W/2, r.Y+r.H/2
				path := new(ui.Path).Circle(cx, cy, 7)
				p.StrokePath(path, 2, t.Elevated)
				if u.ContextAvailable && u.ContextPercent >= .8 {
					const k float32 = .5522847498
					path = new(ui.Path).MoveTo(cx, cy-7).CubeTo(cx+7*k, cy-7, cx+7, cy-7*k, cx+7, cy).CubeTo(cx+7, cy+7*k, cx+7*k, cy+7, cx, cy+7).CubeTo(cx-7*k, cy+7, cx-7, cy+7*k, cx-7, cy).CubeTo(cx-7, cy-7*k, cx-7*k, cy-7, cx, cy-7).Close().Trim(0, float32(u.ContextPercent/100))
					if u.ContextPercent >= 99.7 {
						path.Trim(0, 1)
					}
					p.StrokePath(path, 2, fill)
				}
			})
		})
	})
	if anchor.Clicked() && c.Now().Sub(a.contextDismissed) >= 200*time.Millisecond {
		a.contextOpen = !a.contextOpen
		if a.contextOpen && a.Model.Connected {
			a.Dispatch(Action{Kind: "RefreshContextStats"})
		}
	}
	if a.contextOpen {
		panel := ui.PopoverBase(c.Key("context-details"), ring, &a.contextOpen, func(panel ui.Element) {
			panel.AttachTo(ring, ui.AnchorTopRight, ui.AnchorBottomRight).Margin(0, 0, 8, 0).Column().Width(280).Padding(t.Space(12), 14).Gap(10).Background(t.Elevated).Border(1, t.Border).Radius(t.Px(12)).Shadow(0, 6, 12, 0, ui.RGBA(0, 0, 0, .12)).Label("Context details")
			a.contextDetails(c)
		})
		if panel.Dismissed() {
			a.contextDismissed = c.Now()
		}
	} else if anchor.Hovered() {
		ui.Overlay(c.Key("context-hover"), func() {
			ui.Column(c).AttachTo(ring, ui.AnchorTopRight, ui.AnchorBottomRight).Margin(0, 0, 8, 0).Width(200).Padding(t.Space(8), t.Space(12)).Gap(2).Background(t.Elevated).Border(1, t.Border).Radius(t.Px(12)).Shadow(0, 6, 12, 0, ui.RGBA(0, 0, 0, .12)).Label("Context summary").Children(func() {
				title := a.Tr("context_meter.context_usage")
				if label := contextPercentLabel(u); label != "" {
					title = a.contextText("context_used", map[string]string{"percent": label})
				}
				ui.Text(c, title).FontSize(t.FontSize).TextColor(t.Text).SingleLine()
				ui.Text(c, a.contextTotals(u)).FontSize(t.Px(12)).TextColor(t.TextMuted).SingleLine()
				if u.SessionAvailable && u.CostAvailable {
					ui.Text(c, a.contextText("spent", map[string]string{"cost": contextCost(u.Cost)})).FontSize(t.Px(12)).TextColor(t.TextMuted).SingleLine()
				}
			})
		})
	}
}
func (a *App) contextDetails(c *ui.Context) {
	t, u := a.Theme, a.Model.Usage
	ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Children(func() {
		ui.Text(c, a.Tr("context_meter.context_usage")).Grow(1).FontSize(t.FontSize).TextColor(t.TextSecondary)
		if a.iconButton(c, "Close context details", "icons/x.svg").Size(t.Px(18), t.Px(18)).Clicked() {
			a.contextOpen = false
			a.contextDismissed = c.Now()
		}
	})
	ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Children(func() {
		text := a.Tr("context_meter.unknown")
		if label := contextPercentLabel(u); label != "" {
			text = a.contextText("percent_full", map[string]string{"percent": label})
		}
		ui.Text(c, text).Grow(1).FontSize(t.FontSize).TextColor(t.Text)
		if u.ContextWindow > 0 {
			ui.Text(c, a.contextTotals(u)).FontSize(t.Px(12)).TextColor(t.TextMuted)
		}
	})
	type slice struct {
		key    string
		tokens int64
		color  ui.Color
	}
	var slices []slice
	if u.ContextWindow > 0 && u.ContextTokensAvailable && u.ContextTokens > 0 {
		conversation := min(a.conversationEstimate(), u.ContextTokens)
		if other := u.ContextTokens - conversation; other > 0 {
			key, color := "other_context", t.Accent.Alpha(.55)
			if conversation == 0 {
				key, color = "in_use", a.contextFill(u)
			}
			slices = append(slices, slice{key, other, color})
		}
		if conversation > 0 {
			slices = append(slices, slice{"conversation", conversation, t.Accent})
		}
	}
	ui.Row(c).Label("Context allocation").AlignSelf(ui.Stretch).Height(6).Radius(3).Clip().Background(t.SurfaceSecondary).Children(func() {
		if u.ContextWindow > 0 {
			for _, s := range slices {
				ui.Box(c).HeightPercent(100).WidthPercent(float32(math.Min(1, float64(s.tokens)/float64(u.ContextWindow)) * 100)).Background(s.color)
			}
		}
	})
	ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(6)).Children(func() {
		if len(slices) == 0 {
			ui.Text(c, a.Tr("context_meter.no_usage_reported_for_this_session_yet")).FontSize(t.Px(12)).TextColor(t.TextMuted)
		}
		for _, s := range slices {
			ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(t.Space(8)).Children(func() {
				ui.Box(c).Size(8, 8).Radius(4).Background(s.color)
				ui.Text(c, a.Tr("context_meter."+s.key)).Grow(1).FontSize(t.Px(12)).TextColor(t.TextSecondary)
				ui.Text(c, contextTokens(s.tokens)).FontSize(t.Px(12)).TextColor(t.Text)
			})
		}
	})
	if u.SessionAvailable {
		ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(6)).Children(func() {
			ui.Box(c).AlignSelf(ui.Stretch).Height(1).Background(t.Border)
			ui.Text(c, a.Tr("context_meter.session_usage")).FontSize(t.Px(12)).TextColor(t.TextMuted)
			cache := contextTokenValue(u.CacheRead, u.TokensAvailable || u.CacheReadAvailable)
			if total := float64(u.Input) + float64(u.CacheRead) + float64(u.CacheWrite); total > 0 && (u.TokensAvailable || u.InputAvailable && u.CacheReadAvailable && u.CacheWriteAvailable) {
				rate := float64(u.CacheRead) / total * 100
				cache = a.contextText("cache_read_share", map[string]string{"tokens": cache, "percent": contextHitPercent(rate)})
			}
			cost := "—"
			if u.CostAvailable {
				cost = contextCost(u.Cost)
			}
			for _, row := range []struct{ key, icon, value string }{{"input", "usage-input", contextTokenValue(u.Input, u.TokensAvailable || u.InputAvailable)}, {"output", "usage-output", contextTokenValue(u.Output, u.TokensAvailable || u.OutputAvailable)}, {"cache_read", "cache-read", cache}, {"cache_write", "cache-write", contextTokenValue(u.CacheWrite, u.TokensAvailable || u.CacheWriteAvailable)}, {"total", "usage-total", contextTokenValue(u.Total, u.TotalAvailable)}, {"cost", "usage-cost", cost}} {
				ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(t.Space(8)).Children(func() {
					a.icon(c, "icons/"+row.icon+".svg", t.Px(12), t.TextMuted)
					ui.Text(c, a.Tr("context_meter."+row.key)).Grow(1).FontSize(t.Px(12)).TextColor(t.TextSecondary)
					ui.Text(c, row.value).FontSize(t.Px(12)).TextColor(t.Text)
				})
			}
		})
	}
	if a.Model.Connected {
		label := a.Tr("context_meter.compact_now")
		if a.Model.Compacting {
			label = a.Tr("context_meter.compacting")
		}
		b := ui.ButtonBase(c.Key("context-compact")).Label("Compact context").AlignSelf(ui.Stretch).Height(t.Px(28)).PaddingX(t.Space(8)).Border(1, t.Border).Background(t.Elevated).Radius(t.Px(6)).Center().Disabled(a.Model.Compacting || a.Model.SessionID == "")
		color := t.TextSecondary
		if a.Model.Compacting {
			color = t.TextMuted
		} else if b.Hovered() {
			b.Background(t.Hover)
		}
		b.Children(func() { ui.Text(c, label).FontSize(t.FontSize).FontWeight(500).TextColor(color) })
		if b.Clicked() {
			a.Dispatch(Action{Kind: "CompactContext"})
		}
	}
}
