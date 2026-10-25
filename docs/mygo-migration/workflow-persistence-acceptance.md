# Workflow selection before Pi session identity

Functional **PARTIAL**; visual and native OS **UNTESTED**. These eight narrowly
executed assertions do not constitute a complete workflow-feature pass.

## Pinned original contract

Reference `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`:
`crates/orbit-pi/src/app/composer_ops.rs::choose_workflow_mode` (891–902),
`app/runtime.rs::reset_quota_entries` (170–183), and
`app/session.rs::begin_new_task` (638–679). A choice made before a new session
has an ID belongs to that pending task, not the previously parked session.
New Task returns to Build; a resumed task loads its own persisted choice.

## Executed procedures

| Procedure and expected result | Go regression |
|---|---|
| Choose Plan before first authoritative ID; resolve ID; persist only that task | `TestWorkflowSelectionBeforeFirstIdentityPersistsOnlyNewTask` |
| Request New Session with delayed RPC; old ID becomes unavailable before waiting; later choice belongs to new ID | `TestWorkflowNewSessionClearsOldIdentityBeforeAwaitingRPC` |
| Fail child spawn; retain pending choice; explicit New Task resets to Build | `TestWorkflowFailedSpawnRetainsPendingAndNewTaskResetsIt` |
| Cancel New Session; restore previous choice without corrupting persisted mode | `TestWorkflowCancelledNewSessionDoesNotMutatePreviousChoice` |
| Delay persisted-mode hydration and make a newer choice; late read cannot overwrite it; report corrupt-store failure | `TestWorkflowStaleHydrationCannotOverwriteNewChoice` |
| Resume a saved session; use its own mode, not another draft's pending choice | `TestWorkflowSavedSessionLoadsItsOwnModeNotPendingOtherTask` |
| Capture a task target, switch to a different draft, then execute delayed worker; reject stale target | `TestWorkflowCapturedTargetCannotWriteLaterDraft` |
| Retry child startup after failure; carry pending selection across the retry | `TestWorkflowRetryPreservesChoiceAfterFailedSpawn` |

Implementation: `internal/agent/workflow.go`, lifecycle integration in
`internal/agent/manager.go`, and native click-time opaque target capture in
`internal/ui/app.go`. Disk I/O runs outside the UI/state mutex, serialized across
sessions to prevent read/modify/write lost updates. Version and generation checks
reject stale asynchronous results. This also hardens stale-click races without
changing the intended original workflow choice semantics.

Verification: the canonical patched command
`scripts/go-native.sh test -race ./internal/agent -count=1` passed on
2026-10-25 at the dirty-tree source snapshot copied as checkpoint-0552, with
`ORBIT_PI_INTEGRATION_BIN` pointing to the official pinned Pi1.4.0. Full agent
suite duration32.551s; provider inference was not requested and no credentials
were used. Exact file hashes are recorded alongside the immutable source snapshot;
the subsequently published commit identifies those bytes. This is not an earlier
published runtime's test result. Visual workflow picker matching, extension-driven
workflow enforcement and real provider execution remain separate acceptance work.
