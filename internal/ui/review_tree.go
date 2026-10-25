package orbitui

import (
	"sort"
	"strings"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

// Source: review.rs::directory_paths/tree_rows. Ordering is ASCII-insensitive,
// stable for equal paths, and filtering reveals ancestors without mutating the
// user's unfiltered expansion state.
type reviewTreeRow struct {
	path, name          string
	depth, file         int
	directory, expanded bool
}

func reviewFoldASCII(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, s)
}
func reviewDirectories(files []workspace.ReviewFile) map[string]bool {
	out := map[string]bool{}
	for _, f := range files {
		parts := strings.Split(f.Path, "/")
		for i := 1; i < len(parts); i++ {
			out[strings.Join(parts[:i], "/")] = true
		}
	}
	return out
}
func reviewTreeRows(files []workspace.ReviewFile, expanded map[string]bool, filter string) []reviewTreeRow {
	filter = reviewFoldASCII(strings.TrimSpace(filter))
	var indexes []int
	for i, f := range files {
		if filter == "" || strings.Contains(reviewFoldASCII(f.Path), filter) {
			indexes = append(indexes, i)
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		return reviewFoldASCII(files[indexes[i]].Path) < reviewFoldASCII(files[indexes[j]].Path)
	})
	var rows []reviewTreeRow
	emitted := map[string]bool{}
	for _, index := range indexes {
		parts := strings.Split(files[index].Path, "/")
		visible := true
		for depth, part := range parts[:len(parts)-1] {
			path := strings.Join(parts[:depth+1], "/")
			open := filter != "" || expanded[path]
			if !emitted[path] && visible {
				rows = append(rows, reviewTreeRow{path: path, name: part, depth: depth, file: -1, directory: true, expanded: open})
				emitted[path] = true
			}
			if !open {
				visible = false
				break
			}
		}
		if visible {
			rows = append(rows, reviewTreeRow{path: files[index].Path, name: parts[len(parts)-1], depth: len(parts) - 1, file: index})
		}
	}
	return rows
}
func (s *reviewState) syncTreeRows() {
	var key strings.Builder
	for _, f := range s.data.Files {
		key.WriteString(f.Path)
		key.WriteByte(0)
	}
	if s.treeFileKey != key.String() || s.expandedDirs == nil {
		dirs := reviewDirectories(s.data.Files)
		expanded := map[string]bool{}
		for path := range dirs {
			if s.expandedDirs[path] || !s.treeDirs[path] {
				expanded[path] = true
			}
		}
		s.treeDirs, s.expandedDirs = dirs, expanded
		s.treeFileKey = key.String()
		s.treeDirty = true
	}
	if s.treeDirty || s.treeLastQuery != s.treeQuery {
		s.treeRows = reviewTreeRows(s.data.Files, s.expandedDirs, s.treeQuery)
		s.treeLastQuery, s.treeDirty = s.treeQuery, false
	}
}
func (s *reviewState) toggleTreeDirectory(path string) {
	if s.expandedDirs[path] {
		delete(s.expandedDirs, path)
	} else {
		s.expandedDirs[path] = true
	}
	s.treeDirty = true
	s.syncTreeRows()
}
func (a *App) reviewPaneWidth() float32 {
	a.initReviewLayout()
	return a.reviewState.layout.width
}
func (a *App) reviewTree(c *ui.Context) {
	s, t := &a.reviewState, a.Theme
	s.syncTreeRows()
	width := max(float32(180), min(a.reviewPaneWidth()*.42, float32(240)))
	ui.Column(c.Key("review-tree-column")).Label("Review file tree").Width(width).Shrink(0).FillHeight().MinHeight(0).BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		ui.Row(c).Height(t.Px(36)).PaddingX(t.Px(10)).Gap(t.Space(6)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			a.icon(c, "icons/search.svg", t.Px(14), t.TextMuted)
			input := ui.TextInputBase(c.Key("review-tree-filter"), &s.treeQuery).Label("Filter review files").Placeholder(a.text("sidepane.filter_files", "Filter files…")).Grow(1).MinWidth(0).FontSize(t.FontSize)
			if input.Shortcut(0, ui.KeyEscape) {
				s.treeQuery = ""
			}
		})
		s.syncTreeRows()
		tree := ui.Column(c.Key("review-file-tree")).Bind(&s.treeHandle).Focusable().Label("Review files").Grow(1).MinHeight(0)
		tree.Children(func() {
			// A row click may rebuild the flattened tree during this build.
			// Keep this frame's count/key/renderer on one immutable slice.
			rows := s.treeRows
			s.treeList.Key = func(i int) any {
				r := rows[i]
				return struct {
					Path string
					File int
				}{r.path, r.file}
			}
			ui.List(c.Key("review-tree-rows"), &s.treeList, len(rows), func(index int) {
				r := rows[index]
				ui.Row(c).Height(30).PaddingX(t.Space(6)).AlignItems(ui.Center).Children(func() {
					pad := float32(23) + float32(r.depth)*14
					if r.directory {
						pad = 7 + float32(r.depth)*14
					}
					label := "Review file " + r.path
					if r.directory {
						label = "Review directory " + r.path
					}
					b := ui.ButtonBase(c.Key(r.path)).Label(label).Height(26).Grow(1).MinWidth(0).Padding(0, 7, 0, pad).Radius(t.Px(6)).Gap(t.Space(6)).AlignItems(ui.Center).Justify(ui.Start)
					cursor := s.treeCursor != nil && *s.treeCursor == index
					selected := !r.directory && s.selected == r.file
					if selected {
						b.Background(t.Selection)
					} else if cursor && r.directory {
						b.Background(t.OverlayStrong())
					} else if cursor || b.Hovered() {
						b.Background(t.Hover)
					}
					b.Children(func() {
						icon := "icons/file.svg"
						if r.directory {
							arrow := "icons/chevron-right.svg"
							if r.expanded {
								arrow = "icons/chevron-down.svg"
							}
							a.icon(c, arrow, t.Px(10), t.TextMuted)
							icon = "icons/folder.svg"
						}
						a.icon(c, icon, t.Px(14), t.TextMuted)
						color := t.TextSecondary
						if selected {
							color = t.Text
						}
						label := ui.Text(c, r.name).FontSize(t.Px(12)).TextColor(color).Grow(1).MinWidth(0).MaxLines(1)
						if r.directory {
							label.FontWeight(500)
						} else {
							status, color := "M", t.Warning
							switch s.data.Files[r.file].Status {
							case "added":
								status, color = "A", t.Success
							case "deleted":
								status, color = "D", t.Error
							case "binary":
								status = "B"
							}
							ui.Box(c).Size(18, 18).Shrink(0).Radius(t.Px(4)).Border(1, color.Alpha(.65)).Center().Children(func() { ui.Text(c, status).FontSize(t.Px(12)).FontWeight(600).TextColor(color) })
						}
					})
					if b.Clicked() {
						c.AnimationFrame()
						tree.Focus()
						s.treeCursor = &index
						if r.directory {
							s.toggleTreeDirectory(r.path)
						} else {
							s.selected = r.file
							a.reviewFile(0)
						}
					}
				})
			}).Grow(1).MinHeight(0).PaddingY(t.Space(4))
		})
	})
}
