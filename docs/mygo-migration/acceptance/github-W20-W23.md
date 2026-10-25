# GitHub issues and pull requests: W20–W23

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W20 — GitHub integration probes gh availability/auth/version and shows honest setup states on Issues/Pulls when absent or signed out. Refresh retries probe; CLI path can be overridden by `GH_BIN`.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/gh.rs::GH_BIN_ENV](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L20); [crates/orbit-pi/src/gh.rs::MIN_GH_VERSION](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L23); [crates/orbit-pi/src/gh.rs::status](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L56); [crates/orbit-pi/src/gh.rs::binary](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L93)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W20.behavior [UNTESTED; ROW_SPECIFIC]** Open Issues/Pulls with gh absent, too old, unauthenticated and authenticated; refresh after state changes.
   - Expected: GitHub integration probes gh availability/auth/version and shows honest setup states on Issues/Pulls when absent or signed out. Refresh retries probe; CLI path can be overridden by `GH_BIN`.
2. **W20.edge [UNTESTED; ROW_SPECIFIC]** Set GH_BIN to a controlled executable and fail probe; display exact setup state without performing external login in tests.
   - Expected: Known minimum is 2.20.0; old version produces detail warning, not an unsupported-capability success claim. Prompts/update notifier/pager/color disabled, console hidden on Windows. Signed-in current user fetched only after auth succeeds.
3. **W20.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Known minimum is 2.20.0; old version produces detail warning, not an unsupported-capability success claim. Prompts/update notifier/pager/color disabled, console hidden on Windows. Signed-in current user fetched only after auth succeeds.
4. **W20.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W20.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W20.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W20.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W21 — Issue list has Open/Closed/All, search, label filter, clear filters, status/author/label/time/comment metadata and detail selection.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/gh.rs::IssueFilter](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L246); [crates/orbit-pi/src/gh.rs::list_issues](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L268); [crates/orbit-pi/src/gh.rs::list_labels](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L389); [crates/orbit-pi/src/git_panel/mod.rs::refresh_issues](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1068)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W21.behavior [UNTESTED; ROW_SPECIFIC]** Filter issue list by state, text and labels; clear filters; open an issue and inspect metadata.
   - Expected: Issue list has Open/Closed/All, search, label filter, clear filters, status/author/label/time/comment metadata and detail selection.
2. **W21.edge [UNTESTED; ROW_SPECIFIC]** Return empty, delayed and failing responses; change repository/filter before completion and reject old results.
   - Expected: gh fetch limit 50, label catalog limit 100; empty/loading/error displayed. Backend assignee filter exists but no assignee picker found in list UI. No continuous issue monitoring or list pagination beyond this fetched window.
3. **W21.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: gh fetch limit 50, label catalog limit 100; empty/loading/error displayed. Backend assignee filter exists but no assignee picker found in list UI. No continuous issue monitoring or list pagination beyond this fetched window.
4. **W21.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W21.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W21.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W21.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W22 — Issue detail renders Markdown body/comments and metadata; send comment, toggle labels, close/reopen, open on GitHub, return to list.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::open_issue](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1102); [crates/orbit-pi/src/git_panel/mod.rs::issue_detail_view](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L5001); [crates/orbit-pi/src/git_panel/mod.rs::comment_issue](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1156); [crates/orbit-pi/src/git_panel/mod.rs::toggle_issue_state](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1169)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W22.behavior [UNTESTED; ROW_SPECIFIC]** Open issue body/comments; post a fixture comment, toggle label and close/reopen using a fake gh receiver.
   - Expected: Issue detail renders Markdown body/comments and metadata; send comment, toggle labels, close/reopen, open on GitHub, return to list.
2. **W22.edge [UNTESTED; ROW_SPECIFIC]** Cancel or fail a mutation, refresh and retry; preserve typed text, exact issue identity and no duplicate comment.
   - Expected: Mutation busy state, success/failure banner, list/detail reload from GitHub. Existing comment handler clears composer before network result; failure restoration is not implemented there. Close reason backend supported but UI uses default.
3. **W22.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Mutation busy state, success/failure banner, list/detail reload from GitHub. Existing comment handler clears composer before network result; failure restoration is not implemented there. Close reason backend supported but UI uses default.
4. **W22.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W22.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W22.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W22.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W23 — New issue form has title/body/notes, repository template selector and generation, submit, cancel/back. Requires nonblank title; selected template labels pass to create. Successful creation clears fields, reloads list and opens resulting issue number.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::open_issue_new](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1318); [crates/orbit-pi/src/git_panel/mod.rs::issue_new_view](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L5319); [crates/orbit-pi/src/git_panel/mod.rs::create_issue](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1202); [crates/orbit-pi/src/git_panel/mod.rs::generate_issue_draft](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1773)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W23.behavior [UNTESTED; ROW_SPECIFIC]** Start New Issue, choose template, edit title/body/notes, generate, submit to fixture gh and cancel.
   - Expected: New issue form has title/body/notes, repository template selector and generation, submit, cancel/back. Requires nonblank title; selected template labels pass to create. Successful creation clears fields, reloads list and opens resulting issue number.
2. **W23.edge [UNTESTED; ROW_SPECIFIC]** Reject blank title, preserve template labels and failed submission text; success clears form and opens correct created issue.
   - Expected: Writes via `gh issue create`; no auto-send merely on generation. Generic backend can target explicit repo, assignees and attachment paths, but ordinary Git-page create passes current repo and empty assignee/attachment arrays.
3. **W23.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Writes via `gh issue create`; no auto-send merely on generation. Generic backend can target explicit repo, assignees and attachment paths, but ordinary Git-page create passes current repo and empty assignee/attachment arrays.
4. **W23.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W23.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W23.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W23.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

