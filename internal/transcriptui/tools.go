package transcriptui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/go-text/typesetting/segmenter"
)

const (
	DetailTextCap       = 4000
	OutputTextCap       = 100000
	OutputCollapseLines = 16
	OutputPreviewLines  = 12
	OutputPaintLines    = 400
	DiffCollapseLines   = 40
	DiffPreviewLines    = 24
	DiffPaintLines      = 400
)

func MCPParts(name string) (server, tool string, ok bool) {
	rest, ok := strings.CutPrefix(name, "mcp__")
	if !ok {
		return "", "", false
	}
	server, tool, ok = strings.Cut(rest, "__")
	return server, tool, ok && server != "" && tool != ""
}
func ActionLabel(name string) string {
	for _, e := range []struct{ names, label string }{{"bash shell terminal exec run", "Run"}, {"read view", "Read"}, {"edit", "Edit"}, {"write", "Write"}, {"grep search", "Search"}, {"find glob", "Find files"}, {"list ls tree", "List"}, {"task agent subagent", "Task"}, {"web_search websearch search_web browse", "Web search"}, {"web_fetch webfetch fetch open_url", "Fetch"}, {"skill", "Skill"}, {"ask ask_user question elicit", "Ask"}, {"todo todo_write plan update_plan", "Plan"}, {"notebook eval execute_code codemode", "Run code"}, {"tool_search", "Find tools"}, {"list_mcp_resources list_mcp_resource_templates", "List resources"}, {"read_mcp_resource", "Read resource"}} {
		if in(name, e.names) {
			return e.label
		}
	}
	if strings.HasPrefix(name, "mcp") {
		return "MCP"
	}
	r := []rune(name)
	if len(r) == 0 {
		return "Tool"
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
func Icon(name string) string {
	for _, e := range []struct{ names, icon string }{{"edit", "edit"}, {"write", "write"}, {"read view", "read"}, {"grep search tool_search", "search"}, {"find glob", "find"}, {"list ls tree", "list"}, {"bash shell terminal exec run", "bash"}, {"task agent subagent", "task"}, {"web_search websearch search_web browse", "web"}, {"web_fetch webfetch fetch open_url", "fetch"}, {"skill", "skill"}, {"ask ask_user question elicit", "ask"}, {"todo todo_write plan update_plan", "todo"}, {"notebook eval execute_code codemode", "code"}} {
		if in(name, e.names) {
			return "icons/tools/" + e.icon + ".svg"
		}
	}
	if strings.HasPrefix(name, "mcp") || strings.Contains(name, "mcp_resource") {
		return "icons/tools/mcp.svg"
	}
	return "icons/tools/tool.svg"
}
func in(s, set string) bool {
	for _, v := range strings.Fields(set) {
		if v == s {
			return true
		}
	}
	return false
}
func (t ToolCall) Command() (string, bool) {
	if !in(t.Name, "bash shell terminal exec run") {
		return "", false
	}
	v, ok := obj(t.Arguments)["command"]
	if !ok {
		return "", false
	}
	var s string
	err := json.Unmarshal(v, &s)
	return s, err == nil
}
func (t ToolCall) Preview() string {
	if t.Path != "" {
		return t.Path
	}
	o := obj(t.Arguments)
	if in(t.Name, "web_fetch webfetch fetch open_url browse") {
		if url := str(o["url"]); url != "" {
			url = strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
			return strings.TrimPrefix(url, "www.")
		}
	}
	if in(t.Name, "web_search websearch search_web browse") {
		if q := str(o["query"]); q != "" {
			return q
		}
	}
	if t.Name == "codemode" {
		for _, l := range strings.Split(str(o["code"]), "\n") {
			if l = strings.TrimSpace(l); l != "" {
				return snippet(l, 72)
			}
		}
	}
	if command, ok := t.Command(); ok {
		return snippet(command, 72)
	}
	return snippet(t.Summary, 72)
}
func snippet(s string, n int) string {
	normalized := strings.Join(strings.Fields(s), " ")
	var seg segmenter.Segmenter
	if err := seg.InitWithString(normalized); err != nil {
		return capChars(normalized, n)
	}
	it := seg.GraphemeIterator()
	end, count := 0, 0
	for it.Next() {
		g := it.Grapheme()
		// The pinned segmenter can split an adjacent emoji ZWJ sequence after
		// its joiner. Never cut a preview at that incomplete joiner boundary.
		continuation := end > 0 && strings.HasSuffix(normalized[:end], "\u200d")
		if count == n && !continuation {
			return normalized[:end] + "…"
		}
		end = g.OffsetInBytes + g.LengthInBytes
		if !continuation {
			count++
		}
	}
	return normalized
}

func (t ToolCall) ErrorText() string { return firstError(t.Output) }
func (t ToolCall) TruncationLabel() string {
	if !t.Facts.Truncated {
		return ""
	}
	if t.Facts.OutputLines != nil && t.Facts.TotalLines != nil && *t.Facts.TotalLines > *t.Facts.OutputLines {
		return fmt.Sprintf("truncated · %d/%d", *t.Facts.OutputLines, *t.Facts.TotalLines)
	}
	return "truncated"
}

type Section struct {
	ID, Label, Text string
	Language        highlight.Language
	Prompt          bool
}

func (t ToolCall) Sections() []Section {
	var result []Section
	if cmd, ok := t.Command(); ok {
		result = append(result, Section{"arguments", "Command", cmd, highlight.Shell, true})
	} else if len(t.Diff()) == 0 && present(t.Arguments) {
		result = append(result, Section{"arguments", "Arguments", display(t.Arguments, DetailTextCap), sectionLanguage(t.Arguments), false})
	}
	if present(t.Output) {
		text := display(t.Output, OutputTextCap)
		if strings.TrimSpace(text) != "" {
			result = append(result, Section{"output", "Output", text, sectionLanguage(t.Output), false})
		}
	}
	return result
}
func sectionLanguage(raw json.RawMessage) highlight.Language {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return highlight.Unknown
	}
	return highlight.JSON
}

type LineWindow struct {
	Visible                     []string
	Total                       int
	Foldable, Expanded, Clipped bool
}

func SectionLines(s string, expanded bool) LineWindow {
	lines := strings.Split(s, "\n")
	w := LineWindow{Visible: lines, Total: len(lines), Foldable: len(lines) > OutputCollapseLines}
	w.Expanded = w.Foldable && expanded
	if w.Foldable && !expanded {
		w.Visible = lines[:OutputPreviewLines]
	} else if len(lines) > OutputPaintLines {
		w.Visible = lines[:OutputPaintLines]
		w.Clipped = expanded
	}
	return w
}

type DiffKind uint8

const (
	Context DiffKind = iota
	Addition
	Deletion
	Break
)

type DiffRow struct {
	Kind   DiffKind
	Text   string
	Tokens []highlight.Token
}

func (t ToolCall) Diff() []DiffRow {
	o := obj(t.Arguments)
	lang := highlight.LanguageForTag(strings.TrimPrefix(filepath.Ext(t.Path), "."))
	if t.Name == "write" {
		var content string
		if json.Unmarshal(o["content"], &content) != nil {
			return nil
		}
		lines := diffLines(content)
		tokens := highlight.Tokenize(lang, content)
		var rows []DiffRow
		for i, l := range lines {
			r := DiffRow{Kind: Addition, Text: l}
			if i < len(tokens) {
				r.Tokens = tokens[i]
			}
			rows = append(rows, r)
		}
		return rows
	}
	if t.Name != "edit" {
		return nil
	}
	var edits []object
	_ = json.Unmarshal(o["edits"], &edits)
	var rows []DiffRow
	for _, e := range edits {
		var old, new string
		if json.Unmarshal(e["oldText"], &old) != nil || json.Unmarshal(e["newText"], &new) != nil {
			continue
		}
		if len(rows) > 0 {
			rows = append(rows, DiffRow{Kind: Break})
		}
		rows = append(rows, lineDiff(old, new, lang)...)
	}
	return rows
}
func diffLines(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}
func lineDiff(old, new string, lang highlight.Language) []DiffRow {
	a, b := diffLines(old), diffLines(new)
	at, bt := highlight.Tokenize(lang, old), highlight.Tokenize(lang, new)
	n, m := len(a), len(b)
	rows := make([]DiffRow, 0, n+m)
	add := func(k DiffKind, i int) {
		var text string
		var tokens []highlight.Token
		if k == Addition {
			text = b[i]
			if i < len(bt) {
				tokens = bt[i]
			}
		} else {
			text = a[i]
			if i < len(at) {
				tokens = at[i]
			}
		}
		rows = append(rows, DiffRow{k, text, tokens})
	}
	if n > 0 && m > 250000/n {
		for i := range a {
			add(Deletion, i)
		}
		for j := range b {
			add(Addition, j)
		}
		return rows
	}
	lcs := make([]uint32, (n+1)*(m+1))
	get := func(i, j int) uint32 { return lcs[i*(m+1)+j] }
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			v := max(get(i+1, j), get(i, j+1))
			if a[i] == b[j] {
				v = get(i+1, j+1) + 1
			}
			lcs[i*(m+1)+j] = v
		}
	}
	i, j := 0, 0
	for i < n && j < m {
		if a[i] == b[j] {
			add(Context, i)
			i++
			j++
		} else if get(i+1, j) >= get(i, j+1) {
			add(Deletion, i)
			i++
		} else {
			add(Addition, j)
			j++
		}
	}
	for ; i < n; i++ {
		add(Deletion, i)
	}
	for ; j < m; j++ {
		add(Addition, j)
	}
	return rows
}
func DiffText(rows []DiffRow) string {
	var out strings.Builder
	for _, r := range rows {
		p := " "
		switch r.Kind {
		case Addition:
			p = "+"
		case Deletion:
			p = "-"
		case Break:
			continue
		}
		out.WriteString(p)
		out.WriteString(r.Text)
		out.WriteByte('\n')
	}
	return out.String()
}
