package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
	"time"
)

func mountedComposerMode(t *testing.T) (*App, *ui.Tester, *[]Action) {
	t.Helper()
	a := fixture(t)
	a.access = "supervised"
	a.workflow = "build"
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	return a, tt, &actions
}
func modeClick(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
}
func TestComposerModePaletteFiniteChoicesKeyboardAndFocus(t *testing.T) {
	a, tt, actions := mountedComposerMode(t)
	a.paletteOpen = true
	a.paletteQuery = "access"
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if a.promptOpen || a.composerMode == nil || a.composerMode.kind != "access" {
		t.Fatal("palette did not open finite picker")
	}
	for _, v := range []string{"Supervised", "Auto-accept edits", "Full access"} {
		if !tt.HasText(v) {
			t.Fatal("missing finite option", v)
		}
	}
	tt.Key(0, ui.KeyUp)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(*actions) != 1 || (*actions)[0].Value != "full-access" || a.access != "supervised" {
		t.Fatal("wrap/dispatch or snapshot projection", *actions, a.access)
	}
	if !tt.Focused("Message") {
		t.Fatal("composer focus not restored")
	}
	modeClick(t, tt, "Build")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(*actions) != 2 || (*actions)[1].Kind != "SetWorkflow" || (*actions)[1].Value != "ask" || a.workflow != "build" {
		t.Fatal(*actions, a.workflow)
	}
	modeClick(t, tt, "Build")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.composerMode != nil || len(*actions) != 2 {
		t.Fatal("Escape applied mode")
	}
}
func TestComposerModeHiddenCommandDoesNotCreateGlobalPrompt(t *testing.T) {
	a, tt, _ := mountedComposerMode(t)
	a.Navigate("settings")
	a.command("ChooseWorkflowMode", "")
	tt.Frame()
	if a.composerMode == nil || a.promptOpen || a.page != "settings" {
		t.Fatal("hidden command route")
	}
	if _, ok := tt.Find("Choose Workflow Mode"); ok {
		t.Fatal("global popup on settings")
	}
	a.Navigate("chat")
	tt.Frame()
	if _, ok := tt.Find("Choose Workflow Mode"); !ok {
		t.Fatal("same-owner intent not resumed")
	}
}
func TestComposerModeCompactRetentionAndOwnerCancellation(t *testing.T) {
	a := fixture(t)
	a.Options.OnAction = func(Action) {}
	a.command("ChooseAccess", "")
	narrow := ui.NewTester(a.View, 800, 800)
	narrow.Frame()
	if _, ok := narrow.Find("Choose Access Mode"); ok {
		t.Fatal("compact popup mounted")
	}
	if a.composerMode == nil {
		t.Fatal("compact lost intent")
	}
	next := a.Model
	next.RuntimeID = "new"
	a.SetSnapshot(next)
	if a.composerMode != nil {
		t.Fatal("new owner retained hidden intent")
	}
}
func TestComposerModeHeldChoiceAndWorkspaceABA(t *testing.T) {
	a, tt, actions := mountedComposerMode(t)
	modeClick(t, tt, "Supervised")
	r, ok := tt.Find("Auto-accept edits")
	if !ok {
		t.Fatal("missing choice")
	}
	tt.Press(r.X+r.W/2, r.Y+r.H/2)
	old := a.Model
	next := old
	next.Workspace = "other"
	a.SetSnapshot(next)
	a.SetSnapshot(old)
	a.command("ChooseAccess", "")
	tt.Frame()
	tt.Release(r.X+r.W/2, r.Y+r.H/2)
	tt.Frame()
	if len(*actions) != 0 {
		t.Fatal("old held choice retargeted", *actions)
	}
	modeClick(t, tt, "Auto-accept edits")
	if len(*actions) != 1 {
		t.Fatal("fresh choice failed")
	}
}
func TestComposerModeCompetitionAndUnknownValue(t *testing.T) {
	a, tt, actions := mountedComposerMode(t)
	modeClick(t, tt, "Supervised")
	s := a.composerMode
	a.chooseComposerMode(s, "bypass")
	if len(*actions) != 0 || a.composerMode != s {
		t.Fatal("unknown accepted")
	}
	a.command("ChooseWorkflowMode", "")
	if a.composerMode == s || a.composerMode.kind != "workflow" {
		t.Fatal("competing mode not exclusive")
	}
	a.chooseComposerMode(s, "full-access")
	if len(*actions) != 0 {
		t.Fatal("stale mode dispatch")
	}
	modeClick(t, tt, "Add attachment")
	if a.composerMode != nil {
		t.Fatal("Add menu left mode open")
	}
	tt.CloseMenu()
	tt.Frame()
	a.command("ChooseAccess", "")
	a.openModelPicker("model", time.Now())
	if a.composerMode != nil {
		t.Fatal("model picker left mode intent")
	}
}
func TestComposerModeFullReviewRetainsButDoesNotMount(t *testing.T) {
	a, tt, _ := mountedComposerMode(t)
	a.command("ChooseAccess", "")
	a.ReviewVisible = true
	a.reviewState.layout.full = true
	tt.Frame()
	if a.composerMode == nil {
		t.Fatal("hidden review lost intent")
	}
	if _, ok := tt.Find("Choose Access Mode"); ok {
		t.Fatal("picker mounted without composer")
	}
	a.reviewState.layout.full = false
	tt.Frame()
	if _, ok := tt.Find("Choose Access Mode"); !ok {
		t.Fatal("picker did not return")
	}
}

func TestComposerModeImmediateCloseAndReopenWithoutTimeGate(t *testing.T) {
	a, tt, _ := mountedComposerMode(t)
	modeClick(t, tt, "Supervised")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	modeClick(t, tt, "Supervised")
	if a.composerMode == nil {
		t.Fatal("legitimate immediate click suppressed")
	}
	modeClick(t, tt, "Message")
	if a.composerMode != nil {
		t.Fatal("outside click did not dismiss")
	}
	modeClick(t, tt, "Supervised")
	if a.composerMode == nil {
		t.Fatal("immediate reopening after outside dismissal suppressed")
	}
}

func TestComposerModeSameTriggerTogglesWithoutReopen(t *testing.T) {
	a, tt, _ := mountedComposerMode(t)
	modeClick(t, tt, "Supervised")
	modeClick(t, tt, "Supervised")
	if a.composerMode != nil {
		t.Fatal("same trigger should close without reopening")
	}
	modeClick(t, tt, "Supervised")
	if a.composerMode == nil {
		t.Fatal("fresh trigger should open")
	}
}

func TestComposerModeStationaryPointerDoesNotOverrideKeyboard(t *testing.T) {
	a, tt, actions := mountedComposerMode(t)
	modeClick(t, tt, "Build")
	r, ok := tt.Find("Plan")
	if !ok {
		t.Fatal("no Plan row")
	}
	tt.Move(r.X+r.W/2, r.Y+r.H/2)
	tt.Frame()
	if a.composerMode.cursor != 0 {
		t.Fatal("pointer did not select Plan")
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	tt.Frame()
	if a.composerMode.cursor != 1 {
		t.Fatal("stationary pointer overrode Down", a.composerMode.cursor)
	}
	tt.Key(0, ui.KeyUp)
	tt.Frame()
	tt.Frame()
	if a.composerMode.cursor != 0 {
		t.Fatal("Up did not restore Plan")
	}
	tt.Key(0, ui.KeyDown)
	tt.Frame()
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(*actions) != 1 || (*actions)[0].Value != "build" {
		t.Fatal("stationary pointer retargeted keyboard choice", *actions)
	}
}

func TestComposerModeSourceWidthAndLocalizedPopupLabel(t *testing.T) {
	a, tt, _ := mountedComposerMode(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	modeClick(t, tt, "Supervised")
	title := a.Tr("command_palette.choose_access")
	if title == "Choose Access Mode" {
		t.Fatal("fixture did not load Chinese")
	}
	r, ok := tt.Find(title)
	if !ok {
		t.Fatal("missing localized public popup label", title)
	}
	if r.W != 300 {
		t.Fatal("source MODE_POPUP_W differs", r.W)
	}
	if !tt.HasText(a.Tr("access.supervised_hint")) {
		t.Fatal("source localized description missing")
	}
	a.Theme.FontSize = 18
	tt.Frame()
	r, ok = tt.Find(title)
	if !ok || r.W != 300 {
		t.Fatal("source raw popup width changed with font setting", r)
	}
}
