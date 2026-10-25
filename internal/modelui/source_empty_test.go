package modelui

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSourceEmptyResourceGeometryAndStateSeparation(t *testing.T) {
	s := fixture(t)
	s.SetSnapshot(Snapshot{Connected: true})
	tt := nativeTester(s, "models", 912, 700)
	tt.Frame()
	saveSourceCapture(t, tt, "models-empty")
	card, ok := tt.Find("Model empty state")
	if !ok {
		t.Fatal("source empty card missing")
	}
	tile, _ := tt.Find("Model empty glyph tile")
	glyph, _ := tt.Find("Model empty glyph")
	title, _ := tt.Find("No models reported")
	if card.W != 912 || tile.W != 64 || tile.H != 64 || glyph.W != 32 || glyph.H != 32 || math.Abs(float64(tile.Y-card.Y-41)) > .1 || math.Abs(float64(tile.X+32-456)) > .1 || math.Abs(float64(title.Y-tile.Y-76)) > .1 {
		t.Fatalf("source empty anatomy: card=%+v tile=%+v glyph=%+v title=%+v", card, tile, glyph, title)
	}
	if card.H < 225 || card.H > 235 {
		t.Fatalf("two-line source empty height: %+v", card)
	}
	s.SetSnapshot(Snapshot{Loading: true})
	tt.Frame()
	if tt.HasText("Model empty state") || !tt.HasText("Loading models…") {
		t.Fatal("loading must not use empty claim/card")
	}
	s.SetSnapshot(Snapshot{Connected: true, Error: "catalog fixture failed"})
	tt.Frame()
	errRect, ok := tt.Find("catalog fixture failed")
	empty, _ := tt.Find("Model empty state")
	if !ok || errRect.Y >= empty.Y {
		t.Fatal("error must remain separate before empty card", errRect, empty)
	}
}

func TestSourceModelNoMatchesReturnsViaRealSearch(t *testing.T) {
	s := fixture(t)
	tt := nativeTester(s, "models", 912, 700)
	click(t, tt, "Search model catalog")
	tt.Type("missing-fixture-model")
	if !tt.HasText("No models match") || !tt.HasText("Model empty glyph tile") {
		t.Fatal("no-match source card missing")
	}
	if len(s.Snapshot.Models) != 4 {
		t.Fatal("filter mutated real catalog")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Key(0, ui.KeyBackspace)
	if tt.HasText("No models match") || !tt.HasText("Select model first / alpha") {
		t.Fatal("cleared search did not restore catalog", s.Query)
	}
}

func saveSourceCapture(t *testing.T, tt *ui.Tester, name string) {
	t.Helper()
	out := os.Getenv("ORBIT_SOURCE_EMPTY_CAPTURE_DIR")
	if out == "" {
		return
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, name+"-cpu.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
