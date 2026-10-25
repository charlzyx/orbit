# GitHub issues and pull requests: W24–W27

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W24 — Repository issue/PR templates follow `.github/*_TEMPLATE/` directory then legacy single-file lookup. Markdown frontmatter supplies name/about/title/labels; YAML issue forms flatten to Markdown; config.yml excluded.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/gh_templates.rs::issue_templates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh_templates.rs#L72); [crates/orbit-pi/src/gh_templates.rs::pull_templates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh_templates.rs#L83); [crates/orbit-pi/src/gh_templates.rs::parse_markdown](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh_templates.rs#L170); [crates/orbit-pi/src/gh_templates.rs::parse_issue_form](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh_templates.rs#L277)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W24.behavior [UNTESTED; ROW_SPECIFIC]** Discover directory and legacy issue/PR templates; apply Markdown frontmatter and YAML issue form.
   - Expected: Repository issue/PR templates follow `.github/*_TEMPLATE/` directory then legacy single-file lookup. Markdown frontmatter supplies name/about/title/labels; YAML issue forms flatten to Markdown; config.yml excluded.
2. **W24.edge [UNTESTED; ROW_SPECIFIC]** Exclude config.yml, reject malformed frontmatter safely and preserve expected title/labels/body precedence across template changes.
   - Expected: Templates capped at 256 KiB. Multiple templates selectable; one template can apply automatically only to untouched body, including late async arrival. Template removal does not blindly erase edited bodies.
3. **W24.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Templates capped at 256 KiB. Multiple templates selectable; one template can apply automatically only to untouched body, including late async arrival. Template removal does not blindly erase edited bodies.
4. **W24.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W24.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W24.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W24.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Tool-free Pi RPC generation and repository templates are implemented; fake-process race tests verify isolation, prompt bounds, fallback, cancellation and no staging mutation. Host workflow/live inference and exact template UI acceptance remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W25 — Pull request list has Open/Closed/Merged/All, search, clear filters, draft/author/labels/ref/diffstat/review/check summary and detail selection.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/gh.rs::PrFilter](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L665); [crates/orbit-pi/src/gh.rs::list_pulls](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L720); [crates/orbit-pi/src/gh.rs::GhPull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L562); [crates/orbit-pi/src/gh.rs::checks_summary](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L618)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W25.behavior [UNTESTED; ROW_SPECIFIC]** Filter PR list by Open/Closed/Merged/All and search; inspect refs/draft/reviews/check summaries.
   - Expected: Pull request list has Open/Closed/Merged/All, search, clear filters, draft/author/labels/ref/diffstat/review/check summary and detail selection.
2. **W25.edge [UNTESTED; ROW_SPECIFIC]** Return absent check fields, no results and stale responses; clearing filters restores actual repository list.
   - Expected: gh limit 50. Backend base/author filter fields exist; exposed list UI is state/search, while base selection belongs to New PR. Loading/error/empty states separate.
3. **W25.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: gh limit 50. Backend base/author filter fields exist; exposed list UI is state/search, while base selection belongs to New PR. Loading/error/empty states separate.
4. **W25.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W25.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W25.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W25.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W26 — PR detail renders Markdown description, comments, reviews, commits, changed files, CI checks with links, draft/ref/state metadata and additions/deletions.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::pr_detail_view](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L5686); [crates/orbit-pi/src/gh.rs::GhPull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L562); [crates/orbit-pi/src/gh.rs::GhCheck](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L454); [crates/orbit-pi/src/gh.rs::GhReview](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/gh.rs#L523)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W26.behavior [UNTESTED; ROW_SPECIFIC]** Open a PR with body/comments/reviews/commits/files/checks and follow fixture links.
   - Expected: PR detail renders Markdown description, comments, reviews, commits, changed files, CI checks with links, draft/ref/state metadata and additions/deletions.
2. **W26.edge [UNTESTED; ROW_SPECIFIC]** Use pending/failed/unknown checks and unavailable file details; represent unknown state rather than a false pass.
   - Expected: File rendering capped at 300; unknown/missing check data is not treated as a passing check. Checks combine failure first, then pending, passed, skipped. No native full PR patch review or inline line-comment editor located.
3. **W26.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: File rendering capped at 300; unknown/missing check data is not treated as a passing check. Checks combine failure first, then pending, passed, skipped. No native full PR patch review or inline line-comment editor located.
4. **W26.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W26.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W26.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W26.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W27 — PR comment, approve/request changes, close/reopen, checkout, merge by merge/squash/rebase and optional delete-head-branch. Merge requires confirmation showing method.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::comment_pull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1936); [crates/orbit-pi/src/git_panel/mod.rs::review_pull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1949); [crates/orbit-pi/src/git_panel/mod.rs::toggle_pull_state](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1975); [crates/orbit-pi/src/git_panel/mod.rs::checkout_pull](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1993)

**Fixture:** Temporary repository with fixture templates and fake gh executable recording argv/JSON; no real GitHub writes/auth.

1. **W27.behavior [UNTESTED; ROW_SPECIFIC]** Comment, approve/request changes, close/reopen and checkout through fixture gh; request merge using each method.
   - Expected: PR comment, approve/request changes, close/reopen, checkout, merge by merge/squash/rebase and optional delete-head-branch. Merge requires confirmation showing method.
2. **W27.edge [UNTESTED; ROW_SPECIFIC]** Cancel confirmation and repeat clicks; confirm dialog must name method and branch deletion choice, with one resulting mutation.
   - Expected: Checkout refuses dirty worktree or in-progress Git operation. Mutation busy states; list/detail re-read. Existing comment/review input is cleared before success. gh enforces repository-side merge rules; UI is not a CI-bypass implementation.
3. **W27.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return nonzero/malformed/empty fixture gh results and a late response after filter/repository change.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Checkout refuses dirty worktree or in-progress Git operation. Mutation busy states; list/detail re-read. Existing comment/review input is cleared before success. gh enforces repository-side merge rules; UI is not a CI-bypass implementation.
4. **W27.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W27.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W27.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W27.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

