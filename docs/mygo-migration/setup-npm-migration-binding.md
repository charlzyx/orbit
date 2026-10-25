# P03 explicit npm migration card (separate binding increment)

Baseline local `9117035`: frozen read-only setup, reviewed unexposed coordinator
`4924ed…`, reviewed quit settlement `17c6aac…`, main-column correction `461c5b…`.
The independently reviewed New Task admission correction remains a separate patch;
this binding does not edit its action/draft behavior.

## Actual pinned source entry, not an invented installer settings page

Rust `app/view.rs:2742–2745` mounts `render_pi_migration` only for npm installation
ownership on the setup page. `2903` provides the title/note/action/progress label.
`app/pi_update_ui.rs:178–252` verifies the resulting managed installation and shows
PATH shadowing, while updating dependency/install snapshots. It does not spawn or
restart a task on migration completion. Missing-Pi dependency hints remain
copy-only. No new generic Install Pi button, About route or legacy runtimeNative
surface is exposed by this patch.

The read-only dependency worker now carries the already-probed PiInstallation
alongside facts/dependencies, allowing the actual ownership result to select the
card. Clicking it opens an additional explicit app confirmation with the pinned
npm-removal/data-retention note. The callback captures workflow, binary, generation
and backend; cancel starts nothing. It rechecks lifecycle/intent and setup source
eligibility before invoking the reviewed coordinator. A second intent check after
maintenance acquisition closes the remaining admission interval without IO under
a manager/global lock.

Progress disables duplicate migration. Receipt adoption is owner-only and remains
inside the maintenance lease. Only verified managed installation yields success;
runner failure/exit-zero verification failure remains an error. Node readiness is
separate: missing required Node keeps setup visible, even after Pi migrated. PATH
shadowing has its own warning. No ResumeSetupFor, generic RuntimeChanged, NewSession
or Restart is invoked by receipt adoption. Stale target/generation results cannot
overwrite a newer report; a new read-only refresh can observe actual disk state.

## Safety boundaries

The actual installer adapter is dormant until explicit confirmation. Tests always
inject its runner/ownership/version/dependency results; no test executes an
installer, network request, auth flow or configuration migration. A native-target
admission test deliberately uses no injected backend and proves it rejects the
isolated fixture home before confirmation or runner creation.

The existing native installer uses process HOME/PI_CODING_AGENT_DIR. Therefore its
binding refuses a separately selected Store.UserHome/PiDir instead of silently
migrating another home. This conservative textual path check may reject aliases;
it does not resolve or modify user paths, PATH, permissions or security settings.
The injected fixture backend remains isolated. Custom selected Pi launchers/PI_BIN
remain explicitly unsupported by the reviewed coordinator rather than being
silently replaced. Unsupported cases report an error, never claim migration.

Closing the window retains the host operation as designed; application Quit uses
the reviewed cancel + actual-runner settlement channel and bounded typed-diagnostic
join. A timeout is uncertainty, not rollback or cleanup success. Physical native
installer execution and native macOS acceptance remain unverified.

## Evidence

Mounted RED `/tmp/orbit-setup-migration-red.log` (3.881s) proved the actual npm-only
entry was missing; managed/other/missing-Pi/fully-ready controls did not require it.
Initial mounted confirmation/receipt suite passed 8.850s. Expanded fixture suite
covers cancel, stale generation, actual selected-binary change, actual task-only
change with child execution prevented, Close, held maintenance, duplicate pending
clicks, successful/failed verification, runner failure, remaining Node, PATH shadow,
post-writer intent recheck, and async probe ownership propagation. Existing
coordinator, quit and onboarding tests run alongside it. A successful receipt is
explicitly tested not to create or replace a runtime/task.

### Refresh/migration consistency boundary (deliberate Go safety difference)

Rust refresh_setup guards only its own `refreshing` flag, independently of
pi_migrating. With Go's captured probe generation, allowing a refresh during the
installer could both read half-installed readiness and invalidate the final
migration receipt. Actual mounted RED confirmed both directions of this race
(`/tmp/orbit-setup-migration-refresh-red.log`, 1.560s).

Go now explicitly serializes them: Refresh is disabled and controller admission
rejects a dependency probe while an installation is pending, including the queued
receipt/owner-application interval. Conversely, migration is disabled and its
controller admission rejects while a dependency probe is active. No refresh
request advances the generation or reads a half-installed runtime in that period.
The final receipt is preserved; after actual settlement, Refresh is usable again.
This is an intentional stricter consistency rule, not a claim that Rust used the
same two-way guard.

Final focused race three repetitions passed 37.509s in
`/tmp/orbit-setup-migration-final3-focused.log`, including held runner, queued
receipt, verification/error, runner-reported deadline, actual canceled operation
context, subsequent explicit refresh, and Close during execution. No test waits
for or claims an actual ten-minute native installer timeout. CPU fixture
`/tmp/orbit-setup-migration.png` was inspected at 1200×1100; the actual card, note,
button, host facts and remaining required-Node count fit the main column. Native
OS execution/pixel acceptance remains separate. The stable English accessibility
label is a remaining localization detail; visible strings use the pinned catalog.
