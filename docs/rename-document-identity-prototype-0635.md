# Rename identity: bounded native-cache and callback protection (not atomic rename)

## Current integration result — 2026-10-26 07:13 UTC

This limited ID/cache/receipt change is now approved for the normal feature branch after the complete integrated gate; it is NOT in the delivered c040 .20261026.2 package and does not create a new release. The prototype-stage evidence below is historical, not the latest gate status. The final test-only truncated-viewer fixture uses real Editor.Open identity and preserves bounded content assertions; the earlier ID0 fixture panic remains recorded as a failed gate.

Final integrated gate: 1816 non-document file hashes unchanged; 100 critical package/test identities actually passed; full race suite 3369 test/subtest PASS events and 19 conditional skips, UI 354.696 seconds; toolkit 19 PASS events/zero skips; terminal 115 PASS events/4 conditional skips; vet and Darwin arm64 cross-build EXIT0. See rename-identity-integration-gates.json. No native Mac execution acceptance is inferred.

Filesystem rename still precedes domain remapping: the FS-to-remap gap, namespace read generations and Save/Close identity transactions remain unresolved. This is not atomic rename.

## Historical prototype record

Base HEAD898f37c801c9655e332ed4f085b9968126497722, tree d764fc66b02950b53cbaba8a9b83b21eca4fba04 (public c6e equivalent per coordinator). Scope approved: stable DocumentID, ID-owned native cache, fixed-ID deferred UI receipt. No release/ref/shared-source changes.

## Proven problem and current narrow fix

Fresh `rename-identity-red.jsonl` uses actual UI prompt/Async rename, holds UI callback after domain completion, recreates oldpath, and runs actual OpenFile. With and without a render before reopening, oldpath reused original native State before callback; old callback then moved newoldpath state into destination and removed newoldpath cache. Both subtests RED, package1.561s. Disk bytes remained correct at observation: not evidence of irreversible disk data loss.

Prototype:
- Editor assigns monotonic nonzero DocumentID to newly opened document; rename retains it, close/reopen allocates a new one. SetTextID and SetText share the same internal validation/history/dirty/debounce implementation.
- Native records own State, acknowledged version and image by editor-local ID inside filesState. Existing path maps are current indexes, not permanent aliases. Binding another document at oldpath never gives it the previous ID's State. Active/explicit existing-document opens bind by ID before deciding whether native seeding is necessary.
- Both rename entrypoints capture owner/editor, original target IDs/paths/native states and navigation generation before dispatch. Callback applies only captured IDs still at expected destination, with matching owner and original native state; it does not iterate arbitrary old-prefix states. Metadata bindings update even after new navigation, but focus/desired changes are suppressed for superseded navigation. Late pending native versions are not acknowledged by receipt.
- Closed native records are removed at actual close or pruned by live IDs during subsequent binding. No global registry or cross-editor IDs.

## Focused evidence

Final UI `prototype-final-focused2.jsonl`: 17 top-level / 23 test PASS events including six subtests, no fail/skip, race package3.203s. Includes two original identity REDs now GREEN, pending actual Buffer.Undo/Redo ownership and newer navigation, existing renamed-ID open before callback, existing subtree rename/caret/Undo and cancel/failure, preflight, watcher pending protections, read-only failed sync and native edit/copy behavior.

Domain `prototype-domain.jsonl`: 2 exact tests PASS, race1.018s: DocumentIdentityRenameReopenAndStaleEdit; RenameGapSnapshotLateEditAndRecoveryProbe. Domain shared setText implementation tested through new ID routing; closed ID cannot retarget a new oldpath document. UI/workspace vet PASS; whitespace diffcheck PASS. No full suite or native capture.

## Intentionally unresolved / next independent stage

Filesystem rename STILL precedes Editor.RenameUnder's ioMu; no reservation, namespace generation, load rejection or Save/Close intent-ID transaction has been added. `RenameGapSnapshotLateEditAndRecoveryProbe` intentionally still shows oldpath Save ENOENT in held post-FS/pre-domain gap; late domain text stays Dirty, snapshot is nonblocking, and explicit newpath save recovers it. This is not a held OS syscall or actual timer reproduction. Open/Reload stale-load and operation-error surfacing semantics remain unchanged. Domain RenameUnder still discovers paths at execution; fixed IDs constrain UI apply, NOT the entire domain rename.

Thus this patch closes native pre-callback reuse and wrong deferred UI movement only; **do not describe it as atomic rename or complete rename safety**. Further receipt/ID transaction scope requires review, including validation of the exact admission target set before FS. Existing delete confirmations and delete implementation are untouched; no toolkit changes.

Pinned Rust source299a32b has stable tab identity and entity continuity but its rename reconciliation scans paths too; this prototype protects identity without claiming stronger Rust runtime behavior was proved.
