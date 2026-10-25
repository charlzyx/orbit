# Audited visual candidates: usage

Numbered candidates, not validated screenshot recipes. State validity, fixture execution, capture, comparison and parity are separate gates.

## V285 · Usage model filter popup
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::multi_menu:L353](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L353)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Search, All, Select all, Clear; Duplicate model IDs distinguishable by provider subtitle; Only selected pair checked.
- Setup: Open Model; select providerA/shared-model, not providerB/shared-model.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V285. Forbidden claims: Raw model ID conflates providers.
- Planned pair: `../probes/visual/V285/rust.png` ↔ `../probes/visual/V285/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V286 · Usage workspace filter popup
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::multi_menu:L353](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L353)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Search, All, workspace rows, checkmarks; Select all and Clear.
- Setup: Open Workspace; select workspace A, leave popup open.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V286. Forbidden claims: .
- Planned pair: `../probes/visual/V286/rust.png` ↔ `../probes/visual/V286/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V287 · Usage filter search no matches
- Features: U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/filters.rs::multi_menu:L353](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/filters.rs#L353)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Query retained; All row remains; No match for “zz-no-match”; Select all and Clear footer; Underlying selected chip persists.
- Setup: Select provider A; open Provider and type zz-no-match.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V287. Forbidden claims: All options including All disappear; Selection lost because search has no results.
- Planned pair: `../probes/visual/V287/rust.png` ↔ `../probes/visual/V287/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V288 · Usage chart tokens
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Tokens active in eight-way segmented selector; Source five y labels and adaptive x labels; Filled area and straight stroke; Latency response line when timed data exists.
- Setup: Simple; scroll to Trend; choose Tokens.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V288. Forbidden claims: .
- Planned pair: `../probes/visual/V288/rust.png` ↔ `../probes/visual/V288/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V289 · Usage chart requests
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Requests active in eight-way segmented selector; Source five y labels and adaptive x labels; Filled area and straight stroke; Latency response line when timed data exists.
- Setup: Simple; scroll to Trend; choose Requests.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V289. Forbidden claims: .
- Planned pair: `../probes/visual/V289/rust.png` ↔ `../probes/visual/V289/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V290 · Usage chart input output cache
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **NEEDS_SPLIT**. Candidate combines mutually exclusive source states; cannot be one screenshot. Freeze candidate ID; do not count it as a valid recipe.
- Expected: Input/output/cache-series chooser and selected series.
- Setup: Open the isolated usage surface in fixture.json. Configure the named source state: Usage chart input output cache. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V290. Forbidden claims: Output and Cached simultaneously active; append separate cases.
- Planned pair: `../probes/visual/V290/rust.png` ↔ `../probes/visual/V290/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V291 · Usage chart cost complete
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Cost active in eight-way segmented selector; Source five y labels and adaptive x labels; Filled area and straight stroke; Latency response line when timed data exists.
- Setup: Simple; scroll to Trend; choose Cost.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V291. Forbidden claims: Partial/complete coverage badge.
- Planned pair: `../probes/visual/V291/rust.png` ↔ `../probes/visual/V291/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V292 · Usage chart cost partial
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Cost selectable because priced_requests>0; Only known cost subtotal plotted.
- Setup: Simple Trend; choose Cost with one priced and one unpriced request in same window.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V292. Forbidden claims: Partial-pricing warning/badge: not rendered; Fully unavailable Cost is this same state.
- Planned pair: `../probes/visual/V292/rust.png` ↔ `../probes/visual/V292/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V293 · Usage chart latency
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Latency active in eight-way segmented selector; Source five y labels and adaptive x labels; Filled area and straight stroke; Latency response line when timed data exists.
- Setup: Simple; scroll to Trend; choose Latency.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V293. Forbidden claims: .
- Planned pair: `../probes/visual/V293/rust.png` ↔ `../probes/visual/V293/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V294 · Usage chart latency stat menu
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Average, p50, p95, p99 segments; p95 selected; Duration axis; Measured-response summary.
- Setup: Simple Trend; choose Latency, then p95 in inline selector.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V294/rust.png` ↔ `../probes/visual/V294/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V294. Forbidden claims: Dropdown popup; Percentiles enabled with fewer than20 samples.
- Note: Independent source challenge corrected first-pass expectations; no image comparison has run.
## V295 · Usage chart errors
- Features: U02. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Errors active in eight-way segmented selector; Source five y labels and adaptive x labels; Filled area and straight stroke; Latency response line when timed data exists.
- Setup: Simple; scroll to Trend; choose Errors.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V295. Forbidden claims: .
- Planned pair: `../probes/visual/V295/rust.png` ↔ `../probes/visual/V295/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V296 · Usage chart tooltip focus
- Features: U02, U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/chart.rs::timeline:L103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/chart.rs#L103)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Selected time-focus band; Hover marker and readout card; Exact bucket readout; Focus chip above if viewport permits.
- Setup: Simple Trend Tokens; click bucket October22, leave pointer over same bucket.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V296. Forbidden claims: Rust keyboard-selected band: timeline has only mouse handlers.
- Planned pair: `../probes/visual/V296/rust.png` ↔ `../probes/visual/V296/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V297 · Usage heatmap day tooltip
- Features: U02, U03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/heatmap.rs::calendar:L126](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/heatmap.rs#L126)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Day selected outline; Day tooltip and metric readout; Monday-first grid, month labels; Padding outside calendar has no fabricated activity.
- Setup: Hover then click active heatmap day October22; leave pointer there.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V297. Forbidden claims: Keyboard heatmap navigation as Rust baseline.
- Planned pair: `../probes/visual/V297/rust.png` ↔ `../probes/visual/V297/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V298 · Usage table search results
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Query Alpha; Only matching sessions; Matched/total count; Filtered totals and pager reset to page1.
- Setup: Details Sessions; type Alpha in session search; scroll table into view.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V298. Forbidden claims: Global filter changed by table search.
- Planned pair: `../probes/visual/V298/rust.png` ↔ `../probes/visual/V298/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V299 · Usage table search empty
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Query retained; No sessions match search in table viewport; No matching rows; Pagination/metadata appropriate to zero results.
- Setup: Details Sessions; type zz-no-match; scroll table into view.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V299. Forbidden claims: Whole-page filtered-empty message replacing table.
- Planned pair: `../probes/visual/V299/rust.png` ↔ `../probes/visual/V299/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V300 · Usage table sort states
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **NEEDS_SPLIT**. Candidate combines mutually exclusive source states; cannot be one screenshot. Freeze candidate ID; do not count it as a valid recipe.
- Expected: Source ascending/descending/default indicators across visible headers.
- Setup: Open the isolated usage surface in fixture.json. Configure the named source state: Usage table sort states. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V300. Forbidden claims: Ascending/descending/default active simultaneously; append Title descending and default-reset captures.
- Planned pair: `../probes/visual/V300/rust.png` ↔ `../probes/visual/V300/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V301 · Usage table columns menu
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Source Sessions column titles; Shown checkmarks; Reset columns action.
- Setup: Details Sessions; click Columns; keep native/source popup visible.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V301. Forbidden claims: Headless MyGoTester Menu() metadata counts as painted native popup.
- Planned pair: `../probes/visual/V301/rust.png` ↔ `../probes/visual/V301/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V302 · Usage table hidden resized columns
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: Workspace header/cells absent; Provider width override156px; Aligned cells and remaining headers.
- Setup: Details Sessions; hide Workspace via Columns and dismiss; drag Provider divider to156px. Ensure enough width for Provider in source responsive plan.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V302. Forbidden claims: Hidden column still consumes width; Resize state proves persistence without reload test.
- Planned pair: `../probes/visual/V302/rust.png` ↔ `../probes/visual/V302/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V303 · Usage table page size menu
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **VALID_SOURCE_BRANCH**. Complete domain fixture, capture steps, forbidden claims and exact Rust lines in usage-visual-source-review.json. Source branch reviewed; actual fixture/capture not executed.
- Expected: 10,25,50,100 options; 10 checked; Associated table footer.
- Setup: Details Sessions; current page size10; click rows-per-page trigger.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V303. Forbidden claims: All five independent table menus covered by this single capture.
- Planned pair: `../probes/visual/V303/rust.png` ↔ `../probes/visual/V303/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
## V304 · Usage table next page selected row
- Features: U04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/usage/table.rs::data_table:L181](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/table.rs#L181)
- State validity: **INVALID_COMBINATION**. Source page.rs1580 set_filter resets to page1. The combined nonfirst-page+session-scoped selection candidate is invalid.
- Expected: Nonfirst page;  Paging controls;  Selected row.
- Setup: Open the isolated usage surface in fixture.json. Configure the named source state: Usage table next page selected row. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Independent recipe: [Usage source review](../usage-visual-source-review.json), V304. Forbidden claims: Selected scoped session on page2: row click filters and resets page1; append separate selected-page1 case.
- Planned pair: `../probes/visual/V304/rust.png` ↔ `../probes/visual/V304/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.