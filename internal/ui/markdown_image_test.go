package orbitui

import (
	"bytes"
	"encoding/base64"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func markdownDataFixture(t *testing.T) string { return markdownSizedDataFixture(t, 160, 80) }
func markdownSizedDataFixture(t *testing.T, width, height int) string {
	t.Helper()
	im := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			im.SetNRGBA(x, y, color.NRGBA{R: 240, G: 30, B: 20, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
}
func TestMountedMarkdownDataImageLoadsInsteadOfPermanentAlt(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	a.SetSnapshot(Snapshot{SessionID: "image-owner", Messages: []Message{{ID: "image", Role: "assistant", Text: "![C09 raster fixture](" + markdownDataFixture(t) + ")"}}})
	tt := ui.NewTester(a.View, 1200, 800)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		if r, ok := tt.Find("C09 raster fixture"); ok && r.H >= 79 && math.Abs(float64(r.W/r.H-2)) < .05 {
			red, green, blue, _ := tt.Image().At(int(r.X+r.W/2), int(r.Y+r.H/2)).RGBA()
			if red>>8 < 200 || green>>8 > 80 || blue>>8 > 80 {
				t.Fatal("native image bounds contain placeholder pixels, not decoded raster", red, green, blue)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	r, _ := tt.Find("C09 raster fixture")
	t.Fatal("HTTP/data Markdown images still remain alt placeholders", r)
}

func waitMarkdownRaster(t *testing.T, tt *ui.Tester, alt string) ui.Rect {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		tt.Frame()
		if r, ok := tt.Find(alt); ok && r.H >= 79 && math.Abs(float64(r.W/r.H-2)) < .05 {
			return r
		}
		time.Sleep(time.Millisecond * 5)
	}
	r, _ := tt.Find(alt)
	t.Fatal("native raster not loaded", alt, r)
	return ui.Rect{}
}
func TestMountedMarkdownHTMLDataPreviewAndUnsafeFallback(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	a.SetSnapshot(Snapshot{SessionID: "html", Messages: []Message{{ID: "html-image", Role: "assistant", Text: `<img alt="HTML raster" src="` + markdownDataFixture(t) + `" onerror="never_execute()">`}}})
	tt := ui.NewTester(a.View, 1200, 800)
	waitMarkdownRaster(t, tt, "HTML raster")
	if err := tt.Click("Open Markdown image HTML raster"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen || a.imagePreview == nil {
		t.Fatal("safe data raster not opened in native preview")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if a.imageOpen {
		t.Fatal("native preview did not close")
	}
	a.SetSnapshot(Snapshot{SessionID: "unsafe", Messages: []Message{{ID: "svg", Role: "assistant", Text: "![unsafe SVG](data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(`<svg onload="never_execute()"/>`)) + ")"}}})
	tt.Frame()
	time.Sleep(10 * time.Millisecond)
	tt.Frame()
	if _, ok := tt.Find("Open Markdown image unsafe SVG"); ok {
		t.Fatal("active-content format became clickable image")
	}
	if !tt.HasText("unsafe SVG") {
		t.Fatal("blocked format lost safe alt fallback")
	}
}
func TestMountedMarkdownMixedLocalFilePathRemainsUnchanged(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	data := markdownDataFixture(t)
	_, encoded, _ := strings.Cut(data, ",")
	raw, e := base64.StdEncoding.DecodeString(encoded)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "local.png")
	if e = os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	a.SetSnapshot(Snapshot{SessionID: "mixed", Messages: []Message{{ID: "mixed", Role: "assistant", Text: "![local raster](" + path + ")\n![remote raster](" + data + ")"}}})
	tt := ui.NewTester(a.View, 1200, 800)
	waitMarkdownRaster(t, tt, "local raster")
	waitMarkdownRaster(t, tt, "remote raster")
}
func TestMountedMarkdownImageHeldPressCannotOpenReplacementOwner(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	data := markdownDataFixture(t)
	snapshot := Snapshot{SessionID: "A", RuntimeID: "runtime-A", Messages: []Message{{ID: "same-id", Role: "assistant", Text: "![owner raster](" + data + ")"}}}
	a.SetSnapshot(snapshot)
	tt := ui.NewTester(a.View, 1200, 800)
	waitMarkdownRaster(t, tt, "owner raster")
	button, _ := tt.Find("Open Markdown image owner raster")
	x, y := button.X+button.W/2, button.Y+button.H/2
	tt.Press(x, y)
	a.SetSnapshot(Snapshot{SessionID: "B", RuntimeID: "runtime-B"})
	a.SetSnapshot(snapshot)
	waitMarkdownRaster(t, tt, "owner raster")
	tt.Release(x, y)
	tt.Frame()
	if a.imageOpen {
		t.Fatal("press from old owner opened new owner's preview")
	}
	if err := tt.Click("Open Markdown image owner raster"); err != nil {
		t.Fatal(err)
	}
	if !a.imageOpen {
		t.Fatal("fresh owner image click failed")
	}
}

func TestMountedMarkdownLargeRasterClampsWithoutDistortingOrCropping(t *testing.T) {
	a := fixture(t)
	t.Cleanup(a.markdownImages.Close)
	source := "![wide raster](" + markdownSizedDataFixture(t, 1200, 600) + ")"
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Fill().Padding(10).Children(func() { markdown.View(c, source, &a.md) })
	}, 420, 400)
	r := waitMarkdownRaster(t, tt, "wide raster")
	if r.W > 400 || r.X+r.W > 410.5 || math.Abs(float64(r.W/r.H-2)) > .02 {
		t.Fatal("large remote/data image overflowed or distorted", r)
	}
	red, green, blue, _ := tt.Image().At(int(r.X+r.W/2), int(r.Y+r.H/2)).RGBA()
	if red>>8 < 200 || green>>8 > 80 || blue>>8 > 80 {
		t.Fatal("native image not rendered", r)
	}
}

func TestMountedMarkdownImageServiceRoundtripInvalidatesHeldOwner(t *testing.T) {
	x := workspaceDialogFixture(t)
	a := x.UI
	initial := x.workspace
	data := markdownDataFixture(t)
	a.SetSnapshot(Snapshot{SessionID: "same", RuntimeID: "same-runtime", Workspace: initial, Messages: []Message{{ID: "same", Role: "assistant", Text: "![service raster](" + data + ")"}}})
	tt := ui.NewTester(a.View, 1200, 800)
	waitMarkdownRaster(t, tt, "service raster")
	button, _ := tt.Find("Open Markdown image service raster")
	px, py := button.X+button.W/2, button.Y+button.H/2
	tt.Press(px, py)
	x.changeWorkspaceServices(t.TempDir())
	x.changeWorkspaceServices(initial)
	waitMarkdownRaster(t, tt, "service raster")
	tt.Release(px, py)
	tt.Frame()
	if a.imageOpen {
		t.Fatal("service A-B-A bypassed owner epoch and activated held old image")
	}
	button, _ = tt.Find("Open Markdown image service raster")
	px, py = button.X+button.W/2, button.Y+button.H/2
	tt.Press(px, py)
	x.changeWorkspaceServices(initial) // same owner must not invalidate normal input
	tt.Frame()
	tt.Release(px, py)
	tt.Frame()
	if !a.imageOpen || a.imagePreview == nil {
		t.Fatal("fresh service owner preview did not open")
	}
	x.changeWorkspaceServices(t.TempDir())
	if a.imageOpen || a.imagePreview != nil {
		t.Fatal("workspace services retained old owner's preview")
	}
}

func TestMarkdownPreviewClearsOnEachSnapshotOwnerDimension(t *testing.T) {
	for _, dimension := range []string{"runtime", "session", "path", "workspace"} {
		t.Run(dimension, func(t *testing.T) {
			a := fixture(t)
			t.Cleanup(a.markdownImages.Close)
			s := Snapshot{SessionID: "same", RuntimeID: "runtime", SessionPath: "path", Workspace: "workspace"}
			a.SetSnapshot(s)
			a.imageOpen = true
			a.imagePreview = ui.NewBitmap(image.NewNRGBA(image.Rect(0, 0, 2, 2)))
			switch dimension {
			case "runtime":
				s.RuntimeID += "-new"
			case "session":
				s.SessionID += "-new"
			case "path":
				s.SessionPath += "-new"
			case "workspace":
				s.Workspace += "-new"
			}
			a.SetSnapshot(s)
			if a.imageOpen || a.imagePreview != nil {
				t.Fatal("old preview survived owner", dimension)
			}
		})
	}
}
