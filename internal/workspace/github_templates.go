package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const MaxTemplateBytes = 256 * 1024

type RepositoryTemplate struct {
	Name, Path, Body, TitlePrefix string
	Labels                        []string
}

func (t RepositoryTemplate) Label() string {
	if strings.TrimSpace(t.Name) != "" {
		return t.Name
	}
	return templateName(filepath.Base(t.Path))
}
func IssueTemplates(root string) []RepositoryTemplate {
	return repositoryTemplates(root, "ISSUE_TEMPLATE", true, []string{".github/ISSUE_TEMPLATE.md", ".github/issue_template.md", ".github/ISSUE_TEMPLATE", "docs/ISSUE_TEMPLATE.md", "ISSUE_TEMPLATE.md"})
}
func PullTemplates(root string) []RepositoryTemplate {
	return repositoryTemplates(root, "PULL_REQUEST_TEMPLATE", false, []string{".github/PULL_REQUEST_TEMPLATE.md", ".github/pull_request_template.md", "docs/PULL_REQUEST_TEMPLATE.md", "PULL_REQUEST_TEMPLATE.md", "pull_request_template.md"})
}
func repositoryTemplates(root, dir string, issue bool, fallback []string) []RepositoryTemplate {
	var out []RepositoryTemplate
	entries, _ := os.ReadDir(filepath.Join(root, ".github", dir))
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if issue && (lower == "config.yml" || lower == "config.yaml") {
			continue
		}
		markdown := strings.HasSuffix(lower, ".md")
		form := issue && (strings.HasSuffix(lower, ".yml") || strings.HasSuffix(lower, ".yaml"))
		if !markdown && !form {
			continue
		}
		relative := filepath.ToSlash(filepath.Join(".github", dir, name))
		content, ok := readTemplate(root, relative)
		if !ok {
			continue
		}
		if markdown {
			out = append(out, ParseMarkdownTemplate(content, relative))
		} else {
			out = append(out, ParseIssueFormTemplate(content, relative))
		}
	}
	if len(out) > 0 {
		return out
	}
	for _, relative := range fallback {
		if content, ok := readTemplate(root, relative); ok {
			return []RepositoryTemplate{ParseMarkdownTemplate(content, relative)}
		}
	}
	return nil
}
func readTemplate(root, relative string) (string, bool) {
	path := filepath.Join(root, relative)
	info, e := os.Stat(path)
	if e != nil || !info.Mode().IsRegular() || info.Size() > MaxTemplateBytes {
		return "", false
	}
	resolved, e := filepath.EvalSymlinks(path)
	if e != nil {
		return "", false
	}
	base, e := filepath.EvalSymlinks(root)
	if e != nil {
		return "", false
	}
	rel, e := filepath.Rel(base, resolved)
	if e != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	raw, e := os.ReadFile(path)
	if e != nil || !utf8.Valid(raw) || strings.TrimSpace(string(raw)) == "" {
		return "", false
	}
	return string(raw), true
}
func templateScalar(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		return s[1 : len(s)-1]
	}
	return s
}
func templateLabels(s string) []string {
	var result []string
	for _, part := range strings.Split(templateScalar(s), ",") {
		part = strings.TrimSpace(strings.Trim(strings.TrimSpace(part), "[]"))
		part = strings.TrimSpace(strings.Trim(part, "\"'"))
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
func templateName(name string) string {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	if strings.ToUpper(stem) == stem {
		stem = strings.ToLower(stem)
	}
	stem = strings.TrimSpace(strings.NewReplacer("_", " ", "-", " ").Replace(stem))
	if stem == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(stem)
	return string(unicode.ToUpper(r)) + stem[n:]
}
func ParseMarkdownTemplate(content, path string) RepositoryTemplate {
	content = strings.TrimPrefix(content, "\ufeff")
	t := RepositoryTemplate{Name: templateName(filepath.Base(path)), Path: path, Body: strings.TrimSpace(content)}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return t
	}
	end := 0
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == 0 {
		return t
	}
	t.Body = strings.TrimSpace(strings.Join(lines[end+1:], "\n"))
	for _, line := range lines[1:end] {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "name":
			if name := strings.TrimSpace(templateScalar(value)); name != "" {
				t.Name = name
			}
		case "title":
			t.TitlePrefix = templateScalar(value)
		case "labels":
			t.Labels = templateLabels(value)
		}
	}
	return t
}

type templateField struct {
	kind, label, description, placeholder, value, render string
	options                                              []string
}

func ParseIssueFormTemplate(content, path string) RepositoryTemplate {
	t := RepositoryTemplate{Name: templateName(filepath.Base(path)), Path: path}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	inBody := false
	var fields []templateField
	var field *templateField
	for i := 0; i < len(lines); {
		raw := lines[i]
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			i++
			continue
		}
		if !inBody {
			if strings.HasPrefix(trimmed, "body:") {
				inBody = true
				i++
				continue
			}
			if raw == strings.TrimLeft(raw, " \t") {
				key, value, ok := strings.Cut(raw, ":")
				if ok {
					switch strings.ToLower(strings.TrimSpace(key)) {
					case "name":
						if name := strings.TrimSpace(templateScalar(value)); name != "" {
							t.Name = name
						}
					case "title":
						t.TitlePrefix = templateScalar(value)
					case "labels":
						t.Labels = templateLabels(value)
					}
				}
			}
			i++
			continue
		}
		if value, ok := strings.CutPrefix(trimmed, "- type:"); ok {
			if field != nil {
				fields = append(fields, *field)
			}
			field = &templateField{kind: strings.ToLower(strings.TrimSpace(value))}
			i++
			continue
		}
		if field == nil {
			i++
			continue
		}
		if value, ok := strings.CutPrefix(trimmed, "- label:"); ok {
			field.options = append(field.options, templateScalar(value))
			i++
			continue
		}
		if value, ok := strings.CutPrefix(trimmed, "- "); ok {
			field.options = append(field.options, templateScalar(value))
			i++
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			i++
			continue
		}
		if (key == "placeholder" || key == "value") && (strings.HasPrefix(strings.TrimSpace(value), "|") || strings.HasPrefix(strings.TrimSpace(value), ">")) {
			indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
			value, i = templateBlock(lines, i+1, indent)
		} else {
			value = templateScalar(value)
			i++
		}
		switch key {
		case "label":
			field.label = value
		case "description":
			field.description = value
		case "placeholder":
			field.placeholder = value
		case "value":
			field.value = value
		case "render":
			field.render = value
		}
	}
	if field != nil {
		fields = append(fields, *field)
	}
	var out strings.Builder
	for _, f := range fields {
		switch f.kind {
		case "markdown":
			if strings.TrimSpace(f.value) != "" {
				out.WriteString("<!--\n" + strings.TrimSpace(f.value) + "\n-->\n\n")
			}
		case "checkboxes":
			if f.label != "" {
				out.WriteString("### " + strings.TrimSpace(f.label) + "\n\n")
			}
			for _, option := range f.options {
				out.WriteString("- [ ] " + option + "\n")
			}
			out.WriteString("\n")
		default:
			if f.label != "" {
				out.WriteString("### " + strings.TrimSpace(f.label) + "\n\n")
			}
			if f.description != "" {
				out.WriteString("<!-- " + strings.TrimSpace(f.description) + " -->\n\n")
			}
			if f.placeholder != "" {
				out.WriteString(strings.TrimSpace(f.placeholder) + "\n\n")
			}
			if f.kind == "dropdown" && len(f.options) > 0 {
				out.WriteString("<!-- one of: " + strings.Join(f.options, ", ") + " -->\n\n")
			}
			if f.render == "shell" {
				out.WriteString("```shell\n\n```\n\n")
			}
		}
	}
	t.Body = strings.TrimRight(out.String(), " \t\n")
	return t
}
func templateBlock(lines []string, start, keyIndent int) (string, int) {
	end := start
	indent := -1
	for end < len(lines) {
		raw := lines[end]
		if strings.TrimSpace(raw) == "" {
			end++
			continue
		}
		n := len(raw) - len(strings.TrimLeft(raw, " \t"))
		if n <= keyIndent {
			break
		}
		if indent < 0 || n < indent {
			indent = n
		}
		end++
	}
	if indent < 0 {
		indent = 0
	}
	var block []string
	for _, raw := range lines[start:end] {
		if len(raw) >= indent {
			block = append(block, raw[indent:])
		} else {
			block = append(block, strings.TrimLeft(raw, " \t"))
		}
	}
	return strings.TrimRight(strings.Join(block, "\n"), " \t\n"), end
}
