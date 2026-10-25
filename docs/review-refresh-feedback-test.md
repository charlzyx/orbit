# Review feedback regression test isolation

Pinned Rust299a32b sidepane.rs749–769 assigns a650ms minimum feedback deadline and clears it on expiry; a second click extends it. Production Go matches that lifetime and is unchanged here.

Original Wave10 full-race failure `TestReviewRefreshKeepsSource650msFeedback` at source_gap_test.go50,2.15s, remains recorded in /workspace/shared/orbit-review-refresh-audit-2206/original-failure.jsonl. Its assertion captured a deadline, drained arbitrary real async callbacks, then queried the mutable state at old deadline-1ns. Expiry can legitimately clear that state. Original output did not record the exact clear time, so this is a source-supported timing hazard, not proof of the original scheduler sequence. Focused unchanged Wave9 and Wave10 each passed3/3; that does not erase the full-gate failure.

Test-only revision separates three contracts:
- Immutable deadline boundary checks retain649ms,deadline-1ns,exact-deadline assertions and add after-deadline/loading/cleared-state controls.
- Actual mounted refresh button receives two native clicks. Both deadlines must fall within before/after click+650ms; second must extend first; glyph must be present. Controlled loading prevents unrelated Git completion timing from changing the fixture. No pending async expiry is drained before historical snapshot assertions.
- Complete App with a disposable real Git repository still receives two native refresh clicks, drains actual reload completions and asserts generation advances without error. This preserves real completion coverage without imposing a false sub650ms Git execution requirement.

No production clock/timer,650ms constant,timeout or waiting behavior changed. The old test's5ms sleep is removed; no sleep was added. Existing mounted single-glyph/rotation/reduced-motion test remains unchanged. No native capture or full-suite pass is claimed by this scoped result.

Exact-target `-race -count=3 -json` validation: each of the3 named tests passed3 times (9 actual test hits),0 failures,0 skips; package6.868s. Exact hitset/hash recorded in /workspace/shared/orbit-refresh-test-2209/focused-proof.json, raw focused.jsonl. Final aggregate rerun is the integrator's responsibility.
