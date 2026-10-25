package orbitui

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// The list remains vertically virtualized. The outer viewport owns only the
// horizontal axis, matching sidepane.rs:1413–1483. Keeping the same hierarchy
// while toggling wrap also keeps the list's stable row identities and anchor.
func (a *App) reviewDiff(c *ui.Context) {
	s := &a.reviewState
	ui.Column(c.Key("review-diff-viewport")).Bind(&s.diffViewport).Grow(1).MinHeight(0).MinWidth(0).Clip().Children(func() {
		width := float32(0)
		var viewport ui.Element
		if !s.wrap {
			width = a.reviewUnwrappedWidth(c)
			viewport = ui.ScrollHorizontal(c.Key("review-horizontal")).TrackScroll(&s.horizontal)
		} else {
			s.horizontal = ui.ScrollState{}
			viewport = ui.Row(c.Key("review-horizontal")).Clip()
		}
		viewport.Label("Review horizontal viewport").Grow(1).MinHeight(0).MinWidth(0).Children(func() { a.reviewDiffRows(c, width) })
		if s.vertical.Y != s.stickyPanY {
			s.stickyPanY = s.vertical.Y
			c.AnimationFrame()
		}
	})
}

// Cache the whole snapshot, including hidden context, rather than deriving
// width from the currently mounted rows (which would jump while scrolling).
// Native text measurement accounts for the actual code font and wide glyphs.
type reviewWidthCache struct {
	first       *workspace.ReviewLine
	count       int
	generation  uint64
	font        string
	size, width float32
}

func (a *App) reviewUnwrappedWidth(c *ui.Context) float32 {
	s := &a.reviewState
	var first *workspace.ReviewLine
	if len(s.data.Lines) > 0 {
		first = &s.data.Lines[0]
	}
	cache := &s.widthCache
	if cache.first != first || cache.count != len(s.data.Lines) || cache.generation != s.generation || cache.font != a.Fonts.CodeFamily || cache.size != a.reviewCodeFontSize() {
		*cache = reviewWidthCache{first: first, count: len(s.data.Lines), generation: s.generation, font: a.Fonts.CodeFamily, size: a.reviewCodeFontSize()}
		measure := func(line workspace.ReviewLine) {
			size := a.Theme.CodeSize
			if line.Kind == "context" || line.Kind == "addition" || line.Kind == "deletion" {
				size = a.reviewCodeFontSize()
			}
			w, _ := c.MeasureText(0, ui.Span{Text: line.Content, Font: a.Fonts.CodeFamily, Size: size})
			cache.width = max(cache.width, w)
		}
		for _, line := range s.data.Lines {
			measure(line)
			if line.Gap != nil {
				for _, hidden := range line.Gap.Hidden {
					measure(hidden)
				}
			}
		}
	}
	// Match actual code-row chrome, retaining room for the unified change edge.
	width := cache.width + reviewCodeGutter + a.Theme.Space(12) + 10 + 2
	if s.split {
		width = (cache.width+reviewCodeGutter+2*a.Theme.Space(12))*2 + 1
	}
	return width
}

// Counter-shift only the inner header, not the virtual-list item root (the
// list owns that root's position). Source sidepane.rs:1455–1469 pinned_row.
func (a *App) reviewPinnedRow(c *ui.Context, draw func()) {
	s := &a.reviewState
	viewport := s.horizontal.Bounds(c)
	if s.wrap || viewport.W <= 0 {
		draw()
		return
	}
	ui.Column(c).FillWidth().Children(func() {
		ui.Column(c).Width(viewport.W).Shrink(0).Translate(s.horizontal.X, 0).Children(draw)
	})
}
