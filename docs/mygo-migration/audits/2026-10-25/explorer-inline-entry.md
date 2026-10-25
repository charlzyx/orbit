# Explorer inline entry: bounded second-wave correction

Base: `5e4e1425` plus frozen Explorer menu patch `orbit-explorer-menu-final-1648.patch`.
Unreleased migration source, not shipped `.4`. No release, main merge, real user files or external applications were used.

## Source contract

Pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`:
- `explorer/panel.rs:89–109,592–687`: one EntryEditor, invalid name retains text/focus, edits clear error, Enter commits and Escape restores tree focus.
- `panel.rs:348–354,418–423,494–499`: absent/filtered rename row cancels its editor; tree keys do not also act while entering a name; context-menu interaction retires an entry.
- `panel.rs:957–1014,1079–1110`: a new-file/folder input is pinned above the tree; rename lives inside its own row; clicking another row cancels rather than activates.
- `app/session.rs:2486–2568`: safe filesystem operation on worker, dirty rename rejection, refresh/open after completion; existing buffers are remapped rather than recreated.

## Implementation boundary

Explorer toolbar/menu now use `files_entry.go`; generic viewer rename prompts remain unchanged. The input is a real single-line native TextInputBase, not a screenshot or web overlay. New entries bind their original root, workspace menu epoch and identity; normal tree polling does not cancel the intent. A workspace roundtrip does. Only the current owner may report completion or open/refresh a file. A rename still reconciles existing buffer-path metadata on the UI owner, but cannot restore old focus after losing ownership.

Validation delegates to existing `workspace.ValidateName`. Create remains exclusive, rename retains existing no-overwrite checks, and both share the existing mutation gate. Duplicate submission is rejected. Dirty state is checked at entry start and again at submission. No delete implementation or confirmation was changed.

Deliberate bounded strengthening: a valid name rejected by the filesystem (e.g. existing destination) retains the inline value for correction/retry; Rust removes its entry before the worker and reports a notice. Cancel before commit writes nothing. Cancel/navigation **after** commit only relinquishes UI ownership; it does not claim to undo the already accepted filesystem operation.

## Verification and limits

- Mounted baseline red: `TestMountedExplorerCreateIsInlineNotModal` failed because the old global modal opened (`inline-red-1649.log`).
- Mounted initial green covers usable inline row, invalid-name retention/error clearing, keyboard Enter/Escape, click-away cancellation, filename collision/retry, actual rename, and workspace-roundtrip rejection (`inline-focused-1652.log`).
- Additional cases cover narrow dock width, selected folder parent, duplicate submission, delayed completion/cancellation, newly dirty buffer and filter disappearance. Race/vet final results are recorded in the handoff; do not infer aggregate acceptance from a focused test.

All fixtures use temporary files. No live user data or delete execution. This is functionality and mounted input/geometry evidence, not pixel parity or hosted macOS acceptance. Existing generic filesystem TOCTOU semantics are unchanged. Final combined source still needs aggregate and native recapture.

Final scoped verification (canonical pinned `scripts/go-native.sh` wrapper):
- `test -race ./internal/ui -run 'Explorer|MountedFileRename|DirtyPaths' -count=1`: PASS, 46.070s (`/workspace/shared/orbit-toolchains-continued/inline-all-race-1654.log`).
- `vet ./internal/ui`: PASS, no diagnostics (`/workspace/shared/orbit-toolchains-continued/inline-vet-1654.log`).
- `git diff --check`: PASS.
