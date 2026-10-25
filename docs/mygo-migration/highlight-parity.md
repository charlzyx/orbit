# Paint-only lexical compatibility evidence

Status as of 2026-10-25 04:51 UTC: independently implemented and scoped tests
passed. **Not full GUI, transcript, editor or pixel parity.**

## Implemented contract

`internal/highlight` recognizes the pinned Orbit 16 language families, aliases,
file-name exceptions and friendly labels. It returns UTF-8 byte spans and small
cross-line carry state, with no font, layout, sizing or wrapping dependencies.
Its bounded 256-entry immutable block memo avoids re-lexing unchanged streaming
blocks. `Runs` restores plain gaps losslessly and rejects overlapping, out-of-range
or mid-rune externally supplied spans. Source quirks are intentionally preserved,
including case-sensitive SQL and Go primitive-type literal classes.

`internal/workspace/review.go` computes line tokens while parsing off the render
thread, maintains separate old/new carries, resets both per file/hunk, paints
context from the new side, and retains tokens through hidden-context expansion,
grouping and split conversion. `internal/gitworkbench/embedded_diff.go` paints
only run colors on the existing monospace text element; its palette callback is
evaluated at paint time. Review and transcript mounting are owned by the native
shell integration and must be separately verified, not inferred from this backend.

## Executed evidence

- The pinned Rust `highlight.rs` 15 built-in tests passed with rustc 1.94.0,
  edition 2021, without starting the Rust UI.
- `go test -race ./internal/highlight`: 1,204 exact executable-source oracle
  cases, 16 languages, complete declared vocabularies, quote/comment carry states,
  numeric/shell/diff/Unicode edge cases; additional streaming-prefix, boundary,
  lossless-run and bounded concurrent cache tests passed.
- `go test -race ./internal/workspace`: independent-side/hunk/file carry,
  context expansion/grouping/split preservation and existing workspace suite passed.
- `go test -race ./internal/gitworkbench`: syntax spans reproduce every byte,
  override color alone, change palette without re-tokenization, and existing
  native controller suite passed.

Reproducible source fixtures and generator are in
`internal/highlight/testdata/`. Each fixture is small and language-specific. The
source hash is pinned and checked before regenerating expectations.

## Licensing provenance caveat

The original Orbit `highlight.rs` header and inherited NOTICE describe Waku as
MIT licensed. Waku's current [manifest](https://github.com/egoist/waku/blob/main/Cargo.toml)
says GPL-3.0-only. Its first root LICENSE change found in the public history,
[commit c0ea87b](https://github.com/egoist/waku/commit/c0ea87b5592b2bcbbe5c7d53eb26656ec66f868c),
added GPL-3.0 on 2026-08-11. This does not establish a permissive historical license
for the adapted Orbit code. The new Go scanner is independently written from the
observable pinned behavior and language vocabulary, not mechanically translated
or copied from Waku. The historical claim in inherited Rust files is unresolved;
we neither silently replace the original notice nor claim a licensing clearance.

## Explicit remaining scope

The fixture suite establishes the tested lexical contract only. Rendered baseline
comparison, full transcript Markdown code-fence/custom-render integration, native
editor language behavior, source diff parser variants, all theme-switch interactions
and desktop platform runs require their own evidence. No parity matrix row becomes
fully verified solely because these lexical/domain tests pass.
