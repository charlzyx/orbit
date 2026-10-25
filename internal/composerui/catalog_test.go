package composerui

import (
	"encoding/json"
	"reflect"
	"testing"
)

func raws(s ...string) []json.RawMessage {
	out := make([]json.RawMessage, len(s))
	for i, v := range s {
		out[i] = json.RawMessage(v)
	}
	return out
}
func TestSourceTriggerBoundaries(t *testing.T) {
	for _, tt := range []struct {
		text  string
		caret int
		want  Trigger
		ok    bool
	}{
		{"/rev", 4, Trigger{Slash, 0, 4, "rev"}, true}, {"/review args", 2, Trigger{Slash, 0, 7, "r"}, true}, {"/", 1, Trigger{Slash, 0, 1, ""}, true}, {"/rev ", 5, Trigger{}, false}, {"run /rev", 8, Trigger{}, false},
		{"look at @src/comp", 17, Trigger{At, 8, 17, "src/comp"}, true}, // oversized caret clamps to rune length
		{"a b @c d", 8, Trigger{}, false}, {"abc@x", 5, Trigger{}, false}, {"@", 1, Trigger{At, 0, 1, ""}, true}, {"@one @tw", 8, Trigger{At, 5, 8, "tw"}, true}, {"done\n@re", 8, Trigger{At, 5, 8, "re"}, true},
		{"你好\u2003@文件尾", 6, Trigger{At, 3, 6, "文件"}, true}, {"/", 0, Trigger{}, false}, {"/", -1, Trigger{}, false}, {"", 0, Trigger{}, false}, {"@a@b", 4, Trigger{}, false},
	} {
		got, ok := DetectTrigger(tt.text, tt.caret)
		if got != tt.want || ok != tt.ok {
			t.Errorf("Detect(%q,%d)=%+v,%v want %+v,%v", tt.text, tt.caret, got, ok, tt.want, tt.ok)
		}
	}
}
func TestSourceMentionPaintSpans(t *testing.T) {
	got := TokenizeMentions("/review fix @src/main.rs and @lib.rs")
	want := []Span{{0, 7, Slash}, {12, 24, At}, {29, 36, At}}
	if !reflect.DeepEqual(got, want) {
		t.Fatal(got)
	}
	for _, s := range []string{"/usr/local/bin", "mail a@b.com", "/", "@"} {
		if len(TokenizeMentions(s)) != 0 {
			t.Fatal(s)
		}
	}
	if got := TokenizeMentions("你好\n@文件 /plain @two"); !reflect.DeepEqual(got, []Span{{3, 6, At}, {14, 18, At}}) {
		t.Fatal(got)
	}
	if got := TokenizeMentions("@a@b"); len(got) != 1 || got[0].End != 4 {
		t.Fatal(got)
	}
}
func TestSourceFuzzyScoresOrderingAndLimit(t *testing.T) {
	files := []string{"crates/app/main.rs", "src/main.rs", "README.md", "main.py"}
	entries := FilterEntries("main", files, nil, 10)
	got := []string{}
	for _, e := range entries {
		got = append(got, e.Path)
	}
	if !reflect.DeepEqual(got, []string{"crates/app/main.rs", "src/main.rs", "main.py"}) {
		t.Fatal(got)
	}
	files = []string{"crates/app/main.rs", "src/App.tsx"}
	if got := FilterEntries("app", files, nil, 10); len(got) != 2 || got[0].Path != "src/App.tsx" {
		t.Fatal(got)
	}
	if len(FilterEntries("APPX", files, nil, 10)) != 1 {
		t.Fatal("fuzzy/case")
	}
	if len(FilterEntries("zzq", files, nil, 10)) != 0 {
		t.Fatal("nonmatch")
	}
	commands := []Command{{Name: "review", Description: "zzz"}, {Name: "reset"}, {Name: "read"}}
	if len(FilterEntries("zzz", nil, commands, 10)) != 0 {
		t.Fatal("description incorrectly searched")
	}
	if got := FilterEntries("", files, commands, 4); len(got) != 4 || got[0].Name != "review" || got[3].Path != files[0] {
		t.Fatal(got)
	}
	if len(FilterEntries("", files, nil, -1)) != 0 {
		t.Fatal("negative limit")
	}
}
func TestSourceCommandProvenanceAndTolerantRPC(t *testing.T) {
	rows := ParseCommands(raws(
		`{"name":"skill:review","source":"skill","sourceInfo":{"scope":"project"}}`,
		`{"name":"generate-title","source":"extension","sourceInfo":{"scope":"temporary","path":"/home/u/.orbit-pi/title/index.js"}}`,
		`{"name":"project","description":"local","sourceInfo":{"scope":"project"}}`,
		`{"name":"package","sourceInfo":{"scope":"user","origin":"package"}}`,
		`{"name":"custom","sourceInfo":{"scope":"user"}}`,
		`{"name":"builtin","sourceInfo":{"scope":"user","path":"<inline:thing>"}}`,
		`{"name":"odd-description","description":false,"sourceInfo":[]}`,
		`{"name":3}`, `{"name":null}`, `{}`, `null`, `{"name":""}`), "/work/orbit", "/home/u")
	if len(rows) != 8 {
		t.Fatalf("rows=%+v", rows)
	}
	for i, want := range []Scope{Skill, Orbit, Project, Builtin, Custom, Builtin, Builtin, Builtin} {
		if rows[i].Scope != want {
			t.Errorf("row%d %+v", i, rows[i])
		}
	}
	if rows[2].Project != "orbit" || rows[2].Description != "local" {
		t.Fatal(rows[2])
	}
	if scope, _ := ClassifyScope("extension", "user", "", "/home/u/.orbit-pi-other/x", "/work/orbit", "/home/u"); scope != Custom {
		t.Fatal("path prefix false positive")
	}
}
func TestSourceFileBadgesAndRowMetadata(t *testing.T) {
	for path, want := range map[string]string{"a.md": "MD", "a.jsx": "JSX", "index.html": "HTML", "a.tsx": "TSX", "a.rs": "RS", "a.py": "PY", "a.scss": "SCSS", "a.png": "IMG", "Cargo.toml": "TOML", "Makefile": "FILE", "NOTES.MD": "MD"} {
		if got := FileBadge(path); got.Label != want || got.Dark == 0 || got.Light == 0 {
			t.Fatal(path, got)
		}
	}
	e := Entry{Kind: At, Path: "src/main.go"}
	if e.Title() != "main.go" || e.Subtitle() != "src/" || e.Insertion() != "@src/main.go " {
		t.Fatal(e)
	}
	e = Entry{Kind: Slash, Name: "skill:review", Description: "Review", Scope: Skill}
	if e.Title() != "/skill:review" || e.Subtitle() != "Review" || e.ScopeLabel() != "skills" {
		t.Fatal(e)
	}
}

func TestUnicodeLowercaseMatchesSourceExpansion(t *testing.T) {
	if got := FilterEntries("i\u0307", []string{"İ.go"}, nil, Limit); len(got) != 1 {
		t.Fatal("Unicode full lowercase expansion was lost", got)
	}
}
func FuzzTriggerRuneBounds(f *testing.F) {
	for _, s := range []string{"/review args", "你好 @文件😀 suffix", "@x", "abc@x", ""} {
		f.Add(s, 3)
	}
	f.Fuzz(func(t *testing.T, text string, caret int) {
		if tr, ok := DetectTrigger(text, caret); ok {
			r := []rune(text)
			if tr.Start < 0 || tr.End < tr.Start || tr.End > len(r) {
				t.Fatalf("invalid range %+v", tr)
			}
			_ = r[tr.Start:tr.End]
		}
	})
}

func TestSourceDeviconsAndFontPreference(t *testing.T) {
	if len(devicons) != 591 {
		t.Fatalf("devicons table count=%d", len(devicons))
	}
	if glyph := glyphForFile("notes.txt"); glyph.Glyph != "󰈙" || glyph.Dark != 0x89e051 || glyph.Light != 0x447028 {
		t.Fatal(glyph)
	}
	if glyphForFile("README.md") != glyphForFile("dir/README.md") {
		t.Fatal("basename mapping")
	}
	if glyphForFile("a.TSX") != glyphForFile("a.tsx") {
		t.Fatal("extension lower-case fallback")
	}
	if glyphForFile("unknown.zzzzz").Glyph != "*" {
		t.Fatal("default glyph")
	}
	if got := chooseNerdFamily([]string{"Plain mono", "Hack Nerd Font", "Symbols Nerd Font Mono", "A NF"}); got != "Hack Nerd Font" {
		t.Fatal(got)
	}
	if got := chooseNerdFamily([]string{"Plain mono", "sans-serif"}); got != "" {
		t.Fatal("fake Nerd font", got)
	}
}
