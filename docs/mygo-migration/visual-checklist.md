# Orbit audited visual candidates

**490 numbered first-pass candidate states, not490 validated screenshot recipes. The provisional arithmetic placeholder is980 images, not a finalized capture requirement:471 in-app candidates +19 native-OS candidates. Actual captures:4 images. Same-geometry diagnostic pairs compared:2. Fully fixture-verified pairs:0. Passed:0.**

Rust reference: [`v0.2.12` at `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`](https://github.com/imrj05/orbit/tree/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac). Source-state recipes remain under independent review. This is a first-pass candidate inventory, not a claim that every render branch is proven covered. All139 functional scopes are cross-referenced; four have no independent renderer and remain behavioral tests. All14 Settings sections have a baseline case.

## What this count means

- One V-number initially identifies a candidate source render/setup state, intended to become one Rust/Go image pair after state validation, not a feature or a single control. Shared controls are deduplicated, not multiplied through every data/theme/keyboard combination.
- The 19 native-OS cases include individually opened App/File/Edit/View menus on Linux/Windows/macOS, native pickers, notifications and macOS reopen. They are blocked, not omitted or counted as captured.
- Every source path and handler was checked against the pinned Git object. Fixture generation and state-specific independent review are still pending. Source-anchor existence is not proof that an implementation or capture passes.
- Arbitrary third-party extension frames, arbitrary repository contents and arbitrary custom themes are unbounded. Deterministic protocol/render-equivalence fixtures represent them; this finite list does not claim to enumerate every possible user input.
- If actual1470×922 capture shows required controls below the viewport, add a separately numbered scroll-state case. Do not claim hidden controls are compared. New cases append IDs; old IDs never silently change.
- Earlier two diagnostic pairs had different client sizes and are excluded. Existing CPU images/components are useful evidence but are not matched Rust/Go full-window pairs.

## Fixed comparison contract

English; dark orbit-azure;100% UI scale;1470×922 client pixels; DPI96; fixed UTC clock and animation phase; source bundled default fonts with actual resolved names/hashes recorded. Isolated disposable home, repositories and session store only. No user Mac and no real remote mutations. See [fixture.json](visual/fixture.json).

Each case records source handler/line, feature IDs, setup, expected controls, Go mapping, planned image paths, actual image/hash slots, geometry/pixel metrics, manual review, visual/native verdict and last verified Go SHA. V001 and V070 now contain actual image hashes, source commits, raw metrics and manual differences; other cases have no actual captures. No source or Go code was changed for this register.

## Category totals

| Category | Cases | Register |
|---|---:|---|
| shell | 32 | [V001–V025](visual/shell-01.md), [V026–V469](visual/shell-02.md) |
| composer | 18 | [V030–V056](visual/composer-01.md) |
| transcript | 69 | [V044–V168](visual/transcript-01.md), [V169–V193](visual/transcript-02.md), [V194–V212](visual/transcript-03.md) |
| extensions | 13 | [V057–V069](visual/extensions-01.md) |
| settings | 42 | [V070–V094](visual/settings-01.md), [V095–V111](visual/settings-02.md) |
| updates-privacy | 11 | [V112–V478](visual/updates-privacy-01.md) |
| platform | 17 | [V119–V490](visual/platform-01.md) |
| models | 29 | [V124–V148](visual/models-01.md), [V149–V152](visual/models-02.md) |
| resources | 21 | [V213–V233](visual/resources-01.md) |
| mcp | 22 | [V234–V475](visual/mcp-01.md) |
| bug-report | 9 | [V253–V477](visual/bug-report-01.md) |
| usage | 45 | [V260–V284](visual/usage-01.md), [V285–V304](visual/usage-02.md) |
| git | 40 | [V305–V329](visual/git-01.md), [V330–V344](visual/git-02.md) |
| github | 33 | [V345–V369](visual/github-01.md), [V370–V377](visual/github-02.md) |
| review | 23 | [V378–V400](visual/review-01.md) |
| files | 30 | [V401–V425](visual/files-01.md), [V426–V430](visual/files-02.md) |
| worktrees | 23 | [V431–V453](visual/worktrees-01.md) |
| terminal | 13 | [V454–V466](visual/terminal-01.md) |

## Scope and evidence crosswalk

[All139 scopes](visual/feature-crosswalk.json) · [Machine-readable index](visual/index.json) · [Functional acceptance checklist](acceptance-checklist.md)

Behavior-only scopes: S09 session dispatch/navigation; S14 refresh fan-out; W02 filesystem watcher/invalidation; W63 terminal input bytes. Their visible outcomes reuse existing cases, but screenshots cannot prove their behavior. Backend obligations within mixed visual scopes also remain functional tests.

## Capture acceptance, one case at a time

1. Implement and independently validate the exact fixture; confirm all expected controls are actually visible.
2. Capture pinned Rust and the exact Go SHA/source manifest at identical client bounds, DPI, state, fonts, clock and animation phase.
3. Record both paths and SHA256s; reject geometry/fixture mismatch instead of rescaling.
4. Generate a diff image and record differing pixels, max channel delta, mean error, optional SSIM, geometry deltas and clipped text. No hidden masks.
5. Review manually, record the mismatch/fix and rerun that pair. Only then may this visual subcase pass. Functional and native-OS verdicts stay separate.

## Independently removed false candidates

Pinned source has no stash-create dialog, stash-drop confirmation or temp-cleanup confirmation. An unavailable Open-in control is absent, not an empty app menu. Pi update feedback is toast-based, not a separate update page. MCP exposure is five radio choices, not an invented dropdown. Native menus are distinct captures, not one impossible image with four open menus. These corrections are already reflected in the490-candidate register.

## Independent state challenge and actual capture progress

Source paths/handlers verified:490. Source render branches independently reviewed as valid:73 (Models, Usage, V039 autocomplete absence, V001/V070 actual screens); only2 rendered cases have actually been observed, both diagnostic. Full deterministic fixture/font/clock verification:0. Actual native images captured:4. Same-geometry diagnostic pairs compared:2. Passed:0. Handler existence is not a validated screen state.

V290 combines mutually exclusive Input/Output/Cache metrics and V300 combines mutually exclusive sort states: **NEEDS_SPLIT**. V304 combines page2 with session selection even though selection resets to page1: **INVALID_COMBINATION**. Usage and Models expectations now link their concrete independently reviewed fixtures and forbidden claims; none of those recipes has yet established matched-image parity. V274 is **UNREACHABLE_SOURCE_PRODUCTION**, not a Rust refresh-failure screen. The first-load scan-error branch can only be injected, not reached by the production scanner. Additional source branches such as undersampled latency, fully unpriced cost and six Models variants remain recorded discoveries, not silently omitted or prematurely counted captures. The490 count is frozen as candidates, not an exhaustive finalized screenshot total.

## First actual native comparisons

Both Rust299a32b and Go1821f9829e20a620bfb7863cba7bba48ebe9f239 were captured at1470×922 in Linux Xvfb/Openbox. Four PNG hashes were checked and all four images were manually inspected. No resizing, crop alignment or masks were used.

| Case | Changed pixels | Changed area | Manual result |
|---|---:|---:|---|
| V001 New task |187,498|13.834020%|Different: wordmark/sidebar geometry, central card, composer controls, titlebar and focus state|
| V070 General |47,020|3.469240%|Different: card/separator offsets, glyph metrics, window controls and version text|

These are useful measured diagnostics, not fully fixture-verified parity results. The manifests lack resolved font-file hashes and frozen clock/animation records, and the `dpi` field incorrectly records scale1 although Xvfb uses-dpi96. Both pairs require metadata correction and controlled reshoot before strict acceptance. Actual paths/hashes and complete metrics are in the V001/V070 category records. No pass is claimed.

Capture priority remains V001 New task, V070 General, then the remaining14-section Settings sequence. Keep functional, state-validity, capture, raw comparison, visual pass and native-OS gates separate.

The independent [Models source review](model-visual-source-review.json) supplies concrete synthetic catalogs, selections and expected children for all29 existing model candidates; it also lists six not-yet-enumerated branches. Inaccessible/unreachable or mixed candidates remain flagged, not promoted to valid screens.

The complete [45-candidate Usage review](usage-visual-source-review.json) records exact fixtures, source lines, forbidden claims and14 additional missing/split candidate families. The frozen490 includes unresolved candidates; it is not a promise that exactly980 captures will exhaust the final scope.
