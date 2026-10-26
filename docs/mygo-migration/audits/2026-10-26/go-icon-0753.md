# Distinct Go application identity (requested by user)

Base `80c0421ef637c669d6eb5bd6fec5d59b25439cfc`, tree `5d49fa8ae01ab4a45afc4c035639caf515541888` (same reviewed/published ea43 source). This isolated change excludes the unreviewed rename transaction. No version/tag, bundle identifier, OrbitHome, session schema or runtime behavior change.

## Artwork and reference chain

The Go-only source `assets/icons/go/orbit-go.svg` derives the existing OP paths from `marketing/src/app/icon.svg`. Bright cyan tile, dark navy OP and yellow GO corner badge distinguish it from black Alpha at 16/32/64/128/512 pixels. All letters are paths; no installed fonts or external images. Original Rust/Alpha icons, macOS asset catalog, marketing logos and shared wordmarks remain unmodified. Inset remains 824/1024.

- Source SVG → `scripts/make-go-icon.py` → Go-only `orbit-go.icns` (11 standard/retina PNG slots from 16 through 1024) and embedded `crates/orbit-pi/assets/go-app-icon.png` (512 square RGBA).
- Current release path `scripts/macos-app/bundle.py` validates source/output hashes and every slot dimension, installs the exact new ICNS at `Contents/Resources/OrbitTest.icns`, checks copied bytes and retains its manifest under Provenance. Existing CFBundleIconFile=OrbitTest and CFBundleIdentifier=dev.orbit.go.experimental are unchanged. Legacy `scripts/macos-test/package.py`/Rust Alpha packaging are deliberately untouched.
- Go `LoadAssets` separately decodes the new embedded PNG; About uses it at the existing 48×48 footprint. Sidebar retains the original light/dark wordmark, adds an unselectable high-contrast GO badge without moving its existing measured x/y/width or New Task geometry.
- Generator renders vector with installed librsvg/cairo via ctypes; no downloads or iconutil needed. `--check` performs mandatory structural/hash verification without renderer dependencies. Regeneration test is conditional only where renderer libraries are absent; local regeneration executed and byte-matched all checked-in ICNS sizes and the embedded 512 PNG. Native OS ICNS rendering remains a hosted packaging/smoke concern.

## Verification and previews

Actual mounted light/dark Sidebar and About PNGs: `/workspace/shared/orbit-go-icon-preview-0754/{sidebar,about}-{light,dark}.png`. New/old size comparison: `old-new-sizes.png` in that directory. These were visually inspected, not inferred from code; root and user accepted the color/badge preview before wiring. Images are CPU-rendered app previews, not screenshots of a user's Mac.

Final exact hitset `go-icon-hitset.json`: three top-level UI tests plus both light/dark subtests each run/pass once under race, zero skip/fail, 7.050s. Original wordmark/NewTask and About identity/footprint assertions remain unchanged. UI vet and diff-check pass. macOS-app offline Python discovery: 36 tests PASS (including five new icon tests and a nested nine-test existing suite), no local skip. Actual temporary `.app/Contents/Resources` copy and manifest bytes asserted; all ICNS slots decoded structurally; Pillow also decoded all 11 sizes as a separate inspection. No actual release occurred in this worker.

Initial layout attempt placed badge in normal row flow, causing the unchanged wordmark geometry test to fail (`71,48.6645` instead of `86,46`). This is retained in `/tmp/orbit-go-brand-focused-0758.jsonl`. Corrected badge is an absolute noninteractive sibling of a fixed-width wordmark row; the original assertion now passes. Failed ImageMagick rendering due missing rsvg-convert executable was replaced by direct installed-library vector rendering, not by an alternate artwork or downloaded dependency.

Logs: `/tmp/orbit-go-brand-focused-final.jsonl`, `/tmp/orbit-go-brand-vet.log`, `/tmp/orbit-go-icon-package-final3.log`, `/tmp/orbit-go-icon-check-final.json`; Darwin arm64 cross-build passed (`/tmp/orbit-go-icon-arm64.log`). Final combined/package/native gates are still required for the new release source; previous release validation is not reused as approval of these new bytes.
