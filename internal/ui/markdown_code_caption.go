package orbitui

import "strings"

// Exact caption aliases from pinned Rust highlight.rs::language_label; these
// labels never replace the parser/tokenizer's original language tag.
var sourceCodeCaptions = map[string]string{
	"rust":        "Rust",
	"rs":          "Rust",
	"js":          "JavaScript",
	"javascript":  "JavaScript",
	"mjs":         "JavaScript",
	"cjs":         "JavaScript",
	"node":        "JavaScript",
	"jsx":         "JSX",
	"ts":          "TypeScript",
	"typescript":  "TypeScript",
	"mts":         "TypeScript",
	"cts":         "TypeScript",
	"tsx":         "TSX",
	"py":          "Python",
	"python":      "Python",
	"python3":     "Python",
	"go":          "Go",
	"golang":      "Go",
	"c":           "C",
	"h":           "C",
	"cc":          "C++",
	"cpp":         "C++",
	"c++":         "C++",
	"cxx":         "C++",
	"hpp":         "C++",
	"cs":          "C#",
	"csharp":      "C#",
	"c#":          "C#",
	"java":        "Java",
	"kt":          "Kotlin",
	"kotlin":      "Kotlin",
	"ruby":        "Ruby",
	"rb":          "Ruby",
	"gemfile":     "Ruby",
	"rake":        "Ruby",
	"swift":       "Swift",
	"json":        "JSON",
	"jsonc":       "JSON",
	"json5":       "JSON",
	"yaml":        "YAML",
	"yml":         "YAML",
	"toml":        "TOML",
	"ini":         "INI",
	"cfg":         "INI",
	"sh":          "Shell",
	"shell":       "Shell",
	"shellscript": "Shell",
	"console":     "Shell",
	"bash":        "Bash",
	"zsh":         "Zsh",
	"fish":        "Fish",
	"dockerfile":  "Dockerfile",
	"docker":      "Dockerfile",
	"makefile":    "Makefile",
	"make":        "Makefile",
	"sql":         "SQL",
	"postgres":    "SQL",
	"postgresql":  "SQL",
	"mysql":       "SQL",
	"sqlite":      "SQL",
	"html":        "HTML",
	"htm":         "HTML",
	"xml":         "XML",
	"svg":         "SVG",
	"vue":         "Vue",
	"svelte":      "Svelte",
	"css":         "CSS",
	"scss":        "SCSS",
	"sass":        "Sass",
	"less":        "Less",
	"md":          "Markdown",
	"markdown":    "Markdown",
	"diff":        "Diff",
	"patch":       "Diff",
}

func asciiCodeTag(tag string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, tag)
}

func (a *App) codeCaption(tag string) string {
	if tag == "" {
		return a.Tr("transcript_view.plain_text")
	}
	if asciiCodeTag(tag) == "mermaid" {
		return a.Tr("transcript_view.mermaid_source")
	}
	if caption, ok := sourceCodeCaptions[asciiCodeTag(strings.TrimSpace(tag))]; ok {
		return caption
	}
	return tag
}
