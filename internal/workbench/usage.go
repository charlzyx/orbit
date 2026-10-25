package workbench

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const usageMaxFileBytes = 256 << 20

type Tokens struct{ Input, Output, CacheRead, CacheWrite, Reasoning, Total uint64 }

func (t Tokens) CacheRate() float64 {
	den := float64(t.Input) + float64(t.CacheRead) + float64(t.CacheWrite)
	if den == 0 {
		return 0
	}
	return float64(t.CacheRead) * 100 / den
}

type UsageRecord struct {
	At                                                                             time.Time
	Session, Workspace, Provider, Model, Outcome, Error, SessionPath, SessionTitle string
	Tokens                                                                         Tokens
	Cost                                                                           float64
	HasReasoning, HasCost, ModelPriced                                             bool
	Duration                                                                       time.Duration
	HasDuration                                                                    bool
}
type ToolRun struct {
	At                                                                       time.Time
	Session, Workspace, Name, ID, Provider, Model, SessionPath, SessionTitle string
	Success                                                                  bool
	HasResult                                                                bool
	Duration                                                                 time.Duration
	HasDuration                                                              bool
}
type UsageTurn struct {
	At                                            time.Time
	Session, Workspace, SessionPath, SessionTitle string
}
type UsageError struct {
	At                                                                                  time.Time
	Session, Workspace, Provider, Model, SessionPath, SessionTitle, Kind, Message, Tool string
}
type UsageSession struct {
	ID, Path, Title, Workspace string
	Started, Ended             time.Time
}
type UsageIndex struct {
	// ModelCatalog preserves source interning order, including assistant messages
	// without billable usage. It is scanner metadata, never persisted/exported.
	ModelCatalog           []UsageModelRef `json:"-"`
	Sessions               []UsageSession
	ScannedAt              time.Time
	Errors                 []UsageError
	Records                []UsageRecord
	Tools                  []ToolRun
	Turns                  []UsageTurn
	Warnings               []string
	Files, Parsed, Skipped int
	UnreadableFiles        int `json:"-"`
}
type usageFile struct {
	size       int64
	modified   time.Time
	index      UsageIndex
	unreadable bool
	warning    string
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
	if err := ctx.Err(); err != nil {
		return out, err
	}
	seen := map[string]bool{}
	modelSeen := map[UsageModelRef]bool{}
	// Match usage/collect.rs::scan_store: one workspace-group level only.
	// ReadDir intentionally follows a linked group, but never recursively walks
	// nested session trees or includes a JSONL file directly in the store root.
	groups, e := os.ReadDir(s.Root)
	if e != nil && !errors.Is(e, fs.ErrNotExist) {
		out.Warnings = append(out.Warnings, e.Error())
	}
	var paths []string
	for _, group := range groups {
		if ctx.Err() != nil {
			break
		}
		files, err := os.ReadDir(filepath.Join(s.Root, group.Name()))
		if err != nil {
			// Root files are not workspace groups; the Rust reader ignores them.
			if group.IsDir() {
				out.Warnings = append(out.Warnings, err.Error())
			}
			continue
		}
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".jsonl" {
				paths = append(paths, filepath.Join(s.Root, group.Name(), file.Name()))
			}
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		if ctx.Err() != nil {
			break
		}
		st, err := os.Stat(path)
		if err != nil {
			continue
		}
		out.Files++
		if st.Size() > usageMaxFileBytes {
			out.Skipped++
			out.Warnings = append(out.Warnings, filepath.Base(path)+": session exceeds 256 MiB")
			continue
		}
		seen[path] = true
		f, ok := s.cache[path]
		if !ok || f.size != st.Size() || !f.modified.Equal(st.ModTime()) {
			index, err := readUsageFile(ctx, path)
			if err != nil {
				if ctx.Err() != nil {
					break
				}
				f = usageFile{size: st.Size(), modified: st.ModTime(), unreadable: true, warning: filepath.Base(path) + ": " + err.Error()}
				s.cache[path] = f
			} else {
				f = usageFile{size: st.Size(), modified: st.ModTime(), index: index}
				s.cache[path] = f
				out.Parsed++
			}
		}
		if f.unreadable {
			out.UnreadableFiles++
			out.Skipped++
			if f.warning != "" {
				out.Warnings = append(out.Warnings, f.warning)
			}
			continue
		}
		for _, ref := range f.index.ModelCatalog {
			if !modelSeen[ref] {
				modelSeen[ref] = true
				out.ModelCatalog = append(out.ModelCatalog, ref)
			}
		}
		out.Records = append(out.Records, f.index.Records...)
		out.Tools = append(out.Tools, f.index.Tools...)
		out.Turns = append(out.Turns, f.index.Turns...)
		out.Errors = append(out.Errors, f.index.Errors...)
		out.Sessions = append(out.Sessions, f.index.Sessions...)
		out.Warnings = append(out.Warnings, f.index.Warnings...)
	}
	if errors.Is(e, fs.ErrNotExist) {
		e = nil
	}
	for p := range s.cache {
		if !seen[p] {
			delete(s.cache, p)
		}
	}
	out.markPricing()
	out.ScannedAt = time.Now()
	sort.SliceStable(out.Errors, func(i, j int) bool { return out.Errors[i].At.After(out.Errors[j].At) })
	sort.SliceStable(out.Records, func(i, j int) bool { return out.Records[i].At.Before(out.Records[j].At) })
	sort.SliceStable(out.Tools, func(i, j int) bool { return out.Tools[i].At.Before(out.Tools[j].At) })
	if ctx.Err() != nil {
		return out, ctx.Err()
	}
	return out, e
}
func ts(v any) time.Time {
	switch x := v.(type) {
	case string:
		t, _ := time.Parse(time.RFC3339Nano, x)
		return t
	case json.Number:
		if n, e := x.Int64(); e == nil {
			return time.UnixMilli(n)
		}
	case float64:
		if !math.IsNaN(x) && !math.IsInf(x, 0) && math.Trunc(x) == x && x >= -math.Exp2(63) && x < math.Exp2(63) {
			return time.UnixMilli(int64(x))
		}
	}
	return time.Time{}
}
func number(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case json.Number:
		f, _ := x.Float64()
		return f
	}
	return 0
}
func unsignedCount(v any) (uint64, bool) {
	switch x := v.(type) {
	case json.Number:
		n, e := strconv.ParseUint(string(x), 10, 64)
		if e != nil {
			return 0, false
		}
		return n, true
	case float64:
		if !math.IsNaN(x) && !math.IsInf(x, 0) && x >= 0 && x < math.Exp2(64) && math.Trunc(x) == x {
			return uint64(x), true
		}
	case uint64:
		return x, true
	case int:
		if x >= 0 {
			return uint64(x), true
		}
	}
	return 0, false
}
func count(v any) uint64 { n, _ := unsignedCount(v); return n }
func addCounts(values ...uint64) uint64 {
	var total uint64
	for _, n := range values {
		if math.MaxUint64-total < n {
			return math.MaxUint64
		}
		total += n
	}
	return total
}
func usageCost(v any) (float64, bool) {
	if m, ok := v.(map[string]any); ok {
		v = m["total"]
	}
	switch v.(type) {
	case json.Number, float64:
	default:
		return 0, false
	}
	n := number(v)
	return n, !math.IsNaN(n) && !math.IsInf(n, 0) && n >= 0
}
func (i *UsageIndex) markPricing() {
	priced := map[[2]string]bool{}
	for _, r := range i.Records {
		if r.Cost > 0 && !math.IsInf(r.Cost, 0) && !math.IsNaN(r.Cost) {
			priced[[2]string{r.Provider, r.Model}] = true
		}
	}
	for n := range i.Records {
		r := &i.Records[n]
		r.ModelPriced = priced[[2]string{r.Provider, r.Model}]
	}
}
func readUsageFile(ctx context.Context, path string) (UsageIndex, error) {
	if err := ctx.Err(); err != nil {
		return UsageIndex{}, err
	}
	absolute, e := filepath.Abs(path)
	if e != nil {
		return UsageIndex{}, e
	}
	path = absolute
	f, e := os.Open(path)
	if e != nil {
		return UsageIndex{}, e
	}
	defer f.Close()
	return readUsageStream(ctx, path, f)
}

// The pinned collector retains a valid prefix after a reader failure. A missing
// or wrong first valid JSON header is unreadable; malformed lines alone are not.
func readUsageStream(ctx context.Context, path string, reader io.Reader) (UsageIndex, error) {
	scanner := bufio.NewScanner(usageContextReader{ctx: ctx, Reader: reader})
	// A valid file can contain one long JSONL line; do not impose a smaller
	// record-size cap than the source file limit. One byte allows the delimiter.
	scanner.Buffer(make([]byte, 64<<10), usageMaxFileBytes+1)
	var out UsageIndex
	modelSeen := map[UsageModelRef]bool{}
	session, workspace, title := "", "", ""
	var previous, ended, sessionStarted time.Time
	type call struct {
		id, name, provider, model string
		at                        time.Time
	}
	var calls []call
	sawHeader := false
	for {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if !scanner.Scan() {
			break
		}
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if !utf8.Valid(scanner.Bytes()) {
			break
		}
		var value any
		dec := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		dec.UseNumber()
		if dec.Decode(&value) != nil || dec.Decode(new(any)) != io.EOF {
			continue
		}
		entry, _ := value.(map[string]any)
		if !sawHeader {
			if stringValue(entry["type"]) != "session" {
				return out, errors.New("missing session header")
			}
			sawHeader = true
			session = stringValue(entry["id"])
			workspace = stringValue(entry["cwd"])
			ended = ts(entry["timestamp"])
			if ended.IsZero() {
				ended = time.UnixMilli(0)
			}
			sessionStarted = ended
			if workspace == "" {
				workspace = "unknown"
			}
			continue
		}
		if stringValue(entry["type"]) != "message" {
			continue
		}
		message, ok := entry["message"].(map[string]any)
		if !ok {
			continue
		}
		at := ts(entry["timestamp"])
		started := ts(message["timestamp"])
		if at.IsZero() {
			at = started
		}
		if at.IsZero() {
			at = ended
		}
		if at.After(ended) {
			ended = at
		}
		switch stringValue(message["role"]) {
		case "user":
			if title == "" {
				title = usageSessionTitle(message["content"])
			}
			out.Turns = append(out.Turns, UsageTurn{At: at, Session: session, Workspace: workspace, SessionPath: path})
		case "assistant":
			provider, model := stringValue(message["provider"]), stringValue(message["model"])
			ref := UsageModelRef{Provider: provider, Model: model}
			if !modelSeen[ref] {
				modelSeen[ref] = true
				out.ModelCatalog = append(out.ModelCatalog, ref)
			}
			if usage, ok := message["usage"].(map[string]any); ok {
				tokens := Tokens{Input: count(usage["input"]), Output: count(usage["output"]), CacheRead: count(usage["cacheRead"]), CacheWrite: count(usage["cacheWrite"]), Reasoning: count(usage["reasoning"])}
				total, valid := unsignedCount(usage["totalTokens"])
				if valid {
					tokens.Total = total
				} else {
					tokens.Total = addCounts(tokens.Input, tokens.Output, tokens.CacheRead, tokens.CacheWrite)
				}
				cost, hasCost := usageCost(usage["cost"])
				if !hasCost {
					cost = 0
				}
				outcome := stringValue(message["stopReason"])
				switch outcome {
				case "toolUse", "length", "error", "aborted":
				default:
					outcome = "stop"
				}
				if tokens.Total > 0 || cost > 0 || message["errorMessage"] != nil || outcome == "error" {
					r := UsageRecord{At: at, Session: session, Workspace: workspace, SessionPath: path, Provider: provider, Model: model, Outcome: outcome, Error: stringValue(message["errorMessage"]), Tokens: tokens, HasReasoning: tokens.Reasoning > 0, Cost: cost, HasCost: hasCost}
					start := previous
					if start.IsZero() {
						start = started
					}
					delta := at.Sub(start)
					if !start.IsZero() && delta > 0 && delta <= 30*time.Minute {
						r.Duration = delta
						r.HasDuration = true
					}
					out.Records = append(out.Records, r)
				}
			}
			content, _ := message["content"].([]any)
			for _, raw := range content {
				if b, ok := raw.(map[string]any); ok && stringValue(b["type"]) == "toolCall" {
					name, ok := b["name"].(string)
					if !ok {
						name = "tool"
					}
					calls = append(calls, call{stringValue(b["id"]), name, provider, model, at})
				}
			}
			if messageText, ok := message["errorMessage"].(string); ok {
				out.Errors = append(out.Errors, UsageError{At: at, Session: session, Workspace: workspace, Provider: provider, Model: model, SessionPath: path, Kind: "provider", Message: usageTextCap(messageText, 160)})
			}
		case "toolResult":
			name, ok := message["toolName"].(string)
			if !ok {
				name = "tool"
			}
			r := ToolRun{At: at, Session: session, Workspace: workspace, SessionPath: path, ID: stringValue(message["toolCallId"]), Name: name, Success: message["isError"] != true, HasResult: true}
			if len(out.Records) > 0 {
				prev := out.Records[len(out.Records)-1]
				r.Provider, r.Model = prev.Provider, prev.Model
			}
			for n, c := range calls {
				if c.id == r.ID {
					r.At, r.Name, r.Provider, r.Model = c.at, c.name, c.provider, c.model
					delta := at.Sub(c.at)
					if !c.at.IsZero() && delta > 0 && delta <= 30*time.Minute {
						r.Duration = delta
						r.HasDuration = true
					}
					calls = append(calls[:n], calls[n+1:]...)
					break
				}
			}
			out.Tools = append(out.Tools, r)
			if !r.Success {
				out.Errors = append(out.Errors, UsageError{At: at, Session: session, Workspace: workspace, SessionPath: path, Provider: r.Provider, Model: r.Model, Kind: "tool", Tool: r.Name, Message: r.Name + " returned an error"})
			}
		}
		previous = at
	}
	if err := ctx.Err(); err != nil {
		return out, err
	}
	if !sawHeader {
		return out, errors.New("missing session header")
	}
	if session == "" {
		return UsageIndex{}, nil
	}
	for _, c := range calls {
		out.Tools = append(out.Tools, ToolRun{At: c.at, Session: session, Workspace: workspace, SessionPath: path, ID: c.id, Name: c.name, Provider: c.provider, Model: c.model, Success: true})
	}
	for n := range out.Records {
		out.Records[n].SessionTitle = title
	}
	for n := range out.Tools {
		out.Tools[n].SessionTitle = title
	}
	for n := range out.Turns {
		out.Turns[n].SessionTitle = title
	}
	for n := range out.Errors {
		out.Errors[n].SessionTitle = title
	}
	out.Sessions = []UsageSession{{ID: session, Path: path, Title: title, Workspace: workspace, Started: sessionStarted, Ended: ended}}
	out.markPricing()
	return out, nil
}

func usageTextCap(text string, limit int) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit]) + "…"
	}
	return text
}
func usageSessionTitle(content any) string {
	text, _ := content.(string)
	if blocks, ok := content.([]any); ok {
		for _, raw := range blocks {
			if block, ok := raw.(map[string]any); ok && stringValue(block["type"]) == "text" {
				text = stringValue(block["text"])
				break
			}
		}
	}
	trimmed := strings.TrimLeft(text, " \t\r\n")
	if rest, ok := strings.CutPrefix(trimmed, "<skill"); ok {
		if _, tail, ok := strings.Cut(rest, `name="`); ok {
			if name, _, ok := strings.Cut(tail, `"`); ok {
				return usageTextCap("Skill: "+name, 80)
			}
		}
		return "Skill loaded"
	}
	if strings.HasPrefix(trimmed, "<") {
		if _, tail, ok := strings.Cut(trimmed, ">"); ok {
			inner := strings.TrimLeft(tail, " \t\r\n")
			if inner != "" {
				text = inner
			}
		}
	}
	return usageTextCap(strings.ReplaceAll(text, "\n", " "), 80)
}

// UsageModelRef identifies a model in its provider namespace. Raw model IDs are
// not unique across providers; this scope is transient, never a preference.
type UsageModelRef struct{ Provider, Model string }

type UsageFilter struct {
	ModelRefs                                                         []UsageModelRef
	Start, End                                                        time.Time
	Workspace, Provider, Model, Outcome, Search, Session, SessionPath string
	Workspaces, Providers, Models                                     []string
	FocusStart, FocusEnd                                              time.Time
	ErrorsOnly, CachedOnly                                            bool
}

func containsScope(values []string, value string) bool {
	if len(values) == 0 {
		return true
	}
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
func (f UsageFilter) matchesTime(at time.Time) bool {
	return (f.Start.IsZero() || !at.Before(f.Start)) && (f.End.IsZero() || at.Before(f.End)) && (f.FocusStart.IsZero() || !at.Before(f.FocusStart)) && (f.FocusEnd.IsZero() || at.Before(f.FocusEnd))
}
func (f UsageFilter) matchesModelRef(provider, model string) bool {
	if len(f.ModelRefs) == 0 {
		return true
	}
	for _, ref := range f.ModelRefs {
		if ref.Provider == provider && ref.Model == model {
			return true
		}
	}
	return false
}
func (f UsageFilter) matchesScope(session, workspace, provider, model string) bool {
	return f.matchesModelRef(provider, model) && (f.Session == "" || session == f.Session) && (f.Workspace == "" || workspace == f.Workspace) && containsScope(f.Workspaces, workspace) && (f.Provider == "" || provider == f.Provider) && containsScope(f.Providers, provider) && (f.Model == "" || model == f.Model) && containsScope(f.Models, model)
}
func (f UsageFilter) Matches(r UsageRecord) bool {
	return (f.SessionPath == "" || f.SessionPath == r.SessionPath) && f.matchesTime(r.At) && f.matchesScope(r.Session, r.Workspace, r.Provider, r.Model) && (f.Outcome == "" || r.Outcome == f.Outcome) && (!f.ErrorsOnly || r.Outcome == "error") && (!f.CachedOnly || r.Tokens.CacheRead > 0) && (f.Search == "" || strings.Contains(strings.ToLower(r.Session+" "+r.SessionTitle+" "+r.Workspace+" "+r.Provider+" "+r.Model+" "+r.Error), strings.ToLower(f.Search)))
}
func (f UsageFilter) MatchesTool(r ToolRun) bool {
	return (f.SessionPath == "" || f.SessionPath == r.SessionPath) && !f.ErrorsOnly && !f.CachedOnly && f.Outcome == "" && f.matchesTime(r.At) && f.matchesScope(r.Session, r.Workspace, r.Provider, r.Model) && (f.Search == "" || strings.Contains(strings.ToLower(r.Session+" "+r.SessionTitle+" "+r.Workspace+" "+r.Provider+" "+r.Model+" "+r.Name), strings.ToLower(f.Search)))
}

// Prompts have no provider/model. Request-only error/cache filters exclude them.
func (f UsageFilter) MatchesTurn(r UsageTurn) bool {
	return (f.SessionPath == "" || f.SessionPath == r.SessionPath) && !f.ErrorsOnly && !f.CachedOnly && f.Outcome == "" && f.matchesTime(r.At) && (f.Session == "" || r.Session == f.Session) && (f.Workspace == "" || r.Workspace == f.Workspace) && containsScope(f.Workspaces, r.Workspace) && (f.Search == "" || strings.Contains(strings.ToLower(r.Session+" "+r.SessionTitle+" "+r.Workspace), strings.ToLower(f.Search)))
}
func (i UsageIndex) FilteredRecords(f UsageFilter) []UsageRecord {
	var out []UsageRecord
	for _, r := range i.Records {
		if f.Matches(r) {
			out = append(out, r)
		}
	}
	return out
}
func (i UsageIndex) FilteredTools(f UsageFilter) []ToolRun {
	var out []ToolRun
	for _, r := range i.Tools {
		if f.MatchesTool(r) {
			out = append(out, r)
		}
	}
	return out
}
func (i UsageIndex) FilteredTurns(f UsageFilter) []UsageTurn {
	var out []UsageTurn
	for _, r := range i.Turns {
		if f.MatchesTurn(r) {
			out = append(out, r)
		}
	}
	return out
}

// Error rows use scope/time filters, not request-only cache/error toggles.
func (f UsageFilter) MatchesError(r UsageError) bool {
	return (f.SessionPath == "" || f.SessionPath == r.SessionPath) && f.matchesTime(r.At) && f.matchesScope(r.Session, r.Workspace, r.Provider, r.Model) && (f.Search == "" || strings.Contains(strings.ToLower(r.Session+" "+r.SessionTitle+" "+r.Workspace+" "+r.Provider+" "+r.Model+" "+r.Message), strings.ToLower(f.Search)))
}
func (i UsageIndex) FilteredErrors(f UsageFilter) []UsageError {
	var out []UsageError
	for _, r := range i.Errors {
		if f.MatchesError(r) {
			out = append(out, r)
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].At.After(out[b].At) })
	return out
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
	priced := map[[2]string]bool{}
	for _, r := range i.Records {
		if r.Cost > 0 && !math.IsInf(r.Cost, 0) && !math.IsNaN(r.Cost) {
			priced[[2]string{r.Provider, r.Model}] = true
		}
	}
	for _, r := range i.Records {
		if !f.Matches(r) {
			continue
		}
		s.Requests++
		s.Tokens.Input = addCounts(s.Tokens.Input, r.Tokens.Input)
		s.Tokens.Output = addCounts(s.Tokens.Output, r.Tokens.Output)
		s.Tokens.CacheRead = addCounts(s.Tokens.CacheRead, r.Tokens.CacheRead)
		s.Tokens.CacheWrite = addCounts(s.Tokens.CacheWrite, r.Tokens.CacheWrite)
		s.Tokens.Reasoning = addCounts(s.Tokens.Reasoning, r.Tokens.Reasoning)
		s.Tokens.Total = addCounts(s.Tokens.Total, r.Tokens.Total)
		s.Cost += r.Cost
		if r.ModelPriced || priced[[2]string{r.Provider, r.Model}] {
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
		if f.MatchesTurn(r) {
			s.Turns++
		}
	}
	for _, r := range i.Tools {
		if f.MatchesTool(r) {
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
		b.Tokens = addCounts(b.Tokens, r.Tokens.Total)
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
	_ = c.Write([]string{"timestamp", "session_id", "session", "workspace", "provider", "model", "input", "output", "cache_read", "cache_write", "total", "reasoning", "cost_usd", "duration_ms", "outcome"})
	for _, r := range i.Records {
		if !f.Matches(r) {
			continue
		}
		cost := ""
		if r.HasCost {
			cost = fmt.Sprintf("%.6f", r.Cost)
		}
		duration := int64(0)
		if r.HasDuration {
			duration = r.Duration.Milliseconds()
		}
		outcome := r.Outcome
		if outcome == "toolUse" {
			outcome = "tool_use"
		}
		_ = c.Write([]string{r.At.Local().Format(time.RFC3339), csvSafe(r.Session), csvSafe(r.SessionTitle), csvSafe(r.Workspace), csvSafe(usageProviderLabel(r.Provider)), csvSafe(usageModelLabel(r.Model)), strconv.FormatUint(r.Tokens.Input, 10), strconv.FormatUint(r.Tokens.Output, 10), strconv.FormatUint(r.Tokens.CacheRead, 10), strconv.FormatUint(r.Tokens.CacheWrite, 10), strconv.FormatUint(r.Tokens.Total, 10), strconv.FormatUint(r.Tokens.Reasoning, 10), cost, strconv.FormatInt(duration, 10), csvSafe(outcome)})
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
		return d, now
	case "yesterday":
		return d.Add(-24 * time.Hour), d
	case "last7":
		return d.Add(-6 * 24 * time.Hour), now
	case "last14":
		return d.Add(-13 * 24 * time.Hour), now
	case "last30":
		return d.Add(-29 * 24 * time.Hour), now
	case "this-week":
		monday := d.Add(-time.Duration((int(d.Weekday())+6)%7) * 24 * time.Hour)
		return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location()), now
	case "this-month":
		return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location()), now
	case "all", "all_time":
		return time.UnixMilli(0), now
	case "previous-month":
		end := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
		return end.AddDate(0, -1, 0), end
	}
	return d.Add(-6 * 24 * time.Hour), now
}

// Check cancellation between bounded scanner reads as well as parsed entries.
// Ordinary read failures still preserve the pinned collector's valid prefix.
type usageContextReader struct {
	ctx context.Context
	io.Reader
}

func (r usageContextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.Reader.Read(p)
}
