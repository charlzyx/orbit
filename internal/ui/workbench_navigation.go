package orbitui

// openReviewDock mirrors session.rs::open_review and SidePane::show_review:
// opening is idempotent, replaces Git, and excludes the Explorer dock.
func (a *App) openReviewDock() {
	if a.page == "git" {
		a.Navigate("chat")
	}
	wasOpen := a.ReviewVisible
	a.ReviewVisible, a.ExplorerVisible = true, false
	if !wasOpen {
		a.Dispatch(Action{Kind: "RefreshReview"})
	}
}

// Explicit close uses SidePane::close's restore-point semantics. The header's
// toggle is deliberately separate: source toggle does not reset full width.
func (a *App) closeReviewDock() {
	if !a.ReviewVisible {
		return
	}
	a.ReviewVisible = false
	a.reviewState.sourceMenuOpen = false
	a.initReviewLayout()
	s := &a.reviewState.layout
	// Source close restores even a normal dock that was temporarily narrowed
	// by the viewport, unlike leave_full_width's early return when not full.
	s.full = false
	s.width = max(float32(300), s.restore)
}
