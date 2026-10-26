# C09 top-level transcript code folding candidate

Scope: top-level transcript prose fences only. This does not claim full Markdown,
Mermaid, nested-container, document-preview, typography, or native pixel parity.
No toolkit, decoder, loader, network, or URL changes.

## Pinned source contract

Rust reference 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac:
`transcript_view.rs:118–121,6624–6630` uses >28 lines and a 24-line preview;
`:6683–6725` provides remaining/less toggles and remeasurement;
`:2117–2126` uses semantic assistant `!paint.live`, not whether a fence closes;
`:6550,6589–6597` copies the full joined body independently of expansion.
Expansion identity includes message/prose/block ordinal. `transcript.rs:1128,1166`
clears expansion on load/reset. Nested quote/alert source paths at
`transcript_view.rs:6234–6262,6350–6406` render container lines rather than recursively
folding fenced blocks. Standalone documents pass collapsible=false at :6135–6155.

## Bounded implementation

The existing caller slot and owner-keyed Markdown subtree own fold state. A
source revision/document identity caches top-level parser metadata; expansion
uses top-level block ordinal, not flattened parser indices, and survives source
edits. Missing slots are pruned. Ownership reset discards the subtree. A held
button is content-revision-keyed and checks the current owner/revision on activation.

Only the first 24 parsed literal lines are copied into a cached safe tilde fence,
longer than every tilde run in the preview. The language tag is separated from
the opening fence and preserved for the original tokenizer. Public Markdown
renders this literal-only code body with its usual selection/highlighting.
There is no recursive app hook, external image/URL/HTML loader, second code
renderer, or clip-only hidden selectable tail. The expanded body is the original
`b.Body`. Original full-copy reads `b.Code()` only on activation. No full-body
`Code()` or hash is used to prepare a frame. Preview allocation is source-linear,
not a rows-by-columns projection. Unchanged preparation is tested allocation-free.

Ordinary fold clicks deliberately clear old transcript selection; no retained
selection interaction privilege is added. Real List layout measures the changed
body; no guessed height, forced scroll, or fake footer-height acceptance.

## Evidence and status

The original no-fold mounted RED is preserved in
`/workspace/shared/orbit-code-fold-label-red-final-2117.log`.
Earlier isolated owner/threshold tests passed race x3 (21 actual top/subtest
names), and real App List remeasure/anchor and live-tail tests each passed x3.
Actual transcript drag selection, collapsed full-copy, settled open-fence and
semantic-live cases passed x3 with the JSON hitset checked.

The initial multi-block test failed because Tester.Texts contains both the
button accessibility label and visible text; it counted duplicate labels as
controls. That failure log is preserved at
`/workspace/shared/orbit-code-fold-list-2138.jsonl`. The assertion now observes
separate A/B expanded tail lines and C's still-hidden tail instead. This is not
silently relabeled a pass.

Final source adds a parser-document identity check, themed padding, actual
keyboard activation, idle allocation assertion, and the corrected multi-block
assertions. Final focused race x3 passed: 26 exact top/subtest names, 78 run/78 pass,
zero skip/fail, UI 11.644s (`orbit-code-fold-final-2144.jsonl`; hitset
`orbit-code-fold-final-hitset-2145.json`). Copy/caption regressions race x3
passed: 27 exact names, 81 run/81 pass, zero skip/fail, UI 8.320s
(`orbit-code-fold-regressions-2145.jsonl`; corresponding hitset JSON).
Targeted UI vet passed (`orbit-code-fold-vet-2145.log`). Per coordinator, no duplicate full-UI or cross build
is run while the combined aggregate owns those gates. Hosted Rust fixture
execution/native captures remain separate and are not claimed here.
