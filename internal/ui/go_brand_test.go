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
	r, g, b, a := im.At(256, 90).RGBA()
	if a == 0 || g <= r || b <= r {
		t.Fatal("Go cyan identity missing")
	}
	for _, mode := range []string{"light", "dark"} {
		t.Run(mode, func(t *testing.T) {
			app := fixture(t)
			app.Appearance.Mode = mode
			tt := ui.NewTester(app.View, 1470, 922)
			tt.Frame()
			badge, ok := tt.Find("Go implementation badge")
			if !ok || badge.W <= 0 || badge.H <= 0 {
				t.Fatal("Go badge not actually mounted", badge)
			}
			if _, ok := tt.Find("Sidebar wordmark"); !ok {
				t.Fatal("original wordmark missing")
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
			icon, ok := tt.Find("Orbit Go app icon")
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
