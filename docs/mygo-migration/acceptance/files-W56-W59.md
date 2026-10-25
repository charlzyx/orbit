# Explorer and file editor: W56–W59

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W56 — Rename file/folder within parent with open-tab subtree reconciliation; Delete confirms before dispatch, closes affected tabs without writing, refreshes tree.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/ops.rs::rename](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/ops.rs#L100); [crates/orbit-pi/src/explorer/panel.rs::start_rename](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L569); [crates/orbit-pi/src/explorer/panel.rs::request_delete](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L698); [crates/orbit-pi/src/explorer/panel.rs::delete_prompt](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1426)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W56.behavior [UNTESTED; ROW_SPECIFIC]** Rename a directory containing open tabs, then delete selected paths with confirmation.
   - Expected: Rename file/folder within parent with open-tab subtree reconciliation; Delete confirms before dispatch, closes affected tabs without writing, refreshes tree.
2. **W56.edge [UNTESTED; ROW_SPECIFIC]** Dirty affected buffers block destructive actions or follow source warning contract; cancellation must not rename/delete/write buffers.
   - Expected: Refuses dirty open tabs below target; no-overwrite rename. macOS NSFileManager Trash recoverable, Windows/Linux permanent deletion; confirmation wording is conditional. Failures remain Explorer notice. No file drag/drop move/copy implementation in audited panel.
3. **W56.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Refuses dirty open tabs below target; no-overwrite rename. macOS NSFileManager Trash recoverable, Windows/Linux permanent deletion; confirmation wording is conditional. Failures remain Explorer notice. No file drag/drop move/copy implementation in audited panel.
4. **W56.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W56.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W56.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W56.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W57 — Files has multi-tab editor, existing-tab activation preserves buffer, scrolling tab strip, close tab selects neighbor; last close exits Files. Hide preserves tabs and flushes dirty edits.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/viewer.rs::FileViewer](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L348); [crates/orbit-pi/src/explorer/viewer.rs::show](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L464); [crates/orbit-pi/src/explorer/viewer.rs::activate](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L548); [crates/orbit-pi/src/explorer/viewer.rs::close_tab](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L513)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W57.behavior [UNTESTED; ROW_SPECIFIC]** Open multiple files, reopen existing file, close middle/last tab and hide/show Files.
   - Expected: Files has multi-tab editor, existing-tab activation preserves buffer, scrolling tab strip, close tab selects neighbor; last close exits Files. Hide preserves tabs and flushes dirty edits.
2. **W57.edge [UNTESTED; ROW_SPECIFIC]** Preserve existing unsaved buffers and neighbor selection; slow old-workspace file load must not steal current active selection.
   - Expected: Tabs/buffers are in-memory entity state, no reopen-last-tabs disk store located. First load spinner, later loads keep old content; per-tab generation rejects obsolete reads. `secondary-W` tab close, `secondary-shift-W` surface close.
3. **W57.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Tabs/buffers are in-memory entity state, no reopen-last-tabs disk store located. First load spinner, later loads keep old content; per-tab generation rejects obsolete reads. `secondary-W` tab close, `secondary-shift-W` surface close.
4. **W57.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W57.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W57.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W57.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- **PASS** `TestEditorInactiveLoadNeverStealsSelection`: Background file load registers a document without changing active selection. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W58 — Code/text editing reuses ComposerInput with syntax paint, gutter, no wrap, fill-area layout; selection/navigation, cut/copy/paste, undo/redo, newlines and save shortcuts work in Editor context without sending chat.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/viewer.rs::show](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L464); [crates/orbit-pi/src/explorer/viewer.rs::seed_editor](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L626); [crates/orbit-pi/src/explorer/viewer.rs::editor_body](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1367)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W58.behavior [UNTESTED; ROW_SPECIFIC]** Edit Unicode text with gutter/no-wrap, selection, cut/paste, undo/redo/newline and Save keys.
   - Expected: Code/text editing reuses ComposerInput with syntax paint, gutter, no wrap, fill-area layout; selection/navigation, cut/copy/paste, undo/redo, newlines and save shortcuts work in Editor context without sending chat.
2. **W58.edge [UNTESTED; ROW_SPECIFIC]** Editor Enter must never send chat; long files, tabs and syntax painting must preserve source metrics and read-only constraints.
   - Expected: `Editor Files` context overrides Composer behaviors. No language-server completion/refactor/multi-file search/replacement/three-way merge or debugger implemented in audited editor.
3. **W58.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `Editor Files` context overrides Composer behaviors. No language-server completion/refactor/multi-file search/replacement/three-way merge or debugger implemented in audited editor.
4. **W58.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W58.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W58.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W58.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W59 — Autosave 500 ms debounce flushes ALL dirty tabs; manual save does same. Status chips show saved/unsaved/saving/failed/changed-on-disk. Watch reload updates clean buffer; dirty buffer retained and conflict flag shown; self-save echo preserves caret/scroll.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/viewer.rs::on_editor_changed](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L680); [crates/orbit-pi/src/explorer/viewer.rs::schedule_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L703); [crates/orbit-pi/src/explorer/viewer.rs::flush_all](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L717); [crates/orbit-pi/src/explorer/viewer.rs::flush](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L731)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W59.behavior [UNTESTED; ROW_SPECIFIC]** Edit several tabs, wait 500 ms from latest edit and manually save all; trigger self-save and external watcher reloads.
   - Expected: Autosave 500 ms debounce flushes ALL dirty tabs; manual save does same. Status chips show saved/unsaved/saving/failed/changed-on-disk. Watch reload updates clean buffer; dirty buffer retained and conflict flag shown; self-save echo preserves caret/scroll.
2. **W59.edge [UNTESTED; ROW_SPECIFIC]** Fail a save, edit during save and externally modify a dirty buffer; preserve undo history/draft, expose conflict and never discard newer edits.
   - Expected: Sibling temp-write+rename with direct-write fallback. Failure marks still-open tab dirty. Close dispatches async flush then removes tab, so close-time save failure cannot remain visible in removed tab. Conflict badge is not an autosave lock/merge resolver.
3. **W59.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Sibling temp-write+rename with direct-write fallback. Failure marks still-open tab dirty. Close dispatches async flush then removes tab, so close-time save failure cannot remain visible in removed tab. Conflict badge is not an autosave lock/merge resolver.
4. **W59.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W59.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W59.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W59.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- **PASS** `TestEditorSelfSaveReloadPreservesUndoHistory`: Self-save watcher echoes preserve undo/redo history. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestSlowSaveDoesNotBlockSnapshotsOrLoseConcurrentEdits`: A slow save does not block snapshots or erase newer edits. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestEditorDirtyReloadAndFailedSaveKeepBuffer`: External dirty conflict and failed save retain user buffer. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Global debounce, atomic mode-preserving saves, external conflict flags, edits-during-slow-save, nonblocking snapshots and retaining tabs changed during close are race-tested. Full native editor keyboard/status/visual parity remains open.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

