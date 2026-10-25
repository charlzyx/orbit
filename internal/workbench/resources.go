package workbench

import (
	"context"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Skill struct {
	Name, Description, Path, Scope, Content string
	ManualOnly, Enabled                     bool
	Size                                    int64
	Modified                                time.Time
}

func (s *Service) Skills() ([]Skill, []error) {
	type root struct{ path, scope string }
	roots := []root{}
	if s.Workspace != "" {
		roots = append(roots, root{filepath.Join(s.Workspace, ".pi", "skills"), "project"})
		for d := s.Workspace; ; d = filepath.Dir(d) {
			roots = append(roots, root{filepath.Join(d, ".agents", "skills"), "project"})
			if _, e := os.Stat(filepath.Join(d, ".git")); e == nil {
				break
			}
			if filepath.Dir(d) == d {
				break
			}
		}
	}
	roots = append(roots, root{filepath.Join(s.Preferences.PiDir, "skills"), "global"}, root{filepath.Join(s.Preferences.Home, ".agents", "skills"), "user"})
	var out []Skill
	var errs []error
	seenName := map[string]bool{}
	seenFile := map[string]bool{}
	for _, r := range roots {
		base, e := filepath.Abs(r.path)
		if e != nil {
			continue
		}
		filepath.WalkDir(base, func(path string, d fs.DirEntry, walkerr error) error {
			if walkerr != nil {
				if !errors.Is(walkerr, fs.ErrNotExist) {
					errs = append(errs, walkerr)
				}
				return nil
			}
			rel, _ := filepath.Rel(base, path)
			if d.IsDir() {
				if strings.Count(rel, string(filepath.Separator)) > 6 {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() != "SKILL.md" {
				return nil
			}
			canonical, _ := filepath.EvalSymlinks(path)
			if seenFile[canonical] {
				return nil
			}
			seenFile[canonical] = true
			b, e := os.ReadFile(path)
			if e != nil {
				errs = append(errs, e)
				return nil
			}
			st, e := d.Info()
			if e != nil {
				return nil
			}
			skill := Skill{Name: filepath.Base(filepath.Dir(path)), Path: path, Scope: r.scope, Content: string(b), Enabled: true, Size: st.Size(), Modified: st.ModTime()}
			if strings.HasPrefix(skill.Content, "---\n") {
				if _, tail, ok := strings.Cut(skill.Content[4:], "\n---"); ok {
					front := skill.Content[4 : len(skill.Content)-len(tail)-4]
					var m struct {
						Name        string `yaml:"name"`
						Description string `yaml:"description"`
						Manual      bool   `yaml:"disable-model-invocation"`
					}
					if e = yaml.Unmarshal([]byte(front), &m); e != nil {
						errs = append(errs, fmt.Errorf("%s: invalid skill frontmatter: %w", path, e))
					} else {
						if m.Name != "" {
							skill.Name = m.Name
						}
						skill.Description = m.Description
						skill.ManualOnly = m.Manual
					}
				}
			}
			if seenName[skill.Name] {
				return nil
			}
			seenName[skill.Name] = true
			settings, _ := s.settingsPath(skill.Scope)
			config, e := readObject(settings)
			if e != nil {
				errs = append(errs, e)
			}
			entries, _ := config["skills"].([]any)
			for _, v := range entries {
				pattern := stringValue(v)
				if m, ok := v.(map[string]any); ok {
					pattern = stringValue(m["path"])
				}
				if strings.HasPrefix(pattern, "-") {
					p := strings.TrimPrefix(pattern, "-")
					if !filepath.IsAbs(p) {
						p = filepath.Join(filepath.Dir(settings), p)
					}
					p = filepath.Clean(p)
					if p == path || p == filepath.Dir(path) {
						skill.Enabled = false
					}
				}
			}
			out = append(out, skill)
			return nil
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, errs
}
func (s *Service) SetSkillEnabled(skill Skill, enabled bool) error {
	path, e := s.settingsPath(skill.Scope)
	if e != nil {
		return e
	}
	return s.edit(path, func(m map[string]any) error {
		entries, _ := m["skills"].([]any)
		var keep []any
		for _, v := range entries {
			p := stringValue(v)
			if strings.HasPrefix(p, "-") || strings.HasPrefix(p, "+") {
				p = p[1:]
				if !filepath.IsAbs(p) {
					p = filepath.Join(filepath.Dir(path), p)
				}
				p = filepath.Clean(p)
				if p == skill.Path || p == filepath.Dir(skill.Path) {
					continue
				}
			}
			keep = append(keep, v)
		}
		if !enabled {
			keep = append(keep, "-"+skill.Path)
		}
		m["skills"] = keep
		return nil
	})
}
func (s *Service) DeleteSkill(path string) error {
	skills, _ := s.Skills()
	for _, skill := range skills {
		if filepath.Clean(path) == filepath.Clean(skill.Path) {
			return os.RemoveAll(filepath.Dir(skill.Path))
		}
	}
	return errors.New("skill is no longer in a discovered skill root")
}

type Plugin struct {
	Source, Name, Scope, Kind, InstallPath, Version string
	Installed                                       bool
}

func (s *Service) Plugins() ([]Plugin, []error) {
	var out []Plugin
	var errs []error
	seen := map[string]bool{}
	for _, scope := range []string{"project", "global"} {
		p, e := s.settingsPath(scope)
		if e != nil {
			continue
		}
		m, e := readObject(p)
		if e != nil {
			errs = append(errs, e)
			continue
		}
		items, _ := m["packages"].([]any)
		for _, v := range items {
			source := stringValue(v)
			if m, ok := v.(map[string]any); ok {
				source = stringValue(m["source"])
			}
			if strings.TrimSpace(source) == "" || seen[source] {
				continue
			}
			seen[source] = true
			plugin := resolvePlugin(source, scope, filepath.Dir(p))
			if st, e := os.Stat(plugin.InstallPath); e == nil && st.IsDir() {
				plugin.Installed = true
				if m, e := readObject(filepath.Join(plugin.InstallPath, "package.json")); e == nil {
					plugin.Version = stringValue(m["version"])
				}
			}
			out = append(out, plugin)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, errs
}
func resolvePlugin(source, scope, base string) Plugin {
	p := Plugin{Source: source, Scope: scope}
	if spec, ok := strings.CutPrefix(source, "npm:"); ok {
		p.Kind = "npm"
		name := spec
		start := 0
		if strings.HasPrefix(spec, "@") {
			start = strings.Index(spec, "/") + 1
		}
		if i := strings.Index(spec[start:], "@"); i >= 0 {
			name = spec[:start+i]
		}
		p.Name = name
		p.InstallPath = filepath.Join(base, "npm", "node_modules", name)
		return p
	}
	rest := source
	git := false
	for _, prefix := range []string{"git:", "github:", "https://", "http://", "ssh://", "git://"} {
		if strings.HasPrefix(source, prefix) {
			rest = strings.TrimPrefix(source, prefix)
			git = true
			if prefix == "github:" {
				rest = "github.com/" + rest
			}
			break
		}
	}
	if git {
		p.Kind = "git"
		rest, _, _ = strings.Cut(rest, "#")
		rest = strings.TrimPrefix(rest, "git@")
		rest = strings.Replace(rest, ":", "/", 1)
		rest = strings.TrimSuffix(rest, ".git")
		p.Name = filepath.Base(rest)
		p.InstallPath = filepath.Join(base, "git", rest)
		return p
	}
	p.Kind = "local"
	p.Name = filepath.Base(source)
	p.InstallPath = source
	if !filepath.IsAbs(source) {
		p.InstallPath = filepath.Join(base, source)
	}
	return p
}
func (s *Service) PluginAction(ctx context.Context, action, source, scope string) (string, error) {
	if action != "install" && action != "update" && action != "remove" {
		return "", errors.New("invalid plugin action")
	}
	source = strings.TrimSpace(source)
	if source == "" || strings.HasPrefix(source, "-") || strings.ContainsRune(source, 0) {
		return "", errors.New("invalid package source")
	}
	args := []string{action, source}
	if scope == "project" && action != "update" {
		args = append(args, "-l", "-a")
	}
	return s.Pi(ctx, args...)
}
