# Worktrees and setup lifecycle: W44–W47

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W44 — First setup-script run per repository requires explicit AllowSetup dialog; create dialog can disable setup for that worktree.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::setup_allowed](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L261); [crates/orbit-pi/src/worktree.rs::allow_setup](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L267); [crates/orbit-pi/src/worktree.rs::setup_repo_key](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L277); [crates/orbit-pi/src/app/worktrees.rs::submit_create_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L323)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W44.behavior [UNTESTED; ROW_SPECIFIC]** Create with setup enabled in an untrusted repository; cancel Allow Setup, then explicitly allow.
   - Expected: First setup-script run per repository requires explicit AllowSetup dialog; create dialog can disable setup for that worktree.
2. **W44.edge [UNTESTED; ROW_SPECIFIC]** Trust canonical repository once; disallow per-create setup and fail trust-store persistence without silently running the script.
   - Expected: Permission remembered under canonical repo root in `allowed_setup_repos`. Script executes arbitrary user-controlled commands; do not bypass the permission merely while restyling create UI. Permission config writes currently best-effort in controller.
3. **W44.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Permission remembered under canonical repo root in `allowed_setup_repos`. Script executes arbitrary user-controlled commands; do not bypass the permission merely while restyling create UI. Permission config writes currently best-effort in controller.
4. **W44.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W44.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W44.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W44.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestSetupAllowRemembersCanonicalRepositoryBeforeCreation`: Explicit consent persists canonical repository trust before setup-enabled creation. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W45 — Setup runs after successful Git creation and before automatic opening; expandable captured stdout/stderr/exit result, success/warning toasts, retry, dismiss and manual open.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree_setup.rs::SetupOutcome](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree_setup.rs#L26); [crates/orbit-pi/src/worktree_setup.rs::script_command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree_setup.rs#L66); [crates/orbit-pi/src/worktree_setup.rs::run](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree_setup.rs#L90); [crates/orbit-pi/src/app/worktrees.rs::begin_worktree_setup](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L474)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W45.behavior [UNTESTED; ROW_SPECIFIC]** Run successful/failed/cancelled setup after worktree creation; inspect stdout/stderr/exit and retry/dismiss/manual open.
   - Expected: Setup runs after successful Git creation and before automatic opening; expandable captured stdout/stderr/exit result, success/warning toasts, retry, dismiss and manual open.
2. **W45.edge [UNTESTED; ROW_SPECIFIC]** Verify setup completes before auto-open, failed retry uses same worktree and repeated navigation cannot rerun creation.
   - Expected: Background subprocess cwd is new worktree; `ORBIT_ROOT_PATH`, `ORBIT_WORKTREE_PATH`, `GIT_TERMINAL_PROMPT=0`. Unix `/bin/sh`, Windows `cmd /C` (POSIX script not magically supported). Failure keeps created worktree and suppresses auto-open. No timeout or live output streaming in this module.
3. **W45.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Background subprocess cwd is new worktree; `ORBIT_ROOT_PATH`, `ORBIT_WORKTREE_PATH`, `GIT_TERMINAL_PROMPT=0`. Unix `/bin/sh`, Windows `cmd /C` (POSIX script not magically supported). Failure keeps created worktree and suppresses auto-open. No timeout or live output streaming in this module.
4. **W45.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W45.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W45.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W45.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestSetupConsentFailureAndRetry`: Setup consent, failure and retry do not duplicate worktree creation. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `4cffa4086622c4bf76ab8da24cf144f678119ec1b2f926fbdb09dca2d126d0a3`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W46 — Open worktree adds workspace, persists current directory and synchronizes Review/Git/Explorer/new-terminal context; historical sessions retain their recorded workspace. Missing cold-session cwd shows unavailable dialog with locate/worktrees action, no silent substitution for a deleted directory.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/worktrees.rs::open_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L564); [crates/orbit-pi/src/app/worktrees.rs::worktree_unavailable_dialog_body](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2316); [crates/orbit-pi/src/app.rs::set_current_workspace](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2048); [crates/orbit-pi/src/app.rs::sync_review_context](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L2069)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W46.behavior [UNTESTED; ROW_SPECIFIC]** Open main/linked worktree and switch between historical sessions; create a terminal after switching.
   - Expected: Open worktree adds workspace, persists current directory and synchronizes Review/Git/Explorer/new-terminal context; historical sessions retain their recorded workspace. Missing cold-session cwd shows unavailable dialog with locate/worktrees action, no silent substitution for a deleted directory.
2. **W46.edge [UNTESTED; ROW_SPECIFIC]** Missing cold-session cwd shows unavailable; active services use selected workspace while old terminal/session cwd remains original.
   - Expected: `set_current_workspace` canonicalizes and resets cached worktree root/fetch generation. Opening a worktree itself does not explicitly respawn an existing pi process in this method; distinguish selected-workspace update from a new RPC-session cwd guarantee.
3. **W46.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `set_current_workspace` canonicalizes and resets cached worktree root/fetch generation. Opening a worktree itself does not explicitly respawn an existing pi process in this method; distinguish selected-workspace update from a new RPC-session cwd guarantee.
4. **W46.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W46.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W46.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W46.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W47 — Rename worktree directory under configured root or move to explicit resolved destination, preserving branch; reject blank/existing/no-op target. Active selected workspace follows moved path.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::WorktreeManager::move_to](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L913); [crates/orbit-pi/src/app/worktrees.rs::start_rename_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L620); [crates/orbit-pi/src/app/worktrees.rs::start_move_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L631); [crates/orbit-pi/src/app/worktrees.rs::submit_move_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L647)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W47.behavior [UNTESTED; ROW_SPECIFIC]** Rename under configured root and move to relative/absolute/tilde destination; inspect branch preservation.
   - Expected: Rename worktree directory under configured root or move to explicit resolved destination, preserving branch; reject blank/existing/no-op target. Active selected workspace follows moved path.
2. **W47.edge [UNTESTED; ROW_SPECIFIC]** Reject blank/existing/no-op paths; move active workspace updates all relevant service paths and errors retain the original worktree.
   - Expected: Uses `git worktree move`, not blind filesystem rename; mutation runs off-thread and refreshes list. Main-worktree menu omits rename/move.
3. **W47.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Uses `git worktree move`, not blind filesystem rename; mutation runs off-thread and refreshes list. Main-worktree menu omits rename/move.
4. **W47.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W47.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W47.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W47.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

