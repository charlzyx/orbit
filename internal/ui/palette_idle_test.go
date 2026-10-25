package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestPaletteAbortSourceAvailabilityMatrix(t *testing.T) {
	cases := []struct {
		name    string
		s       agent.Snapshot
		visible bool
	}{
		{name: "idle"},
		{name: "running", s: agent.Snapshot{Running: true}, visible: true},
		// Manager.Snapshot includes awaiting in Running; use that public projection,
		// not a fabricated direct access to the manager's private session state.
		{name: "awaiting projected running", s: agent.Snapshot{Running: true, Status: "Waiting for response"}, visible: true},
		{name: "compaction only", s: agent.Snapshot{Compacting: true}},
		{name: "running compaction", s: agent.Snapshot{Running: true, Compacting: true}, visible: true},
		{name: "retry only", s: agent.Snapshot{Retrying: true}},
		{name: "running retry", s: agent.Snapshot{Running: true, Retrying: true}, visible: true},
		{name: "no current session idle"},
		{name: "no current session running", s: agent.Snapshot{Running: true}, visible: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := fixture(t)
			a.SetSnapshot(adaptSnapshot(tc.s, "/workspace/fixture", nil, nil))
			a.paletteOpen = true
			a.paletteQuery = "abort stop cancel run agent working"
			var got []Action
			a.Options.OnAction = func(v Action) { got = append(got, v) }
			tt := ui.NewTester(a.View, 1222, 878)
			tt.Frame()
			_, visible := tt.Find(a.Tr("command_palette.abort_run"))
			if visible != tc.visible {
				t.Fatal("source availability", visible, tc.visible)
			}
			if visible {
				if err := tt.Click(a.Tr("command_palette.abort_run")); err != nil {
					t.Fatal(err)
				}
				tt.Frame()
				if len(got) != 1 || got[0].Kind != "AbortRun" {
					t.Fatal(got)
				}
			}
		})
	}
}
func TestPaletteAbortRunningSettlesWhilePaletteOpen(t *testing.T) {
	a := fixture(t)
	a.Model.Running = true
	a.paletteOpen = true
	a.paletteQuery = "abort stop cancel run agent working"
	tt := ui.NewTester(a.View, 1222, 878)
	tt.Frame()
	if _, ok := tt.Find(a.Tr("command_palette.abort_run")); !ok {
		t.Fatal("running action missing")
	}
	s := a.Model
	s.Running = false
	a.SetSnapshot(s)
	tt.Frame()
	if _, ok := tt.Find(a.Tr("command_palette.abort_run")); ok {
		t.Fatal("settled action still offered")
	}
}
func TestPaletteAbortFilterDoesNotChangeEscapePageOrControllerGuard(t *testing.T) {
	for _, page := range []string{"settings", "worktrees"} {
		t.Run(page, func(t *testing.T) {
			a := fixture(t)
			var got []Action
			a.Options.OnAction = func(v Action) { got = append(got, v) }
			a.Navigate(page)
			tt := ui.NewTester(a.View, 1222, 878)
			tt.Frame()
			tt.Key(0, ui.KeyEscape)
			tt.Frame()
			if a.page != "chat" || len(got) != 0 {
				t.Fatal("idle Escape page behavior changed", a.page, got)
			}
		})
	}
	// Nil Agent is intentional: an idle Controller must not call into Abort.
	a := fixture(t)
	x := &Controller{UI: a}
	x.Action(Action{Kind: "AbortRun"})
}
