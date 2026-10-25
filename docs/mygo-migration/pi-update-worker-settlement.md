# Automatic Pi worker actual completion on Quit

The existing explicit install runner had a settlement handle, but the automatic
launch-update goroutine did not. Canceling Controller background context was not
proof that the automatic worker's subprocess Wait/reap had returned before host
exit. A fixture of the actual tracked checker demonstrated Close returning while
a cancellation-aware-but-not-yet-returned worker was still active.

## Small lifecycle extension

`native.CheckPiUpdateOnLaunch` now returns an actual-completion channel. Offline
and custom PI_BIN skips return a closed channel. All other worker early returns,
normal returns and cancellation exits close it via a defer around the unchanged
worker body. A private worker-start helper permits deterministic fixture testing
without contacting a release endpoint or running an installer. Existing callers
may ignore the returned channel.

Controller retains that handle once per launch. Close first cancels background
work, then cancels any explicit operation, then waits outside manager/global/UI
locks on actual completion, never owner notification acknowledgment. Explicit and
automatic runners share a single three-second total deadline; this is not an
additional three seconds for each operation. The existing explicit-runner waiter
and automatic join share one small typed-diagnostic helper. Nil is a no-op;
already-completed channels are checked before an exhausted/zero budget so they
cannot randomly lose to a ready timer.

The appended allowlisted `runtime_auto_update_shutdown` diagnostic records only
started/ok/timeout and duration, no arbitrary errors, output, paths or credentials.
Timeout explicitly means uncertain cleanup: it neither closes completion nor
releases a still-running maintenance lease. An arbitrary injected worker that
ignores cancellation can outlive the bounded join; this is diagnosed, not called
cleanup success. Native commands keep their existing context process-group
cancellation and WaitDelay behavior. Update plans, release lookup, idle delay,
network target and business timeouts are unchanged.

## Tests and limits

* Actual Close remains blocked after cancellation until injected worker return.
* Already queued owner receipts do not block Quit or reappear after host close.
* Both uncompleted handles consume one deadline, timeout is typed/uncertain,
  and nil/already-settled handles remain successful even with no budget left.
* A second startup cannot overwrite the tracked completion handle.
* Native custom/offline/canceled-before-IO paths settle without a release query;
  private worker fixtures verify normal return versus still-held execution.
* A fixed harmless `/bin/sh` + `sleep` fixture uses native cancellation, then
  verifies its own descendant is terminated/zombie and the maintenance writer
  is released only after command return. This Linux procfs assertion is not
  physical macOS installer acceptance.

No real Pi update, installer script, auth action, network release request or user
configuration migration was executed by these fixtures. Focused native/UI race,
type build, vet and Darwin arm64 compile gate this isolated delta; the integrator
owns the final combined full suite.
