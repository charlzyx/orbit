# Visual inventory: bug-report (V253–V477)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V253 · Settings report a bug baseline
- Features: P01, U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Bug/feature/Other selector; Context with Generate; Title; What; Steps fields; Screenshot controls; Diagnostics disclosure (not an invented consent switch).
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Settings report a bug baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V253/rust.png` ↔ `../probes/visual/V253/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V254 · Report feature mode
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Feature request type;  Source-specific labels.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report feature mode. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V254/rust.png` ↔ `../probes/visual/V254/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V255 · Report validation error
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Required fields marked;  Draft retained.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report validation error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V255/rust.png` ↔ `../probes/visual/V255/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V256 · Report attachment preview
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Attached screenshot/file list;  Remove attachment.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report attachment preview. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V256/rust.png` ↔ `../probes/visual/V256/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V257 · Report submitting
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Submitting state;  Duplicate disabled.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report submitting. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V257/rust.png` ↔ `../probes/visual/V257/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V258 · Report submit error
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Failure detail;  Retry;  Draft retained.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report submit error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V258/rust.png` ↔ `../probes/visual/V258/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V259 · Report submit success
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Created toast or browser-fallback notice; Source form cleared only after success.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report submit success. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V259/rust.png` ↔ `../probes/visual/V259/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V476 · Report other mode
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Other report type;  Source label changes.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report other mode. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V476/rust.png` ↔ `../probes/visual/V476/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V477 · Report generating draft
- Features: U12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/bug_report_ui.rs::bug_report_rows:L27](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/bug_report_ui.rs#L27)
- Expected: Generate busy;  Current model label;  Duplicate action disabled.
- Setup: Open the isolated bug-report surface in fixture.json. Configure the named source state: Report generating draft. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V477/rust.png` ↔ `../probes/visual/V477/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
