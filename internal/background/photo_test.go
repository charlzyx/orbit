package background

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func photoFixture(w, h int) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.SetNRGBA(x, y, color.NRGBA{uint8((x*37 + y*19) % 256), uint8((x*7 + y*61 + 83) % 256), uint8((x*71 + y*5 + 12) % 256), uint8((x*13 + y*47 + 90) % 256)})
		}
	}
	return im
}
func writePhoto(t *testing.T, path string, im image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPhotoChooserSupportsSourceBMPAndTIFF(t *testing.T) {
	for _, tc := range []struct {
		name   string
		encode func(io.Writer, image.Image) error
	}{
		{"source.bmp", bmp.Encode}, {"source.tiff", func(w io.Writer, im image.Image) error { return tiff.Encode(w, im, nil) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tc.name)
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			err = tc.encode(f, photoFixture(19, 13))
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
			s := NewPhotoStore(t.TempDir())
			if _, err = s.Choose(context.Background(), path); err != nil {
				t.Fatal(err)
			}
			if im, err := s.Image(context.Background()); err != nil || im == nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPhotoConfigPerFieldAndSourceDefaults(t *testing.T) {
	s := NewPhotoStore(t.TempDir())
	if s.Config().Tuning != DefaultTuning() {
		t.Fatal(s.Config())
	}
	for _, tc := range []struct {
		json string
		want Tuning
	}{
		{`{"source":"camera.png","blur":1.26,"cell":6,"fade":0.22}`, Tuning{1.5, 6, .22}},
		{`{"blur":16,"cell":5,"fade":0.9}`, Tuning{6, 4, .14}},
		{`{"blur":-2,"cell":8,"fade":0}`, Tuning{0, 8, 0}},
		{`{"blur":"bad","cell":2,"fade":0.08}`, Tuning{0, 2, .08}},
		{`invalid`, DefaultTuning()},
	} {
		if err := os.WriteFile(s.configPath(), []byte(tc.json), 0600); err != nil {
			t.Fatal(err)
		}
		if got := s.Config().Tuning; got != tc.want {
			t.Fatalf("%s: got%+v want%+v", tc.json, got, tc.want)
		}
	}
	for i, v := range FadeHeights {
		if got := (Tuning{Fade: v}).FadePeak(); got != FadePeaks[i] {
			t.Fatal(v, got)
		}
	}
}

func TestPhotoChooseTuneCacheClearAndFailure(t *testing.T) {
	ctx := context.Background()
	s := NewPhotoStore(t.TempDir())
	source := filepath.Join(t.TempDir(), "source.png")
	writePhoto(t, source, photoFixture(19, 11))
	if _, err := s.SetTuning(Tuning{1.5, 6, .22}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Choose(ctx, source)
	if err != nil {
		t.Fatal(err)
	}
	if c.Source == nil || *c.Source != "source.png" || c.Tuning != (Tuning{1.5, 6, .22}) {
		t.Fatal(c)
	}
	first, err := s.Image(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if first.Rect.Dx() != 19 || first.Rect.Dy() != 11 {
		t.Fatal(first.Rect)
	}
	if _, err = s.SetTuning(Tuning{1.5, 6, .32}); err != nil {
		t.Fatal(err)
	}
	second, err := s.Image(ctx)
	if err != nil || first != second {
		t.Fatalf("fade rebuilt pixels: %v", err)
	}
	if _, err = s.SetTuning(Tuning{2, 6, .32}); err != nil {
		t.Fatal(err)
	}
	third, err := s.Image(ctx)
	if err != nil || third == second {
		t.Fatalf("blur failed to rebuild: %v", err)
	}
	old, _ := os.ReadFile(s.imagePath())
	if _, err = s.Choose(ctx, filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Fatal("missing source accepted")
	}
	now, _ := os.ReadFile(s.imagePath())
	if !bytes.Equal(old, now) {
		t.Fatal("failed selection replaced image")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = s.Choose(canceled, source); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = s.Clear(); err != nil {
		t.Fatal(err)
	}
	if err = s.Clear(); err != nil {
		t.Fatal("repeat clear", err)
	}
	im, err := s.Image(ctx)
	if err != nil || im != nil {
		t.Fatal(im, err)
	}
	if s.Config().Source != nil || s.Config().Tuning != DefaultTuning() {
		t.Fatal(s.Config())
	}
}

func TestPhotoDimensionsAndProtectedWrite(t *testing.T) {
	ctx := context.Background()
	src := photoFixture(2001, 17)
	fitted, err := fitPhoto(ctx, src)
	if err != nil {
		t.Fatal(err)
	}
	if fitted.Bounds().Dx() != 1600 || fitted.Bounds().Dy() != 14 {
		t.Fatal(fitted.Bounds())
	}
	small := photoFixture(12, 17)
	same, err := fitPhoto(ctx, small)
	if err != nil || same != small {
		t.Fatal("small image enlarged")
	}
	root := t.TempDir()
	blocked := filepath.Join(root, "file")
	if err = os.WriteFile(blocked, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	s := NewPhotoStore(blocked)
	if _, err = s.SetTuning(Tuning{2, 4, .14}); err == nil {
		t.Fatal("failed write claimed success")
	}
}

func TestPhotoDitherPreservesColorAlphaAndPartialCells(t *testing.T) {
	ctx := context.Background()
	src := photoFixture(19, 13)
	for _, cell := range CellSizes {
		out, err := Dither(ctx, src, cell)
		if err != nil {
			t.Fatal(err)
		}
		for y := 0; y < 13; y++ {
			for x := 0; x < 19; x++ {
				cx, cy := x/cell, y/cell
				p := out.NRGBAAt(x, y)
				origin := src.NRGBAAt(min(18, cx*cell+cell/2), min(12, cy*cell+cell/2))
				threshold := (float32(bayer4[cy%4][cx%4]) + .5) / 16
				want := color.NRGBA{quantize(origin.R, threshold), quantize(origin.G, threshold), quantize(origin.B, threshold), origin.A}
				if p != want {
					t.Fatalf("cell%d,%d,%d=%v want%v", cell, x, y, p, want)
				}
				for _, v := range []uint8{p.R, p.G, p.B} {
					if v%85 != 0 {
						t.Fatal("non-quantized", v)
					}
				}
			}
		}
	}
}

func TestPhotoBlurMatchesPinnedRustOracle(t *testing.T) {
	b, err := os.ReadFile("testdata/blur-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Width, Height int
		Sigma         float32
		SHA256        string
	}
	if err = json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		out, err := Soften(context.Background(), photoFixture(tc.Width, tc.Height), tc.Sigma)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(out.Pix)
		if hex.EncodeToString(digest[:]) != tc.SHA256 {
			t.Fatalf("%dx%d sigma%g: %x want%s", tc.Width, tc.Height, tc.Sigma, digest, tc.SHA256)
		}
	}
	im := photoFixture(9, 7)
	same, err := Soften(context.Background(), im, 0)
	if err != nil || same != im {
		t.Fatal("blur off changed pixels")
	}
}

func TestPhotoFitMatchesPinnedRustOracle(t *testing.T) {
	b, err := os.ReadFile("testdata/resize-oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Width, Height int
		OutWidth      int `json:"out_width"`
		OutHeight     int `json:"out_height"`
		SHA256        string
	}
	if err = json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		out, err := fitPhoto(context.Background(), photoFixture(tc.Width, tc.Height))
		if err != nil {
			t.Fatal(err)
		}
		im := out.(*image.NRGBA)
		digest := sha256.Sum256(im.Pix)
		if im.Rect.Dx() != tc.OutWidth || im.Rect.Dy() != tc.OutHeight || hex.EncodeToString(digest[:]) != tc.SHA256 {
			t.Errorf("%dx%d: got%s %x want%dx%d %s", tc.Width, tc.Height, im.Rect, digest, tc.OutWidth, tc.OutHeight, tc.SHA256)
		}
	}
}

func TestPhotoConcurrentReadsTuningAndCancellation(t *testing.T) {
	s := NewPhotoStore(t.TempDir())
	source := filepath.Join(t.TempDir(), "source.png")
	writePhoto(t, source, photoFixture(31, 19))
	if _, err := s.Choose(context.Background(), source); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Go(func() {
			for n := 0; n < 4; n++ {
				if _, err := s.Image(context.Background()); err != nil {
					t.Error(err)
				}
				s.Config()
				if _, err := s.SetTuning(Tuning{float32(n) / 2, CellSizes[n], FadeHeights[n]}); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.Image(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
