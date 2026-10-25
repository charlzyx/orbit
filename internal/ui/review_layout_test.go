package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestNativeReviewFullWidthRestoreAndViewport(t *testing.T) {
	a, tt := mountedReview(t)
	a.ReviewWidth = 500
	a.SidebarVisible = false
	tt.Frame()
	if err := tt.Click("Review file two.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	if !a.SidebarVisible || !a.reviewFull() || a.reviewPaneWidth() != 1222 {
		t.Fatal("full review did not share window with opened sidebar", a.SidebarVisible, a.reviewPaneWidth())
	}
	if _, ok := tt.Find("Message"); ok {
		t.Fatal("composer still mounted under full review")
	}
	if _, ok := tt.Find("New task backdrop"); ok {
		t.Fatal("backdrop remained under full review")
	}
	tt.SetSize(1100, 922)
	if a.reviewPaneWidth() != 852 {
		t.Fatal("full review did not follow viewport", a.reviewPaneWidth())
	}
	a.command("ToggleSidebar", "")
	tt.Frame()
	if a.reviewPaneWidth() != 1100 {
		t.Fatal("full review did not follow sidebar toggle", a.reviewPaneWidth())
	}
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	if a.reviewFull() || a.reviewPaneWidth() != 500 || a.reviewState.selected != 1 {
		t.Fatal("restore lost width or selected file", a.reviewPaneWidth(), a.reviewState.selected)
	}
	if _, ok := tt.Find("Message"); !ok {
		t.Fatal("restore did not return session view")
	}
	if a.Options.Store.Layout().SidepaneWidth != 0 {
		t.Fatal("transient full width was persisted")
	}
}
func TestNativeReviewDragClampsAndPersists(t *testing.T) {
	a, tt := mountedReview(t)
	drag := func(dx float32) {
		t.Helper()
		r, ok := tt.Find("Resize review")
		if !ok {
			t.Fatal("missing resize handle")
		}
		x, y := r.X+r.W/2, r.Y+200
		tt.Press(x, y)
		tt.Move(x+dx, y)
		tt.Release(x+dx, y)
		tt.Frame()
	}
	drag(-100)
	if a.reviewPaneWidth() != 560 {
		t.Fatal("left-edge drag width", a.reviewPaneWidth())
	}
	drainRegressionUpdates(t, a, func() bool { return a.Options.Store.Layout().SidepaneWidth == 560 })
	drag(-1000)
	if a.reviewPaneWidth() != 900 {
		t.Fatal("drag exceeded source900 cap", a.reviewPaneWidth())
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	if a.reviewPaneWidth() <= 900 {
		t.Fatal("full width incorrectly capped like drag")
	}
	drag(20)
	if a.reviewFull() || a.reviewPaneWidth() != 900 {
		t.Fatal("drag did not leave full width at cap", a.reviewPaneWidth())
	}
	tt.SetSize(800, 700)
	drag(-100)
	if a.reviewPaneWidth() != 320 {
		t.Fatal("drag ignored source480px main reserve", a.reviewPaneWidth())
	}
	drag(1000)
	if a.reviewPaneWidth() != 300 {
		t.Fatal("drag below source300 minimum", a.reviewPaneWidth())
	}
}
func TestNativeReviewFeaturePageYieldsAndKeepsRestoreWidth(t *testing.T) {
	a, tt := mountedReview(t)
	a.ReviewWidth = 520
	tt.Frame()
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	a.Navigate("files")
	tt.Frame()
	if a.reviewState.layout.full || a.reviewPaneWidth() != 520 {
		t.Fatal("new feature page did not restore dock width")
	}
	a.Navigate("chat")
	tt.Frame()
	if a.reviewFull() || a.reviewPaneWidth() != 520 {
		t.Fatal("return from feature silently expanded pane")
	}
}
func TestNativeReviewFullWidthRetainsOnlySameDraftSelection(t *testing.T) {
	a, tt := mountedReview(t)
	a.Model.SessionID = "one"
	a.Draft = "alpha 中文 omega"
	tt.Frame()
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	selection := a.reviewState.layout.composer
	if selection.anchor == selection.caret {
		t.Fatal("fixture has no selected text")
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		tt.Frame()
	} // Prune the hidden text-input node, not one-frame hiding.
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	restored := a.reviewState.layout.composer
	if restored.anchor != selection.anchor || restored.caret != selection.caret {
		t.Fatal("full-width cycle lost native selection", selection, restored)
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	a.SetSnapshot(Snapshot{SessionID: "two", Workspace: a.Model.Workspace})
	a.Draft = "a different longer draft"
	tt.Frame()
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	restored = a.reviewState.layout.composer
	if restored.anchor != restored.caret || restored.session != "two" {
		t.Fatal("restored another session's selection", restored)
	}
}

func TestNativeReviewFullWidthDoesNotStealTreeFocus(t *testing.T) {
	a, tt := mountedReview(t)
	if err := tt.Click("Review file one.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		tt.Frame()
	}
	if err := tt.Click("Review file one.go"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if a.reviewState.selected != 1 {
		t.Fatal("restoring full review stole the tree's keyboard focus", a.reviewState.selected)
	}
}

func TestNativeReviewFullWidthKeepsReviewScroll(t *testing.T) {
	a, tt := mountedReview(t)
	a.reviewState.list.ScrollTo(50, ui.Start)
	tt.Frame()
	first, _ := a.reviewState.list.Visible()
	if first != 50 {
		t.Fatal("fixture scroll", first)
	}
	for range 2 {
		if err := tt.Click("Full width review"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if got, _ := a.reviewState.list.Visible(); got != first {
			t.Fatal("expand lost review scroll", got, first)
		}
		if err := tt.Click("Restore review width"); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if got, _ := a.reviewState.list.Visible(); got != first {
			t.Fatal("restore lost review scroll", got, first)
		}
	}
}

func TestNativeReviewDockCoexistsWithFeaturePages(t *testing.T) {
	for _, page := range []string{"files", "git", "worktrees"} {
		t.Run(page, func(t *testing.T) {
			a, tt := mountedReview(t)
			a.Options.Pages = map[string]func(*ui.Context){page: func(c *ui.Context) {
				ui.Column(c).Label("Feature page body").Fill().Children(func() { ui.Button(c, "Feature keyboard target") })
			}}
			a.Navigate(page)
			tt.Frame()
			if _, ok := tt.Find("Review file one.go"); !ok {
				t.Fatal("source dock missing beside", page)
			}
			if _, ok := tt.Find("Feature page body"); !ok {
				t.Fatal("feature renderer missing", page)
			}
			if err := tt.Click("Feature keyboard target"); err != nil {
				t.Fatal(err)
			}
			if err := tt.Click("Full width review"); err != nil {
				t.Fatal(err)
			}
			if !a.reviewFull() {
				t.Fatal("existing feature prevented full-width review")
			}
			if _, ok := tt.Find("Feature page body"); ok {
				t.Fatal("full-width review did not hide existing feature")
			}
			if err := tt.Click("Restore review width"); err != nil {
				t.Fatal(err)
			}
			if _, ok := tt.Find("Feature page body"); !ok {
				t.Fatal("restore lost feature page")
			}
		})
	}
}

func TestNativeReviewDockExcludesSettingsAndUsage(t *testing.T) {
	for _, page := range []string{"settings", "usage"} {
		t.Run(page, func(t *testing.T) {
			a, tt := mountedReview(t)
			a.Navigate(page)
			tt.Frame()
			if _, ok := tt.Find("Review file one.go"); ok {
				t.Fatal("review should not be mounted on", page)
			}
		})
	}
}

func TestNativeReviewFullWidthKeepsComposerUndo(t *testing.T) {
	a, tt := mountedReview(t)
	a.Model.SessionID = "undo-session"
	a.Draft = "original 中文"
	tt.Frame()
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("replacement 中文")
	if a.Draft != "replacement 中文" {
		t.Fatal("fixture edit", a.Draft)
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		tt.Frame()
	}
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != "original 中文" {
		t.Fatal("full-width cycle discarded composer undo history", a.Draft)
	}
	tt.Key(ui.Cmd|ui.Shift, ui.KeyZ)
	if a.Draft != "replacement 中文" {
		t.Fatal("full-width cycle discarded composer redo history", a.Draft)
	}
}

func TestNativeReviewHiddenComposerTakesNoInput(t *testing.T) {
	a, tt := mountedReview(t)
	a.Draft = "keep this draft"
	a.Model.Connected = true
	var actions []Action
	a.Options.OnAction = func(action Action) { actions = append(actions, action) }
	tt.Frame()
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	tt.Type("hidden input")
	tt.SetClipboard("hidden paste")
	tt.Key(ui.Cmd, ui.KeyV)
	for _, mods := range []ui.Modifiers{0, ui.Alt, ui.Cmd, ui.Cmd | ui.Shift} {
		tt.Key(mods, ui.KeyEnter)
	}
	if a.Draft != "keep this draft" {
		t.Fatal("invisible input changed draft", a.Draft)
	}
	for _, action := range actions {
		if action.Kind == "Submit" || action.Kind == "PasteImage" {
			t.Fatal("invisible composer dispatched", action)
		}
	}
	if _, ok := tt.Find("Message"); ok {
		t.Fatal("invisible composer exposed text/accessibility label")
	}
}

func TestNativeReviewFullWidthRestoresFileEditorFocusAndUndo(t *testing.T) {
	a, _, path := mountedFile(t, "file original 中文")
	a.ReviewVisible = true
	a.reviewState.root = a.Model.Workspace
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("File editor"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("file replacement 中文")
	st := a.files.states[path]
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	tt.Type("invisible")
	if st.Buffer.String() != "file replacement 中文" {
		t.Fatal("hidden editor accepted input")
	}
	if err := tt.Click("Restore review width"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if st.Buffer.String() != "file original 中文" {
		t.Fatal("restored file editor lost focus or undo", st.Buffer.String())
	}
}

func TestNativeReviewHiddenContentDoesNotRequestFrames(t *testing.T) {
	a, tt := mountedReview(t)
	a.Prefs.ReduceMotion = false
	a.Model.Messages = []Message{{ID: "streaming", Role: "assistant", Text: "ongoing reply", Streaming: true}}
	tt.Frame()
	if err := tt.Click("Full width review"); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		tt.Frame()
	}
	if tt.FramePending() {
		t.Fatal("invisible transcript/composer still requests animation frames")
	}
}
