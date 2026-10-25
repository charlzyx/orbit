package workspace

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestSnapshotIgnoresHiddenSymlinkAndNestedRules(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"src", "empty", "build", "src/ignored"} {
		os.MkdirAll(filepath.Join(root, d), 0700)
	}
	for p, s := range map[string]string{".gitignore": "build/\n*.tmp\n!keep.tmp\n", ".hidden": "x", "a.tmp": "x", "keep.tmp": "x", "src/ok.go": "package p", "src/.ignore": "ignored/\n", "src/ignored/a": "x", "build/output": "x"} {
		os.WriteFile(filepath.Join(root, p), []byte(s), 0600)
	}
	if runtime.GOOS != "windows" {
		os.Symlink(root, filepath.Join(root, "src", "loop"))
	}
	s, err := Snapshot(context.Background(), root, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.Files, []string{"keep.tmp", "src/ok.go"}) {
		t.Fatal(s)
	}
	if !reflect.DeepEqual(s.Dirs, []string{"empty", "src"}) {
		t.Fatal(s)
	}
	shown, err := Snapshot(context.Background(), root, true)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range shown.Files {
		if p == ".hidden" {
			found = true
		}
	}
	if !found {
		t.Fatal(shown)
	}
}
func TestSnapshotCancelledAndMissing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e := Snapshot(ctx, t.TempDir(), false); e == nil {
		t.Fatal("cancel")
	}
	if _, e := Snapshot(context.Background(), filepath.Join(t.TempDir(), "missing"), false); e == nil {
		t.Fatal("missing")
	}
}
func TestTreeRowsFilterKeepsAncestors(t *testing.T) {
	s := TreeSnapshot{Dirs: []string{"src", "src/deep", "empty"}, Files: []string{"README.md", "src/deep/a.go"}}
	rows := TreeRows(s, nil, "a.go", nil)
	if len(rows) != 3 || rows[0].Path != "src" || rows[2].Depth != 2 {
		t.Fatal(rows)
	}
	rows = TreeRows(s, nil, "", nil)
	if len(rows) != 3 || !rows[0].Directory {
		t.Fatal(rows)
	}
}
func TestWatcherReportsChangeAndCloses(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "a")
	os.WriteFile(p, []byte("a"), 0600)
	w := WatchWorkspace(context.Background(), root, 20*time.Millisecond)
	time.Sleep(80 * time.Millisecond)
	os.WriteFile(p, []byte("changed"), 0600)
	select {
	case <-w.Changes:
	case <-time.After(2 * time.Second):
		t.Fatal("missing change")
	}
	w.Close()
	w.Close()
}

func TestSnapshotAncestorIgnoreRelativePatterns(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "nested", "project")
	os.MkdirAll(root, 0700)
	os.WriteFile(filepath.Join(parent, ".ignore"), []byte("/nested/project/excluded\n*.tmp\n"), 0600)
	for _, name := range []string{"excluded", "test.tmp", "keep.txt"} {
		os.WriteFile(filepath.Join(root, name), []byte("x"), 0600)
	}
	s, e := Snapshot(context.Background(), root, false)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(s.Files, []string{"keep.txt"}) {
		t.Fatal(s)
	}
}

func TestWatcherDoesNotObserveCheckpointScratchOrRefs(t *testing.T) {
	root := wtTestRepo(t)
	gitdir := filepath.Join(root, ".git")
	roots := []string{root, gitdir}
	before := watchState(context.Background(), roots)
	for _, p := range []string{"orbit-checkpoint-index-test", "orbit-checkpoint-index-test.lock", "refs/orbit/session-test-turn-1", "objects/test/object", "logs/HEAD"} {
		path := filepath.Join(gitdir, p)
		os.MkdirAll(filepath.Dir(path), 0700)
		os.WriteFile(path, []byte("scratch"), 0600)
	}
	after := watchState(context.Background(), roots)
	if !sameWatchState(before, after) {
		t.Fatal("self-refresh loop from internal Git writes")
	}
	os.WriteFile(filepath.Join(gitdir, "MERGE_HEAD"), []byte("head"), 0600)
	if sameWatchState(after, watchState(context.Background(), roots)) {
		t.Fatal("operation metadata ignored")
	}
}
func TestWorktreeMutationLockCancellation(t *testing.T) {
	root := wtTestRepo(t)
	m := &WorktreeManager{Dir: root}
	unlock, e := m.lockMutation(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	defer unlock()
	other := &WorktreeManager{Dir: root}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, e = other.Create(ctx, CreateWorktreeOptions{Name: "cancelled", Branch: "cancelled", NewBranch: true}); e == nil {
		t.Fatal("cancelled lock wait mutated")
	}
	if _, e = os.Stat(filepath.Join(root, ".wt", "cancelled")); !os.IsNotExist(e) {
		t.Fatal("created during cancelled lock wait")
	}
}

func TestIgnoreFileOverridesGitRulesAndGlobalExcludes(t *testing.T) {
	root := wtTestRepo(t)
	os.MkdirAll(filepath.Join(root, "keepdir"), 0700)
	os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.tmp\nkeepdir/\n"), 0600)
	os.WriteFile(filepath.Join(root, ".ignore"), []byte("!keep.tmp\n!keepdir/\n"), 0600)
	for _, p := range []string{"keep.tmp", "skip.tmp", "keepdir/file"} {
		os.WriteFile(filepath.Join(root, p), []byte("x"), 0600)
	}
	s, e := Snapshot(context.Background(), root, false)
	if e != nil {
		t.Fatal(e)
	}
	want := []string{"keep.tmp", "keepdir/file", "seed.txt"}
	if !reflect.DeepEqual(s.Files, want) {
		t.Fatal(s.Files)
	}
}

func TestSnapshotPrunesGitGlobalAndInfoExcludesBeforeWalk(t *testing.T) {
	root := wtTestRepo(t)
	global := filepath.Join(t.TempDir(), "ignore")
	os.WriteFile(global, []byte("global-cache/\n*.global\n"), 0600)
	g := Git{Dir: root}
	if _, e := g.Run(context.Background(), "config", "core.excludesfile", global); e != nil {
		t.Fatal(e)
	}
	os.WriteFile(filepath.Join(root, ".git", "info", "exclude"), []byte("info-cache/\n"), 0600)
	for _, dir := range []string{"global-cache", "info-cache"} {
		os.MkdirAll(filepath.Join(root, dir), 0700)
		os.WriteFile(filepath.Join(root, dir, "hidden"), []byte("x"), 0600)
	}
	os.WriteFile(filepath.Join(root, "hidden.global"), []byte("x"), 0600)
	s, e := Snapshot(context.Background(), root, false)
	if e != nil || !reflect.DeepEqual(s.Files, []string{"seed.txt"}) || len(s.Dirs) != 0 {
		t.Fatal(s, e)
	}
}

func TestWatcherTracksEmptyDirectoriesAndOrdinaryLogsObjects(t *testing.T) {
	root := t.TempDir()
	before := watchState(context.Background(), []string{root})
	for _, name := range []string{"empty", "logs", "objects", "orbit-checkpoint-not-git"} {
		if err := os.Mkdir(filepath.Join(root, name), 0700); err != nil {
			t.Fatal(err)
		}
	}
	after := watchState(context.Background(), []string{root})
	if sameWatchState(before, after) {
		t.Fatal("empty directory changes not observed")
	}
	for _, name := range []string{"empty", "logs", "objects", "orbit-checkpoint-not-git"} {
		if _, ok := after[filepath.Join(root, name)]; !ok {
			t.Fatal("ordinary workspace directory ignored", name)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "node_modules"), 0700); err != nil {
		t.Fatal(err)
	}
	if !sameWatchState(after, watchState(context.Background(), []string{root})) {
		t.Fatal("generated directory triggered refresh")
	}
}
