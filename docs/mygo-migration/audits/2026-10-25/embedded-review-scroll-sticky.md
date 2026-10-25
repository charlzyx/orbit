# Embedded Git Changes: long lines and sticky headers

Scope: the expandable staged/unstaged diffs inside Git Changes, not the docked
Review pane. This batch remains independent of dock Review changes.

## Source contract

Pinned Rust baseline `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`:
`sidepane.rs:1413–1483` keeps a vertical virtual list inside wrap-off horizontal
scroll, with headers counter-shifted; `review.rs:191–204` includes hidden context
in its width; `sidepane.rs:1485–1574` stacks section and file headers and pushes
them at the next section/file boundary.

## Implementation

- Native horizontal viewport around the original virtual list; the existing
  selectable RichText switches SingleLine to NoWrap, without changing copy data.
- Native measured width includes both staged/unstaged snapshots and retained
  hidden context; split doubles the existing row extent. Cache invalidates on
  snapshot replacement/count/generation or native/theme font metrics.
- Headers counter-pan; vertical scrollbar painting and pointer hit-testing use
  public ScrollbarInsets to remain at the visible right edge at every pan.
- Real absolute List children reuse existing section/file actions. Section sits
  above file; next file pushes file only; next section pushes the section and
  retires the file. Native wheel events bubble through the actual List.
- Header identities include staged/path; header handles/heights are pruned on
  dataset/filter replacement. Workspace/tab transitions reset viewport state.
- No internal geometry labels override the existing readable RichText content.

## Mounted regression evidence

Synthetic isolated fixtures, no live data: initial viewport test failed on the
old renderer, then native pointer pan reached the tail and Ctrl+A/C copied it.
Tests cover unified/split; virtualization; vertical scrollbar drag at left/middle/
right pan; wrap reset; hidden context; same-count replacement; font metric changes;
split toggles; resize near the horizontal end; scaled header geometry; next-section
and next-file push; section/file actions; and sticky Discard retaining its existing
confirmation (Cancel performs no mutation). A separate real header-wheel test
failed with a sibling overlay and passed after moving it into List.Children.

Logs under `/workspace/shared/orbit-toolchains-continued/`:
`embedded-red.log`, `embedded-scroll.log`, `embedded-sticky-red.log`,
`embedded-sticky2.log`, `embedded-review-fixes.log`,
`embedded-header-wheel.log`, `embedded-wheel-fixed.log`.
Final package race/vet outcome is recorded in the handoff, not inferred here.

## Limits / remaining work

These mounted tests are native toolkit layout/input evidence, not an actual macOS
screenshot comparison or full migration acceptance. The embedded renderer still
uses its prior explicit monospace 12px font; editor-font preference parity is not
claimed. Existing delayed context-menu callbacks can combine a captured old path
with a new workspace; this separate lifecycle issue is not fixed by this batch.
Directional context expansion and unrelated Git/GitHub workflows are unchanged.

Manual native checks: expand a long staged and unstaged diff; wrap off and pan to
the tail; select/copy a visible line; drag the right-edge vertical scrollbar at all
horizontal positions; wheel over both sticky headers; resize and toggle split;
scroll across file and staged/unstaged boundaries; cancel Discard; switch workspace
and confirm no inherited pan/header. Confirm against Rust, not merely this checklist.
