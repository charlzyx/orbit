package highlight

import "unicode/utf8"

// Run is a lossless paint segment. An empty Class uses the default foreground.
// All metrics remain the caller's existing font/size/weight/line-height.
type Run struct {
	Text  string
	Class Class
}

// Runs fills plain gaps around trusted token spans. Invalid, overlapping, or
// non-UTF-8-boundary spans are ignored rather than losing or duplicating text.
// It preserves every byte, including malformed input from an external file.
func Runs(line string, tokens []Token) []Run {
	var result []Run
	at := 0
	boundary := func(n int) bool { return n == 0 || n == len(line) || n > 0 && n < len(line) && utf8.RuneStart(line[n]) }
	for _, t := range tokens {
		if t.Start < at || t.End <= t.Start || t.End > len(line) || !boundary(t.Start) || !boundary(t.End) {
			continue
		}
		if at < t.Start {
			result = append(result, Run{Text: line[at:t.Start]})
		}
		result = append(result, Run{line[t.Start:t.End], t.Class})
		at = t.End
	}
	if at < len(line) || len(result) == 0 {
		result = append(result, Run{Text: line[at:]})
	}
	return result
}
