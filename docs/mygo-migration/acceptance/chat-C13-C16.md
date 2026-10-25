# Composer, transcript and interactive agent surfaces: C13–C16

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## C13 — In-transcript find

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/search.rs::TranscriptSearch](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/search.rs#L12); [crates/orbit-pi/src/app/search.rs::sync_transcript_search](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/search.rs#L71); [crates/orbit-pi/src/transcript.rs::find_messages](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L2388); [crates/orbit-pi/src/main.rs::bind_keys](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L357)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C13.behavior [UNTESTED; ROW_SPECIFIC]** Open transcript Find; search plain text, navigate matches and dismiss back to prior focus.
   - Expected: Primary+F opens find; Enter/Shift+Enter next/previous, hit count, selected row wash, Escape close and focus return.
2. **C13.edge [UNTESTED; ROW_SPECIFIC]** Search while streaming and after changing session; test case behavior, no result and hidden/collapsed content according to source.
   - Expected: Case-insensitive message-level substring matches, not per-occurrence highlighting or global content search. Refresh when query/message count changes; matching streaming text with unchanged count is not guaranteed.
3. **C13.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Case-insensitive message-level substring matches, not per-occurrence highlighting or global content search. Refresh when query/message count changes; matching streaming text with unchanged count is not guaranteed.
4. **C13.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C13.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C13.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C13.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C09 [PARTIAL]:** Search implementation exists, but dedicated interaction test and source highlight/active-hit keyboard behavior are not yet verified. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C14 — Stop, retry and compact states

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/session.rs::on_abort](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L417); [crates/orbit-pi/src/app/view.rs::run_status_strip](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3882); [crates/orbit-pi/src/app/runtime.rs::on_command_failure](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L50); [crates/orbit-pi/src/app/settings.rs::compact_now](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5811)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C14.behavior [UNTESTED; ROW_SPECIFIC]** Abort a busy turn, exercise retry lifecycle and request compaction through the source control.
   - Expected: Stop button/Escape abort, retry banner with attempt/maximum/provider error and cancel retry, manual compact and auto-compaction/retry controls.
2. **C14.edge [UNTESTED; ROW_SPECIFIC]** Abort during retry wait and compaction; fail or cancel RPC and ensure status, draft and subsequent prompt recover.
   - Expected: Awaiting-first-token, streaming, compacting, retrying, aborted and exited are distinct; error banner is dismissible; runtime failure clears busy so Stop cannot stay stuck.
3. **C14.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Awaiting-first-token, streaming, compacting, retrying, aborted and exited are distinct; error banner is dismissible; runtime failure clears busy so Stop cannot stay stuck.
4. **C14.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C14.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C14.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C14.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C10 [PARTIAL]:** Pure adapters and agent compaction reducer tests pass. Full native retry/compaction indicators and every button/state transition remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C15 — Context/cost meter

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/context_meter.rs::ContextMeterData](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/context_meter.rs#L633); [crates/orbit-pi/src/context_meter.rs::usage_slices](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/context_meter.rs#L120); [crates/orbit-pi/src/app/runtime.rs::refresh_context_stats](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L123); [crates/orbit-pi/src/app/view.rs::context_button](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3195)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C15.behavior [UNTESTED; ROW_SPECIFIC]** Populate context/token/cost metrics and hover the context meter; vary active model limits.
   - Expected: Used context ring, hover summary, detail breakdown, compact action, session tokens/cost/cache rate.
2. **C15.edge [UNTESTED; ROW_SPECIFIC]** Use unknown price, zero/absent context limits and overflow; label unavailable values instead of fabricated estimates.
   - Expected: pi stats are authoritative. Conversation-versus-other uses labelled chars/4 estimate; rules/skills/MCP split is not invented. Unknown is not zero; very small nonzero percentages render <1%.
3. **C15.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: pi stats are authoritative. Conversation-versus-other uses labelled chars/4 estimate; rules/skills/MCP split is not invented. Unknown is not zero; very small nonzero percentages render <1%.
4. **C15.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C15.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C15.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C15.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C10 [PARTIAL]:** Pure adapters and agent compaction reducer tests pass. Full native retry/compaction indicators and every button/state transition remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C16 — Blocking extension dialogs

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/dialog.rs::Dialog](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/dialog.rs#L171); [crates/orbit-pi/src/app/dialogs.rs::handle_extension_ui_request](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L10); [crates/orbit-pi/src/app/dialogs.rs::open_dialog](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L205); [crates/orbit-pi/src/app/dialogs.rs::respond_to_dialog](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L236)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C16.behavior [UNTESTED; ROW_SPECIFIC]** Issue each blocking extension dialog: confirm, select, input and editor; submit using buttons and keyboard.
   - Expected: Native select/confirm/input/editor modal; arrows/Enter/Escape, text editing and cancel return exact extension_ui_response ID.
2. **C16.edge [UNTESTED; ROW_SPECIFIC]** Cancel, time out, replace with a newer request and switch session; respond once to the original request id.
   - Expected: Cancel on teardown/switch; if an existing blocking dialog/approval is open, cancel newcomer instead of stranding first. UI request pending can issue background notification.
3. **C16.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Cancel on teardown/switch; if an existing blocking dialog/approval is open, cancel newcomer instead of stranding first. UI request pending can issue background notification.
4. **C16.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C16.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C16.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C16.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S04 [PARTIAL]:** Live custom test requires ORBIT_TERMINAL_TEST and pinned Ghostty; it was run previously. Ordinary default suite skips opt-in test. OS-native menu entry/accelerator dispatch remains separate. [Scoped steps/tests](../native-ui-acceptance.json).

**Owner challenge UI-C11 [PARTIAL]:** General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

