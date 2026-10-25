# S05 non-chat/files keyboard audit (bounded, read-only production)

Reference: Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
Go test snapshot: isolated copy of wave4 frozen source; every entry in
`/workspace/shared/orbit-wave4-source-freeze-1846.json` matches byte-for-byte.
Worktree: `/workspace/scratch/6a74e5402e48/orbit-focus-pages-1852`.
No production or toolkit changes; only two new audit test files.

## Source findings

* `app/view.rs:3775–3818` delegates next/previous to the window and blocks
  keyboard-modal surfaces. It does not specify a separate Settings/Usage/
  Worktrees surface ring. Extending the chat/files custom ring to these pages
  without a demonstrated defect would not follow this contract.
* `app/settings.rs:6143–6225`: appearance is one focusable segmented radio
  group, with arrows and Home/End selecting modes. Go already implements those
  keys in `sourceSettingsSegments`. Go additionally has focusable native segment
  buttons: that is a remaining tab-order granularity difference, not a proven
  trap or unreachable setting. No claim of identical global tab order.
* `app/settings.rs:7042–7063`: picker arrows move cursor and Enter chooses;
  `app/session.rs:498–507` Escape dismisses dropdown before leaving Settings.
  Go's existing input-local handling and native popover enforce these layers.
* `usage/filters.rs:180–210`: Enter dismisses a multi-select menu; selection is
  via controls, not an Enter toggle of the hover-highlighted row. Do not change
  this behavior as a supposed accessibility fix. Go native buttons can instead
  be reached by Tab and activated by Space.
* `app/worktrees.rs:2652–2664` explicitly focuses name in quick-create;
  `app/session.rs:444–449` cancels a worktree modal before the underlying page
  or running agent. Go uses native modal/text-input focus and cancellation.

## Mounted checks

New tests use actual `ui.NewTester` keyboard dispatch and `Focused` owner
assertions, not action registry invocation:

1. Settings search → Tab Appearance → Enter → Tab appearance group → End
   selects System → Escape returns to actual Message owner.
2. Settings language selector reached by Tab → Enter focuses filter → empty
   result query + Enter does not select a phantom item → Escape dismisses only
   dropdown → second Escape leaves Settings and focuses Message.
3. Usage model filter focuses search → Tab Select all → Space changes the
   actual filter → Escape dismisses → Tab Back to task → Enter leaves Usage.
4. Worktrees filter → Tab New worktree → Enter focuses Directory name → Tab
   branch-mode checkbox → Tab Branch name → Escape cancels without starting
   creation or setup. Synthetic root only; no Git/create/delete/setup action.

Initial exploratory failures were invalid test assumptions (wrong appearance
label, missing synthetic Root disabled create, and overlooking the intervening
branch-mode checkbox), NOT product RED evidence. Corrected fixtures preserve
those actual controls. No production defect has been demonstrated in this pass.

## Scope / next decision

No minimum production correction is justified by these checks, so none is
proposed. Keep S05 scope limited to its existing chat/files surface change.
These checks do not prove all controls, all Settings sections, tiny windows,
OS accessibility menus, native Mac dispatch, or full Rust pixel/keyboard parity.
The optional test-only patch is separate from the already frozen wave4 gate.

## Final gate result

`source /workspace/shared/orbit-toolchains-continued/env.sh; scripts/go-native.sh test -p 2 -race ./internal/ui ./internal/worktreeui -run '^TestAudit(Settings|Usage|Worktree)' -count=1 -v`

PASS: UI 14.077s; Worktree UI 1.345s. Full log:
`/workspace/shared/orbit-toolchains-continued/logs/s05-nonchat-audit-v3.log`.
This is a focused audit gate only; no full-suite or native Mac execution claim.
Test-only patch SHA256:
`21934323ba7883680574c59727db7f7e8e57332662a1e4cb07e5c01a55231990`.
