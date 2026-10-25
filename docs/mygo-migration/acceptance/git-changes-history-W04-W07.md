# Git changes, history, graph and repository operations: W04–W07

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W04 — Git has six tabs: Changes, History, Graph, Stashes, Issues and Pull requests. Header/back, branch row, refresh, contextual busy/empty/error states and shortcuts invoke controllers.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::GitTab](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L85); [crates/orbit-pi/src/git_panel/mod.rs::set_tab](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L589); [crates/orbit-pi/src/git_panel/mod.rs::show](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L568); [crates/orbit-pi/src/git_panel/mod.rs::hide](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L576)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W04.behavior [UNTESTED; ROW_SPECIFIC]** Visit all six Git tabs via clicks and Primary+Alt+1…6; trigger refresh and back navigation.
   - Expected: Git has six tabs: Changes, History, Graph, Stashes, Issues and Pull requests. Header/back, branch row, refresh, contextual busy/empty/error states and shortcuts invoke controllers.
2. **W04.edge [UNTESTED; ROW_SPECIFIC]** Use a non-repository, unborn repository and delayed command; tab-local loading/error/empty state must be accurate.
   - Expected: `show` wakes embedded Review; `hide` parks it. Workspace changes clear relevant cached lists/details/filters/templates. Manual refresh re-probes gh and spins for at least 650 ms. Watch refresh does not accidentally grow History/Graph pagination.
3. **W04.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `show` wakes embedded Review; `hide` parks it. Workspace changes clear relevant cached lists/details/filters/templates. Manual refresh re-probes gh and spins for at least 650 ms. Watch refresh does not accidentally grow History/Graph pagination.
4. **W04.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W04.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W04.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W04.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W05 — Git branch/status discovery includes staged and unstaged status columns, original rename path, per-side additions/deletions, untracked counts, ahead/behind and unborn/detached conditions.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L41); [crates/orbit-pi/src/git.rs::current_branch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L57); [crates/orbit-pi/src/git.rs::status_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L541); [crates/orbit-pi/src/git.rs::StatusRow](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L490)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W05.behavior [UNTESTED; ROW_SPECIFIC]** Create staged/unstaged/renamed/untracked files with known line counts; inspect branch and ahead/behind status.
   - Expected: Git branch/status discovery includes staged and unstaged status columns, original rename path, per-side additions/deletions, untracked counts, ahead/behind and unborn/detached conditions.
2. **W05.edge [UNTESTED; ROW_SPECIFIC]** Use unborn HEAD, detached HEAD, missing upstream and binary files; do not fabricate branch or line counts.
   - Expected: Git CLI, not libgit2. `GIT_OPTIONAL_LOCKS=0` avoids read-triggered index writes; untracked binary or >2 MiB files report zero counted lines. Status collection failure paints unavailable state.
3. **W05.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Git CLI, not libgit2. `GIT_OPTIONAL_LOCKS=0` avoids read-triggered index writes; untracked binary or >2 MiB files report zero counted lines. Status collection failure paints unavailable state.
4. **W05.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W05.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W05.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W05.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W06 — Branch picker searches and groups current/recent/local/remote refs; checkout local or create/check out a local tracking branch for remote. Recent selection is derived from HEAD reflog.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::list_branches](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L65); [crates/orbit-pi/src/git.rs::recent_branches](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L92); [crates/orbit-pi/src/git.rs::checkout_branch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L83); [crates/orbit-pi/src/git.rs::checkout_remote_branch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L118)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W06.behavior [UNTESTED; ROW_SPECIFIC]** Search current/recent/local/remote refs; check out local and a remote-only ref in temporary repositories.
   - Expected: Branch picker searches and groups current/recent/local/remote refs; checkout local or create/check out a local tracking branch for remote. Recent selection is derived from HEAD reflog.
2. **W06.edge [UNTESTED; ROW_SPECIFIC]** Use conflicting worktree branch, dirty overwrite and missing remote ref; retain actionable error and current branch truth.
   - Expected: `switch` falls back to `checkout`; refs are read from Git. Error banners show actual failure; branch operation busy state, menu focus and 200 ms outside-dismiss guards prevent accidental reopening.
3. **W06.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `switch` falls back to `checkout`; refs are read from Git. Error banners show actual failure; branch operation busy state, menu focus and 200 ms outside-dismiss guards prevent accidental reopening.
4. **W06.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W06.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W06.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W06.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W07 — New branch, rename current branch, safe branch delete; branch merge/rebase ref picker includes local/remote/tag refs. Merge offers normal fast-forward-permitted, no-ff and squash modes. Dirty rebase uses autostash.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::open_branch_prompt](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2827); [crates/orbit-pi/src/git_panel/mod.rs::confirm_branch_prompt](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2850); [crates/orbit-pi/src/git_panel/mod.rs::delete_branch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2876); [crates/orbit-pi/src/git_panel/mod.rs::merge_ref](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2799)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W07.behavior [UNTESTED; ROW_SPECIFIC]** Create/rename/delete disposable branches; merge normal/no-ff/squash and rebase with dirty tracked content.
   - Expected: New branch, rename current branch, safe branch delete; branch merge/rebase ref picker includes local/remote/tag refs. Merge offers normal fast-forward-permitted, no-ff and squash modes. Dirty rebase uses autostash.
2. **W07.edge [UNTESTED; ROW_SPECIFIC]** Cancel deletion/abort operations; invalid refs and unmerged delete must fail safely; branch picker includes source ref groups.
   - Expected: Safe delete `-d` first; “not fully merged” triggers explicit force-delete `-D` confirmation. Git validates/checks operations; UI closes menus before dispatch.
3. **W07.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Safe delete `-d` first; “not fully merged” triggers explicit force-delete `-D` confirmation. Git validates/checks operations; UI closes menus before dispatch.
4. **W07.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W07.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W07.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W07.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

