# Watcher native pending protection — recovered candidate

Base: `469021fe08cc3e68dff0ee2e2b7dbdb1e9caee23` (includes rename preflight `0864f5b` and success-only native version acknowledgement). This is a separately reviewed adapter correction; no changes to frozen gate/source trees.

## Source contract and correction

`run.go` previously ran `Editor.Reload` on all documents in an Async worker, then its UI completion enumerated all then-current documents and called `syncDocument` unconditionally. `syncDocument` replaces differing native text and acknowledges its version. Domain `Dirty` cannot describe edits still pending in the native buffer or edits between worker completion and UI callback.

The watcher block now calls `reloadWatchedFiles`:

1. Synchronize existing pending editable buffers via existing `syncFileBuffers`, abort on genuine error (no forced save).
2. Capture actual original path/state targets and editor/files owner; skip targets currently closing. Reads remain background and no UI state is accessed by the worker.
3. At UI apply, require the same owner, same original nonnil native state, not closing, and **current native version equals acknowledged version**. Only then sync current domain state. New tabs, replaced native states and pending edits are not overwritten/acknowledged.
4. Readonly/error pending drafts are retained. Disk reads remain enabled so a later valid UTF8 transition can restore editability; the ordinary subsequent renderer can then synchronize retained draft through the existing real SetText path. No permanent retry/timer, force-save, stable-ID architecture, watcher generation or lock change.

Pinned Rust `299a32b`: `app/events.rs:85–100` reloads active file only; `viewer.rs:568–611` checks tab ID and load generation; `seed_editor:624–676` reads actual native text/revision, preserves dirty edits and leaves noneditable buffers alone. This Go patch does NOT change all-tabs scope or claim exact Rust reload parity.

## Actual evidence

`red.jsonl`: after a behavior-preserving extraction of the existing Async block, `TestWatcherLateNativeEditPreserved` runs a real background Reload, receives/holds its queued UI completion, edits the native Buffer, then applies completion. Expected pending text was replaced by external text; acknowledgement advanced from 0 to 2 despite pending version 1. Package FAIL 1.737s. This is a controlled real buffer/worker scheduling test, not a claim of native OS key timing.

MyGo Buffer.Set makes an undoable full-text edit. The RED proves visible draft replacement/incorrect acknowledgement, NOT irrevocable data loss or erased Undo.

`green.jsonl` and `green2.jsonl` retain an intermediate fixture failure: close navigation queued other jobs, so blindly taking the next callback did not prove OpenFile completion. The final close/reopen fixture waits for actual state removal/replacement using the existing bounded update-drain helper. This was a test scheduling correction, not a claimed production close/reopen fix.

`green-final.jsonl`: initial eight watcher top-level tests, -race -count=2 PASS (package 2.929s). Final focused JSON and vet are recorded in proof.json after completion, including defensive owner coverage.

## Evidence boundaries / remaining scope

- The owner-replacement negative fixture is defensive. Current production workspace service switching retains filesState; no reachable workspace cache replacement bug is asserted.
- Callback validation covers actual UI close/reopen state replacement. It does not add document identities or solve domain path reuse while Reload itself is in flight.
- Concurrent Reload jobs still lack Rust-style per-document generation; older reads may finish later. This remains separate.
- Keeping late native text does not retroactively derive a conflict after clean domain content already adopted external bytes. Existing domain conflict policy remains unchanged.
- File read errors live in Content.Error; Reload mostly returns nil. This patch preserves existing error behavior, not a new error-reporting/retry policy.
- No new image/capture, full-suite gate, release or external publication. Passing focused tests is not full functional/pixel parity.

Final fresh gate: `final-focused.jsonl` **11 top-level tests / 13 test PASS events** (two readonly/error subtests included), no skip/fail, race package **2.887s**. `vet.log` UI vet PASS; `git diff --check` PASS. `proof.json` records exact hitset and hashes. No full suite run by this worker.
