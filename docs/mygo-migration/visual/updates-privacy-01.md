# Visual inventory: updates-privacy (V112–V478)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V112 · Updater checking
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Updater checking;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater checking. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V112/rust.png` ↔ `../probes/visual/V112/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V113 · Updater current
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Updater current;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater current. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V113/rust.png` ↔ `../probes/visual/V113/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V114 · Updater available release
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Check-origin Available dialog; Version/release notes; Install and Cancel.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater available release. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V114/rust.png` ↔ `../probes/visual/V114/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V115 · Updater installing status
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Updating/installing state; Source spinner/label, no invented percentage bar.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater installing status. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V115/rust.png` ↔ `../probes/visual/V115/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V116 · Updater ready to install
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Available dialog opened by staged-download control; Later secondary action versus Cancel in check-origin dialog.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater ready to install. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V116/rust.png` ↔ `../probes/visual/V116/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V117 · Updater error
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Updater error;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V117/rust.png` ↔ `../probes/visual/V117/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V118 · Updater release history modal
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_section:L403](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L403); [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Updater release history modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater release history modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V118/rust.png` ↔ `../probes/visual/V118/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Signed native Go feed/key is unconfigured. Deterministic fixture captures do not establish live update parity.

## V470 · Pi update available
- Features: U10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/toast_ui.rs::toast_layer:L54](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/toast_ui.rs#L54)
- Expected: Background Pi update available toast; No invented manual update page.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Pi update available. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V470/rust.png` ↔ `../probes/visual/V470/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V471 · Pi update deferred busy toast
- Features: U10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/toast_ui.rs::toast_layer:L54](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/toast_ui.rs#L54)
- Expected: Agent remained busy until wait deadline; Update deferred until next launch toast.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Pi update deferred busy toast. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V471/rust.png` ↔ `../probes/visual/V471/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V472 · Pi update error
- Features: U10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/toast_ui.rs::toast_layer:L54](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/toast_ui.rs#L54)
- Expected: Pi update/install error toast; Actual failure detail.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Pi update error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V472/rust.png` ↔ `../probes/visual/V472/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V478 · Updater unavailable install
- Features: U08, U09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/updater_ui.rs::updater_dialog_layer:L565](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/updater_ui.rs#L565)
- Expected: Development/unmanaged installation notice;  No unsafe install action.
- Setup: Open the isolated updates-privacy surface in fixture.json. Configure the named source state: Updater unavailable install. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V478/rust.png` ↔ `../probes/visual/V478/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
