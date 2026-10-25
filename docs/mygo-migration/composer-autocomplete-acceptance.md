# C04: Composer autocomplete source contract

Reference: Orbit `299a32bc87fe3d4851bdb902d9c03dcf9bf99cac`;
`crates/orbit-pi/src/mentions.rs`, `app/composer_ops.rs:9–300`,
`app/events.rs:913–953`, `app/session.rs:88–105,378–416,456–460`,
`composer.rs:297–342,465–497,1365–1405`, `theme/tokens.rs:589–675`.

This acceptance contract was written before the Go implementation. Rust is read
as a reference only, never executed as the delivered runtime.

## Exact behavior to port

1. `/` triggers only at byte zero, while the caret is within the leading token.
   Query stops at the caret; replacement extends through the entire slash token,
   preserving later whitespace and arguments. `/` alone lists all commands.
2. `@` triggers only at start or following Unicode whitespace. The last `@`
   before the caret wins, and no query whitespace is allowed. Email addresses
   stay plain. Replacement stops at the caret, preserving a token's suffix.
3. Fuzzy subsequence scoring is case-insensitive: consecutive +3, beginning or
   slash-segment +2, each gap penalized by min(gap,12). Commands add 10. Files
   use max(2 × basename score, whole-path score). Stable original catalog order
   breaks ties. Descriptions are not searched. At most eight rows render.
4. Real Pi `get_commands` supplies name, description, source and sourceInfo.
   Missing name is omitted, missing description is empty. Scope precedence is
   skill; home/.orbit-pi path; project basename; package/builtin/synthetic path;
   user custom; fallback builtin. A refreshed catalog updates an open popup.
5. File catalog is real local workspace enumeration, relative slash-separated
   sorted paths, excluding hidden entries and dependency/build/VCS directories;
   at most 4,000 files, directory depth 12. Dot directories are skipped, not
   `.gitignore` matching. Directory symlinks are not followed. Baseline caches
   only the current workspace and refreshes on workspace change.
6. Up/Down wrap while open; no candidates leaves ordinary editor keys alone.
   Enter, alternate-send Enter, explicit-steer Enter and Tab accept the
   highlight before submission. Shift+Enter stays a newline. Closed Tab uses
   ordinary focus traversal. Completion adds exactly one trailing space and
   leaves the caret immediately after it. It is an undoable text edit.
7. Escape and outside press dismiss until trigger kind/query changes. Source
   code clamps the existing highlight after filtering, despite a misleading
   comment saying it resets. No-trigger state clears highlight to zero.
8. Only the main composer owns the menu; other inputs cannot accept its rows.
   IME composition is never submitted or completed. No caret, selection or
   Unicode splitting is introduced by mention coloring or replacement.
9. Command display: extension icon, `/name`, dim description and scope badge.
   File display: file-type glyph/badge, basename then dim directory including
   trailing slash, no scope. Selected and hover colors follow Orbit tokens.
10. Popup anchors above composer, full 960px content width (window-constrained),
    5px separation, 8px window margin, at most eight fixed 31px default rows,
    4px outer row inset, 6px inner horizontal padding/radius, 4px vertical list
    padding, 18px scope badge, modal-surface elevation. Density/font scaling
    must preserve the source token formulas rather than generic app styling.
11. Paint-only mentions color a leading command only if its name is ASCII
    alphanumeric or `-_:`, and every nonempty token-boundary `@file`. Leading
    absolute paths and emails stay normal. Source spans use UTF-8 byte offsets;
    native MyGo adapters must convert to its documented rune offsets.

## Native safety/lifecycle obligations

The source scans synchronously and has no loading/error UI. The Go port moves
I/O off the UI thread without inventing success or results. Loading, cancelled,
empty, failed and partially readable scans remain distinguishable diagnostics;
no fake row is selectable. Workspace/session changes cancel old work and reject
late results; close stops outstanding scans. File lists remain cached for the
current workspace. Explicit refresh is allowed but not auto-triggered every
frame. A failed catalog request must not resurrect another session's commands.

MyGo Element handles are build-scoped. Keyboard callbacks must not read them
between frames. Deferred completions are revalidated against live input text,
focus, composition and caret in the next native build, then preserve the new
caret after the editor receives the replacement. Newer input cancels stale
pending completion instead of editing the wrong token.

## Evidence gates

Pure Go source-rule tests; real temporary filesystem and cancellation/fault
fixtures; MyGo CPU native input tests for key/click/focus/IME/undo/repeated flows;
real Manager get_commands projection (not fabricated production data). CPU
rendering is not native OS execution or measured GPUI pixel equality. C04 stays
partial until the mounted application and OS/pixel gates are independently
verified. Test fixtures are explicitly fixtures, not claims of live inference.

## Go implementation and verified evidence (2026-10-25)

Implementation: `internal/composerui/{catalog,files,state,native,glyphs}.go`.
The host owner mounted it in `internal/ui/chat.go`, preserved raw RPC commands
in `Snapshot.CommandsRaw`, added topmost Escape handling and lifecycle Close,
and added the mandatory compatibility check in `cmd/orbit/main.go`.

- Source rules: detection/replacement, full Unicode lowercase, fuzzy ranking,
  stable ties, 8-row cap, provenance precedence, paint spans, file badges,
  591 exact devicons filename/extension mappings with both palettes.
- Real filesystem: excluded trees, hidden files, symlinks without traversal,
  depth/count caps, missing-root errors, partial results, explicit refresh,
  cache retention, context cancellation, stale workspace/session rejection.
- Native CPU interaction: Up/Down wrapping, Enter/Alt-Enter/Cmd-Shift-Enter,
  Tab completion, closed Tab focus traversal, no-results handling, Escape,
  outside dismissal, pointer acceptance mid-token, other-input focus isolation,
  IME non-submission, Shift-Enter newline, Unicode caret/suffix preservation,
  repeated completion/send, exact completion undo/redo and subsequent typing.
- Toolkit regression: live queued navigation/text before a frame, string and
  indexed buffers, reversed selection, range clamping, pointer transactions,
  readonly/disabled/IME refusal and callback lifetime. Existing OnTextKey,
  TextArea undo and indexed-input tests also passed with the patch applied.
- Real Pi: `TestLivePiCommandCatalogCredentialFree` starts pinned Pi 1.4.0
  through the actual Go Manager in an isolated HOME and workspace. Controlled
  `.pi/prompts/contract.md` and `.pi/skills/scan/SKILL.md` are loaded offline;
  `get_commands` returns real descriptions/provenance and drives the controller.
  Extensions/tools/context files are disabled. No prompt, inference, credentials
  or external writes are used. Project resource loading requires the Manager's
  existing `--approve`; adding `--no-approve` correctly returns no project
  commands and was removed from this isolated test fixture.

Canonical commands (Go 1.27.3; `ORBIT_GO` can select its executable):

    scripts/go-native.sh test -race ./internal/composerui
    scripts/go-native.sh vet ./internal/composerui
    scripts/go-native.sh test -race github.com/egoist/mygo/ui -run 'TestOrbitComposer|TestOnTextKey|TestTextAreaUndo|TestBufferInputs'
    scripts/go-native.sh test -race ./internal/ui -run 'TestNativeComposerCompletionPrecedesSendAndAbort|TestIMECompositionDoesNotSubmit'
    scripts/go-native.sh -C third_party/mygo-terminal test -race ./...
    python3 scripts/test-go-native-overlay.py

For the non-skipped real Pi test, set `ORBIT_PI_INTEGRATION_BIN` to the pinned
1.4.0 binary. `ORBIT_REQUIRE_COMPOSER_PATCH=1` makes a missing exact-edit API fail
rather than skip. All 24 component tests plus fuzz seed coverage passed; the
final component race run including real Pi passed in 2.805 seconds; patched toolkit regressions passed in 1.545 seconds; mounted
host completion/IME tests passed in 4.346 seconds. The terminal submodule
wrapper path passed all packages (main terminal package 66.570 seconds).
These timings are recorded observations, not performance assertions.

The wrapper verifies module sum identities, `go mod verify`, seven upstream
source hashes and patch/test hashes. It uses a cleaned-up temporary verified
source copy + temporary modfile + Go overlay, because Go 1.27 forbids overlays
inside GOMODCACHE. The repository go.mod/go.sum and module cache are not edited.
A cold-cache check used a separate temporary GOMODCACHE with official archives
and no extracted MyGo; exact `go mod download` materialization succeeded with
GOPROXY=off. Inherited GOFLAGS and local terminal replacement are preserved.
The patch and its tests are reviewable under `third_party/mygo-composer/`.

## Visual evidence and remaining verification limits

`probes/composerui/composer-slash-native-cpu.png` and
`probes/composerui/composer-files-native-cpu.png` are isolated MyGo CPU
component fixtures, not screenshots of a running GPUI/native OS application.
The geometry tests check 960px width, 31px default row metric, exactly 5px
separation above the composer and source 8px viewport clamping after resize.
A zero-size native anchor is positioned at the source-clamped bounds before
MyGo places the popup, avoiding toolkit side-flipping or its own 4px margin. Native file rows use registered/installed Nerd Font families in source preference order, otherwise the source 17px extension
badge. Static icon data is attributed to devicons 0.6.13 under Apache-2.0;
its original source hashes and full license are retained in the package.

Still not established: same-state GPUI/native OS pixel equality on each target,
platform IME event delivery, and the source's actual PUA-glyph availability
probe for externally installed fonts (the current public MyGo font API lists
families but does not expose a no-fallback glyph-presence test). The bundled Nerd Font is separately parsed with sfnt to verify the exact
PUA probe glyph, registered with MyGo, selected and rendered in a real CPU
file-popup test. Externally installed fonts still require native OS coverage
verification, not an invented glyph-success result.

C04 is implemented and source/CPU/real-RPC tested, not a claim of completed
native OS or pixel-perfect parity. Plain unpatched application startup refuses
to run and tells the user to use the canonical wrapper.
