package orbitui

import (
	"github.com/charlzyx/orbit/internal/diagnosticlog"
	"time"
)

// OnQuit runs after MyGo's event loop has stopped. Join IO settlement, never
// owner acknowledgement: queued owner callbacks may now be permanently dropped.
// Timeout is explicitly uncertain; it neither frees the lease nor claims that
// an arbitrary cancellation-ignoring runner has terminated.
func (x *Controller) waitSetupInstallRunner(limit time.Duration) bool {
	op := x.setupInstall.pending
	if op == nil {
		return true
	}
	op.cancel()
	return x.waitRuntimeWorker(op.runnerSettled, diagnosticlog.EventRuntimeInstallShutdown, limit)
}

// Background cancellation has already been requested by Close. Cancel the
// explicit operation too before either join, and share one total time budget.
func (x *Controller) waitRuntimeRunners(limit time.Duration) {
	deadline := time.Now().Add(limit)
	if op := x.setupInstall.pending; op != nil {
		op.cancel()
	}
	x.waitSetupInstallRunner(time.Until(deadline))
	x.waitRuntimeWorker(x.piUpdateDone, diagnosticlog.EventRuntimeAutoUpdateShutdown, time.Until(deadline))
}

func (x *Controller) waitRuntimeWorker(done <-chan struct{}, event diagnosticlog.Event, limit time.Duration) bool {
	if done == nil {
		return true
	}
	started := time.Now()
	x.diagnostics.Record(event, diagnosticlog.ResultStarted, 0)
	finish := func(settled bool) bool {
		result := diagnosticlog.ResultOK
		if !settled {
			result = diagnosticlog.ResultTimeout
		}
		x.diagnostics.Record(event, result, time.Since(started))
		return settled
	}
	// The shared budget may be exhausted by the other runner. A completed
	// channel still settles: do not race a zero-duration timer against it.
	select {
	case <-done:
		return finish(true)
	default:
	}
	if limit <= 0 {
		return finish(false)
	}
	timer := time.NewTimer(limit)
	defer timer.Stop()
	select {
	case <-done:
		return finish(true)
	case <-timer.C:
		return finish(false)
	}
}
