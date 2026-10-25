# P20 second bounded batch: existing chat/workspace keys

Base: clean wave4 a6eec11 plus approved first P20 patch (local baseline
01806346a7fdeb61abdbd56ced8fb10f2644a298). No edits to integration trees.
Unreleased; not delivered .4 and not proof of whole P20 or migration parity.

## Source anchors and changes

Pinned Rust 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac:
- access.rs:58–65 and workflow.rs:61–67 return localized enum display labels.
  Go chat.go now uses the same access/workflow keys for painted chip text,
  tooltips and menu choices; retains public English chip Label, exact protocol
  values, persisted access value and dispatch sequence. Unknown future values
  keep prior fallback rather than exposing a fabricated translation key.
- app/search.rs:42 supplies search.find_in_transcript as placeholder. Go does
  the same without changing Find in transcript public identity or input logic.
- message_scroller.rs:360 uses message_scroller.jump_to_latest; Go visible jump
  button now routes through it, retaining its English public Label and action.
- transcript_view.rs:2362,2639 use transcript.thinking. Go streaming fallback
  renderer now uses that key; completed Thought process is deliberately unchanged
  because this batch has no exact source equivalent for that distinction.
- app/sidebar.rs:1546–1602 workspace actions use sidebar.copy_path,
  icon_and_color, remove_from_sidebar; picker Reset uses sidebar.reset. Go
  connects these original keys, retaining root+epoch/menu owner safeguards,
  workspace action identity, popup Label, Reset public Label and data behavior.

No new translation keys or YAML edits. No generic label replacement. No protocol,
filesystem, worker, native toolkit, destructive-confirmation or appearance-control
changes. Existing source Chinese values `workflow.build` and
`message_scroller.jump_to_latest` are English; routing through these keys is fixed,
but those source translation gaps intentionally remain. Worktrees, New task here,
Open in file manager, Thought process, Paste image, and other previous crosswalk
rows without an exact semantic match stay pending.

## Evidence

Actual Chinese mounted RED before production changes:
`p20-chat-red-1901.log` — composer display and workspace actions fail. The
initial search assertion in that run used Tester.HasText, which does not inspect
native input placeholder paint; that failure is not valid regression evidence.
Search is checked separately by comparing only empty-input text-region pixels
across language reloads, with a production-hook-negative control. Tests assert localized native choices emit original
SetAccess/SetWorkflow values, switching English reloads correctly, clipboard is
exact workspace path, popup/Reset identity remains callable and reset retains its
workspace. Existing source English and sticky owner tests run alongside these.

CPU mounted tests are behavioral evidence, not physical macOS captures, screen
reader language acceptance, complete international typography or pixel parity.

Search negative-control log `p20-search-paint-negative-1904.log` failed with zero
changed input-region pixels when only the placeholder translation hook was
removed. Restored hook passed (`p20-search-paint-green-1903.log`, 5.850s).
Final repeated mounted checks, complete assets/UI race and vet logs are
`p20-chat-final-{focused,full,vet}-1904.log` under the canonical toolchains log
root. Read their terminal results; existence of a log is not success evidence.

## Remaining accessibility localization boundary

The preserved English public Label is also the native accessible name. Thus
painted text and tooltips can be localized while some screen-reader names remain
English. This is a remaining accessibility localization gap, not a requirement
to retain English permanently for tests. A separate source-first audit should
separate stable element/test identity from translated accessibility semantics
before changing those labels and their selector contracts. This batch does not
claim localized screen-reader acceptance.

Final exact frozen source gate: repeated mounted race (2x) 10.839s, full UI race
460.069s, assets race 2.332s, vet exit 0. Four frozen file hashes and patch were
verified unchanged after the gates. This section is a separate doc-only additive.
