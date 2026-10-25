package transcriptui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/charlzyx/orbit/internal/agent"
)

type object map[string]json.RawMessage

func obj(raw json.RawMessage) object   { var v object; _ = json.Unmarshal(raw, &v); return v }
func str(raw json.RawMessage) string   { var v string; _ = json.Unmarshal(raw, &v); return v }
func boolean(raw json.RawMessage) bool { var v bool; _ = json.Unmarshal(raw, &v); return v }
func present(raw json.RawMessage) bool {
	return len(raw) > 0 && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}
func clone(raw json.RawMessage) json.RawMessage { return bytes.Clone(raw) }
func number(raw json.RawMessage) *float64 {
	var n float64
	if json.Unmarshal(raw, &n) != nil || !present(raw) || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
		return nil
	}
	return &n
}
func uintValue(raw json.RawMessage) *uint64 {
	n := number(raw)
	if n == nil || *n >= math.Exp2(64) {
		return nil
	}
	v := uint64(*n)
	return &v
}
func count(raw json.RawMessage) uint64 {
	if n := uintValue(raw); n != nil {
		return *n
	}
	return 0
}
func messageObject(m agent.Message) object {
	o := obj(m.Raw)
	if inner := obj(o["message"]); inner != nil {
		return inner
	}
	return o
}

func ParseUsage(raw json.RawMessage) *Usage {
	o := obj(raw)
	if o == nil {
		return nil
	}
	u := &Usage{Input: count(o["input"]), Output: count(o["output"]), CacheRead: count(o["cacheRead"]), CacheWrite: count(o["cacheWrite"])}
	if v := uintValue(o["totalTokens"]); v != nil {
		u.Total = *v
	} else {
		u.Total = u.Input + u.Output + u.CacheRead + u.CacheWrite
	}
	u.Cost = number(o["cost"])
	if u.Cost == nil {
		u.Cost = number(obj(o["cost"])["total"])
	}
	if u.Total == 0 && (u.Cost == nil || *u.Cost <= 0) {
		return nil
	}
	return u
}
func (r Row) Usage() *Usage {
	var total *Usage
	for _, s := range r.Steps {
		if s.Usage == nil {
			continue
		}
		u := s.Usage
		if total == nil {
			total = &Usage{}
		}
		total.Input += u.Input
		total.Output += u.Output
		total.CacheRead += u.CacheRead
		total.CacheWrite += u.CacheWrite
		total.Total += u.Total
		if u.Cost != nil {
			if total.Cost == nil {
				v := 0.
				total.Cost = &v
			}
			*total.Cost += *u.Cost
		}
	}
	return total
}
func (r Row) GenerationTime() *time.Duration {
	var total time.Duration
	found := false
	for _, s := range r.Steps {
		if s.Duration != nil {
			total += *s.Duration
			found = true
		}
	}
	if !found {
		return nil
	}
	return &total
}
func (u Usage) CachePercent() *float64 {
	n := u.Input + u.CacheRead + u.CacheWrite
	if n == 0 {
		return nil
	}
	v := 100 * float64(u.CacheRead) / float64(n)
	return &v
}
func (r Row) Text() string {
	var parts []string
	for _, s := range r.Steps {
		if s.Text != "" {
			parts = append(parts, s.Text)
		}
	}
	return strings.Join(parts, "\n\n")
}
func (r Row) HasWork() bool {
	for _, s := range r.Steps {
		if s.Thinking != "" || len(s.Tools) > 0 {
			return true
		}
	}
	return false
}
func (r Row) ChangedFiles() []ChangedFile { return ChangedFilesOf([]Row{r}) }
func ChangedFilesOf(rows []Row) []ChangedFile {
	var files []ChangedFile
	indices := map[string]int{}
	for _, r := range rows {
		for _, s := range r.Steps {
			for _, t := range s.Tools {
				if t.Path == "" || (t.Added == 0 && t.Removed == 0) {
					continue
				}
				if i, ok := indices[t.Path]; ok {
					files[i].Added += t.Added
					files[i].Removed += t.Removed
				} else {
					indices[t.Path] = len(files)
					files = append(files, ChangedFile{t.Path, t.Added, t.Removed})
				}
			}
		}
	}
	return files
}

// Project mirrors Transcript::load_from: ignore context-only messages, merge
// consecutive assistant steps, and attach results to the newest matching call.
// Live tool_execution_* rows emitted by agent are folded into those same calls.
// Orphan tool executions remain visible rather than fabricating an assistant.
func Project(messages []agent.Message) []Row {
	var rows []Row
	for mi, m := range messages {
		o := messageObject(m)
		role := m.Role
		if role == "" {
			role = str(o["role"])
		}
		id := m.ID
		if id == "" {
			id = fmt.Sprintf("message-%d", mi)
		}
		if role == "tool" || role == "toolResult" {
			tool := ToolFromMessage(m)
			if attachTool(rows, tool, role == "tool") {
				continue
			}
			// Snapshot-only orphan toolResult is context-only, exactly as Rust. An
			// execution event without an assistant call gets a recoverable tool row.
			if role == "tool" {
				rows = append(rows, Row{ID: id, Role: "tool", Streaming: m.Streaming, Steps: []Step{{ID: id, Tools: []ToolCall{tool}, Streaming: m.Streaming}}})
			}
			continue
		}
		if role == "compactionSummary" || role == "branchSummary" {
			rows = append(rows, Row{ID: id, Role: role, FinishedAt: timestamp(o["timestamp"], m.Timestamp), Summary: &SummaryRow{Branch: role == "branchSummary", Text: str(o["summary"]), TokensBefore: uintValue(o["tokensBefore"])}})
			continue
		}
		if role != "user" && role != "assistant" {
			continue
		}
		step := Step{ID: id, Text: m.Text, Thinking: m.Thinking, Streaming: m.Streaming, Timestamp: timestamp(o["timestamp"], m.Timestamp), Usage: ParseUsage(m.Usage)}
		if step.Usage == nil {
			step.Usage = ParseUsage(o["usage"])
		}
		if ms := number(o["durationMs"]); ms != nil && *ms < float64(math.MaxInt64)/float64(time.Millisecond) {
			v := time.Duration(math.Round(*ms)) * time.Millisecond
			step.Duration = &v
		}
		blocks := m.Blocks
		if len(blocks) == 0 {
			var rawBlocks []json.RawMessage
			_ = json.Unmarshal(o["content"], &rawBlocks)
			for _, b := range rawBlocks {
				bo := obj(b)
				blocks = append(blocks, agent.ContentBlock{Type: str(bo["type"]), Text: str(bo["text"]), Thinking: str(bo["thinking"]), ToolCallID: str(bo["id"]), ToolName: str(bo["name"]), Arguments: clone(bo["arguments"]), Raw: b})
			}
		}
		var texts, thoughts []string
		var images []agent.Image
		for bi, b := range blocks {
			switch b.Type {
			case "text":
				texts = append(texts, b.Text)
			case "thinking":
				thoughts = append(thoughts, b.Thinking)
			case "toolCall", "tool_call":
				tool := toolFromBlock(b)
				if tool.ID == "" {
					tool.ID = fmt.Sprintf("%s:tool-%d", id, bi)
				}
				tool.Running = m.Streaming && !tool.Complete
				step.Tools = append(step.Tools, tool)
			case "image":
				if b.Image != nil {
					images = append(images, *b.Image)
				} else {
					bo := obj(b.Raw)
					mime := str(bo["mimeType"])
					if mime == "" {
						mime = str(bo["mime_type"])
					}
					if data := str(bo["data"]); data != "" {
						images = append(images, agent.Image{Type: "image", Data: data, MIMEType: mime})
					}
				}
			}
		}
		if len(texts) > 0 {
			step.Text = strings.Join(texts, "\n\n")
		}
		if len(thoughts) > 0 {
			step.Thinking = strings.Join(thoughts, "")
		}
		if step.Text == "" {
			step.Text = str(o["content"])
		}
		if role == "user" {
			step.Text = compactUserPrompt(step.Text)
		}
		r := Row{ID: id, Role: role, Steps: []Step{step}, Images: images, Streaming: m.Streaming, FinishedAt: step.Timestamp, Aborted: str(o["stopReason"]) == "aborted"}
		if role == "assistant" && !r.Aborted {
			r.Error = strings.TrimSpace(str(o["errorMessage"]))
			if r.Error == "" {
				r.Error = m.Error
			}
		}
		if role == "assistant" && len(rows) > 0 && rows[len(rows)-1].Role == "assistant" {
			prev := &rows[len(rows)-1]
			prev.Steps = append(prev.Steps, step)
			prev.Images = append(prev.Images, images...)
			prev.Streaming = prev.Streaming || r.Streaming
			prev.Aborted = prev.Aborted || r.Aborted
			if !r.FinishedAt.IsZero() {
				prev.FinishedAt = r.FinishedAt
			}
			if r.Error != "" {
				prev.Error = r.Error
			}
		} else {
			rows = append(rows, r)
		}
	}
	return rows
}
func timestamp(raw json.RawMessage, fallback time.Time) time.Time {
	if s := str(raw); s != "" {
		if t, e := time.Parse(time.RFC3339Nano, s); e == nil {
			return t
		}
	}
	if n := number(raw); n != nil {
		v := int64(*n)
		if *n > float64(math.MaxInt64) {
			return fallback
		}
		if v <= 1_000_000_000_000 {
			v *= 1000
		}
		return time.UnixMilli(v)
	}
	return fallback
}
func attachTool(rows []Row, in ToolCall, live bool) bool {
	if in.ID == "" {
		return false
	}
	for ri := len(rows) - 1; ri >= 0; ri-- {
		for si := len(rows[ri].Steps) - 1; si >= 0; si-- {
			for ti := range rows[ri].Steps[si].Tools {
				t := &rows[ri].Steps[si].Tools[ti]
				if t.ID != in.ID {
					continue
				}
				if present(in.Arguments) {
					t.Arguments = in.Arguments
					t.Path = in.Path
					t.Summary = in.Summary
					t.Added = in.Added
					t.Removed = in.Removed
				}
				if present(in.Output) && (live || !present(t.Output)) {
					t.Output = in.Output
				}
				t.Failed = in.Failed
				t.Facts = in.Facts
				t.Running = in.Running
				t.Complete = in.Complete
				rows[ri].Streaming = rows[ri].Streaming || in.Running
				return true
			}
		}
	}
	return false
}
func toolFromBlock(b agent.ContentBlock) ToolCall {
	o := obj(b.Raw)
	id := b.ToolCallID
	if id == "" {
		id = str(o["toolCallId"])
	}
	name := b.ToolName
	if name == "" {
		name = str(o["name"])
	}
	if name == "" {
		name = "tool"
	}
	args := b.Arguments
	if !present(args) {
		args = o["arguments"]
	}
	t := newTool(id, name, args)
	output := o["result"]
	if !present(output) {
		output = o["output"]
	}
	if present(output) {
		t.Output = clone(output)
		t.Facts = factsFromResult(output)
		t.Complete = true
	}
	t.Failed = boolean(o["isError"])
	return t
}
func newTool(id, name string, args json.RawMessage) ToolCall {
	if len(args) > 0 && !json.Valid(args) {
		args = nil
	}
	t := ToolCall{ID: id, Name: name, Arguments: clone(args)}
	o := obj(args)
	t.Path = str(o["path"])
	if present(args) {
		var compact bytes.Buffer
		if json.Compact(&compact, args) == nil {
			t.Summary = capChars(compact.String(), 160)
		} else {
			t.Summary = capChars(string(args), 160)
		}
	}
	switch name {
	case "write":
		t.Added = countLines(str(o["content"]))
	case "edit":
		var edits []object
		_ = json.Unmarshal(o["edits"], &edits)
		for _, e := range edits {
			t.Added += countLines(str(e["newText"]))
			t.Removed += countLines(str(e["oldText"]))
		}
	}
	return t
}
func ToolFromMessage(m agent.Message) ToolCall {
	o := messageObject(m)
	name, id := m.ToolName, m.ToolCallID
	if name == "" {
		name = str(o["toolName"])
	}
	if id == "" {
		id = str(o["toolCallId"])
	}
	args := o["args"]
	for _, b := range m.Blocks {
		if b.Type == "toolCall" || b.Type == "tool_call" {
			if present(b.Arguments) {
				args = b.Arguments
			}
			if name == "" {
				name = b.ToolName
			}
			if id == "" {
				id = b.ToolCallID
			}
			break
		}
	}
	t := newTool(id, name, args)
	t.Running = m.Streaming
	t.Complete = !m.Streaming
	t.Failed = m.Error != "" || boolean(o["isError"])
	if flag, ok := o["isError"]; ok {
		t.Failed = boolean(flag)
	}
	result := o["result"]
	if !present(result) {
		result = o["partialResult"]
	}
	if !present(result) && str(o["role"]) == "toolResult" {
		result, _ = json.Marshal(o)
	}
	if present(result) {
		t.Output = normalizeResult(result)
		if m.Role == "toolResult" || str(o["role"]) == "toolResult" {
			if _, ok := obj(result)["content"]; !ok {
				t.Output = nil
			} else if str(t.Output) == "" {
				t.Output = nil
			}
		}
		t.Facts = factsFromResult(result)
	} else if m.ToolOutput != "" {
		t.Output, _ = json.Marshal(m.ToolOutput)
	} else if m.Text != "" {
		t.Output, _ = json.Marshal(m.Text)
	}
	return t
}
func normalizeResult(raw json.RawMessage) json.RawMessage {
	o := obj(raw)
	if o == nil {
		return clone(raw)
	}
	var blocks []object
	if json.Unmarshal(o["content"], &blocks) == nil && present(o["content"]) && bytes.HasPrefix(bytes.TrimSpace(o["content"]), []byte("[")) {
		var parts []string
		for _, b := range blocks {
			if str(b["type"]) == "text" {
				if v := str(b["text"]); v != "" {
					parts = append(parts, v)
				}
			}
		}
		v, _ := json.Marshal(strings.Join(parts, "\n"))
		return v
	}
	return clone(raw)
}
func factsFromResult(raw json.RawMessage) ToolFacts {
	d := obj(obj(raw)["details"])
	var f ToolFacts
	if tr := d["truncation"]; present(tr) {
		o := obj(tr)
		f.Truncated = true
		if _, ok := o["truncated"]; ok {
			f.Truncated = boolean(o["truncated"])
		}
		f.OutputLines = uintValue(o["outputLines"])
		f.TotalLines = uintValue(o["totalLines"])
	}
	for _, k := range []string{"matchLimitReached", "linesTruncated", "entryLimitReached"} {
		f.Truncated = f.Truncated || boolean(d[k])
	}
	return f
}
func countLines(s string) uint64 {
	if s == "" {
		return 0
	}
	return uint64(strings.Count(s, "\n") + 1)
}
func capChars(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}
func display(raw json.RawMessage, limit int) string {
	if !present(raw) {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return capChars(s, limit)
	}
	var v any
	if json.Unmarshal(raw, &v) == nil {
		b, _ := json.MarshalIndent(v, "", "  ")
		return capChars(string(b), limit)
	}
	return capChars(string(raw), limit)
}
func firstError(raw json.RawMessage) string {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return ""
	}
	var find func(any) string
	find = func(v any) string {
		switch x := v.(type) {
		case string:
			for _, l := range strings.Split(x, "\n") {
				if l = strings.TrimSpace(l); l != "" {
					return capChars(l, 200)
				}
			}
		case []any:
			for _, e := range x {
				if s := find(e); s != "" {
					return s
				}
			}
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if s := find(x[k]); s != "" {
					return s
				}
			}
		}
		return ""
	}
	return find(v)
}

// compactUserPrompt removes only Pi's trailing image transport hints and its
// explicit injected <skill ...>...</skill> wrapper. Ordinary prose is unchanged.
func compactUserPrompt(text string) string {
	text = strings.TrimRight(text, " \t\r\n")
	for {
		sep := strings.LastIndex(text, "\n\n")
		if sep < 0 {
			break
		}
		tail := strings.TrimSpace(text[sep+2:])
		all := true
		for _, l := range strings.Split(tail, "\n") {
			l = strings.TrimSpace(l)
			all = all && (strings.HasPrefix(l, "[Image:") || strings.HasPrefix(l, "[Image converted") || strings.HasPrefix(l, "[Image omitted"))
		}
		if !all {
			break
		}
		text = strings.TrimRight(text[:sep], " \t\r\n")
	}
	rest, ok := strings.CutPrefix(strings.TrimLeft(text, " \t\r\n"), "<skill")
	if !ok {
		return text
	}
	end, close := strings.Index(rest, ">"), strings.Index(rest, "</skill>")
	if end < 0 || close < 0 {
		return text
	}
	name := ""
	if _, after, found := strings.Cut(rest[:end], `name="`); found {
		name, _, _ = strings.Cut(after, `"`)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = "skill"
	}
	trailing := strings.TrimSpace(rest[close+len("</skill>"):])
	text = "/skill:" + name
	if trailing != "" {
		text += " " + trailing
	}
	return text
}
