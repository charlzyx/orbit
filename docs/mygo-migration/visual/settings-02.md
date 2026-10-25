# Visual inventory: settings (V095–V111)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V095 · Settings provider remove confirmation
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- Expected: Settings provider remove confirmation;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider remove confirmation. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V095/rust.png` ↔ `../probes/visual/V095/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V096 · Settings provider oauth method selection
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- Expected: Provider card authentication-method buttons; Source methods available from capability metadata; not an invented modal.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider oauth method selection. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V096/rust.png` ↔ `../probes/visual/V096/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Authentication gate: Real authentication remains unverified; only deterministic noncredential visual fixtures are permitted without the relevant authorization.

## V097 · Settings provider oauth in progress device code
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- Expected: Settings provider oauth in progress device code;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider oauth in progress device code. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V097/rust.png` ↔ `../probes/visual/V097/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Authentication gate: Real authentication remains unverified; only deterministic noncredential visual fixtures are permitted without the relevant authorization.

## V098 · Settings provider oauth error
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- Expected: Settings provider oauth error;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider oauth error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V098/rust.png` ↔ `../probes/visual/V098/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Authentication gate: Real authentication remains unverified; only deterministic noncredential visual fixtures are permitted without the relevant authorization.

## V099 · Settings provider quota detail modal
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- Expected: Settings provider quota detail modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider quota detail modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V099/rust.png` ↔ `../probes/visual/V099/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V100 · Settings appearance light theme select
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance light theme select;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance light theme select. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V100/rust.png` ↔ `../probes/visual/V100/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V101 · Settings appearance dark theme select
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance dark theme select;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance dark theme select. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V101/rust.png` ↔ `../probes/visual/V101/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V102 · Settings appearance ui font select
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance ui font select;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance ui font select. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V102/rust.png` ↔ `../probes/visual/V102/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V103 · Settings appearance code font select
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance code font select;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance code font select. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V103/rust.png` ↔ `../probes/visual/V103/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V104 · Settings appearance language select
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance language select;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance language select. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V104/rust.png` ↔ `../probes/visual/V104/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V105 · Settings appearance photo processing
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance photo processing;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance photo processing. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V105/rust.png` ↔ `../probes/visual/V105/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V106 · Settings appearance photo preview tuning
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance photo preview tuning;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance photo preview tuning. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V106/rust.png` ↔ `../probes/visual/V106/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V107 · Settings appearance photo load error
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::appearance_rows:L5926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5926); [crates/orbit-pi/src/app/settings.rs::settings_select_popup:L6907](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L6907)
- Expected: Settings appearance photo load error;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance photo load error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V107/rust.png` ↔ `../probes/visual/V107/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V108 · Settings privacy retention mode menu
- Features: P01, U06. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::retention_mode_select:L5671](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L5671)
- Expected: Archive/Delete retention action options; Current mode; No invented preview page.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings privacy retention mode menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V108/rust.png` ↔ `../probes/visual/V108/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V109 · Settings privacy temp scan results
- Features: P01, U07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/temp_ui.rs::temporary_files_section:L79](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/temp_ui.rs#L79)
- Expected: Settings privacy temp scan results;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings privacy temp scan results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V109/rust.png` ↔ `../probes/visual/V109/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V110 · Settings privacy temp empty
- Features: P01, U07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/temp_ui.rs::temporary_files_section:L79](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/temp_ui.rs#L79)
- Expected: Empty temporary files state; Open folder; No cleanup confirmation exists in pinned source.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings privacy temp empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V110/rust.png` ↔ `../probes/visual/V110/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V111 · Settings privacy temp cleanup partial failure
- Features: P01, U07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/toast_ui.rs::toast_layer:L54](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/toast_ui.rs#L54)
- Expected: Removed/skipped-active/failed counts in warning toast; Refreshed temporary-file list.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings privacy temp cleanup partial failure. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V111/rust.png` ↔ `../probes/visual/V111/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
