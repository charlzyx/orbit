package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"slices"
	"testing"
)

func TestSessionMenuSourceActionsAndExactTargets(t *testing.T) {
	a := sourceSidebarFixture(t)
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{a.Tr("sidebar.unpin_session"), a.Tr("sidebar.archive_session"), a.Tr("sidebar.copy_path"), a.Tr("sidebar.reveal_in_finder")} {
		if !slices.Contains(tt.Menu(), want) {
			t.Fatalf("missing %q: %v", want, tt.Menu())
		}
	}
	if err := tt.ChooseMenuItem(a.Tr("sidebar.copy_path")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "CopyText" || actions[0].Value != "/store/pin.jsonl" {
		t.Fatal(actions)
	}
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy session ID"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 2 || actions[1].Kind != "CopyText" || actions[1].Value != "pin" {
		t.Fatal(actions)
	}
}
func TestSessionMenuHidesDeleteForParkedRuntime(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Runtime.Parked = []agent.RuntimeProcess{{Key: "parked", SessionPath: "/store/pin.jsonl", Alive: true}}
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	for _, label := range []string{"Delete…", a.Tr("sidebar.delete_session_menu")} {
		if slices.Contains(tt.Menu(), label) {
			t.Fatal("live runtime exposes deletion", tt.Menu())
		}
	}
}

func TestSessionMenuUpdatesPinArchiveAndPreservesConfirmation(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.Sessions = []SessionRow{a.Model.Sessions[8], a.Model.Sessions[7]}
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	for _, label := range []string{a.Tr("sidebar.unpin_session"), a.Tr("sidebar.pin_session")} {
		if err := tt.RightClick("pin task"); err != nil {
			t.Fatal(err)
		}
		if err := tt.ChooseMenuItem(label); err != nil {
			t.Fatal(err)
		}
		tt.Frame()
		if len(actions) == 0 || actions[len(actions)-1].Kind != "PinSession" || actions[len(actions)-1].Value != "/store/pin.jsonl" {
			t.Fatal(actions)
		}
		a.Model.Sessions[1].Pinned = !a.Model.Sessions[1].Pinned
		tt.Frame()
	}
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.Frame()
	if len(actions) != 2 {
		t.Fatal("dismiss dispatched", actions)
	}
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem(a.Tr("sidebar.delete_session_menu")); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 2 {
		t.Fatal("delete bypassed confirmation", actions)
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 2 {
		t.Fatal("cancel deleted", actions)
	}
}
