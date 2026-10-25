package transcriptui

import (
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

// SelectionMenu supplies the pinned source's Copy Selection / Copy Message
// menu. Pass the shared Selectable ancestor and the currently hit message's
// answer text, not a different nested selection container. The host identifies
// the hovered logical row; the selected range may span several rows.
func SelectionMenu(c *ui.Context, scope ui.Element, messageText string, o Options) {
	scope.CopyData(SelectionData).ContextMenu(func(m *ui.Menu) {
		selected := SelectionText(scope.SelectedTexts())
		if m.Item(o.tr("transcript.copy_selection", "Copy Selection")).Disabled(selected == "").Chosen() {
			if o.Copy != nil {
				o.Copy(selected)
			} else {
				c.WriteClipboard(selected)
			}
		}
		if m.Item(o.tr("transcript.copy_message", "Copy Message")).Disabled(messageText == "").Chosen() {
			if o.Copy != nil {
				o.Copy(messageText)
			} else {
				c.WriteClipboard(messageText)
			}
		}
	})
}
