# Settings, providers, skills, packages and MCP: P05–P08

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## P05 — Provider catalog and filtering

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::provider_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::ensure_provider_metadata](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7639); [crates/orbit-pi/src/app/settings.rs::provider_refresh](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7665)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P05.behavior [UNTESTED; ROW_SPECIFIC]** Browse providers, filter by configured/available/name and inspect capability/status indicators.
   - Expected: Discover provider metadata from installed pi-ai with curated fallback, provider/model counts and marks, connected/auth states, search/filter cards.
2. **P05.edge [UNTESTED; ROW_SPECIFIC]** Use providers with no models, duplicate names and unavailable credentials; clearing filter restores the full list.
   - Expected: Do not hardcode author-specific accounts/models. Unknown provider ID is humanized with fallback glyph. Refresh runs off-thread and malformed file is an error, not an empty successful reset.
3. **P05.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Do not hardcode author-specific accounts/models. Unknown provider ID is humanized with fallback glyph. Refresh runs off-thread and malformed file is an error, not an empty successful reset.
4. **P05.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P05.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P05.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P05.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P03 [PARTIAL]:** These native/disk fixtures pass. Full provider card geometry, source account detail/quota embed and real credential flows remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P06 — API credentials and custom providers

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::provider_key_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7772); [crates/orbit-pi/src/app/settings.rs::provider_save](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L8027); [crates/orbit-pi/src/app/settings.rs::provider_remove](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L8103)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P06.behavior [UNTESTED; ROW_SPECIFIC]** Add, edit and remove a fixture custom provider with dummy credentials; validate URL/model fields.
   - Expected: Masked API-key editor, provider CRUD including ID/base URL/API family/model list, save/remove, refresh runtime credentials; special Ollama Cloud session input is stored separately.
2. **P06.edge [UNTESTED; ROW_SPECIFIC]** Cancel deletion; reject invalid config and failed saves; redact dummy secret markers from logs, screenshots and errors.
   - Expected: Writes preserve unrelated models.json/auth.json fields and per-model fields by ID. Invalid JSON/empty keys rejected rather than clobbered. Credentials remain local; avoid logging raw secret.
3. **P06.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Writes preserve unrelated models.json/auth.json fields and per-model fields by ID. Invalid JSON/empty keys rejected rather than clobbered. Credentials remain local; avoid logging raw secret.
4. **P06.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P06.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P06.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P06.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P03 [PARTIAL]:** These native/disk fixtures pass. Full provider card geometry, source account detail/quota embed and real credential flows remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P07 — Provider OAuth lifecycle

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/auth.rs::AuthManager](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/auth.rs#L140); [crates/orbit-pi/src/app/runtime.rs::handle_auth_effects](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L395); [crates/orbit-pi/src/app/settings.rs::provider_oauth_login](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L7805); [crates/orbit-pi/src/platform.rs::open_terminal_command](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/platform.rs#L957)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P07.behavior [UNTESTED; ROW_SPECIFIC]** Start OAuth using a local fixture issuer, open/paste the returned URL and complete or cancel the flow.
   - Expected: Auth capability/status probe, sign-in browser/device flow, cancel, timeout/restart recovery, sign-out; supported namespace uses auth.* RPC.
2. **P07.edge [UNTESTED; ROW_SPECIFIC]** Delay success until after cancellation/restart; old callbacks must not authenticate the new request; never use real accounts in automated tests.
   - Expected: State reducer contains nonsecret display state only. Unsupported auth RPC falls back to supported terminal login path; failed opening or cancellation surfaced. Never claim sign-in from an opened URL alone.
3. **P07.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: State reducer contains nonsecret display state only. Unsupported auth RPC falls back to supported terminal login path; failed opening or cancellation surfaced. Never claim sign-in from an opened URL alone.
4. **P07.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P07.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P07.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P07.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P04 [PARTIAL]:** Pure quota rules pass; live OAuth, full quota popup grouping/hidden-provider/MCP tabs and OS browser-return lifecycle remain unverified or incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P08 — Models catalog page

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/settings.rs::model_rows](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1433); [crates/orbit-pi/src/app/settings.rs::models_refresh](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1369); [crates/orbit-pi/src/app/settings.rs::model_toolbar](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L1082)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P08.behavior [UNTESTED; ROW_SPECIFIC]** List models by provider; filter/search; toggle supported availability/default fields and inspect metadata.
   - Expected: Full model listing, search name/ID/provider, provider filter, favorites-only, favorite cards, active model chooser and catalog refresh.
2. **P08.edge [UNTESTED; ROW_SPECIFIC]** Return loading/error/empty results and a missing currently-selected model; preserve actual provider state.
   - Expected: Refresh blocked while run active; runtime/pi catalog is actual source. No results differs from no models reported; active selection agrees with composer picker.
3. **P08.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Refresh blocked while run active; runtime/pi catalog is actual source. No results differs from no models reported; active selection agrees with composer picker.
4. **P08.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P08.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P08.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P08.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-C05 [PARTIAL]:** Source modelui catalog, filters/favorites, grouped model/thinking popovers and Agent/title searchable defaults are now mounted. Authoritative raw model capabilities retained. Configured update --models action serializes, cancels, restarts and resumes saved session with an explicit zero-inference RPC fixture. CPU captures inspected; exact native pixels and OS interaction remain unverified. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

