# Autosave debounce test settlement (test-only)

Base `c03f0ea7faf54eedc1fb33aaed7059ddb589760b`; release marker/source runtime unchanged by this candidate. Only `internal/workspace/files_test.go` changes test behavior. No production, sleep-duration, timeout, file-deletion, release/ref/workflow modification.

## Evidence and limits

Preserved hosted failure: `/workspace/shared/orbit-release-failure-0507/logs/root-race.jsonl`, run 40271341712. `TestEditorDebounceAcrossTabsAndClosingLastEdited` read `a` instead of `changed a` at files_test.go:228. Assertion timestamp was about 192 ms after start; failure elapsed 0.46 s includes cleanup. This does not prove the timer's exact schedule or where I/O was then blocked.

Fresh same-source/runtime original test: `baseline.jsonl`, race count5, five PASS, package1.979s. The actual CI failure was NOT reproduced.

Controlled proof `controlled-red.jsonl` plus `controlled-red-probe.go.txt`: real AfterFunc/flushAll enters the existing writeFile seam; only the neighbor write is held on a channel. The original 60/60/70 ms timeline is preserved using testing/synctest. At the old assertion point actual timer has entered, disk still reads `a`, and original expectation fails (package0.020s). This proves sleep/deadline is not a write-completion barrier, not that CI had precisely this controlled delay. The failing probe is archived outside Go package discovery, not left as a disabled/skip test.

## New executable contract

- Original global debounce scenario keeps the exact 60/60/70 ms intervals with virtual time; at120ms the second edit's160ms deadline has not arrived, checked both by no entered-write signal and real unchanged disk bytes.
- Close(b) remains real and synchronous. At190ms the actual timer must have entered a's write; deliberate hold demonstrates incomplete disk state. Release, receive explicit completion AFTER SaveExisting returns, then synctest.Wait settles save bookkeeping. Only then assert a=`changed a`, b=`changed b`, active a clean and not Saving. Real filesystem completion is not assumed from synctest.Wait alone.
- Independent epoch test calls real flushAll with the old token (both disk files unchanged), then current token (both dirty files written). No bypass of actual disk assertions.
- Final exact hitset: TestEditorDebounceAcrossTabsAndClosingLastEdited and TestEditorAutosaveEpochRejectsOldAndSavesAllDirty, each ten race repetitions: 20 PASS, zero skip/fail; package1.063s. UI/full aggregate not run. Workspace vet and diff whitespace check PASS. Logs/hashes in local proof.json.

A successful focused gate does not establish the CI failure's precise scheduler cause or final release validation; parent owns integration and aggregate rerun.
