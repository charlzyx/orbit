# Settings, providers, skills, packages and MCP: P01–P04

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## P01 — Settings hub and search

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app.rs::SettingsSection](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2679); [crates/orbit-pi/src/app/settings.rs::SETTINGS_SECTIONS](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L50); [crates/orbit-pi/src/app/settings.rs::SETTINGS_SEARCH_INDEX](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L116)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P01.behavior [UNTESTED; ROW_SPECIFIC]** Visit all 14 source Settings sections; search by section name and synonym; choose with keyboard.
   - Expected: 14 sections: General, Appearance, Shortcuts, Privacy, Runtime, Agent, Worktrees, Providers, Models, Skills, Plugins, MCP, About, Report Bug; search indexes labels/synonyms across sections, keyboard selects results.
2. **P01.edge [UNTESTED; ROW_SPECIFIC]** Test empty/no-result query, back navigation, scrolling and section change while a popup is open.
   - Expected: Section navigation, back and scroll state; empty-query/no-result behavior and popup focus must match. Existing docs that list fewer sections are stale.
3. **P01.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Section navigation, back and scroll state; empty-query/no-result behavior and popup focus must match. Existing docs that list fewer sections are stale.
4. **P01.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P01.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P01.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P01.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P01 [PARTIAL]:** 14 sections and retained native scroll pass; fixed missing FillHeight that previously prevented long pages scrolling. Exact-row/synonym search is missing; current category-only search fails source P01. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P02 — General and runtime inspection

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::runtime_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931); [crates/orbit-pi/src/app/runtime.rs::runtime_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L488); [crates/orbit-pi/src/app/runtime.rs::runtime_stop](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L518); [crates/orbit-pi/src/app/runtime.rs::runtime_restart](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L532)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P02.behavior [UNTESTED; ROW_SPECIFIC]** Inspect binary/PID/uptime/stderr and parked processes; Start, Stop and Restart the active runtime.
   - Expected: Show local session store/current workspace, process status/PID/binary/uptime/stdio, last error and recent stderr, parked process count; Start/Stop/Restart.
2. **P02.edge [UNTESTED; ROW_SPECIFIC]** Use missing binary and crashed child; Restart must resume the same session, not display Connected or erase history.
   - Expected: Runtime restart must resume current session rather than erase history. Missing executable/dead child is actionable, never Connected. State refreshes on heartbeat.
3. **P02.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Runtime restart must resume current session rather than erase history. Missing executable/dead child is actionable, never Connected. State refreshes on heartbeat.
4. **P02.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P02.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P02.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P02.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P02 [PARTIAL]:** Host lacks actual PID/uptime/stderr/process-row adapter and complete auto-retry/follow-up/title-model controls. Native dependency/install helpers exist; action/OS integration needs execution. Do not fabricate metrics. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../runtime-acceptance.md](../runtime-acceptance.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P03 — First-run dependency setup

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/view.rs::render_onboarding](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2664); [crates/orbit-pi/src/app/view.rs::render_pi_migration](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2903)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P03.behavior [UNTESTED; ROW_SPECIFIC]** Run first-use dependency detection with Pi/Node present and missing; inspect each offered setup action.
   - Expected: Host OS/architecture/store paths, Node/pi/git/gh detection/install guidance, refresh checks, official pi installer and npm-to-managed migration UI.
2. **P03.edge [UNTESTED; ROW_SPECIFIC]** Cancel installation, fail a command and retry in a disposable environment; never report ready before verification succeeds.
   - Expected: Required versus optional dependencies differ. Managed/npm/other/custom PI_BIN installs are distinguished; unsupported package-manager installs are not silently replaced. Verify install outcome and on-PATH identity.
3. **P03.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Required versus optional dependencies differ. Managed/npm/other/custom PI_BIN installs are distinguished; unsupported package-manager installs are not silently replaced. Verify install outcome and on-PATH identity.
4. **P03.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P03.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P03.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P03.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P02 [PARTIAL]:** Host lacks actual PID/uptime/stderr/process-row adapter and complete auto-retry/follow-up/title-model controls. Native dependency/install helpers exist; action/OS integration needs execution. Do not fabricate metrics. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P04 — Agent defaults and titles

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::agent_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5196); [crates/orbit-pi/src/app/runtime.rs::apply_default_model](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L217)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P04.behavior [UNTESTED; ROW_SPECIFIC]** Edit agent defaults, prompt/title-generation settings and restore source defaults; start a new session.
   - Expected: Default model/provider/thinking for NEW sessions, active model selection, Enter delivery policy, one/all follow-ups, auto-compaction/retry, automatic session titles and dedicated title model.
2. **P04.edge [UNTESTED; ROW_SPECIFIC]** Load malformed stored values, fail persistence and override settings in an existing session; compare scope boundaries.
   - Expected: session-defaults.json affects spawn and new session but not resumed model. auto-title.json supports active-model fallback; title generation after first settled exchange only.
3. **P04.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: session-defaults.json affects spawn and new session but not resumed model. auto-title.json supports active-model fallback; title generation after first settled exchange only.
4. **P04.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P04.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P04.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P04.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P02 [PARTIAL]:** Host lacks actual PID/uptime/stderr/process-row adapter and complete auto-retry/follow-up/title-model controls. Native dependency/install helpers exist; action/OS integration needs execution. Do not fabricate metrics. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../runtime-acceptance.md](../runtime-acceptance.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

