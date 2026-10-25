# Visual inventory: git (V330–V344)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V330 · Git failure secret lfs hook
- Features: W13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Safety-specific failure text;  No inappropriate force recovery.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git failure secret lfs hook. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V330/rust.png` ↔ `../probes/visual/V330/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V331 · Git failure unknown
- Features: W13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Raw unknown failure detail;  Copy and dismiss.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git failure unknown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V331/rust.png` ↔ `../probes/visual/V331/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V332 · Git operation conflicts
- Features: W14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::operation_bar:L3755](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3755)
- Expected: Merge/rebase state;  Conflict count;  Up to8 file links;  Continue disabled.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git operation conflicts. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V332/rust.png` ↔ `../probes/visual/V332/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V333 · Git operation resolved
- Features: W14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::operation_bar:L3755](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3755)
- Expected: No conflicts;  Continue;  Applicable Skip only.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git operation resolved. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V333/rust.png` ↔ `../probes/visual/V333/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V334 · Git operation abort confirm
- Features: W14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::confirm_popup:L6983](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6983)
- Expected: Abort operation warning;  Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git operation abort confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V334/rust.png` ↔ `../probes/visual/V334/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V335 · Git stashes list
- Features: W15. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::stashes_tab:L4238](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4238)
- Expected: Newest-first stash rows;  Create;  Pop/Apply/Drop.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git stashes list. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V335/rust.png` ↔ `../probes/visual/V335/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V336 · Git stashes empty
- Features: W15. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::stashes_tab:L4238](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4238)
- Expected: No stashes;  Create action.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git stashes empty. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V336/rust.png` ↔ `../probes/visual/V336/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V337 · Git history list
- Features: W16, W19. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::history_tab:L4312](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4312)
- Expected: Commit rows;  Avatars/initials;  All refs;  Load more.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git history list. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V337/rust.png` ↔ `../probes/visual/V337/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V338 · Git history author path filter
- Features: W16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::history_filter_bar:L4426](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4426)
- Expected: Author/path chips;  Removable filters;  Follow-renames history.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git history author path filter. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V338/rust.png` ↔ `../probes/visual/V338/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V339 · Git history empty filter
- Features: W16. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::history_tab:L4312](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4312)
- Expected: No matching commits;  Filter retained.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git history empty filter. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V339/rust.png` ↔ `../probes/visual/V339/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V340 · Git history expanded commit
- Features: W17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::commit_detail_body:L4494](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4494)
- Expected: Subject/body/hash/author/committer/parents/refs;  Changed-file stats.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git history expanded commit. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V340/rust.png` ↔ `../probes/visual/V340/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V341 · Git history file menu
- Features: W17. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::file_actions_menu:L4612](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4612)
- Expected: Open current file;  Uncommitted diff;  Reveal;  Copy path;  File history.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git history file menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V341/rust.png` ↔ `../probes/visual/V341/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V342 · Git graph local
- Features: W18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::graph_tab:L4694](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4694)
- Expected: Colored branch lanes;  Merge edges;  Ref badges.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git graph local. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V342/rust.png` ↔ `../probes/visual/V342/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V343 · Git graph all refs
- Features: W18. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::graph_tab:L4694](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L4694)
- Expected: Remote-only commits included;  All refs selected;  Load more.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git graph all refs. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V343/rust.png` ↔ `../probes/visual/V343/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V344 · Git account avatar fallback
- Features: W19, W20. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_bar:L3299](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3299)
- Expected: Signed-in account chip;  Stable initials/avatar fallback.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git account avatar fallback. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V344/rust.png` ↔ `../probes/visual/V344/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
