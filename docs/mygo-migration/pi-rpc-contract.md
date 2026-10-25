# Pi RPC contract and migration verification

Audit date: 2026-10-25. This is a verification report and implementation contract, not a claim that a MyGo port exists.

## Pinned sources

- Orbit baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
- MyGo reference: `9c69cf829ece196c5b3a3601f1c7bbe70526376f`.
- Runtime tested: `@earendil-works/pi-coding-agent@1.4.0`, installed from the official npm package into a disposable cloud-only prefix, with lifecycle scripts disabled.
- Official Pi tag: [`v1.4.0`, `4a080252f06caf63a525d7873c5d94d3c23887e0`](https://github.com/earendil-works/pi/tree/4a080252f06caf63a525d7873c5d94d3c23887e0).
- npm tarball: `https://registry.npmjs.org/@earendil-works/pi-coding-agent/-/pi-coding-agent-1.4.0.tgz`.
- Registry integrity: `sha512-Z41LpZ3yYHSC/UUps6Mr39UMFPwH8elhJBG9DY1P7kuzsE5qSK/6eMj2EHzO5evJRQWLbKSqUMw+4nCNtIyiQg==`.
- Runtime: Node `v24.22.1`; package requirement: Node `>=22.19.0`.

Authoritative contract was read from the **installed 1.4.0 package**: `docs/rpc.md`, `docs/rpc-commands.md`, `docs/rpc-extension-ui.md`, `docs/json.md`, `dist/modes/rpc/rpc-types.d.ts`, and `dist/modes/rpc/rpc-mode.js`. The equivalent source reference is [Pi RPC at the pinned commit](https://github.com/earendil-works/pi/blob/4a080252f06caf63a525d7873c5d94d3c23887e0/packages/coding-agent/docs/rpc.md). Do not infer future compatibility from an unpinned `main` document.

## Existing Orbit architecture is already correct

`crates/orbit-rpc/src/client.rs` owns a long-lived Pi subprocess, not an HTTP service, a JSON-RPC 2.0 endpoint, or an in-process model SDK. The GUI talks JSON objects over the child's stdin/stdout. Pi remains the sole agent runtime after a UI migration.

The baseline launch is:

```text
pi --mode rpc --approve [default provider/model flags] [--session-dir DIR] [--extension FILE ...]
```

`--approve` grants **project trust**, allowing Pi to load project-local configuration/resources. It is not blanket tool approval. Orbit's tool approval behavior comes from its guard extension; it is not an operating-system sandbox.

Baseline transport behavior to preserve:

- One process per live session, with workspace `cwd` and process-scoped environment overrides.
- `PI_BIN` resolution, managed/npm package discovery, and PATH augmentation for GUI launches.
- Independent stdin writer, stdout reader, and capped stderr drain.
- Unique command IDs, ID-correlated replies, and a shared stream including responses and asynchronous events.
- Nonblocking event draining on an app-owned heartbeat: 16 ms while active, 90 ms while idle (`main_window.rs`). It continues draining parked sessions with no window open.
- Process shutdown/exit surfaced to the UI; no fabricated completion or connection status.
- Stored sessions remain Pi-owned. Orbit's UI-specific layout/settings remain separate.

## Wire contract

### Framing and correlation

- Each record is one JSON object terminated by LF. Accept CRLF by stripping one trailing CR. Do not treat U+2028 or U+2029 in JSON strings as record boundaries.
- Responses use `type: "response"`, `command`, `success`, and the command's optional string `id`; success may include `data`, failure includes `error`.
- Commands can complete out of order. The live probe sent `abort` third and received its response after several later read commands. Arrival order is not a correlation mechanism.
- Stdout is protocol data; stderr is diagnostic data. Both must keep draining while the GUI is busy.
- Preserve unknown JSON fields/event types. The current Rust parser is permissive, including malformed records surfaced as unknown payloads rather than crashing the GUI.
- Extension UI responses are **one-way** records, not ordinary commands awaiting a `response`.

### Command surface used by Orbit

| Area | Existing commands |
|---|---|
| Prompt and run | `prompt`, `steer`, `follow_up`, `abort`, `clear_queue` |
| Queues/recovery | `set_steering_mode`, `set_follow_up_mode`, `compact`, `set_auto_compaction`, `set_auto_retry`, `abort_retry` |
| Session | `new_session`, `switch_session`, `get_state`, `get_messages`, `set_session_name`, `get_session_stats`, `get_fork_messages`, `get_entries`, `fork`, `clone` |
| Models | `get_available_models`, `set_model`, `cycle_model`, `get_available_thinking_levels`, `set_thinking_level`, `cycle_thinking_level` |
| Discoverability | `get_commands` |
| Orbit patched capabilities | `auth.list`, `auth.status`, `auth.login`, `auth.logout`, `auth.cancel`, `quota.list` |
| Escape hatch | Raw version-specific JSON for extension UI and other known Pi commands |

Pi 1.4.0 also defines commands such as `bash`, `abort_bash`, `export_html`, `get_tree`, and `get_last_assistant_text`. These do not automatically become new product features just because the protocol exposes them.

### Run completion and streaming

A successful `prompt` reply acknowledges handling/acceptance, not a finished response. Inspect its `data.disposition`; an extension-handled prompt need not start an agent run. `agent_end` ends a lower-level run, while `agent_settled` means no automatic work remains. Retry/compaction/queued work can occur in between.

`message_update.assistantMessageEvent` contains content-block deltas. A future Go model should retain `contentIndex`, unknown fields, and the raw envelope, then replace reconstructed content with authoritative `message_end.message`. Tool lifecycle records correlate by `toolCallId`. Do not flatten multiple content blocks into one string or assume usage arrives on every delta.

### Extension UI

Stock Pi supports dialog methods `select`, `confirm`, `input`, `editor` via `extension_ui_request`. Responses must reuse the exact request ID:

```json
{"type":"extension_ui_response","id":"request-id","value":"selected option"}
{"type":"extension_ui_response","id":"request-id","confirmed":true}
{"type":"extension_ui_response","id":"request-id","cancelled":true}
```

Cancellation is a top-level `cancelled` field. Putting it inside `value` can leave the extension waiting forever. Fire-and-forget methods (`notify`, `setStatus`, `setWidget`, `setTitle`, `set_editor_text`) must not receive synthetic dialog replies. Parked sessions need a deliberate fail-closed dialog policy, as the existing `cancel_dialog` path supplies.

Stock Pi 1.4.0 still has no custom TUI modal support in RPC. Orbit's patch adds `method: "custom"` open/render/close frames and one-way `extension_ui_input` / `extension_ui_resize` records. The native host paints server-provided ANSI lines, sends keyboard bytes/viewport dimensions, and does not reimplement the component. Gate this on `get_state.capabilities.extension_ui`, never assume it.

## Patches and bundled extensions are part of functional parity

The baseline's `crates/orbit-pi/src/rpc_patches.rs` automatically applies three best-effort, marker-delimited patches before each session spawn, with a fingerprint/stat fast path. It embeds the scripts from `contrib/`, maintains `.orbit-orig` backup/revert behavior, and degrades when unavailable. A UI port that retains only `prompt`/`get_state` loses real Orbit behavior.

| Feature | Upstream 1.4.0 | Orbit mechanism | Cloud verification |
|---|---|---|---|
| Provider auth | `auth.list` is unknown | `contrib/pi-auth-rpc` | Installer succeeded; discovery returned 42 unauthenticated providers |
| Account quota | `quota.list` is unknown | `contrib/pi-quota-rpc`; bundled quota bridge is an independent fallback | Installer succeeded; empty-credential query returned empty providers |
| Custom TUI modal | `custom()` unavailable | `contrib/pi-custom-ui-rpc` | Installer succeeded; `get_state` advertised `custom` capability |
| Tool approval | Not implied by `--approve` | `contrib/orbit-guard-extension` | Offline decision/extension tests passed |
| Plan/Build/Ask | Not a stock access sandbox | `contrib/orbit-workflow-extension` | Offline policy/extension tests passed |
| Session title | Separate model-backed extension | `contrib/orbit-title-extension` | Offline tests only; no title model call made |
| Account usage bridge | Custom session entries | `contrib/orbit-quota-extension` | Offline adapter/extension tests passed |

All three installers found their expected seams in Pi 1.4.0's bundle `dist/bundle/chunks/chunk-MM3LKT73.js`. They modified only the disposable cloud package, not a user's global installation. Patch success does **not** prove OAuth login, a provider's usage endpoint, or every custom UI interaction works end to end. Those remain untested here.

Credential handling stays inside Pi. Auth discovery should expose supported methods, credential kinds, and status, not tokens. Quota data must remain normalized and sanitized. The quota bridge uses `get_entries` cursors scoped to a session; reset the cursor on session change and recover from a stale cursor rather than inventing totals.

## Executed verification

### Real Pi, no credentials and no inference

The subprocess was given a clean, isolated HOME and `PI_CODING_AGENT_DIR`, an allowlisted environment (no inherited provider credentials), and settings with install telemetry/analytics disabled. Arguments:

```text
--mode rpc --no-session --no-approve --offline --no-extensions --no-skills --no-prompt-templates --no-context-files --no-tools
```

`PI_OFFLINE=1` and `PI_SKIP_VERSION_CHECK=1` were also set. No project code, discovered extensions, external MCP servers, authentication, model inference, shell tool, purchase, or payment was requested.

| Check | Stock 1.4.0 | Patched 1.4.0 |
|---|---|---|
| `get_state` | Success; idle, zero messages | Same, plus extension UI capabilities |
| `get_available_models` | Success, `models: []` | Same |
| `abort` while idle | Success | Success |
| `get_available_thinking_levels` | Success, `["off"]` | Same |
| `get_messages` | Success, empty | Same |
| `get_session_stats` | Success, zero tokens/cost/tools | Same |
| `get_entries` | Success, initial thinking-level entry | Same |
| `get_commands` | Success, empty | Same |
| `auth.list` | Explicit unknown-command failure | Success, 42 providers, all unauthenticated |
| `quota.list` | Explicit unknown-command failure | Success, empty providers |
| Close stdin | Clean exit code 0 | Clean exit code 0 |
| Agent/message/tool start events | None | None |

An empty model list is the honest no-credential/offline result, not a populated-catalog test. Idle `abort` verifies framing and server handling; it does not verify interrupting a real model stream.

Reproducible audit tooling and sanitized evidence are checked in under [`probes/`](probes/): `pi_rpc_probe.py`, `stock-pi-1.4.0.json`, and `patched-pi-1.4.0.json`. Session IDs/timestamps are real observed ephemeral values; machine-specific paths are replaced with placeholders. The probe accepts an executable and evidence destination and rejects a Pi version other than 1.4.0. It always creates and cleans up its own empty home/workspace.

```sh
python3 docs/mygo-migration/probes/pi_rpc_probe.py \
  --pi /path/to/pi --expect stock --output /tmp/pi-stock-evidence.json
# After the three Orbit patches have been applied to an authorized test install:
python3 docs/mygo-migration/probes/pi_rpc_probe.py \
  --pi /path/to/pi --expect patched --output /tmp/pi-patched-evidence.json
```

### Bundled extensions and handlers

Executed against the pinned Orbit source:

```sh
node --test contrib/orbit-quota-extension/*.test.mjs \
  contrib/pi-quota-rpc/*.test.mjs contrib/pi-custom-ui-rpc/*.test.mjs \
  contrib/orbit-guard-extension/*.test.mjs contrib/orbit-title-extension/*.test.mjs \
  contrib/orbit-workflow-extension/*.test.mjs
```

Result: **122 passed, 0 failed, 0 skipped** across the 10 test files. These are network-free/fake-provider tests, not proof of authenticated provider behavior.

### Avoid accidental paid tests

Do not run all baseline live tests with a real authenticated Pi executable. `crates/orbit-rpc/tests/live_pi.rs::live_prompt_streams_agent_events` and `live_switch_probe.rs` send actual prompts. Use `--lib`, explicit fake test targets, and individually selected no-inference live tests. The automatic title extension is also model-backed, so it must stay disabled in credential-free probes.

## Portable Go transport feasibility and acceptance criteria

The transport can use only Go's standard library (`os/exec`, `encoding/json`, `bufio`, `context`, `sync`, `io`). It has no dependency on Cocoa, MyGo, a browser, or a Node SDK. Node is required by the installed Pi CLI, not by the Go UI host.

The native Go/MyGo reconstruction has been confirmed. Transport implementation is assigned to `internal/pirpc`, with session/domain reducers in `internal/agent`. These remain independently testable on Linux even though the MyGo UI requires macOS. Transport implementation/testing status is tracked separately from the completed audit above.

Required transport properties:

1. Process ownership with configured executable, argument prefix for Node launchers, cwd, isolated test environment, repeated extension flags, and safe shutdown escalation.
2. LF-only framing without `bufio.Scanner`'s default 64 KiB ceiling. Image attachments, transcripts, and custom UI records can be much larger. Retain unknown/raw JSON and distinguish a malformed record from an unsupported event.
3. Serialized writes and independent continuous stdout/stderr drains. UI consumption must not block the reader or response matching.
4. ID correlation under concurrent/out-of-order responses. Always remove pending requests on cancellation, write failure, and process exit.
5. A separate one-way write method for extension responses/input/resize, preserving request IDs and not registering phantom pending calls.
6. A bounded stderr history and explicit process-exited/error events. Do not mistake stdout EOF alone for a known exit status; reap the child.
7. UI-safe event draining and cancellation. Avoid an unbounded hidden goroutine or blocking pipe after the window/session closes.
8. Keep transport independent of protocol reducers: session/transcript/auth/quota state and capability gating belong above it.

Test gates before enabling the port:

- Fake-process tests: concurrent/out-of-order responses, unknown events, stderr draining, process exit, malformed records, large frames, U+2028/U+2029, CRLF, cancellation, one-way UI records, write failure, arguments/env propagation, and shutdown.
- Race detector for process/event/request lifecycle; test under both slow UI consumption and sudden child exit.
- Opt-in real Pi checks for the same credential-free command list above, pinned runtime, empty environment, and explicit offline/no-tool flags.
- Separate authorized provider streaming tests only when credentials and model usage are intentionally enabled; do not claim those from fake or offline tests.
- MyGo integration must render actual state/events and preserve extension behavior. Pixel/interaction parity is a separate gate from transport correctness.

### Baseline transport details worth improving rather than copying blindly

The Rust transport registers a pending reply before the writer sends a command, and does not explicitly clear every pending receiver on writer failure/EOF. A Go port should fail outstanding calls immediately on terminal transport failure. `Raw` commands can preserve an existing JSON ID while the current `send()` pending-map key uses a generated ID; one-way UI writes should never use that normal request path. These are source-audit observations, not regressions introduced by this migration.

## Go implementation and executed checks

The confirmed Go reconstruction now contains two independently testable packages:

- `internal/pirpc`: real child ownership, 39 named Pi/Orbit command constants, flat command encoding, raw-preserving events, concurrent correlation, one-way extension UI records, process cleanup, bounded event/diagnostic storage, context cancellation, and platform launch handling.
- `internal/agent`: asynchronous active/background session manager, Pi session-store discovery, model/thinking/queue/transcript/auth/quota reducers, custom UI/dialog/status requests, exact embedded extension materialization, patch lifecycle, access/workflow persistence, session defaults, and per-turn Git checkpoint integration.

`Manager.Snapshot()` returns a deep copy, including partially streamed tool arguments that are not yet valid JSON. `Tick()` drains background sessions too, cancels their unpresentable dialogs, and returns whether the app should use its active heartbeat. Busy sessions are never evicted to satisfy the warm-process cap; idle parked processes expire after five minutes by default. Closing a window must not close this manager; quitting the application must. `LiveSessionPaths()` protects all manager-owned session files from user-enabled retention. `AnyBusy()` includes active/background startup, unsent prompts, runs, compaction, and retry; UI controllers must serialize retention/update transitions against new opens or prompts because these read APIs are point-in-time snapshots, not locks.

Both packages passed the race detector in the cloud, including three real **credential-free, offline** Pi 1.4.0 integration tests for transport, manager bootstrap/new-session/abort, and saved-session metadata rename. Fake subprocess tests cover out-of-order/concurrent responses, explicit and duplicate IDs, cancellation, exit, stderr saturation, large/Unicode/CRLF records, malformed/unknown events, overflow, blocked writers, one-way dialogs, spawn arguments/environment, graceful/forced shutdown, session pooling, and snapshot isolation. Reducer tests cover content-block indices, final-message authority, tool-result merging, auth isolation, quota-entry cursors, and activity-based session ordering.

The final recorded run contains **53 passing test records (including subtests), no failures or skips**, with `go vet` clean. [Sanitized Go verification evidence](probes/go-runtime-verification.json) records exact test names, package timings, source hashes, and separate Darwin/arm64 and Windows/amd64 **compile-only** checks. Immediate Escape cancels even an unsent prompt still waiting for the Git checkpoint; queued inputs share that cancellation context, and interactive abort clears/restores queued text before stopping Pi. Dynamic MCP secrets are reloaded on every spawn and redacted from diagnostic snapshots, including explicit overrides of resolved secret keys.

Run the offline default suite:

```sh
go test -race ./internal/pirpc ./internal/agent
```

The real integration cases are **opt-in**. `PI_BIN` alone does not enable them:

```sh
ORBIT_PI_INTEGRATION_BIN=/path/to/official/pi-1.4.0 \
  go test -race ./internal/pirpc ./internal/agent -count=1
```

The tests create an empty HOME, disable automatic network activity/extensions/tools/telemetry, and never send an inference prompt. This is runtime correctness evidence, not authenticated model-stream, provider-login, provider-quota, native rendering, or pixel-parity evidence. GUI assembly and remaining feature verification are separate acceptance gates.

Successful compaction invalidates only the cached context-window baseline immediately, preserving cumulative counters and cost. A `compaction_end` event requests fresh session stats even for manual compaction with no `agent_settled`. Regression tests preserve the old baseline on failed/aborted compaction, and verify that successful compaction never flashes its old percentage while the refresh is pending. The UI usage adapter keeps session totals separate from per-message usage, does not replace null context with transcript-derived guesses, and suppresses stale prior assistant output while a new response has not reported usage.
