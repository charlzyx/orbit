package worktreeui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
	"path/filepath"
	"strings"
	"time"
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

// WorkInPicker is a native anchored popup, not an operating-system menu. The
// source owns arrow/Enter/Escape behavior and restores composer focus on dismiss;
// a native MenuButton has no dismissal callback and cannot preserve that contract.
func WorkInPicker(c *ui.Context, s *State, onCreate func()) ui.Element {
	s.Tick()
	if !s.initialized && !s.Busy && !s.closed {
		s.Refresh()
	}
	rows := s.WorkInRows()
	if len(rows) == 0 || s.closed {
		s.workInOpen = false
		return ui.Box(c).Size(0, 0)
	}
	local := s.text("worktree.work_in.local", "Local")
	label := local
	for i, r := range rows {
		if r.Selected {
			label = r.Label
			if i == 0 {
				label = local
			}
		}
	}
	anchor := ui.Button(c.Key("work-in-trigger"), s.text("worktree.work_in.title", "Work in")+": "+label).Tooltip("Choose a working tree")
	opened := false
	if anchor.Clicked() {
		if s.workInOpen {
			s.closeWorkIn(true)
		} else if s.workInDismissedAt.IsZero() || c.Now().Sub(s.workInDismissedAt) >= 200*time.Millisecond {
			s.workInOpen = true
			opened = true
			s.workInCursor = 0
			for i, r := range rows {
				if r.Selected {
					s.workInCursor = i
					break
				}
			}
			if s.Options.OnPickerOpen != nil {
				s.Options.OnPickerOpen()
			}
		}
	}
	// An out-of-band dialog must not leave a keyboard popup above it.
	if s.HasModal() {
		s.workInOpen = false
	}
	popup := ui.Popover(c.Key("work-in-popup"), anchor, &s.workInOpen, func() {
		panel := ui.Column(c.Key("work-in-rows")).Focusable().MinWidth(220).MaxWidth(420).Gap(2)
		if opened {
			panel.Focus()
		}
		if panel.Shortcut(0, ui.KeyDown) {
			s.workInCursor = (s.workInCursor + 1) % len(rows)
		}
		if panel.Shortcut(0, ui.KeyUp) {
			s.workInCursor = (s.workInCursor + len(rows) - 1) % len(rows)
		}
		s.workInCursor = max(0, min(s.workInCursor, len(rows)-1))
		if panel.Shortcut(0, ui.KeyEnter) {
			s.chooseWorkIn(rows[s.workInCursor], onCreate)
		}
		panel.Children(func() {
			for i, r := range rows {
				title := r.Label
				if r.Create {
					ui.Box(c).Height(1).FillWidth().Background(c.Theme().Border)
					title = s.text("worktree.work_in.new", "New worktree…")
				} else if i == 0 {
					title = local
				}
				item := ui.Button(c.Key("work-in-"+r.Path+title), title).AlignSelf(ui.Stretch).Disabled(s.Busy || (!r.Create && s.Options.OpenWorkspace == nil))
				if i == s.workInCursor {
					item.Background(c.Theme().SurfaceHover)
				}
				if item.Clicked() {
					s.workInCursor = i
					s.chooseWorkIn(r, onCreate)
				}
			}
		})
	})
	if popup.Dismissed() {
		if popup.PressedOutside() {
			s.workInDismissedAt = c.Now()
		}
		s.closeWorkIn(true)
	}
	return anchor
}
func (s *State) closeWorkIn(restore bool) {
	s.workInOpen = false
	if restore && !s.HasModal() && s.Options.FocusComposer != nil {
		s.Options.FocusComposer()
	}
}
func (s *State) chooseWorkIn(row WorkInRow, onCreate func()) {
	if s.Busy {
		return
	}
	s.workInOpen = false
	if row.Create {
		s.quickCreateDialog()
		if onCreate != nil {
			onCreate()
		}
		return // The new dialog, not the composer, owns focus.
	}
	if !row.Selected && s.Options.OpenWorkspace != nil {
		s.Options.OpenWorkspace(row.Path)
	}
}
