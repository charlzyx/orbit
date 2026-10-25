# Installer quit settlement (additive, still unexposed)

Baseline local `00ab0d0`, reviewed coordinator patch `4924ed…`.

MyGo's pinned `app.go` documents and fires OnQuit after its event loop has stopped.
Therefore a queued owner receipt cannot be the shutdown join target. The former
coordinator correctly retained maintenance during actual work but Controller.Close
could return immediately after requesting cancellation, before a native child
was reaped. A deterministic held-runner regression failed against that baseline:
`/tmp/orbit-setup-install-quit-red.log` (Quit returned before actual settlement).

This additive change separates `runnerSettled` from owner acknowledgement and
maintenance completion. It closes only after the entire IO phase returns.
Controller.Close first closes host admission and cancels background work, then
requests operation cancellation and waits at most three seconds for that channel.
No owner event-loop callback, manager lock or UI-global lock is required to settle.
The existing worker still releases maintenance exactly once only after actual
return; the shutdown wait does not manufacture completion or free a running lease.

Typed, allowlisted diagnostics record `runtime_install_shutdown` with started and
ok/timeout only; no arbitrary errors, paths, commands or installer output enter
that log. Timeout explicitly means cleanup remains uncertain, not that the
installer stopped. The normal backend uses context-aware download and the existing
native command runner's private Unix process-group cancellation plus 2s WaitDelay.
A cancellation-ignoring injected runner can exceed the bound; this implementation
does not claim it can roll back partial installation or stop arbitrary escaped
processes. No real installation is executed or exposed in this increment.

Fixtures cover a held runner observing cancellation but not yet returning; the
close path must wait for actual return, not just context cancellation. A short
bounded-join fixture verifies timeout diagnostics without releasing the active
lease. Existing queued-result and closed-host tests cover shutdown after receipt
admission and prevent owner-ack deadlock. A Linux-only fixed `/bin/sh` plus `sleep`
fixture uses native RunInstallerInvocation solely as a process-control seam,
checks actual runner settlement before Quit returns and verifies its own child is
terminated (or an inert orphan zombie). No installer script, endpoint, auth,
registration, real-home configuration or security setting is used. Physical
macOS installer execution and child reaping remain unverified.
