package highlight

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func diffLine(line string) []Token {
	if line == "" {
		return nil
	}
	var class Class
	switch {
	case line[0] == '+' && !strings.HasPrefix(line, "+++"):
		class = Added
	case line[0] == '-' && !strings.HasPrefix(line, "---"):
		class = Removed
	case line[0] == '@':
		class = Meta
	case strings.ContainsRune("+-di", rune(line[0])):
		class = Comment
	default:
		return nil
	}
	return []Token{{0, len(line), class}}
}

var shellOperators = strings.Fields("2>> 2>&1 2> 1>> 1>&2 1> &>> &> >> >& |& || && > < | ; & ( ) { } !")

func shellLine(line string) []Token {
	var result []Token
	command := true
	for at := 0; at < len(line); {
		rest := line[at:]
		r, size := utf8.DecodeRuneInString(rest)
		if unicode.IsSpace(r) {
			at += size
			continue
		}
		end := at
		var class Class
		switch r {
		case '#':
			end = len(line)
			class = Comment
		case '$':
			end = at + 1
			class = Meta
			if strings.HasPrefix(rest, "${") {
				n := strings.IndexByte(rest, '}')
				if n < 0 {
					end = len(line)
				} else {
					end = at + n + 1
				}
			} else {
				for end < len(line) && (letter(line[end]) || digit(line[end]) || strings.ContainsRune("_?#@*!$", rune(line[end]))) {
					end++
				}
			}
		case '\'', '"':
			end = quoteEnd(line, at+1, quote{string(r), false, r == '"'})
			if end < 0 {
				end = len(line)
			}
			class = String
			command = false
		default:
			op := ""
			for _, candidate := range shellOperators {
				if strings.HasPrefix(rest, candidate) {
					op = candidate
					break
				}
			}
			if op != "" {
				end = at + len(op)
				class = Operator
				switch op {
				case "|", "|&", "||", "&&", ";", "&", "(":
					command = true
				}
			} else {
				for end < len(line) {
					ch, n := utf8.DecodeRuneInString(line[end:])
					if unicode.IsSpace(ch) || strings.ContainsRune("|&;<> (){}\"'`", ch) {
						break
					}
					end += n
				}
				if end == at {
					at += size
					continue
				}
				word := line[at:end]
				switch {
				case strings.Contains(word, "=") && !strings.HasPrefix(word, "-") && !strings.HasPrefix(word, "/"):
					class = Type
				case command:
					class = Command
					command = false
				case strings.HasPrefix(word, "-") && len(word) > 1:
					class = Flag
				case strings.HasPrefix(word, "~") || strings.Contains(word, "/"):
					class = Path
				}
			}
		}
		if class != "" {
			result = appendToken(result, at, end, class)
		}
		at = end
	}
	return result
}
