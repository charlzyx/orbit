package transcriptui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestSelectionMenuOpeningSnapshotAndReopen(t *testing.T) {
	source := "ORIGINAL_MESSAGE"
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Selectable().Size(600, 300)
		scope.Children(func() { ui.Text(c, source) })
		SelectionMenu(c, scope, source, nativeOptions())
	}, 600, 300)
	if err := tt.RightClick(source); err != nil {
		t.Fatal(err)
	}
	source = "APPENDED_MESSAGE"
	tt.Frame()
	tt.Move(590, 290)
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "ORIGINAL_MESSAGE" {
		t.Fatal("replay did not use opening payload", tt.Clipboard())
	}
	if err := tt.RightClick(source); err != nil {
		t.Fatal(err)
	}
	tt.Key(0, ui.KeyEscape)
	if tt.Clipboard() != "ORIGINAL_MESSAGE" {
		t.Fatal("cancel performed choice")
	}
	source = "REOPENED_MESSAGE"
	tt.Frame()
	if err := tt.RightClick(source); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != source {
		t.Fatal("reopen retained prior intent", tt.Clipboard())
	}
}
func TestSelectionMenuDisabledAndImmutableSelection(t *testing.T) {
	source := "SELECT_THIS_TEXT"
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Selectable().Size(600, 300)
		scope.Children(func() { ui.Text(c, source) })
		SelectionMenu(c, scope, source, nativeOptions())
	}, 600, 300)
	if err := tt.RightClick(source); err != nil {
		t.Fatal(err)
	}
	tt.SetClipboard("sentinel")
	if err := tt.ChooseMenuItem("Copy Selection"); err == nil {
		t.Fatal("empty selection enabled")
	}
	if tt.Clipboard() != "sentinel" {
		t.Fatal("disabled selection copied")
	}
	tt.CloseMenu()
	r, ok := tt.Find(source)
	if !ok {
		t.Fatal("missing text")
	}
	tt.Press(r.X, r.Y+r.H/2)
	tt.Move(r.X+r.W, r.Y+r.H/2)
	tt.Release(r.X+r.W, r.Y+r.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	captured := tt.Clipboard()
	if captured == "" || captured == "sentinel" {
		t.Fatal("drag selection missing")
	}
	if err := tt.RightClick(source); err != nil {
		t.Fatal(err)
	}
	source = "REPLACED_TEXT"
	tt.Frame()
	tt.Move(590, 290)
	if err := tt.ChooseMenuItem("Copy Selection"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != captured {
		t.Fatalf("selection changed %q→%q", captured, tt.Clipboard())
	}
}

func TestSelectionMenuEmptyMessageRemainsDisabledWithSelection(t *testing.T) {
	const body = "VISIBLE_TOOL_TEXT"
	tt := ui.NewTester(func(c *ui.Context) {
		scope := ui.Column(c).Selectable().Size(600, 300)
		scope.Children(func() { ui.Text(c, body) })
		SelectionMenu(c, scope, "", nativeOptions())
	}, 600, 300)
	r, ok := tt.Find(body)
	if !ok {
		t.Fatal("missing body")
	}
	tt.Press(r.X, r.Y+r.H/2)
	tt.Move(r.X+r.W, r.Y+r.H/2)
	tt.Release(r.X+r.W, r.Y+r.H/2)
	if err := tt.RightClick(body); err != nil {
		t.Fatal(err)
	}
	tt.SetClipboard("sentinel")
	if err := tt.ChooseMenuItem("Copy Message"); err == nil {
		t.Fatal("empty message action enabled")
	}
	if tt.Clipboard() != "sentinel" {
		t.Fatal("disabled message wrote clipboard")
	}
	if err := tt.ChooseMenuItem("Copy Selection"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != body {
		t.Fatal("disabled message shifted selection action", tt.Clipboard())
	}
}
