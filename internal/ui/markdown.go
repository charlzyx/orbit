package orbitui

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"github.com/egoist/mygo/ui/markdown/parser"
)

// GitHub-style alerts are Orbit syntax, not a default MyGo extension.
func (a *App) markdownBlock(c *ui.Context, d *parser.Document, i int) bool {
	if a.markdownRaggedTable(c, d, i) {
		return true
	}
	if a.markdownImageBlock(c, d, i) {
		return true
	}
	b := d.Block(i)
	if b.Kind != parser.Quote {
		return false
	}
	raw := d.Source()[b.Start:b.End]
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, ">")
		lines[i] = strings.TrimPrefix(line, " ")
	}
	if len(lines) == 0 {
		return false
	}
	kind := strings.TrimSpace(lines[0])
	colors := map[string]ui.Color{"[!NOTE]": a.Theme.Info, "[!TIP]": a.Theme.Success, "[!IMPORTANT]": a.Theme.Accent, "[!WARNING]": a.Theme.Warning, "[!CAUTION]": a.Theme.Error}
	color, ok := colors[kind]
	if !ok {
		return false
	}
	title := titleCase(strings.ToLower(strings.TrimSuffix(strings.TrimPrefix(kind, "[!"), "]")))
	ui.Column(c).Padding(12, 16).Gap(6).BorderWidth(0, 0, 0, 3).BorderColor(color).Background(color.Alpha(.06)).Radius(4).Children(func() {
		ui.Text(c, title).TextColor(color).FontWeight(600)
		opts := markdown.Options{ShareSelection: true, Theme: a.md.Theme, OnLink: a.md.OnLink, LoadImage: a.md.LoadImage, WrapCode: a.md.WrapCode, Block: a.markdownBlock}
		markdown.View(c, strings.Join(lines[1:], "\n"), &opts)
	})
	return true
}
