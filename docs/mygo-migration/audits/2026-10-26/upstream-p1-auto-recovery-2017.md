# Bounded selected-runtime recovery candidate

Independent increment on local combined P1 f4789e435a0857ce768d54384ee3a50c83ee7963 (event/request/status work). Not a release, not a full-parity/native-pixel claim. Exact upstream #87 source: imrj05/orbit `1a22457eb76dff9cdad2a2669213b7aaeea1d821`, app/runtime.rs:549–676 and auto_restart_tests.rs; source files and verified Git blob manifest remain in `evidence/orbit-upstream-p1-evidence-1919/rust`.

## Policy and intentional Go mappings

- Only an observed actual death of the currently selected owned child schedules recovery. A wire `process_exited` while ProcessSnapshot still reports no exit cannot authorize spawning. The local transport-error/exit pair is consumed once per child.
- At most three admitted automatic attempts per runtime lineage, with a 400ms delay. A cancelled pending attempt consumes no attempt. Automatic replacement, manual Start/Restart and MCP replacement carry the existing count; merely creating the automatic replacement entry cannot refill it.
- Refill requires an actual child that reached Ready and survived at least ten seconds thereafter. Use the child's actual ExitedAt, never a later UI drain time. This is deliberately stricter than upstream's spawn-time uptime check.
- A genuine explicit user Open may begin a new lineage only after the old owned child/lifecycle is retired. Dead, lease-free entries are removed on explicit Open; a cancelled automatic startup has a reap barrier before a same-file writer can start. There is no historical global session-budget table and no claim of “three attempts forever across all user actions.” Ordinary warm reuse is not a reset. A worker finishing retirement never removes the still-selected stopped task; it must remain manually restartable.
- Go has multiple active/parked runtimes whereas the source app controller is singular. Parked exits and eviction never respawn, and navigation cancels pending automatic intent rather than retaining a timer that can revive a later-selected old task.
- An empty initial draft with no messages and Turn=0 can bootstrap anew. If conversation/run history exists but no reliable session path is available, remain disconnected with a diagnostic instead of pairing the old transcript with a fresh empty session.
- No prompt, steer, follow-up, queue item, tool request, or business side effect is replayed. Recovery only uses the existing bootstrap path: SwitchSession for a saved path, then read/catalog/capability probes. Saved-session switch must acknowledge success and not cancellation before Ready; failed spawn/reattach/bootstrap uses the same remaining budget and its child is closed/reaped.

## Ownership and cancellation

A recovery lease binds selected key, exact liveSession pointer, manager navigation generation and due instant. Admission rechecks these atomically before detaching/creating a replacement. Startup-only lease checks guard pre-spawn, adoption, reattach/bootstrap and Ready/error publication. Ordinary non-recovery Open retains its prior ability to finish loading a parked session.

Open (including warm navigation), NewSession, accepted identity commands, Stop, Close, Start/Restart and MCP replacement cancel obsolete recovery intent. An auto-owned bootstrap is marked stopped/loading=false so it cannot be reused as a warm entry. Client Close runs on workers, not under the manager lock. A child created after cancellation is rejected and reaped before receiving session commands.

A cancelled automatic attempt owns a completion barrier through closing the old client, spawning and reaping any late child. Subsequent same-file startup waits on a worker, not the UI or manager mutex, and can be cancelled or time out with a retryable failure. Timeout is not treated as retirement success. Unrelated paths do not wait. Canonical path plus inode identity is cached before an automatic attempt can spawn: cancellation does no filesystem lookup, and retirement never re-reads an obsolete path that could now be deleted, renamed or a symlink loop. Known hardlink aliases still wait. An attempt cancelled before identity capture is structurally prevented from spawning. The intended new target's own disk errors remain errors; a not-yet-created saved path keeps the existing Open contract.

Terminal cleanup clears queue/restoration/entry-poll pending state and stale streaming markers. Diagnostics retain a bounded pre-redacted copy of prior stderr. Once Tick has observed this exact client's actual death, late command/UI/compact/abort transport completions cannot overwrite the terminal recovery/refusal/budget diagnostic. Live-child command failures keep their original behavior.

## Evidence and retained failures

All processes are disposable test executables speaking real stdio JSONL with temporary home/workspace paths, no provider credentials or actual tool execution. Traces record exact child PID and command kinds/paths.

External evidence directory: `evidence/orbit-auto-recovery-evidence-1945`.
- `reattach-red.jsonl`: unchanged prior candidate never respawns after a real child exits.
- `final-fast.jsonl`: retained intermediate repeated-run failure exposed late Send transport completion overwriting the missing-session-path refusal. This was fixed in production, not by weakening the assertion or increasing a timeout.
- `review-red.jsonl`: deterministic held model-command worker reproduces that order; removed/renamed old paths already pass, while an unrelated retired symlink-loop path blocks a new target. Cached owner identity fixes the latter without ignoring target errors or same-inode ownership.
- `terminal-order-green.jsonl`: both diagnostic-order cases ×10 under race after the narrow completion guard.
- `real-health.jsonl`: actual ten-second healthy child refills, actual quick crash delayed until after ten seconds before Tick does not refill. These are real process-time checks, not synthetic “now” injection.
- `ownership-first.jsonl`: broken stdin transport plus exit consumes once; parked death never revives; held pre-spawn then Stop/Open really launches a late child, verifies it receives zero commands, is reaped, and only then permits the new same-path writer.
- Other first-pass logs cover exact switch acknowledgment/cancellation/failure, bounded spawn failures, delay cancellation, navigation ABA, Close, retirement timeout/Stop, different-path independence and MCP validation versus exit producing one replacement.

Final focused hitsets, regression results and vet are recorded alongside the frozen patch. No second full suite or publication is performed by this worker. Source compatibility does not establish native pixel parity; automatic recovery deliberately does not retry a user's operation.

## Final isolated gates (2026-10-26 20:24 UTC)

- `final-fast-v2.jsonl`: 22 top-level tests + 10 subtests, each run/pass three times under race (96/96; 132.392s), exact expected hitset, no skips/failures.
- `final-real-health.jsonl`: both real-process ten-second tests run/pass once (2/2; 25.102s), no skips/failures.
- `final-regression.jsonl`: 22 existing lifecycle/MCP/abort/request-ID top-level tests + 10 subtests, each run/pass three times under race (96/96; 116.699s), exact expected hitset, no skips/failures.
- Each log has a sibling `.hitset.json` listing every verified name. The canonical script's one-line verification preamble is not treated as a test event.
- The same sequential command chain then completed agent `go vet` and darwin/arm64 application build, exit 0. These are isolated focused gates, not aggregate or native-runtime acceptance.
