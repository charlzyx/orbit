# Editor disk-line summary and actual save status (unreleased delta)

## Source and observed issue

Native captures at Go `987eacb` vs Rust `299a32bc` WB03/WB04 show the same
fixture text and editable blank terminal line. Rust header says 5 lines; Go
footer says 6. This is **not evidence of an extra byte or line written to disk**.
Rust `crates/orbit-pi/src/explorer/viewer.rs:114,189–195,932` uses loaded
`FileContent.lines`, dropping only the empty split item caused by a terminal LF.
Its editor retains the actual LF and an editable final line. `on_editor_changed`
(680–700) does not update disk metadata: unsaved edits mark a separate save state;
loaded metadata changes after save/watch refresh.

Go already uses the same file-content split rule in `internal/workspace/files.go`
and updates `Content.Lines` on successful save in `editor.go`. The incorrect
comparison was the UI footer's editor `Buffer.LineCount()` versus Rust's disk
metadata. The new header exposes the existing disk line count; the footer keeps
live Ln/Col without a misleading second total. Unsaved multiline insertion and
undo intentionally leave the loaded-disk header count unchanged. Successful save
updates it. No buffer bytes, IO, line endings, cursor, wrapping or focus change.

## Save state

Rust `viewer.rs:991–1011` chooses read-only, disk conflict, then Saved/Unsaved/
Saving/Failed. Go had shown `Saving…` for every Dirty document, including a failed
save. The bounded footer correction uses actual `Document` flags:
read-only → conflict → active Saving → LastError → Dirty/Unsaved → Saved.
An active retry supersedes its old failure, as Rust's SaveState::Saving does.
Raw filesystem error details remain unmodified; existing source locale keys are
reused. Existing explicit Save button and Wrap control remain unchanged.

## Verification

Actual mounted RED tests (`editor-status-red-1854.log`) demonstrated missing
5-line header and failed-save status incorrectly labeled as Saving. GREEN tests
cover actual mounted source count while preserving six editable buffer lines;
real typing, unsaved metadata, undo/redo, saved count update, exact terminal LF
bytes; real failing save through a temporary directory at the fixture file path;
and real external-write conflict retaining the user's synthetic draft.

Status priority table covers active retry/in-flight/error/read-only combinations;
it is not a claim that the whole UI blocked a real disk write. Existing workspace
`TestSlowSaveDoesNotBlockSnapshotsOrLoseConcurrentEdits` exercises actual in-flight
Saving via the package's controlled writer seam and is included in the package
race gate. The native editor/Explorer regression subset is also rerun.

Logs: `/workspace/shared/orbit-toolchains-continued/editor-status-*185*.log`.
Final combined-tree aggregate and new hosted native WB03/WB04 captures are still
parent-owned. No complete header/tab/icon visual parity claim; no publication.

## Image-mode boundary confirmed in independent review

The pinned Rust source does **not** exclude images from its line metadata:
Image content has `lines: Vec::new()` (viewer.rs158–160), toolbar only excludes
Binary/TooLarge (955), and Render includes toolbar unconditionally (1426).
Therefore image headers retain source's `0 lines` behavior rather than silently
introducing a different rule. The additional real PNG mounted test asserts that
header, actual distinctive raster pixels in the body, and unchanged PNG bytes.
No image-dimensions-only header is claimed for this Rust pin.
