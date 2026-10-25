# S12 sidebar sticky workspace headers — bounded source batch

Base: frozen S13 workspace actions plus test-only action routing additive (local baseline `5100cdf`). No agent/lease changes; no toolkit-core patch.

## Source contract
Pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`:
- `app/sidebar.rs:467–529`: pin the last workspace header at/above the list's first visible row, only for expanded groups. The active workspace has no special privilege: an expanded background workspace also pins. Flat mode and collapsed groups do not pin; a real header exactly at the viewport top needs no copy.
- `app/sidebar.rs:511–523`: use the next measured header's Y and height relative to the viewport to push the outgoing header away. Unmeasured far-away items cannot trigger push-off.
- `app/view.rs:282–311,662–708`: render the real list and pinned native layer with common row content; the pinned owner alone gets the workspace popup. Clip the outgoing layer at the list viewport.

## Implementation
The original List, row keys, row count, selection and ScrollTo/ScrollIntoView stay in place. A native absolute List child paints the pinned header; it does **not** use List.Header, whose header-space compensation changes logical start anchors. Keeping the layer under the real scrolling ancestor preserves wheel delivery.

Header/viewport handles provide current native geometry. Row order/collapse, font/size/density, sidebar width and window height changes discard stale bindings and request a settled frame. Removed paths cannot retain actionable headers. While pinned, the ordinary underlying header is noninteractive presentation; only the native pinned copy owns click/context-menu/menu/popup handlers. Both copies retain the same logical workspace identity, with S13's epoch/path/entry menu guard unchanged. Worktrees keeps the source's current-repository semantics.

The existing six-DIP sidebar resize grip occupied the same right-edge hit region as the native List scrollbar. The public ScrollbarInsets moves both bar painting and hit testing left of the grip. Header actions reserve the complete inset scrollbar hit gutter (public ScrollbarWidth plus native track padding), not just the visible thumb, so an ellipsis edge click cannot turn into a scrollbar click. The original resize grip location/behavior is unchanged.

## Evidence
Logs under `/workspace/shared/orbit-toolchains-continued`:
- `sidebar-sticky-red-1802.log`: deterministic mounted baseline failure after actual vertical scrolling — expanded workspace header disappeared.
- `sidebar-sticky-first-1803.log`: first sticky regression green under race.
- `sidebar-scrollbar-1806.log`: real hover/frame/press/move/release still failed before moving the bar away from the resize grip.
- `sidebar-sticky-gutter-1810.log`: sticky behavior race twice PASS 23.126s (before final complete track-gutter reservation).
- `sidebar-sticky-final-focused-1811.log`: final focused race PASS 31.593s, including existing SourceSidebar navigation/filter/order/focus and S13 actions.
- Final full UI race/vet status is supplied with the frozen patch manifest, not inferred from earlier focused passes.

Mounted assertions cover expanded active/nonactive groups, exact native ScrollTo(42, Start), sticky wheel delivery, incoming-group push-off, collapsed groups retaining pinned sessions without sticky headers, flat mode, stale menu and held-press target changes, font/resize/dataset invalidation, scrollbar and resize grip dragging, active-session focus/Down/Escape, one popup while the real header is partially visible, and the rightmost pixel of the pinned action button avoiding the scrollbar track.

Independent source review found no blocker in the bounded sticky contract. It is not a full platform Tab-order or accessibility acceptance test. No hosted macOS visual recapture or pixel-parity claim; native Mac checks remain separate. No filesystem business operation was added: tests exercise only existing resize persistence in disposable preference homes.
