# Wave 7: target and content safety, source UI corrections

Feature-branch source only; `.4` installer remains unchanged. The included patches were independently reviewed and combined against wave6 (`918805`). Exact patch hashes and final gate results are in [wave7-integration-evidence.json](wave7-integration-evidence.json).

- Palette Delete binds displayed target/path and one-shot confirmation identity; owner changes, replacement/Cancel and held-pointer retargets are rejected. Synthetic tests capture actions; no real session was deleted.
- Palette Pin uses the current saved path rather than an empty default. Idle Abort is hidden according to Rust's busy projection; wire guard and Escape behavior are unchanged.
- Inline image slots now invalidate on immutable content/MIME changes, removed/re-added slots and ABA. Callback and click guards check both owner and content revision; preview provenance prevents closing unrelated Markdown previews. No decoder/network change or global cache-budget claim.
- Review file headers show source-style glyph fallback and additions/deletions, including zero. Nerd glyph availability and native font metrics remain separate; this is not full pixel parity.
- npm-only setup migration UI uses an explicit captured confirmation and safe coordinator. Dependency refresh and migration are mutually exclusive through actual receipt settlement. Tests use fixture backends only; **no actual installer, authentication or migration was executed**. Dependency readiness, managed-Pi verification and PATH shadow warnings remain distinct.

## Combined final gate

581 Go/module hashes were frozen and rechecked unchanged. Whole-repository vet and race tests passed; UI 587.024 s, native 86.430 s, workbench 18.277 s. Darwin arm64 cross-build passed; final command chain exit 0. UI duration approached the test-process limit under concurrent cloud suites; no assertion or business timeout was weakened. Future batches should avoid redundant concurrent full suites.

Exact-final-source hosted native capture is pending this publication's normal CI. The previous wave6 normal run passed all four jobs and both raw native archives were verified; two actual reviewed states and remaining wrap differences are documented in [native-918805-reviewed-subset.md](native-918805-reviewed-subset.md).

Not included: Rename prompt target correction, visible-model refresh, removal of dead legacy runtime UI, Pi notification/source lifecycle refinements, ragged Markdown table preservation, process-launch receipt harness, or Review fractional glyph policy. Those remain isolated next-wave work and cannot be counted as delivered here.
