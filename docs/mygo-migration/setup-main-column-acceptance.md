# P03 setup belongs to the main column

Independent additive on frozen read-only P03 (`d54db58` local baseline).
Pinned Rust `app/view.rs:532` mounts the session sidebar whenever Settings is not
open. Its dependency/setup conditional at `795–800` selects only the main child.
The original Go setup gate returned before the sidebar, incorrectly hiding task
navigation and resize while dependencies were missing.

The additive moves that conditional around `a.main` only, preserving the existing
root/sidebar ancestry and resize handle. Settings still owns the full surface;
collapsed-sidebar preference remains respected. Setup scroll has MinWidth(0) so
it fits the remaining column instead of forcing sidebar overflow.

A real mounted RED failed before correction:
`/tmp/orbit-setup-main-column-red.log`, "required-dependency gate incorrectly
removed source sidebar". Corrected full onboarding focused race three times
passed 18.440s (`/tmp/orbit-setup-main-column-focused.log`), including unchanged
input/drop/Settings behavior, sidebar geometry/resize presence, visible New Task,
and formerly focused sidebar CmdShiftB→setup→Down/Enter selecting the exact session
without exposing the hidden composer. Read-only dependency refresh and task
ownership guards are unchanged; no installer code is included.

CPU fixture `/tmp/orbit-setup-main-column.png` was inspected at 1200×800: sidebar
remains visible and dependency cards, facts, and refresh footer occupy the main
column. This is not physical macOS acceptance or full pixel parity.
