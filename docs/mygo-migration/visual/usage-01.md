# Audited visual candidates: usage

Numbered candidates, not validated screenshot recipes. State validity, fixture execution, capture, comparison and parity are separate gates.

## V260 · Usage simple top
- Features: U01, U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Header Export and Refresh; Range, Workspace, Provider, Model controls; Simple selected; Daily activity heatmap; Summary board: Requests, Total tokens, Cache hit rate, Avg response; Summary secondary strip.
- Setup: Open Usage, Simple, Last 7 days; keep scroll at top. Capture full client; do not claim trend if below viewport.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V260. Forbidden claims: All lower cards simultaneously visible.
- Planned pair: `../probes/visual/V260/rust.png` ↔ `../probes/visual/V260/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V261 · Usage simple lower health
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Source-generated insight rows; Input/output/cache-read/cache-write composition; Cache hit-rate details and Show cached only; Hit-rate trend when at least two readable buckets.
- Setup: Simple, Last 7 days. Scroll so Signals and Token health are visible; if their complete bounds exceed client height use separate appended capture.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V261. Forbidden claims: Signals for arbitrary data: conditional on insights; Always-present token health: conditional on tokens/cache.
- Planned pair: `../probes/visual/V261/rust.png` ↔ `../probes/visual/V261/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V262 · Usage details models
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Models selected among Models/Workspaces/Providers/Tools; Ranked model bars; Search and Columns; Model rows, totals and paging.
- Setup: Details; Breakdown Models. Scroll to Breakdown heading.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V262. Forbidden claims: Every optional column at every width.
- Planned pair: `../probes/visual/V262/rust.png` ↔ `../probes/visual/V262/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V263 · Usage details workspaces
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Workspaces selected; Ranked workspace bars; Search, Columns, table, totals, paging.
- Setup: Details; Breakdown Workspaces. Scroll to Breakdown heading.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V263. Forbidden claims: .
- Planned pair: `../probes/visual/V263/rust.png` ↔ `../probes/visual/V263/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V264 · Usage details providers
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Providers selected; Ranked provider bars; Search, Columns, table, totals, paging.
- Setup: Details; Breakdown Providers. Scroll to Breakdown heading.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V264. Forbidden claims: .
- Planned pair: `../probes/visual/V264/rust.png` ↔ `../probes/visual/V264/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V265 · Usage details tools
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Tools selected; Ranked tool call bars; Search, Columns, tool-specific columns, totals, paging.
- Setup: Details; Breakdown Tools. Scroll to Breakdown heading.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V265. Forbidden claims: Reuse request/token columns as tool schema.
- Planned pair: `../probes/visual/V265/rust.png` ↔ `../probes/visual/V265/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V266 · Usage records sessions
- Features: U02, U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Sessions selected; Search, Columns; Responsive source column plan including trailing Open; Totals distinguish filtered set and current page; Page-size and navigation controls.
- Setup: Details; Records Sessions. Scroll Records heading to top of content.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V266. Forbidden claims: All 13 enum columns forced visible: responsive plan has 12 named columns plus Open; Tools enum not admitted.
- Planned pair: `../probes/visual/V266/rust.png` ↔ `../probes/visual/V266/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V267 · Usage records daily
- Features: U02, U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Daily selected; Date/day bucket schema; Independent search, Columns, totals, page-size and pager.
- Setup: Details; Records Daily; Last 7 days. Scroll Records heading to top.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V267. Forbidden claims: All-time necessarily daily: granularity adapts to range.
- Planned pair: `../probes/visual/V267/rust.png` ↔ `../probes/visual/V267/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V268 · Usage records failures
- Features: U02, U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Failures selected; Provider/tool/stopped metadata; When/session/provider/model/type/message source columns; Search, Columns, totals, pager and retry explanation.
- Setup: Details; Records Failures. Scroll Records heading to top.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V268. Forbidden claims: Failure-row detail modal: no source click handler.
- Planned pair: `../probes/visual/V268/rust.png` ↔ `../probes/visual/V268/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V269 · Usage loading first
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Source skeleton cards; Header and range/mode controls; Disabled export.
- Setup: Hold initial scan result before publication; open Usage Simple. Capture before first index arrives.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V269. Forbidden claims: Fabricated numeric data.
- Planned pair: `../probes/visual/V269/rust.png` ↔ `../probes/visual/V269/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V270 · Usage empty store
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: No usage data yet; Hint names isolated store; Disabled export.
- Setup: Scan empty isolated session directory successfully; open Usage.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V270. Forbidden claims: Numeric zeros pretending observed usage; Body retry.
- Planned pair: `../probes/visual/V270/rust.png` ↔ `../probes/visual/V270/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V271 · Usage unreadable store
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Unable to read usage data; Unreadable count and store hint; Header Refresh.
- Setup: Publish empty index with unreadable_files=2, representing two malformed/unreadable fixture files.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V271/rust.png` ↔ `../probes/visual/V271/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V271. Forbidden claims: Body Retry/Try again: not rendered here; First-load scan-error branch.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.
## V272 · Usage filtered empty
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: No usage matches these filters; Active narrowing chips retained; Clear filters action; Disabled export.
- Setup: Select provider A and workspace B that have no intersection, but each exists globally.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V272. Forbidden claims: No usage data yet.
- Planned pair: `../probes/visual/V272/rust.png` ↔ `../probes/visual/V272/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V273 · Usage refresh prior data
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Previous statistics remain; Header Refreshing status and button; Existing filters and chart remain.
- Setup: First publish core index; request refresh and hold next scan result.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V273. Forbidden claims: First-load skeleton replacing prior data.
- Planned pair: `../probes/visual/V273/rust.png` ↔ `../probes/visual/V273/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V274 · Usage refresh error
- Features: U02. Lane: `canonical`. Visual: **BLOCKED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **UNREACHABLE_SOURCE_PRODUCTION**. usage/collect.rs83–146 and page.rs1206: source scanner cannot produce the proposed prior-data error state.
- Expected: NOT A REACHABLE PRODUCTION RUST STATE; The scanner sends UsageIndex, not Result; no previous-data refresh-failure branch with Retry exists.
- Setup: Open the isolated usage surface in fixture.json. Configure the named source state: Usage refresh error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V274/rust.png` ↔ `../probes/visual/V274/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V274. Forbidden claims: Refresh failure banner while prior data remains; Retry button alongside existing data.
- Blocker: Excluded from source-parity capture until repurposed or superseded: source production state is unreachable; Go-only extra UI cannot establish Rust parity.
## V275 · Usage removed session notice
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Header: That session is no longer in the store — scope cleared; Session chip removed; Remaining usage shown.
- Setup: Scope last session in index; remove that session from isolated store; refresh. Capture within six seconds after new index publishes; keep other sessions present.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V275. Forbidden claims: Missing-session modal; Open-session error fabricated by this Usage handler.
- Planned pair: `../probes/visual/V275/rust.png` ↔ `../probes/visual/V275/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V276 · Usage active filter chips
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Provider, model, workspace removable chips; Separate time-focus removable chip; Clear all; Range selector remains above.
- Setup: Select matching provider/model/workspace values and click a populated trend bucket. Keep active chips near viewport top.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V276. Forbidden claims: Ordinary range is a removable date chip; Error and cache chips simultaneously.
- Planned pair: `../probes/visual/V276/rust.png` ↔ `../probes/visual/V276/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V277 · Usage error only
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Failed requests chip; Error-scoped values; No Cached only chip.
- Setup: Click failed-request action in Summary secondary strip; keep active filters visible.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V277. Forbidden claims: Generic Filters checkbox menu absent from Rust.
- Planned pair: `../probes/visual/V277/rust.png` ↔ `../probes/visual/V277/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V278 · Usage cache only
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Cached only chip; No Failed requests chip; When health action visible it reads Show all requests.
- Setup: In Token health click Show cached only; capture active-filter region, or separately append health action view.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V278. Forbidden claims: Generic Filters menu; Active chip and distant health action guaranteed in same viewport.
- Planned pair: `../probes/visual/V278/rust.png` ↔ `../probes/visual/V278/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V279 · Usage export menu
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: CSV filtered requests action plus explanatory subtitle; JSON current view action plus explanatory subtitle.
- Setup: Click header Export, leave popover open.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V279. Forbidden claims: Native save dialog already opened.
- Planned pair: `../probes/visual/V279/rust.png` ↔ `../probes/visual/V279/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V280 · Usage session context menu
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/view.rs::body:L622](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/view.rs#L622)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Open session; Scope to session; Copy session ID; Filter by provider; Filter by model; Filter by workspace; Source separators.
- Setup: Details Sessions; right-click an unscoped session with a known model; do not left-click first.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V280. Forbidden claims: Right-click automatically selects/scopes session.
- Planned pair: `../probes/visual/V280/rust.png` ↔ `../probes/visual/V280/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V281 · Usage date preset menu
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::range_menu:L863](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L863)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Today, Yesterday, Last 7 days, Last 14 days, Last 30 days, This week, This month, Previous month, Custom, All time; Current selection mark; Source comparison hints.
- Setup: Click date range selector with Last 7 days current.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V281. Forbidden claims: Calendar shown before Custom chosen.
- Planned pair: `../probes/visual/V281/rust.png` ↔ `../probes/visual/V281/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V282 · Usage custom calendar start
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::calendar:L921](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L921)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Preset list remains; Month previous/next; Monday-first weekday row; Day12 selected, no end; Pick the end date.
- Setup: Open range, Custom; navigate to October2026; click day12 once.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V282. Forbidden claims: Separate modal; Apply or Clear controls.
- Planned pair: `../probes/visual/V282/rust.png` ↔ `../probes/visual/V282/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V283 · Usage custom calendar range
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::calendar:L921](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L921)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Start/end day fills; In-range shading; Selected range label footer; Month navigation.
- Setup: Choose October12 then October18: second click auto-applies and closes. Reopen range and Custom to capture completed range.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V283/rust.png` ↔ `../probes/visual/V283/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V283. Forbidden claims: Open pending Apply state immediately after second click; Apply/Clear footer buttons.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.
## V284 · Usage provider filter popup
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::multi_menu:L353](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L353)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Focused search; All row; Provider rows with checks and keyboard highlight; Select all and Clear footer.
- Setup: Open Provider; select provider A; leave menu open and focus a list row with ArrowDown.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V284. Forbidden claims: Exclusive single-select behavior.
- Planned pair: `../probes/visual/V284/rust.png` ↔ `../probes/visual/V284/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.