# Composer, transcript and interactive agent surfaces: C01–C04

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## C01 — Native multiline text editing

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/composer.rs::ComposerInput](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/composer.rs#L55); [crates/orbit-pi/src/main.rs::bind_keys](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L357); [crates/orbit-pi/src/app/session.rs::submit_composer_as](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L88)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C01.behavior [UNTESTED; ROW_SPECIFIC]** Type Unicode, combining marks, emoji and multiline text; select by word/line; cut, paste, undo and redo.
   - Expected: Wrapping/auto-grow/scroll; grapheme-aware caret/selection, word/line/document movement, copy/cut/paste/select-all, undo/redo and IME composition. Shift+Enter inserts newline.
2. **C01.edge [UNTESTED; ROW_SPECIFIC]** Use IME composition and candidate Enter, Shift+Enter and editor focus; none may accidentally submit a chat prompt.
   - Expected: IME candidate Enter must never submit; focused surface gates submission. Same input primitive is reused for editor, search and modal fields but key context differs.
3. **C01.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: IME candidate Enter must never submit; focused surface gates submission. Same input primitive is reused for editor, search and modal fields but key context differs.
4. **C01.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C01.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C01.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C01.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C01 [PARTIAL]:** Headless composition event and multiline submission pass. Real Chinese/Japanese/Korean OS IME candidates, clipboard and accessibility remain unexecuted. Source22/28px frames, compact omission, one-to-eight lines, surface_secondary, terse full-access and absent unavailable thinking now covered; current native recapture pending. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C02 — Prompt, steer and follow-up policy

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/composer_send.rs::SendMode](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/composer_send.rs#L10); [crates/orbit-pi/src/app/session.rs::submit_as](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L129); [crates/orbit-pi/src/theme.rs::UiPrefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/theme.rs#L324)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C02.behavior [UNTESTED; ROW_SPECIFIC]** Send an idle prompt; while busy send with Enter, Alt+Enter and Primary+Shift+Enter under each configured send mode.
   - Expected: Idle send starts prompt. While working, Enter uses configurable Follow-up/Steer mode, Alt+Enter its opposite, Primary+Shift+Enter explicitly steers; hints and button labels follow preference.
2. **C02.edge [UNTESTED; ROW_SPECIFIC]** Make the RPC send fail before acknowledgement; retain exact draft and attachments; reject whitespace-only input.
   - Expected: UI enters busy/awaiting immediately before first agent event; clear input only after send succeeds; failed send retains text and attachments. Empty text is not sent.
3. **C02.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: UI enters busy/awaiting immediately before first agent event; clear input only after send succeeds; failed send retains text and attachments. Empty text is not sent.
4. **C02.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C02.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C02.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C02.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C02 [PARTIAL]:** Exact tested sends/rejection pass. Source individual queue editing/removal and one/all queued delivery controls are incomplete in host UI. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C03 — Queue visualization and delivery

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/view.rs::queue_bar](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L4532); [crates/orbit-pi/src/app/view.rs::on_clear_queue](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L4719); [crates/orbit-pi/src/app/session.rs::reset_queue](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L284); [crates/orbit-pi/src/app/settings.rs::set_follow_up_mode](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5580)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C03.behavior [UNTESTED; ROW_SPECIFIC]** Queue multiple steer and follow-up messages, inspect order, edit/remove supported entries and let the agent settle.
   - Expected: Pending steering/follow-up queue appears above composer, offers clear_queue; set_follow_up_mode supports one-at-a-time/all delivery.
2. **C03.edge [UNTESTED; ROW_SPECIFIC]** Switch sessions with queued messages, fail one delivery and repeat the action; verify no duplicate or wrong-session delivery.
   - Expected: Optimistic queue display reconciles with queue_update; parked sessions maintain queue. Input not rendered as delivered until runtime actually delivers it.
3. **C03.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Optimistic queue display reconciles with queue_update; parked sessions maintain queue. Input not rendered as delivered until runtime actually delivers it.
4. **C03.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C03.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C03.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C03.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C02 [PARTIAL]:** Exact tested sends/rejection pass. Source individual queue editing/removal and one/all queued delivery controls are incomplete in host UI. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C04 — Slash commands and file mentions

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/composer_ops.rs::autocomplete_entries](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L9); [crates/orbit-pi/src/app/composer_ops.rs::commit_entry](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L113); [crates/orbit-pi/src/app/composer_ops.rs::attach_file](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L471)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C04.behavior [UNTESTED; ROW_SPECIFIC]** Type slash and file-mention prefixes, navigate results, accept a command/file, then submit.
   - Expected: Slash command catalog from pi; @ file autocomplete, query highlighting and arrow/Tab/Enter acceptance; + menu Attach image / Attach file / Mention file.
2. **C04.edge [UNTESTED; ROW_SPECIFIC]** Use spaces, Unicode and ambiguous paths; Escape and backspace must restore draft; test empty and stale result lists.
   - Expected: Autocomplete list capped at eight rows; path discovery is background/git-aware. Accepting completion takes precedence over prompt send. Attach file inserts a mention, not arbitrary contents upload.
3. **C04.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Autocomplete list capped at eight rows; path discovery is background/git-aware. Accepting completion takes precedence over prompt send. Attach file inserts a mention, not arbitrary contents upload.
4. **C04.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C04.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C04.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C04.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C03 [PARTIAL]:** Native source-derived slash/skill/command/file completion is mounted; real command metadata and off-thread bounded filesystem scan, focus/caret/Enter/AltEnter/steer/Escape paths have tests. Canonical hash-checked MyGo compatibility supplies undoable caret-aware replacement; cmd refuses unpatched startup. Actual OS IME candidate panel and full source visual comparison remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

