# Orbit source-first acceptance checklist: initial executable inventory

Baseline: [`299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`](https://github.com/imrj05/orbit/tree/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac). Native Go + MyGo target; cloud-only fixtures.

**139 feature scopes · 297 row-specific procedures · 676 explicitly unexpanded draft gates · 24 narrow assertions passed · 0 full feature passes.**

This is an initial checklist to expand and execute, not a completion report. Gates marked DRAFT_NEEDS_SOURCE_EXPANSION still need exact source-specific faults, files, timing and chords; they are not finished behavioral specifications. Every scope links pinned Rust handlers, row-specific behavior/edge actions, Go mapping and evidence. Functional, visual and native-OS verdicts are separate.

**Status meanings:** NOT_IMPLEMENTED = no working path; PARTIAL = named subset verified or known gap; UNTESTED = acceptance scope not run; PASS = exact scope executed on identified revision with evidence; BLOCKED = explicit unavailable prerequisite. A failed assertion is recorded as an observed failure under PARTIAL, not hidden in a package-level success.

**Safety:** use isolated HOME/session stores, temporary repositories, local bare remotes and fake gh/provider/auth/update endpoints. No user repositories, real credentials, external mutations or user Mac. Native OS tests require an appropriate separately authorized cloud host.

| Category | Rows | Functional | Visual | Native OS | Procedures |
|---|---:|---|---|---|---|
| Shell, navigation, sessions and workspaces | 14 | PARTIAL 14 | UNTESTED 14 | UNTESTED 14 | [S01–S04](acceptance/shell-S01-S04.md) · [S05–S08](acceptance/shell-S05-S08.md) · [S09–S12](acceptance/shell-S09-S12.md) · [S13–S14](acceptance/shell-S13-S14.md) |
| Composer, transcript and interactive agent surfaces | 20 | PARTIAL 19, NOT_IMPLEMENTED 1 | UNTESTED 20 | UNTESTED 20 | [C01–C04](acceptance/chat-C01-C04.md) · [C05–C08](acceptance/chat-C05-C08.md) · [C09–C12](acceptance/chat-C09-C12.md) · [C13–C16](acceptance/chat-C13-C16.md) · [C17–C20](acceptance/chat-C17-C20.md) |
| Settings, providers, skills, packages and MCP | 20 | PARTIAL 16, UNTESTED 4 | UNTESTED 20 | UNTESTED 20 | [P01–P04](acceptance/settings-P01-P04.md) · [P05–P08](acceptance/settings-P05-P08.md) · [P09–P12](acceptance/settings-P09-P12.md) · [P13–P16](acceptance/settings-P13-P16.md) · [P17–P20](acceptance/settings-P17-P20.md) |
| Usage analytics, update, privacy and release behavior | 14 | PARTIAL 8, UNTESTED 6 | UNTESTED 14 | UNTESTED 14 | [U01–U04](acceptance/updates-privacy-U01-U04.md) · [U05–U08](acceptance/updates-privacy-U05-U08.md) · [U09–U12](acceptance/updates-privacy-U09-U12.md) · [U13–U14](acceptance/updates-privacy-U13-U14.md) |
| Workbench lifecycle and shared state | 3 | UNTESTED 2, PARTIAL 1 | UNTESTED 3 | UNTESTED 3 | [W01–W03](acceptance/workbench-lifecycle-W01-W03.md) |
| Git changes, history, graph and repository operations | 16 | UNTESTED 16 | UNTESTED 16 | UNTESTED 16 | [W04–W07](acceptance/git-changes-history-W04-W07.md) · [W08–W11](acceptance/git-changes-history-W08-W11.md) · [W12–W15](acceptance/git-changes-history-W12-W15.md) · [W16–W19](acceptance/git-changes-history-W16-W19.md) |
| GitHub issues and pull requests | 11 | UNTESTED 11 | UNTESTED 11 | UNTESTED 11 | [W20–W23](acceptance/github-W20-W23.md) · [W24–W27](acceptance/github-W24-W27.md) · [W28–W30](acceptance/github-W28-W30.md) |
| Review and turn checkpoints | 9 | UNTESTED 7, PARTIAL 2 | UNTESTED 9 | UNTESTED 9 | [W31–W34](acceptance/review-W31-W34.md) · [W35–W38](acceptance/review-W35-W38.md) · [W39–W39](acceptance/review-W39-W39.md) |
| Worktrees and setup lifecycle | 12 | UNTESTED 4, PARTIAL 8 | UNTESTED 12 | UNTESTED 12 | [W40–W43](acceptance/worktrees-W40-W43.md) · [W44–W47](acceptance/worktrees-W44-W47.md) · [W48–W51](acceptance/worktrees-W48-W51.md) |
| Explorer and file editor | 9 | UNTESTED 6, PARTIAL 3 | UNTESTED 9 | UNTESTED 9 | [W52–W55](acceptance/files-W52-W55.md) · [W56–W59](acceptance/files-W56-W59.md) · [W60–W60](acceptance/files-W60-W60.md) |
| Terminal, platform and window lifetime | 11 | UNTESTED 11 | UNTESTED 11 | UNTESTED 11 | [W61–W64](acceptance/terminal-platform-W61-W64.md) · [W65–W68](acceptance/terminal-platform-W65-W68.md) · [W69–W71](acceptance/terminal-platform-W69-W71.md) |

## Revision and evidence rules

Published implementation reference: [`80580ee3cffe8eab19b9399d7813fa6be61ec6d9`](https://github.com/charlzyx/orbit/commit/80580ee3cffe8eab19b9399d7813fa6be61ec6d9). Current acceptance evidence is a separately hashed dirty-tree snapshot, **not** evidence for that published SHA. [Exact files, command, timestamps and named test outcomes](acceptance/workspace-evidence.json). The manifest was unchanged before/after the executed suite.

Per-row `last_verified_sha` is null until that scope is actually tested. Narrow assertions identify the manifest SHA and test name separately. Existing source/Go tests are candidates, not automatic passes.

Visual procedure for each row: capture pinned Rust and Go at identical OS, scale, installed fonts, locale, viewport and fixture state; compare geometry, ordering, text, focus, clipping and interaction states. Native-OS procedure: repeat the row on every advertised host, including actual PTY/window/menu/clipboard/dialog lifetimes. Cross-compilation and headless CPU rendering are not substitutes.

## First independently challenged cases

- W34: fixed and regression-tested reversed numstat/file identity, suppressed raw diff metadata, synthetic unpositioned hunks, escaped paths and asymmetric compact gaps. 1,204 lexical oracle cases plus separate old/new carry and paint-only spans pass; W34 as a whole remains PARTIAL.
- W59: self-save watcher reload had reset undo history. Regression now proves undo history survives the save echo; slow-save and dirty-conflict assertions pass. Full editor interaction/OS parity remains open.
- W51: a throttled worktree watch event could be dropped without another event. The pending refresh regression passes. Full Worktrees page parity remains open.
- The independent Usage settings round-trip run initially failed; its owner reports a correction and new scoped evidence in [settings backend verification](settings-backend-implementation.md). This index does not silently convert that whole category to PASS.

## Additional owner evidence, separately scoped

- [Settings/backend source checklist](settings-backend-implementation.md) and [exact source hashes](settings-backend-source-evidence.json).
- [Native UI source challenges:29 scenarios and concrete omissions](native-ui-acceptance.md).
- [Runtime/Agent exact lifecycle acceptance](runtime-acceptance.md).
- [MCP settings:19 independently scoped cases](mcp-ui-implementation.md).
- [Backdrop source acceptance, host integration pending](background-acceptance.md).
- [Original full source inventory and all remaining caveats](feature-parity.md).
- [Paint-only lexical oracle and provenance caveat](highlight-parity.md).
- [Machine-readable checklist index](acceptance/index.json).
