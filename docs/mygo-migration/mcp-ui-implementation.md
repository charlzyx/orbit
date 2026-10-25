# Native MCP settings and lifecycle

The Go implementation lives in `internal/mcpui` and uses the pinned MyGo
`v0.6.4-0.20261025001514-9c69cf829ece` native `ui.Element` API. It does not import
`internal/ui`, embed a browser, invoke a Rust backend, or manufacture connection
status/tool data. The behavior reference is `crates/orbit-pi/src/app/mcp_ui.rs`,
`crates/orbit-pi/src/mcp/{config,manager}.rs`, `docs/mcp.md`, and the source theme
metrics in `crates/orbit-pi/src/theme/tokens.rs`.

## Mount contract

```go
state := mcpui.New(store, workspace, mcpui.Options{
    PiBinary: piBinary,
    OnChange: invalidateWindow, // goroutine-safe wake/invalidation only
    OnModified: scheduleSafePiReload, // UI thread, after a persisted mutation
    OnApply: scheduleSafePiReload, // UI thread, explicit reload request
    Text: translateWithFallback,
    Theme: func() mcpui.Theme {
        return mcpui.Theme{
            Elevated: elevatedColor,
            CodeFont: codeFamily,
            Px: scalePx,
            Space: resolveSpacing,
        }
    },
})
```

Mount `state.View(c)` directly in the Settings content slot, not inside a second
scroll view. The module owns the pinned title, subtitle, toolbar and status row,
its scrolling content, and its native modal layers. It takes standard palette,
font and accessibility colors from `c.Theme()` and only injects additional Orbit
geometry through `Theme`.

The host calls `SetWorkspace`, `SetPiBinary`, `Tick` on its UI heartbeat, and
`Close` at shutdown. `Tick() bool` reports a view change. `OnChange` only requests
a frame; worker goroutines never mutate UI state. `HasModal` and `Escape` let the
host guard native menu commands. `AddServer` / `BeginAdd`, `BeginEdit`, `Refresh`,
`Probe`, and `SetEnabled` provide command-palette entry points.

`OnModified` and `OnApply` must use the host's debounced, busy-gated,
session-preserving Pi reload path. They must not kill an active turn or drop a
parked session's work. Calling a restart API is not proof that extensions are
ready. `Applied()` exists only for an observed successful reload; until then the
UI retains a truthful pending/requested notice. The package does not claim that
all active/parked Pi processes are refreshed merely because a callback ran.

## Implemented controls

- Pinned title/subtitle and native 32px search/refresh/add toolbar; selectable
  server detail and separate global/project groups. Source-derived 960px content
  cap, 24px gutters, grouped boards, 560px editor, and modal header/footer metrics.
- Real merged global/project `mcp.json` definitions, source filenames, description
  metadata, enabled switch, all five exposure modes, real status/error/tool and
  resource counts, connection test, editable typed fields, and explicit removal
  confirmation. Configuration errors and Pi probe errors are visible.
- Typed stdio command, one-argument-per-line arguments, environment lines, working
  directory, enabled state and request timeout **in seconds**. Typed streamable
  HTTP URL and header lines, with optional OAuth client ID, client secret,
  callback port, loopback URL and scope. Legacy SSE is rejected. Per-tool exposure
  overrides and unknown definition fields survive edits.
- Source-compatible scope migration: save the destination first, then remove the
  original entry and metadata. Partial failures are reported and disk state is
  reloaded rather than pretending a multi-file transaction was atomic.
- Literal env/header/OAuth secrets are masked when editing and require a native
  confirmation before entering the 0600 private store. Pi receives `${NAME}`
  references, not literal values in `mcp.json`. References and `!command` values
  survive unchanged. A dedicated password-masked secret prompt names the server,
  scope and exact reference and warns that other configs using that name share
  the value. Cancellation clears the transient secret draft.
- Actual `pi mcp list --json` probes, `pi mcp login NAME --timeout 300`, and
  `pi mcp logout NAME`; login cancellation and logout confirmation are native.
  The standalone probe intentionally does not add `--approve`: Pi's real trust
  note remains visible. No separate MCP implementation is invented in Go.
- External config rechecks while the page is visible, codemode-disabled warnings,
  deferred/apply affordances, keyboard row navigation, editable native text
  controls, linked field labels, and nested Escape/cancel behavior.

## State and persistence guarantees

All file reads/writes and subprocesses are worker jobs. Each result carries a
UI generation; workspace changes, Pi executable changes, cancellation and close
reject stale results. A slow read-only probe can be canceled by an editing
operation rather than locking the entire configuration page for its 120-second
budget. Busy writes remain serialized. CLI operations inherit the workbench
process limits and cancellation behavior. On supported Unix platforms the
shared process helper kills the subprocess group on cancellation; Windows hides
the console, while full Windows descendant-job cleanup remains a separate
platform gate.

The backend additions in `internal/workbench/mcp.go` are shared rather than
UI-only persistence:

- `SetMCPDescription(scope, name, description)` reads/writes the upstream
  `mcp-metadata.json` shape, including `version: 1` and
  `descriptions["scope|workspace|name"]`. `MCP()` loads descriptions. Malformed
  sidecars are surfaced and not silently replaced.
- `MCPAutoEnableCodemode() (*bool, error)` distinguishes unset from an explicit
  bool, with project precedence; `SetMCPAutoEnableCodemode` edits only that flag.
- `MCPSecretNamespace` and `MCPGeneratedSecretName` preserve the upstream FNV-1a
  project namespace. `RemoveMCP` removes metadata and prunes generated orphaned
  secrets only in its own namespace. `PruneMCPSecrets` scans both raw config
  files, so a project-shadowed global server still protects its referenced key;
  other project namespaces and user-named secrets are retained.
- `ProbeMCP` accepts valid servers JSON on exit 1, which Pi uses for individual
  server failures. Other exit errors and invalid output remain failures. Shared
  `Service.Command` returns sanitized stdout alongside its existing error,
  preserving callers' error semantics. Decoded JSON strings are scrubbed so
  escaped quotes/newlines cannot make known secrets evade redaction.
- MCP timeout validation accepts native Go integer values as well as JSON
  numbers; OAuth shape, string fields, loopback URL and port are validated.

The editor does not copy raw values into an output preview or generic JSON
editor. The native UI never shows a stored secret value as status/tool text.

## Verification

Focused tests use temporary homes/workspaces and fake CLI executables, never
real credentials or external MCP services. They cover:

- typed stdio/HTTP/OAuth validation, one-argument-per-line serialization,
  unknown/per-tool field retention, masking, scope isolation and exact FNV names;
- metadata round trips, malformed-file refusal, codemode precedence, hidden
  global/shared-secret preservation, generated-orphan pruning and exit-1 probes;
- decoded JSON secret redaction, actual fake-Pi save-to-probe flow, mutation
  callback and invalidation behavior;
- real MyGo text binding, label clicks, HTTP field switching, password-masked
  text snapshots, filtering/detail selection, removal confirmation, nested
  Escape, no writes on cancel, and reachable modal footer at 800×560;
- stale workspace/executable results, shutdown, slow-probe interruption, and
  the distinction between a reload request and actual completion.

Run:

```sh
go test -race ./internal/mcpui ./internal/workbench ./internal/ui ./cmd/orbit
ORBIT_MCP_CAPTURE_DIR=/tmp/orbit-mcp-captures go test ./internal/mcpui -run TestNativeCaptureMCP
```

`TestNativeCaptureMCP` produces explicitly labeled native CPU rendering fixtures
for list/detail, stdio editor and HTTP editor. These renders were visually
inspected. They are not screenshots of a live MCP service and are not proof of
pixel-identical native GPU behavior on macOS/Windows.

## Remaining acceptance gates

Live OAuth/browser sign-in and real external stdio/HTTP server connection tests
were not run. The host still owns readiness-confirmed session restart and
multi-session/parked-process refresh. macOS/Windows native GPU comparison,
platform cancellation behavior and full application-level pixel diffs remain
acceptance gates. Extended Go-only labels use translation fallbacks; translating
new fallback-only strings into every locale is not claimed complete. The
source's top-bar MCP popup/status chip is a separate host surface and is not
implemented merely by mounting this settings page.

## Source-first acceptance checklist (current worktree)

Every source anchor below is in `crates/orbit-pi/src/app/mcp_ui.rs` unless a
backend source is named. The commands are independently runnable from the repo
root; fixture verification does not imply live-provider or OS acceptance.

| Case | Exact source handler | Executable verification / expected result | Functional status |
|---|---|---|---|
| Add stdio server | `mcp_add`, `open_mcp_editor`, `mcp_editor_save` | `go test ./internal/mcpui -run TestTypedDraftStdioAndSecretReferences`: command/args/env become typed config; one argument per line; literal values are replaced with scoped references. | Passed |
| Edit without losing fields | `mcp_edit`, `mcp_editor_save` | `go test ./internal/mcpui -run TestDraftMasksAndPreservesUnknownAndToolFields`: name stays immutable, untouched mask preserves secret, unknown fields/toolExposure/timeout survive. | Passed |
| HTTP + OAuth fields | `open_mcp_editor`, `mcp_editor_save`, `mcp_editor_layer` | `go test ./internal/mcpui -run 'TestHTTPDraftOAuthValidationAndTransportSwitch|TestNativeHTTPFieldsAndPasswordMasking'`: HTTP fields replace stdio controls; loopback URL/port validation rejects invalid values; no secret in rendered text. | Passed, fixture only |
| Configuration validation | `mcp_editor_save`, `mcp/config.rs::validate_def` | `go test ./internal/mcpui -run TestTypedDraftRejectsInvalidFields`; `go test ./internal/workbench -run 'TestMCPValidation|TestMCPTimeoutAcceptsTypedIntegers'`: malformed names, pairs, scope, port and timeout are rejected before writes. | Passed |
| Secret consent + persistence | `mcp_editor_save`, `mcp/config.rs::resolve_typed_secret` | `go test ./internal/mcpui -run TestSaveLiteralRequiresExplicitConfirmationAndUsesReference`: no secret/config write before confirmation; confirmed secret goes to private store; only reference enters Pi config; fake Pi reports actual fixture probe result. | Passed |
| Scoped secrets | `mcp/config.rs::generated_secret_name`, `secret_namespace` | `go test ./internal/mcpui -run TestGeneratedSecretsScopeNamespaceMatchesUpstream`: exact FNV-1a namespace; distinct global/project/workspace names. | Passed |
| Secret prompt cancellation | `mcp_editor_cancel`, `mcp_editor_close` (additional native prompt) | `go test ./internal/mcpui -run 'TestNativeSecretPromptRequiresConfirmationAndMasksValue|TestSecretConfirmationEscapeKeepsEditorAndDoesNotSave'`: secret is masked; top confirmation alone closes on first Escape; no write on cancellation; final dismissal clears draft. | Passed |
| Description sidecar | `mcp/manager.rs::McpManager::define`, `McpMetadata::key/set` | `go test ./internal/workbench -run TestMCPMetadataRoundtripDoesNotPollutePiConfig`: source-compatible sidecar key; description never enters Pi config; malformed metadata is not replaced. | Passed |
| Remove + shared secret safety | `mcp_request_remove`, `mcp_cancel_remove`, `mcp_confirm_remove`; `McpManager::remove` | `go test ./internal/mcpui -run TestNativeListSelectionFilterAndRemovalConfirmation`; `go test ./internal/workbench -run 'TestMCPPruningKeepsHiddenGlobalAndOtherNamespaces|TestMCPRemoveCleansDescriptionAndOrphansWithoutTouchingSharedSecret'`: Cancel leaves server; confirmed backend remove cleans description/orphans but preserves shared/hidden-global/other-project keys. | Passed |
| Status/probe truth | `spawn_mcp_probe`, `on_mcp_probe`, `mcp_test_server` | `go test ./internal/mcpui -run TestProbeParsesTruthAndUnknownStates`; `go test ./internal/workbench -run TestMCPProbeRetainsExitOneServerResults`: tools/resources/exposure/errors come from output; unknown state stays unknown; exit-1 valid JSON is shown while unrelated failures remain errors. | Passed, fake CLI |
| Error redaction and unreadable secrets | `on_mcp_probe`, `McpManager::redact_text`; `mcp/secrets.rs` | `go test ./internal/workbench -run 'TestMCPProbeRedactsEscapedSecretValues|TestMCPSecretStoreWrongShapeIsNotSilentlyReplaced|TestSecretsRedactionAndCorruptStore'`: escaped/multiline secrets never leak; malformed/null/non-string stores remain unchanged. | Passed |
| Codemode warning | `mcp_rows` | `go test ./internal/workbench -run TestMCPCodemodeProjectPrecedence`: unset remains nil; project explicit value wins; view warns for disabled codemode-exposed servers. | Persistence passed; warning render implemented |
| Busy probe + editor | `spawn_mcp_probe`, `mcp_add`, `mcp_edit` | `go test ./internal/mcpui -run TestSlowProbeDoesNotBlockEditingOrLeakItsResult`: Add stays usable during a slow probe; canceled generation cannot apply a stale result. | Passed |
| Workspace/close/runtime override | `mcp_section_opened`, `tick_mcp` | `go test ./internal/mcpui -run 'TestWorkspaceGenerationRejectsLateResults|TestCloseRejectsResultsAndRemovesSecretDraft|TestPiOverrideInvalidatesOldProbeGeneration'`: old workspace/executable results do not affect current state; shutdown clears drafts. | Passed |
| Safe apply acknowledgment | `mcp_config_changed`, `mcp_apply_now`, `mcp_after_settle`, `mcp_apply_clicked` | `go test ./internal/mcpui -run TestApplyRequestIsNotReportedAsCompleted`: request retains pending state; only explicit observed-success `Applied()` settles it. Host must separately verify no running/parked turn is interrupted and session restored. | UI contract passed; live host readiness pending |
| OAuth CLI mapping | `mcp_sign_in`, `mcp_sign_out`, `mcp_cancel_sign_in` | `go test ./internal/mcpui -run TestMCPOAuthCLIUsesExactPiCommands`: fake CLI receives login with 300-second timeout, logout, and trust-preserving probe argv. A real browser sign-in/cancel/logout remains to be exercised against an owned test MCP service. | Command mapping passed; live browser lifecycle not run |
| Keyboard/editing | `mcp_handle_key`, `mcp_editor_cancel`, `mcp_editor_layer` | `go test ./internal/mcpui -run 'TestMCPNativeKeyboardNavigationEditAndToggle|TestNativeFieldLabelsFocusTheirInputs|TestNativeEditorMinimumHeightAndCancel'`: Down selects next row, Enter opens it, Space persists enabled toggle, labels focus actual inputs, Escape cancels without write. | Passed |
| Native geometry | `settings.rs::settings_body/settings_header`; `mcp_toolbar`, `mcp_server_row`, `mcp_editor_layer` | `ORBIT_MCP_CAPTURE_DIR=/tmp/orbit-mcp-captures go test ./internal/mcpui -run TestNativeCaptureMCP`: inspect pinned title/toolbar/status, scope groups and 560px stdio/HTTP modal; footer stays reachable at minimum tested window. | CPU fixture visual inspection passed; pixel-identical OS GPU claim not made |
| Top-bar MCP | `mcp_status_chip`, `mcp_popup_panel`, `mcp_popup_row` | Open host top-bar MCP tab, compare statuses and navigation with settings; no implementation is inferred from this settings package. | Separate host acceptance, not completed here |

Visual evidence is the locally rendered native CPU fixtures, not an upstream
pixel-diff score. OS acceptance (macOS Metal, Windows GPU, native OAuth browser,
Windows descendant cleanup) remains separate and unverified in this cloud run.

### Exact source hashes tested

These SHA-256 hashes identify dirty worktree contents, **not a commit**. Later
host or backend changes require rerunning affected checks. Hashes were captured
after the MCP-focused race tests passed and no subsequent MCP edits are planned.

```text
c5b318fc3f90339202fd212a4dfd4f5002adad6fec8334d0051185da1419a764  internal/mcpui/state.go
6dc14b6bcf89488de90ea600c02d2aa963132c48de651679b0ecb70556d013c2  internal/mcpui/draft.go
1d275bf761f2a96a760dece97b7b99780c473ca0ff3f0f85a9ba1dc58b6e9d69  internal/mcpui/view.go
1a9c0b593a7fc441c2f7aaa72902dcbfda34daeaace29b1b079a2a16bc869397  internal/mcpui/editor.go
b544922e5f74d58994d0004a1d51d6b9b7f853c325022a7b0bc9c0f9b4041df4  internal/mcpui/state_test.go
2bb3909d30ec86e3ba03b568a4fe703f1ca15516605fc51c24693fd507f85a16  internal/mcpui/draft_test.go
73f4cb69b7a777a974c4f36c49725bd80712bc9ab9446b0736ac2b4f16653782  internal/workbench/mcp.go
4d0e167a5d782ff0a1bdf899a3d8742573f4a6123ba72d1ac52ca0e421387d8a  internal/workbench/mcp_parity_test.go
56c17ee90e6da552d3a0d0c94803051ae21fd876124a2a3b27003a29513bc03e  internal/workbench/service.go

```
