package background

// Compatibility implementation of image-rs0.25.10 Triangle sampling, used by
// the pinned Orbit reference. See licenses/image-rs-MIT.txt. Straight RGBA,
// float32 intermediate values, normalized edge weights and final byte rounding
// are intentional: Go's premultiplied bilinear scaler has different pixels.

import (
	"context"
	"image"
	"image/color"
	"math"
)

type triangleWeights struct {
	first  int
	values []float32
}

func triangleAt(input, output, index int) triangleWeights {
	ratio := float32(input) / float32(output)
	support := max(float32(1), ratio)
	center := (float32(index) + .5) * ratio
	left := min(input-1, max(0, int(math.Floor(float64(center-support)))))
	right := min(input, max(left+1, int(math.Ceil(float64(center+support)))))
	center -= .5
	w := triangleWeights{left, make([]float32, right-left)}
	var sum float32
	for i := range w.values {
		d := float32(i+left) - center
		d /= support
		if d < 0 {
			d = -d
		}
		v := max(float32(0), 1-d)
		w.values[i] = v
		sum += v
	}
	for i := range w.values {
		w.values[i] /= sum
	}
	return w
}

func resizeTriangle(ctx context.Context, src image.Image, width, height int) (*image.NRGBA, error) {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	xw := make([]triangleWeights, width)
	for x := range xw {
		xw[x] = triangleAt(sw, width, x)
	}
	// One vertically filtered row rather than a full float32 intermediate
	// image avoids an unnecessary large allocation for wide source pictures.
	row := make([]float32, sw*4)
	for y := 0; y < height; y++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		yw := triangleAt(sh, height, y)
		clear(row)
		for x := 0; x < sw; x++ {
			for k, w := range yw.values {
				p := color.NRGBAModel.Convert(src.At(b.Min.X+x, b.Min.Y+yw.first+k)).(color.NRGBA)
				row[4*x] += float32(p.R) * w
				row[4*x+1] += float32(p.G) * w
				row[4*x+2] += float32(p.B) * w
				row[4*x+3] += float32(p.A) * w
			}
		}
		for x, w := range xw {
			var p [4]float32
			for k, v := range w.values {
				for c := 0; c < 4; c++ {
					p[c] += row[(w.first+k)*4+c] * v
				}
			}
			at := out.PixOffset(x, y)
			for c := 0; c < 4; c++ {
				out.Pix[at+c] = uint8(math.Round(float64(min(255, max(0, p[c])))))
			}
		}
	}
	return out, nil
}
