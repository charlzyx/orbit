# Supplementary font measurements, 2026-10-25

The display-free specimens use the pinned real Linux text engines. No X11,
Wayland or native application window was started. These are not the two hosted
native screenshots and cannot retrospectively identify their selected fonts.

## Verified Latin faces

For both engines, `.ZedSans` requests 400 and 500 selected the bundled Plex Sans
regular face, index 0, SHA-256
`975dcda37d80f038dcd143c22e33ca2d97a0cc5a929aace1c749153b0fe1afa5`.
Requests 600 and 700 selected bundled semibold, index 0, SHA-256
`a20caf8286023a6a7a85e40b1d2a4ae9fc3e3b1f9eda8f4c542dd4986af67bb1`.
Regular/italic and semibold/semibold-italic behaved consistently. Lilex requests
400/500 selected regular and 600/700 selected bold, with matching loaded bytes
and face indices across the engines. MyGo's selected file hashes agreed with
its complete HarfBuzz face-blob hashes; GPUI observed actual CosmicText data.

Thus the legacy `IBM Plex Sans SmBld` family name alone is not evidence of a
runtime weight mismatch. Equal selected bytes do not establish equal synthesis,
hinting, shaping, positioning, or pixels.

## Actual metric differences despite equal Latin font bytes

Measured at size 14, upright `.ZedSans`, unconstrained single-line shape:

| String | Requested weight | Pango width | GPUI width |
| --- | ---: | ---: | ---: |
| General | 400 or 500 | 50 | 49.41999816894531 |
| This machine | 400 or 500 | 84 | 83.41200256347656 |
| General | 600 or 700 | 51 | 51.1140022277832 |
| This machine | 600 or 700 | 85 | 86.08599853515625 |

Pango reported ascent/descent 15/4; GPUI reported
14.34999942779541/3.8500001430511475. In the Go specimen's requested 1.3 line-height,
its actual layout used height 19 and local baseline 15. The GPUI specimen records
shaping metrics only; no baseline from a painted application was observed.

Width differences change magnitude and sign by string/weight. No global offset,
font-size adjustment or width multiplier follows from these observations.

## Fallback and limitations

The mixed string `AV office fi 0123456789 → 中文 🙂` produced different observed
fallback face sets. The traces preserve hashes and collection indices, including
Noto CJK/emoji and GPUI's additional italic fallback. Do not attribute this to a
single cause yet: Go registers the entire asset font tree; frozen Rust initially
loads nine faces, and language/locale selection also differs in the specimen.

The final probes produced 122 MyGo events and 445 GPUI events without drops.
Each renderer's 48-case shape/metrics JSON was byte-identical with tracing on and
off. Final diagnostic MyGo UI selection regressions passed with -race, and vet passed
for its text/UI packages. Eleven diagnostic-tool tests passed. The GPUI
diagnostic library compiled and its specimen executable ran. These checks are
not native application paint acceptance. `validation.json` records exact hashes.

Required next evidence: a separately identified hosted diagnostic run of each
actual app, equal verified scene inputs, actual painted baselines, effective
GTK/Cairo settings, same-renderer stock-versus-diagnostic native captures, and
complete unmodified full-image comparisons. Only after that evidence should a
minimal typography normalization be selected. No production normalization is
implemented by this directory.
