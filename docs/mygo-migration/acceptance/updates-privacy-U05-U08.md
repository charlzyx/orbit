# Usage analytics, update, privacy and release behavior: U05–U08

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## U05 — Notifications and toasts

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/events.rs::post_notification](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L1193); [crates/orbit-pi/src/app/session.rs::activate_session_from_notification](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L1215)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U05.behavior [UNTESTED; ROW_SPECIFIC]** Trigger success/warning/error toasts and desktop notifications for source events; dismiss and reopen affected surface.
   - Expected: Completion/failure/input-needed background notifications, three independent desktop/sound/toast preferences, click bundled banner opens session; deduped capped toast stack with dismiss/TTL/fade.
2. **U05.edge [UNTESTED; ROW_SPECIFIC]** Disable notification preference, fail permission request and repeat same event; no unsolicited duplicate notifications.
   - Expected: Frontmost window suppresses OS banner/sound in favor of toast. macOS bundled UNUserNotificationCenter; unbundled osascript attributed to Script Editor. Non-macOS native banner/sound stubs remain limitations.
3. **U05.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Frontmost window suppresses OS banner/sound in favor of toast. macOS bundled UNUserNotificationCenter; unbundled osascript attributed to Script Editor. Non-macOS native banner/sound stubs remain limitations.
4. **U05.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U05.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U05.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U05.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U05 [PARTIAL]:** Guard and temporary retention tests pass. Whole native update/install/temp dialogs and real OS notification lifecycle remain platform acceptance gates. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U06 — Session retention and privacy controls

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::privacy_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5618)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U06.behavior [UNTESTED; ROW_SPECIFIC]** Change retention/privacy settings and preview cleanup on a copied session store; cancel then confirm eligible deletions.
   - Expected: Privacy explains local data, opt-in launch archive/delete by age presets and manual cleanup; eligible sessions based on message activity.
2. **U06.edge [UNTESTED; ROW_SPECIFIC]** Protect pinned/active/recent sessions, malformed timestamps and outside-root paths; failed cleanup must remain actionable.
   - Expected: Default off/archive/30 d; presets 7,14,30,60,90,180,365. Live sessions skipped. Permanent deletion removes JSONL and pin/archive state; archive is reversible. Outcome reports touched/skipped/failures.
3. **U06.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Default off/archive/30 d; presets 7,14,30,60,90,180,365. Live sessions skipped. Permanent deletion removes JSONL and pin/archive state; archive is reversible. Outcome reports touched/skipped/failures.
4. **U06.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U06.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U06.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U06.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U05 [PARTIAL]:** Guard and temporary retention tests pass. Whole native update/install/temp dialogs and real OS notification lifecycle remain platform acceptance gates. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U07 — Temporary file cleanup

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::privacy_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5618)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U07.behavior [UNTESTED; ROW_SPECIFIC]** Create expired and current temporary artifacts under fixture roots; run cleanup and inspect retained files.
   - Expected: List Orbit temp files and legacy orbit-* separately from pi AI scratch, sizes/age, refresh/reveal/cleanup.
2. **U07.edge [UNTESTED; ROW_SPECIFIC]** Use symlink escapes, unrelated files and a locked target; cleanup must stay inside its allowed paths and not remove current artifacts.
   - Expected: Protect files belonging to live Orbit process; AI scratch touched within 1 h retained; scan depth bound; do not delete arbitrary temp directories. Failures and kept items reported.
3. **U07.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Protect files belonging to live Orbit process; AI scratch touched within 1 h retained; scan depth bound; do not delete arbitrary temp directories. Failures and kept items reported.
4. **U07.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U07.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U07.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U07.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U05 [PARTIAL]:** Guard and temporary retention tests pass. Whole native update/install/temp dialogs and real OS notification lifecycle remain platform acceptance gates. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U08 — Signed Orbit app updater

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/updater.rs::download_to](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/updater.rs#L1432); [crates/orbit-pi/src/updater.rs::verify_artifact](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/updater.rs#L1360); [crates/orbit-pi/src/updater.rs::check_for_updates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/updater.rs#L260)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U08.behavior [UNTESTED; ROW_SPECIFIC]** Check signed update manifests with newer, equal and older versions; download and verify a fixture payload.
   - Expected: Automatic check preference, manual check/Primary+Shift+U, version/history/search modal, available pill, Update now, download/stage/quit/install and failure messages.
2. **U08.edge [UNTESTED; ROW_SPECIFIC]** Reject bad signatures/hash/architecture, offline/rate-limit errors and partial download; never install unverified bytes.
   - Expected: Ed25519 signature verification before staging; bounded feed/artifact/unpack limits. Debug dormant unless explicitly forced; missing/invalid key disables. Fork must not accidentally install upstream Orbit artifacts.
3. **U08.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Ed25519 signature verification before staging; bounded feed/artifact/unpack limits. Debug dormant unless explicitly forced; missing/invalid key disables. Fork must not accidentally install upstream Orbit artifacts.
4. **U08.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U08.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U08.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U08.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Signed updater/staging logic and disposable tests exist; a Go-specific signed feed/key/release artifact pipeline is not configured. Must not use upstream Rust updates.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

