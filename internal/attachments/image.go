// Package attachments preserves Orbit's image-ingest policy before Pi RPC.
package attachments

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"math"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	_ "golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// MaxDecodedPixels preserves the existing preparation threshold. It is not a
// resident-memory budget: this many RGBA pixels alone can approach 1 GiB.
const MaxDecodedPixels = 268435456

func DecodeDimensionsAllowed(width, height int) bool {
	return width > 0 && height > 0 && width <= MaxDecodedPixels && height <= MaxDecodedPixels && width <= MaxDecodedPixels/height
}

const MaxEdge = 1400
const MaxBase64Bytes = 512 * 1024

type Image struct {
	Bytes         []byte
	MIME          string
	Compressed    bool
	Width, Height int
}

func (i Image) Base64() string { return base64.StdEncoding.EncodeToString(i.Bytes) }

// Prepare preserves undecodable/animated inputs. The payload target is soft for
// lossless alpha images and JPEGs at the original quality floor of 40.
func Prepare(input []byte, mime string) Image {
	result := Image{Bytes: bytes.Clone(input), MIME: mime}
	switch strings.ToLower(strings.Split(mime, ";")[0]) {
	case "image/gif", "image/svg+xml", "image/tiff":
		return result
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(input))
	if err != nil {
		return result
	}
	result.Width, result.Height = cfg.Width, cfg.Height
	if !DecodeDimensionsAllowed(cfg.Width, cfg.Height) {
		return result
	}
	if cfg.Width <= MaxEdge && cfg.Height <= MaxEdge && base64.StdEncoding.EncodedLen(len(input)) <= MaxBase64Bytes {
		return result
	}
	decoded, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return result
	}
	alpha := cfg.ColorModel == color.NRGBAModel || cfg.ColorModel == color.NRGBA64Model || cfg.ColorModel == color.RGBAModel || cfg.ColorModel == color.RGBA64Model || cfg.ColorModel == color.AlphaModel || cfg.ColorModel == color.Alpha16Model
	if palette, ok := cfg.ColorModel.(color.Palette); ok {
		for _, c := range palette {
			_, _, _, a := c.RGBA()
			if a < 65535 {
				alpha = true
				break
			}
		}
	}
	w, h := Fit(cfg.Width, cfg.Height, MaxEdge)
	if w != cfg.Width || h != cfg.Height {
		dst := image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.BiLinear.Scale(dst, dst.Bounds(), decoded, decoded.Bounds(), draw.Src, nil)
		decoded = dst
	}
	var out bytes.Buffer
	if alpha {
		err = nativewebp.Encode(&out, decoded, nil)
		mime = "image/webp"
	} else {
		mime = "image/jpeg"
		for quality := 75; quality >= 40; quality -= 5 {
			out.Reset()
			err = jpeg.Encode(&out, decoded, &jpeg.Options{Quality: quality})
			if err != nil || base64.StdEncoding.EncodedLen(out.Len()) <= MaxBase64Bytes {
				break
			}
		}
	}
	if err != nil {
		return result
	}
	return Image{Bytes: out.Bytes(), MIME: mime, Compressed: true, Width: w, Height: h}
}
func Fit(width, height, limit int) (int, int) {
	longest := max(width, height, 1)
	if longest <= limit {
		return width, height
	}
	scale := float64(limit) / float64(longest)
	return max(1, int(math.Round(float64(width)*scale))), max(1, int(math.Round(float64(height)*scale)))
}
