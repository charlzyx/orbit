package resourceui

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
)

func emptyResourceFixture(t *testing.T) *State {
	t.Helper()
	s := New(preferences.New(t.TempDir()), t.TempDir(), Options{Offline: true})
	s.Refresh()
	waitState(t, s)
	t.Cleanup(s.Close)
	return s
}

func TestSourcePluginEmptyGeometryAndLoadingIsolation(t *testing.T) {
	s := emptyResourceFixture(t)
	tt := fixedNavTester(s, "plugins", 1470, 922)
	tt.Frame()
	saveSourceCapture(t, tt, "plugins-empty")
	card, ok := tt.Find("Plugin empty state")
	tile, _ := tt.Find("Plugin empty glyph tile")
	title, _ := tt.Find("No plugins installed")
	if !ok || card.X != 399 || card.W != 912 || math.Abs(float64(card.Y-238)) > 2 || tile.W != 64 || tile.H != 64 || math.Abs(float64(tile.Y-card.Y-41)) > .1 || math.Abs(float64(tile.X+32-855)) > .1 || math.Abs(float64(title.Y-tile.Y-76)) > .1 || card.H < 210 || card.H > 214 {
		t.Fatalf("source plugin empty geometry: card=%+v tile=%+v title=%+v", card, tile, title)
	}
	t.Logf("Plugin empty=%+v tile=%+v title=%+v", card, tile, title)
	s.Busy = true
	s.loaded = false
	s.activity = "Loading resource fixture…"
	tt.Frame()
	if tt.HasText("Plugin empty state") || !tt.HasText("Plugin loading state") || !tt.HasText("Cancel") {
		t.Fatal("loading rendered false empty state or lost cancel")
	}
	s.Busy = false
	s.loaded = true
	s.Error = "fixture installation failed"
	tt.Frame()
	if !tt.HasText("Plugin action failed") || !tt.HasText("fixture installation failed") || !tt.HasText("Refresh") {
		t.Fatal("error/retry controls disappeared")
	}
}

func TestSourceSkillsEmptyGeometryAndSearch(t *testing.T) {
	s := emptyResourceFixture(t)
	tt := fixedNavTester(s, "skills", 1470, 922)
	saveSourceCapture(t, tt, "skills-empty")
	search, _ := tt.Find("Skills search row")
	empty, _ := tt.Find("Skills no matches state")
	text, _ := tt.Find("No skills match")
	if search.X != 252 || search.Y != 12 || search.W != 295 || search.H != 36 || math.Abs(float64(empty.Y-96)) > .1 || math.Abs(float64(text.Y-154)) > .1 {
		t.Fatal("source Skills search/list empty bounds", search, empty, text)
	}
	if !tt.HasText("Select a skill") {
		t.Fatal("empty detail missing")
	}
	click(t, tt, "Search skills")
	tt.Type("missing")
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if s.Selected != "" || s.HasModal() {
		t.Fatal("empty search created selection or modal")
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

func TestSourcePluginEmptyUsesSecondarySurfaceAndFixedTile(t *testing.T) {
	s := emptyResourceFixture(t)
	secondary, elevated := ui.Hex("#242424"), ui.Hex("#292929")
	s.Options.Theme = func() Theme { return Theme{SurfaceSecondary: secondary, Elevated: elevated} }
	tt := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(ui.DarkTheme())
		s.pluginEmpty(c, "extensions", "Empty fixture", "Hint")
	}, 912, 400)
	card, _ := tt.Find("Plugin empty state")
	tile, _ := tt.Find("Plugin empty glyph tile")
	wantSurface := tt.Image().RGBAAt(int(card.X+20), int(card.Y+20))
	wantTile := tt.Image().RGBAAt(int(tile.X+10), int(tile.Y+10))
	if wantSurface.R != 0x24 || wantSurface.G != 0x24 || wantTile.R != 0x29 || wantTile.G != 0x29 {
		t.Fatal("empty surface/tile tokens not forwarded", wantSurface, wantTile)
	}
}

func TestSourceBlankInstallRemainsVisibleButCannotLaunch(t *testing.T) {
	s := emptyResourceFixture(t)
	tt := fixedNavTester(s, "plugins", 1470, 922)
	click(t, tt, "Install")
	if s.Busy || s.Error == "" || len(s.Plugins) != 0 {
		t.Fatal("blank source launched instead of validating", s.Busy, s.Error)
	}
}
