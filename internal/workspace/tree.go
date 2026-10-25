package workspace

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const MaxTreeEntries = 50000
const MaxTreeDepth = 24

type TreeSnapshot struct {
	Root      string   `json:"root"`
	Dirs      []string `json:"dirs"`
	Files     []string `json:"files"`
	Truncated bool     `json:"truncated"`
	Warnings  []string `json:"warnings,omitempty"`
}
type ignoreRule struct {
	base      string
	prefix    string
	pattern   *regexp.Regexp
	negative  bool
	directory bool
	priority  int
}

// gitGlob implements the path wildcard form shared by .ignore and .gitignore.
func gitGlob(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '*':
			if i+1 < len(s) && s[i+1] == '*' {
				i++
				if i+1 < len(s) && s[i+1] == '/' {
					i++
					b.WriteString("(?:.*/)?")
				} else {
					b.WriteString(".*")
				}
			} else {
				b.WriteString("[^/]*")
			}
		case '?':
			b.WriteString("[^/]")
		case '[':
			end := strings.IndexByte(s[i+1:], ']')
			if end >= 0 {
				end += i + 1
				chunk := s[i : end+1]
				if strings.HasPrefix(chunk, "[!") {
					chunk = "[^" + chunk[2:]
				}
				b.WriteString(chunk)
				i = end
			} else {
				b.WriteString(`\[`)
			}
		case '\\':
			if i+1 < len(s) {
				i++
				b.WriteString(regexp.QuoteMeta(string(s[i])))
			}
		default:
			b.WriteString(regexp.QuoteMeta(string(s[i])))
		}
	}
	return b.String()
}
func readIgnore(path, base string) []ignoreRule {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var rules []ignoreRule
	s := bufio.NewScanner(f)
	for s.Scan() {
		p := strings.TrimRight(s.Text(), "\r ")
		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		neg := strings.HasPrefix(p, "!")
		if neg {
			p = p[1:]
		}
		dir := strings.HasSuffix(p, "/")
		p = strings.TrimSuffix(p, "/")
		anchored := strings.HasPrefix(p, "/")
		p = strings.TrimPrefix(p, "/")
		if p == "" {
			continue
		}
		prefix := "(?:^|/)"
		if anchored || strings.Contains(p, "/") {
			prefix = "^"
		}
		r, err := regexp.Compile(prefix + gitGlob(p) + "$")
		if err == nil {
			priority := 0
			if filepath.Base(path) == ".ignore" {
				priority = 1
			}
			rules = append(rules, ignoreRule{base: base, pattern: r, negative: neg, directory: dir, priority: priority})
		}
	}
	return rules
}
func ignoreDecision(rel string, isDir bool, rules []ignoreRule) (value, matched bool, priority int) {
	priority = -100
	for _, r := range rules {
		local := rel
		if r.base != "" {
			if !strings.HasPrefix(rel, r.base+"/") {
				continue
			}
			local = strings.TrimPrefix(rel, r.base+"/")
		}
		if r.prefix != "" {
			local = r.prefix + "/" + local
		}
		found := false
		for p, dir := local, isDir; p != "" && p != "."; p, dir = filepath.ToSlash(filepath.Dir(p)), true {
			if (!r.directory || dir) && r.pattern.MatchString(p) {
				found = true
				break
			}
		}
		if found && r.priority >= priority {
			value = !r.negative
			matched = true
			priority = r.priority
		}
	}
	return
}
func ignored(rel string, isDir bool, rules []ignoreRule) bool {
	value, _, _ := ignoreDecision(rel, isDir, rules)
	return value
}

// Snapshot never follows symlinks and bounds depth/entry count. It honors
// ancestor/nested .gitignore/.ignore and Git's global/exclude rules when Git
// is available. Filesystem read failures are reported, not silently success.
func Snapshot(ctx context.Context, root string, showHidden bool) (TreeSnapshot, error) {
	abs, err := filepath.Abs(root)
	out := TreeSnapshot{Root: abs}
	if err != nil {
		return out, err
	}
	fi, err := os.Stat(abs)
	if err != nil {
		return out, err
	}
	if !fi.IsDir() {
		return out, errors.New("workspace is not a directory")
	}
	var baseRules []ignoreRule
	// Read low-precedence Git global/exclude rules before walking, so a huge
	// ignored tree cannot consume the entry cap before visible source files.
	gitPath := func(args ...string) string {
		cmd := exec.CommandContext(ctx, "git", append([]string{"-C", abs}, args...)...)
		hideGitConsole(cmd)
		raw, e := cmd.Output()
		if e != nil {
			return ""
		}
		return strings.TrimSpace(string(raw))
	}
	global := gitPath("config", "--path", "--get", "core.excludesfile")
	if global == "" {
		base := os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			if home, e := os.UserHomeDir(); e == nil {
				base = filepath.Join(home, ".config")
			}
		}
		if base != "" {
			global = filepath.Join(base, "git", "ignore")
		}
	}
	top := gitPath("rev-parse", "--show-toplevel")
	prefix := ""
	if top != "" {
		if p, e := filepath.Rel(top, abs); e == nil && p != "." {
			prefix = filepath.ToSlash(p)
		}
	}
	addLow := func(path string, priority int) {
		if path == "" {
			return
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(abs, path)
		}
		rules := readIgnore(path, "")
		for i := range rules {
			rules[i].priority = priority
			rules[i].prefix = prefix
		}
		baseRules = append(baseRules, rules...)
	}
	addLow(global, -2)
	addLow(gitPath("rev-parse", "--git-path", "info/exclude"), -1)
	ignoreOverrides := map[string]bool{}
	// Ancestor ignore files apply relative to their own directory; translate
	// each matching pattern through the actual path prefix when walking below.
	ancestors := []string{}
	for p := filepath.Dir(abs); p != filepath.Dir(p); p = filepath.Dir(p) {
		ancestors = append(ancestors, p)
	}
	for i := len(ancestors) - 1; i >= 0; i-- {
		p := ancestors[i]
		prefix, _ := filepath.Rel(p, abs)
		for _, name := range []string{".gitignore", ".ignore"} {
			rules := readIgnore(filepath.Join(p, name), "")
			for j := range rules {
				rules[j].prefix = filepath.ToSlash(prefix)
			}
			baseRules = append(baseRules, rules...)
		}
	}
	var walk func(string, int, []ignoreRule) error
	walk = func(dir string, depth int, rules []ignoreRule) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, _ := filepath.Rel(abs, dir)
		if rel == "." {
			rel = ""
		}
		rel = filepath.ToSlash(rel)
		local := append([]ignoreRule(nil), rules...)
		for _, name := range []string{".gitignore", ".ignore"} {
			local = append(local, readIgnore(filepath.Join(dir, name), rel)...)
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if depth == 0 {
				return err
			}
			out.Warnings = append(out.Warnings, err.Error())
			return nil
		}
		for _, ent := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			if out.Truncated {
				return nil
			}
			name := ent.Name()
			if !utf8.ValidString(name) || name == ".git" || (!showHidden && strings.HasPrefix(name, ".")) {
				continue
			}
			if ent.Type()&fs.ModeSymlink != 0 {
				continue
			}
			child := name
			if rel != "" {
				child = rel + "/" + name
			}
			isDir := ent.IsDir()
			excluded, matched, priority := ignoreDecision(child, isDir, local)
			if excluded {
				continue
			}
			if matched && priority == 1 {
				ignoreOverrides[child] = true
			}
			if isDir {
				out.Dirs = append(out.Dirs, child)
			} else {
				info, err := ent.Info()
				if err != nil {
					out.Warnings = append(out.Warnings, err.Error())
					continue
				}
				if !info.Mode().IsRegular() {
					continue
				}
				out.Files = append(out.Files, child)
			}
			if len(out.Dirs)+len(out.Files) >= MaxTreeEntries {
				out.Truncated = true
				return nil
			}
			if isDir && depth+1 < MaxTreeDepth {
				if err := walk(filepath.Join(dir, name), depth+1, local); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err = walk(abs, 0, baseRules); err != nil {
		return out, err
	}
	// check-ignore handles global excludes and exact Git semantics beyond the
	// local parser, in one NUL-framed subprocess (filenames may contain LF).
	all := append(append([]string(nil), out.Dirs...), out.Files...)
	if len(all) > 0 {
		cmd := exec.CommandContext(ctx, "git", "-C", abs, "check-ignore", "--no-index", "--stdin", "-z")
		cmd.Stdin = strings.NewReader(strings.Join(all, "\x00") + "\x00")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		raw, e := cmd.Output()
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if e == nil || exitStatus(e) == 1 {
			excluded := map[string]bool{}
			for _, p := range strings.Split(string(raw), "\x00") {
				excluded[p] = true
			}
			filter := func(xs []string) []string {
				ys := xs[:0]
				for _, x := range xs {
					if !excluded[x] || ignoreOverrides[x] {
						ys = append(ys, x)
					}
				}
				return ys
			}
			out.Dirs = filter(out.Dirs)
			out.Files = filter(out.Files)
		}
	}
	sort.Strings(out.Dirs)
	sort.Strings(out.Files)
	return out, nil
}
func exitStatus(err error) int {
	if err == nil {
		return 0
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		return e.ExitCode()
	}
	return -1
}

type TreeRow struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Directory bool   `json:"directory"`
	Depth     int    `json:"depth"`
	Expanded  bool   `json:"expanded"`
	GitStatus string `json:"gitStatus,omitempty"`
}

// TreeRows keeps ancestors of a filtered match, directories first, and only
// traverses expanded folders when no filter is active.
func TreeRows(s TreeSnapshot, expanded map[string]bool, query string, badges map[string]string) []TreeRow {
	query = strings.ToLower(strings.TrimSpace(query))
	dirs := map[string]bool{}
	for _, d := range s.Dirs {
		dirs[d] = true
	}
	visible := map[string]bool{}
	all := append(append([]string(nil), s.Dirs...), s.Files...)
	for _, p := range all {
		if query == "" || strings.Contains(strings.ToLower(p), query) {
			visible[p] = true
			for parent := filepath.ToSlash(filepath.Dir(p)); parent != "." && parent != ""; parent = filepath.ToSlash(filepath.Dir(parent)) {
				visible[parent] = true
			}
		}
	}
	children := map[string][]string{}
	for p := range visible {
		parent := filepath.ToSlash(filepath.Dir(p))
		if parent == "." {
			parent = ""
		}
		children[parent] = append(children[parent], p)
	}
	var out []TreeRow
	var walk func(string, int)
	walk = func(parent string, depth int) {
		xs := children[parent]
		sort.Slice(xs, func(i, j int) bool {
			if dirs[xs[i]] != dirs[xs[j]] {
				return dirs[xs[i]]
			}
			return strings.ToLower(xs[i]) < strings.ToLower(xs[j])
		})
		for _, p := range xs {
			open := expanded[p] || query != ""
			out = append(out, TreeRow{p, filepath.Base(p), dirs[p], depth, open, badges[p]})
			if dirs[p] && open {
				walk(p, depth+1)
			}
		}
	}
	walk("", 0)
	return out
}
