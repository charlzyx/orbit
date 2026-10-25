package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"path/filepath"
)

// Settings and Worktrees share the same persisted Rust-compatible configuration.
// Apply it on the UI thread, after successful persistence and again on entry.
// A running worktree job retains its already-copied configuration.
func (x *Controller) reloadWorktreePreferences() {
	if x.Store == nil || x.Worktrees == nil {
		return
	}
	cfg := workspace.LoadWorktreeConfig(filepath.Join(x.Store.RootDir, "worktrees.json"))
	cfg.AllowedSetupRepos = append([]string(nil), cfg.AllowedSetupRepos...)
	x.Worktrees.Config = cfg
}
