> New source-first acceptance entry point: [initial executable checklist and explicit unexpanded gates](acceptance-checklist.md). This inventory remains supporting evidence, not a completion claim.

# Orbit → mygo full-feature parity baseline

Baseline: `imrj05/orbit` commit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, inspected in the cloud on 2026-10-25. No user computer was used.

This is the preservation contract for **all current Orbit functionality**, not an MVP proposal. It inventories implementation and interaction behavior from source, with current test evidence. The original source inventory is not an implementation claim. The Go evidence column and catalogue below now record specific reconstructed code and passing test scopes; **every full-parity acceptance gate remains open**. No full-feature or pixel-equivalence claim is made. Source presence and domain tests are not end-to-end validation. The compact companion `feature-parity.json` indexes stable category JSON files, expanded source paths and the mechanical file/test inventory. Generate the complete download with `python3 docs/mygo-migration/build-parity.py`.

## Go reconstruction evidence

Snapshot: **2026-10-25 04:35 UTC**, cloud-only, uncommitted working tree. “Partial” means the stated code/test scope exists, not that the entire baseline row is complete. The indexed JSON categories preserve all **139 individual rows**, each with `full_parity_verified: false`, paths, test files and an explicit remaining condition.

| Evidence | Implemented scope | Executed evidence and limits |
|---|---|---|
| G1 | `internal/workspace/{files,editor,tree,watch}.go`; native Files adapter | Race tests pass for create/no-clobber, limits/binary/lossy/read-only, global 500 ms last-edit autosave/all-dirty flush, atomic mode-preserving saves, external-change conflict flags, undo/redo, rename/close, failed-save retention, nested/ancestor ignores and `.ignore` precedence, watcher cancellation, empty-directory additions/removal and checkpoint-loop exclusions. Ordinary workspace `logs`/`objects` directories are not mistaken for Git internals. Native recursive/debounced watching now uses pinned fsnotify 1.10.1; resource/backend/overflow failure closes it and enables bounded polling. Native tests detect same-size/mtime-preserved writes, new recursive directories, linked/private/common Git metadata and Git initialization after startup. Poll fallback cannot detect metadata-preserving writes; NFS/SMB/FUSE delivery is not guaranteed. Windows/macOS workspace test binaries cross-compile, but only Linux watcher execution has been exercised. Complete explorer/editor keyboard/visual behavior is not established. |
| G2 | `internal/workspace/{git,git_ops,github}.go`; `internal/gitworkbench` | Race tests pass with temporary repositories/local bare remotes, fake `gh` runners, six-tab render/controller guards, mutation serialization, generation rejection and confirmation behavior. Unix cancellation now kills Git/gh descendant processes; a temporary-shell test confirms they cannot write later after cancellation. Windows process-tree cancellation remains a separate platform gate. No external remote writes or authenticated GitHub mutation tests were performed. Tool-free Pi RPC generation and repository issue/PR templates now have fake-process tests for argument isolation, 48 KiB/file and 192 KiB total original-code bounds, 96 KiB diffs, cancellation/fallback and no index mutation. Avatar identity overrides/async cached photos and embedded grouped diff controls now exist with scoped tests. Live draft workflow, exact native keyboard/timing, syntax paint, localization and pixel geometry remain separate acceptance work. |
| G3 | `internal/workspace/{checkpoint,review}.go`; agent lifecycle hooks; native Review adapter | Race tests pass for all six sources, immutable last-turn snapshots, switched/new/unborn branches, unchanged real index/HEAD, root-commit and literal-path review, grouped/split/binary models, compact fallback gaps, stable expansion and 50k visible-line cap. Native full-width/sticky headers/hunk keyboard/collapse/syntax paint and pixel parity remain incomplete or unverified. |
| G4 | `internal/workspace/{worktrees,setup}.go`; `internal/worktreeui` | Race tests pass for create/rename/move/lock/unlock/remove/prune/repair in disposable repositories, setup consent before creation and canonical trust persistence, relative moves resolved from main root and rename targets under configured root, failure preserves tree, cancellation, branch preservation, cross-manager locks, stale UI-result rejection, failed-action persistence, operation-specific retry, transactional settings persistence, native Escape/confirm and settings-cancel behavior. Quick-create slug defaults and auto-derived `orbit/` branches stop after a manual edit; new/existing branch choices stay separate, and checked-out branches offer an explicit open-existing dialog. The native Work-in picker is mounted and has derived-row/native-menu selection/dismissal tests. The source guide-dismiss preference and translator are wired by the host; source-key labels and interpolation have scoped tests, but extra labels remain English. Actual OS keyboard/focus behavior, complete locale coverage and packaged-platform acceptance remain separate gates. |
| G5 | `internal/pirpc`, `internal/agent` | Package race suites pass, covering fake-Pi transport/state/session/dialog/environment/parking and lifecycle contracts. This does not imply live model-provider inference, OAuth, custom terminal UI, image workflows or every bundled extension UI is verified. Optional credential-free live tests are separate. |
| G6 | `internal/preferences`, `internal/workbench`, `internal/native` | Package race suites pass for persistence/discovery/accounting/native helper/security/update/temp-cleanup domain tests. UI wiring, remote auth/install flows and cross-platform packaged behavior must be independently exercised. Signed Go release feed/key/artifacts are not configured. |
| G7 | `internal/terminal` | Package race tests pass. Presence of optional live PTY tests is not evidence they ran; host shell/clipboard/IME/selection/mouse/platform packaging acceptance remains separate. |
| G8 | `internal/ui`, `cmd/orbit/main.go` | Native shell/settings/chat/files/review/controllers exist. An earlier aggregate run found an in-progress `parser.BlockQuote` compile error; it was corrected. The 04:34 UTC workspace/worktreeui/gitworkbench/ui race rerun and workspace/worktreeui vet checks passed. These scoped tests still do not establish all contexts, native-platform behavior or pixel equivalence. Screenshot fixtures are synthetic test content, not live agent results or baseline pixel comparisons. |

### Prioritized integration findings

- Deliberate safety differences remain explicit: Go retains dirty tabs after close-time save failure rather than dropping them, and setup trust persistence must succeed before the approved create continues rather than best-effort persistence. Native-watch backend failure uses bounded polling rather than only manual refresh. These are not claimed as byte-for-byte source failure behavior.
- The later audit found collapsed/split Review navigation errors, cached-review error handling, stale file-open activation, dirty-file destructive overrides, unbounded truncated viewers, and missing relative-path copying. Corrections are present; `internal/ui/viewer_parity_test.go` adds five focused regressions. These are narrow fixes, not complete keyboard/pixel acceptance.
- The audit identified cross-session workspace drift that could route Git actions to the previous repository, wrong last-turn inference from counting visible user messages, and stale Review results after source/workspace changes. The fixes are present. New `internal/ui/workspace_review_regression_test.go` cases pass under `-race` for rapid hidden/source/session/workspace changes and persisted turn counters after transcript compaction. This closes those narrow regressions, not the whole Review/Explorer rows.
- File watcher fan-out, safe dirty-tab warnings, active rename/clone action IDs, commit-ref/path Review callbacks, `gh` authentication routing and all-dirty manual saving were explicitly sent to the UI owner. A domain API alone does not close these rows.
- Preferences now normalize the source `last_updated` workspace-sort key rather than the incompatible `last_activity`; legacy filesystem creation-time behavior and the full sidebar sort/marks UI still need acceptance.
- Hard-coded English labels, incomplete contextual keyboard/focus behavior, backdrop reconstruction, avatar presentation and exact Work-in native-platform behavior are known user-visible gaps. They are not hidden behind “all features implemented.”
- Required environment validation remains: real cloud Linux app exercise, macOS/Windows builds/packaging and native system integration, credential-free live Pi probes, then explicitly authorized provider/GitHub flows. No user's Mac was accessed.

## Scope and architecture findings

- App source: 148 Rust files / 149,251 lines in `crates/orbit-pi/src`; transport: 3 files / 3,245 lines in `crates/orbit-rpc/src`. Combined: **152,496 lines**. The 8 RPC integration test files add 1,092 lines; 23 JS/MJS extension/patch files add 6,970 lines. These are physical lines, including comments and tests, not production-only SLOC.
- Static declarations: 981 ordinary Rust tests and 148 GPUI tests in the two `src` trees, plus 13 ordinary Rust integration tests. These counts are not a pass count, unique test-case count after expansion, or code coverage. Live tests can skip without prerequisites.
- The UI is not an independent thin view: `OrbitApp` owns live/parked pi processes, queues, transcript, editor, terminal, Git/Review, settings, watchers, notifications and updater coordination. Feature modules reach shared private fields. `ComposerInput`, GPUI entity/focus lifetime, anchored popovers, virtualized lists and paint-only text are common infrastructure across nearly every surface.
- pi is already an external child process using JSONL over stdio. Replacing the renderer does not require replacing the agent runtime. It does require preserving host behavior around native filesystem access, PTYs, git/gh CLIs, background tasks, dialogs, security-sensitive credentials and process lifetime. A browser-only mygo frontend cannot own those features by itself without a native host/bridge.
- The user subsequently confirmed **Go + MyGo native reconstruction**, not a Rust skin. This is an intentional replacement of the existing Rust/GPUI architecture. The baseline still preserves Orbit behavior; changing implementation language does not waive native features or test gates.
- Pixel parity must have an explicit target. Preserving Orbit's layout/interaction semantics while applying mygo visual design is not the same as reproducing Orbit's original pixels. No pixel-comparison evidence was produced in this source audit.

## Reading the matrix

`Primary` = Cmd on macOS, Ctrl on Windows/Linux. Source references are repository-relative; `::symbol` is a search anchor. Test IDs refer to the test-evidence catalogue below. A row marked source-present includes its supported/degraded states, not a guarantee that the feature works on every platform. Migration acceptance must verify pointer, keyboard, persistence, loading/empty/error/cancel and reconnect behavior, not merely the happy-path screenshot.


## Full preservation contract by category

Every category below is part of the all-feature contract. Nothing is an optional MVP slice.

| Category | Rows | Human-readable contract | Structured records |
|---|---|---|---|
| Shell, navigation, sessions and workspaces | S01–S14 | [Read](parity/shell.md) | [JSON](parity/shell.json) |
| Composer, transcript and interactive agent surfaces | C01–C20 | [Read](parity/chat.md) | [JSON](parity/chat.json) |
| Settings, providers, skills, packages and MCP | P01–P20 | [Read](parity/settings.md) | [JSON](parity/settings.json) |
| Usage analytics, update, privacy and release behavior | U01–U14 | [Read](parity/updates-privacy.md) | [JSON](parity/updates-privacy.json) |
| Workbench lifecycle and shared state | W01–W03 | [Read](parity/workbench-lifecycle.md) | [JSON](parity/workbench-lifecycle.json) |
| Git changes, history, graph and repository operations | W04–W19 | [Read](parity/git-changes-history.md) | [JSON](parity/git-changes-history.json) |
| GitHub issues and pull requests | W20–W30 | [Read](parity/github.md) | [JSON](parity/github.json) |
| Review and turn checkpoints | W31–W39 | [Read](parity/review.md) | [JSON](parity/review.json) |
| Worktrees and setup lifecycle | W40–W51 | [Read](parity/worktrees.md) | [JSON](parity/worktrees.json) |
| Explorer and file editor | W52–W60 | [Read](parity/files.md) | [JSON](parity/files.json) |
| Terminal, platform and window lifetime | W61–W71 | [Read](parity/terminal-platform.md) | [JSON](parity/terminal-platform.json) |

[Source test catalogue, keyboard/persistence requirements, shipped limitations, and acceptance gates](parity/acceptance-and-source-tests.md).

The versioned index also links the complete source inventory and exact verification command outcomes. `build-parity.py` verifies unique IDs and reconstructs the monolithic JSON/Markdown under `generated/`; generated files are not needed in commits. No full-parity status is inferred by assembly.

### Scoped update: paint-only syntax (04:51 UTC)

The independent Go lexical contract now matches 1,204 executable pinned-source cases across 16 languages, with race-tested review carries/context retention and Git color-only spans. This does not prove whole-row, full parser, transcript/editor or pixel parity. [Evidence, reproduction and provenance caveat](highlight-parity.md). All 139 full-parity flags remain false.
