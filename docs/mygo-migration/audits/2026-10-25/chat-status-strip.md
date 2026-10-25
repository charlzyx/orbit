# Retry and compaction status strips — unreleased increment

Pinned Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `crates/orbit-pi/src/app/events.rs:193–220`, `app/view.rs:3882–3980` and placement at `view.rs:962`.

The Go snapshot now carries immutable retry attempt/max/error detail. Chat mounts the retry-priority strip with a Cancel button, or the compaction strip without an invented Cancel action. Missing/null/malformed provider errors use localized `events.transient_error`; a supplied literal is retained. Observed end, settlement, disconnect and Stop clear detail.

Tests cover reducer event defaults and immutable snapshots; mounted controls; and actual JSONL subprocess → Manager → Controller.View → native mounted Cancel → `abort_retry` protocol, including rejection, newer attempts, busy parked sessions, compaction Stop and Chinese fallback text. The unchanged renderer failed a mounted compaction assertion; the new renderer passed. Final localized protocol fixtures passed five repetitions. These are credential-free fixtures, not real provider inference or complete hosted-desktop/pixel acceptance.

## Explicit remaining source differences

- Rust `app/settings.rs:5828–5835` clears retry presentation optimistically on Cancel. This increment still keeps Go's observed-state presentation until an event arrives. The existing test describes that current behavior; it is not proof of Rust parity. A separate source-equivalent cancellation increment is required.
- Retry exhaustion fallback/prefix remain English in the reducer.
- Context usage ring/details/quick Compact are a separate pending increment.

C14 is not declared fully closed. None of these changes modify the shipped `.4` installation.
