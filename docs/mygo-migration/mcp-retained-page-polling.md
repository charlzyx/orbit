# Retained MCP page polling without a rendered window

Reference: Rust `299a32b`, `app/mcp_ui.rs:90–117` and
`app/settings.rs::set_settings_section` (7415 onward).
This is a separate follow-on to `windowless-host-maintenance.md`, not a change
to its frozen delivery or a claim of full MCP parity.

Rust gates its 60-second external configuration check on retained navigation:
Settings is open and the selected section is MCP. Closing the native window
does not close the retained settings page. Leaving Settings or selecting another
section stops new external checks; already-scheduled work still drains. Opening
the MCP section schedules a refresh. External edits are read-only until the user
explicitly requests apply/restart.

The Go host previously used the component's `lastViewed < 2 seconds` heuristic,
so merely restoring a windowless host tick did not preserve this behavior. The
host now supplies its retained page/section intent with `SetHostVisible` before
`MCP.Tick`. Standalone embedded components retain their existing recent-render
fallback when no host visibility has been supplied. Navigation-driven refresh
was already implemented by `loadSettings("mcp")` and is not duplicated.

The due comparison includes the exact 60-second boundary, matching Rust's
`Instant::now() >= at`. No polling interval is shortened, no MCP process is
launched from the owner callback, and external edits still do not invoke the
host's modified/apply callbacks automatically. As before, workers read files
and perform probes; owner ticks apply their completions.

Evidence uses disposable configuration directories and a synthetic Pi probe
executable. The pre-correction retained-page regression failed without a recent
render. Race-enabled tests cover the 60-second boundary with Go's virtual clock,
page/section exit despite a recent frame, completion draining after exit, and
an external configuration edit followed by explicit manual apply. No user data,
credentials, real MCP endpoints or native macOS window were exercised.

This correction retains the existing component fingerprint/probe semantics;
it does not establish raw-file fingerprint parity in the settings UI, real
server connectivity, or atomic coordination with external filesystem writes.
