package composerui

import "strings"

type Badge struct {
	Label       string
	Dark, Light uint32
}

var badges = map[string]Badge{
	"md":            {"MD", 0x519ABA, 0x3B72A8},
	"markdown":      {"MD", 0x519ABA, 0x3B72A8},
	"mdx":           {"MD", 0x519ABA, 0x3B72A8},
	"txt":           {"TXT", 0x8A8A94, 0x6E6E78},
	"pdf":           {"PDF", 0xD93831, 0xB02A24},
	"ts":            {"TS", 0x3178C6, 0x2A66A8},
	"tsx":           {"TSX", 0x3178C6, 0x2A66A8},
	"d.ts":          {"TS", 0x3178C6, 0x2A66A8},
	"js":            {"JS", 0xF1E05A, 0xA9941A},
	"mjs":           {"JS", 0xF1E05A, 0xA9941A},
	"cjs":           {"JS", 0xF1E05A, 0xA9941A},
	"jsx":           {"JSX", 0xF1E05A, 0xB29310},
	"json":          {"JSON", 0xCBCB41, 0x8F8F1E},
	"jsonc":         {"JSON", 0xCBCB41, 0x8F8F1E},
	"json5":         {"JSON", 0xCBCB41, 0x8F8F1E},
	"html":          {"HTML", 0xE44D26, 0xC43D1B},
	"htm":           {"HTML", 0xE44D26, 0xC43D1B},
	"css":           {"CSS", 0x563D7C, 0x563D7C},
	"scss":          {"SCSS", 0xC6538C, 0xA8447A},
	"sass":          {"SCSS", 0xC6538C, 0xA8447A},
	"less":          {"LESS", 0x2A5D8F, 0x2A5D8F},
	"vue":           {"VUE", 0x41B883, 0x35986C},
	"svelte":        {"SVELTE", 0xFF3E00, 0xD63400},
	"astro":         {"ASTRO", 0xFF5D01, 0xD64E00},
	"rs":            {"RS", 0xDEA584, 0xB5744F},
	"py":            {"PY", 0x3572A5, 0x2A5B85},
	"go":            {"GO", 0x00ADD8, 0x008DB2},
	"rb":            {"RB", 0xCC342D, 0xA82A24},
	"php":           {"PHP", 0x4F5D95, 0x414C7C},
	"java":          {"JAVA", 0xB07219, 0x8E5B13},
	"kt":            {"KT", 0xA97BFF, 0x7C4FD9},
	"kts":           {"KT", 0xA97BFF, 0x7C4FD9},
	"swift":         {"SWIFT", 0xF05138, 0xC93F28},
	"c":             {"C", 0x8D8D9B, 0x62626E},
	"h":             {"C", 0x8D8D9B, 0x62626E},
	"cpp":           {"C++", 0xF34B7D, 0xD22A5E},
	"hpp":           {"C++", 0xF34B7D, 0xD22A5E},
	"cc":            {"C++", 0xF34B7D, 0xD22A5E},
	"cxx":           {"C++", 0xF34B7D, 0xD22A5E},
	"cs":            {"CS", 0x178600, 0x106B00},
	"sh":            {"SH", 0x89E051, 0x4E9424},
	"bash":          {"SH", 0x89E051, 0x4E9424},
	"zsh":           {"SH", 0x89E051, 0x4E9424},
	"fish":          {"SH", 0x89E051, 0x4E9424},
	"toml":          {"TOML", 0x9C4121, 0x7C331A},
	"yaml":          {"YML", 0xCB171E, 0xA11218},
	"yml":           {"YML", 0xCB171E, 0xA11218},
	"lock":          {"LOCK", 0x7D7D7D, 0x6E6E78},
	"env":           {"ENV", 0xECD53F, 0x8F8414},
	"sql":           {"SQL", 0xDD7C37, 0xB85F20},
	"csv":           {"CSV", 0x237346, 0x1B5C38},
	"tsv":           {"CSV", 0x237346, 0x1B5C38},
	"png":           {"IMG", 0xA074C4, 0x7C4FA0},
	"jpg":           {"IMG", 0xA074C4, 0x7C4FA0},
	"jpeg":          {"IMG", 0xA074C4, 0x7C4FA0},
	"gif":           {"IMG", 0xA074C4, 0x7C4FA0},
	"webp":          {"IMG", 0xA074C4, 0x7C4FA0},
	"bmp":           {"IMG", 0xA074C4, 0x7C4FA0},
	"ico":           {"IMG", 0xA074C4, 0x7C4FA0},
	"svg":           {"SVG", 0xFFB13B, 0xD68F1F},
	"hbs":           {"HBS", 0x8A8A94, 0x6E6E78},
	"gitignore":     {"GIT", 0xF14E32, 0xD13E22},
	"gitattributes": {"GIT", 0xF14E32, 0xD13E22},
}

func FileBadge(path string) Badge {
	i := strings.LastIndex(path, ".")
	if i >= 0 {
		if b, ok := badges[strings.ToLower(path[i+1:])]; ok {
			return b
		}
	}
	return Badge{"FILE", 0x8A8A94, 0x6E6E78}
}
