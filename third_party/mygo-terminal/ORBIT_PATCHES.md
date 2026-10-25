# Orbit terminal compatibility fork

This is an Orbit-maintained compatibility copy, **not an upstream MyGo
release**. It copies only `plugins/terminal` from MyGo commit
`9c69cf829ece196c5b3a3601f1c7bbe70526376f`. The rest of MyGo remains the exact
pinned upstream module. The local module path
`github.com/egoist/mygo/orbit-terminal` avoids an ambiguous import with the
original plugin while retaining Go's permitted access to MyGo's internal
localization, download, and test helpers. It is resolved locally by Orbit's
`go.mod`; it is not a published module under the upstream author's account.

## Changes

- Self-import paths point to this local module.
- `Terminal.SetVisible(bool)` gates redraw invalidation without closing,
  suspending, replacing, or replaying the PTY or VT state. The host must report
  tab, pane, page, and window visibility. Output and metadata still update.
- `Options.ScrollbackLines` binds the **existing** Ghostty ABI option 28,
  `SCROLLBACK_MAX_LINES`, with ABI-manifest validation. A positive value uses
  a line budget instead of the default byte budget.
- `Terminal.ScrollbackRows()` exposes actual retained physical history rows
  so tests and callers do not mistake a configured budget for exact retention.
- `orbit_compat_test.go` covers hidden invalidation, retained output, return to
  visibility, and the actual native line-budget behavior.

All native library URLs and SHA-256 hashes in `mygo-plugin.json` are unchanged.
The native engine remains Ghostty commit
`befcdfd2c3a1cb24d9ec886e93c95b2b5daa7028`; no patched engine binary is supplied.

## Remaining difference: exact history retention

Orbit's Rust reference uses an exact 10,000-line Alacritty history. Ghostty's
native limit evicts whole pages and has a minimum allocation, rather than
maintaining an exact line ring. In a Linux test using the pinned library, after
20,050 one-cell lines and a 24-row viewport, a configured 10,000-line budget
retained 20,027 / 7,238 / 9,623 / 9,939 history rows at widths 1 / 10 / 80 / 240.
These are characterization results, **not source-equivalence assertions**.

Clipping excess visible rows alone cannot restore rows that page eviction has
already removed. Exact last-10,000 semantics would require either a native
strict-line eviction facility or a carefully tested logical history layer
with extra native retention, a persistent tracked discard boundary across
reflow, and consistent clipping of viewport, selection, search, and exports.
This copy does not disguise that remaining difference through lossy replay,
whole-history clearing, or an inaccurately reported count.

Orbit's separate adapter calibrates row pitch against real font metrics:
bottom terminals use font size × 1.4, and custom Pi surfaces use the source
12/18 editor-size scale. Ghostty still quantizes cells to integer device
pixels; fractional GPUI row pitches are not claimed to be pixel-identical.

## Verification

This directory is a nested module, so root `go test ./...` does not test its
own upstream suite. Run both root Orbit tests and:

```sh
go -C third_party/mygo-terminal test -race ./...
CGO_ENABLED=0 go -C third_party/mygo-terminal build ./...
```

Native tests use the original hash-verified libghostty-vt asset. Set a writable
`XDG_CACHE_HOME` for Linux test environments whose real home is read-only.
Cross-compilation is not evidence of running the macOS or Windows backends.

## Notices

- `LICENSE`: original MyGo MIT license, copyright 2026 MyGo contributors.
- `GHOSTTY-LICENSE`: MIT license from the exact Ghostty source commit above.
- `ghostty_themes.go`: original embedded iTerm2-Color-Schemes license and
  attribution, retained unchanged.
- This document identifies Orbit's modifications; original source headers,
  build provenance, and native-library hashes remain intact.
