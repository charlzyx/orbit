# Worktrees and setup lifecycle: W40–W43

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W40 — Worktrees page lists Git-discovered main/linked/external worktrees with independent name/path/branch, active/bare/detached/locked/prunable markers and reasons.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::Worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L141); [crates/orbit-pi/src/worktree.rs::parse_worktree_list](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L504); [crates/orbit-pi/src/worktree.rs::WorktreeManager::repository_root](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L682); [crates/orbit-pi/src/worktree.rs::WorktreeManager::list](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L739)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W40.behavior [UNTESTED; ROW_SPECIFIC]** List main, linked and externally created worktrees; inspect path/branch/active/locked/prunable markers.
   - Expected: Worktrees page lists Git-discovered main/linked/external worktrees with independent name/path/branch, active/bare/detached/locked/prunable markers and reasons.
2. **W40.edge [UNTESTED; ROW_SPECIFIC]** Use bare/detached and missing paths; refresh after external changes without fabricating membership or current selection.
   - Expected: `git worktree list --porcelain -z`, older-Git fallback; main root found from Git even when opened in linked worktree. Non-repo is actionable error; canonical/lexical paths prevent alias confusion.
3. **W40.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `git worktree list --porcelain -z`, older-Git fallback; main root found from Git even when opened in linked worktree. Non-repo is actionable error; canonical/lexical paths prevent alias confusion.
4. **W40.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W40.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W40.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W40.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W41 — Create worktree from existing local/remote branch or a new branch with optional start point; separate editable display/directory name and branch name, auto-derived new branch until user edits it; quick-create suggests unused branch slug plus suffix.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/worktrees.rs::open_create_worktree_dialog](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L197); [crates/orbit-pi/src/app/worktrees.rs::submit_create_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L323); [crates/orbit-pi/src/app/worktrees.rs::perform_worktree_create](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L393); [crates/orbit-pi/src/app/worktrees.rs::open_new_worktree_from_work_in](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2653)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W41.behavior [UNTESTED; ROW_SPECIFIC]** Create from local/remote existing branch and a new branch/start point; edit display name and branch independently.
   - Expected: Create worktree from existing local/remote branch or a new branch with optional start point; separate editable display/directory name and branch name, auto-derived new branch until user edits it; quick-create suggests unused branch slug plus suffix.
2. **W41.edge [UNTESTED; ROW_SPECIFIC]** Auto-derive branch until manual edit, use quick-create unique names and cancel; no pre-confirmation worktree creation.
   - Expected: New target under configured root; branch refs validated by Git. Branch choices async; default avoids active branch where possible. Root created when absent. Existing checked-out branch produces dialog to open its existing worktree.
3. **W41.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: New target under configured root; branch refs validated by Git. Branch choices async; default avoids active branch where possible. Root created when absent. Existing checked-out branch produces dialog to open its existing worktree.
4. **W41.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W41.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W41.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W41.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestNativeCreateNameTypingDerivesBranchUntilManualEdit`: Typing the worktree name derives branch only until the user edits branch explicitly. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W42 — Worktree validation rejects empty/dot/parent/separator names, root escape, existing paths, missing repo and branch conflict. Creation errors keep actionable context rather than pretending success.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::WorktreeError](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L45); [crates/orbit-pi/src/worktree.rs::validate_name](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L328); [crates/orbit-pi/src/worktree.rs::normalize_lexically](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L393); [crates/orbit-pi/src/worktree.rs::canonical_or](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L416)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W42.behavior [UNTESTED; ROW_SPECIFIC]** Submit empty/dot/parent/separator names, escape paths, existing targets and branches used by another worktree.
   - Expected: Worktree validation rejects empty/dot/parent/separator names, root escape, existing paths, missing repo and branch conflict. Creation errors keep actionable context rather than pretending success.
2. **W42.edge [UNTESTED; ROW_SPECIFIC]** Each error preserves context; branch-in-use offers explicit Open Existing and never opens automatically.
   - Expected: Windows also rejects reserved device names/characters; normalization and canonical same-path checks matter on symlinks. Branch-in-use dialog offers navigation, not automatic branch duplication.
3. **W42.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Windows also rejects reserved device names/characters; normalization and canonical same-path checks matter on symlinks. Branch-in-use dialog offers navigation, not automatic branch duplication.
4. **W42.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W42.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W42.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W42.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestNativeBranchInUseOpenRequiresExplicitClick`: Branch-in-use handling does not open another worktree until explicit Open Existing click. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W43 — Settings configure worktree root, setup-script path, run-setup automatically and open-after-create. Repository `.orbit/worktree.json` overrides root/script, not global booleans.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::WorktreeConfig](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L186); [crates/orbit-pi/src/worktree.rs::WorktreeRepoConfig](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L284); [crates/orbit-pi/src/worktree.rs::expand_path](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L309); [crates/orbit-pi/src/worktree.rs::effective_directory](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L246)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W43.behavior [UNTESTED; ROW_SPECIFIC]** Set global worktree root/script/run/open preferences and repository root/script overrides.
   - Expected: Settings configure worktree root, setup-script path, run-setup automatically and open-after-create. Repository `.orbit/worktree.json` overrides root/script, not global booleans.
2. **W43.edge [UNTESTED; ROW_SPECIFIC]** Reload settings and fail a save; repository overrides must not replace global booleans or falsely persist a toggled value.
   - Expected: Global `~/.orbit-pi/worktrees.json`, defaults `.wt` and `.orbit/worktree-setup.sh`, auto-setup/open true. Absolute and `~/` paths supported. Missing/malformed config falls back safely; setup must be existing file.
3. **W43.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Global `~/.orbit-pi/worktrees.json`, defaults `.wt` and `.orbit/worktree-setup.sh`, auto-setup/open true. Absolute and `~/` paths supported. Missing/malformed config falls back safely; setup must be existing file.
4. **W43.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W43.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W43.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W43.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestSettingsPersistenceFailureIsTransactionalAndRetryable`: Failed worktree settings persistence leaves prior config intact and supports retry. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

