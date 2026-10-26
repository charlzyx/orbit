## Current integration validation — 2026-10-26 19:29 UTC

The reconstructed candidate (not a byte recovery of the lost candidate) completed the serial local gate on `ebb9e62ea6508bdc1d800d5e8447766bb2cdf9f6`, tree `eebf53fce12962d92698808f56c4b10e558bc3d5`. All 1,721 non-document file hashes remained unchanged. Root independently reviewed the result and approved ordinary feature-branch publication; no release or native retry is included.

- Full race: UI 446.544s; 3,583 passing test events, 19 enumerated conditional skips, zero failures. Previous 266 plus new coverage gives 343 exact package/test identities, each RUN/PASS once with no skip/failure.
- Composer/selection 19 events; required menu 22 top-level / 32 events, zero skip; terminal 115 passing events / four conditional skips. Wrapper 25, packaging 40, icon 5, legacy packaging 21, vet and Darwin ARM64 build passed.
- Existing harness offline tests: 102 Python and 31 Go policy names each three times passed. This is not Swift or native Mac execution; both prior unassigned Mac attempts remain NOT_RUN.
- Full JSONL SHA256: `eecdb324776a5ecb17c18d849b4158c7f88f90bf43df9e92cfd8b7939db7b853`. The vanished 17:21 run remains UNKNOWN; no historical result was substituted.

The historical reconstruction notes below are retained. External filesystem actors and deletion-before-CloseUnder remain explicitly outside the transaction guarantee.

---

# Rename transaction recovery candidate — new source and new evidence

Base remote commit `6a9c9937512387740f7a4e3cc4ca245e78d712df`, tree `5da4f267d2999f0e33598c1d89620dce39c6853c`. This is a fresh behavioral reconstruction after the second executor reset, **not** a byte-identical restoration of lost `cb80` / patch `8d115` or the later lost integration tree. Their old logs are unavailable. The old 17:21 full-chain terminal result is **UNKNOWN** and is not used as evidence here.

## Reconstructed contract

Both rename UI entrypoints first synchronize native pending buffers and preserve the dirty-edit refusal. Their original owner/editor/fixed-ID receipt now submits RenameDocuments, rather than filesystem rename followed by a separately locked domain remap. The existing ioMu serializes this Editor's rename and writes; mu is held only for short admission/reservation/commit, never filesystem I/O. New members, stale target IDs, dirty/saving documents, a closed Editor and occupied destinations are refused. Late SetTextID remains responsive and is preserved through success/failure; autosave and SaveID resolve the original ID's committed path.

Open/Reload continue reading off the state lock, with namespace epoch and per-document reload sequence checks around the read. Global epoch is intentionally conservative: unrelated old reads may receive an explicit namespace error and require a new user attempt. CloseUnder increments epoch even with zero tabs; successful actual CloseID removal increments it; failed and already-absent CloseID do not. CloseAll failure remains recoverable.

CloseIDReceipt shares the close implementation with the old error-only wrapper and returns a nonzero `{ID, Path}` only for actual successful removal. Private admitFileClose returns the original complete async job; closeFile still dispatches it with Async. Native metadata cleanup remains separate from navigation eligibility. A files-local navigation generation changes only on real page transitions through the existing shared Navigate/goHistory path, preventing same-page history ABA without changing global scheduling or openGeneration.

Reopen admission records the original ID or a replacement the user explicitly selected later. Original-ID reopening requires its successful close receipt to match the requested path. ReopenClosed uses normal open's reader with vacant-only pre/post checks and existing namespace fences: it never returns a newly occupied path's existing document. Explicit replacement and newly reopened results use lock-protected ActivateID, never a path fallback. Owner/editor/workspace/open-generation/navigation checks precede UI completion effects. Canceled completions may leave an inactive registered tab, as existing OpenInactive does; no rollback is claimed.

Watcher reports coalesce only NamespaceBusy/NamespaceChanged by errors.Is within that one batch, retain individual non-namespace errors and cancellation behavior, preserve native pending versions, and do not retry. Actual LoadFile OS errors still use the existing FileContent.Error representation; injected error-reporting tests do not claim those are newly returned by ReloadID.

## Fresh RED before correction

Only a behavior-preserving private read seam preceded the first baseline tests:

- `readfence-red.jsonl`: two actual temporary-file/held-LoadFile cases, **2 RUN /2 FAIL**, package **0.010s**. Deleted zero-tab namespace and a newer successful Close were both undone by an older read. The baseline uses its existing Close(path) API; the unchanged test later exercises the shared CloseID implementation through that wrapper.
- `navigation-red.jsonl`: real close worker/held owner-queue callback plus rename/recreate; Settings, Chat, new-file selection, reopen→Settings and both history-ABA branches failed. Explicit replacement reopen was the positive control. Package **1.968s**, six failing subcases, not a claim of seven failures.
- `reopen-order-red.jsonl`: after initial transaction/navigation reconstruction but before fixed-ID reopen hookup, actual admit→request original reopen→RenameDocuments with held UI receipt→register replacement→close job/completion first→rename receipt last. **1 RUN /1 FAIL**, package **1.703s**. This independently reproduced the review counterexample instead of relying on the lost old RED.

All these exact new assertions remain and pass after correction. No waits, debounce values, skips or test assertions were weakened to obtain GREEN.

## Fresh final focused evidence

Go **1.27.3** and the canonical go-native wrapper were used after the coordinator restored the pinned environment. Actual top-level RUN names are compared to explicit expected lists, not inferred from a successful regex invocation. New `docs/rename-recovery-1910-proof.json` records the exact hitsets and local evidence hashes.

- Workspace: **20 top-level /40 test PASS events**, zero fail/skip, race **1.050s**.
- UI: **39 top-level /65 test PASS events**, zero fail/skip, race **4.849s**.
- Linux workspace/UI vet PASS; Darwin arm64 UI test-binary compilation PASS (Mach-O arm64); Darwin workspace/UI vet PASS; validation chain exit 0; diffcheck PASS.

Coverage includes held rename success/failure with responsive snapshots and late edit/identity save, actual timer write completion to the renamed path, changed/new-member/dirty/closed/destination admission, out-of-order reload, namespace away/back and failed-rename fences, failed/idempotent close and unrelated-read effects, CloseAll recovery, save identity against replacement, real held reopen-read replacement/rename/CloseAll, successful/no-op/failed close receipts, normal reopen before/after worker, explicit replacement subsequently replaced again, delayed callbacks versus new owner/navigation, actual same-tuple goHistory ABA, and the published native ID/pending/Undo/image/watcher/preflight/render-close regressions.

The full published fileEditorContent/fileEditorDocument rendering section remains byte-identical to base (SHA256 `64410d843e51dcf9d5309dea2233dfd01a57150135f9c380786e85b8c8178c43`), retaining live-ID rejection and the second nil guard. Late native pending tests use real editor buffers/adapters, not an OS key/focus capture. The timer test confirms actual completed disk output; it does not claim to observe the timer's exact mutex-entry instant.

## Remaining limits and authority

No full suite, native Mac run, user file, credential, release, ref, workflow or production integration was performed. This candidate requires new independent review. It does not solve external filesystem actor races, destination check→rename replacement by another process, or deletion's filesystem→CloseUnder interval before domain invalidation. It is an Editor transaction, not an inode-atomic filesystem protocol or proof of all migration parity. Native Mac runner0 failures remain NOT_RUN, independent of these tests.
