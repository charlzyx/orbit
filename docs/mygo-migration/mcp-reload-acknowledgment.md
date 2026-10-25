# MCP runtime reload acknowledgment: bounded functional correction

Source comparison: Rust baseline `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`
versus Go `.4` commit `0af37538a766d57243f9170f28c726327d053127`.
This correction is not a claim of complete MCP, provider, native macOS or visual
acceptance. All regression configurations and JSONL children are synthetic.

## Source contract and observed failure

Rust `crates/orbit-pi/src/app/mcp_ui.rs::mcp_apply_now` (lines 210–273) only
restarts the active live, settled client with a known session ID and path. A
busy active turn/compaction defers it; an unrelated parked run is not stopped.
The new client switches back to the same saved conversation. `mcp_after_settle`
schedules deferred work outside the event-drain loop.

The Go `.4` host called asynchronous `Agent.Restart()` and reported only a
request. No production path called `mcpui.State.Applied()`, leaving actual
completed reloads unacknowledged. The real Controller → Manager → synthetic
JSONL child regression `TestMCPHostAcknowledgesOnlyCompletedSessionPreservingReload`
failed on that source with `host never acknowledged actual reload`.

## Correction

- Reload requests carry the active runtime, workspace, saved session path/ID and
  a host request epoch. MCP requests additionally carry the settings revision,
  observed fingerprint and selected Pi binary; generic resource/runtime reloads
  do not acknowledge MCP state.
- `RestartIdle` checks target identity and busy state under the same manager lock
  that detaches the old active entry. A newer session or newly started turn cannot
  be stopped by a stale request. A busy parked runtime does not block active apply.
- A nil restart return is admission only. The host waits for that exact replacement
  runtime to finish bootstrap and be connected with the saved session identity.
  New edits, session/workspace changes, or Pi binary changes invalidate an old ACK.
- Flight observation survives newer edits even though its old ACK is invalid.
  A failing old bootstrap retains explicit retry eligibility for the newer edit,
  without automatically restarting repeatedly. Failure checks use the active
  runtime-local error, never an unrelated saved-session operation diagnostic.
- Bootstrap failure leaves pending state and a retry message. It does not enter an
  automatic restart loop. An explicit retry may replace that known failed reload;
  an arbitrary initial disconnected/missing-identity process is not fabricated.
- Toolbar reconnect first refreshes configuration asynchronously, then schedules
  apply with the refreshed token. It does not bind a request to stale page data.
  A read-only probe is interrupted; an in-progress configuration write is not
  canceled and instead gives an explicit retry-after-completion message.
- Successful acknowledgment is conditional and requests a fresh probe. The token
  protects observed settings changes; it is not an atomic snapshot of arbitrary
  unobserved external file edits or proof of real MCP server connectivity.

The stronger completed-bootstrap UI acknowledgment is a deliberate honest
readiness boundary. Rust records its loaded stamp at successful spawn and sends
switch_session asynchronously; the implementation does not claim identical ACK
internals merely because the visible result is similar.

## Executable bounded coverage

- `TestMCPHostAcknowledgesOnlyCompletedSessionPreservingReload`
- `TestMCPReloadWaitsForBootstrapAndNewEditSupersedesOld`
- `TestMCPReloadFailureRemainsPendingAndExplicitRetrySucceeds`
- `TestMCPReloadNewEditThenBootstrapFailureAllowsExplicitRetry`
- `TestMCPReloadIgnoresUnrelatedSavedSessionDiagnostic`
- `TestMCPReloadDefersBusyActiveButNotBusyParked`
- `TestMCPReloadSessionSwitchCannotAcknowledgeOldCompletion`
- `TestGenericRuntimeReloadDoesNotAcknowledgeMCPPending`
- `TestConfigApplyWithoutLiveSavedIdentityDoesNotManufactureSession`
- `TestApplyAcknowledgmentRejectsNewEditWorkspaceAndBinary`
- `TestRefreshAndApplyCapturesFreshConfigBeforeHostRequest`
- `TestRefreshAndApplyInterruptsProbeAndReportsOtherBusy`
- `TestConfigReloadAtomicGuardsKeepBusyAndSupersededRuntime`

The child-process helper is a fixture, not an additional product acceptance test.
Existing saved-session catalog reload and runtime park/restart regressions are
also exercised. Aggregate tests and native acceptance must be rerun on the final
combined source, not inferred from this focused suite.

## Known remaining gap: lazy parked configuration invalidation

Rust `crates/orbit-pi/src/app/session.rs:792–805` fingerprints the **target**
workspace and reuses a parked client only if it is busy or its loaded MCP stamp
matches. A stale idle client cold-starts on reopen; a busy one is never killed.
Go `Manager.open` still reuses a matching alive/loading client without a loaded
configuration stamp. This patch does not fix that separate lifecycle gap.

A safe follow-on needs asynchronous target-workspace fingerprint validation,
loaded-stamp tracking, and race-safe lazy idle replacement. Do not insert disk
reads on the UI open path or eagerly restart parked/busy processes. Required
cases include unchanged cross-workspace reuse, project-local versus global
edits, stale idle reopen, busy stale preservation and a newer session switch
while validation is in flight. No full MCP parity claim is warranted meanwhile.
