# Runtime / Agent settings acceptance

## Source-first contract

Baseline: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, official Pi 1.4.0 (`4a080252f06caf63a525d7873c5d94d3c23887e0`).

- `crates/orbit-pi/src/app/settings.rs:4933–5182`: Runtime status, child PID, binary, running-only uptime, stdio transport, last error, parked process PID/busy state, and latest eight stderr lines. There is no socket host/port to invent.
- `crates/orbit-pi/src/app/runtime.rs:127–142`: child adoption establishes a process start time. The original `session.rs:809` also calls adoption when returning to a parked child, resetting its displayed uptime. The Go API deliberately uses the actual child-spawn time instead; same-PID navigation does not manufacture a fresh process age.
- `crates/orbit-pi/src/app/settings.rs:5239–5249,5781–5797`: auto-retry sends `set_auto_retry` with `enabled`; the original UI updates optimistically.
- Official installed `docs/rpc-commands.md#set_auto_retry` and `dist/modes/rpc/rpc-types.d.ts:148–161`: Pi 1.4.0 supports the setter but `get_state` does **not** report the current auto-retry setting. A truthful port therefore starts with an unknown value. Only a successful acknowledgement of a requested setting establishes a known value. No startup setter/probe changes user preferences just to populate a checkbox.

## Scenarios and expected results (defined before implementation)

1. **No child:** active process absent; configured executable may be displayed; no PID, uptime, or invented stderr.
2. **Loading:** tracked session row exists; before successful spawn PID/start time are absent. If the child exists but bootstrap is pending, its real PID/start are visible with `loading` state.
3. **Active idle/busy:** a successful real spawn reports its PID, actual launched command path, Pi launcher, immutable start time, elapsed running uptime, and observed idle/busy flags.
4. **Parked:** switching workspaces preserves the old child as a parked row. Returning to the saved live session reuses its PID and start time; it is not counted twice.
5. **Same-process new session:** session identity changes, but PID/start/confirmed process-level retry setting do not reset.
6. **Failed spawn:** state is `failed`, error is visible but redacted, PID/start/uptime remain unavailable.
7. **Exit/restart:** exited process has known exit code and no running uptime; restart receives a newly observed PID/start and unknown retry setting. No elapsed duration is copied from the prior child.
8. **Recent stderr:** return only the latest eight bounded lines, detached from internal buffers. Redact resolved secrets, sensitive child environment values, and common credential assignment/Bearer formats. Never export full child environment or command arguments.
9. **Retry success:** explicit setter starts pending, preserves the previous confirmed value while pending, and marks the requested value known/supported only on success. Two concurrent setter clicks cannot race a newer/older state.
10. **Retry failure:** unknown-command failure marks unsupported; other failures remain errors, not a fabricated capability conclusion. Failed setters never change the confirmed value. Restart clears process-level retry knowledge; live session reuse preserves it.
11. **Real Pi, credential-free:** pinned 1.4.0 in an empty HOME accepts explicit off/on setter calls, reports no retry value in get_state, emits no model/tool events, and incurs no inference. Temporary settings are removed with the test home.
12. **Concurrency/immutability:** runtime snapshots can be read while Tick, switching, requests, or Close run; race detector passes, caller mutations cannot affect manager state, no filesystem/process-launch work runs in the snapshot API.

Execution evidence is appended after implementation. These scenarios do not prove native rendering, paid inference, authenticated provider retries, or macOS/Windows process behavior.

## Executed item-by-item checks

All scenarios above passed in the cloud. The final command produced **70 passing test/subtest records, zero failures/skips**, with the race detector enabled. `go vet ./internal/pirpc ./internal/agent` and `git diff --check` passed. Darwin/arm64 and Windows/amd64 transport/runtime test binaries compiled successfully; they were not executed on those platforms.

| Acceptance case | Executed test | Result |
|---|---|---|
| no child, idle child, parked switch, same-process new session, restart | `TestRuntimeSnapshotLifecycleParkSwitchReuseAndRestart` | Passed |
| active-only stop/start preserves parked child | `TestRuntimeStopStartKeepsParkedAndManagerUsable` | Passed |
| loading and stop before spawn | `TestRuntimeLoadingAndStopBeforeSpawn` | Passed |
| failed spawn and exit code7 | `TestRuntimeFailedSpawnAndUnexpectedExitAreObserved` | Passed |
| last8 bounded/redacted/detached stderr lines | `TestRuntimeStderrIsBoundedRedactedAndDetached` | Passed |
| busy parked process remains visible | `TestRuntimeBusyProcessRemainsVisibleWhenParked` | Passed |
| setter ACK, pending prior value, duplicate rejection | `TestRetrySetterWaitsForAckAndRejectsConcurrentRequests` | Passed |
| unsupported versus ordinary rejected request | `TestRetryFailureDoesNotInventValueOrSupport` | Passed |
| rejection preserves known value; timeout marks unknown | `TestRetryRejectedRequestPreservesKnownValueButTimeoutDoesNot` | Passed |
| concurrent readers and stop | `TestRuntimeSnapshotConcurrentReaders` | Passed |
| credential assignment, Bearer, multiline secret redaction | `TestDiagnosticCredentialRedaction` | Passed |
| observed child start/exit and detached exit-code pointer | `TestProcessSnapshotReportsObservedStartExitAndDetachedCode` | Passed |
| auto-compaction value only known when get_state field exists | `TestAutoCompactionKnownRequiresObservedStateField` | Passed |
| real Pi metadata, setter off/on, absent getter, same-process reuse, no inference | `TestLiveManagerCredentialFree` | Passed |

Structured evidence: [runtime-settings-acceptance.json](probes/runtime-settings-acceptance.json), including exact test names, package timings, and source hashes. The real Pi test creates an empty temporary home, checks version1.4.0, disables resources/tools/network activity, verifies the missing getter, explicitly sets off then on, and observes zero messages/runs. No model request or credential use occurs.

`Manager.Start()` and `Stop()` affect only the active process; parked children remain alive. Stop during loading cancels the pending spawn before it is installed. Snapshots distinguish loading/stopping/stopped/failed/exited/idle/busy; missing PID, start, exit code, and initial retry values remain absent/unknown. Setter rejection preserves the prior confirmed value; timeout or transport loss marks it unknown because execution may have occurred. `Snapshot.AutoCompactionKnown` separately distinguishes an observed false value from a missing `get_state` field.

No claim is made here about native pixel alignment, authenticated provider retry behavior, or on-device Windows/macOS process execution.
