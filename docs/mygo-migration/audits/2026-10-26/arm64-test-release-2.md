# Requested arm64 release retry: 2026-10-26.2

The first marker run40271341712 failed validation at TestEditorDebounceAcrossTabsAndClosingLastEdited; native packaging and publication were skipped. Its intended .20261026.1 tag/release remained absent. The failed run and assertion are preserved, not overwritten or counted as success.

A same-source repeated original test passed5 times, which does not establish the original CI failure timing. A controlled real timer/write experiment independently proved the old fixed-sleep assertion could precede disk completion. The reviewed test-only correction preserves the original virtual time sequence, real file bytes, global debounce and closing-neighbor requirements, adds explicit write completion and rejects stale epochs. Exact2tests ran10 times each under race, with vet passing. No application production code was changed.

The existing fingerprint intentionally includes tests. Old1015-file fingerprint `f3200d7081d74d1c01369fddd2e925929494b71af1366d15865a8f7efd8d754e` becomes `4d55a9a74d8354a98442b979e06250ef905b24a964560d0ca4adeb4607f90189`; the only changed input is `internal/workspace/files_test.go`. All other inputs, including every non-test production Go file, match the reviewed a8 runtime. No fingerprint guard was relaxed.

The new marker/tag `macos-test-v0.2.12-go-test.20261026.2` runs the same complete validation and hosted arm64 DMG install/reopen/Pi/log gates before publication. It does not rerun an old SHA while claiming new tests. User-Mac picker hang, full functionality and visual parity remain unconfirmed/incomplete.
