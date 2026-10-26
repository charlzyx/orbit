package orbitui

import (
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

type chatMarkdownState struct {
	options  markdown.Options // owned in place; never copy its internal sync.Map
	source   string
	revision uint64
	fold     codeFoldState
}
type chatMarkdownStateKey struct{}
type codeCopyFeedback struct{ at time.Time }
type codeCopyFeedbackKey struct{}

// chatMarkdown supplies immutable source identity without calling Code() on
// every frame. Only the header is revision-keyed: Body keeps toolkit selection,
// incremental highlighting and horizontal scroll state unchanged.
func (a *App) chatMarkdown(c *ui.Context, source string, streaming bool) {
	// Preserve the caller's slot key before adding an owner-specific subtree.
	a.chatMarkdownWithFolding(c, source, streaming, true)
}

func (a *App) chatMarkdownWithFolding(c *ui.Context, source string, streaming, allowFold bool) {
	ui.Column(c).Children(func() { a.chatMarkdownOwned(c, source, streaming, allowFold) })
}

func (a *App) chatMarkdownOwned(c *ui.Context, source string, streaming, allowFold bool) {
	owner := a.inlineImageToken()
	root := ui.Column(c).Key("chat-markdown-owner-" + strconv.FormatUint(owner.generation, 10))
	state := ui.Local(root, chatMarkdownStateKey{}, func() *chatMarkdownState { return &chatMarkdownState{} })
	s := *state
	if s.revision == 0 || s.source != source {
		s.source = source
		s.revision++
	}
	revision := s.revision
	o := &s.options
	o.Theme = a.md.Theme
	o.Extensions = a.md.Extensions
	o.Code = a.md.Code
	o.Math = a.md.Math
	o.HTML = a.md.HTML
	o.ShareSelection = a.md.ShareSelection
	o.LoadImage = a.md.LoadImage
	o.OnLink = a.md.OnLink
	o.Block = func(c *ui.Context, d *parser.Document, index int) bool {
		s.fold.prepareEligible(d, revision, streaming, allowFold)
		if a.md.Block != nil {
			return a.md.Block(c, d, index)
		}
		return false
	}
	o.WrapCode = a.md.WrapCode
	o.Reveal = a.md.Reveal
	o.CodeBlock = func(c *ui.Context, b *markdown.CodeBlockView) {
		th := o.Theme
		box := ui.Column(c).Direction(ui.LeftToRight).Background(th.CodeBlockBackground).Border(1, th.CodeBlockBorder).Radius(th.CodeBlockRadius).AlignSelf(ui.Stretch).Clip().Role(ui.RoleGroup)
		label := "Code"
		if b.Language != "" {
			label += ", " + b.Language
		}
		box.Label(label)
		box.Children(func() {
			a.codeCopyHeader(c, b.Language, owner, revision, func() bool { return s.revision == revision }, b.Code)
			a.codeFoldBody(c, b, s, owner, revision, streaming, allowFold)
		})
	}
	root.Children(func() {
		if streaming {
			markdown.Stream(c, source, o)
		} else {
			markdown.View(c, source, o)
		}
	})
}

func (a *App) codeCopyHeader(c *ui.Context, language string, owner inlineImageToken, revision uint64, current func() bool, read func() string) {
	th := a.md.Theme
	em := th.FontSize
	head := ui.Row(c).Key("code-copy-revision-"+strconv.FormatUint(revision, 10)).AlignItems(ui.Center).Padding(em*.35, em*.4, 0, em*th.CodeBlockPadding)
	state := ui.Local(head, codeCopyFeedbackKey{}, func() codeCopyFeedback { return codeCopyFeedback{} })
	head.Children(func() {
		ui.Text(c, a.codeCaption(language)).FontSize(em * .8).TextColor(th.Muted).Unselectable()
		ui.Spacer(c)
		copied := !state.at.IsZero() && c.Now().Sub(state.at) < 2*time.Second
		button := ui.ButtonBase(c).Padding(em*.2, em*.45).Radius(5).Gap(em * .3).Label(a.Tr("transcript.copy_code"))
		if button.Hovered() {
			button.Background(th.Text.Alpha(.08))
		}
		button.Children(func() {
			icon := "icons/copy.svg"
			if copied {
				icon = "icons/check.svg"
			}
			a.icon(c, icon, em*.95, th.Muted)
			ui.Text(c, a.Tr("common.copy")).FontSize(em * .8).TextColor(th.Muted)
		})
		if button.Clicked() && current() && a.ownsInlineImage(owner) {
			// Document.Lines adds exactly one newline after every line. Remove only
			// that artificial final newline, preserving genuine final blank lines.
			c.WriteClipboard(strings.TrimSuffix(read(), "\n"))
			state.at = c.Now()
			c.After(2 * time.Second)
		}
	})
}
