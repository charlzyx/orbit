# C10/C11/C12: Typed transcript source contract and evidence

Reference Orbit: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
Native dependency: MyGo `9c69cf829ece196c5b3a3601f1c7bbe70526376f`.
Implementation: `internal/transcriptui`, mounted by `internal/ui/chat.go`.
All rendering inputs in this report's tests/captures are explicitly synthetic
Pi protocol fixtures. They are not model output, live provider usage, or proof
that a model ran. A separate credential-free actual Pi transport suite is recorded
below; it does not request model inference. No Rust sidecar, browser/webview, API credential, paid model call, user Mac,
GitHub write, or production session was used.

**Verdict:** source-derived implementation and scoped native CPU tests pass.
**C10/C11/C12 full-row parity is not verified.** The concrete upstream C11
virtualization mismatch is reproduced and closed by the opt-in, hash-verified
selection addition below. Pixel equality and actual macOS/Linux/Windows native
input/lifecycle verdicts remain separate and unexecuted here.

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
  independent selection scopes. The canonical build additionally mounts the
  attributed `third_party/mygo-selection` snapshot-lifetime addition; ordinary
  upstream Selectable behavior remains unchanged unless explicitly enabled.

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

Remaining: exact file-type devicons and source shimmer are not implemented in
this component. Full alias permutation, click-through URL OS launch, theme/density
typography, and matched card pixels are not independently verified. Existing
semantic SVG icons and paint-only lexical colors are wired, but this is not a
pixel-equality claim.

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
partial hits to 100%. Settled rate requires reported output and a duration >=1 second: Pi durationMs
takes precedence over measured local delta time. Live rate sums settled step
clocks plus the current request receive clock, never the UI heartbeat.
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

Receive-clock implementation: `pirpc.Client.readStdout` timestamps each actual
record before parsing/enqueue as local-only `Event.ReceivedAt`. Pure ParseEvent
and reloaded messages have no receive clock. `agent.sessionState.applyAt` keeps
per-message reasoning and generation clocks, counts only delta gaps <=1s, pauses
at tool execution, resets on boundaries/cancellation, and preserves measured
values and real streamed usage when the final snapshot omits them. Snapshot
cloning also clones optional duration pointers. `turn_start` supplies the actual
request origin, including prefill; absent clocks are never reconstructed from
heartbeat batches. Reloaded reasoning uses the source's adjacent timestamp-gap
estimate only in the reasoning header, with an explicit estimate tooltip that
notes tool time is included, never as provider telemetry or a rate denominator.

Native thought cards follow appended text only while already at the end; upward
scroll detaches and reaching the bottom resumes following. Fold state and scroll
state persist independently and reset with actual session identity. The source's
unlabelled chars/4 token estimator is deliberately not represented as reported
usage: without supplied output tokens the live rate remains absent. This truth
boundary is explicit rather than a fabricated token count or a pixel-parity claim.

Additional evidence: `TestStreamClock*`, `TestEventParseDoesNotInventReceiveTime`,
`TestReaderStampsArrivalBeforeDelayedDrain`,
`TestMeasuredAndEstimatedThinkingClocksAndRates`, and
`TestNativeThinkingTimingAndStreamFollowDetach`. Credential-free actual Pi1.4.0
protocol integration passed; it submits no prompt/model request and asserts zero
provider cost/tokens/tool calls.

Remaining: source shimmer and exact all-theme/font-density visual matrix remain
open. MyGo's ordinary nested scroller hands an already-blocked wheel to the
parent, but does not forward the unused fraction of a wheel that crosses the
inner boundary in the same event (`ui/input.go::scrollBy` vs source
`split_thinking_scroll`). That precise residual-wheel behavior remains a concrete
interaction gap; no full-row pass is claimed. Awaiting-first-reply tail state and
complete navigation rail remain host obligations, not established by these tests.

### T07 · Cross-message selection, context menu, and virtualized retention (C11)

Source: `transcript_view.rs:302::TextSelection`, `:518::selected_text`,
`:827::register_text_selection`, `:912::text_selection_menu`; `app.rs:1333`
pre-focus copy interceptor; pinned MyGo `ui/textselection.go`.

1. Drag from a Chinese-containing user paragraph through multiple Markdown
   blocks and another mounted message; press Cmd/Ctrl+C.
2. Append or replace live text, reopen the context menu over another message,
   copy the selection/full clicked message, then dismiss without a clipboard write.
3. Select three virtual-list rows, scroll all selected endpoints away, copy,
   and scroll back. The selected row nodes must actually unmount; copied text
   remains immutable and highlights return at source-correct UTF-8 boundaries.
4. Drag after scrolling new rows into view. Only blocks in the initial mounted
   snapshot can extend this drag, exactly as Rust `extend_selection` specifies.
5. Switch actual session identity, remove/disable the scope or its opt-in, click
   outside, and repeat selection. Verify no stale selection or node retention.
6. Move focus programmatically to the composer. Source Cmd+C still copies the
   transcript selection until an outside click clears it; afterward normal input
   copy resumes. A right-click elsewhere in the transcript preserves selection.

Expected: shared selection, source single-newline joining with empty ranges
skipped, correct menu targets, current-layout block-boundary gutter hit testing,
no unselected controls in clipboard, no write on dismissal, and no virtualization
bypass. The snapshot contains immutable strings/numeric IDs, never row nodes,
editors, layouts or callbacks. The initial mounted block count bounds its size.

Evidence: `TestNativeCrossMessageMarkdownSelectionAndMenu`,
`TestNativeVirtualizedCrossRowSelectionAndEndpointLimit` (unchanged upstream
behavior intentionally reproduced), `TestNativeRetainedVirtualizedTranscriptSelection`,
`TestSelectionCompatibilityGuard`, and `TestOrbitSelection*` in
`third_party/mygo-selection/orbit_selection_test.go.txt`. Mounted host cases are
maintained separately in `internal/ui/transcript_integration_test.go`.

**Why a narrow compatibility addition was required:** MyGo `syncTextSelection`
clears missing endpoints. Its public API cannot restore a cross-container range.
The explicit opt-in `Element.RetainTextSelection(sessionIdentity)` preserves Rust's
cloned block semantics; `internal/transcriptui.RetainSelection` mounts it on the
actual shared session scope. The canonical wrapper validates original/result
hashes and loads only five explicitly named selection files into a temporary
verified dependency copy. No module-cache or go.mod/go.sum changes occur.
`cmd/orbit` fails before application startup if either composer or selection
compatibility is absent. A stale clipboard cache is not used as a substitute.
See `third_party/mygo-selection/README.md` and `baseline.json` for design,
attribution, file hashes and lifecycle/security regression cases.

Actual OS clipboard, accessibility, IME, and matched visual evidence remain
separate gates. The full upstream toolkit race suite has one allocation-budget
failure also reproduced on untouched pinned MyGo: steady-frame 3 and warmed
scroll-frame 11 allocations versus its expected 2. This is documented as an
upstream/environment failure, not reported as a full-suite pass.

## Latest integration verification

Canonical `scripts/go-native.sh test -race ./internal/ui ./internal/transcriptui
./cmd/orbit -count=1` passed on 2026-10-25: host 20.766s, transcript 1.335s, and
application compiled. This includes the mounted real-session selection scope
and the actual-receive-clock message projection. Agent/Pi/typed race with the
credential-free actual Pi executable passed separately (33.519s / 4.977s /
1.317s). Optional captures were executed explicitly, not counted from skips.

## Reproduction and artifacts

Use the pinned toolchain environment:

```
export PATH=/workspace/shared/orbit-toolchains/go/bin:$PATH
export GOPATH=/workspace/shared/orbit-toolchains/gopath
export GOCACHE=/workspace/shared/orbit-toolchains/go-cache
export XDG_CACHE_HOME=/workspace/shared/orbit-toolchains/cache
scripts/go-native.sh test -race ./internal/transcriptui -count=1
scripts/go-native.sh vet ./internal/transcriptui
python3 scripts/test-go-native-overlay.py
scripts/go-native.sh test -race github.com/egoist/mygo/ui \
  -run '^(TestOrbitComposer|TestOrbitSelection|TestSelectable)' -count=1
ORBIT_TRANSCRIPT_CAPTURE_DIR="$PWD/docs/mygo-migration/probes/transcriptui" \
  scripts/go-native.sh test ./internal/transcriptui \
  -run '^TestNative(TranscriptCapture|RetainedSelectionCapture)$' -count=1
```

`probes/transcriptui/typed-tools-expanded.png`, `typed-tools-failure.png`, and
`selection-restored.png` are native MyGo CPU fixture images reviewed for clipping/content. They are explicitly
labeled in-image. They are **not matched Rust captures** and establish neither
pixel parity nor a native OS/GPU execution pass. The host owner maintains mounted
application tests/captures independently. No skipped optional test is counted as
executed. The machine-readable report keeps full-row parity false.
