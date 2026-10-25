# Automatic update target safety gate

The native automatic worker resolves its installation and installer target using
process HOME / PI_CODING_AGENT_DIR (`native/runtime.go:151,224,371`). Controller
Store and Agent can instead select UserHome / PiAgentDir independently. Blocking
only the eventual receipt would still allow mutation of the wrong installation.

This separate delta reuses the explicit migration binding's conservative
`nativeSetupTargetMatches` check before starting the automatic worker. A mismatch
silently skips this automatic operation (it is not a failed user-requested
installation). After the existing nonblocking maintenance writer admission, it
rechecks target, selected binary and closed host; rejection releases the lease
without running any update. Matching defaults and configured matching Pi agent
directories retain the existing worker and owner event delivery.

No installer target, PATH, home, configuration file, automatic delay, network
request, release lookup, update plan or business timeout is changed. Textual
normalized-path equality is conservative and may reject aliases rather than risk
another target; it does not change filesystem/security settings. Store target
fields are immutable during normal Controller lifetime. This does not claim to
serialize arbitrary external process-environment mutations after admission.

Tests first fail with the original extracted wrapper: both mismatched initial
targets start the injected checker, and a later directory change still admits the
writer. Fixture-only tests then cover home/directory mismatch, matching default
and explicit directory, post-admission recheck/release, custom binary, Close,
reader exclusion, and actual owner-queue notification delivery. The injected
checker never calls the native updater, performs network traffic or installs Pi.
Worker actual-completion shutdown tracking remains a separate follow-up.
