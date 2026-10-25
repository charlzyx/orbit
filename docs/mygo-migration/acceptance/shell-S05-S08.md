# Shell, navigation, sessions and workspaces: S05–S08

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## S05 — Keyboard surface traversal

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/main.rs::bind_keys](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L357); [crates/orbit-pi/src/app/view.rs::on_focus_next](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3775); [crates/orbit-pi/src/app/view.rs::focus_traversal_blocked](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3804); [crates/orbit-pi/src/app/composer_ops.rs::close_add_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L537)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S05.behavior [UNTESTED; ROW_SPECIFIC]** Traverse all focusable shell surfaces with Tab and Shift+Tab; dismiss the active popover with Escape.
   - Expected: Tab/Shift+Tab traverse focusable surfaces; terminal and autocomplete consume Tab for their own meaning; modal/select/picker contexts own Enter/Escape/arrows.
2. **S05.edge [UNTESTED; ROW_SPECIFIC]** Open nested overlays, return to the composer and verify text selection plus input focus; repeat after switching workspace.
   - Expected: Do not let Escape abort a run while a dialog/custom UI/menu should close; closing a popup restores focus and suppresses click-through reopen.
3. **S05.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Do not let Escape abort a run while a dialog/custom UI/menu should close; closing a popup restores focus and suppresses click-through reopen.
4. **S05.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S05.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S05.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S05.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S03 [PARTIAL]:** These exact native CPU flows pass, not every command/context or OS accelerator. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S06 — Read and watch shared pi session store

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/sessions.rs::SessionInfo](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sessions.rs#L29); [crates/orbit-pi/src/sessions.rs::SessionWatcher](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sessions.rs#L229); [crates/orbit-pi/src/app/events.rs::tick](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L18)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S06.behavior [UNTESTED; ROW_SPECIFIC]** Populate a fixture pi session directory; open Orbit; append, rename, truncate and remove session files externally.
   - Expected: Discover pi JSONL sessions, names, workspace, message preview and last actual message activity; updates from CLI/external writers refresh sidebar.
2. **S06.edge [UNTESTED; ROW_SPECIFIC]** Use malformed JSONL, a temporarily unreadable file and a partially written final line; verify watcher recovery without duplicate sessions.
   - Expected: Uses pi default session store, not a new incompatible database. Debounced background scan and cached metadata. Opening/bookkeeping writes do not count as message activity.
3. **S06.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Uses pi default session store, not a new incompatible database. Debounced background scan and cached metadata. Opening/bookkeeping writes do not count as message activity.
4. **S06.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S06.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S06.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S06.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S05 [PARTIAL]:** Draft/inactive file generation tests pass. Whole native controller repository-switch interaction with fake Pi and all domain mutations still needs a dedicated end-to-end fixture. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S07 — New task and workspace selection

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/session.rs::on_new_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L581); [crates/orbit-pi/src/app/session.rs::can_reuse_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L61); [crates/orbit-pi/src/app/pickers.rs::start_task_in_folder](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pickers.rs#L118)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S07.behavior [UNTESTED; ROW_SPECIFIC]** Create a task in each of two temporary workspaces and in no-workspace mode; choose from recent and directory picker paths.
   - Expected: Primary+N, sidebar New Task, recent folder picker and native Browse; selected folder sets session cwd and Explorer/Git/Review context; custom workspace logo or fallback glyph.
2. **S07.edge [UNTESTED; ROW_SPECIFIC]** Cancel the picker, select a missing directory and repeat New Task while the current task is running.
   - Expected: Reuse only alive idle pi process; park a running session before new_session; a dead process is replaced. New-task transient panels close; draft session placeholder visible before first prompt.
3. **S07.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Reuse only alive idle pi process; park a running session before new_session; a dead process is replaced. New-task transient panels close; draft session placeholder visible before first prompt.
4. **S07.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S07.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S07.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S07.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S05 [PARTIAL]:** Draft/inactive file generation tests pass. Whole native controller repository-switch interaction with fake Pi and all domain mutations still needs a dedicated end-to-end fixture. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../background-acceptance.md](../background-acceptance.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S08 — Switch/resume and parallel sessions

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app.rs::ParkedSession](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L150); [crates/orbit-pi/src/app/session.rs::park_active_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L299); [crates/orbit-pi/src/app/runtime.rs::reap_idle_parked](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L569); [crates/orbit-pi/src/app/events.rs::drain_parked_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L541)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S08.behavior [UNTESTED; ROW_SPECIFIC]** Start two fixture Pi sessions; switch while one streams; return and resume a previously closed session.
   - Expected: Open session by sidebar/palette/notification; preserve live transcript, pending queue, widgets and busy status when parked. Sidebar shows background work.
2. **S08.edge [UNTESTED; ROW_SPECIFIC]** Crash one child and deliver its late event after switching; ensure the other session and its composer state remain isolated.
   - Expected: At most six warm background sessions, idle TTL 300 s; evict/reap idle only, never active runs. Pending parks before a session path is known are drained/rekeyed. Stale MCP fingerprint forces fresh process on reopen.
3. **S08.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: At most six warm background sessions, idle TTL 300 s; evict/reap idle only, never active runs. Pending parks before a session path is known are drained/rekeyed. Stale MCP fingerprint forces fresh process on reopen.
4. **S08.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S08.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S08.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S08.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S05 [PARTIAL]:** Draft/inactive file generation tests pass. Whole native controller repository-switch interaction with fake Pi and all domain mutations still needs a dedicated end-to-end fixture. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

