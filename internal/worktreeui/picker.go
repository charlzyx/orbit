package worktreeui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"strings"
)

// WorkInRow is derived only from Git's worktree list, never placeholder paths.
type WorkInRow struct {
	Label, Path      string
	Selected, Create bool
}

func (s *State) WorkInRows() []WorkInRow {
	if len(s.Rows) == 0 {
		return nil
	}
	active := -1
	for i, w := range s.Rows {
		if workspace.SamePath(w.Path, s.Workspace) {
			active = i
			break
		}
	}
	if active < 0 {
		for i, w := range s.Rows {
			if w.Main {
				rel, e := filepath.Rel(w.Path, s.Workspace)
				if e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					active = i
				}
			}
		}
	}
	rows := []WorkInRow{}
	for i, w := range s.Rows {
		if w.Main {
			rows = append(rows, WorkInRow{Label: "Local", Path: w.Path, Selected: active == i})
		}
	}
	for i, w := range s.Rows {
		if !w.Main {
			rows = append(rows, WorkInRow{Label: w.Name, Path: w.Path, Selected: active == i})
		}
	}
	return append(rows, WorkInRow{Label: "New worktree…", Create: true})
}

// WorkInPicker is a native anchored menu. MyGo owns arrow/Enter/Escape and
// restores focus to its trigger; onCreate should reveal the Worktrees page.
func WorkInPicker(c *ui.Context, s *State, onCreate func()) ui.Element {
	s.Tick()
	if !s.initialized && !s.Busy && !s.closed {
		s.Refresh()
	}
	rows := s.WorkInRows()
	if len(rows) == 0 {
		return ui.Box(c).Size(0, 0)
	}
	label := s.text("worktree.work_in.local", "Local")
	for i, row := range rows {
		if row.Selected {
			label = row.Label
			if i == 0 {
				label = s.text("worktree.work_in.local", "Local")
			}
		}
	}
	return ui.MenuButton(c, s.text("worktree.work_in.title", "Work in")+": "+label, func(m *ui.Menu) {
		for i, r := range rows {
			if r.Create {
				m.Separator()
			}
			title := r.Label
			if r.Create {
				title = s.text("worktree.work_in.new", "New worktree…")
			} else if i == 0 {
				title = s.text("worktree.work_in.local", "Local")
			}
			if m.Item(title).Checked(r.Selected).Disabled(s.Busy || (!r.Create && s.Options.OpenWorkspace == nil)).Chosen() {
				if r.Create {
					s.quickCreateDialog()
					if onCreate != nil {
						onCreate()
					}
				} else {
					s.Options.OpenWorkspace(r.Path)
				}
			}
		}
	}).Tooltip("Choose a working tree")
}
