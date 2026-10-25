package orbitui

import "github.com/egoist/mygo/ui"

type attachmentPathInsertion struct {
	owner attachmentOwner
	text  string
}

// Retain data/owner only, never MyGo's build-scoped Element or TextKey. Native
// dialogs and drop delivery can complete outside the composer's current build.
func (a *App) queueAttachmentPath(owner attachmentOwner, text string) {
	a.pendingAttachmentPaths = append(a.pendingAttachmentPaths, attachmentPathInsertion{owner, text})
	if w := a.window.Load(); w != nil && !w.IsDestroyed() {
		w.Invalidate()
	}
}

// Discard obsolete intents even while Settings/Review hides the composer.
func (a *App) pruneAttachmentPaths() {
	kept := a.pendingAttachmentPaths[:0]
	for _, pending := range a.pendingAttachmentPaths {
		if a.ownsAttachment(pending.owner) {
			kept = append(kept, pending)
		}
	}
	clear(a.pendingAttachmentPaths[len(kept):])
	a.pendingAttachmentPaths = kept
	if len(kept) == 0 {
		a.pendingAttachmentPaths = nil
	}
}
func (a *App) flushAttachmentPaths(composing bool, caret func() (int, int), replace func(int, int, string) bool) bool {
	changed, consumed := false, 0
	for _, pending := range a.pendingAttachmentPaths {
		if !a.ownsAttachment(pending.owner) {
			consumed++
			continue
		}
		if composing {
			break
		}
		_, at := caret()
		// Rust ComposerInput::insert_at_caret inserts at the active end, without
		// deleting selected text. Each file is its own native undo transaction.
		if !replace(at, at, pending.text) {
			break
		}
		consumed++
		changed = true
	}
	if consumed == len(a.pendingAttachmentPaths) {
		a.pendingAttachmentPaths = nil
	} else {
		a.pendingAttachmentPaths = a.pendingAttachmentPaths[consumed:]
	}
	if changed {
		a.autocomplete.Dismiss()
	}
	return changed
}
func (a *App) flushAttachmentPathInput(c *ui.Context, input ui.Element) {
	if len(a.pendingAttachmentPaths) == 0 {
		return
	}
	editor, ok := any(input).(interface{ ReplaceTextRange(int, int, string) bool })
	if !ok {
		return
	}
	if a.flushAttachmentPaths(input.Composing(), input.TextAnchorCaret, editor.ReplaceTextRange) {
		input.Focus()
		a.retainReviewComposer(c, input)
		c.AnimationFrame()
	}
}
func (a *App) flushAttachmentPathKey(k ui.TextKey) bool {
	if len(a.pendingAttachmentPaths) == 0 {
		return false
	}
	editor, ok := any(k).(interface {
		AnchorCaret() (int, int)
		Composing() bool
		ReplaceRange(int, int, string) bool
	})
	if !ok {
		return k.Key == ui.KeyEnter
	}
	if a.flushAttachmentPaths(editor.Composing(), editor.AnchorCaret, editor.ReplaceRange) {
		a.composer.Focus()
	}
	// Native platforms filter IME candidate keys before OnTextKey (Darwin
	// surface.go keyDown). This guards a raw/synthetic Enter from submitting
	// the old draft; committed TextInput remains the editor's responsibility.
	return len(a.pendingAttachmentPaths) > 0 && k.Key == ui.KeyEnter
}
