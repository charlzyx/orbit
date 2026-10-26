package orbitui

import (
	"strconv"
	"strings"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

const codeFoldThreshold = 28
const codeFoldPreviewLines = 24

type codeFoldBlock struct {
	ordinal, lines int
	preview        string
	options        markdown.Options // owned in place; never copy Options' cache
}
type codeFoldState struct {
	document *parser.Document
	revision uint64
	blocks   map[int]*codeFoldBlock
	expanded map[int]bool
}

func (s *codeFoldState) prepareEligible(d *parser.Document, revision uint64, streaming, allowFold bool) {
	if allowFold && !streaming {
		s.prepare(d, revision)
	}
}

// prepare is source/revision bounded, not per-frame Code() or body hashing.
// Parser identity matters if the public renderer replaces its document without
// changing the supplied string (for example an extension configuration change).
func (s *codeFoldState) prepare(d *parser.Document, revision uint64) {
	if s.document == d && s.revision == revision {
		return
	}
	s.document, s.revision = d, revision
	blocks := make(map[int]*codeFoldBlock)
	live := make(map[int]bool)
	ordinal := 0
	for i := range d.Children(0) {
		b := d.Block(i)
		if b.Kind == parser.CodeBlock && b.Fenced {
			live[ordinal] = true
			lines := d.LineCount(i)
			if lines > codeFoldThreshold {
				blocks[i] = &codeFoldBlock{ordinal: ordinal, lines: lines, preview: codeFoldLiteralPreview(d, i)}
			}
		}
		ordinal++
	}
	s.blocks = blocks
	for ordinal := range s.expanded {
		if !live[ordinal] {
			delete(s.expanded, ordinal)
		}
	}
}

// The preview is only literal code, including any backticks, HTML, links or
// images in its payload. A tilde delimiter longer than every payload run makes
// it impossible for a supplied line to escape into a second renderer/loader.
func codeFoldLiteralPreview(d *parser.Document, index int) string {
	lines := min(codeFoldPreviewLines, d.LineCount(index))
	var body strings.Builder
	longest := 2
	for k := 0; k < lines; k++ {
		line, _ := d.Line(index, k)
		run := 0
		for _, r := range line {
			if r == '~' {
				run++
				longest = max(longest, run)
			} else {
				run = 0
			}
		}
		body.WriteString(line)
		body.WriteByte('\n')
	}
	fence := strings.Repeat("~", longest+1)
	return fence + " " + d.Block(index).Language() + "\n" + body.String() + fence + "\n"
}

func (a *App) codeFoldBody(c *ui.Context, b *markdown.CodeBlockView, state *chatMarkdownState, owner inlineImageToken, revision uint64, streaming, allowFold bool) {
	block := state.fold.blocks[b.Index]
	if !allowFold || streaming || block == nil {
		b.Body(c)
		return
	}
	expanded := state.fold.expanded[block.ordinal]
	if expanded {
		b.Body(c)
	} else {
		o := &block.options
		o.Theme = a.md.Theme
		o.ShareSelection = a.md.ShareSelection
		o.WrapCode = a.md.WrapCode
		// This literal-only document intentionally has no app Block hook, image
		// loader, HTML renderer or nested code fold/copy header.
		o.CodeBlock = func(c *ui.Context, preview *markdown.CodeBlockView) { preview.Body(c) }
		markdown.View(c.Key("code-fold-literal-prefix-"+strconv.FormatUint(revision, 10)), block.preview, o)
	}
	label := a.Tr("transcript.show_less")
	if !expanded {
		label = orbitassets.Translate(a.catalog, a.english, "transcript_view.show_remaining_lines", map[string]string{"count": strconv.Itoa(block.lines - codeFoldPreviewLines)})
	}
	t := a.Theme
	button := ui.ButtonBase(c).Key("code-fold-"+strconv.Itoa(block.ordinal)+"-revision-"+strconv.FormatUint(revision, 10)).AlignSelf(ui.Stretch).Height(30).PaddingX(t.Space(12)).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).AlignItems(ui.Center).Gap(t.Space(6)).Label(label)
	color := t.TextMuted
	if button.Hovered() {
		button.Background(t.Hover)
		color = t.Text
	}
	button.Children(func() {
		ui.Text(c, label).FontSize(t.Px(12)).FontWeight(500).TextColor(color).Unselectable()
		ui.Spacer(c)
		glyph := "icons/chevron-right.svg"
		if expanded {
			glyph = "icons/chevron-down.svg"
		}
		a.icon(c, glyph, t.Px(12), t.TextMuted)
	})
	if button.Clicked() && state.revision == revision && a.ownsInlineImage(owner) {
		if state.fold.expanded == nil {
			state.fold.expanded = make(map[int]bool)
		}
		state.fold.expanded[block.ordinal] = !expanded
	}
}
