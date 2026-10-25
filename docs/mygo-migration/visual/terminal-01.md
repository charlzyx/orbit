# Visual inventory: terminal (V454–V466)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V454 · Terminal docked shell
- Features: W01, W03, W61. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Bottom panel;  Actual shell prompt;  Active tab;  New/close controls.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal docked shell. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/terminal`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V454/rust.png` ↔ `../probes/visual/V454/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V455 · Terminal multiple tabs
- Features: W61, W65. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::tab_chip:L1877](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1877)
- Expected: Distinct cwd/title tabs;  Active selection;  Per-tab close.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal multiple tabs. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V455/rust.png` ↔ `../probes/visual/V455/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V456 · Terminal overflow tabs
- Features: W65. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Overflow tab strip scrolled;  Active tab visible.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal overflow tabs. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V456/rust.png` ↔ `../probes/visual/V456/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V457 · Terminal title tooltip
- Features: W65. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::tab_chip:L1877](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1877)
- Expected: OSC title versus folder fallback;  Original cwd tooltip.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal title tooltip. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V457/rust.png` ↔ `../probes/visual/V457/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V458 · Terminal ansi colors attributes
- Features: W62. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: ANSI/indexed/truecolor swatches;  Bold/italic/underline/inverse text.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal ansi colors attributes. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V458/rust.png` ↔ `../probes/visual/V458/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V459 · Terminal wide unicode grid
- Features: W62. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Wide CJK/emoji/combining cells;  Column alignment.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal wide unicode grid. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V459/rust.png` ↔ `../probes/visual/V459/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V460 · Terminal cursor focused
- Features: W62. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Focused block/bar/underline cursor fixture;  Exact cell bounds.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal cursor focused. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V460/rust.png` ↔ `../probes/visual/V460/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V461 · Terminal cursor unfocused
- Features: W62. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Unfocused cursor appearance;  Same grid.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal cursor unfocused. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V461/rust.png` ↔ `../probes/visual/V461/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V462 · Terminal selection
- Features: W62, W64. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Mouse-selected grid text;  Correct cell-boundary highlight.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal selection. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V462/rust.png` ↔ `../probes/visual/V462/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V463 · Terminal scrollback
- Features: W62, W64. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Scrolled history;  Scrollback indicator;  View no longer at live bottom.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal scrollback. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V463/rust.png` ↔ `../probes/visual/V463/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V464 · Terminal resized panel
- Features: W03, W62. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Changed dock height;  Reflowed PTY dimensions;  Same full client size.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal resized panel. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/terminal`, `internal/workbench`
- Planned pair: `../probes/visual/V464/rust.png` ↔ `../probes/visual/V464/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V465 · Terminal exited
- Features: W65. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: Exited shell overlay;  Restart active tab action.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal exited. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V465/rust.png` ↔ `../probes/visual/V465/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V466 · Terminal start error
- Features: W65. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/terminal.rs::render:L1336](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/terminal.rs#L1336)
- Expected: PTY/shell startup failure;  Honest error message.
- Setup: Open the isolated terminal surface in fixture.json. Configure the named source state: Terminal start error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/terminal`
- Planned pair: `../probes/visual/V466/rust.png` ↔ `../probes/visual/V466/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
