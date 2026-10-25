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
	started := time.Now()
	x.diagnostics.Record(diagnosticlog.EventRuntimeInstallShutdown, diagnosticlog.ResultStarted, 0)
	op.cancel()
	timer := time.NewTimer(limit)
	defer timer.Stop()
	result := diagnosticlog.ResultOK
	settled := true
	select {
	case <-op.runnerSettled:
	case <-timer.C:
		result = diagnosticlog.ResultTimeout
		settled = false
	}
	x.diagnostics.Record(diagnosticlog.EventRuntimeInstallShutdown, result, time.Since(started))
	return settled
}
