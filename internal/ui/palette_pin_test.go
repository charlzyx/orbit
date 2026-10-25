package orbitui

import (
	"github.com/egoist/mygo/ui"
	"slices"
	"testing"
)

func TestPalettePinTargetsActiveAndTogglesOnlyItsStoredPath(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Model.SessionPath = "/store/active.jsonl"
	a.Model.SessionID = "active"
	if err := a.Options.Store.SetSession(false, "/store/other.jsonl", true); err != nil {
		t.Fatal(err)
	}
	x := &Controller{UI: a, Store: a.Options.Store, pinned: map[string]bool{}, archived: map[string]bool{}}
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v); x.Action(v) }
	tt := ui.NewTester(a.View, 1222, 878)
	for turn := 0; turn < 2; turn++ {
		a.paletteOpen = true
		a.paletteQuery = "pin session"
		tt.Frame()
		if turn == 0 {
			if err := tt.Click(a.Tr("sidebar.pin_session")); err != nil {
				t.Fatal(err)
			}
		} else {
			tt.Key(0, ui.KeyEnter)
		}
		tt.Frame()
		saved := a.Options.Store.SessionSet(false)
		if len(got) != turn+1 || got[turn].Value != "/store/active.jsonl" || got[turn].SessionID != "active" || slices.Contains(saved, "") || !slices.Contains(saved, "/store/other.jsonl") || slices.Contains(saved, "/store/active.jsonl") != (turn == 0) {
			t.Fatalf("wrong target/toggle actions=%+v saved=%q", got, saved)
		}
	}
}
func TestPalettePinWithoutSavedPathDoesNotDispatch(t *testing.T) {
	a := fixture(t)
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v) }
	a.Model.SessionID = "unsaved"
	a.Model.SessionPath = ""
	a.command("PinSession", "")
	if len(got) != 0 || len(a.Options.Store.SessionSet(false)) != 0 {
		t.Fatal("pathless pin must be inert", got)
	}
}
