# Native attachment/export waits and delayed attachment ownership

## Source match

Pinned Rust `app/composer_ops.rs::attach_image/attach_file` and `usage/page.rs::export` explicitly use asynchronous file dialogs to avoid holding the current UI entity inside a nested modal loop. Usage captures its index/filter/view before waiting. Go previously called `BrowseAttachments` and `SaveFile` synchronously from the immediate-mode action handler.

This increment moves only native waits off the view thread, through MyGo's documented any-goroutine dialog API. It keeps native overwrite prompts and caller-selected paths. Window and task ownership checks and UI mutations remain on the UI queue. This does not establish real macOS dialog-confirm acceptance or resolve the separately unconfirmed workspace hang.

## Behaviors

- Attachments, Usage export and workspace selection do not stack native sheets on repeated/cross-kind clicks.
- Task-bound dialog completion compares captured workspace/intent epoch and the manager's existing opaque entry key/task generation. Empty session IDs are not treated as a stable task identity; ordinary ID hydration alone does not invalidate the target.
- Cancel/error, shutdown or a closed/replaced parent cannot apply a late native result.
- Usage export clones its mutable index/filter/options collections before waiting. A later page/filter/task change cannot silently alter the data selected for export. The body is serialized before opening/truncating the destination; native SaveFile retains overwrite confirmation. Once writing has begun it is an accepted operation, not an OS-cancellable transaction.
- Session HTML export retains its existing synchronous implementation and remains an explicit follow-up. A UI-only captured-target check before generic Manager.Send is insufficient because the runtime/task may change between validation and actual command submission; no unsafe async conversion is included.

Attachment image reads had a separate ownership defect: `AddPaths` previously called `AddImage` only after reading, so the second stage captured whichever task was active then. Both read and decode now retain the manager's opaque runtime key/task generation, workspace, native parent and UI intent generation. Normal same-task session-ID/path hydration does not discard a valid attachment; standalone App hosts without the managed identity callback conservatively compare their snapshot identity. A stale read/error/decode is discarded rather than attached/reported in a later task, including another empty-ID draft. Non-image path insertion remains the existing append behavior; caret parity is not claimed in this increment.

## Verification scope

Focused race tests cover asynchronous caller return, duplicate/cross-kind requests, task/window/shutdown changes, cancel/error/success export, unchanged existing file after cancellation/error, captured export collections and delayed image read/decode ownership. Tests use controlled native-dialog functions and temporary data, not user credentials, actual OS confirmation or real Pi export inference. Native dialog execution and final combined-tree aggregate validation remain separate gates.
