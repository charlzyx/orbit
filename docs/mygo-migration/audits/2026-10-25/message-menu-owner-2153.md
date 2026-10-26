# Transcript message menu: owner-scoped replay

Source-first audit against Rust 299a32bc87fe3d4851bdb902d9c03dcf9bf99cac,
Go wave11 candidate. No claim of native OS/pixel or complete C08 acceptance.

## What is already present (useful negative result)

Rust transcript_view.rs:905–967 has Copy Selection and Copy Message, not a
universal quote/edit-message menu. Go internal/transcriptui/selection.go already
implements those items. Rust ChatMessage::text (transcript.rs:107–115) and Go
Row.Text (model.go:125–133) both join nonempty step answers with two newlines.
Normal mounted first/second/first-message right-click copying passes. No new
quote/edit feature is justified by this audit.

## Actual ownership defect and narrow correction

MyGo's public ContextMenu contract (ui/menu.go:13–17,77–98) runs its builder at
open and again in the current frame after selection. The Go transcript's fixed
selectable/menu host key let a menu opened for an old task replay into another
current task. Actual mounted runtime, runtime A→B→A and session-switch tests
copied NEW_MESSAGE from the new task via the old menu. This is retargeting, not
an observation of copying the old text.

Rust transcript.rs:1082,1150 clears text selection/menu on load/reset;
transcript_view.rs:465–470 clears the menu. The candidate uses the existing
actual owner generation (runtime/session/path/workspace/live task/service),
including roundtrips and Close, as the selectable/menu host identity. Old menu
IDs can no longer replay into the new host. The common transcript Copy callback
also rejects a no-longer-live owner. Reset discards the old hover target. It does
not change toolkit code, menu items, copy formatting, or normal pointer behavior.

## Explicitly separate remaining gaps

- Go still caches last row hover. A separate real mounted RED right-clicked
  blank transcript space 100 DIP above a previously hovered message and copied
  the old hover target. Rust message_at (:475–490) requires a text hit or a
  nearest-block gutter distance below 24 DIP. A whole row is not that contract.
- Rust captures TextMenu.message_text at right-click (:855–864), then copies that
  captured string (:931–941). Current Go menu builder reads current-frame text.
  Same-owner content updates and pointer motion require an independent immutable
  menu-intent solution. Clearing hover every frame is NOT included: it could
  disable valid menu selection after moving the pointer to the native menu.

## Evidence

Preserved RED `/workspace/shared/orbit-message-menu-audit-2148.jsonl`: normal
current-hit pass; three old-owner cases fail by copying NEW_MESSAGE.
Blank-hit RED `/workspace/shared/orbit-message-menu-blank-red-2150.jsonl`.
An intermediate frame-cleared-hover experiment passed a subset but was withdrawn
before freezing; `/workspace/shared/orbit-message-menu-owner-green-2151.jsonl`
is not acceptance for the final owner-only patch.

Final owner-only source: `/workspace/shared/orbit-message-menu-owner-final-2152.jsonl`
race x3, UI 6.800s, exact 10 top/subtest names / 30 run and pass / no skip or fail.
Includes normal targets, runtime/ABA/session/workspace/Close, pointer moving off
message into the native menu, actual cross-message selection and folding selection.
Hitset: `/workspace/shared/orbit-message-menu-owner-hitset-2153.json`.
Actual service/task/Close matrix passed race x3 (4 names/12 run/pass), UI 2.133s.
Final all-menu matrix additionally verifies a freshly opened current-owner menu
still works after the rejected old selection: race x3, exact 12 names / 36 run
and pass / zero skip or fail, UI 6.215s. Log: `orbit-message-menu-owner-final-v2-2154.jsonl`;
hitset: `orbit-message-menu-owner-final-hitset-2154.json`. Targeted UI vet passed
(`orbit-message-menu-owner-vet-final-2154.log`). No duplicate full UI or cross-build.
