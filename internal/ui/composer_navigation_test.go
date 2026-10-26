package orbitui

import (
	"encoding/json"
	"github.com/egoist/mygo/ui"
	"testing"
)

func composerNavigationFixture(t *testing.T) (*App, *ui.Tester) {
	t.Helper()
	a := railFixture(t)
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.Click("Message"); err != nil {
		t.Fatal(err)
	}
	return a, tt
}
func requireNavigationRow(t *testing.T, a *App, want int) {
	t.Helper()
	got, _ := a.transcript.Visible()
	if got != want {
		t.Fatalf("visible row=%d want %d", got, want)
	}
}
func TestComposerNativeTurnNavigation(t *testing.T) {
	a, tt := composerNavigationFixture(t)
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 20)
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 18)
	tt.Key(ui.Cmd, ui.KeyDown)
	requireNavigationRow(t, a, 20)
	a.transcript.ScrollTo(0, ui.Start)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 0)
	tt.Key(ui.Cmd, ui.KeyDown)
	requireNavigationRow(t, a, 2)
	// The final user row clamps against the true bottom instead of reaching top.
	a.transcript.ScrollTo(22, ui.Start)
	tt.Frame()
	last, _ := a.transcript.Visible()
	a.transcript.ScrollTo(20, ui.Start)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyDown)
	requireNavigationRow(t, a, last)
	tt.Key(ui.Cmd, ui.KeyDown)
	requireNavigationRow(t, a, last)
}
func TestComposerTurnNavigationPreservesDraftSelectionUndo(t *testing.T) {
	a, tt := composerNavigationFixture(t)
	draft := "甲😀乙\n第二行tail"
	tt.Type(draft)
	for range 4 {
		tt.Key(ui.Shift, ui.KeyLeft)
	}
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 20)
	if a.Draft != draft {
		t.Fatal("navigation changed draft", a.Draft)
	}
	tt.Type("X")
	if a.Draft != "甲😀乙\n第二行X" {
		t.Fatal("navigation lost active selection", a.Draft)
	}
	tt.Key(ui.Cmd, ui.KeyZ)
	if a.Draft != draft {
		t.Fatal("navigation changed undo history", a.Draft)
	}
}
func TestComposerTurnNavigationIMEAndPlainArrowNegativeControls(t *testing.T) {
	a, tt := composerNavigationFixture(t)
	tt.Type("first\nsecond")
	a.transcript.ScrollTo(8, ui.Start)
	tt.Frame()
	before, _ := a.transcript.Visible()
	tt.Compose("候", 1)
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, before)
	tt.Compose("", 0)
	for _, k := range []struct {
		mods ui.Modifiers
		key  ui.Key
	}{{0, ui.KeyUp}, {0, ui.KeyDown}, {0, ui.KeyHome}, {0, ui.KeyEnd}, {ui.Cmd, ui.KeyHome}, {ui.Cmd, ui.KeyEnd}} {
		tt.Key(k.mods, k.key)
		requireNavigationRow(t, a, before)
	}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt.Compose("候", 1)
	tt.Key(0, ui.KeyEnter)
	if len(actions) != 0 {
		t.Fatal("IME Enter submitted", actions)
	}
}
func TestTurnNavigationOutsideComposerUsesGlobalShortcut(t *testing.T) {
	a := railFixture(t)
	tt := ui.NewTester(func(c *ui.Context) { a.messages(c); a.shortcuts(c) }, 1200, 600)
	tt.Frame()
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 20)
	tt.Key(ui.Cmd, ui.KeyDown)
	tt.Frame()
	first, _ := a.transcript.Visible()
	if first <= 20 {
		t.Fatal("global next turn did not advance", first)
	}
}
func TestComposerTurnNavigationAutocompletePrecedence(t *testing.T) {
	a, tt := composerNavigationFixture(t)
	a.Model.CommandsRaw = []json.RawMessage{json.RawMessage(`{"name":"review","description":"Review changes","source":"extension"}`), json.RawMessage(`{"name":"reload","description":"Reload extensions","source":"extension"}`)}
	tt.Type("/re")
	tt.Frame()
	if !a.autocomplete.IsOpen() {
		t.Fatal("autocomplete not mounted")
	}
	before, _ := a.transcript.Visible()
	tt.Key(0, ui.KeyDown)
	requireNavigationRow(t, a, before)
	tt.Key(ui.Cmd, ui.KeyUp)
	requireNavigationRow(t, a, 20)
	if a.Draft != "/re" {
		t.Fatal("navigation accepted or changed completion", a.Draft)
	}
}
