package orbitui

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type reviewState struct {
	source                     workspace.ReviewSource
	data                       workspace.ReviewSnapshot
	list                       ui.ListState
	loading, split, wrap, tree bool
	err, root, ref, path       string
	selected                   int
	generation                 uint64
	pending                    bool
	handle                     ui.Handle
	collapsed                  map[int]bool
	dataIdentity               string
}

func (a *App) RefreshReview() {
	s := &a.reviewState
	s.generation++
	generation := s.generation
	if s.loading {
		s.pending = true
		return
	}
	if s.source == "" {
		s.source = workspace.ReviewUncommitted
	}
	s.loading = true
	root := a.Model.Workspace
	ref, path := s.ref, s.path
	req := workspace.ReviewRequest{Source: s.source, Session: a.Model.SessionID, Turn: a.Model.Turn}
	identity := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s", root, req.Source, req.Session, req.Turn, ref, path)
	if s.dataIdentity != identity {
		s.data = workspace.ReviewSnapshot{}
		s.err = ""
	}
	a.Async(func() func() {
		g := workspace.Git{Dir: root}
		var result workspace.ReviewSnapshot
		var err error
		if ref != "" || path != "" {
			d, e := g.CollectReviewAt(context.Background(), req.Source, ref, path)
			err = e
			result = workspace.ParseReview(d)
		} else if req.Source == workspace.ReviewUncommitted {
			st, e := g.CollectReview(context.Background(), workspace.ReviewRequest{Source: workspace.ReviewStaged})
			if e != nil {
				err = e
			} else {
				un, e := g.CollectReview(context.Background(), workspace.ReviewRequest{Source: workspace.ReviewUnstaged})
				err = e
				result = workspace.GroupReviews(workspace.ParseReview(st), workspace.ParseReview(un))
			}
		} else {
			d, e := g.CollectReview(context.Background(), req)
			err = e
			result = workspace.ParseReview(d)
		}
		return func() {
			s.loading = false
			if generation != s.generation || root != a.Model.Workspace || req.Source != s.source || ref != s.ref || path != s.path || req.Session != a.Model.SessionID {
				s.pending = false
				a.RefreshReview()
				return
			}
			s.root = root
			s.err = ""
			if err != nil {
				s.err = err.Error()
			} else {
				s.data = result
				s.dataIdentity = identity
				s.collapsed = map[int]bool{}
				s.list = ui.ListState{}
			}
		}
	})
}
func (a *App) review(c *ui.Context) {
	s := &a.reviewState
	t := a.Theme
	if s.root != a.Model.Workspace && !s.loading {
		a.RefreshReview()
	}
	ui.Column(c.Key("review-pane")).Bind(&s.handle).Width(max(360, a.ReviewWidth)).FillHeight().BorderWidth(0, 0, 0, 1).BorderColor(t.Border).Children(func() {
		ui.Row(c).Height(40).PaddingX(10).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.MenuButton(c, reviewLabel(s.source), func(m *ui.Menu) {
				for _, src := range []workspace.ReviewSource{workspace.ReviewLastTurn, workspace.ReviewUncommitted, workspace.ReviewUnstaged, workspace.ReviewStaged, workspace.ReviewCommitted, workspace.ReviewBranch} {
					if m.Item(reviewLabel(src)).Checked(src == s.source).Chosen() {
						s.source = src
						s.ref = ""
						s.path = ""
						a.RefreshReview()
					}
				}
			})
			ui.Spacer(c)
			if s.loading {
				ui.Spinner(c).Size(12, 12)
			}
			if a.iconButton(c, "Refresh review", "icons/refresh.svg").Clicked() {
				a.RefreshReview()
			}
			if a.iconButton(c, "Review options", "icons/more.svg").Menu(func(m *ui.Menu) {
				if m.Item("Split diff").Checked(s.split).Chosen() {
					s.split = !s.split
				}
				if m.Item("Wrap lines").Checked(s.wrap).Chosen() {
					s.wrap = !s.wrap
				}
				if m.Item("File list").Checked(s.tree).Chosen() {
					s.tree = !s.tree
				}
			}).Clicked() {
			}
			if a.iconButton(c, "Close review", "icons/x.svg").Clicked() {
				a.ReviewVisible = false
			}
		})
		ui.Row(c).Padding(6, 12).Gap(8).Children(func() {
			ui.Text(c, fmt.Sprintf("%d files", len(s.data.Files))).TextColor(t.TextMuted)
			ui.Text(c, fmt.Sprintf("+%d", s.data.Additions)).TextColor(t.Success)
			ui.Text(c, fmt.Sprintf("−%d", s.data.Deletions)).TextColor(t.Error)
		})
		if s.err != "" {
			if len(s.data.Lines) == 0 {
				a.Empty(c, "Review unavailable", s.err)
				return
			}
			ui.Text(c, "Refresh failed; showing the previous snapshot. "+s.err).TextColor(t.Warning).Padding(8).Selectable()
		}
		if len(s.data.Lines) == 0 {
			a.Empty(c, "No changes", "No changes in this review source.")
			return
		}
		if s.tree {
			ui.Scroll(c).MaxHeight(180).Children(func() {
				for i, f := range s.data.Files {
					f := f
					if a.button(c, f.Path, "icons/file.svg", false).AlignSelf(ui.Stretch).Clicked() {
						s.selected = i
						a.reviewFile(0)
					}
				}
			})
		}
		if s.split {
			visible, indexes := s.visible()
			_ = indexes
			rows := workspace.SplitReview(visible)
			ui.List(c.Key("split-review"), &s.list, len(rows), func(i int) {
				r := rows[i]
				if r.Full != nil {
					a.reviewLine(c, *r.Full, -1)
				} else {
					ui.Row(c).Children(func() {
						ui.Column(c).WidthPercent(50).MinWidth(0).Children(func() {
							if r.Left != nil {
								a.reviewLine(c, *r.Left, -1)
							}
						})
						ui.Column(c).WidthPercent(50).MinWidth(0).Children(func() {
							if r.Right != nil {
								a.reviewLine(c, *r.Right, -1)
							}
						})
					})
				}
			}).Grow(1).MinHeight(0)
		} else {
			visible, indexes := s.visible()
			ui.List(c.Key("unified-review"), &s.list, len(visible), func(i int) { a.reviewLine(c, visible[i], indexes[i]) }).Grow(1).MinHeight(0)
		}
		if s.data.Truncated {
			ui.Text(c, "Diff display limited to 50,000 lines.").Padding(8).TextColor(t.Warning)
		}
		if c.Shortcut(ui.Alt, ui.KeyDown) {
			a.reviewFile(1)
		}
		if c.Shortcut(ui.Alt, ui.KeyUp) {
			a.reviewFile(-1)
		}
	})
}
func reviewLabel(s workspace.ReviewSource) string {
	switch s {
	case workspace.ReviewLastTurn:
		return "Last turn"
	case workspace.ReviewStaged:
		return "Staged"
	case workspace.ReviewUnstaged:
		return "Unstaged"
	case workspace.ReviewCommitted:
		return "Committed"
	case workspace.ReviewBranch:
		return "Branch changes"
	default:
		return "Uncommitted changes"
	}
}
func (a *App) reviewFile(delta int) {
	s := &a.reviewState
	if len(s.data.Files) == 0 {
		return
	}
	s.selected = (s.selected + delta + len(s.data.Files)) % len(s.data.Files)
	if s.collapsed == nil {
		s.collapsed = map[int]bool{}
	}
	s.collapsed[s.selected] = false
	for _, marker := range s.markers("file_header") {
		if marker.file == s.selected {
			s.list.ScrollTo(marker.row, ui.Start)
			break
		}
	}
}
func (a *App) reviewLine(c *ui.Context, l workspace.ReviewLine, index int) {
	t := a.Theme
	s := &a.reviewState
	switch l.Kind {
	case "file_header":
		path := l.Content
		if l.File >= 0 && l.File < len(s.data.Files) {
			path = s.data.Files[l.File].Path
		}
		ui.Row(c).Height(34).Background(t.Surface).PaddingX(10).AlignItems(ui.Center).Children(func() {
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
			ui.Text(c, path).MaxLines(1).FontWeight(500).Grow(1)
			if a.iconButton(c, "Open file", "icons/external-link.svg").Clicked() {
				a.OpenFile(filepath.Join(a.Model.Workspace, path))
			}
		})
		return
	case "section_header":
		ui.Text(c, l.Content).Padding(8, 10).Background(t.Elevated).FontWeight(600)
		return
	case "gap":
		if l.Gap == nil {
			return
		}
		b := a.button(c, fmt.Sprintf("Show %d unchanged lines", l.Gap.Count), "", false).AlignSelf(ui.Stretch).Background(t.Surface)
		b.Disabled(!l.Gap.Expandable())
		if b.Clicked() {
			if index < 0 {
				for i, x := range s.data.Lines {
					if x.Gap != nil && x.Gap.ID == l.Gap.ID && x.File == l.File {
						index = i
						break
					}
				}
			}
			s.data.ExpandGap(index, "both")
		}
		return
	}
	color, bg, prefix := t.TextSecondary, t.Background, " "
	switch l.Kind {
	case "addition":
		color = t.Text
		bg = t.Success.Alpha(.10)
		prefix = "+"
	case "deletion":
		color = t.Text
		bg = t.Error.Alpha(.10)
		prefix = "−"
	case "hunk_header", "meta":
		color = t.TextMuted
		bg = t.Surface
	}
	ui.Row(c).MinHeight(20).Background(bg).Gap(6).PaddingX(6).Children(func() {
		num := ""
		if l.NewLine > 0 {
			num = fmt.Sprint(l.NewLine)
		} else if l.OldLine > 0 {
			num = fmt.Sprint(l.OldLine)
		}
		ui.Text(c, num).Width(36).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize).TextColor(t.TextMuted)
		ui.Text(c, prefix).Width(10).Font(a.Fonts.CodeFamily).TextColor(color)
		text := ui.Text(c, l.Content).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize).TextColor(color).Grow(1).MinWidth(0)
		if !s.wrap {
			text.MaxLines(1)
		}
	})
}

func (s *reviewState) visible() ([]workspace.ReviewLine, []int) {
	var lines []workspace.ReviewLine
	var indexes []int
	for i, l := range s.data.Lines {
		if l.Kind == "file_header" || l.Kind == "section_header" || !s.collapsed[l.File] {
			lines = append(lines, l)
			indexes = append(indexes, i)
		}
	}
	return lines, indexes
}

type reviewMarker struct{ row, file int }

// List navigation addresses the actual rendered rows, not original diff
// offsets: collapsed bodies and split deletion/addition pairs change indexes.
func (s *reviewState) markers(kind string) []reviewMarker {
	visible, _ := s.visible()
	var out []reviewMarker
	if s.split {
		for i, r := range workspace.SplitReview(visible) {
			if r.Full != nil && r.Full.Kind == kind {
				out = append(out, reviewMarker{i, r.Full.File})
			}
		}
	} else {
		for i, l := range visible {
			if l.Kind == kind {
				out = append(out, reviewMarker{i, l.File})
			}
		}
	}
	return out
}
func (a *App) reviewCommand(id string) {
	s := &a.reviewState
	switch id {
	case "ReviewClose":
		a.ReviewVisible = false
		a.composer.Focus()
	case "ReviewFileNext", "ReviewTreeNext":
		a.reviewFile(1)
	case "ReviewFilePrev", "ReviewTreePrev":
		a.reviewFile(-1)
	case "ReviewTreeToggle":
		if s.collapsed == nil {
			s.collapsed = map[int]bool{}
		}
		s.collapsed[s.selected] = !s.collapsed[s.selected]
	case "ReviewExpandAll":
		s.collapsed = map[int]bool{}
	case "ReviewCollapseAll":
		s.collapsed = map[int]bool{}
		for i := range s.data.Files {
			s.collapsed[i] = true
		}
	case "ReviewHunkNext", "ReviewHunkPrev":
		delta := 1
		if id == "ReviewHunkPrev" {
			delta = -1
		}
		first, _ := s.list.Visible()
		markers := s.markers("hunk_header")
		if delta > 0 {
			for _, m := range markers {
				if m.row > first {
					s.list.ScrollTo(m.row, ui.Start)
					break
				}
			}
		} else {
			for i := len(markers) - 1; i >= 0; i-- {
				if markers[i].row < first {
					s.list.ScrollTo(markers[i].row, ui.Start)
					break
				}
			}
		}
	}
}
