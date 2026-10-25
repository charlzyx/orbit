# Composer, transcript and interactive agent surfaces: C05–C08

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## C05 — Image attachments and drag/paste

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/composer_ops.rs::attach_image](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L435); [crates/orbit-pi/src/app/composer_ops.rs::on_file_drop](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L405); [crates/orbit-pi/src/app/events.rs::drain_pasted_images](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L1108); [crates/orbit-pi/src/attachments.rs::compress_image](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/attachments.rs#L55)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C05.behavior [UNTESTED; ROW_SPECIFIC]** Attach images via picker, clipboard and drag/drop; remove one and submit another to a fixture Pi receiver.
   - Expected: Async native picker, external file drop, clipboard image paste, removable previews/chips, attached image lightbox; images travel with prompt/follow_up/steer.
2. **C05.edge [UNTESTED; ROW_SPECIFIC]** Reject unsupported/unreadable/oversized data without losing text; cancel picker and fail sending while attachments remain selected.
   - Expected: Maximum eight images. PNG/JPEG compression targets 1400 px and 512 KiB base64, transparency uses WebP; caps are best effort, GIF/SVG/TIFF pass through unchanged. Source/decoding failures must not crash.
3. **C05.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Maximum eight images. PNG/JPEG compression targets 1400 px and 512 KiB base64, transparency uses WebP; caps are best effort, GIF/SVG/TIFF pass through unchanged. Source/decoding failures must not crash.
4. **C05.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C05.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C05.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C05.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C04 [PARTIAL]:** Attachment Prepare and transport domain tests exist; host native add/drop/clipboard/partial-failure end-to-end tests and source text-file mention semantics remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Transport/state support is not proof of image drag/paste, structured asks or custom UI/widget rendering parity.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C06 — Model and thinking selectors

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/model_selector.rs::ModelSelector](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L159); [crates/orbit-pi/src/app/pickers.rs::sync_model_selector](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/pickers.rs#L761); [crates/orbit-pi/src/app/session.rs::can_choose_thinking](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L52)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C06.behavior [UNTESTED; ROW_SPECIFIC]** Open model/provider and thinking menus; select available models, cycle thinking level and start a prompt.
   - Expected: Searchable model picker with provider groups, current model, favorites and recent sections; provider scope, favorite toggle, metadata and keyboard scrolling. Reasoning levels are runtime-provided.
2. **C06.edge [UNTESTED; ROW_SPECIFIC]** Return empty model lists, unavailable current model and failed set_model response; preserve real server state and clear loading.
   - Expected: No hardcoded available-model truth. Thinking picker hidden/disabled for empty or [off]-only levels. Favorites/recent stores survive restart and model provider/id form identity.
3. **C06.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: No hardcoded available-model truth. Thinking picker hidden/disabled for empty or [off]-only levels. Favorites/recent stores survive restart and model provider/id form identity.
4. **C06.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C06.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C06.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C06.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C05 [PARTIAL]:** Source modelui catalog, filters/favorites, grouped model/thinking popovers and Agent/title searchable defaults are now mounted. Authoritative raw model capabilities retained. Configured update --models action serializes, cancels, restarts and resumes saved session with an explicit zero-inference RPC fixture. CPU captures inspected; exact native pixels and OS interaction remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C07 — Stream reducer and snapshot equivalence

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript.rs::load_from](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L1081); [crates/orbit-pi/src/transcript.rs::apply_event](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L1181); [crates/orbit-pi/src/app/events.rs::apply_messages_snapshot](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/events.rs#L645)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C07.behavior [UNTESTED; ROW_SPECIFIC]** Replay recorded text/thinking/tool deltas and an equivalent final snapshot; compare transcript state.
   - Expected: Snapshot get_messages and live text/thinking/tool events feed one transcript model; per-user-turn fold, tool-result association, custom messages and compaction summary cards.
2. **C07.edge [UNTESTED; ROW_SPECIFIC]** Replay duplicated, out-of-order, partial and late old-session events; malformed records must not corrupt other rows.
   - Expected: Recover partial/missing live events and process stop; unknown payloads do not panic. Avoid duplicate optimistic user messages when snapshot arrives.
3. **C07.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Recover partial/missing live events and process stop; unknown payloads do not panic. Avoid duplicate optimistic user messages when snapshot arrives.
4. **C07.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C07.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C07.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C07.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C06 [PARTIAL]:** Native variable-height follow/read-position test passes. Logical assistant turns now project across tool-result messages. Source navigation rail and full long-session performance/OS acceptance remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C08 — Virtualized scrolling and navigation rail

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript.rs::jump_turn](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L2278); [crates/orbit-pi/src/transcript_view.rs::render_navigation_rail](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1396); [crates/orbit-pi/src/transcript.rs::persist_hint](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L877)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C08.behavior [UNTESTED; ROW_SPECIFIC]** Render thousands of transcript rows; scroll away from bottom while streaming; use jump-to-bottom and navigation rail.
   - Expected: Tail-follow while at bottom, scroll-away inspection, Jump to latest, previous/next user turn, clickable rail markers/hover previews and one-time rail hint.
2. **C08.edge [UNTESTED; ROW_SPECIFIC]** Append tokens above and below the viewport, expand tool details and switch sessions; preserve the intended scroll anchor.
   - Expected: Streaming row remeasurement must keep stable scroll behavior. Rail hint dismissal persists; no full transcript rerender cost for every token.
3. **C08.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Streaming row remeasurement must keep stable scroll behavior. Rail hint dismissal persists; no full transcript rerender cost for every token.
4. **C08.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C08.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C08.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C08.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C06 [PARTIAL]:** Native variable-height follow/read-position test passes. Logical assistant turns now project across tool-result messages. Source navigation rail and full long-session performance/OS acceptance remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

