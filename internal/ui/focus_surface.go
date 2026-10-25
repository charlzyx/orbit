package orbitui

import "github.com/egoist/mygo/ui"

// Only current-build handles participate. Native text editors and the terminal
// own Tab; this does not install a window shortcut while either has focus.
type workbenchFocusState struct{ terminal, terminalRestart, editor, emptySidebar ui.Handle }

func (a *App) surfaceTraversal(c *ui.Context) bool {
	if a.page != "chat" && a.page != "files" || a.modalOpen() {
		return false
	}
	terminal := &a.surfaceFocus.terminal
	terminalEligible, terminalExited := false, false
	if p := a.Options.Terminal; p != nil {
		s := p.Snapshot()
		if s.Open && s.Active >= 0 && s.Active < len(s.Tabs) {
			t := s.Tabs[s.Active]
			terminalEligible = !t.Loading && t.Error == ""
			terminalExited = t.Exited
			if terminalExited {
				terminal = &a.surfaceFocus.terminalRestart
			}
		}
	}
	if !terminalExited && a.surfaceFocus.terminal.FocusWithin(c) || a.surfaceFocus.editor.FocusWithin(c) || a.searchInput.FocusWithin(c) {
		return false
	}
	completionOpen := a.autocomplete.IsOpen() && a.composer.FocusWithin(c)
	sidebar := &a.sessionList.Handle
	if len(a.sidebarRows()) == 0 {
		sidebar = &a.surfaceFocus.emptySidebar
	}
	handles := []*ui.Handle{&a.composer, sidebar, &a.reviewState.treeHandle, terminal}
	if a.files != nil {
		handles = append(handles, &a.files.treeHandle, &a.surfaceFocus.editor, &a.files.handle)
	}
	visible := make([]*ui.Handle, 0, len(handles))
	current := -1
	for _, h := range handles {
		if a.reviewFull() && (h == &a.composer || h == &a.surfaceFocus.editor || h == terminal || a.files != nil && h == &a.files.handle) {
			continue
		}
		if !c.Resolve(*h).Valid() {
			continue
		}
		if h == terminal && !terminalEligible {
			continue
		}
		// Source editor and preview root represent the same viewer surface.
		if a.files != nil && h == &a.files.handle && c.Resolve(a.surfaceFocus.editor).Valid() {
			continue
		}
		if h == sidebar && !a.SidebarVisible {
			continue
		}
		if h.FocusWithin(c) || h == terminal && terminalExited && a.surfaceFocus.terminal.FocusWithin(c) {
			current = len(visible)
		}
		visible = append(visible, h)
	}
	// Inputs, popover/menu controls and unrelated pages retain toolkit traversal.
	if current < 0 || len(visible) == 0 {
		return false
	}
	direction := 0
	if !completionOpen && c.Shortcut(0, ui.KeyTab) {
		direction = 1
	} else if c.Shortcut(ui.Shift, ui.KeyTab) {
		direction = -1
	}
	if direction == 0 {
		return false
	}
	visible[(current+direction+len(visible))%len(visible)].Focus(c)
	return true
}
