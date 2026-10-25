# Native visual fixture audit

Audit date: 2026-10-25 UTC. Scope: the first two actual Linux desktop capture pairs, not the later, un-captured working-tree fixes. This audit does not change the UI, capture workflow, frozen Rust source, or comparison implementation.

## Verdict

- Both applications produced actual native-window PNGs at 1470 × 922. A software-rendered native window is still a native execution; this is distinct from the project's offscreen CPU component fixtures.
- Independently recomputed whole-image differences reproduce the existing report: new task has **187,498 changed pixels (13.834019508%)**; General settings has **47,020 (3.469240191%)**, out of 1,355,340 pixels each. There are **zero exact-pixel matches**.
- These results are useful raw image comparisons. They are **not a certified normalized-state comparison**: the supplied manifests do not establish actual font resolution, loaded state, input focus/caret phase, repeatability, or full render-environment identity. Agreement between two incomplete manifests is not proof of equivalence.
- There are source-confirmed geometry, fill, control, and content differences, independently of font rasterization. The separate Linux text stacks also have source-confirmed metric/placement differences. Neither finding excuses leaving layout or behavior mismatched. No evidence here supports “pixel perfect,” nor does it prove that different font files were selected.

## Evidence identity and limits

| Item | Identity |
| --- | --- |
| Frozen original Orbit | `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`, v0.2.12 |
| Captured Go implementation | `1821f9829e20a620bfb7863cba7bba48ebe9f239` |
| MyGo source inspected | `9c69cf829ece196c5b3a3601f1c7bbe70526376f` |
| Reference artifact directory | `reference-capture-1821f982/reference-desktop/` |
| Go artifact directory | `native-capture-1821f982/go-desktop/` |
| Existing raw comparison | `comparison-1821f982/comparison.json` |

The artifact directories above are under the cloud investigation's `orbit-toolchains` directory. The native PNGs were individually inspected, not inferred from source or descriptions. The separately captured desktop workflow's reported success means capture/build execution succeeded, not that pixel parity passed. Current working-tree geometry fixes are deliberately not credited against these older PNGs.

PNG SHA-256 values:

| Scene | Rust | Go |
| --- | --- | --- |
| New task | `d4accf98a16cf6f95044be8f0a4cdfcc710a58e1cdce9fac1c465bc85e31256d` | `5a2563ba498f6f54221e75c54fb4cd03a5203f7a89d3d37c9a801d93aac6d975` |
| General | `97b2186bf23698998f9eb66c6fd67762e7720acfb347a2c87485265f9a0fee70` | `22d2dbd09a6af6b43f83f05e31db79333146d2acadb8364ed8289bfb518f9c5c` |

Both graphics logs list X.Org 21.1.11, Mesa 25.2.8, and llvmpipe LLVM 20.1.2. This reports available graphics software, not necessarily the selected compositor/raster path in each process. Go's process log explicitly says `mygo: native UI draws without the GPU: there is none`. The old manifests incorrectly call scale 1 `dpi: 1`; the current capture script has changed that to `dpi: 96, uiScale: 1`. That correction does not retrospectively measure either old process's actual DPI or logical/device scale. The saved first 30 `xdpyinfo` lines do not include its dimensions/resolution report.

## Measured geometry and visible differences

Coordinates are zero-based native pixels. Rectangle right/bottom coordinates below are exclusive. Pixel probes are diagnostic measurements, never masks or exclusions from whole-image comparison.

### New task

| Measurement | Rust | Captured Go |
| --- | --- | --- |
| Sidebar separator at y = 700 | x = 247 | x = 248 |
| Logo bright-ink bounds, threshold RGB ≥ 180 inside [30,40,220,80) | [88,47,160,61) | [73,53,174,73) |
| Composer flat fill at (700,820) | `#242424` | `#1f1f1f` |
| Composer flat fill interval at x = 700 | y = 783…876 | y = 773…880 |
| Top-right app chrome | No Linux minimize/maximize/close glyphs in this reference | Three additional window-control glyphs |
| Empty composer controls | Compact lock/workflow/model controls | Visible Full access, Off, and Send labels; different arrangement |
| Empty composer input | Visible caret in this exposure | No visible caret in this exposure; focus/phase is unproven |

The logo difference is not evidence of a font mismatch: it is an image. Frozen Rust `app/view.rs` requests `embedded_image_w(..., px(76.))`; captured Go `internal/ui/app.go` requests width 104 and height 25 inside a 38 px column. Frozen Rust's composer explicitly uses `theme.surface_secondary`; the captured Go `composerView` uses `t.Surface`. Those two source differences agree with the measured pixels.

A fixed diagnostic region containing the composer and footer, [370,750,1350,922), has 119,740 differing pixels out of 168,560 (71.037%). A backdrop-only probe [1080,100,1400,700) has 5,240/192,000 (2.729%) differing pixels. These percentages are not pass thresholds and cannot be summed as a complete attribution. The backdrop discrepancy must still be separated into content-origin/geometry, phase, and native paint behavior before assigning a cause.

### General settings

Settings navigation is structurally much closer. Both sidebar separators are x = 239 at y = 700, and both machine-card side borders are x = 399 and x = 1310 at y = 200.

At x = 700, exact `#343434` horizontal rule positions are:

| Rule | Rust y | Captured Go y | Delta |
| --- | ---: | ---: | ---: |
| Page header rule | 76 | 77 | +1 |
| This machine card top | 121 | 123 | +2 |
| First machine row separator | 191 | 193 | +2 |
| Second machine row separator | 284 | 285 | +1 |
| Machine card bottom | 376 | 378 | +2 |
| Notifications card top | 443 | 444 | +1 |
| First notification row separator | 513 | 514 | +1 |
| Second notification row separator | 583 | 584 | +1 |
| Notifications card bottom | 653 | 654 | +1 |

The notification separators are present in both images; their low contrast must not be mistaken for absence. Text ink differs in addition to these physical row displacements. Both screenshots use the same sampled main background `#181818`, sidebar `#1f1f1f`, grouped-card fill `#242424`, and border `#343434`.

Visible non-state-equivalent content remains: Rust's footer says v0.2.12; Go's says v0.2.12-go-dev. This is a real text difference, not noise to crop or blur away. Any diagnostic fixture that supplies an equal display-version string must disclose that explicit input separately from the real build's provenance. Until then retain the difference, even if considered an intentional development-build distinction.

## Font evidence: inputs are not resolved faces

### What is proven

- Rust `assets.rs::register_zed_fonts` loads eight bundled Zed font files. GPUI maps `.ZedSans` to IBM Plex Sans and `.ZedMono` to Lilex.
- Go embeds the original asset tree through `assets.go`. `internal/ui/theme.go::LoadAssets` registers Plex Sans bytes with `.ZedSans` and Lilex bytes with `.ZedMono`.
- The font-asset diff between the frozen reference and captured Go source is empty.
- The actual TTF metadata contains IBM Plex Sans regular/italic at weight 400 and semibold/semibold-italic at weight 600; there is no bundled Plex Sans medium 500 face. Lilex regular/italic are 400 and bold/bold-italic are 700.
- The semibold file has legacy family `IBM Plex Sans SmBld` and typographic family `IBM Plex Sans`. A local file metadata inspection exposes that distinction; it is not a native runtime face-resolution trace.

Representative bundled-byte SHA-256 values:

| Font input | SHA-256 |
| --- | --- |
| IBMPlexSans-Regular.ttf | `975dcda37d80f038dcd143c22e33ca2d97a0cc5a929aace1c749153b0fe1afa5` |
| IBMPlexSans-SemiBold.ttf | `a20caf8286023a6a7a85e40b1d2a4ae9fc3e3b1f9eda8f4c542dd4986af67bb1` |
| Lilex-Regular.ttf | `8a467ad5a1941dc7c9cebaee34e3ed9364a73b48393034e363e4b528bd711d04` |
| Lilex-Bold.ttf | `3369214e2ecddb45ae5024b11d45dd637e99e9d2127ab729817d41d72b8a9c68` |

An externally executed `fc-match` is not proof of MyGo's face: MyGo adds application fonts and alias rules in its own process, then refreshes its Pango font map. The external process does not necessarily see those registrations. It is not proof of GPUI's face either: GPUI Linux uses its own CosmicText font database.

### Source-proven renderer differences

The sources inspected are the actual pinned MyGo checkout and cached pinned Cargo dependencies, not current online documentation.

| Area | GPUI 0.2.2 Linux reference | MyGo pinned Linux implementation |
| --- | --- | --- |
| Shaping and font system | CosmicText 0.14.2, Rustybuzz 0.14.1, fontdb 0.16.2 | Pango/HarfBuzz/Fontconfig system shared libraries |
| Rasterization | Swash 0.2.10 through CosmicText SwashCache | Cairo/FreeType |
| Hinting | CosmicText `swash_image` builds scaler with `.hint(true)` | Cairo hinted metrics enabled; actual hint style and antialias mode come from GTK XSETTINGS/defaults |
| Horizontal placement | GPUI has four subpixel variants; shaped `glyph.x` and `glyph.y` are retained as floats | `pango_context_set_round_glyph_positions(true)`, `positions()` returns one horizontal position |
| Vertical placement | One GPUI subpixel variant on Linux; Swash path truncates vertical fractional shift | `baseline(y)` floors y with 0.001 epsilon |
| Line metrics | `layout_line` returns CosmicText layout width/max ascent/max descent | Font metrics are hinted; line height is max(requested multiplier × font size, ascent + descent); baseline centers that ink extent |
| Weight selection | GPUI first uses `zed-font-kit` best-match properties, then records actual CosmicText fallback runs | Requested numeric Pango weight passed into the font description; actual per-run Pango font selected downstream |

Do not describe GPUI as “unhinted”: the pinned Swash path explicitly enables hinting. Do not assume MyGo's initial grayscale/medium default was the actual captured setting: `windowHost.uiFont` supplies GTK's runtime settings, potentially overriding it, and these values were not logged.

For a requested weight 500, `zed-font-kit::matching::find_best_match` explicitly prefers 400 if that candidate exists and an exact 500 does not. That is source-level selection behavior, not proof of the captured final face: system-font duplicates and per-glyph fallback still require observing the resolved run. Pango's selected face and possible synthesis are unrecorded. The images alone cannot distinguish altered rasterization from a different font selection.

These architectural differences can explain residual text width, baseline, wrap, and coverage differences even with identical font input bytes. They do not establish which cause dominates the current images, do not guarantee that byte equality can be reached by changing preferences alone, and do not justify leaving actual geometry, fill, control, or interaction differences unresolved.

### Smallest trustworthy read-only runtime instrumentation

“Read-only” here means logging the selections and geometry the renderer already used, not substituting faces, changing hinting, overriding clocks, or repainting text with another engine.

1. **MyGo:** At the existing `pangoEngine.shape` loop, observe `item.font` for every final shaped run, not just `font(style)`. At `fontOf`, attach a stable trace identity to each distinct native font. Query that same PangoFcFont's pattern with `pango_fc_font_get_pattern` and record file, face index, family, PostScript name, weight/slant, variations, embolden/matrix flags when available. Keep Fontconfig and CSS/Pango weight units explicitly distinct. Hash the selected file's bytes. For stronger loaded-byte evidence, obtain the HarfBuzz face from the existing `fontGetHBFont` result and hash a valid complete referenced face blob; also record its index, or explicitly mark loaded bytes unavailable. Do not replace a missing result with the requested family name.
2. **GPUI:** At `CosmicTextSystemState::layout_line`, log the final font only **after** the `glyph.font_id` fallback correction. Resolve its `CosmicTextFont::id()` to fontdb `FaceInfo`; record actual PostScript/family/weight/style/index and hash `CosmicTextFont::data()` or the database's `with_face_data` result. Logging only `font_id(&Font)` misses fallback chosen during shaping.
3. **Both:** Log requested family, size, weight, language, features, run text range, final face identity, glyph IDs/clusters, advances/positions, line width/height/ascent/descent, baseline, element content box, scale, and pixel-snapping policy. Limit text to explicitly synthetic fixture strings or their hashes. Tie records to case ID and completed frame ID. Record actual Cairo scaled-font options/matrix and GTK source values for Go, and Swash/subpixel policy for Rust.
4. Emit deduplicated trace data after the relevant frame so log I/O does not perturb startup/animation timing. Preserve stock production captures. If instrumentation requires a temporary dependency overlay, label that build diagnostic, hash the overlay, retain exact upstream pins, and check same-renderer instrumented versus stock captures in a stable scene before treating its trace as representative. Do not silently redefine a patched build as the frozen original.

Useful shared fixture strings include real labels such as General, This machine, Workspace, Connected, the exact synthetic path, and a distinct test string exercising 400/500/600/700, italic, punctuation, ligatures, digits, code font, and intended Unicode fallback. A font specimen is supplementary; the actual application runs must also be traced.

## Remaining fixture gaps

| Gap | Why current evidence is insufficient | Minimum evidence to close it |
| --- | --- | --- |
| Build identity | Script labels a binary from current HEAD, or hardcodes the Rust SHA, without verifying the binary | Binary hash; exact source tree and clean/dirty state; build command/toolchain/flags; lockfile and compatibility-overlay hashes; runtime build identity when available |
| Reference font input root | Current script hashes `CAPTURE_REPO` from the migration checkout even when Rust was built under `/tmp/orbit-reference` | Hash assets from each binary's actual build source; verify embedded bytes or build artifact inputs. Current equality of source assets does not make that inference safe for future runs |
| Font selection/rendering | Shared asset list does not show selected face, fallback, synthesis, hinting or shaping | Final shaped-run trace described above; system library versions and effective text settings |
| Actual DPI/viewport | 1470 × 922 output is measured; DPI/scale are mostly declared | App logical client bounds and device scale; X window geometry/ID/class; `xdpyinfo` dimensions/resolution; Xft DPI and relevant GTK settings; capture client, not root/frame decoration |
| Theme/state | JSON preferences and scene filename do not establish loaded values | Read-back active palette/mode and semantic color tokens, UI/code fonts/sizes, density, locale/direction, sidebar width, selected page, scroll offsets, focus, dialog/popover states |
| Startup readiness | Fixed 3 s / 1 s sleeps can capture different asynchronous stages | App state/frame-ready acknowledgement for the named scene, required catalog/connection response completion, no pending fixture mutations, and settled geometry |
| Pointer/hover | Pointer position is not normalized; settings traversal leaves it over the row clicked | Move outside client or to a documented inert root coordinate after each input; wait for hover/tooltip dismissal; record actual pointer location |
| Window/input focus | Activating the window does not prove the same widget owns focus | Record X active window plus application focus owner; match focused/unfocused state deliberately; verify caret visible/hidden in a separate focused-input case |
| Motion/clock | Reduce motion does not globally freeze caret, spinners, toast timers, elapsed counters, or clocks | Classify each scene's volatile fields and either prove they cannot affect the scene or provide explicit equal fixture inputs/phase with provenance |
| Pi/environment | Real Pi processes, extensions, discovery, environment and update settings can differ despite identical visible Connected | Pi binary/version/hash, args, allowlisted environment, synthetic files hashes, expected RPC replies/state, offline status and initialization completion. No credentials or paid inference |
| Version/PID/uptime | Actual version labels differ; future Runtime/About captures contain volatile fields | Explicitly equal synthetic display inputs in a separately identified diagnostic fixture, or keep the scene non-normalized and report differences. No post-capture masking |
| Repeatability | One exposure per renderer cannot isolate nondeterminism | Repeated captures across a documented interval covering applicable timer cycles and at least one clean restart, after establishing semantic state equivalence; retain all hashes and whole-image diffs |
| Comparator completeness | Current comparison tests equality of optional fields; both missing values compare equal; uiScale/timezone/font/clock evidence is not checked | Separate schema/coverage gate requiring each field, known evidence quality, expected case roster, and nonempty captures; do not relabel raw equality as accepted parity |

### Clock nuance

Frozen Rust initializes `backdrop_phase` to 0 and does not advance it while reduce motion is true. Go's `background.Clock` starts with zero elapsed time and pauses when reduce motion is true. Given successful startup loading of that preference, phase 0 is therefore a source-supported hypothesis for these initial scenes, stronger than an arbitrary sleep but weaker than a runtime observation. Do not mark all animation globally frozen from this fact. The supplied manifests correctly need to retain `clockFrozen: false` and `animationPhaseFrozen: false` until their broader claim is supported.

## Normalized-fixture acceptance contract

Use separate outcomes for **fixture validity**, **pixel equality**, **geometry**, and **behavior**. A raw diff remains available even when comparability fails, but it must not count as a normalized acceptance pass.

1. **Provenance gate:** Exact Rust, Go, MyGo, Pi and compatibility input identities; stock versus diagnostic build explicitly identified; binary and source-input hashes attached. Expected scene roster is nonempty and complete. Images are original, readable, and match their manifest dimensions.
2. **Environment gate:** Same target OS/architecture/display configuration and actual logical/device viewport; effective font/settings/renderer versions known. Different text engines are recorded, not falsely declared identical. Asset and resolved-face identity must be sufficient to attribute any residual difference; unverified selection blocks a font-normalized claim.
3. **State gate:** Matching visible data, locale, theme tokens, density, font request, workspace/path, layout, scroll, selection, focus, and transient state. Equal input files alone are insufficient. Any intentional content deviation is enumerated and remains in the unmodified diff.
4. **Stability gate:** For static cases, show that the named state is settled and repeats without changes across relevant timer periods and a clean launch. For dynamic cases, require an explicit equal observable phase/state; if unavailable, label the case non-normalized rather than hiding the volatile region. Existing frozen reference remains the authoritative unmodified baseline.
5. **Pixels:** Compare all pixels at identical native coordinates and size, with exact equality as `changed_pixels == 0`. No resize, crop alignment, translation, masking, blur, thresholded image replacement, or new percentage tolerance. Preserve max delta, RGB/RGBA metric definition, difference image, and image hashes. The current MAE/RMSE are over four RGBA channels; opaque alpha therefore lowers those averages relative to RGB-only formulas, although changed-pixel counts are unchanged.
6. **Geometry:** Independently compare named element bounds, hairlines, spacing, radii, glyph-run positions, wrapping, clipping, and hit targets against frozen source/native evidence. Diagnostic regions can explain failures but never exclude them from the full-image gate. Record fractional logical coordinates and final physical pixels separately so rounding is observable rather than guessed.
7. **Behavior:** Independently verify corresponding controls, focus/keyboard navigation, hover/pressed/disabled/loading/error/cancel states, and repeated flows. Matching a screenshot does not prove that the pictured controls work or cover offscreen content.

A scene can be a geometry pass with an unresolved raster difference, or a stable raw exact-pixel match with a missing behavioral test. Neither is complete migration acceptance. No claim in this audit certifies either of the two current scenes as passing.

## Smallest next trustworthy capture

1. Recapture the corrected Go source together with the same frozen Rust version in the hosted cloud environment only. Bind asset hashes to the correct source roots and record binaries/toolchains/renderer settings. Do not reuse these older PNGs as evidence for pending code changes.
2. Begin with the static General scene: matching read-back state, idle pointer outside client, explicit focus policy, no transient notifications. Preserve truthful build-version difference unless an explicitly separate fixture supplies equal display text. Record repeated frames, app bounds, active theme and final font runs.
3. Repeat new task with sidebar/content-origin geometry corrected and runtime-proven phase/focus. Inspect composer/control identity separately from text rasterization. The unfocused static case and initial focused-caret behavior need distinct evidence, not an erased caret.
4. Enforce schema/coverage failure separately from raw diff execution. Add regression tests for missing required fields on both sides, differing scale/fonts/state, empty case rosters, manifest/image dimension mismatch, and missing expected screenshots. A green CI capture job must never stand in for comparison acceptance.
5. Expand native evidence to every required surface/state, including below-the-fold content, other themes/densities/scales/locales, popovers, dialogs and focus states. The two first screenshots are a starting diagnostic, not evidence for full feature parity or other operating systems.

No native window was started during this audit, and no restricted display, user computer, local browser, or local Codex path was retried. The independent Pillow diagnostic read the original PNGs without editing them; its whole-image counts exactly reproduce the existing comparison. Its scratch measurement script and JSON are retained with the cloud investigation, not substituted for native execution.

### Audit validation

- Verified that `Cargo.lock`, the referenced Rust asset/view/settings/backdrop files, and font assets have no diff between frozen Rust `299a32b` and captured Go tree `1821f982`.
- Verified the inspected MyGo checkout is exactly `9c69cf829ece196c5b3a3601f1c7bbe70526376f`.
- Independent Pillow full-image counts, percentages and four-channel MAE equal the prior comparison JSON exactly for both scenes.
- Existing `scripts/test-compare-native-captures.py`: 2 tests passed. This checks current comparator behavior, not the proposed stronger schema gates, native interaction acceptance, or application regressions.
