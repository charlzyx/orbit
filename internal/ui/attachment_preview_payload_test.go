package orbitui

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/HugoSmits86/nativewebp"
	"github.com/charlzyx/orbit/internal/attachments"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

func attachmentPreviewHeader(w, h uint32) []byte {
	b := []byte("\x89PNG\r\n\x1a\n")
	chunk := make([]byte, 17)
	copy(chunk, "IHDR")
	binary.BigEndian.PutUint32(chunk[4:8], w)
	binary.BigEndian.PutUint32(chunk[8:12], h)
	chunk[12] = 8
	chunk[13] = 6
	b = binary.BigEndian.AppendUint32(b, 13)
	b = append(b, chunk...)
	return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(chunk))
}
func TestAttachmentPreviewOversizedHeaderPreservesPayloadWithoutPixelDecode(t *testing.T) {
	raw := attachmentPreviewHeader(32768, 32768)
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || cfg.Width != 32768 {
		t.Fatal(cfg, err)
	}
	prepared := attachments.Prepare(raw, "image/png")
	if !bytes.Equal(prepared.Bytes, raw) || prepared.MIME != "image/png" {
		t.Fatal("raw fallback changed")
	}
	// This header is intentionally not sent to a real pixel decoder.
	a := fixture(t)
	a.AddImage(raw, "image/png", "header")
	drainAttachmentOrder(t, a)
	if len(a.attachments) != 1 || a.attachments[0].Preview != nil || a.attachments[0].Image.Data != base64.StdEncoding.EncodeToString(raw) || a.attachments[0].Image.MIMEType != "image/png" {
		t.Fatal("raw attachment lost or unsafe preview admitted")
	}
}
func TestAttachmentPreviewSmallFormatsPreserveWirePayload(t *testing.T) {
	im := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	im.Set(0, 0, color.NRGBA{R: 255, A: 255})
	cases := []struct {
		name, mime string
		encode     func(*bytes.Buffer) error
	}{
		{"png", "image/png", func(b *bytes.Buffer) error { return png.Encode(b, im) }},
		{"jpeg", "image/jpeg", func(b *bytes.Buffer) error { return jpeg.Encode(b, im, nil) }},
		{"gif", "image/gif", func(b *bytes.Buffer) error { return gif.Encode(b, im, nil) }},
		{"bmp", "image/bmp", func(b *bytes.Buffer) error { return bmp.Encode(b, im) }},
		{"webp", "image/webp", func(b *bytes.Buffer) error { return nativewebp.Encode(b, im, nil) }},
		{"tiff", "image/tiff", func(b *bytes.Buffer) error { return tiff.Encode(b, im, nil) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			if err := tc.encode(&b); err != nil {
				t.Fatal(err)
			}
			raw := b.Bytes()
			a := fixture(t)
			a.AddImage(raw, tc.mime, tc.name)
			drainAttachmentOrder(t, a)
			if len(a.attachments) != 1 || a.attachments[0].Preview == nil {
				t.Fatal("tiny registered raster lost preview")
			}
			got := a.attachments[0].Image
			if got.MIMEType != tc.mime || got.Data != base64.StdEncoding.EncodeToString(raw) {
				t.Fatal("preview changed wire payload")
			}
		})
	}
}
func TestAttachmentPreviewUnsupportedAndInvalidKeepRaw(t *testing.T) {
	for _, tc := range []struct {
		name, mime string
		raw        []byte
	}{{"unknown", "image/unknown", []byte("opaque")}, {"invalid", "image/png", []byte("bad png")}} {
		t.Run(tc.name, func(t *testing.T) {
			a := fixture(t)
			a.AddImage(tc.raw, tc.mime, tc.name)
			drainAttachmentOrder(t, a)
			if len(a.attachments) != 1 {
				t.Fatal("attachment dropped")
			}
			got := a.attachments[0]
			if got.Preview != nil || got.Image.MIMEType != tc.mime || got.Image.Data != base64.StdEncoding.EncodeToString(tc.raw) {
				t.Fatal("invalid preview changed payload")
			}
		})
	}
}
