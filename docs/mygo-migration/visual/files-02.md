# Visual inventory: files (V426–V430)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V426 · Files readonly lossy truncated
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::text_body:L1145](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1145)
- Expected: Read-only decoded/truncated content;  Warning and toolbar metadata.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files readonly lossy truncated. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V426/rust.png` ↔ `../probes/visual/V426/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V427 · Files binary
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: Binary content unavailable message;  No gibberish editor.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files binary. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V427/rust.png` ↔ `../probes/visual/V427/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V428 · Files too large
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: Too-large state;  Size limit message.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files too large. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V428/rust.png` ↔ `../probes/visual/V428/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V429 · Files read error
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: Unreadable file error;  Path retained.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files read error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V429/rust.png` ↔ `../probes/visual/V429/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V430 · Files loading
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::loading_state:L1252](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1252)
- Expected: Loading indicator;  Active tab/path context.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files loading. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V430/rust.png` ↔ `../probes/visual/V430/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
