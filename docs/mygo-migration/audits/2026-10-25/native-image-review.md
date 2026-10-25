# 32083a3 bounded native image review (2026-10-25)

Read-only visual inspection via view_image; no source/image edits, no execution of UI actions by this reviewer.

## Provenance and limits
Evidence: [run 40201277104 artifacts](https://github.com/charlzyx/orbit/actions/runs/40201277104). Reference ZIP SHA256 `1651866912451973acd6e1c7e9afe1978148ad28b3d723a04b2da9c80d387fd5`; Go ZIP `5e755ad082609a118ca4d6bb86cc885aa085fd9a8c803eb7f0c67b3e318f1b50`.
- R = `12602355698/reference-desktop`, source clean `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
- G = `12601098875/go-desktop`, source clean `32083a3cfe5f6dd76b760bd4a12c831251bfa6c3` (unreleased frozen 17-file batch, NOT shipped .4).
- Both workbench manifests: 1470×922, Xvfb/Openbox Linux software graphics, density100/dpi96; time/caret not frozen. Synthetic local Git/session, no credentials/inference. Native desktop rendering is not a Mac interaction test.
- Manifests' OCR readiness is not acceptance. Go WB08 missing due None.strip harness exception, no terminal conclusion from this capture. Retry is separate.
- Inspected 7 WB pairs + 2 Usage pairs, and three old d9c78f Go images for directional comparison. No pixel-difference computation or whole-feature PASS claim.

## Per-pair bounded result
All paths below relative to R and G above (same filename each side).

| Image | Direct observation | Remaining limitation |
|---|---|---|
| `workbench/wb01-transcript-dark.png` | Both contain fixed user message, assistant list/code and no-inference note. | Go user/assistant block sits lower, bullet spacing tighter, user timestamp/copy on left rather than right; sidebar lacks Rust current/workspace/age presentation; extra Work in: Local footer. Not full transcript parity. |
| `workbench/wb02-explorer-dark.png` | New Go Explorer card width/header/filter/footer and row/status placement now substantially track Rust. Five files count and collapsed directory state agree. | New GO/TXT badge text displaced below own squares. Native image does not exercise resize, keyboard expand/collapse, rename, default-app/context menus or persistence. |
| `workbench/wb03-file-alpha-dark.png` | Both open alpha.go and show staged contents; Explorer selected alpha. | Go absolute path vs Rust relative name; Save/toggle vs Saved plus byte count; Go reports 6 lines vs Rust 5 for trailing newline convention; no tab language badge; typography and geometry differ. No edit/save semantics exercised. |
| `workbench/wb04-file-tabs-dark.png` | Both retain alpha and bravo tabs, show unstaged bravo contents, select bravo. | Same editor gaps, plus confirmed badge displacement persists and is partly obscured by selected row. Only open/switch snapshot, not unsaved/close flow. |
| `workbench/wb05-review-uncommitted-dark.png` | Both show same three changed files and overall +3/-2; tree and dock occupy comparable region. | Go extra diff-sign gutter leaves less content width, so return strings/notes wrap earlier; row section height diverges; per-file counts absent from Go section headers. This wrapped short fixture cannot test known no-wrap horizontal visibility/sticky header/context expansion gaps. |
| `workbench/wb06-git-changes-dark.png` | Both show 1 staged file and 2 other changes with matching per-file +/− counts, commit form visible. | Go names are centered across rows instead of Rust left-aligned disclosure rows; different toolbar/action hierarchy and much looser row spacing. Screenshot does NOT prove stage/unstage/discard/commit behavior. |
| `workbench/wb07-git-history-dark.png` | Two same commits/hash labels present. Go short list now starts near y205, not old d9c78f near y500: top-alignment fix visibly effective. | Metadata centered rather than under subject, branch badge at right vs next to title, different filter/toolbar/tab geometry. Graph/pagination/actions not exercised. |
| `usage-simple-dark.png` | Same 33 requests,55.4K tokens,9.5% cache,5.4s average and synthetic heatmap/series; major card bounds align. | Typography, header/control baseline and spacing still differ. Lower chart/interaction/export/filter correctness not established by image. |
| `usage-details-dark.png` | Same four breakdown rows/totals and shares; new Go Breakdown bottom ~845 and Records top~865 align Rust, vs old Go~820/~840. Segmented tabs and natural content-height changes visibly improve geometry. | Name column allocation and table-header alignment differ; headers are visibly clipped (Rust also clips some). No sorting/search/columns/pagination or records drilling exercised. |

## New badge defect (bounded source diagnosis)
Frozen `internal/ui/files.go:470–479`: `FontSize(t.Px(7.5)).LineHeight(t.Px(8))` inside 17×17 badge. MyGo LineHeight is a font-size multiplier; FixedLineHeight is DIP. API documented in available MyGo source `/workspace/shared/orbit-toolchains/review-visibility-probe-wcbaJv/ui/element.go:1176–1189`. Current value yields roughly 60px line height at default font, not 8px. Original contract: `crates/orbit-pi/src/app/helpers.rs:968–990`, font ui_px(7.5), line_height ui_px(8.), flex centered 17px square; Explorer calls it at `explorer/panel.rs:1169`.

Minimal correction candidate: FixedLineHeight(t.Px(8)), no broader icon/menu style changes. Parent owns implementation. Mounted regression must inspect each badge AND text geometry for alpha.go, bravo.go and notes.txt: text bounds contained in matching square, centerY difference <=1px, correct GO/TXT, no next-row intrusion; default/scaled fonts, selected/unselected and later frame. Current test checks row heights but not descendant text geometry, explaining missed defect. Native recapture still required after fix.

Prior evidence `/workspace/shared/orbit-toolchains/native-capture-d9c78f/go/go-desktop/workbench/wb02-explorer-dark.png` used generic file glyphs, no language badges: this is a newly introduced badge implementation defect, not an old matching-badge regression. The old History and Usage Details images in that same root directly support above before/after direction.

## Tomorrow practical manual checks
Prioritize real functions: choose/open workspace; open two files, edit/save/reopen, cancel dirty close; stage/unstage disposable change and verify disk/index; History open commit and paginate; Review long line with wrap off and pan, expand context; terminal exit/restart preserves cwd; Usage filter/sort/export expected fixture counts. These screenshots only validate the selected states, not these operations. Keep .4 vs unreleased changes explicit. No full visual or functional parity claim.

## OCR-only retry and WB08 (14:26 UTC)

[Run 40201741107](https://github.com/charlzyx/orbit/actions/runs/40201741107), source `e577c31c91780a0367f2cfe676b98e5ac2ba12f8`, changes only the TSV parser and regression test from runtime `32083a3`. Both native jobs completed eight workbench captures; reports are `READY_FOR_MANUAL_REVIEW`, not acceptance.

- Go artifact12601189266 ZIP SHA256 `3bb32bf8a219f0dc721dc0eac0faddd4b5ff041686845bfa3b05c1c122ba2831`.
- Rust artifact12602216710 ZIP SHA256 `7c649e8731325a2f2a133badeca881f6ab14dd63cbbe04b0262e9fb3da519dc9`.
- Direct review of both `workbench/wb08-terminal-dark.png`: pane begins at same y≈662; both show synthetic ready line and shell prompt. Go lacks Rust header restart, with different tab/icon sizing and hide/close presentation. This shows an opened live terminal, not exit/restart, clipboard, IME, scrollback or multi-tab correctness. Separate terminal-host patch is not in these images.
- Aggregate job FAILED at `TestAbortPendingRestorationCancelledByExplicitClear`, abort_pipeline_test.go:138, after six seconds without draining workers. Successful capture jobs do not override that failure. Investigation is separate; whole workflow is not PASS.

Minimal badge correction and MCP/terminal-host patches require final-source gates and fresh native images. None changes downloaded `.4`.

## ecba1eb badge correction: actual frame and conservative click gate

Run40203054192 Go artifact12602855942 ZIP SHA256 `b73765ab71213b49657766c03a4e08cd1358fb2d1cdd55114a1a1537dc3b223f` contains a fresh full-window Explorer frame. Direct visual review confirms GO/TXT now centered inside their own badge; this is bounded evidence of the new fix, not Explorer parity.

Capture stopped before WB03 pointer input: PSM11 recognized exact alpha.go at [1286,190,44,11] but confidence52.055969 was below the unchanged60 gate. No click was sent. A second PSM3 read of that same unmodified whole-window PNG recognized the same exact literal and bbox at80.610741. The follow-up harness permits only this narrowly defined second-pass confirmation, records both raw TSV/hash proofs, and still rejects missing/duplicate literals, mismatched bounds, changed images and confidence below60. No cropping, fuzzy label matching, threshold lowering, or coordinate-only click is allowed. Actual synthetic TSVs and artifact provenance are preserved in scripts/workbench-native/testdata/wb03-go-*; tests cover no-pointer-on-failure.

This harness failure does not establish a runtime regression. The updated harness must still finish a new actual capture on the identical runtime before native acceptance claims.

Safety review additionally tightened exact-literal uniqueness across all in-region OCR candidates, including a second low-confidence duplicate; nonfinite/out-of-range confidence is rejected. This is the same OCR engine with another segmentation mode, not statistically independent recognition. Hash equality proves the saved image is unchanged, not that a dynamic live window cannot change while OCR runs; this harness uses the bounded static synthetic fixture and does not generalize to arbitrary dynamic UI clicking. Local alternate OCR used Tesseract5.5.0; CI must independently pass its own engine and unchanged gate.
