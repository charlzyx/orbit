package orbitui

import "github.com/egoist/mygo/ui"

// Source sidepane.rs keeps one dock restore point and a separate transient
// window ceiling. Only an explicit drag writes the persisted layout width.
type reviewComposerSelection struct {
	valid, focused bool
	session, draft string
	anchor, caret  int
}

type reviewLayoutState struct {
	fileFocused                                       bool
	filePath, fileWorkspace                           string
	content                                           ui.Handle
	composer                                          reviewComposerSelection
	restoreComposer                                   bool
	initialized, full, featureWasOpen, fullWasVisible bool
	width, restore, available, configured             float32
	resizeDirty                                       bool
}

func (a *App) initReviewLayout() {
	s := &a.reviewState.layout
	if !s.initialized || s.configured != a.ReviewWidth {
		w := a.ReviewWidth
		if w <= 0 {
			w = 460
		}
		w = max(float32(300), w)
		s.width, s.restore, s.configured = w, w, a.ReviewWidth
		s.initialized = true
	}
}
func (a *App) reviewFull() bool {
	return a.ReviewVisible && a.page != "settings" && a.page != "usage" && a.reviewState.layout.full
}
func (a *App) syncReviewLayout(c *ui.Context) {
	a.initReviewLayout()
	s := &a.reviewState.layout
	feature := a.page == "files" || a.page == "git" || a.page == "usage" || a.page == "worktrees"
	visible := a.ReviewVisible && a.page != "settings" && a.page != "usage"
	if visible && feature && !s.featureWasOpen && s.full {
		a.leaveReviewFull()
	}
	s.featureWasOpen = feature
	fullVisible := visible && s.full
	if fullVisible && !s.fullWasVisible {
		a.SidebarVisible = true
	}
	if s.fullWasVisible && !fullVisible {
		s.restoreComposer = a.page == "chat"
		if a.page == "files" && s.fileFocused && s.fileWorkspace == a.Model.Workspace && a.files != nil && !a.reviewState.handle.FocusWithin(c) {
			if d, ok := a.files.editor.Active(); ok && d.Content.Path == s.filePath {
				a.files.focusPending = s.filePath
			}
		}
	}
	s.fullWasVisible = fullVisible
	w, _ := c.Size()
	if a.SidebarVisible {
		w -= a.SidebarWidth
	}
	if a.ExplorerVisible && !a.ReviewVisible {
		w -= max(float32(200), a.ExplorerWidth)
	}
	s.available = max(float32(300), w)
	if s.full {
		s.width = s.available
	} else if s.width > s.available {
		s.width = s.available
	}
}
func (a *App) leaveReviewFull() {
	s := &a.reviewState.layout
	if !s.full {
		return
	}
	s.full = false
	s.width = max(float32(300), s.restore)
}
func (a *App) toggleReviewFull() {
	a.initReviewLayout()
	s := &a.reviewState.layout
	if s.full {
		a.leaveReviewFull()
		return
	}
	s.restore = s.width
	s.full = true
	s.width = max(float32(300), s.available)
	a.SidebarVisible = true
	a.autocomplete.Dismiss()
}
func (a *App) setReviewDragWidth(width, viewportWidth float32) {
	a.initReviewLayout()
	s := &a.reviewState.layout
	s.full = false
	width = max(float32(300), min(width, viewportWidth-480, float32(900)))
	if width == s.width {
		return
	}
	s.width, s.restore = width, width
	a.ReviewWidth = width
	s.configured = width
	a.Layout.SidepaneWidth = float64(width)
	s.resizeDirty = true
}
func (a *App) persistReviewDrag() {
	s := &a.reviewState.layout
	if !s.resizeDirty {
		return
	}
	s.resizeDirty = false
	layout := a.Layout
	a.Async(func() func() { err := a.Options.Store.SaveLayout(layout); return func() { a.save(err) } })
}
func (a *App) reviewResizeHandle(c *ui.Context) {
	t := a.Theme
	h := ui.Box(c.Key("review-resize")).Label("Resize review").Absolute().Left(9).Top(0).Bottom(0).Width(6).Cursor(ui.CursorResizeEW)
	if h.Hovered() {
		h.Background(t.Accent.Alpha(.4))
	}
	if dx, _, dragging := h.Dragged(); dragging && dx != 0 {
		w, _ := c.Size()
		a.setReviewDragWidth(a.reviewPaneWidth()-dx, w)
		c.AnimationFrame()
	}
	if _, _, ended := h.DragEnded(); ended {
		a.persistReviewDrag()
	}
}

// A remounted composer must not steal focus from the Review control the user
// is still using. Ordinary first appearance keeps its existing autofocus.
func (a *App) reviewComposerAutoFocus(c *ui.Context) bool {
	if a.reviewFull() {
		return false
	}
	s := &a.reviewState.layout
	p := s.composer
	if s.restoreComposer && p.valid && p.session == a.Model.SessionID && p.draft == a.Draft {
		return p.focused && !a.reviewState.handle.FocusWithin(c)
	}
	return true
}

// Retain only the same session/draft's caret while full-page Review hides
// the composer. This never changes ordinary input, IME or autocomplete edits.
func (a *App) retainReviewComposer(c *ui.Context, input ui.Element) {
	if a.reviewFull() {
		return
	}
	s := &a.reviewState.layout
	if s.restoreComposer {
		previous := s.composer
		s.restoreComposer = false
		if previous.valid && previous.session == a.Model.SessionID && previous.draft == a.Draft {
			input.SetTextSelection(previous.anchor, previous.caret)
			if previous.focused && !a.reviewState.handle.FocusWithin(c) {
				input.Focus()
			}
		}
	}
	anchor, caret := input.TextAnchorCaret()
	s.composer = reviewComposerSelection{valid: true, focused: input.Focused(), session: a.Model.SessionID, draft: a.Draft, anchor: anchor, caret: caret}
}

func (a *App) trackReviewFileFocus(c *ui.Context) {
	if a.reviewFull() {
		return
	}
	s := &a.reviewState.layout
	s.fileFocused = false
	if a.page == "files" && a.files != nil {
		if d, ok := a.files.editor.Active(); ok {
			s.fileFocused = a.files.handle.FocusWithin(c)
			s.filePath, s.fileWorkspace = d.Content.Path, a.Model.Workspace
		}
	}
}
