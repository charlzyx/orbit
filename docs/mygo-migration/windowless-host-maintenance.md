# Windowless non-rendering host maintenance

Reference: Rust `299a32b`, `main_window.rs:136–160` app-owned heartbeat,
`app/events.rs:18–120` full non-window-bound application tick.
Implementation: unreleased continuation of Go `5e4e1425` plus the parked-MCP
fingerprint correction. This is not a native macOS Dock-reopen acceptance claim.

## Confirmed gap

Go's windowless timer previously called only `Agent.Tick`. Pi events, parked
runtime cleanup and native notifications continued, but no frame existed to
consume `App.updates` or run controller `sync`. Consequently completed
retention/update/other async callbacks could remain queued; MCP/resources/Git/
worktree completions, deferred config reload acknowledgment, host snapshot and
usage-store synchronization could stall until reopening a window. Toast expiry
also depended on rendering.

The production `TestWindowlessHostDrainsAsyncCompletion` reproduced this before
the correction: the completed callback remained unconsumed. It was not a test
of rendered widgets or a fabricated new maintenance responsibility.

## Correction and owner-thread boundary

The existing adaptive timer (90 ms idle, 16 ms active, delay after work) now uses
MyGo's actual `RunOnMain` synchronous dispatcher for owner-thread eligibility
checking and non-rendering host maintenance. The middle `Agent.Tick` stays on
the timer worker: the manager serializes runtime state, and native notification
I/O may itself wait for the UI loop or invoke the unpackaged macOS helper. No
controller/widget maintenance executes on that worker.

After runtime work, the owner callback checks stop, native-window existence and
window-opening handoff **again**. If reopening won, the visible frame owns host
maintenance. No lock is held while waiting for the dispatcher or running a
completion. Close continues not to join a potentially main-thread-waiting
heartbeat, so native quit cannot deadlock on its own queued callback.

The shared non-rendering host step drains completed jobs, snapshots runtime
state, advances existing service completion queues and deferred config apply,
and expires toasts using a supplied timestamp. It does not invoke `View`, open
native dialogs, render a backdrop or perform layout. Pending picker results
retain their own captured-parent/intent cancellation checks.

## Completion mailbox tradeoff

Controller callbacks now enter a mutex-protected mailbox rather than blocking
on the old 64-slot UI channel. Synchronous notification/completion bursts cannot
self-deadlock an owner-thread runtime drain. At most 256 mailbox callbacks are
removed per tick, then invoked **outside** the mutex. A callback that enqueues
another callback leaves it for a later batch. Existing `App.Async` worker
completions retain their buffered channel; its drain takes only the queue length
observed at batch start. Neither route endlessly drains newly arriving work.

Controller mailbox admission is intentionally unbounded: authoritative accepted
completions are not silently dropped, but a sustained producer rate above owner
throughput can grow memory. This is a conscious tradeoff, not a claim that all
backpressure risks vanished or that Rust uses the same queue implementation.
Close atomically disables admission and discards remaining UI completions; jobs
already executing retain their existing cancellation behavior. `App.Async` still
has its independent buffered-channel/`Window.Update` paths; this does not unify
all background queues or eliminate their existing shutdown-time producer waits.

## Scope and evidence

Synthetic-process tests cover actual MCP replacement bootstrap acknowledgment
without a rendered frame. Headless owner-dispatch tests cover the production
timer boundary, reopening/quit before dispatch, 700 accepted completions,
256-item fairness, recursive enqueue, callback-triggered Close and toast expiry.
Existing virtual-clock heartbeat tests continue checking adaptive cadence, and
existing real JSONL child tests cover busy/idle/retry/compaction/parked activity.

No new retention sweep is scheduled. Existing automatic retention still starts
once at readiness under the user's saved policy; explicit deletion confirmation
is unchanged. This change only allows its existing completion to apply without
a window. No live user sessions or retention data were used by these tests.

Native rendering/layout/backdrop behavior is not moved to background work.
MCP external-edit polling still uses the existing recent-render visibility
heuristic; retaining an open MCP page while windowless is a separate remaining
source difference, despite completion and explicit apply now progressing.
Foreground notification delivery's pre-existing threading semantics are not
redesigned here. Real OS window-close/reopen and notification behavior require
same-source hosted native verification; headless callback tests do not replace it.
