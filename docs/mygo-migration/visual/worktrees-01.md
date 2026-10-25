# Visual inventory: worktrees (V431–V453)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V431 · Worktrees list baseline
- Features: W40, W51. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::render_worktrees_page:L1042](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1042)
- Expected: Main/linked/external rows;  Name/path/branch;  Active markers.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees list baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V431/rust.png` ↔ `../probes/visual/V431/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V432 · Worktrees marker gallery
- Features: W40. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktrees_list:L1583](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1583)
- Expected: Bare/detached/locked/prunable rows;  Reasons.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees marker gallery. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V432/rust.png` ↔ `../probes/visual/V432/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V433 · Worktrees intro guide
- Features: W51. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_howto_card:L1506](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1506)
- Expected: Introductory guide;  Dismiss control.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees intro guide. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V433/rust.png` ↔ `../probes/visual/V433/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V434 · Worktrees empty
- Features: W51. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_empty_state:L3103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L3103)
- Expected: No available linked worktrees;  Create action.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V434/rust.png` ↔ `../probes/visual/V434/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V435 · Worktrees loading
- Features: W51. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_empty_state:L3103](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L3103)
- Expected: Loading/busy state;  No fake worktree rows.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees loading. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V435/rust.png` ↔ `../probes/visual/V435/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V436 · Worktrees refresh error
- Features: W51. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktrees_error_banner:L1460](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1460)
- Expected: Refresh failure;  Retry;  Existing context.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees refresh error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V436/rust.png` ↔ `../probes/visual/V436/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V437 · Worktrees row menu
- Features: W49. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_row_menu:L2927](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2927)
- Expected: Open;  Rename/Move;  Lock/Unlock;  Remove;  Reveal/Copy.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktrees row menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V437/rust.png` ↔ `../probes/visual/V437/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V438 · Worktree create existing branch
- Features: W41. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::create_worktree_dialog_body:L1875](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1875)
- Expected: Directory/display name;  Existing branch;  Setup/open choices.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree create existing branch. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V438/rust.png` ↔ `../probes/visual/V438/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V439 · Worktree create new branch
- Features: W41. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::create_worktree_dialog_body:L1875](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1875)
- Expected: Independent branch name;  Start point;  Auto-derived until edited.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree create new branch. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V439/rust.png` ↔ `../probes/visual/V439/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V440 · Worktree create remote branch
- Features: W41. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::create_worktree_dialog_body:L1875](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1875)
- Expected: Remote branch selection;  Local tracking behavior labels.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree create remote branch. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V440/rust.png` ↔ `../probes/visual/V440/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V441 · Worktree create validation error
- Features: W42. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_dialog_error:L1863](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1863)
- Expected: Invalid name/path/branch error;  Draft retained.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree create validation error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V441/rust.png` ↔ `../probes/visual/V441/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V442 · Worktree allow setup
- Features: W44. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::allow_setup_dialog_body:L2116](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2116)
- Expected: Repository script;  Explicit Allow;  Cancel.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree allow setup. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V442/rust.png` ↔ `../probes/visual/V442/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V443 · Worktree setup running
- Features: W45. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_setup_banner:L1301](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1301)
- Expected: Setup running;  Captured output;  No premature success.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree setup running. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V443/rust.png` ↔ `../probes/visual/V443/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V444 · Worktree setup success expanded
- Features: W45. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_setup_banner:L1301](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1301)
- Expected: Successful exit;  Expanded stdout/stderr;  Open/dismiss.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree setup success expanded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V444/rust.png` ↔ `../probes/visual/V444/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V445 · Worktree setup failure expanded
- Features: W45. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_setup_banner:L1301](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L1301)
- Expected: Exit failure;  Captured output;  Retry/open/dismiss.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree setup failure expanded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V445/rust.png` ↔ `../probes/visual/V445/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V446 · Worktree unavailable session
- Features: W46. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::worktree_unavailable_dialog_body:L2316](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2316)
- Expected: Missing recorded cwd;  Locate/worktrees actions;  No silent replacement.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree unavailable session. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V446/rust.png` ↔ `../probes/visual/V446/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V447 · Worktree branch in use
- Features: W41, W42. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::branch_in_use_dialog_body:L2373](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2373)
- Expected: Branch already checked out;  Existing worktree action.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree branch in use. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V447/rust.png` ↔ `../probes/visual/V447/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V448 · Worktree rename dialog
- Features: W47. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::rename_worktree_dialog_body:L2159](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2159)
- Expected: Current name;  New name field;  Confirm/Cancel.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree rename dialog. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V448/rust.png` ↔ `../probes/visual/V448/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V449 · Worktree move dialog
- Features: W47. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::move_worktree_dialog_body:L2197](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2197)
- Expected: Explicit destination path;  Confirm/Cancel.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree move dialog. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V449/rust.png` ↔ `../probes/visual/V449/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V450 · Worktree remove clean confirm
- Features: W48. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::remove_worktree_dialog_body:L2235](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2235)
- Expected: Clean worktree path;  Explicit removal;  Branch retained.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree remove clean confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V450/rust.png` ↔ `../probes/visual/V450/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V451 · Worktree remove dirty confirm
- Features: W48. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::remove_worktree_dialog_body:L2235](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2235)
- Expected: Dirty warning;  Force-removal disclosure;  Cancel.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Worktree remove dirty confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V451/rust.png` ↔ `../probes/visual/V451/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V452 · Work in picker
- Features: W50. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::work_in_popup:L2764](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2764)
- Expected: Local/main;  Linked worktrees;  New;  Current choice.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Work in picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V452/rust.png` ↔ `../probes/visual/V452/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V453 · Work in picker keyboard selection
- Features: W50. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/app/worktrees.rs::work_in_popup:L2764](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/app/worktrees.rs#L2764)
- Expected: Keyboard-highlighted different row;  Current-workspace marker retained.
- Setup: Open the isolated worktrees surface in fixture.json. Configure the named source state: Work in picker keyboard selection. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/workspace/setup.go`, `internal/workspace/worktrees.go`, `internal/worktreeui/labels.go`, `internal/worktreeui/picker.go`, `internal/worktreeui/state.go`, `internal/worktreeui/view.go`
- Planned pair: `../probes/visual/V453/rust.png` ↔ `../probes/visual/V453/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
