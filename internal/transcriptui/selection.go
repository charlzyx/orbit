package transcriptui

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
)

// SelectionText implements TextSelection::selected_text's clipboard contract:
// skip empty ranges and join selected visible text blocks with one newline.
// MyGo uses rune offsets, unlike Rust's UTF-8 byte offsets, so slicing is by rune.
func SelectionText(parts []ui.SelectedText) string {
	var selected []string
	for _, p := range parts {
		r := []rune(p.Text)
		from, to := max(0, min(p.Start, len(r))), max(0, min(p.End, len(r)))
		if from > to {
			from, to = to, from
		}
		if from < to {
			selected = append(selected, string(r[from:to]))
		}
	}
	return strings.Join(selected, "\n")
}
func SelectionData(parts []ui.SelectedText) transfer.Data {
	return transfer.TextData(SelectionText(parts))
}

type selectionMenuIntent struct {
	host               uint64
	message, selection string
}
type selectionMenuIntentKey struct{}

// SelectionMenu supplies the pinned source's Copy Selection / Copy Message
// menu. Pass the shared Selectable ancestor and the currently hit message's
// answer text, not a different nested selection container. The host identifies
// the hovered logical row; the selected range may span several rows.
func SelectionMenu(c *ui.Context, scope ui.Element, messageText string, o Options) {
	// Local lifetime is the actual owner-keyed selectable/menu host. Retiring
	// that host retires its sole intent; each new open replaces these strings.
	intent := ui.Local(scope, selectionMenuIntentKey{}, func() selectionMenuIntent { return selectionMenuIntent{} })
	copyText := func(text string) {
		if o.Copy != nil {
			o.Copy(text)
		} else {
			c.WriteClipboard(text)
		}
	}
	scope.CopyData(SelectionData).ContextMenu(func(m *ui.Menu) {
		// Keep indices and labels fixed across the public collect/Chosen phases.
		selected := m.Item(o.tr("transcript.copy_selection", "Copy Selection"))
		message := m.Item(o.tr("transcript.copy_message", "Copy Message"))
		if selected.Chosen() {
			if intent.host == scope.ID() && intent.selection != "" {
				copyText(intent.selection)
			}
			*intent = selectionMenuIntent{}
			return
		}
		if message.Chosen() {
			if intent.host == scope.ID() && intent.message != "" {
				copyText(intent.message)
			}
			*intent = selectionMenuIntent{}
			return
		}
		// ContextMenu calls the builder only to collect a newly opened menu or
		// replay a chosen item. No chosen item means this is the opening payload.
		*intent = selectionMenuIntent{host: scope.ID(), message: messageText, selection: SelectionText(scope.SelectedTexts())}
		selected.Disabled(intent.selection == "")
		message.Disabled(intent.message == "")
	})
}

// RetainedSelectionAvailable reports whether the canonical hash-checked MyGo
// compatibility addition is present. Unpatched standalone tests may still
// exercise mounted selection, but the application must fail before startup.
func RetainedSelectionAvailable() bool {
	_, ok := any(ui.Element{}).(interface {
		RetainTextSelection(string) ui.Element
		RetainTextSelectionInteraction(ui.Element) ui.Element
	})
	return ok
}

// RetainSelection enables the source's immutable selection snapshot semantics
// on the shared transcript scope. identity must be the actual stable session
// identity, never a row count or latest turn. It must be called every frame.
func RetainSelection(scope ui.Element, identity string) bool {
	if patched, ok := any(scope).(interface{ RetainTextSelection(string) ui.Element }); ok {
		patched.RetainTextSelection(identity)
		return true
	}
	return false
}

func RequireCompatibility() error {
	if !RetainedSelectionAvailable() {
		return fmt.Errorf("Orbit requires its pinned MyGo transcript selection compatibility patch; use scripts/go-native.sh run ./cmd/orbit or scripts/go-native.sh build ./cmd/orbit")
	}
	return nil
}

// RetainSelectionInteraction grants the rail's explicitly scoped pointer
// interaction, not ordinary controls, the source's selection preservation.
func RetainSelectionInteraction(control, scope ui.Element) bool {
	if patched, ok := any(control).(interface{ RetainTextSelectionInteraction(ui.Element) ui.Element }); ok {
		patched.RetainTextSelectionInteraction(scope)
		return true
	}
	return false
}
