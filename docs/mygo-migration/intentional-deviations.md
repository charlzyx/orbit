# Explicit implementation deviations requiring review

These are **not unqualified 1:1 parity passes**. The pinned behavioral source is
Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. The Go choices below deliberately
avoid data loss, false success or fabricated runtime facts. This document records
what differs; it does not assert that the user accepted a changed product contract.
Keep each affected acceptance scope PARTIAL until the exception is explicitly
resolved and its remaining ordinary behavior is verified.

## D01 — Failed tab close keeps the user's draft (W57, W59)

- Source: [`explorer/viewer.rs::close_tab`, lines513–535](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L513)
  schedules `flush`, then immediately removes the tab. `flush`, lines731–767,
  reports errors only if that tab still exists when the asynchronous write returns.
- Go: `internal/workspace/editor.go::Editor.Close` waits for that document's
  write off the rendering thread and retains the tab/draft after write failure or
  an edit arriving during the write. A later successful retry closes it.
- Reason: removing the sole in-memory copy before write success risks losing a
  draft. We do not reproduce that unsafe ordering merely to obtain a parity label.
- Narrow evidence: `TestEditorFailedCloseKeepsDraftAndCanRetry`,
  `TestCloseKeepsTabIfUserEditsDuringSlowSave`. Native close-button/host lifecycle
  and full editor acceptance remain separate.

## D02 — Discard reports failure rather than claiming success (W09, W13)

- Source: [`git.rs::discard_paths`, lines672–686](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L672)
  ignores both checkout and clean results and returns success. `discard_unstaged`
  ignores checkout failure before returning the clean result.
- Go: `internal/workspace/git.go::Discard` checks selected tracked paths, restores
  worktree content from the index, stops on restore failure, and returns clean
  failures. Cancelled and unavailable commands are errors, not successful discard.
- Preserved contract: staged entries stay intact; only selected unstaged/untracked
  data is discarded after the caller's confirmation. Ignored files and symlink
  targets outside the temporary workspace are not cleaned.
- Narrow evidence: `TestGitDiscardFailuresDoNotPretendSuccess`,
  `TestGitDiscardAllPreservesStagedIgnoredAndExternalTargets`,
  `TestGitDiscardPreservesIndex`. No external repository is used by these tests.

## D03 — Runtime uptime uses actual child-spawn time (P02, S08)

- Source: [`app/runtime.rs::adopt_client`, lines127–142](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L127)
  resets `started_at` to now. `app/session.rs` also adopts a previously parked
  child when returning to it, so its displayed uptime resets without a new PID.
- Go: the runtime snapshot keeps the observed child-spawn timestamp for the
  lifetime of that PID. Returning to a live session does not invent a new process
  age; an actual restart gets a new PID/start time.
- Evidence and exact lifecycle scenarios: [Runtime acceptance](runtime-acceptance.md).
  This is a truthfulness correction, not a claim to match the old displayed age.

## D04 — Unobserved retry state stays unknown (P04)

- Source Agent settings update the auto-retry setting optimistically. Official
  Pi1.4.0 supports `set_auto_retry`, but its `get_state` does not report that value.
- Go starts with unknown retry state, does not mutate user settings to discover
  it, and confirms a value only after the setter acknowledgement. Failed, timed
  out or interrupted requests retain/clear knowledge according to whether the
  actual outcome is known; they do not invent an enabled value.
- Evidence: [Runtime acceptance](runtime-acceptance.md), including explicit
  credential-free setter tests and the actual RPC command contract.

## How to assess these differences

Run the ordinary source behavior and the deviation-specific safety assertion
separately. A safety assertion may PASS while strict source equivalence remains
PARTIAL. Do not erase an observed difference, relabel the source behavior, or
silently use these exceptions to justify unrelated UX changes. This list is not
permission to change further product behavior without source review.

## D05 — Worktree preferences and setup consent fail transactionally (W43, W44)

- Source: `app/worktrees.rs::toggle_worktree_run_setup_auto` and
  `toggle_worktree_open_after_create`, lines968–978, update memory and ignore
  `persist()` failures. `confirm_worktree_dialog` AllowSetup, lines273–289,
  likewise ignores failed trust persistence before creating the worktree.
- Go: `internal/worktreeui/state.go` only adopts a new persisted configuration
  after save success. Failed consent persistence prevents setup-enabled creation;
  retry retains the original requested create operation rather than running it
  without a durable trust record.
- Evidence: `TestSettingsPersistenceFailureIsTransactionalAndRetryable`,
  `TestSetupAllowRemembersCanonicalRepositoryBeforeCreation`, and
  `TestSetupConsentFailureAndRetry`. This is a documented failure-path difference,
  not a reason to change successful setup semantics or mark full parity.

## D06 — A plain folder does not advertise a nonexistent completed Git diff (W31, W32)

- Source: `checkpoint.rs::capture_turn` returns success for a non-Git directory;
  `app/session.rs::finish_turn` then assigns `latest_turn` on that success. A later
  Last Turn collection still fails because the folder has no Git repository.
- Go: completion additionally confirms that its ending ref exists before exposing
  `Snapshot.LatestTurn`. Ordinary non-Git prompts remain supported and capture
  remains a silent no-op, but Last Turn stays unavailable instead of advertising
  a nonexistent saved diff.
- Narrow evidence: `TestCheckpointPlainFolderDoesNotAdvertiseMissingDiff`.
  This honest-availability difference remains separate from completed-turn
  behavior in actual repositories, which follows the pinned source contract.

## D07 — Turn recovery gates new prompts instead of risking ref-number reuse (W32)

- Source: `app/session.rs::recover_latest_turn` starts a background lookup and
  later raises `turn_count` to the recovered maximum. It does not explicitly gate
  a prompt while that counter is unknown; failed lookup yields no completed turn.
- Go: identity recovery marks the session temporarily not-ready. Prompt attempts
  are rejected before reserving a number, preserving the composer draft. A genuine
  Git ref/probe failure keeps this gate with a visible recovery error; restart or
  reopen provides a bounded retry after the repository fault is corrected. Plain
  non-Git folders clear the gate normally.
- Reason: starting from a guessed zero while saved refs exist could overwrite a
  previous completed checkpoint. This is an explicit ordering/failure-path safety
  difference, not a claim that Rust displayed the same interim status.
- Narrow evidence: `TestManagerPromptGatePreservesExistingRefDuringCheckpointRecovery`,
  `TestCheckpointIdentityRecoveryRejectsStaleGenerationAndLiveTurn`, and
  `TestCheckpointRecoveryFailureStaysGatedUntilSuccessfulRetry`.

## D08 — Capture the start before Pi can mutate the worktree (W32)

- Source: `app/session.rs::submit_composer_as` first sends the prompt, then
  `begin_turn` schedules its start snapshot asynchronously. A very fast tool can
  therefore begin before that snapshot finishes.
- Go: the prompt worker awaits its reserved start-capture attempt before sending
  the RPC prompt. Capture failure is still reported without blocking a normal
  non-Git prompt; successive end/start captures are serialized off the UI thread.
- This slightly changes dispatch latency but prevents early agent edits from
  being accidentally included in the supposed pre-turn baseline. UI waiting
  feedback begins immediately; it is not falsely equated with provider start.
- Narrow evidence: `TestCheckpointStartPrecedesImmediatePiToolMutation` makes an actual
  fixture Pi child write immediately on receiving prompt, before emitting
  `agent_start`, and verifies that write appears in the completed diff.
