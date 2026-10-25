# Visual inventory: platform (V119–V490)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V119 · Macos parked window dock reopen
- Features: S01, W69, W70. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main_window.rs::open:L77](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main_window.rs#L77)
- Expected: Macos parked window dock reopen;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Macos parked window dock reopen. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V119/rust.png` ↔ `../probes/visual/V119/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Platform gate; screenshot inventory is not proof these actual OS states executed.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V120 · Os folder picker
- Features: S01, W69. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/app/session.rs::browse_for_folder:L724](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/session.rs#L724)
- Expected: Os folder picker;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Os folder picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V120/rust.png` ↔ `../probes/visual/V120/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Platform gate; screenshot inventory is not proof these actual OS states executed.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V121 · Os attachment picker
- Features: S01, W69. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/app/composer_ops.rs::attach_image:L435](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/composer_ops.rs#L435)
- Expected: Os attachment picker;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Os attachment picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V121/rust.png` ↔ `../probes/visual/V121/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Platform gate; screenshot inventory is not proof these actual OS states executed.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V122 · Os export save dialog
- Features: U04, W69. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/usage/page.rs::export:L2457](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/usage/page.rs#L2457)
- Expected: Os export save dialog;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Os export save dialog. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V122/rust.png` ↔ `../probes/visual/V122/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Platform gate; screenshot inventory is not proof these actual OS states executed.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V123 · Os notification banner
- Features: U05, W69. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/notifications.rs::notify:L182](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/notifications.rs#L182)
- Expected: Os notification banner;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Os notification banner. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V123/rust.png` ↔ `../probes/visual/V123/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Platform gate; screenshot inventory is not proof these actual OS states executed.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V479 · Native linux app menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: App menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native linux app menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V479/rust.png` ↔ `../probes/visual/V479/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual linux native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V480 · Native linux file menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: File menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native linux file menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V480/rust.png` ↔ `../probes/visual/V480/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual linux native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V481 · Native linux edit menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: Edit menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native linux edit menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V481/rust.png` ↔ `../probes/visual/V481/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual linux native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V482 · Native linux view menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: View menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native linux view menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V482/rust.png` ↔ `../probes/visual/V482/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual linux native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V483 · Native windows app menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: App menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native windows app menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V483/rust.png` ↔ `../probes/visual/V483/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual windows native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V484 · Native windows file menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: File menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native windows file menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V484/rust.png` ↔ `../probes/visual/V484/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual windows native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V485 · Native windows edit menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: Edit menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native windows edit menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V485/rust.png` ↔ `../probes/visual/V485/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual windows native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V486 · Native windows view menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: View menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native windows view menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V486/rust.png` ↔ `../probes/visual/V486/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual windows native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V487 · Native macos app menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: App menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native macos app menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V487/rust.png` ↔ `../probes/visual/V487/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual macos native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V488 · Native macos file menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: File menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native macos file menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V488/rust.png` ↔ `../probes/visual/V488/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual macos native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V489 · Native macos edit menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: Edit menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native macos edit menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V489/rust.png` ↔ `../probes/visual/V489/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual macos native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.

## V490 · Native macos view menu
- Features: S01, S03, W70, W71. Lane: `native-os`. Visual: **BLOCKED**. Native OS: **BLOCKED**.
- Rust: [crates/orbit-pi/src/main.rs::app_menus:L574](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/main.rs#L574)
- Expected: View menu open;  Source menu items and separators;  Correct shortcuts and enabled states.
- Setup: Open the isolated platform surface in fixture.json. Configure the named source state: Native macos view menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V490/rust.png` ↔ `../probes/visual/V490/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Actual macos native menu required. CPU descriptor capture does not paint this menu.
- Blocker: Actual native OS menu/dialog/notification execution not yet captured; headless CPU descriptors are insufficient.
