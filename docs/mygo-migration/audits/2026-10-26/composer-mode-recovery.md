# Reconstructed finite composer mode pickers

Base: published `cee007d9309815489bfaa11529fa3047a2b612f8`. The previous
executor's unuploaded UI files disappeared. These are reconstructed bytes,
not the old Wave11 hash; old gates are historical only. Final validation is
recorded separately by the recovery integrator.

## Byte-exact recovered objects (unchanged)

- `internal/agent/workflow.go`: `695d0bcde70e88e5c139f534dff15b8b8b365bde`
- `internal/agent/workflow_capture_test.go`: `0b5d916d3468a616ae127403078645f51fe66c77`
- `internal/ui/actions.go`: `7ae07a9b5350014bed03655b3b024fbfefbd552c`

These match partial recovery commit `b41d2c2058c0316194a5af9f3672b0c5a959c988`.

## Reconstructed scope and source

Pinned Rust `299a32bc`: composer_ops.rs 686–901, pickers.rs palette toggles,
view.rs Access/Workflow popups and visible noncompact composer. Finite values,
current marker and source description catalog, keyboard wrapping/Enter/Escape,
composer focus, competing menus, hidden same-owner intent, and opaque task
capture are restored with public MyGo APIs. No free-text permission prompt,
new permissions, toolkit changes or global page navigation.

Access persists only through the existing Controller/Manager. Neither picker
optimistically assigns local mode: confirmed Snapshot is the sole projection.
Legitimate empty-ID Workflow selection is retained and persists after hydration.
The exact recovered Manager accepts Workflow into memory before trying its
file write; failure reports an error but does not roll back that memory value.
Rust composer_ops.rs `set_workflow_mode` likewise assigns mode before persist.
This patch does not change that existing contract or claim the file was saved.

Full Review retains an invisible composer for native undo. A reconstructed
mounted regression found that an overlay could escape this invisible subtree;
explicit visible-composer gating now retains intent without mounting the popup.
There is no 200ms/time debounce: immediate Escape/outside dismissal and fresh
trigger reopening have actual mounted tests. Same-trigger toggle is covered.

## Evidence boundaries

Recovery logs retain compile mistakes and the real full-Review RED. A newly
written Workflow-write-failure test initially expected rollback; it failed
against unchanged recovered Manager behavior. The corrected test explicitly
checks pre-snapshot old UI, diagnostic plus confirmed in-memory mode, and failed
file write, rather than claiming persistence success. That earlier failure log
is preserved as `/tmp/orbit-recover-modes-async.jsonl`.

Fixtures are disposable shell/test-process JSONL children and temporary Stores;
no credentials, inference, real settings or installers. New focused JSON tests
must be checked by exact name and run/pass counts. They are not full combined
acceptance, actual native macOS input, screenshot or pixel parity.

## Recovery revision after independent review

The first reconstruction `0eac25e7` was not accepted. A real mounted regression
(`orbit-recover-modes-hover-red.jsonl`) confirmed stationary pointer hover reset
keyboard Down on subsequent frames. The revised menu tracks the previous hovered
row and changes keyboard cursor only on entering a different row (or returning
from outside), not every frame. Up/Down/Enter while pointer remains stationary
are exercised with actual input and extra frames.

Pinned Rust view.rs:158 sets `MODE_POPUP_W = 300` raw px; tokens.rs:255–258 uses
scaled 14px Default and 12px Small text. This revision uses those exact values
rather than the reconstructed 260px / FontSize-minus-one. Popover public label
uses existing localized `command_palette.choose_access` or
`command_palette.choose_workflow_mode`; it does not add a visible heading absent
from the source. The earlier lost candidate's reported 320px is not treated as
source truth. These are new reviewed bytes, never the lost hash.
