# Session file ownership / metadata lease integration (frozen 17:32 UTC)

Incremental patch: `orbit-session-lease-incremental-1732.patch`, based on isolated local commit `e55b13cad7ca02d7d9d5d4aa01c89edcc78b2c30` (complete prior 1717 ownership/session-menu snapshot). Apply AFTER the 1717 agent ownership patch. This delta already incorporates agent-only export a71d2c and protocol settlement f782b4e: do not apply those agent hunks a second time. UI export additions remain separately owned by the export patch. No release or merge performed.

## Bounded behavior
- Worker-side canonical path + SameFile ownership from 1717 remains intact. Active/loading/stopping/alive entries and retiring paths remain protected; stopped background records can still be cleaned up.
- Reusable metadata target selection and lease reservation happen in the SAME manager lock after the full worker-side ownership snapshot is revalidated. No second writer is opened for a live directory alias/hardlink owner.
- Saved-session SetSessionName and GetState retain the captured original child lease through actual completion. A timeout only reports delay; it cannot prove cancellation. Mismatched response command retains the lease until child/manager termination.
- Generic SetSessionName shares the post-marshal target/generation admission and lease, but does not increment task generation. NewSession during a lease takes a fresh runtime rather than redirecting the held write. Other identity transitions cannot reuse that child.
- Metadata errors from a retired/stopped target do not become the replacement task's global connection error.
- Restart closes the old child off the UI thread, then waits for captured lease.done. Metadata releases its path+FileInfo reservation before signaling done. Replacement remains restartClosing and retains retired-path ownership until cleanup completes. Shutdown interrupts this wait; it does not launch a replacement child after Close.
- Warm existing-runtime Open is conservatively rejected while ANY deletion reservation exists. This can temporarily reject reuse of an unrelated session, but is explicit/retryable and restores normal reuse after release; it performs no UI filesystem I/O.
- Temporary metadata child Close runs before the outer file gate is released.

## Evidence and limits
Focused race includes held rename past timeout, same-child stale marshaling, warm alias reservation/recovery, protocol mismatch retention, held identity stop/restart/close. `session-lease-restart-green-1728.log`: PASS 31.295s. Deterministic Close while restart waits for owner cleanup and held Rename→Restart clean gate assertions passed 3x in `session-cleanup-close-1731.log`.
The negative `session-metadata-stale-error-red-1726.log` reproduces the genuine replacement-start/file-gate cleanup race (not a stale diagnostic-only failure); done handshake fixes that path.
Final frozen agent/pirpc race and vet PASS (agent 131.358s; pirpc cached): `session-lease-full-race-1732.log`, `session-lease-full-vet-1732.log` under `/workspace/shared/orbit-toolchains-continued`.

All filesystem tests use disposable fixtures. Case-insensitive variant test is skipped on this Linux filesystem; macOS case-insensitive validation remains pending. Internal reservations are not a filesystem transaction against arbitrary external writers or atomic symlink replacement races. No native UI acceptance or full Rust parity claim follows from these tests.

Fresh uncached pirpc race rerun also passed; see `session-lease-pirpc-fresh-1734.log`.
