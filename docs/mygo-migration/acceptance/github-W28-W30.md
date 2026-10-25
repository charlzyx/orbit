# GitHub issues and pull requests: W28–W30

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W28 — New PR form has title/body/notes, draft toggle, base-branch picker, template selection, generation, submit/cancel. Base picker is anchored in form; default prefers main/master then available branch.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::open_pull_new](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1332); [crates/orbit-pi/src/git_panel/mod.rs::pr_new_view](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6279); [crates/orbit-pi/src/git_panel/mod.rs::base_branch_selector](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1618); [crates/orbit-pi/src/git_panel/mod.rs::pr_base_popup](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1666)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W28.behavior [UNTESTED; ROW_SPECIFIC]** Open New PR with main/master, only-current, no-refs and alternate-local-branch fixtures; open its anchored base chip, type a trimmed mixed-case query, choose a local row, enable Draft and submit through fake gh.
   - Expected: main wins, then master; otherwise the first local branch not equal to current wins, or main if none. The 320px above-chip picker shows local refs only in stable lowercase order. Pointer selection changes base, closes picker, and sends the selected --base plus --draft only after Create.
2. **W28.edge [UNTESTED; ROW_SPECIFIC]** Use no main/master, blank title, failed submission and changing base picker; preserve form state and correct default branch.
   - Expected: Required title; selected template labels passed to gh. Creation returns URL, opens PR detail, clears fields. No automatic publish/push is performed by this controller before `gh pr create`.
3. **W28.error [UNTESTED; ROW_SPECIFIC]** Fail gh pr create while preserving the open form, correct the fake failure, then retry.
   - Expected: Failure does not close or clear the form. Successful retry clears/closes once; no second request is dispatched by repeated busy clicks.
4. **W28.loading [UNTESTED; ROW_SPECIFIC]** Hold PR creation, attempt duplicate native Create and inspect draft/form before releasing.
   - Expected: Source pr_busy keeps form visible and disables duplicate submit. Full modal-vs-inline visual difference remains unresolved in Go.
5. **W28.persistence [UNTESTED; ROW_SPECIFIC]** Cancel/reopen New PR in the same instance; reconstruct panel after full app restart.
   - Expected: Draft prose is in memory, not persisted in a new config file. On reopen, source chooses default base anew while existing draft toggle remains its stored field.
6. **W28.keyboard [UNTESTED; ROW_SPECIFIC]** Focus the branch filter and type Unicode text; send Up, Down and Enter, including zero results, without clicking a branch. Separately inspect Escape dispatch on the original GPUI host.
   - Expected: GitPanel has no PickerConfirm/Prev/Next listener despite the Picker key context, so these actions must not select a branch or create a PR. Native Go containment is tested; original GPUI Escape propagation remains an explicit open gate, not assumed from MyGo PopoverBase.
7. **W28.repeat [UNTESTED; ROW_SPECIFIC]** Run cancel/reopen, failed create/retry and Issue↔PR navigation three times, using different text for each form.
   - Expected: No cross-form draft replacement or duplicate creation. Template-on-open behavior remains source-specific rather than indiscriminate draft erasure.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- **PASS** `TestGitHubNativeCreateFailureKeepsFormAndRetryClearsOnce`: Actual native Issue and PR Create stay open during/error, retain drafts, reject duplicate busy dispatch, trim the title and clear only after successful retry. [Evidence](github-evidence.json). Tested dirty-tree manifest `c80aaa612b72d9a162821cc26a51bd0d5997a2954b2f974bb6d7307203f1f8e3`, not a published commit.
- **PASS** `TestGitHubSourceDefaultBase`: Eight fixtures match unchanged Rust default_base, including main/master precedence, skip-current local fallback and empty fallback. [Evidence](github-state-evidence.json). Tested dirty-tree manifest `07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`, not a published commit.
- **PASS** `TestGitHubBaseFilterMatchesRustUnicodeAndStableSort`: Local refs use trimmed full Unicode lowercase substring filtering and stable lowercase sort, including Greek final sigma and dotted capital I. [Evidence](github-state-evidence.json). Tested dirty-tree manifest `07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`, not a published commit.
- **PASS** `TestGitHubNativeBasePickerFiltersLocalBranchesWithoutImplicitSelection`: Native anchored search excludes tags/remotes, focuses the query, sorts rows and selects only on pointer click; arrow/Enter do not mutate or submit. [Evidence](github-state-evidence.json). Tested dirty-tree manifest `07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`, not a published commit.
- **PASS** `TestGitHubBasePickerEmptyDismissRepeatAndErrorDoNotMutateDraft`: Native 320px above-chip panel, no-match Enter, retained draft during injected loading/error, outside dismissal and 200ms reopening guard; Escape containment is tested but source GPUI dispatch remains unverified. [Evidence](github-state-evidence.json). Tested dirty-tree manifest `07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`, not a published commit.
- **PASS** `TestGitHubNativeSelectedBaseSurvivesCreateFailureAndExactArgvRetry`: Pointer-selected Unicode base and draft flag survive actual fake-gh create failure; exact base/title/body/draft argv are repeated only on explicit retry. [Evidence](github-state-evidence.json). Tested dirty-tree manifest `07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`, not a published commit.

**Additional source-specific procedures:** [../github-acceptance.md](../github-acceptance.md), [github-evidence.json](github-evidence.json), [../github-base-picker-acceptance.md](../github-base-picker-acceptance.md), [github-state-evidence.json](github-state-evidence.json)

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W29 — AI issue/PR draft uses optional user notes, branch commits/diff and chosen template; preserves template prefix/body structure.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/issue_message.rs::generate](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/issue_message.rs#L66); [crates/orbit-pi/src/issue_message.rs::collect](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/issue_message.rs#L210); [crates/orbit-pi/src/issue_message.rs::run_pi](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/issue_message.rs#L528); [crates/orbit-pi/src/issue_message.rs::heuristic](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/issue_message.rs#L192)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W29.behavior [UNTESTED; ROW_SPECIFIC]** Generate issue/PR draft with notes, branch diff/commits and selected template.
   - Expected: AI issue/PR draft uses optional user notes, branch commits/diff and chosen template; preserves template prefix/body structure.
2. **W29.edge [UNTESTED; ROW_SPECIFIC]** Return truncated/invalid generation and cancel; preserve template prefix/sections and do not overwrite later manual edits.
   - Expected: Isolated tool-free `pi -p` subprocess, 180 s timeout, 24 KiB context and max 30 commits, same selected model/provider, heuristic fallback. No active chat-session mutation. A strict “all pi calls over RPC” conversion must include this path and W11.
3. **W29.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Isolated tool-free `pi -p` subprocess, 180 s timeout, 24 KiB context and max 30 commits, same selected model/provider, heuristic fallback. No active chat-session mutation. A strict “all pi calls over RPC” conversion must include this path and W11.
4. **W29.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W29.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W29.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W29.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Additional source-specific procedures:** [../github-acceptance.md](../github-acceptance.md), [github-evidence.json](github-evidence.json), [../github-base-picker-acceptance.md](../github-base-picker-acceptance.md), [github-state-evidence.json](github-state-evidence.json)

**Remaining:** Tool-free Pi RPC generation and repository templates are implemented; fake-process race tests verify isolation, prompt bounds, fallback, cancellation and no staging mutation. Host workflow/live inference and exact template UI acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W30 — Review can open from toolbar diffstat, transcript per-turn summary, command/palette, Git file callback; select desired file after load. Separate embedded instance supplies Changes overview.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/sidepane.rs::show_uncommitted](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L539); [crates/orbit-pi/src/sidepane.rs::show_file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L555); [crates/orbit-pi/src/sidepane.rs::show_review](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L571); [crates/orbit-pi/src/sidepane.rs::show_review_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L816)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W30.behavior [UNTESTED; ROW_SPECIFIC]** Open Review from toolbar, per-turn summary, palette and Git file/commit callbacks.
   - Expected: Review can open from toolbar diffstat, transcript per-turn summary, command/palette, Git file callback; select desired file after load. Separate embedded instance supplies Changes overview.
2. **W30.edge [UNTESTED; ROW_SPECIFIC]** Delay collection, switch source/workspace/session and select another file; display only latest intended source/ref/path.
   - Expected: Full-page Git embedded mode suppresses dock resize/full-width/close chrome and source popup, groups Staged and Changes, starts files collapsed; dock starts expanded.
3. **W30.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Full-page Git embedded mode suppresses dock resize/full-width/close chrome and source popup, groups Staged and Changes, starts files collapsed; dock starts expanded.
4. **W30.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W30.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W30.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W30.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Additional source-specific procedures:** [../github-acceptance.md](../github-acceptance.md), [github-evidence.json](github-evidence.json), [../github-base-picker-acceptance.md](../github-base-picker-acceptance.md), [github-state-evidence.json](github-state-evidence.json)

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

