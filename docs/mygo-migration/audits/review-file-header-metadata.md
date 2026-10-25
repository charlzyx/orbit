# Review file-header metadata (isolated wave 7 candidate)

Scope: dock Review normal and sticky file headers only. The 987 native WB05 pair showed the Go header lacked a file glyph and per-file additions/deletions. That capture predates this change; this is not a new native/pixel parity result.

## Source contract

Pinned Rust `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, `crates/orbit-pi/src/diff_view.rs:35–105` renders a glyph, file path, `+additions`, and `-deletions` unconditionally. Nonzero additions/deletions use success/error; zeros use muted. There is no binary/empty/renamed-file header exemption or special status badge. `review.rs:810–830` parses numstat `-` into zero while retaining binary status; missing numstat records get zero-count files at `review.rs:458–466`. Path/counts come from the parsed file, not the currently expanded diff rows. Rename parsing itself is outside this patch: the renderer displays the parser's path verbatim.

`app/helpers.rs:939–965` uses a devicon if a Nerd Font and matching glyph are available; otherwise the caller's plain `icons/file.svg` fallback. This patch provides that fallback, **not Nerd-Font glyph parity or a claim about the installed font availability in the old native capture**. Explorer extension badges are not the Review fallback. TextSize::Small is UI-scaled 12px and IconSize::Small is 14px (`theme/tokens.rs:237–262,344–372`). Header source spacing is density-scaled 12px outer and 8px between items.

## Change and limits

Only `review_sticky.go` production code changes. Add a noninteractive file SVG, source-sized monospace path, and per-file +/- counts. Counts and glyph do not shrink; the path shrinks/one-line clips before covering them. Preserve the existing fixed 34px header, background, toggle/Open targets (28px), action callbacks/keys, owner behavior, sticky attachment, pan/list anchors, parser, and text-selection behavior. Existing Go Open action and target sizing differ from Rust's whole-header interaction; this patch does not redesign those interactions. No font engine changes; known glyph metric/wrapping residue remains.

## Evidence

- `review-header-red-1945.log`: actual mounted normal/sticky metadata test fails before production change because source file glyph is missing.
- `review-header-focused-1945.log`: initial ten metadata/geometry branches pass (1.504s).
- `review-header-complete-1946.log`: Open-route test initially compared a relative expected path with OpenFile's normalized absolute path; corrected test to use a controlled temporary workspace and real fixture file, not production.
- `review-header-complete-1947.log`: all three new mounted tests PASS (2.057s).
- `review-header-race-1947.log`: `^(TestMountedReview|TestReview)` race subset PASS (123.434s), including existing sticky, held-press, scrollbar, pan/split/context regressions.
- `review-header-vet-1947.log`: UI vet PASS; code whitespace check PASS.
- `review-header-cpu-1948.log`: existing native CPU render fixture PASS (2.171s). Actually inspected unchanged PNG `/workspace/shared/orbit-review-header-cpu-1948/review-default-fixture.png`: glyph, path, both zero counts and Open visible without overlap in 254px body. Its synthetic Files metadata is zero despite some add/delete rows, as intended. This is not the old WB05 fixture or a hosted native recapture.

Logs live under `/workspace/shared/orbit-toolchains-continued/`. Production was unchanged during these final gates.

Mounted tests cover nonzero and zero values, empty and binary metadata, rename-style and long paths, icon/path/count/action nonoverlap, real toggle, parsed binary/empty/pure-rename patches, full Review vertical sticky and horizontal pan, new-slice snapshot metadata replacement, and retained Open route. Existing Review sticky wheel/held-press/push, scrollbar, pan/split/context tests remain part of the regression gate. Synthetic controlled fixtures only; no user files or network. Test passes are not proof of exact pixels or hosted Mac behavior.
