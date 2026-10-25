# Native Skills and Plugins: source-first acceptance

This report concerns the Go `internal/resourceui` settings surfaces only. It is
not a claim that the complete migration or pixel-identical rendering has passed.

## Reference contract

- Orbit reference: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
- MyGo native module: `9c69cf829ece196c5b3a3601f1c7bbe70526376f`.
- Behavioral sources: `crates/orbit-pi/src/app/skills_ui.rs`,
  `crates/orbit-pi/src/skills.rs`, `crates/orbit-pi/src/app/settings.rs`
  (`refresh_plugins`, `plugin_install`, `plugin_run`, `apply_plugin_action`),
  `crates/orbit-pi/src/app/plugin_update_ui.rs`, and the plugin update notice in
  `crates/orbit-pi/src/app/toast_ui.rs`.
- Geometry and typography: `crates/orbit-pi/src/theme/tokens.rs`, the settings
  body/navigation in `settings.rs`, and both actual marketing images
  `marketing/public/screens/skills.png` and `plugins.png` (visually inspected).

The checklist was derived from those handlers before marking individual cases.
“Pass” below means the specified fixture test passed, not unrestricted OS or
external package-manager acceptance.

## Checklist and evidence

| ID | Source behavior and expected outcome | Go acceptance evidence | Result |
|---|---|---|---|
| R01 | Discover actual skills; list project/global/user groups; filter name and description without changing the selected detail | `TestSourceSkillsDiscoveryBodySelectionAndFilter`; real temporary `.pi/skills`, `.agents/skills`, settings files | Pass |
| R02 | Select by file path, choose first discovered skill initially, retain existing selection after refresh, choose remaining skill after deletion | `TestSourceSkillsDiscoveryBodySelectionAndFilter`, `TestNativeSkillTogglePersistsAndDeletionRequiresSecondAction` | Pass |
| R03 | Render frontmatter-free Markdown; expose invalid description and manual-only state, invocation, directory, bytes and modification age | `TestSourceSkillsDiscoveryBodySelectionAndFilter`, captures; `Skill.Body/Valid/InvokeCommand` backend tests | Pass for native fixture; exact Markdown text raster equivalence not established |
| R04 | Toggle writes real pi overrides, re-reads disk, does not optimistically claim success, preserves malformed/unrelated settings | `TestNativeSkillTogglePersistsAndDeletionRequiresSecondAction`, `TestSkillsReadWriteErrorsRemainVisibleAndNoOptimism`; backend resource parity tests | Pass |
| R05 | Open full SKILL.md path with default app, reveal that file, copy displayed full file path, not its directory | `TestNativeSkillActionsCallbacksAndCancel` verifies exact callback arguments; host adapter inspected separately | Callback contract passes; actual external OS opener/clipboard remains platform acceptance |
| R06 | Delete first asks inline; Cancel/Escape do not write; confirm deletes only the captured skill; changing selection dismisses stale confirmation | `TestNativeSkillActionsCallbacksAndCancel`, `TestNativeDeletedSelectionDoesNotReuseStaleTarget`, real disposable file removal | Pass |
| R07 | Partial deletion is not misreported as success or retried: actual deleted row disappears, settings-cleanup failure remains explicit | independent `TestAuditNativeDeletedSkillCleanupFailureShowsActualDiskState` | Pass |
| R08 | Installed plugins come from scoped settings; source/version search; real resolved version, path, installed/missing badges | `TestPluginScopeSearchVersionsAndRemovalCancel`; local package.json fixtures | Pass |
| R09 | Choose Global/Project installation; source field Enter is not a commit; only Install button starts action | `TestNativePluginEnterDoesNotCommitButTabButtonDoes`; source `app.rs` explicitly leaves Enter unhandled | Pass |
| R10 | Run Pi off UI thread; project install/remove append `-l -a`; update has no scope flags; repeated clicks start one operation | `TestPluginInstallRealSettingsFakeProcessLifecycle`, `TestSkipAndResetRealPersistenceAndBatchPartialFailure`; backend exact-argv tests | Pass using explicitly named fake-process fixture, not a real installation |
| R11 | Remove requires inline confirmation; cancel leaves packages unchanged | `TestPluginScopeSearchVersionsAndRemovalCancel`; real settings file reread | Pass |
| R12 | Preserve progress/error state, show cancellation only after process reap, allow retry after failure/cancel | `TestPluginFailureCancellationAndRepeatedRetry`; actual POSIX fixture child process | Pass on Linux; Windows/macOS process behavior not claimed |
| R13 | A workspace change must reject stale completion without allowing overlapping old/new mutating workers | `TestWorkspaceGenerationAndCloseRejectStaleCompletion`, `TestNavigationWaitsForCanceledWorkerBeforeAllowingNewOperation`, independent actual Git cancellation test | Pass |
| R14 | Source Refresh rereads packages and checks metadata; launch check waits five seconds; no checks under `PI_OFFLINE` or CLI offline | `RefreshPlugins`, `Tick` deadline; `TestOfflineLaunchAndExplicitCheckNeverProbe`; actual local Git check below | Pass for offline/dispatch; automatic real registry calls deliberately not exercised |
| R15 | Check actual upstream Git metadata read-only, retain installed checkout/remote-tracking refs, suppress exactly the skipped revision | independent `TestAuditNativeLocalGitUpdateOfferSkipAndReset` creates a local bare remote and installed checkout | Pass; no network registry/credential use |
| R16 | Source bottom-right notice persists until explicit Update/Skip/Close; closing persists exact version skip; no automatic install | `TestNativeUpdateNoticeRequiresActionAndReservesGeometry`, `TestSkipAndResetRealPersistenceAndBatchPartialFailure` | Pass |
| R17 | Batch Update now runs sequentially, clears skips only for successes, retains failures and reports partial success | `TestSkipAndResetRealPersistenceAndBatchPartialFailure`; real skip JSON and explicit fake Pi argv log | Pass |
| R18 | Native Tab/input editing, selection arrows/Home/End, focused row scroll visibility, Escape cancellation; list must not consume search-input arrows | `TestNativeKeyboardSkillsAndFilterIsolation`, `TestKeyboardEndFocusesAndRevealsLastSkill`, `TestNativePluginEnterDoesNotCommitButTabButtonDoes` | Pass |
| R19 | Fixed 240-DIP host slot + 320-DIP skill master; pinned Plugins header; all minimum-window controls reachable; long errors stay below toolbar | `TestMinimumWindowSourceGeometryAndControlReachability`, `TestLongErrorAndBusyRemainReachable`; 800×560 captures | Pass for measured fixture layout |
| R20 | Render original embedded SVGs, source type/spacing/radius and host palette callbacks, without webview/Rust | `view.go`, `skills.go`, `plugins.go`, `updates.go`; 1470×922 and 800×560 CPU-native captures inspected | Native implementation verified; same-state GPUI pixel comparison remains open |

The per-plugin “Reset skip” control is an explicitly added recovery affordance,
not an interaction found in the Rust reference. The component has no production
data fixtures, no simulated install success, and no alternate agent runtime.

## Integration and ownership

`resourceui.New(prefs, workspace, Options)` owns retained data/state, not retained
MyGo elements. `State.View(c, "skills"|"plugins")` mounts outside the generic
settings scroll. `State.Overlay(c)` mounts once at the host root;
`NoticeHeight(c)` reserves the exact notice height below ordinary toasts.

The host calls `Tick`, `SetWorkspace`, `SetPiBinary`, and `Close`; category entry
calls `Refresh` or `RefreshPlugins`. `Options.Offline` supplements `PI_OFFLINE`.
Theme callbacks and open/reveal/copy callbacks run only on the UI thread.
Worker closures capture their service and immutable target. Only `Tick` applies
generation-checked results. Navigation keeps the operation guard until a canceled
worker has actually returned, even though its stale results are discarded.

`OnModified` asks the host for its existing debounced, busy-gated,
session-preserving Pi reload. It does not claim saved changes have already been
applied to a running session. Canceled/failed package processes are reread because
they may have partially modified configuration.

## Defects exposed and corrected during checking

1. Replacing Delete with inline confirmation orphaned native focus; Escape did
   not dismiss. Confirmation-scoped Escape handling now survives that transition.
2. A canceled child returned `signal: killed` rather than `context.Canceled`.
   Cancellation is normalized at the worker boundary after the child is reaped.
3. MyGo row cross-axis defaults vertically centered the entire page. Both pages
   now explicitly stretch in the host slot.
4. Calling `Children` twice appended a duplicate icon instead of replacing it.
   Icon-only controls now have their own single builder.
5. Alphabetically ordered backing data did not match grouped arrow navigation.
   Navigation now follows the displayed project/global/user order, focuses and
   scrolls the chosen row into view.
6. Workspace invalidation dropped Busy before a canceled process returned.
   A retired-generation guard now prevents a second package operation from
   overlapping it.
7. Initial visual assumptions used 12px compact button text and 20px page title.
   Source token inspection corrected these to 14px Medium/Large button labels
   and a 16px Plugins title, with the source 28/32px heights and 8px radii.

## Reproduction and boundaries

Use the Go 1.27.3 toolchain and exact module already pinned by the repository:

```sh
go test -race ./internal/resourceui -count=1
go vet ./internal/resourceui
ORBIT_RESOURCE_CAPTURE_DIR=/absolute/output/path go test ./internal/resourceui -run 'TestNativeCapture' -count=1
```

`TestMain` uses a disposable native cache. Every resource test uses temporary
homes/workspaces, no actual credential store, no real installation, and no
inference. Local Git tests contact only an explicitly created local bare remote.
Fake Pi executables are named and confined to tests.

The included probe images in `probes/resourceui/` are native MyGo CPU renders
of real disposable resource files. Their left-hand 240-DIP area is visibly
labeled “Settings navigation fixture”, not presented as an actual account or
complete host screenshot. They show the true page slot at full/minimum window
dimensions. Marketing images were inspected, but contain different data and
display scale; they cannot establish exact pixel parity. GPU capture, OS-native
file opening/reveal/clipboard, real approved package install/update/remove, and
same-state Rust/MyGo image deltas remain separate acceptance requirements.
