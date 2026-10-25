# Abort pipeline CI failure: test synchronization correction

## Observed
- Reported CI aggregate: e577c31, runtime 32083a3, run 40201741107 / job121194174997; `TestAbortPendingRestorationCancelledByExplicitClear` failed at abort_pipeline_test.go:138, `abort/clear workers did not drain`.
- Isolated source copied from orbit-functional-integration-1419 to /tmp/orbit-abort-audit-1426. No runtime edits, credentials, provider use or shared-source mutation.
- Original focused command: `go test -p 1 ./internal/agent -run '^TestAbortPendingRestorationCancelledByExplicitClear$' -count=30 -timeout=180s` reproduced identical failure (5.02s case, 5.718s total). Terminal output was captured by tool transcript; no original log file was created.

## Causal source evidence
- manager.go:984–985 returns `Snapshot.Running = snapshot.Running || entry.state.awaiting`.
- Submit/Send marks awaiting before asynchronously writing prompt (manager.go:513 and subsequent launch).
- The three tests waited only for `s.Running`, so they could proceed without a fixture `agent_start`.
- Abort clears awaiting synchronously (manager.go:1328), then asynchronously queues clear/abort. Therefore the next `!s.Running` could pass before either command arrived.
- Fixture manager_test.go:63–74 releases only currently held clear/abort responses; it does not remember an early release. A premature release RPC can complete, then the first clear response becomes held until the 15s command timeout, exceeding the test's 5s worker-drain guard.
- Existing runtime CallOrdered still queues clear before abort without awaiting clear; no evidence here requires a runtime modification.

## Correction and validation
- Test-only patch: require `s.Running && !s.Awaiting` before Abort in each of the three delayed-clear tests. This waits for real fixture `agent_start`; subsequent !Running then means its settled event, not local awaiting cancellation.
- Preserve all timeout bounds and restoration/leak assertions.
- `git apply --check` against original integration source passed.
- Focused command after correction: `go test -p 1 ./internal/agent -run '^TestAbort(PendingRestorationCancelledByExplicitClear|DelayedClearRejectsReusedStateButRestoresParkedSession|PendingPipelineStopDoesNotRestoreExpectedCloseError)$' -count=100 -timeout=180s`
- PASS: `ok github.com/charlzyx/orbit/internal/agent 12.615s`.
- The first test has reused/parked subtests, so all four concrete scenario paths were repeated 100 times.
- Cached canonical toolchain/env used, serial package compilation, no new cache. Output: orbit-abort-sync-1429.log.
- Patch SHA256: 52ce101bb6b624e19356fdc0eb6e6d36074ffc0217a1d8735730402c11e8a7a5.

## Scope limit
This establishes a reproducible test barrier flaw and focused regression stability, not full aggregate/native acceptance or proof of all abort semantics. Immediate Abort while prompt is awaiting is a different scenario and should not be conflated with the held-response fixture. Shared-source application and aggregate rerun remain integration owner's responsibility.
