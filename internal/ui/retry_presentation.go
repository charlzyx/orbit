package orbitui

// Presentation cancellation is scoped to one observed retry event, never to a
// provider/attempt string. Identical later retry events must show again.
type retryCancellation struct {
	active               bool
	runtimeID, sessionID string
	revision             uint64
}

func (a *App) cancelRetryPresentation() {
	a.retryCancellation = retryCancellation{active: true, runtimeID: a.Model.RuntimeID, sessionID: a.Model.SessionID, revision: a.Model.RetryDetail.Revision}
}
func (a *App) retryPresentationCancelled() bool {
	v := a.retryCancellation
	return v.active && v.runtimeID == a.Model.RuntimeID && v.sessionID == a.Model.SessionID && v.revision == a.Model.RetryDetail.Revision
}
func (a *App) retryPresentationActive() bool {
	return a.Model.Retrying && !a.retryPresentationCancelled()
}
