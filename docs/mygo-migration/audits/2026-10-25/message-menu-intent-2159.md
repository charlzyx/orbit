# C08 message-menu opening snapshot (increment after owner fix)

Requires the separately reviewed owner-keyed menu host patch `dbadc9ae`.
Only menu payload lifetime changes; no new items, toolkit private API, resource
loader, raw-message mutation, or rendering/selection policy changes.

## Source and actual mismatch

Pinned Rust 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac:
`transcript_view.rs:855–864` captures the right-clicked ChatMessage::text into
TextMenu.message_text; :931–941 copies that captured value. :913–928 captures
selected text for Copy Selection. It does not re-resolve the message at native
menu selection time.

The real App test opened an assistant streaming message's menu, adopted appended
text under the same owner, moved the pointer away and selected Copy Message.
Go copied the appended current text rather than the opening snapshot. Preserved
RED: `/workspace/shared/orbit-menu-intent-sameowner-red-2156.jsonl` (one named test,
actual failure; no scope/owner change). That same actual path is now the passing
`TestMessageMenuSameOwnerStreamUsesOpeningSnapshot` regression.

## Public two-phase implementation

MyGo public ContextMenu contract (menu.go:13–17,77–98) invokes the builder on
open and again for a choice. Two fixed items are constructed first; a matching
Chosen branch consumes the captured payload. Only the no-choice opening branch
replaces the local intent and sets disabled presentation. A disabled native
item cannot invoke copying; guards also reject an empty captured payload.
No dependence on private collect flags, pointer location during selection,
current hovered text during selection, or changing item indices.

There is only one intent per live owner-keyed menu host, containing immutable
message/selected strings and the host ID. Opening replaces it, choosing clears
it, and retiring the owner host retires its Local state. Cancel leaves at most
that one intent until reopen or host retirement; it does not write clipboard.
The existing App Copy guard still requires the current actual live owner.

## Explicit remaining boundary

**Blank-space targeting remains unfixed.** Its independent RED is retained in
`orbit-message-menu-blank-red-2150.jsonl`. The app still supplies last row hover;
Rust uses actual text blocks or <24 DIP gutter. SelectedTexts exposes IDs/text
and rune ranges, not all Markdown block geometry. This patch neither guesses
row geometry nor adds a second renderer. It corrects the lifetime of a supplied
opening target, not whether that initial target was hit correctly.

## Verification

Public-only prototype: 2 exact names x3, 6 run/pass, no skip/fail; opening versus
choice counters, same-owner edits, pointer exit, disabled selection, cancel and
reopen (`orbit-menu-intent-prototype-2156.jsonl` and matching hitset).

Final production tests: 19 exact names x3, 57 run/pass, no skip/fail, transcriptui
1.333s / UI 7.965s (`orbit-menu-intent-final-2158.jsonl`, hitset
`orbit-menu-intent-final-hitset-2159.json`). Includes original owner/actual-service
matrix, fresh menus, cross-message/virtualized selection and fold selection.
Final three SelectionMenu tests x3 (9 run/pass, no skip/fail) also cover disabled
empty-message with active selection (`orbit-menu-intent-disabled-final-2159.jsonl`
and hitset). Final UI/transcriptui vet passed (`orbit-menu-intent-vet-2159.log`). No full UI duplicate,
new hosted run, native OS/pixel, or complete C08 acceptance is claimed.
