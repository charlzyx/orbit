# Git changes, history, graph and repository operations: W08–W11

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W08 — Changes uses grouped staged/unstaged Review sections with per-file Stage/Unstage and bulk Stage all/Unstage all; counts and action availability follow actual status.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::stage_paths](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L644); [crates/orbit-pi/src/git.rs::stage_all](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L653); [crates/orbit-pi/src/git.rs::unstage_paths](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L657); [crates/orbit-pi/src/git.rs::unstage_all](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L666)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W08.behavior [UNTESTED; ROW_SPECIFIC]** Stage one path, unstage one path, Stage all and Unstage all in staged/unstaged grouped Changes.
   - Expected: Changes uses grouped staged/unstaged Review sections with per-file Stage/Unstage and bulk Stage all/Unstage all; counts and action availability follow actual status.
2. **W08.edge [UNTESTED; ROW_SPECIFIC]** Exercise partly staged files, untracked and binary content; status counts and embedded Review must update from the actual index.
   - Expected: Stage uses path-separated Git arguments; stage-all includes untracked; unstage uses reset. File/section actions are rebuilt after status loads and cleared on workspace/clean-tree changes.
3. **W08.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Stage uses path-separated Git arguments; stage-all includes untracked; unstage uses reset. File/section actions are rebuilt after status loads and cleared on workspace/clean-tree changes.
4. **W08.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W08.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W08.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W08.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W09 — Discard individual unstaged path or all unstaged changes through explicit confirmation; staged changes are intended to remain.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::PendingConfirm](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L109); [crates/orbit-pi/src/git_panel/mod.rs::confirm_pending](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2770); [crates/orbit-pi/src/git_panel/mod.rs::discard](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2341); [crates/orbit-pi/src/git_panel/mod.rs::discard_all_changes](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2352)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W09.behavior [UNTESTED; ROW_SPECIFIC]** In a temporary repository commit a.txt=base, stage a.txt=staged and staged-new.txt=staged-new, then replace both worktree copies and add scratch/nested.txt. Confirm Discard All.
   - Expected: Worktree a.txt and staged-new.txt become their indexed bytes; git ls-files --stage -z is identical before/after. Untracked scratch content is removed. Repeat with one selected literal path and verify unrelated files are unchanged.
2. **W09.edge [UNTESTED; ROW_SPECIFIC]** Check staged blob remains unchanged and cancellation performs no mutation; failure banner must preserve the affected path.
   - Expected: Tracked files restore from index; untracked files are deleted by `git clean`. This is not OS Trash. Existing per-path implementation ignores both checkout/clean errors and returns success; bulk implementation ignores checkout error but propagates clean error. Preserve action semantics, do not mark failure handling verified.
3. **W09.error [UNTESTED; ROW_SPECIFIC]** Set Git executable to a missing fixture binary, then cancel a second discard before dispatch.
   - Expected: Strict source difference: Rust discard_paths ignores command errors and returns success. Go must retain its explicit error and unmodified fixture content per intentional deviation D02; do not mark exact source-equivalence PASS.
4. **W09.loading [UNTESTED; ROW_SPECIFIC]** Pause the discard subprocess after confirmation, then click the same action again and switch Git tabs.
   - Expected: One mutation starts. Busy state lasts until the real operation finishes; tab navigation cannot cancel/replay a mutation that already succeeded. Error remains visible on failure.
5. **W09.persistence [UNTESTED; ROW_SPECIFIC]** Before discard record HEAD, index entries, ignored/keep.txt and an outside temporary target reached through an untracked symlink; discard then recreate the page.
   - Expected: HEAD/index entries and ignored/outside target bytes remain intact. Reconstructed status shows only the actual remaining staged changes; no saved preference is needed for this operation.
6. **W09.keyboard [UNTESTED; ROW_SPECIFIC]** Open the discard confirmation by the file action; press Escape/Cancel; repeat and explicitly Confirm.
   - Expected: Cancellation invokes no discard callback and changes no file. Confirmation invokes once; global New Session/Settings/Terminal actions cannot mutate context behind the active confirmation.
7. **W09.repeat [UNTESTED; ROW_SPECIFIC]** Confirm Discard All twice after the first succeeds, then request discard for a path deleted externally.
   - Expected: Second discard is idempotent with respect to HEAD/index and remaining ignored files. Missing paths do not select a different path or broaden deletion.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- **PASS** `TestGitDiscardAllPreservesStagedIgnoredAndExternalTargets`: Discard All preserves index, ignored files and outside symlink targets; repeating remains safe. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestGitDiscardFailuresDoNotPretendSuccess`: Failed/cancelled discard reports error and retains bytes, intentional deviation D02. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Additional source-specific procedures:** [../intentional-deviations.md](../intentional-deviations.md)

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W10 — Commit bar has editable message, include-unstaged checkbox, Commit and Commit-and-push, blank-message generation, and state-driven Push/Publish/Pull/Merge/up-to-date actions.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::on_commit](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2404); [crates/orbit-pi/src/git_panel/mod.rs::run_git_action](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2589); [crates/orbit-pi/src/git_panel/mod.rs::commit_bar](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3905); [crates/orbit-pi/src/git_panel/mod.rs::bar_actions](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L7161)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W10.behavior [UNTESTED; ROW_SPECIFIC]** Edit commit message and include-unstaged toggle; Commit and Commit-and-push to a local bare remote.
   - Expected: Commit bar has editable message, include-unstaged checkbox, Commit and Commit-and-push, blank-message generation, and state-driven Push/Publish/Pull/Merge/up-to-date actions.
2. **W10.edge [UNTESTED; ROW_SPECIFIC]** Blank-message generation, no changes and rejected push must not create duplicate commits or falsely report successful push.
   - Expected: No staged work and include-unstaged off prevents commit. Generating/pending/operation flags gate keyboard commits too. Success clears message only for commit flows; failures retain it. Commit-and-push is sequential and can partially succeed.
3. **W10.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: No staged work and include-unstaged off prevents commit. Generating/pending/operation flags gate keyboard commits too. Success clears message only for commit flows; failures retain it. Commit-and-push is sequential and can partially succeed.
4. **W10.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W10.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W10.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W10.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W11 — Generate conventional commit message from staged diff, optional unstaged context, original files, branch/repo and recent repository/user commit style. Generate can ask Stage all versus use already-staged content; blank-message commit chains generation into commit.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/commit_message.rs::generate](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commit_message.rs#L41); [crates/orbit-pi/src/commit_message.rs::run_pi](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commit_message.rs#L423); [crates/orbit-pi/src/commit_message.rs::heuristic](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commit_message.rs#L72); [crates/orbit-pi/src/commit_message.rs::system_prompt](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commit_message.rs#L244)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W11.behavior [UNTESTED; ROW_SPECIFIC]** Generate commit text from staged changes with/without unstaged inclusion and repository/user history.
   - Expected: Generate conventional commit message from staged diff, optional unstaged context, original files, branch/repo and recent repository/user commit style. Generate can ask Stage all versus use already-staged content; blank-message commit chains generation into commit.
2. **W11.edge [UNTESTED; ROW_SPECIFIC]** Return empty/malformed generator output, cancellation and absent Pi; verify staged-only choice never silently stages unrelated files.
   - Expected: Existing auxiliary path is `pi -p --no-tools --no-session --no-skills --no-context-files --no-approve`, NOT RPC, with selected provider/model, 180 s timeout and heuristic fallback. Extensions remain enabled for custom providers. Bounded diff/original context; result stays editable.
3. **W11.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Existing auxiliary path is `pi -p --no-tools --no-session --no-skills --no-context-files --no-approve`, NOT RPC, with selected provider/model, 180 s timeout and heuristic fallback. Extensions remain enabled for custom providers. Bounded diff/original context; result stays editable.
4. **W11.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W11.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W11.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W11.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Tool-free Pi RPC generation and repository templates are implemented; fake-process race tests verify isolation, prompt bounds, fallback, cancellation and no staging mutation. Host workflow/live inference and exact template UI acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

