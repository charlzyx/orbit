# Local Review text compatibility

Pinned MyGo `9c69cf829ece`; exact manifest SHA256
`9b8d49dcb3f95af6aef6e2e2d34fe7e6afcb3f2f16e8b8353cc1cef39dadd86c`.
Prototype patch `d9d2f0547a8b08a6144d43065fadba5ffac20db45791215f6fb7c43573ecde79`.

Canonical `scripts/go-native.sh` verifies these eleven explicit source paths,
patch/license/baseline/overlay hashes and composes them with existing composer and
selection compatibility. It uses a temporary source copy; module cache is never
modified. No upstream upgrade. Unreviewed caller replacements remain rejected.

The new API defaults off. Only main Review context/addition/deletion rows opt in,
including their line number and syntax runs, after native capability resolution.
Linux shaping uses an owned temporary context with glyph rounding off and hint
metrics on. No shared context toggle. Non-Linux and unsupported Pango use normal
measure/render and report a fixed unchanged/unsupported policy status internally.
Headers, metadata, embedded Git, composer/editor/chat are not opted in.

The older process-wide `--diagnostic-text-metrics[-observed]` modes conflict with
this exact Pango replacement and are now explicitly rejected by the canonical
wrapper. Their historical source/evidence remains, but cannot silently override
the local policy or substantiate current native font options. No build flag or
runtime global enable is introduced.

Known prototype gate exception: unmodified upstream `TestDiffFrameAllocs` fails
at steady allocations=3 (expected<=2) under this Go1.27.3/race environment on both
pristine pin and candidate. Original assertions and logs retained; the full UI
suite is not called green. Other763 toolkit UI tests passed in two original-
timeout processes, full text race, composed composer/selection, vet and arm64
compile passed. This is not hosted application/native input acceptance.
