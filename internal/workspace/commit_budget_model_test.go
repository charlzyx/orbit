package workspace

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCommitUnknownWindowSharesDiffBudget(t *testing.T) {
	g := testGit(t)
	for i := 0; i < 8; i++ {
		writeGitFile(t, g, fmt.Sprintf("file%d.txt", i), strings.Repeat("synthetic long line\n", 12000))
	}
	p, rows, e := CommitGenerationPrompt(context.Background(), g, false)
	if e != nil {
		t.Fatal(e)
	}
	if len(rows) != 8 {
		t.Fatal(rows)
	}
	if len(p.System)+len(p.User) > 256*1024 {
		t.Fatalf("unknown-window prompt unbounded across files: %d bytes", len(p.System)+len(p.User))
	}
}

func TestCommitContextWindowBudgetBoundaries(t *testing.T) {
	for _, tc := range []struct{ window, total int }{{0, 256 * 1024}, {-1, 256 * 1024}, {1, 24 * 1024}, {32000, 48000}, {128000, 192000}, {1000000, 1500000}, {int(^uint(0) >> 1), 2 * 1024 * 1024}} {
		if got := commitTotal(tc.window); got != tc.total {
			t.Fatal(tc, got)
		}
		b := newCommitBudget(tc.window, 0, 6)
		diff, orig := 0, 0
		for i := 0; i < 6; i++ {
			d := b.takeDiff(strings.Repeat("界", 100000), 6-i)
			o := b.takeOriginal(strings.Repeat("语", 100000), 6-i)
			diff += len(d)
			orig += len(o)
			if len(d) > 64*1024 || len(o) > 48*1024 {
				t.Fatal("per-file budget")
			}
		}
		if diff > tc.total/2 || orig > tc.total*3/10 {
			t.Fatal("shared budget exceeded", tc, diff, orig)
		}
	}
	for i := 0; i < 100; i++ {
		v := commitBound(strings.Repeat("你好", 100), i)
		if len(v) > i || !utf8.ValidString(v) {
			t.Fatal(i, len(v))
		}
	}
}
func TestCommitSelectedWindowCapsActualMultiFilePrompt(t *testing.T) {
	g := testGit(t)
	for i := 0; i < 12; i++ {
		writeGitFile(t, g, fmt.Sprintf("f%d.txt", i), strings.Repeat("change\n", 20000))
	}
	var previous int
	for _, window := range []int{32000, 128000, 1000000} {
		p, rows, e := commitGenerationPrompt(context.Background(), g, false, window)
		if e != nil {
			t.Fatal(e)
		}
		size := len(p.User) + len(p.System)
		if size > commitTotal(window) || size <= previous {
			t.Fatal(window, size, previous)
		}
		previous = size
		for _, row := range rows {
			if !strings.Contains(p.User, row.Path) {
				t.Fatal("file summary lost", row)
			}
		}
	}
}

func TestCommitUnknownSelectedModelFailureIsExplicit(t *testing.T) {
	g := testGit(t)
	writeGitFile(t, g, "fixture.txt", "only synthetic\n")
	cfg := generationConfig(t, "ok", "must not succeed")
	cfg.Provider, cfg.Model = "absent-provider", "absent-model"
	result, e := GenerateCommitRPC(context.Background(), g, cfg, false)
	if e != nil || !result.Fallback || !strings.Contains(result.Warning, "wrong selected model") || result.Text == "must not succeed" {
		t.Fatal(result, e)
	}
}
