# Audited visual candidates: shell

These are numbered candidates. State validity, capture, comparison and parity are separate gates. See the master register and fixture.json.

## V001 · New task current workspace empty projects
- Features: S01, S02, S07. Lane: `canonical`. Visual: **PARTIAL**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_empty_state:L2374](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2374)
- State validity: **VALID_SOURCE_BRANCH**. Pinned handler read and both native screenshots inspected. Actual screen exists; strict fixture/font/clock normalization is still incomplete.
- Expected: New task current workspace empty projects;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: New task current workspace empty projects. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/ui`, `internal/workbench`.
- Actual diagnostic Rust image: `/workspace/shared/orbit-toolchains/reference-capture-1821f982/reference-desktop/new-task-dark.png`; SHA256 `d4accf98a16cf6f95044be8f0a4cdfcc710a58e1cdce9fac1c465bc85e31256d`.
- Actual diagnostic Go image: `/workspace/shared/orbit-toolchains/native-capture-1821f982/go-desktop/new-task-dark.png`; SHA256 `5a2563ba498f6f54221e75c54fb4cd03a5203f7a89d3d37c9a801d93aac6d975`; commit `1821f9829e20a620bfb7863cba7bba48ebe9f239`.
- Compared at1470×922 without crop/rescale/masks: 187498 changed pixels (13.834020%). Exact pixels: **DIFFERENT**, not PASS.
- Manual review: DIFFERENT: Go wordmark is larger; New Task row and empty-projects block are lower; Ctrl+N label is missing; central badge/card geometry differs; Go composer adds Full access/Off/Send text absent in this Rust fixture; Go has extra titlebar/window controls. Caret/focus and backdrop animation are not fully normalized.
- Full fixture gate: BLOCKED: no actual resolved font/file hashes or frozen clock/animation record; manifest dpi=1 mislabeled scale although Xvfb launched with -dpi96.

## V002 · New task sidebar hidden
- Features: S01, S02, S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_empty_state:L2374](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2374)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: New task sidebar hidden;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: New task sidebar hidden. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/ui`, `internal/workbench`.
- Planned pair: `../probes/visual/V002/rust.png` ↔ `../probes/visual/V002/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V003 · New task photo backdrop
- Features: S01, S02, S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_empty_state:L2374](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2374)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: New task photo backdrop;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: New task photo backdrop. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/ui`, `internal/workbench`.
- Planned pair: `../probes/visual/V003/rust.png` ↔ `../probes/visual/V003/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V004 · Onboarding missing required dependencies
- Features: P03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_onboarding:L2664](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2664); [crates/orbit-pi/src/app/view.rs::render_pi_migration:L2903](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2903); [crates/orbit-pi/src/app/view.rs::render_dependency_row:L2955](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2955)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Onboarding missing required dependencies;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Onboarding missing required dependencies. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V004/rust.png` ↔ `../probes/visual/V004/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V005 · Onboarding installing dependency
- Features: P03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_onboarding:L2664](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2664); [crates/orbit-pi/src/app/view.rs::render_pi_migration:L2903](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2903); [crates/orbit-pi/src/app/view.rs::render_dependency_row:L2955](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2955)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Onboarding installing dependency;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Onboarding installing dependency. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V005/rust.png` ↔ `../probes/visual/V005/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V006 · Onboarding dependency install failure
- Features: P03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_onboarding:L2664](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2664); [crates/orbit-pi/src/app/view.rs::render_pi_migration:L2903](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2903); [crates/orbit-pi/src/app/view.rs::render_dependency_row:L2955](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2955)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Onboarding dependency install failure;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Onboarding dependency install failure. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V006/rust.png` ↔ `../probes/visual/V006/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V007 · Onboarding pi npm migration offer
- Features: P03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_onboarding:L2664](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2664); [crates/orbit-pi/src/app/view.rs::render_pi_migration:L2903](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2903); [crates/orbit-pi/src/app/view.rs::render_dependency_row:L2955](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L2955)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Onboarding pi npm migration offer;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Onboarding pi npm migration offer. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V007/rust.png` ↔ `../probes/visual/V007/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V008 · Sidebar grouped active pinned background running sessions
- Features: S06, S08, S12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::render_side_row:L546](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L546); [crates/orbit-pi/src/app/sidebar.rs::empty_sessions_state:L2063](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2063)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar grouped active pinned background running sessions;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar grouped active pinned background running sessions. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V008/rust.png` ↔ `../probes/visual/V008/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V009 · Sidebar flat all sessions
- Features: S06, S08, S12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::render_side_row:L546](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L546); [crates/orbit-pi/src/app/sidebar.rs::empty_sessions_state:L2063](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2063)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar flat all sessions;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar flat all sessions. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V009/rust.png` ↔ `../probes/visual/V009/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V010 · Sidebar archived only
- Features: S06, S08, S12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::render_side_row:L546](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L546); [crates/orbit-pi/src/app/sidebar.rs::empty_sessions_state:L2063](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2063)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar archived only;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar archived only. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V010/rust.png` ↔ `../probes/visual/V010/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V011 · Sidebar expanded project sticky header after scroll
- Features: S06, S08, S12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::render_side_row:L546](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L546); [crates/orbit-pi/src/app/sidebar.rs::empty_sessions_state:L2063](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L2063)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar expanded project sticky header after scroll;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar expanded project sticky header after scroll. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V011/rust.png` ↔ `../probes/visual/V011/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V012 · Sidebar session context menu
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar session context menu;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar session context menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V012/rust.png` ↔ `../probes/visual/V012/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V013 · Sidebar session delete confirmation
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar session delete confirmation;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar session delete confirmation. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V013/rust.png` ↔ `../probes/visual/V013/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V014 · Sidebar session inline rename
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar session inline rename;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar session inline rename. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V014/rust.png` ↔ `../probes/visual/V014/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V015 · Sidebar project context menu
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar project context menu;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar project context menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V015/rust.png` ↔ `../probes/visual/V015/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V016 · Sidebar project list options popup
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar project list options popup;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar project list options popup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V016/rust.png` ↔ `../probes/visual/V016/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V017 · Sidebar project appearance mark and color picker
- Features: S11, S12, S13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/sidebar.rs::session_menu_popup:L1250](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1250); [crates/orbit-pi/src/app/sidebar.rs::workspace_menu_popup:L1550](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1550); [crates/orbit-pi/src/app/sidebar.rs::sidebar_options_popup:L1925](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/sidebar.rs#L1925)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Sidebar project appearance mark and color picker;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Sidebar project appearance mark and color picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/native`, `internal/pirpc`, `internal/preferences`, `internal/workbench`.
- Planned pair: `../probes/visual/V017/rust.png` ↔ `../probes/visual/V017/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V018 · Command palette empty query sessions and commands
- Features: S04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/command_palette.rs::render:L742](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/command_palette.rs#L742)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Command palette empty query sessions and commands;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Command palette empty query sessions and commands. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`.
- Planned pair: `../probes/visual/V018/rust.png` ↔ `../probes/visual/V018/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V019 · Command palette filtered results
- Features: S04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/command_palette.rs::render:L742](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/command_palette.rs#L742)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Command palette filtered results;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Command palette filtered results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`.
- Planned pair: `../probes/visual/V019/rust.png` ↔ `../probes/visual/V019/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V020 · Command palette zero results
- Features: S04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/command_palette.rs::render:L742](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/command_palette.rs#L742)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Command palette zero results;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Command palette zero results. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`.
- Planned pair: `../probes/visual/V020/rust.png` ↔ `../probes/visual/V020/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V021 · Workspace picker recent current and browse
- Features: S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/workspace_picker.rs::render:L249](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/workspace_picker.rs#L249)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Workspace picker recent current and browse;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workspace picker recent current and browse. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Planned pair: `../probes/visual/V021/rust.png` ↔ `../probes/visual/V021/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V022 · Workspace picker filtered zero
- Features: S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/workspace_picker.rs::render:L249](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/workspace_picker.rs#L249)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Workspace picker filtered zero;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workspace picker filtered zero. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Planned pair: `../probes/visual/V022/rust.png` ↔ `../probes/visual/V022/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V023 · Workspace picker four row scroll limit
- Features: S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/workspace_picker.rs::render:L249](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/workspace_picker.rs#L249)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Workspace picker four row scroll limit;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workspace picker four row scroll limit. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Planned pair: `../probes/visual/V023/rust.png` ↔ `../probes/visual/V023/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V024 · Header more actions popup
- Features: S10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_header_more_popup:L3469](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3469)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Header more actions popup;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Header more actions popup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Planned pair: `../probes/visual/V024/rust.png` ↔ `../probes/visual/V024/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.

## V025 · Header inline title edit
- Features: S10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render_header_more_popup:L3469](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L3469)
- State validity: **NOT_REVIEWED**. Independent render-branch review pending.
- Expected: Header inline title edit;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Header inline title edit. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/pirpc`.
- Planned pair: `../probes/visual/V025/rust.png` ↔ `../probes/visual/V025/go.png`. Actual images absent. Metrics not measured; manual review not performed; verified Go SHA none.
