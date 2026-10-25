# Main Review local advances — isolated candidate

Source-first scope: Rust `diff_view.rs:278–323,380–464` applies code font metrics
on code-row ancestry; Go main Review unified and split share `reviewCodeRow`.
The number/body geometry (52px gutter, 12/10 body padding, 12.5 scaled font,
requested16 line height/min19) is unchanged. Native Linux Pango rounded each
7.5px Lilex advance to8; the23-character fixture therefore measured184 rather
than172.5 and wrapped at178px. Turning off glyph rounding alone retains hint
metrics and natural17px height. No shrinking, letterspacing or width compensation.

## App boundary

`review_text_policy.go` resolves capability on the UI owner before either width
measurement or layout. Its immutable fixed status is `enabled`,
`unsupported-native-api`, or `unchanged-platform`. Unsupported hosts remain on
ordinary methods; they do not invoke the optional API or panic at startup.
`reviewCodeRow` alone inherits the policy; `reviewUnwrappedWidth` measures only
matching context/addition/deletion kinds, including hidden gap lines, using the
same method. Cache identity includes the effective policy as well as existing
snapshot/generation/font/size. Metadata/header/fallback text remains ordinary.

Embedded Git has separate `paintDiffLine`/`unwrappedDiffWidth` and still uses its
older12px/two-number-column geometry. It is deliberately untouched, not claimed
aligned by this change. Main Review does not gain new selectability.

## Packaging and evidence

`third_party/mygo-review-text` preserves the reviewed11-file prototype verbatim,
including independent Pango-context lifetime, inherited layout/cache policy and
selection/hit/caret tests. Canonical wrapper verifies exact manifest/source/hash
and composes existing overlays; old global diagnostic combination is explicitly
rejected rather than allowed to supersede it. No cache writes/upstream upgrade.

App tests cover unsupported/non-Linux admission, ordinary metadata, cache policy
invalidation, hidden lines, unchanged geometry, normal neighbor and existing
unified/split/wrap/pan/sticky regressions. Original source-geometry assertions
remain; its expected text width now uses the actual same rendering policy.

Final app gates: canonical `test -race ./...` PASS (UI468.039s,
agent154.856s, native79.634s, gitworkbench49.309s); no timeout changes or retries.
Focused race3x14.921s and final broader Review102.110s passed. Wrapper5new plus
21existing tests, all-app vet and Darwin arm64 build passed. Source diffcheck
passes excluding the transported original unified patch, whose legitimate
context-space/tab lines must retain their exact reviewed bytes/hash. Historical
toolkit prototype allocation assertion still fails equally on pristine/candidate;
not relabeled as passing or suppressed. App full-suite success does not imply
that independently recorded toolkit baseline assertion became green.

## Native verification remains pending

No hosted run, new installer, pointer input or publication was performed for
this isolated candidate. Future exact-SHA capture must preserve fixture/font
hashes,1470x922 dimensions/raw PNGs and compare with pinned Rust. Verify fixture
note and return wrapping at identical code coordinates, then split/no-wrap,
resize/pan/sticky and unrelated text negative controls. Existing line-height/
baseline differences may remain. Darwin compile means unchanged-policy build,
not physical Mac acceptance; mounted tests do not prove native IME/selection.
