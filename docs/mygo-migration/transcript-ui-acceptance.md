# C10/C11/C12: Typed transcript source contract and evidence

Reference Orbit: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
Native dependency: MyGo `9c69cf829ece196c5b3a3601f1c7bbe70526376f`.
Implementation: `internal/transcriptui`, mounted by `internal/ui/transcript.go`.
All inputs in this report's tests/captures are explicitly synthetic Pi protocol
fixtures. They are not model output, live provider usage, or proof that a model
ran. No Rust sidecar, browser/webview, API credential, paid model call, user Mac,
GitHub write, or production session was used.

**Verdict:** source-derived implementation and scoped native CPU tests pass.
**C10/C11/C12 full-row parity is not verified.** The concrete C11 virtualization
limit below is reproduced, not waived. Pixel equality and actual macOS/Linux/
Windows native input/lifecycle verdicts remain separate and unexecuted here.

## Source-first findings that change the implementation

- `transcript.rs:1081::load_from` merges consecutive assistant messages into one
  logical turn, despite intervening tool-result rows. Tool results attach to the
  newest matching call ID. Context-only system/custom records are not chat rows.
  An assistant carrying `toolCall` still carries its prose/thinking: classifying
  it by `ToolName` alone loses that content. The old Go host had that defect.
- `transcript.rs:282,383` makes compaction and branch summaries their own context
  boundaries. They are not assistant answers and do not contribute answer copy.
- `transcript_view.rs:2040–2211` preserves pre-answer and post-answer activity
  sequence. A settled turn fold can hide pre-answer work, not later prose.
- `transcript_view.rs:2211,7053` renders changed files **only once at the task
  tail**, not as historical per-turn cards. `transcript.rs:2146,2469` aggregates
  changed paths across the task. `app/session.rs:2648::review_opener` captures
  authoritative `latest_turn`; no visible-message count is a checkpoint ID.
- `transcript_view.rs:1044,1242` suppresses only the final answer's ordinary
  footer when a tail card exists. The single copy/time/usage footer follows the
  file card. It copies the final answer, not a file list or raw tool envelopes.
- `transcript.rs:554–598` deliberately discards tool-result image blocks from
  transcript text. Output images are not an omitted source feature. User and
  assistant image blocks remain available to the host's real image/lightbox UI.
- Pinned MyGo already has container-wide text selection. Its public
  `Element.Selectable`, `SelectedTexts`, and `CopyData`, plus Markdown
  `Options.ShareSelection`, avoid a core fork for mounted cross-message text.
  Individual Markdown views, including nested alert views, must not establish
  independent selection scopes.

## Exact behavior cases

Each case is a scoped contract, not a replacement for the master acceptance row.

### T01 · Projection, IDs, ordering, and content boundaries (C10/C12)

Source: `transcript.rs:1081::load_from`, `:630::merge_step`, `:526::tool_result_parts`.

1. Load user → assistant thinking/call → matching toolResult → assistant answer.
2. Include context-only messages, two text blocks, and a compaction boundary.
3. Repeat the same immutable snapshot, then apply cumulative live output shapes.

Expected: one assistant row, ordered steps, one tool, joined result text without
transport envelope/base64; stable first-step representative ID. A summary breaks
merging. Repeated rendering never adds duplicate steps/results. A live execution
without its assistant call stays a real orphan tool row; historical orphan
`toolResult` remains context-only, as in Rust. Invalid partial argument JSON is
not promoted to completed arguments. Source-normalized image hints and explicit
injected skill documents do not leak protocol/skill-body text into user bubbles.

Evidence: `TestProjectSourceLogicalTurnsAndBoundary`,
`TestLiveToolEnvelopeUpdatesAreCumulativeNotDuplicated`,
`TestSourcePostAnswerGroupingAndChangedFiles`,
`TestSnapshotToolImagesAndUserMetadataStayOutOfText`.

### T02 · Typed cards, paths, aliases, and unknown tools (C10)

Source: `transcript_view.rs:3152::render_activity_card`,
`:7249::activity_icon`, `:7391::mcp_tool_parts`, `:5048::activity_preview`.

1. Render shell/read/edit/write/grep/find/list/web/code/task/skill aliases.
2. Use a multiline shell command, path, fetch URL, query, Unicode code preview,
   `mcp__server__tool`, and an unknown tool name.
3. Open, close, reopen, copy, and receive newer actual arguments/results.

Expected: semantic action/icon; path takes preview priority; shell displays the
command, not JSON; fetch URL opens independently without expanding; code preview
uses its first nonempty line; MCP server/tool split preserves full raw name in
model data; unknown names have a faithful generic Arguments/Output fallback.
Independent card/section fold state uses stable identity and resets on session
change. Headers never fabricate results. Truncation is shown only from Pi's
`details.truncation`, `matchLimitReached`, `linesTruncated`, or `entryLimitReached`.

Evidence: `TestTypedToolsAndOutputBounds`,
`TestNativeToolFailureFoldCopyRepeatAndStreaming`,
`TestStateSessionResetAndIndependentToggles`.

Remaining: full alias permutation, click-through URL OS launch, exact file-type
devicons, theme/density typography, shimmer, and matched card pixels are not all
independently verified. Existing native icons and paint-only lexical colors are
wired, but this is not a pixel-equality claim.

### T03 · Actual errors, partial output, abort, and absence (C10/C12)

Source: `transcript.rs:692::message_error`,
`transcript_view.rs:3512::render_tool_error_strip`, `:2250::render_assistant_error`.

1. Fail a tool with multiline actual output while the card is closed.
2. Expand, then replace partial output with its cumulative update and final data.
3. End an assistant with `stopReason=aborted` and its standard errorMessage.
4. Repeat with a provider error, missing output, and missing/null/zero usage.

Expected: failure strip visible while folded; first nonempty error line capped
at 200 Unicode characters; generic failure text only when none was captured.
No concatenation of cumulative snapshots. Cancelled partial answer survives with
quiet Stopped, never a false provider-error card. Missing telemetry stays absent.
No token count, cost, completion result, or historical duration is invented.

Evidence: `TestSourceFailuresAbortAndMissingMetrics`,
`TestNativeToolFailureFoldCopyRepeatAndStreaming`,
`TestNativeLogicalTurnFoldAbortAndMissingUsage`.

### T04 · Arguments/output bounds, syntax, diff, and copy (C10)

Source: `transcript_view.rs:97–138`, `:3551::render_tool_detail`,
`:3973::build_edit_diff`, `:4108::edit_patch_text`.

1. Show 16 and 17 output lines, then 401 lines; toggle repeatedly and copy folded.
2. Use edit batches with unchanged/added/deleted lines and a write ending in LF.
3. Diff old/new grids over 250,000 cells; include Chinese/emoji and malformed JSON.

Expected: argument cap 4,000 characters; output cap 100,000; fold after 16 lines,
preview 12, expanded paint 400. Diff fold after 40, preview 24, paint 400.
Copy uses the whole captured/capped section, not just its painted preview.
Shell/JSON/file syntax is color-only. A trailing file LF does not paint a phantom
row. LCS deletions win ties; oversized grids use whole-block replacement.
Diff clipboard uses `+`/`-`/space prefixes and final LF; cosmetic breaks are
omitted. Source edit arguments are `edits[].oldText/newText`; write uses content.
Source added/removed counters count argument lines including a trailing split,
not a fabricated Git diff. Distinct argument/output/diff toggles do not collapse
the containing card. Diff rows retain the source's 24px marker gutter and
full-row change washes; selection copies code text, not the decorative gutter.

Evidence: `TestTypedToolsAndOutputBounds`, native repeated-fold/copy fixture.
Additional fault/large-grid/property coverage is recorded in the JSON evidence.

### T05 · Read-only questionnaire history (C10, not C17 submission)

Source: `ask.rs:223–356`; `transcript_view.rs:2924::render_ask_card`.

1. Render valid multi-question arguments, malformed questions, and no output.
2. Deliver canonical selected-label/custom-answer envelope, then declined output.

Expected: valid questionnaires render question/chip/numbered options, chosen
marks, custom answer, Waiting or No answer. Invalid questions retain the generic
tool card. This component never sends authorization answers: live questionnaire
buffering/replay remains the host's separate C17 obligation.

Evidence: `TestQuestionMalformedFallbackAndExactAnswer`,
`TestNativeQuestionRecordWaitingAnsweredDeclined`.

### T06 · Logical turn, live clocks, metrics, and file-tail ownership (C12)

Source: `transcript.rs:130::usage`, `:145::generation_time`,
`transcript_view.rs:4432::usage_metric`, `:4500::usage_breakdown_card`,
`:4820::render_working_indicator`, `:7008::summary_time_label_at`.

1. Aggregate provider usage across two assistant steps, including cache/cost.
2. Observe a live turn, a tool/round-trip gap, settlement, repeated settled
   snapshots, and a session change. Reload history without a live observation.
3. Hover metrics; render task-tail changed files and copy the final answer.

Expected: sum only supplied usage. Cost accepts a scalar or `cost.total` and
retains sub-cent precision. Cache share uses input+read+write and never rounds
partial hits to 100%. Rate requires reported output and durationMs >=1 second.
The live wall clock is measured from observed streaming, includes tool gaps,
stops at settlement, and is never derived from historical timestamp gaps.
One fold and one footer per logical turn; source-ordered work groups. The task
tail merges paths, shows 3 then at most 12 files, captures the host's authoritative
latest checkpoint for Review, and owns the single final footer. Tail time reads
Today/Yesterday plus 12-hour clock, otherwise short date. No historical ordinal
is inferred from compacted visible prompts.

Evidence: `TestObservedWallTimeDoesNotInventReloadTiming`,
`TestSourceMetricFormattingAndDateBoundaries`,
`TestNativeUsageHoverTailOwnsSingleFooter`,
`TestNativeSummaryAndChangedFileLimits`.

Remaining: source per-step thinking timer/reload-estimate and active-generation
fallback clocks are not yet ported; no substitute metric is fabricated. Source
shimmer and the exact all-theme/font-density visual matrix remain open. Awaiting
first reply tail state and complete navigation-rail behavior are host obligations,
not established by these component tests.

### T07 · Cross-message selection, context menu, and virtualized limit (C11)

Source: `transcript_view.rs:302::TextSelection`, `:518::selected_text`,
`:912::text_selection_menu`; pinned MyGo `ui/textselection.go`.

1. Drag from a Chinese-containing user paragraph through multiple Markdown
   blocks and another mounted message; press Cmd/Ctrl+C.
2. Append streamed text, reopen the context menu, copy the full clicked message,
   then dismiss with Escape. Empty selected ranges must not add blank lines.
3. Repeat across three actual virtualized-list rows, then scroll the selection's
   endpoints out of the list's mounted range.

Expected for mounted text: one selection scope, Rune-safe copies, single-newline
block joining, Copy Selection and Copy Message target the correct content, no
unselected controls in clipboard, no clipboard mutation on dismissal.

Evidence: `TestNativeCrossMessageMarkdownSelectionAndMenu`,
`TestNativeVirtualizedCrossRowSelectionAndEndpointLimit` and separate host tests.

**Reproduced mismatch, not a pass:** MyGo `syncTextSelection` clears its selection
when a selected endpoint is unmounted (`ui/textselection.go:166–173`). Rust keeps
cloned SelectedBlock text and can copy the existing selection after scrolling
(`transcript_view.rs:518`). MyGo exposes no public setter/restoration API for a
container-spanning selection or a virtual-list pin for selected text endpoints.
Retaining only a clipboard string would risk copying stale selections after
focus/click changes and would not restore highlight semantics, so no such fake
fix is installed. Closing this requires a narrowly reviewed upstream/public
selection-lifetime hook or an app-owned selectable text/layout subsystem; a
wholesale MyGo core fork was not introduced. Drag gutter hit testing also uses
MyGo glyph bounds rather than Rust's registered block bounds; exact gutter
behavior is not claimed. Real OS clipboard, accessibility, and IME remain gates.

## Reproduction and artifacts

Use the pinned toolchain environment:

```
export PATH=/workspace/shared/orbit-toolchains/go/bin:$PATH
export GOPATH=/workspace/shared/orbit-toolchains/gopath
export GOCACHE=/workspace/shared/orbit-toolchains/go-cache
export XDG_CACHE_HOME=/workspace/shared/orbit-toolchains/cache
go test -race ./internal/transcriptui -count=1
go vet ./internal/transcriptui
ORBIT_TRANSCRIPT_CAPTURE_DIR="$PWD/docs/mygo-migration/probes/transcriptui" \
  go test ./internal/transcriptui -run TestNativeTranscriptCapture -count=1
```

`probes/transcriptui/typed-tools-expanded.png` and `typed-tools-failure.png` are
native MyGo CPU fixture images reviewed for clipping/content. They are explicitly
labeled in-image. They are **not matched Rust captures** and establish neither
pixel parity nor a native OS/GPU execution pass. The host owner maintains mounted
application tests/captures independently. No skipped optional test is counted as
executed. The machine-readable report keeps full-row parity false.
