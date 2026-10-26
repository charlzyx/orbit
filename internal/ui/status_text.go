package orbitui

import "strings"

// statusPlainText projects only the text of upstream #89 widgets::parse_line.
// It is not a terminal interpreter: CSI and string controls (including OSC52)
// are discarded, and no clipboard, URL or other side effect is performed.
func statusPlainText(line string) string {
	if !strings.ContainsRune(line, '\x1b') {
		return line
	}
	var out strings.Builder
	out.Grow(len(line))
	for i := 0; i < len(line); {
		if line[i] != '\x1b' {
			out.WriteByte(line[i])
			i++
			continue
		}
		i++
		if i == len(line) {
			break
		}
		switch line[i] {
		case '[':
			i++
			for i < len(line) && !(line[i] >= 0x40 && line[i] <= 0x7e) {
				i++
			}
			if i < len(line) {
				i++
			}
		case ']', '_', 'P', '^', 'X':
			i++
			for i < len(line) {
				if line[i] == '\a' {
					i++
					break
				}
				if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}
		default:
			// Keep source behavior for a two-byte ESC introducer; do not invent a
			// broader sanitizer that changes ordinary status text/Unicode semantics.
			i++
		}
	}
	return out.String()
}
