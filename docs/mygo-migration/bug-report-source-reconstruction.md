# Report a bug: source reconstruction

Status: **PARTIAL**, not 1:1 accepted. The source-shaped native MyGo form and
bounded interaction tests are implemented. Mandatory public-report review,
opt-in diagnostics/identity, screenshot redaction, and the fork repository are
retained deviations. A new actual-window capture is still required for this
working-tree change; the attached renders are CPU test evidence only.

## Reference inspected

Pinned Rust: `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`.

- `crates/orbit-pi/src/app/bug_report_ui.rs`: six-row report board,
  segments, contextual field labels, notes-only generation, clear, image cap,
  inline errors, and submission lifecycle.
- `crates/orbit-pi/src/app/settings.rs::settings_section_desc/settings_group`:
  fixed 960px settings shell, 24px content insets, section text, bordered board,
  16px-inset 1px row separators.
- `crates/orbit-pi/src/app/helpers.rs::button_frame/input_field_frame` and
  `theme/tokens.rs`: source dimensions, phi line height, input padding/radius,
  medium buttons, lightness-gradient button faces/contact shadows.
- `crates/orbit-pi/src/issue_message.rs::ReportKind/generate_report` and
  `bug_report.rs`: kind labels/prompts, actual report body headings,
  fallback drafts, diagnostics collection, GitHub/browser behavior.
- Actual 1470×922 native PNGs were viewed before implementation:
  `reference-capture-f61e00f/reference-desktop/settings-report-bug-dark.png`
  and `native-capture-81a5114/go-desktop/settings-report-bug-dark.png` in the
  cloud verification workspace. The latter is the old generic Go page, not
  evidence for this reconstruction.

## Reconstructed form

`internal/bugreportui/view.go` now owns a fixed settings header and separately
scrolling body, matching the host's other source settings pages. No webview or
browser substitute is used.

| Source contract | Current Go implementation/evidence |
|---|---|
| 960 outer width, 912px content at default viewport | 960 max-width shell with 24px horizontal insets |
| Uppercase section caption + description, one bordered board | Six rows; 12px board radius; 16px inset dividers |
| Issue type on left; three abutting segments on right | Bug report / Feature request / Other; Bug default; real pointer/keyboard buttons |
| Context row action on the right | Live `Using <model>` caption; omit empty or `…`; Generate draft / Generating… |
| Context 96, description 120, steps 84 px | Source fixed field heights, whole-box multiline editing, scaled input minimum/padding |
| Single-line title input | Source phi text line, 6px/8px input padding, border and 8px radius |
| Kind-specific labels | What happened? / Steps to reproduce; What problem does this solve? / Proposed solution; Details / Additional details |
| Screenshot row/chips and six-image cap | Add screenshot at right; up to six unique names; valid images survive a bad selection; remove restores capacity |
| Bottom Clear + Create issue on GitHub | Both reachable in the 922px default viewport; create opens retained review gate |
| Errors and loading | Empty Context/title/body errors; worker errors preserve fields; busy actions reject duplicates |
| Clear | Clears context/title/body/steps/screenshots/error; retains selected kind and privacy choices |
| Cancellation | Cancel keeps draft, waits for worker exit before unlocking repeated actions; stale results after Clear/Close ignored |

The primary button uses source lightness lift/sink and short contact/ambient
shadows. The host supplies live palette/density/model callbacks through
`bugreportui.Options.Theme` and `ModelLabel` in `internal/ui/run.go`.

## Privacy deviations preserved, not claimed as parity

The pinned source automatically collects Orbit version, App ID, host OS,
architecture, installation type, pi/gh versions, and the authenticated GitHub
username at submission. Its main Create action files directly or opens a
prefilled external page. Its footer explicitly promises automatic attachment.

This fork's pre-existing behavior is intentionally **not silently removed**:

1. Diagnostics and GitHub identity remain off by default; recognizable-secret
   and home-path text redaction remains on. The extra controls are in a clear
   `Privacy…` dialog instead of expanding the source form beyond the viewport.
2. A local report preview and explicit public-submission checkbox remain
   mandatory. The source-labeled Create button opens review; it does not create
   an issue or open a browser by itself.
3. Header/section/footer copy is truthful about review and selected diagnostics,
   rather than claiming disabled diagnostics will be attached automatically.
4. Current-window capture requires an explicit checkbox and action in Privacy.
   Manually selected images are normalized to PNG with metadata removed;
   screenshot preview/redaction stays available. Source screenshot chips only
   remove; Go splits preview and removal into separate explicit targets.
5. Revoking identity inclusion removes any cached GitHub username from the
   prepared report even if an old diagnostic snapshot still contains it.
6. Existing default repository is `charlzyx/orbit`, while the pinned upstream
   source uses `imrj05/orbit`. No destination change is made here.
7. Generation keeps the existing isolated, notes-only Pi JSONL RPC path rather
   than Rust's one-shot `pi -p`. Model errors produce an honestly labeled local
   fallback with its warning. No workspace/transcript/diagnostic/image context
   is implicitly sent for drafting.
8. Cancellation of submission cannot establish whether a remote create already
   happened. The UI keeps the report and explicitly asks the user to inspect
   the repository before retrying; it never auto-retries that uncertain write.

The top description and footer, public-review step, options dialog, chip
preview action, and repository therefore remain explicit **PARTIAL** items.
New supplemental privacy/cancellation copy is English fallback text rather
than a complete source-locale translation set. The source labels and
placeholders use the existing locale keys.

## Verification

Focused commands (canonical pinned-overlay wrapper):

- `scripts/go-native.sh test -race ./internal/bugreportui -count=1`
- `scripts/go-native.sh vet ./internal/bugreportui`
- `git diff --check -- internal/bugreportui`

The package contains 18 scoped tests. New source tests cover:

- empty-field validation without invoking a model;
- kind labels and error clear without dropping entered text;
- generation retry/error preservation and notes-only input;
- cancel/repeated-click gate and stale result protection;
- first-six screenshot cap, name dedupe, mixed valid/invalid selections,
  cancelled picker, preview, pixel-redaction action, and removal;
- diagnostics/identity defaults and cached-identity revocation;
- failed submission preserving all fields and requiring a fresh review;
- full native-shell CPU pointer, Tab, Space, Escape, local Copy, and clear;
- model caption changes and fixed-board/footer geometry;
- fake GitHub/URL counters proving opening, generating locally, selecting
  images, preparing/reviewing/copying/dismissing do not transmit anything.

Tests use isolated temporary homes, local generated PNGs, controlled callbacks,
and fake GitHub/URL adapters. They do not create real issues, open a real
external browser, call a paid model, or access user credentials.

### Geometry evidence, not actual-window acceptance

The full-app CPU fixture uses the real host theme, bundled fonts, settings
navigation, and mounted report view at 1470×922, density 100, 14px UI type.
`probes/bugreportui/bug-report-source-geometry.json` records control bounds.

| Element | Pinned native reference | Go CPU fixture |
|---|---:|---:|
| Board x / y | 399 / 143 | 399 / 142.81 |
| Board width / height | 912 / 676 | 912 / 676.08 |
| Generate y / height | 209 / 28 | 208.81 / 28 |
| Add screenshot y / height | 756 / 28 | 755.71 / 28 |
| Footer button y / height | 839 / 28 | 838.90 / 28 |

CPU PNGs in `probes/bugreportui/` include source-shaped idle, feature,
generating, selected screenshots, screenshot preview, public review, and
Privacy. These were visually inspected. They do not establish exact GPU glyph
rasterization, native file-picker operation, real-model generation, actual
GitHub attachment/browser behavior, or actual OS accessibility. No new
native-to-native changed-pixel percentage is claimed from CPU images.

Remaining: hosted actual-window recapture for this checkpoint, comparison
against the pinned native reference, native picker/window capture/platform
keyboard exercises, complete locales, and explicit resolution of the privacy
and repository deviations before this category can be called 1:1.
