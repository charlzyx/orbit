# Visual candidate follow-up corrections

This is a follow-up to the frozen490-candidate publication snapshot, not a new
screenshot count. It does not alter stable IDs, claim a capture, or change the
frozen publication manifest. Apply these expectations in the next revision.
Rust reference remains `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

## Usage reviewer corrections received after the snapshot

- **V301:** the columns popup footer reads **Show all**, not Reset columns.
  `usage/filters.rs::columns_menu`, lines523–575. The popup must actually be
  painted by the native host; a CPU Tester menu descriptor is not its screenshot.
- **V284:** use pointer hover for the highlighted option, not ArrowDown.
  `usage/filters.rs`, lines178–212, registers PickerCancel and PickerConfirm
  to dismiss; it does not register arrow selection. Enter must not toggle a row.
- **V284–V287:** Select all selects all globally available IDs even when the
  search narrows visible rows. All remains selectable when the query has no
  matching concrete option. See `usage/filters.rs:435` and `usage/page.rs:1815`.

- **V268:** source Failures has five columns: When, Type, Model, Session, Message.
  `usage/view.rs:3848`; no separate Provider column. Provider information can
  appear inside model presentation. The independent Usage review corrected its
  previous six-column wording.

- **V265 / Tools table:** `usage/view.rs:3729–3755` marks every Tools column
  nonsortable and nonresizable. Fixed call-descending order is not a user sort
  interaction. Do not apply generic table sort/resize candidates to this table.
- **V268 / Message:** `usage/view.rs:3848–3904` makes the Failures Message
  column nonsortable. Only its actually sortable headers may use sort cases.

These are source expectations, not proof that current Go behavior matches.
The Usage owner is correcting its source-review artifact separately.

## GitHub render branches challenged against source

### V345 · Missing gh

- `git_panel/mod.rs::tab_bar`, lines3203–3240, omits Issues and Pulls tabs while
  `gh_installed=false`; it does not render six visible tab buttons in that state.
- `set_tab`, lines589–608, still accepts indices4/5. A controlled fixture may
  therefore use the registered Issues action to reach the missing-gh body.
- `gh_setup_tab`, lines6406–6434, renders the informational missing-gh note.
  Do not invent an inline installer button or fake remote data.

### V346 · Signed out

- With `gh_installed=true` and `gh_authenticated=false`, `gh_setup_tab` renders
  the auth-required title plus `gh_detail` or the default explanatory body.
- It is an `empty_note`, not a sign-in form or an Authenticate button.
  The Go convenience Authenticate button is not evidence of source pixel parity.

### V352 and V363 · GitHub list failure

- `issues_list_view`, lines4946–5005, and `pulls_list_view`, lines5633–5693,
  prioritize their error note over loading/empty/list branches.
- Capture the error note with the filter bar and the normal global Refresh
  control. There is no separately invented inline Retry button.
- V363 should mean the PR fetch-error state, not an impossible simultaneous
  loading/error presentation. Initial PR loading is a distinct source branch
  that must remain on the discovery list until the candidate ledger is revised.

### V368 · PR review body

- `review_pull`, lines1942–1959, uses the existing shared `pr_comment` text;
  the detail's Approve/Request changes actions dispatch that value.
- Do not invent a separate source review modal or a second independent body
  editor merely because the Go implementation currently uses one.

### V369 · Merge confirmation

- `request_merge_pull`, lines1963–1976, captures the method and delete-branch
  setting into PendingConfirm.
- `confirm_popup`, lines7001–7011, displays PR number and selected method.
  It does **not** render a delete-head checkbox inside the confirmation.
- The checkbox and method choice belong to PR detail before opening the popup.
  Verify delete-branch argv functionally; do not claim the modal image proves it.

## Known functional source differences, not new requirements

`git_panel/mod.rs::set_context`, lines624–672, hides new forms and clears
list/filter/template state, but does not explicitly clear creation text entities
or `issue_labels`. The previous Go implementation cleared those values.
[Later bounded source acceptance](github-base-picker-acceptance.md) now proves
the failure and correction with unchanged-source Rust decisions plus fake-gh
and native tests. This does not alter the frozen visual register or claim images.

No authenticated GitHub operation was used to inspect these render branches.
