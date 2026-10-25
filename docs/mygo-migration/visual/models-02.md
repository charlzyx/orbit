# Audited visual candidates: models

Numbered candidates, not validated screenshot recipes. State validity, fixture execution, capture, comparison and parity are separate gates.

## V149 · Composer thinking options
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Only reported levels are rendered; five rows fit. This does not cover every possible icon alias or unknown-level fallback. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Flat Off, Low, Medium, High, Max icon/label/hint rows; Current High check; Search and5 levels footer; No scope row.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer thinking options. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V149. Overrides: `{"surface": "Composer", "picker": "Thinking", "levels": ["off", "low", "medium", "high", "max"], "current_level": "high"}`.
- Planned pair: `../probes/visual/V149/rust.png` ↔ `../probes/visual/V149/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V150 · Composer thinking filtered
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Source-render branch exists. Concrete fixture specified; actual paired capture still unverified. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Typed high query; Only High label/hint row; Current check; 1 levels footer.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer thinking filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V150. Overrides: `{"surface": "Composer", "picker": "Thinking", "picker_query": "high"}`.
- Planned pair: `../probes/visual/V150/rust.png` ↔ `../probes/visual/V150/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V151 · Composer thinking no matches
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/model_selector.rs::render:L885](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/model_selector.rs#L885)
- State validity: **VALID_SOURCE_BRANCH**. Explicit no-match entry is rendered; no invented choice. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: No matching levels title; Try a reasoning level hint; Zero levels footer.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer thinking no matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V151. Overrides: `{"surface": "Composer", "picker": "Thinking", "picker_query": "zzz-no-match"}`.
- Planned pair: `../probes/visual/V151/rust.png` ↔ `../probes/visual/V151/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.

## V152 · Composer thinking unavailable
- Features: C06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::thinking_chip:L2199](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2199)
- State validity: **VALID_SOURCE_BRANCH**. Zero levels and one level share the same absent branch. A disabled Thinking chip is invented in this state; running with a multi-level ladder is a distinct disabled-chip branch. Exact fixture and source ranges: model-visual-source-review.json.
- Expected: Model chip remains; No thinking chip and no thinking popup.
- Setup: Open the isolated models surface in fixture.json. Configure the named source state: Composer thinking unavailable. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Concrete fixture: [independent Models review](../model-visual-source-review.json), V152. Overrides: `{"surface": "Composer", "levels": ["off"], "current_level": "off", "picker": null}`.
- Planned pair: `../probes/visual/V152/rust.png` ↔ `../probes/visual/V152/go.png`. No actual capture for this candidate. Metrics not measured; no visual pass.
