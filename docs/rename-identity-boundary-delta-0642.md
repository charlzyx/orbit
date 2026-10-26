# Rename identity boundary correction

## Current integration result — 2026-10-26 07:13 UTC

This limited ID/cache/receipt change is now approved for the normal feature branch after the complete integrated gate; it is NOT in the delivered c040 .20261026.2 package and does not create a new release. The prototype-stage evidence below is historical, not the latest gate status. The final test-only truncated-viewer fixture uses real Editor.Open identity and preserves bounded content assertions; the earlier ID0 fixture panic remains recorded as a failed gate.

Final integrated gate: 1816 non-document file hashes unchanged; 100 critical package/test identities actually passed; full race suite 3369 test/subtest PASS events and 19 conditional skips, UI 354.696 seconds; toolkit 19 PASS events/zero skips; terminal 115 PASS events/4 conditional skips; vet and Darwin arm64 cross-build EXIT0. See rename-identity-integration-gates.json. No native Mac execution acceptance is inferred.

Filesystem rename still precedes domain remapping: the FS-to-remap gap, namespace read generations and Save/Close identity transactions remain unresolved. This is not atomic rename.

## Historical prototype record

Base prototype f8789e719f1c4f30ce077743502caaf6c231f530. Domain transaction/FS gap still unchanged; this delta closes UI binding and navigation defects found by independent review.

## Fresh RED evidence

- `identity-boundary-red.jsonl` FAIL1.756s: syncDocument with snapshot of closed ID recreates closed record; snapshot of same ID at now-old path recreates old path index; real queued OpenFile completion overwrites pre-existing native pending text and acknowledges it.
- The Open pending fixture explicitly seeds native adapter state after actual OpenInactive worker completion, then edits that buffer before applying actual queued callback. It proves helper/callback contract, NOT an OS-key focus sequence avoiding navigation-generation changes.
- `identity-open-red.jsonl` FAIL1.546s: real OpenFile worker registers oldpath, callback held; real FS/domain rename + receipt apply; oldpath recreated and actual new document registered. The old Open callback then activates replacement ID and does not invalidate/remap accepted old intent. This is a regression introduced by omission of the previous generation increment.
- Initial stale-path RED used one path change after snapshot. Final renamed_again regression captures after first rename and rejects after second rename of same ID.

## Correction

- bindFileNative and captureFileNative reject non-live IDs or mismatched current path before creating/pruning caches. syncDocument aborts if rejected; for a valid ID/path it reads current domain metadata rather than installing older text/content from a queued copy.
- syncDocument retains any existing native buffer whose version differs from acknowledged version; neither Buffer.Set nor path/ID-record acknowledgement occurs. Includes readonly/error modes.
- OpenFile background completion validates captured files owner and returned document ID still at requested path before activating; cannot retarget a newly reopened oldpath document.
- Receipt still binds original live ID metadata irrespective of newer navigation. Only when the admission navigation generation is still current and desired matches this operation's old target does it remap desired and increment generation once, invalidating older pending Open callbacks. Empty native-state targets also receive this navigation update.

## Coverage/limits

Focused includes actual close then stale snapshot, same-ID second rename, pending queued-open seam, queued-oldpath replacement navigation, readonly/missing-file ID acknowledgement, real PNG source cache separation across rename/recreate/close, all prior original-ID/native Undo and navigation tests, existing rename/watcher/version tests. Logs and exact hitset in identity-delta-proof.json; no full suite, capture, integration, ref or release action.

No claim of global atomicity between domain snapshots and UI bindings: filesystem transaction/reservation, Open/Reload generation checks and Save/Close intent binding remain separate unimplemented stage. No permanent path alias, no state mutex over filesystem I/O.

Final exact gate: `identity-delta-final2.jsonl` 22 top-level / 32 PASS test events (10 subtests), zero fail/skip, race3.411s. Final UI/workspace vet and diffcheck PASS. No production changes outside these UI adapter boundaries.
