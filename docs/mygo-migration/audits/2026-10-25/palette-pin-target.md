# Palette Pin: active saved-path payload

Pinned Rust app/pickers.rs:437–446 toggles current_session_path. Go command
palette static rows carry empty values; App.command previously forwarded that
empty value to Controller.SetSession, which stored an empty path instead of
pinning the active session.

This tiny independent patch adds only the PinSession command case: capture the
current saved path/session ID immediately, emit the exact target, and do nothing
when no saved path exists. The row-menu commandSession path and Controller's
existing toggle behavior are unchanged. Pin/Unpin dynamic labels, live enabled
filtering, palette search and Delete are explicitly outside this increment.

Actual mounted palette→synthetic Controller→temporary Store RED:
`palette-pin-red-1944.log` stored `["/store/other.jsonl", ""]` and emitted a
pathless action for unsaved session. Corrected tests exercise pointer pin and
keyboard unpin, preserving another pinned session; pathless command emits no
action. No real user files, runtime or OS actions are used.

This is NOT a data migration. Existing empty-key pollution in a real previous
preferences file is not removed or silently rewritten by this patch; any cleanup
needs separate evidence/authorization. Unrelated stored entries are preserved.
