package highlight

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"
)

// Fixtures are output of the pinned Rust executable, not expectations generated
// from the Go scanner. See testdata/README.md for provenance and reproduction.
func TestPinnedSourceOracle(t *testing.T) {
	files, err := filepath.Glob("testdata/*.json")
	if err != nil || len(files) != 16 {
		t.Fatalf("fixtures: %v %v", files, err)
	}
	count := 0
	for _, file := range files {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		var fixtures []struct {
			Name, Tag, Text string
			Carry, Next     Carry
			Tokens          []Token
		}
		if e = json.Unmarshal(data, &fixtures); e != nil {
			t.Fatal(e)
		}
		for _, f := range fixtures {
			count++
			t.Run(f.Tag+"/"+f.Name, func(t *testing.T) {
				got, next := TokenizeLine(LanguageForTag(f.Tag), f.Text, f.Carry)
				if len(got) == 0 {
					got = []Token{}
				}
				if !reflect.DeepEqual(got, f.Tokens) || next != f.Next {
					t.Fatalf("text=%q carry=%+v\ngot %+v %+v\nwant %+v %+v", f.Text, f.Carry, got, next, f.Tokens, f.Next)
				}
			})
		}
	}
	if count != 1204 {
		t.Fatalf("fixture count %d", count)
	}
}
func TestLanguageAliasesLabelsAndPaths(t *testing.T) {
	for tag, want := range tagLanguages {
		if got := LanguageForTag("  " + strings.ToUpper(tag) + " "); got != want {
			t.Errorf("%s = %v, want %v", tag, got, want)
		}
	}
	for tag, want := range languageLabels {
		if got := LanguageLabel(" " + strings.ToUpper(tag) + " "); got != want {
			t.Errorf("label %s = %s", tag, got)
		}
	}
	for path, want := range map[string]Language{"src/MAIN.RS": Rust, "Makefile": Shell, "Dockerfile": Shell, "cargo.lock": TOML, "package-lock.json": JSON, "composer.lock": JSON, "note.md": Unknown, "noext": Unknown, "/a/.gitignore": Unknown} {
		if got := LanguageForPath(path); got != want {
			t.Errorf("path %s = %v want %v", path, got, want)
		}
	}
	if LanguageForTag("brainfuck") != Unknown || LanguageLabel("unknown") != "" || LanguageLabel("objc") != "" || LanguageLabel("scala") != "" {
		t.Fatal("unknown labels must stay unknown")
	}
}
func TestEveryStreamingUnicodePrefixIsLossless(t *testing.T) {
	source := "let π = 0xFA; /* é😀 */ @decorator(世界) 'hello' \"escaped\\\"tail\" `multi\nline`\n𐐀( x𝟎 😀 // tail"
	for lang := Rust; lang <= Diff; lang++ {
		for n := 0; n <= len(source); n++ {
			if !utf8.ValidString(source[:n]) {
				continue
			}
			lines := strings.Split(source[:n], "\n")
			tokens := Tokenize(lang, source[:n])
			if len(lines) != len(tokens) {
				t.Fatal("line count")
			}
			for i, line := range lines {
				previous := 0
				for _, token := range tokens[i] {
					if token.Start < previous || token.End <= token.Start || token.End > len(line) || !utf8.ValidString(line[:token.Start]) || !utf8.ValidString(line[:token.End]) {
						t.Fatalf("lang %v invalid token %+v in %q", lang, token, line)
					}
					previous = token.End
				}
				var rebuilt strings.Builder
				for _, run := range Runs(line, tokens[i]) {
					rebuilt.WriteString(run.Text)
				}
				if rebuilt.String() != line {
					t.Fatalf("lost bytes %q", line)
				}
			}
		}
	}
}
func TestRunsRejectInvalidSpansWithoutChangingText(t *testing.T) {
	text := "éclair 世界"
	tokens := []Token{{1, 2, Keyword}, {0, 2, String}, {1, 4, Comment}, {3, 999, Type}, {7, 10, Number}, {len(text), len(text), Meta}}
	var b strings.Builder
	for _, r := range Runs(text, tokens) {
		b.WriteString(r.Text)
		if !utf8.ValidString(r.Text) {
			t.Fatalf("invalid UTF8 span %q", r.Text)
		}
	}
	if b.String() != text {
		t.Fatal(b.String())
	}
	for lang := Rust; lang <= Diff; lang++ {
		Tokenize(lang, "a\xff'\xfe'")
	}
}
func TestCarryAndCacheRemainIndependent(t *testing.T) {
	code := "/* open\nend */ let x = true;\n"
	first := TokenizeCached(Rust, code)
	if len(first.Lines) != 3 || first != TokenizeCached(Rust, code) {
		t.Fatal("cache miss or trailing line lost")
	}
	if first == TokenizeCached(C, code) || first == TokenizeCached(Rust, code+"x") {
		t.Fatal("incorrect cache key")
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				TokenizeCached(Rust, code)
			}
		}()
	}
	wg.Wait()
	for i := 0; i < 300; i++ {
		TokenizeCached(Rust, strings.Repeat("x", i))
	}
	blockCache.Lock()
	n := len(blockCache.entries)
	blockCache.Unlock()
	if n > 256 {
		t.Fatal(n)
	}
	if !reflect.DeepEqual(first.Lines, Tokenize(Rust, code)) {
		t.Fatal("eviction mutated old block")
	}
}
