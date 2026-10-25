# Visual inventory: shell (V026–V469)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V026 · Branch picker local remote current filter
- Features: W06, S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/branch_picker.rs::render:L238](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/branch_picker.rs#L238)
- Expected: Branch picker local remote current filter;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Branch picker local remote current filter. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/pirpc`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V026/rust.png` ↔ `../probes/visual/V026/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V027 · Branch picker no results create branch
- Features: W06, S07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/branch_picker.rs::render:L238](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/branch_picker.rs#L238)
- Expected: Branch picker no results create branch;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Branch picker no results create branch. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent`, `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/pirpc`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V027/rust.png` ↔ `../probes/visual/V027/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V028 · Open in app menu discovered apps
- Features: W66, W67, W68. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/open_in.rs::render_open_in_control:L126](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L126)
- Expected: Open in app menu discovered apps;  Source labels, icons, selected and disabled states visible;  No unrelated dialog or stale response.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Open in app menu discovered apps. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V028/rust.png` ↔ `../probes/visual/V028/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Verify exact open-in symbol; OS discovery is platform dependent.

## V029 · Open in no supported app control absent
- Features: W66, W67, W68. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/open_in.rs::render_open_in_control:L126](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/open_in.rs#L126)
- Expected: No installed supported app; Open-in control absent, not an invented empty menu.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Open in no supported app control absent. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/native`, `internal/preferences`, `internal/workbench`
- Planned pair: `../probes/visual/V029/rust.png` ↔ `../probes/visual/V029/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
- Note: Verify exact open-in symbol; OS discovery is platform dependent.

## V467 · Workbench sidebar review terminal layout
- Features: S02, W01, W03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Sidebar/Review/terminal simultaneously visible;  Source dividers and sizes.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workbench sidebar review terminal layout. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V467/rust.png` ↔ `../probes/visual/V467/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V468 · Workbench files explorer layout
- Features: W01, W03. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Files main area plus Explorer;  Docked Review closed.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workbench files explorer layout. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/workbench`
- Planned pair: `../probes/visual/V468/rust.png` ↔ `../probes/visual/V468/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V469 · Workbench focus traversal
- Features: S05, W71. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/view.rs::render:L166](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/view.rs#L166)
- Expected: Visible focused input/control ring;  Active surface unchanged.
- Setup: Open the isolated shell surface in fixture.json. Configure the named source state: Workbench focus traversal. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V469/rust.png` ↔ `../probes/visual/V469/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
