package orbitui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

type reviewState struct {
	layout                                reviewLayoutState
	turn                                  int // Zero follows latest completed turn; positive pins historical review.
	source                                workspace.ReviewSource
	data                                  workspace.ReviewSnapshot
	list                                  ui.ListState
	horizontal                            ui.ScrollState
	widthCache                            reviewWidthCache
	vertical                              ui.ScrollState
	diffViewport                          ui.Handle
	headerHandles                         map[int]*ui.Handle
	stickyPanY                            float32
	loading, split, wrap, tree            bool
	err, root, ref, path                  string
	pendingPath, pendingRoot              string
	selected                              int
	diffCursor                            *int // Original snapshot index, independent of viewport and split rows.
	treeCursor                            *int
	treeRows                              []reviewTreeRow
	treeList                              ui.ListState
	treeQuery, treeLastQuery, treeFileKey string
	treeDirs, expandedDirs                map[string]bool
	treeDirty                             bool
	generation                            uint64
	pending                               bool
	handle, treeHandle                    ui.Handle
	collapsed                             map[int]bool
	dataIdentity                          string
	refreshSpinUntil                      time.Time
	refreshStarted                        time.Time
	sourceMenuOpen                        bool
	sourceMenuDismissedAt                 time.Time
}

const reviewRefreshFeedback = 650 * time.Millisecond

// Loading controls content; click feedback has an independent minimum lifetime.
// Source: sidepane.rs::refresh_from_button. Repeated clicks extend the deadline.
func (s *reviewState) refreshFeedback(now time.Time) bool {
	return s.loading || now.Before(s.refreshSpinUntil)
}
func (a *App) refreshReviewFromButton(now time.Time) {
	s := &a.reviewState
	s.refreshSpinUntil = now.Add(reviewRefreshFeedback)
	if !s.loading {
		a.RefreshReview()
	}
	a.Async(func() func() {
		time.Sleep(reviewRefreshFeedback)
		return func() {
			if !time.Now().Before(s.refreshSpinUntil) {
				s.refreshSpinUntil = time.Time{}
			}
		}
	})
}

func (a *App) RefreshReview() {
	s := &a.reviewState
	if s.source == workspace.ReviewLastTurn && a.Model.LatestTurn == 0 {
		s.source = workspace.ReviewUncommitted
		s.turn = 0
	}
	if s.source == workspace.ReviewLastTurn && s.turn == 0 {
		s.turn = a.Model.LatestTurn
	}
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
	if s.pendingRoot != root {
		s.pendingPath, s.pendingRoot = "", ""
	}
	ref, path := s.ref, s.path
	req := workspace.ReviewRequest{Source: s.source, Session: a.Model.SessionID, Turn: a.reviewTurn()}
	identity := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s", root, req.Source, req.Session, req.Turn, ref, path)
	if s.dataIdentity != identity {
		s.clearReviewSnapshot()
	}
	a.Async(func() func() {
		g := workspace.Git{Dir: root}
		var result workspace.ReviewSnapshot
		var err error
		if ref != "" || path != "" {
			d, e := g.CollectReviewAt(context.Background(), req.Source, ref, path)
			err = e
			result = workspace.ParseReview(d)
		} else {
			// Docked Review compares HEAD with the working tree once. Only
			// the source embedded Git browser groups Staged/Changes sections.
			d, e := g.CollectReview(context.Background(), req)
			err = e
			result = workspace.ParseReview(d)
		}
		return func() {
			s.loading = false
			if generation != s.generation || root != a.Model.Workspace || req.Source != s.source || ref != s.ref || path != s.path || req.Session != a.Model.SessionID || (req.Source == workspace.ReviewLastTurn && req.Turn != a.reviewTurn()) {
				s.pending = false
				a.RefreshReview()
				return
			}
			s.root = root
			s.err = ""
			if err != nil {
				s.err = err.Error()
			} else {
				a.applyReviewSnapshot(result)
				s.dataIdentity = identity
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
	ui.Column(c.Key("review-pane")).Bind(&s.handle).Width(a.reviewPaneWidth()).FillHeight().Padding(0, 0, 0, t.Space(12)).Children(func() {
		// The source dock is an opaque rounded card with a separate left
		// resize gutter, not a transparent overlay on the new-task backdrop.
		ui.Column(c.Key("review-card")).Label("Review card").Fill().MinHeight(0).Background(t.Background).Border(1, t.Border).Radius(t.Px(12)).Clip().Children(func() {
			a.reviewToolbar(c)
			if s.err != "" {
				if len(s.data.Lines) == 0 {
					a.Empty(c, "Review unavailable", s.err)
					return
				}
				ui.Text(c, "Refresh failed; showing the previous snapshot. "+s.err).TextColor(t.Warning).Padding(8).Selectable()
			}
			if len(s.data.Lines) == 0 {
				if s.loading && s.dataIdentity == "" {
					a.Empty(c, a.text("sidepane.loading_changes", "Loading changes…"), "")
					return
				}
				a.Empty(c, a.text("sidepane.no_changes", "No changes"), a.reviewEmptyDescription())
				return
			}
			ui.Row(c).Grow(1).MinHeight(0).Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).FillHeight().Children(func() { a.reviewDiff(c) })
				if s.tree && a.reviewPaneWidth() >= 440 {
					a.reviewTree(c)
				}
			})
			if c.Shortcut(ui.Alt, ui.KeyDown) {
				a.reviewFile(1)
			}
			if c.Shortcut(ui.Alt, ui.KeyUp) {
				a.reviewFile(-1)
			}
		})
		a.reviewResizeHandle(c)
	})
}
func reviewLabel(s workspace.ReviewSource) string {
	switch s {
	case workspace.ReviewLastTurn:
		return "Last Turn"
	case workspace.ReviewStaged:
		return "Staged"
	case workspace.ReviewUnstaged:
		return "Unstaged"
	case workspace.ReviewCommitted:
		return "Committed"
	case workspace.ReviewBranch:
		return "Branch"
	default:
		return "Uncommitted"
	}
}
func (a *App) reviewFile(delta int) {
	s := &a.reviewState
	if len(s.data.Files) == 0 {
		return
	}
	// sidepane.rs::review_move_file clamps at the ends rather than wrapping.
	s.selected = max(0, min(s.selected+delta, len(s.data.Files)-1))
	s.diffCursor = nil
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

// Source apply_snapshot keeps the selected path across a refreshed/reordered
// snapshot, then resets its hunk cursor and expands the docked review body.
func (a *App) applyReviewSnapshot(data workspace.ReviewSnapshot) {
	s := &a.reviewState
	path := ""
	if s.selected >= 0 && s.selected < len(s.data.Files) {
		path = s.data.Files[s.selected].Path
	}
	s.data = data
	s.selected = 0
	// show_file selects a path inside the complete review, not a filtered diff.
	// A missing pending path falls back to the prior selection, as Rust does.
	for _, wanted := range []string{s.pendingPath, path} {
		found := false
		for i, file := range data.Files {
			if wanted != "" && file.Path == wanted {
				s.selected, found = i, true
				break
			}
		}
		if found {
			break
		}
	}
	s.pendingPath, s.pendingRoot = "", ""
	s.collapsed = map[int]bool{}
	s.list = ui.ListState{}
	s.treeCursor = nil
	a.reviewFile(0)
}

// Full-context snapshots suppress @@ metadata. The source also treats the
// first addition/deletion in each same-file change run as a hunk boundary.
func reviewHunkStart(lines []workspace.ReviewLine, i int) bool {
	l := lines[i]
	if l.Kind == "hunk_header" {
		return true
	}
	if l.Kind != "addition" && l.Kind != "deletion" {
		return false
	}
	return i == 0 || lines[i-1].File != l.File || (lines[i-1].Kind != "addition" && lines[i-1].Kind != "deletion")
}

// Convert a semantic snapshot index to the current virtual row. In split
// mode a deletion/addition pair occupies one row; collapsed content maps to
// its retained header, rather than to an unrelated later file.
func (s *reviewState) renderedIndex(index int) (int, bool) {
	if index < 0 || index >= len(s.data.Lines) {
		return 0, false
	}
	if s.collapsed[s.data.Lines[index].File] && s.data.Lines[index].Kind != "file_header" {
		file := s.data.Lines[index].File
		for i, l := range s.data.Lines {
			if l.Kind == "file_header" && l.File == file {
				index = i
				break
			}
		}
	}
	lines, indexes := s.visible()
	row := 0
	for i := 0; i < len(lines); {
		if s.split && (lines[i].Kind == "addition" || lines[i].Kind == "deletion") {
			file := lines[i].File
			var deletes, adds []int
			for i < len(lines) && lines[i].File == file && (lines[i].Kind == "addition" || lines[i].Kind == "deletion") {
				if lines[i].Kind == "deletion" {
					deletes = append(deletes, indexes[i])
				} else {
					adds = append(adds, indexes[i])
				}
				i++
			}
			for pair := 0; pair < max(len(deletes), len(adds)); pair++ {
				if (pair < len(deletes) && deletes[pair] == index) || (pair < len(adds) && adds[pair] == index) {
					return row, true
				}
				row++
			}
		} else {
			if indexes[i] == index {
				return row, true
			}
			i++
			row++
		}
	}
	return 0, false
}
func (a *App) reviewLine(c *ui.Context, l workspace.ReviewLine, index int) {
	if l.Kind == "file_header" || l.Kind == "section_header" {
		a.reviewPinnedRow(c, func() { a.reviewLineContent(c, l, index) })
		return
	}
	a.reviewLineContent(c, l, index)
}
func (a *App) reviewLineContent(c *ui.Context, l workspace.ReviewLine, index int) {
	t := a.Theme
	s := &a.reviewState
	switch l.Kind {
	case "file_header":
		a.reviewFileHeader(c, l, false)
		return
	case "section_header":
		ui.Text(c, l.Content).Padding(8, 10).Background(t.Elevated).FontWeight(600)
		return
	case "gap":
		a.reviewGap(c, l, index)
		return
	case "context", "addition", "deletion":
		number := l.NewLine
		if number == 0 {
			number = l.OldLine
		}
		a.reviewCodeRow(c, l, number, false)
		return
	}
	color, bg, prefix := t.TextSecondary, t.Background, " "
	switch l.Kind {
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
		spans := a.reviewSpans(l)
		text := ui.RichText(c, spans...).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize).TextColor(color).Grow(1).MinWidth(0)
		if !s.wrap {
			text.NoWrap()
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

type reviewLineIdentity struct {
	file, old, next int
	kind, content   string
	gap             uint64
}

func reviewLineKey(l workspace.ReviewLine) reviewLineIdentity {
	key := reviewLineIdentity{file: l.File, old: l.OldLine, next: l.NewLine, kind: l.Kind, content: l.Content}
	if l.Gap != nil {
		key.gap = l.Gap.ID
	}
	return key
}

func (a *App) toggleReviewSplit() {
	s := &a.reviewState
	first, _ := s.list.Visible()
	index := -1
	visible, indexes := s.visible()
	if !s.split && first >= 0 && first < len(indexes) {
		index = indexes[first]
	} else if s.split {
		rows := workspace.SplitReview(visible)
		if first >= 0 && first < len(rows) {
			line := rows[first].Full
			if line == nil {
				line = rows[first].Left
			}
			if line == nil {
				line = rows[first].Right
			}
			if line != nil {
				key := reviewLineKey(*line)
				for i, l := range s.data.Lines {
					if reviewLineKey(l) == key {
						index = i
						break
					}
				}
			}
		}
	}
	s.split = !s.split
	if row, ok := s.renderedIndex(index); ok {
		s.list.ScrollTo(row, ui.Start)
	}
}

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
		// view.rs::on_review_close leaves the dock open and only hands
		// keyboard focus back to the composer, like sidebar Escape.
		a.composer.Focus()
	case "ReviewFileNext":
		a.reviewFile(1)
	case "ReviewFilePrev":
		a.reviewFile(-1)
	case "ReviewTreeNext", "ReviewTreePrev":
		s.syncTreeRows()
		if len(s.treeRows) == 0 {
			return
		}
		next := 0
		if s.treeCursor != nil {
			delta := 1
			if id == "ReviewTreePrev" {
				delta = -1
			}
			next = max(0, min(*s.treeCursor+delta, len(s.treeRows)-1))
		}
		s.treeCursor = &next
		s.treeList.ScrollIntoView(next)
	case "ReviewTreeToggle":
		s.syncTreeRows()
		if s.treeCursor != nil && *s.treeCursor >= 0 && *s.treeCursor < len(s.treeRows) {
			r := s.treeRows[*s.treeCursor]
			if r.directory {
				s.toggleTreeDirectory(r.path)
			} else {
				s.selected = r.file
				a.reviewFile(0)
			}
		}
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
		current := 0
		if s.diffCursor != nil {
			current = *s.diffCursor
		} else if s.selected >= 0 && s.selected < len(s.data.Files) {
			current = max(0, s.data.Files[s.selected].DiffLine)
		}
		for i := current + delta; i >= 0 && i < len(s.data.Lines); i += delta {
			if !reviewHunkStart(s.data.Lines, i) {
				continue
			}
			s.diffCursor = &i
			s.selected = s.data.Lines[i].File
			if row, ok := s.renderedIndex(i); ok {
				s.list.ScrollTo(row, ui.Start)
			}
			break
		}
	}
}

func (a *App) reviewSpans(line workspace.ReviewLine) []ui.Span {
	runs := highlight.Runs(line.Content, line.Tokens)
	spans := make([]ui.Span, 0, len(runs))
	for _, run := range runs {
		span := ui.Span{Text: run.Text}
		if run.Class != "" {
			span.Color = a.Theme.TokenColor(run.Class)
		}
		spans = append(spans, span)
	}
	return spans
}

func (a *App) reviewTurn() int {
	if a.reviewState.source != workspace.ReviewLastTurn {
		return 0
	}
	if a.reviewState.turn > 0 {
		return a.reviewState.turn
	}
	return a.Model.LatestTurn
}
func (a *App) currentReviewLabel() string {
	if a.reviewState.source == workspace.ReviewLastTurn && a.reviewState.turn > 0 && a.reviewState.turn != a.Model.LatestTurn {
		return fmt.Sprintf("Turn %d", a.reviewState.turn)
	}
	return reviewLabel(a.reviewState.source)
}

// OpenTurnReview pins an actual captured turn; zero selects the current newest.
func (a *App) OpenTurnReview(turn int) {
	if turn < 0 || turn > a.Model.LatestTurn {
		return
	}
	if turn == 0 {
		turn = a.Model.LatestTurn
	}
	a.reviewState.turn = turn
	a.reviewState.source = workspace.ReviewLastTurn
	a.reviewState.ref = ""
	a.reviewState.path = ""
	a.reviewState.pendingPath, a.reviewState.pendingRoot = "", ""
	a.Navigate("chat")
	a.ReviewVisible = true
	a.ExplorerVisible = false
	a.RefreshReview()
}

// OpenReviewDiff routes Git's working-tree file action through show_file's
// full-snapshot/pending-selection semantics. An explicit commit ref retains
// the deliberate committed-file comparison API.
func (a *App) OpenReviewDiff(source, ref, path string) {
	s := &a.reviewState
	s.source, s.ref, s.path = workspace.ReviewSource(source), ref, path
	s.turn = 0
	s.pendingPath, s.pendingRoot = "", ""
	if ref == "" && path != "" {
		s.source = workspace.ReviewUncommitted
		s.path = ""
		s.pendingPath, s.pendingRoot = filepath.ToSlash(path), a.Model.Workspace
	}
	a.Navigate("chat")
	a.ReviewVisible, a.ExplorerVisible = true, false
	a.RefreshReview()
}

func (a *App) reviewDiffRows(c *ui.Context, width float32) {
	s := &a.reviewState
	// The inner list is wider than the viewport when unwrapped. Keep its
	// native vertical scrollbar at the visible right edge, for painting AND
	// pointer hit-testing; otherwise it is only reachable after panning right.
	rightInset := float32(0)
	if !s.wrap {
		viewport := s.horizontal.Bounds(c)
		if viewport.W > 0 {
			rightInset = max(0, max(width, viewport.W)-viewport.W-s.horizontal.X)
		}
	}
	if s.split {
		visible, indexes := s.visible()
		_ = indexes
		rows := workspace.SplitReview(visible)
		s.list.Header = nil
		s.list.Key = func(i int) any {
			r := rows[i]
			if r.Full != nil {
				return reviewLineKey(*r.Full)
			}
			if r.Left != nil {
				return reviewLineKey(*r.Left)
			}
			return reviewLineKey(*r.Right)
		}
		ui.List(c.Key("review-lines"), &s.list, len(rows), func(i int) {
			r := rows[i]
			if r.Full != nil {
				a.reviewLine(c, *r.Full, -1)
			} else {
				ui.Row(c).MinHeight(19).AlignItems(ui.Stretch).Children(func() {
					ui.Column(c).Grow(1).Basis(0).MinWidth(0).Background(a.Theme.overlay()).Children(func() {
						if r.Left != nil {
							a.reviewCodeRow(c, *r.Left, r.Left.OldLine, true).Grow(1)
						}
					})
					ui.Box(c).Width(1).Shrink(0).Background(a.Theme.Border)
					ui.Column(c).Grow(1).Basis(0).MinWidth(0).Background(a.Theme.overlay()).Children(func() {
						if r.Right != nil {
							a.reviewCodeRow(c, *r.Right, r.Right.NewLine, true).Grow(1)
						}
					})
				})
			}
		}).WidthPercent(100).MinWidth(width).FillHeight().Shrink(0).ScrollbarInsets(0, rightInset, 0, 0).TrackScroll(&s.vertical).Children(func() { a.reviewStickyHeader(c) })
	} else {
		visible, indexes := s.visible()
		s.list.Header = nil
		s.list.Key = func(i int) any { return reviewLineKey(visible[i]) }
		ui.List(c.Key("review-lines"), &s.list, len(visible), func(i int) { a.reviewLine(c, visible[i], indexes[i]) }).WidthPercent(100).MinWidth(width).FillHeight().Shrink(0).ScrollbarInsets(0, rightInset, 0, 0).TrackScroll(&s.vertical).Children(func() { a.reviewStickyHeader(c) })
	}
}
