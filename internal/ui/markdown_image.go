package orbitui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charlzyx/orbit/internal/markdownimage"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
	"golang.org/x/net/html"
)

type markdownImage struct {
	url, alt, title, raw string
	pending              bool
}

func remoteMarkdownImage(raw string) bool {
	lower := strings.ToLower(raw)
	return strings.HasPrefix(lower, "http:") || strings.HasPrefix(lower, "https:") || strings.HasPrefix(lower, "data:")
}
func markdownImageInputs(d *parser.Document, i int) ([]markdownImage, bool) {
	b := d.Block(i)
	if b.Kind == parser.Paragraph {
		in := d.Inlines(i)
		var images []markdownImage
		remote := false
		for k := 0; k < len(in); k = in[k].Next {
			n := in[k]
			switch n.Kind {
			case parser.Image:
				images = append(images, markdownImage{url: n.Dest, alt: parser.PlainText(in[k+1 : n.Next]), title: n.Title, raw: d.Source()[n.Start:n.End], pending: n.Pending})
				remote = remote || remoteMarkdownImage(n.Dest)
			case parser.SoftBreak, parser.LineBreak:
			case parser.Text:
				if strings.TrimSpace(n.Text) != "" {
					return nil, false
				}
			default:
				return nil, false
			}
		}
		return images, len(images) > 0 && remote
	}
	if b.Kind != parser.HTMLBlock {
		return nil, false
	}
	// Only a standalone img tag: no HTML container/script/style execution or
	// reparsing of arbitrary prose. Ignore non-source attributes such as onerror.
	raw := strings.TrimSpace(d.Source()[b.Start:b.End])
	z := html.NewTokenizer(strings.NewReader(raw))
	var found *markdownImage
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() != io.EOF || found == nil {
				return nil, false
			}
			return []markdownImage{*found}, remoteMarkdownImage(found.url)
		case html.TextToken:
			if strings.TrimSpace(string(z.Text())) != "" {
				return nil, false
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			token := z.Token()
			if token.Data != "img" || found != nil {
				return nil, false
			}
			found = &markdownImage{}
			for _, a := range token.Attr {
				switch a.Key {
				case "src":
					found.url = a.Val
				case "alt":
					found.alt = a.Val
				case "title":
					found.title = a.Val
				}
			}
		case html.EndTagToken:
			if z.Token().Data != "img" {
				return nil, false
			}
		default:
			return nil, false
		}
	}
}
func (a *App) markdownImageBlock(c *ui.Context, d *parser.Document, i int) bool {
	images, ok := markdownImageInputs(d, i)
	if !ok {
		return false
	}
	t := a.Theme
	ui.Row(c).Wrap().Gap(8).AlignSelf(ui.Stretch).MinWidth(0).Children(func() {
		for index, input := range images {
			if !remoteMarkdownImage(input.url) {
				// Preserve the existing local-file path and its toolkit cache unchanged,
				// even when one paragraph mixes local and remote image elements.
				markdown.View(c.Key(fmt.Sprintf("local-markdown-image-%d", index)), input.raw, &a.md)
				continue
			}
			services := c.Services()
			result := a.markdownImages.Request(input.url, services.Invalidate)
			alt := input.alt
			if strings.TrimSpace(alt) == "" {
				alt = a.text("markdown.image_unavailable", "Image unavailable")
			}
			if result.Bitmap == nil || input.pending {
				ui.Row(c.Key(fmt.Sprintf("markdown-image-placeholder-%d", index))).Padding(8).MinWidth(0).MaxWidthPercent(100).Radius(12).Border(1, t.Border).Background(t.Surface).Role(ui.RoleImage).Label(input.alt).Children(func() { ui.Text(c, alt).TextColor(t.TextMuted).FontSize(t.Px(12)) })
				continue
			}
			width, height := result.Bitmap.Size()
			button := ui.ButtonBase(c.Key(fmt.Sprintf("markdown-image-%d-%d-%x", index, result.Owner, result.Key))).Label("Open Markdown image "+input.alt).MinWidth(0).MaxWidthPercent(100).Padding(0).Border(1, t.Border).Radius(12).Clip().Children(func() {
				ui.Image(c, result.Bitmap).MinWidth(0).MaxWidthPercent(100).AspectRatio(float32(width) / float32(height)).Label(input.alt)
			})
			if input.title != "" {
				button.Tooltip(input.title)
			}
			if button.Clicked() {
				// Data rasters have no external navigation target; show the decoded safe
				// image in the existing native preview instead of launching a data URL.
				if strings.HasPrefix(strings.ToLower(input.url), "data:") {
					a.inlineImages.preview = inlineImageContentToken{}
					a.imagePreview = result.Bitmap
					a.imageOpen = true
				} else {
					a.Dispatch(Action{Kind: "OpenLink", Value: markdownimage.Normalize(input.url)})
				}
			}
		}
	})
	return true
}

// Reset at both snapshot and service ownership boundaries. Services update
// Model.Workspace before the next snapshot, including A→B→A without a frame.
func (a *App) resetMarkdownImageOwner() {
	a.resetInlineImages()
	a.markdownImages.Reset()
	a.imageOpen = false
	a.imagePreview = nil
}
