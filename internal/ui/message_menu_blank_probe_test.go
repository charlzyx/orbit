package orbitui

import (
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestMessageMenuMeasuredBlankDoesNotReusePreviousHover(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(menuAuditSnapshot())
	tt := ui.NewTester(a.View, 1470, 922)
	first, ok := tt.Find("FIRST_MESSAGE")
	if !ok || first.W <= 0 || first.H <= 0 {
		t.Fatal("first bounds absent")
	}
	second, ok := tt.Find("SECOND_MESSAGE")
	if !ok || second.W <= 0 || second.H <= 0 {
		t.Fatal("second bounds absent")
	}
	composer, ok := tt.Find("Composer box")
	if !ok || composer.H <= 0 {
		t.Fatal("composer bounds absent")
	}
	x, y := first.X+first.W/2, first.Y-80
	distance := func(r ui.Rect) float32 { return max(r.X-x, x-r.X-r.W, 0) + max(r.Y-y, y-r.Y-r.H, 0) }
	if y <= 44 || y >= composer.Y || distance(first) < 24 || distance(second) < 24 {
		t.Fatalf("fixture not blank first=%v second=%v composer=%v point=%v,%v", first, second, composer, x, y)
	}
	t.Logf("actual first=%v second=%v composer=%v blank=(%v,%v), distances=(%v,%v)", first, second, composer, x, y, distance(first), distance(second))
	if err := tt.RightClick("FIRST_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "FIRST_MESSAGE" {
		t.Fatal("positive hit not first")
	}
	tt.SetClipboard("UNCHANGED_SENTINEL")
	tt.Move(x, y)
	tt.RightClickAt(x, y)
	if len(tt.Menu()) != 0 {
		_ = tt.ChooseMenuItem("Copy Message")
		t.Fatalf("blank opened menu / stale payload: clipboard=%q", tt.Clipboard())
	}
	if tt.Clipboard() != "UNCHANGED_SENTINEL" {
		t.Fatal("blank changed clipboard")
	}
}

func TestMessageMenuBlankPreservesActualSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(menuAuditSnapshot())
	tt := ui.NewTester(a.View, 1470, 922)
	first, ok := tt.Find("FIRST_MESSAGE")
	if !ok || first.H <= 0 {
		t.Fatal("fixture text missing")
	}
	tt.Press(first.X+2, first.Y+first.H/2)
	tt.Move(first.X+70, first.Y+first.H/2)
	tt.Release(first.X+70, first.Y+first.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	before := tt.Clipboard()
	if before == "" {
		t.Fatal("fixture actual selection empty")
	}
	tt.SetClipboard("BLANK_SENTINEL")
	tt.RightClickAt(first.X+first.W/2, first.Y-80)
	tt.Key(ui.Cmd, ui.KeyC)
	if tt.Clipboard() != before {
		t.Fatalf("blank changed selected Copy: before=%q after=%q", before, tt.Clipboard())
	}
}

func TestMessageMenuCancelBlankAndOtherTextSelection(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(menuAuditSnapshot())
	tt := ui.NewTester(a.View, 1470, 922)
	first, ok := tt.Find("FIRST_MESSAGE")
	if !ok {
		t.Fatal("first missing")
	}
	tt.Press(first.X+2, first.Y+first.H/2)
	tt.Move(first.X+70, first.Y+first.H/2)
	tt.Release(first.X+70, first.Y+first.H/2)
	tt.Key(ui.Cmd, ui.KeyC)
	selected := tt.Clipboard()
	if selected == "" {
		t.Fatal("selection missing")
	}
	if err := tt.RightClick("SECOND_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy Selection"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != selected {
		t.Fatalf("other-text right click changed selection: %q -> %q", selected, tt.Clipboard())
	}
	if err := tt.RightClick("SECOND_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("Copy Message"); err != nil {
		t.Fatal(err)
	}
	if tt.Clipboard() != "SECOND_MESSAGE" {
		t.Fatalf("wrong clicked message: %q", tt.Clipboard())
	}
	if err := tt.RightClick("FIRST_MESSAGE"); err != nil {
		t.Fatal(err)
	}
	tt.CloseMenu()
	tt.SetClipboard("CANCEL_SENTINEL")
	tt.RightClickAt(first.X+first.W/2, first.Y-80)
	if len(tt.Menu()) != 0 {
		t.Fatal("cancel then blank reused menu")
	}
	if tt.Clipboard() != "CANCEL_SENTINEL" {
		t.Fatal("cancel then blank copied old intent")
	}
}
