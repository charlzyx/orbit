package orbitui

import "fmt"

// This is a UI confirmation intent, not permission to delete. The backend
// retains its active/live/busy/path ownership checks at actual execution.
type paletteDeleteIntent struct {
	path, title, session, runtime, workspace string
}

func (i *paletteDeleteIntent) owns(s Snapshot) bool {
	return i.path == s.SessionPath && i.session == s.SessionID && i.runtime == s.RuntimeID && i.workspace == s.Workspace
}
func (a *App) cancelPaletteDelete() {
	if a.paletteDelete == nil {
		return
	}
	a.paletteDelete = nil
	a.confirmOpen = false
	a.confirmCommit = nil
}
func (a *App) invalidatePaletteDelete(s Snapshot) {
	if i := a.paletteDelete; i != nil && !i.owns(s) {
		a.cancelPaletteDelete()
	}
}
func (a *App) confirmPaletteDelete() {
	if a.Model.SessionPath == "" {
		return
	}
	i := &paletteDeleteIntent{path: a.Model.SessionPath, title: a.Model.SessionName, session: a.Model.SessionID, runtime: a.Model.RuntimeID, workspace: a.Model.Workspace}
	title := i.title
	if title == "" {
		title = i.path
	}
	a.Confirm("Delete task?", fmt.Sprintf("Delete %q?\n%s\nThis removes the saved session file. Sessions with a live runtime cannot be deleted.", title, i.path), func() {
		if a.paletteDelete != i || !i.owns(a.Model) {
			return
		}
		a.paletteDelete = nil // one-shot, before dispatch can trigger reentrant updates
		a.Dispatch(Action{Kind: "DeleteSession", Value: i.path, SessionID: i.session})
	})
	a.paletteDelete = i
}
