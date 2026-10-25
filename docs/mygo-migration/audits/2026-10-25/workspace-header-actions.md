# S13 workspace header actions and marks — source-derived candidate

Baseline: `ce4d49a`. Source: pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

## Source contract and implementation
- `app/sidebar.rs:1546–1602,2362–2398`: Copy path / Worktrees / Icon & color / Remove from sidebar. Removal is an Orbit project-list edit, **not disk deletion or runtime stop**; Rust does not add a destructive-delete confirmation. Existing file/session deletion confirmations are unchanged.
- `app.rs:2158–2170`: remove the selected project and forget its mark/stamp; re-adding the directory is supported. The Go store preserves other projects and sort/group/archive preferences and validates the selected path, addition stamp and captured in-process list revision under the shared file lock.
- `workspace_mark.rs:31–145`: curated 27 icon stems, 8 semantic tints; unknown manually edited values fall back safely; explicit icon wins over automatic status. No arbitrary asset path is accepted by the writer.
- `app/view.rs:220–261`: workspace status priority current / running / all archived / idle.
- `app/sidebar.rs:963–1070`: flat-list second line shows workspace name and mark using that session's own active/running/archived status. Grouped rows retain their existing summary behavior.
- `app/sidebar.rs:1665–1814`: six-column icon grid, tint row, Reset, immediate persisted selection. Candidate uses native anchored PopoverBase, not a screenshot overlay. Right-click and a clickable ellipsis menu both reach the actions.

## Ownership / asynchronous boundary
Native menu element IDs include workspace epoch, exact path, entry identity and addition stamp. Both SetSnapshot and Controller.changeWorkspaceServices invalidate the epoch on real workspace changes, including A→B→A before render. Same-snapshot menus remain usable. This does not change the source's duplicate-basename group merging rule; the first listed project still owns that group, and its exact path owns the menu.

Preferences canonicalization and disk writes run through App.Async. Only one workspace edit is admitted at a time; redundant clicks cannot enqueue concurrent edits. Store.EditWorkspace changes one matching project, not a stale whole-list snapshot. All Store instances share a workspaces.json lock and revision; every successful Write/Update (including Save/Add/Remove) advances it. Snapshot and revision are captured together. Directory aliases and case variants resolve by canonical path or directory SameFile identity. Any intervening list write makes an old edit fail with an explicit retry message, including same-second remove/re-add with identical AddedAt. Completion refreshes only the original UI list owner; if it was replaced while the write ran, the newer UI is untouched. Revision conflicts refresh the original owner for a fresh retry. Other write/read failures retain the visible project list and report the error, rather than replacing it with an empty failed-read snapshot. The preferences edit itself remains accepted for the captured project if still present; navigating away is not cancellation of an already accepted write.

A popup closes when its target vanishes, the workspace changes, its header stops rendering (hidden sidebar/settings/flat grouping), or Escape/outside dismissal occurs. Hidden popups cannot leave the app's modal shortcut guard stuck. No manager, session lease, child process or live user directory is modified by this batch.

## Evidence
- Baseline mounted red: Copy path absent (`workspace-menu-red-1737.log`).
- Mounted menu/picker regression race twice: PASS 22.218s (`workspace-menu-final-focused-1743.log`), before final services-entry hook and flat-row marker addition.
- Preference targeted edit/invalid marker/stale stamp/keep-session/re-add race: PASS 1.294s (`workspace-mark-prefs-1740.log`).
- Final exact-source owner-focused and full-package race/vet logs are recorded with the delivered manifest; do not infer final status from earlier passes.

Tests use only disposable temporary project/store files. Remaining limits: no hosted macOS popover/native recapture yet, no pixel-parity claim, no new source-wide sidebar layout/sticky rewrite. The revision is process-local and excluded from JSON; external processes or arbitrary direct filesystem writers do not participate. The shared lock serializes workspace-file operations across stores; the revision registry retains directory identities for the process lifetime. Existing on-disk AddedAt remains seconds and is not used as a unique lifetime token. Case-insensitive variant coverage skips on Linux and remains pending on macOS.

Revision evidence: disabling only the comparison deterministically allowed a stale snapshot to edit a same-stamp re-added entry (`workspace-revision-aba-red-1749.log`); restored shared-store revision rejects it and permits a fresh retry. Final evidence includes duplicate Store instances, symlinked RootDir, direct Write, Add/Save/Remove, stale completion and malformed-store failure.

## Explicit action routing checks (test-only additive)
Rust `sidebar.rs:1575–1584` opens Worktrees without passing the clicked header's cwd; `app/worktrees.rs:42–58,85–94` obtains the **current** workspace. Thus right-clicking inactive B while A is active intentionally opens A's worktrees; this candidate does not silently switch to B. Rust `app/session.rs:619–635` starts a new task even in the same workspace (reusing an idle child); Go's menu dispatches SetWorkspace → Controller.SetWorkspace → Agent.NewSession without a same-path early return.

Mounted actual-controller tests use a credential-free JSONL child: inactive B's Worktrees action keeps A as the controller/worktree workspace; New task here on both A and B must produce a different session ID and align the active services with the requested directory. These are behavioral assertions, not menu-item presence checks. All fixtures explicitly register both A and B in the private project list. No production action change was necessary.

Final frozen production gate: `workspace-menu-frozen-full-race-1751.log` PASS UI 393.644s / preferences 1.320s; `workspace-menu-frozen-vet-1751.log` PASS. Independent revision guard race 3x PASS in `/tmp/orbit-workspace-revision-independent.log`. Action test-only additive evidence: `workspace-menu-controller-final2-1758.log` (final status delivered with additive manifest).
