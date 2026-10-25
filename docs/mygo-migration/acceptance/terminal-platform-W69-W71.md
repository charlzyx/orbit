# Terminal, platform and window lifetime: W69–W71

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W69 — Platform helpers reveal files, default-open paths, launch interactive external terminal commands, open browser links and find home directory.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/platform.rs::reveal_in_file_manager](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L605); [crates/orbit-pi/src/platform.rs::open_path_default](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L664); [crates/orbit-pi/src/platform.rs::open_terminal_command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L957); [crates/orbit-pi/src/platform.rs::open_url](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L1013)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W69.behavior [UNTESTED; ROW_SPECIFIC]** Reveal file, default-open path, launch interactive terminal command and open browser URL in controlled native stubs.
   - Expected: Platform helpers reveal files, default-open paths, launch interactive external terminal commands, open browser links and find home directory.
2. **W69.edge [UNTESTED; ROW_SPECIFIC]** Reject unsafe URL schemes/arguments and missing home/app; record exact target without launching unrequested external actions in CI.
   - Expected: macOS open/AppleScript, Windows explorer-select/cmd start, Linux xdg-open parent and available terminal fallback. Browser helper permits HTTP(S) only. Home fallback includes USERPROFILE and HOMEDRIVE/HOMEPATH, not HOME alone.
3. **W69.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: macOS open/AppleScript, Windows explorer-select/cmd start, Linux xdg-open parent and available terminal fallback. Browser helper permits HTTP(S) only. Home fallback includes USERPROFILE and HOMEDRIVE/HOMEPATH, not HOME alone.
4. **W69.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W69.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W69.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W69.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native helpers/domain tests pass; real platform integrations and UI selection flows remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W70 — One shell entity survives closing and reopening main window (macOS Dock reopen); main window opens maximized with screen-relative restore size/minimum 960×640. Background heartbeat continues windowless. Windows/Linux have own minimize/maximize/restore/close/drag behavior; macOS native traffic-light chrome.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/main_window.rs::Shell](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L30); [crates/orbit-pi/src/main_window.rs::shell](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L58); [crates/orbit-pi/src/main_window.rs::open](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L77); [crates/orbit-pi/src/main_window.rs::install_reopen_handler](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L42)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W70.behavior [UNTESTED; ROW_SPECIFIC]** Launch/close/reopen native window while a session runs, inspect maximized/restore/minimum bounds.
   - Expected: One shell entity survives closing and reopening main window (macOS Dock reopen); main window opens maximized with screen-relative restore size/minimum 960×640. Background heartbeat continues windowless. Windows/Linux have own minimize/maximize/restore/close/drag behavior; macOS native traffic-light chrome.
2. **W70.edge [UNTESTED; ROW_SPECIFIC]** Run actual macOS Dock reopen and Windows/Linux closure checks separately; cross-build cannot satisfy windowless heartbeat/process identity.
   - Expected: Shell Global retains pi/transcript/feature entities. App-bound heartbeat is 90 ms idle, 16 ms active, not old documented window-bound behavior. macOS minimum 13, Windows build floor 17763 in host support checks.
3. **W70.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Shell Global retains pi/transcript/feature entities. App-bound heartbeat is 90 ms idle, 16 ms active, not old documented window-bound behavior. macOS minimum 13, Windows build floor 17763 in host support checks.
4. **W70.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W70.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W70.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W70.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native helpers/domain tests pass; real platform integrations and UI selection flows remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W71 — Keyboard commands, menu/palette actions and UI buttons share registered actions; Git tab shortcuts are actually platform+Alt+1…6, not stale prose saying Cmd+1…6.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/commands.rs::CommandId](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commands.rs#L22); [crates/orbit-pi/src/commands.rs::BINDINGS](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commands.rs#L794)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W71.behavior [UNTESTED; ROW_SPECIFIC]** Enumerate source CommandId/BINDINGS and compare menu/palette/button dispatch with actual key events.
   - Expected: Keyboard commands, menu/palette actions and UI buttons share registered actions; Git tab shortcuts are actually platform+Alt+1…6, not stale prose saying Cmd+1…6.
2. **W71.edge [UNTESTED; ROW_SPECIFIC]** Verify Git tabs use Primary+Alt+1…6; test editor/terminal/dialog focus isolation and each platform's displayed chord labels.
   - Expected: `secondary` resolves platform vocabulary; terminal context and Review context override global bindings deliberately. First-run new task closes panels and menus as a coordinated reset.
3. **W71.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `secondary` resolves platform vocabulary; terminal context and Review context override global bindings deliberately. First-run new task closes panels and menus as a coordinated reset.
4. **W71.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W71.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W71.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W71.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

