package gitworkbench

import (
	"fmt"
	"github.com/egoist/mygo/ui"
)

// Native menus deliver their choice by element identity to a rebuilt frame,
// not to the closure that opened them. Key the owning button by repository job
// epoch, dataset and logical target so stale choices cannot bind a new target.
// The normal same-snapshot frame keeps its identity and menu functionality.
func (s *State) menuContext(c *ui.Context, kind, target string, dataset any) *ui.Context {
	return c.Key(fmt.Sprintf("git-menu:%s:%d:%s:%p:%s", s.Git.Dir, s.generation, kind, dataset, target))
}
func menuDataset[T any](rows []T) *T {
	if len(rows) == 0 {
		return nil
	}
	return &rows[0]
}

// Navigation from a list-row menu must not replace the ListState while the
// native virtual list is still building rows. Apply it at the next View entry.
type historyNavigation struct {
	workspace  string
	generation uint64
	path       string
	follow     bool
}

func (s *State) queueFileHistory(c *ui.Context, path string, follow bool) {
	s.pendingHistory = &historyNavigation{s.Git.Dir, s.generation, path, follow}
	c.AnimationFrame()
}
func (s *State) applyPendingHistory() {
	n := s.pendingHistory
	s.pendingHistory = nil
	if n == nil || n.workspace != s.Git.Dir || n.generation != s.generation {
		return
	}
	s.HistoryOptions.Path = n.path
	if n.follow {
		s.HistoryOptions.Follow = true
		s.HistoryOptions.Limit = 80
		s.CommitDetail = nil
	}
	if s.Tab != 1 {
		s.SetTab(1)
	} else {
		s.Refresh()
	}
}
