# Pending image work must precede Submit

Isolated baseline: frozen wave8 source snapshot, local commit `263e258`.
Pinned Rust `app/events.rs:1105–1127` explicitly drains pasted images at tick and
again on submit so paste immediately followed by Enter still attaches;
`app/session.rs:135` drains before taking attachments. Its dropped-image
`Attachment::from_path` path is synchronous before submit can run.

## Actual regression evidence

Go's asynchronous read/decode had no corresponding pending-image submit guard.
A controlled JSONL subprocess recorded a real prompt containing only text while
a read was held, and mounted Enter cleared the draft. A second test held the
finished decoder's owner callback: real Submit sent text without images, then
applying the callback attached the image to the next draft. These are synthetic
local RPC fixtures, not real model/provider calls.

## Narrow correction

One current-owner pending batch and one idempotent token per existing read/decode
stage track readiness. The read callback reserves decode work before releasing
its own token, so no submit gap exists between stages. The last token releases
only after its owner callback has applied or discarded the result. All token and
batch changes occur on the UI owner, not background workers.

Both App.submit (keyboard/button path) and Controller's direct Submit reject
while the current owner's image work is pending. Text, native selection/caret,
existing attachments and existing error display remain unchanged. Completion
does not automatically submit: the user explicitly sends again once images are
ready. This is a deliberate safe asynchronous adaptation of Rust's synchronous
drain, not deferred send/optimistic clear followed by restoration.

Task/window owner changes prune the old batch even with the composer hidden;
Close cancels it explicitly. Cancellation detaches and invalidates that batch.
A late or repeated callback cannot attach to the new draft or decrement/unlock a
new owner's counter. Failed reads release their token and retain existing error
handling, allowing a later explicit text submit. Ordinary same-task session ID
hydration continues to use the stable opaque attachment owner.

This changes neither the current max-eight checks, file suffix classification,
compression/decoder/raw fallback, image preview, nor file caret insertion. It
adds no worker pool or additional read/decode jobs; the existing asynchronous IO
is not forcibly canceled. Capacity admission, source drop-at-cap behavior and
input-order versus completion-order remain separate audits, not claimed fixed.

## Tests and limits

Mounted held-read Enter and queued-decode actual Submit regressions now retain
the draft. Actual subprocess receipts verify a later explicit send includes the
image. Tests cover read-to-decode handoff, direct Controller guard, duplicate
callbacks, unchanged errors, failed-read retry, Unicode selection/caret, existing
images, hidden new-task ownership, old/new completion ordering, cancellation and
Close. No physical clipboard/OS picker/IME or real model call was performed.
Focused race/type/vet/arm64 gates this delta; combined full UI remains integrator
owned to avoid redundant concurrent long suites.
