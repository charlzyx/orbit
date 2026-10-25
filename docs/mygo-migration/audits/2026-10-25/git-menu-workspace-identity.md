# Git menu identity and file-history navigation

This increment follows the independent embedded-diff patch. No toolkit/private
API changes, credentials, live repositories, remote mutations or release changes.

## Findings and correction

MyGo `ui/menu.go:buildMenu` resolves an OS-menu choice against the newly built
button identity, not the old closure. A captured-generation check inside a fresh
closure is therefore insufficient. Buttons owning Changes, History/Graph,
commit-file, issue and pull-request menus now have public keys incorporating the
repository/job generation, dataset identity and logical target. Changes also key
staged/unstaged status; issue/pull targets include URL. Ordinary same-snapshot
frames preserve menu behavior; cancel/refresh/workspace epochs or replacement
snapshots invalidate stale choices. No hidden path is copied/opened after an
invalidated choice.

A mounted normal Changes → File history menu action reproduced a nil dereference:
`SetTab` reset the ListState while its rows were being built. File-history requests
now carry workspace/generation/path/follow and apply at the next View entry before
building lists. Superseded requests are discarded. Rust `git_panel/mod.rs:4688`
requests `filter_history_path(path, true)`; `2207–2214` clears selected detail and
refreshes history. Changes now follows renames as well, rather than inheriting a
false Follow value. Commit-file history uses the same safe navigation route.
Existing Discard confirmations and Git operation handlers are unchanged.

## Evidence

- `git-menu-red2.log`: same-workspace replacement snapshot left the old menu active.
  The A→B and A→B→A cases in this fixture already invalidated naturally; they are
  defensive regression coverage, not claimed reproduced bugs.
- `git-menu-matrix.log`: normal File history menu nil-dereference stack in native
  `listFrame.build`, triggered by resetting active list state.
- Mounted tests: stale/fresh Changes menus, A→B→A before a new frame, same-path stage
  changes, same-workspace replacement, cancellation, stable normal frames, same-
  workspace no-op switch, History/Graph/commit-file identity, issue/pull menus,
  normal copy/open/history actions, and safe deferred history/rename following.
- Unit guard tests complement mounted navigation: queued history cannot cross
  cancellation, Close, or workspace roundtrip epochs.

Fixtures use temporary repositories and callback collectors; URLs are inert
`example.invalid` strings and no external browser/GitHub operation is performed.
Package race/vet results are recorded with the handoff. These checks do not imply
native macOS menu/pixel acceptance or full Git feature parity.

## Synthetic fixture isolation and real-refresh coverage

Under concurrent race-test load the static embedded-diff fixture crossed the
production four-second refresh interval. Its intentionally non-repository temp
folder then produced a Git error banner, moving the viewport from
`{0,145,1222,574}` to `{0,295,1222,424}` while the drag test retained its original
coordinates. The data itself remained 151 lines per source. A forced-expiry
native-layout diagnostic confirmed this (`embedded-fixture-expiry2.log`); the
initial suspicion that refresh cleared the data was disproved.

Only that synthetic geometry fixture now uses a future refreshed timestamp to
avoid unrelated background loading. Production refresh and all geometry/drag
assertions remain unchanged. `TestEmbeddedReviewRealRefreshReplacesWidthsAndHeaders`
uses an actual isolated Git repository: committed baseline → unstaged long line →
restore old path and stage a new longer file → real Refresh. It verifies updated
width, new staged header and removal of the old header. Existing real-refresh/
mutation tests remain in the full package race run.
