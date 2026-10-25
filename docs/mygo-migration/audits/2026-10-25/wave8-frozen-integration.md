# Wave 8: source behavior and bounded verification

This feature-branch source batch is based on `0173c5e6`. The `.4` installer is unchanged. App-managed installation/update/migration and real user credentials were not exercised; temporary pinned CI/toolchain dependencies are distinct from those user operations.

## Included source scope

- Remove the proven unreachable legacy runtime UI closure, retaining live runtime facts, maintenance, updater and privacy behavior. The old Pi-binary action is not a reachable native chooser and is not evidence of the user's picker hang.
- Explicit Refresh follows the visible page: Review/Usage use existing asynchronous ownership/coalescing; MCP follows the source read-and-probe contract; Plugins only rediscovers local resources, without silently adding an update check. Tests use recording hooks/disposable state, not real configured endpoints.
- Pi notices use known source messages/severity. Verified automatic-update receipts no longer restart the existing agent process; explicit Restart remains. Automatic checker admission verifies target HOME/PiDir and rechecks after maintenance admission. Full background worker settlement remains next-wave work.
- Ragged Markdown tables preserve extra body cells/links through a bounded public-renderer projection. All output-row widths count toward cell/byte budgets; over-budget content falls back to complete selectable raw text, not silent truncation. Owner/content revisions reject stale clicks. Ordinary tables remain unchanged; this is not a claim of complete Markdown/GFM geometry or image-inline parity.
- Static Chat launch receipts verify an independent pinned vendor archive plus source-patched file inventory, exact Node/launcher identities and owned live process ancestry. The new **metadata-only required CI job** must run two real tests with zero skips before either native job. Chat opt-in stays off on normal pushes; no fresh opt-in run is authorized by local results alone.
- The existing pinned Rust reference job runs a verbatim parser fixture after checking the original source hash and extracted segments. No new Rust dependency, network operation or GUI action is introduced by that step.

## Final combined gate and honest limits

595 Go/module hashes were frozen and rechecked unchanged. Whole-repository vet, serial race tests and Darwin arm64 cross-build passed, final chain exit 0. UI race: **516.840 s**. Full package timings and incoming patch hashes are in [wave8-integration-evidence.json](wave8-integration-evidence.json).

The canonical test command now uses `-p 1 -timeout 20m` for the growing **whole-package suite budget**. No case, assertion, application TTL or individual operation timeout was removed or increased. The earlier standalone notice suite timed out at 600.055 s as a new case started; that case plus notices passed isolated, and this final combined suite passed. Original failure logs remain retained. Serial execution here finished UI below the old limit; the larger total budget is not a claim that this run required over ten minutes.

Local Python: 50 tests, 47 pass and 3 explicit opt-in/platform skips. An additional actual Node22/Pi metadata run passed for both renderers (38.553 s), but local `/proc/task/children` is unavailable: complete ancestry remains **NOT VERIFIED locally**. The required hosted receipt job and actual Rust parser execution are **NOT RUN before publication**, not inferred passes.

## Deliberately excluded

The incomplete `90ebb` Rename fix is not shipped: real offline Pi proved it rejected legitimate unpersisted live-task rename. Its fully target-bound replacement is separately reviewed for the next batch. Background updater settlement, typed deferral, pending-image submit protection and palette dynamic titles are also separate. The fractional-glyph toolkit prototype is not enabled: its known allocation assertion fails identically on pristine and candidate and is not reported as an all-green suite.
