# C04 plus menu: Mention file

Rust `app.rs:125` defines three add-menu rows: attach image, attach file,
mention file. `composer_ops.rs:586–606` inserts `@` at the active caret and
focuses the composer; the normal file autocomplete opens from that trigger.
The Go menu lacked this third action.

The additive menu action reuses the reviewed owner-safe native edit queue from
`attachment-caret-acceptance.md`; it does not submit a prompt or perform external
tool execution. The native menu button has an opaque attachment-owner key:
MyGo `menu.go` builds menus when opened and again after selection, so a stable
unowned button could otherwise apply an old choice to a new task. All existing
menu actions remain available. No image decoding or transcript changes.

Mounted regressions cover the real button/menu path, Unicode active caret,
independent undo/redo, a real isolated project file index opening completion,
cancel/repeat, stale native selection after a task generation change, and IME
composition committed before insertion. MyGo's menu/editor click handling
commits marked input when opening the menu; typing it again would be a second
input, not an IME commit. A separate caret test covers delayed insertion when
composition remains active.

This is not a claim of full plus-menu parity: the source's image-only native
picker is still a separate gap. Native desktop menu rendering and physical
platform IME remain distinct from mounted pinned-editor tests.
