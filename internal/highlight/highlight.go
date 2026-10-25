// Package highlight implements the pinned Orbit paint-only lexical contract.
// It returns UTF-8 byte spans, never fonts, dimensions, wrapping or layout.
// This is an independent Go scanner checked against Orbit's executable baseline;
// no Waku implementation is incorporated (its historical license is unverified).
package highlight

import (
	"crypto/sha256"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

type Language uint8

const (
	Unknown Language = iota
	Rust
	Script
	Python
	Go
	C
	Java
	Ruby
	Swift
	JSON
	YAML
	TOML
	Shell
	CSS
	HTML
	SQL
	Diff
)

type Class string

const (
	Keyword  Class = "keyword"
	Literal  Class = "literal"
	String   Class = "string"
	Comment  Class = "comment"
	Number   Class = "number"
	Type     Class = "type"
	Function Class = "function"
	Meta     Class = "meta"
	Command  Class = "command"
	Flag     Class = "flag"
	Path     Class = "path"
	Operator Class = "operator"
	Added    Class = "added"
	Removed  Class = "removed"
)

type Token struct {
	Start, End int
	Class      Class
}
type CarryKind uint8

const (
	NoCarry CarryKind = iota
	BlockComment
	MultilineString
)

type Carry struct {
	Kind       CarryKind
	StringSlot uint8
}

type quote struct {
	delimiter          string
	multiline, escapes bool
}
type languageSpec struct {
	comments                  []string
	blockOpen, blockClose     string
	quotes                    []quote
	keywords, literals, types map[string]bool
	extra, meta               string
	capitalTypes, calls       bool
}

func words(s string) map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(s) {
		m[w] = true
	}
	return m
}
func defaultSpec() languageSpec {
	return languageSpec{comments: []string{"//"}, blockOpen: "/*", blockClose: "*/", quotes: []quote{{"\"", false, true}, {"'", false, true}}, keywords: words(""), literals: words("true false null"), types: words(""), extra: "_", capitalTypes: true, calls: true}
}

func LanguageForTag(tag string) Language {
	return tagLanguages[strings.ToLower(strings.TrimSpace(tag))]
}
func LanguageForPath(path string) Language {
	name := strings.ToLower(filepath.Base(path))
	switch name {
	case "makefile":
		return Shell
	case "dockerfile":
		return Shell
	case "cargo.lock":
		return TOML
	case "package-lock.json", "composer.lock":
		return JSON
	}
	return LanguageForTag(strings.TrimPrefix(filepath.Ext(path), "."))
}
func LanguageLabel(tag string) string { return languageLabels[strings.ToLower(strings.TrimSpace(tag))] }

func appendToken(tokens []Token, start, end int, class Class) []Token {
	if start >= end {
		return tokens
	}
	if n := len(tokens); n > 0 && tokens[n-1].End == start && tokens[n-1].Class == class {
		tokens[n-1].End = end
		return tokens
	}
	return append(tokens, Token{start, end, class})
}
func alphabetic(r rune) bool { return unicode.IsLetter(r) || unicode.Is(unicode.Other_Alphabetic, r) }
func uppercase(r rune) bool  { return unicode.IsUpper(r) || unicode.Is(unicode.Other_Uppercase, r) }
func identifier(s string, extra string) int {
	end := 0
	for i, r := range s {
		allowed := alphabetic(r) || strings.ContainsRune(extra, r)
		if i > 0 {
			allowed = allowed || unicode.IsNumber(r)
		}
		if !allowed {
			break
		}
		end = i + utf8.RuneLen(r)
	}
	return end
}
func nextRune(s string) rune {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return r
		}
	}
	return 0
}
func digit(b byte) bool  { return b >= '0' && b <= '9' }
func letter(b byte) bool { return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' }
func numberEnd(s string) int {
	i := 1
	if len(s) > 1 && s[0] == '0' && strings.ContainsRune("xXbBoO", rune(s[1])) {
		i = 2
		for i < len(s) && (digit(s[i]) || letter(s[i]) || s[i] == '_') {
			i++
		}
		return i
	}
	dot := false
	for i < len(s) {
		b := s[i]
		switch {
		case digit(b) || b == '_':
			i++
		case b == '.' && !dot && i+1 < len(s) && digit(s[i+1]):
			dot = true
			i++
		case (b == 'e' || b == 'E') && i+1 < len(s) && (digit(s[i+1]) || s[i+1] == '+' || s[i+1] == '-'):
			i += 2
		case letter(b):
			i++
		default:
			return i
		}
	}
	return i
}
func quoteEnd(line string, start int, q quote) int {
	for i := start; i < len(line); i++ {
		if q.escapes && line[i] == '\\' {
			i++
			continue
		}
		if strings.HasPrefix(line[i:], q.delimiter) {
			return i + len(q.delimiter)
		}
	}
	return -1
}

// TokenizeLine is linear in the line length. Plain text is represented by gaps
// between tokens. The Carry refers only to the selected language's quote slots.
func TokenizeLine(lang Language, line string, carry Carry) ([]Token, Carry) {
	if lang == Unknown || int(lang) >= len(specs) {
		return nil, Carry{}
	}
	if lang == Diff {
		return diffLine(line), Carry{}
	}
	if lang == Shell {
		return shellLine(line), Carry{}
	}
	sp := specs[lang]
	var tokens []Token
	i := 0
	if carry.Kind == BlockComment {
		if sp.blockClose == "" {
			return nil, Carry{}
		}
		n := strings.Index(line, sp.blockClose)
		if n < 0 {
			return appendToken(tokens, 0, len(line), Comment), carry
		}
		i = n + len(sp.blockClose)
		tokens = appendToken(tokens, 0, i, Comment)
		carry = Carry{}
	} else if carry.Kind == MultilineString {
		if int(carry.StringSlot) >= len(sp.quotes) {
			return nil, Carry{}
		}
		n := quoteEnd(line, 0, sp.quotes[carry.StringSlot])
		if n < 0 {
			return appendToken(tokens, 0, len(line), String), carry
		}
		i = n
		tokens = appendToken(tokens, 0, i, String)
		carry = Carry{}
	}
	hashComment := false
	for _, c := range sp.comments {
		hashComment = hashComment || c == "#"
	}
	if i == 0 && strings.ContainsRune(sp.meta, '#') && !hashComment && strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "#") {
		return appendToken(tokens, 0, len(line), Meta), Carry{}
	}
	for i < len(line) {
		rest := line[i:]
		isComment := false
		for _, prefix := range sp.comments {
			if strings.HasPrefix(rest, prefix) {
				isComment = true
				break
			}
		}
		if isComment {
			tokens = appendToken(tokens, i, len(line), Comment)
			break
		}
		if sp.blockOpen != "" && strings.HasPrefix(rest, sp.blockOpen) {
			start := i + len(sp.blockOpen)
			n := strings.Index(line[start:], sp.blockClose)
			if n < 0 {
				return appendToken(tokens, i, len(line), Comment), Carry{Kind: BlockComment}
			}
			end := start + n + len(sp.blockClose)
			tokens = appendToken(tokens, i, end, Comment)
			i = end
			continue
		}
		slot := -1
		for k, q := range sp.quotes {
			if strings.HasPrefix(rest, q.delimiter) {
				slot = k
				break
			}
		}
		if slot >= 0 {
			q := sp.quotes[slot]
			end := quoteEnd(line, i+len(q.delimiter), q)
			if end < 0 {
				tokens = appendToken(tokens, i, len(line), String)
				if q.multiline {
					return tokens, Carry{MultilineString, uint8(slot)}
				}
				break
			}
			tokens = appendToken(tokens, i, end, String)
			i = end
			continue
		}
		ch, size := utf8.DecodeRuneInString(rest)
		if digit(rest[0]) {
			end := i + numberEnd(rest)
			tokens = appendToken(tokens, i, end, Number)
			i = end
			continue
		}
		if strings.ContainsRune(sp.meta, ch) {
			n := identifier(rest[size:], sp.extra)
			if n > 0 {
				end := i + size + n
				tokens = appendToken(tokens, i, end, Meta)
				i = end
				continue
			}
		}
		if n := identifier(rest, sp.extra); n > 0 {
			word := rest[:n]
			var class Class
			switch {
			case sp.keywords[word]:
				class = Keyword
			case sp.literals[word]:
				class = Literal
			case sp.types[word]:
				class = Type
			case sp.calls && nextRune(rest[n:]) == '(':
				class = Function
			case sp.capitalTypes && uppercase(ch):
				class = Type
			}
			if class != "" {
				tokens = appendToken(tokens, i, i+n, class)
			}
			i += n
			continue
		}
		i += size
	}
	return tokens, carry
}
func Tokenize(lang Language, code string) [][]Token {
	lines := strings.Split(code, "\n")
	result := make([][]Token, len(lines))
	carry := Carry{}
	for i, line := range lines {
		result[i], carry = TokenizeLine(lang, line, carry)
	}
	return result
}

type cacheKey struct {
	lang Language
	hash [32]byte
}

// Block is immutable after construction. Callers must not mutate Lines/tokens.
type Block struct{ Lines [][]Token }

var blockCache = struct {
	sync.Mutex
	entries map[cacheKey]*Block
}{entries: map[cacheKey]*Block{}}

// TokenizeCached avoids repeated tokenization of unchanged streaming blocks. It
// never changes text layout. As in the source, the memo is bounded to 256 blocks.
func TokenizeCached(lang Language, code string) *Block {
	key := cacheKey{lang, sha256.Sum256([]byte(code))}
	blockCache.Lock()
	defer blockCache.Unlock()
	if b := blockCache.entries[key]; b != nil {
		return b
	}
	b := &Block{Lines: Tokenize(lang, code)}
	if len(blockCache.entries) >= 256 {
		clear(blockCache.entries)
	}
	blockCache.entries[key] = b
	return b
}
