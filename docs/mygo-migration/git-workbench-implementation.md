# Native Git workbench implementation and remaining parity

Source contract: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, principally `crates/orbit-pi/src/git.rs`, `git_ops.rs`, `gh.rs`, `gh_templates.rs`, `commit_message.rs`, `issue_message.rs`, `bug_report.rs`, `git_panel/mod.rs`, and `git_panel/failure.rs`. This note describes Go code and targeted test evidence, not a full-parity or pixel-parity acceptance result. See W04–W29 in `feature-parity.md` for the original behavior inventory.

## Implemented backend

- `internal/workspace/git.go`: cancellable Git CLI runner with typed stdout/stderr/exit failures, optional index locks disabled; NUL-delimited status/rename/numstat parsing; literal pathspec handling; staged/unstaged counts; unborn/detached handling; stage/unstage/discard/commit; branch checkout/create and remote tracking; ahead/behind, fetch, ff-only pull, merge/rebase upstream, publish/push, force-with-lease and sync; filtered history, commit metadata/files, graph records and lane layout; known-forge commit URLs. `Git.Run` is exported for other workspace services.
- `git_ops.go`: worktree-aware in-progress merge/rebase/cherry-pick/revert/bisect detection, conflict paths, guarded continue/skip/abort, branch refs/rename/safe or explicit force deletion, regular/no-ff/squash merge, rebase/autostash, stash create/apply/pop/drop, classified actionable failures.
- `github.go`: GitHub CLI discovery, version/auth probe, identity, typed issues/pulls/details/comments/labels/checks/reviews/files/commits, filters and create/comment/state/label/review/merge/checkout operations. Invalid arguments fail before dispatch. PR checkout refuses dirty or in-progress repositories. Refresh performs reads only.
- `github_templates.go`: sorted repository issue/PR template lookup, legacy paths, Markdown front matter, source-style limited issue-form YAML conversion, title/label prefill, 256 KiB and UTF-8 limits, outside-symlink rejection.
- `avatars.go`: email overrides from `~/.orbit-pi/git-avatars.json`, authoritative GitHub noreply login, handle-shaped-name fallback, public profile image URL, stable initials/tint. `internal/gitworkbench/avatars.go` supplies bounded asynchronous image loading and fallback caching.
- `generation.go` and `report_generation.go`: editable commit, issue/PR and app-report drafts. Commit context includes bounded original code, per-file patches, repository/branch and recent repository/author commit style. Draft context includes notes, selected template, commits and diff statistics. Report generation samples **only supplied notes**, never repository files, Git history, conversation, attachments or diagnostics.

### Public integration points

`gitworkbench.New(workspace, Options)` returns a controller; `View(c, state)` returns a native MyGo element. `SetWorkspace`, `Refresh`, `Tick`, `Cancel`, `Close`, `SetTab`, `RunGit`, `RunGitHub`, `SelectCommit`, `SelectIssue`, `SelectPull` and `CloseTransient` are host entry points. The host routes Escape through `CloseTransient` before leaving the page.

Options include thread-safe invalidation (`OnChange` or `Invalidate`), `OpenFile`, `RevealFile`, `OpenReview(source, ref, path)`, `OpenURL`, optional attachment selection, avatar loader/home, and generation callbacks. Generation callbacks must be captured for an immutable workspace and updated on workspace switches. Workers capture callbacks before dispatch; UI state is applied from a result queue on `Tick`. Stale generations are dropped. Workspace changes cancel old work and clear pending confirmations, drafts, details and filters so a stale action cannot target the new repository.

Generation signatures:

- `GenerateCommitRPC(ctx, Git, GenerationConfig, stagedOnly) (GeneratedText, error)`
- `GenerateDraftRPC(ctx, Git, GenerationConfig, DraftRequest) (Draft, error)`
- `GenerateReportDraftRPC(ctx, GenerationConfig, kind, notes) (Draft, error)` where kind is `bug`, `feature` or `other`.

`GeneratedText` and `Draft` carry `Fallback` and `Warning`; callers must display the warning instead of implying a model succeeded. Explicit caller cancellation returns an error. Helpers do not commit, push, create an issue or submit a report.

### Intentional Pi transport migration

The original auxiliary generators invoke `pi -p`. The Go implementation intentionally uses a **separate** `pi --mode rpc` process because the requested architecture uses Pi RPC. Each generator disables tools, skills, context files, project trust and session persistence; supplies an explicit system prompt; optionally selects the active provider/model with `set_model`; sends `prompt`; waits for `agent_end`; reads `get_last_assistant_text`; and always closes/reaps the process. Default deadline is 180 seconds. Extensions remain enabled for custom provider resolution, as in the source helper. Interactive extension requests are canceled rather than silently approved. The active chat process/session is never reused. Missing Pi/auth/model, malformed output and helper timeout produce an explicitly marked editable heuristic, not a successful-model claim.

## Implemented native surfaces

All six source tabs are present: **Changes, History, Graph, Stashes, Issues, Pull requests**. Worktrees remains a separate host surface.

- Changes: staged/unstaged groups, individual/bulk stage and unstage, explicit discard confirmation, actual parsed embedded diff, unified/split, wrap, expand/collapse files and unchanged gaps, path/folder filtering, file-context actions and Review routing. Commit controls follow repository state, support include-unstaged, blank-message generation and staged-only/stage-all choice. Commit-and-push reports partial success accurately and does not recommit on retry.
- History/Graph: current/all branch options, author/path/follow filters, initial 80 and +40 Load more, real graph lane/edge/node painting, refs and author identity, expandable **inline** commit detail, 300-file cap, copy/forge links, current-file/reveal/path-copy/file-history and Review actions.
- Stashes: real Git stack, create with optional message/untracked choice, apply/pop and confirmed drop.
- Issues/Pulls: honest CLI setup states, filters and fetched limit 50; details replace the list, with return-to-list controls, native Markdown bodies/comments/reviews, metadata, CI check status/links, files and commits; edit/comment/state/review/merge/checkout operations with real results. Creation forms support notes, templates, editable generation, required title, draft/base/labels, selected-base defaults and created-item detail opening. A late template does not overwrite text already edited.
- Persistent full error detail is copyable and survives refresh; recovery actions are classification-specific. In-progress operation controls and up to eight conflict-file links use real repository state. Force push, discard, abort, skip, branch destruction, PR merge/checkout and stash drop have explicit UI confirmation.
- Asynchronous operations do not mutate UI state from workers. Duplicate actions are gated. Changing tabs does not cancel an already-dispatched mutation; it schedules a refresh for the selected tab. A successful mutation followed by refresh failure cannot be replayed by Retry. Keyboard routes include contextual Cmd/Ctrl+1…6, commit Cmd/Ctrl+Enter while its text area has focus, and transient Escape handling.

## Evidence actually exercised

Cloud Linux, Go toolchain in the workspace. Most recent targeted command:

`go test -race ./internal/gitworkbench ./internal/workspace -count=1`

Both packages passed on 2026-10-25. These tests include other workspace services owned by the broader migration, not only this Git implementation.

- `git_test.go`: **real temporary Git repositories**; newborn index, literal/unusual filenames, staged/unstaged preservation, rename parsing, branches, history and graph layout, stash round trips, merge/rebase/cherry-pick/bisect recovery, stale-operation guards, local-bare-remote fetch/pull/push/publish/sync/checkout, errors and cancellation. Local bare repositories only; no external remote write.
- `github_test.go`: **fake command execution** for JSON/probe/version/mutation argument and failure assertions; local Git for dirty/operation checkout guards. This does not establish compatibility with a live authenticated GitHub account.
- `generation_test.go`: **fake Pi JSONL child process**, selected-model and no-tools/no-session/RPC flags, normalized response, fallback, timeout/cancellation, original/diff context budgets, no index changes, report notes-only prompt and Unicode title fallback. This is not real model inference or credential validation.
- `avatars_test.go`: identity resolution and override/fallback behavior. `parity_test.go` uses a generated in-memory PNG with a fake avatar loader; no public avatar network fetch in tests.
- `state_test.go`, `parity_test.go`: real temporary Git controller operations, stale result suppression, error retention, mutation refresh failures, duplicate/cancel behavior, all six native element trees, confirmation cancel/confirm, form title validation, editable/template/generation flows, actual split/unified patch text, inline/full-width detail routing, transient Escape, and workspace-switch pending-action reset.

No original Rust application launch, cross-application pixel comparison, interactive macOS/Windows run, live GitHub write, live Pi model generation, or real platform file-manager interaction is established by this evidence.

## Exact remaining gaps and acceptance work

These are **not** full-parity rows:

1. **Localization and visual treatment (W04–W29):** Git UI labels are currently English and not wired to the shell locale catalog. Native spacing, fonts, icons, adaptive toolbar wrapping, hover affordances, filter-chip styling, branch/ref picker anchoring/group headings and create-form presentation still differ. Issue/PR creation is a modal, whereas source creation occupies the tab. History details and issue/PR details have been moved to the source's inline/full-tab structure, but no pixel acceptance has been performed.
2. **Embedded Review fidelity (W08, W30+):** real diff text and actions exist, but paint-only syntax-token coloring and exact shared Review toolbar/layout-preference propagation remain. The local native tree is not the source's complete shared Review renderer. Path selection/scroll position, source-specific responsive layout and all theme/scale combinations need integration QA.
3. **Timing/loading boundaries (W04, W06, W18, W20):** one controller busy state replaces source per-tab loading state. Four-second visible Git refresh is implemented, but source minimum 650 ms manual-refresh feedback and 200 ms picker-dismiss guard are not mirrored. GitHub identity is first probed on a GitHub tab, rather than on initial Git page entry. Live watcher integration belongs to the host.
4. **Identity/metadata styling (W19, W21, W25):** identity resolution and photos are implemented, but fallback hue palette/size are not yet the original FNV64/eight-hue/14px treatment. GitHub bodies/comments use login-derived photos rather than preferring every explicit event avatar URL. Some relative-time/updated/comment metadata and label-color rendering differ from source.
5. **Context and navigation polish (W06, W07, W16, W17, W28):** per-file context actions exist and host reveal callback is exposed. Native menus/dialogs do not yet reproduce all source popover focus/dismiss behavior. Safe-delete failure is surfaced, with a separate confirmed Force delete choice; the original automatic escalation prompt after an unmerged error is not mirrored. Parent-commit navigation outside the current loaded window and every scroll-restoration case need verification. PR base selection currently uses a text field, not the anchored branch picker.
6. **Graph acceptance (W18):** real lane layout handles octopus merges/window boundaries and draws edges, but original lane geometry, theme colors, selected-row expansion continuity and actual render screenshots have not been compared. There is not yet a dedicated local-remote-only/all-refs regression or dedicated stale force-with-lease/revert recovery regression in the Go test set, despite implemented backend support.
7. **Platform/service acceptance:** Windows command-hide build code exists but was not run here; macOS/Windows shell/reveal/clipboard/dialog behavior remains host/platform testing. GitHub CLI JSON and flags have mocked coverage, not a live service acceptance run. Source's advisory gh 2.20 minimum does not prove all modern flags work with every such version. Real configured Pi inference remains untested.
8. **Safe intentional differences:** per-path discard failures are propagated instead of the source's swallowed errors; commit-and-push partial success clears a successfully committed draft and warns about push failure rather than suggesting no commit occurred; stash drop and PR checkout get explicit confirmation; malformed/oversized/outside-symlink templates are rejected safely. These are disclosed differences, not grounds for an unqualified 1:1 claim.

Known-forge URLs mean GitHub/GitLab/Bitbucket public hosts and their supported URL syntax variants. Source intentionally does not build arbitrary self-hosted forge URLs.
