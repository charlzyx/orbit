package transcriptui

import (
	"encoding/json"
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charlzyx/orbit/internal/highlight"
)

func TestBoundedDiffReconstructionAndLargeGrid(t *testing.T) {
	// Deterministic synthetic inputs, never model results. Each emitted patch
	// must reconstruct both sides, including repeated lines and tie cases.
	rng := rand.New(rand.NewPCG(23, 91))
	words := []string{"same", "different", "用户", "", "👨‍👩‍👧‍👦"}
	for n := 0; n < 300; n++ {
		generate := func() string {
			var lines []string
			for i, count := 0, rng.IntN(24); i < count; i++ {
				lines = append(lines, words[rng.IntN(len(words))])
			}
			if len(lines) == 0 {
				return ""
			}
			return strings.Join(lines, "\n") + "\n"
		}
		old, next := generate(), generate()
		rows := lineDiff(old, next, highlight.Unknown)
		var before, after []string
		for _, row := range rows {
			if row.Kind == Context || row.Kind == Deletion {
				before = append(before, row.Text)
			}
			if row.Kind == Context || row.Kind == Addition {
				after = append(after, row.Text)
			}
		}
		if !reflect.DeepEqual(before, diffLines(old)) || !reflect.DeepEqual(after, diffLines(next)) {
			t.Fatalf("case %d failed reconstruction", n)
		}
	}
	// 501x501 exceeds the source's 250,000-cell LCS budget. Fallback is
	// deterministic deletion-then-addition, not a lossy first-400-lines diff.
	old, next := strings.Repeat("old\n", 501), strings.Repeat("new\n", 501)
	rows := lineDiff(old, next, highlight.Unknown)
	if len(rows) != 1002 {
		t.Fatalf("fallback retained %d rows", len(rows))
	}
	for i, row := range rows {
		want := Deletion
		if i >= 501 {
			want = Addition
		}
		if row.Kind != want {
			t.Fatal("fallback order")
		}
	}
	if strings.Count(DiffText(rows), "\n") != 1002 {
		t.Fatal("copy truncated to paint budget")
	}
}

func TestUnicodeCapsAndMalformedToolFallback(t *testing.T) {
	text := strings.Repeat("用", OutputTextCap+1)
	raw, _ := json.Marshal(text)
	capped := display(raw, OutputTextCap)
	if !utf8.ValidString(capped) || len([]rune(capped)) < OutputTextCap || strings.Contains(capped, "�") {
		t.Fatal("cap split Unicode")
	}
	malformed := newTool("fixture-malformed", "edit", json.RawMessage(`{"path":"x.go","edits":[{"oldText":null,"newText":3}]}`))
	if len(malformed.Diff()) != 0 || len(malformed.Sections()) != 1 {
		t.Fatal("malformed edit erased inspectable arguments")
	}
	unknown := newTool("fixture-unknown", "custom_tool", json.RawMessage(`{"some":"actual fixture value"}`))
	if unknown.Preview() == "" || len(unknown.Sections()) != 1 {
		t.Fatal("unknown tool lost fallback")
	}
}
