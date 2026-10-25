package workbench

import (
	"encoding/json"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

// UsageExportOptions identifies the view without persisting transient scopes.
// Insights contains already-computed on-screen insights, never invented text.
type UsageExportOptions struct {
	Preset, Granularity, SessionSearch string
	Location                           *time.Location
	Now                                time.Time
	Insights                           []string
}

func usageProviderLabel(id string) string {
	if id == "" {
		return "unknown"
	}
	if id == "chatgpt" {
		return "ChatGPT"
	}
	r := []rune(strings.NewReplacer("_", " ", "-", " ").Replace(id))
	if len(r) > 0 {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}
func usageModelLabel(id string) string {
	if _, rest, ok := strings.Cut(id, "/"); ok && rest != "" {
		return rest
	}
	return id
}
func UsageToolClass(tool string) string {
	switch tool {
	case "bash":
		return "terminal"
	case "read":
		return "read"
	case "edit", "write":
		return "edit"
	case "grep", "find", "ls":
		return "search"
	case "web_fetch", "web_search":
		return "web"
	case "todo", "plan_complete", "artifact":
		return "plan"
	case "ask_user_question", "ask_question", "ask_user":
		return "ask"
	default:
		return "other"
	}
}
func usageBucketStart(at time.Time, granularity string, loc *time.Location) time.Time {
	at = at.In(loc)
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, loc)
	switch granularity {
	case "hour":
		return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), 0, 0, 0, loc)
	case "week":
		return day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
	case "month":
		return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, loc)
	default:
		return day
	}
}
func usageNextBucket(at time.Time, granularity string) time.Time {
	switch granularity {
	case "hour":
		return at.Add(time.Hour)
	case "week":
		return at.AddDate(0, 0, 7)
	case "month":
		return at.AddDate(0, 1, 0)
	default:
		return at.AddDate(0, 0, 1)
	}
}
func usageBucketLabel(at time.Time, granularity string) string {
	switch granularity {
	case "hour":
		return at.Format("Jan 2, 15:04")
	case "month":
		return at.Format("January 2006")
	case "week":
		return "Week of " + at.Format("Jan 2, 2006")
	default:
		return at.Format("Jan 2, 2006")
	}
}
func usageCacheRate(t Tokens) any {
	if t.CacheRead == 0 && t.CacheWrite == 0 {
		return nil
	}
	if float64(t.Input)+float64(t.CacheRead)+float64(t.CacheWrite) == 0 {
		return nil
	}
	return t.CacheRate()
}
func usageAverage(s UsageSummary) any {
	if s.TimedRequests == 0 {
		return nil
	}
	return float64(s.Duration) / float64(time.Millisecond) / float64(s.TimedRequests)
}
func usageKey(session, path string) string {
	if path != "" {
		return path
	}
	return session
}
func usageTotals(records []UsageRecord) UsageSummary {
	return (UsageIndex{Records: records}).Summary(UsageFilter{})
}
func usagePointJSON(at time.Time, s UsageSummary, granularity string) map[string]any {
	return map[string]any{"start_ms": at.UnixMilli(), "label": usageBucketLabel(at, granularity), "requests": s.Requests, "input": s.Tokens.Input, "output": s.Tokens.Output, "cache_read": s.Tokens.CacheRead, "cache_write": s.Tokens.CacheWrite, "total": s.Tokens.Total, "errors": s.Errors, "cost_usd": s.Cost, "avg_duration_ms": usageAverage(s)}
}

// ExportJSON follows the baseline aggregate export schema. All rows obey the
// shared scope; calendar intentionally ignores the selected date/focus window.
func (index UsageIndex) ExportJSON(w io.Writer, filter UsageFilter, options UsageExportOptions) error {
	if options.Location == nil {
		options.Location = time.Local
	}
	if options.Now.IsZero() {
		options.Now = time.Now()
	}
	switch options.Granularity {
	case "hour", "day", "week", "month":
	default:
		options.Granularity = "day"
	}
	if options.Preset == "" {
		options.Preset = "custom"
	}
	// Mark a detached copy: exports must not race live UI reads of the scanner's
	// immutable snapshot when a test or caller constructed records directly.
	index.Records = append([]UsageRecord(nil), index.Records...)
	index.markPricing()
	records := index.FilteredRecords(filter)
	tools := index.FilteredTools(filter)
	summary := index.Summary(filter)
	groups := []map[string][]UsageRecord{{}, {}, {}, {}}
	for _, r := range records {
		keys := []string{usageKey(r.Session, r.SessionPath), r.Provider + "\x00" + r.Model, r.Provider, r.Workspace}
		for n, key := range keys {
			groups[n][key] = append(groups[n][key], r)
		}
	}
	breakdown := func(n int) []map[string]any {
		out := make([]map[string]any, 0, len(groups[n]))
		for key, rows := range groups[n] {
			s := usageTotals(rows)
			label, id := key, key
			var sub any
			switch n {
			case 1:
				id = rows[0].Model
				label = usageModelLabel(id)
				sub = usageProviderLabel(rows[0].Provider)
			case 2:
				label = usageProviderLabel(id)
			case 3:
				label = filepath.Base(id)
				sub = filepath.Dir(id)
			}
			share := float64(0)
			if summary.Tokens.Total > 0 {
				share = float64(s.Tokens.Total) / float64(summary.Tokens.Total)
			}
			out = append(out, map[string]any{"id": id, "label": label, "sub": sub, "requests": s.Requests, "tokens": s.Tokens.Total, "input": s.Tokens.Input, "output": s.Tokens.Output, "cache_read": s.Tokens.CacheRead, "cache_write": s.Tokens.CacheWrite, "cost_usd": s.Cost, "share": share})
		}
		sort.Slice(out, func(a, b int) bool {
			aa, bb := out[a]["tokens"].(uint64), out[b]["tokens"].(uint64)
			if aa != bb {
				return aa > bb
			}
			return out[a]["label"].(string) < out[b]["label"].(string)
		})
		return out
	}
	metadata := map[string]UsageSession{}
	for _, s := range index.Sessions {
		metadata[usageKey(s.ID, s.Path)] = s
	}
	perSessionTools := map[string][2]int{}
	toolErrors := 0
	for _, r := range tools {
		key := usageKey(r.Session, r.SessionPath)
		n := perSessionTools[key]
		n[0]++
		if r.HasResult && !r.Success {
			n[1]++
			toolErrors++
		}
		perSessionTools[key] = n
	}
	sessions := make([]map[string]any, 0, len(groups[0]))
	needle := strings.ToLower(strings.TrimSpace(options.SessionSearch))
	for key, rows := range groups[0] {
		s := usageTotals(rows)
		first := rows[0]
		meta, hasMeta := metadata[key]
		title, workspace := first.SessionTitle, first.Workspace
		var started, ended any
		if hasMeta {
			title, workspace = meta.Title, meta.Workspace
			started, ended = meta.Started.UnixMilli(), meta.Ended.UnixMilli()
		}
		models := map[[2]string]uint64{}
		for _, r := range rows {
			pair := [2]string{r.Provider, r.Model}
			models[pair] = addCounts(models[pair], r.Tokens.Total)
		}
		top := [2]string{}
		var max uint64
		for pair, tokens := range models {
			if tokens > max || (tokens == max && pair[1] > top[1]) {
				top, max = pair, tokens
			}
		}
		topModel := usageModelLabel(top[1]) + " · " + usageProviderLabel(top[0])
		if needle != "" && !strings.Contains(strings.ToLower(title+" "+workspace+" "+topModel), needle) {
			continue
		}
		counts := perSessionTools[key]
		sessions = append(sessions, map[string]any{"session_id": first.Session, "title": title, "workspace": filepath.Base(workspace), "top_model": topModel, "models": len(models), "started_ms": started, "ended_ms": ended, "requests": s.Requests, "input": s.Tokens.Input, "output": s.Tokens.Output, "cache_read": s.Tokens.CacheRead, "cache_write": s.Tokens.CacheWrite, "total": s.Tokens.Total, "cost_usd": s.Cost, "errors": s.Errors, "tool_calls": counts[0], "tool_errors": counts[1]})
	}
	sort.Slice(sessions, func(a, b int) bool {
		aa, bb := sessions[a]["total"].(uint64), sessions[b]["total"].(uint64)
		if aa != bb {
			return aa > bb
		}
		return sessions[a]["session_id"].(string) < sessions[b]["session_id"].(string)
	})
	bucketRecords := map[time.Time][]UsageRecord{}
	for _, r := range records {
		at := usageBucketStart(r.At, options.Granularity, options.Location)
		bucketRecords[at] = append(bucketRecords[at], r)
	}
	start, end := filter.Start, filter.End
	if end.IsZero() {
		end = options.Now
	}
	if start.IsZero() || options.Preset == "all" {
		start = time.Time{}
		for _, r := range index.Records {
			if start.IsZero() || r.At.Before(start) {
				start = r.At
			}
		}
	}
	if start.IsZero() {
		start = end
	}
	series := make([]map[string]any, 0)
	for at, n := usageBucketStart(start, options.Granularity, options.Location), 0; at.Before(end) && n < 4000; at, n = usageNextBucket(at, options.Granularity), n+1 {
		series = append(series, usagePointJSON(at, usageTotals(bucketRecords[at]), options.Granularity))
	}
	tableGranularity := options.Granularity
	if tableGranularity == "hour" {
		tableGranularity = "day"
	}
	table := map[time.Time][]UsageRecord{}
	for _, r := range records {
		at := usageBucketStart(r.At, tableGranularity, options.Location)
		table[at] = append(table[at], r)
	}
	buckets := make([]map[string]any, 0, len(table))
	for at, rows := range table {
		s := usageTotals(rows)
		buckets = append(buckets, map[string]any{"start_ms": at.UnixMilli(), "label": usageBucketLabel(at, tableGranularity), "requests": s.Requests, "tokens": s.Tokens.Total, "input": s.Tokens.Input, "output": s.Tokens.Output, "cache_hit_rate": usageCacheRate(s.Tokens), "errors": s.Errors})
	}
	sort.Slice(buckets, func(a, b int) bool { return buckets[a]["start_ms"].(int64) > buckets[b]["start_ms"].(int64) })
	calendarEnd := index.ScannedAt
	if calendarEnd.IsZero() {
		for _, r := range index.Records {
			if r.At.After(calendarEnd) {
				calendarEnd = r.At
			}
		}
	}
	if calendarEnd.IsZero() {
		calendarEnd = time.UnixMilli(0)
	}
	lastDay := usageBucketStart(calendarEnd, "day", options.Location)
	windowStart := lastDay.Add(-365 * 24 * time.Hour)
	firstDay := usageBucketStart(windowStart, "day", options.Location)
	calendarFilter := filter
	calendarFilter.Start, calendarFilter.End = windowStart, lastDay.AddDate(0, 0, 1)
	calendarFilter.FocusStart, calendarFilter.FocusEnd = time.Time{}, time.Time{}
	calendarRecords := map[time.Time][]UsageRecord{}
	for _, r := range index.FilteredRecords(calendarFilter) {
		at := usageBucketStart(r.At, "day", options.Location)
		calendarRecords[at] = append(calendarRecords[at], r)
	}
	calendar := make([]map[string]any, 0, 366)
	for at := firstDay; !at.After(lastDay) && len(calendar) < 380; at = at.AddDate(0, 0, 1) {
		// Source export excludes padding cells whose local midnight falls
		// before the raw fixed-duration window, including a DST boundary.
		if at.Before(windowStart) {
			continue
		}
		point := usagePointJSON(at, usageTotals(calendarRecords[at]), "day")
		delete(point, "label")
		delete(point, "avg_duration_ms")
		calendar = append(calendar, point)
	}
	errorsDetail := make([]map[string]any, 0)
	for _, r := range index.FilteredErrors(filter) {
		if len(errorsDetail) >= 400 {
			break
		}
		errorsDetail = append(errorsDetail, map[string]any{"timestamp_ms": r.At.UnixMilli(), "session_id": r.Session, "kind": r.Kind, "model": usageModelLabel(r.Model), "message": r.Message})
	}
	type toolTotals struct {
		calls, errors, timed int
		duration, max        time.Duration
	}
	toolGroups := map[string]toolTotals{}
	for _, r := range tools {
		s := toolGroups[r.Name]
		s.calls++
		if r.HasResult && !r.Success {
			s.errors++
		}
		if r.HasDuration {
			s.timed++
			s.duration += r.Duration
			if r.Duration > s.max {
				s.max = r.Duration
			}
		}
		toolGroups[r.Name] = s
	}
	toolRows := make([]map[string]any, 0, len(toolGroups))
	for name, s := range toolGroups {
		var average any
		if s.timed > 0 {
			average = float64(s.duration) / float64(time.Millisecond) / float64(s.timed)
		}
		toolRows = append(toolRows, map[string]any{"tool": name, "class": UsageToolClass(name), "calls": s.calls, "errors": s.errors, "avg_duration_ms": average, "max_duration_ms": s.max.Milliseconds()})
	}
	sort.Slice(toolRows, func(a, b int) bool {
		aa, bb := toolRows[a]["calls"].(int), toolRows[b]["calls"].(int)
		if aa != bb {
			return aa > bb
		}
		return toolRows[a]["tool"].(string) < toolRows[b]["tool"].(string)
	})
	var previous any
	if options.Preset != "all" && !filter.Start.IsZero() && filter.End.After(filter.Start) {
		hasHistory := false
		for _, r := range index.Records {
			if r.At.Before(filter.Start) {
				hasHistory = true
				break
			}
		}
		if hasHistory {
			f := filter
			f.Start, f.End = filter.Start.Add(-filter.End.Sub(filter.Start)), filter.Start
			f.FocusStart, f.FocusEnd = time.Time{}, time.Time{}
			s := index.Summary(f)
			previous = map[string]any{"requests": s.Requests, "total": s.Tokens.Total, "input": s.Tokens.Input, "output": s.Tokens.Output, "cache_read": s.Tokens.CacheRead, "cache_write": s.Tokens.CacheWrite, "cost_usd": s.Cost, "errors": s.Errors, "avg_duration_ms": usageAverage(s)}
		}
	}
	insights := options.Insights
	if insights == nil {
		insights = []string{}
	}
	payload := map[string]any{"generated_at_ms": options.Now.UnixMilli(), "range": map[string]any{"preset": options.Preset, "label": options.Preset, "start_ms": usageRangeStart(filter.Start), "end_ms": end.UnixMilli(), "granularity": options.Granularity}, "summary": map[string]any{"requests": summary.Requests, "turns": summary.Turns, "sessions": len(groups[0]), "models": len(groups[1]), "providers": len(groups[2]), "workspaces": len(groups[3]), "input": summary.Tokens.Input, "output": summary.Tokens.Output, "cache_read": summary.Tokens.CacheRead, "cache_write": summary.Tokens.CacheWrite, "total": summary.Tokens.Total, "reasoning": summary.Tokens.Reasoning, "cache_hit_rate": usageCacheRate(summary.Tokens), "errors": summary.Errors, "aborted": summary.Aborted, "avg_duration_ms": usageAverage(summary), "duration_samples": summary.TimedRequests, "tool_calls": summary.ToolRuns, "bash_calls": summary.BashRuns, "tool_errors": toolErrors, "cost_usd": summary.Cost, "priced_requests": summary.PricedRequests}, "previous": previous, "series": series, "models": breakdown(1), "providers": breakdown(2), "workspaces": breakdown(3), "sessions": sessions, "buckets": buckets, "calendar": calendar, "tools": toolRows, "errors_detail": errorsDetail, "insights": insights}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}

func usageRangeStart(start time.Time) int64 {
	if start.IsZero() {
		return 0
	}
	return start.UnixMilli()
}
