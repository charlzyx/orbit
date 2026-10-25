# MyGo native UI feasibility and visual baseline

Inspection date: 2026-10-25 (UTC). Cloud-only source inspection and headless rendering. No user Mac, desktop browser, or user server was used.

## Verdict and boundaries

**A genuine native Go port is technically plausible with the pinned MyGo toolkit. It is not a skin change, an automatic translation, or completed parity.** MyGo has the necessary native layout, input, virtual-list, editor, Markdown, painting and platform abstractions. Orbit's GPUI render functions, entity lifetimes, controllers, domain models and platform integrations still need deliberate migration and behavior tests.

The user subsequently confirmed a full Go + MyGo native reconstruction, not a Rust backend, skin change or webview. This report records the inspection before implementation; the existing Rust app is the parity reference. The migration must update the GPUI-only decisions in `INTENT.md` and `AGENT.md` explicitly. Inspection captures are not evidence of a completed Orbit port.

### Pinned sources

- [Orbit source, release v0.2.12](https://github.com/imrj05/orbit/tree/299a32bc87fe3d4851bdb902d9c03dcf9bf99cac): `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.
- [MyGo source](https://github.com/egoist/mygo/tree/9c69cf829ece196c5b3a3601f1c7bbe70526376f): `9c69cf829ece196c5b3a3601f1c7bbe70526376f`; `mygo.Version` is `0.6.3` at this commit.
- Sources below are paths within those pinned trees, not claims about an unspecified `latest` version. Project instructions and MyGo `docs/architecture.md` were read before the inspection harness was created.

### Material risks identified now

1. **The reference images are not a fresh capture of the pinned app.** The General screenshot visibly says `Orbit v0.2.5`; the checkout is v0.2.12. The captured sidebar is resized, and current default themes are Azure/Azure Light. Do not combine screenshot geometry, current defaults and user-customized settings without recording which one is authoritative.
2. **One upstream native UI performance test fails in this cloud environment.** `TestDiffFrameAllocs` measured 9 allocations during warm scrolling where it requires at most 2. Editor/Markdown/demos passed, but the native UI suite is not fully green. See verification below.
3. **The terminal changes emulator and deployment contract.** Orbit uses `alacritty_terminal`; MyGo's plugin uses a pinned `libghostty-vt` shared library. The library's supported runtime floor and packaging must be honored, and terminal escape/input behavior must be compared, not assumed identical.
4. **Matching widgets is insufficient.** MyGo defaults differ in fonts, metrics, colors, Markdown spacing, selection behavior and motion. GFM alert callouts need an Orbit-specific block renderer. An exact platform screenshot comparison cannot be inferred from Linux CPU rendering.

## Native architecture, not a webview

MyGo offers two independent window paths. For this port, use `WindowOptions.Content: ui.View(...)`. Do not use `URL`, a web frontend, `ui.WebView`, or Go↔TypeScript bindings to substitute for native UI. The typed web IPC advertised by MyGo is unrelated to pi's RPC transport and is unnecessary here.

The native path is:

`Go application state → ui.Context builders → layout / text shaping → scene → Metal, D3D11, OpenGL or CPU rasterizer`.

The platform contributes a native surface, window events, IME and accessibility. `ui` does not require `mygo` or a browser backend. Native-only Linux windows do not load WebKitGTK, JavaScriptCore or libsoup. Source: MyGo `docs/architecture.md`, Native UI and Linux sections; `ui/window.go`, `content.go`, `internal/surface/`.

### Confirmed native Go port seam

1. Keep pi as the only agent runtime: one child process per active runtime/session policy, `pi --mode rpc`, stdin/stdout newline-delimited JSON. Port the existing Rust process lifecycle and RPC model to a Go package; retain raw/unknown event data and request IDs. Do not replace pi with an SDK daemon or confuse MyGo web IPC with pi RPC.
2. Separate Go domain state/reducers from view builders. Replay captured/synthetic RPC fixtures through both the current Rust reference and Go reducers. Preserve session switching, stream/tool ordering, extension dialogs, abort/restart, queued messages, provider authentication and quota extension behavior before styling them.
3. Run process I/O, file watching, Git, parsing and other blocking work off the UI thread. Transfer bounded batches into state using `Window.Update`; preserve the existing approximately 90 ms coalescing policy first, then measure. MyGo builds on demand, so an always-running UI heartbeat should not be introduced merely to repaint idle windows.
4. Treat `ui.Element` as a build-pass value. Keep `ui.Handle`, list/editor state, stable IDs and application data instead of saving elements between frames. GPUI `Entity`, `WeakEntity`, event subscriptions and deferred layers do not transfer directly into this ownership model.
5. Port Orbit design tokens into a Go theme layer and compose unstyled/base controls where needed. Keep behavior tests and deterministic state fixtures independent of rendering, with the existing Rust app available until parity gates pass.

The existing `crates/orbit-rpc` crate is useful as an executable specification and fixture oracle, not a directly importable Go dependency. A Rust helper process would add a second protocol and another process lifecycle; it should not become the default architecture by accident. Direct FFI also conflicts with MyGo's no-cgo design unless explicitly engineered and justified.

## Capability mapping

| Area | Orbit reference | Pinned MyGo native capability | Required migration / validation |
|---|---|---|---|
| Shell, panes, popovers | `app/view.rs`, `app/sidebar.rs`, `sidepane.rs`, `command_palette.rs` | `Row`, `Column`, flex/grid layout, split containers, `Popover`, dialogs, menus, `ui.Router`, custom `Painter` | Rebuild pane sizing, collapse/expand, focus restore, resize handles, layering and titlebar hit areas. Default widgets are not Orbit geometry. |
| Virtual transcript | `transcript.rs`, `transcript_view.rs`, `message_scroller.rs` | `ui.List`, persistent `ListState`, variable row heights, stable `Key`, `FollowEnd`, `AtEnd`, `ScrollToEnd`, sticky `Header` | Preserve reader anchor while rows stream, expand, prepend or disappear. Use a bounded-height list, stable message IDs and no nested all-message render. Test jump-to-latest and selection across streaming rows. |
| Sidebar/file/review collections | `app/sidebar.rs`, `explorer/panel.rs`, `sidepane.rs` | Virtual `List`, `Tree`/`Outline`, tables, selection, labels and reorder | Port sorting, grouping, per-project collapse state, file identity, badges, keyboard navigation and sticky headers; do not equate a stock tree demo with workspace behavior. |
| Composer and IME | `composer.rs`, `app/composer_ops.rs`, `app/session.rs` | `TextAreaBase`, `TextAreaBufferBase`, `.Lines(min,max)`, `OnTextKey`, `OnPaste`, `Composing`, `TextInputClient` | Preserve Enter vs Shift+Enter, steering, history, autocomplete, image paste, attachments, undo and focus. Composition Enter must never submit. Test Chinese/Japanese candidate selection, cancellation, surrounding replacement, emoji/UTF-16/grapheme cases, and long drafts. |
| Markdown | `transcript_view.rs` block parser/renderer | `ui/markdown.View`, `Stream`, incremental parser, GFM tables/task lists/autolinks/strikethrough, footnotes, selectable text, code copying, custom block hooks | Match Orbit paragraphs, heading/list rhythm, code/header sizes and link/image policy. MyGo's defaults include additional syntax and reveal animation, so explicitly configure rather than silently changing behavior. Orbit GFM alerts need `Options.Block` or a custom renderer; they are not built into the inspected MyGo parser. |
| Streaming presentation | Approximately 90 ms drain; paint-only highlighting | Incremental Markdown append parsing, 32-line code chunks, configurable reveal (`AtOnce`, fade), list end-following | Preserve raw RPC text as truth. Disable or deliberately match the default word reveal; do not introduce delayed text as an accidental behavior change. Test partially written fences/tables and finalization. |
| Syntax / code editor | `highlight.rs`, `explorer/viewer.rs` | `ui/editor`, virtual visible-line layout, tokenizers/highlighters, edits, multi-caret, folds, search, completion hooks, diagnostics | Map token classes and exact colors; keep code highlighting paint-only. `editor.TokenStyle` can apply bold/italic, so do not let those change geometry in surfaces where Orbit currently colors only. Autosave, binary/size guards and file conflict policy remain application work. |
| Review / diff | `review.rs`, `sidepane.rs`, `checkpoint.rs`, `git.rs` | `editor.DiffView`, unified/side-by-side, word differences, collapsed context and patch helpers | Reuse only where semantics match. Preserve Git-derived sources, Last Turn checkpoint refs, file headers, context expansion, staging actions and source selection. A buffer diff is not a replacement for Orbit's Git model. |
| Terminal | `terminal.rs`: tabbed PTYs and Alacritty grid | `plugins/terminal`: Ghostty VT, PTY/ConPTY, selection, clipboard controls, mouse/Kitty protocols, reflow, themes, callbacks and `Conn` option | Rebuild tab lifecycle, active/background behavior, shell profiles, working directories, termination, shortcuts and settings. Replay VT fixtures and interactive workloads. Package the pinned shared library; app must not require a first-run network download. |
| Accessibility | Orbit `INTENT.md` P5 still records focus rings/screen-reader labeling as unfinished | Roles, names, values, states, descriptions, actions, announcements, virtual row positions; macOS AX, Linux ATK/AT-SPI, Windows UIA | Preserve keyboard behavior and improve labels without claiming Orbit already had full accessibility. Custom-painted controls still need semantic nodes. Real VoiceOver/Orca/Narrator and focus tests remain necessary. |
| Icons / fonts | `crates/orbit-pi/assets`, `theme.rs` | `ui.ParseSVG`, `Icon`, `Image`, `RegisterFont(data,family)`, `FontFamilies` | Embed original assets; register the original font bytes under stable names including `.ZedSans`/`.ZedMono` if keeping settings compatibility. Retain licenses. Validate fallback, weight and shaping on each OS. |
| Scaling and themes | `theme/tokens.rs`, `theme/appearance.rs` | DIP-based layout, surface scale, explicit theme/font/spacing, `Tester.SetScale`, system contrast/text-scale/reduced-motion preferences | Port the 14 px-authored token math, density anchors 80/100/120%, 10 palettes and separate primary-action/accent roles. Do not substitute MyGo's single `Spacing` knob for Orbit's token mapping. Test fractional DPI and live changes. |
| Platform services | `main_window.rs`, `notifications.rs`, settings/worktree/provider/MCP modules | App lifecycle, windows, menus, dialogs, filesystem watches, clipboard/drag, notifications, secrets, updater primitives | Each controller/config format and platform fallback remains application code. A toolkit API does not preserve Orbit's policy, extension bridge, signing identity or update protocol automatically. |

Sources for the MyGo column: `ui/list.go`, `docs/ui/list.md`, `ui/textarea.go`, `docs/ui/text-input.md`, `ui/markdown/{view,code,theme}.go`, `ui/markdown/parser/document.go`, `ui/editor/{doc,syntax,theme,diffview}.go`, `ui/access.go`, `ui/svg.go`, `ui/window.go`, `ui/preferences.go`, `plugins/terminal/terminal.go`, and their adjacent tests.

### Specific compatibility notes

- `ListState.Key` preserves row identity and anchor across insertion/removal; `FollowEnd` stops following when the reader scrolls away. This maps well to Orbit, but its existing anchor regression fixtures should be ported rather than deleted.
- MyGo exposes real IME adapters: `NSTextInputClient`, `GtkIMContext` and Windows IMM32. Its existing composition tests are evidence of implementation, not proof that OS candidate windows behave correctly in the new application.
- MyGo Markdown does not fetch images by default. `LoadImage` must be supplied deliberately; match Orbit's local/remote image handling and lightbox policy. Mermaid and TeX are not automatically diagram/math renderers. Preserve Orbit's honest copyable Mermaid code fallback unless a real renderer is separately implemented.
- MyGo SVG intentionally omits text, embedded images, patterns, markers and filters. An XML inventory of all 165 SVGs in Orbit's application assets found none of those unsupported element tags and no parse errors. This is a compatibility screen, not a pixel render test of every icon.
- Orbit contains 95 `.ttf`/`.otf` application font files plus license/metadata files. Its defaults are bundled IBM Plex Sans and Lilex; a system-font approximation is not pixel parity. MyGo's own default font size is 13 DIP on macOS and 14 elsewhere, which must be overridden if reproducing Orbit's authored 14 px UI.

## Toolchain and runtime support

MyGo's `go.mod` requires **Go 1.27.3**; its README says 1.27.3+. The looser getting-started text saying 1.27 does not override the module. [Official Go downloads](https://go.dev/dl/?mode=json) listed Go 1.27.3 during inspection. The cloud used `go version go1.27.3 linux/amd64`, with `CGO_ENABLED=0`. Native UI requires no Bun or JavaScript runtime.

| Target | Native renderer / text / accessibility | Runtime caveats |
|---|---|---|
| macOS amd64/arm64 | Metal or CPU; Core Text; AX | Full terminal port requires macOS 13+ according to the pinned terminal library documentation. Signing, notifications and native input need actual macOS CI/runtime validation, not cross-compilation alone. |
| Linux amd64/arm64 | OpenGL 3.3/ES3 or CPU+cairo; Pango/HarfBuzz/fontconfig; ATK/AT-SPI | Native windows need GTK3 and supporting system libraries, not WebKit. Terminal library requires glibc 2.28+. X11 and Wayland differ in placement, global shortcuts and decoration behavior. |
| Windows amd64/arm64 | D3D11, WARP fallback; DirectWrite; UIA | Terminal uses ConPTY and requires Windows 10 1809+. Native-only UI does not need WebView2. Per-monitor DPI, IMM32, caption/snap/resize and shell integration need Windows runtime testing. |
| js/wasm | WebGL2/canvas and browser text | Useful for toolkit previews but **not** an equivalent deployment of Orbit's local desktop app. A browser cannot spawn the local pi process or access the user's workspace like the desktop app. It is not a shortcut to full parity. |

Terminal assets are declared in MyGo `plugins/terminal/mygo-plugin.json`, pinning Ghostty `befcdfd2c3a1cb24d9ec886e93c95b2b5daa7028` with per-platform SHA-256 values. `mygo build` packages them; raw `go run`/`go test` can download them into cache. Release builds should resolve and verify assets in the build pipeline, not silently depend on users downloading them at runtime. Reference: MyGo `docs/plugins/terminal.md`, Library and Testing sections.

## Measured Orbit screenshot baseline

All 30 PNGs under `marketing/public/screens/` were decoded and measured. Representative actual pixels were visually inspected for `session.png`, `new-task-light.png`, `review.png` and `general.png`. Measurement inventory and hashes are in [the manifest](baselines/orbit-reference-manifest.json).

- Every image is **3164 × 2068 RGBA**.
- Every image's bounding box of pixels with alpha exactly 255 is **[112,76,3052,1920)**: 2940 × 1844 physical pixels. Alpha is nonzero outside that rectangle because of the window shadow; the ordinary alpha bounding box is the entire image and is not the content rectangle.
- Embedded resolution is approximately 144 DPI. A 2× capture is consistent with the source's 44 DIP header and 960 DIP composer width, but DPI metadata alone is not proof of the display scale. The following raw coordinates are authoritative observations; DIP conversions are explicitly inferred.

Coordinates use original-image physical pixels, top-left origin, and half-open ranges:

| Feature / screenshot | Measured pixel reference | Interpretation |
|---|---|---|
| Session/new-task/review sidebar divider | x = 770–771; main background begins x = 772 | Window-left-to-main is 660 px, about 330 DIP at inferred 2×. This is not the current source's default 248 DIP sidebar; layout is user-resizable/persisted. |
| Settings divider, `general.png` | x = 590–591; main begins x = 592 | 480 px from window edge, consistent with the source's 240 DIP settings nav. |
| Shared topbar, `review.png` | Window top y = 76; review card top border y = 164–165 | 88 px, consistent with `TOP_BAR_H = 44` at inferred 2×. |
| Session composer | Straight side borders x = 952–953 and 2870–2871 at y = 1700; top y = 1624–1625 and bottom y = 1827–1828 at x = 1912 | Outer width 1920 px, consistent with `CONTENT_MAX_W = 960`; captured outer height about 205 px. Rounded corners/shadows are not included in the straight-edge measurements. |
| Sidebar footer line, `session.png` | y = 1832–1833 at x = 200 | 88 px to opaque window bottom, consistent with the source's 44 DIP footer. |
| Review file-tree divider | x = 2570–2571; tree body begins x = 2572 | Approximately 240 DIP tree width at inferred 2×; verify actual persisted pane widths before turning this into a default. |
| General settings first card | Top border y = 336–337; bottom y = 882–883 at x = 1600; side border near x = 913–914 / 2729–2730 | A concrete reference rectangle for later capture alignment, not a global fixed-height layout rule. |

Sampled dark reference colors, away from text and antialiasing: sidebar/composer fill `(24,24,24)`, main fill `(26,26,26)`, common border `(42,42,42)`, composer outline `(52,52,52)`. `new-task-light.png` sidebar is `(243,243,243)` and main background `(246,245,246)`. These are screenshot samples, **not** replacements for current semantic palettes.

Current source geometry is independently available in `crates/orbit-pi/src/app.rs` (`SIDEBAR_DEFAULT_W=248`, `SIDEBAR_MIN_W=200`, `CONTENT_MAX_W=960`), `app/view.rs` (`TOP_BAR_H=44`, footer 44, resize hit strip 6), `app/settings.rs` (settings nav 240) and `theme/tokens.rs`. Preserve the distinction between a captured user's state and source defaults.

### Pixel-parity acceptance must be defined before claiming it

Capture the pinned Rust application with deterministic fixture state, original font bytes, exact palette/density, window content size, scale, OS/text engine and animation clock. Freeze those as the functional/visual reference. Compare the Go port at the same conditions, with explicit masks only for nondeterministic values and OS-owned chrome. Do not stretch screenshots or silently tolerate shifted text baselines to produce a favorable diff.

Use exact geometry assertions and per-region image differences. Any tolerance for text antialiasing or GPU rasterization must be recorded and reviewed; a Linux Pango image must not be sold as an exact macOS Core Text match. Marketing screenshots are useful visual guidance but do not cover terminal, worktrees, all dialogs, interrupted flows, themes, localization, focus and error states.

## Actual MyGo native demo captures

The upstream `examples/gallery` view was rendered in this cloud with `ui.NewTester`, the Linux Pango text engine and MyGo's CPU renderer at **980 × 720 DIP, scale 1**, producing real native toolkit pixels. No HTML/browser mockup was used.

- [Overview](baselines/mygo-gallery-overview-linux-cpu-1x.png)
- [Virtual list](baselines/mygo-gallery-list-linux-cpu-1x.png)
- [Markdown](baselines/mygo-gallery-markdown-linux-cpu-1x.png)

These are original gallery view functions under a deterministic inspection fixture, not screenshots of an Orbit port. The fixture sets a fixed timestamp, disables route transitions and does not start the demo's live-data goroutine, so the chart is intentionally empty. Source: [capture harness](baselines/mygo-gallery-capture_test.go.txt). The images were opened and visually checked after creation. They show that native layout, readable text, SVG icons, controls, list rows and Markdown/code rendering work in this environment. They do not validate a real desktop window, GPU driver, OS IME candidate panel, screen reader or terminal session.

## Verification performed

Environment: cloud Linux amd64, Go 1.27.3, `CGO_ENABLED=0`. Go module/build caches and font cache were directed to cloud workspace storage. The initial snapshot-only run emitted fontconfig cache warnings but returned success and rendered visible text; the subsequent test command set a writable `XDG_CACHE_HOME`.

1. **Passed:** custom inspection harness renders three native gallery snapshots. The harness was temporary in the MyGo checkout and is retained as documentation only.
2. **Executed, not fully passing:** `go test ./ui/... ./examples/gallery ./examples/markdown-chat ./examples/code-editor`.
   - Passed packages: `ui/chart`, `ui/editor`, `ui/markdown`, `ui/markdown/parser`, `ui/richtext`, `ui/shader`, `examples/gallery`, `examples/markdown-chat`.
   - `examples/code-editor` compiles with no test files.
   - Failed package: `ui`, due to `TestDiffFrameAllocs` in `ui/bench_test.go:299–304`. Observed steady=2, sidebar scrolling=123, sidebar+diff=125, warm within-laid-out-rows=9; assertion requires steady and warm ≤2. The failure is an upstream performance-regression signal in this environment, not proof that all scrolling is slow and not an Orbit migration failure. No thresholds were changed and no passing rerun was substituted.
3. **Not performed:** Orbit UI port, pi end-to-end UI test, actual desktop GUI launch, GPU parity, terminal runtime/download, OS IME or assistive-technology runtime checks, code signing, installer/update validation. Cross-build results are recorded separately below once the commands finish.

### Reproduce native demo captures

Copy `baselines/mygo-gallery-capture_test.go.txt` into the pinned MyGo checkout as `examples/gallery/orbit_inspection_snapshot_test.go`, set `MYGO_INSPECTION_OUTPUT` to a writable output directory, and run:

```sh
CGO_ENABLED=0 go test ./examples/gallery -run '^TestOrbitInspectionSnapshots$' -count=1 -v
```

Use Go 1.27.3 and writable `GOPATH`, `GOCACHE`, `XDG_CACHE_HOME`. Remove the temporary test from the reference checkout afterward. This is a demo evidence harness, not a production app or visual parity test.

## Gated path to an honest 1:1 result

1. Native Go is confirmed. Define the exact pinned source/state used as the reference and record the architecture change.
2. Establish Go pi-RPC conformance and domain reducer tests, then native window/theme/font/asset infrastructure.
3. Port shell, sidebar, transcript and composer with stream/IME/scroll-anchor fixtures; produce same-platform matched screenshots.
4. Port remaining settings, Git/review, explorer/editor, worktrees, usage, provider/MCP/extension dialogs, notifications and updater behaviors against the full feature inventory. Do not call a shell/chat demo “all features.”
5. Validate terminal emulator/PTY behavior and native runtime packaging, then run actual macOS/Windows/Linux UI and accessibility tests in authorized cloud CI. Cross-compilation alone does not close this gate.
6. Require a green, reviewed behavior matrix plus per-state visual diffs before replacing the Rust reference or claiming pixel-level parity. Preserve Orbit's Apache-2.0/NOTICE, MyGo's MIT notice and each redistributed asset/library license.

### Completed cross-compilation

Both upstream `examples/gallery` and `examples/terminal` cross-built successfully with Go 1.27.3 and `CGO_ENABLED=0` for all six desktop targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64. These are build-only checks. They do not claim that a real GUI or terminal was run on those operating systems, nor that the native Ghostty asset loaded.

## Native implementation checkpoint (2026-10-25)

The native choice is now implemented, not just investigated. `cmd/orbit` runs a
Go/MyGo `ui.View`, with the direct `internal/agent.Manager` Pi JSONL runtime.
It does not load a webview, use a Rust backend, or replace Pi inference. The old
Rust sources and marketing PNGs remain reference material.

The current native view tree includes the shell, virtualized grouped session
sidebar, transcript, composer, model/command pickers, fourteen settings
categories, real Files/editor and six-source review, the six-tab Git page,
worktrees and Work-in picker, local usage, pooled terminals, native dialogs,
notifications/lifecycle and signed-update integration. Backend results, failures,
empty states and unsupported capabilities are rendered explicitly. The bug-report
form is mounted with explicit public-submit review. Custom extension UI uses the
native Ghostty ANSI bridge without starting a PTY. These are implementation facts,
not proof of complete functional or visual parity.

### Reproducible native capture

Run from the repository root, using the pinned Go toolchain:

```sh
ORBIT_CAPTURE_DIR=/absolute/output go test ./internal/ui -run '^TestNativeCapture$' -count=1
```

The test supplies an isolated temporary preference home/cache, English, Azure
dark, 14-DIP UI font, density 100, reduced motion, and a 1470×922-DIP viewport at
1×. General's displayed session path is fixed to `/home/orbit/.pi/agent` only in
the explicit capture fixture, so temporary test-directory names cannot change
its pixels. No Pi process, provider prompt, account, or invented usage is involved.
The transcript capture is labeled test content in the rendered image.

The four PNGs are in `baselines/native/`. Two independent captures using different
temporary homes produced identical SHA-256 hashes after that normalization:

| Capture | SHA-256 |
|---|---|
| `new-task-dark.png` | `ce97e3aa1cda472b0154dd8402ada24c55d54dea303677a99d3076173b0d9ee7` |
| `settings-appearance-dark.png` | `de5165512bd7dce9a9bb221dc0bc51e761d3537ea29d439ee7cc5f2fdd66acff` |
| `settings-general-dark.png` | `b422c6a7cce3ea49955896748c3f93c557b29eac565516f5909a646c5876b5b8` |
| `transcript-fixture-dark.png` | `07e23caacdc347c328f9bd2a57bb81ed36d3fb92c1d70881b85093713037ebbb` |

This records a specific checkpoint; later intentional UI changes must regenerate
and review hashes, rather than silently declare the old hashes current.

### Tests and remaining acceptance

Native tests exercise every parsed shortcut chord, palette uniqueness, fourteen
settings categories, command/model keyboard selection, CJK IME composition versus
submission, editor typing through the real native editor into disk persistence,
unsent drafts across session switches, GFM alerts, bottom-anchored composer and
2× raster dimensions. Disposable real-Git tests exercise rapid hidden-filter /
workspace scan replacement, review source/session/workspace replacement, and
persisted turn checkpoints after transcript compaction. The UI package passes
`go test -race` on the cloud Linux environment.

Still required before a one-to-one completion claim: an aligned rendered GPUI
baseline at the same current source revision; image-difference review for every
major surface and theme; real desktop event-loop/GPU validation on Linux, macOS
and Windows; OS IME, accessibility/screen-reader, drag/drop, native modal,
notifications, terminal and update acceptance; large-history performance; and the
per-feature behaviors still marked partial in `feature-parity.md`. CPU snapshots
and cross-compilation do not establish those results. The existing upstream
MyGo allocation-test failure described above is not waived by application tests.


### Integration corrections verified at 04:29 UTC

The UI/cmd checkpoint passes `go test -race ./internal/ui ./cmd/orbit`.
A second isolated capture matches all four PNG hashes above. The two Settings
hashes changed intentionally when source nav ordering/icons and Privacy-only
retention placement were corrected.

Additional checks now cover the actual Ghostty custom component receiving ESC
without closing its native modal, explicit cancellation, maintenance serialization
with preserved drafts, asynchronous retention/archive reload, split/collapsed
review navigation indexes, source-identity-safe cached review failures, stale
file-load selection after a workspace switch, dirty-path rename/delete refusal,
and capped truncated-file rendering. The maintenance gate protects idle parked
sessions from retention and prevents new prompts/process changes during runtime
replacement. Runtime quota and context data now comes from authoritative Pi
statistics, preserving unknown/null context after compaction instead of displaying
an invented zero. Native menus route through the same modal guard as shortcuts.

Remaining acceptance items are not removed by these checks. Newly corrected
filtered-picker cursor behavior, minimum-window modal reachability, native menu
accelerator precedence, modal focus restoration and non-integer scaling still
need their focused regression/desktop acceptance passes. The terminal platform
worker's native library integration is opt-in via `ORBIT_TERMINAL_TEST=1` and
`MYGO_GHOSTTY_VT` pointing at the pinned library; ordinary UI tests do not download
or start a terminal.
