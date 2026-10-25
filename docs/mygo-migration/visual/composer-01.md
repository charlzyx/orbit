# Visual inventory: composer (V030–V056)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V030 · Composer idle multiline draft
- Features: C01, C02, C18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer idle multiline draft;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer idle multiline draft. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V030/rust.png` ↔ `../probes/visual/V030/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V031 · Composer running steer followup primary alternate
- Features: C01, C02, C18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer running steer followup primary alternate;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer running steer followup primary alternate. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V031/rust.png` ↔ `../probes/visual/V031/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V032 · Composer access mode menu
- Features: C01, C02, C18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer access mode menu;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer access mode menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V032/rust.png` ↔ `../probes/visual/V032/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V033 · Composer workflow build plan ask menu
- Features: C01, C02, C18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer workflow build plan ask menu;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer workflow build plan ask menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V033/rust.png` ↔ `../probes/visual/V033/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V034 · Composer queued steering and followup messages
- Features: C03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer queued steering and followup messages;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer queued steering and followup messages. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V034/rust.png` ↔ `../probes/visual/V034/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V035 · Composer queued message edit remove
- Features: C03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer queued message edit remove;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer queued message edit remove. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`
- Planned pair: `../probes/visual/V035/rust.png` ↔ `../probes/visual/V035/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V036 · Composer leading slash command popup
- Features: C04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/composer_ops.rs::autocomplete_popup:L147](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L147)
- Expected: Composer leading slash command popup;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer leading slash command popup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V036/rust.png` ↔ `../probes/visual/V036/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V037 · Composer skill command with description
- Features: C04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/composer_ops.rs::autocomplete_popup:L147](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L147)
- Expected: Composer skill command with description;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer skill command with description. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V037/rust.png` ↔ `../probes/visual/V037/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V038 · Composer file mention popup
- Features: C04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/composer_ops.rs::autocomplete_popup:L147](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L147)
- Expected: Composer file mention popup;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer file mention popup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V038/rust.png` ↔ `../probes/visual/V038/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V039 · Composer autocomplete zero results
- Features: C04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/composer_ops.rs::autocomplete_popup:L147](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L147)
- Expected: Unmatched slash/file-mention draft remains in the normal composer; No autocomplete overlay or empty-results panel: autocomplete_popup returns None when count is zero or entries are empty.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer autocomplete zero results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V039/rust.png` ↔ `../probes/visual/V039/go.png`. Neither image exists in this register.
- Note: Independently challenged by host owner: pinned app/composer_ops.rs:147–159 suppresses the popup entirely. This is an absence-state capture, not a fake empty popup.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
## V040 · Composer image attachment strip
- Features: C05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer image attachment strip;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer image attachment strip. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V040/rust.png` ↔ `../probes/visual/V040/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V041 · Composer image preview modal
- Features: C05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer image preview modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer image preview modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V041/rust.png` ↔ `../probes/visual/V041/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V042 · Composer drag over highlight
- Features: C05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer drag over highlight;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer drag over highlight. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V042/rust.png` ↔ `../probes/visual/V042/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V043 · Composer partial attachment failure
- Features: C05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::composer_row:L1414](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L1414)
- Expected: Composer partial attachment failure;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Composer partial attachment failure. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V043/rust.png` ↔ `../probes/visual/V043/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V053 · Context meter known usage cost
- Features: C15, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Context meter known usage cost;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Context meter known usage cost. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V053/rust.png` ↔ `../probes/visual/V053/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V054 · Context meter unknown after compaction
- Features: C15, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Context meter unknown after compaction;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Context meter unknown after compaction. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V054/rust.png` ↔ `../probes/visual/V054/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V055 · Quota popup reports windows balances reset
- Features: C15, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Quota popup reports windows balances reset;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Quota popup reports windows balances reset. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V055/rust.png` ↔ `../probes/visual/V055/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V056 · Quota popup loading unsupported error
- Features: C15, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Quota popup loading unsupported error;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated composer surface in fixture.json. Configure the named source state: Quota popup loading unsupported error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V056/rust.png` ↔ `../probes/visual/V056/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
