package composerui

import (
	"fmt"
	"math"
	"time"

	"github.com/egoist/mygo/ui"
)

// Style is supplied by Orbit's active semantic palette. Text localizes the
// source's five scope labels; Icon uses the application's bundled native SVG.
type Style struct {
	FontSize                                                         float32
	Density                                                          int
	Font                                                             string
	Dark                                                             bool
	Surface, Border, Selection, Hover, Text, Secondary, Muted, Badge ui.Color
	CommandColor, FileColor                                          ui.Color
	TextFor                                                          func(string, string) string
	Icon                                                             func(*ui.Context, string, float32, ui.Color)
}

func (s Style) px(n float32) float32 {
	size := s.FontSize
	if size <= 0 {
		size = 14
	}
	return n * size / 14
}
func (s Style) space(n int) float32 {
	values := map[int][3]float32{4: {2, 4, 6}, 6: {3, 6, 8}, 8: {4, 8, 10}}
	v := values[n]
	density := s.Density
	if density == 0 {
		density = 100
	}
	d := max(float32(-1), min(float32(1), float32(density-100)/20))
	x := v[1]
	if d < 0 {
		x += (v[1] - v[0]) * d
	} else {
		x += (v[2] - v[1]) * d
	}
	return s.px(x)
}
func (s Style) RowHeight() float32 { return float32(math.Round(float64(s.px(14)*1.618))) + s.px(8) }
func (s Style) label(scope Scope, project string) string {
	if scope == Project && project != "" {
		return project
	}
	key := "autocomplete.scope_" + string(scope)
	if s.TextFor != nil {
		return s.TextFor(key, string(scope))
	}
	return string(scope)
}

// SyncInput is called while building, after creating the actual main TextArea.
// It never queries an Element from between-frame keyboard callbacks.
// Returns true when the bound draft was changed and a fresh frame is needed.
func (c *Controller) SyncInput(input ui.Element, draft *string) bool {
	if draft == nil || !input.Valid() {
		return false
	}
	if c.restoreCaret {
		input.SetTextSelection(c.pendingCaret, c.pendingCaret)
		c.restoreCaret = false
	}
	_, caret := input.TextAnchorCaret()
	c.Sync(*draft, caret, input.Focused(), input.Composing())
	pending := c.pending
	c.pending = nil
	for _, p := range pending {
		// A key's text must still equal the native text; newer input cancels it.
		if p.text != *draft || !input.Focused() || input.Composing() {
			continue
		}
		if p.delta != 0 {
			c.Move(p.delta)
		}
		if p.accept {
			if text, at, ok := c.Commit(c.state.Highlighted); ok {
				*draft = text
				c.pendingCaret = at
				c.restoreCaret = true
				return true
			}
		}
	}
	return false
}

// HandleKey is called before the composer's submit branches. It consumes only
// completion-owned keys and never keeps the short-lived TextKey or Element.
type keyRangeEditor interface {
	AnchorCaret() (int, int)
	Composing() bool
	ReplaceRange(int, int, string) bool
}
type elementRangeEditor interface{ ReplaceTextRange(int, int, string) bool }

// NativeRangeEditsAvailable reports whether the reviewed native compatibility
// addition is linked. The fallback is safe, but does not promise undo parity.
func NativeRangeEditsAvailable() bool { _, ok := any(ui.TextKey{}).(keyRangeEditor); return ok }

func (c *Controller) HandleKey(k ui.TextKey) bool {
	if editor, ok := any(k).(keyRangeEditor); ok {
		_, caret := editor.AnchorCaret()
		c.Sync(k.Text(), caret, true, editor.Composing())
		if c.composing && k.Key == ui.KeyEnter {
			return true
		}
		if !c.state.Open {
			return false
		}
		switch {
		case k.Mods == 0 && k.Key == ui.KeyUp:
			return c.Move(-1)
		case k.Mods == 0 && k.Key == ui.KeyDown:
			return c.Move(1)
		case k.Mods == 0 && k.Key == ui.KeyEscape:
			return c.Dismiss()
		case k.Mods == 0 && k.Key == ui.KeyTab, k.Key == ui.KeyEnter && (k.Mods == 0 || k.Mods == ui.Alt || k.Mods == (ui.Cmd|ui.Shift)):
			trigger := c.state.Trigger
			entry := c.state.Entries[c.state.Highlighted]
			if editor.ReplaceRange(trigger.Start, trigger.End, entry.Insertion()) {
				c.Dismiss()
				c.lastKey = ""
				return true
			}
		}
		return false
	}
	if c.composing && k.Key == ui.KeyEnter {
		return true
	}
	if !c.state.Open || c.composing || !c.focused {
		return false
	}
	switch {
	case k.Mods == 0 && k.Key == ui.KeyUp:
		c.pending = append(c.pending, pendingKey{text: k.Text(), delta: -1})
		return true
	case k.Mods == 0 && k.Key == ui.KeyDown:
		c.pending = append(c.pending, pendingKey{text: k.Text(), delta: 1})
		return true
	case k.Mods == 0 && k.Key == ui.KeyEscape:
		c.Dismiss()
		return true
	case k.Mods == 0 && k.Key == ui.KeyTab, k.Key == ui.KeyEnter && (k.Mods == 0 || k.Mods == ui.Alt || k.Mods == (ui.Cmd|ui.Shift)):
		c.pending = append(c.pending, pendingKey{text: k.Text(), accept: true})
		return true
	}
	return false
}

// CommitInput is for the send button. It preserves the menu-before-send rule.
func (c *Controller) CommitInput(input ui.Element, draft *string) bool {
	if draft == nil || !input.Valid() {
		return false
	}
	_, caret := input.TextAnchorCaret()
	c.Sync(*draft, caret, input.Focused(), input.Composing())
	return c.commitNative(input, draft, c.state.Highlighted)
}
func (c *Controller) commitNative(input ui.Element, draft *string, index int) bool {
	if editor, ok := any(input).(elementRangeEditor); ok {
		if !c.state.Open || c.composing || index < 0 || index >= len(c.state.Entries) {
			return false
		}
		trigger := c.state.Trigger
		if editor.ReplaceTextRange(trigger.Start, trigger.End, c.state.Entries[index].Insertion()) {
			c.Dismiss()
			c.lastKey = ""
			return true
		}
		return false
	}
	text, at, ok := c.Commit(index)
	if !ok {
		return false
	}
	*draft = text
	c.pendingCaret = at
	c.restoreCaret = true
	return true
}

// PaintMentions applies source-compatible, paint-only native text ranges.
func PaintMentions(input ui.Element, text string, style Style) {
	ranges := []ui.TextRange{}
	for _, span := range TokenizeMentions(text) {
		color := style.CommandColor
		if span.Kind == At {
			color = style.FileColor
		}
		ranges = append(ranges, ui.TextRange{Start: span.Start, End: span.End, Color: color})
	}
	input.TextRanges(ranges...)
}

// Popup is mounted after the composer is built; it never changes layout below
// it. Loading requests another modestly timed frame, without a fake candidate.
func (c *Controller) Popup(ctx *ui.Context, anchor, input ui.Element, draft *string, style Style) {
	if c.state.Files.State == Loading {
		ctx.After(30 * time.Millisecond)
	}
	if c.restoreCaret {
		ctx.AnimationFrame()
	}
	if !c.state.Open || len(c.state.Entries) == 0 {
		return
	}
	width, height := ctx.Size()
	bounds := anchor.Bounds()
	anchorKey := [4]float32{bounds.X, bounds.Y, bounds.W, bounds.H}
	if anchorKey != c.popupAnchor {
		c.popupAnchor = anchorKey
		ctx.AnimationFrame()
	}
	placement := popupBounds(bounds, width, height, len(c.state.Entries), style)
	open := true
	var panel ui.Element
	// Position a zero-size native anchor inside the source's 8px inset. MyGo
	// then has no reason to apply its own 4px edge snap or side flip. The
	// deferred pass observes any newly resized composer bounds before settling.
	ui.Overlay(ctx, func() {
		positioned := ui.Box(ctx.Key("completion-source-anchor")).Absolute().Left(placement.X).Top(placement.Y+placement.H).Size(0, 0)
		panel = ui.PopoverBase(ctx.Key("composer-autocomplete"), positioned, &open, func(panel ui.Element) {
			panel.AttachTo(positioned, ui.AnchorTopLeft, ui.AnchorBottomLeft).Width(placement.W).Column().Padding(style.space(4), 0).Radius(style.px(12)).Border(1, style.Border).Background(style.Surface).Font(style.Font).FontSize(style.px(14)).Clip().Label("Composer completions")
			alpha1, alpha2, alpha4 := float32(.12), float32(.08), float32(.12)
			if !style.Dark {
				alpha1, alpha2, alpha4 = .06, .06, .04
			}
			panel.Shadow(0, 2, 3, 0, ui.RGBA(0, 0, 0, alpha1)).Shadow(0, 3, 6, 0, ui.RGBA(0, 0, 0, alpha2)).Shadow(0, 6, 12, 0, ui.RGBA(0, 0, 0, .04)).Shadow(0, 1, 0, 0, ui.RGBA(0, 0, 0, alpha4))
			for i, entry := range c.state.Entries {
				row := ui.Row(ctx.Key(fmt.Sprintf("completion-%d", i))).Label(entry.Title()).AlignSelf(ui.Stretch).AlignItems(ui.Center).Height(style.RowHeight()).Shrink(0).MarginX(style.space(4)).Padding(style.px(4), style.space(6)).Gap(style.space(6)).Radius(style.px(6)).Cursor(ui.CursorPointer)
				clicked := row.Clicked()
				selected := i == c.state.Highlighted
				if selected {
					row.Background(style.Selection)
				}
				if row.Hovered() {
					row.Background(style.Hover)
				}
				row.Children(func() {
					if entry.Kind == Slash && style.Icon != nil {
						style.Icon(ctx, "icons/extensions.svg", style.px(14), style.Muted)
					} else if !c.fileGlyph(ctx, entry, style) {
						b := FileBadge(entry.Path)
						color := b.Light
						if style.Dark {
							color = b.Dark
						}
						if entry.Kind == Slash {
							b.Label = "/"
							color = uint32(style.Muted.R)<<16 | uint32(style.Muted.G)<<8 | uint32(style.Muted.B)
						}
						ink := ui.RGB(uint8(color>>16), uint8(color>>8), uint8(color))
						ui.Row(ctx).Size(17, 17).Shrink(0).Radius(style.px(4)).Background(ink.Alpha(.16)).Center().Children(func() { ui.Text(ctx, b.Label).FontSize(style.px(7.5)).FontWeight(600).TextColor(ink).MaxLines(1) })
					}
					ui.Row(ctx).Grow(1).MinWidth(0).AlignItems(ui.Center).Gap(style.space(8)).Children(func() {
						color := style.Secondary
						if selected {
							color = style.Text
						}
						ui.Text(ctx, entry.Title()).FontWeight(500).TextColor(color).MaxWidth(480).Shrink(1).MaxLines(1)
						if sub := entry.Subtitle(); sub != "" {
							ui.Text(ctx, sub).FontSize(style.px(12)).TextColor(style.Muted).Grow(1).MinWidth(0).MaxLines(1)
						}
					})
					if entry.Kind == Slash {
						ui.Row(ctx).Height(18).Padding(0, style.space(6)).Radius(style.px(6)).Background(style.Badge).AlignItems(ui.Center).Children(func() {
							ui.Text(ctx, style.label(entry.Scope, entry.Project)).FontSize(style.px(10)).TextColor(style.Muted).MaxLines(1)
						})
					}
				})
				if clicked && draft != nil {
					c.focused = true
					if c.commitNative(input, draft, i) {
						input.Focus()
						ctx.AnimationFrame()
					}
				}
			}
		})
	})
	if !open || panel.Dismissed() {
		c.Dismiss()
	}
}

func popupBounds(anchor ui.Rect, width, height float32, count int, style Style) ui.Rect {
	w := min(float32(960), max(float32(0), width-16))
	h := float32(min(count, Limit))*style.RowHeight() + 2*style.space(4) + 2
	x := max(float32(8), min(anchor.X, width-8-w))
	y := max(float32(8), min(anchor.Y-5-h, height-8-h))
	return ui.Rect{X: x, Y: y, W: w, H: h}
}

// RequireCompatibility prevents shipping the intentionally limited unpatched
// fallback. Use the repository's hash-checked wrapper for every native binary.
func RequireCompatibility() error {
	if !NativeRangeEditsAvailable() {
		return fmt.Errorf("Orbit requires its pinned MyGo composer edit compatibility patch; use scripts/go-native.sh run ./cmd/orbit or scripts/go-native.sh build ./cmd/orbit")
	}
	return nil
}

// MentionFileColor mirrors theme.rs: readable accent complement, with a 70%
// neutral fallback for saturation below .15. Conversion is paint-only.
func MentionFileColor(color ui.Color) ui.Color {
	r, g, b := float64(color.R)/255, float64(color.G)/255, float64(color.B)/255
	high, low := max(r, g, b), min(r, g, b)
	light := (high + low) / 2
	delta := high - low
	sat, hue := float64(0), float64(0)
	if delta > 0 {
		sat = delta / (1 - math.Abs(2*light-1))
		switch high {
		case r:
			hue = math.Mod((g-b)/delta, 6)
		case g:
			hue = (b-r)/delta + 2
		default:
			hue = (r-g)/delta + 4
		}
		hue /= 6
		if hue < 0 {
			hue++
		}
	}
	if sat < .15 {
		return color.Alpha(float32(color.A) / 255 * .7)
	}
	hue = math.Mod(hue+.5, 1)
	sat = max(sat, .5)
	chroma := (1 - math.Abs(2*light-1)) * sat
	x := chroma * (1 - math.Abs(math.Mod(hue*6, 2)-1))
	m := light - chroma/2
	switch {
	case hue < 1.0/6:
		r, g, b = chroma, x, 0
	case hue < 2.0/6:
		r, g, b = x, chroma, 0
	case hue < 3.0/6:
		r, g, b = 0, chroma, x
	case hue < 4.0/6:
		r, g, b = 0, x, chroma
	case hue < 5.0/6:
		r, g, b = x, 0, chroma
	default:
		r, g, b = chroma, 0, x
	}
	return ui.RGBA(uint8(math.Round((r+m)*255)), uint8(math.Round((g+m)*255)), uint8(math.Round((b+m)*255)), float32(color.A)/255)
}
