package orbitui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

// Baseline299a32b: session.rs::on_open_git toggles, while
// pickers.rs::run_command(OpenGit) calls the idempotent open_git destination.
func TestWorkbenchGitShortcutTogglesButPaletteDestinationStaysOpen(t *testing.T) {
	a := fixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyG)
	if a.page != "git" {
		t.Fatal("Git shortcut did not open Git", a.page)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyG)
	if a.page != "chat" {
		t.Fatal("second Git shortcut did not leave Git", a.page)
	}
	for range 2 {
		a.paletteOpen = true
		a.paletteQuery = "git commit push branch history graph changes"
		tt.Frame()
		tt.Key(0, ui.KeyEnter)
		if a.page != "git" || a.paletteOpen {
			t.Fatal("palette destination toggled or failed to dismiss", a.page, a.paletteOpen)
		}
	}
}

// view.rs::on_review_changes -> session.rs::open_review -> show_review;
// source close_git_for_review prevents simultaneous Git and docked Review.
func TestWorkbenchReviewShortcutIsIdempotentAndReplacesGit(t *testing.T) {
	a := fixture(t)
	a.page = "git"
	a.ExplorerVisible = true
	refreshes := 0
	a.Options.OnAction = func(v Action) {
		if v.Kind == "RefreshReview" {
			refreshes++
		}
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyR)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyR)
	if !a.ReviewVisible || a.ExplorerVisible || a.page != "chat" {
		t.Fatal("Review command did not idempotently replace Git/Explorer", a.ReviewVisible, a.ExplorerVisible, a.page)
	}
	if refreshes != 1 {
		t.Fatal("repeated open triggered another explicit load", refreshes)
	}
}

func TestWorkbenchGitDestinationClosesReviewAndRestoresDockWidth(t *testing.T) {
	a, tt := mountedReview(t)
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	a.reviewState.sourceMenuOpen = true
	a.command("OpenGit", "")
	if a.page != "git" || a.ReviewVisible || a.reviewState.layout.full || a.reviewState.sourceMenuOpen {
		t.Fatal("Git did not use source close lifecycle", a.page, a.ReviewVisible, a.reviewState.layout.full)
	}
}

// session.rs::on_git_tab is a no-op when Git is not open, including palette.
func TestWorkbenchGitTabCommandsDoNotNavigateFromChat(t *testing.T) {
	for _, id := range []string{"GitTabChanges", "GitTabHistory", "GitTabGraph", "GitTabStashes", "GitTabIssues", "GitTabPulls"} {
		t.Run(id, func(t *testing.T) {
			a := fixture(t)
			calls := 0
			a.Options.OnAction = func(Action) { calls++ }
			a.command(id, "")
			if a.page != "chat" || calls != 0 {
				t.Fatal("closed Git tab command changed chat", a.page, calls)
			}
			a.page = "git"
			a.command(id, "")
			if a.page != "git" || calls != 1 {
				t.Fatal("open Git tab did not dispatch once", a.page, calls)
			}
		})
	}
}

func TestWorkbenchGitTabShortcutOnlyDispatchesWhileGitVisible(t *testing.T) {
	a := fixture(t)
	calls := 0
	a.Options.OnAction = func(v Action) {
		if v.Kind == "GitTabHistory" {
			calls++
		}
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Alt, ui.Key2)
	if a.page != "chat" || calls != 0 {
		t.Fatal("tab shortcut escaped chat", a.page, calls)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyG)
	tt.Key(ui.Cmd|ui.Alt, ui.Key2)
	if a.page != "git" || calls != 1 {
		t.Fatal("tab shortcut did not reach Git", a.page, calls)
	}
}

func TestWorkbenchModalStillBlocksGitShortcutAndNativeAction(t *testing.T) {
	a := fixture(t)
	a.Options.HasModal = func() bool { return true }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	tt.Key(ui.Cmd|ui.Shift, ui.KeyG)
	a.nativeCommand("OpenGit")
	if a.page != "chat" {
		t.Fatal("Git action escaped modal guard", a.page)
	}
}

// view.rs::on_review_close returns focus only. The first Escape must not
// close the pane or abort Pi; a second Escape from the composer remains abort.
func TestWorkbenchReviewTreeEscapeOnlyReturnsComposerFocus(t *testing.T) {
	a, tt := mountedReview(t)
	aborts := 0
	a.Options.OnAction = func(v Action) {
		if v.Kind == "AbortRun" {
			aborts++
		}
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if !a.ReviewVisible || !tt.Focused("Message") || aborts != 0 {
		t.Fatal("Review Escape changed visibility or aborted instead of focusing", a.ReviewVisible, tt.Focused("Message"), aborts)
	}
	tt.Key(0, ui.KeyEscape)
	if aborts != 1 {
		t.Fatal("composer Escape did not retain genuine abort", aborts)
	}
}

func TestWorkbenchReviewHeaderRemainsToggleAndCloseResetsFullWidth(t *testing.T) {
	a, tt := mountedReview(t)
	if err := tt.Click("Review changes"); err != nil {
		t.Fatal(err)
	}
	if a.ReviewVisible {
		t.Fatal("header toggle did not hide Review")
	}
	if err := tt.Click("Review changes"); err != nil {
		t.Fatal(err)
	}
	if !a.ReviewVisible {
		t.Fatal("header toggle did not reopen Review")
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Close review"); err != nil {
		t.Fatal(err)
	}
	if a.ReviewVisible || a.reviewState.layout.full {
		t.Fatal("explicit close retained full width")
	}
}

func TestWorkbenchExplicitCloseRestoresViewportClampedNormalDock(t *testing.T) {
	a, tt := mountedReview(t)
	a.ReviewWidth = 700
	tt.Frame()
	tt.SetSize(800, 922)
	if a.reviewPaneWidth() != 552 || a.reviewState.layout.restore != 700 {
		t.Fatal("narrow fixture did not retain restore point", a.reviewPaneWidth(), a.reviewState.layout.restore)
	}
	tt.SetSize(1470, 922)
	if a.reviewPaneWidth() != 552 {
		t.Fatal("fixture unexpectedly restored before close")
	}
	if err := tt.Click("Close review"); err != nil {
		t.Fatal(err)
	}
	if a.reviewState.layout.width != 700 || a.reviewState.layout.full {
		t.Fatal("explicit close did not restore ordinary clamped width", a.reviewState.layout)
	}
	a.command("ReviewChanges", "")
	tt.Frame()
	if a.reviewPaneWidth() != 700 || a.Options.Store.Layout().SidepaneWidth != 0 {
		t.Fatal("reopen lost restore point or persisted transient width", a.reviewPaneWidth())
	}
}
