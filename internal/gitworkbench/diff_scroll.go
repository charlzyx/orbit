package gitworkbench

import (
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type diffViewportState struct {
	dataset              diffDatasetKey
	headers              map[diffHeaderKey]*ui.Handle
	headerHeights        map[diffHeaderKey]float32
	headerTheme          ui.Theme
	lastY                float32
	horizontal, vertical ui.ScrollState
	viewport             ui.Handle
	width                diffWidthCache
}
type diffWidthCache struct {
	staged, unstaged           *workspace.ReviewLine
	stagedCount, unstagedCount int
	generation                 uint64
	width                      float32
	metrics                    diffFontMetrics
}

type diffFontMetrics struct {
	font                string
	size, spacing, cell float32
}

func firstReviewLine(lines []workspace.ReviewLine) *workspace.ReviewLine {
	if len(lines) == 0 {
		return nil
	}
	return &lines[0]
}
func (s *State) unwrappedDiffWidth(c *ui.Context) float32 {
	a, b := s.Data.StagedReview.Lines, s.Data.UnstagedReview.Lines
	cache := &s.diffView.width
	cell, _ := c.MeasureText(0, ui.Span{Text: "M", Font: "monospace", Size: 12})
	metrics := diffFontMetrics{c.Theme().Font, c.Theme().FontSize, c.Theme().Spacing, cell}
	if cache.staged != firstReviewLine(a) || cache.unstaged != firstReviewLine(b) || cache.stagedCount != len(a) || cache.unstagedCount != len(b) || cache.generation != s.generation || cache.metrics != metrics {
		*cache = diffWidthCache{staged: firstReviewLine(a), unstaged: firstReviewLine(b), stagedCount: len(a), unstagedCount: len(b), generation: s.generation, metrics: metrics}
		measure := func(line workspace.ReviewLine) {
			w, _ := c.MeasureText(0, ui.Span{Text: line.Content, Font: "monospace", Size: 12})
			cache.width = max(cache.width, w)
		}
		for _, lines := range [][]workspace.ReviewLine{a, b} {
			for _, line := range lines {
				measure(line)
				if line.Gap != nil {
					for _, hidden := range line.Gap.Hidden {
						measure(hidden)
					}
				}
			}
		}
	}
	// The existing embedded renderer has two 35px number columns, a 10px sign,
	// three6px gaps and two10px paddings. Preserve its selectable rich text.
	width := cache.width + 118
	if s.DiffSplit {
		width *= 2
	}
	return width
}
func (s *State) changesDiffViewport(c *ui.Context, items []changeItem) {
	d := &s.diffView
	s.prepareDiffHeaders(items)
	ui.Column(c.Key("changes-viewport")).Bind(&d.viewport).Grow(1).MinHeight(0).MinWidth(0).Clip().Children(func() {
		width := float32(0)
		var viewport ui.Element
		if s.DiffWrap {
			d.horizontal = ui.ScrollState{}
			viewport = ui.Row(c.Key("changes-horizontal")).Clip()
		} else {
			width = s.unwrappedDiffWidth(c)
			viewport = ui.ScrollHorizontal(c.Key("changes-horizontal")).TrackScroll(&d.horizontal)
		}
		viewport.Label("Git diff horizontal viewport").Grow(1).MinHeight(0).MinWidth(0).Children(func() {
			right := float32(0)
			bounds := d.viewport.Bounds(c)
			if !s.DiffWrap && bounds.W > 0 {
				right = max(0, max(width, bounds.W)-bounds.W-d.horizontal.X)
			}
			ui.List(c.Key("changes-list"), &s.list, len(items), func(i int) {
				item := items[i]
				if !s.DiffWrap && item.line == nil && item.split == nil && bounds.W > 0 {
					ui.Column(c).FillWidth().Children(func() {
						ui.Column(c).Width(bounds.W).Translate(d.horizontal.X, 0).Children(func() { s.changeItem(c, item, false) })
					})
				} else {
					s.changeItem(c, item, false)
				}
			}).WidthPercent(100).MinWidth(width).FillHeight().Shrink(0).ScrollbarInsets(0, right, 0, 0).TrackScroll(&d.vertical).Children(func() { s.stickyChangesHeaders(c, items) })
		})
		if d.vertical.Y != d.lastY {
			d.lastY = d.vertical.Y
			c.AnimationFrame()
		}
	})
}

// Header handles are view-local, not a history of every path ever visited.
// Prune with the current rendered dataset (also handles filter/collapse) and
// invalidate measured heights when refreshed snapshots replace their rows.
func (s *State) prepareDiffHeaders(items []changeItem) {
	d := &s.diffView
	a, b := s.Data.StagedReview.Lines, s.Data.UnstagedReview.Lines
	key := diffDatasetKey{firstReviewLine(a), firstReviewLine(b), len(a), len(b), s.generation}
	if d.dataset != key {
		d.dataset = key
		d.headers = nil
		d.headerHeights = nil
	}
	live := make(map[diffHeaderKey]bool)
	for _, item := range items {
		if isChangeSection(item) || isChangeFile(item) {
			live[changeHeaderKey(item)] = true
		}
	}
	for key := range d.headers {
		if !live[key] {
			delete(d.headers, key)
			delete(d.headerHeights, key)
		}
	}
}

type diffDatasetKey struct {
	staged, unstaged           *workspace.ReviewLine
	stagedCount, unstagedCount int
	generation                 uint64
}
