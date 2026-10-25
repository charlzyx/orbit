package orbitui

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Presentation cancellation is scoped to one observed retry event, never to a
// provider/attempt string. Identical later retry events must show again.
type retryCancellation struct {
	active               bool
	runtimeID, sessionID string
	revision             uint64
}

func (a *App) cancelRetryPresentation() {
	a.retryStatusAt = a.retryFeedbackTime()
	a.retryCancellation = retryCancellation{active: true, runtimeID: a.Model.RuntimeID, sessionID: a.Model.SessionID, revision: a.Model.RetryDetail.Revision}
}
func (a *App) retryPresentationCancelled() bool {
	v := a.retryCancellation
	return v.active && v.runtimeID == a.Model.RuntimeID && v.sessionID == a.Model.SessionID && v.revision == a.Model.RetryDetail.Revision
}
func (a *App) retryPresentationActive() bool {
	return a.Model.Retrying && !a.retryPresentationCancelled()
}

// This is only the local Retry aborted feedback, not a global status owner.
// Rust app.rs STATUS_MESSAGE_TTL is six seconds; view.rs3171 puts it before
// send hints and the context meter, muted, truncated, and at most 320px wide.
func (a *App) retryCancellationStatus(c *ui.Context) bool {
	if a.retryStatusAt.IsZero() {
		return false
	}
	if a.Model.ExtensionStatusAt.After(a.retryStatusAt) {
		a.retryStatusAt = time.Time{}
		return false
	}
	elapsed := a.retryFeedbackTime().Sub(a.retryStatusAt)
	if elapsed >= 6*time.Second {
		a.retryStatusAt = time.Time{}
		return false
	}
	c.After(6*time.Second - elapsed)
	ui.Box(c.Key("retry-cancellation-status")).Label("Retry cancellation status").MaxWidth(320).Children(func() {
		ui.Text(c, a.Tr("settings.retry_aborted")).MaxWidth(320).SingleLine().TextColor(a.Theme.TextMuted).FontSize(a.Theme.Px(12))
	})
	return true
}

// Pi's ReceivedAt is time.Now in the local reader. Use that same monotonic
// clock at the user action, not c.Now's potentially earlier display-frame
// timestamp. The private override keeps this one feedback timer deterministic
// in mounted tests; no global clock or runtime timing is changed.
func (a *App) retryFeedbackTime() time.Time {
	if a.retryFeedbackNow != nil {
		return a.retryFeedbackNow()
	}
	return time.Now()
}
