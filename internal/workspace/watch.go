package workspace

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WorkspaceChange struct {
	Root string
	At   time.Time
}
type Watcher struct {
	Changes <-chan WorkspaceChange
	// Ready closes after initial registrations, or the fallback baseline, exist.
	Ready  <-chan struct{}
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
	native atomic.Bool
}

func (w *Watcher) Close() {
	if w == nil {
		return
	}
	w.once.Do(w.cancel)
	<-w.done
}

// Native reports the active mechanism, not a guarantee of notifications on a
// remote filesystem. Network filesystems may require the explicit poll fallback.
func (w *Watcher) Native() bool { return w != nil && w.native.Load() }

type nativeWatcher struct {
	events <-chan fsnotify.Event
	errors <-chan error
	add    func(string) error
	remove func(string) error
	close  func() error
}
type watchFactory func() (*nativeWatcher, error)

func newNativeWatcher() (*nativeWatcher, error) {
	w, err := fsnotify.NewBufferedWatcher(256)
	if err != nil {
		return nil, err
	}
	return &nativeWatcher{w.Events, w.Errors, w.Add, w.Remove, w.Close}, nil
}

// WatchWorkspace follows the workspace and linked/common Git metadata using
// native directory notifications, recursively registered and trailing-debounced.
// Generated/object/checkpoint trees are excluded. If native registration, limits,
// or event delivery fail, it closes the backend and uses bounded metadata polling.
// That fallback cannot detect changes preserving both size and mtime. No system
// watch limits or other security-sensitive settings are changed by this code.
func WatchWorkspace(parent context.Context, root string, interval time.Duration) *Watcher {
	return watchWorkspace(parent, root, interval, newNativeWatcher)
}
func watchWorkspace(parent context.Context, root string, interval time.Duration, factory watchFactory) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	ctx, cancel := context.WithCancel(parent)
	changes := make(chan WorkspaceChange, 1)
	ready := make(chan struct{})
	w := &Watcher{Changes: changes, Ready: ready, cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(w.done)
		defer close(changes)
		defer w.native.Store(false)
		if strings.TrimSpace(root) == "" {
			close(ready)
			<-ctx.Done()
			return
		}
		roots := workspaceWatchRoots(ctx, root)
		emit := func() {
			select {
			case changes <- WorkspaceChange{root, time.Now()}:
			default:
			}
		}
		nw, err := factory()
		if err == nil {
			registrations := map[string]bool{}
			err = registerWatchTree(ctx, nw, roots, registrations, "")
			if err == nil {
				w.native.Store(true)
				close(ready)
				err = runNativeWatch(ctx, nw, roots, registrations, interval, emit)
				w.native.Store(false)
				_ = nw.close()
				if ctx.Err() != nil {
					return
				}
				// A lost/overflowing event stream may have omitted a real change. Force
				// exactly one refresh after establishing the fallback polling baseline.
				roots = workspaceWatchRoots(ctx, root)
				pollWorkspace(ctx, roots, interval, emit, emit)
				return
			}
			_ = nw.close()
		}
		pollWorkspace(ctx, roots, interval, emit, func() { close(ready) })
	}()
	return w
}
func workspaceWatchRoots(ctx context.Context, root string) []string {
	root = canonicalPath(root)
	roots := []string{root}
	g := Git{Dir: root}
	for _, arg := range []string{"--absolute-git-dir", "--git-common-dir"} {
		p, err := g.text(ctx, "rev-parse", arg)
		if err != nil || p == "" {
			continue
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		p = canonicalPath(p)
		seen := false
		for _, prior := range roots[1:] {
			seen = seen || prior == p
		}
		if !seen {
			roots = append(roots, p)
		}
	}
	return roots
}
func pollWorkspace(ctx context.Context, roots []string, interval time.Duration, emit func(), ready func()) {
	last := watchState(ctx, roots)
	if ready != nil {
		ready()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			next := watchState(ctx, roots)
			if !sameWatchState(last, next) {
				last = next
				emit()
			}
		}
	}
}

var errWatchLimit = errors.New("workspace native watch entry limit reached")

func registerWatchTree(ctx context.Context, nw *nativeWatcher, roots []string, registered map[string]bool, subtree string) error {
	count := 0
	for i, root := range roots {
		start := root
		if subtree != "" {
			if !pathUnder(subtree, root) {
				continue
			}
			start = subtree
		}
		err := filepath.WalkDir(start, func(p string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err != nil {
				return err
			}
			if !watchPathRelevant(root, p, i > 0, d.IsDir()) || d.Type()&os.ModeSymlink != 0 {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			count++
			if count > MaxTreeEntries {
				return errWatchLimit
			}
			if d.IsDir() && !registered[p] {
				if len(registered) >= MaxTreeEntries {
					return errWatchLimit
				}
				if err := nw.add(p); err != nil {
					return err
				}
				registered[p] = true
			}
			return nil
		})
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if len(registered) == 0 {
		return os.ErrNotExist
	}
	return nil
}
func nativePathRelevant(roots []string, path string, isDir bool) bool {
	// A linked worktree's private metadata lives under the common .git dir;
	// choose its more specific root before the common root's worktrees exclusion.
	match := -1
	for i, root := range roots {
		if pathUnder(path, root) && (match < 0 || len(root) >= len(roots[match])) {
			match = i
		}
	}
	if match < 0 {
		return false
	}
	return watchPathRelevant(roots[match], path, match > 0, isDir)
}
func runNativeWatch(ctx context.Context, nw *nativeWatcher, roots []string, registered map[string]bool, delay time.Duration, emit func()) error {
	var timer *time.Timer
	var deadline <-chan time.Time
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()
	dirty := func() {
		if timer == nil {
			timer = time.NewTimer(delay)
		} else {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(delay)
		}
		deadline = timer.C
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err, ok := <-nw.errors:
			if !ok {
				return errors.New("native watch errors channel closed")
			}
			if err != nil {
				return err
			}
		case event, ok := <-nw.events:
			if !ok {
				return errors.New("native watch events channel closed")
			}
			path := filepath.Clean(event.Name)
			// A folder can become a repository after the watch starts. A .git
			// directory or linked-worktree pointer introduces new metadata roots.
			if path == filepath.Join(roots[0], ".git") && event.Has(fsnotify.Create|fsnotify.Write) {
				roots = workspaceWatchRoots(ctx, roots[0])
				if len(roots) == 1 {
					if info, err := os.Stat(path); err == nil && info.IsDir() {
						roots = append(roots, canonicalPath(path))
					}
				}
				if err := registerWatchTree(ctx, nw, roots, registered, ""); err != nil {
					return err
				}
				dirty()
				continue
			}
			info, statErr := os.Lstat(path)
			isDir := registered[path] || (statErr == nil && info.IsDir())
			if statErr == nil && info.Mode()&os.ModeSymlink != 0 {
				continue
			}
			if !nativePathRelevant(roots, path, isDir) {
				continue
			}
			if event.Has(fsnotify.Remove|fsnotify.Rename) && registered[path] {
				for prior := range registered {
					if pathUnder(prior, path) {
						_ = nw.remove(prior)
						delete(registered, prior)
					}
				}
			}
			if event.Has(fsnotify.Remove | fsnotify.Rename) {
				for _, root := range roots {
					if path == root {
						return os.ErrNotExist
					}
				}
			}
			if event.Has(fsnotify.Create) && isDir {
				if err := registerWatchTree(ctx, nw, roots, registered, path); err != nil {
					return err
				}
			}
			// Directory Write/Chmod is parent-mtime noise on kqueue/Windows. Child
			// create/delete events provide the relevant change without build/git loops.
			if isDir && !event.Has(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) {
				continue
			}
			if event.Has(fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename | fsnotify.Chmod) {
				dirty()
			}
		case <-deadline:
			deadline = nil
			emit()
		}
	}
}

func watchPathRelevant(root, path string, metadata, isDir bool) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	if rel == "." {
		return true
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if metadata {
		switch parts[0] {
		case "HEAD", "index", "ORIG_HEAD", "packed-refs", "config", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD":
			return len(parts) == 1
		case "refs":
			return len(parts) == 1 || parts[1] == "heads" || parts[1] == "remotes" || parts[1] == "tags"
		case "rebase-merge", "rebase-apply":
			return true
		default:
			return false
		}
	}
	for _, part := range parts {
		switch part {
		case ".git", "node_modules", "target", "dist", "build", ".next", "__pycache__", ".venv", "venv", ".cache", ".gradle", "DerivedData":
			return false
		}
	}
	name := parts[len(parts)-1]
	return name != ".DS_Store" && !strings.HasSuffix(name, "~") && !strings.HasSuffix(name, ".swp") && !strings.HasSuffix(name, ".swx")
}

type fileStamp struct {
	size     int64
	modified int64
	mode     fs.FileMode
}

func watchState(ctx context.Context, roots []string) map[string]fileStamp {
	out := map[string]fileStamp{}
	for i, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil || len(out) >= MaxTreeEntries {
				return fs.SkipAll
			}
			if err != nil {
				return nil
			}
			if !watchPathRelevant(root, p, i > 0, d.IsDir()) || d.Type()&os.ModeSymlink != 0 {
				if d.IsDir() {
					return fs.SkipDir
				}
				return nil
			}
			// Track empty directory existence, but not mtimes from excluded children.
			if d.IsDir() {
				out[p] = fileStamp{mode: os.ModeDir}
				return nil
			}
			fi, err := d.Info()
			if err == nil {
				out[p] = fileStamp{fi.Size(), fi.ModTime().UnixNano(), fi.Mode()}
			}
			return nil
		})
	}
	return out
}
func sameWatchState(a, b map[string]fileStamp) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
