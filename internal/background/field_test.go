package background

import (
	"image"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestProductionMarkGeometry(t *testing.T) {
	m, e := ProductionMark()
	if e != nil {
		t.Fatal(e)
	}
	if m.columns != 64 || m.Aspect() <= 2 || m.Aspect() >= 2.4 {
		t.Fatal(m.columns, m.rows, m.Aspect())
	}
	var sum float32
	for _, v := range m.cells {
		sum += v
	}
	mean := sum / float32(len(m.cells))
	if mean <= .55 || mean >= .85 {
		t.Fatal(mean)
	}
	if m.Coverage(.21, .45) >= .5 || m.Coverage(.56, .90) <= .5 {
		t.Fatal("production monogram counter/stem altered")
	}
	if _, e = DeriveMark(image.NewNRGBA(image.Rect(0, 0, 32, 32))); e == nil {
		t.Fatal("empty mark accepted")
	}
}
func TestFieldCoverageMotionAndMarkSweep(t *testing.T) {
	f, e := NewField()
	if e != nil {
		t.Fatal(e)
	}
	fr := f.frame(1280, 820)
	for _, u := range []float32{.004, .05, .25, .5, .75, .95, .996} {
		for _, v := range []float32{.004, .05, .25, .5, .75, .95, .996} {
			for _, phase := range []float32{0, 7, 31.5, 300} {
				x, _ := f.sample(u, v, phase, fr)
				if x <= .02 {
					t.Fatal("empty field", u, v, phase, x)
				}
			}
		}
	}
	var marked, total, changed int
	for y := 0; y < 80; y++ {
		for x := 0; x < 80; x++ {
			u, v := (float32(x)+.5)/80, (float32(y)+.5)/80
			a, _ := f.sample(u, v, 0, frame{})
			b, _ := f.sample(u, v, 1, frame{})
			if a >= .60 {
				t.Fatal("saturated field")
			}
			la, lb := level(a, bayer8[y%8][x%8]), level(b, bayer8[y%8][x%8])
			if la > 0 {
				marked++
			}
			if la != lb {
				changed++
			}
			total++
		}
	}
	ratio := float64(changed) / float64(total)
	if ratio <= .002 || ratio >= .04 {
		t.Fatal("motion changed", ratio)
	}
	if float64(marked)/float64(total) < .6 {
		t.Fatal("insufficient screen coverage")
	}
	at := func(mu, phase float32) float32 {
		v, _ := f.sample(fr.minU+(fr.maxU-fr.minU)*mu, fr.minV+(fr.maxV-fr.minV)*.2, phase, fr)
		return v
	}
	if at(.05, 5) <= at(.95, 5) || at(.95, 15) <= at(.05, 15) {
		t.Fatal("light no longer crosses production mark")
	}
	if f.frame(360, 300).mark || f.frame(1280, 360).mark || !fr.mark {
		t.Fatal("small frame mark bounds")
	}
}
func TestDitherLevelsAndNoFrameAllocation(t *testing.T) {
	var previous float32
	seen := map[int]bool{}
	for step := 0; step <= 40; step++ {
		v := float32(step) / 40
		var sum float32
		for _, row := range bayer8 {
			for _, threshold := range row {
				l := level(v, threshold)
				seen[l] = true
				sum += strengths[l]
			}
		}
		if sum < previous-float32(math.SmallestNonzeroFloat32) {
			t.Fatal("nonmonotonic screen")
		}
		previous = sum
	}
	if len(seen) != 4 {
		t.Fatal(seen)
	}
	f, _ := NewField()
	count := 0
	visit := func(Dot) { count++ }
	allocs := testing.AllocsPerRun(10, func() { f.Dots(1280, 820, 0, 9, visit) })
	if allocs != 0 {
		t.Fatal("field allocates per frame", allocs)
	}
	if count == 0 {
		t.Fatal("field blank")
	}
}
func TestReduceMotionFreezesCurrentPhase(t *testing.T) {
	var c Clock
	t0 := time.Unix(100, 0)
	if c.Phase(t0, false) != 0 {
		t.Fatal("initialphase")
	}
	if c.Phase(t0.Add(3*time.Second), true) != 3 {
		t.Fatal("freeze jumps")
	}
	if c.Phase(t0.Add(50*time.Second), true) != 3 {
		t.Fatal("moves while reduced")
	}
	if c.Phase(t0.Add(51*time.Second), false) != 3 {
		t.Fatal("unfreeze catches up")
	}
	if c.Phase(t0.Add(52*time.Second), false) != 4 {
		t.Fatal("unfreeze did not resume")
	}
}
func TestNativeFieldPaint(t *testing.T) {
	f, e := NewField()
	if e != nil {
		t.Fatal(e)
	}
	tester := ui.NewTester(func(c *ui.Context) {
		ui.Box(c).Fill().Background(ui.Hex("#181818")).Draw(func(p *ui.Painter, r ui.Rect) { f.Paint(p, r, 12.5, 1, ui.Hex("#808080"), true) })
	}, 900, 620)
	tester.Frame()
	im := tester.Image()
	if im.Bounds().Dx() != 900 || im.Bounds().Dy() != 620 {
		t.Fatal(im.Bounds())
	}
	seen := map[uint32]bool{}
	for y := 0; y < 620; y += 3 {
		for x := 0; x < 900; x += 3 {
			r, _, _, _ := im.At(x, y).RGBA()
			seen[r] = true
		}
	}
	if len(seen) < 4 {
		t.Fatal("no native dither levels", seen)
	}
}
