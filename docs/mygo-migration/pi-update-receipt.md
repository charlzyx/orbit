# Automatic Pi update receipts preserve existing processes

This isolated delta is based on a read-only frozen wave7 source snapshot plus
reviewed cleanup and notification mappings (local baseline `1ed4179`). It does
not change the automatic update worker, installer, timing, busy/admission checks,
or the user's explicit RestartRuntime / RuntimeChanged actions.

## Reproduced behavior difference

Pinned Rust `app/pi_update_ui.rs:138–172` explicitly keeps existing Pi processes;
only subsequently created processes use the installation on disk. Go's live
startup `updated` callback instead dispatched RuntimeChanged, whose idle branch
restarted the active process. Extracting that exact switch into the actual owner
handler and injecting a receipt reproduced replacement of live idle task
`live-1`/PID 174 with loading `live-2`/PID 0. This is a real harmless JSONL
subprocess fixture, not an assertion only against an action mock.

The corrected automatic branch never dispatches a runtime/task action. It emits
the source success receipt with verified version/path and an independent error
warning when the verified installation is not on PATH. Existing failure detail
handling stays unchanged; unchanged-version completion updates facts silently.

## Ownership and truthful facts

The startup callback captures manager identity, configured binary and setup probe
generation on the owner. After Close, manager replacement, binary selection, a
newer dependency refresh, or a pending explicit migration, stale events cannot
change newer UI facts. Active task identity is deliberately not an ownership
condition: this is a global installation receipt and must not mutate any active
or parked idle/busy process.

A matching terminal dependency snapshot is copied. It supersedes/cancels an older
in-flight probe before replacing only Pi host facts, retaining other host facts.
Pi installation verification does not imply required Node readiness. No task is
resumed or created, and absent dependency snapshots do not invent readiness.
Newer explicit probe generations win over earlier startup receipts; the worker
orchestration itself is intentionally unchanged by this delta.

## Fixture coverage and limits

* The real idle child PID/opaque key and local draft/attachment remain unchanged.
* A busy parked child and a newer active child both survive the global receipt.
* A harmless temporary shell launcher is rewritten by the test itself. The
  injected receipt launches nothing; a subsequent explicitly created process
  uses the new fixture launcher, while the old child stays alive. Explicit
  RestartRuntime and RuntimeChanged continue to replace their requested child.
* Stale manager/binary/probe generation, queued old receipt, closed host and
  pending migration are rejected. Missing Node, copied dependencies, preserved
  non-Pi facts, probe cancellation, PATH warning and failure/unchanged handling
  are checked without any installer or update network request.

The synthetic process test is not native installer acceptance and does not
resolve the previously documented setup-entry contradiction. Actual automatic
update target/environment selection remains an independent orchestration audit;
this patch changes only owner receipt application.
