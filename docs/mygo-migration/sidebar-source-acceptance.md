# Sidebar organization and navigation: source-derived acceptance

Reference: unchanged Rust baseline `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
Implementation: `internal/ui/sidebar.go`, the mounted sidebar in `app.go`, and
its command routing in `actions.go`. This is a focused S09/S12 behavioral
reconstruction, not a claim that every sidebar feature or native platform is
accepted.

## Implemented contract

- `app.rs::sidebar_sessions` and `sidebar.rs::sessions_with_placeholder`: only
  explicitly saved projects contribute disk sessions. A started active task
  may be represented before the first store scan, but an unsent draft is not
  listed. Its live title supersedes a stale disk title.
- `session_lead_rank`: active session, pins, running sessions, then newest
  ordinary sessions. Equal-rank recency order is stable. A background run's
  writes cannot displace the active row.
- `sidebar_groups`, `sort_workspace_groups`: grouped project identity uses
  basename, the first saved path wins duplicate labels, empty saved projects
  remain present, and unlisted projects are not invented from session files.
  Manual, last-updated, alphabetical ascending/descending, date-added and
  session-count sorts are implemented. Last-updated floats the active group;
  date-added reads persisted timestamps, not reversed list position.
- `is_workspace_group_collapsed`, `emit_workspace_group`: the current workspace
  starts open and other workspaces closed. Collapsing keeps active, pinned and
  running tasks reachable. Expanded groups start with exactly three sessions,
  reveal up to three per Show more, offer a right-edge Show less after a partial
  expansion and retain Show less after the entire history is shown.
- `build_flat_sidebar_rows`, `filter_archived`: flat mode contains sessions only,
  without workspace headers or overflow rows. Hide/show/only archive filters
  run before projection, counting and navigation.
- `navigable_sessions`, `open_session_slot`, `cycle_session`: primary+1…9 and
  Ctrl+Tab/Ctrl+Shift+Tab use displayed sessions rather than raw snapshot order.
  Invalid slots, empty lists and one-session cycles are no-ops. When the current
  task is absent, cycling starts at the first/last displayed task.
- `on_focus_sessions`, `on_sidebar_confirm`, `on_sidebar_close`: focusing reveals
  the sidebar and selects the active row; native list arrows/Home/End navigate
  headers, sessions and overflow. Enter uses the same action as clicking.
  A session destination returns focus to the composer. Focused-sidebar Escape
  leaves navigation rather than aborting Pi; feature-page Escape and genuine
  composer Escape retain their existing semantics.
- Sort/group/archive choices still persist through the existing workspace
  store. Disclosure and extra-visible-row state remain transient, as in the
  reference.

The projection carries complete session rows into both rendering and navigation.
It deliberately does not reproduce the reference's filtered-row indices being
looked up in an unfiltered session array: that could open a different task from
the one displayed when archived sessions precede it.

## Executable evidence

`internal/ui/sidebar_source_test.go` covers the contract with intentionally
out-of-display-order fixtures: an old active task, newer ordinary tasks, a pin,
a background run, an archived task and an unlisted project. Mounted MyGo tests
assert actual painted order, hidden rows, selected cursor, focus and dispatched
`SwitchSession` path, not just helper return values. They exercise primary-number
and cycle shortcuts, group click/keyboard disclosure, Show more/less, native
project-menu persistence and reloading a fresh preference store. Additional
cases cover filtered-empty navigation, stale titles and Unicode-safe placeholders.

These tests use isolated temporary preference homes and CPU-mounted MyGo views.
They do not launch Pi, access user stores, exercise a real OS desktop or establish
pixel-identical rendering. Existing page-Escape tests are retained unchanged.

## Not established by this evidence

Sticky group headers, workspace icon/tint customization, drag/manual reordering,
complete session/workspace context menus, automatic active-row scroll on every
external refresh, real-OS menu/focus behavior, Pi switching lifecycle and measured
same-state Rust/Go visual parity remain separate acceptance work. The source's
legacy date-added filesystem-timestamp fallback is not added here.
