# Terminal, platform and window lifetime: W61–W64

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W61 — Integrated bottom terminal owns one real alacritty PTY/shell per tab. New tab uses current workspace; old tabs keep original shells/cwd across workspace/session switch and hidden panel. Toggle focuses shell; close tab ends that shell; closing last tab closes panel.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/terminal.rs::TerminalSession::with_shell](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L354); [crates/orbit-pi/src/terminal.rs::TerminalSession::drop](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L451); [crates/orbit-pi/src/terminal.rs::default_shell](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L515); [crates/orbit-pi/src/terminal.rs::TerminalPanel::set_workspace](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1675)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W61.behavior [UNTESTED; ROW_SPECIFIC]** Create two real PTY tabs in different temporary workspaces; hide/show terminal and switch session/workspace.
   - Expected: Integrated bottom terminal owns one real alacritty PTY/shell per tab. New tab uses current workspace; old tabs keep original shells/cwd across workspace/session switch and hidden panel. Toggle focuses shell; close tab ends that shell; closing last tab closes panel.
2. **W61.edge [UNTESTED; ROW_SPECIFIC]** Existing PTYs retain original shell/cwd; closing one terminates only it, including child process cleanup, while new tab uses current cwd.
   - Expected: Unix login shell from SHELL then zsh/bash/sh, `-l`; Windows COMSPEC/cmd. TERM=xterm-256color and COLORTERM=truecolor. Drop sends Shutdown; process state is in-memory, not persisted across app restart.
3. **W61.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Unix login shell from SHELL then zsh/bash/sh, `-l`; Windows COMSPEC/cmd. TERM=xterm-256color and COLORTERM=truecolor. Drop sends Shutdown; process state is in-memory, not persisted across app restart.
4. **W61.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W61.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W61.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W61.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/terminal`
**Existing test candidates, not automatic acceptance:** `internal/terminal/panel_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Terminal package race tests pass; optional live tests and macOS/Windows packaged delivery not established by this run.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W62 — Terminal renders real VT grid on canvas: ANSI/indexed/truecolor, attributes, wide cells, selections, cursor shapes/focus, scrollback indicator and resizing.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/terminal.rs::prepare_grid](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L594); [crates/orbit-pi/src/terminal.rs::Palette](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L100); [crates/orbit-pi/src/terminal.rs::EventProxy::send_event](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L279); [crates/orbit-pi/src/terminal.rs::TerminalView::poll](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1085)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W62.behavior [UNTESTED; ROW_SPECIFIC]** Feed ANSI/truecolor/wide-character/attribute/cursor sequences and scroll beyond visible rows; resize repeatedly.
   - Expected: Terminal renders real VT grid on canvas: ANSI/indexed/truecolor, attributes, wide cells, selections, cursor shapes/focus, scrollback indicator and resizing.
2. **W62.edge [UNTESTED; ROW_SPECIFIC]** Compare VT grid, selection, scrollback indicator and cursor in focused/unfocused states; no ANSI text simulation passes this gate.
   - Expected: 10,000 scrollback lines; min 2 columns/1 row; 16 ms channel drain; dirty flag avoids unnecessary frames, hidden tabs still drain. OSC color/size/cursor requests answered; Bell currently ignored.
3. **W62.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: 10,000 scrollback lines; min 2 columns/1 row; 16 ms channel drain; dirty flag avoids unnecessary frames, hidden tabs still drain. OSC color/size/cursor requests answered; Bell currently ignored.
4. **W62.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W62.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W62.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W62.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/terminal`
**Existing test candidates, not automatic acceptance:** `internal/terminal/panel_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Terminal package race tests pass; optional live tests and macOS/Windows packaged delivery not established by this run.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W63 — Terminal translates Enter/backspace/tab/shift-tab/arrows/home/end/page/insert/delete/F1–F12, application-cursor mode, C0 control keys and Alt/meta; app-platform shortcuts not sent to shell. Escape specifically goes to terminal so vim/less can exit, not global AbortRun.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/terminal.rs::control_byte](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L799); [crates/orbit-pi/src/terminal.rs::encode_key](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L829); [crates/orbit-pi/src/terminal.rs::on_key_down](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1207); [crates/orbit-pi/src/terminal.rs::on_terminal_escape](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1195)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W63.behavior [UNTESTED; ROW_SPECIFIC]** Send all declared terminal keys, modifiers, application-cursor mode and Escape to a controlled byte-recording PTY.
   - Expected: Terminal translates Enter/backspace/tab/shift-tab/arrows/home/end/page/insert/delete/F1–F12, application-cursor mode, C0 control keys and Alt/meta; app-platform shortcuts not sent to shell. Escape specifically goes to terminal so vim/less can exit, not global AbortRun.
2. **W63.edge [UNTESTED; ROW_SPECIFIC]** App-global platform shortcuts must not leak bytes to shell; Ctrl/Alt and F1–F12 encodings match source behavior.
   - Expected: Cmd/Super combinations remain app territory; Ctrl-C remains terminal control, copy/paste use Cmd or Ctrl+Shift. Alt uses base key/meta escape rather than OS option-produced glyph. Typing returns view to bottom and resets cursor blink.
3. **W63.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Cmd/Super combinations remain app territory; Ctrl-C remains terminal control, copy/paste use Cmd or Ctrl+Shift. Alt uses base key/meta escape rather than OS option-produced glyph. Typing returns view to bottom and resets cursor blink.
4. **W63.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W63.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W63.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W63.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/terminal`
**Existing test candidates, not automatic acceptance:** `internal/terminal/panel_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Terminal package race tests pass; optional live tests and macOS/Windows packaged delivery not established by this run.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## W64 — Mouse drag simple text selection, precise left-side cell boundary, copy selection and clipboard paste with bracketed-paste framing when enabled; smooth pixel/line scroll accumulator.

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/terminal.rs::grid_point](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1127); [crates/orbit-pi/src/terminal.rs::on_mouse_down](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1230); [crates/orbit-pi/src/terminal.rs::on_mouse_move](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1252); [crates/orbit-pi/src/terminal.rs::on_mouse_up](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1268)

**Fixture:** Disposable native host or CPU test harness, real temporary PTY where supported, controlled shell/byte recorder and stub external apps.

1. **W64.behavior [UNTESTED; ROW_SPECIFIC]** Drag/select at left cell boundary, copy selection, paste with bracketed mode and scroll fractional wheel deltas.
   - Expected: Mouse drag simple text selection, precise left-side cell boundary, copy selection and clipboard paste with bracketed-paste framing when enabled; smooth pixel/line scroll accumulator.
2. **W64.edge [UNTESTED; ROW_SPECIFIC]** Use wide characters, backward drag and hidden/unfocused terminal; no selection offset drift or unframed bracketed paste.
   - Expected: OSC clipboard store/load bridged through GPUI clipboard. No distinct rectangular/word/line selection or terminal search implementation located. Selection movement clamps to viewport/grid, scrollback coordinates corrected.
3. **W64.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Exit/fail the active child or external-app stub while hidden/unfocused; exercise restart and missing-platform capability.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: OSC clipboard store/load bridged through GPUI clipboard. No distinct rectangular/word/line selection or terminal search implementation located. Selection movement clamps to viewport/grid, scrollback coordinates corrected.
4. **W64.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **W64.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W64.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **W64.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/terminal`
**Existing test candidates, not automatic acceptance:** `internal/terminal/panel_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Remaining:** Terminal package race tests pass; optional live tests and macOS/Windows packaged delivery not established by this run.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

