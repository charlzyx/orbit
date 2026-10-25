# Source test catalogue, limitations and acceptance gates

## Workbench interaction inventory (source-only)

Baseline: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `/workspace/shared/orbit-mygo-rpc`. Inspected `AGENT.md`, `.agents/skills/gpui/SKILL.md`, `.agents/skills/gpui-testing/SKILL.md`, and the source anchors below. No application launch, test execution, pixel comparison, Git mutation, or user's-computer access was performed. “Existing” means implemented in this source, not verified working at runtime and not migration parity PASS. Test evidence names source tests, not execution results. Unless otherwise stated, paths below are relative to `crates/orbit-pi/src/`.

### Coupling that the redesign must preserve

The workbench is a shared-controller system, not a group of detachable screenshots. `OrbitApp::sync_review_context` pushes workspace/session/turn/model into two Review instances, Git, Explorer and terminal. Workspace watches refresh several surfaces together. Git Changes embeds the same Review renderer with stage/unstage/discard callbacks; history/conflict rows can open Files; transcript summaries can open captured turn diffs. Files and terminals hold live buffers/processes independently of the currently visible page. A UI replacement must retain those actions and lifecycle boundaries, including keyboard precedence, error recovery and background operations.

### Source-backed limitations and migration verification risks

1. Documentation is not a feature oracle. `git_panel/mod.rs` header still says three tabs while `GitTab` defines six. `AGENT.md` links `docs/explorer.md` and `docs/worktrees.md`, absent at this HEAD, and has stale heartbeat/single-line-input guidance. Use source anchors and current handlers rather than importing those obsolete claims.
2. Current auxiliary generation is not RPC: W11 and W29 spawn one-shot `pi -p`. Their current behavior must remain, and any strict RPC-only requirement needs an explicit implementation/test plan for these flows.
3. Distinguish backend capability from exposed UI: History grep, Issue assignee filter/create assignees/attachments, PR list author/base filters, stash custom messages and worktree lock reasons exist in backend types/functions but are not exposed by corresponding audited Git/worktree controls. Do not promise them as working UI just because functions exist. The `gh::create_issue` attachment path may be used by the separate bug-report feature, outside this slice.
4. Existing read-only review does not provide rewind/restore; historical file menu Open diff currently opens live uncommitted Review and ignores selected hash. There is no native full PR patch/inline-comment review, three-way merge editor, LSP IDE feature set, shell-profile picker, split terminal, persistent terminal restart recovery, or Linux Open-in catalog in this source slice. These are not missing redesign parity if not present in the baseline; label additions separately.
5. Important source-quality risks are not demonstrated failures: Git async `spawn_data` has no workspace/request generation guard, unlike Review; checkpoint start is spawned detached and not awaited before sending prompt, and its error is discarded; closing Files tab starts save then drops visible state, so delayed failure has no tab to report in; autosave conflict badge does not pause save; per-path discard ignores errors; some worktree setup/preferences persistence errors are ignored. They need deliberate tests and truthful error design during migration, not assumed correctness or a demand to reproduce bugs.
6. `open_worktree` sets selected workspace but does not itself replace a currently live pi child. Existing cold-session open does spawn with its recorded cwd and separately refuses a missing cwd. Verify new-task/switch/session integration before claiming every active agent always reroots immediately.
7. Tests listed are inspected source evidence only. `.github/workflows/ci.yml` builds macOS/Linux/Windows, clippy on macOS, secret-material guard and dependency audit, but explicitly does NOT run tests. Some tests require git, platform APIs, GPUI visual test context or a real shell; worktree tests skip when git missing. No executed result or pixel-level parity claim is justified by this inventory.
8. Platform differences are genuine contracts: macOS Trash versus permanent Linux/Windows delete; Linux no Open-in catalog; Windows cmd setup-script behavior; contextual Cmd versus Ctrl+Shift terminal shortcuts; macOS windowless shell persistence. A single-platform build alone cannot establish these behaviors.


## Test evidence catalogue (source only, not execution results)

Tests below are current upstream regression evidence, not migration completion. `::tests` means inline Rust tests in that file; GPUI tests require a suitable build/runtime. See the JSON inventory for individual names and source lines.

| ID | Current test locations and scope |
|---|---|
| T01 | `main_window.rs::tests` shell reuse; `app/titlebar_layout_tests.rs`, `app/window_watch_tests.rs`, `app/backdrop_layout_tests.rs`, `app/new_task_panels_tests.rs`, `app/sidepane_full_width_tests.rs`: geometry/lifetime regressions. |
| T02 | `layout.rs::tests`: parse bounds/settled writes; related panel resize tests live in corresponding source modules. |
| T03 | `commands.rs::tests`, `command_palette.rs::tests`, `platform.rs::tests`; popup/key-routing regressions in `app/popup_layout_tests.rs`, `model_selector.rs`. |
| T04 | `composer.rs::tests`, `composer_send.rs::tests`, `mentions.rs::tests`, `attachments.rs::tests`, `app/composer_layout_tests.rs`. |
| T05 | `sessions.rs::tests`, `pins.rs::tests`, `archive.rs::tests`, `workspace_logo.rs::tests`, `workspace_mark.rs::tests`; `app/sidebar_{active_reveal,group,placeholder,remove,sort}_tests.rs`, sidebar inline tests. |
| T06 | `app/session_park_tests.rs`, `app/new_task_reconnect_tests.rs`, `app/send_progress_tests.rs`, `app/session_default_apply_tests.rs`, `app/refresh_tests.rs`, `app/error_label_tests.rs`. |
| T07 | `model_selector.rs::tests`, `model_selector_match.rs::tests`, `favorites.rs::tests`, `recent_models.rs::tests`, `session_defaults.rs::tests`, model-filter inline tests in `app/settings.rs`. |
| T08 | Large inline suites in `transcript.rs`, `transcript_view.rs`, `message_scroller.rs`, `highlight.rs`; include parsing, streaming, interrupted/failure behavior, selection, markdown, changed files, formatting and model-layer 10k-message performance, not a measured GPU scroll benchmark. |
| T09 | `context_meter.rs::tests`, `quota.rs::tests`, `quota_hidden.rs::tests`; `crates/orbit-rpc/tests/quota_fake_pi.rs`; `contrib/orbit-quota-extension/{index,adapters}.test.mjs`, `contrib/pi-quota-rpc/quota-handler.test.mjs`. |
| T10 | `dialog.rs::tests`, `ask.rs::tests`, `custom_ui.rs::tests`, `widgets.rs::tests`; `contrib/pi-custom-ui-rpc/custom-ui-handler.test.mjs`. |
| T11 | `access.rs::tests`, `workflow.rs::tests`, `auto_title.rs::tests`, `bundled_extensions.rs::tests`, `rpc_patches.rs::tests`; all guard/workflow/title `.test.mjs` files. |
| T12 | Inline `app/settings.rs` selector, theme catalog, model filtering and search tests. These do not establish click-through coverage of every setting. |
| T13 | `onboarding.rs::tests`, `pi_install.rs::tests`, `pi_update.rs::tests`; fake installer tests preserve user data. Live network/install detection tests may skip or depend on environment. |
| T14 | `providers.rs::tests` catalog/credential/config preservation; `auth.rs::tests` auth reducer; `crates/orbit-rpc/tests/auth_fake_pi.rs`. Live introspection requires installed pi. |
| T15 | `skills.rs::tests` discovery/precedence/overrides/deletion; `plugins.rs::tests` source parsing/install paths/update parsing; source-level UI does not imply real registry installation coverage. |
| T16 | `mcp.rs`, `mcp/config.rs`, `mcp/secrets.rs`, `mcp/manager.rs`, `app/mcp_ui.rs` inline suites: config merge, secrets/masking/redaction, parser, fingerprints/status/hints. Real OAuth/server integration still needs end-to-end runs. |
| T17 | `theme.rs::tests`, `theme/appearance.rs::tests`, `theme/tokens.rs`, `i18n.rs::tests`, `app/theme_switch_tests.rs`, `app/devicons_tests.rs`, `dither.rs`, `backdrop.rs`; locale key/placeholder script `scripts/check_i18n.py`. |
| T18 | `usage/tests.rs` plus inline `usage/page.rs`, aggregate/model/collect/chart/format/heatmap suites: accounting, indexing, date/filter/query/sort/pagination/export/preferences and geometry. |
| T19 | `notifications.rs::tests`, `toast.rs::tests`; system notification delivery/click routing requires actual packaged platform testing. |
| T20 | `session_retention.rs::tests`, `temp_files.rs::tests`: age/liveness/deletion guards and malformed settings. |
| T21 | `updater.rs::tests`, `pi_update.rs::tests`, `pi_install.rs::tests`, `rpc_patches.rs::tests`; signed archive/path validation and helper tests are not proof of successful install on every target. |
| T22 | `bug_report.rs::tests`, `issue_message.rs::tests`; generated-body/URL/fallback parsing, not actual remote issue creation. |
| T23 | `crates/orbit-rpc/src/{client,types}.rs`, `crates/orbit-rpc/tests/{spawn_args,auth_fake_pi,quota_fake_pi,live_pi,live_agent_control,live_catalog,live_model_match,live_switch_probe}.rs`: transport/control contract. |

All unqualified source paths in this catalogue are under `crates/orbit-pi/src/`. `.github/workflows/ci.yml` builds on macOS/Linux/Windows and runs clippy on macOS, but does not currently run the full test suite. “CI green” therefore cannot substitute for parity tests.

## Keyboard contract that must not be lost

The authoritative complete tables are `crates/orbit-pi/src/commands.rs::BINDINGS` and `crates/orbit-pi/src/main.rs::bind_keys`; canonical platform strings are `platform.rs::shortcuts`. Preserve the actual registered contexts, not just labels copied from AGENT.md.

- Global: Primary+N new session; Primary+R refresh; Primary+P/K palette; Primary+, settings; Primary+/ shortcut help; Primary+B sidebar; Primary+Shift+B focus sessions; Primary+1…9 open visible session; Primary+F transcript find; Primary+U usage; Primary+Up/Down turn navigation; Primary+Shift+C copy last response; Primary+Shift+M model; Primary+Shift+T thinking; Primary+Shift+R Review; Primary+Shift+G Git; Primary+J terminal; Primary+Shift+E Explorer; Primary+Shift+U updates; Primary+Q quit.
- Composer: Enter and Primary+Enter submit; Shift+Enter newline; Alt+Enter alternate running-send mode; Primary+Shift+Enter explicit steer; Tab accepts completion before traversing; Escape and Primary+. abort only after higher-priority surfaces are handled. Preserve IME candidate composition.
- Text fields/editor: native copy/cut/paste/select-all, grapheme and word/line selections, undo/redo, Home/End and extended selection, editor tab indentation and literal context-specific bindings. Do not globally map these to agent commands.
- Popovers: arrows, Enter, Escape; click-out dismiss; filter input retains text editing, hover must not undo keyboard-highlight pin. Context registration order is part of current behavior.
- Terminal: Tab/Escape go to shell rather than traversal/abort; Mac Primary+T/W create/close tab; other OS Primary+Shift+T/W to avoid shell Ctrl+T/W; Primary+{ and Primary+} switch tabs (actual shifted punctuation event).
- Sidebar: arrows/Home/End, Enter activate, Escape return composer; Review tree: arrows, Space/Enter toggle as bound, n/p files, ]/[ hunks, e/c expand/collapse, Escape return focus.
- Files: Primary+W closes active tab (last closes surface), Primary+Shift+W closes Files, Primary+S saves active editor immediately.
- Git and questionnaire/dialog-specific bindings must follow their local contexts; a broad global shortcut handler that bypasses these is not equivalent.

## Persistence and external-state acceptance

Preserve (or explicitly migrate) all existing state formats, including bad/missing-file defaults and scope precedence. A new application must not overwrite user-owned unknown fields just because it cannot parse them.

- pi: `~/.pi/agent/sessions/**.jsonl`, `~/.pi/agent/{settings,models,auth,mcp}.json`, project `.pi/{settings,mcp}.json`, resources/package installations and credentials owned by pi.
- Orbit: `~/.orbit-pi/{layout,ui,fonts,workspaces,pinned-sessions,archived-sessions,favorites,recent-models,session-defaults,auto-title,access,access-allow,workflow,notifications,hidden-quota-providers,usage,session-retention,plugin-update-skips,mcp-secrets,mcp-metadata}.json`, theme/backdrop/hint settings, RPC-patch config/state/materialized scripts, bundled extension files, worktree settings/trust.
- Repository: `.orbit/worktree.json`, optional setup script and `refs/orbit/...` checkpoint refs. Name/branch/path are separate worktree identities. Never substitute branch names for worktree names.
- Ephemeral state: queues, pending dialogs/approval, streaming/transcript folding/selection, terminal PTYs, editor buffers/undo/autosave, busy/error indicators, in-flight Git/setup/update operations and parked process fingerprints need explicit lifetime rules. They are not all safe to persist or restore blindly.
- Mixed stores have different failure policy. Some upstream small preferences are best-effort writes, while credential/provider/MCP config operations reject corruption and preserve unknown fields. Do not flatten these into one generic “save successful” response.

## Shipped limitations versus new work

These are current baseline boundaries, not licenses to remove existing functionality:

1. Rewind/fork RPC types exist, but rewind/fork UI is not shipped; clone is shipped. Do not count a typed command as a completed UI feature.
2. Mermaid fences are labelled source, not diagrams. Screen-reader/focus accessibility and on-device scrolling benchmarks remain incomplete.
3. Access/workflow guards are extension-level policies, not an OS sandbox. Default access is Full access; invalid/missing access config also falls back to Full access, despite a misleading comment implying it cannot widen access. Treat the code as the baseline and record any intentional safety change separately.
4. Custom extension UI paints terminal frames and supports legacy keyboard input; Kitty protocol/full mouse fidelity and terminal `setTitle` are not equivalent native implementations.
5. Desktop notification/banner sound implementation is macOS-specific; Linux open-in app discovery is empty in this HEAD. Explorer deletion uses macOS Trash but permanent deletion on Windows/Linux. A new UI must not claim cross-platform availability solely because code compiles.
6. Runtime auth/quota/custom UI can depend on installed-pi patch compatibility; failure degrades capabilities explicitly. Do not imply unmodified upstream pi provides every extension namespace.
7. Main conversation uses RPC, but AI commit/issue/PR drafting currently uses isolated tool-free `pi -p --no-session` children. A requirement that **all** agent invocation use RPC introduces real migration work here.
8. Review checkpoint failures, partial usage/pricing coverage, absent historical retries and unsupported providers must stay visibly unknown/unavailable instead of becoming dummy success/data.

## Migration and verification gates

Confirmed direction: **Go + MyGo native reconstruction**, not a Rust restyling. Existing Rust source is a frozen behavioral reference; it is not the Go implementation. Every row remains pending until the Go/native path is wired and tested.

1. Trace each row to its new host API, UI state/action handlers, settings ownership, and platform capability branch. Domain-only code is not completed interaction parity.
2. Replay representative real/fake pi event transcripts through both reducers, including first-send latency, queue changes, failure/abort/retry/compaction, process death, restart and parallel-session switches. Compare semantic state and command wire payloads.
3. Exercise filesystem/Git/worktree operations in temporary repositories only. Cover dirty/locked/in-use worktrees, conflicting names, missing files, invalid/binary/large text, repeated actions, cancellation and failures. Remote push/issue creation must not occur in tests without explicit isolated fixtures.
4. Add end-to-end keyboard/pointer coverage for every modal/menu/surface context, focus restoration, outside-click dismissal, double/repeated actions and disabled states. Test original data migration and malformed-file behavior.
5. Capture screenshots from both actual apps at fixed viewport, OS scale, fonts, locale, theme, density and seeded state. Compare requested visual target across empty/loading/active/error/modal/narrow/wide states. Source-derived geometry tests alone are not pixel parity.
6. Test packaged native lifecycle on declared supported OS/architecture targets: launch/reopen/quit, PTY lifetime, clipboard/IME, system permissions, notifications, file picker, updater signature rejection/install/rollback. Cloud-only development does not authorize use of the user's Mac.
7. Gate completion on every feature row, not an MVP subset. Track “source mapped”, “Go domain implemented”, “UI wired”, “tests executed”, “platform verified” and “visual verified” separately. No unfinished/unsupported row may silently be marked PASS.

## Audit caveats and source/document drift

- AGENT.md describes a fixed/window heartbeat in older sections, but `main_window.rs` is app-owned and adaptive 16/90 ms. It also contains older single-line/IME wording contradicted by current multiline composer source.
- Git documentation headings say three tabs while the implementation has six. Bundled extension header says three but lists/loads four. `docs/explorer.md` and `docs/worktrees.md` referenced in AGENT.md are absent at this revision.
- This is a full source-derived product-surface inventory, not a formal proof of every possible internal branch. The machine inventory makes omitted modules visible and the acceptance gates forbid claiming completion from this audit alone.
