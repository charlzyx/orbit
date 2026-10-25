# Actual-byte attachment file bound

This separate delta follows the approved attachment ingress FIFO; it does not
change image decoding, clipboard admission, caret insertion, or origin policies.
The old Stat <= 64 MiB followed by os.ReadFile was not an authoritative read
bound: the file can grow, and nonregular inputs can have misleading sizes.

The production limit remains 64 MiB. Regular-file policy is checked before open
and on the opened handle; symlinks resolving to regular files remain accepted.
The actual reader consumes at most limit+1 bytes, rejects the sentinel byte and
returns no partial bytes on failure. Existing size-error copy stays the same.
Drop/general-file origins keep their source-matched path fallback for read errors;
the strict image origin keeps its error/no-path behavior.

Darwin/Linux use O_RDONLY|O_NONBLOCK only for open, preventing a stat-to-open FIFO
replacement from waiting for a writer. Regular file semantics are unchanged.
Other platforms retain plain os.Open plus both regular checks and byte bound;
there is no claim of closing their special-file replacement window. No platform
gets a promise that arbitrary filesystem syscalls or network mounts can be
cancelled. Rejecting nonregular input and the size limit are deliberate Go safety
differences, not claimed original Rust constraints. Decoded pixels/Base64/SVG
memory and io.ReadAll buffer capacity are outside this raw-file input bound.
The 64 MiB limit is accepted raw bytes (plus one rejection sentinel consumed),
not an overall resident-memory budget.

Tests use limit=8 rather than repeated 64 MiB allocations. The extracted old
unbounded reader seam RED accepted 9/32 bytes and consumed all 64 simulated grown
bytes; it did not test a user file or execute an installer. Green tests exercise
the production reader with exact-limit, over-limit, a bounded growing reader,
regular symlink, directory rejection, growth between first stat/open and actual
Darwin/Linux regular-to-FIFO replacement with no writer. No physical native
picker confirmation or user-Mac acceptance is claimed.
