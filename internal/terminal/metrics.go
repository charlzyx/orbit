package terminal

import (
	"math"
	"strings"

	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

// Orbit lays terminal rows out at font size × 1.4, not at the font's own
// ascent + descent + line gap × 1.4. Ghostty's LineHeight uses the latter,
// so passing 1.4 directly changes the number of PTY rows with the font family.
const sourceTerminalLineHeight float32 = 1.4

func terminalFont(f ghostty.Font) ghostty.Font {
	if f.Size <= 0 {
		f.Size = 13
	}
	return fontAtRowHeight(f, f.Size*sourceTerminalLineHeight)
}

// fontAtRowHeight converts Orbit's DIP pitch into the pinned plugin's metric
// multiplier. Ghostty still quantizes cells to whole device pixels using
// ceil(pitch*scale-.05); fractional-DIP GPUI rows are not exactly representable.
func fontAtRowHeight(f ghostty.Font, pitch float32) ghostty.Font {
	if strings.TrimSpace(f.Family) == "" {
		f.Family = "monospace"
	}
	if f.Size <= 0 {
		f.Size = 13
	}
	if f.Weight <= 0 {
		f.Weight = 400
	}
	f.Weight = min(max(f.Weight, 100), 900)
	f.Features = append([]string(nil), f.Features...)
	m := (ui.Font{Family: f.Family, Size: f.Size, Weight: f.Weight}).Metrics()
	natural := m.Ascent + m.Descent + m.LineGap
	if natural > 0 && pitch > 0 {
		f.LineHeight = pitch / natural
	} else {
		f.LineHeight = 1
	}
	return f
}

// gridRowHeight mirrors the pinned plugin's device-pixel quantization. Keep
// it explicit rather than calling fractional rows an exact source match.
func gridRowHeight(pitch, scale float32) float32 {
	if scale <= 0 {
		scale = 1
	}
	return max(float32(math.Ceil(float64(pitch*scale-.05))), 1) / scale
}
