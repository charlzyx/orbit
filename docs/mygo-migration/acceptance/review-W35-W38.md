# Review and turn checkpoints: W35–W38

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W35 — Long unchanged context collapses around changes; directional expand from either edge, both edges/count, or all reveals retained lines.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/review.rs::Gap](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L76); [crates/orbit-pi/src/review.rs::ExpansionDirection](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L68); [crates/orbit-pi/src/review.rs::Snapshot::expand_gap](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L208); [crates/orbit-pi/src/review.rs::collapse_context](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L655)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W35.behavior [UNTESTED; ROW_SPECIFIC]** Collapse long leading/between/trailing context and expand start/end/both/all repeatedly.
   - Expected: Long unchanged context collapses around changes; directional expand from either edge, both edges/count, or all reveals retained lines.
2. **W35.edge [UNTESTED; ROW_SPECIFIC]** Near the 50k visible cap and compact fallback, disable impossible expansion; retain stable gap IDs and unchanged source lines.
   - Expected: 3-line context around changes; directional increments of 100; stable gap IDs; file/section header indexes recomputed and virtual row count spliced correctly. Compact fallback never claims expandable bytes it lacks.
3. **W35.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: 3-line context around changes; directional increments of 100; stable gap IDs; file/section header indexes recomputed and virtual row count spliced correctly. Compact fallback never claims expandable bytes it lacks.
4. **W35.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W35.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W35.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W35.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- **PASS** `TestReviewHighlightSurvivesHiddenExpansionAndGrouping`: Token spans survive retained context expansion, grouping and split conversion. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestReviewVisibleLimitAndExpansionCap`: Visible-line cap remains 50,000 and expansion cannot exceed it. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W36 — Review wrap/no-wrap, unified/split diff, per-file and all-file collapse/expand, sticky file/section headers, horizontal pan with pinned header, dock resize and full-width toggle.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/sidepane.rs::toggle_wrap](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L392); [crates/orbit-pi/src/sidepane.rs::toggle_split](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L399); [crates/orbit-pi/src/sidepane.rs::toggle_all_files](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L406); [crates/orbit-pi/src/sidepane.rs::toggle_file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L423)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W36.behavior [UNTESTED; ROW_SPECIFIC]** Toggle wrap, unified/split, file/all collapse, full-width and horizontal pan across several changed files.
   - Expected: Review wrap/no-wrap, unified/split diff, per-file and all-file collapse/expand, sticky file/section headers, horizontal pan with pinned header, dock resize and full-width toggle.
2. **W36.edge [UNTESTED; ROW_SPECIFIC]** Collapse a preceding file while viewing a later hunk, switch layout and resize; sticky headers and selection/scroll must stay on semantic targets.
   - Expected: Pair deletion/addition replacements across two columns; unbalanced additions/deletions get filler. Width constraints change stats/tree visibility; full-width can exceed dock cap and restores prior width. Wrap/split/collapse choices are entity state, not shown persisted to disk.
3. **W36.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Pair deletion/addition replacements across two columns; unbalanced additions/deletions get filler. Width constraints change stats/tree visibility; full-width can exceed dock cap and restores prior width. Wrap/split/collapse choices are entity state, not shown persisted to disk.
4. **W36.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W36.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W36.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W36.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending. Full-width mode, per-file collapse, sticky headers and horizontal panning are not fully wired.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W37 — Changed-files tree groups folders, expands/collapses directories, case-insensitive filter, selection/highlight, jump to file and autoexpansion of new directories.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/review.rs::directory_paths](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L921); [crates/orbit-pi/src/review.rs::tree_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L940); [crates/orbit-pi/src/sidepane.rs::sync_tree_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L717); [crates/orbit-pi/src/sidepane.rs::apply_snapshot](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L661)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W37.behavior [UNTESTED; ROW_SPECIFIC]** Filter changed-files tree, expand folders and select nested files; refresh with added directories.
   - Expected: Changed-files tree groups folders, expands/collapses directories, case-insensitive filter, selection/highlight, jump to file and autoexpansion of new directories.
2. **W37.edge [UNTESTED; ROW_SPECIFIC]** Use same basename in different directories and collapse ancestors; selection/jump/autoexpansion must not target row ordinals.
   - Expected: Tree list virtualized and cached, recomputed only filter/structure change. Refresh attempts to preserve selected path; pending file choice wins. Tree is hidden below responsive pane width.
3. **W37.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Tree list virtualized and cached, recomputed only filter/structure change. Refresh attempts to preserve selected path; pending file choice wins. Tree is hidden below responsive pane width.
4. **W37.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W37.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W37.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W37.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending. Current native file list is not the complete source tree/filter interaction.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W38 — Review keyboard context: Up/Down tree cursor, Enter/Space toggle/open, n/p next/previous file, ]/[ next/previous hunk, e/c expand/collapse all, Escape returns to composer/closes Review rather than aborting agent.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/commands.rs::BINDINGS](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commands.rs#L794); [crates/orbit-pi/src/sidepane.rs::review_tree_move](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L908); [crates/orbit-pi/src/sidepane.rs::review_tree_toggle](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L922); [crates/orbit-pi/src/sidepane.rs::review_move_file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L933)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W38.behavior [UNTESTED; ROW_SPECIFIC]** Focus Review and exercise Up/Down, Enter/Space, n/p, ]/[, e/c and Escape.
   - Expected: Review keyboard context: Up/Down tree cursor, Enter/Space toggle/open, n/p next/previous file, ]/[ next/previous hunk, e/c expand/collapse all, Escape returns to composer/closes Review rather than aborting agent.
2. **W38.edge [UNTESTED; ROW_SPECIFIC]** Repeat with split view and collapsed earlier files; Escape restores composer/Review state without aborting the agent.
   - Expected: Context binding order wins over global abort; filter Escape clears filter appropriately. Selected file is expanded before scrolling.
3. **W38.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Context binding order wins over global abort; filter Escape clears filter appropriately. Selected file is expanded before scrolling.
4. **W38.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W38.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W38.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W38.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Source-specific n/p, bracket hunk navigation, expand/collapse and focus-return keyboard acceptance remains open.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

