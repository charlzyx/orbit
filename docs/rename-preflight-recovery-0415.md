# Reconstructed rename preflight — new evidence only

Base ea98724a833bd01b74ca99a27ac3c3882f13c850. Executor recovery removed earlier unpublished prototype sources; no old hash or old PASS is used as evidence for this candidate. The frozen recovery checkout is untouched.

Current source has two rename commit paths: inline commitExplorerEntry and the legacy renameFilePrompt callback. Both checked domain HasDirtyUnder before observing pending native editor Buffer changes. Existing syncFileBuffers already preserves those changes for save/close. This six-line production patch calls it before both rename dirty checks and returns on failure. Dirty refusal, error presentation, and normal debounce autosave remain; no forced save, deletion or filesystem mutation API change.

New exact test TestRenamePreflightRecoveryPendingNativeBuffer mounts the actual file editor, changes its public Buffer before domain synchronization, and invokes each actual commit callback. This isolates the buffer/domain seam; it is not a captured OS input ordering claim. Baseline inline and prompt each fail because rename starts despite pending edits. With the patch both refuse rename, synchronize the original text, preserve native Undo, retain the original path and create no destination.

The test logs actual original-path bytes. Normal500ms autosave is allowed: original or pending-edit bytes are valid after refusal, but unexpected bytes or a destination fail. It deliberately does not turn a scheduler delay into a false assertion that normal autosave must never run. The production diff contains no Save/SaveAll call. Synchronization-error early-return is source-reviewed; the new test directly exercises dirty refusal, not an artificially injected SetText error.

Known remaining risks, not fixed here: filesystem RenamePath still runs before RenameUnder obtains ioMu; Document has no stable identity; deferred UI reconciliation blindly moves current path-prefix slots; native SetText failure still advances version bookkeeping; Open/Reload lack rename namespace generation and watcher ignores Reload errors. Stable-ID native indexing, short-state-lock reservation with ioMu across background rename, affected-ID receipts for callbacks, and explicit bounded retry handling require a separate reviewed candidate. No state mutex may span filesystem I/O; no permanent old-path alias or delete expansion is proposed.

New evidence lives in this isolated tree: red.jsonl (both subtests fail on unchanged ea98724); green.jsonl (race count2, both subtests pass twice, package2.782s). No full suite or native capture was run by this worker.

Final same-tree scoped validation:5 exact top-level rename tests,7 PASS events including2 subtests,0 failure/skip; race4.427s. UI vet and git diff --check pass. proof.json records exact hitsets and log hashes. This is not an aggregate gate or permission to integrate before independent review.
