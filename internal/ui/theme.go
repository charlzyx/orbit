package orbitui

import (
	"fmt"
	"io/fs"
	"strings"
	"sync"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/preferences"
	ghostty "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/egoist/mygo/ui/markdown"
)

const HeaderHeight float32 = 44
const ContentMaxWidth float32 = 960
const SidebarMinWidth float32 = 200
const SettingsNavWidth float32 = 240

type Theme struct {
	Palette
	ID                 string
	Dark               bool
	FontSize, CodeSize float32
	Density            int
	Font, CodeFont     string
}

func NewTheme(id string, prefs preferences.UI, fonts preferences.Fonts) Theme {
	p, ok := Palettes[id]
	if !ok {
		id = "orbit-azure"
		p = Palettes[id]
	}
	return Theme{Palette: p, ID: id, Dark: !strings.Contains(id, "light") && id != "orbit-paper" && id != "orbit-contrast", FontSize: float32(prefs.UIFontSize), CodeSize: float32(prefs.EditorFontSize), Density: prefs.SpacingDensity, Font: fonts.UIFamily, CodeFont: fonts.CodeFamily}
}
func (t Theme) Px(n float32) float32 { return n * t.FontSize / 14 }
func (t Theme) Space(n int) float32 {
	table := map[int][3]float32{0: {0, 0, 0}, 1: {1, 1, 2}, 2: {1, 2, 4}, 3: {2, 3, 4}, 4: {2, 4, 6}, 6: {3, 6, 8}, 8: {4, 8, 10}, 12: {10, 12, 14}, 16: {14, 16, 18}, 20: {18, 20, 22}, 24: {20, 24, 28}, 32: {28, 32, 36}, 40: {36, 40, 44}, 48: {44, 48, 52}}
	v, ok := table[n]
	if !ok {
		return t.Px(float32(n))
	}
	d := max(float32(-1), min(float32(1), float32(t.Density-100)/20))
	x := v[1]
	if d < 0 {
		x += (v[1] - v[0]) * d
	} else {
		x += (v[2] - v[1]) * d
	}
	return t.Px(x)
}
func (t Theme) Native() *ui.Theme {
	return &ui.Theme{Dark: t.Dark, Background: t.Background, Surface: t.Surface, SurfaceHover: t.Hover, SurfacePressed: t.Active, Border: t.Border, Text: t.Text, TextMuted: t.TextMuted, Accent: t.Accent, AccentHover: t.AccentHover, AccentPressed: t.AccentPressed, AccentText: t.AccentFg, Danger: t.Error, Warning: t.Warning, Success: t.Success, Selection: t.Selection, Focus: t.FocusRing, Radius: t.Px(6), Spacing: t.Space(4), FontSize: t.FontSize, Font: t.Font, Scrollbar: t.TextMuted.Alpha(.45), ScrollbarWidth: 6}
}
func (t Theme) Markdown() *markdown.Theme {
	m := markdown.DefaultTheme(t.Native())
	m.Font = t.Font
	m.CodeFont = t.CodeFont
	m.FontSize = t.FontSize
	m.Text = t.AssistantText
	m.Muted = t.TextMuted
	m.Link = t.AccentText
	m.Code = t.AccentText
	m.CodeBackground = t.Accent.Alpha(.1)
	m.CodeBlockBackground = t.CodeBg
	m.CodeBlockBorder = t.Border
	m.CodeBlockRadius = t.Px(8)
	m.LineHeight = 1.55
	m.CodeLineHeight = 1.4
	return m
}
func (t Theme) Code() *editor.Theme {
	e := editor.DefaultTheme(t.Native())
	e.Background = t.Background
	e.Text = t.CodeText
	e.LineNumber = t.TextMuted
	e.Added = t.Success
	e.Deleted = t.Error
	e.Modified = t.Warning
	colors := map[editor.TokenKind]ui.Color{editor.TokenKeyword: t.AccentText, editor.TokenString: t.SynString, editor.TokenNumber: t.SynNumber, editor.TokenFunction: t.SynFunction, editor.TokenType: t.SynType, editor.TokenComment: t.SynComment, editor.TokenConstant: t.SynLiteral, editor.TokenMeta: t.SynMeta, editor.TokenOperator: t.SynOperator}
	for k, c := range colors {
		e.Tokens[k] = editor.TokenStyle{Color: c}
	}
	return e
}

var assetsOnce sync.Once
var assetsError error
var icons map[string]*ui.SVG
var logos map[string]*ui.Bitmap

func LoadAssets() error {
	assetsOnce.Do(func() {
		icons = map[string]*ui.SVG{}
		logos = map[string]*ui.Bitmap{}
		assetsError = fs.WalkDir(orbitassets.Assets(), ".", func(p string, d fs.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.IsDir() {
				return nil
			}
			switch {
			case strings.HasSuffix(p, ".svg"):
				b, e := orbitassets.Asset(p)
				if e != nil {
					return e
				}
				v, e := ui.ParseSVG(b)
				if e != nil {
					return fmt.Errorf("icon %s: %w", p, e)
				}
				icons[p] = v
			case strings.HasSuffix(p, ".ttf") || strings.HasSuffix(p, ".otf"):
				b, e := orbitassets.Asset(p)
				if e != nil {
					return e
				}
				family := ""
				if strings.Contains(strings.ToLower(p), "plexsans") {
					family = ".ZedSans"
				}
				if strings.Contains(strings.ToLower(p), "lilex") {
					family = ".ZedMono"
				}
				if e = ui.RegisterFont(b, family); e != nil {
					return fmt.Errorf("font %s: %w", p, e)
				}
			case p == "logo.png" || p == "logo-dark.png" || p == "app-icon.png":
				b, e := orbitassets.Asset(p)
				if e != nil {
					return e
				}
				v, e := ui.DecodeBitmap(b)
				if e != nil {
					return e
				}
				logos[p] = v
			}
			return nil
		})
	})
	return assetsError
}

// Terminal maps the same sixteen semantic ANSI slots as the pinned Rust view.
func (t Theme) Terminal() *ghostty.Theme {
	return &ghostty.Theme{Foreground: t.CodeText, Background: t.CodeBg, Cursor: t.AccentText, Selection: t.Selection, SelectionText: t.SelectionFg, Palette: [16]ui.Color{t.TextMuted, t.Error, t.Success, t.SynNumber, t.SynFunction, t.SynString, t.SynType, t.Text, t.TextMuted, t.Error, t.Success, t.SynLiteral, t.SynFunction, t.SynString, t.SynType, t.Text}}
}
