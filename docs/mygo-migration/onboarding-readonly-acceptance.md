# P03 read-only first-run setup gate

Baseline: `ade11f3879abd04f0e26630b236466cf9886787e`.

## Source contract and bounded implementation

Pinned Rust `crates/orbit-pi/src/app/view.rs:795–800` replaces the chat page with
setup when required dependencies are missing. `2592–2658` refreshes asynchronously,
ignores concurrent refreshes, and starts a client if dependencies become ready.
`2664–2830` shows required/optional dependencies, versions, copyable installation
hints, host/storage/installation facts, and refresh status.

The Go native-host ready callback now starts a read-only background probe; no
filesystem or child-version query runs in the render function. Unknown/missing
required dependencies replace chat with setup; optional Git does not block chat.
Settings remain reachable. Refresh adopts an owner-queued snapshot and returns to
the preserved chat/draft once required dependencies are available. Copy buttons
write only the displayed hint to the clipboard; they do not execute it.

A selected Pi executable (including PI_BIN when there is no explicit selection)
is probed instead of an unrelated PATH installation. A missing explicit launcher
reports its error, without offering a PATH installer as a purported repair.
Legacy CheckDependencies retains its existing PATH semantics for other callers.

Refresh is singleflight per binary generation, cancellable on controller close,
and rejects late results after binary replacement. A missing-to-ready refresh may
resume only its captured active workflow generation and selected binary. Manager
admission occurs under its lock, is restricted to a failed never-live entry, and
uses the existing start lifecycle checks. It cannot replace a task, restart a
previously live/stopped child, or duplicate a pending launch. Runtime maintenance
excludes automatic resume. A rejected resume does not synthesize success or start
a different task; existing task/runtime controls retain responsibility.

## Evidence

* Mounted missing-required/copy-button tests failed before implementation:
  `/tmp/orbit-onboarding-red.log`; optional-Git control already passed.
* Final focused race suite, three repetitions:
  `/tmp/orbit-onboarding-final2-focused.log`: native 1.203s, agent 7.857s,
  UI 5.712s. Includes real mounted refresh/copy/settings/draft restoration,
  singleflight, errors, actual binary switch, task switch, queued completion after
  close, maintenance, and exact-entry resume including held Stop/Close.
* Optional CPU render `/tmp/orbit-onboarding-missing.png` inspected at 1200×800:
  all three dependency cards, host facts and refresh footer visible. This is a
  fixture render, not native macOS acceptance or a pixel-parity assertion.
* Native dependency tests execute isolated version-only fixture scripts; manager
  tests launch their existing test child. No installer/auth/user configuration
  mutation is performed.

## Deliberate limits and source contradictions

This closes the missing required-dependency gate, not all P03. Installer progress,
installation/migration actions and the existing unreachable Runtime settings UI
remain separate work. No fake install/migration button is exposed. Native
first-run/panel/IME acceptance remains unverified. Rust's artificial 250ms refresh
spinner floor is not reproduced; Go displays actual in-flight state.

The catalog mentions `gh`, but actual pinned dependency checks contain only Pi,
Node and Git. Rust comments mention About reopening setup, yet `setup_open` has
no true assignment in the pinned tree; this patch does not invent that route or
a redundant Done button. The ready automatic page returns directly to chat.

The manager guard is deliberately stricter than Rust's bare `client.is_none()`:
old refreshes cannot restart new or intentionally stopped tasks. Current required
Node behavior follows the pinned source even when a standalone Pi installation
might independently run without Node. Executable presence retains the existing
native dependency API's installed semantics; a version-query error is displayed,
not silently converted into an installation action.

### Input and retained-host semantics

Source root `view.rs:523–525` registers file drops outside the setup conditional;
`composer_ops.rs:405–433` excludes Settings only. Setup therefore retains a user
file drop for the hidden draft, without automatically submitting it. Go preserves
this non-obvious behavior rather than blocking all input. Root global actions
(`view.rs:1287–1328`, `session.rs:581–611`) also keep explicit task/navigation keys
available. Enter/alternate send belongs to the actual composer focus
(`session.rs:88–103`); a formerly focused but now unmounted Go composer does not
receive text/send keys. Mounted regression distinguishes setup vs Settings drops
and checks that returning ready does not submit the retained draft.

Closing the last macOS window retains the host by design (`native/lifecycle.go`),
so its owner heartbeat may complete the read-only probe while windowless. Reopen
uses the same controller, not a second probe/session. Application Quit invokes
Controller.Close, cancels the probe context and rejects queued completion. This
is intentionally different from native file-dialog results, which belong to a
specific initiating window.

### Final gates

Final production full native/agent/UI race passed in
`/tmp/orbit-onboarding-full-race.log` (73.849s / 151.624s / 403.152s).
After adding the input and retained-host regression tests, final-byte full UI race
passed again in `/tmp/orbit-onboarding-final-full-ui.log` (398.492s), focused UI
three times in `/tmp/orbit-onboarding-final-ui-focused.log` (9.389s), and input
three times in `/tmp/orbit-onboarding-input.log` (14.621s). Final native/agent/UI
vet and Darwin arm64 cross-build passed (`/tmp/orbit-onboarding-final-vet.log`,
`/tmp/orbit-onboarding-final-arm64.log`); `git diff --check` passed. Cross-build is
not a physical macOS launch or user first-run acceptance.
