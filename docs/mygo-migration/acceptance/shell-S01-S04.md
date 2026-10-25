# Shell, navigation, sessions and workspaces: S01–S04

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## S01 — Native shell and one-window lifetime

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/main_window.rs::open](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L77); [crates/orbit-pi/src/main_window.rs::spawn_heartbeat](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L136); [crates/orbit-pi/src/app/view.rs::render](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166); [crates/orbit-pi/src/platform.rs::titlebar_options](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L1507)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S01.behavior [UNTESTED; ROW_SPECIFIC]** Launch against displays 1280×800 and 1920×1080 and a no-display-metadata fixture; inspect initial window options, then unmaximize.
   - Expected: Initially maximized and focused, minimum960×640. Restore bounds are min(displayWidth×0.85,1440) by min(displayHeight×0.9,920), centered; no-display fallback1240×840. Composer is focused and app id is dev.orbit.pi.
2. **S01.edge [UNTESTED; ROW_SPECIFIC]** Verify same session and process identity after reopening; test a window smaller than the source minimum.
   - Expected: App-owned heartbeat survives window closure: 16 ms active, 90 ms idle. Closing a window must not silently kill a running session. Windows/Linux closure semantics remain platform-specific.
3. **S01.error [UNTESTED; ROW_SPECIFIC]** Inject failure from the native open_window adapter, then retry opening without creating another app controller.
   - Expected: The source logs 'Orbit: could not open the main window' and returns; no success window handle, duplicate OrbitApp or duplicate heartbeat is created.
4. **S01.loading [UNTESTED; ROW_SPECIFIC]** Start a fixture streaming Pi session, close all windows, advance the heartbeat, finish the session and reopen.
   - Expected: The app-owned heartbeat still drains while windowless:90ms idle and16ms active after each observed tick; repaint is unnecessary without a window. This is not a window-loading spinner.
5. **S01.persistence [UNTESTED; ROW_SPECIFIC]** Record OrbitApp identity and child PID; close the macOS window without quitting; reopen from Dock and compare.
   - Expected: The same global shell entity and existing Pi child are reused. This gate concerns in-process retention, not disk session restoration after full app exit.
6. **S01.keyboard [UNTESTED; ROW_SPECIFIC]** Use native window close and the macOS application reopen event; inspect app command bindings for a dedicated resize/reopen action.
   - Expected: No app-specific resize or reopen keyboard chord is invented. Composer focus is restored by open(); normal OS window controls retain their native behavior.
7. **S01.repeat [UNTESTED; ROW_SPECIFIC]** Close and reopen the main window three times while a controlled Pi child emits numbered events.
   - Expected: One retained controller and child, one heartbeat; every event is reduced once and returning window shows the current session. Actual macOS Dock behavior remains a native-OS gate.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S01 [PARTIAL]:** Source minimum/restore bounds and native traffic-light clearance pass. Native menus now auto-hide on Linux/Windows while retaining edit roles and shortcuts. Actual initial-maximize/Alt-F10/Dock lifetime on each OS remains separate acceptance. Linux window buttons are now hidden like pinned source; initial composer focus and conditional Open-in controls are native-CPU tested. Actual recapture pending. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S02 — Resizable workbench geometry

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/layout.rs::Layout](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/layout.rs#L29); [crates/orbit-pi/src/app/view.rs::render](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166); [crates/orbit-pi/src/app.rs::SIDEBAR_MIN_W](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app.rs#L87)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S02.behavior [UNTESTED; ROW_SPECIFIC]** Drag Sidebar, Review, Explorer and Terminal below minimum, to a normal size and above maximum; toggle full-width Review independently.
   - Expected: Sidebar default248/min200; Review default460/min300/drag max900; Explorer default248/min172/max460; Terminal default260/min96. Main content retains480px reserve where the source resize handler applies it. Full-width Review preserves its separate restore width.
2. **S02.edge [UNTESTED; ROW_SPECIFIC]** Persist a valid size, then supply negative, huge and nonfinite saved values in a copied home; relaunch.
   - Expected: Four sizes in ~/.orbit-pi/layout.json; writes settle after 500 ms. Invalid/nonfinite/out-of-range stored sizes discarded. Open/closed surfaces are not all persisted in this store.
3. **S02.error [UNTESTED; ROW_SPECIFIC]** Make the fixture ~/.orbit-pi/layout.json parent unwritable; drag and allow the settle flush. Restore permission, change a size again and flush.
   - Expected: Pinned Layout.persist ignores write errors: current live size remains, old disk state is not falsely claimed updated, and a later changed size can persist. This action has no session-scan or Pi dependency.
4. **S02.loading [UNTESTED; ROW_SPECIFIC]** Record a size change at t=0; flush at499ms and500ms. Record an identical value at400ms and repeat.
   - Expected: No disk write before500ms; one flush when changed value has settled. Recording the same value does not rearm the timer. Resizing is immediate and has no loading spinner.
5. **S02.persistence [UNTESTED; ROW_SPECIFIC]** Persist sidebar_width248, sidepane_width460, explorer_width300, terminal_height260 in ~/.orbit-pi/layout.json; relaunch. Repeat with missing, text,10,99999 and nonfinite parsed fields.
   - Expected: Only those four sizes belong to this store. Numeric finite values in inclusive80…4000 survive parsing before panel-specific clamps; invalid fields become absent/default independently. Malformed file defaults all fields. Open panels, fonts and theme are not restored by this file.
6. **S02.keyboard [UNTESTED; ROW_SPECIFIC]** After resizing use Cmd+B/Ctrl+B, Cmd+Shift+E/Ctrl+Shift+E and Cmd+J/Ctrl+J to toggle Sidebar, Explorer and Terminal; reopen each.
   - Expected: The registered source panel actions toggle visibility without treating those keys as resize commands; remembered size is retained. Review ToggleSidePanel has no chord in this command entry, so no invented key is required.
7. **S02.repeat [UNTESTED; ROW_SPECIFIC]** Perform100 pointer size updates, stop, then leave the splitter stationary and reopen the panel three times.
   - Expected: Live geometry updates without writing per pixel; one settled write per changed burst. Stationary input does not repeatedly rewrite layout.json; visibility toggles retain size.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S02 [PARTIAL]:** DIP composer and modal bounds pass, including800×560 stress below source minimum. Full panel drag/relaunch/font-density matrix and matched pixels remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S03 — Global command/action registry

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/commands.rs::COMMANDS](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commands.rs#L224); [crates/orbit-pi/src/commands.rs::BINDINGS](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/commands.rs#L794); [crates/orbit-pi/src/main.rs::bind_keys](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L357); [crates/orbit-pi/src/app/pickers.rs::run_command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pickers.rs#L415)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S03.behavior [UNTESTED; ROW_SPECIFIC]** Enumerate every BINDINGS entry; invoke the matching menu, palette action and keyboard chord from its declared focus context.
   - Expected: Command palette, native menu, shortcuts help and pointer controls dispatch the same action identities; labels use Cmd on macOS and Ctrl elsewhere.
2. **S03.edge [UNTESTED; ROW_SPECIFIC]** Repeat each chord from composer, editor, terminal and dialog focus; verify out-of-context actions are not dispatched.
   - Expected: Context-specific bindings override global behavior. Access/workflow/MCP and worktrees commands include palette-only entries without fabricated chords.
3. **S03.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Context-specific bindings override global behavior. Access/workflow/MCP and worktrees commands include palette-only entries without fabricated chords.
4. **S03.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S03.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S03.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S03.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S03 [PARTIAL]:** These exact native CPU flows pass, not every command/context or OS accelerator. [Scoped steps/tests](../native-ui-acceptance.json).

**Owner challenge UI-S04 [PARTIAL]:** Native menu guard includes module-owned Git/Worktrees/MCP/Resources/report review/preview and Usage filter overlays. CPU routing tests pass; actual platform accelerator ordering remains separate. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## S04 — Command palette versus transcript search

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/command_palette.rs::CommandPalette](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/command_palette.rs#L313); [crates/orbit-pi/src/app/pickers.rs::toggle_command_palette](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pickers.rs#L288); [crates/orbit-pi/src/app/pickers.rs::run_palette_command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pickers.rs#L387)

**Fixture:** Two temporary workspaces, an isolated HOME/PI_CODING_AGENT_DIR, fixture session JSONL files and controllable Pi RPC children.

1. **S04.behavior [UNTESTED; ROW_SPECIFIC]** Open sidebar Search and the command palette; search a session title, command label and settings synonym; select each result.
   - Expected: Primary+P or Primary+K opens fuzzy-search sessions/commands/settings and dynamic MCP actions. Sidebar Search opens this palette, not transcript-content search. Up/down, Enter, Escape, scrim dismissal and focus return must match.
2. **S04.edge [UNTESTED; ROW_SPECIFIC]** Enter text that appears only inside a transcript; confirm palette does not pretend to be transcript search; test no matches and Escape.
   - Expected: State-driven visibility, disabled actions and active-title variants; no backend content-search service.
3. **S04.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay/fail session scan or runtime child used by this action; repeat while changing workspace.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: State-driven visibility, disabled actions and active-title variants; no backend content-search service.
4. **S04.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **S04.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **S04.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **S04.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S03 [PARTIAL]:** These exact native CPU flows pass, not every command/context or OS accelerator. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

