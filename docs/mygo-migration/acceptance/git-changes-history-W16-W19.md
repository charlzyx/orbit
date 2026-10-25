# Git changes, history, graph and repository operations: W16–W19

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W16 — History shows current branch commit list, optional all-refs filter, author filtering by author interaction, path/file history with follow-renames, removable filter chips and Load more.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::HistoryFilter](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L955); [crates/orbit-pi/src/git.rs::history_filtered](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L979); [crates/orbit-pi/src/git.rs::history_by_author](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L799); [crates/orbit-pi/src/git_panel/mod.rs::refresh_history](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2155)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W16.behavior [UNTESTED; ROW_SPECIFIC]** Load branch/all-refs history, filter author/path with rename following, remove chips and Load more.
   - Expected: History shows current branch commit list, optional all-refs filter, author filtering by author interaction, path/file history with follow-renames, removable filter chips and Load more.
2. **W16.edge [UNTESTED; ROW_SPECIFIC]** Use equal timestamps, merge commits, empty filters and external branch changes; avoid duplicate or skipped commits.
   - Expected: `HISTORY_PAGE=40`; refresh expands requested window (initial request uses max(existing,40)+40). Watch-driven status refresh intentionally avoids growing it. Backend also supports grep, but no History grep control was located.
3. **W16.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `HISTORY_PAGE=40`; refresh expands requested window (initial request uses max(existing,40)+40). Watch-driven status refresh intentionally avoids growing it. Backend also supports grep, but no History grep control was located.
4. **W16.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W16.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W16.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W16.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W17 — Expand/collapse commit detail with subject/body/hash/author/committer/parents/refs and changed-file statistics; copy/open external commit links. File menu opens current file, current uncommitted diff, OS reveal, path copy or follow-renames history.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::CommitDetail](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L850); [crates/orbit-pi/src/git.rs::commit_detail](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L867); [crates/orbit-pi/src/git.rs::commit_files](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L899); [crates/orbit-pi/src/git.rs::RemoteWeb](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L1238)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W17.behavior [UNTESTED; ROW_SPECIFIC]** Expand commit details, inspect parents/ref/file stats and use copy/open-file/open-diff/reveal actions.
   - Expected: Expand/collapse commit detail with subject/body/hash/author/committer/parents/refs and changed-file statistics; copy/open external commit links. File menu opens current file, current uncommitted diff, OS reveal, path copy or follow-renames history.
2. **W17.edge [UNTESTED; ROW_SPECIFIC]** Select a root/merge commit or deleted file; selected commit diff must not accidentally show HEAD/worktree content.
   - Expected: File details capped at 300 rendered rows with remainder summary. `file_actions_menu` receives a hash but currently ignores it for Open diff, so this is NOT a historical per-commit diff viewer. Remote URLs cover GitHub/GitLab/Bitbucket/known GitHub/GitLab/Bitbucket hosts; URL syntax variants.
3. **W17.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: File details capped at 300 rendered rows with remainder summary. `file_actions_menu` receives a hash but currently ignores it for Open diff, so this is NOT a historical per-commit diff viewer. Remote URLs cover GitHub/GitLab/Bitbucket/known GitHub/GitLab/Bitbucket hosts; URL syntax variants.
4. **W17.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W17.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W17.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W17.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W18 — Graph draws branch lanes, merge edges, refs and commit metadata; all-refs option includes remote-only commits and Load more deepens history.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git.rs::GraphCommit](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L1051); [crates/orbit-pi/src/git.rs::GraphRow](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L1117); [crates/orbit-pi/src/git.rs::graph_history](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L1066); [crates/orbit-pi/src/git.rs::layout_graph](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git.rs#L1142)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W18.behavior [UNTESTED; ROW_SPECIFIC]** Load a repository with branches, merges and remote-only commits; toggle all refs and deepen graph history.
   - Expected: Graph draws branch lanes, merge edges, refs and commit metadata; all-refs option includes remote-only commits and Load more deepens history.
2. **W18.edge [UNTESTED; ROW_SPECIFIC]** Compare lane/edge relationships and ref labels after pagination, narrow viewport and checkout; do not infer graph parity from a flat log.
   - Expected: Layout tolerates parent commits outside current window and octopus merges. Graph has independent loading/error states.
3. **W18.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Layout tolerates parent commits outside current window and octopus merges. Graph has independent loading/error states.
4. **W18.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W18.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W18.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W18.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Git/temp-repo and fake-gh tests plus scoped six-tab/native controller tests pass; live authenticated GitHub and exact UI interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W19 — Git author avatars resolve from email-to-login override, GitHub noreply address, then handle-shaped name; fallback is stable colored initials. GitHub signed-in account chip shows current gh identity.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/git_panel/mod.rs::github_avatar_url](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L8510); [crates/orbit-pi/src/git_panel/mod.rs::resolve_github_login](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L8518); [crates/orbit-pi/src/git_panel/mod.rs::avatar_login_overrides](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L8535); [crates/orbit-pi/src/git_panel/mod.rs::monogram_chip](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L8589)

**Fixture:** Temporary repositories plus local bare remote/clones, fixture users and conflict branches; no external remote writes.

1. **W19.behavior [UNTESTED; ROW_SPECIFIC]** Resolve fixture avatar identities via override, GitHub noreply and name fallback; inspect signed-in gh account.
   - Expected: Git author avatars resolve from email-to-login override, GitHub noreply address, then handle-shaped name; fallback is stable colored initials. GitHub signed-in account chip shows current gh identity.
2. **W19.edge [UNTESTED; ROW_SPECIFIC]** Fail avatar download, use colliding/invalid identities and switch theme; stable initials remain and unrelated account is not shown.
   - Expected: `~/.orbit-pi/git-avatars.json` read once; malformed map/invalid handles ignored. Remote avatar loading/failure falls back to monogram, not wrong arbitrary user.
3. **W19.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Return a controlled nonzero Git exit, cancellation or stale completion; retry only the operation still outstanding.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `~/.orbit-pi/git-avatars.json` read once; malformed map/invalid handles ignored. Remote avatar loading/failure falls back to monogram, not wrong arbitrary user.
4. **W19.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W19.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W19.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W19.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/github_templates.go`, `internal/gitworkbench/templates.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/git_test.go`, `internal/workspace/github_test.go`, `internal/gitworkbench/state_test.go`, `internal/workspace/generation_test.go`, `internal/workspace/process_cancel_unix_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Avatar email/login overrides, async cache/photo/fallback and author filters are implemented; live external image behavior, visual geometry and full interactions remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

