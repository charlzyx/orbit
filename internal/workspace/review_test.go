package workspace

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestReviewSixSourcesAndUntrackedPaths(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	cpRun(t, g, "switch", "-c", "task")
	cpWrite(t, root, "committed.txt", "committed\n")
	cpRun(t, g, "add", ".")
	cpRun(t, g, "commit", "-m", "task")
	cpWrite(t, root, "staged.txt", "staged\n")
	cpRun(t, g, "add", "staged.txt")
	cpWrite(t, root, "untracked\t\n.txt", "unstaged\n")
	cpWrite(t, root, "seed.txt", "modified\n")
	for source, want := range map[ReviewSource]int{ReviewUncommitted: 3, ReviewUnstaged: 2, ReviewStaged: 1, ReviewCommitted: 1, ReviewBranch: 4} {
		d, e := g.CollectReview(ctx, ReviewRequest{Source: source})
		if e != nil {
			t.Fatal(source, e)
		}
		s := ParseReview(d)
		if len(s.Files) != want {
			t.Fatalf("%s got %#v", source, s.Files)
		}
		for _, f := range s.Files {
			if f.DiffLine < 0 {
				t.Fatal("missing navigation", f)
			}
		}
	}
	if _, e := g.CollectReview(ctx, ReviewRequest{Source: "invalid"}); e == nil {
		t.Fatal("invalid source")
	}
}
func TestReviewGapsExpandStableIDsAndSplit(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	var old strings.Builder
	for i := 1; i <= 450; i++ {
		fmt.Fprintf(&old, "line %d\n", i)
	}
	cpWrite(t, root, "long.txt", old.String())
	cpRun(t, g, "add", ".")
	cpRun(t, g, "commit", "-m", "long")
	modified := strings.Replace(old.String(), "line 225\n", "changed\n", 1)
	cpWrite(t, root, "long.txt", modified)
	d, e := g.CollectReview(context.Background(), ReviewRequest{Source: ReviewUncommitted})
	if e != nil {
		t.Fatal(e)
	}
	s := ParseReview(d)
	gap := -1
	for i, l := range s.Lines {
		if l.Gap != nil {
			gap = i
			break
		}
	}
	if gap < 0 {
		t.Fatal("no gap")
	}
	id := s.Lines[gap].Gap.ID
	before := len(s.Lines)
	if n := s.ExpandGap(gap, "end"); n != 101 || len(s.Lines) != before+100 {
		t.Fatal(n, len(s.Lines), before)
	}
	if s.Lines[gap].Gap.ID != id {
		t.Fatal("unstable ID")
	}
	for i := 0; i < len(s.Lines); {
		if s.Lines[i].Gap != nil {
			if s.ExpandGap(i, "all") == 0 {
				t.Fatal("cannot expand")
			}
		} else {
			i++
		}
	}
	if len(s.Lines) < 450 {
		t.Fatal(len(s.Lines))
	}
	split := SplitReview(s.Lines)
	paired := false
	for _, row := range split {
		if row.Left != nil && row.Right != nil && row.Left.Kind == "deletion" && row.Right.Kind == "addition" {
			paired = true
		}
	}
	if !paired {
		t.Fatal("split replacement unpaired")
	}
	if _, _, ok := s.FileHeadersAround(20); !ok {
		t.Fatal("sticky header")
	}
}
func TestReviewBinaryGroupedAndCompactFallback(t *testing.T) {
	d := ReviewDiff{Source: ReviewUncommitted, Numstat: "-\t-\tbinary\x001\t0\tnew\x00", Patch: "diff --git a/binary b/binary\nBinary files a/binary and b/binary differ\ndiff --git a/new b/new\nnew file mode 100644\n@@ -0,0 +1 @@\n+new\n", CompleteContext: true}
	s := ParseReview(d)
	if len(s.Files) != 2 || s.Files[0].Status != "binary" || s.Files[1].Status != "added" {
		t.Fatal(s)
	}
	group := GroupReviews(s, s)
	if !group.Grouped() || len(group.Files) != 4 || group.Additions != 2 {
		t.Fatal(group)
	}
	if _, _, ok := group.SectionHeadersAround(len(group.Lines) - 1); !ok {
		t.Fatal("group header")
	}
	compact := ParseReview(ReviewDiff{Numstat: "1\t1\ta\x00", Patch: "diff --git a/a b/a\n@@ -100 +100 @@\n-old\n+new\n", CompleteContext: false})
	found := false
	for i, l := range compact.Lines {
		if l.Gap != nil {
			found = true
			if l.Gap.Expandable() || compact.ExpandGap(i, "all") != 0 {
				t.Fatal("fake compact expansion")
			}
		}
	}
	if !found {
		t.Fatal("missing compact gap")
	}
}
func TestReviewVisibleLimitAndExpansionCap(t *testing.T) {
	var b strings.Builder
	b.WriteString("diff --git a/a b/a\n@@ -0,0 +1,50002 @@\n")
	for i := 0; i < 50002; i++ {
		b.WriteString("+line\n")
	}
	s := ParseReview(ReviewDiff{Numstat: "50002\t0\ta\x00", Patch: b.String(), CompleteContext: true})
	if !s.Truncated || len(s.Lines) != MaxRenderedDiffLines {
		t.Fatal(len(s.Lines), s.Truncated)
	}
	s.Lines[1] = ReviewLine{Kind: "gap", Gap: &ReviewGap{ID: 1, Count: 2, Hidden: []ReviewLine{{Kind: "context"}, {Kind: "context"}}}}
	if s.ExpandGap(1, "all") != 0 {
		t.Fatal("over cap")
	}
}

func TestReviewSelectedCommitRootAndLiteralPath(t *testing.T) {
	root := wtTestRepo(t)
	g := Git{Dir: root}
	ctx := context.Background()
	initial := cpRun(t, g, "rev-parse", "HEAD")
	d, e := g.CollectReviewAt(ctx, ReviewCommitted, initial, "seed.txt")
	if e != nil {
		t.Fatal(e)
	}
	s := ParseReview(d)
	if len(s.Files) != 1 || s.Files[0].Status != "added" || s.Files[0].Path != "seed.txt" {
		t.Fatal(s)
	}
	cpWrite(t, root, "literal[1].txt", "first\n")
	cpWrite(t, root, "literal1.txt", "unrelated\n")
	cpRun(t, g, "add", ".")
	cpRun(t, g, "commit", "-m", "files")
	target := cpRun(t, g, "rev-parse", "HEAD")
	cpWrite(t, root, "literal[1].txt", "later workspace\n")
	d, e = g.CollectReviewAt(ctx, ReviewCommitted, target, "literal[1].txt")
	if e != nil {
		t.Fatal(e)
	}
	s = ParseReview(d)
	if len(s.Files) != 1 || strings.Contains(d.Patch, "later workspace") || strings.Contains(d.Patch, "unrelated") {
		t.Fatal(d)
	}
	if _, e = g.CollectReviewAt(ctx, ReviewCommitted, "--help", ""); e == nil {
		t.Fatal("revision option injection")
	}
	if _, e = g.CollectReviewAt(ctx, ReviewCommitted, target, "../outside"); e == nil {
		t.Fatal("outside path")
	}
}

func TestCompactGapCountsAndEachFileReset(t *testing.T) {
	s := ParseReview(ReviewDiff{Numstat: "2\t2\ta\x001\t1\tb\x00", Patch: "diff --git a/a b/a\n@@ -1 +1 @@\n-old\n+new\n@@ -5 +5 @@\n-old\n+new\ndiff --git a/b b/b\n@@ -10 +10 @@\n-old\n+new\n"})
	var counts []int
	for _, l := range s.Lines {
		if l.Gap != nil {
			counts = append(counts, l.Gap.Count)
		}
	}
	if len(counts) != 2 || counts[0] != 3 || counts[1] != 9 {
		t.Fatal(counts)
	}
}
