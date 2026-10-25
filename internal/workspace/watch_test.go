package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
)

func readyWatcher(t *testing.T, w *Watcher) {
	t.Helper()
	select {
	case <-w.Ready:
	case <-time.After(3 * time.Second):
		t.Fatal("watch initialization timed out")
	}
}
func expectWatchChange(t *testing.T, w *Watcher) {
	t.Helper()
	select {
	case _, ok := <-w.Changes:
		if !ok {
			t.Fatal("watch closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("missing workspace notification")
	}
}
func expectWatchQuiet(t *testing.T, w *Watcher, d time.Duration) {
	t.Helper()
	select {
	case event := <-w.Changes:
		t.Fatal("unexpected workspace notification", event)
	case <-time.After(d):
	}
}
func TestNativeWatchDetectsPreservedMetadataAndNewNestedTrees(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "same.txt")
	if err := os.WriteFile(path, []byte("before"), 0600); err != nil {
		t.Fatal(err)
	}
	stamp, _ := os.Stat(path)
	w := WatchWorkspace(context.Background(), root, 25*time.Millisecond)
	defer w.Close()
	readyWatcher(t, w)
	if !w.Native() {
		t.Skip("native filesystem notifications unavailable; fallback separately tested")
	}
	if err := os.WriteFile(path, []byte("after!"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, stamp.ModTime(), stamp.ModTime()); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	nested := filepath.Join(root, "new", "nested")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	if err := os.WriteFile(filepath.Join(nested, "child"), []byte("new file"), 0600); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	// Atomic replacement remains observable because directories, not files, are watched.
	if err := SaveExisting(path, "atomic"); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
}
func TestNativeWatchFollowsLinkedMetadataAndIgnoresCheckpointChurn(t *testing.T) {
	root := wtTestRepo(t)
	m := WorktreeManager{Dir: root}
	linked, err := m.Create(context.Background(), CreateWorktreeOptions{Name: "watched", Branch: "watched", NewBranch: true})
	if err != nil {
		t.Fatal(err)
	}
	w := WatchWorkspace(context.Background(), linked.Worktree.Path, 20*time.Millisecond)
	defer w.Close()
	readyWatcher(t, w)
	if !w.Native() {
		t.Skip("native filesystem notifications unavailable")
	}
	g := Git{Dir: linked.Worktree.Path}
	own, err := g.text(context.Background(), "rev-parse", "--absolute-git-dir")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(own, "MERGE_HEAD"), []byte("test metadata"), 0600); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	if _, err := g.Run(context.Background(), "update-ref", "refs/heads/external-watched", "HEAD"); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	for _, p := range []string{filepath.Join(root, ".git", "orbit-checkpoint-index-test"), filepath.Join(root, ".git", "refs", "orbit", "session-test-turn-1"), filepath.Join(root, ".git", "objects", "test", "object"), filepath.Join(linked.Worktree.Path, "node_modules", "generated")} {
		os.MkdirAll(filepath.Dir(p), 0700)
		if err := os.WriteFile(p, []byte("ignored"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	expectWatchQuiet(t, w, 120*time.Millisecond)
}
func TestWatchFallbackOnInitializationFailureAndCancellation(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	w := watchWorkspace(ctx, root, 10*time.Millisecond, func() (*nativeWatcher, error) { return nil, errors.New("fixture resource limit") })
	readyWatcher(t, w)
	if w.Native() {
		t.Fatal("failed native watcher reported active")
	}
	if err := os.Mkdir(filepath.Join(root, "empty"), 0700); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	cancel()
	w.Close()
	w.Close()
	if _, ok := <-w.Changes; ok {
		t.Fatal("changes not closed")
	}
}
func TestNativeWatchRootRecreationFallsBack(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "workspace")
	os.Mkdir(root, 0700)
	w := WatchWorkspace(context.Background(), root, 10*time.Millisecond)
	defer w.Close()
	readyWatcher(t, w)
	if !w.Native() {
		t.Skip("native filesystem notifications unavailable")
	}
	if err := os.Remove(root); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	deadline := time.Now().Add(time.Second)
	for w.Native() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if w.Native() {
		t.Fatal("dead root kept an unusable native watch")
	}
	os.Mkdir(root, 0700)
	os.WriteFile(filepath.Join(root, "restored"), []byte("restored"), 0600)
	expectWatchChange(t, w)
}

type fakeNativeWatch struct {
	events chan fsnotify.Event
	errors chan error
	mu     sync.Mutex
	paths  map[string]bool
	closed bool
}

func newFakeNative() *fakeNativeWatch {
	return &fakeNativeWatch{events: make(chan fsnotify.Event, 64), errors: make(chan error, 1), paths: map[string]bool{}}
}
func (f *fakeNativeWatch) factory() (*nativeWatcher, error) {
	return &nativeWatcher{events: f.events, errors: f.errors, add: func(p string) error { f.mu.Lock(); defer f.mu.Unlock(); f.paths[p] = true; return nil }, remove: func(p string) error { f.mu.Lock(); defer f.mu.Unlock(); delete(f.paths, p); return nil }, close: func() error { f.mu.Lock(); defer f.mu.Unlock(); f.closed = true; return nil }}, nil
}
func TestNativeWatchDebouncesBurstAndFallsBackOnOverflow(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.go")
	os.WriteFile(path, []byte("start"), 0600)
	fake := newFakeNative()
	w := watchWorkspace(context.Background(), root, 50*time.Millisecond, fake.factory)
	defer w.Close()
	readyWatcher(t, w)
	if !w.Native() {
		t.Fatal("fake watch not initialized")
	}
	for i := 0; i < 4; i++ {
		fake.events <- fsnotify.Event{Name: path, Op: fsnotify.Write}
		time.Sleep(15 * time.Millisecond)
	}
	expectWatchQuiet(t, w, 20*time.Millisecond)
	expectWatchChange(t, w)
	expectWatchQuiet(t, w, 70*time.Millisecond)
	fake.errors <- fsnotify.ErrEventOverflow
	expectWatchChange(t, w)
	if w.Native() {
		t.Fatal("overflow did not activate fallback")
	}
	// The overflow itself triggers a refresh; wait until the fallback has sampled
	// before introducing the next change so its polling delta is deterministic.
	time.Sleep(70 * time.Millisecond)
	os.WriteFile(path, []byte("changed during polling"), 0600)
	expectWatchChange(t, w)
	fake.mu.Lock()
	closed := fake.closed
	fake.mu.Unlock()
	if !closed {
		t.Fatal("failed native backend leaked")
	}
}
func TestEmptyWorkspaceDoesNotWatchProcessDirectory(t *testing.T) {
	called := false
	w := watchWorkspace(context.Background(), "", time.Millisecond, func() (*nativeWatcher, error) { called = true; return nil, errors.New("should not run") })
	readyWatcher(t, w)
	w.Close()
	if called {
		t.Fatal("empty workspace initialized a filesystem watch")
	}
}

func TestNativeWatchDiscoversGitCreatedAfterStartup(t *testing.T) {
	root := t.TempDir()
	w := WatchWorkspace(context.Background(), root, 20*time.Millisecond)
	defer w.Close()
	readyWatcher(t, w)
	if !w.Native() {
		t.Skip("native filesystem notifications unavailable")
	}
	if _, err := (Git{Dir: root}).Run(context.Background(), "init", "-b", "main"); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
	if err := os.WriteFile(filepath.Join(root, ".git", "HEAD"), []byte("ref: refs/heads/changed\n"), 0600); err != nil {
		t.Fatal(err)
	}
	expectWatchChange(t, w)
}
