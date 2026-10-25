package workbench

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Tokens struct{ Input, Output, CacheRead, CacheWrite, Reasoning, Total uint64 }

func (t Tokens) CacheRate() float64 {
	den := t.Input + t.CacheRead + t.CacheWrite
	if den == 0 {
		return 0
	}
	return float64(t.CacheRead) * 100 / float64(den)
}

type UsageRecord struct {
	At                                                  time.Time
	Session, Workspace, Provider, Model, Outcome, Error string
	Tokens                                              Tokens
	Cost                                                float64
	Duration                                            time.Duration
	HasDuration                                         bool
}
type ToolRun struct {
	At                           time.Time
	Session, Workspace, Name, ID string
	Success                      bool
	Duration                     time.Duration
	HasDuration                  bool
}
type UsageTurn struct {
	At                 time.Time
	Session, Workspace string
}
type UsageIndex struct {
	Records                []UsageRecord
	Tools                  []ToolRun
	Turns                  []UsageTurn
	Warnings               []string
	Files, Parsed, Skipped int
}
type usageFile struct {
	size     int64
	modified time.Time
	index    UsageIndex
}
type UsageScanner struct {
	Root  string
	mu    sync.Mutex
	cache map[string]usageFile
}

func NewUsageScanner(root string) *UsageScanner {
	return &UsageScanner{Root: root, cache: map[string]usageFile{}}
}
func (s *UsageScanner) Scan(ctx context.Context) (UsageIndex, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out UsageIndex
	seen := map[string]bool{}
	e := filepath.WalkDir(s.Root, func(path string, d fs.DirEntry, e error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if e != nil {
			if errors.Is(e, fs.ErrNotExist) {
				return nil
			}
			out.Warnings = append(out.Warnings, e.Error())
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".jsonl") {
			return nil
		}
		st, e := d.Info()
		if e != nil {
			return nil
		}
		out.Files++
		seen[path] = true
		if st.Size() > 256<<20 {
			out.Skipped++
			out.Warnings = append(out.Warnings, filepath.Base(path)+": session exceeds256MiB")
			return nil
		}
		f, ok := s.cache[path]
		if !ok || f.size != st.Size() || !f.modified.Equal(st.ModTime()) {
			index, e := readUsageFile(ctx, path)
			if e != nil {
				out.Warnings = append(out.Warnings, filepath.Base(path)+": "+e.Error())
				out.Skipped++
				return nil
			}
			f = usageFile{st.Size(), st.ModTime(), index}
			s.cache[path] = f
			out.Parsed++
		}
		out.Records = append(out.Records, f.index.Records...)
		out.Tools = append(out.Tools, f.index.Tools...)
		out.Turns = append(out.Turns, f.index.Turns...)
		out.Warnings = append(out.Warnings, f.index.Warnings...)
		return nil
	})
	for p := range s.cache {
		if !seen[p] {
			delete(s.cache, p)
		}
	}
	sort.Slice(out.Records, func(i, j int) bool { return out.Records[i].At.Before(out.Records[j].At) })
	return out, e
}
func ts(v any) time.Time {
	switch x := v.(type) {
	case string:
		t, _ := time.Parse(time.RFC3339Nano, x)
		return t
	case float64:
		return time.UnixMilli(int64(x))
	default:
		return time.Time{}
	}
}
func number(v any) float64 { f, _ := v.(float64); return f }
func count(v any) uint64 {
	f := number(v)
	if f <= 0 {
		return 0
	}
	return uint64(f)
}
func readUsageFile(ctx context.Context, path string) (UsageIndex, error) {
	f, e := os.Open(path)
	if e != nil {
		return UsageIndex{}, e
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<10), 64<<20)
	var out UsageIndex
	session, workspace := filepath.Base(path), ""
	var previous time.Time
	type call struct {
		name string
		at   time.Time
	}
	calls := map[string]call{}
	for scanner.Scan() {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		var entry map[string]any
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		at := ts(entry["timestamp"])
		kind := stringValue(entry["type"])
		if kind == "session" {
			session = stringValue(entry["id"])
			workspace = stringValue(entry["cwd"])
			continue
		}
		message, ok := entry["message"].(map[string]any)
		if !ok {
			continue
		}
		if at.IsZero() {
			at = ts(message["timestamp"])
		}
		role := stringValue(message["role"])
		switch role {
		case "user":
			out.Turns = append(out.Turns, UsageTurn{at, session, workspace})
		case "assistant":
			content, _ := message["content"].([]any)
			for _, raw := range content {
				if b, ok := raw.(map[string]any); ok && stringValue(b["type"]) == "toolCall" {
					calls[stringValue(b["id"])] = call{stringValue(b["name"]), at}
				}
			}
			if usage, ok := message["usage"].(map[string]any); ok {
				tokens := Tokens{Input: count(usage["input"]), Output: count(usage["output"]), CacheRead: count(usage["cacheRead"]), CacheWrite: count(usage["cacheWrite"]), Reasoning: count(usage["reasoning"]), Total: count(usage["totalTokens"])}
				if _, ok := usage["totalTokens"]; !ok {
					tokens.Total = tokens.Input + tokens.Output + tokens.CacheRead + tokens.CacheWrite
				}
				cost := 0.
				if c, ok := usage["cost"].(map[string]any); ok {
					cost = number(c["total"])
				}
				if tokens.Total > 0 || cost > 0 {
					r := UsageRecord{At: at, Session: session, Workspace: workspace, Provider: stringValue(message["provider"]), Model: stringValue(message["model"]), Outcome: stringValue(message["stopReason"]), Error: stringValue(message["errorMessage"]), Tokens: tokens, Cost: cost}
					delta := at.Sub(previous)
					if !previous.IsZero() && delta >= 0 && delta <= 30*time.Minute {
						r.Duration = delta
						r.HasDuration = true
					}
					out.Records = append(out.Records, r)
				}
			}
		case "toolResult":
			id := stringValue(message["toolCallId"])
			r := ToolRun{At: at, Session: session, Workspace: workspace, ID: id, Name: stringValue(message["toolName"]), Success: message["isError"] != true}
			if c, ok := calls[id]; ok {
				if r.Name == "" {
					r.Name = c.name
				}
				delta := at.Sub(c.at)
				if !c.at.IsZero() && delta >= 0 && delta <= 30*time.Minute {
					r.Duration = delta
					r.HasDuration = true
				}
				delete(calls, id)
			}
			out.Tools = append(out.Tools, r)
		}
		if !at.IsZero() {
			previous = at
		}
	}
	return out, scanner.Err()
}

type UsageFilter struct {
	Start, End                                  time.Time
	Workspace, Provider, Model, Outcome, Search string
}

func (f UsageFilter) Matches(r UsageRecord) bool {
	return (f.Start.IsZero() || !r.At.Before(f.Start)) && (f.End.IsZero() || r.At.Before(f.End)) && (f.Workspace == "" || r.Workspace == f.Workspace) && (f.Provider == "" || r.Provider == f.Provider) && (f.Model == "" || r.Model == f.Model) && (f.Outcome == "" || r.Outcome == f.Outcome) && (f.Search == "" || strings.Contains(strings.ToLower(r.Session+" "+r.Workspace+" "+r.Provider+" "+r.Model+" "+r.Error), strings.ToLower(f.Search)))
}

type UsageSummary struct {
	Requests, Turns, ToolRuns, BashRuns, Errors, Aborted, PricedRequests int
	Tokens                                                               Tokens
	Cost                                                                 float64
	Duration                                                             time.Duration
	TimedRequests                                                        int
}

func (i UsageIndex) Summary(f UsageFilter) UsageSummary {
	var s UsageSummary
	for _, r := range i.Records {
		if !f.Matches(r) {
			continue
		}
		s.Requests++
		s.Tokens.Input += r.Tokens.Input
		s.Tokens.Output += r.Tokens.Output
		s.Tokens.CacheRead += r.Tokens.CacheRead
		s.Tokens.CacheWrite += r.Tokens.CacheWrite
		s.Tokens.Reasoning += r.Tokens.Reasoning
		s.Tokens.Total += r.Tokens.Total
		s.Cost += r.Cost
		if r.Cost > 0 {
			s.PricedRequests++
		}
		if r.Outcome == "error" {
			s.Errors++
		}
		if r.Outcome == "aborted" {
			s.Aborted++
		}
		if r.HasDuration {
			s.Duration += r.Duration
			s.TimedRequests++
		}
	}
	for _, r := range i.Turns {
		if f.Matches(UsageRecord{At: r.At, Workspace: r.Workspace, Session: r.Session}) {
			s.Turns++
		}
	}
	for _, r := range i.Tools {
		if f.Matches(UsageRecord{At: r.At, Workspace: r.Workspace, Session: r.Session}) {
			s.ToolRuns++
			if r.Name == "bash" {
				s.BashRuns++
			}
		}
	}
	return s
}

type UsageBucket struct {
	At       time.Time
	Requests int
	Tokens   uint64
	Cost     float64
}

func (i UsageIndex) Series(f UsageFilter, granularity string, loc *time.Location) []UsageBucket {
	if loc == nil {
		loc = time.Local
	}
	m := map[time.Time]UsageBucket{}
	for _, r := range i.Records {
		if !f.Matches(r) {
			continue
		}
		t := r.At.In(loc)
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		switch granularity {
		case "hour":
			day = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc)
		case "week":
			day = day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
		case "month":
			day = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
		b := m[day]
		b.At = day
		b.Requests++
		b.Tokens += r.Tokens.Total
		b.Cost += r.Cost
		m[day] = b
	}
	out := make([]UsageBucket, 0, len(m))
	for _, b := range m {
		out = append(out, b)
	}
	sort.Slice(out, func(a, b int) bool { return out[a].At.Before(out[b].At) })
	return out
}
func (i UsageIndex) ExportCSV(w io.Writer, f UsageFilter) error {
	c := csv.NewWriter(w)
	c.Write([]string{"timestamp", "session", "workspace", "provider", "model", "input", "output", "cache_read", "cache_write", "total", "cost_usd", "outcome"})
	for _, r := range i.Records {
		if !f.Matches(r) {
			continue
		}
		c.Write([]string{r.At.Format(time.RFC3339), csvSafe(r.Session), csvSafe(r.Workspace), csvSafe(r.Provider), csvSafe(r.Model), strconv.FormatUint(r.Tokens.Input, 10), strconv.FormatUint(r.Tokens.Output, 10), strconv.FormatUint(r.Tokens.CacheRead, 10), strconv.FormatUint(r.Tokens.CacheWrite, 10), strconv.FormatUint(r.Tokens.Total, 10), fmt.Sprintf("%.8f", r.Cost), csvSafe(r.Outcome)})
	}
	c.Flush()
	return c.Error()
}
func csvSafe(s string) string {
	if strings.ContainsAny(s[:min(1, len(s))], "=+-@\t\r") {
		return "'" + s
	}
	return s
}
func UsageRange(preset string, now time.Time) (time.Time, time.Time) {
	d := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch preset {
	case "today":
		return d, d.AddDate(0, 0, 1)
	case "yesterday":
		return d.AddDate(0, 0, -1), d
	case "last7":
		return d.AddDate(0, 0, -6), d.AddDate(0, 0, 1)
	case "last14":
		return d.AddDate(0, 0, -13), d.AddDate(0, 0, 1)
	case "last30":
		return d.AddDate(0, 0, -29), d.AddDate(0, 0, 1)
	case "this-week":
		return d.AddDate(0, 0, -(int(d.Weekday())+6)%7), d.AddDate(0, 0, 1)
	case "this-month":
		return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location()), d.AddDate(0, 0, 1)
	case "previous-month":
		end := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
		return end.AddDate(0, -1, 0), end
	}
	return time.Time{}, time.Time{}
}
