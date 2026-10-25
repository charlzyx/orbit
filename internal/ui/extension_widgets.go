package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strconv"
	"strings"
)

type widgetSGR struct {
	fg, bg                                        *ui.Color
	bold, italic, underline, strike, inverse, dim bool
}

func widgetSpans(line string, t Theme) []ui.Span {
	var spans []ui.Span
	var state widgetSGR
	emit := func(text string) {
		if text == "" {
			return
		}
		fg, bg := t.CodeText, ui.Color{}
		if state.fg != nil {
			fg = *state.fg
		}
		if state.bg != nil {
			bg = *state.bg
		}
		if state.inverse {
			fg = t.CodeBg
			if state.bg != nil {
				fg = *state.bg
			}
			bg = t.CodeText
			if state.fg != nil {
				bg = *state.fg
			}
		}
		if state.dim {
			fg = fg.Alpha(.7)
		}
		weight := 400
		if state.bold {
			weight = 700
		}
		spans = append(spans, ui.Span{Text: text, Color: fg, Background: bg, Weight: weight, Italic: state.italic, Underline: state.underline, Strikethrough: state.strike})
	}
	start := 0
	for i := 0; i < len(line); {
		if line[i] != 0x1b {
			i++
			continue
		}
		emit(line[start:i])
		i++
		if i >= len(line) {
			start = i
			break
		}
		switch line[i] {
		case '[':
			begin := i + 1
			j := begin
			for j < len(line) && (line[j] < 0x40 || line[j] > 0x7e) {
				j++
			}
			if j < len(line) {
				if line[j] == 'm' {
					state.apply(line[begin:j], t)
				}
				i = j + 1
			} else {
				i = len(line)
			}
		case ']', '_', 'P', '^', 'X':
			i++
			for i < len(line) {
				if line[i] == 7 {
					i++
					break
				}
				if line[i] == 0x1b && i+1 < len(line) && line[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}
		default:
			i++
		}
		start = i
	}
	emit(line[start:])
	if len(spans) == 0 {
		emit(" ")
	}
	return spans
}
func widgetIndexed(index int, t Theme) ui.Color {
	if index < 0 {
		index = 0
	}
	if index < 16 {
		return t.Terminal().Palette[index]
	}
	if index < 232 {
		n := index - 16
		component := func(n int) uint8 {
			if n == 0 {
				return 0
			}
			return uint8(55 + 40*n)
		}
		return ui.RGB(component(n/36), component(n/6%6), component(n%6))
	}
	if index < 256 {
		v := uint8(8 + 10*(index-232))
		return ui.RGB(v, v, v)
	}
	return t.CodeText
}
func (s *widgetSGR) apply(params string, t Theme) {
	pieces := strings.Split(params, ";")
	nums := make([]int, len(pieces))
	for i, p := range pieces {
		nums[i], _ = strconv.Atoi(p)
	}
	color := func(fg bool, c ui.Color) {
		if fg {
			s.fg = &c
		} else {
			s.bg = &c
		}
	}
	for i := 0; i < len(nums); i++ {
		n := nums[i]
		switch {
		case n == 0:
			*s = widgetSGR{}
		case n == 1:
			s.bold = true
		case n == 2:
			s.dim = true
		case n == 3:
			s.italic = true
		case n == 4:
			s.underline = true
		case n == 7:
			s.inverse = true
		case n == 9:
			s.strike = true
		case n == 22:
			s.bold = false
			s.dim = false
		case n == 23:
			s.italic = false
		case n == 24:
			s.underline = false
		case n == 27:
			s.inverse = false
		case n == 29:
			s.strike = false
		case n >= 30 && n <= 37:
			color(true, widgetIndexed(n-30, t))
		case n == 39:
			s.fg = nil
		case n >= 40 && n <= 47:
			color(false, widgetIndexed(n-40, t))
		case n == 49:
			s.bg = nil
		case n >= 90 && n <= 97:
			color(true, widgetIndexed(n-90+8, t))
		case n >= 100 && n <= 107:
			color(false, widgetIndexed(n-100+8, t))
		case n == 38 || n == 48:
			if i+2 < len(nums) && nums[i+1] == 5 {
				color(n == 38, widgetIndexed(nums[i+2], t))
				i += 2
			} else if i+4 < len(nums) && nums[i+1] == 2 {
				color(n == 38, ui.RGB(uint8(nums[i+2]), uint8(nums[i+3]), uint8(nums[i+4])))
				i += 4
			}
		}
	}
}
