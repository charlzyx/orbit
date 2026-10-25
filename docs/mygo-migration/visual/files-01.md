# Visual inventory: files (V401–V425)

Initial source-state register. Recipes are not yet executed; no matched pair is complete. Common environment: [fixture.json](fixture.json).

## V401 · Explorer tree baseline
- Features: W01, W52, W53. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render:L60](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L60)
- Expected: Nested folders;  Empty directory;  Git status/dirty ancestor;  Active file.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer tree baseline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `cmd/orbit/main.go`, `internal/ui`, `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V401/rust.png` ↔ `../probes/visual/V401/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V402 · Explorer hidden files shown
- Features: W52. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render:L60](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L60)
- Expected: Hidden files enabled;  Ignored paths excluded per source.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer hidden files shown. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V402/rust.png` ↔ `../probes/visual/V402/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V403 · Explorer filtered tree
- Features: W52. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render:L60](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L60)
- Expected: Matching file and ancestors;  Query retained.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer filtered tree. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V403/rust.png` ↔ `../probes/visual/V403/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V404 · Explorer filter no match
- Features: W52. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render:L60](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L60)
- Expected: No matching file;  Filter input.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer filter no match. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V404/rust.png` ↔ `../probes/visual/V404/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V405 · Explorer keyboard selection
- Features: W53. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render_row:L1044](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1044)
- Expected: Keyboard selected row;  Expanded ancestor;  Active-file highlight.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer keyboard selection. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V405/rust.png` ↔ `../probes/visual/V405/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V406 · Explorer folder context menu
- Features: W54. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::context_menu:L1217](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1217)
- Expected: Expand/collapse;  New File/Folder;  Rename/Delete;  Open/reveal/copy.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer folder context menu. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V406/rust.png` ↔ `../probes/visual/V406/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V407 · Explorer file context menu flipped
- Features: W54. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::context_menu:L1217](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1217)
- Expected: Pointer near bottom;  Menu flips into viewport;  File actions.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer file context menu flipped. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V407/rust.png` ↔ `../probes/visual/V407/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V408 · Explorer new file inline
- Features: W55. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render_row:L1044](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1044)
- Expected: Inline new-file field in selected parent;  Confirm/cancel hints.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer new file inline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V408/rust.png` ↔ `../probes/visual/V408/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V409 · Explorer new folder inline
- Features: W55. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render_row:L1044](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1044)
- Expected: Inline folder field;  Existing siblings retained.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer new folder inline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V409/rust.png` ↔ `../probes/visual/V409/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V410 · Explorer rename inline
- Features: W56. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render_row:L1044](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1044)
- Expected: Current filename selected in inline field.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer rename inline. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V410/rust.png` ↔ `../probes/visual/V410/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V411 · Explorer delete confirm
- Features: W56. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::delete_prompt:L1426](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L1426)
- Expected: Named file/folder;  Delete and Cancel.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer delete confirm. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V411/rust.png` ↔ `../probes/visual/V411/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V412 · Explorer operation error
- Features: W55, W56. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::notice_banner:L998](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L998)
- Expected: Invalid/failed mutation notice;  Existing tree preserved.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer operation error. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V412/rust.png` ↔ `../probes/visual/V412/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V413 · Explorer empty root
- Features: W52. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/panel.rs::render:L60](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/panel.rs#L60)
- Expected: Empty directory state;  Creation controls.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Explorer empty root. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V413/rust.png` ↔ `../probes/visual/V413/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V414 · Files editor tabs
- Features: W57, W58. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::render:L1383](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1383)
- Expected: Multiple file tabs;  Active editor;  Gutter;  Relative path;  Save state.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files editor tabs. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V414/rust.png` ↔ `../probes/visual/V414/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V415 · Files tabs overflow
- Features: W57. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::tab_strip:L768](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L768)
- Expected: Overflowing tabs;  Horizontal strip scroll;  Active tab visible.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files tabs overflow. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V415/rust.png` ↔ `../probes/visual/V415/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V416 · Files code selection
- Features: W58. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::editor_body:L1367](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1367)
- Expected: Syntax-painted editable buffer;  Selection;  Caret;  No wrap.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files code selection. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V416/rust.png` ↔ `../probes/visual/V416/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V417 · Files code horizontal scroll
- Features: W58. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::editor_body:L1367](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1367)
- Expected: Long source line;  Horizontal scroll;  Stable gutter.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files code horizontal scroll. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V417/rust.png` ↔ `../probes/visual/V417/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V418 · Files dirty unsaved
- Features: W59. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::status_chip:L992](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L992)
- Expected: Unsaved chip;  Dirty tab state.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files dirty unsaved. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V418/rust.png` ↔ `../probes/visual/V418/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V419 · Files saving
- Features: W59. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::status_chip:L992](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L992)
- Expected: Saving chip;  Buffer retained.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files saving. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V419/rust.png` ↔ `../probes/visual/V419/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V420 · Files save failed
- Features: W59. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::status_chip:L992](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L992)
- Expected: Failed save chip/error;  Dirty buffer retained.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files save failed. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V420/rust.png` ↔ `../probes/visual/V420/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V421 · Files changed on disk
- Features: W59. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::status_chip:L992](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L992)
- Expected: Changed-on-disk warning;  Dirty buffer not overwritten.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files changed on disk. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V421/rust.png` ↔ `../probes/visual/V421/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V422 · Files markdown preview
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: Rendered live Markdown;  Source/preview toggle.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files markdown preview. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V422/rust.png` ↔ `../probes/visual/V422/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V423 · Files markdown source
- Features: W58, W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::editor_body:L1367](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1367)
- Expected: Editable Markdown source;  Same live buffer.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files markdown source. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V423/rust.png` ↔ `../probes/visual/V423/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V424 · Files raster image
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: PNG/JPEG fixture image;  Correct fit;  File size.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files raster image. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V424/rust.png` ↔ `../probes/visual/V424/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.

## V425 · Files vector image
- Features: W60. Lane: `canonical`. Visual: **UNTESTED**. Native OS: **UNTESTED**.
- Rust: [crates/orbit-pi/src/explorer/viewer.rs::body:L1031](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/explorer/viewer.rs#L1031)
- Expected: SVG fixture;  Correct fit;  No fake placeholder.
- Setup: Open the isolated files surface in fixture.json. Configure the named source state: Files vector image. Use the smallest deterministic data set that makes every expected control visible; scroll to the relevant region without changing client size. Freeze clock/animation and capture the complete client at1470x922. If controls cannot coexist visibly, append another stable case instead of claiming this image covers hidden controls.
- Go mapping: `internal/ui/files.go`, `internal/workspace/editor.go`, `internal/workspace/files.go`, `internal/workspace/tree.go`, `internal/workspace/watch.go`
- Planned pair: `../probes/visual/V425/rust.png` ↔ `../probes/visual/V425/go.png`. Neither image exists in this register.
- Metrics: not measured. Manual review: not reviewed. Last verified Go SHA: none.
