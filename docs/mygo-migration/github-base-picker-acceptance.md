# GitHub source acceptance: PR base picker and retained workspace state

Reference: Orbit0.2.12, commit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
Six additional narrow Go assertions pass after concrete failures were reproduced.
Three standalone Rust oracle tests execute unchanged source decisions. **W24/W28
remain PARTIAL; visual and native-OS parity remain UNTESTED.**

## Exact revision and verification

[github-state-evidence.json](acceptance/github-state-evidence.json) records the
before/after source hashes, ordinary complete Git workbench `-race` result, vet,
test names, red-run log hash and oracle output. Its scoped manifest is
`07dc34d0411e17fa29e1d4d048207adb9472ccfcb6aecbf088ffd37151a19e72`.
The manifest was unchanged during verification. This is **not** a published
Git commit or evidence for1821f982/f61e00f; publication must identify the actual
matching tree separately. Full package race test took4.413s; vet exited0.

The [standalone oracle](acceptance/run-github-state-oracle.py) verifies the pinned
GitPanel SHA256 `bbfc71243ee6d2f93eb9cddf3dc8bb0749b677244f478cdce69f2c836bcfef85`,
extracts `default_base` and `set_context` unchanged, and compiles them with
std-only Context/Review adapters. The filter/sort uses the unchanged contiguous
expression from `pr_base_popup`; the GPUI node construction is not executed.
This is source-decision evidence, **not** an original-GUI screenshot or key-route
oracle. No dependencies are downloaded and no user repository is accessed.

## GH-B01 · Default base matches actual source branches (W28)

- Rust: [`git_panel/mod.rs::default_base:2139–2154`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2139).
- Steps: open New PR with empty refs; only current local `feature`; current plus
  `release`; master present; main and master; main itself current; a tag named
  main plus release; and only remote/tag refs.
- Expected, in that order: main, main, release, master, main, main, main, main.
  Source's preferred-name pass deliberately does not test ref kind; its fallback
  alone requires a local branch unequal to the current branch.
- Red: five of these eight fixtures differed. Go selected empty/current/release
  where source selected main/release.
- Fixed: `State.defaultBase` reads actual refs with the exact preference/fallback.
- Test: `TestGitHubSourceDefaultBase` with eight subcases; corresponding Rust
  oracle `default_precedence_and_current_fallback`.

## GH-B02 · Local-only, Unicode-aware anchored branch search (W28)

- Rust: [`base_branch_selector:1618`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1618),
  [`pr_base_popup:1666–1770`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1666).
- Steps: native-click the base chip with mixed-case local refs plus a remote and
  tag. Type `  bEt  `; separately test Greek `ΟΣ` using `ος`, and dotted capital
  `İ` using `i` plus combining dot. Send Down then Enter without selecting, then
  click Beta explicitly.
- Expected: 320px popup above the chip, search focused, Local branches header,
  local rows only, stable Unicode-lowercase sorting/filtering, selected-row check,
  click changes base and dismisses. Search Enter must not select or create a PR.
- Red: the base was a free-text field and no anchored popup existed.
- Fixed: native lookless PopoverBase, real embedded branch/search/check icons,
  source local-ref/filter/sort and pointer-selection semantics. `cases.Lower`
  handles mappings that simple Go `strings.ToLower` would miss.
- Tests: `TestGitHubNativeBasePickerFiltersLocalBranchesWithoutImplicitSelection`
  and `TestGitHubBaseFilterMatchesRustUnicodeAndStableSort`; corresponding Rust
  oracle `literal_local_filter_sort_preserves_source_case_ties`.

## GH-B03 · Empty/loading/error and repeat dismissal (W28)

- Steps: open the native picker while refresh state is pending; type an absent
  name and press Enter; inject a refresh error while keeping the form; click
  outside the popup, immediately click the chip, then repeat after200ms.
- Expected: No matching branches appears; no branch mutation or submission;
  title/body/Draft remain. Outside press closes only the popup. Same-gesture
  reopen is suppressed; a later reopen clears the query.
- Source close uses `menu_dismissed_at`, with the200ms guard in
  `base_branch_selector`; `close_ref_picker:2818` stores dismissal time.
- Test: `TestGitHubBasePickerEmptyDismissRepeatAndErrorDoNotMutateDraft` checks
  actual pointer/key events and popup geometry. The pending/error transition in
  this case is injected state, not a claim that an actual refresh subprocess
  was paused. Actual subprocess failure is covered by GH-B05 below.
- Open keyboard gate: Picker bindings exist in `main.rs:448–451` but GitPanel
  has no corresponding listener. Go tests ensure Up/Down/Enter do not choose or
  submit. MyGo Escape closes its popup without closing the form; that containment
  test is **not** proof of original GPUI Escape propagation. Native GPUI execution
  is still needed before this keyboard dimension can pass completely.

## GH-B04 · Workspace changes retain the original in-memory drafts/cache (W24/W21)

- Rust: [`set_context:624–672`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L624)
  closes creation forms and clears results/templates but does not clear Issue/PR
  title/body/notes entities, `pr_new_draft` or `issue_labels`.
- Steps: load one label using fake gh in temporary repositoryA; write distinct
  Issue/PR inputs and enable PR Draft; switch to temporary repositoryB. Reopen
  each form, refresh, then construct a new panel instance.
- Expected: forms close on switch and old templates clear. Independent prose and
  Draft survive within the original panel; cached labels avoid a second lookup
  even across repositories. No Create occurs. A fresh panel has empty drafts and
  labels. Reopening PR recomputes its default base as source does.
- Red: Go cleared all Issue/PR draft values and cache on workspace switch.
- Fixed: remember the active draft before changing context; retain independent
  draft objects and lazy label cache, while clearing result/template/transient
  state. This matches the source cache quirk; it does not assert that labels were
  freshly fetched from repositoryB. No disk persistence was added.
- Test: `TestGitHubWorkspaceChangeRetainsSourceDraftsAndLabelCache`; corresponding
  Rust oracle `changed_workspace_retains_inputs_and_labels_but_clears_lists_templates`.
  The older template test's Go-isolation assertion was corrected; its old
  historical evidence is not relabeled as source parity.

## GH-B05 · Selected base and Draft reach exact argv after failure/retry (W28)

- Rust: [`create_pull:2008–2056`](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2008).
- Steps: native-click branch `release/雪`, keep multiline-capable prose and
  whitespace-padded title, enable Draft, then click Create with fake gh failure.
  Click Create again while busy; wait for failure; remove the failure and retry.
- Expected: pending/error keep the form, chosen base and Draft. Exactly two fake
  `pr create` calls occur, one failed and one explicit retry, with `--base
  release/雪`, trimmed title, exact body and `--draft`. Success closes/clears.
- Test: `TestGitHubNativeSelectedBaseSurvivesCreateFailureAndExactArgvRetry`.
  Fake gh records argv and never invokes gh/network. Real temporary Git
  repositories are used only for local metadata refresh. No authenticated
  remote mutation or user's repository was touched.

## Remaining boundaries

The source forms are inline; Go still wraps them in a modal. Whole form fields,
failure-refresh fan-out, source dispatch, full density/theme geometry and matched
Rust/Go images remain open. In-flight workspace races outside these retained
fields are not silently proved by the pure `set_context` oracle. Passing GitPanel
package tests are not counted as whole-feature passes; this supplement
contributes exactly six new named Go assertions, not
one pass per subcase or per Rust assert. The139-scope checklist remains honest
about functional, visual and actual native-OS verdicts separately.
