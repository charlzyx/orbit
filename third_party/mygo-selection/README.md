# Opt-in MyGo retained transcript selection

Status: mounted by the explicit hash-verified canonical build wrapper. The host
shared transcript scope enables it using the actual session identity.
Upstream: `github.com/egoist/mygo` at
`9c69cf829ece196c5b3a3601f1c7bbe70526376f`, MIT, license retained in `LICENSE`.
Behavioral reference: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

## Source contract and minimal API

Orbit `transcript_view.rs:302–543` snapshots the text/identity/order of blocks
mounted when selection begins. Copy remains independent of later rendering,
virtualization, and text mutation. Only initially captured blocks can extend
that drag. On remount the selected UTF-8 byte ranges clamp to the current text.
`register_text_selection:827` clears on outside left-click and opens the Copy
Selection / Copy Message menu without replacing a selection on right-click.
`app.rs:1333` intercepts Cmd+C before focus dispatch when a selection exists,
including when the composer gained focus programmatically.

Pinned MyGo `ui/textselection.go:166–173` instead clears a container selection
when an endpoint unmounts. Keeping all rows rendered violates virtualization;
a clipboard cache alone cannot restore highlights or follow selection lifetime.

The added opt-in native API is:

    scope.RetainTextSelection(sessionIdentity)

Call it on the shared Selectable container every build with the stable actual
session/document identity. Removing the call, changing identity, removing or
disabling/hiding the scope, or clicking outside clears the snapshot. Ordinary
Selectable containers are unchanged. A retained scope keeps only immutable
strings and numeric IDs, never editors, nodes, layouts or callbacks. The initial
mounted block count is the lifetime bound, independent of full transcript size.

After drag release the focus moves to the owning container. This also avoids
MyGo's existing focused-row pin retaining a virtual row. A retained selection's
Cmd+C intercept is active only when a nonempty selection exists. Clicking outside
restores normal native text-input copy. Ctrl+C/Cmd+C inside an IME/text field is
otherwise unchanged. Existing native SelectionData/SelectedTexts and context
menu callbacks receive the original captured strings and source-correct ranges.

## Changes and attribution

`selection-retention.patch` is the reviewable unified diff. Explicit overlay
files carry the same resulting source; their hashes and original pinned hashes
are listed in `baseline.json`. This artifact changes only:

- `ui/context.go`: one opt-in scope metadata pointer in native state.
- `ui/textselection.go`: opt-in snapshot projection, copy/selection helpers,
  current-layout gutter hit testing, and unchanged default path.
- `ui/input.go`: scope lifetime on outside click and retained-copy interception.
- Added `ui/orbit_selection.go`: opt-in API and bounded snapshot helpers.
- Added `ui/orbit_selection_test.go`: native deterministic regression fixtures.

No GPU/platform/initialization/transport code, source pin, public existing API,
module-cache file, repository go.mod/go.sum, or composer addition is changed.
No unsafe/reflection, background I/O, network calls, model invocation, credentials,
or OS clipboard integration is used by these tests.

The canonical wrapper must verify the exact module sums and each upstream hash,
then apply only the explicitly listed overlays to its temporary verified source
copy. It must reject unexpected existing additive files, unknown paths, or hash
mismatches; no wildcard discovery. The `.go.txt` suffix avoids accidentally
including review artifacts as application packages. The canonical `scripts/go-native.sh` wrapper now mounts these explicit overlays
alongside the composer addition. `cmd/orbit` fails loudly if either required
compatibility capability is absent. Original upstream behavior is independently
regression-tested without enabling the opt-in in those fixtures.

## Tests

Canonical proof: `scripts/go-native.sh test -race github.com/egoist/mygo/ui -run '^TestOrbitSelection'`. The same fixtures were first tested in an isolated pinned source copy:

- Virtualized endpoint removal preserves exact Unicode clipboard text.
- Nodes actually unmount; retained state count stays bounded; remounted highlights
  restore without retaining off-screen editors.
- Appending/replacing current text leaves copied snapshots immutable, while
  highlights clamp to new text with source UTF-8 byte-boundary semantics.
- Session identity, removed/disabled scope, removed opt-in and outside clicks
  invalidate selection; repeated selection does not reuse old text.
- Dragging into newly mounted rows cannot add blocks absent at drag start.
- Gutter hit testing uses current paragraph boxes and Manhattan distance.
- Native Copy Selection menu, dismissal, copy priority over programmatic composer
  focus and normal input copying after outside-click are exercised.

The existing toolkit suite is also run without enabling this opt-in in existing
fixtures. CPU/native UI tests are not actual OS clipboard/IME/GPU execution or
matched Rust screenshot evidence. Those remain distinct application gates.

## Allocation baseline limitation

The full upstream UI race suite was executed. Its sole failure was
`TestDiffFrameAllocs`: steady 3 and warmed scrolling 11 allocations against an
expected ceiling of 2. The exact test was repeated on untouched pinned upstream
and failed with the same steady/warm values (135.286s). This is not a claimed
full-suite pass or a hidden skip. The short functional suite and all compatibility
regressions are recorded separately from that upstream/environment budget failure.
