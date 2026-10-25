package orbitui

import (
	"fmt"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestSourceEscapeLeavesFeaturePageBeforeAgentAbort(t *testing.T) {
	for _, page := range []string{"settings", "worktrees"} {
		for _, running := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/running=%v", page, running), func(t *testing.T) {
				a := fixture(t)
				a.Model.Running = running
				a.page = page
				a.Draft = "preserve draft"
				calls := []Action{}
				a.Options.OnAction = func(v Action) { calls = append(calls, v) }
				tt := ui.NewTester(a.View, 1470, 922)
				tt.Frame()
				tt.Key(0, ui.KeyEscape)
				if a.page != "chat" || a.Draft != "preserve draft" || len(calls) != 0 {
					t.Fatalf("Escape affected agent instead of page: %s %q %+v", a.page, a.Draft, calls)
				}
			})
		}
	}
}
func TestSourceSettingsPopupEscapeClosesOnlyPopupBeforePage(t *testing.T) {
	a := fixture(t)
	a.Model.Running = true
	a.page = "settings"
	a.settingsPage = "appearance"
	calls := 0
	a.Options.OnAction = func(Action) { calls++ }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Theme selector"); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if a.page != "settings" || a.nativeSettings.selectMenu.id != "" || calls != 0 {
		t.Fatal("first Escape escaped popup scope or sent action")
	}
	tt.Key(0, ui.KeyEscape)
	if a.page != "chat" || calls != 0 {
		t.Fatal("second Escape did not close Settings without abort")
	}
}
func TestDirectAbortDispatchAlsoRespectsFeaturePage(t *testing.T) {
	x := maintenanceFixture(t)
	x.UI.page = "settings"
	x.UI.Model.Running = true
	x.Action(Action{Kind: "AbortRun"})
	if x.UI.page != "chat" || x.Agent.Snapshot().ConnectionError != "" {
		t.Fatal("direct menu dispatch touched disconnected Pi")
	}
}
