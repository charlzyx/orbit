package background

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"

	"github.com/charlzyx/orbit/internal/preferences"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

var CellSizes = [...]int{2, 3, 4, 6, 8}
var CellLabels = [...]string{"Fine", "Small", "Medium", "Large", "Coarse"}
var FadeHeights = [...]float32{0, .08, .14, .22, .32}
var FadePeaks = [...]float32{0, .55, .72, .88, 1}
var FadeLabels = [...]string{"None", "Subtle", "Medium", "Deep", "Full"}

const MaxEdge = 1600
const BlurMax float32 = 6
const BlurStep float32 = .5

// Tuning is compatible with the flat source background.json format.
type Tuning struct {
	Blur float32 `json:"blur"`
	Cell int     `json:"cell"`
	Fade float32 `json:"fade"`
}
type PhotoConfig struct {
	Source *string `json:"source"`
	Tuning
}

func DefaultTuning() Tuning { return Tuning{Cell: 4, Fade: .14} }
func BlurFromNorm(n float32) float32 {
	if math.IsNaN(float64(n)) {
		return 0
	}
	return float32(math.Round(float64(min(1, max(0, n))*BlurMax/BlurStep))) * BlurStep
}
func (t Tuning) Normalize() Tuning {
	r := DefaultTuning()
	if !math.IsNaN(float64(t.Blur)) && !math.IsInf(float64(t.Blur), 0) {
		r.Blur = BlurFromNorm(t.Blur / BlurMax)
	}
	for _, v := range CellSizes {
		if t.Cell == v {
			r.Cell = v
		}
	}
	for _, v := range FadeHeights {
		if t.Fade == v {
			r.Fade = v
		}
	}
	return r
}
func (t Tuning) CellIndex() int {
	for i, v := range CellSizes {
		if t.Cell == v {
			return i
		}
	}
	return 0
}
func (t Tuning) FadeIndex() int {
	for i, v := range FadeHeights {
		if math.Abs(float64(t.Fade-v)) < .001 {
			return i
		}
	}
	return 0
}
func (t Tuning) FadePeak() float32 { return FadePeaks[t.FadeIndex()] }

// PhotoStore performs filesystem and image work synchronously: hosts MUST call
// it from workers, never a render callback. Returned images are immutable.
// Separate store instances should not mutate the same configuration directory.
type PhotoStore struct {
	dir    string
	mu     sync.Mutex
	key    photoKey
	cached *image.NRGBA
	valid  bool
}
type photoKey struct {
	modified, size int64
	blur           float32
	cell           int
}

func NewPhotoStore(dir string) *PhotoStore { return &PhotoStore{dir: dir} }
func (s *PhotoStore) imagePath() string    { return filepath.Join(s.dir, "new-task-background.png") }
func (s *PhotoStore) configPath() string   { return filepath.Join(s.dir, "background.json") }
func (s *PhotoStore) Config() PhotoConfig  { s.mu.Lock(); defer s.mu.Unlock(); return s.readConfig() }
func (s *PhotoStore) readConfig() PhotoConfig {
	r := PhotoConfig{Tuning: DefaultTuning()}
	b, err := os.ReadFile(s.configPath())
	if err != nil {
		return r
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(b, &raw) != nil {
		return r
	}
	_ = json.Unmarshal(raw["source"], &r.Source)
	var blur float32
	if json.Unmarshal(raw["blur"], &blur) == nil {
		r.Blur = BlurFromNorm(blur / BlurMax)
	}
	var cell int
	if json.Unmarshal(raw["cell"], &cell) == nil {
		for _, v := range CellSizes {
			if cell == v {
				r.Cell = cell
			}
		}
	}
	var fade float32
	if json.Unmarshal(raw["fade"], &fade) == nil {
		for _, v := range FadeHeights {
			if fade == v {
				r.Fade = fade
			}
		}
	}
	return r
}
func (s *PhotoStore) writeConfig(c PhotoConfig) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return preferences.AtomicWrite(s.configPath(), append(b, '\n'), 0600)
}

// Choose validates before touching the active image. The application only
// stores the basename, not the selected source's private absolute path.
func (s *PhotoStore) Choose(ctx context.Context, path string) (PhotoConfig, error) {
	im, err := decodePhoto(ctx, path)
	if err != nil {
		return PhotoConfig{}, fmt.Errorf("read background: %w", err)
	}
	im, err = fitPhoto(ctx, im)
	if err != nil {
		return PhotoConfig{}, err
	}
	var b bytes.Buffer
	if err = png.Encode(&b, im); err != nil {
		return PhotoConfig{}, err
	}
	if err = ctx.Err(); err != nil {
		return PhotoConfig{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err = ctx.Err(); err != nil {
		return PhotoConfig{}, err
	}
	c := s.readConfig()
	label := filepath.Base(path)
	c.Source = &label
	if err = preferences.AtomicWrite(s.imagePath(), b.Bytes(), 0600); err != nil {
		return PhotoConfig{}, fmt.Errorf("save background: %w", err)
	}
	if err = s.writeConfig(c); err != nil {
		s.valid = false
		return PhotoConfig{}, fmt.Errorf("save background settings: %w", err)
	}
	s.valid = false
	return c, nil
}
func (s *PhotoStore) SetTuning(t Tuning) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.readConfig()
	t = t.Normalize()
	if c.Tuning == t {
		return false, nil
	}
	c.Tuning = t
	if err := s.writeConfig(c); err != nil {
		return false, err
	}
	return true, nil
}
func (s *PhotoStore) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for _, path := range []string{s.imagePath(), s.configPath()} {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
		}
	}
	s.valid = false
	s.cached = nil
	return errors.Join(errs...)
}
func (s *PhotoStore) Image(ctx context.Context) (*image.NRGBA, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c := s.readConfig()
	st, err := os.Stat(s.imagePath())
	if os.IsNotExist(err) {
		s.cached = nil
		s.valid = false
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	k := photoKey{st.ModTime().UnixNano(), st.Size(), c.Blur, c.Cell}
	if s.valid && s.key == k {
		return s.cached, nil
	}
	im, err := decodePhoto(ctx, s.imagePath())
	if err != nil {
		return nil, err
	}
	im, err = fitPhoto(ctx, im)
	if err != nil {
		return nil, err
	}
	out, err := Dither(ctx, im, c.Cell)
	if err != nil {
		return nil, err
	}
	out, err = Soften(ctx, out, c.Blur)
	if err != nil {
		return nil, err
	}
	s.key = k
	s.cached = out
	s.valid = true
	return out, nil
}

func decodePhoto(ctx context.Context, path string) (image.Image, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(io.LimitReader(f, 128<<20))
	if err != nil {
		return nil, err
	}
	// Avoid decompression bombs while accepting ordinary large camera sources.
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 64<<20 {
		return nil, errors.New("background image exceeds64 megapixels")
	}
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	im, _, err := image.Decode(io.LimitReader(f, 128<<20))
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return im, nil
}
func fitPhoto(ctx context.Context, src image.Image) (image.Image, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= MaxEdge && h <= MaxEdge {
		return src, nil
	}
	ratio := float64(MaxEdge) / float64(max(w, h))
	w = max(1, int(math.Round(float64(w)*ratio)))
	h = max(1, int(math.Round(float64(h)*ratio)))
	return resizeTriangle(ctx, src, w, h)
}

var bayer4 = [4][4]uint8{{0, 8, 2, 10}, {12, 4, 14, 6}, {3, 11, 1, 9}, {15, 7, 13, 5}}

func quantize(v uint8, threshold float32) uint8 {
	x := min(1, max(0, (float32(v)/255*.70-.5)*1.3+.5)) * 3
	base := float32(math.Floor(float64(x)))
	if x-base > threshold {
		base++
	}
	return uint8(math.Round(float64(min(3, max(0, base)) / 3 * 255)))
}

// Dither is source-compatible color ordered dithering, not theme monochrome.
func Dither(ctx context.Context, src image.Image, grid int) (*image.NRGBA, error) {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 || int64(w)*int64(h) > 64<<20 {
		return nil, errors.New("invalid background dimensions")
	}
	grid = max(1, grid)
	out := image.NewNRGBA(image.Rect(0, 0, w, h))
	for cy := 0; cy < (h+grid-1)/grid; cy++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for cx := 0; cx < (w+grid-1)/grid; cx++ {
			p := color.NRGBAModel.Convert(src.At(b.Min.X+min(w-1, cx*grid+grid/2), b.Min.Y+min(h-1, cy*grid+grid/2))).(color.NRGBA)
			t := (float32(bayer4[cy%4][cx%4]) + .5) / 16
			p.R = quantize(p.R, t)
			p.G = quantize(p.G, t)
			p.B = quantize(p.B, t)
			for y := cy * grid; y < min(h, (cy+1)*grid); y++ {
				for x := cx * grid; x < min(w, (cx+1)*grid); x++ {
					out.SetNRGBA(x, y, p)
				}
			}
		}
	}
	return out, nil
}

// Soften uses the same three-box Gaussian approximation and rounded byte
// passes as image0.25.10 fast_blur. Sigma0 returns the original immutable image.
func Soften(ctx context.Context, src *image.NRGBA, sigma float32) (*image.NRGBA, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if sigma <= 0 {
		return src, nil
	}
	if math.IsNaN(float64(sigma)) || math.IsInf(float64(sigma), 0) {
		return nil, errors.New("invalid blur")
	}
	sigma = min(BlurMax, sigma)
	ideal := float32(math.Sqrt(float64(4*sigma*sigma + 1)))
	lo := int(math.Floor(float64(ideal)))
	if lo%2 == 0 {
		lo--
	}
	hi := lo + 2
	m := int(math.Round(float64(.75*float32(lo+3) - 3*sigma*sigma/float32(lo+1))))
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	cur := src
	for pass := 0; pass < 3; pass++ {
		box := hi
		if pass < m {
			box = lo
		}
		r := max(0, (box-1)/2)
		if r%2 == 0 {
			r++
		}
		a := image.NewNRGBA(image.Rect(0, 0, w, h))
		z := image.NewNRGBA(a.Bounds())
		weight := float32(1) / float32(2*r+1)
		// Sliding windows keep work linear even at the largest blur setting.
		for y := 0; y < h; y++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			for c := 0; c < 4; c++ {
				sum := 0
				for dx := -r; dx <= r; dx++ {
					sum += int(cur.Pix[cur.PixOffset(cur.Rect.Min.X+min(w-1, max(0, dx)), cur.Rect.Min.Y+y)+c])
				}
				for x := 0; x < w; x++ {
					a.Pix[a.PixOffset(x, y)+c] = uint8(math.Round(float64(float32(sum) * weight)))
					left, right := max(0, x-r), min(w-1, x+r+1)
					sum += int(cur.Pix[cur.PixOffset(cur.Rect.Min.X+right, cur.Rect.Min.Y+y)+c]) - int(cur.Pix[cur.PixOffset(cur.Rect.Min.X+left, cur.Rect.Min.Y+y)+c])
				}
			}
		}
		for x := 0; x < w; x++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			for c := 0; c < 4; c++ {
				sum := 0
				for dy := -r; dy <= r; dy++ {
					sum += int(a.Pix[a.PixOffset(x, min(h-1, max(0, dy)))+c])
				}
				for y := 0; y < h; y++ {
					z.Pix[z.PixOffset(x, y)+c] = uint8(math.Round(float64(float32(sum) * weight)))
					top, bottom := max(0, y-r), min(h-1, y+r+1)
					sum += int(a.Pix[a.PixOffset(x, bottom)+c]) - int(a.Pix[a.PixOffset(x, top)+c])
				}
			}
		}
		cur = z
	}
	return cur, nil
}
