# Shell, navigation, sessions and workspaces: S09–S12

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## S09 — Session navigation

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/sidebar.rs::open_session_slot](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2509); [crates/orbit-pi/src/app/sidebar.rs::cycle_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2521); [crates/orbit-pi/src/app/sidebar.rs::on_sidebar_confirm](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2618)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S09.behavior [UNTESTED; ROW_SPECIFIC]** Navigate next/previous session and session history using source commands; return to a running session.
   - Expected: Primary+1…9 targets visible sidebar order; next/previous session shortcuts; sidebar focus, up/down/home/end/Enter/Escape; active row auto-reveal through collapsed groups.
2. **S09.edge [UNTESTED; ROW_SPECIFIC]** Exercise first/last boundaries, archived sessions and a deleted history target without opening the wrong workspace.
   - Expected: Navigation observes current grouping, sorting, archived filtering and visible session rows. It must not point at stale hidden items.
3. **S09.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Navigation observes current grouping, sorting, archived filtering and visible session rows. It must not point at stale hidden items.
4. **S09.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S09.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S09.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S09.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S10 — Session metadata and title controls

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/session.rs::clone_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L613); [crates/orbit-pi/src/app/settings.rs::rename_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5836); [crates/orbit-pi/src/app/events.rs::tick](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L18); [crates/orbit-pi/src/app/sidebar.rs::on_menu_clone_session](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2188)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S10.behavior [UNTESTED; ROW_SPECIFIC]** Rename a session, allow automatic title generation, edit metadata and inspect title in sidebar and header.
   - Expected: Rename, generate title, copy ID/path, reveal session file, inspect session details, clone entire conversation. Active header and sidebar reconcile session_info_changed.
2. **S10.edge [UNTESTED; ROW_SPECIFIC]** Return an empty or failed generated title; manually rename while generation is pending; reopen the session.
   - Expected: Clone is shipped, rewind/fork UI is not. Clone preserves session data with new identity; invalid input/partial data is guarded. Title generation spinner and rename success feedback are distinct.
3. **S10.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Clone is shipped, rewind/fork UI is not. Clone preserves session data with new identity; invalid input/partial data is guarded. Title generation spinner and rename success feedback are distinct.
4. **S10.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S10.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S10.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S10.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S11 — Pin, archive and delete sessions

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/sidebar.rs::on_menu_delete_confirm](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2231); [crates/orbit-pi/src/app/sidebar.rs::on_menu_toggle_archive](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2214)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S11.behavior [UNTESTED; ROW_SPECIFIC]** Pin and unpin; archive and restore; request deletion and first cancel, then confirm on a disposable session.
   - Expected: Session row menu supports pin/unpin, archive/unarchive, copy path, reveal and explicit delete confirmation; archived filter changes default visible list.
2. **S11.edge [UNTESTED; ROW_SPECIFIC]** Filter archived sessions, delete a currently selected fixture session and verify fallback selection and on-disk recovery contract.
   - Expected: Pins and archives are path-keyed Orbit stores. Session deletion removes JSONL and related pin/archive/workflow state; running-session guards matter. Archive does not rewrite pi history.
3. **S11.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Pins and archives are path-keyed Orbit stores. Session deletion removes JSONL and related pin/archive/workflow state; running-session guards matter. Archive does not rewrite pi history.
4. **S11.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S11.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S11.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S11.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S12 — Workspace sidebar organization

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/sidebar.rs::build_sidebar_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L176); [crates/orbit-pi/src/app/sidebar.rs::build_flat_sidebar_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L370); [crates/orbit-pi/src/app.rs::WorkspaceSort](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2248); [crates/orbit-pi/src/app.rs::SidebarGroupBy](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2339)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S12.behavior [UNTESTED; ROW_SPECIFIC]** Sort/group workspace sidebar, drag manual order if offered, collapse a group and toggle archived visibility.
   - Expected: Workspace groups or one flat list, collapsed groups, show-more session rows, sticky workspace header, workspace selection/markers; row and context-menu actions.
2. **S12.edge [UNTESTED; ROW_SPECIFIC]** Use duplicate display names and sessions with missing workspace paths; relaunch from the same fixture preference store.
   - Expected: workspaces.json records explicit project list, added timestamps, grouping, archive filter, marks and sort. Sort supports manual, updated, added, alphabetical ascending/descending, session count.
3. **S12.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: workspaces.json records explicit project list, added timestamps, grouping, archive filter, marks and sort. Sort supports manual, updated, added, alphabetical ascending/descending, session count.
4. **S12.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S12.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S12.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S12.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

