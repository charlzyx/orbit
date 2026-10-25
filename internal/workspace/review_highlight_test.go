package workspace

import (
	"reflect"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/highlight"
)

func TestReviewHighlightCarriesSeparateSidesAndResetsAtHunks(t *testing.T) {
	s := ParseReview(ReviewDiff{Numstat: "3\t3\ta.rs\x001\t1\tb.py\x00", Patch: "diff --git a/a.rs b/a.rs\n@@ -1,3 +1,3 @@\n-/* old\n+let fresh = true;\n untouched\n-*/ let old = 2;\n+let next = 3;\n@@ -8 +8 @@\n let resumed = 4;\ndiff --git a/b.py b/b.py\n@@ -1 +1 @@\n+@cache\n"})
	find := func(text string) ReviewLine {
		for _, l := range s.Lines {
			if l.Content == text {
				return l
			}
		}
		t.Fatalf("missing %q", text)
		return ReviewLine{}
	}
	if got := find("let fresh = true;").Tokens; len(got) != 2 || got[0].Class != highlight.Keyword || got[1].Class != highlight.Literal {
		t.Fatal(got)
	}
	if got := find("untouched").Tokens; len(got) != 0 {
		t.Fatal("context should paint new side, not old comment", got)
	}
	if got := find("*/ let old = 2;").Tokens; len(got) != 3 || got[0].Class != highlight.Comment || got[1].Class != highlight.Keyword {
		t.Fatal("old carry lost", got)
	}
	if got := find("let resumed = 4;").Tokens; len(got) != 2 || got[0].Class != highlight.Keyword {
		t.Fatal("hunk did not reset carry", got)
	}
	if got := find("@cache").Tokens; len(got) != 1 || got[0].Class != highlight.Meta {
		t.Fatal("new file did not resolve language", got)
	}
}
func TestReviewHighlightSurvivesHiddenExpansionAndGrouping(t *testing.T) {
	var patch strings.Builder
	patch.WriteString("diff --git a/code.rs b/code.rs\n@@ -1,12 +1,12 @@\n-/* old\n+/* new\n")
	for i := 0; i < 10; i++ {
		patch.WriteString(" unchanged comment\n")
	}
	patch.WriteString(" */\n")
	s := ParseReview(ReviewDiff{CompleteContext: true, Numstat: "1\t1\tcode.rs\x00", Patch: patch.String()})
	found := false
	for i, l := range s.Lines {
		if l.Gap != nil && len(l.Gap.Hidden) > 0 {
			found = true
			for _, h := range l.Gap.Hidden {
				if len(h.Tokens) != 1 || h.Tokens[0].Class != highlight.Comment {
					t.Fatal(h)
				}
			}
			s.ExpandGap(i, "all")
			break
		}
	}
	if !found {
		t.Fatal("no retained gap")
	}
	grouped := GroupReviews(s, s)
	for _, l := range grouped.Lines {
		if l.Content == "unchanged comment" {
			want, _ := highlight.TokenizeLine(highlight.Rust, l.Content, highlight.Carry{Kind: highlight.BlockComment})
			if !reflect.DeepEqual(l.Tokens, want) {
				t.Fatal("tokens lost", l)
			}
		}
	}
	rows := SplitReview(s.Lines)
	for _, row := range rows {
		for _, line := range []*ReviewLine{row.Left, row.Right, row.Full} {
			if line != nil && line.Content == "unchanged comment" && len(line.Tokens) != 1 {
				t.Fatal("split lost tokens")
			}
		}
	}
}

func TestReviewSourceHeaderPathsMetadataAndSyntheticHunks(t *testing.T) {
	s := ParseReview(ReviewDiff{Numstat: "1\t1\tsecond.rs\x001\t1\tfirst.py\x00", Patch: "diff --git a/first.py b/first.py\nindex abc..def 100644\n--- a/first.py\n+++ b/first.py\n@@\n-def old():\n+def new():\n\\ No newline at end of file\ndiff --git a/second.rs b/second.rs\n@@ -1 +1 @@\n-let old = 1;\n+let new = 2;\ndiff --git a/extra.txt b/extra.txt\nGIT binary patch\nignored metadata\n", CompleteContext: true})
	if len(s.Files) != 3 || s.Files[2].Status != "binary" {
		t.Fatal(s.Files)
	}
	for _, line := range s.Lines {
		if line.Kind == "hunk_header" || strings.HasPrefix(line.Content, "index ") || strings.HasPrefix(line.Content, "--- ") || strings.HasPrefix(line.Content, "+++ ") || line.Content == "ignored metadata" {
			t.Fatal("source-suppressed metadata leaked", line)
		}
		if strings.HasPrefix(line.Content, "def ") && (line.File != 1 || line.OldLine != 0 || line.NewLine != 0) {
			t.Fatal("synthetic position or filename mismatch", line)
		}
		if strings.HasPrefix(line.Content, "let ") && line.File != 0 {
			t.Fatal("numstat order used instead of path", line)
		}
		if line.Kind == "meta" && strings.HasPrefix(line.Content, "\\") {
			t.Fatal("no-newline marker was not removed", line)
		}
	}
}
func TestReviewHeaderEscapesAndCompactAsymmetricGaps(t *testing.T) {
	for raw, want := range map[string]string{
		"diff --git a/space b/name b/space b/name":                 "space b/name",
		"diff --git \"a/tab\\tline\\n.rs\" \"b/tab\\tline\\n.rs\"": "tab\tline\n.rs",
		"diff --git \"a/\\303\\251.rs\" \"b/\\303\\251.rs\"":       "é.rs",
	} {
		got, ok := reviewHeaderPath(raw)
		if !ok || got != want {
			t.Fatalf("%q => %q %v, want %q", raw, got, ok, want)
		}
	}
	s := ParseReview(ReviewDiff{Numstat: "1\t1\ta\x00", Patch: "diff --git a/a b/a\n@@ -2 +10 @@\n-old\n+new\n@@ -5 +20 @@\n-old\n+new\n"})
	var gaps []*ReviewGap
	for _, l := range s.Lines {
		if l.Gap != nil {
			gaps = append(gaps, l.Gap)
		}
	}
	if len(gaps) != 2 || gaps[0].Count != 9 || gaps[0].Position != "leading" || gaps[1].Count != 9 || gaps[1].Position != "between" {
		t.Fatal(gaps)
	}
}
