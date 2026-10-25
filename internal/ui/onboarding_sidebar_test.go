package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestOnboardingSetupReplacesMainNotSidebar(t *testing.T) {
	a := fixture(t)
	a.SidebarVisible = true
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 800)
	before, ok := tt.Find("Sidebar wordmark")
	if !ok {
		t.Fatal("fixture sidebar absent")
	}
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	tt.Frame()
	after, ok := tt.Find("Sidebar wordmark")
	if !ok {
		t.Fatal("required-dependency gate incorrectly removed source sidebar")
	}
	if before != after {
		t.Fatal("setup reparented sidebar", before, after)
	}
	if _, ok := tt.Find("Resize sidebar"); !ok {
		t.Fatal("setup removed source sidebar resize")
	}
	title, ok := tt.Find(a.Tr("view.set_up_orbit"))
	if !ok || title.X < a.SidebarWidth {
		t.Fatal("setup is not in main column", title, a.SidebarWidth)
	}
	if _, ok := tt.Find("Message"); ok {
		t.Fatal("main composer mounted beneath setup")
	}
	if err := tt.Click(a.text("sidebar.new_session", "New Task")); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "NewSession" {
		t.Fatal("visible source task action lost", actions)
	}
	a.SidebarVisible = false
	tt.Frame()
	if _, ok := tt.Find("Sidebar wordmark"); ok {
		t.Fatal("collapsed sidebar forced open by setup")
	}
	if _, ok := tt.Find(a.Tr("view.set_up_orbit")); !ok {
		t.Fatal("sidebar collapse removed setup")
	}
	a.SidebarVisible = true
	a.command("OpenSettings", "")
	tt.Frame()
	if _, ok := tt.Find("Sidebar wordmark"); ok {
		t.Fatal("Settings must own entire surface without session sidebar")
	}
	a.Navigate("chat")
	a.setup.dependencies = setupDeps(true, true, true)
	tt.Frame()
	if _, ok := tt.Find("Message"); !ok {
		t.Fatal("ready main composer not restored")
	}
	after, ok = tt.Find("Sidebar wordmark")
	if !ok || after != before {
		t.Fatal("ready transition changed sidebar ancestry", before, after)
	}
}

func TestOnboardingSidebarFocusSurvivesMainReplacement(t *testing.T) {
	a := sourceSidebarFixture(t)
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Key(ui.Cmd|ui.Shift, ui.KeyB)
	if a.sessionCursor != 1 {
		t.Fatal("fixture sidebar focus missing")
	}
	a.setup = onboardingState{enabled: true, loaded: true, dependencies: setupDeps(false, true, true)}
	tt.Frame()
	tt.Key(0, ui.KeyDown)
	if a.sessionCursor != 2 {
		t.Fatal("main replacement stole sidebar keyboard focus", a.sessionCursor)
	}
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 1 || actions[0].Kind != "SwitchSession" || actions[0].Value != "/store/pin.jsonl" {
		t.Fatal("setup sidebar navigation changed target", actions)
	}
	if _, ok := tt.Find("Message"); ok {
		t.Fatal("sidebar selection bypassed setup gate")
	}
}
