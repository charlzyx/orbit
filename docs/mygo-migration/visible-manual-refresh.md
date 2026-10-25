# Explicit Refresh: visible read-side models

Scope: pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac` versus Go wave5 base `3fddcce`; this isolated increment includes the approved retry receipt test correction in its baseline, not its patch.

## Source contract and correction

Rust `crates/orbit-pi/src/app/session.rs:2702–2740` refreshes sessions, Explorer, visible Review and Usage, Git/worktrees, and the current MCP/Plugins settings section. Go `internal/ui/run.go` previously refreshed sessions/files/Git/worktrees only. The new `visible_refresh.go` dispatches the missing visible models through their existing asynchronous admission, coalescing and owner guards.

Visibility follows Go's actual paint branches (`app.go:472–480` and `main`): Settings renders before setup gating; Usage renders its own page; a retained `ReviewVisible` flag does not imply the pane is painted behind Settings, Usage or onboarding. No hidden settings categories are refreshed.

* Review: existing `review.go:71` generation/workspace handling rejects previous-workspace completions.
* Usage: existing `usage.go:161` refresh is a global data-store scan, not a workspace-owned list. Repeated requests coalesce; failures retain the last good dataset and do not navigate back to the requesting page.
* MCP: Rust `app/mcp_ui.rs:309–315` rereads configuration and schedules an immediate probe. The existing Go MCP Refresh hook intentionally retains that read-side probe. It may use subprocesses/network; this is **not** a claim of a network-free Refresh. It does not Apply/restart the agent or grant credentials.
* Plugins: Rust `app/settings.rs:7455–7461` calls only `plugins::discover` (`plugins.rs:99–132`). Go uses `Resources.Refresh` (`resourceui/state.go:333`) to read the local resource snapshot. It deliberately does **not** use `RefreshPlugins` (`:331`), which schedules `checkPending`/`CheckUpdates`. Existing independent startup/update scheduling is unchanged. No install, upgrade, or upstream metadata check is added by this command.

Rust's later agent auth/catalog/context requests are outside this bounded increment: no new agent RPC is introduced, and this change does not claim complete Refresh parity. Existing filesystem watchers and Usage fingerprint refresh are not equivalent to an explicit command.

## Evidence

* Actual mounted Cmd/Ctrl+R and palette selection use the real Controller action and synthetic MCP hook. Before production hooks both routes remained at zero calls (`visible-refresh-red-2002.log`). No real MCP endpoint or token is used.
* Page matrix checks current visible page versus retained dock flags, including Settings ahead of blocked setup; Plugins must use its actual local State admission and must not call the old update-check hook.
* Actual resource State tests cover generation/pending coalescing, unchanged CheckedAt, false checkPending/Checking, and replacement-workspace snapshot ownership. Independent startup timer is explicitly disabled in these synthetic fixtures so it cannot confound the command's side-effect assertion.
* Usage synthetic held scan/failure preserves global last good data across newer navigation; temporary Git workspaces verify only the newer Review result is adopted.
* Final focused race, two repetitions: UI 6.167s and resourceui 2.034s PASS; vet PASS. Logs: `/workspace/shared/orbit-toolchains-continued/visible-refresh-strict-focused-2007.log` and `visible-refresh-strict-vet-2007.log`.

Focused evidence is not a full package/native macOS/pixel parity pass. Full integration gates remain separate. No user filesystem or external service was exercised.
