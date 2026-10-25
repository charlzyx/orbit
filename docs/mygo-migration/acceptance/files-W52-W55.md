# Explorer and file editor: W52–W55

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W52 — Explorer scans workspace with nested/global Git ignores and `.ignore`, keeps empty directories, directories-first case-insensitive order, hidden-files toggle, Git status badges and dirty-ancestor markings, filter including matching subtree/ancestors.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/walk.rs::snapshot](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/walk.rs#L41); [crates/orbit-pi/src/explorer/walk.rs::Snapshot](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/walk.rs#L23); [crates/orbit-pi/src/explorer/tree.rs::TreeIndex](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/tree.rs#L74); [crates/orbit-pi/src/explorer/tree.rs::StatusBadge](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/tree.rs#L19)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W52.behavior [UNTESTED; ROW_SPECIFIC]** Scan tree with nested/global ignores and .ignore, hidden toggle, empty dirs, status badges and descendant filter.
   - Expected: Explorer scans workspace with nested/global Git ignores and `.ignore`, keeps empty directories, directories-first case-insensitive order, hidden-files toggle, Git status badges and dirty-ancestor markings, filter including matching subtree/ancestors.
2. **W52.edge [UNTESTED; ROW_SPECIFIC]** Rename/delete/add directories, use symlink escape and rapid hidden/workspace switch; latest valid snapshot wins.
   - Expected: Off-thread snapshot; max 50,000 entries and depth 24, entry-cap truncation surfaced. Symlinks/special/non-UTF8 paths omitted; missing root returns empty snapshot; scan errors are skipped. Not a recursive search of file contents.
3. **W52.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Off-thread snapshot; max 50,000 entries and depth 24, entry-cap truncation surfaced. Symlinks/special/non-UTF8 paths omitted; missing root returns empty snapshot; scan errors are skipped. Not a recursive search of file contents.
4. **W52.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W52.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W52.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W52.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending. Dedicated native-controller stale source/session/workspace/hidden-toggle regressions pass under -race.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W53 — Explorer virtualized rows, active-file highlight/reveal and ancestor expansion; mouse activate expands folder or opens file. Up/Down/Home/End, Left/Right folder collapse/expand, platform+Left/Right all collapse/expand, Enter/Space activate.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/panel.rs::set_workspace](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L251); [crates/orbit-pi/src/explorer/panel.rs::set_active](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L279); [crates/orbit-pi/src/explorer/panel.rs::reveal_active](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L405); [crates/orbit-pi/src/explorer/panel.rs::on_key_down](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L417)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W53.behavior [UNTESTED; ROW_SPECIFIC]** Navigate Explorer with Up/Down/Home/End/Left/Right and platform-modified expand/collapse commands.
   - Expected: Explorer virtualized rows, active-file highlight/reveal and ancestor expansion; mouse activate expands folder or opens file. Up/Down/Home/End, Left/Right folder collapse/expand, platform+Left/Right all collapse/expand, Enter/Space activate.
2. **W53.edge [UNTESTED; ROW_SPECIFIC]** Reveal an active deep file after filtering/collapse; keyboard and pointer target same node, not stale visible indexes.
   - Expected: Inline name editor owns Enter/Escape; tree Escape closes row menu or clears filter before global abort. Workspace change resets tree/cursor/active/entry/delete/notice state.
3. **W53.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Inline name editor owns Enter/Escape; tree Escape closes row menu or clears filter before global abort. Workspace change resets tree/cursor/active/entry/delete/notice state.
4. **W53.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W53.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W53.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W53.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W54 — Explorer row context menu at pointer, viewport-flipping: folder expand/collapse, New File/Folder, Rename/Delete; file Open/Rename/Delete; both default-app open, reveal, copy absolute and relative path. Header has creation/refresh/collapse/hidden controls.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/panel.rs::header](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L756); [crates/orbit-pi/src/explorer/panel.rs::context_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1217); [crates/orbit-pi/src/explorer/panel.rs::open_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L493); [crates/orbit-pi/src/explorer/panel.rs::dismiss_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L507)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W54.behavior [UNTESTED; ROW_SPECIFIC]** Open folder/file context menus at all viewport edges; use create/rename/delete/open/reveal/copy actions.
   - Expected: Explorer row context menu at pointer, viewport-flipping: folder expand/collapse, New File/Folder, Rename/Delete; file Open/Rename/Delete; both default-app open, reveal, copy absolute and relative path. Header has creation/refresh/collapse/hidden controls.
2. **W54.edge [UNTESTED; ROW_SPECIFIC]** Cancel destructive operations and compare absolute versus workspace-relative clipboard paths; menu must stay in viewport.
   - Expected: Context menu outside-dismiss guard; OS integration varies by platform. Menu dispatch stops inappropriate row propagation.
3. **W54.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Context menu outside-dismiss guard; OS integration varies by platform. Menu dispatch stops inappropriate row propagation.
4. **W54.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W54.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W54.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W54.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W55 — New file/folder inline name field in selected folder/root, Enter confirm/Escape cancel. File creation opens new editable tab; folder creation refreshes tree; success toast/error strip.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/panel.rs::start_new_file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L535); [crates/orbit-pi/src/explorer/panel.rs::start_new_folder](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L540); [crates/orbit-pi/src/explorer/panel.rs::begin_entry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L594); [crates/orbit-pi/src/explorer/panel.rs::commit_entry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L636)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W55.behavior [UNTESTED; ROW_SPECIFIC]** Create file/folder inline in root and selected folder; confirm Enter and cancel Escape.
   - Expected: New file/folder inline name field in selected folder/root, Enter confirm/Escape cancel. File creation opens new editable tab; folder creation refreshes tree; success toast/error strip.
2. **W55.edge [UNTESTED; ROW_SPECIFIC]** Reject invalid/existing names and failed writes; successful file opens editable tab while folder refreshes tree only.
   - Expected: Name rejects dot/parent/slash/backslash/NUL/control/>255-byte names. `create_new` prevents truncation; mkdir one level and fails if existing. Disk I/O off UI thread.
3. **W55.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Name rejects dot/parent/slash/backslash/NUL/control/>255-byte names. `create_new` prevents truncation; mkdir one level and fails if existing. Disk I/O off UI thread.
4. **W55.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W55.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W55.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W55.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

