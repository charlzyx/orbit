# Usage analytics, update, privacy and release behavior: U09–U12

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## U09 — Platform update installation safety

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/updater.rs::run_install_helper](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/updater.rs#L694); [crates/orbit-pi/src/updater.rs::StagedUpdate](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/updater.rs#L113)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U09.behavior [UNTESTED; ROW_SPECIFIC]** Exercise updater preparation/rollback using local staged packages for each platform without replacing a user's app.
   - Expected: Unix helper validates archive/install path, waits normal quit, swaps application, relaunches and rolls back if ready signal fails; Windows installer launch.
2. **U09.edge [UNTESTED; ROW_SPECIFIC]** Simulate archive traversal, locked destination, helper failure and cancellation; preserve old runnable version and report unsupported hosts honestly.
   - Expected: macOS .app tarball, Linux managed prefix, Windows silent installer; platform/arch feeds differ. Download/staged temp cleanup and signature rejection are part of parity.
3. **U09.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: macOS .app tarball, Linux managed prefix, Windows silent installer; platform/arch feeds differ. Download/staged temp cleanup and signature rejection are part of parity.
4. **U09.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U09.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U09.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U09.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Disposable signature/archive/install-helper tests pass; actual installed macOS/Linux/Windows update/rollback remains unverified.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U10 — Pi update/install management

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/pi_update_ui.rs::check_pi_update_on_launch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pi_update_ui.rs#L30); [crates/orbit-pi/src/app/pi_update_ui.rs::finish_pi_update](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pi_update_ui.rs#L139)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U10.behavior [UNTESTED; ROW_SPECIFIC]** Detect Pi version, offer supported install/update and run controlled package command with fixture output.
   - Expected: Delayed newer-version check for installed pi, official managed update, npm migration prompt, reported unsupported package manager, install/version/PATH verification.
2. **U10.edge [UNTESTED; ROW_SPECIFIC]** Simulate missing npm, broken executable, failed version check and cancellation; restart only after verified success.
   - Expected: PI_BIN custom builds and PI_OFFLINE skip. Wait up to 120 s for all sessions/compaction to quiet, otherwise defer next launch. Existing children keep old version; new children use new version.
3. **U10.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: PI_BIN custom builds and PI_OFFLINE skip. Wait up to 120 s for all sessions/compaction to quiet, otherwise defer next launch. Existing children keep old version; new children use new version.
4. **U10.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U10.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U10.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U10.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U05 [PARTIAL]:** Guard and temporary retention tests pass. Whole native update/install/temp dialogs and real OS notification lifecycle remain platform acceptance gates. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U11 — Automatic RPC patches

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/bundled_extensions.rs::spawn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/bundled_extensions.rs#L97)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U11.behavior [UNTESTED; ROW_SPECIFIC]** Apply each bundled RPC patch to the exact supported Pi fixture version; reapply and inspect markers.
   - Expected: Embed/materialize custom-ui/quota/auth patch scripts, identify pi bundle anchors, apply before every spawn, fingerprint/lock fast path, Runtime opt-out/status.
2. **U11.edge [UNTESTED; ROW_SPECIFIC]** Use already-patched, unsupported-version and changed-source input; patches must be idempotent and fail closed without partial corruption.
   - Expected: Best-effort missing Node/read-only install/anchor drift → report unavailable capability, not broken core launch. Backups/revert and patch scripts remain source of truth. Config default enabled; ORBIT_NO_RPC_PATCHES override.
3. **U11.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Best-effort missing Node/read-only install/anchor drift → report unavailable capability, not broken core launch. Backups/revert and patch scripts remain source of truth. Config default enabled; ORBIT_NO_RPC_PATCHES override.
4. **U11.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U11.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U11.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U11.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U12 — Bug/feature reporting

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/bug_report.rs::file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/bug_report.rs#L116); [crates/orbit-pi/src/bug_report.rs::Diagnostics](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/bug_report.rs#L26)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U12.behavior [UNTESTED; ROW_SPECIFIC]** Open Report Bug/Feature, populate diagnostics and preview outgoing content; use a non-sending test receiver.
   - Expected: Settings Report Bug accepts context/title/description/repro or problem/solution; AI draft using active model; up to six screenshots; clear/submit and created-issue link.
2. **U12.edge [UNTESTED; ROW_SPECIFIC]** Verify redaction, omitted private paths/secrets, cancellation and launch failure; do not submit externally during acceptance runs.
   - Expected: Build diagnostics disclosed; gh issue create --attach when available, browser prefilled fallback and saved screenshots otherwise. No silent issue creation on AI draft. Busy and inline errors separate.
3. **U12.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Build diagnostics disclosed; gh issue create --attach when available, browser prefilled fallback and saved screenshots otherwise. No silent issue creation on AI draft. Busy and inline errors separate.
4. **U12.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U12.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U12.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U12.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Original report wizard context/redaction/action workflow is not yet evidenced in the native host.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

