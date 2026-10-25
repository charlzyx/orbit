# Remove the unused legacy runtime UI closure

Exact clean baseline: `baf9b63b8f35ded873e4ef028dd0dd57fc5ba548`.
Work is isolated; original implementations remain in Git history. No user state,
configuration, sessions, credentials, filesystem data or migration schema is
removed. This is not a restoration of the old runtime settings page.

## Reference audit before deletion

Whole-repository Go reference scans, including tests, found:

| Symbol | Baseline references | Decision |
| --- | --- | --- |
| `runtimeNative` | declaration only, `native_settings.go:39` | Remove unused view |
| `loadRuntime` | declaration plus calls inside that view and `runPiOperation` | Remove closed helper |
| `confirmPiInstall` | declaration plus the unused view's install/migrate buttons | Remove closed helper |
| `runPiOperation` | declaration plus the unused view and its confirm callback | Remove closed helper |
| `checkPiUpdate` | declaration, unused view, `run.go`'s `CheckPiUpdate` case | Remove with unemitted string case |
| `Options.BeginRuntimeOperation` | field, one New-controller assignment, two reads in `runPiOperation` | Remove orphan option/wiring |

The exact `"CheckPiUpdate"` action string existed only in that switch case: no
command catalog entry, shortcut binding, native menu item, action emitter, test
call or serialized preference schema uses it. The real application update route
is **CheckForUpdates**, retained in `commands.go` and `native/menu.go`; automatic
Pi checking is **native.CheckPiUpdateOnLaunch**, also retained.

The remaining legacy fields `dependencies`, `installation`, `runtimeLoaded`,
`busy`, `latest`, `output` had readers only in the removed closure. External
writes were the startup Pi OnEvent callback, a discarded model-refresh output
buffer, and one dummy `runtimeLoaded=true` assignment in `runtime_view_test.go`.
Those writes/fields are removed, not their live surrounding behavior. These were
private, in-memory UI fields, not persisted or exported schema. The Options hook
was code wiring only. No preference type, JSON tag, setting file or CLI flag is
changed. After deletion, the whole-repository Go scan for these removed method,
option and action identifiers has zero matches.

## Live paths explicitly retained

* Actual `runtimeSettings` / `runtime_view.go`, snapshot-based PID/process/start/
  stop/restart/patch controls, and every assertion in its tests. Only the unused
  fixture-state assignment is removed.
* Controller.beginRuntimeOperation, maintenance exclusions, new guarded setup
  coordinator/quit settlement, and the independent npm setup binding path.
* Automatic Pi startup checking: its OnEvent notifications and RuntimeChanged
  dispatch remain byte-for-byte behaviorally unchanged. ChoosePiBinary remains.
* Actual model catalog refresh: environment/secret resolution, native catalog
  operation, task restart/open, redacted error handling, progress, and success/
  failure notifications. Only stdout/stderr's already unread legacy UI copy is
  discarded; the native command still consumes its output and reports errors.
* Privacy/temp cleanup and the Orbit updater UI in the rest of
  `native_settings.go`; the file is deliberately retained.
* Native InstallPi / RunPiUpdate / LatestPiVersion / PlanPiUpdate, their active
  consumers and all native security/process/update tests. No native API is cut.

Before adding this audit document the source delta is 16 added / 167 removed lines
across five files; most additions are gofmt field alignment. No test is deleted.
All-package type build `scripts/go-native.sh test ./... -run '^$'` passes in
`/tmp/orbit-legacy-runtime-typebuild.log`. Full race tests, vet and Darwin arm64
cross-build are required before integration. These checks are not actual native
macOS installation or UI acceptance.
