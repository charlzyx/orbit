# C04 attachment paths at the native caret

Pinned Rust `composer_ops.rs` (drop around line 425; attach files around 488)
uses `insert_at_caret(path + " ")` for non-images. `composer.rs:379` inserts
at the active cursor end, not over selected text. Images remain attachments.

Go now retains only pending path text and the existing opaque attachment owner.
It applies each path through the reviewed composer overlay's range-edit API in
the current build or TextKey callback, never retaining either native object.
This preserves Unicode rune offsets, selection text, per-path undo, caret and
focus. Session hydration preserves the managed owner; task generation, window,
workspace and shutdown changes reject obsolete edits. Snapshot adoption and
View prune stale intents even when the composer is hidden. Image decoding and
its ownership checks are unchanged.

IME composition defers range edits and submission. Pinned MyGo Darwin
`surface.go` keyDown filters marked-input candidate keys before OnTextKey and
passes them to `interpretKeyEvents`; committed text arrives as TextInput.
The raw Enter guard mirrors existing completion protection. Mounted tests use
`Tester.Type` while composed (the documented commit event), not cancellation.
The normal executable calls `composerui.RequireCompatibility` before startup:
a missing range-edit overlay is a fail-fast error, not a supported frozen-draft
fallback. No new toolkit changes are included.

Tests cover forward/reversed selections, Unicode and spaces, two independent
undo/redo steps, queued Enter before a new frame, committed IME text, direct
submission while pending, stale owners, normal hydration, mixed images/files
and retained typing focus. The asynchronous dialog test now mounts a composer
frame before checking native insertion rather than assuming direct Draft append.

These are mounted pinned-editor tests, not physical macOS drag/drop or IME
validation. Native Save/selection confirmation remains separately gated; this
change does not establish that the user's reported UI hang is resolved.
