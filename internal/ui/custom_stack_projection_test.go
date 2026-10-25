package orbitui

import (
	"testing"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

// This verifies foreground projection/action routing, not the pixel composition
// of all underlying source layers or the external component's render protocol.
func TestCustomStackProjectionNewestCustomAboveGeneric(t *testing.T) {
	generic := agent.PendingDialog{ID: "generic", Method: "input", Title: "Generic"}
	older := agent.PendingDialog{ID: "older", Method: "custom", Title: "Older custom", Lines: []string{"older frame"}}
	newer := agent.PendingDialog{ID: "newer", Method: "custom", Title: "Newer custom", Lines: []string{"newer frame"}}
	for _, rows := range [][]agent.PendingDialog{{generic, older, newer}, {older, newer, generic}, {older, generic, newer}} {
		got := adaptSnapshot(agent.Snapshot{Dialogs: rows}, "", nil, nil)
		if got.Dialog == nil || got.Dialog.ID != "newer" || got.Dialog.Lines[0] != "newer frame" {
			t.Fatalf("foreground = %#v", got.Dialog)
		}
	}
	a := fixture(t)
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	rows := []agent.PendingDialog{older, newer, generic}
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Dialogs: rows}, "", nil, nil))
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if len(actions) != 1 || actions[0].Kind != "RespondDialog" || actions[0].Data["id"] != "newer" {
		t.Fatalf("wrong top response: %#v", actions)
	}
	// The resulting Pi snapshot reveals the previous custom surface first.
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Dialogs: []agent.PendingDialog{older, generic}}, "", nil, nil))
	tt.Frame()
	if a.Model.Dialog.ID != "older" {
		t.Fatalf("revealed %s", a.Model.Dialog.ID)
	}
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{Dialogs: []agent.PendingDialog{generic}}, "", nil, nil))
	tt.Frame()
	if a.Model.Dialog.ID != "generic" {
		t.Fatalf("generic lost: %#v", a.Model.Dialog)
	}
	a.SetSnapshot(adaptSnapshot(agent.Snapshot{}, "", nil, nil))
	tt.Frame()
	if a.Model.Dialog != nil {
		t.Fatal("closed custom remained active")
	}
}
