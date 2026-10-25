package orbitui

import "github.com/egoist/mygo/ui"

// Source sidebar.rs session_menu_popup: stateful labels and exact row targets.
func (a *App) sessionRowMenu(m *ui.Menu, s SessionRow) {
	pin, archive := "sidebar.pin_session", "sidebar.archive_session"
	if s.Pinned {
		pin = "sidebar.unpin_session"
	}
	if s.Archived {
		archive = "sidebar.unarchive_session"
	}
	for _, cmd := range []struct{ id, label string }{
		{"PinSession", a.Tr(pin)}, {"ArchiveSession", a.Tr(archive)},
		{"CopySessionPath", a.Tr("sidebar.copy_path")}, {"RevealSession", a.Tr("sidebar.reveal_in_finder")},
		{"CloneSession", a.Tr("sidebar.clone_session")},
		// Retain existing Go convenience actions without confusing ID with path.
		{"RenameSession", a.Tr("go.session.rename")}, {"CopySessionId", a.Tr("usage.ctx_copy_session_id")},
	} {
		if m.Item(cmd.label).Chosen() {
			a.commandSession(cmd.id, s)
		}
	}
	if !a.sessionHasRuntime(s.Path) {
		m.Separator()
		if m.Item(a.Tr("sidebar.delete_session_menu")).Chosen() {
			a.commandSession("DeleteSession", s)
		}
	}
}
func (a *App) sessionHasRuntime(path string) bool {
	if path == "" {
		return true
	}
	if p := a.Runtime.Active; p != nil && p.SessionPath == path {
		return true
	}
	for _, p := range a.Runtime.Parked {
		if p.SessionPath == path && (p.Alive || p.Loading || p.Busy || p.State == "stopping") {
			return true
		}
	}
	return false
}
