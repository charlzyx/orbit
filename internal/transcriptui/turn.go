package transcriptui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
)

// SearchText exposes the actual text, thinking, tool argument/result data and
// summary text without dropping those facts when rows are grouped.
func (r Row) SearchText() string {
	var parts []string
	if r.Summary != nil {
		parts = append(parts, r.Summary.Text)
	}
	for _, s := range r.Steps {
		parts = append(parts, s.Text, s.Thinking)
		for _, t := range s.Tools {
			parts = append(parts, t.Name, t.Path, display(t.Arguments, OutputTextCap), display(t.Output, OutputTextCap))
		}
	}
	parts = append(parts, r.Error)
	return strings.Join(parts, "\n")
}

// Part preserves the source ordering of pre-answer and post-answer work groups.
// Group indexes Steps[Start:End]; prose uses Steps[Start].Text.
type Part struct {
	Kind               string
	Start, End         int
	HiddenByFold, Live bool
}

func (r Row) Parts() []Part {
	firstAnswer, firstWork := -1, -1
	for i, s := range r.Steps {
		if firstAnswer < 0 && strings.TrimSpace(s.Text) != "" {
			firstAnswer = i
		}
		if firstWork < 0 && (s.Thinking != "" || len(s.Tools) > 0) {
			firstWork = i
		}
	}
	var parts []Part
	covered := 0
	for i, s := range r.Steps {
		if i == firstWork {
			end := len(r.Steps)
			if firstAnswer >= 0 {
				end = firstAnswer + 1
			}
			if end > 0 {
				has := false
				for _, v := range r.Steps[:end] {
					has = has || v.Thinking != "" || len(v.Tools) > 0
				}
				if has {
					parts = append(parts, Part{"work", 0, end, firstAnswer < 0 || i < firstAnswer, r.Streaming && end == len(r.Steps)})
				}
			}
		}
		if s.Text != "" {
			parts = append(parts, Part{"text", i, i + 1, false, r.Streaming && i == len(r.Steps)-1})
		}
		if firstAnswer >= 0 && i > firstAnswer && i >= covered && (s.Thinking != "" || len(s.Tools) > 0) {
			end := i + 1
			for end < len(r.Steps) && strings.TrimSpace(r.Steps[end].Text) == "" && (r.Steps[end].Thinking != "" || len(r.Steps[end].Tools) > 0) {
				end++
			}
			covered = end
			parts = append(parts, Part{"work", i, end, false, r.Streaming && end == len(r.Steps)})
		}
	}
	return parts
}
func activityTitle(steps []Step, live bool, o Options) string {
	counts := map[string]int{}
	for _, s := range steps {
		if s.Thinking != "" {
			counts["thought"]++
		}
		for _, t := range s.Tools {
			kind := "tool"
			switch {
			case in(t.Name, "bash shell terminal exec run"):
				kind = "command"
			case in(t.Name, "read view grep find glob search list ls tree"):
				kind = "file_read"
			case in(t.Name, "edit write"):
				kind = "file_edit"
			}
			counts[kind]++
		}
	}
	var labels []string
	for _, kind := range []string{"command", "file_read", "file_edit", "tool", "thought"} {
		n := counts[kind]
		if n == 0 {
			continue
		}
		if kind == "thought" && live {
			labels = append(labels, o.tr("transcript.thinking", "Thinking"))
			continue
		}
		word := strings.ReplaceAll(kind, "_", " ")
		key := "transcript_view.ran_" + kind
		fallback := "Ran %{count} " + word
		if kind == "thought" {
			key = "transcript_view.thought"
			fallback = "%{count} thought"
		}
		if n == 1 {
			key += "_one"
		} else {
			key += "_other"
			fallback += "s"
		}
		labels = append(labels, o.tr(key, fallback, "count", fmt.Sprint(n)))
	}
	if len(labels) == 0 {
		if live {
			return o.tr("transcript.working", "Working")
		}
		return o.tr("transcript.worked", "Worked")
	}
	return strings.Join(labels, " · ")
}
func activityIcons(steps []Step) []string {
	seen := map[string]bool{}
	var icons []string
	thought := false
	for _, s := range steps {
		thought = thought || s.Thinking != ""
		for _, t := range s.Tools {
			icon := Icon(t.Name)
			if !seen[icon] {
				seen[icon] = true
				icons = append(icons, icon)
			}
		}
	}
	rank := func(icon string) int {
		switch {
		case strings.Contains(icon, "bash"):
			return 0
		case strings.Contains(icon, "read") || strings.Contains(icon, "search") || strings.Contains(icon, "find") || strings.Contains(icon, "list"):
			return 1
		case strings.Contains(icon, "edit") || strings.Contains(icon, "write"):
			return 2
		case strings.Contains(icon, "web") || strings.Contains(icon, "fetch"):
			return 3
		}
		return 4
	}
	sort.SliceStable(icons, func(i, j int) bool { return rank(icons[i]) < rank(icons[j]) })
	if thought {
		icons = append(icons, "icons/tools/thinking.svg")
	}
	return icons[:min(6, len(icons))]
}

// Turn paints the assistant's complete logical turn. User bubbles remain the
// host's concern. TextSelection must be mounted on the transcript ancestor.
func Turn(c *ui.Context, s *State, row Row, o Options) ui.Element {
	o = defaults(c, o)
	c = c.Key("turn:" + row.ID)
	key := "turn:" + row.ID
	if elapsed := s.elapsed(row.ID, o.now(c)); elapsed != nil {
		row.Elapsed = elapsed
	}
	return ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Gap(8).Children(func() {
		if row.Summary != nil {
			Summary(c, s, row, o)
			return
		}
		for i, img := range row.Images {
			if o.Image != nil {
				o.Image(c.Key(fmt.Sprintf("image-%d", i)), img)
			}
		}
		open := s.Open(key, false)
		if !row.Streaming && row.HasWork() {
			label := o.tr("transcript.worked", "Worked")
			if row.Elapsed != nil {
				label = o.tr("transcript_view.worked_for", "Worked for %{duration}", "duration", formatDuration(*row.Elapsed))
			}
			toggle(c, s, key, label, false, o)
		}
		for _, part := range row.Parts() {
			pc := c.Key(fmt.Sprintf("part-%s-%d", part.Kind, part.Start))
			if part.Kind == "text" {
				o.prose(pc, row.Steps[part.Start].Text, part.Live)
				continue
			}
			if part.HiddenByFold && !row.Streaming && !open {
				continue
			}
			activity(pc, s, key, row.Steps, part, o)
		}
		if row.Error != "" {
			ui.Column(c).AlignSelf(ui.Stretch).Padding(8, 12).Border(1, o.Colors.Error.Alpha(.45)).Radius(8).Background(o.Colors.Error.Alpha(.1)).Children(func() {
				ui.Text(c, o.tr("transcript_view.agent_error", "Agent error")).Unselectable().FontWeight(600).TextColor(o.Colors.Text)
				ui.Text(c, row.Error).Selectable().TextColor(o.Colors.Muted)
			})
		}
		if row.Aborted && !row.Streaming {
			ui.Text(c, o.tr("transcript_view.stopped", "Stopped")).Unselectable().FontSize(o.FontSize - 1).TextColor(o.Colors.Muted)
		}
		if row.Streaming {
			ui.Row(c).Unselectable().AlignItems(ui.Center).Gap(8).Children(func() {
				elapsed := time.Duration(0)
				if row.Elapsed != nil {
					elapsed = *row.Elapsed
				}
				ui.Row(c).Gap(3).AlignItems(ui.Center).Children(func() {
					for i := 0; i < 3; i++ {
						color := o.Colors.Border
						if (int64(elapsed/time.Millisecond)/180+int64(i)*2)%6 < 3 {
							color = o.Colors.Accent
						}
						ui.Box(c).Size(4, 4).Radius(2).Background(color)
					}
				})
				label := o.tr("transcript.working", "Working")
				if row.Elapsed != nil {
					label = o.tr("transcript_view.working_for", "Working for %{duration}", "duration", workingElapsed(*row.Elapsed))
					c.After(180 * time.Millisecond)
				}
				if len(row.Steps) > 0 {
					if activity := workingActivity(row.Steps[len(row.Steps)-1], o); activity != "" {
						label = activity + " · " + label
					}
				}
				ui.Text(c, label).MinWidth(0).TextColor(o.Colors.Accent).FontSize(o.FontSize - .5).FontWeight(500)
			})
		} else if row.Text() != "" && !o.SuppressFooter {
			Footer(c, s, row, o)
		}
	})
}
func activity(c *ui.Context, s *State, rowKey string, all []Step, part Part, o Options) {
	key := fmt.Sprintf("%s:activity:%d", rowKey, part.Start)
	steps := all[part.Start:part.End]
	open := s.Open(key, part.Live)
	title := activityTitle(steps, part.Live, o)
	ui.Column(c).AlignSelf(ui.Stretch).MinWidth(0).Gap(6).Children(func() {
		b := ui.ButtonBase(c.Key("header")).AlignSelf(ui.Stretch).Padding(5, 8).Gap(5).Label(title)
		b.Children(func() {
			for _, icon := range activityIcons(steps) {
				o.icon(c, icon, 14, o.Colors.Muted)
			}
			ui.Text(c, title).MinWidth(0).FontSize(o.FontSize - 1).TextColor(o.Colors.Muted)
			ui.Spacer(c)
			icon := "icons/chevron-right.svg"
			if open {
				icon = "icons/chevron-down.svg"
			}
			o.icon(c, icon, 12, o.Colors.Muted)
		})
		if b.Clicked() {
			s.Toggle(key, part.Live)
		}
		if open {
			for i, step := range steps {
				sc := c.Key(fmt.Sprintf("step-%d", part.Start+i))
				if step.Thinking != "" {
					thoughtKey := fmt.Sprintf("%s:thought:%d", rowKey, part.Start+i)
					ui.Column(sc.Key("thought")).AlignSelf(ui.Stretch).MinWidth(0).Border(1, o.Colors.Border).Radius(8).Padding(8, 12).Children(func() {
						label := o.tr("transcript.thought", "Thought")
						if step.Streaming && step.Text == "" && len(step.Tools) == 0 {
							label = o.tr("transcript.thinking", "Thinking")
						}
						if toggle(sc, s, thoughtKey, label, true, o) {
							ui.Scroll(sc).AlignSelf(ui.Stretch).MaxHeight(260).Children(func() {
								ui.Text(sc, capChars(step.Thinking, 40000)).Selectable().Font(o.CodeFont).FontSize(o.CodeSize).TextColor(o.Colors.Muted)
							})
						}
					})
				}
				for j, t := range step.Tools {
					t.ID = rowKey + ":" + t.ID
					Tool(sc.Key(fmt.Sprintf("call-%d", j)), s, t, o)
				}
			}
		}
	})
}
func FormatTokens(n uint64) string {
	unit, suffix := uint64(1), ""
	if n >= 1000000 {
		unit, suffix = 1000000, "M"
	} else if n >= 1000 {
		unit, suffix = 1000, "K"
	} else {
		return fmt.Sprint(n)
	}
	whole, tenths := n/unit, ((n%unit)+(unit/20))/(unit/10)
	if tenths == 0 {
		return fmt.Sprintf("%d%s", whole, suffix)
	}
	if tenths == 10 {
		return fmt.Sprintf("%d%s", whole+1, suffix)
	}
	return fmt.Sprintf("%d.%d%s", whole, tenths, suffix)
}

func formatCost(cost float64) string {
	if math.IsNaN(cost) || math.IsInf(cost, 0) || cost <= 0 {
		return "$0.00"
	}
	if cost < .01 {
		return fmt.Sprintf("$%.4f", cost)
	}
	if cost < 100 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return fmt.Sprintf("$%.0f", cost)
}
func formatDuration(d time.Duration) string {
	sec := max(1, int64(d/time.Second))
	unit := func(n int64, u string) string {
		if n != 1 {
			u += "s"
		}
		return fmt.Sprintf("%d %s", n, u)
	}
	if sec < 60 {
		return unit(sec, "second")
	}
	if sec < 3600 {
		v := unit(sec/60, "minute")
		if sec%60 > 0 {
			v += " " + unit(sec%60, "second")
		}
		return v
	}
	v := unit(sec/3600, "hour")
	if (sec%3600)/60 > 0 {
		v += " " + unit((sec%3600)/60, "minute")
	}
	return v
}
func (r Row) UsageLabel() string { return usageLabel(r, Options{}) }

func Footer(c *ui.Context, s *State, row Row, o Options) ui.Element {
	o = defaults(c, o)
	c = c.Key("footer:" + row.ID)
	return ui.Row(c).AlignSelf(ui.Stretch).Unselectable().AlignItems(ui.Center).Gap(6).Children(func() {
		copyButton(c, s, "message:"+row.ID, row.Text(), o.tr("transcript.copy_message", "Copy Message"), o)
		if !row.FinishedAt.IsZero() {
			ui.Text(c, footerTimestamp(row, o)).FontSize(o.FontSize - 2).TextColor(o.Colors.Muted)
		}
		if usage := row.Usage(); usage != nil {
			label := usageLabel(row, o)
			anchor := ui.Text(c, label).FontSize(o.FontSize - 2).TextColor(o.Colors.Muted)
			if anchor.Hovered() {
				ui.Overlay(c, func() {
					ui.Column(c.Key("usage-breakdown")).Absolute().PassThrough().AttachTo(anchor, ui.AnchorTop, ui.AnchorBottom).Width(224).Padding(8, 12).Gap(6).Radius(8).Border(1, o.Colors.Border).Background(o.Colors.Surface).Children(func() {
						ui.Text(c, o.tr("transcript_view.message_usage", "Message usage")).FontSize(o.FontSize - 1).TextColor(o.Colors.Muted)
						metric := func(icon, label, value string) {
							ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(6).Children(func() {
								o.icon(c, icon, 12, o.Colors.Muted)
								ui.Text(c, label).Grow(1).TextColor(o.Colors.Muted).FontSize(o.FontSize - 1)
								ui.Text(c, value).TextColor(o.Colors.Text).FontSize(o.FontSize - 1)
							})
						}
						metric("icons/usage-input.svg", o.tr("transcript.input", "Input"), FormatTokens(usage.Input))
						metric("icons/usage-output.svg", o.tr("transcript.output", "Output"), FormatTokens(usage.Output))
						cache := FormatTokens(usage.CacheRead)
						if percent := usage.CachePercent(); percent != nil {
							cache = o.tr("transcript_view.hit_share", "%{tokens} · %{percent}% hit", "tokens", cache, "percent", hitPercent(*percent))
						}
						metric("icons/cache-read.svg", o.tr("transcript.cache_read", "Cache read"), cache)
						metric("icons/cache-write.svg", o.tr("transcript.cache_write", "Cache write"), FormatTokens(usage.CacheWrite))
						ui.Box(c).AlignSelf(ui.Stretch).Height(1).Background(o.Colors.Border)
						metric("icons/usage-total.svg", o.tr("transcript.total", "Total"), FormatTokens(usage.Total))
						if rate := generationRate(row); rate != "" {
							metric("icons/usage-output.svg", o.tr("transcript_view.rate", "Rate"), o.tr("transcript_view.tps_value", "%{rate} tok/s avg", "rate", rate))
						}
						cost := "—"
						if usage.Cost != nil {
							cost = formatCost(*usage.Cost)
						}
						metric("icons/usage-cost.svg", o.tr("usage.cost", "Cost"), cost)
					})
				})
			}
		}
	})
}

func footerTimestamp(row Row, o Options) string {
	if o.TimestampLabel != "" {
		return o.TimestampLabel
	}
	return row.FinishedAt.Local().Format("15:04")
}
func SummaryTimestamp(at, now time.Time, o Options) string {
	at = at.Local()
	now = now.In(at.Location())
	ay, am, ad := at.Date()
	ny, nm, nd := now.Date()
	a := time.Date(ay, am, ad, 12, 0, 0, 0, time.UTC)
	n := time.Date(ny, nm, nd, 12, 0, 0, 0, time.UTC)
	days := int(n.Sub(a) / (24 * time.Hour))
	switch days {
	case 0:
		return o.tr("transcript_view.today_clock", "Today %{clock}", "clock", at.Format("3:04 PM"))
	case 1:
		return o.tr("transcript_view.yesterday_clock", "Yesterday %{clock}", "clock", at.Format("3:04 PM"))
	}
	return at.Format("Jan 2")
}

// Tail owns the one final run footer after the changed-files card. Set
// Options.SuppressFooter on the last assistant's Turn when this is mounted.
func Tail(c *ui.Context, s *State, files []ChangedFile, last Row, o Options) ui.Element {
	o = defaults(c, o)
	return ui.Column(c.Key("tail-summary")).AlignSelf(ui.Stretch).Gap(10).Children(func() {
		Files(c, s, files, o)
		if !last.FinishedAt.IsZero() {
			o.TimestampLabel = SummaryTimestamp(last.FinishedAt, o.now(c), o)
		}
		o.SuppressFooter = false
		Footer(c, s, last, o)
	})
}

func workingElapsed(d time.Duration) string {
	s := int64(d / time.Second)
	if s < 60 {
		return fmt.Sprintf("%ds", s)
	}
	if s < 3600 {
		if s%60 == 0 {
			return fmt.Sprintf("%dm", s/60)
		}
		return fmt.Sprintf("%dm %ds", s/60, s%60)
	}
	if (s%3600)/60 == 0 {
		return fmt.Sprintf("%dh", s/3600)
	}
	return fmt.Sprintf("%dh %dm", s/3600, (s%3600)/60)
}
func workingActivity(step Step, o Options) string {
	if len(step.Tools) > 0 {
		t := step.Tools[len(step.Tools)-1]
		verb := ""
		switch {
		case in(t.Name, "bash shell terminal exec run"):
			verb = o.tr("transcript_view.verb_running", "Running")
		case in(t.Name, "read view"):
			verb = o.tr("transcript_view.verb_reading", "Reading")
		case in(t.Name, "grep find glob search"):
			verb = o.tr("transcript_view.verb_searching", "Searching")
		case in(t.Name, "edit write"):
			verb = o.tr("transcript_view.verb_editing", "Editing")
		default:
			verb = ActionLabel(t.Name)
			if t.Preview() == "" {
				return o.tr("transcript_view.using_tool", "Using %{action}", "action", verb)
			}
		}
		if p := t.Preview(); p != "" {
			verb += " " + p
		}
		return verb
	}
	if step.Thinking != "" && strings.TrimSpace(step.Text) == "" {
		return o.tr("transcript_view.thinking_ellipsis", "Thinking…")
	}
	return ""
}
func hitPercent(p float64) string {
	if math.IsNaN(p) || math.IsInf(p, 0) {
		return "0"
	}
	v := math.Round(max(0., min(100., p))*10) / 10
	if p < 100 && v >= 100 {
		v = 99.9
	}
	return strings.TrimSuffix(fmt.Sprintf("%.1f", v), ".0")
}
func generationRate(row Row) string {
	u, d := row.Usage(), row.GenerationTime()
	if u == nil || d == nil || *d < time.Second || u.Output == 0 {
		return ""
	}
	rate := float64(u.Output) / d.Seconds()
	if rate >= 10 {
		return fmt.Sprintf("%.0f", rate)
	}
	return fmt.Sprintf("%.1f", rate)
}
func usageLabel(row Row, o Options) string {
	u := row.Usage()
	if u == nil {
		return ""
	}
	label := "↑" + FormatTokens(u.Input) + " ↓" + FormatTokens(u.Output)
	if u.CacheRead > 0 {
		if p := u.CachePercent(); p != nil {
			label += o.tr("transcript_view.cached_share", " · %{percent}% cached", "percent", hitPercent(*p))
		}
	}
	if u.Cost != nil && *u.Cost > 0 {
		label += " · " + formatCost(*u.Cost)
	}
	if rate := generationRate(row); rate != "" {
		label += " · " + o.tr("transcript_view.tps_value", "%{rate} tok/s avg", "rate", rate)
	}
	return label
}
