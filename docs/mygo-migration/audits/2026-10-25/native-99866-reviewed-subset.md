# Hosted evidence at 99866 — bounded subset

Commit `99866d8611aac61fcef9267e6d7cfc534c050944`, migration run [40233429813](https://github.com/charlzyx/orbit/actions/runs/40233429813): all five jobs succeeded. Standard CI [40233429762](https://github.com/charlzyx/orbit/actions/runs/40233429762): all six jobs succeeded. These results are not whole-feature or pixel-parity acceptance.

## Pi launch provenance regression

Required Ubuntu job `121289877644` executed exactly two real metadata-only tests in 9.924 seconds, with zero skips. Both renderers' controlled launch fixtures retained a verified owned ancestry chain although Pi changed its process command line to `pi`. The expected Node 22.23.4, fixed vendor archive, actual Go/Rust patch sources and independently prepared 1,280-file package inventories matched their prior hashes. Real metadata RPC succeeded. The tests explicitly report `NATIVE_UI_NOT_RUN`; no inference or authentication was used.

Artifact `12614011387` ZIP SHA256: `fbc52fc4c95531eaf6573e295bb0196d5c178a47cb32bff6f82dcaa58e895213`. This closes the hosted process-ownership proof gap, not the static Chat capture gap. Earlier zero-input capture failures remain recorded; they are not retroactively converted to passes.

## Original Rust parser execution

Reference job `121290016417` verified the original transcript source hash and verbatim parser excerpts, compiled them with the pinned Rust toolchain and executed the fixture. Output retained the third body cell despite two header cells: `header=["A", "B"]`, body including `[lost](https://example.invalid/extra)`, and `retained cells=5`. This is actual original-parser execution, not only source reading. It does not validate every Markdown construct. The later code-copy fixture extension was not part of this run.

## One actually inspected Go image

WB05 `wb05-review-uncommitted-dark.png` was opened and inspected. Its raw SHA256 is `085e9909de26e378835f1ecc3a7d98351b50976ccd198c805610ab6642f5ef99`, byte-identical to the inspected 0173 Go WB05 frame. Fixture metadata/hash, recipe, 1470×922 geometry, 14px font, 100 density and 96dpi matched, while application source/binary identities differed.

File-glyph fallback and +1/-1 / +1/-0 counters are visible. The staged and unstaged return text and fixture note still wrap as before. This single Go→Go frame is an eligible baseline for a later exact-source local Review glyph-policy comparison, subject to matching all state inputs. It is not a Rust pixel pass or an inspection of all captured states. Runtime owned Pango policy was not measured; the old shared-context observer is not valid evidence for it.

## Subsequent static Chat comparison: useful failure, not acceptance

The separately reviewed one-shot test commit `bfc1579f63d84a5b2385bf5b17e5594b29d35d02` preserved the 99866 runtime and ran [40234988331](https://github.com/charlzyx/orbit/actions/runs/40234988331). Its required receipt regression again passed two real tests with zero skips before either native job. Both renderers used the same controlled persisted session and only the two approved keyboard inputs.

Actual raw Go CHAT01 displayed the tail and `CHAT TAIL READY`. After Ctrl+Up, Go CHAT02 remained at that tail; the before-input and settled PNGs were byte-identical. Original Rust CHAT02 visibly displayed `PROMPT ELEVEN` and `RESPONSE NUMBER 11`. Thus this was not merely a shared OCR expectation failure. Go's CHAT02 is a **FAIL**, while Rust's two-state native job succeeded. Context remains **BLOCKED**, not captured through fake model/auth state.

Read-only source inspection found both renderers declare secondary-Up and Linux maps the Go modifier to Ctrl. In Go, the focused multiline editor can consume Up/Down before the later global shortcut handler. The existing direct command tests did not establish real focused-key routing. A separate mounted-key runtime correction is being investigated; neither this evidence nor the native run proves that correction yet. No OCR threshold or expected target was changed and no automatic rerun was made. This does not reproduce or resolve the user's separate macOS workspace-picker hang.
