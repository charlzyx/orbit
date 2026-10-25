package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"strings"
)

type sidebarStickyState struct {
	viewport ui.Handle
	headers  map[string]*ui.Handle
	layout   string
	scroll   ui.ScrollState
	lastY    float32
}
type sidebarStickyPosition struct {
	index  int
	offset float32
}

func (a *App) sidebarHeaderHandle(path string) *ui.Handle {
	s := &a.sidebarState.sticky
	if s.headers == nil {
		s.headers = map[string]*ui.Handle{}
	}
	if s.headers[path] == nil {
		s.headers[path] = &ui.Handle{}
	}
	return s.headers[path]
}

// Rust sidebar.rs:482–529. Keep the native List's own logical rows and anchor;
// an absolute child supplies only the pinned presentation, never List.Header.
func (a *App) prepareSidebarSticky(c *ui.Context, rows []sidebarRow) sidebarStickyPosition {
	s := &a.sidebarState.sticky
	none := sidebarStickyPosition{index: -1}
	first, _ := a.sessionList.Visible()
	var key strings.Builder
	_, windowH := c.Size()
	fmt.Fprintf(&key, "%s:%g:%d:%g:%g:%s|", a.Fonts.UIFamily, a.Theme.FontSize, a.Theme.Density, a.SidebarWidth, windowH, a.Workspaces.GroupBy)
	for _, r := range rows {
		fmt.Fprintf(&key, "%s:%s:%t|", r.key, r.path, r.collapsed)
	}
	if key.String() != s.layout {
		s.layout = key.String()
		s.headers = map[string]*ui.Handle{}
		c.AnimationFrame()
		return none
	}
	if a.Workspaces.GroupBy == "one_list" || len(rows) == 0 || first < 0 {
		return none
	}
	first = min(first, len(rows)-1)
	header := -1
	for i := first; i >= 0; i-- {
		if rows[i].group {
			header = i
			break
		}
	}
	if header < 0 || rows[header].collapsed {
		return none
	}
	viewport := s.viewport.Bounds(c)
	if viewport.H <= 0 {
		return none
	}
	normal := ui.Rect{}
	if h := s.headers[rows[header].path]; h != nil {
		normal = h.Bounds(c)
	}
	if first == header && (normal.H <= 0 || normal.Y >= viewport.Y) {
		return none
	}
	offset := float32(0)
	for i := header + 1; i < len(rows); i++ {
		if !rows[i].group {
			continue
		}
		if h := s.headers[rows[i].path]; h != nil {
			next := h.Bounds(c)
			if next.H > 0 {
				offset = min(float32(0), next.Y-viewport.Y-next.H)
			}
		}
		break
	}
	return sidebarStickyPosition{header, offset}
}
func (a *App) renderSidebarSticky(c *ui.Context, rows []sidebarRow, p sidebarStickyPosition) {
	if p.index < 0 || p.index >= len(rows) {
		return
	}
	r := rows[p.index]
	viewport := a.sidebarState.sticky.viewport.Bounds(c)
	// Descendant of the real List: wheel bubbling and native scrollbar behavior
	// remain intact; the list viewport clips the next-group push-off.
	ui.Column(c.Key("sidebar-sticky:" + r.path)).Absolute().Top(p.offset).Left(0).Width(viewport.W).Background(a.Theme.Surface).Label("Sidebar sticky workspace " + r.path).Children(func() { a.workspaceHeader(c, r, true, true) })
}

// MyGo's public ScrollbarWidth plus its six-DIP track hit padding. Reserve
// the complete track, not merely the visible thumb, before action buttons.
func sidebarScrollbarGutter(c *ui.Context) float32 {
	width := c.Theme().ScrollbarWidth
	if width <= 0 {
		width = 6
	}
	return max(float32(8), sidebarResizeGripWidth+width+6)
}
