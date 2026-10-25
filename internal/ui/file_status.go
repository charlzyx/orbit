package orbitui

import "github.com/charlzyx/orbit/internal/workspace"

// Match viewer.rs::status_chip without treating debounce-pending or failed
// writes as actively saving. A retry in flight supersedes its previous error.
func (a *App) fileSaveStatus(d workspace.Document) string {
	switch {
	case !d.Content.Editable():
		return a.text("explorer.read_only", "Read-only")
	case d.Conflict:
		return a.text("explorer.changed_on_disk", "Changed on disk")
	case d.Saving:
		return a.text("explorer.saving", "Saving…")
	case d.LastError != "":
		return a.text("explorer.save_failed", "Save failed") + ": " + d.LastError
	case d.Dirty:
		return a.text("explorer.unsaved", "Unsaved")
	default:
		return a.text("explorer.saved", "Saved")
	}
}
