package orbitui

import "sync/atomic"

// A positive acknowledgement of synchronous host admission, NOT confirmation
// that a remote new_session RPC or the new process has completed successfully.
// The callback is owner-thread synchronous; an absent/no-op controller cannot
// authorize discarding local draft state.
type newTaskAdmission struct{ accepted atomic.Bool }

func (action Action) acceptNewTask() {
	if action.Kind == "NewSession" && action.newTaskAdmission != nil {
		action.newTaskAdmission.accepted.Store(true)
	}
}
func (a *App) admitNewTask() bool {
	before := a.Model
	text := a.Draft
	ids := make([]string, len(a.attachments))
	for i, attachment := range a.attachments {
		ids[i] = attachment.ID
	}
	admission := &newTaskAdmission{}
	a.Dispatch(Action{Kind: "NewSession", Value: before.Workspace, newTaskAdmission: admission})
	if !admission.accepted.Load() {
		return false
	}
	// A synchronous custom controller or re-entrant callback can already present
	// another task/draft. Never clear that newer state after accepting an old intent.
	now := a.Model
	if now.RuntimeID != before.RuntimeID || now.SessionID != before.SessionID || now.SessionPath != before.SessionPath || now.Workspace != before.Workspace || a.Draft != text || len(a.attachments) != len(ids) {
		return false
	}
	for i, id := range ids {
		if a.attachments[i].ID != id {
			return false
		}
	}
	return true
}
