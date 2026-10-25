# Explicit setup install/migration coordinator (not exposed in UI)

Baseline: local `d54db58`, the frozen read-only P03 candidate applied to
`ade11f3879abd04f0e26630b236466cf9886787e`. This is a separate additive increment.

## Source and contract

Rust `app/pi_update_ui.rs:178–252` admits explicit install/npm migration, refuses
unsupported platforms/duplicate managed/third-party ownership, runs off-thread,
and verifies the actual installation before success. Its verification receipt
also warns about PATH shadowing. `app/view.rs:2903` is the explicit migration
panel. Existing Go `runtimeNative` is not mounted and `runPiOperation` is not a
safe integration shortcut: it treats installer exit success as sufficient and
releases maintenance before owner completion. This increment does not wire that
legacy path or expose buttons.

`setup_install.go` provides a narrow future confirmation coordinator:

* Capture task, selected binary and setup-probe generation before confirmation;
  reject stale confirmation before nonblocking maintenance admission.
* At most one operation per controller; no queue of installer intents.
* Re-detect installation ownership before invoking an explicit runner. Npm
  replacement requires the captured migration confirmation. Managed duplicates,
  third-party ownership and unsupported platforms never reach the runner.
* Run with a ten-minute context derived from host lifetime. Context cancellation
  is a request, not settlement: a runner that has not returned retains the
  maintenance lease. No manager/global mutex is held during IO or owner waits.
* Re-probe installation and dependency versions after runner completion. Exit
  zero alone is not success: require an actual managed executable with a working
  version result. Nonzero status or runner error remains an error even when a
  managed executable exists afterward. Preserve diagnostic output and explicit
  PATH-shadow warning; never modify PATH, permissions, or security settings.
* Keep maintenance through owner application of the receipt. Application Close
  may discard the owner callback; after the actual runner returns, canceled host
  context releases the lease exactly once. Windowless retained hosts use their
  normal owner mailbox. No generic RuntimeChanged/Restart is emitted.

## Fixture evidence and limits

Tests only inject in-memory installer functions and fixture installation facts.
The fixture manager is never opened. No test calls the native installer adapter,
network endpoint, registration/auth flow or real-home migration. Existing native
runner cancellation uses a private process group on Unix and waits for command
return; this coordinator does not claim a physical installer was exercised.

A deliberate negative control removed the managed/version verification branch:
`/tmp/orbit-setup-install-verification-red.log` failed for missing, npm and broken
version results despite exit zero. The exact branch was restored; focused race
three repetitions passed in `/tmp/orbit-setup-install-final-focused.log` (3.120s).
An additional explicit nonzero-status fixture is in the final gate.

Tests cover confirmation/generation, custom launcher, platform, busy lease,
managed/third-party/npm ownership, confirmed npm, runner failure, exit-zero false
success, PATH shadowing, duplicate admission, pending owner application, stale
owner, Close during a deliberately cancellation-ignoring runner, and Close after
completion was queued. No button/success toast or automatic task launch is added.

Deliberate stricter safety: custom configured launchers or PI_BIN are unsupported
for this managed-path coordinator, rather than installing an unrelated PATH Pi.
A future UI must explain this, not silently replace the user's selection. The
working-version requirement is stronger than Rust's kind-only `managed()` check.
A ten-minute timeout cannot undo an install already partially performed; errors
need refresh/recovery, not a fake rollback claim. A runner ignoring cancellation
can retain one lease until it actually returns; tests explicitly prove that a
cancellation signal alone does not release it.
