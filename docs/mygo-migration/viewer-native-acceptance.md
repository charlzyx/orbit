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
- Dock resize/full-width controls and exact source toolbar/menu composition
  still need reconstruction and desktop captures.
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
