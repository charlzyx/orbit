# Settings, providers, skills, packages and MCP: P09–P12

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## P09 — Quota headline and provider cards

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/app/runtime.rs::poll_quota_entries](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/runtime.rs#L317); [crates/orbit-pi/src/app/mcp_ui.rs::QuotaPopupTab](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/mcp_ui.rs#L2447)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P09.behavior [UNTESTED; ROW_SPECIFIC]** Refresh quota for supported fixture providers and inspect headline/detail cards and reset times.
   - Expected: Top-bar quota pill/popover and provider quota cards show windows/reset times/balance/spend, account notes/errors; active-provider priority, refresh, hidden-provider group and MCP tab.
2. **P09.edge [UNTESTED; ROW_SPECIFIC]** Exercise unsupported, unauthenticated, expired and rate-limited responses; show unknown quota honestly and stop stale polling.
   - Expected: quota.list and bundled orbit:quota session entries merge nonsecret facts. Unknown/unsupported not fabricated. Refresh spinner minimum 450 ms and timeout 4 s; bridge polls 60 s with 3 s bootstrap polls. Hidden IDs persist.
3. **P09.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: quota.list and bundled orbit:quota session entries merge nonsecret facts. Unknown/unsupported not fabricated. Refresh spinner minimum 450 ms and timeout 4 s; bridge polls 60 s with 3 s bootstrap polls. Hidden IDs persist.
4. **P09.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P09.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P09.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P09.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P04 [PARTIAL]:** Pure quota rules pass; live OAuth, full quota popup grouping/hidden-provider/MCP tabs and OS browser-return lifecycle remain unverified or incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P10 — Skills discovery and master/detail

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/skills.rs::discover](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/skills.rs#L88); [crates/orbit-pi/src/skills.rs::Skill](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/skills.rs#L52); [crates/orbit-pi/src/app/skills_ui.rs::render_skills_page](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/skills_ui.rs#L32); [crates/orbit-pi/src/app/skills_ui.rs::apply_skill_action](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/skills_ui.rs#L722)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P10.behavior [UNTESTED; ROW_SPECIFIC]** Discover global/project skills; search, select and inspect metadata/content/location in master-detail view.
   - Expected: Discover project .pi/skills, ancestor .agents/skills up to git root, global pi skills and user .agents/skills; dedupe names by precedence; filter/scope/detail markdown, invoke /skill:name, open/reveal/copy path.
2. **P10.edge [UNTESTED; ROW_SPECIFIC]** Use duplicate names, missing SKILL.md, malformed frontmatter and inaccessible directories; preserve scope attribution.
   - Expected: Missing required description flagged invalid; manual-only distinct from disabled. Discovery is not a package marketplace and exact +/- overrides are supported, not general glob parity with pi loader.
3. **P10.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Missing required description flagged invalid; manual-only distinct from disabled. Discovery is not a package marketplace and exact +/- overrides are supported, not general glob parity with pi loader.
4. **P10.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P10.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P10.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P10.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P05 [PARTIAL]:** Dedicated resourceui replaces generic host implementation and is mounted. Worker source checklist/native/race tests are in progress; do not claim completed acceptance until handoff. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P11 — Skill enable/disable/delete

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/skills.rs::set_enabled](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/skills.rs#L338); [crates/orbit-pi/src/skills.rs::delete](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/skills.rs#L440); [crates/orbit-pi/src/app/skills_ui.rs::delete_selected_skill](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/skills_ui.rs#L772)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P11.behavior [UNTESTED; ROW_SPECIFIC]** Enable/disable a selected skill and request deletion in a disposable copied skill directory.
   - Expected: Toggle scope-correct skills overrides preserving unrelated settings; delete selected skill after confirmation and clear exact override.
2. **P11.edge [UNTESTED; ROW_SPECIFIC]** Cancel deletion, simulate permission failure and refresh externally renamed skills; do not delete outside resolved skill roots.
   - Expected: Deletes skill directory (or only SKILL.md at root), not OS Trash. Do not confuse this with Explorer deletion. Legacy non-array skills schema refuses overwrite.
3. **P11.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Deletes skill directory (or only SKILL.md at root), not OS Trash. Do not confuse this with Explorer deletion. Legacy non-array skills schema refuses overwrite.
4. **P11.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P11.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P11.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P11.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P05 [PARTIAL]:** Dedicated resourceui replaces generic host implementation and is mounted. Worker source checklist/native/race tests are in progress; do not claim completed acceptance until handoff. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## P12 — Plugin package management

Functional **UNTESTED** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/plugins.rs::discover](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/plugins.rs#L99); [crates/orbit-pi/src/plugins.rs::install](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/plugins.rs#L135); [crates/orbit-pi/src/plugins.rs::remove](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/plugins.rs#L147); [crates/orbit-pi/src/plugins.rs::update](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/plugins.rs#L157)

**Fixture:** Isolated HOME and project config, dummy secret marker values, controlled package/provider/MCP executables and fake OAuth issuer; no real credentials.

1. **P12.behavior [UNTESTED; ROW_SPECIFIC]** Install, remove and inspect a fixture plugin package at each supported scope using a controlled package command.
   - Expected: List npm/git/local packages with scope/kind/installed/version, search; install source, project/global scope, update/remove, reveal/open paths.
2. **P12.edge [UNTESTED; ROW_SPECIFIC]** Cancel/fail installation, repeat clicks and return malformed package metadata; keep real package status rather than optimistic success.
   - Expected: pi CLI owns fetch/lockfile/settings changes; 240 s command timeout; malformed one-scope config still lists other scope. Project source wins duplicates; one operation at a time shows progress/result.
3. **P12.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Make this setting's file write or provider/package/MCP operation fail; cancel then retry once after recovery.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: pi CLI owns fetch/lockfile/settings changes; 240 s command timeout; malformed one-scope config still lists other scope. Project source wins duplicates; one operation at a time shows progress/result.
4. **P12.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **P12.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **P12.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **P12.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-P05 [PARTIAL]:** Dedicated resourceui replaces generic host implementation and is mounted. Worker source checklist/native/race tests are in progress; do not claim completed acceptance until handoff. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

