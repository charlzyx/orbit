# Terminal per-tab close: source-derived reachability

Pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `terminal.rs`:

* 1880–1964: each tab has its own close control. Close targets that tab index,
  including inactive tabs, and stops propagation so selection cannot run on a
  stale index after removal.
* 1735–1759: closing a tab left of active retains the active runtime; closing
  active chooses its right neighbor or the new last tab. Closing the final tab
  hides the panel. The shell receives SIGHUP through entity drop; no dirty or
  running-job confirmation is present here. No middle-button tab-close binding
  was found; none is added.
* 1974–1991: tab strip scrolls horizontally, preserving toolbar space.
* 2055–2070: toolbar x calls toggle (hide), NOT close-active. Existing Go used a
  close-active x plus separate hide chevron: they were not source-equivalent.

Go Panel.CloseTab already implements the removal and neighbor semantics. This
increment only exposes it per tab: sibling select and close buttons inside a
stable-ID chip avoid nested click bubbling; a horizontal strip keeps controls
reachable without displacing the fixed toolbar. The extra close-active toolbar
button is removed and the existing hide action receives the source x glyph.
Keyboard close-active remains unchanged.

## Target lifetime

The synchronously executed UI helper resolves a tab ID against the current
same Panel snapshot, then invokes CloseTab without any await or callback gap.
Missing IDs do nothing. Select uses the same current-ID lookup. Structural
operations are owner-UI operations; async startup only updates its own tab.
This helper is NOT a new general atomic cross-thread close API.

Panel.nextID increments only on newTab; SetWorkspace changes only the next
shell's workspace and leaves existing shells running. No Clear/reset path
reuses IDs. Close permanently closes the Panel, so that object cannot admit
new tabs afterward. The tests exercise removal and new tabs through A→B→A,
including a pointer-down held across removal and a captured stale-ID intent.
No stale ID may select/kill either replacement.

## Evidence and limits

`terminal-tabs-red.log`: three real disposable PTYs ready; mounted UI fails
because inactive tab's close control does not exist. Final focused tests cover
per-tab containment, inactive close and exact live owner retention, termination
of only the intended shell, active neighbor selection, final panel close,
held pointer/ID ABA, and existing exit/restart/startup recovery. Toolbar hide
and reopen retain the same Terminal object, unchanged PID-bearing shell output
and grid content, with Done still open for every running PTY.

All shells use synthetic temporary directories. No user terminal, credentials,
setup/deletion data or toolkit modification. The shell reports its PID into its
output; no host-wide process inspection is used. Native Mac rendering and
pixel parity still need the independent capture lane; these mounted tests do
not certify either.
