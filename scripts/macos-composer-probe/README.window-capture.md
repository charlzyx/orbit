# Isolated own-window OS capture layer — PHASE CANDIDATE / NOT_RUN_NATIVE

This additive candidate is separate from the four-case driver and its content
rerenders. The candidate now wires the fixed phase handshake, but no native operation has been executed. No native action, permission query or screenshot was run locally.

The readonly Go CLI accepts `window-proof PID START_US WINDOW_NUMBER`. It checks
kernel PID/start/executable before and after, calls only the public readonly
CGPreflightScreenCaptureAccess (never Request), then
CGWindowListCopyWindowInfo with **optionIncludingWindow alone and the specific
nonzero window ID**. A non-singleton result is rejected without iterating it.
Only that dictionary's ownerPID/window number, onscreen flag and bounds are
read; titles/other windows/full screen are not read. Wrong owner/ID is rejected
before bounds. Missing permission or hidden window is BLOCKED.

OwnWindowReceipt is a main-thread-only unwired primitive reading the existing
owned NSWindow's number, visibility and frame, with current-process start identity
checks. A native pointer is NOT a window number. The phase binding calls this in the actual app with nonce/phase/time; the Python callback is not an authorization source by itself. The complete harness must be reviewed again before it is run.

Python accepts only before/seed/final/undo under the nonce's fresh private root;
requires before/after independent CLI proofs and fresh owned app receipts; and
uses exactly `/usr/sbin/screencapture -x -l ID phase-os.png`. Fixed ID and private
path are not caller-selectable capture scope. Failure stops, with no content
rerender fallback. CG bounds and Cocoa frame use different Y origins: compare
each source before/after independently, and cross-check only width/height.
No inferred Y conversion or screenshot tolerance is used. Capture output is bounded (16 MiB encoded, each dimension <=4096, <=16,777,216 pixels). Every PNG chunk CRC and complete IDAT/IEND are checked; Pillow verify() followed by reopen/load() actually decodes pixels read-only. Pillow 12.3.0 accepted a synthetic short-scanline PNG, so a bounded zlib byte-count/EOF/no-tail guard additionally requires exact noninterlaced scanline length for legal PNG bit-depth/color combinations. Adam7 is explicitly unsupported/BLOCKED, not silently converted; no filter/pixel codec is implemented here. Decoded PNG pixel dimensions are independent evidence, not equated to NS/CG window bounds (Retina/shadow). Screenshots corroborate appearance, never replace the state oracle.

Public API references:
- https://developer.apple.com/documentation/coregraphics/cgwindowlistoption/optionincludingwindow
- https://developer.apple.com/documentation/coregraphics/cgpreflightscreencaptureaccess()

Evidence: 34 offline/mock Python cases, including rejected foreign owner/window,
permission, stale receipt, coordinate changes, unsafe output and failed command.
The original Pillow diagnostic (27 PASS / 1 FAIL) is retained at /tmp/orbit-window-capture-pillow-diagnostic; final 34-case evidence is /tmp/orbit-window-capture-pillow-final. Local Pillow is 12.3.0; no dependency was installed. A future reviewed workflow must explicitly pin the official Pillow test dependency. A direct spawned verifier is terminated and reaped with a bounded 2-second cleanup; failure records its PID and BLOCKED, no SIGKILL or unknown-PID signals.

Darwin ARM64 CLI cross-compilation and vet are not native execution. Swift helper
still requires actual hosted compilation. Same-UID interference and a window
change between checks remain TOCTOU limits; this design assumes a dedicated
credential-free runner, does not freeze the WindowServer, and cannot bypass
screen-recording permissions. No codesign/quarantine/TCC changes are included.

## Fixed phase handshake

After each content rerender, the app proof goroutine publishes a one-shot ready token for before/seed/final (plus undo only for text). The runner requests a fresh before receipt, independently verifies the exact CG/kernel window, captures that one window, requests a fresh after receipt, and only after successful proof/decode writes the PASS acknowledgment. The app does not continue to the next allowed input until that acknowledgment. Tokens bind nonce/PID/start/window number/phase, unknown or duplicate JSON fields are rejected by the app, and each artifact is created without overwrite in the private root. Phase operation deadline is 10 seconds inside the existing case deadline; waiting for ordinary initial app readiness retains the existing case limit. No UI thread waits on the runner. Any failure prevents later input and never substitutes the content rerender for the OS capture.

No workflow is added. A future reviewed job must keep an exact ee258/tree047611 baseline worktree, import only hash-verified test scripts from a separately reviewed harness commit, and record both source identities; prepare.py's exact HEAD tree check remains unchanged.
