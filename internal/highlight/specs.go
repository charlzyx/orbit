package highlight

import "strings"

// The language vocabulary and aliases are compatibility data from pinned Orbit,
// not a parser grammar. Intentionally retain source quirks (case-sensitive SQL,
// primitive Go types painted as literals, and no Markdown lexer).
var tagLanguages = func() map[string]Language {
	m := map[string]Language{}
	for lang, tags := range map[Language]string{
		Rust: "rust rs", Script: "js jsx mjs cjs javascript node ts tsx mts cts typescript", Python: "py python python3", Go: "go golang", C: "c h cc cpp c++ cxx hpp objc m", Java: "java kt kotlin scala cs csharp c#", Ruby: "rb ruby gemfile rake", Swift: "swift", JSON: "json jsonc json5", YAML: "yaml yml", TOML: "toml ini cfg", Shell: "sh bash zsh shell shellscript console fish dockerfile docker makefile make", CSS: "css scss sass less", HTML: "html htm xml svg vue svelte", SQL: "sql postgres postgresql mysql sqlite", Diff: "diff patch",
	} {
		for _, tag := range strings.Fields(tags) {
			m[tag] = lang
		}
	}
	return m
}()
var languageLabels = func() map[string]string {
	m := map[string]string{}
	for label, tags := range map[string]string{
		"Rust": "rust rs", "JavaScript": "js javascript mjs cjs node", "JSX": "jsx", "TypeScript": "ts typescript mts cts", "TSX": "tsx", "Python": "py python python3", "Go": "go golang", "C": "c h", "C++": "cc cpp c++ cxx hpp", "C#": "cs csharp c#", "Java": "java", "Kotlin": "kt kotlin", "Ruby": "ruby rb gemfile rake", "Swift": "swift", "JSON": "json jsonc json5", "YAML": "yaml yml", "TOML": "toml", "INI": "ini cfg", "Shell": "sh shell shellscript console", "Bash": "bash", "Zsh": "zsh", "Fish": "fish", "Dockerfile": "dockerfile docker", "Makefile": "makefile make", "SQL": "sql postgres postgresql mysql sqlite", "HTML": "html htm", "XML": "xml", "SVG": "svg", "Vue": "vue", "Svelte": "svelte", "CSS": "css", "SCSS": "scss", "Sass": "sass", "Less": "less", "Markdown": "md markdown", "Diff": "diff patch",
	} {
		for _, tag := range strings.Fields(tags) {
			m[tag] = label
		}
	}
	return m
}()
var specs = func() [17]languageSpec {
	var all [17]languageSpec
	for i := range all {
		all[i] = defaultSpec()
	}
	set := func(lang Language, keywords, literals string) *languageSpec {
		s := &all[lang]
		s.keywords = words(keywords)
		s.literals = words(literals)
		return s
	}
	noBlock := func(s *languageSpec) { s.blockOpen = ""; s.blockClose = "" }
	simple := func(s *languageSpec) { s.capitalTypes = false; s.calls = false }
	single := quote{"'", false, true}
	double := quote{"\"", false, true}
	tick := quote{"`", true, true}
	triple := quote{"\"\"\"", true, true}
	tripleSingle := quote{"'''", true, true}
	s := set(Rust, "as async await break const continue crate dyn else enum extern fn for if impl in let loop match mod move mut pub ref return self Self static struct super trait type unsafe use where while yield macro_rules", "true false None Some Ok Err")
	s.types = words("bool char f32 f64 i8 i16 i32 i64 i128 isize str u8 u16 u32 u64 u128 usize")
	s.meta = "'"
	s = set(Script, "as async await break case catch class const continue debugger declare default delete do else enum export extends finally for from function get if implements import in infer instanceof interface keyof let namespace new of private protected public readonly return satisfies set static super switch this throw try type typeof var void while yield", "true false null undefined NaN Infinity")
	s.quotes = []quote{double, single, tick}
	s.extra = "_$"
	s = set(Python, "and as assert async await break class continue def del elif else except finally for from global if import in is lambda match nonlocal not or pass raise return try while with yield", "True False None self cls")
	s.comments = []string{"#"}
	noBlock(s)
	s.quotes = []quote{triple, tripleSingle, double, single}
	s.meta = "@"
	s = set(Go, "break case chan const continue default defer else fallthrough for func go goto if import interface map package range return select struct switch type var", "true false nil iota bool byte error float32 float64 int int8 int16 int32 int64 rune string uint uint8 uint16 uint32 uint64")
	s.quotes = []quote{double, tick, single}
	s = set(C, "auto break case class const constexpr continue default delete do else enum explicit extern for friend goto if inline namespace new operator override private protected public register return sizeof static struct switch template this throw try typedef typename union using virtual volatile while", "true false NULL nullptr bool char double float int long short signed unsigned void size_t uint8_t uint32_t uint64_t")
	s.meta = "#"
	s = set(Java, "abstract as assert break case catch class companion const continue data def default do else enum extends final finally for fun if implements import instanceof interface internal is lateinit native new object open operator override package private protected public return sealed static super suspend switch synchronized this throw throws trait transient try val var volatile when while yield", "true false null boolean byte char double float int long short void String Unit")
	s.meta = "@"
	s = set(Ruby, "alias and begin break case class def defined? do else elsif end ensure for if in module next not or redo rescue retry return then undef unless until when while yield require require_relative attr_accessor attr_reader attr_writer", "true false nil self __FILE__ __dir__")
	s.comments = []string{"#"}
	noBlock(s)
	s.extra = "_?!"
	s.meta = "@$:"
	s = set(Swift, "actor as associatedtype async await break case catch class continue default defer deinit do else enum extension fallthrough fileprivate final for func guard if import in indirect init inout internal is lazy let mutating nonisolated open operator private protocol public repeat rethrows return self static struct subscript super switch throw throws try typealias var where while", "true false nil Any Bool Double Int String Void some")
	s.quotes = []quote{triple, double}
	s.meta = "@#"
	s = set(JSON, "", "true false null")
	s.quotes = []quote{double}
	simple(s)
	s = set(YAML, "", "true false null yes no on off ~")
	s.comments = []string{"#"}
	noBlock(s)
	simple(s)
	s = set(TOML, "", "true false")
	s.comments = []string{"#"}
	noBlock(s)
	s.quotes = []quote{triple, double, single}
	simple(s)
	// Shell has its own role-aware scanner. Its generic vocabulary is unused.
	s = set(CSS, "important media keyframes import supports font-face from to and not only", "auto none inherit initial unset transparent")
	s.comments = nil
	s.extra = "_-"
	s.meta = "@$#.&"
	s.capitalTypes = false
	s = set(HTML, "", "")
	s.comments = nil
	s.blockOpen = "<!--"
	s.blockClose = "-->"
	s.extra = "_-:"
	simple(s)
	s = set(SQL, "add all alter and as asc begin between by case cast column commit constraint create cross default delete desc distinct drop else end exists foreign from full group having if in index inner insert into is join key left like limit not offset on or order outer primary references returning right rollback select set table then transaction union unique update values view when where with", "null true false")
	s.comments = []string{"--", "#"}
	s.quotes = []quote{single, double}
	s.capitalTypes = false
	return all
}()
