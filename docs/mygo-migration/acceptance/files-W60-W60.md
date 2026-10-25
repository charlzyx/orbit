# Explorer and file editor: W60–W60

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W60 — Markdown rendered preview/source toggle uses live edited buffer; image preview for PNG/JPEG/GIF/WebP/BMP/TIFF/SVG; code/text read-only fallback for lossy/truncated content; binary/too-large/read-error honest states. Toolbar shows relative path, byte size/lines/truncation/save status.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/explorer/viewer.rs::Mode](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L52); [crates/orbit-pi/src/explorer/viewer.rs::FileContent::editable](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L91); [crates/orbit-pi/src/explorer/viewer.rs::load](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L122); [crates/orbit-pi/src/explorer/viewer.rs::is_image](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L234)

**Fixture:** Temporary directory with Unicode/binary/images/large files, nested ignores, editable documents and controllable I/O/watch callbacks.

1. **W60.behavior [UNTESTED; ROW_SPECIFIC]** Toggle Markdown preview/source from live buffer and load all source image formats, binary and oversized files.
   - Expected: Markdown rendered preview/source toggle uses live edited buffer; image preview for PNG/JPEG/GIF/WebP/BMP/TIFF/SVG; code/text read-only fallback for lossy/truncated content; binary/too-large/read-error honest states. Toolbar shows relative path, byte size/lines/truncation/save status.
2. **W60.edge [UNTESTED; ROW_SPECIFIC]** Use invalid UTF-8, truncated/unreadable file and missing image decoder; read-only fallback must stay bounded and never save truncated text.
   - Expected: 2 MiB file cap before read (including images); 200,000-line cap; NUL sniff first 8,000 bytes; invalid UTF-8 decoded lossily but never editable. Image format guard does not imply ICO/AVIF decode support.
3. **W60.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail read/save/rename/delete in the temporary root, then deliver an external watch event or late file load.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: 2 MiB file cap before read (including images); 200,000-line cap; NUL sniff first 8,000 bytes; invalid UTF-8 decoded lossily but never editable. Image format guard does not imply ICO/AVIF decode support.
4. **W60.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W60.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W60.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W60.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/files.go`, `internal/workspace/editor.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`, `internal/ui/files.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/files_test.go`, `internal/workspace/tree_test.go`, `internal/workspace/watch_test.go`

**Executed narrow assertions:**
- **PASS** `TestFileLoadingModesAndGuards`: Text/image/binary/loading guards use actual file content and bounds. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestEditorReadonly`: Read-only documents reject edits and saves. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Remaining:** File/ignore/editor/autosave/read-only/failure tests pass; native explorer/editor keyboard and visual acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

