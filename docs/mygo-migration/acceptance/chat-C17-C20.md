# Composer, transcript and interactive agent surfaces: C17–C20

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## C17 — Inline permission approval

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/dialog.rs::ApprovalRequest](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/dialog.rs#L127); [crates/orbit-pi/src/app/dialogs.rs::open_approval](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L260); [crates/orbit-pi/src/app/view.rs::approval_bar](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L4379)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C17.behavior [UNTESTED; ROW_SPECIFIC]** Request each permission kind, inspect requested scope, allow once/session and deny via inline approval.
   - Expected: Access guard requests show tool/detail with Allow once / Always allow this tool / Deny; keyboard selection and Escape deny; MCP tool labels include server.
2. **C17.edge [UNTESTED; ROW_SPECIFIC]** Repeat or cancel pending approval and deliver a stale response; never grant wider scope than selected.
   - Expected: Global Supervised / Auto-accept edits / Full access mode, default Full access. Per-mode allowlist persists. Not a sandbox; extension unavailable is visible and cannot be portrayed as enforced protection.
3. **C17.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Global Supervised / Auto-accept edits / Full access mode, default Full access. Per-mode allowlist persists. Not a sandbox; extension unavailable is visible and cannot be portrayed as enforced protection.
4. **C17.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C17.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C17.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C17.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C11 [PARTIAL]:** General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C18 — Plan/Build/Ask workflow

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/composer_ops.rs::set_workflow_mode](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L891)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C18.behavior [UNTESTED; ROW_SPECIFIC]** Switch Plan, Build and Ask modes; start and continue a turn in each; inspect resulting extension state.
   - Expected: Composer mode chip, keyboard picker and palette action; Plan/Ask scope mutation tools and shell access, Build restores only workflow-removed tools.
2. **C18.edge [UNTESTED; ROW_SPECIFIC]** Reject unsupported protocol capability honestly; cancellation and mode changes must not silently broaden execution permission.
   - Expected: Per-session workflow.json is read live by extension and custom entries record mode; prune deleted-session keys. Default Build. This is extension policy, not OS isolation.
3. **C18.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Per-session workflow.json is read live by extension and custom entries record mode; prune deleted-session keys. Default Build. This is extension policy, not OS isolation.
4. **C18.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C18.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C18.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C18.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C11 [PARTIAL]:** General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C19 — Structured ask questionnaire

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/ask.rs::AskPrompt](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/ask.rs#L70); [crates/orbit-pi/src/ask.rs::AskReplayState](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/ask.rs#L140); [crates/orbit-pi/src/app/view.rs::ask_panel](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3982); [crates/orbit-pi/src/transcript_view.rs::render_ask_card](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L2928)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C19.behavior [UNTESTED; ROW_SPECIFIC]** Present a structured questionnaire with single, multiple and freeform answers; navigate and submit.
   - Expected: ask_user_question inline panel supports multiple questions, single/multi select, custom text, previews, previous/next, final submit and cancel; finished transcript question card.
2. **C19.edge [UNTESTED; ROW_SPECIFIC]** Test required unanswered question, Other, back, cancellation and stale request replacement; preserve exact answer structure.
   - Expected: Buffer answers locally until whole questionnaire commits, then replay extension select/input requests; retain prior answer when navigating back; reject unexpected replay sequences.
3. **C19.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Buffer answers locally until whole questionnaire commits, then replay extension select/input requests; retain prior answer when navigating back; reject unexpected replay sequences.
4. **C19.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C19.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C19.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C19.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C11 [PARTIAL]:** General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Transport/state support is not proof of image drag/paste, structured asks or custom UI/widget rendering parity.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## C20 — Extension widgets and custom UI

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/dialogs.rs::apply_extension_widget](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L76); [crates/orbit-pi/src/app/dialogs.rs::apply_custom_frame](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/dialogs.rs#L114)

**Fixture:** Isolated HOME, deterministic fixture Pi RPC stream, long transcript/Unicode/image fixtures and no live provider inference.

1. **C20.behavior [UNTESTED; ROW_SPECIFIC]** Emit extension notification, status, widget and custom UI events at all source placements.
   - Expected: setWidget styled ANSI/SGR text above/below composer; replace/clear by key; custom open/render/close stack with monospace terminal frames, cursor, input and resize notifications. notify toast, setStatus line, set_editor_text composer.
2. **C20.edge [UNTESTED; ROW_SPECIFIC]** Replace/remove keyed widgets, emit unsupported UI payloads and end a session; do not keep stale cross-session content.
   - Expected: Custom UI is capability/patch-dependent; advertises 96×20 viewport, scrolls oversized content. Legacy keyboard encoder only; Kitty mode metadata is recorded but not fully implemented. setTitle and widget component factories have no equivalent.
3. **C20.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Delay, reject, cancel or disconnect the fixture RPC operation; send a late event for the previous request/session.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Custom UI is capability/patch-dependent; advertises 96×20 viewport, scrolls oversized content. Legacy keyboard encoder only; Kitty mode metadata is recorded but not fully implemented. setTitle and widget component factories have no equivalent.
4. **C20.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **C20.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **C20.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **C20.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/pirpc`, `internal/agent`
**Existing test candidates, not automatic acceptance:** `internal/pirpc/client_test.go`, `internal/agent/manager_test.go`, `internal/agent/reducer_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-S04 [PARTIAL]:** Native menu guard includes module-owned Git/Worktrees/MCP/Resources/report review/preview and Usage filter overlays. CPU routing tests pass; actual platform accelerator ordering remains separate. [Scoped steps/tests](../native-ui-acceptance.json).

**Owner challenge UI-C11 [PARTIAL]:** General modal and ANSI custom bridge exist. Structured questionnaire, specialized inline approval/replay, widget placement and unavailable workflow/access presentation are not fully wired. No full extension parity claim. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Transport/state support is not proof of image drag/paste, structured asks or custom UI/widget rendering parity.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

