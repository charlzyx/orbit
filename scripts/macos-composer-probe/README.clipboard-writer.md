## External synthetic clipboard writer (test-only, uncompiled)

**Status: NOT_COMPILED / NOT_RUN_NATIVE.** Authored on Linux without Swift or a
macOS SDK. The offline Python tests validate the fixtures and contract model,
not AppKit runtime behavior. Hosted execution remains gated on independent
review and the fully verified immutable Provider base.

`clipboard_writer.swift` is an external, long-lived AppKit process. Its sole
argv argument is `png-text`, `tiff-text`, or `text-only`; it also requires the
explicit `ORBIT_COMPOSER_DISPOSABLE_RUNNER=1` opt-in. This opt-in is not proof of
runner isolation: the enclosing admission gate must establish a fresh dedicated
macOS runner. Do not run on a user's Mac. It accepts no paths or input payloads.

One eager `NSPasteboardItem` carries fixed `clipboard-text` plus either the fixed
PNG or TIFF bytes (or text only). TIFF mode never supplies PNG. The 2x2 fixture
is opaque red/green on row 0, blue/white on row 1. RGBA SHA-256:
`c21b35e3f28e676cedf24c13575a7346682e101a2d26aad9598d0cdbcee9ee3b`.
Static encoded hashes and exact pixel checks live in `writer_contract.py` and
`test_writer_contract.py`. No image encoder runtime variability is involved.

After writing, stdout emits one JSON line with `schema:1`, `status:WRITTEN`, own
`pid`, `mode`, `changeCount`, actual advertised `types`, `fixture`,
`cleanupToken`, and `receipt`. Fixture metadata includes encoded length/hash,
2x2 dimensions and row-major RGBA bytes/hash; text-only includes only text hash.
The harness must validate against its child PID, mode and expected metadata,
retain the exact process/record pair, bound stdout to 16 KiB and use a deadline.
No error/timeout is success; no alternate payload or input retry is allowed.

Cleanup is a single stdin command `cleanup <cleanupToken>\n` to that same live
process. There is deliberately **no standalone cleanup CLI** and no caller-
supplied generation. Authority is the token and generation retained in memory.
The audit receipt is freshly created via `mkdtemp` (0700) and O_EXCL/O_NOFOLLOW
(0600) under fixed `/private/tmp`; it is never read or trusted for cleanup.
A forged receipt cannot claim an ambient generation. Changed generation or bad
command refuses clearing. EOF refuses clearing; helper death does not authorize
a different process to clear. On normal completion its receipt directory is
removed; failed/terminated helpers may leave only their private synthetic audit
receipt for ephemeral runner disposal. No ambient clipboard content is read,
saved or restored. Cleanup emits `CLEARED` or `CLEANUP_REFUSED`; failure is not
permission to force cleanup or retry native input.

AppKit provides no atomic compare-and-clear or compare-and-write. The metadata
checks detect races but cannot prevent a concurrent writer between check and
mutation. Dedicated runner isolation is mandatory; this is not a security
boundary against malicious same-UID processes.

Offline check (safe on Linux, no subprocess or clipboard access):

    PYTHONDONTWRITEBYTECODE=1 python3 -m unittest -v test_writer_contract.py

The process protocol unit test uses only a mock process and in-memory stdin.
Swift source assertions are not a compile or native proof. Compiling against
public AppKit/CryptoKit and exercising the native protocol are future gated
macOS tasks, not actions performed by this implementation.
