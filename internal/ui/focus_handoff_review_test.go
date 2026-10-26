package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestFocusHandoffQueuedOwnerDoesNotApplyOldHideIntent(t *testing.T) {
	a := sourceSidebarFixture(t)
	showNew := false
	other := ""
	a.Options.Overlay = func(c *ui.Context) {
		if showNew {
			ui.TextInput(c, &other).Label("New owner input").AutoFocus().Absolute().Left(900).Top(40).Size(180, 26)
		}
	}
	sidebar := false
	tt := ui.NewTester(func(c *ui.Context) { a.View(c); sidebar = a.sessionList.Handle.FocusWithin(c) }, 1470, 922)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if !sidebar {
		t.Fatal("fixture no actual sidebar focus")
	}
	a.SidebarVisible = false
	a.updates <- func() {
		s := a.Model
		s.RuntimeID = "new-runtime"
		s.SessionID = "new-owner"
		a.SetSnapshot(s)
		showNew = true
	}
	tt.Frame()
	tt.Frame()
	if a.Model.SessionID != "new-owner" {
		t.Fatal("queued owner not adopted")
	}
	tt.Type("new owner text")
	if !tt.Focused("New owner input") || other != "new owner text" || a.Draft != "" {
		t.Fatalf("old hide intent stole new owner: other=%q draft=%q focused=%v", other, a.Draft, tt.Focused("New owner input"))
	}
}
func TestFocusHandoffPaletteDispatchExactlyOnce(t *testing.T) {
	for _, pointer := range []bool{false, true} {
		t.Run(map[bool]string{false: "Enter", true: "pointer"}[pointer], func(t *testing.T) {
			a := sourceSidebarFixture(t)
			var actions []Action
			a.Options.OnAction = func(v Action) { actions = append(actions, v) }
			tt := ui.NewTester(a.View, 1470, 922)
			tt.Key(ui.Cmd, ui.KeyP)
			tt.Type("copy session id")
			tt.Frame()
			if pointer {
				if err := tt.Click(a.Tr("command_palette.copy_session_id")); err != nil {
					t.Fatal(err)
				}
			} else {
				tt.Key(0, ui.KeyEnter)
			}
			tt.Frame()
			tt.Frame()
			if len(actions) != 1 || actions[0].Kind != "CopyText" || actions[0].Value != a.Model.SessionID {
				t.Fatalf("dispatch count/payload %v", actions)
			}
		})
	}
}

func TestFocusHandoffQueuedPagePreservesCurrentInput(t *testing.T) {
	a := sourceSidebarFixture(t)
	showNew := false
	other := ""
	a.Options.Overlay = func(c *ui.Context) {
		if showNew {
			ui.TextInput(c, &other).Label("Persistent new owner input").AutoFocus().Absolute().Left(900).Top(40).Size(180, 26)
		}
	}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	a.SidebarVisible = false
	a.updates <- func() {
		s := a.Model
		s.RuntimeID = "new-runtime"
		s.SessionID = "new-owner"
		a.SetSnapshot(s)
		a.Navigate("settings")
		showNew = true
	}
	tt.Frame()
	tt.Frame()
	if !tt.Focused("Persistent new owner input") {
		t.Fatal("new owner input not focused")
	}
	tt.Type("settings input")
	if a.page != "settings" || other != "settings input" || a.Draft != "" {
		t.Fatalf("queued page lost current input: page=%q other=%q draft=%q", a.page, other, a.Draft)
	}
}

func TestFocusHandoffQueuedPageLeavesNoPendingHandleRequest(t *testing.T) {
	for _, mode := range []string{"candidate", "explicit Focus positive", "CancelFocus negative"} {
		t.Run(mode, func(t *testing.T) { focusHandoffPendingProbe(t, mode) })
	}
}
func focusHandoffPendingProbe(t *testing.T, mode string) {
	a := sourceSidebarFixture(t)
	probe := false
	probeText, other := "", ""
	sidebar := false
	tt := ui.NewTester(func(c *ui.Context) {
		if probe {
			ui.TextInput(c, &probeText).Bind(&a.composer).Label("Non autofocus composer probe").Size(240, 30)
		} else {
			a.View(c)
			sidebar = a.sessionList.Handle.FocusWithin(c)
		}
		ui.TextInput(c.Key("persistent-observer"), &other).Label("Persistent focus observer").Absolute().Left(900).Top(40).Size(180, 26)
	}, 1470, 922)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	tt.Frame()
	if !sidebar {
		t.Fatal("actual sidebar not focused")
	}
	a.composer.CancelFocus() // eliminate any unrelated fixture-construction request before the tested event
	a.SidebarVisible = false
	a.updates <- func() {
		s := a.Model
		s.RuntimeID = "replacement-runtime"
		s.SessionID = "replacement-session"
		a.SetSnapshot(s)
		a.Navigate("settings")
	}
	tt.Frame()
	tt.Frame()
	if err := tt.Click("Persistent focus observer"); err != nil {
		t.Fatal(err)
	}
	if !tt.Focused("Persistent focus observer") {
		t.Fatal("observer not actually focused")
	}
	// Isolate public Handle's pending-request behavior from App's deliberate
	// default AutoFocus on returning to Chat. No private toolkit state is read.
	if mode == "explicit Focus positive" {
		a.composer.Focus()
	} else if mode == "CancelFocus negative" {
		a.composer.CancelFocus()
	}
	probe = true
	tt.Frame()
	tt.Frame()
	tt.Type("observer text")
	if mode == "explicit Focus positive" {
		if probeText != "observer text" || other != "" || !tt.Focused("Non autofocus composer probe") {
			t.Fatalf("positive focus control not observed: probe=%q other=%q", probeText, other)
		}
		return
	}
	if probeText != "" || other != "observer text" || !tt.Focused("Persistent focus observer") {
		t.Fatalf("old focus request survived owner/page change: probe=%q other=%q", probeText, other)
	}
}

func TestFocusHandoffOffChatPaletteLeavesNoDeferredRequest(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Navigate("settings")
	probe := false
	probeText, other := "", ""
	tt := ui.NewTester(func(c *ui.Context) {
		if probe {
			ui.TextInput(c, &probeText).Bind(&a.composer).Label("Offchat non autofocus probe").Size(240, 30)
		} else {
			a.View(c)
		}
		ui.TextInput(c.Key("persistent-offchat-observer"), &other).Label("Offchat focus observer").Absolute().Left(900).Top(40).Size(180, 26)
	}, 1470, 922)
	a.composer.CancelFocus()
	if err := tt.Click("Offchat focus observer"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Cmd, ui.KeyP)
	tt.Frame()
	if !a.paletteOpen {
		t.Fatal("palette not open")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.page != "settings" {
		t.Fatal("dismiss navigated")
	}
	if err := tt.Click("Offchat focus observer"); err != nil {
		t.Fatal(err)
	}
	s := a.Model
	s.RuntimeID = "new-runtime"
	s.SessionID = "new-owner"
	a.SetSnapshot(s)
	tt.Frame()
	if !tt.Focused("Offchat focus observer") {
		t.Fatal("later explicit new focus not established")
	}
	probe = true
	tt.Frame()
	tt.Frame()
	tt.Type("new observer text")
	if probeText != "" || other != "new observer text" {
		t.Fatalf("offchat palette request outlived explicit focus and owner: probe=%q other=%q", probeText, other)
	}
}

func TestFocusHandoffHiddenComposerSidebarLeavesNoDeferredRequest(t *testing.T) {
	for _, state := range []string{"setup", "full Review"} {
		t.Run(state, func(t *testing.T) {
			a := sourceSidebarFixture(t)
			if state == "setup" {
				a.setup.enabled = true
			} else {
				a.ReviewVisible = true
				a.initReviewLayout()
				a.reviewState.layout.full = true
			}
			probe := false
			sidebar := false
			probeText, other := "", ""
			tt := ui.NewTester(func(c *ui.Context) {
				if probe {
					ui.TextInput(c, &probeText).Bind(&a.composer).Label("Hidden composer probe").Size(240, 30)
				} else {
					a.View(c)
					sidebar = a.sessionList.Handle.FocusWithin(c)
				}
				ui.TextInput(c.Key("hidden-observer"), &other).Label("Hidden focus observer").Absolute().Left(900).Top(40).Size(180, 26)
			}, 1470, 922)
			tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
			tt.Frame()
			if !sidebar {
				t.Fatal("sidebar not focused")
			}
			if state == "setup" && !a.setupBlocked() {
				t.Fatal("setup not blocked")
			}
			if state == "full Review" && !a.reviewFull() {
				t.Fatal("Review not full")
			}
			a.composer.CancelFocus()
			tt.Key(ui.Cmd, ui.KeyB)
			tt.Frame()
			if a.SidebarVisible {
				t.Fatal("sidebar not hidden")
			}
			if err := tt.Click("Hidden focus observer"); err != nil {
				t.Fatal(err)
			}
			probe = true
			tt.Frame()
			tt.Frame()
			tt.Type("observer text")
			if probeText != "" || other != "observer text" {
				t.Fatalf("hidden composer retained sidebar request: probe=%q other=%q", probeText, other)
			}
		})
	}
}
