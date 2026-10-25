# Docked Review context, sticky headers and scrollbar correction

Unreleased increment on `108f4fb`, Rust reference `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`. No release/main changes. Embedded Git Changes is a separate renderer and is **not** covered by these dock fixes.

## Source-derived behavior

- `sidepane.rs:1750–1854`: leading gap expands from end, trailing from start; between gaps <=100 lines have one both-edge action, larger ones have start/end actions; the label expands both, Shift on either target expands all. Unavailable retained context stays disabled. Added production native gutter/button wiring, preserving the existing domain cap and no filesystem reads during expansion.
- Gap virtual-row identities now include stable Gap.ID. Previously two gaps in one file had the same `(file,0,0,gap,empty)` key; partial expansion must preserve one gap's identity while distinguishing different gaps.
- `sidepane.rs:1485–1550`: an actual native absolute file-header child stays outside horizontal scrolling and is pushed out by the next file header. It reuses the real collapse/open-file handlers and uses a file-specific native key, so a press held across a scroll to another file cannot activate that other file on release.
- Source list anchors are unchanged. An initial MyGo List.Header implementation failed existing ScrollTo/split/hunk tests because that API reserves sticky-header height on ScrollTo(Start). It was rejected, not accepted by changing the old expected indexes. The native absolute-child implementation follows the Rust layer structure instead.
- Independent review found the first long-line patch placed the inner list's native vertical scrollbar beyond the outer viewport. A real right-edge Press/Move/Release test reproduced this at pan=0. Public ScrollbarInsets now shifts both painting and hit-testing to the visible right edge at left/middle/far-right pan; no toolkit patch or fake screenshot overlay.

## Evidence

Red logs in `/workspace/shared/orbit-toolchains-continued/`:
- `review-gap-red.log`: missing directional/Shift entry points.
- `review-sticky-red.log`: file header disappears after real vertical wheel.
- `review-bar-red.log`: right-edge vertical drag does not scroll at pan=0.

Mounted tests now exercise direction/Shift, small/large between gaps, repeated split expansion, disabled gaps, stable gap identity, dual-axis sticky header actions, next-header push, held press across files, visible-edge scrollbar dragging, and unchanged full-width/split/hunk logical anchors. `review-sticky-anchor.log` passed all focused old/new navigation regressions; `review-sticky-press.log` passed the final identity-specific gesture. `review-final3-race.log` passed all Review tests before the final file-specific identity addition (167.400s).

Final exact-tree canonical `test -race ./internal/ui -run Review -count=1`: PASS (193.161s), log `review-final4-race.log`. Canonical `vet ./internal/ui`: PASS, log `review-final-vet.log`. Native desktop same-state capture, full repository aggregate, and Mac behavior remain integration gates. These mounted input/geometry tests are not pixel parity or complete W35/W36 acceptance. Header styling/counts and grouped embedded section/file behavior remain separate differences.

## Sticky header input follow-up

A real mounted pointer-wheel test exposed that an absolute sibling header swallowed
wheel events instead of reaching the vertical diff list. The sticky child now belongs
to that same native List (public `Children`), counter-panned by the horizontal offset.
This preserves both native wheel ancestors without forwarding events or introducing
`List.Header`'s different logical scroll-anchor behavior.

Mounted coverage includes header-origin vertical and horizontal wheel, unified/split,
near-end horizontal pan followed by resize, vertical-anchor preservation, real collapse,
held press crossing file identity, next-file push-off, and visible-edge scrollbar drag.
Existing editor/transcript selection tests are run as adjacent regressions; dock Review
RichText itself was not selectable before this change, so readable accessibility names
are not evidence of new diff selection support. Native desktop recapture remains separate.

## Directional gutter fixed-height correction

Rust `sidepane.rs:1788–1794` explicitly fixes the first of two directional controls
to 16px. A mounted geometry test showed that equal flex growth instead divided
the 32px gutter into border-inclusive 16.5/15.5px controls. The first control now
has an explicit 16px height/min/max and no flex grow/shrink, preserving a 16/16
hit-target split. Width remains the source 46px gutter including its 1px border.
This is a bounded geometry correction, not a new functional or native screenshot
parity claim. The red evidence is `review-gap-height-red.log`.
