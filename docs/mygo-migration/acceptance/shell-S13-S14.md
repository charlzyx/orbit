# Shell, navigation, sessions and workspaces: S13–S14

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## S13 — Workspace management and appearance

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/sidebar.rs::on_workspace_remove](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2383); [crates/orbit-pi/src/app/sidebar.rs::on_workspace_open_picker](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2373); [crates/orbit-pi/src/workspace_mark.rs::WorkspaceMark](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/workspace_mark.rs#L13)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S13.behavior [UNTESTED; ROW_SPECIFIC]** Add, rename, recolor and remove a temporary workspace using row menus; inspect its icon and active highlight.
   - Expected: Add folders, open-workspace picker, copy workspace path, remove from Orbit sidebar, choose curated icon and semantic tint; active/running/archived glyph fallback.
2. **S13.edge [UNTESTED; ROW_SPECIFIC]** Cancel removal, select a missing workspace, and exercise a linked worktree with the same repository root.
   - Expected: Removing a workspace is not deleting its folder or session files. marks validate known assets, persist with workspace list, and resolve tint against active theme.
3. **S13.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Removing a workspace is not deleting its folder or session files. marks validate known assets, persist with workspace list, and resolve tint against active theme.
4. **S13.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S13.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S13.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S13.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Owner challenge UI-S08 [PARTIAL]:** These focused cases pass and1470x922CPU popup is inspected. Actual native folder dialog, workspace logo lookup, project mark customization, sidebar sticky headers and matched pixels remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S14 — Refresh fan-out

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/session.rs::on_refresh](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L2702); [crates/orbit-pi/src/app/events.rs::sync_workspace_watcher](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L1237)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S14.behavior [UNTESTED; ROW_SPECIFIC]** Trigger the registered global Refresh action after external session, branch, file and settings changes.
   - Expected: Primary+R refreshes sessions and the current workspace-dependent surfaces/catalogs rather than merely repainting.
2. **S14.edge [UNTESTED; ROW_SPECIFIC]** Delay each refresh independently; switch workspace before old results return and verify correct visible targets refresh once.
   - Expected: Active workspace must remain settled and background results must not overwrite a newer workspace. Session-store and workspace-watch invalidations also update usage/branch/Git/Review/Explorer.
3. **S14.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Active workspace must remain settled and background results must not overwrite a newer workspace. Session-store and workspace-watch invalidations also update usage/branch/Git/Review/Explorer.
4. **S14.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S14.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S14.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S14.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S06 [PARTIAL]:** Fresh cwd no longer manufactures a sidebar project; saved workspaces and actual sessions still populate groups. Source empty-state geometry is mounted. Full marks/icons/sticky active reveal/sort-state persistence remains incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

