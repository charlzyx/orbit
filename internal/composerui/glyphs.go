package composerui

import (
	"path"
	"sort"
	"strings"

	"github.com/egoist/mygo/ui"
)

type fileGlyph struct {
	Glyph       string
	Dark, Light uint32
}

func glyphForFile(filename string) fileGlyph {
	name := path.Base(filename)
	if glyph, ok := devicons[name]; ok {
		return glyph
	}
	if strings.HasSuffix(filename, "/") {
		return fileGlyph{"\uf115", 0x7e8ea8, 0x7e8ea8}
	}
	ext := strings.TrimPrefix(path.Ext(name), ".")
	if glyph, ok := devicons[ext]; ok {
		return glyph
	}
	if glyph, ok := devicons[strings.ToLower(ext)]; ok {
		return glyph
	}
	return fileGlyph{"*", 0x7e8ea8, 0x7e8ea8}
}

var preferredNerdFamilies = []string{"SymbolsNerdFont", "Symbols Nerd Font", "JetBrainsMono Nerd Font", "Hack Nerd Font", "FiraCode Nerd Font", "CaskaydiaCove Nerd Font"}

func chooseNerdFamily(families []string) string {
	candidates := []string{}
	for _, name := range families {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "nerd font") || strings.Contains(lower, "nerdfont") || strings.HasSuffix(lower, " nf") {
			candidates = append(candidates, name)
		}
	}
	rank := func(name string) int {
		for i, v := range preferredNerdFamilies {
			if strings.EqualFold(name, v) {
				return i
			}
		}
		return len(preferredNerdFamilies)
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		ra, rb := rank(a), rank(b)
		if ra != rb {
			return ra < rb
		}
		as, bs := strings.HasPrefix(strings.ToLower(a), "symbols"), strings.HasPrefix(strings.ToLower(b), "symbols")
		if as != bs {
			return as
		}
		return a < b
	})
	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}
func (c *Controller) nerdFamily() string {
	if !c.fontsChecked {
		names := []string{}
		for _, f := range ui.FontFamilies() {
			names = append(names, f.Name)
		}
		c.nerdFont = chooseNerdFamily(names)
		c.fontsChecked = true
	}
	return c.nerdFont
}
func (c *Controller) fileGlyph(ctx *ui.Context, entry Entry, style Style) bool {
	if entry.Kind != At {
		return false
	}
	family := c.nerdFamily()
	if family == "" {
		return false
	}
	glyph := glyphForFile(entry.Path)
	color := glyph.Light
	if style.Dark {
		color = glyph.Dark
	}
	ui.Row(ctx).Width(style.px(14) + 4).Shrink(0).Center().Children(func() {
		ui.Text(ctx, glyph.Glyph).Font(family).FontSize(style.px(14)).TextColor(ui.RGB(uint8(color>>16), uint8(color>>8), uint8(color))).MaxLines(1)
	})
	return true
}
