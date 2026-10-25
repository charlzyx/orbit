# Git changes, history, graph and repository operations: W12–W15

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W12 — Fetch, push/publish with upstream creation, ff-only pull, merge upstream, rebase upstream with autostash, Sync and force-with-lease. Sync fetches, pulls/merges when behind, then pushes when ahead.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::fetch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L230); [crates/orbit-pi/src/git.rs::push](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L166); [crates/orbit-pi/src/git.rs::pull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L205); [crates/orbit-pi/src/git.rs::merge_upstream](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L214)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W12.behavior [UNTESTED; ROW_SPECIFIC]** Fetch/publish/pull/sync/merge/rebase and force-with-lease using temporary clones and local bare remote.
   - Expected: Fetch, push/publish with upstream creation, ff-only pull, merge upstream, rebase upstream with autostash, Sync and force-with-lease. Sync fetches, pulls/merges when behind, then pushes when ahead.
2. **W12.edge [UNTESTED; ROW_SPECIFIC]** Create ahead, behind, diverged, rejected and stale-lease states; verify each recovery command and never force-push without lease.
   - Expected: Push with an existing upstream does not reinterpret auth/network failure as missing-upstream. Forced push never uses bare `--force` and requires UI confirmation. Rejected push triggers best-effort fetch to update counts and available recovery.
3. **W12.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Push with an existing upstream does not reinterpret auth/network failure as missing-upstream. Forced push never uses bare `--force` and requires UI confirmation. Rejected push triggers best-effort fetch to update counts and available recovery.
4. **W12.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W12.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W12.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W12.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W13 — Persistent copyable Git failure banner classifies push rejection/divergence/conflicts/dirty-overwrite/auth/network/no-upstream, secret scanning, file-size/LFS/protected-branch/hooks, missing remote/repository and unknown failures. Offers only matching recovery.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/failure.rs::FailureKind](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/failure.rs#L20); [crates/orbit-pi/src/git_panel/failure.rs::ActionError](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/failure.rs#L138); [crates/orbit-pi/src/git_panel/failure.rs::RecoveryAction](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/failure.rs#L105); [crates/orbit-pi/src/git_panel/failure.rs::classify](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/failure.rs#L184)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W13.behavior [UNTESTED; ROW_SPECIFIC]** Feed one fixture stderr for every classified Git failure; inspect banner, Copy and offered recovery action.
   - Expected: Persistent copyable Git failure banner classifies push rejection/divergence/conflicts/dirty-overwrite/auth/network/no-upstream, secret scanning, file-size/LFS/protected-branch/hooks, missing remote/repository and unknown failures. Offers only matching recovery.
2. **W13.edge [UNTESTED; ROW_SPECIFIC]** Trigger recovery after changing repository; preserve original diagnosis on refresh failure and never replay an already-successful mutation.
   - Expected: Full raw stderr survives until dismiss or later success; not a fleeting/truncated toast. Recovery may pull/merge/rebase/force-with-lease/continue/skip/abort/open conflict/publish/retry. Reauthenticate only copies `gh auth login`; it does not automatically run login. Server-rule failures have no fake repair button.
3. **W13.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Full raw stderr survives until dismiss or later success; not a fleeting/truncated toast. Recovery may pull/merge/rebase/force-with-lease/continue/skip/abort/open conflict/publish/retry. Reauthenticate only copies `gh auth login`; it does not automatically run login. Server-rule failures have no fake repair button.
4. **W13.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W13.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W13.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W13.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Additional source-specific procedures:** [../intentional-deviations.md](../intentional-deviations.md)

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W14 — Detect and display in-progress merge/rebase/cherry-pick/revert/bisect, conflict count and up to eight file links. Open conflicts in Files; continue once no conflicts; skip only rebase/cherry-pick; confirm abort.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_ops.rs::InProgress](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L18); [crates/orbit-pi/src/git_ops.rs::operation_state](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L64); [crates/orbit-pi/src/git_ops.rs::conflicted_files](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L86); [crates/orbit-pi/src/git_ops.rs::continue_operation](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L133)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W14.behavior [UNTESTED; ROW_SPECIFIC]** Create merge/rebase/cherry-pick/revert/bisect states in disposable repositories; open conflict files and continue/skip/abort.
   - Expected: Detect and display in-progress merge/rebase/cherry-pick/revert/bisect, conflict count and up to eight file links. Open conflicts in Files; continue once no conflicts; skip only rebase/cherry-pick; confirm abort.
2. **W14.edge [UNTESTED; ROW_SPECIFIC]** Leave unresolved conflicts, cancel abort and try unsupported skip; source gating and operation detection must match actual Git state.
   - Expected: Worktree-safe Git-dir resolution. `GIT_EDITOR`/`GIT_SEQUENCE_EDITOR=:` avoid interactive editor blocking. Bisect has no meaningful continue and returns an explicit error; abort is `bisect reset`. This is conflict assistance, not a three-way merge editor or new cherry-pick/revert/bisect UI.
3. **W14.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Worktree-safe Git-dir resolution. `GIT_EDITOR`/`GIT_SEQUENCE_EDITOR=:` avoid interactive editor blocking. Bisect has no meaningful continue and returns an explicit error; abort is `bisect reset`. This is conflict assistance, not a three-way merge editor or new cherry-pick/revert/bisect UI.
4. **W14.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W14.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W14.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W14.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W15 — Stashes tab lists newest first; create stash including untracked, pop, apply without removal and drop.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_ops.rs::StashEntry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L269); [crates/orbit-pi/src/git_ops.rs::list_stashes](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L277); [crates/orbit-pi/src/git_ops.rs::stash_push](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L296); [crates/orbit-pi/src/git_ops.rs::stash_pop](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_ops.rs#L320)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W15.behavior [UNTESTED; ROW_SPECIFIC]** Create stash with tracked/untracked changes, list, apply, pop and drop with confirmations where source requires.
   - Expected: Stashes tab lists newest first; create stash including untracked, pop, apply without removal and drop.
2. **W15.edge [UNTESTED; ROW_SPECIFIC]** Use empty stash list, conflicting pop and stale stash index after refresh; never apply/drop a different entry silently.
   - Expected: Git owns stash persistence. UI currently creates no custom message and drops without a separate PendingConfirm path. Backend supports optional stash message even though UI does not expose it.
3. **W15.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Git owns stash persistence. UI currently creates no custom message and drops without a separate PendingConfirm path. Backend supports optional stash message even though UI does not expose it.
4. **W15.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W15.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W15.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W15.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

