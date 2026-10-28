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
