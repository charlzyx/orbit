# Terminal, platform and window lifetime: W65–W68

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W65 — Terminal tabs display login-shell icon, OSC title or folder fallback, own cwd tooltip; tab strip scrolls; next/previous wrap; resize height; restart only active tab in original cwd, preserving others; exited overlay and startup error state.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/terminal.rs::ShellProfile](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L466); [crates/orbit-pi/src/terminal.rs::label](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1052); [crates/orbit-pi/src/terminal.rs::tooltip](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1041); [crates/orbit-pi/src/terminal.rs::cycle_tab](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1782)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W65.behavior [UNTESTED; ROW_SPECIFIC]** Change OSC title, cycle tabs, scroll tab strip, resize and restart active terminal in original cwd.
   - Expected: Terminal tabs display login-shell icon, OSC title or folder fallback, own cwd tooltip; tab strip scrolls; next/previous wrap; resize height; restart only active tab in original cwd, preserving others; exited overlay and startup error state.
2. **W65.edge [UNTESTED; ROW_SPECIFIC]** Show exited overlay/exit status and relaunch only active tab; hidden terminal invalidation and other shell state remain intact.
   - Expected: Restart minimum 650 ms visible feedback. Hidden/full-page-suppressed terminal stops requesting frames without killing shells. Platform-specific new/close tab chords avoid Ctrl-T/Ctrl-W shell conflicts on non-macOS.
3. **W65.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Restart minimum 650 ms visible feedback. Hidden/full-page-suppressed terminal stops requesting frames without killing shells. Platform-specific new/close tab chords avoid Ctrl-T/Ctrl-W shell conflicts on non-macOS.
4. **W65.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W65.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W65.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W65.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/terminal`
**Existing test candidates, not automatic acceptance:** `internal/terminal/panel_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Terminal package race tests pass; optional live tests and macOS/Windows packaged delivery not established by this run.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W66 — Open-in discovers installed editors/terminals/file manager once off-thread with real app icons. macOS catalog VS Code/Cursor/Zed/Finder/Terminal/Termy/iTerm2/Kitty/Ghostty/Warp/Xcode/Rider/Android Studio; Windows catalog VS Code/Cursor/Zed/Explorer/Windows Terminal/Rider/Android Studio.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/platform.rs::ExternalApp](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L22); [crates/orbit-pi/src/platform.rs::OPEN_IN_CATALOG](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L41); [crates/orbit-pi/src/platform.rs::detect_open_in_apps](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L138); [crates/orbit-pi/src/app/open_in.rs::detect_open_in_apps](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L18)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W66.behavior [UNTESTED; ROW_SPECIFIC]** Run installed-app detection off-thread with fixture platform locations; inspect recognized apps and real icons.
   - Expected: Open-in discovers installed editors/terminals/file manager once off-thread with real app icons. macOS catalog VS Code/Cursor/Zed/Finder/Terminal/Termy/iTerm2/Kitty/Ghostty/Warp/Xcode/Rider/Android Studio; Windows catalog VS Code/Cursor/Zed/Explorer/Windows Terminal/Rider/Android Studio.
2. **W66.edge [UNTESTED; ROW_SPECIFIC]** Missing/invalid app binaries and platform-only catalog entries are not presented as installed; cache refresh follows source contract.
   - Expected: macOS Launch Services bundle IDs incl alternate variants; icons rasterized at 36 px for 18 px slot. Windows executable/PATH/environment expansion and serialized icon extraction/fallback. Linux currently returns empty catalog; UI absent, not simulated installed apps.
3. **W66.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: macOS Launch Services bundle IDs incl alternate variants; icons rasterized at 36 px for 18 px slot. Windows executable/PATH/environment expansion and serialized icon extraction/fallback. Linux currently returns empty catalog; UI absent, not simulated installed apps.
4. **W66.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W66.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W66.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W66.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native helpers/domain tests pass; real platform integrations and UI selection flows remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W67 — Split open-in control: preferred app primary action and searchable installed-app menu; selecting app opens current workspace and remembers choice. Primary fallback opening does not overwrite unavailable preference.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/open_in.rs::on_open_in_primary](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L77); [crates/orbit-pi/src/app/open_in.rs::on_open_in_caret](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L96); [crates/orbit-pi/src/app/open_in.rs::open_workspace_in_app](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L40); [crates/orbit-pi/src/app/open_in.rs::render_open_in_control](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L126)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W67.behavior [UNTESTED; ROW_SPECIFIC]** Use preferred Open In primary action and searchable menu; select another installed app.
   - Expected: Split open-in control: preferred app primary action and searchable installed-app menu; selecting app opens current workspace and remembers choice. Primary fallback opening does not overwrite unavailable preference.
2. **W67.edge [UNTESTED; ROW_SPECIFIC]** Temporarily remove saved app and use fallback; primary fallback must not overwrite remembered unavailable preference.
   - Expected: Requires workspace and installed catalog; dropdown focuses filter, excludes Work-in menu, 200 ms dismiss guard. macOS Termy uses encoded `termy://new?dir=…`; Windows Terminal uses `-d`. Native launch itself is best-effort.
3. **W67.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Requires workspace and installed catalog; dropdown focuses filter, excludes Work-in menu, 200 ms dismiss guard. macOS Termy uses encoded `termy://new?dir=…`; Windows Terminal uses `-d`. Native launch itself is best-effort.
4. **W67.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W67.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W67.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W67.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native helpers/domain tests pass; real platform integrations and UI selection flows remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W68 — Open-in preference resolution: workspace-specific choice, legacy global fallback, platform file manager, then first installed app. Temporarily unavailable saved app stays remembered for reinstall.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/platform.rs::OpenInPrefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L887); [crates/orbit-pi/src/platform.rs::preferred_app](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L913); [crates/orbit-pi/src/platform.rs::remember](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L930); [crates/orbit-pi/src/platform.rs::persist_to](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L941)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W68.behavior [UNTESTED; ROW_SPECIFIC]** Set workspace-specific then legacy global Open In preferences; remove/reinstall preferred app fixture.
   - Expected: Open-in preference resolution: workspace-specific choice, legacy global fallback, platform file manager, then first installed app. Temporarily unavailable saved app stays remembered for reinstall.
2. **W68.edge [UNTESTED; ROW_SPECIFIC]** Resolve workspace/global/file-manager/first-installed precedence exactly and preserve unavailable choice across restart.
   - Expected: `~/.orbit-pi/open-in.json`; serialized async saves and temp-file atomic replacement; save failure toast. Missing file defaults, malformed/unreadable logs and defaults.
3. **W68.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: `~/.orbit-pi/open-in.json`; serialized async saves and temp-file atomic replacement; save failure toast. Missing file defaults, malformed/unreadable logs and defaults.
4. **W68.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W68.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W68.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W68.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Native helpers/domain tests pass; real platform integrations and UI selection flows remain pending.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

