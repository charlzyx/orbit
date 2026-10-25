# Files and Review: mounted native interaction checks

Reference: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`; MyGo
`9c69cf829ece196c5b3a3601f1c7bbe70526376f`. These are native MyGo CPU interaction
tests, not same-state Rust/Go screenshot comparisons or desktop OS acceptance.

## Source-derived checks

The tests in `internal/ui/viewer_native_test.go` mount the actual file editor,
tab strip, review tree and virtual diff list. They send pointer/keyboard input
and use disposable files and Git repositories. Tests do not invoke a model.

| Case | Pinned Rust behavior | Executable check |
| --- | --- | --- |
| File text selection, undo/redo and caret | `explorer/viewer.rs` editable buffer, `activate` at 552; editor commands at `main.rs:436` | `TestMountedFileSelectionUndoRedoAndTabCaret`: Unicode selection/copy/replacement, undo restores selection, redo restores UTF-8 caret, switching tabs restores caret and typing focus |
| Failed close/save | `viewer.rs::flush`, `close_tab` at 513 | `TestMountedFileFailedCloseKeepsDraftAndCanRetry`: write failure retains draft/tab and exposes error; successful retry writes the same draft. Keeping the draft is the documented safety divergence from Rust's unconditional close |
| Rename retains editor identity | `viewer.rs::reconcile_rename` at 416 | `TestMountedFileRenameSubtreeKeepsEveryBufferCaretAndUndo`, `TestMountedFileRenameCancelAndFailureKeepState`: native rename prompt, subtree remapping, unrelated dirty tab, undo/caret/focus, cancellation, same-name no-op, and destination-exists failure |
| Cached open and queued reload | `viewer.rs::open` at 448 and `reload` at 537 | `TestMountedFileOpenCachedTabIsImmediateAndKeepsNativeBuffer`, `TestMountedFileOpenCompletionUsesCurrentDocumentAfterReload`: cached tabs activate without stale worker copies; a newer document reload wins over an older open completion |
| Old close completion | `viewer.rs::close_tab`, stable tab identities | `TestMountedFileStaleCloseCompletionKeepsReopenedState`: applying an old completion cannot delete a newly reopened editor's state or caret |
| Hide versus close tab | `viewer.rs::hide` at 506, `on_close_tab_action` at 1223; `main.rs:436–441` | `TestMountedFilePrimaryWAndShiftWHaveDistinctScopes`: primary-W closes active tab; primary-Shift-W hides Files but retains tabs; last tab close leaves Files |
| Close/open/activate ordering | `viewer.rs::open`, `activate`, `close_tab` | `TestMountedFileReopenQueuedBehindCloseCannotStealLaterSelection`: a later tab selection supersedes an earlier queued reopen |
| File navigation | `sidepane.rs::review_move_file` at 932 | `TestMountedReviewFileKeysClampAtEnds`: N/P clamp at the first/last file, not wrap |
| Tree cursor and submission | `sidepane.rs::review_tree_move` at 909 and `review_tree_toggle` at 923; common tree focus at 1908 | `TestMountedReviewTreeCursorWaitsForEnter`: arrows move the cursor only, Enter selects/expands the file, repeated Down clamps |
| Scoped tree keyboard | `sidepane.rs` `ReviewDiffTree` context at 1908 | `TestMountedReviewTreeKeysDoNotHijackToolbar`: leaving tree focus for the toolbar stops tree navigation shortcuts |
| Full-context hunk navigation | `sidepane.rs::review_move_hunk` at 949, `is_hunk_start` at 2359 | `TestMountedReviewHunkKeysNavigateFullContextRuns`: brackets navigate change-run starts even when full-context output omits `@@` headers; deletion/addition pairs use rendered split indexes |
| Logical position across split mode | `sidepane.rs::remeasure_diff_rows` at 433 | `TestMountedReviewSplitPreservesLogicalContextRow`: a context row remains the first visible logical row when paired changes shorten the split representation; fractional offset is not yet proven |
| Refresh/reorder selection | `sidepane.rs::apply_snapshot` at 661 | `TestMountedReviewRefreshRetainsSelectedPathAfterReorder`: adding a preceding changed file retains the selected path, not its old ordinal |
| Markdown preview focus | `explorer/viewer.rs::toggle_markdown_preview` at 1240 and deferred target at 1387 | `TestMountedFileMarkdownPreviewRestoresEditorFocusAndCaret`: preview/source roundtrips preserve buffer/caret, focus the editor or viewer root, and do not leave Enter activating the toggle |
| Hierarchical right-hand tree | `review.rs::tree_rows` at 940; `sidepane.rs::render_tree` at 1875 | `TestReviewTreeSourceOrderingFilterAndExpansion`, `TestReviewTreeRefreshKeepsClosedDirsAndExpandsNewOnes`, `TestMountedReviewTreeHierarchyFilterKeyboardAndResponsiveBounds`: stable ASCII path sorting, ancestor reveal, preserved collapsed directories, keyboard directory toggles and exact source width thresholds |
| Git file diff action | `app.rs:1876`, `sidepane.rs::show_file` at 555 | `TestMountedReviewOpenFileKeepsWholeSnapshotAndCommitFilter`: working-tree action loads the whole Uncommitted snapshot and selects the requested path; explicit commit-ref comparison retains its distinct file scope |

## Reproduced failures and fixes

The first mounted-control run failed four cases: a stale close completion
removed reopened native state; N/P wrapped; bracket navigation did nothing in
full-context diffs; and refreshed files changed the selected path by ordinal.
Two further open tests reproduced an unnecessary asynchronous cached-tab snapshot
and a stale initial-open completion overwriting a newer watcher reload. Both
now use the live document/buffer rather than the old value copy.

A subsequent strict Enter test exposed another issue: the previously clicked
native Button consumed Enter, overriding the moved tree cursor. The tree now
has one focus target, matching the reference's shared focus handle. These
failures have ordinary red/green regressions, not opt-in expected failures.

Native buffers are synchronized before Save/Close/hide begins asynchronous disk
I/O. Close actions are deduplicated per path; a queued reopen carries workspace
and request generation, and old completions do not erase live reopened state.
Diff virtual rows use semantic line keys within the loaded snapshot, and hunk
commands convert original snapshot positions to current rendered rows.

## Evidence and limits

At 2026-10-25 07:29 UTC, the canonical `scripts/go-native.sh test -race
./internal/ui ./cmd/orbit -count=1` passed (UI 31.734 seconds; cmd compiled),
followed by clean `scripts/go-native.sh vet ./internal/ui ./cmd/orbit` and
`git diff --check`. This includes 19 focused Files/Review test functions.
Concurrent changes to other surfaces after compilation require their own
immutable-checkpoint validation; this result is not a blanket branch verdict.

The 07:27 CPU capture pass produced and visually inspected the Review and editor
fixtures. It caught backdrop dots bleeding through the formerly transparent
Review pane. The dock now has the source opaque background, 12px left gutter,
border and rounded shell, covered by
`TestMountedReviewCardOccludesNewTaskBackdrop`.

Remaining acceptance gates include:

- The docked Review tree is now a hierarchical, filterable right column with
  source 30px rows, 26px controls, indentation, status badges and directory
  expansion state. Exact glyph/font rendering and matched Review screenshots
  remain unverified; the file icon currently uses the source's fallback SVG
  rather than Nerd Font-specific glyph selection.
- Dock resize/full-width interactions now have mounted CPU coverage (below).
  Exact source toolbar/menu composition and matched desktop captures remain open.
- Unwrapped diff horizontal scrolling, sticky file headers and fractional
  scroll-offset preservation need source-derived mounted checks and fixes.
- File preview/readonly/image geometry, loading/error states and exact tab
  chrome still need matched native captures. Passing text editing tests does
  not establish font/IME/OS accessibility parity.
- The GitHub PR Files rows still lack the source Open Diff action. The host
  callback is corrected, but an unmounted caller is not a completed workflow.
- Same-state OS screenshots, window focus/IME, native menus and assistive
  technology require desktop acceptance; these CPU tests do not prove them.

## Capture command

`ORBIT_VIEWER_CAPTURE_DIR=<output> scripts/go-native.sh test ./internal/ui
-run '^TestNativeViewerCapture$'` creates three explicitly named CPU fixture
images: default-width Review, wide Review and editor selection. The fixtures
contain generated text and disposable paths, not production session data.
They are visual QA inputs only, not replacements for the V### desktop pairs.

## Rename follow-up (2026-10-25 07:36 UTC)

The mounted subtree-rename test initially failed because the host cleared the
entire native editor-state map after any successful rename. The corrected
adapter remaps only affected path keys, retaining the existing editor objects,
undo/redo history, caret, scroll state, buffer versions and decoded images.
Unrelated tabs are untouched. A same-name rename is a no-op, and failed or
cancelled renames retain both source and destination contents and native state.
The renamed active editor receives focus after the prompt closes.

The complete canonical UI/cmd race suite passed in 35.335 seconds, followed by
clean UI/cmd vet and diff checks. There are now 21 focused Files/Review test
functions plus the optional fixture-capture test. This follow-up changes only
`internal/ui/files.go`, `internal/ui/viewer_native_test.go` and this document.


## Review dock/full-width follow-up (2026-10-25 08:11 UTC)

Source contract: `sidepane.rs:287–391` separates the persisted dock restore
point from the current available window width; `app/view.rs:327–406` opens the
sidebar on expansion and yields full width when a feature page opens;
`app.rs:91` and `app/view.rs:1260` reserve 480px during pointer resize.

The native left-edge handle now implements the source 300–900px drag range
and window reserve, and persists only a completed user drag. Full width is
uncapped by the drag maximum, follows the available viewport/sidebar width,
opens the sidebar, hides the chat/backdrop/terminal, and restores the prior
docked width without overwriting its preference. File identity and Review
scroll survive repeated expansion/restoration.

Twelve mounted checks exercise actual pointer movement/release, repeated
full-width controls, sidebar and viewport changes, feature-page transitions,
selection restoration, session switching while the composer is hidden, Review
tree keyboard focus, native editor undo/redo, hidden-input isolation and idle
frame behavior. The focus test first failed: an unkeyed content row changed
identity when the backdrop disappeared, and the remounted composer took focus
unconditionally. Stable content keys, a keep-focus expansion control and
conditional focus restoration fix that failure.

Tests:
- `TestNativeReviewFullWidthRestoreAndViewport`
- `TestNativeReviewDragClampsAndPersists`
- `TestNativeReviewFeaturePageYieldsAndKeepsRestoreWidth`
- `TestNativeReviewFullWidthRetainsOnlySameDraftSelection`
- `TestNativeReviewFullWidthDoesNotStealTreeFocus`
- `TestNativeReviewFullWidthKeepsReviewScroll`
- `TestNativeReviewDockCoexistsWithFeaturePages`
- `TestNativeReviewDockExcludesSettingsAndUsage`
- `TestNativeReviewFullWidthKeepsComposerUndo`
- `TestNativeReviewHiddenComposerTakesNoInput`
- `TestNativeReviewFullWidthRestoresFileEditorFocusAndUndo`
- `TestNativeReviewHiddenContentDoesNotRequestFrames`

The undo test also first failed: MyGo `ui/runtime.go::prune` discards the
private input editor when its node is no longer built. The fix retains a
single invisible, disabled content column at its measured size, outside flow
layout. It neither paints nor receives input, but preserves native undo and
scroll state. This uses only the public API, not a Router or dependency patch.
The same-session/draft caret bridge and session/workspace keyed composer stop
selection from leaking into a different task. Files/Git/Worktrees now actually
share the content row with the dock; Settings/Usage remain exclusive.

A separate, reproducible toolkit probe in
`probes/review-layout/visibility_access_test.go.txt` was run against an isolated
copy of exact pinned MyGo. Its actual accessibility tree omits the invisible
editor and its value/focus, input/paste/Enter do nothing, no frames remain
pending, and undo works when shown again. It passed `-race` in 1.359 seconds.
The dependency cache and production toolkit sources were not modified. This
is a CPU accessibility-tree check, not a screen-reader/OS desktop verdict.

At 08:23 UTC, the complete canonical UI/cmd race suite passed in 46.857 seconds,
followed by clean UI/cmd vet and diff checks. All twelve layout tests passed in
13.586 seconds; the total focused Files/Review coverage is now 33 functions
plus the optional fixture capture. No hidden-content animation requests were
observed in the explicit streaming fixture. Other ancestry changes, such as
toggling the root sidebar Split while the editor is retained, still need their
own undo/focus acceptance; this batch does not claim general state retention
across every navigation or root layout replacement.

This batch is separate from publication checkpoint0736 and the rename
follow-up. It changes `app.go`, `chat.go`, `review.go`, `review_tree.go`, and
adds `review_layout.go` / `review_layout_test.go`; this document records it.
Matched desktop geometry and OS input remain pending, not inferred from CPU
interaction tests.
