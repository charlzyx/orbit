# Recoverable hidden quota providers

Reference: Rust `299a32b`, `app/session.rs:1738–1749` (partition only the popover),
`1938–1980` (visibility controls), `2035–2088` (collapsed Hidden group), and
`quota_hidden.rs:1–88` (provider-ID preference format and best-effort persistence).
Go base: `5e4e1425`; this is an isolated unreleased increment.

## Corrected functional gap

The quota popover now provides native hide/show controls. Hidden reports move
into a collapsed, recoverable group; opening that group shows the same reported
windows/balances/errors at the source's 0.6 opacity, with a restore action. The
visible provider count follows the same partition. Hiding every provider does
not remove the usage control or its recovery group.

Visibility is keyed by provider ID in Orbit's
`hidden-quota-providers.json`, shaped as `{ "providers": ["provider-id"] }`.
Missing/unreadable/malformed files default to no hidden providers; mixed arrays
retain nonempty strings. Duplicate IDs are normalized so one restore always
works. IDs absent from the current report/catalog remain stored, matching the
source's stable-ID intent rather than treating a temporary missing report as a
request to show that provider again.

The header usage summary is computed from the original reports. Settings usage
cards also remain unfiltered and do not acquire these popover-only toggles.
No provider is disconnected, disabled, queried, or modified by hiding it. The
implementation consumes existing quota fields only and invents no quota data.

Writes happen off the rendering thread. A serialized, revision-guarded worker
prevents an older toggle snapshot from overwriting newer intent. Current UI
state changes immediately; write failure leaves that state in memory and reports
an error, without pretending persistence succeeded. Persistence remains
best-effort, not a durability guarantee against immediate process termination.

## Evidence and limits

The pre-correction mounted-popover regression could not find a hide action.
Headless native-widget interaction tests exercise hide, collapsed recovery,
show and reconstruction from the saved preference. Additional race tests cover
all-hidden partition, unchanged headline/settings reports, 201 rapid toggles,
an unavailable provider ID, malformed/mixed preference files and deterministic
write failure. All data and storage paths are synthetic temporary fixtures.

This does **not** add the missing MCP tab. Rust's two-pane popover and MCP runtime
projection (`session.rs:1775` onward and `mcp_ui.rs::mcp_popup_tab`) remain a
separate functional gap; adding a placeholder tab or inferred server state would
not satisfy it. Native popup placement, visual parity and hosted OS interaction
still require final-source native acceptance, not these headless tests alone.
