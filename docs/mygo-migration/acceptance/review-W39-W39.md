# Review and turn checkpoints: W39–W39

Pinned baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. Every procedure uses disposable cloud fixtures. A narrow test PASS never passes the whole row.

## W39 — Review loading starts immediately, stale invalidation reloads when visible, generation rejects older source requests, manual refresh feedback minimum 650 ms, source-specific empty/error states.

Functional **PARTIAL** · Visual **UNTESTED** · Native OS **UNTESTED**

**Rust source:** [crates/orbit-pi/src/sidepane.rs::load_review](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L582); [crates/orbit-pi/src/sidepane.rs::mark_review_stale](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L515); [crates/orbit-pi/src/sidepane.rs::set_source](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L786); [crates/orbit-pi/src/sidepane.rs::clear_diff](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/sidepane.rs#L800)

**Fixture:** Temporary Git repository with differing HEAD/index/worktree/branch/turn snapshots, Unicode/quoted paths and complete/compact/binary patches.

1. **W39.behavior [UNTESTED; ROW_SPECIFIC]** Open a Review source without cached data; manually refresh a loaded source; then change to another source while the first load is delayed.
   - Expected: Loading begins immediately; a source change resets unrelated snapshot. Completion for the obsolete generation cannot replace the selected source.
2. **W39.edge [UNTESTED; ROW_SPECIFIC]** Keep same-source snapshot on refresh error, show first-load error honestly and maintain minimum source refresh feedback timing.
   - Expected: On an error after an existing snapshot, code keeps previous snapshot and only sets `review_error` if no snapshot exists. Selection/scroll and collapses are refreshed per mode. This cached-error behavior is source truth, not proof fresh content is shown.
3. **W39.error [UNTESTED; ROW_SPECIFIC]** Fail first Review collection; then load successfully and fail refreshing that same source.
   - Expected: First load shows unavailable/error rather than an empty-success diff. Same-source refresh preserves the previous snapshot with a visible error; failure is not cleared merely by navigating back.
4. **W39.loading [UNTESTED; ROW_SPECIFIC]** Complete a manually requested refresh in less than50ms using a controlled collector and clock; inspect refresh feedback at100ms,649ms and650ms.
   - Expected: Source refresh_from_button feedback remains active for at least650ms even when Git finishes immediately. A slow request stays loading until its actual completion; the time floor is not a fake completion delay for content.
5. **W39.persistence [UNTESTED; DRAFT_NEEDS_SOURCE_EXPANSION]** Run this row in two cloned fixture homes, baseline then Go. Record touched files and active state; relaunch both; repeat the action and compare restored/reset state.
   - Expected: Restore/reset behavior and file contents must match the pinned handler. In-memory-only state stays in-memory; persistence is not inferred from an available store helper. Apply the specific source state contract above.
6. **W39.keyboard [UNTESTED; ROW_SPECIFIC]** Focus Review and press Escape during load, then reopen and invoke the registered Refresh action.
   - Expected: Escape follows Review cancellation/focus behavior rather than sending agent abort. Refresh reloads the current source and keeps its source/ref/path selection.
7. **W39.repeat [UNTESTED; ROW_SPECIFIC]** Queue sourceA, sourceB and sourceC refreshes; release callbacks in C,A,B order while also switching workspace.
   - Expected: Only current generation/workspace/session/source/ref/path can apply. If a pending request was deferred while loading, it runs once after stale completion rather than being lost.

**Go mapping:** `internal/workspace/checkpoint.go`, `internal/workspace/review.go`, `internal/ui/review.go`, `internal/agent/manager.go`
**Existing test candidates, not automatic acceptance:** `internal/workspace/checkpoint_test.go`, `internal/workspace/review_test.go`

**Executed narrow assertions:**
- **PASS** `TestReviewRefreshKeepsSource650msFeedback`: Native refresh spinner remains active650ms independently of completed content load; repeated clicks extend deadline. [Evidence](ui-feedback-evidence.json). Tested dirty-tree manifest `feb5fa4354edc6faf70b383f3c0ca59e0d1155d3b47c27052c040efbd7fd53db`, not a published commit.

**Owner challenge UI-W01 [PARTIAL]:** Initial acceptance failed at25.821182ms. Corrected with independent deadline and actual rendering predicate; default native regression now passes. Full Review animation/pixels remain open. [Scoped steps/tests](../native-ui-acceptance.json).

**Additional source-specific procedures:** [strict-results.json](strict-results.json), [ui-feedback-evidence.json](ui-feedback-evidence.json)

**Remaining:** Six-source immutable snapshots/checkpoint attribution/diff model tested; native review interaction/presentation parity remains pending. Dedicated native-controller stale source/session/workspace/hidden-toggle regressions pass under -race.
**Visual/OS gates:** No matched pinned-Rust/Go image baseline run at equal OS, scale, font, viewport and state. This row has not passed its complete native host interaction sequence. Linux CPU/domain tests and cross-builds do not establish macOS/Windows runtime parity.

