# Explorer context-menu actions — unreleased source increment

Reference: Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `crates/orbit-pi/src/explorer/panel.rs:1238–1406` (directory toggle, create/rename/delete, default-app/reveal/copy-path actions). This increment is not in the shipped `.4` package.

## Changes

- Directory menus expose Expand/Collapse before create actions; both file and directory menus expose Open in default app.
- Default-app dispatch captures the selected absolute path, schedules the launcher outside the view handler and returns errors through the existing owner-thread completion queue. The native OS adapter may marshal to its main thread; this is not a guarantee of a nonblocking OS launcher.
- Workspace changes invalidate native menu identity, including A→B→A before the next rendered frame, so an old choice cannot act on a same-named row in a newer workspace.
- Existing create/rename flows and deletion confirmation remain unchanged. No test opens an external application or accesses user files.

## Evidence

`internal/ui/explorer_menu_test.go` mounts the real MyGo context-menu path (`Menu`/`ChooseMenuItem`), covering repeated toggles, cancel/dismiss, exact file/folder targets, workspace replacement and round trips, and deletion cancellation with temporary fixtures. The injected launcher test checks caller scheduling and error delivery, not native OS latency.

Baseline negative tests failed for missing Expand/default-app entries. A separate round-trip regression failed before the workspace menu epoch was added. Final focused `-race` suite passed (20.998 s) and `go-native.sh vet ./internal/ui` passed on the final source. Full integrated gates and hosted macOS behavior are tracked separately; this does not establish complete Explorer or visual parity. Inline create/rename parity remains a separate gap.
