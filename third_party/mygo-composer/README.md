# MyGo composer compatibility addition (proposal/test overlay)

Pinned upstream: `github.com/egoist/mygo`
`9c69cf829ece196c5b3a3601f1c7bbe70526376f`, MIT (LICENSE retained).

The single added source file in `composer-edit.patch` extends only opt-in native
text editing. It does not change framework initialization, GPU backends,
platform bindings, source pin, or any existing API semantics.

## Why the public API is insufficient

- `ui/textinput.go`: TextKey offers Text/SetText, but no live caret or edit range.
- `ui/frame.go`: Element is build-scoped, invalid between input callbacks.
- `ui/editor_navigation.go`: textKey prepares queued input before invoking the
  callback, making a callback-local caret API the correct event-order boundary.
- `ui/editor_history.go`: SetText merges app changes into the prior undo step;
  `record(false)` + `replace` creates a separate, reversible completion edit.
- `ui/textbuffer_input.go`: externally replacing indexed TextBuffer content
  resets native undo, not a solution for composer transaction parity.

## Added methods

- TextKey.AnchorCaret() (anchor, caret int), rune offsets, callback-only.
- TextKey.Composing() bool, callback-only.
- TextKey.ReplaceRange(start, end int, text string) bool, one undo transaction.
- Element.ReplaceTextRange(start, end int, text string) bool, build-only for
  pointer completion. Stale, disabled, read-only and composing inputs refuse.

The common helper uses native editor operations, then publishes through the
existing widget input path, marks change, and requests a frame. It preserves
undo/redo and selection metadata, and has no reflection/unsafe access.

`baseline.json` records the exact hashes of every upstream implementation used
for the design. The implementation is a `.go.txt` to keep normal application
package discovery from treating this review artifact as a separate package.

Canonical entry point: `scripts/go-native.sh <go arguments>`, optionally with
`ORBIT_GO=/path/to/go`. It verifies module sum identities, `go mod verify`,
baseline file hashes and patch/test hashes. Go 1.27 rejects overlays directly
in GOMODCACHE, so the wrapper copies the verified source to a temporary
directory, makes a temporary modfile with absolute replacements (including the
existing terminal fork), and adds only the reviewed Go file through an overlay.
The source copy, modfile and overlay are removed after the command. Neither the
repository go.mod/go.sum nor the module cache is written. Leading `-C` and the
original CLI arguments/environment are preserved; GOFLAGS quoting follows Go's
own grammar, and a nonconflicting caller overlay is merged safely.

Application startup calls RequireCompatibility and fails with the canonical
wrapper instruction if the patch is absent. The unpatched path exists only for
isolated regression tests and is never presented as exact undo/caret parity.

The additional test source runs via:

    scripts/go-native.sh test -race github.com/egoist/mygo/ui -run TestOrbitComposer

It covers queued native events before a render, indexed and string controls,
Unicode, reversed selection, exact undo/redo transactions, typing after a
completion, pointer edits, read-only/disabled/IME boundaries and TextKey lifetime.
