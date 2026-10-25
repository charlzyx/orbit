# Wave 10 integration — first complete gate failed; correction gate pending

Based on the completed Wave 9 runtime; the public parent is `b81869a9fc30bee4e357ed32096206a65f421a50`. No installer release is changed. This batch is not full function/visual parity.

## Bounded changes

- Local main-Review code-line glyph rounding policy and matching width measurement/cache use the reviewed fixed-hash compatibility overlay. Defaults and other text regions remain unchanged; unsupported platforms/API retain original behavior. All original toolkit allocation known-failure evidence remains documented separately from passing app tests. Native raw before/after comparison is still pending; no live owned Pango option was measured.
- Five existing command-palette titles follow actual panel-open / exact-path cached pin state, with original locale keys and no paint-time filesystem I/O.
- Code-copy header retains the original body renderer, copies open/live code, and removes only the parser-added final newline. Owner/content keys reject stale held activations. Friendly source language captions preserve original tokenizer tags and unknown input. Folding is excluded.
- Attachment ingress preserves source-dependent drop/file-picker/image-picker fallback behavior and FIFO order; bounded queued admission and two I/O slots are explicit Go safety differences. Native source file selection is single-file; drop remains multi-file. The actual file reader uses limit+1 and regular-file validation, including a nonblocking Unix open against FIFO replacement. These are raw ingress limits, not total decoded-pixel or process-memory bounds. A separate preview DecodeConfig gate is excluded.

## Validation boundary

The final combined focused run executed 72 actual tests/subtests, all passing; the recorded expected complete names and emitted JSON run/pass sets were checked. The canonical wrapper composition/hash negative tests passed. The production/build/test source freeze contains 1,592 file hashes, including compatibility assets and wrapper scripts. The complete single serialized vet/race/arm64 gate is pending.

The first caption-only regex was incorrect and did not run the caption tests. Its claimed caption-locale pass was withdrawn. Exact tests on both the author's tree and integration tree reproduced the missing locale-test precondition. The corrected test uses the real Appearance language selector, checks the temporary Store's zh-CN persistence, returns to Chat and retains the Chinese-caption assertion. Seven exact test/subtest names each ran and passed five repetitions. Production caption code was unchanged; original failed and misselected logs are preserved.

The Wave 9 focus-test correction is inherited unchanged. New Composer mode menus/Access admission, focused Ctrl-Up navigation, preview pixel admission and code folding are excluded from this frozen batch. Original `.4`, the macOS picker-hang uncertainty and unverified native interaction states remain explicit.

## First complete gate and reviewed corrections (22:14 UTC)

The original freeze `bf20275a211dabf80ef85833f8c447e7c892b6bd` did **not** pass its complete gate. Vet passed, but the serialized race run failed in three tests: `TestMCPParkedUsesTargetWorkspaceAndGlobalRawBytes` (timeout), `TestMountedMarkdownImageServiceRoundtripInvalidatesHeldOwner` (fresh preview did not open), and `TestReviewRefreshKeepsSource650msFeedback` (extended feedback ended early). Agent took 159.637 seconds and UI 504.316 seconds; the following arm64 stage was not reached. Raw original JSON remains `wave10-race-2146.jsonl`; subsequent focused results do not erase this failure.

The unchanged MCP case passed three repetitions on each Wave 9/Wave 10 baseline and ten in an isolated diagnostic copy. Its original output omitted the actual Active fields, so no root cause is established. A test-only timeout diagnostic now prints these fields without changing the original five-second deadline.

The Review test previously queried mutable state at a saved historical deadline after draining asynchronous callbacks that may clear that state. This is a source-supported test hazard, not proof of the exact failed interleaving. The reviewed test-only revision separates immutable 650ms boundary checks, real two-click deadline extension/glyph checks, and real disposable-repository reload completion. Three exact targets each passed three repetitions, three adjacent regressions and vet passed. Production timing, business TTLs and waiting budgets are unchanged.

The fresh-image preview failure is still under independent exact/baseline diagnosis. No image production correction or new complete pass is claimed. A new source freeze and one full serialized gate will follow the reviewed disposition of all three failures.

### Reviewed image correction and freeze v2

The identical image test passed five times on Wave 9 and failed five times on Wave 10. Isolated instrumentation found stable bounds/cache owner but a changed inline owner and button identity: the new owner-keyed Markdown wrapper exposed an unconditional epoch bump during an idempotent same-workspace service refresh. The minimal reviewed correction advances service ownership only for an actual path/model mismatch; explicit admitted SetWorkspace intent still advances once even for the same path. New/OpenSession and opaque task generations remain unchanged. The original image test is byte-identical and now passes five repetitions; the full 41-name picker/dialog/attachment/image matrix passed three repetitions, with no skipped cases, and vet passed. See `workspace-refresh-owner-2218.md` for callsites and causal evidence.

Freeze v2 includes this production correction, the expanded MCP failure diagnostic and the Review test-only separation. The original three-failure gate remains recorded. The new complete gate is pending; focused results are not its replacement.

### Final v2 complete gate

The reviewed v2 source completed the entire serialized chain successfully: vet, all 25 race-tested packages (UI 488.659s, agent 155.127s), and Darwin arm64 cross-build. All 1,593 frozen source/build/test/asset hashes were rechecked unchanged. The original three failing tests now pass in the complete run; the MCP case remains an unreproduced original failure rather than a claimed root-cause fix. The JSON records 19 conditional skips separately in `wave10-v2-gate-evidence.json`; a package pass is not a claim that these skipped cases executed. The earlier focused six complete critical names each ran/passed three times with zero skips. Original v1 failure evidence is preserved. No new native image, geometry or keyboard acceptance is inferred from this local gate.
