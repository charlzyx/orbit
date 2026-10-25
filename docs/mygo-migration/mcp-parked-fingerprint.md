# MCP parked process configuration invalidation

Behavioral reference: Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`,
`crates/orbit-pi/src/app/session.rs:792–805`,
`crates/orbit-pi/src/mcp/config.rs:777–795`, and `mcp.rs:61–66`.
Implementation base: Go `5e4e1425` (unreleased continuation after `.4`).

## Contract and implementation

Rust retains a parked process if it is busy **or** its loaded MCP fingerprint
matches the target workspace. A stale idle process is discarded on reopen;
settling alone is not an instruction to restart anything. The stamp includes
the global `~/.pi/agent/mcp.json` and target `<workspace>/.pi/mcp.json` paths and
raw bytes, not the outgoing workspace or the settings page's merged server list.
Absence/read failure is a distinct marker; formatting and malformed bytes still
participate. Go uses SHA-256 for the same equality domain, not Rust's hash value.

Go now captures this stamp in its startup worker immediately before launching
Pi. Idle parked reopen performs the desired-stamp read in a manager worker.
The target becomes selected immediately but is temporarily not ready for RPC
submission while this check is outstanding; `Open`, `Snapshot` and `Tick` do
not perform these filesystem reads. Busy and already bootstrapping processes
are reused without stale-config termination.

The completion must still own the same live entry, validation token and latest
selection epoch. A late completion may release its own readiness gate but never
steal a newer selection. A running/compacting/retrying/awaiting turn arriving
during the check wins the final busy recheck. Detaching the stale idle process
and assigning the replacement happen under the same manager lock, preventing
a prompt from slipping between the check and replacement. Command admission
is revalidated under the lock after caller-owned JSON marshaling; a command
captured before replacement/selection change is rejected with its draft kept. Child close and
bootstrap remain outside the lock. Unrelated parked children remain untouched.

Failure stays visible on the replacement runtime and does not create a retry
loop. Reopening can cold-start again. A successful replacement captures its own
loaded stamp, so later edits continue to invalidate it. A process reused for a
new conversation retains its actual loaded stamp, not an invented fresh one.

## Evidence and boundaries

`TestMCPIdleParkedReopenReplacesStaleConfig` failed before implementation with
"stale idle parked process was reused" using the actual Manager and synthetic
JSONL child processes. The regression suite additionally covers:

- unrelated outgoing-project edits versus target-project/global edits;
- raw whitespace changes, missing/empty/malformed files and deletion;
- busy stale preservation, settlement without eager replacement, idle reopen;
- blocked asynchronous reads, nonblocking selection, and RPC readiness gating;
- newer edits/selections superseding a held earlier validation;
- a running event arriving after idle admission but before validation completes;
- failed replacement (exited child, live RPC failure, canceled session switch),
  explicit reopen recovery, unrelated parked PID retention;
- a command paused in JSON marshaling while its captured runtime is replaced.

Fixtures use disposable homes/workspaces and executable helper processes, never
user MCP endpoints, credentials or conversations. These tests establish manager
lifecycle behavior, not real MCP server connectivity or native macOS acceptance.
The existing runtime lifecycle test now waits for asynchronous reopen readiness
before requesting idle `new_session` reuse.

Like the reference, the fingerprint is an observed disk snapshot, not an atomic
transaction with Pi's configuration read. An arbitrary external edit after that
snapshot can require a later explicit apply or idle reopen. This work does not
claim complete provider/configuration/native/visual parity.
