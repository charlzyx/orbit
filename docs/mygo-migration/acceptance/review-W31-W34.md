# Review and turn checkpoints: W31–W34

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W31 — Review sources: Last Turn, Uncommitted, Unstaged, Staged, Committed and Branch. Comparisons respectively use turn refs; HEAD→worktree; index→worktree; HEAD→index; branch base→HEAD; branch base→worktree.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/review.rs::Source](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L28); [crates/orbit-pi/src/git.rs::collect_review_diff](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L304); [crates/orbit-pi/src/git.rs::resolve_range](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L333); [crates/orbit-pi/src/git.rs::branch_base](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L426)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W31.behavior [UNTESTED; ROW_SPECIFIC]** Create distinct baseline/index/worktree/branch/turn states; select all six Review sources.
   - Expected: Review sources: Last Turn, Uncommitted, Unstaged, Staged, Committed and Branch. Comparisons respectively use turn refs; HEAD→worktree; index→worktree; HEAD→index; branch base→HEAD; branch base→worktree.
2. **W31.edge [UNTESTED; ROW_SPECIFIC]** Use unborn HEAD, no main/master and absent turn checkpoint; compare exact trees and honest source-specific unavailable state.
   - Expected: Default branch comparison searches local main/master distinct from current, else HEAD; unborn HEAD uses empty tree. Last Turn unavailable without session/completed checkpoint and falls back on session switch. Exact existing semantics must not be replaced by a generic `git diff`.
3. **W31.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Default branch comparison searches local main/master distinct from current, else HEAD; unborn HEAD uses empty tree. Last Turn unavailable without session/completed checkpoint and falls back on session switch. Exact existing semantics must not be replaced by a generic `git diff`.
4. **W31.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W31.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W31.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W31.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W32 — Snapshot tracked and untracked files at turn start/end into session-keyed `refs/orbit/session-…` checkpoint commits; recover latest completed turn after restart and continue numbering.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/checkpoint.rs::capture_turn_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L114); [crates/orbit-pi/src/checkpoint.rs::capture_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L155); [crates/orbit-pi/src/checkpoint.rs::capture_worktree_commit](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L189); [crates/orbit-pi/src/checkpoint.rs::checkpoint_ref](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L90)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W32.behavior [UNTESTED; ROW_SPECIFIC]** Capture tracked/untracked turn-start/end snapshots, restart manager and continue turn numbering.
   - Expected: Snapshot tracked and untracked files at turn start/end into session-keyed `refs/orbit/session-…` checkpoint commits; recover latest completed turn after restart and continue numbering.
2. **W32.edge [UNTESTED; ROW_SPECIFIC]** Verify real index/HEAD/worktree unchanged, concurrent checkpoints serialized, non-repository no-op and failed capture diagnostic.
   - Expected: Isolated temporary index preserves real index/HEAD/worktree; ref-safe session key and atomic ref updates. Non-repo capture no-ops. Index read uses copied temporary index to avoid file-watch loops.
3. **W32.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Isolated temporary index preserves real index/HEAD/worktree; ref-safe session key and atomic ref updates. Non-repo capture no-ops. Index read uses copied temporary index to avoid file-watch loops.
4. **W32.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W32.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W32.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W32.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W33 — Last Turn attribution handles branch switches during or between turns and dirty files on target branch; starting metadata retains HEAD, branch and repository refs.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/checkpoint.rs::TurnStartMetadata](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L30); [crates/orbit-pi/src/checkpoint.rs::prepare_turn_diff_base](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L241); [crates/orbit-pi/src/checkpoint.rs::target_branch_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L285); [crates/orbit-pi/src/checkpoint.rs::virtual_branch_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L317)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W33.behavior [UNTESTED; ROW_SPECIFIC]** Switch branches during a turn and start with dirty changes on the destination branch.
   - Expected: Last Turn attribution handles branch switches during or between turns and dirty files on target branch; starting metadata retains HEAD, branch and repository refs.
2. **W33.edge [UNTESTED; ROW_SPECIFIC]** Last Turn must exclude unrelated branch switch and preexisting destination dirt while retaining actual agent changes.
   - Expected: Branch-aware virtual start/diff base prevents a whole branch switch from appearing as agent edits. Checkpoint display is read-only review, not a user-facing rewind/restore workflow.
3. **W33.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Branch-aware virtual start/diff base prevents a whole branch switch from appearing as agent edits. Checkpoint display is read-only review, not a user-facing rewind/restore workflow.
4. **W33.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W33.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W33.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W33.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W34 — Review parses file/hunk/section headers, added/deleted/context/meta lines, old/new line numbers, binary rows, totals and paint-only syntax. Diff/stat reads are off-thread.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::MAX_HYDRATED_PATCH_BYTES](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L31); [crates/orbit-pi/src/git.rs::collect_review_diff_inner](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L312); [crates/orbit-pi/src/git.rs::diff_output](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L445); [crates/orbit-pi/src/review.rs::parse_collected](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/review.rs#L350)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W34.behavior [UNTESTED; ROW_SPECIFIC]** Parse multi-file, binary, grouped, quoted-path, synthetic and full/compact patches; inspect line numbers and syntax.
   - Expected: Review parses file/hunk/section headers, added/deleted/context/meta lines, old/new line numbers, binary rows, totals and paint-only syntax. Diff/stat reads are off-thread.
2. **W34.edge [UNTESTED; ROW_SPECIFIC]** Reverse numstat order, omit a file from stats, cross multiline token carries, exceed 32 MiB/50k limits and expand hidden context.
   - Expected: Full context retained only up to 32 MiB hydrated patch, otherwise 3-line compact diff with honest nonexpandable gaps; visible rendering cap 50,000 rows and truncation flag. No fabricated content past cap.
3. **W34.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail a first load and same-source refresh separately; finish an earlier request after a newer source/workspace selection.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Full context retained only up to 32 MiB hydrated patch, otherwise 3-line compact diff with honest nonexpandable gaps; visible rendering cap 50,000 rows and truncation flag. No fabricated content past cap.
4. **W34.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W34.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W34.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W34.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`, `internal/highlight`, `internal/gitworkbench/embedded_diff.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`, `internal/highlight/highlight_test.go`, `internal/workspace/review_highlight_test.go`, `internal/gitworkbench/highlight_test.go`

**Executed narrow assertions:**
- **PASS** `TestPinnedSourceOracle`: All 1,204 exact pinned-source lexical cases match across 16 languages. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestReviewHighlightCarriesSeparateSidesAndResetsAtHunks`: Old/new carries are independent; context uses new side; file/hunk boundaries reset state. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestReviewSourceHeaderPathsMetadataAndSyntheticHunks`: File identity follows diff paths, source-hidden metadata stays hidden, synthetic hunks have no invented positions. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestReviewHeaderEscapesAndCompactAsymmetricGaps`: Quoted/space paths decode and compact gaps account for both old and new ranges. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.
- **PASS** `TestDiffColorSpansPreserveEveryByteAndFontMetrics`: Git paint spans concatenate exact text and change color only, including dynamic palette callback. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Paint-only lexer matches 1,204 pinned Rust oracle cases; separate-side carries, hidden context, split/group retention and Git color-only spans tested. Full parser variant audit, transcript/editor mounting and native rendered parity remain pending. See G9 and highlight-parity.md; historical Waku license claim remains unverified.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

