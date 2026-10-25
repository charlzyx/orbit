# Usage analytics, update, privacy and release behavior: U13–U14

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## U13 — Privacy and network boundaries

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/bug_report.rs::issue_body](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/bug_report.rs#L262); [crates/orbit-pi/src/http.rs::avatar_client](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/http.rs#L22)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U13.behavior [UNTESTED; ROW_SPECIFIC]** Record network/file activity for a local idle session, provider call, avatar, update check and report preview.
   - Expected: No telemetry/installation-ID feature in source; keep local sessions/config and actual provider/auth/quota/update/GitHub behavior.
2. **U13.edge [UNTESTED; ROW_SPECIFIC]** Disable applicable settings and compare observed endpoints to source; never claim offline/privacy behavior from UI labels alone.
   - Expected: Network is not zero: provider calls, quota/auth, app/pi/plugin update checks, GitHub metadata/avatars/reporting, and requested remote content. Preserve actual timing/PI_OFFLINE guards instead of claiming no background traffic.
3. **U13.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Network is not zero: provider calls, quota/auth, app/pi/plugin update checks, GitHub metadata/avatars/reporting, and requested remote content. Preserve actual timing/PI_OFFLINE guards instead of claiming no background traffic.
4. **U13.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U13.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U13.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U13.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U14 — Packaging and platform identity

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/assets.rs::collect_fonts](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/assets.rs#L103); [crates/orbit-pi/src/main.rs::main](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L630)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U14.behavior [UNTESTED; ROW_SPECIFIC]** Build each advertised OS/architecture target, inspect app identity/assets and start in disposable native environments.
   - Expected: Bundled assets/fonts/icon, macOS debug Alpha identity, app menu/window controls, Windows manifest/installer and Linux bundle/deb scripts; release target identity and native integration.
2. **U14.edge [UNTESTED; ROW_SPECIFIC]** Verify missing system dependency and unsupported platform diagnostics; cross-build success alone is not native runtime acceptance.
   - Expected: Platform-limited functions must show unavailability, not fake success. Fork identifiers, update key/feed, report target and branding must be explicit migration decisions.
3. **U14.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Platform-limited functions must show unavailability, not fake success. Fork identifiers, update key/feed, report target and branding must be explicit migration decisions.
4. **U14.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U14.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U14.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U14.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native entry point/assets exist; target installers, identity/signing and full packaged-platform acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

