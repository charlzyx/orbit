# Palette Delete: capture and revoke exact confirmation intent

Base: clean wave5 3fddcce plus independently approved test-only receipt-clock
27733b14 additive (local baseline commit recorded in manifest). No other pending
palette corrections are included: Pin, idle gating, search and refresh remain
separate. No production agent/file-deletion code is changed.

## Source requirement and original failure

Pinned Rust app/sidebar.rs:2691–2725 captures active session path/title into the
same SessionMenu used by row deletion; explicit confirmation remains necessary.
Go actions.go:278–279 instead read current Model.SessionPath inside Continue's
closure. Actual palette mount first→second snapshot then Continue dispatched
second in `command-audit-red-1936.log`. The test only captured Action; it did NOT
execute deletion. Backend active/live protection remains a separate gate and
this evidence does not establish successful unintended deletion.

## Minimal implementation

Capture path/title/session/runtime/workspace into a single paletteDeleteIntent
before showing named path+title confirmation. Reject empty saved path. Continue
checks exact intent plus current owner and emits only captured path/session ID;
it consumes intent before dispatch (one-shot/reentrancy safe).

SetSnapshot invalidates on any task-owner dimension change. Actual workspace
services invalidate before directly mutating Model.Workspace, including A→B→A
between frames; same-owner refresh remains valid. Title-only updates do not
silently change the already-displayed victim. Generic newer Confirm replaces
intent. Cancel/Escape/dismissal revoke it; non-delete confirmation behavior is
otherwise unchanged.

A closure token alone is insufficient: native Continue identity must also be
scoped to that intent. The deterministic held-press test first failed after the
closure guard: Press on first Continue→replace owner and arm second before next
frame→Release dispatched second (`palette-delete-held-negative-1942.log`). Only
palette-delete modal Context.Key now includes the token; generic confirmations
retain their prior key. A fresh click on the replacement still works.

## Verification scope

Tests capture emitted Actions or use disposable controller fixtures with an
offline missing Pi binary. No test executes DeleteSession on real/user files.
Pointer/Enter palette admission, exact displayed path/title, one-shot Continue,
pathless refusal, title-only refresh, session/path/runtime/workspace ABA,
Cancel/Escape/new-confirmation replacement, actual service workspace roundtrip,
held native Press/Release, and fresh replacement confirmation are covered.
Existing session row menu/confirmation behavior is included in focused gates.

Agent active/live/busy/file ownership refusal and explicit confirmation are not
weakened. This is UI intent correctness, not an atomic filesystem transaction or
proof of every session lifecycle. Physical macOS menu/AX capture is unverified.
