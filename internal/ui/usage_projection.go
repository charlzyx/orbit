package orbitui

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"time"

	"github.com/charlzyx/orbit/internal/workbench"
)

type usageAggregate struct {
	Key, ID, Name, Path, Provider, Model, Workspace                  string
	Requests, Errors, Aborted, Timed, Priced, ToolFailures, ToolRuns int
	Tokens                                                           workbench.Tokens
	Cost                                                             float64
	Duration                                                         time.Duration
	At, Started, Ended                                               time.Time
	MaxDuration                                                      time.Duration
}
type usagePoint struct {
	At, End time.Time
	Latency usageLatency
	usageAggregate
}
type usageProjection struct {
	PeakPrompt                                          uint64
	ToolFailures                                        int
	ReasoningReported                                   bool
	Granularity                                         string
	SeriesTruncated                                     bool
	Previous                                            *workbench.UsageSummary
	Latency, ToolLatency                                usageLatency
	CachedRequests, CacheDataRequests                   int
	ModelRefOptions                                     []workbench.UsageModelRef
	Errors                                              []workbench.UsageError
	Summary                                             workbench.UsageSummary
	Records                                             []workbench.UsageRecord
	Tools                                               []workbench.ToolRun
	Sessions, Models, Providers, Workspaces, ToolGroups []usageAggregate
	Series, Daily, Calendar                             []usagePoint
	ProviderOptions, ModelOptions, WorkspaceOptions     []string
	P95                                                 time.Duration
	Timed                                               int
}

func usageAdd(a, b uint64) uint64 {
	if math.MaxUint64-a < b {
		return math.MaxUint64
	}
	return a + b
}

func addUsage(a *usageAggregate, r workbench.UsageRecord) {
	a.Requests++
	a.Tokens.Input = usageAdd(a.Tokens.Input, r.Tokens.Input)
	a.Tokens.Output = usageAdd(a.Tokens.Output, r.Tokens.Output)
	a.Tokens.CacheRead = usageAdd(a.Tokens.CacheRead, r.Tokens.CacheRead)
	a.Tokens.CacheWrite = usageAdd(a.Tokens.CacheWrite, r.Tokens.CacheWrite)
	a.Tokens.Reasoning = usageAdd(a.Tokens.Reasoning, r.Tokens.Reasoning)
	a.Tokens.Total = usageAdd(a.Tokens.Total, r.Tokens.Total)
	a.Cost += r.Cost
	if r.ModelPriced {
		a.Priced++
	}
	if r.Outcome == "error" {
		a.Errors++
	}
	if r.Outcome == "aborted" {
		a.Aborted++
	}
	if r.HasDuration {
		a.Timed++
		a.Duration += r.Duration
		a.MaxDuration = max(a.MaxDuration, r.Duration)
	}
	if r.At.After(a.At) {
		a.At = r.At
	}
}
func bucketStart(at time.Time, granularity string, location *time.Location) time.Time {
	at = at.In(location)
	day := time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, location)
	switch granularity {
	case "hour":
		return time.Date(at.Year(), at.Month(), at.Day(), at.Hour(), 0, 0, 0, location)
	case "week":
		return day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
	case "month":
		return time.Date(at.Year(), at.Month(), 1, 0, 0, 0, 0, location)
	default:
		return day
	}
}
func nextUsageBucket(at time.Time, granularity string) time.Time {
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
func buildUsageProjection(index workbench.UsageIndex, filter workbench.UsageFilter, granularity string, location *time.Location) usageProjection {
	if location == nil {
		location = time.Local
	}
	p := usageProjection{Granularity: granularity, Errors: index.FilteredErrors(filter), Summary: index.Summary(filter), Records: index.FilteredRecords(filter), Tools: index.FilteredTools(filter)}
	pricedModels := map[[2]string]bool{}
	for _, r := range index.Records {
		if r.ModelPriced || (r.Cost > 0 && !math.IsInf(r.Cost, 0) && !math.IsNaN(r.Cost)) {
			pricedModels[[2]string{r.Provider, r.Model}] = true
		}
	}
	modelRefs := map[workbench.UsageModelRef]bool{}
	providers, models, workspaces := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range index.Records {
		if r.Provider != "" {
			providers[r.Provider] = true
		}
		if r.Model != "" {
			models[r.Model] = true
			modelRefs[workbench.UsageModelRef{Provider: r.Provider, Model: r.Model}] = true
		}
		if r.Workspace != "" {
			workspaces[r.Workspace] = true
		}
	}
	for _, e := range index.Errors {
		if e.Provider != "" {
			providers[e.Provider] = true
		}
		if e.Model != "" {
			modelRefs[workbench.UsageModelRef{Provider: e.Provider, Model: e.Model}] = true
		}
		if e.Workspace != "" {
			workspaces[e.Workspace] = true
		}
	}
	for _, r := range index.Tools {
		if r.Provider != "" {
			providers[r.Provider] = true
		}
		if r.Model != "" {
			modelRefs[workbench.UsageModelRef{Provider: r.Provider, Model: r.Model}] = true
		}
		if r.Workspace != "" {
			workspaces[r.Workspace] = true
		}
	}
	for ref := range modelRefs {
		p.ModelRefOptions = append(p.ModelRefOptions, ref)
	}
	sort.Slice(p.ModelRefOptions, func(i, j int) bool {
		a, b := p.ModelRefOptions[i], p.ModelRefOptions[j]
		if a.Model != b.Model {
			return a.Model < b.Model
		}
		return a.Provider < b.Provider
	})
	for s := range providers {
		p.ProviderOptions = append(p.ProviderOptions, s)
	}
	for s := range models {
		p.ModelOptions = append(p.ModelOptions, s)
	}
	for s := range workspaces {
		p.WorkspaceOptions = append(p.WorkspaceOptions, s)
	}
	sort.Strings(p.ProviderOptions)
	sort.Strings(p.ModelOptions)
	sort.Strings(p.WorkspaceOptions)
	groups := []map[string]*usageAggregate{{}, {}, {}, {}}
	series, daily := map[time.Time]*usagePoint{}, map[time.Time]*usagePoint{}
	var durations, toolDurations []time.Duration
	bucketDurations := map[time.Time][]time.Duration{}
	sessionModels := map[string]map[workbench.UsageModelRef]uint64{}
	tableGranularity := granularity
	if tableGranularity == "hour" {
		tableGranularity = "day"
	}
	for _, r := range p.Records {
		p.PeakPrompt = max(p.PeakPrompt, usageAdd(usageAdd(r.Tokens.Input, r.Tokens.CacheRead), r.Tokens.CacheWrite))
		if r.HasReasoning || r.Tokens.Reasoning > 0 {
			p.ReasoningReported = true
		}
		r.ModelPriced = pricedModels[[2]string{r.Provider, r.Model}]
		sessionKey := r.SessionPath
		if sessionKey == "" {
			sessionKey = r.Session
		}
		if sessionModels[sessionKey] == nil {
			sessionModels[sessionKey] = map[workbench.UsageModelRef]uint64{}
		}
		ref := workbench.UsageModelRef{Provider: r.Provider, Model: r.Model}
		sessionModels[sessionKey][ref] = usageAdd(sessionModels[sessionKey][ref], r.Tokens.Total)
		if r.Tokens.CacheRead > 0 {
			p.CachedRequests++
		}
		if r.Tokens.CacheRead > 0 || r.Tokens.CacheWrite > 0 {
			p.CacheDataRequests++
		}
		keys := []string{sessionKey, r.Provider + "\x00" + r.Model, r.Provider, r.Workspace}
		for i, key := range keys {
			g := groups[i][key]
			if g == nil {
				g = &usageAggregate{Key: key, ID: r.Session, Name: key, Path: r.SessionPath, Provider: r.Provider, Model: r.Model, Workspace: r.Workspace}
				if i == 0 {
					g.Name = r.SessionTitle
					if g.Name == "" {
						g.Name = r.Session
					}
				}
				if i == 1 {
					g.Name = r.Model + " · " + r.Provider
				}
				if i == 3 {
					g.Name = filepath.Base(r.Workspace)
				}
				groups[i][key] = g
			}
			addUsage(g, r)
		}
		for _, spec := range []struct {
			m           map[time.Time]*usagePoint
			granularity string
		}{{series, granularity}, {daily, tableGranularity}} {
			at := bucketStart(r.At, spec.granularity, location)
			point := spec.m[at]
			if point == nil {
				point = &usagePoint{At: at, End: nextUsageBucket(at, spec.granularity)}
				spec.m[at] = point
			}
			addUsage(&point.usageAggregate, r)
		}
		if r.HasDuration {
			durations = append(durations, r.Duration)
			at := bucketStart(r.At, granularity, location)
			bucketDurations[at] = append(bucketDurations[at], r.Duration)
		}
	}
	toGroups := func(m map[string]*usageAggregate) []usageAggregate {
		out := make([]usageAggregate, 0, len(m))
		for _, r := range m {
			out = append(out, *r)
		}
		sort.Slice(out, func(i, j int) bool {
			if out[i].Tokens.Total != out[j].Tokens.Total {
				return out[i].Tokens.Total > out[j].Tokens.Total
			}
			if out[i].Requests != out[j].Requests {
				return out[i].Requests > out[j].Requests
			}
			if out[i].Name != out[j].Name {
				return out[i].Name < out[j].Name
			}
			return out[i].Key < out[j].Key
		})
		return out
	}
	metadata := map[string]workbench.UsageSession{}
	for _, v := range index.Sessions {
		key := v.Path
		if key == "" {
			key = v.ID
		}
		metadata[key] = v
	}
	for key, g := range groups[0] {
		if m, ok := metadata[key]; ok {
			g.ID = m.ID
			g.Started, g.Ended = m.Started, m.Ended
		}
		var maxTokens uint64
		first := true
		for ref, tokens := range sessionModels[key] {
			if first || tokens > maxTokens || (tokens == maxTokens && ref.Provider+"/"+ref.Model > g.Provider+"/"+g.Model) {
				g.Provider, g.Model = ref.Provider, ref.Model
				maxTokens = tokens
				first = false
			}
		}
	}
	p.Sessions, p.Models, p.Providers, p.Workspaces = toGroups(groups[0]), toGroups(groups[1]), toGroups(groups[2]), toGroups(groups[3])
	toolGroups := map[string]*usageAggregate{}
	for _, r := range p.Tools {
		g := toolGroups[r.Name]
		if g == nil {
			g = &usageAggregate{Key: r.Name, Name: r.Name}
			toolGroups[r.Name] = g
		}
		g.Requests++
		if r.HasResult && !r.Success {
			g.ToolFailures++
			p.ToolFailures++
		}
		if r.HasDuration {
			g.Timed++
			g.Duration += r.Duration
			g.MaxDuration = max(g.MaxDuration, r.Duration)
			toolDurations = append(toolDurations, r.Duration)
		}
	}
	sessionPositions := map[string]int{}
	for i := range p.Sessions {
		sessionPositions[p.Sessions[i].Key] = i
	}
	for _, tool := range p.Tools {
		key := tool.SessionPath
		if key == "" {
			key = tool.Session
		}
		if i, ok := sessionPositions[key]; ok {
			g := &p.Sessions[i]
			g.ToolRuns++
			if tool.HasResult && !tool.Success {
				g.ToolFailures++
			}
		}
	}
	p.ToolGroups = toGroups(toolGroups)
	sort.SliceStable(p.ToolGroups, func(i, j int) bool { return p.ToolGroups[i].Requests > p.ToolGroups[j].Requests })
	toPoints := func(m map[time.Time]*usagePoint) []usagePoint {
		out := make([]usagePoint, 0, len(m))
		for _, r := range m {
			out = append(out, *r)
		}
		sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
		return out
	}
	for at, point := range series {
		point.Latency = usageLatencyFrom(bucketDurations[at])
	}
	p.Series, p.Daily = toPoints(series), toPoints(daily)
	start, end := filter.Start, filter.End
	if len(p.Series) > 0 {
		if start.IsZero() || start.UnixMilli() <= 0 {
			start = p.Series[0].At
			for _, r := range index.Records {
				if r.At.Before(start) {
					start = bucketStart(r.At, granularity, location)
				}
			}
		}
		if end.IsZero() {
			end = p.Series[len(p.Series)-1].End
		}
	}
	if !start.IsZero() && !end.IsZero() && start.Before(end) {
		p.Series = nil
		at := bucketStart(start, granularity, location)
		for len(p.Series) < 4000 && at.Before(end) {
			point := series[at]
			if point == nil {
				point = &usagePoint{At: at, End: nextUsageBucket(at, granularity)}
			}
			p.Series = append(p.Series, *point)
			at = nextUsageBucket(at, granularity)
		}
		p.SeriesTruncated = at.Before(end)
	}
	calendarFilter := filter
	calendarFilter.Start, calendarFilter.End = time.Time{}, time.Time{}
	calendarFilter.FocusStart, calendarFilter.FocusEnd = time.Time{}, time.Time{}
	calendar := map[time.Time]*usagePoint{}
	for _, r := range index.FilteredRecords(calendarFilter) {
		r.ModelPriced = pricedModels[[2]string{r.Provider, r.Model}]
		at := bucketStart(r.At, "day", location)
		point := calendar[at]
		if point == nil {
			point = &usagePoint{At: at, End: nextUsageBucket(at, "day")}
			calendar[at] = point
		}
		addUsage(&point.usageAggregate, r)
	}
	p.Calendar = toPoints(calendar)
	sort.Slice(p.Records, func(i, j int) bool { return p.Records[i].At.After(p.Records[j].At) })
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	p.Timed = len(durations)
	if len(durations) > 0 {
		p.P95 = durations[min(len(durations)-1, int(math.Ceil(float64(len(durations))*.95))-1)]
	}
	p.Latency = usageLatencyFrom(durations)
	p.ToolLatency = usageLatencyFrom(toolDurations)
	if !filter.Start.IsZero() && filter.Start.UnixMilli() > 0 && filter.End.After(filter.Start) {
		earlier := false
		for _, r := range index.Records {
			if r.At.Before(filter.Start) {
				earlier = true
				break
			}
		}
		if earlier {
			previous := filter
			previous.End = filter.Start
			previous.Start = filter.Start.Add(-filter.End.Sub(filter.Start))
			previous.FocusStart, previous.FocusEnd = time.Time{}, time.Time{}
			summary := index.Summary(previous)
			p.Previous = &summary
		}
	}
	return p
}

func usageMetricValue(point usageAggregate, metric string) (float64, bool) {
	switch metric {
	case "requests":
		return float64(point.Requests), true
	case "input":
		return float64(point.Tokens.Input), true
	case "output":
		return float64(point.Tokens.Output), true
	case "cache":
		return float64(usageAdd(point.Tokens.CacheRead, point.Tokens.CacheWrite)), true
	case "cost":
		return point.Cost, point.Priced > 0
	case "latency":
		if point.Timed > 0 {
			return point.Duration.Seconds() / float64(point.Timed), true
		}
		return 0, false
	case "errors":
		return float64(point.Errors), true
	default:
		return float64(point.Tokens.Total), true
	}
}

func sortUsageRows(rows []usageAggregate, field string, descending bool) {
	less := func(a, b usageAggregate) int {
		switch field {
		case "name":
			if a.Name < b.Name {
				return -1
			}
			if a.Name > b.Name {
				return 1
			}
		case "requests":
			if a.Requests < b.Requests {
				return -1
			}
			if a.Requests > b.Requests {
				return 1
			}
		case "cost":
			if a.Cost < b.Cost {
				return -1
			}
			if a.Cost > b.Cost {
				return 1
			}
		case "latest":
			if a.At.Before(b.At) {
				return -1
			}
			if a.At.After(b.At) {
				return 1
			}
		default:
			if a.Tokens.Total < b.Tokens.Total {
				return -1
			}
			if a.Tokens.Total > b.Tokens.Total {
				return 1
			}
		}
		return 0
	}
	sort.SliceStable(rows, func(i, j int) bool {
		n := less(rows[i], rows[j])
		if n == 0 {
			return rows[i].Key < rows[j].Key
		}
		if descending {
			return n > 0
		}
		return n < 0
	})
}

func usageLatencyHint(p95 time.Duration, count int) string {
	if count == 0 {
		return "No timed responses"
	}
	if count < 20 {
		return fmt.Sprintf("%d measured · p95 needs 20 samples", count)
	}
	return fmt.Sprintf("p95 %s · %d measured", p95.Round(time.Millisecond), count)
}

func usageCalendarBounds(index workbench.UsageIndex, location *time.Location) (grid, first, last time.Time) {
	if location == nil {
		location = time.Local
	}
	end := index.ScannedAt
	if end.IsZero() {
		for _, r := range index.Records {
			if end.IsZero() || r.At.After(end) {
				end = r.At
			}
		}
	}
	if end.IsZero() {
		end = time.Unix(0, 0)
	}
	last = bucketStart(end, "day", location)
	first = last.Add(-365 * 24 * time.Hour)
	localFirst := bucketStart(first, "day", location)
	grid = localFirst.AddDate(0, 0, -(int(localFirst.Weekday())+6)%7)
	return
}
