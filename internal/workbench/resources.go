package workbench

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charlzyx/orbit/internal/runtimepath"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Skill struct {
	Name, Description, Path, Scope, Content string
	ManualOnly, Enabled                     bool
	Size                                    int64
	Modified                                time.Time
}

type skillRoot struct{ path, scope string }

func (s *Service) skillRoots() []skillRoot {
	var roots []skillRoot
	homeAgents := filepath.Join(s.Preferences.Home, ".agents", "skills")
	if s.Workspace != "" {
		roots = append(roots, skillRoot{filepath.Join(s.Workspace, ".pi", "skills"), "project"})
		start := s.Workspace
		if canonical, e := filepath.EvalSymlinks(start); e == nil {
			start = canonical
		}
		for d := start; ; d = filepath.Dir(d) {
			path := filepath.Join(d, ".agents", "skills")
			if path != homeAgents {
				roots = append(roots, skillRoot{path, "project"})
			}
			if _, e := os.Stat(filepath.Join(d, ".git")); e == nil {
				break
			}
			if filepath.Dir(d) == d {
				break
			}
		}
	}
	return append(roots, skillRoot{filepath.Join(s.Preferences.PiDir, "skills"), "global"}, skillRoot{homeAgents, "user"})
}

// Like pi, follow skill directory links, stop at a SKILL.md root, and bound
// traversal depth so symlink cycles cannot turn discovery into an infinite walk.
func collectSkillFiles(dir string, depth int, out *[]string, errs *[]error) {
	if depth > 6 {
		return
	}
	path := filepath.Join(dir, "SKILL.md")
	if st, e := os.Stat(path); e == nil && st.Mode().IsRegular() {
		*out = append(*out, path)
		return
	}
	entries, e := os.ReadDir(dir)
	if e != nil {
		if !errors.Is(e, fs.ErrNotExist) {
			*errs = append(*errs, e)
		}
		return
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if st, e := os.Stat(path); e == nil && st.IsDir() {
			collectSkillFiles(path, depth+1, out, errs)
		}
	}
}
func normalizedSkill(content string) string {
	return strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\r", "\n")
}
func (s Skill) Valid() bool           { return strings.TrimSpace(s.Description) != "" }
func (s Skill) InvokeCommand() string { return "/skill:" + s.Name }
func (s Skill) Body() string {
	normalized := normalizedSkill(s.Content)
	if strings.HasPrefix(normalized, "---") {
		if end := strings.Index(normalized[3:], "\n---"); end >= 0 {
			return strings.TrimLeft(normalized[3+end+4:], "\n")
		}
	}
	return normalized
}
func skillEntry(v any) string {
	if m, ok := v.(map[string]any); ok {
		return stringValue(m["path"])
	}
	return stringValue(v)
}
func skillExactPath(file, pattern, base string) bool {
	pattern = strings.TrimPrefix(pattern, "./")
	// Rust Path equality ignores '.'/duplicate separators but deliberately does
	// not resolve '..' or symlinks. Keep that exact override contract.
	if !filepath.IsAbs(pattern) {
		pattern = base + string(filepath.Separator) + pattern
	}
	return skillPathKey(pattern) == skillPathKey(file) || skillPathKey(pattern) == skillPathKey(filepath.Dir(file))
}
func skillPathKey(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	out := make([]string, 0, len(parts))
	for n, p := range parts {
		if p == "." || (p == "" && n > 0) {
			continue
		}
		out = append(out, p)
	}
	return strings.Join(out, "/")
}

// This deliberately mirrors Orbit's permissive flat-scalar parser rather than
// treating unrelated or future skill metadata as a schema error.
func skillFrontmatter(content string) (name, description string, manual bool) {
	normalized := normalizedSkill(content)
	if !strings.HasPrefix(normalized, "---") {
		return
	}
	end := strings.Index(normalized[3:], "\n---")
	if end < 0 || 3+end < 4 {
		return
	}
	lines := strings.Split(normalized[4:3+end], "\n")
	unquote := func(v string) string {
		v = strings.TrimSpace(v)
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			return v[1 : len(v)-1]
		}
		return v
	}
	apply := func(key, value string) {
		switch key {
		case "name":
			name = unquote(value)
		case "description":
			description = unquote(value)
		case "disable-model-invocation":
			manual = strings.EqualFold(value, "true")
		}
	}
	collect := func(start int) ([]string, int) {
		var parts []string
		i := start
		for ; i < len(lines); i++ {
			line := lines[i]
			if strings.TrimSpace(line) == "" {
				parts = append(parts, "")
			} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				parts = append(parts, strings.TrimSpace(line))
			} else {
				break
			}
		}
		return parts, i
	}
	for i := 0; i < len(lines); {
		raw := lines[i]
		if strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t") || strings.TrimSpace(raw) == "" || strings.HasPrefix(strings.TrimSpace(raw), "#") {
			i++
			continue
		}
		key, value, ok := strings.Cut(raw, ":")
		if !ok {
			i++
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		block := value == "|" || value == ">" || strings.HasPrefix(value, "|-") || strings.HasPrefix(value, ">-") || strings.HasPrefix(value, "|+") || strings.HasPrefix(value, ">+")
		if block {
			parts, next := collect(i + 1)
			sep := "\n"
			if strings.HasPrefix(value, ">") {
				sep = " "
			}
			apply(key, strings.TrimSpace(strings.Join(parts, sep)))
			i = next
			continue
		}
		if value == "" && (key == "name" || key == "description" || key == "disable-model-invocation") {
			parts, next := collect(i + 1)
			joined := strings.TrimSpace(strings.Join(parts, " "))
			if joined != "" {
				apply(key, joined)
				i = next
				continue
			}
		}
		apply(key, value)
		i++
	}
	return
}

func (s *Service) Skills() ([]Skill, []error) {
	var out []Skill
	var errs []error
	seenName, seenFile := map[string]bool{}, map[string]bool{}
	settings := map[string][]any{}
	for _, scope := range []string{"project", "global"} {
		if path, e := s.settingsPath(scope); e == nil {
			m, e := readObject(path)
			if e != nil {
				errs = append(errs, e)
			}
			settings[scope], _ = m["skills"].([]any)
		}
	}
	for _, r := range s.skillRoots() {
		base, e := filepath.Abs(r.path)
		if e != nil {
			errs = append(errs, e)
			continue
		}
		var files []string
		collectSkillFiles(base, 0, &files, &errs)
		sort.Strings(files)
		for _, path := range files {
			canonical, e := filepath.EvalSymlinks(path)
			if e != nil {
				canonical = path
			}
			if seenFile[canonical] {
				continue
			}
			b, e := os.ReadFile(path)
			if e != nil {
				errs = append(errs, e)
				continue
			}
			st, e := os.Stat(path)
			if e != nil {
				errs = append(errs, e)
				continue
			}
			skill := Skill{Name: filepath.Base(filepath.Dir(path)), Path: path, Scope: r.scope, Content: string(b), Enabled: true, Size: st.Size(), Modified: st.ModTime()}
			name, description, manual := skillFrontmatter(skill.Content)
			if strings.TrimSpace(name) != "" {
				skill.Name = name
			}
			skill.Description = description
			skill.ManualOnly = manual

			seenFile[canonical] = true
			if seenName[skill.Name] {
				continue
			}
			seenName[skill.Name] = true
			settingsPath, _ := s.settingsPath(skill.Scope)
			scope := skill.Scope
			if scope == "user" {
				scope = "global"
			}
			for _, v := range settings[scope] {
				pattern := skillEntry(v)
				if strings.HasPrefix(pattern, "-") && skillExactPath(path, pattern[1:], filepath.Dir(settingsPath)) {
					skill.Enabled = false
				}
			}
			out = append(out, skill)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, errs
}
func (s *Service) SetSkillEnabled(skill Skill, enabled bool) error {
	path, e := s.settingsPath(skill.Scope)
	if e != nil {
		return e
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if enabled {
		if _, e := os.Stat(path); errors.Is(e, fs.ErrNotExist) {
			return nil
		}
	}
	m, e := readObject(path)
	if e != nil {
		return e
	}
	entries, ok := m["skills"].([]any)
	if _, present := m["skills"]; present && !ok {
		return errors.New("settings skills must be an array")
	}
	keep := make([]any, 0, len(entries)+1)
	removed := false
	for _, v := range entries {
		p := stringValue(v)
		if (strings.HasPrefix(p, "-") || strings.HasPrefix(p, "+")) && skillExactPath(skill.Path, p[1:], filepath.Dir(path)) {
			removed = true
			continue
		}
		keep = append(keep, v)
	}
	if enabled && !removed {
		return nil
	}
	if !enabled {
		keep = append(keep, "-"+skill.Path)
	}
	m["skills"] = keep
	return writeObject(path, m)
}
func (s *Service) DeleteSkill(path string) error {
	skills, _ := s.Skills()
	for _, skill := range skills {
		if filepath.Clean(path) == filepath.Clean(skill.Path) {
			dir := filepath.Dir(skill.Path)
			var e error
			// A skill declared directly at a discovery root must not delete its siblings.
			root := filepath.Base(dir) == "skills"
			for _, r := range s.skillRoots() {
				if filepath.Clean(r.path) == filepath.Clean(dir) {
					root = true
				}
			}
			if root {
				e = os.Remove(skill.Path)
			} else {
				e = os.RemoveAll(dir)
			}
			if e != nil {
				return e
			}
			if e = s.SetSkillEnabled(skill, true); e != nil {
				return fmt.Errorf("skill deleted, but its settings override could not be cleared: %w", e)
			}
			return nil
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
			if _, e := os.Stat(plugin.InstallPath); e == nil {
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
	source = strings.TrimSpace(source)
	if spec, ok := strings.CutPrefix(source, "npm:"); ok {
		spec = strings.TrimSpace(spec)
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
	if scope != "global" && scope != "project" && scope != "user" {
		return "", errors.New("invalid package scope")
	}
	if scope == "project" && s.Workspace == "" {
		return "", errors.New("select a workspace first")
	}
	args := []string{action, source}
	if scope == "project" && action != "update" {
		args = append(args, "-l", "-a")
	}
	return s.Pi(ctx, args...)
}

// PluginUpdate is an observed upstream version, never an installation promise.
type PluginUpdate struct{ Source, Name, Scope, Current, Latest string }

func (s *Service) SkippedPluginUpdates() (map[string]string, error) {
	m, e := readObject(filepath.Join(s.Preferences.RootDir, "plugin-update-skips.json"))
	if e != nil {
		return nil, e
	}
	out := map[string]string{}
	for k, v := range m {
		if text, ok := v.(string); ok {
			out[k] = text
		}
	}
	return out, nil
}
func (s *Service) SkipPluginUpdate(source, version string) error {
	if strings.TrimSpace(source) == "" {
		return errors.New("package source is required")
	}
	return s.edit(filepath.Join(s.Preferences.RootDir, "plugin-update-skips.json"), func(m map[string]any) error {
		if version == "" {
			delete(m, source)
		} else {
			m[source] = version
		}
		return nil
	})
}

// FindPluginUpdates performs read-only npm/Git metadata probes off the UI thread.
// Failures are returned separately so one inaccessible registry does not hide
// updates discovered from other sources. No install directory is modified.
func (s *Service) FindPluginUpdates(ctx context.Context) ([]PluginUpdate, []error) {
	packages, errs := s.Plugins()
	skipped, e := s.SkippedPluginUpdates()
	if e != nil {
		errs = append(errs, e)
	}
	found, probeErrors := s.findPluginUpdates(ctx, packages, skipped, s.resourceProbe)
	return found, append(errs, probeErrors...)
}

type resourceProbeFunc func(context.Context, string, string, ...string) (string, error)

func (s *Service) findPluginUpdates(ctx context.Context, packages []Plugin, skipped map[string]string, probe resourceProbeFunc) ([]PluginUpdate, []error) {
	type result struct {
		update *PluginUpdate
		err    error
	}
	results := make([]result, len(packages))
	limit := make(chan struct{}, 4)
	var wg sync.WaitGroup
	for n, p := range packages {
		if !p.Installed || p.Kind == "local" {
			continue
		}
		wg.Add(1)
		go func(n int, p Plugin) {
			defer wg.Done()
			select {
			case limit <- struct{}{}:
			case <-ctx.Done():
				results[n].err = ctx.Err()
				return
			}
			defer func() { <-limit }()
			update, e := pluginUpdate(ctx, p, s.Workspace, skipped, probe)
			results[n] = result{update, e}
		}(n, p)
	}
	wg.Wait()
	var out []PluginUpdate
	var errs []error
	for n, r := range results {
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", packages[n].Name, r.err))
		}
		if r.update != nil {
			out = append(out, *r.update)
		}
	}
	return out, errs
}
func pluginUpdate(ctx context.Context, p Plugin, workspace string, skipped map[string]string, probe resourceProbeFunc) (*PluginUpdate, error) {
	current, latest := p.Version, ""
	switch p.Kind {
	case "npm":
		name := npmUpdateName(p.Source)
		if name == "" || current == "" {
			return nil, nil
		}
		binary, args := npmViewCommand(name)
		out, e := probe(ctx, workspace, binary, args...)
		if e != nil {
			return nil, e
		}
		latest = npmVersion(out)
		if latest == "" {
			return nil, errors.New("registry returned no valid version")
		}
		if !newerPluginVersion(latest, current) {
			return nil, nil
		}
	case "git":
		if strings.Contains(p.Source, "#") {
			return nil, nil
		}
		out, e := probe(ctx, p.InstallPath, "git", "rev-parse", "HEAD")
		if e != nil {
			return nil, e
		}
		current = strings.TrimSpace(out)
		if !gitRevision(current) {
			return nil, errors.New("installed Git HEAD is not a revision")
		}
		remote, e := probe(ctx, p.InstallPath, "git", "remote", "get-url", "origin")
		if e != nil {
			return nil, e
		}
		remote = strings.TrimSpace(remote)
		if remote == "" {
			return nil, errors.New("Git origin is empty")
		}
		out, e = probe(ctx, p.InstallPath, "git", "ls-remote", "--symref", "--", remote, "HEAD")
		if e != nil {
			return nil, e
		}
		for _, line := range strings.Split(out, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[1] == "HEAD" && gitRevision(fields[0]) {
				latest = fields[0]
				break
			}
		}
		if latest == "" {
			return nil, errors.New("Git remote did not advertise HEAD")
		}
	default:
		return nil, nil
	}
	if current == latest || skipped[p.Source] == latest {
		return nil, nil
	}
	return &PluginUpdate{Source: p.Source, Name: p.Name, Scope: p.Scope, Current: current, Latest: latest}, nil
}
func npmUpdateName(source string) string {
	spec, ok := strings.CutPrefix(source, "npm:")
	if !ok {
		return ""
	}
	spec = strings.TrimSpace(spec)
	name := resolvePlugin("npm:"+spec, "global", "").Name
	if spec != name || name == "" || strings.HasPrefix(name, "-") {
		return ""
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("@/._-", c)) {
			return ""
		}
	}
	return name
}
func npmViewCommand(name string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd.exe", []string{"/D", "/S", "/C", "npm view --json " + name + "@latest version"}
	}
	return "npm", []string{"view", "--json", name + "@latest", "version"}
}
func npmVersion(out string) string {
	var v any
	value := ""
	if json.Unmarshal([]byte(out), &v) == nil {
		switch x := v.(type) {
		case string:
			value = x
		case []any:
			for _, item := range x {
				if text, ok := item.(string); ok {
					value = text
					break
				}
			}
		case map[string]any:
			value = stringValue(x["version"])
		}
	}
	if value == "" {
		value = strings.Trim(strings.TrimSpace(out), "\"")
	}
	if value != "" && value[0] >= '0' && value[0] <= '9' {
		return value
	}
	return ""
}
func gitRevision(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
func newerPluginVersion(a, b string) bool {
	fields := func(v string) []string {
		v, _, _ = strings.Cut(v, "-")
		v, _, _ = strings.Cut(v, "+")
		return strings.Split(v, ".")
	}
	aa, bb := fields(a), fields(b)
	for i := 0; i < max(len(aa), len(bb)); i++ {
		var x, y uint64
		if i < len(aa) {
			x, _ = strconv.ParseUint(strings.TrimSpace(aa[i]), 10, 64)
		}
		if i < len(bb) {
			y, _ = strconv.ParseUint(strings.TrimSpace(bb[i]), 10, 64)
		}
		if x != y {
			return x > y
		}
	}
	return false
}
func (s *Service) resourceProbe(ctx context.Context, dir, binary string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()
	env := append(os.Environ(), "HOME="+s.Preferences.Home, "PI_CODING_AGENT_DIR="+s.Preferences.PiDir, "PI_SKIP_VERSION_CHECK=1", "GIT_TERMINAL_PROMPT=0", "NO_COLOR=1", "CI=1")
	preferred := ""
	if s.PiBinary != "" && strings.ContainsAny(s.PiBinary, "/\\") {
		preferred = filepath.Dir(s.PiBinary)
	}
	env = runtimepath.AugmentEnvironment(env, preferred)
	if resolved := runtimepath.ResolveIn(binary, env, runtime.GOOS); resolved != "" {
		binary = resolved
	}
	cmd := exec.CommandContext(ctx, binary, args...)
	configureResourceProcess(cmd)
	cmd.Dir = dir
	cmd.Env = env
	cmd.WaitDelay = time.Second
	var stdout, stderr limitedBuffer
	stdout.remaining = 4 << 20
	stderr.remaining = 1 << 20
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if e := cmd.Run(); e != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		return "", fmt.Errorf("%s: %w: %s", filepath.Base(binary), e, s.Redact(strings.TrimSpace(stderr.String())))
	}
	if stdout.truncated {
		return "", errors.New("metadata output exceeded 4 MiB limit")
	}
	return stdout.String(), nil
}
