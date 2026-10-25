# Inline image payload revision — scoped correction

Status: implemented with mounted/race regressions; **not full C09 or hosted
pixel acceptance**. This increment follows the independent inline-owner epoch
fix and changes neither decoder support nor network policy/cache total budget.

## Source contract

Pinned Rust299a32b `transcript.rs:1081–1111` reconstructs messages on `load_from`.
`ChatMessage::from_value:331` parses image blocks via `image_from_block:497–519`
using their actual data/MIME. A reloaded full snapshot is not an ID/index-only
bitmap lookup. This does not assert equivalent partial-image stream-delta
semantics; the Rust assistant-step merging path has separate behavior.

## Observed regression and correction

Real mounted PNG red→green replacement at the same message ID/index kept the
old red pixels. Empty/invalid base64 and MIME-only corrections retained sticky
errors. A held old decoder completion could publish after slot replacement.

`inline_image_content.go` records immutable data/MIME references only for
requested slots. Snapshot adoption compares values and assigns a monotonically
increasing content revision when data or MIME changes; removed slots are pruned.
Rendering an existing slot only reads its small revision token, with no payload
hashing, base64 decoding or payload-sized cache key allocation on that path.
Data references are not duplicated into keys or logged. Equal values remain
equal even with distinct backing storage; no unsafe pointer identity is used.

Bitmap/error completion and native button activation require both the existing
live owner token and the current content revision. Button identity includes
both revisions. Hidden A→B→A and removal/re-add cannot revive an old completion.
An inline preview records its slot/revision and closes when that payload changes;
opening a Markdown preview clears inline provenance so unrelated replacement
cannot close it. Reopening valid new content still works.

## Regression coverage

`TestInlineImageContentAudit*` uses production PNG/base64 decoding and mounted
buttons/pixels: same-length different content, empty/invalid/MIME-only recovery,
held stale success/error, hidden payload ABA, removal/re-add, unchanged image
with streaming text updates, held pointer, fresh preview, own-preview invalidation,
and independent Markdown preview preservation. Existing
`TestInlineImageAudit*`, `TestMountedMarkdown*`, `TestAttachment*` and
`TestImagePicker*` remain applicable. These are headless native toolkit tests,
not actual hosted desktop image comparisons.

## Remaining boundaries

No aggregate per-session cache budget/LRU/worker limit is claimed. Payload
references for removed requested slots are dropped for content correctness,
not as a comprehensive memory policy. Decoder dimensions/MIME/data-URL parity,
network resource rules and generic App.Async cancellation are unchanged.
