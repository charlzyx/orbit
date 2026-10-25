// Package background implements Orbit's original procedural and photo backdrops.
// Geometry/formulas follow crates/orbit-pi/src/{backdrop,dither}.rs at 299a32b
// (Apache-2.0); native painting is Go/MyGo, without a webview or Rust runtime.
package background

import (
	"bytes"
	"errors"
	"image"
	"math"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

var bayer8 = [8][8]uint8{{0, 32, 8, 40, 2, 34, 10, 42}, {48, 16, 56, 24, 50, 18, 58, 26}, {12, 44, 4, 36, 14, 46, 6, 38}, {60, 28, 52, 20, 62, 30, 54, 22}, {3, 35, 11, 43, 1, 33, 9, 41}, {51, 19, 59, 27, 49, 17, 57, 25}, {15, 47, 7, 39, 13, 45, 5, 37}, {63, 31, 55, 23, 61, 29, 53, 21}}
var radii = [4]float32{0, .10, .15, .20}
var strengths = [4]float32{0, .45, .70, 1}

type Mark struct {
	columns, rows int
	cells         []float32
}

func (m *Mark) Aspect() float32 { return float32(m.columns) / float32(m.rows) }
func (m *Mark) Coverage(u, v float32) float32 {
	x, y := u*float32(m.columns)-.5, v*float32(m.rows)-.5
	x0, y0 := int(math.Floor(float64(x))), int(math.Floor(float64(y)))
	fx, fy := x-float32(x0), y-float32(y0)
	at := func(x, y int) float32 {
		if x < 0 || y < 0 || x >= m.columns || y >= m.rows {
			return 0
		}
		return m.cells[y*m.columns+x]
	}
	raw := at(x0, y0)*(1-fx)*(1-fy) + at(x0+1, y0)*fx*(1-fy) + at(x0, y0+1)*(1-fx)*fy + at(x0+1, y0+1)*fx*fy
	return smooth(.35, .65, raw)
}

// DeriveMark thresholds/erodes the shipped production artwork, then averages it
// into the same64-column mask as Orbit. It never redraws the logo by hand.
func DeriveMark(img image.Image) (*Mark, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w < 7 || h < 7 || int64(w)*int64(h) > 16<<20 {
		return nil, errors.New("invalid backdrop mark dimensions")
	}
	on := make([]bool, w*h)
	horizontal := make([]bool, w*h)
	eroded := make([]bool, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			l := (2126*(r>>8) + 7152*(g>>8) + 722*(bl>>8)) / 10000
			on[y*w+x] = l > 128
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			keep := on[y*w+x]
			for d := 0; keep && d <= 3; d++ {
				if x < d || x+d >= w {
					keep = false
					break
				}
				keep = on[y*w+x-d] && on[y*w+x+d]
			}
			horizontal[y*w+x] = keep
		}
	}
	minx, miny, maxx, maxy := w, h, 0, 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			keep := horizontal[y*w+x]
			for d := 0; keep && d <= 3; d++ {
				if y < d || y+d >= h {
					keep = false
					break
				}
				keep = horizontal[(y-d)*w+x] && horizontal[(y+d)*w+x]
			}
			eroded[y*w+x] = keep
			if keep {
				minx = min(minx, x)
				miny = min(miny, y)
				maxx = max(maxx, x)
				maxy = max(maxy, y)
			}
		}
	}
	if minx >= maxx || miny >= maxy {
		return nil, errors.New("backdrop mark has no eroded glyphs")
	}
	bw, bh := maxx-minx+1, maxy-miny+1
	m := &Mark{columns: 64, rows: max(1, int(math.Round(float64(float32(64)*float32(bh)/float32(bw)))))}
	m.cells = make([]float32, m.columns*m.rows)
	for row := 0; row < m.rows; row++ {
		for col := 0; col < m.columns; col++ {
			x0 := minx + int(float32(col)*float32(bw)/float32(m.columns))
			x1 := minx + min(bw, int(math.Ceil(float64(float32(col+1)*float32(bw)/float32(m.columns)))))
			y0 := miny + int(float32(row)*float32(bh)/float32(m.rows))
			y1 := miny + min(bh, int(math.Ceil(float64(float32(row+1)*float32(bh)/float32(m.rows)))))
			var covered, total float32
			for y := y0; y < max(y1, y0+1); y++ {
				for x := x0; x < max(x1, x0+1); x++ {
					if x <= maxx && y <= maxy && eroded[y*w+x] {
						covered++
					}
					total++
				}
			}
			if total > 0 {
				m.cells[row*m.columns+col] = covered / total
			}
		}
	}
	return m, nil
}

var markOnce sync.Once
var defaultMark *Mark
var markError error

func ProductionMark() (*Mark, error) {
	markOnce.Do(func() {
		var b []byte
		b, markError = orbitassets.Asset("app-icon.png")
		if markError != nil {
			return
		}
		var im image.Image
		im, _, markError = image.Decode(bytes.NewReader(b))
		if markError == nil {
			defaultMark, markError = DeriveMark(im)
		}
	})
	return defaultMark, markError
}

type Field struct{ Mark *Mark }

func NewField() (*Field, error) {
	m, e := ProductionMark()
	if e != nil {
		return nil, e
	}
	return &Field{Mark: m}, nil
}

type frame struct {
	mark                   bool
	minU, minV, maxU, maxV float32
}

func (f *Field) frame(width, height float32) frame {
	if f.Mark == nil || width <= 0 || height <= 0 {
		return frame{}
	}
	mw := min(width*.36, float32(560))
	mh := mw / f.Mark.Aspect()
	mw = min(mw, width*.8)
	if mw < 200 || mh >= height*.45 {
		return frame{}
	}
	hu, hv := mw/width/2, mh/height/2
	return frame{true, .5 - hu, .57 - hv, .5 + hu, .57 + hv}
}
func smooth(a, b, x float32) float32 {
	t := max(float32(0), min(float32(1), (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}
func sin(x float32) float32 { return float32(math.Sin(float64(x))) }
func cos(x float32) float32 { return float32(math.Cos(float64(x))) }
func (f *Field) sample(u, v, phase float32, fr frame) (value, coverage float32) {
	const tau = float32(2 * math.Pi)
	broad := sin(tau*(u*.75+phase*.016)) * cos(tau*(v*.55-phase*.011))
	cross := sin(tau*((u*.45-v*.32)+phase*.007)) * cos(tau*((v*.38+u*.21)-phase*.005))
	wave := .68*broad + .32*cross
	vignette := smooth(1.15, .05, float32(math.Hypot(float64((u-.5)/.80), float64((v-.52)/.86))))
	vignette = .65 + (1-.65)*vignette
	page := max(float32(0), min(float32(1), (.30+.22*wave+.10*(v-.5))*vignette))
	if !fr.mark {
		return page, 0
	}
	mu, mv := (u-fr.minU)/(fr.maxU-fr.minU), (v-fr.minV)/(fr.maxV-fr.minV)
	if mu < 0 || mu > 1 || mv < 0 || mv > 1 {
		return page, 0
	}
	coverage = f.Mark.Coverage(mu, mv)
	if coverage <= 0 {
		return page, 0
	}
	phaseFraction := phase/20 - float32(math.Trunc(float64(phase/20)))
	sweep := -.4 + phaseFraction*1.8
	lit := smooth(.55, .30, float32(math.Abs(float64(mu-sweep))))
	glyph := .18 + (.85-.18)*lit
	return max(float32(0), min(float32(1), page*(1-coverage)+glyph*coverage)), coverage
}
func level(value float32, threshold uint8) int {
	scaled := value * 4
	base := float32(math.Floor(float64(scaled)))
	if scaled-base > (float32(threshold)+.5)/64 {
		base++
	}
	return max(0, min(3, int(base)))
}

type Dot struct{ X, Y, Radius, Alpha float32 }

// Dots visits native logical-pixel geometry without allocating a per-frame list.
func (f *Field) Dots(width, height, phase, pitch float32, visit func(Dot)) {
	if pitch <= 0 || width < pitch || height < pitch || math.IsNaN(float64(phase)) {
		return
	}
	fr := f.frame(width, height)
	cols, rows := int(math.Ceil(float64(width/pitch))), int(math.Ceil(float64(height/pitch)))
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			cx, cy := (float32(x)+.5)*pitch, (float32(y)+.5)*pitch
			value, mark := f.sample(cx/width, cy/height, phase, fr)
			l := level(value, bayer8[y%8][x%8])
			radius := pitch * radii[l]
			if radius > 0 {
				visit(Dot{cx, cy, radius, strengths[l] * (1 + (.85-1)*mark)})
			}
		}
	}
}
func (f *Field) Paint(p *ui.Painter, r ui.Rect, phase, scale float32, ink ui.Color, dark bool) {
	strength := float32(.62)
	if dark {
		strength = .55
	}
	ink = ink.Alpha(strength)
	f.Dots(r.W, r.H, phase, 9*scale, func(d Dot) {
		p.Fill(ui.Rect{X: r.X + d.X - d.Radius, Y: r.Y + d.Y - d.Radius, W: d.Radius * 2, H: d.Radius * 2}, ink.Alpha(d.Alpha), d.Radius)
	})
}

// Clock freezes the current phase for Reduce Motion instead of jumping to zero.
// Call on the UI thread; paint callbacks use Painter.Now for deterministic tests.
type Clock struct {
	last    time.Time
	elapsed time.Duration
	paused  bool
}

func (c *Clock) Phase(now time.Time, reduce bool) float32 {
	if !c.last.IsZero() && !c.paused && now.After(c.last) {
		c.elapsed += now.Sub(c.last)
	}
	c.last = now
	c.paused = reduce
	return float32(c.elapsed.Seconds())
}
