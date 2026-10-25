// Package composerui implements Orbit's source-derived composer completion menu.
// Offsets exposed by this package are Unicode rune offsets, matching MyGo.
package composerui

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const Limit = 8
const MaxFiles = 4000
const MaxDepth = 12

type Kind uint8

const (
	None Kind = iota
	Slash
	At
)

type Trigger struct {
	Kind       Kind
	Start, End int
	Query      string
}

// DetectTrigger follows mentions.rs, avoiding its invalid UTF-8/caret-zero panic.
func DetectTrigger(content string, caret int) (Trigger, bool) {
	text := []rune(content)
	caret = min(max(caret, 0), len(text))
	if len(text) > 0 && text[0] == '/' && caret > 0 {
		end := 1
		for end < len(text) && !unicode.IsSpace(text[end]) {
			end++
		}
		if caret <= end {
			return Trigger{Slash, 0, end, string(text[1:caret])}, true
		}
	}
	at := -1
	for i := 0; i < caret; i++ {
		if text[i] == '@' {
			at = i
		}
	}
	if at < 0 || (at > 0 && !unicode.IsSpace(text[at-1])) {
		return Trigger{}, false
	}
	for _, r := range text[at+1 : caret] {
		if unicode.IsSpace(r) {
			return Trigger{}, false
		}
	}
	return Trigger{At, at, caret, string(text[at+1 : caret])}, true
}

type Span struct {
	Start, End int
	Kind       Kind
}

func TokenizeMentions(content string) []Span {
	text := []rune(content)
	out := []Span{}
	if len(text) > 0 && text[0] == '/' {
		end := 1
		valid := true
		for end < len(text) && !unicode.IsSpace(text[end]) {
			r := text[end]
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' || r == ':') {
				valid = false
			}
			end++
		}
		if valid && end > 1 {
			out = append(out, Span{0, end, Slash})
		}
	}
	for i, r := range text {
		if r != '@' || i > 0 && !unicode.IsSpace(text[i-1]) {
			continue
		}
		end := i + 1
		for end < len(text) && !unicode.IsSpace(text[end]) {
			end++
		}
		if end > i+1 {
			out = append(out, Span{i, end, At})
		}
	}
	return out
}

type Scope string

const (
	Builtin Scope = "builtin"
	Skill   Scope = "skills"
	Orbit   Scope = "orbit"
	Custom  Scope = "custom"
	Project Scope = "project"
)

type Command struct {
	Name, Description string
	Scope             Scope
	Project           string
}
type Entry struct {
	Kind                    Kind
	Name, Description, Path string
	Scope                   Scope
	Project                 string
}

func (e Entry) Title() string {
	if e.Kind == Slash {
		return "/" + e.Name
	}
	p := strings.LastIndex(e.Path, "/")
	return e.Path[p+1:]
}
func (e Entry) Subtitle() string {
	if e.Kind == Slash {
		return e.Description
	}
	p := strings.LastIndex(e.Path, "/")
	if p < 0 {
		return ""
	}
	return e.Path[:p+1]
}
func (e Entry) Insertion() string {
	if e.Kind == Slash {
		return "/" + e.Name + " "
	}
	return "@" + e.Path + " "
}
func (e Entry) ScopeLabel() string {
	if e.Scope == Project && e.Project != "" {
		return e.Project
	}
	return string(e.Scope)
}
func ClassifyScope(source, scope, origin, path, workspace, home string) (Scope, string) {
	if source == "skill" {
		return Skill, ""
	}
	if home != "" && path != "" {
		rel, err := filepath.Rel(filepath.Join(home, ".orbit-pi"), path)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return Orbit, ""
		}
	}
	if scope == "project" {
		name := ""
		if workspace != "" {
			name = filepath.Base(filepath.Clean(workspace))
			if name == "." || name == string(filepath.Separator) {
				name = ""
			}
		}
		return Project, name
	}
	if origin == "package" || source == "builtin" || strings.HasPrefix(path, "<") {
		return Builtin, ""
	}
	if scope == "user" {
		return Custom, ""
	}
	return Builtin, ""
}
func ParseCommands(raw []json.RawMessage, workspace, home string) []Command {
	out := make([]Command, 0, len(raw))
	for _, b := range raw {
		// Decode field-by-field: an unrelated wrong-shaped field must not discard a valid name.
		var v map[string]json.RawMessage
		if json.Unmarshal(b, &v) != nil {
			continue
		}
		var name *string
		if json.Unmarshal(v["name"], &name) != nil || name == nil {
			continue
		}
		field := func(m map[string]json.RawMessage, k string) string {
			var s string
			_ = json.Unmarshal(m[k], &s)
			return s
		}
		var info map[string]json.RawMessage
		_ = json.Unmarshal(v["sourceInfo"], &info)
		scope, project := ClassifyScope(field(v, "source"), field(info, "scope"), field(info, "origin"), field(info, "path"), workspace, home)
		out = append(out, Command{*name, field(v, "description"), scope, project})
	}
	return out
}
func fuzzyMatch(needle, haystack string) (int, bool) {
	hay := []rune(haystack)
	start, prev, score := 0, -1, 0
	for _, r := range needle {
		found := -1
		for i := start; i < len(hay); i++ {
			if hay[i] == r {
				found = i
				break
			}
		}
		if found < 0 {
			return 0, false
		}
		if prev >= 0 && prev == found-1 {
			score += 3
		}
		if found == 0 || hay[found-1] == '/' {
			score += 2
		}
		score -= min(found-(prev+1), 12)
		prev = found
		start = found + 1
	}
	return score, true
}
func FilterEntries(query string, files []string, commands []Command, limit int) []Entry {
	type scored struct {
		score int
		entry Entry
	}
	rows := []scored{}
	lower := cases.Lower(language.Und)
	q := lower.String(query)
	for _, c := range commands {
		if s, ok := fuzzyMatch(q, lower.String(c.Name)); ok {
			rows = append(rows, scored{s + 10, Entry{Kind: Slash, Name: c.Name, Description: c.Description, Scope: c.Scope, Project: c.Project}})
		}
	}
	for _, p := range files {
		base := p[strings.LastIndex(p, "/")+1:]
		s, ok := fuzzyMatch(q, lower.String(base))
		s *= 2
		ps, pok := fuzzyMatch(q, lower.String(p))
		if pok && (!ok || ps > s) {
			s = ps
		}
		if ok || pok {
			rows = append(rows, scored{s, Entry{Kind: At, Path: p}})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].score > rows[j].score })
	limit = min(max(limit, 0), len(rows))
	out := make([]Entry, limit)
	for i := range out {
		out[i] = rows[i].entry
	}
	return out
}
