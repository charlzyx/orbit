# Composer, transcript and interactive agent surfaces: C09–C12

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## C09 — Rich transcript text and code

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript_view.rs::parse_blocks](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L5839); [crates/orbit-pi/src/transcript_view.rs::parse_inline](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L5201); [crates/orbit-pi/src/transcript_view.rs::render_block](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L6177); [crates/orbit-pi/src/transcript_view.rs::render_code_block](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L6537)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C09.behavior [UNTESTED; ROW_SPECIFIC]** Render nested Markdown, tables, links, code fences in all supported languages and streaming incomplete fences.
   - Expected: Headings, paragraphs, emphasis/strike/inline code, links/autolinks, quote/alerts, ordered/unordered/task lists, tables, code fences and paint-only syntax highlighting; copy code and whole response.
2. **C09.edge [UNTESTED; ROW_SPECIFIC]** Grow a code block one Unicode prefix at a time; highlighting must alter paint only, not text width, wrapping or scroll anchors.
   - Expected: Mermaid remains explicitly labelled copyable source, not diagram rendering. Markdown images have loading/fallback and URL normalization. Lists/tables need identical wrapping/column constraints.
3. **C09.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Mermaid remains explicitly labelled copyable source, not diagram rendering. Markdown images have loading/fallback and URL normalization. Lists/tables need identical wrapping/column constraints.
4. **C09.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C09.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C09.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C09.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C07 [PARTIAL]:** GFM alert/geometry sample passes; highlight tokenizer has1204 source-oracle cases. Whole source markdown corpus, per-message media/actions and exact spacing remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C10 — Transcript selection and context menu

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript_view.rs::TextSelection](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L302); [crates/orbit-pi/src/transcript_view.rs::text_selection_menu](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L912); [crates/orbit-pi/src/transcript.rs::selected_text](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L1076); [crates/orbit-pi/src/transcript.rs::mark_copied](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript.rs#L2271)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C10.behavior [UNTESTED; ROW_SPECIFIC]** Select across transcript rows; copy through keyboard and context menu; invoke available row actions.
   - Expected: Selectable text across rendered blocks, drag selection, copy feedback, right-click text menu, clickable links and image lightbox. Primary+Shift+C copies newest assistant response.
2. **C10.edge [UNTESTED; ROW_SPECIFIC]** Right-click a link, code block and tool row near viewport edges; preserve exact copied text and no unintended agent action.
   - Expected: UTF-8/grapheme boundaries, virtualized visible blocks and selection registration must remain sound; copy control success feedback expires.
3. **C10.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: UTF-8/grapheme boundaries, virtualized visible blocks and selection registration must remain sound; copy control success feedback expires.
4. **C10.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C10.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C10.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C10.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C08 [PARTIAL]:** Typed tools, logical turns, observed-live elapsed time, source tail/usage footer and shared user/Markdown/alert selection are mounted and native-tested. C11 remains a concrete framework gap: selection clears when a virtualized endpoint unmounts; Rust retains SelectedBlock snapshots. Do not equate mounted-row selection with full parity. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C11 — Thinking/tool activity and details

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript_view.rs::render_activity_group](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2387); [crates/orbit-pi/src/transcript_view.rs::render_tool_detail](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3551); [crates/orbit-pi/src/transcript_view.rs::render_tool_error_strip](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3512); [crates/orbit-pi/src/transcript_view.rs::build_edit_diff](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L3973)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C11.behavior [UNTESTED; ROW_SPECIFIC]** Stream thinking and tool start/update/end/error events; expand/collapse raw details and nested content.
   - Expected: Fold activity chains; live thinking scroll/tail; distinct tool icons/action preview, invocation arguments/result, error/truncation badges, code body, file edit inline diff and changed-file links.
2. **C11.edge [UNTESTED; ROW_SPECIFIC]** Show missing tool results, large output and interrupted tools honestly; switching rows must not reuse another tool's detail.
   - Expected: Tool output caps and truncated indicator preserve honesty; read/write/edit/bash/search/web/MCP and unknown/custom tool paths have meaningful fallbacks; errors remain visible after turn settles.
3. **C11.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Tool output caps and truncated indicator preserve honesty; read/write/edit/bash/search/web/MCP and unknown/custom tool paths have meaningful fallbacks; errors remain visible after turn settles.
4. **C11.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C11.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C11.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C11.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C08 [PARTIAL]:** Typed tools, logical turns, observed-live elapsed time, source tail/usage footer and shared user/Markdown/alert selection are mounted and native-tested. C11 remains a concrete framework gap: selection clears when a virtualized endpoint unmounts; Rust retains SelectedBlock snapshots. Do not equate mounted-row selection with full parity. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C12 — Turn summaries and metrics

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/transcript_view.rs::render_working_indicator](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L4820); [crates/orbit-pi/src/transcript_view.rs::render_message_footer](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L4369); [crates/orbit-pi/src/transcript_view.rs::render_summary_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1811); [crates/orbit-pi/src/transcript_view.rs::render_changed_files](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C12.behavior [UNTESTED; ROW_SPECIFIC]** Complete turns with usage, latency and file changes; inspect per-turn summary and open its Review target.
   - Expected: Working elapsed time, output rate, usage footer/breakdown, added/removed counts, final changed-files summary card and summary click-through to review.
2. **C12.edge [UNTESTED; ROW_SPECIFIC]** Return absent usage or failed checkpoints; do not invent cost, diff or elapsed values; resumed sessions retain turn attribution.
   - Expected: Distinguish estimated/live rates from final provider usage, stopped from failed; avoid showing generated metrics when source does not supply them.
3. **C12.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Distinguish estimated/live rates from final provider usage, stopped from failed; avoid showing generated metrics when source does not supply them.
4. **C12.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C12.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C12.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C12.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/ui`, `cmd/orbit/main.go`
**Existing test candidates, not automatic acceptance:** `internal/ui/snapshot_test.go`, `internal/ui/workspace_review_regression_test.go`, `internal/ui/viewer_parity_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C08 [PARTIAL]:** Typed tools, logical turns, observed-live elapsed time, source tail/usage footer and shared user/Markdown/alert selection are mounted and native-tested. C11 remains a concrete framework gap: selection clears when a virtualized endpoint unmounts; Rust retains SelectedBlock snapshots. Do not equate mounted-row selection with full parity. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

