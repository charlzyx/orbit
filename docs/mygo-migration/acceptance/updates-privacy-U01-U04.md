# Usage analytics, update, privacy and release behavior: U01–U04

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## U01 — Usage data/accounting

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/usage/aggregate.rs::build_calendar](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/aggregate.rs#L475); [crates/orbit-pi/src/usage/collect.rs::build_index](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/collect.rs#L225); [crates/orbit-pi/src/usage/collect.rs::scan_store](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/collect.rs#L154)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U01.behavior [UNTESTED; ROW_SPECIFIC]** Scan fixture sessions containing assistant usage/tool calls, cached tokens, reasoning subsets and timestamps.
   - Expected: Scan pi session assistant usage/tool calls/results into normalized index; requests, turns, tool runs, token buckets/cache hit/cost/time/errors.
2. **U01.edge [UNTESTED; ROW_SPECIFIC]** Include malformed records, duplicate IDs, absent price and >30-minute gaps; exclude unavailable/reasoning double-counted metrics.
   - Expected: Only real billed records; reasoning subset of output not extra tokens. Duration gaps >30 min excluded. Pricing coverage exposed; retries absent from store remain unavailable.
3. **U01.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Only real billed records; reasoning subset of output not extra tokens. Duration gaps >30 min excluded. Pricing coverage exposed; retries absent from store remain unavailable.
4. **U01.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U01.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U01.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U01.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U01 [PARTIAL]:** Focused domain/projection cases pass; this does not close the complete analytics corpus or all UI displays. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U02 — Usage overview/detail surfaces

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/usage/page.rs::UsagePage](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L876)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U02.behavior [UNTESTED; ROW_SPECIFIC]** Open Usage in Simple and Details modes at narrow/wide widths; inspect heatmap, charts, KPIs and records.
   - Expected: Primary+U, Simple/Details modes, daily heatmap, summary KPIs, usage-over-time chart/table, signals, token-health panels, provider/model/workspace/tool breakdown and records.
2. **U02.edge [UNTESTED; ROW_SPECIFIC]** Delay/fail index load and refresh empty data; distinguish first load from stale refresh and preserve accurate last-updated state.
   - Expected: Responsive panels use actual main-area width; first load/refresh/no-data/error/last-updated states distinct. Virtual/data-table performance and chart hover tooltips need separate verification.
3. **U02.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: Responsive panels use actual main-area width; first load/refresh/no-data/error/last-updated states distinct. Virtual/data-table performance and chart hover tooltips need separate verification.
4. **U02.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U02.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U02.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U02.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U02 [PARTIAL]:** Native cards/heatmap/chart/breakdowns/session/failure rows render. Source comparison deltas, signals, detailed cache/latency panels and exact layout remain missing; screenshots are fixtures, not Rust comparisons. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U03 — Usage filters and drill-down

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/usage/model.rs::UsageFilter](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/model.rs#L675); [crates/orbit-pi/src/usage/page.rs::set_filter](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L1580); [crates/orbit-pi/src/usage/page.rs::scope_to_day](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L1758)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U03.behavior [UNTESTED; ROW_SPECIFIC]** Apply each date preset/custom range, provider/model/workspace/session/tool filter; drill down and clear filters.
   - Expected: Date presets and custom calendar range, granularity, searchable multi-select workspace/provider/model, select-all/clear, active chips, day/bucket click narrowing, session/errors-only/cached-only focus.
2. **U03.edge [UNTESTED; ROW_SPECIFIC]** Exercise daylight-boundary dates, empty results and combinations; compare retained filters after navigation and relaunch.
   - Expected: One filter model drives every panel; local-time date buckets, inclusive/exclusive boundaries and zero/unknown values preserved. No separate project dimension invented.
3. **U03.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: One filter model drives every panel; local-time date buckets, inclusive/exclusive boundaries and zero/unknown values preserved. No separate project dimension invented.
4. **U03.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U03.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U03.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U03.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U03 [PARTIAL]:** Multi-select menus, scope fields, calendar normalization and contiguous4k guarded buckets implemented. Searchable native multi-select popup and disambiguating identical model IDs acrossproviders remain incomplete. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

## U04 — Usage tables and export

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/usage/page.rs::query_sessions](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L79); [crates/orbit-pi/src/usage/page.rs::export](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L2457); [crates/orbit-pi/src/usage/page.rs::Prefs](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L2528)

**Fixture:** Copied fixture session store, controlled clock/pricing/update responses, dummy redaction markers and disposable staging directory.

1. **U04.behavior [UNTESTED; ROW_SPECIFIC]** Sort each Usage table column, page/load more and export selected data from a fixture dataset.
   - Expected: Search/sort/direction, column visibility/reset/resize, pagination/page sizes across session/series/daily/failure/breakdown tables, context menu and open source session; export CSV/JSON with native Save dialog.
2. **U04.edge [UNTESTED; ROW_SPECIFIC]** Use equal-sort keys, unsupported prices and CSV-special text; export no unrelated rows or secrets and preserve configured page size.
   - Expected: usage.json persists supported view prefs only, not transient filters/data IDs. Export filtered records/aggregates as implemented; escaping, cancellation and write errors surfaced.
3. **U04.error [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Inject malformed data, unavailable I/O/network response and cancellation at the action boundary.
   - Expected: Compare with the linked Rust handler using the same fault. Preserve its documented guard and error state: usage.json persists supported view prefs only, not transient filters/data IDs. Export filtered records/aggregates as implemented; escaping, cancellation and write errors surfaced.
4. **U04.loading [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Pause the fixture dependency at request start; inspect immediately, finish it, then repeat while navigating away and back. For synchronous handlers record that there is no asynchronous loading phase.
   - Expected: Loading/disabled/empty/error transitions must match the pinned handler; no fabricated success, duplicate operation or stale cross-context result. Record baseline timing rather than assume a universal spinner.
5. **U04.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **U04.keyboard [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Invoke the same operation by each registered source chord/menu/palette/button available for this row. Use Tab/Shift+Tab and Escape; repeat from composer, editor, terminal and modal focus. If no binding exists, explicitly record none.
   - Expected: Dispatch, focus restoration and cancellation match source CommandId/BINDINGS and local handler; no unintended prompt, shell bytes or destructive mutation. The positive-action and edge expectations still hold.
7. **U04.repeat [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Perform the primary sequence three times, including a workspace/session change between runs, with delayed callback from the first run.
   - Expected: No duplicate mutations, leaked child/watchers, stale selection or old-context update; current visible state remains the actual underlying state.

**Go mapping:** `internal/preferences`, `internal/workbench`, `internal/native`
**Existing test candidates, not automatic acceptance:** `internal/preferences/store_test.go`, `internal/preferences/workspaces_test.go`, `internal/workbench/workbench_test.go`, `internal/native/native_test.go`, `internal/native/runtime_test.go`, `internal/native/updater_test.go`, `internal/native/temp_files_test.go`

**Executed narrow assertions:**
- None assigned to this scope yet. Last verified SHA: none.

**Owner challenge UI-U04 [PARTIAL]:** Stable sorting/modal/prefs cases pass; JSON/CSV backend implemented and host chooser wired. Independent table states, column visibility/resize/context menus and actual native Save interaction remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [../settings-backend-implementation.md](../settings-backend-implementation.md), [../settings-backend-source-evidence.json](../settings-backend-source-evidence.json)

**Remaining:** Native code exists; this row is not validated end to end.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

