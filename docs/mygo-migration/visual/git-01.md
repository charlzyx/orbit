# Visual inventory: git (V305–V329)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V305 · Git changes staged unstaged
- Features: W04, W05, W08, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::changes_tab:L3892](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3892)
- Expected: Six tabs;  Branch row;  Staged and unstaged sections;  Per-file and bulk actions.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git changes staged unstaged. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/agent/manager.go`, `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V305/rust.png` ↔ `../probes/visual/V305/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V306 · Git changes clean
- Features: W04, W05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::changes_tab:L3892](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3892)
- Expected: Clean working tree;  Honest empty state;  State-driven footer.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git changes clean. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V306/rust.png` ↔ `../probes/visual/V306/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V307 · Git not repository
- Features: W04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::render:L7098](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L7098)
- Expected: Not a Git repository;  No fake branch/status.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git not repository. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V307/rust.png` ↔ `../probes/visual/V307/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V308 · Git unborn branch
- Features: W05, W10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_bar:L3299](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3299)
- Expected: Unborn branch;  Initial commit affordance.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git unborn branch. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V308/rust.png` ↔ `../probes/visual/V308/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V309 · Git detached head
- Features: W05. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_bar:L3299](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3299)
- Expected: Detached HEAD identity;  Unavailable branch-only actions.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git detached head. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V309/rust.png` ↔ `../probes/visual/V309/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V310 · Git loading
- Features: W04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::render:L7098](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L7098)
- Expected: Initial tab loading state;  No stale fake rows.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git loading. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V310/rust.png` ↔ `../probes/visual/V310/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V311 · Git refresh busy
- Features: W04. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::tab_bar:L3202](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3202)
- Expected: Refresh indicator;  Active tab retained.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git refresh busy. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V311/rust.png` ↔ `../probes/visual/V311/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V312 · Git branch more menu
- Features: W07, W12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_more_menu:L3560](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3560)
- Expected: New/rename/delete branch;  Merge/rebase actions.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git branch more menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V312/rust.png` ↔ `../probes/visual/V312/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V313 · Git new branch prompt
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_prompt_popup:L6793](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6793)
- Expected: New branch name;  Confirm/Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git new branch prompt. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V313/rust.png` ↔ `../probes/visual/V313/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V314 · Git rename branch prompt
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::branch_prompt_popup:L6793](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6793)
- Expected: Current branch name;  Rename draft.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git rename branch prompt. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V314/rust.png` ↔ `../probes/visual/V314/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V315 · Git branch invalid failure banner
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Branch prompt closed after invalid confirm; Persistent failure banner with error text.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git branch invalid failure banner. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V315/rust.png` ↔ `../probes/visual/V315/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V316 · Git delete branch confirm
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::confirm_popup:L6983](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6983)
- Expected: Safe delete refused unmerged branch; Force-delete named branch confirmation; Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git delete branch confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V316/rust.png` ↔ `../probes/visual/V316/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V317 · Git merge ref picker
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::ref_picker_popup:L6674](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6674)
- Expected: Local/remote/tag refs;  Search;  Merge mode choices.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git merge ref picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V317/rust.png` ↔ `../probes/visual/V317/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V318 · Git rebase ref picker
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::ref_picker_popup:L6674](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6674)
- Expected: Searchable refs;  Rebase action.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git rebase ref picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V318/rust.png` ↔ `../probes/visual/V318/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V319 · Git ref picker no match
- Features: W07. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::ref_picker_popup:L6674](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6674)
- Expected: Query retained;  No matching refs.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git ref picker no match. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V319/rust.png` ↔ `../probes/visual/V319/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V320 · Git discard file confirm
- Features: W09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::confirm_popup:L6983](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6983)
- Expected: Named path;  Destructive warning;  Confirm/Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git discard file confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V320/rust.png` ↔ `../probes/visual/V320/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V321 · Git discard all confirm
- Features: W09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::confirm_popup:L6983](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6983)
- Expected: All unstaged warning;  Confirm/Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git discard all confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V321/rust.png` ↔ `../probes/visual/V321/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V322 · Git commit draft
- Features: W10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::commit_bar:L3905](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3905)
- Expected: Editable message;  Include unstaged;  Commit;  Commit and push.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git commit draft. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V322/rust.png` ↔ `../probes/visual/V322/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V323 · Git generate stage choice
- Features: W11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::stage_prompt_popup:L6873](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6873)
- Expected: Stage all versus already staged;  Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git generate stage choice. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V323/rust.png` ↔ `../probes/visual/V323/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V324 · Git generating commit
- Features: W11. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::commit_bar:L3905](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3905)
- Expected: Generating state;  Duplicate action disabled.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git generating commit. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V324/rust.png` ↔ `../probes/visual/V324/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V325 · Git commit push publish states
- Features: W10, W12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::commit_bar:L3905](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3905)
- Expected: Source action strip for upstream/ahead state;  Exact visible current action.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git commit push publish states. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V325/rust.png` ↔ `../probes/visual/V325/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V326 · Git sync menu
- Features: W12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::sync_menu:L3440](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3440)
- Expected: Fetch;  Pull;  Merge/rebase upstream;  Sync;  Force-with-lease.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git sync menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V326/rust.png` ↔ `../probes/visual/V326/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V327 · Git force push confirm
- Features: W12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::confirm_popup:L6983](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6983)
- Expected: Explicit force-with-lease confirmation;  Cancel.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git force push confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V327/rust.png` ↔ `../probes/visual/V327/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V328 · Git failure push rejected
- Features: W13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Copyable persistent failure;  Matching pull/rebase recovery.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git failure push rejected. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V328/rust.png` ↔ `../probes/visual/V328/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V329 · Git failure auth network
- Features: W13. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Authentication/network detail;  Matching recovery only.
- Setup: Open the isolated git surface in fixture.json. Configure the named source state: Git failure auth network. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V329/rust.png` ↔ `../probes/visual/V329/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
