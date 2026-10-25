# Native font fidelity diagnostics (Linux only, opt-in)

This directory adds **observations**, not typography fixes. It is not imported by
the app, canonical Go wrapper, or stock Rust build. It never starts a display or
window. Native application capture belongs to an explicitly identified hosted
CI diagnostic job, separate from the stock reference capture.

Pins: original Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, GPUI `0.2.2`,
MyGo `9c69cf829ece196c5b3a3601f1c7bbe70526376f`, Go `1.27.3`, Rust `1.94.0`.
The GPUI diagnostic copy adds an exact `sha2 = 0.10.9` dependency already present
in Orbit's pinned lockfile. No renderer dependency is upgraded.

## What is observed

- MyGo: final Pango runs after fallback; actual `item.font`; Fontconfig file,
  face index, family/PostScript name, numeric weight/slant, variation string,
  embolden and matrix when available; resolved Pango weight/style; actual loaded
  HarfBuzz face-blob SHA-256 and its index; selected-file SHA-256 separately.
- MyGo: effective per-run requested style, glyph IDs/rune clusters/coordinates/
  advances, context options/language, host font-rendering setting inputs, and
  actual Cairo scaled-font options/matrix. Painting records line boxes, actual
  logical baseline and the already computed physical baseline, scale and clip.
  Glyph bounds are scene bitmap rectangles **before any inline transform**.
- GPUI: original requested `TextRun` inputs and final CosmicText glyphs **after
  the fallback FontId correction** and existing emoji skip. Actual
  `CosmicTextFont::data()` SHA-256, fontdb face index/family/PostScript name/
  weight/style, glyph clusters/positions/advances and line metrics are recorded.
  Paint hooks observe actual baseline origins including wrapping/alignment.
- Fontconfig weight units are not CSS/Pango weight units. GPUI byte clusters and
  MyGo rune clusters are explicitly different. The observer never substitutes
  requested family names for an unavailable resolved face.
- Presentation batches are flushed only after the native presentation call
  returns. This is **not** a GPU fence, compositor completion, or scanout proof.
  It is a process-local present-call sequence, not the app's semantic readiness
  signal. Use one native window/process and a separately verified scene/frame
  correlation. The configured case ID alone is not proof of loaded scene state.

Cairo scaled-font records are emitted when a scaled font is created after its
face was selected by an allowlisted shape. A previously cached scaled font may
have no such record; do not infer its options from a missing record. Start a
clean process for each diagnostic capture and require the expected records.

Only exact strings in the explicit JSON allowlist are traced. The configuration
requires all of `ORBIT_FONT_TRACE_DIR`, `ORBIT_FONT_TRACE_TEXTS`, and
`ORBIT_FONT_TRACE_CASE`; otherwise the observer is inert. Limits are 64 strings,
512 UTF-8 bytes per string, 64 KiB configuration, 128-byte case ID, 256 fonts,
4,096 unique events and eight queued batches. File writes run on a bounded
background worker; dropped observations are reported. Files are created
exclusively, privately, and never overwrite existing output. Kill only after the
expected batch is complete and readable. Repeated records are deduplicated, so
missing repeated paint events do not mean a frame did not repaint.

Hashing loaded bytes and collecting JSON still adds CPU work. This is not a
claim of zero timing perturbation. Before relying on application traces, compare
stock versus diagnostic captures of each renderer in a settled static scene,
without cropping, shifting, masking, blurring or relaxing the exact-pixel gate.

## Prepare isolated source copies

Use a clean Git checkout of the exact MyGo pin and the exact crates.io GPUI 0.2.2
source. The preparer verifies aggregate full-tree hashes and explicit patched
file hashes before and after copying. A source drift, repeated output directory,
or overlapping input/output root is rejected. It does not modify originals or
module caches. The original licenses are retained in each source copy.

```sh
python3 scripts/font-fidelity/prepare.py --diagnostic \
  --mygo-source /path/to/clean/mygo-pinned-checkout \
  --gpui-source /path/to/registry/gpui-0.2.2 \
  --output /tmp/orbit-font-diagnostic
```

The diagnostic MyGo copy includes the same hash-verified composer and selection
compatibility additions used by the canonical build. Both runtime guards remain
present. The canonical wrapper intentionally rejects arbitrary MyGo overlays;
**do not weaken it**. A diagnostic build uses its own copied modfile:

```sh
out=/tmp/orbit-font-diagnostic
cp go.mod "$out/orbit-diagnostic.mod"
cp go.sum "$out/orbit-diagnostic.sum"
go mod edit -modfile="$out/orbit-diagnostic.mod" \
  -replace="github.com/egoist/mygo=$out/mygo" \
  -replace="github.com/egoist/mygo/orbit-terminal=$PWD/third_party/mygo-terminal"
CGO_ENABLED=0 GOWORK=off go build -modfile="$out/orbit-diagnostic.mod" \
  -o "$out/orbit-go-diagnostic" ./cmd/orbit
```

For Rust, use another export of the frozen original, not its stock capture tree.
Append the following to that diagnostic export's root Cargo.toml, with the real
absolute path. Keep its resulting lockfile and build provenance. The only
production-source change is in the temporary dependency copy; the exported
Orbit application's implementation remains unchanged.

```toml
[patch.crates-io]
gpui = { path = "/tmp/orbit-font-diagnostic/gpui" }
```

Use a separate `CARGO_TARGET_DIR` for the diagnostic app binary so it cannot
replace the stock `orbit-pi` executable. Preserve the original capture job,
source pin, binary and all original PNGs. Save `diagnostic-provenance.json`,
actual binary hashes, source/export identity, full build commands, toolchain and
native library versions, final lockfile hashes, and the trace environment.
The preparer's manifest describes dependency source copies, not a built binary.

For hosted native execution, create an empty trace directory and set:

```sh
export ORBIT_FONT_TRACE_DIR=/path/to/new/private/trace-directory
export ORBIT_FONT_TRACE_TEXTS=/path/to/scripts/font-fidelity/fixture-strings.json
export ORBIT_FONT_TRACE_CASE=general-static-diagnostic
```

Do not pass a diagnostic binary through a capture path that silently labels it
as the unmodified original. The hosted coordinator must explicitly mark it
`diagnostic: true`, retain its overlay/binary identities and match its observed
scene/readiness/viewport/focus with the stock run. No native display was started
while developing these tools.

## Display-free supplementary specimens

These execute the actual pinned Linux text engines, but **not the native app**.
They cannot establish application GTK settings, element placement or previous
capture provenance. Both specimens cover 400/500/600/700, upright/italic,
`.ZedSans`/`.ZedMono`, Latin labels, digits/ligatures and Unicode fallback.
Go uses the actual app's lexical full-font-asset registration order; Rust loads
its nine initial bundled fonts in the frozen app's order. The Go specimen sets
language `en` explicitly; GPUI records its actual FontSystem locale. Preserve
these distinctions when attributing fallback differences.

Go: set `ORBIT_FONT_TRACE_ASSETS` to the original `crates/orbit-pi/assets/fonts`
root and `ORBIT_FONT_SPECIMEN_RESULT` to a new JSON result path. Use an isolated
HOME/XDG cache. Run in the prepared MyGo copy:

```sh
env -u DISPLAY -u WAYLAND_DISPLAY go test -race ./internal/text \
  -run '^TestOrbitFontSpecimen$' -count=1
```

Rust: in the separate diagnostic original export, add a temporary
`crates/orbit-pi/src/bin/orbit-font-specimen.rs` containing only:

```rust
fn main() { gpui::orbit_font_specimen(); }
```

Run `cargo run -p orbit-pi --bin orbit-font-specimen` with DISPLAY and
WAYLAND_DISPLAY unset and the same specimen variables. This entry point only
constructs `CosmicTextSystem`, never `Application` or a native window. It does
not require GPUI dev-dependencies. Its batch explicitly says no native present.
Do not mistake it for a native capture or ship this temporary binary.

For each renderer run the same specimen with the three trace variables unset,
then compare its result JSON byte-for-byte to the trace-enabled result. That
checks shaped metrics were not changed; it does not replace native same-renderer
stock/diagnostic screenshot invariance.

## Summaries and checks

```sh
python3 scripts/font-fidelity/test_tools.py -v
python3 scripts/font-fidelity/analyze.py /path/to/rust-trace /path/to/go-trace \
  --output /path/to/summary.json
```

The analyzer rejects empty/incomplete traces, malformed hashes, conflicting
face identities and mixed cases. Missing hashes and dropped events remain
explicit problems. It compares observed face sets per exact source string, not
one-to-one run/weight/span attribution. Same face sets never certify equal
metrics, baselines, rendering options, synthesis, native state or pixels.

See `findings.md` for measured supplementary results and remaining evidence.
Do not normalize font weights, invent offsets or replace either native text
engine on the basis of source assumptions or these specimens alone.
