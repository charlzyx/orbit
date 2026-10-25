package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGitDiscardAllPreservesStagedIgnoredAndExternalTargets(t *testing.T) {
	g := testGit(t)
	ctx := context.Background()
	testCommit(t, g, "a.txt", "committed\n", "initial")
	writeGitFile(t, g, ".gitignore", "ignored/\n")
	testRun(t, g, "add", ".gitignore")
	testRun(t, g, "commit", "-m", "ignore")
	writeGitFile(t, g, "a.txt", "staged\n")
	writeGitFile(t, g, "staged-new.txt", "staged new\n")
	requireNoError(t, g.Stage(ctx, "a.txt", "staged-new.txt"))
	before := testRun(t, g, "ls-files", "--stage", "-z")
	writeGitFile(t, g, "a.txt", "unstaged\n")
	writeGitFile(t, g, "staged-new.txt", "unstaged new\n")
	writeGitFile(t, g, "scratch/nested.txt", "discard\n")
	writeGitFile(t, g, "ignored/keep.txt", "keep ignored\n")
	external := filepath.Join(t.TempDir(), "keep.txt")
	os.WriteFile(external, []byte("keep external"), 0600)
	linked := os.Symlink(filepath.Dir(external), filepath.Join(g.Dir, "outside-link")) == nil
	requireNoError(t, g.DiscardAll(ctx))
	if after := testRun(t, g, "ls-files", "--stage", "-z"); after != before {
		t.Fatal("discard altered index entries")
	}
	for name, want := range map[string]string{"a.txt": "staged\n", "staged-new.txt": "staged new\n", "ignored/keep.txt": "keep ignored\n"} {
		b, e := os.ReadFile(filepath.Join(g.Dir, name))
		requireNoError(t, e)
		if string(b) != want {
			t.Fatalf("%s = %q", name, b)
		}
	}
	if _, e := os.Stat(filepath.Join(g.Dir, "scratch/nested.txt")); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("untracked file retained", e)
	}
	if linked {
		if _, e := os.Lstat(filepath.Join(g.Dir, "outside-link")); !errors.Is(e, os.ErrNotExist) {
			t.Fatal("untracked link retained", e)
		}
	}
	b, e := os.ReadFile(external)
	requireNoError(t, e)
	if string(b) != "keep external" {
		t.Fatal("clean followed external symlink")
	}
	requireNoError(t, g.DiscardAll(ctx))
	if after := testRun(t, g, "ls-files", "--stage", "-z"); after != before {
		t.Fatal("repeated discard altered index")
	}
}

// Intentional safety difference D02: original discard_paths discards both
// command errors. The Go port reports failure instead of claiming a mutation.
func TestGitDiscardFailuresDoNotPretendSuccess(t *testing.T) {
	g := testGit(t)
	testCommit(t, g, "a", "before", "initial")
	writeGitFile(t, g, "a", "unsaved")
	broken := g
	broken.Binary = filepath.Join(t.TempDir(), "missing-git")
	if e := broken.Discard(context.Background(), "a"); e == nil {
		t.Fatal("missing binary falsely succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := g.DiscardAll(ctx); !errors.Is(e, context.Canceled) {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(g.Dir, "a"))
	requireNoError(t, e)
	if string(b) != "unsaved" {
		t.Fatal("failed/cancelled discard changed content")
	}
}

// Intentional safety difference D01: do not remove the only in-memory copy of
// a draft before the disk write is known to have succeeded.
func TestEditorFailedCloseKeepsDraftAndCanRetry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "draft.txt")
	os.WriteFile(path, []byte("old"), 0600)
	e := NewEditor()
	e.delay = time.Hour
	defer e.CloseAll()
	e.Open(path)
	e.SetText(path, "user draft")
	e.writeFile = func(string, string) error { return errors.New("fixture write denied") }
	if err := e.Close(path); err == nil {
		t.Fatal("failed close succeeded")
	}
	docs := e.Documents()
	if len(docs) != 1 || !docs[0].Dirty || docs[0].Text != "user draft" || docs[0].LastError == "" {
		t.Fatal("draft lost", docs)
	}
	e.writeFile = SaveExisting
	if err := e.Close(path); err != nil {
		t.Fatal(err)
	}
	if len(e.Documents()) != 0 {
		t.Fatal("successful retry did not close")
	}
	b, err := os.ReadFile(path)
	requireNoError(t, err)
	if string(b) != "user draft" {
		t.Fatal("retry saved wrong content")
	}
}
