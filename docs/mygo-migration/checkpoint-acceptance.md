# Completed-turn Review: source-first acceptance and observed correction

Scope: W31–W33. Functional **PARTIAL**, visual **UNTESTED**, native OS
**UNTESTED**. The named assertions below pass; these are not whole-feature or
pixel-parity verdicts. All repositories, session paths and children are disposable
cloud fixtures. No provider inference, external remote mutation or user Mac.

## Pinned source contract

Baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

- [`app/session.rs::begin_turn` and `finish_turn`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L200):
  `turn_count` identifies the accepted current prompt; `latest_turn` changes only
  after successful end capture. Duplicate settle closes `turn_open` synchronously.
- [`app.rs::sync_review_context`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2069):
  Review receives the completed number, not the in-flight counter.
- [`sidepane.rs::set_turn_context`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L486),
  `show_review_turn`, `last_turn_source`, `source_label`, `source_menu`:
  no completed checkpoint disables Last Turn; changing to an empty session falls
  back to Uncommitted. A selected `LastTurn{N}` remains that concrete N when a new
  turn completes, then is labeled `Turn N`; it does not silently follow latest.
- [`git.rs::resolve_range`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L333):
  start endpoint preference is diff-base, start, previous end; ending ref is
  required. Other sources use exact HEAD/index/worktree/branch-base endpoints.
- [`checkpoint.rs`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/checkpoint.rs):
  private indexes avoid touching the real index; snapshot refs are session-keyed;
  branch-aware virtual starts exclude preexisting target/dirty changes.

## Failure observed before correction

`TestReviewLastTurnKeepsCompletedDiffDuringNextPrompt` built a real temporary
repository, completed turn1, displayed it, then started turn2 without an end ref.
Refreshing Review failed with `turn's ending checkpoint unavailable` and cleared
the prior files. Go had used the incremented `Snapshot.Turn` as Last Turn.

[Original failure record](acceptance/checkpoint-failure-history.json) retains the
red result. Its hashes were captured after observation while development was
moving, so they are explicitly **not** an immutable pre/post execution manifest.

## Correction and exact executed assertions

Go now keeps `Turn` and `LatestTurn` separate. Successful end capture publishes
the latter only for the same session/workspace; failed capture retains the prior
completed number. Native Review stores a selected historical turn independently,
uses completed availability, and rejects stale context callbacks.

| Procedure and expected result | Executed assertion |
|---|---|
| Start first turn: no completed number; end successfully:1; start2: retain1; fail end-ref update with disposable lock: retain1; remove lock and retry:2 | `TestCheckpointCompletedCounterAdvancesOnlyAfterSuccessfulEnd` |
| Fail first-ever end-ref update: Last Turn remains unavailable, diagnostic retained | `TestCheckpointFirstEndFailureDoesNotEnableLastTurn` |
| Delay an old completion, replace session identity, then finish old capture: old refs persist but new UI count stays0 | `TestCheckpointCompletionDoesNotLeakAcrossSessionIdentity` |
| Receive a different authoritative session ID: reset both old counters | `TestCheckpointIdentityResponseResetsBothTurnCounters` |
| Reopen persisted completed3, submit next prompt: current4/completed3 until actual settle; then completed4 contains only new edits | `TestManagerRestoresCompletedTurnDuringNextPrompt` |
| Pause identity recovery with existing ref1: prompt is rejected without reserving a turn; recover1 then retry uses2, leaving ref1 unchanged | `TestManagerPromptGatePreservesExistingRefDuringCheckpointRecovery` |
| Finish two identity lookups out of order: stale generation cannot release newer readiness gate; a late lookup cannot rewind a newer live counter | `TestCheckpointIdentityRecoveryRejectsStaleGenerationAndLiveTurn` |
| Corrupt only disposable packed-refs so listing really fails: not-ready/error persists; repair and retry recovers4 before readiness | `TestCheckpointRecoveryFailureStaysGatedUntilSuccessfulRetry` |
| Child writes immediately on prompt receipt, before agent_start: the completed diff still contains that write | `TestCheckpointStartPrecedesImmediatePiToolMutation` (D08) |
| Non-Git prompt: no fake completed Git ref and no false capture error | `TestCheckpointPlainFolderDoesNotAdvertiseMissingDiff` (D06) |
| Start next prompt while Last Turn1 is displayed: repeated refreshes still show completed1 | `TestReviewLastTurnKeepsCompletedDiffDuringNextPrompt` |
| Select historical1 with completed2: repeated refresh stays on `Turn 1`; switch to no-completion session: Uncommitted | `TestReviewHistoricalTurnStaysPinnedAndEmptySessionFallsBack` |
| Compact transcript history: persisted completed3 still drives Review, not assistant-message count | `TestLastTurnUsesPersistedCounterAfterCompaction` |
| Snapshot repeatedly: index bytes/mtime, HEAD/status unchanged; temporary index cleaned | `TestCheckpointDoesNotChangeIndexHeadOrWorktree` |
| Later filesystem edits do not change a completed turn's diff; latest ending number recovers | `TestCheckpointTurnFreezesDiffAndRestoresCounter` |
| Existing branch switch excludes target content; newly created branch commits include actual agent changes | `TestCheckpointBranchSwitchExcludesExistingTargetChanges`, `TestCheckpointNewBranchIncludesCommittedAgentWork` |
| Unborn repository and cancellation; plain-folder no-op and legacy baseline fallback | `TestCheckpointUnbornAndCancellation`, `TestCheckpointPlainFolderAndLegacyBaselineAreNonErrors` |
| HEAD/index/worktree/branch source fixtures return exact independently expected changed-file counts | `TestReviewSixSourcesAndUntrackedPaths` |

[Executed command, named outcomes and unchanged-before/after file hashes](acceptance/checkpoint-evidence.json)
identify the dirty-tree content manifest. No evidence is attributed to the older
published implementation SHA. Existing tests elsewhere remain candidates, not
automatic acceptance for this scope.

## Original Rust oracle

`acceptance/run-checkpoint-oracle.py` directly imports the unchanged pinned
`checkpoint.rs`, checks its SHA256, and runs its five existing tests in temporary
repositories. Only a Git subprocess adapter and error-translation stub replace
app dependencies. No GPUI or app runtime is exercised.

All five original tests passed using Rust1.94.0 and already-cached
anyhow1.0.104/serde_json1.0.151, with Cargo offline. This is useful independent
domain reference evidence, not a native UI baseline or proof of every Go case.

## Remaining and deliberate safety differences

- [W31–W34 procedures](acceptance/review-W31-W34.md) contain the exact source
  steps still to execute, including detached/unborn branch attribution, invalid
  metadata, source-menu key routing and parked/background lifecycle combinations.
- D06 keeps a non-Git Last Turn unavailable; D07 gates a prompt during recovery
  and after actual recovery failure to prevent overwriting existing numbered
  refs. D08 completes the start-capture attempt before sending the RPC prompt. See [intentional deviations](intentional-deviations.md).
- Per-turn transcript card identity is a separate host mapping requirement. A
  generic tail Review button is not proof that every historical card targets its
  actual turn. The UI must not fabricate turn identity from visible row ordinals.
- Same-state Rust/Go images, native OS focus/menu behavior, GPU rendering and
  advertised-platform execution remain unverified.
