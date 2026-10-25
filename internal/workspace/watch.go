package workspace

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type WorkspaceChange struct {
	Root string
	At   time.Time
}
type Watcher struct {
	Changes <-chan WorkspaceChange
	cancel  context.CancelFunc
	done    chan struct{}
	once    sync.Once
}

func (w *Watcher) Close() {
	if w == nil {
		return
	}
	w.once.Do(w.cancel)
	<-w.done
}

// WatchWorkspace uses a bounded portable metadata poll, coalescing changes.
// It includes linked-worktree and common Git metadata, but excludes object,
// log, checkpoint and generated trees. Call Close to release its goroutine.
func WatchWorkspace(parent context.Context, root string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	ctx, cancel := context.WithCancel(parent)
	ch := make(chan WorkspaceChange, 1)
	w := &Watcher{Changes: ch, cancel: cancel, done: make(chan struct{})}
	go func() {
		defer close(w.done)
		defer close(ch)
		dirs := []string{root}
		for _, arg := range []string{"--absolute-git-dir", "--git-common-dir"} {
			raw, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", arg).Output()
			if err == nil {
				p := strings.TrimSpace(string(raw))
				if !filepath.IsAbs(p) {
					p = filepath.Join(root, p)
				}
				dirs = append(dirs, p)
			}
		}
		last := watchState(ctx, dirs)
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-tick.C:
				next := watchState(ctx, dirs)
				if !sameWatchState(last, next) {
					last = next
					select {
					case ch <- WorkspaceChange{root, now}:
					default:
					}
				}
			}
		}
	}()
	return w
}

type fileStamp struct {
	size     int64
	modified int64
	mode     fs.FileMode
}

func watchState(ctx context.Context, roots []string) map[string]fileStamp {
	out := map[string]fileStamp{}
	for rootIndex, root := range roots {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if ctx.Err() != nil {
				return fs.SkipAll
			}
			if err != nil {
				return nil
			}
			if len(out) >= MaxTreeEntries {
				return fs.SkipAll
			}
			if p != root && rootIndex > 0 {
				rel, _ := filepath.Rel(root, p)
				parts := strings.Split(filepath.ToSlash(rel), "/")
				keep := false
				switch parts[0] {
				case "HEAD", "index", "ORIG_HEAD", "packed-refs", "config", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD":
					keep = len(parts) == 1
				case "refs":
					keep = len(parts) == 1 || (len(parts) > 1 && (parts[1] == "heads" || parts[1] == "remotes" || parts[1] == "tags"))
				case "rebase-merge", "rebase-apply":
					keep = true
				}
				if !keep {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
			if p != root && d.IsDir() {
				switch d.Name() {
				case ".git", "node_modules", "target", "dist", "build", ".next", "__pycache__", ".venv", "venv", ".cache", ".gradle", "DerivedData":
					return fs.SkipDir
				}
			}
			if d.Name() == ".DS_Store" || strings.HasSuffix(d.Name(), "~") || strings.HasSuffix(d.Name(), ".swp") || strings.HasSuffix(d.Name(), ".swx") {
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			// Directory existence matters to Explorer even when it is empty.
			// Ignore directory mtimes: excluded build/checkpoint writes change
			// their parents, and must not cause self-triggered refresh loops.
			if d.IsDir() {
				out[p] = fileStamp{mode: os.ModeDir}
				return nil
			}
			fi, e := d.Info()
			if e == nil && !d.IsDir() {
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
