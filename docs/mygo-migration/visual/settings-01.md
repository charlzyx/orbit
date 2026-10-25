# Audited visual candidates: settings

These are numbered candidates. State validity, capture, comparison and parity are separate gates. See the master register and fixture.json.

## V070 · Settings general baseline
- Features: P01, P02, U05. Lane: `canonical`. Visual: **PARTIAL**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **VALID_SOURCE_BRANCH**. Pinned handler read and both native screenshots inspected. Actual screen exists; strict fixture/font/clock normalization is still incomplete.
- Expected: Settings general baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings general baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Actual diagnostic Rust image: `/workspace/shared/orbit-toolchains/reference-capture-1821f982/reference-desktop/settings-general-dark.png`; SHA256 `97b2186bf23698998f9eb66c6fd67762e7720acfb347a2c87485265f9a0fee70`.
- Actual diagnostic Go image: `/workspace/shared/orbit-toolchains/native-capture-1821f982/go-desktop/settings-general-dark.png`; SHA256 `22d2dbd09a6af6b43f83f05e31db79333146d2acadb8364ed8289bfb518f9c5c`; commit `1821f9829e20a620bfb7863cba7bba48ebe9f239`.
- Compared at1470×922 without crop/rescale/masks: 47020 changed pixels (3.469240%). Exact pixels: **DIFFERENT**, not PASS.
- Manual review: DIFFERENT: This Machine/Notifications cards and separators have small vertical offsets; text/glyph metrics differ; Go adds top-right native window controls and a -go-dev version suffix. Main labels and section ordering agree, but pixel parity does not.
- Full fixture gate: BLOCKED: no actual resolved font/file hashes or frozen clock/animation record; manifest dpi=1 mislabeled scale although Xvfb launched with -dpi96.

## V071 · Settings appearance baseline
- Features: P01, P18, P19, P20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings appearance baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings appearance baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V071/rust.png` ↔ `../probes/visual/V071/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V072 · Settings shortcuts baseline
- Features: P01, P20, S03, S05, W71. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings shortcuts baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings shortcuts baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`.
- Planned pair: `../probes/visual/V072/rust.png` ↔ `../probes/visual/V072/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V073 · Settings privacy baseline
- Features: P01, U06, U07, U13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings privacy baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings privacy baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V073/rust.png` ↔ `../probes/visual/V073/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V074 · Settings runtime baseline
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V074/rust.png` ↔ `../probes/visual/V074/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V075 · Settings agent baseline
- Features: P01, P04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings agent baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings agent baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V075/rust.png` ↔ `../probes/visual/V075/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V076 · Settings worktrees baseline
- Features: P01, W43. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings worktrees baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings worktrees baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`, `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`.
- Planned pair: `../probes/visual/V076/rust.png` ↔ `../probes/visual/V076/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V077 · Settings providers baseline
- Features: P01, P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings providers baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings providers baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V077/rust.png` ↔ `../probes/visual/V077/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V078 · Settings about baseline
- Features: P01, U08, U10, U14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_body:L676](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L676); [crates/orbit-pi/src/app/settings.rs::settings_rows:L932](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L932)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings about baseline;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings about baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V078/rust.png` ↔ `../probes/visual/V078/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V079 · Settings search label keyword matches
- Features: P01. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_search_results:L766](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L766)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings search label keyword matches;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings search label keyword matches. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V079/rust.png` ↔ `../probes/visual/V079/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V080 · Settings search zero results
- Features: P01. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::settings_search_results:L766](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L766)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings search zero results;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings search zero results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V080/rust.png` ↔ `../probes/visual/V080/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V081 · Settings general notifications denied
- Features: P01, P02, U05. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::notification_rows:L4778](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4778)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings general notifications denied;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings general notifications denied. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V081/rust.png` ↔ `../probes/visual/V081/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Requires macOS packaged/unbundled platform; CPU mock is not OS proof.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V082 · Settings general notifications unbundled macos
- Features: P01, P02, U05. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::notification_rows:L4778](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4778)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings general notifications unbundled macos;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings general notifications unbundled macos. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V082/rust.png` ↔ `../probes/visual/V082/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Requires macOS packaged/unbundled platform; CPU mock is not OS proof.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V083 · Settings runtime active pid uptime stderr
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::runtime_rows:L4931](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime active pid uptime stderr;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime active pid uptime stderr. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V083/rust.png` ↔ `../probes/visual/V083/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Actual process facts required; no invented PID/version/metrics.

## V084 · Settings runtime parked processes
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::runtime_rows:L4931](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime parked processes;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime parked processes. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V084/rust.png` ↔ `../probes/visual/V084/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Actual process facts required; no invented PID/version/metrics.

## V085 · Settings runtime stopped
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::runtime_rows:L4931](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime stopped;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime stopped. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V085/rust.png` ↔ `../probes/visual/V085/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Actual process facts required; no invented PID/version/metrics.

## V086 · Settings runtime failed stderr
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::runtime_rows:L4931](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime failed stderr;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime failed stderr. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V086/rust.png` ↔ `../probes/visual/V086/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Actual process facts required; no invented PID/version/metrics.

## V087 · Settings runtime patch status
- Features: P01, P02, U11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::runtime_rows:L4931](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4931)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings runtime patch status;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings runtime patch status. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V087/rust.png` ↔ `../probes/visual/V087/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Note: Actual process facts required; no invented PID/version/metrics.

## V088 · Settings providers catalog custom connected cards
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings providers catalog custom connected cards;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings providers catalog custom connected cards. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V088/rust.png` ↔ `../probes/visual/V088/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V089 · Settings providers filtered zero
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings providers filtered zero;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings providers filtered zero. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V089/rust.png` ↔ `../probes/visual/V089/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V090 · Settings providers partial catalog error
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings providers partial catalog error;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings providers partial catalog error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V090/rust.png` ↔ `../probes/visual/V090/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V091 · Settings provider add modal
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings provider add modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider add modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V091/rust.png` ↔ `../probes/visual/V091/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V092 · Settings provider edit modal
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings provider edit modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider edit modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V092/rust.png` ↔ `../probes/visual/V092/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V093 · Settings provider invalid save preserves draft
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings provider invalid save preserves draft;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider invalid save preserves draft. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V093/rust.png` ↔ `../probes/visual/V093/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V094 · Settings provider secret key modal
- Features: P05, P06, P07, P09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/settings.rs::provider_rows:L2391](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L2391); [crates/orbit-pi/src/app/settings.rs::provider_editor_layer:L4145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L4145); [crates/orbit-pi/src/app/settings.rs::provider_key_layer:L3926](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3926); [crates/orbit-pi/src/app/settings.rs::provider_usage_layer:L3798](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/settings.rs#L3798)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Settings provider secret key modal;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated settings surface in fixture.json. Configure the named source state: Settings provider secret key modal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V094/rust.png` ↔ `../probes/visual/V094/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
- Authentication gate: Real authentication remains unverified; only deterministic noncredential visual fixtures are permitted without the relevant authorization.
