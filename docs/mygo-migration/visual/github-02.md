# Visual inventory: github (V370–V377)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V370 · Github pull closed reopen
- Features: W27. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_detail_view:L5686](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L5686)
- Expected: Closed state;  Reopen action.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull closed reopen. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V370/rust.png` ↔ `../probes/visual/V370/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V371 · Github pull checkout dirty error
- Features: W27. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::failure_banner:L3630](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L3630)
- Expected: Checkout refused for dirty worktree;  No false success.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull checkout dirty error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V371/rust.png` ↔ `../probes/visual/V371/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V372 · Github new pull
- Features: W28, W29. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_new_view:L6279](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6279)
- Expected: Title/body/notes;  Draft toggle;  Base branch;  Template;  Generate;  Submit/Cancel.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github new pull. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V372/rust.png` ↔ `../probes/visual/V372/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V373 · Github pull base picker
- Features: W28. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_base_popup:L1666](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1666)
- Expected: Anchored searchable local-branch picker;  Current base.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull base picker. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V373/rust.png` ↔ `../probes/visual/V373/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V374 · Github pull base no match
- Features: W28. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_base_popup:L1666](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1666)
- Expected: Unmatched query;  No branch result.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull base no match. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V374/rust.png` ↔ `../probes/visual/V374/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V375 · Github pull template menu
- Features: W28. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::template_picker_popup:L1479](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1479)
- Expected: Discovered PR templates;  None;  Selected template.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull template menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V375/rust.png` ↔ `../probes/visual/V375/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V376 · Github pull generating
- Features: W29. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_new_view:L6279](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6279)
- Expected: Generating draft;  Duplicate disabled;  Notes retained.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull generating. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V376/rust.png` ↔ `../probes/visual/V376/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V377 · Github pull submit error
- Features: W28. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/git_panel/mod.rs::pr_new_view:L6279](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L6279)
- Expected: Failed create keeps draft open;  Actionable error.
- Setup: Open the isolated github surface in fixture.json. Configure the named source state: Github pull submit error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/gitworkbench/state.go`, `internal/gitworkbench/templates.go`, `internal/gitworkbench/view.go`, `internal/workspace/generation.go`, `internal/workspace/git.go`, `internal/workspace/git_ops.go`, `internal/workspace/github.go`, `internal/workspace/github_templates.go`
- Planned pair: `../probes/visual/V377/rust.png` ↔ `../probes/visual/V377/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
