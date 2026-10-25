# Workspace picker: return from the view before waiting

## Source evidence

- Rust `crates/orbit-pi/src/app/session.rs::browse_for_folder` deliberately spawns the native sheet asynchronously. Its comment identifies reentry of the current entity from a blocking nested modal loop as unsafe.
- Go previously called `native.BrowseWorkspace` synchronously from `Controller.Action`, which is dispatched while building the immediate-mode view.
- Pinned MyGo `dialog.go` documents that dialog methods block but may be called from any goroutine. `Dialog.Open` queues native presentation on the main loop; `loop.go::await` pumps native events when called on the main thread, but waits on a channel when called off-thread.

This is a concrete architectural mismatch. It is **not evidence that the user's reported hang has been reproduced or resolved**.

## Correction

Only waiting for the native selection moves to a worker goroutine. Presentation still uses MyGo's supported main-thread dispatch. Completion is put onto the controller's existing UI update queue; `SetWorkspace` and error notification remain UI-owned.

- A pending flag allows only one native workspace dialog at a time.
- A later workspace/session/new-task intent wins over a late answer, even if navigation returns to the same workspace.
- A closed controller or closed/replaced parent window rejects the result.
- Cancellation does not change workspace; native errors cannot apply a returned path.
- An obsolete result does not show an obsolete error.

The native API does not expose cancellation of an already displayed sheet. This change does not force a modal result or claim that all OS close/presentation races have been tested.

## Regression evidence and limits

`TestWorkspaceDialog*` exercises the production action and controller UI queue with a controlled blocking native-dialog function. It checks immediate caller return, another production view frame while waiting, duplicate click suppression, Unicode/space paths, UI-only completion, cancellation/error/reopen, superseding navigation, and shutdown before/after completion is queued.

A synchronous mutation (removing the goroutine around the same native wait) fails `TestWorkspaceDialogReturnsBeforeAnswerAndAppliesOnUIQueue` with “BrowseWorkspace blocked its action/view caller while native sheet was open.” This is a regression sensitivity test, **not a real macOS hang reproduction**.

The prior exact `.4` native probe remains pinned and unchanged. It verified native cancellation and recent-directory switches, but confirmation was BLOCKED because the public remote accessibility tree exposed no confirm button. No new native confirmation acceptance is claimed here; actual macOS confirm/cancel/repeated/close testing remains required after integration.
