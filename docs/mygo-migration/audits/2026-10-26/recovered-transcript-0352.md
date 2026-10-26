# Transcript reconstruction after workspace loss

## Provenance and scope

Baseline: `cee007d9309815489bfaa11529fa3047a2b612f8` (Wave10).
Only `internal/transcriptui/selection.go` and `menu_intent_test.go` were restored byte-for-byte from surviving Git objects in recovery commit `b41d2c2058c0316194a5af9f3672b0c5a959c988`. The UI implementation and tests in this change are newly reconstructed, not recovered old Wave11 bytes. Earlier Wave11 gates do not validate this new tree.

Pinned Rust reference: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `crates/orbit-pi/src/transcript_view.rs` constants at 118–121 and code-fold rendering around 6600–6710. Settled code above 28 lines displays a 24-line preview; streaming remains unfolded. MyGo reference is the existing pinned `9c69cf829ece196c5b3a3601f1c7bbe70526376f` with the project's canonical compatibility overlays; no toolkit source change is introduced here.

## Reconstructed behavior

- Actual inline/task owner generation keys the selectable transcript host, clears hover on owner reset, and guards Copy activation. The exact recovered menu implementation snapshots message/selection at opening and uses fixed public Collect/Chosen items, so same-owner streaming does not retarget an already-open menu.
- Top-level fenced code folds only when settled and above the source threshold. Expansion state follows root-block ordinal within the owner; source/document revisions invalidate derived metadata and removed ordinals are pruned. Streaming, nested alert/quote and indented-code paths remain unfolded.
- A collapsed prefix is a cached 24-line literal fenced document, with a tilde delimiter longer than any payload run. Its renderer contains only a code-body callback: no app Block recursion, image loader or HTML hook. Expanded/live content uses the original Body. Copy always uses the original full block, including hidden lines.
- Fold controls use ordinary selection-clearing pointer/keyboard behavior. Tests exercise a real virtualized App List, visible nonzero RichText bounds, growth/shrink, retained scrolled anchor and live-tail behavior.

## New verification, not historical acceptance

Canonical `scripts/go-native.sh test -race -count=3 -json` scoped to UI/transcriptui passed: UI 10.438 s, transcriptui 1.065 s. The accompanying hitset explicitly checks 29 complete package/test names each ran and passed three times (65 unique tests/subtests actually observed), zero fail/skip. This includes both original caption/locale tests and original code-copy regressions. `scripts/go-native.sh vet ./internal/ui ./internal/transcriptui` and `git diff --check` passed. A new combined integration/full gate is still required; no native OS/pixel acceptance is claimed.

Raw final JSON: `/tmp/orbit-recover-transcript-final3.jsonl`, SHA256 `5682f6a1bb144c75ab5ba2900f691e74f62d1556f51a68062def3c3bfc1a8314`; vet `/tmp/orbit-recover-transcript-vet3.log`. Exact checked names/counts are in `recovered-transcript-hitset.json`.

## Recovery failures retained

1. Initial test compile used nonexistent `Action.Name`; fixed to the real `Kind` field. Original `/tmp/orbit-recover-transcript-menu-first.jsonl` remains.
2. First List assertion queried a zero-sized generic header from overscan/prepass. The fixture now uses a uniquely identifiable actual visible code RichText and still asserts height growth/shrink and anchor retention. Original failure `/tmp/orbit-recover-transcript-final.jsonl`; isolated corrected List test passed three times.
3. A new selection test queried a retained Element outside its current frame and panicked. It now samples SelectedTexts inside the frame, retaining the assertion that folding clears old selection. Original `/tmp/orbit-recover-transcript-final2.jsonl` remains; final3 is the complete corrected run.
4. An earlier regex did not match the actual two caption-test names. That claim was explicitly withdrawn, not treated as PASS. The final exact-name manifest includes both tests and verifies three actual passes each; Chinese locale assertions are unchanged.

## Remaining boundaries

Blank-space/right-click hover hit targeting is unresolved and unchanged. This is not full transcript or pixel parity, not a global cache/memory bound, and not physical native selection/IME acceptance. Nested folding is intentionally outside this bounded top-level change. No original message bytes, remote loaders, authentication, model requests, user files or release artifacts are changed.

## Independent review correction (03:55 UTC)

The first reconstructed candidate 9380a510 was not approved for integration. Independent review identified three deviations from the earlier source contract: unnecessary preview preparation while streaming/disabled; expansion state pruning while a fenced block temporarily dropped to 28 lines; and a fixed collapsed-prefix key that could reuse selection identity across source replacement. The new candidate gates metadata preparation on eligible settled folding, retains expansion state for every top-level fenced ordinal, and revision-keys the collapsed prefix. Three new tests cover ineligible metadata absence, 29→28→29 expanded roundtrip, and actual selected preview replacement with no remapped selection. Footer now uses the existing Translate(count) path and Rust's 30px height, density-scaled 12px horizontal padding, top border, 6px gap, 12px medium label and chevron/hover colors; this is source restoration, not a redesign.

Final review run: `/tmp/orbit-recover-transcript-review-final3.jsonl`, UI 9.500s; exact manifest now contains 32 top-level package/test names each run/pass 3, 68 unique test/subtest names, no skips/failures. Vet `/tmp/orbit-recover-transcript-review-vet.log` and diff-check passed. Earlier 29-name results remain historical scoped evidence, not the final revised tree gate. Two intermediate review-test compile errors (wrong SelectedTexts slice type, then its formatting verb) are retained in review-final/review-final2 JSON logs; neither was a runtime or passing gate. New combined/full integration gate remains pending.
