# Worktrees and setup lifecycle: W48–W51

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W48 — Remove worktree asks explicit clean/dirty confirmation after reading dirty state; dirty confirmation uses `--force`; branch is never deleted. Main/unknown/locked worktrees refused by backend.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::WorktreeManager::is_dirty](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L784); [crates/orbit-pi/src/worktree.rs::WorktreeManager::remove](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L887); [crates/orbit-pi/src/app/worktrees.rs::start_remove_worktree](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L724); [crates/orbit-pi/src/app/worktrees.rs::remove_worktree_dialog_body](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2235)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W48.behavior [UNTESTED; ROW_SPECIFIC]** Remove clean and dirty disposable worktrees after inspecting confirmation; cancel first.
   - Expected: Remove worktree asks explicit clean/dirty confirmation after reading dirty state; dirty confirmation uses `--force`; branch is never deleted. Main/unknown/locked worktrees refused by backend.
2. **W48.edge [UNTESTED; ROW_SPECIFIC]** Main/unknown/locked worktrees are refused; force only follows dirty confirmation and branch remains after removal.
   - Expected: UI never offers remove for main. Locked remains refused even with force. Existing dirty-probe error defaults to false, but backend checks again before unforced removal. Removed-session recovery remains separate.
3. **W48.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: UI never offers remove for main. Locked remains refused even with force. Existing dirty-probe error defaults to false, but backend checks again before unforced removal. Removed-session recovery remains separate.
4. **W48.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W48.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W48.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W48.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W49 — Lock/unlock, prune stale administrative entries, repair metadata after manual move, reveal and copy path. Worktree row click opens; right-click and ellipsis menus share actions.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/worktree.rs::lock](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L925); [crates/orbit-pi/src/worktree.rs::unlock](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L945); [crates/orbit-pi/src/worktree.rs::prune](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L959); [crates/orbit-pi/src/worktree.rs::repair](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/worktree.rs#L966)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W49.behavior [UNTESTED; ROW_SPECIFIC]** Lock/unlock, prune stale metadata, repair manually moved worktree and copy/reveal path.
   - Expected: Lock/unlock, prune stale administrative entries, repair metadata after manual move, reveal and copy path. Worktree row click opens; right-click and ellipsis menus share actions.
2. **W49.edge [UNTESTED; ROW_SPECIFIC]** Failure and cancellation cannot replay a previous mutation; row/right-click/ellipsis actions refer to the same worktree.
   - Expected: Mutation helper serializes one operation and drops a second request rather than queuing stale row targets. Lock reason backend exists; UI currently passes none. Failures route to page banner or branch-in-use dialog.
3. **W49.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Mutation helper serializes one operation and drops a second request rather than queuing stale row targets. Lock reason backend exists; UI currently passes none. Failures route to page banner or branch-in-use dialog.
4. **W49.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W49.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W49.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W49.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestFailedLockRetryDoesNotRepeatPreviousCreate`: Retry of a failed lock cannot replay a previous Create operation. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W50 — Status “Work in” picker has Local/main, linked worktrees and New; Up/Down wrap selection, Enter chooses, Escape closes/restores composer focus; mutually excludes overlapping popovers.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/worktrees.rs::work_in_chip](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2699); [crates/orbit-pi/src/app/worktrees.rs::work_in_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2449); [crates/orbit-pi/src/app/worktrees.rs::toggle_work_in_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2520); [crates/orbit-pi/src/app/worktrees.rs::on_work_in_next](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2565)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W50.behavior [UNTESTED; ROW_SPECIFIC]** Open Work In from composer; inspect Local first, linked worktrees then New; choose current item and another item.
   - Expected: Current worktree is selected/highlighted; selecting it closes without reloading the workspace. Selecting another invokes exactly its path; New opens a quick-create dialog with name suggestion.
2. **W50.edge [UNTESTED; ROW_SPECIFIC]** Wrap at boundaries, exclude overlapping popovers and restore composer focus; new worktree workflow keeps current session identity.
   - Expected: Chip intentionally absent at main checkout and no-repo; shown in linked worktree. Top-bar Worktrees page remains entry point from main. Outside-click mouseup reopening guarded 200 ms.
3. **W50.error [UNTESTED; ROW_SPECIFIC]** Remove the selected linked worktree externally before choosing it; open Work In while another worktree dialog is active.
   - Expected: Missing path follows source unavailable/open handling rather than silently creating it; an existing modal suppresses the overlapping picker and keeps its focus.
4. **W50.loading [UNTESTED; ROW_SPECIFIC]** Delay worktree discovery and render composer repeatedly; release results after changing workspace.
   - Expected: Only current workspace rows become selectable. No fake main/linked entries are rendered while discovery is unavailable; async result guards apply.
5. **W50.persistence [UNTESTED; ROW_SPECIFIC]** Open/dismiss picker, then restart app from identical preferences; choose a different workspace and restart again.
   - Expected: Picker visibility/highlight itself is ephemeral. Current-workspace persistence is owned by the host workspace selection, not a new picker-specific preference file.
6. **W50.keyboard [UNTESTED; ROW_SPECIFIC]** From Local press Up to wrap to New, Down to wrap to Local, Down to first linked and Enter. Reopen and press Escape three times in immediate succession.
   - Expected: Up/Down wrap exact displayed order; Enter selects once. Escape always closes and focuses composer; immediate Escape/reopen is not delayed. New transfers focus to its dialog, never back to composer.
7. **W50.repeat [UNTESTED; ROW_SPECIFIC]** Click trigger again while open; dismiss by outside pointer and immediately release over trigger; open a competing composer popup.
   - Expected: Trigger toggles closed without reopening on same gesture. Source200ms pointer-dismiss guard applies only to mouse dismissal, not keyboard Escape. Opening Work In closes source-overlapping popovers via host callback.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestWorkInDismissalRestoresComposerFocus`: Actual Escape returns focus to composer across three immediate reopen/dismiss cycles. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestWorkInKeyboardWrapNoopAndDialogFocusGuard`: Arrow/Enter wrap, current-worktree no-op and New/existing-dialog focus guard match source. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.
- **PASS** `TestWorkInRowsAndNativeMenuSelection`: Native Work In menu uses actual main/linked/New rows and selections. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Additional source-specific procedures:** [work-in-focus-history.json](work-in-focus-history.json)

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W51 — Worktrees refresh/busy/error/empty UI, setup banner and dismissible introductory guide. External changes trigger throttled list refresh while page visible.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/worktrees.rs::refresh_worktrees](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L116); [crates/orbit-pi/src/app/worktrees.rs::refresh_worktrees_soon](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L168); [crates/orbit-pi/src/app/worktrees.rs::worktree_howto_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1506); [crates/orbit-pi/src/app/worktrees.rs::worktree_empty_state](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L3103)

**Fixture:** Temporary main repository and worktree root, isolated HOME, controlled setup script and no user repositories.

1. **W51.behavior [UNTESTED; ROW_SPECIFIC]** Refresh Worktrees while empty/loading/error, dismiss guide and inspect setup banner.
   - Expected: Worktrees refresh/busy/error/empty UI, setup banner and dismissible introductory guide. External changes trigger throttled list refresh while page visible.
2. **W51.edge [UNTESTED; ROW_SPECIFIC]** Burst external events inside throttle and while busy; one pending refresh must eventually occur without requiring another event.
   - Expected: 2 s watcher-refresh throttle; fetch generation discards stale workspace result and pending refresh reruns. Guide dismissal stored via transcript hint persistence.
3. **W51.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Fail the current Git/setup/config write; retry after navigation and verify no prior mutation is replayed.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: 2 s watcher-refresh throttle; fetch generation discards stale workspace result and pending refresh reruns. Guide dismissal stored via transcript hint persistence.
4. **W51.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W51.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W51.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W51.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/workspace/worktrees.go`, `internal/workspace/setup.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/labels.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/worktrees_test.go`, `internal/worktreeui/state_test.go`, `internal/worktreeui/view_test.go`, `internal/worktreeui/picker_test.go`

**Executed narrow assertions:**
- **PASS** `TestThrottledWorktreeWatchEventuallyRefreshesWithoutAnotherEvent`: A throttled/busy watch event yields an eventual refresh without a second event. [Evidence](workspace-evidence.json). Tested dirty-tree manifest `8a0cce65c52c9e7494d292a39cf6d34c28d7a6f9a75896e80a43b2ab711c5c59`, not a published commit.

**Remaining:** Real temporary worktree operations, setup consent/failure/cancel, transactional persistence and scoped native typing/confirm/retry tests pass. Host guide preference/translation hooks exist. Complete localization, host keyboard/focus and platform/pixel equivalence remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

