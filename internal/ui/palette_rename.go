package orbitui

import "github.com/charlzyx/orbit/internal/agent"

// A rename prompt belongs to its displayed task, not the task active at Save.
// Backend target/lease checks still apply after this UI intent is consumed.
type paletteRenameIntent struct {
	path, name, session, runtime, workspace string
	target                                  *agent.SessionRenameTarget
}

func (i *paletteRenameIntent) owns(s Snapshot) bool {
	return i.path == s.SessionPath && i.session == s.SessionID && i.runtime == s.RuntimeID && i.workspace == s.Workspace
}
func (a *App) cancelPaletteRename() {
	if a.paletteRename == nil {
		return
	}
	a.paletteRename = nil
	a.promptOpen = false
	a.promptCommit = nil
}
func (a *App) invalidatePaletteRename(s Snapshot) {
	if i := a.paletteRename; i != nil && !i.owns(s) {
		a.cancelPaletteRename()
	}
}
func (a *App) promptPaletteRename() {
	if a.Model.SessionPath == "" {
		return
	}
	a.cancelPaletteRename() // A new arm attempt supersedes the old intent, even if capture fails.
	var target *agent.SessionRenameTarget
	if capture := a.Options.CaptureSessionRenameTarget; capture != nil {
		v, err := capture(a.Model.RuntimeID, a.Model.SessionID, a.Model.SessionPath)
		if err != nil {
			a.Notify(err.Error(), true)
			return
		}
		target = &v
	}
	i := &paletteRenameIntent{path: a.Model.SessionPath, name: a.Model.SessionName, session: a.Model.SessionID, runtime: a.Model.RuntimeID, workspace: a.Model.Workspace, target: target}
	a.Prompt("Rename task", i.name, func(name string) {
		if a.paletteRename != i || !i.owns(a.Model) {
			return
		}
		a.paletteRename = nil
		a.promptCommit = nil // Drop the captured live child after the one-shot intent.
		a.Dispatch(Action{sessionRenameTarget: i.target, Kind: "RenameSession", Value: name, SessionID: i.session, Data: map[string]any{"path": i.path}})
	})
	a.paletteRename = i
}
