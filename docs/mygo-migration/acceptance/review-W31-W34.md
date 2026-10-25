# Review and turn checkpoints: W31–W34

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W31 — Review sources: Last Turn, Uncommitted, Unstaged, Staged, Committed and Branch. Comparisons respectively use turn refs; HEAD→worktree; index→worktree; HEAD→index; branch base→HEAD; branch base→worktree.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::resolve_range](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L333); [crates/orbit-pi/src/app.rs::sync_review_context](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2069); [crates/orbit-pi/src/sidepane.rs::source_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L2113); [crates/orbit-pi/src/sidepane.rs::set_turn_context](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L486)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W31.behavior [UNTESTED; ROW_SPECIFIC]** Create baseline HEAD plus one committed branch change, one staged file, one unstaged edit and one untracked file. Choose each of the six sources and compare both endpoints and exact changed paths.
   - Expected: Uncommitted=HEAD→worktree, Unstaged=index→worktree including untracked files, Staged=HEAD→index, Committed=main/master merge-base→HEAD, Branch=base→worktree. Last Turn uses immutable completed refs, not current live worktree or current prompt ordinal.
2. **W31.edge [UNTESTED; ROW_SPECIFIC]** Complete turn1, select Last Turn, start turn2 but withhold its end checkpoint. Refresh three times, then complete turn2. Open the explicit turn1 card afterward.
   - Expected: While turn2 runs, the existing turn1 diff remains available. Selected LastTurn{1} remains pinned after completion2 and is labeled Turn 1; choosing Last Turn again selects2. An empty-session source menu disables Last Turn.
3. **W31.error [UNTESTED; ROW_SPECIFIC]** Remove the ending ref for an explicitly selected turn; separately remove its diff-base ref and then its start ref while keeping the previous ending ref. Switch to a session with no completed refs.
   - Expected: Missing ending ref is unavailable, not an empty-success diff. Starting endpoint prefers turn-diff-N, then turn-start-N, then turn-(N-1). Switching sessions with no completed turn changes Last Turn to Uncommitted rather than retaining an error-prone old turn.
4. **W31.loading [UNTESTED; ROW_SPECIFIC]** Pause the completed-turn lookup and a source load independently; change session/workspace/source before the earlier callback completes.
   - Expected: The completed count is not inferred from transcript rows or an in-flight Turn. Obsolete source/session/workspace callbacks cannot overwrite the current view. Same-source cached diff remains during refresh; manual feedback follows W39.
5. **W31.persistence [UNTESTED; ROW_SPECIFIC]** Complete turn3, compact away old transcript messages, recreate runtime and Review, then start the next prompt. Repeat with only turn-start-9 and turn-diff-9 refs added.
   - Expected: Persisted ending refs restore completed3 and next counter4. Start/diff refs do not count as completed9. Review selected source/historical turn are controller state, not a separately invented disk preference.
6. **W31.keyboard [UNTESTED; ROW_SPECIFIC]** Open the Review source dropdown; select Last Turn by the existing native menu interaction; dismiss without choosing. Inspect source command registration for a dedicated Last Turn chord.
   - Expected: Only enabled selection changes source and records the completed turn number. Dismissal changes nothing. No dedicated Last Turn chord is invented. Historical Review buttons pass their actual turn number, not message ordinal.
7. **W31.repeat [UNTESTED; ROW_SPECIFIC]** Alternate a completed session, a first-run session and a completed historical turn three times while delayed Review loads finish out of order.
   - Expected: No previous-session file content or completed counter leaks into the new context. Source set_turn_context only falls back when the new session has no completed turn; otherwise it retains the numeric selected LastTurn{N}, which may be unavailable in that session. Record this source limitation rather than silently rewriting its selection contract.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- **PASS** `TestReviewSixSourcesAndUntrackedPaths`: Exact HEAD/index/worktree/branch changed-file fixtures, including untracked unusual paths. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestLastTurnUsesPersistedCounterAfterCompaction`: Completed checkpoint number, not transcript row count, drives Last Turn after compaction. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestReviewLastTurnKeepsCompletedDiffDuringNextPrompt`: Repeated refresh retains prior completed diff while a newer prompt has no ending ref. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestReviewHistoricalTurnStaysPinnedAndEmptySessionFallsBack`: Historical Turn1 remains pinned; empty-session switch falls back to Uncommitted. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointPlainFolderAndLegacyBaselineAreNonErrors`: Plain-folder capture no-ops; legacy previous-end baseline remains a usable fallback. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.

**Additional source-specific procedures:** [../checkpoint-acceptance.md](../checkpoint-acceptance.md), [checkpoint-evidence.json](checkpoint-evidence.json), [checkpoint-failure-history.json](checkpoint-failure-history.json), [../intentional-deviations.md](../intentional-deviations.md)

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W32 — Snapshot tracked and untracked files at turn start/end into session-keyed `refs/orbit/session-…` checkpoint commits; recover latest completed turn after restart and continue numbering.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/checkpoint.rs::latest_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L396); [crates/orbit-pi/src/checkpoint.rs::capture_turn_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L114); [crates/orbit-pi/src/app/session.rs::finish_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L227); [crates/orbit-pi/src/app/session.rs::begin_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L200)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W32.behavior [UNTESTED; ROW_SPECIFIC]** Start turn1 in a temporary repository with different HEAD/index/worktree bytes and an untracked file; capture its end after edits. Record bytes, mtime, HEAD and status before/after capture.
   - Expected: Private-index snapshots include tracked and untracked content. Real index bytes/mtime, HEAD and worktree remain unchanged. Turn-start and baseline refs are published together; completed ref and branch-aware diff base are separately recorded.
2. **W32.edge [UNTESTED; ROW_SPECIFIC]** Start a first turn without ending it; then complete it. Start a second turn while reading the previously completed diff. Repeat in an unborn repository and a non-Git folder.
   - Expected: Start does not enable a nonexistent Last Turn. Completed counter advances only after successful end capture; in-flight counter remains separate. Unborn repository can snapshot its files. Plain-folder capture no-ops; Go deliberately avoids advertising a nonexistent Git ref (deviation D06).
3. **W32.error [UNTESTED; ROW_SPECIFIC]** Place a disposable .lock file at refs/orbit/session-fixture-turn-N.lock before end capture. Repeat with no prior completed turn and with completed turn1. Remove lock and retry.
   - Expected: End capture failure retains completed0/1 respectively, emits bounded diagnostic and never claims a successful new diff. Retry can publish N. Partial object writes are not treated as user edits; real index and HEAD remain unchanged.
4. **W32.loading [UNTESTED; ROW_SPECIFIC]** Pause an earlier checkpoint task, reserve end/start tasks for successive turns, then release it. Abort while the next prompt is waiting; send duplicate agent_settled events.
   - Expected: Go reserves checkpoint work off the UI thread in order and gates new prompts during identity recovery (D07). End capture begins only for an open turn; duplicate settles cannot reserve it twice. Source begins capture asynchronously after send acceptance, so Go waiting for start capture before RPC prompt dispatch is an explicit safer ordering difference, not a source timing PASS.
5. **W32.persistence [UNTESTED; ROW_SPECIFIC]** Complete turns1 and3 for sessionA; add start/diff-only refs99 and completed7 for sessionB. Recreate manager and open A.
   - Expected: latest_turn lists only numeric ending suffixes for A; restores3, next prompt uses4, unrelated B and start/diff99 are excluded. Session keys retain ASCII alphanumerics/-/_ and replace each other Unicode scalar with one underscore.
6. **W32.keyboard [UNTESTED; ROW_SPECIFIC]** Submit from composer using its configured Enter send mode, issue Escape/Stop during the turn, then use Review without another prompt.
   - Expected: Checkpoint creation has no separate keyboard shortcut. Submission and abort dispatch use the ordinary controller path; Review remains read-only and never resets or checks out a checkpoint.
7. **W32.repeat [UNTESTED; ROW_SPECIFIC]** Reuse one runtime for a new session while an old completion is delayed; also park one running session and complete it while another is active.
   - Expected: Each checkpoint closure retains its original session/workspace/turn. Old completion may persist old refs but cannot publish its count into a replacement identity. Background sessions keep their own completed counter; reopening recovers the same refs.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- **PASS** `TestCheckpointCompletedCounterAdvancesOnlyAfterSuccessfulEnd`: Completed counter advances on success only; lock failure retains old value and retry succeeds. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointCompletionDoesNotLeakAcrossSessionIdentity`: Old completion persists old refs without publishing into reused session identity. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointFirstEndFailureDoesNotEnableLastTurn`: First end capture failure keeps Last Turn unavailable. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointPlainFolderDoesNotAdvertiseMissingDiff`: Plain-folder no-op never advertises a nonexistent Git ending ref (D06). [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointIdentityResponseResetsBothTurnCounters`: Authoritative new session ID resets in-flight and completed counters. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestManagerRestoresCompletedTurnDuringNextPrompt`: Real manager restores completed3, starts4 while retaining3, then settles an isolated turn4 diff. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointIdentityRecoveryRejectsStaleGenerationAndLiveTurn`: Obsolete recovery cannot release another identity gate or rewind a live counter. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestManagerPromptGatePreservesExistingRefDuringCheckpointRecovery`: Pending recovery rejects prompt before reservation; safe retry preserves prior ref1 and starts2 (D07). [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointRecoveryFailureStaysGatedUntilSuccessfulRetry`: Actual ref-listing failure remains not-ready until repaired recovery publishes a trustworthy counter (D07). [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointDoesNotChangeIndexHeadOrWorktree`: Repeated captures preserve real index bytes/mtime, HEAD/status and remove temporary indexes. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointTurnFreezesDiffAndRestoresCounter`: Completed turn remains immutable after later edits and recovers persisted ending count. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointUnbornAndCancellation`: Unborn capture succeeds; pre-cancelled operation writes no checkpoint ref. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointStartPrecedesImmediatePiToolMutation`: Actual Pi fixture writes immediately on prompt receipt; start baseline still excludes that mutation (D08). [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.

**Additional source-specific procedures:** [../checkpoint-acceptance.md](../checkpoint-acceptance.md), [checkpoint-evidence.json](checkpoint-evidence.json), [checkpoint-failure-history.json](checkpoint-failure-history.json), [../intentional-deviations.md](../intentional-deviations.md)

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W33 — Last Turn attribution handles branch switches during or between turns and dirty files on target branch; starting metadata retains HEAD, branch and repository refs.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/checkpoint.rs::TurnStartMetadata](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L30); [crates/orbit-pi/src/checkpoint.rs::prepare_turn_diff_base](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L241); [crates/orbit-pi/src/checkpoint.rs::target_branch_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L285); [crates/orbit-pi/src/checkpoint.rs::virtual_branch_start](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs#L317)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W33.behavior [UNTESTED; ROW_SPECIFIC]** Create main and an existing target branch with target-only files. Capture a turn start on main, switch to target and add only agent.txt, then capture end.
   - Expected: Last Turn contains agent.txt only, not target branch preexisting files. Metadata captures starting HEAD, symbolic branch and all local/remote/tag commit refs; annotated tags are peeled.
2. **W33.edge [UNTESTED; ROW_SPECIFIC]** Carry an untracked pre-turn dirty file to target; then create a new branch during a separate turn and commit agent changes there. Repeat with detached HEAD and an unborn start.
   - Expected: Virtual branch start retains pre-turn dirty content on the target so it is not attributed to the turn. Newly created branch commits remain attributed. Detached comparison uses HEAD identity; an absent target ancestor uses an empty-tree commit.
3. **W33.error [UNTESTED; ROW_SPECIFIC]** Force virtual merge-tree to fail or corrupt the stored turn-start metadata. Inspect end ref and attempt Review through its fallback endpoints.
   - Expected: capture_turn reports the failure and does not falsely advance completed UI state. Existing ending object/ref may remain because end publication precedes diff-base preparation. Review falls back only through documented diff/start/previous refs; no worktree rollback occurs.
4. **W33.loading [UNTESTED; ROW_SPECIFIC]** Delay branch-aware diff-base preparation while switching the visible session and workspace.
   - Expected: Background capture operates on the original captured path/identity; completion must not relabel another session's Last Turn. The UI may retain its previous completed result until this capture succeeds.
5. **W33.persistence [UNTESTED; ROW_SPECIFIC]** Record metadata and completed diff refs, switch branches again and edit files after settle, then restart and view the original completed turn.
   - Expected: Original immutable refs still produce the same diff; later branch switches/terminal edits are not retroactively attributed. Metadata lives in the start snapshot commit message under Orbit-Turn-Start, not a separate user index or config file.
6. **W33.keyboard [UNTESTED; ROW_SPECIFIC]** Open a completed turn through its Review control and navigate its changed files; inspect available actions for Restore/Rewind.
   - Expected: This is read-only review. No user-facing checkpoint rewind/restore action is invented and no navigation key mutates HEAD/index/worktree.
7. **W33.repeat [UNTESTED; ROW_SPECIFIC]** Perform three turns with a branch switch between1/2 and during2, editing one distinct file each time. Compare exact per-turn paths after turn3.
   - Expected: Each turn compares its own start to its own end. A branch switch between turns is outside both turns; existing target changes are excluded; each new agent change appears once in its own completed turn.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- **PASS** `TestCheckpointBranchSwitchExcludesExistingTargetChanges`: Existing target branch changes and pre-turn dirty files are excluded from the new agent diff. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.
- **PASS** `TestCheckpointNewBranchIncludesCommittedAgentWork`: Agent work committed on a newly created branch remains attributed to that turn. [Evidence](checkpoint-evidence.json). Tested dirty-tree manifest `94755581807bf98b08a0bbd5550473abc7e5a7dc796e5f8654f6428aa4333b93`, not a published commit.

**Additional source-specific procedures:** [../checkpoint-acceptance.md](../checkpoint-acceptance.md), [checkpoint-evidence.json](checkpoint-evidence.json), [checkpoint-failure-history.json](checkpoint-failure-history.json), [../intentional-deviations.md](../intentional-deviations.md)

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
- **PASS** `TestPinnedSourceOracle`: All 1,204 exact pinned-source lexical cases match across 16 languages. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestReviewHighlightCarriesSeparateSidesAndResetsAtHunks`: Old/new carries are independent; context uses new side; file/hunk boundaries reset state. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestReviewSourceHeaderPathsMetadataAndSyntheticHunks`: File identity follows diff paths, source-hidden metadata stays hidden, synthetic hunks have no invented positions. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestReviewHeaderEscapesAndCompactAsymmetricGaps`: Quoted/space paths decode and compact gaps account for both old and new ranges. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestDiffColorSpansPreserveEveryByteAndFontMetrics`: Git paint spans concatenate exact text and change color only, including dynamic palette callback. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Remaining:** Paint-only lexer matches 1,204 pinned Rust oracle cases; separate-side carries, hidden context, split/group retention and Git color-only spans tested. Full parser variant audit, transcript/editor mounting and native rendered parity remain pending. See G9 and highlight-parity.md; historical Waku license claim remains unverified.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

