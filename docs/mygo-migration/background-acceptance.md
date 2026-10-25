# Background source acceptance

Pinned reference: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`,
`crates/orbit-pi/src/backdrop.rs` and `dither.rs`; native placement is
`app/view.rs::new_task_backdrop`, `backdrop_image`, and `dot_backdrop`.
This is an implementation/verification checklist, not matched-image evidence.

| Case | Procedure and source-specific expected result | Status |
|---|---|---|
| BG01 | With no selected photo, derive the mark from shipped app-icon.png, threshold/erode it and paint the source 8×8 Bayer field. Small bounds omit the mark rather than deform it; phase animates the source sweep. | PARTIAL: field unit/native CPU tests only |
| BG02 | Enable Reduce Motion during animation, leave it enabled, then disable. Current phase freezes; elapsed paused time is not accumulated into a jump. | PARTIAL: clock unit test only |
| BG03 | Choose a PNG/JPEG/GIF/WebP in an isolated home. Persist `new-task-background.png` and basename in `background.json`; preserve prior tuning. Fit longest edge to1600, do not enlarge smaller sources. An unreadable/corrupt file must not replace the prior image. | UNTESTED |
| BG04 | Paint all five cell presets2/3/4/6/8 with odd canvas dimensions. Sample each cell center, use source4×4 Bayer thresholds and independent color channels; fill last partial rows/columns and preserve alpha. | UNTESTED |
| BG05 | Blur0 must be a true no-op; values snap to0.5 within0–6. Blur runs after ordered dithering. Invalid cell/fade values fall back to4/.14, independently of valid fields. | UNTESTED |
| BG06 | Choose a photo, alter blur/cell, change only fade, then clear and relaunch. Blur/cell invalidate processed pixels; fade alone reuses them. Clear removes image/config and restores defaults. Reload on changed image mtime/size. | UNTESTED |
| BG07 | Use a wide image in a narrow main column. Cover-center image, clip it inside that column below top bar, never bleed onto sidebar. Fade heights0/.08/.14/.22/.32 pair with peaks0/.55/.72/.88/1. | UNTESTED native/visual |
| BG08 | Repeatedly select images while processing; cancel an older request, navigate away, deny a save path and recover. No stale image/label is installed and real errors remain visible. All I/O/processing stays off the render frame. | UNTESTED host |

No full feature PASS; matched Rust/Go screenshots, native dialogs, and host
integration remain untested. Current implementation changes are uncommitted;
do not attribute them to a previously published SHA.
