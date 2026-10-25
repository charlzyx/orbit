package orbitui

import (
	"fmt"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo/ui"
)

const reviewSourceGesture = 200 * time.Millisecond

// sidepane.rs 1030–1218: title/window actions occupy the first40px row;
// source, direct view controls and responsive totals occupy the second.
func (a *App) reviewToolbar(c *ui.Context) {
	s, t := &a.reviewState, a.Theme
	ui.Row(c.Key("review-header")).Label("Review header").Height(40).Shrink(0).Padding(0, t.Space(6), 0, t.Space(12)).Gap(4).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		ui.Text(c, a.text("sidepane.review", "Review")).FontSize(t.Px(12)).FontWeight(500).Grow(1).MinWidth(0).MaxLines(1)
		if a.reviewPaneWidth() >= 440 {
			if a.reviewToggle(c.Key("review-tree-toggle"), "Toggle review file tree", "icons/folder.svg", s.tree, true).Clicked() {
				s.tree = !s.tree
				c.AnimationFrame()
			}
		}
		a.reviewRefreshButton(c)
		label, icon := "Full width review", "icons/arrow-expand.svg"
		if s.layout.full {
			label, icon = "Restore review width", "icons/arrow-shrink.svg"
		}
		if a.reviewToggle(c.Key("review-full-width"), label, icon, s.layout.full, true).KeepFocus().Clicked() {
			a.toggleReviewFull()
			c.AnimationFrame()
		}
		if a.reviewToggle(c.Key("review-close"), "Close review", "icons/x.svg", false, true).Clicked() {
			a.closeReviewDock()
		}
	})
	ui.Row(c.Key("review-toolbar")).Label("Review toolbar").Height(40).Shrink(0).PaddingX(10).Gap(8).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
		a.reviewSourceControl(c)
		ui.Spacer(c)
		if s.data.Truncated {
			ui.Text(c, a.text("sidepane.partial", "partial")).Label("Partial review diff").Tooltip("Partial diff: display limited to 50,000 lines.").FontSize(t.Px(12)).TextColor(t.Warning).Shrink(0)
		}
		label, icon := "Collapse all", "icons/collapse-all.svg"
		collapsed := false
		for _, v := range s.collapsed {
			collapsed = collapsed || v
		}
		if collapsed {
			label, icon = "Expand all", "icons/expand-all.svg"
		}
		if a.reviewToggle(c.Key("review-files"), label, icon, false, len(s.data.Files) > 0).Clicked() {
			if collapsed {
				a.reviewCommand("ReviewExpandAll")
			} else {
				a.reviewCommand("ReviewCollapseAll")
			}
			c.AnimationFrame()
		}
		label = "Wrap lines"
		if s.wrap {
			label = "Unwrap lines"
		}
		if a.reviewToggle(c.Key("review-wrap"), label, "icons/text-wrap.svg", s.wrap, true).Clicked() {
			s.wrap = !s.wrap
			c.AnimationFrame()
		}
		label, icon = "Side-by-side view", "icons/side-by-side.svg"
		if s.split {
			label, icon = "Unified view", "icons/unified-view.svg"
		}
		if a.reviewToggle(c.Key("review-split"), label, icon, s.split, true).Clicked() {
			a.toggleReviewSplit()
			c.AnimationFrame()
		}
		if a.reviewPaneWidth() >= 380 {
			ui.Row(c.Key("review-change-totals")).Label("Review change totals").Shrink(0).Gap(8).Children(func() {
				ui.Text(c, fmt.Sprintf("+%d", s.data.Additions)).FontSize(t.Px(12)).FontWeight(500).TextColor(t.Success)
				ui.Text(c, fmt.Sprintf("-%d", s.data.Deletions)).FontSize(t.Px(12)).FontWeight(500).TextColor(t.Error)
			})
		}
	})
}

// Source Default buttons are22px with14px glyphs, not the app's28px default.
func (a *App) reviewToggle(c *ui.Context, label, icon string, active, enabled bool) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Tooltip(label).Size(t.Px(22), t.Px(22)).Shrink(0).Radius(t.Px(6)).Center().Disabled(!enabled)
	color := t.TextMuted
	if active {
		color = t.Text
	}
	if !enabled {
		color = t.TextDisabled
	} else if b.Hovered() {
		b.Background(t.Hover)
		color = t.Text
	}
	b.Children(func() { a.icon(c, icon, t.Px(14), color) })
	return b
}

// The refresh glyph itself rotates every700ms; there is never a second loader.
// Reduced motion retains the accent feedback without requesting animation.
func (a *App) reviewRefreshButton(c *ui.Context) {
	s, t := &a.reviewState, a.Theme
	b := ui.ButtonBase(c.Key("review-refresh")).Label("Refresh review").Tooltip("Refresh").Size(t.Px(22), t.Px(22)).Shrink(0).Radius(t.Px(6)).Center()
	if b.Hovered() {
		b.Background(t.Hover)
	}
	if b.Clicked() {
		a.refreshReviewFromButton(time.Now())
		c.AnimationFrame()
	}
	active := s.refreshFeedback(time.Now())
	if active && s.refreshStarted.IsZero() {
		s.refreshStarted = c.Now()
	}
	if !active {
		s.refreshStarted = time.Time{}
	}
	b.Children(func() {
		color := t.TextMuted
		if active {
			color = t.AccentText
		}
		glyph := a.icon(c.Key("review-refresh-glyph"), "icons/refresh.svg", t.Px(14), color)
		if active {
			glyph.Label("Refreshing review")
			if !a.Prefs.ReduceMotion {
				phase := float32(c.Now().Sub(s.refreshStarted)%(700*time.Millisecond)) / float32(700*time.Millisecond)
				glyph.Rotate(phase * 360)
				c.AnimationFrame()
			}
		} else {
			glyph.Label("Review refresh glyph")
		}
	})
}

func (s *reviewState) toggleSourceMenu(now time.Time) {
	if !s.sourceMenuDismissedAt.IsZero() {
		dismissed := s.sourceMenuDismissedAt
		s.sourceMenuDismissedAt = time.Time{}
		if now.Sub(dismissed) < reviewSourceGesture {
			return
		}
	}
	s.sourceMenuOpen = !s.sourceMenuOpen
}

func (a *App) selectReviewSource(source workspace.ReviewSource, now time.Time) {
	s := &a.reviewState
	if source == workspace.ReviewLastTurn && a.Model.LatestTurn == 0 {
		return
	}
	s.sourceMenuOpen = false
	s.sourceMenuDismissedAt = now
	turn := 0
	if source == workspace.ReviewLastTurn {
		turn = a.Model.LatestTurn
	}
	if s.source == source && s.turn == turn {
		return
	}
	s.clearReviewSnapshot()
	s.source, s.turn = source, turn
	s.ref, s.path = "", ""
	s.pendingPath, s.pendingRoot = "", ""
	a.RefreshReview()
}

func (a *App) reviewSourceControl(c *ui.Context) {
	s, t := &a.reviewState, a.Theme
	anchor := ui.ButtonBase(c.Key("review-source")).Label("Review source").Height(t.Px(28)).Shrink(0).PaddingX(t.Space(8)).Gap(t.Space(4)).Radius(t.Px(6)).Border(1, t.Border).Background(t.Elevated).AlignItems(ui.Center)
	if s.sourceMenuOpen {
		anchor.BorderColor(t.BorderStrong())
	}
	if anchor.Hovered() {
		anchor.Background(t.Hover)
	}
	anchor.Children(func() {
		a.icon(c, "icons/file-diff.svg", t.Px(14), t.TextMuted)
		ui.Text(c, a.currentReviewLabel()).MaxWidth(120).MaxLines(1).FontSize(t.FontSize).FontWeight(500)
		a.icon(c, "icons/chevron-down.svg", t.Px(12), t.TextMuted)
	})
	if anchor.Clicked() {
		s.toggleSourceMenu(c.Now())
		c.AnimationFrame()
	}
	wasOpen := s.sourceMenuOpen
	ui.PopoverBase(c.Key("review-source-menu"), anchor, &s.sourceMenuOpen, func(panel ui.Element) {
		panel.Width(200).Margin(4, 0, 0, 0).PaddingY(t.Space(4)).Background(t.Elevated).Border(1, t.BorderStrong()).Radius(t.Px(12)).Label("Review source menu")
		a.usagePopupShadow(panel)
		ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(1)).Children(func() {
			for i, src := range []workspace.ReviewSource{workspace.ReviewLastTurn, workspace.ReviewUncommitted, workspace.ReviewUnstaged, workspace.ReviewStaged, workspace.ReviewCommitted, workspace.ReviewBranch} {
				if i == 1 || i == 4 {
					ui.Box(c.Key(fmt.Sprintf("review-source-separator-%d", i))).Label("Review source separator").Height(1).Margin(t.Space(4), t.Space(8)).Background(t.Border).AlignSelf(ui.Stretch)
				}
				enabled := src != workspace.ReviewLastTurn || a.Model.LatestTurn > 0
				selected := src == s.source || (s.source == "" && src == workspace.ReviewUncommitted)
				if src == workspace.ReviewLastTurn {
					selected = enabled && s.source == src && a.reviewTurn() == a.Model.LatestTurn
				}
				b := ui.ButtonBase(c.Key(string(src))).Label("Review source "+reviewLabel(src)).Disabled(!enabled).Height(t.Px(23)+2*t.Space(4)).Margin(0, t.Space(4)).Padding(0, t.Space(6)).Gap(t.Space(4)).Radius(t.Px(6)).AlignSelf(ui.Stretch).AlignItems(ui.Center)
				if selected {
					b.Background(t.Selection)
				} else if enabled && b.Hovered() {
					b.Background(t.Hover)
				}
				b.Children(func() {
					color := t.TextSecondary
					if selected {
						color = t.Text
					}
					if !enabled {
						color = t.TextMuted
					}
					ui.Text(c, reviewLabel(src)).FontSize(t.FontSize).MaxLines(1).TextColor(color).Grow(1).MinWidth(0)
					if !enabled {
						ui.Text(c, a.text("sidepane.no_turns_yet", "No turns yet")).FontSize(t.Px(12)).TextColor(t.TextMuted).Shrink(0)
					}
					if selected {
						a.icon(c, "icons/check.svg", t.Px(14), t.AccentText).Label("Review selected source " + reviewLabel(src))
					}
				})
				if b.Clicked() {
					a.selectReviewSource(src, c.Now())
					c.AnimationFrame()
				}
			}
		})
	})
	if wasOpen && !s.sourceMenuOpen {
		s.sourceMenuDismissedAt = c.Now()
	}
}

func (a *App) reviewEmptyDescription() string {
	key, fallback := "sidepane.empty_uncommitted", "The working tree matches HEAD."
	switch a.reviewState.source {
	case workspace.ReviewLastTurn:
		key, fallback = "sidepane.empty_last_turn", "This turn didn't change any files."
	case workspace.ReviewUnstaged:
		key, fallback = "sidepane.empty_unstaged", "Everything is staged."
	case workspace.ReviewStaged:
		key, fallback = "sidepane.empty_staged", "Nothing is staged."
	case workspace.ReviewCommitted:
		key, fallback = "sidepane.empty_committed", "No commits on this branch beyond its base."
	case workspace.ReviewBranch:
		key, fallback = "sidepane.empty_branch", "This branch matches its base branch."
	}
	return a.text(key, fallback)
}

// Source clear_diff discards the old identity immediately, even when the old
// asynchronous collection has not returned. The query itself survives.
func (s *reviewState) clearReviewSnapshot() {
	s.data = workspace.ReviewSnapshot{}
	s.dataIdentity, s.err = "", ""
	s.treeDirs, s.expandedDirs = nil, nil
	s.treeFileKey, s.treeRows = "", nil
	s.treeCursor, s.diffCursor = nil, nil
	s.treeDirty = true
	s.selected = 0
	s.collapsed = map[int]bool{}
	s.list = ui.ListState{}
	s.horizontal = ui.ScrollState{}
	s.vertical = ui.ScrollState{}
	s.headerHandles = nil
	s.stickyPanY = 0
	s.widthCache = reviewWidthCache{}
}
