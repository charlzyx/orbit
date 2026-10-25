# Palette-only idle Abort availability

Source pinned Rust command_palette.rs:259–278 gates Abort with snapshot.busy;
:451–455 skips disabled commands entirely (hidden, not a disabled row).
app/pickers.rs:313–321 fills busy from is_running(), which session.rs:42–44
includes busy/streaming/awaiting. Go Manager.Snapshot manager.go:1139 includes
awaiting in its public Running field; this is also Controller's existing Abort
wire admission flag (run.go:531–538).

Only palette result projection skips AbortRun when !Model.Running. No registry,
key binding, Escape navigation, Controller/agent wire guard or other command
availability is changed. No extra path/ID/connected predicate: source does not
require a saved session for aborting an in-flight run. Compaction/retry flags
alone are not the source busy predicate; with Running they remain visible.

`palette-idle-red-1957.log` actual mounted idle/compaction-only/retry-only/no-current
idle and running→settled cases failed before the fix (row still offered).
Tests mount running and awaiting-projected public snapshots, each flag with and
without Running, and missing-current-session conditions; enabled entries emit
AbortRun. The awaiting test validates the UI's public projection, not a newly
executed Pi transport awaiting transition. Existing Manager.Snapshot source
supplies that projection. Tests retain settings/worktrees Escape→chat behavior
and nil-Agent idle Controller admission (no wire call).

This does not implement all Rust palette enabled rules, search, dynamic menus,
RefreshSessions, Pin or Delete. No OS menu/physical Mac or complete palette
parity is claimed.
