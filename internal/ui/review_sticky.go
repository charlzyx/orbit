package orbitui

import (
	"fmt"
	"path/filepath"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

func (a *App) reviewFileHeader(c *ui.Context, l workspace.ReviewLine, sticky bool) {
	s, t := &a.reviewState, a.Theme
	path := l.Content
	additions, deletions := 0, 0
	if l.File >= 0 && l.File < len(s.data.Files) {
		file := s.data.Files[l.File]
		path, additions, deletions = file.Path, file.Additions, file.Deletions
	}
	row := ui.Row(c)
	if sticky {
		row.Label(fmt.Sprintf("Review sticky file header %d", l.File))
	} else {
		if s.headerHandles == nil {
			s.headerHandles = map[int]*ui.Handle{}
		}
		if s.headerHandles[l.File] == nil {
			s.headerHandles[l.File] = &ui.Handle{}
		}
		row.Label(fmt.Sprintf("Review file header %d", l.File)).Bind(s.headerHandles[l.File])
	}
	row.Height(34).Background(t.Surface).PaddingX(t.Space(12)).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
		icon := "icons/chevron-down.svg"
		if s.collapsed[l.File] {
			icon = "icons/chevron-right.svg"
		}
		if a.iconButton(c, "Toggle file diff", icon).Clicked() {
			if s.collapsed == nil {
				s.collapsed = map[int]bool{}
			}
			s.collapsed[l.File] = !s.collapsed[l.File]
			s.selected = l.File
		}
		// Rust diff_view::render_file_header uses a file glyph (plain file SVG
		// when no Nerd Font is installed), never an extension badge.
		a.icon(c, "icons/file.svg", t.Px(14), t.TextMuted).Label("Review file icon").Shrink(0)
		ui.Text(c, path).MaxLines(1).Font(a.Fonts.CodeFamily).FontSize(t.Px(12)).TextColor(t.TextSecondary).FontWeight(500).Grow(1).MinWidth(0)
		addedColor, deletedColor := t.TextMuted, t.TextMuted
		if additions > 0 {
			addedColor = t.Success
		}
		if deletions > 0 {
			deletedColor = t.Error
		}
		// Zero counts remain visible, including empty and binary files. These
		// are snapshot metadata, not counts recomputed from visible/context rows.
		ui.Text(c, fmt.Sprintf("+%d", additions)).FontSize(t.Px(12)).TextColor(addedColor).Shrink(0)
		ui.Text(c, fmt.Sprintf("-%d", deletions)).FontSize(t.Px(12)).TextColor(deletedColor).Shrink(0)
		if a.iconButton(c, "Open file", "icons/external-link.svg").Clicked() {
			a.OpenFile(filepath.Join(a.Model.Workspace, path))
		}
	})
}

// Match Rust sidepane.rs:1485–1550 with a native absolute child, rather than
// List.Header: MyGo's Header changes ScrollTo(Start) to reserve header space,
// which changes the source's logical hunk/split navigation anchors.
// Attach the absolute child to the list, so wheel events over the header
// reach the real vertical list and its horizontal ancestor. Counter-pan keeps
// its screen position equivalent to the source's outside-scroller layer.
func (a *App) reviewStickyHeader(c *ui.Context) {
	s := &a.reviewState
	first, _ := s.list.Visible() // requests a frame when the visible range changes
	visible, _ := s.visible()
	if s.split {
		rows := workspace.SplitReview(visible)
		if first < 0 || first >= len(rows) {
			return
		}
		line := rows[first].Full
		if line == nil {
			line = rows[first].Left
		}
		if line == nil {
			line = rows[first].Right
		}
		if line == nil {
			return
		}
		for i, l := range visible {
			if reviewLineKey(l) == reviewLineKey(*line) {
				first = i
				break
			}
		}
	}
	if first < 0 || first >= len(visible) {
		return
	}
	header := -1
	for i := first; i >= 0; i-- {
		if visible[i].Kind == "file_header" {
			header = i
			break
		}
	}
	if header < 0 {
		return
	}
	viewport := s.diffViewport.Bounds(c)
	if viewport.H <= 0 {
		return
	}
	line := visible[header]
	normal := ui.Rect{}
	if h := s.headerHandles[line.File]; h != nil {
		normal = h.Bounds(c)
	}
	if header == first && (normal.H == 0 || normal.Y >= viewport.Y) {
		return
	}
	offset := float32(0)
	for i := header + 1; i < len(visible); i++ {
		if visible[i].Kind != "file_header" {
			continue
		}
		if h := s.headerHandles[visible[i].File]; h != nil {
			next := h.Bounds(c)
			if next.H > 0 {
				offset = min(float32(0), next.Y-viewport.Y-34)
			}
		}
		break
	}
	ui.Column(c.Key(fmt.Sprintf("review-sticky-file-%d", line.File))).Absolute().Top(offset).Left(s.horizontal.X).Width(viewport.W).Children(func() { a.reviewFileHeader(c, line, true) })
}
