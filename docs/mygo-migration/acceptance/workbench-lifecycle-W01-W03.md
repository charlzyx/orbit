# Workbench lifecycle and shared state: W01–W03

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W01 — One main-area feature surface at a time: Files opens while closing Settings/Git/Usage/Worktrees; return-to-chat preserves the file entity. Docked Review and Explorer are mutually exclusive. Full-width Review keeps/restores a dock width and opens the sessions sidebar.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/session.rs::open_file_in_viewer](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L2448); [crates/orbit-pi/src/app.rs::sync_terminal_active](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2098); [crates/orbit-pi/src/sidepane.rs::close](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L524); [crates/orbit-pi/src/app/view.rs::render](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)

**Fixture:** Two temporary Git workspaces, fixture Pi stream, open files/tabs and independent watcher/subprocess completion controls.

1. **W01.behavior [UNTESTED; ROW_SPECIFIC]** Open Files, Settings, Git, Usage and Worktrees in turn; toggle docked Review/Explorer and full-width Review.
   - Expected: One main-area feature surface at a time: Files opens while closing Settings/Git/Usage/Worktrees; return-to-chat preserves the file entity. Docked Review and Explorer are mutually exclusive. Full-width Review keeps/restores a dock width and opens the sessions sidebar.
2. **W01.edge [UNTESTED; ROW_SPECIFIC]** Hide and return to Files with an unsaved buffer; verify only source-exclusive surfaces close and saved widths/focus restore.
   - Expected: Page changes update terminal visibility rather than killing shells. Full-width close restores dock state. This is cross-surface behavior, not just styling.
3. **W01.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail a watcher or refresh callback and switch visible workspace/surface before completion.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Page changes update terminal visibility rather than killing shells. Full-width close restores dock state. This is cross-surface behavior, not just styling.
4. **W01.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W01.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W01.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W01.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W02 — Workspace edits from agent, user, editor or Git refresh docked Review, visible Git status, Explorer, active Files tab, branch status and visible Worktrees list.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/watch.rs::WorkspaceWatcher](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/watch.rs#L77); [crates/orbit-pi/src/watch.rs::git_metadata_dirs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/watch.rs#L138); [crates/orbit-pi/src/watch.rs::workspace_relevant](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/watch.rs#L206); [crates/orbit-pi/src/watch.rs::git_metadata_relevant](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/watch.rs#L243)

**Fixture:** Two temporary Git workspaces, fixture Pi stream, open files/tabs and independent watcher/subprocess completion controls.

1. **W02.behavior [UNTESTED; ROW_SPECIFIC]** Modify files via agent/editor/external process and stage/checkout via Git while each dependent pane is visible.
   - Expected: Workspace edits from agent, user, editor or Git refresh docked Review, visible Git status, Explorer, active Files tab, branch status and visible Worktrees list.
2. **W02.edge [UNTESTED; ROW_SPECIFIC]** Burst changes, rename directories and switch workspace during debounce; no stale refresh or watch loop from checkpoint refs.
   - Expected: Watch is 500 ms debounced; linked-worktree per-worktree/common Git dirs are watched separately. Generated trees, Git object/log/checkpoint scratch are excluded to prevent refresh loops. No watcher backend means manual refresh remains; startup retried after workspace change, not every heartbeat.
3. **W02.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail a watcher or refresh callback and switch visible workspace/surface before completion.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Watch is 500 ms debounced; linked-worktree per-worktree/common Git dirs are watched separately. Generated trees, Git object/log/checkpoint scratch are excluded to prevent refresh loops. No watcher backend means manual refresh remains; startup retried after workspace change, not every heartbeat.
4. **W02.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W02.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W02.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W02.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- **PASS** `TestNativeWatchDetectsPreservedMetadataAndNewNestedTrees`: Native filesystem events detect metadata-preserving writes and newly nested directories. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestNativeWatchFollowsLinkedMetadataAndIgnoresCheckpointChurn`: Linked-worktree/common refs are watched without refreshing on private checkpoint churn. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestNativeWatchDebouncesBurstAndFallsBackOnOverflow`: Debounced bursts and overflow fallback preserve eventual refresh. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Native fsnotify recursive/trailing-debounce and bounded polling fallback pass Linux filesystem/race tests. Windows/macOS cross-compile only; real host event delivery, network filesystems and all UI refresh fan-out remain acceptance gates.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W03 — Resizable sidebar, Review, Explorer and terminal dimensions survive relaunch.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/layout.rs::Layout](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/layout.rs#L29); [crates/orbit-pi/src/layout.rs::record](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/layout.rs#L115); [crates/orbit-pi/src/layout.rs::flush_if_settled](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/layout.rs#L142); [crates/orbit-pi/src/sidepane.rs::set_width](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L328)

**Fixture:** Two temporary Git workspaces, fixture Pi stream, open files/tabs and independent watcher/subprocess completion controls.

1. **W03.behavior [UNTESTED; ROW_SPECIFIC]** Drag Sidebar, Review, Explorer and Terminal below minimum, to a normal size and above maximum; toggle full-width Review independently.
   - Expected: Sidebar default248/min200; Review default460/min300/drag max900; Explorer default248/min172/max460; Terminal default260/min96. Main content retains480px reserve where the source resize handler applies it. Full-width Review preserves its separate restore width.
2. **W03.edge [UNTESTED; ROW_SPECIFIC]** Write invalid/nonfinite sizes and test save failure; source clamps/defaults apply without overwriting unrelated layout state.
   - Expected: `~/.orbit-pi/layout.json`; 500 ms settle debounce, finite 80–4000 px load sanity plus per-panel clamps. Only dimensions are in this store; do not infer persisted tab/process/buffer state. Current writes are best-effort and errors are ignored.
3. **W03.error [UNTESTED; ROW_SPECIFIC]** Make the fixture ~/.orbit-pi/layout.json parent unwritable; drag and allow the settle flush. Restore permission, change a size again and flush.
   - Expected: Pinned Layout.persist ignores write errors: current live size remains, old disk state is not falsely claimed updated, and a later changed size can persist. This action has no session-scan or Pi dependency.
4. **W03.loading [UNTESTED; ROW_SPECIFIC]** Record a size change at t=0; flush at499ms and500ms. Record an identical value at400ms and repeat.
   - Expected: No disk write before500ms; one flush when changed value has settled. Recording the same value does not rearm the timer. Resizing is immediate and has no loading spinner.
5. **W03.persistence [UNTESTED; ROW_SPECIFIC]** Persist sidebar_width248, sidepane_width460, explorer_width300, terminal_height260 in ~/.orbit-pi/layout.json; relaunch. Repeat with missing, text,10,99999 and nonfinite parsed fields.
   - Expected: Only those four sizes belong to this store. Numeric finite values in inclusive80…4000 survive parsing before panel-specific clamps; invalid fields become absent/default independently. Malformed file defaults all fields. Open panels, fonts and theme are not restored by this file.
6. **W03.keyboard [UNTESTED; ROW_SPECIFIC]** After resizing use Cmd+B/Ctrl+B, Cmd+Shift+E/Ctrl+Shift+E and Cmd+J/Ctrl+J to toggle Sidebar, Explorer and Terminal; reopen each.
   - Expected: The registered source panel actions toggle visibility without treating those keys as resize commands; remembered size is retained. Review ToggleSidePanel has no chord in this command entry, so no invented key is required.
7. **W03.repeat [UNTESTED; ROW_SPECIFIC]** Perform100 pointer size updates, stop, then leave the splitter stationary and reopen the panel three times.
   - Expected: Live geometry updates without writing per pixel; one settled write per changed burst. Stationary input does not repeatedly rewrite layout.json; visibility toggles retain size.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Layout persistence code exists; settle debounce and drag/restore interactions require acceptance.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

