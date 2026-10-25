# Review long-line navigation — scoped corrective implementation

Base: `5e4e1425`, Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. This is unreleased migration work, not a change to shipped `.4`.

## Source contract and implementation

- `sidepane.rs:1413–1483`: wrap-off uses a horizontal viewport around the vertically virtualized list; both split halves have sufficient natural width; in-list file headers are counter-shifted to remain at viewport edges.
- `review.rs:191–204`: hidden context contributes to width, so revealing context does not make the horizontal extent jump.
- Go `review_scroll.go`: same-axis separation, real `ui.ScrollHorizontal`/`ScrollState`, the existing `ui.ListState` retained; width cached per snapshot/generation/font/size and measured using native font layout, including hidden context. Measurement uses the actual renderer rather than Rust's conservative cell-count estimate, and chrome reflects the existing Go row gutters.
- `review.go`: no-wrap uses `NoWrap`, not truncated `MaxLines(1)`; file/section inner header counter-shift; unchanged diff spans and split pairing.
- Wrapping restores viewport-contained flow and clears horizontal offset; new workspace/source snapshot clears pan/cache. File navigation within a snapshot retains pan and selected file. No toolkit/module-cache changes or compatibility guard bypass.

## Verification

Mounted red before implementation: `TestMountedReviewUnwrappedLongLineHasHorizontalViewport` failed because no horizontal viewport existed. Log: `/workspace/shared/orbit-toolchains-continued/review-red.log`.

After implementation, real `ui.Tester` pointer-scroll tests cover unified/split long tails, horizontal movement without changing vertical anchor, continued vertical virtualization/wheel routing, wrap restoration, snapshot reset, fixed file-header position/width, file navigation, hidden Unicode context extent, and code-font size change. Canonical wrapper, exact pin and additive compatibility overlays retained.

`source /workspace/shared/orbit-toolchains-continued/env.sh; scripts/go-native.sh test -race ./internal/ui -run Review -count=1`: PASS (89.818s), log `/workspace/shared/orbit-toolchains-continued/review-all-race.log` (before subsequent context-direction follow-on).

## Explicit limits

This is mounted native-toolkit state/geometry/input evidence, not same-state desktop PNG parity or Mac verification. Existing Go file-header height/style differs from Rust and is not claimed fixed. Sticky vertical headers and directional context controls are follow-on work. Existing review text did not expose text selection/copy; this change does not add or claim that capability. Composer/editor selection paths are unchanged. Full repository aggregate and frozen native recapture remain integration gates.
