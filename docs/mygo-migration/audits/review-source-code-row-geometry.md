# Review code-row geometry: bounded source correction

## Evidence and corrected causal claim

Go `987eacb` WB05 and Rust `299a32bc` full native images show extra wrapping
of added return strings and `Untracked fixture note.` in Go. Those pixels prove
a difference, **not that Go's gutter is wider**. The earlier audit's causal
wording was corrected in `/workspace/shared/orbit-native-987-1846/workbench-actual-review.md`.

Rust `diff_view.rs:20,253–323` declares 12.5px code font (scaled by editorSize/13),
fixed requested 16px line height, minimum row19, a fixed52px number gutter,
12px density-scaled left body padding and10px right padding. Add/delete unified
rows have a2px left edge. There is no +/- text column. Split uses old numbers on
left, new on right,12px body padding both ends and a1px middle divider
(386–464); `sidepane.rs:1681–1750` routes both views through these functions.

The real mounted254px-row RED (`review-row-red-1911.log`) found Go's original
RichText x64/W184/H18. Source column geometry instead gives context x64/W180,
addition/deletion x66/W178. Thus source's body can be **narrower**, not wider.
The patch changes the actual columns and declared font/line-height contract;
it does not shrink gutters opportunistically to force the fixture onto one line.

## Explicit renderer limit, not pixel acceptance

At this environment's resolved `.ZedMono`, MyGo's measured note width is184px
at both12.5 and13px. Its text engine also floors line height at glyph ascent plus
descent (`internal/text/layout.go`, `line.Height=max(requested,natural)`): the
requested16px yields17px per line here, so the note actually occupies34px after
wrapping. This remains different from the Rust capture. No toolkit/font module
was changed, no text clipped to fake16px, and no pixel-parity claim is made.
The mounted test compares row coordinates to the source constants and actual
height to a separate native source-font probe; that probe is not Rust pixel proof.

## Bounded production changes

- New `reviewCodeRow` for context/addition/deletion only, source single gutter,
  side-specific split numbers, number color/edge/washes and body padding.
- Actual split halves share available width around1px divider, retain filler
  and stretch their code row to the partner row's height.
- Unwrapped cache measures code at its new actual font, non-code rows at their
  unchanged font, and reserves actual chrome including change edge/divider.
  Hidden context, font/dataset invalidation, horizontal viewport, list owner/key,
  vertical scrollbar inset, sticky headers and context expansion are preserved.
- Existing tail-reach regression now checks source10px body end-padding rather
  than the old6px. It still performs real pointer pan and checks the endpoint.
- No file-header/menu/icon expansion, no hunk/meta redesign, no toolkit change.

## Selection scope is deliberately not expanded

A real drag across two dock rows followed by Cmd+C left the clipboard sentinel
unchanged (`review-selection-gap-red-1920.log`; diagnostic source archived beside
that log, not hidden inside the green gate). Baseline dock RichText is not marked
Selectable. Rust's inspected code-row path uses StyledText without an identified
text-selection handler; sidepane selection references primarily mean file/tree
selection. Therefore this is **not established as a Rust→Go migration omission**,
and cross-row copying remains unverified for Rust/unsupported by this Go path.
No Selectable container or new virtualization/selection ownership feature is
added. Existing composer/editor selection regression tests are retained.

## Verification scope

Real mounted RED→GREEN source columns and font declaration; split old123/new456
number test; complete MountedReview/NativeReview subset, including long-line
pointer pan, visible-edge vertical scrollbar drag, sticky clicks/wheels,
context expansion, resize, font changes, split toggling and owner resets.
The final gate also runs existing editor and transcript selection tests.
Logs are `/workspace/shared/orbit-toolchains-continued/review-row-*`.

The final combined-source aggregate and hosted native recapture remain parent
responsibilities. WB05 per-file header icon/statistics and exact glyph metrics/
wrapping remain explicit gaps; this change is not whole Review visual parity.
