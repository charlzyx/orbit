# GitHub source acceptance: bounded native failure-path corrections

Reference: Orbit0.2.12 at `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
This report covers six narrow scenarios. It does **not** mark W20–W30 complete,
claim authenticated GitHub execution, or establish visual/native-OS parity.

## Scope and safety

The tests invoke native MyGo controls where stated, an executable fake `gh`,
and real disposable local Git repositories. The fake records argv and returns
fixed JSON or deliberate failures. It never invokes real `gh`, reads credentials,
or contacts GitHub. The Unix-shebang harness is explicitly skipped on Windows;
Windows runtime acceptance remains open. No remote mutation was performed.

Exact current source hashes, command, timing, named results and before/after
stability are in [github-evidence.json](acceptance/github-evidence.json).
These dirty-source results are not attributed to the published1821f982 commit.

## Executed scenarios

### GH-A01 · Cancel/reopen retains independent drafts (W24, W28)

- Rust: [`open_issue_new` and `open_pull_new`, lines1318–1350](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1318).
  Issue and PR have separate text entities. Opening another form does not erase
  their title/body/notes; applying the sole template is a separate rule below.
- Steps: open New Issue without templates; enter distinct title/body/notes;
  click the actual native Cancel button. Do the same for New PR. Visit a comment
  form, then reopen each creation form.
- Expected: each form retains its own original prose, never the comment draft.
- Observed red: reopened Issue values were all empty.
- Correction: separate session-local `githubFormDraft` stores preserve each
  source form across shared native widgets and unrelated dialogs.
- Ordinary regression: `TestGitHubNativeCancelRetainsIndependentDrafts`.

### GH-A02 · Failed create stays open; retry does not duplicate (W24, W28)

- Rust: [`create_issue`, lines1202–1244](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1202)
  and [`create_pull`, lines2008–2056](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L2008).
- Steps, separately for Issue and PR: enter whitespace-padded Unicode title,
  multiline body and notes. Configure fake CLI creation failure. Click native
  Create, then attempt another create while busy. Wait for error; correct the
  fake failure and click Create again.
- Expected: form remains open while request runs and on failure; draft stays
  editable; failure text is visible. Duplicate busy submit does not spawn a
  second mutation. Successful retry closes the form and clears title/body/notes.
  Recorded CLI title is trimmed exactly as the Rust handler does.
- Observed red: both forms immediately closed while `Busy=true`.
- Correction: creation dialogs close only on successful mutation; failure does
  not clear drafts. Actual form displays the error while the modal is open.
- Ordinary regression: `TestGitHubNativeCreateFailureKeepsFormAndRetryClearsOnce`
  with independent `issue` and `pull` subcases.

### GH-A03 · Optional labels cannot reject a successful list (W21, W25)

- Rust: [`refresh_issues`, lines1068–1101](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1068)
  uses `list_labels(...).unwrap_or_default()` only while its cache is empty.
  `refresh_pulls` does not require the issue-label request.
- Steps: fake CLI probe/auth/list succeed with one real fixture item; label
  lookup fails. Refresh Issues and Pulls independently in disposable repositories.
- Expected: both lists display their returned item without a list error.
  Pulls must not request issue labels at all.
- Observed red: both list loads failed with `fixture labels unavailable`.
- Correction: issue labels are best effort; PR refresh is label-independent.
- Ordinary regression: `TestGitHubOptionalLabelsNeverBlockIssuesOrPulls`.

### GH-A04 · Label cache survives repeated navigation (W21)

- Same Rust `refresh_issues` cache contract as GH-A03.
- Steps: load Issues with one label; visit History, Pulls and Issues; refresh again.
- Expected: label remains present and fake CLI records exactly one `label list`.
- Observed red during independent follow-up: two label requests occurred because
  another tab's refresh replaced the aggregate Go data without preserving labels.
- Correction: immutable load specification carries the repository-local cache
  across tab data replacement. The original run still reset cache on workspace changes. That separately challenged difference is corrected and tested in [the later source-state supplement](github-base-picker-acceptance.md); this historical assertion remains within-repository only.
- Ordinary regression: `TestGitHubLabelCacheSurvivesTabRoundTrip`.

### GH-A05 · Sole template and Blank preserve the exact source distinction (W24, W28)

- Rust: [`apply_issue_template`, `apply_pull_template`, `clear_template_body`,
  lines1352–1423](https://github.com/imrj05/orbit/blob/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac/crates/orbit-pi/src/git_panel/mod.rs#L1352).
- Steps: open Issue with exactly one template, edit prose/title/notes, dismiss and
  reopen. Then select Blank while the body equals a known PR skeleton; repeat
  with unrelated user prose.
- Expected: explicit reopen reapplies the sole template body, while preserving
  nonempty title and notes. Blank removes **any known Issue or PR skeleton**,
  not just the most recently selected template, but never arbitrary user prose.
- Correction: separated draft storage does not accidentally override the source
  template-on-open behavior; Blank recognizes both template collections.
- Ordinary regression: `TestGitHubTemplateReopenAndBlankMatchSource`.
  The old extra workspace-reset assertion verified Go isolation, not source parity.
  It has now been corrected to source retention in the separately hashed
  [later source-state supplement](github-base-picker-acceptance.md).

### GH-A06 · Mutation succeeded, follow-up read failed (W24)

- Rust `create_issue` clears/closes on successful creation before opening its
  resulting detail. A subsequent detail-load failure must not turn that success
  into permission to repeat the mutation.
- Steps: fake Issue creation returns a URL for#42; fake detail lookup then fails.
- Expected: created receipt remains, form is closed/cleared, detail error is
  reported, and there is exactly one `issue create`. No generic retry closure
  may replay the successful write.
- Ordinary regression: `TestGitHubCreatedMutationNotRetriedWhenDetailRefreshFails`.
  This particular post-create failure test is Issue-only; PR is not silently
  counted from similar code.

## Commands and outcomes

Canonical commands:

```
ORBIT_GO=/workspace/shared/orbit-toolchains/go/bin/go \
XDG_CACHE_HOME=/workspace/shared/orbit-toolchains/cache \
GOCACHE=/tmp/orbit-go-cache GOPATH=/workspace/shared/orbit-toolchains/gopath \
scripts/go-native.sh test -race -json ./internal/gitworkbench
scripts/go-native.sh vet ./internal/gitworkbench
```

The initial source challenges failed as described above. The ordinary full
Git workbench race suite and vet pass after the bounded corrections; exact
post-correction outcomes and hashes are in the evidence file. Package suite
success does not convert these broad feature scopes into full PASS.

## Remaining source-specific gaps and gates

- W20: real missing/unsupported `gh`, authentication terminal ownership and real
  signed-in account identity still require independent native lifecycle checks.
- W21/W25: source list limits are50 and label limit100. The Rust GitHub lists do
  **not** have Load more pagination; do not invent a pagination parity requirement.
  Repeated filter navigation during an in-flight request still needs native tests.
- W22/W26: detail loading, stale completion after navigation, full Markdown,
  reviews/checks/files/commits and exact menus are not proved by these six cases.
- W23/W27: comments, labels, close/reopen, review, checkout and merge each still
  need action-specific native request/confirmation/error tests. Existing backend
  command tests are not a substitute. Authenticated remote mutations remain unrun.
- W24/W28: Go uses a modal where Rust uses an inline form; extra native fields
  are not source pixel parity. After failed creation, Rust refreshes its list;
  that entire fan-out is not asserted here.
- W28: the free-text base field was replaced by the source-derived anchored
  picker in [later bounded acceptance](github-base-picker-acceptance.md). Original
  GPUI Escape dispatch and matched native pixels remain open.
- W29: draft-generation cancellation and cross-form late callbacks need their own
  independent fixture. These retention tests do not prove model generation.
- W30: Git-to-Review action wiring, exact selected-file identity and native visual
  layout remain separate from the tested workspace backend.
- Same-state Rust/Go GitHub screenshots, actual OS menus/clipboard/browser launch
  and real authenticated operations are unverified. No whole GitHub scope is PASS.
