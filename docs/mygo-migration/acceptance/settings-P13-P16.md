# Settings, providers, skills, packages and MCP: P13–P16

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## P13 — Plugin update offers

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/plugins.rs::find_updates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/plugins.rs#L165); [crates/orbit-pi/src/app/plugin_update_ui.rs::start_plugin_update_check](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/plugin_update_ui.rs#L19); [crates/orbit-pi/src/app/plugin_update_ui.rs::skip_plugin_updates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/plugin_update_ui.rs#L52); [crates/orbit-pi/src/app/plugin_update_ui.rs::apply_plugin_updates](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/plugin_update_ui.rs#L69)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P13.behavior [UNTESTED; ROW_SPECIFIC]** Provide an available plugin update; inspect, apply and dismiss its offer.
   - Expected: Launch-delayed and manual Refresh upstream check; show Update now / skip current versions; batch outcomes and later newer releases.
2. **P13.edge [UNTESTED; ROW_SPECIFIC]** Fail checking/updating and test no-update result; ensure update notification and disabled/loading actions settle correctly.
   - Expected: 5 s launch delay; PI_OFFLINE skips; 12 s probes, max four parallel. Pinned npm/local sources excluded. plugin-update-skips.json stores source→version. Results tied to checked workspace.
3. **P13.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: 5 s launch delay; PI_OFFLINE skips; 12 s probes, max four parallel. Pinned npm/local sources excluded. plugin-update-skips.json stores source→version. Results tied to checked workspace.
4. **P13.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P13.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P13.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P13.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P05 [PARTIAL]:** Dedicated resourceui source master-detail and CRUD/update/skip lifecycle is mounted, including Offline flag, default-app open, Notify completion and global update notice below toasts.24 scoped package tests/race pass. Host/platform full flow and matched pixels remain unverified; see resource module report. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../resource-ui-implementation.md](../resource-ui-implementation.md), [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P14 — MCP manager and server table

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/mcp/manager.rs::McpManager](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/mcp/manager.rs#L74); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1135); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_handle_key](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L1272)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P14.behavior [UNTESTED; ROW_SPECIFIC]** Open MCP manager with global/project/local servers; sort/select/refresh and inspect status/tool counts.
   - Expected: Global/project-scoped server list/filter, detailed transport/status/tool counts/resources, add/edit/toggle/test/reconnect/remove, palette per-server actions and top-bar summary.
2. **P14.edge [UNTESTED; ROW_SPECIFIC]** Provide malformed config and absent transport process; show origin and diagnostic without dropping unrelated server settings.
   - Expected: No Orbit MCP implementation: pi executes and probes. Project names shadow global. Config and live runtime state are separate; errors, untrusted-project note and codemode reachability warning are explicit.
3. **P14.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: No Orbit MCP implementation: pi executes and probes. Project names shadow global. Config and live runtime state are separate; errors, untrusted-project note and codemode reachability warning are explicit.
4. **P14.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P14.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P14.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P14.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P06 [PARTIAL]:** mcpui/workbench scoped race tests and19-case checklist pass. Host mount/deferred reload exists. Real OAuth and authoritative applied/readiness callback are not verified; Applied is deliberately not called on async restart acceptance. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../mcp-ui-implementation.md](../mcp-ui-implementation.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P15 — MCP add/edit fields and exposure

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/mcp.rs::McpTransport](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/mcp.rs#L160); [crates/orbit-pi/src/mcp.rs::McpExposure](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/mcp.rs#L72); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_editor_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L523)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P15.behavior [UNTESTED; ROW_SPECIFIC]** Create/edit stdio and HTTP servers, arguments/env/headers and exposure restrictions; save and reread config.
   - Expected: Stdio command/args/env/cwd; streamable HTTP URL/headers; description/scope/timeout/enabled/tool exposure. Exposures codemode/codemode-deferred/deferred/direct/hidden; per-tool overrides.
2. **P15.edge [UNTESTED; ROW_SPECIFIC]** Reject duplicate/invalid names, unsafe secret entries and incompatible fields; cancel without mutating the fixture files.
   - Expected: Validate names, duplicate/move/rename scope and transport fields; atomic surgical merge preserves unknown fields/unrelated servers. Invalid config cannot be overwritten by empty defaults.
3. **P15.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Validate names, duplicate/move/rename scope and transport fields; atomic surgical merge preserves unknown fields/unrelated servers. Invalid config cannot be overwritten by empty defaults.
4. **P15.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P15.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P15.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P15.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P06 [PARTIAL]:** mcpui/workbench scoped race tests and19-case checklist pass. Host mount/deferred reload exists. Real OAuth and authoritative applied/readiness callback are not verified; Applied is deliberately not called on async restart acceptance. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../mcp-ui-implementation.md](../mcp-ui-implementation.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P16 — MCP secrets and auth

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/mcp_ui.rs::mcp_sign_in](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L797); [crates/orbit-pi/src/app/mcp_ui.rs::mcp_sign_out](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L866)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P16.behavior [UNTESTED; ROW_SPECIFIC]** Use dummy secret references in MCP config; reveal/mask/update/delete through the source workflow.
   - Expected: Literal secret form values become ${NAME} refs backed by local secrets, masks round-trip unchanged; OAuth login/cancel/logout via pi; launch and probe receive secret env.
2. **P16.edge [UNTESTED; ROW_SPECIFIC]** Missing secret, failed secret-store write and OAuth cancellation must not leak values or persist a false authenticated state.
   - Expected: mcp-secrets.json Unix 0600, namespace-safe unused-secret pruning, log/error redaction. No keychain encryption claim. Probe timeout 120 s; login 300 s + 30 s cleanup grace.
3. **P16.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: mcp-secrets.json Unix 0600, namespace-safe unused-secret pruning, log/error redaction. No keychain encryption claim. Probe timeout 120 s; login 300 s + 30 s cleanup grace.
4. **P16.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P16.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P16.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P16.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P06 [PARTIAL]:** mcpui/workbench scoped race tests and19-case checklist pass. Host mount/deferred reload exists. Real OAuth and authoritative applied/readiness callback are not verified; Applied is deliberately not called on async restart acceptance. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../mcp-ui-implementation.md](../mcp-ui-implementation.md)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

