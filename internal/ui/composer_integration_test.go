package orbitui

import (
	"encoding/json"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestNativeComposerCompletionPrecedesSendAndAbort(t *testing.T) {
	for _, mods := range []ui.Modifiers{0, ui.Alt, ui.Cmd | ui.Shift} {
		a := fixture(t)
		a.Model.CommandsRaw = []json.RawMessage{json.RawMessage(`{"name":"review","description":"Review changes","source":"extension"}`), json.RawMessage(`{"name":"reload","description":"Reload extensions","source":"extension"}`)}
		var actions []Action
		a.Options.OnAction = func(x Action) { actions = append(actions, x) }
		tt := ui.NewTester(a.View, 1200, 800)
		if err := tt.Click("Message"); err != nil {
			t.Fatal(err)
		}
		tt.Type("/rev")
		tt.Frame()
		if !a.autocomplete.IsOpen() {
			t.Fatal("source command catalog did not open completion")
		}
		tt.Key(mods, ui.KeyEnter)
		tt.Frame()
		if a.Draft != "/review " || len(actions) != 0 {
			t.Fatalf("completion sent prompt instead: %q %+v", a.Draft, actions)
		}
		tt.Key(0, ui.KeyEnter)
		if len(actions) != 1 || actions[0].Kind != "Submit" {
			t.Fatalf("completed draft not submitted: %+v", actions)
		}
		tt.Type("/re")
		tt.Frame()
		tt.Key(0, ui.KeyEscape)
		if a.autocomplete.IsOpen() || len(actions) != 1 {
			t.Fatal("completion Escape aborted runtime")
		}
	}
}
