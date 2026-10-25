# S05 Appearance: one keyboard stop per source group

Pinned Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
`crates/orbit-pi/src/app/settings.rs:6143–6225` creates one focusable
appearance group. Plain child div segments select on pointer click; the group
prevents mouse-down default focus transfer. Left/Up and Right/Down cycle modes;
Home/End choose Light/System. Tab/ShiftTab are ordinary window traversal.

The Go implementation already handled the selection keys on its focusable
root, but each native ButtonBase child introduced an extra focus stop. A real
mounted regression fails before correction: Tab from `Appearance mode` does
not reach the following `Theme selector`. This was a tab-order difference,
not an inescapable keyboard trap.

## Bounded correction

Only `sourceSettingsSegments` (currently called solely by Appearance mode)
changes: children become styled public Row elements with the same Center,
Shrink(0), size, typography, hover and Clicked behavior; KeepFocus preserves the
current focus during pointer selection. The root's Focusable and existing
selection shortcuts are unchanged. There is no global Tab interception,
roving-group substitution, toolkit patch or change to other Settings controls.

Explicit `Role(ui.RoleButton)` preserves accessible button labels. Upstream
MyGo `ui/access.go:435–437` generates ActionPress for Clicked/flagClickable
independently of Focusable; `access.go:425–432` suppresses actions when disabled.
Thus removing the Tab stop does not remove the accessible press operation.
The public tester does not expose the AX tree: this is source-level evidence,
not a claim that VoiceOver/OS accessibility execution was tested.

## Regression scope

* Mounted full Appearance page: forward Tab skips all children to Theme;
  ShiftTab returns directly to group; Left/Right/Up/Down/Home/End select the
  source values while retaining group focus.
* Pointer selection changes mode while keeping focus in Search settings.
* Disabled group has no Tab stop and pointer/key input cannot invoke apply;
  re-enabling restores one group stop and exactly one selection callback.
* Existing non-chat mounted Settings navigation and dropdown Escape tests
  remain included in the focused gate.

Synthetic preference home only. No credentials, installation/setup, network,
real retention data or publication. No full-workbench keyboard parity claim.

The existing persistence test previously clicked Light then sent Right without
keyboard entry. That assumed pointer focus transfer, contradicting the pinned
Rust prevent_default contract. It now retains the pointer persistence assertion,
then Tabs from Search settings to the actual group before asserting Right's
persisted Dark value. The assertion is not removed or weakened. The old test
failure is retained in `s05-appearance-full-race.log`; final gates use v2 logs.
