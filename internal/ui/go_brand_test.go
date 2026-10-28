package orbitui

import (
	"bytes"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestGoBrandDedicatedIconAndLightDarkScreens(t *testing.T) {
	data, err := orbitassets.Asset("go-app-icon.png")
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if im.Bounds().Dx() != 512 || im.Bounds().Dy() != 512 {
		t.Fatal("icon dimensions", im.Bounds())
	}
	_, _, _, alpha := im.At(0, 0).RGBA()
	if alpha != 0 {
		t.Fatal("icon corners not transparent")
	}
	r, g, b, a := im.At(144, 160).RGBA()
	if a == 0 || r < 45000 || g < 45000 || b < 45000 {
		t.Fatal("Pixel PI silver stem missing")
	}
	// The square P counter and the P/I gap must remain open, not a solid block.
	for _, point := range [][2]int{{206, 216}, {316, 256}} {
		r, g, b, alpha := im.At(point[0], point[1]).RGBA()
		if alpha == 0 || r > 20000 || g > 20000 || b > 20000 {
			t.Fatal("PI counter or letter spacing lost", point)
		}
	}
	r, g, b, a = im.At(364, 256).RGBA()
	if a == 0 || r < 45000 || g < 45000 || b < 45000 {
		t.Fatal("separate pixel I stem missing")
	}
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			app := fixture(t)
			app.Appearance.Mode = mode
			wordmark := icons["icons/gopi-wordmark.svg"]
			if wordmark == nil {
				t.Fatal("GOPI vector not loaded through embedded asset path")
			}
			if w, h := wordmark.Size(); w != 74 || h != 26 {
				t.Fatal("unexpected GOPI vector view box", w, h)
			}
			tt := ui.NewTester(app.View, 1470, 922)
			tt.Frame()
			if _, ok := tt.Find("Go implementation badge"); ok {
				t.Fatal("retired yellow GO badge still mounted")
			}
			if _, ok := tt.Find("Sidebar wordmark"); !ok {
				t.Fatal("GOPI wordmark missing")
			}
			save := func(name string) {
				if dir := os.Getenv("ORBIT_GO_BRAND_CAPTURE_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					f, err := os.Create(filepath.Join(dir, name+"-"+mode+".png"))
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					if err := png.Encode(f, tt.Image()); err != nil {
						t.Fatal(err)
					}
				}
			}
			save("sidebar")
			app.Navigate("settings")
			app.settingsPage = "about"
			tt.Frame()
			if !tt.HasText("GoPi") {
				t.Fatal("About display name not GoPi", tt.Texts())
			}
			icon, ok := tt.Find("GoPi app icon")
			if !ok || icon.W != 48 || icon.H != 48 {
				t.Fatal("About Go icon not mounted at preserved size", icon)
			}
			if logos["go-app-icon.png"] == nil || logos["go-app-icon.png"] == logos["app-icon.png"] {
				t.Fatal("About icon not separately loaded")
			}
			save("about")
		})
	}
}
