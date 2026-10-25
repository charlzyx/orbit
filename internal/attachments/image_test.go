package attachments

import (
	"bytes"
	"golang.org/x/image/webp"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestPreservesSmallAndAnimatedImages(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 12, 13)))
	v := Prepare(b.Bytes(), "image/png")
	if v.Compressed || !bytes.Equal(v.Bytes, b.Bytes()) {
		t.Fatal("small source changed")
	}
	for _, m := range []string{"image/gif", "image/svg+xml", "image/tiff", "image/unknown"} {
		v = Prepare([]byte("original bytes"), m)
		if v.Compressed || string(v.Bytes) != "original bytes" {
			t.Fatal(m)
		}
	}
}
func TestOpaqueResizeAndAlphaPreservation(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2000, 1000))
	for y := 0; y < 1000; y++ {
		for x := 0; x < 2000; x++ {
			src.SetRGBA(x, y, color.RGBA{90, 20, 70, 255})
		}
	}
	var b bytes.Buffer
	jpeg.Encode(&b, src, nil)
	v := Prepare(b.Bytes(), "image/jpeg")
	if !v.Compressed || v.MIME != "image/jpeg" || v.Width != 1400 || v.Height != 700 {
		t.Fatalf("%+v", v)
	}
	src.SetRGBA(0, 0, color.RGBA{0, 0, 0, 0})
	b.Reset()
	png.Encode(&b, src)
	v = Prepare(b.Bytes(), "image/png")
	if !v.Compressed || v.MIME != "image/webp" {
		t.Fatal("alpha lost")
	}
	d, e := webp.Decode(bytes.NewReader(v.Bytes))
	if e != nil {
		t.Fatal(e)
	}
	_, _, _, a := d.At(0, 0).RGBA()
	if a == 65535 {
		t.Fatal("alpha not preserved")
	}
}
func TestFit(t *testing.T) {
	for _, v := range [][4]int{{4000, 2000, 1400, 700}, {2000, 4000, 700, 1400}, {1, 99999, 1, 1400}, {20, 30, 20, 30}} {
		w, h := Fit(v[0], v[1], 1400)
		if w != v[2] || h != v[3] {
			t.Fatal(v, w, h)
		}
	}
}
