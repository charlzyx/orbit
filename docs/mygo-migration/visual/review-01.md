# Visual inventory: review (V378–V400)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V378 · Review docked unified
- Features: W01, W03, W31, W34, W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: Docked diff;  File tree;  Old/new gutters;  Addition/deletion/context colors.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review docked unified. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/gitworkbench/embedded_diff.go`, `internal/highlight`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/ui/review.go`, `internal/workbench`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V378/rust.png` ↔ `../probes/visual/V378/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V379 · Review full width
- Features: W01, W03, W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: Full-width Review;  Sessions sidebar;  Restore-dock control.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review full width. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/native`, `internal/preferences`, `internal/ui`, `internal/ui/review.go`, `internal/workbench`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V379/rust.png` ↔ `../probes/visual/V379/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V380 · Review source menu
- Features: W31. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::source_menu:L2113](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L2113)
- Expected: Last Turn;  Uncommitted;  Unstaged;  Staged;  Committed;  Branch.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review source menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V380/rust.png` ↔ `../probes/visual/V380/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V381 · Review last turn disabled
- Features: W31, W32. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::source_menu:L2113](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L2113)
- Expected: No completed turn;  Last Turn disabled.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review last turn disabled. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V381/rust.png` ↔ `../probes/visual/V381/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V382 · Review historical turn selected
- Features: W31, W32, W33. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_content:L1345](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1345)
- Expected: Concrete Turn N label;  Selected completed diff remains pinned.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review historical turn selected. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V382/rust.png` ↔ `../probes/visual/V382/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V383 · Review staged unstaged sections
- Features: W08, W34. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: Section headers;  Per-file stage/unstage;  Section actions.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review staged unstaged sections. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/gitworkbench/embedded_diff.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/highlight`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V383/rust.png` ↔ `../probes/visual/V383/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V384 · Review split diff
- Features: W34, W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: Old/new side-by-side columns;  Matched line rows;  Separate gutters.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review split diff. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/gitworkbench/embedded_diff.go`, `internal/highlight`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V384/rust.png` ↔ `../probes/visual/V384/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V385 · Review wrapped long lines
- Features: W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff_line:L1681](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1681)
- Expected: Long text wraps;  Correct gutters and row heights.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review wrapped long lines. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V385/rust.png` ↔ `../probes/visual/V385/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V386 · Review unwrapped horizontal pan
- Features: W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff_line:L1681](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1681)
- Expected: Horizontal offset;  Pinned header/gutter;  No clipping mismatch.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review unwrapped horizontal pan. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V386/rust.png` ↔ `../probes/visual/V386/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V387 · Review files collapsed
- Features: W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: All file headers collapsed;  Counts retained.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review files collapsed. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V387/rust.png` ↔ `../probes/visual/V387/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V388 · Review sticky header scroll
- Features: W36. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_sticky_header:L1488](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1488)
- Expected: Scrolled file/section header pinned correctly.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review sticky header scroll. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V388/rust.png` ↔ `../probes/visual/V388/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V389 · Review tree filtered
- Features: W37. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_tree:L1874](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1874)
- Expected: Directory tree;  Query;  Matching subtree and active file.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review tree filtered. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V389/rust.png` ↔ `../probes/visual/V389/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V390 · Review tree no match
- Features: W37. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_tree:L1874](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1874)
- Expected: Unmatched query;  Empty tree result.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review tree no match. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V390/rust.png` ↔ `../probes/visual/V390/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V391 · Review tree keyboard cursor
- Features: W38. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_tree_row:L1944](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1944)
- Expected: Keyboard cursor distinct from active file;  Selected hunk visible.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review tree keyboard cursor. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V391/rust.png` ↔ `../probes/visual/V391/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V392 · Review context gap collapsed
- Features: W35. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_gap:L1751](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1751)
- Expected: Hidden unchanged-line count;  Up/down/both/all controls.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review context gap collapsed. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V392/rust.png` ↔ `../probes/visual/V392/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V393 · Review context gap expanded edge
- Features: W35. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_gap:L1751](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1751)
- Expected: One expanded edge;  Remaining gap count.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review context gap expanded edge. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V393/rust.png` ↔ `../probes/visual/V393/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V394 · Review context all expanded
- Features: W35. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff:L1404](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1404)
- Expected: Full retained context;  Source line numbers.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review context all expanded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V394/rust.png` ↔ `../probes/visual/V394/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V395 · Review binary renamed deleted
- Features: W34. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff_line:L1681](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1681)
- Expected: Binary row;  Rename paths;  Deleted-file header metadata.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review binary renamed deleted. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/gitworkbench/embedded_diff.go`, `internal/highlight`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V395/rust.png` ↔ `../probes/visual/V395/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V396 · Review syntax multiline
- Features: W34. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_diff_line:L1681](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1681)
- Expected: Paint-only syntax spans across multiline tokens;  Red/green backgrounds unchanged.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review syntax multiline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/gitworkbench/embedded_diff.go`, `internal/highlight`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V396/rust.png` ↔ `../probes/visual/V396/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V397 · Review loading
- Features: W39. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_content:L1345](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1345)
- Expected: Immediate loading state;  Selected source label retained.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review loading. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V397/rust.png` ↔ `../probes/visual/V397/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V398 · Review refresh feedback
- Features: W39. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_content:L1345](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1345)
- Expected: Manual refresh spinner at frozen300ms within650ms floor.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review refresh feedback. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V398/rust.png` ↔ `../probes/visual/V398/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V399 · Review source empty
- Features: W31, W39. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_content:L1345](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1345)
- Expected: Selected source-specific empty text;  No stale diff.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review source empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V399/rust.png` ↔ `../probes/visual/V399/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V400 · Review load error
- Features: W39. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/sidepane.rs::render_content:L1345](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L1345)
- Expected: Source-specific error;  Retry;  No stale unrelated diff.
- Setup: Open the isolated review surface in fixture.json. Configure the named source state: Review load error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V400/rust.png` ↔ `../probes/visual/V400/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
