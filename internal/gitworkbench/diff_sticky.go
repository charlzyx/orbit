package gitworkbench

import (
	"fmt"
	"github.com/egoist/mygo/ui"
)

type diffHeaderKey struct {
	staged, section bool
	path            string
}

func changeHeaderKey(item changeItem) diffHeaderKey {
	return diffHeaderKey{staged: item.staged, section: item.header != "" && !item.folder, path: item.path}
}
func (s *State) bindChangeHeader(c *ui.Context, row ui.Element, item changeItem, sticky bool) {
	key := changeHeaderKey(item)
	label := fmt.Sprintf("file %t:%s", item.staged, item.path)
	if key.section {
		label = fmt.Sprintf("section %t", item.staged)
	}
	if sticky {
		row.Label("Git sticky " + label)
		return
	}
	row.Label("Git " + label)
	if s.diffView.headers == nil {
		s.diffView.headers = map[diffHeaderKey]*ui.Handle{}
	}
	if s.diffView.headers[key] == nil {
		s.diffView.headers[key] = &ui.Handle{}
	}
	row.Bind(s.diffView.headers[key])
}
func (s *State) changeHeaderBounds(c *ui.Context, item changeItem) ui.Rect {
	if h := s.diffView.headers[changeHeaderKey(item)]; h != nil {
		return h.Bounds(c)
	}
	return ui.Rect{}
}
func isChangeSection(item changeItem) bool { return item.header != "" && !item.folder }
func isChangeFile(item changeItem) bool {
	return item.row.Path != "" && item.line == nil && item.split == nil
}

// Rust sidepane.rs:1485–1574: section strip above current file header. The
// next section pushes the whole pair away; within a section the next file
// pushes just its file header. Native absolute children reuse live handlers,
// never change the virtual list's logical anchors, and do not pan with code.
func (s *State) stickyChangesHeaders(c *ui.Context, items []changeItem) {
	if s.diffView.headerTheme != *c.Theme() {
		s.diffView.headerTheme = *c.Theme()
		s.diffView.headerHeights = map[diffHeaderKey]float32{}
	}
	if s.diffView.headerHeights == nil {
		s.diffView.headerHeights = map[diffHeaderKey]float32{}
	}
	for key, handle := range s.diffView.headers {
		if r := handle.Bounds(c); r.H > 0 {
			s.diffView.headerHeights[key] = r.H
		}
	}
	first, _ := s.list.Visible()
	if first < 0 || first >= len(items) {
		return
	}
	viewport := s.diffView.viewport.Bounds(c)
	if viewport.H <= 0 {
		return
	}
	section, file := -1, -1
	for i := 0; i <= first; i++ {
		if isChangeSection(items[i]) {
			section, file = i, -1
		} else if isChangeFile(items[i]) {
			file = i
		}
	}
	if section < 0 {
		return
	}
	secBounds := s.changeHeaderBounds(c, items[section])
	sectionHeight := s.changeHeaderHeight(c, items[section])
	sectionNeeds := section < first || (secBounds.H > 0 && secBounds.Y < viewport.Y)
	sectionPush := float32(0)
	nextSection := len(items)
	for i := section + 1; i < len(items); i++ {
		if !isChangeSection(items[i]) {
			continue
		}
		nextSection = i
		r := s.changeHeaderBounds(c, items[i])
		if r.H > 0 {
			sectionPush = min(float32(0), r.Y-viewport.Y-sectionHeight)
		}
		break
	}
	// Draw file first so the section wins both painting and hit-testing where
	// they touch. Distinct keys prevent a held pointer activating a new file.
	if file >= 0 && sectionPush >= 0 {
		fileBounds := s.changeHeaderBounds(c, items[file])
		top := float32(0)
		if sectionNeeds {
			top = sectionHeight
		}
		needs := sectionNeeds || file < first || (fileBounds.H > 0 && fileBounds.Y < viewport.Y+top)
		if needs {
			height := s.changeHeaderHeight(c, items[file])
			push := float32(0)
			for i := file + 1; i <= nextSection && i < len(items); i++ {
				if !isChangeFile(items[i]) && !isChangeSection(items[i]) {
					continue
				}
				r := s.changeHeaderBounds(c, items[i])
				if r.H > 0 {
					push = min(float32(0), r.Y-viewport.Y-top-height)
				}
				break
			}
			item := items[file]
			ui.Column(c.Key(fmt.Sprintf("sticky-file-%t:%s", item.staged, item.path))).Absolute().Top(top + push).Left(s.diffView.horizontal.X).Width(viewport.W).Background(c.Theme().Background).Children(func() { s.changeItem(c, item, true) })
		}
	}
	if sectionNeeds {
		item := items[section]
		ui.Column(c.Key(fmt.Sprintf("sticky-section-%t", item.staged))).Absolute().Top(sectionPush).Left(s.diffView.horizontal.X).Width(viewport.W).Background(c.Theme().Background).Children(func() { s.changeItem(c, item, true) })
	}
}

// Use measured mounted row heights. When jumping over an unmounted group,
// derive the same native button/text chrome instead of a fixed pixel guess.
func (s *State) changeHeaderHeight(c *ui.Context, item changeItem) float32 {
	if h := s.diffView.headerHeights[changeHeaderKey(item)]; h > 0 {
		return h
	}
	_, text := c.MeasureText(0, ui.Span{Text: "Review", Font: c.Theme().Font, Size: c.Theme().FontSize})
	button := text + c.Theme().Space(3)
	if c.Theme().Border.A > 0 {
		button += 2
	}
	if isChangeSection(item) {
		if item.count == 0 {
			return text + 16
		}
		return button + 16
	}
	return max(button, text+10) + 8
}
