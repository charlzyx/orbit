# GoPi display identity, native Copy routing and bounded chrome alignment

## Scope

- User-facing app/window/About/bundle/DMG display names become **GoPi**. The
  executable, bundle identifier, private `Orbit Go Test` state directory,
  configuration keys and existing release identifiers remain unchanged. An
  actual old-bundle → renamed-bundle fixture verifies existing data reuse.
- macOS sends native Edit/Copy as a command, not necessarily as a raw Cmd+C key.
  The retained transcript-selection overlay already handled the latter but
  missed the former. A real mounted cross-message drag followed by
  `Tester.Command("copy")` reproduced an empty clipboard before the five-line
  fix. Input selection/cut/paste, context menu, reset and copy-button regressions
  remain in place. Linux command routing is not physical macOS shortcut proof.
- Collapsed workspace groups intentionally retain current, pinned and running
  sessions, matching the original Rust contract. Only the explanatory English
  and Chinese hover text changes; ordinary sessions still collapse.
- Bounded visual alignment uses the original sidebar header 26px content height,
  12px/2px vertical spacing, 12px chevron/16px folder, separate count, session
  22px leading inset and selected title weight. Thought uses a content-width
  disclosure; ChangedFiles uses the source two-line title/totals and icon tile.
  Preferred external apps use their real discovered icon plus a chevron menu.
  Existing actions, menu entries and visibility conditions remain unchanged.

## Safety and remaining differences

App icons are decoded once on the existing discovery worker from its already
loaded icon bytes. Frame rendering reads the published immutable cache; it does
not read files or decode images. Missing/invalid icons use a generic folder,
with the actual application name retained as the accessible label and localized
hover tooltip. The native menu cancellation test uses the headless host's
`CloseMenu` seam: a raw Escape key there is an app event, not Cocoa tracking.

The first style regression run found two sticky-header failures: newly added
outer vertical margin was absent from the measured header but present in the
floating presentation. Moving the same spacing inside the bound row restores
measured push-off/partial-header ownership; the original assertions are retained.
An earlier new-test compile error (float32 vs the preference's float64) is also
retained in diagnostic logs, not counted as a passing run.

Generic changed-file edit glyphs, Worked/Show-more positioning, full Rust More
menu/search/dirty-count features, new upstream Mermaid/session-review behavior
and a possible MyGo DiffView adapter are not part of this patch. Native DiffView
exists at the current pin, but replacing Review requires a separate contract
review, not a styling substitution.

## Validation

Targeted actual native-command, input, data-identity, locale and packaging tests
passed before style work. Final combined validation is recorded only after the
frozen tree completes its gate; earlier focused results are not a full-suite claim.
Synthetic before/after CPU renders contain no user conversations. They are
in-process Linux renders, not macOS screenshots or evidence of Cocoa menu input.

`BenchmarkChromeFrameInput` uses the same runner/data for empty and 1000-message
UI histories, idle frames and type/backspace operations, with 20 iterations and
three repeats. This is a UI-frame allocation/time check, not a Controller snapshot
benchmark, actual IME input or a claim about the user's observed macOS CPU.

The first combined full race run completed with 3748 passing test events,
19 conditional skips and one failure in the pre-existing recovery fixture.
The same exact test reproduced 2 failures / 20 runs on both the public base and
this candidate. Failure-only real-child tracing showed `auth.list` followed by
`break_stdin` before `quota.list`: Connected was visible while the final startup
probe could still write. Only that test now waits for the actual final probe and
an ordered RPC barrier before deliberately closing stdin. Original failure,
recovery-budget and child-count assertions remain unchanged, with no new sleeps,
production edits, retries or extended deadlines. The fixed case passed 30 race
runs; a fresh complete agent package run also passed. The failed initial full run
is retained, not converted into a pass by these narrower results.

Final v2 gate (2026-10-28): full race 3749 pass / 19 conditional skips;
20 toolkit, 32 menu and 115 terminal tests passed (terminal: 2 conditional
skips). All 1679 enumerated root top-level names ran; all 84 selected exact
regressions passed without skip. Vet, five platform cross-builds, packaging,
icon, wrapper and offline harness suites passed. 1917 non-Markdown source/asset
SHA256 values remained unchanged throughout. These are automated cloud results;
physical macOS Cmd+C and the renamed installed app still require user validation.
