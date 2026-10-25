# Visual inventory: transcript (V194–V212)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V194 · Transcript compaction open
- Features: C12, C14. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_summary_card:L1811](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1811)
- Expected: Expanded compaction summary body.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript compaction open. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent`, `internal/pirpc`, `internal/ui`
- Planned pair: `../probes/visual/V194/rust.png` ↔ `../probes/visual/V194/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V195 · Transcript branch summary folded
- Features: C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_summary_card:L1811](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1811)
- Expected: Collapsed branch summary.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript branch summary folded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V195/rust.png` ↔ `../probes/visual/V195/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V196 · Transcript branch summary open
- Features: C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_summary_card:L1811](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1811)
- Expected: Expanded branch summary body.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript branch summary open. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V196/rust.png` ↔ `../probes/visual/V196/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V197 · Transcript changed files one
- Features: C12, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_changed_files:L7053](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)
- Expected: One changed file;  Relative path;  Added/deleted counts;  Review.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript changed files one. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/ui`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V197/rust.png` ↔ `../probes/visual/V197/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V198 · Transcript changed files three
- Features: C12, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_changed_files:L7053](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)
- Expected: Three file rows;  Source no-expand boundary.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript changed files three. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/ui`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V198/rust.png` ↔ `../probes/visual/V198/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V199 · Transcript changed files four folded
- Features: C12, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_changed_files:L7053](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)
- Expected: Four changed files;  Folded threshold affordance.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript changed files four folded. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/ui`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V199/rust.png` ↔ `../probes/visual/V199/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V200 · Transcript changed files twelve open
- Features: C12, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_changed_files:L7053](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)
- Expected: Twelve file rows;  Expanded boundary;  Single final footer.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript changed files twelve open. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/ui`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V200/rust.png` ↔ `../probes/visual/V200/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V201 · Transcript changed files thirteen open
- Features: C12, W30. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_changed_files:L7053](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L7053)
- Expected: Thirteen files;  Source cap/overflow presentation;  Single footer.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript changed files thirteen open. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/agent/manager.go`, `internal/ui`, `internal/ui/review.go`, `internal/workspace/checkpoint.go`, `internal/workspace/review.go`
- Planned pair: `../probes/visual/V201/rust.png` ↔ `../probes/visual/V201/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V202 · Transcript footer missing telemetry
- Features: C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Unavailable telemetry omitted honestly;  No synthetic token/cost values.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript footer missing telemetry. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V202/rust.png` ↔ `../probes/visual/V202/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V203 · Transcript footer telemetry and time gallery
- Features: C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Actual token/cache/rate/cost;  Today/Yesterday/older dates in visible compact rows.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript footer telemetry and time gallery. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V203/rust.png` ↔ `../probes/visual/V203/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V204 · Transcript usage breakdown hover
- Features: C12. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::usage_breakdown_card:L4499](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L4499)
- Expected: Input/output/cache/cost breakdown;  Anchored hover card.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript usage breakdown hover. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V204/rust.png` ↔ `../probes/visual/V204/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V205 · Transcript code known language
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_code_block:L6537](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L6537)
- Expected: Known language syntax;  Numbered rows;  Copy button;  Horizontal scroll.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript code known language. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V205/rust.png` ↔ `../probes/visual/V205/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V206 · Transcript code unknown language
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_code_block:L6537](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L6537)
- Expected: Unknown language plain code;  Numbered rows;  Same metrics.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript code unknown language. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V206/rust.png` ↔ `../probes/visual/V206/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V207 · Transcript code copied
- Features: C09. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_code_block:L6537](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L6537)
- Expected: Copied feedback;  Code remains unchanged.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript code copied. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V207/rust.png` ↔ `../probes/visual/V207/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V208 · Transcript selection cross block
- Features: C10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Unicode cross-block selection;  Link span;  Selection highlight.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript selection cross block. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V208/rust.png` ↔ `../probes/visual/V208/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V209 · Transcript selection reversed
- Features: C10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Reversed selection endpoints;  Same selected text extent.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript selection reversed. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V209/rust.png` ↔ `../probes/visual/V209/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V210 · Transcript selection context menu enabled
- Features: C10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Copy Selection enabled;  Copy Message;  Source menu geometry.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript selection context menu enabled. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V210/rust.png` ↔ `../probes/visual/V210/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V211 · Transcript selection context menu disabled
- Features: C10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: No selection;  Copy Selection disabled;  Copy Message available.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript selection context menu disabled. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V211/rust.png` ↔ `../probes/visual/V211/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V212 · Transcript selection endpoint offscreen
- Features: C08, C10. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/transcript_view.rs::render_row:L1746](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/transcript_view.rs#L1746)
- Expected: Selected endpoint outside viewport;  Remaining highlight preserved.
- Setup: Open the isolated transcript surface in fixture.json. Configure the named source state: Transcript selection endpoint offscreen. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`
- Planned pair: `../probes/visual/V212/rust.png` ↔ `../probes/visual/V212/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
