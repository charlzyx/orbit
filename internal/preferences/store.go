// Package preferences preserves Orbit's on-disk preferences independently of Pi.
package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

// Store is a concurrency-safe settings repository. Home and PiDir are explicit so
// tests never borrow credentials or write to the user's actual configuration.
type Store struct {
	Home, RootDir, UserHome, PiDir string
	mu                             sync.Mutex
}

func New(home string) *Store {
	return NewWithPi(home, home, "")
}
func Default() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configured := os.Getenv("PI_CODING_AGENT_DIR")
	if err := ValidatePiDir(configured); err != nil {
		return nil, err
	}
	return NewWithPi(home, home, configured), nil
}
func (s *Store) path(name string) (string, error) {
	if name == "" || filepath.Base(name) != name || name == "." || name == ".." {
		return "", fmt.Errorf("invalid preference name %q", name)
	}
	return filepath.Join(s.RootDir, name), nil
}
func (s *Store) Read(name string, target any) error {
	unlockWorkspace := s.lockWorkspaceFile(name)
	defer unlockWorkspace()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.read(name, target)
}
func (s *Store) read(name string, target any) error {
	p, e := s.path(name)
	if e != nil {
		return e
	}
	b, e := os.ReadFile(p)
	if e != nil {
		return e
	}
	if e = json.Unmarshal(b, target); e != nil {
		return fmt.Errorf("read %s: %w", name, e)
	}
	return nil
}
func (s *Store) Write(name string, value any) error {
	unlockWorkspace := s.lockWorkspaceFile(name)
	defer unlockWorkspace()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.write(name, value)
}
func (s *Store) write(name string, value any) error {
	p, e := s.path(name)
	if e != nil {
		return e
	}
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	err := AtomicWrite(p, append(b, '\n'), 0600)
	if err == nil && name == "workspaces.json" {
		workspaceFiles.revision[s.workspaceFileKey()]++
	}
	return err
}

// Update reloads on every write, preserving keys written by Pi or another window.
// A malformed file is not silently replaced: the caller must explicitly repair it.
func (s *Store) Update(name string, edit func(map[string]json.RawMessage) error) error {
	unlockWorkspace := s.lockWorkspaceFile(name)
	defer unlockWorkspace()
	s.mu.Lock()
	defer s.mu.Unlock()
	m := map[string]json.RawMessage{}
	e := s.read(name, &m)
	if e != nil && !errors.Is(e, fs.ErrNotExist) {
		return e
	}
	if m == nil {
		return fmt.Errorf("%s must contain an object", name)
	}
	if e = edit(m); e != nil {
		return e
	}
	return s.write(name, m)
}
func AtomicWrite(path string, data []byte, mode fs.FileMode) error {
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".orbit-*.tmp")
	if e != nil {
		return e
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if e = f.Chmod(mode); e != nil {
		f.Close()
		return e
	}
	if _, e = f.Write(data); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}
func (s *Store) raw(name string) map[string]json.RawMessage {
	m := map[string]json.RawMessage{}
	if s.Read(name, &m) != nil || m == nil {
		return map[string]json.RawMessage{}
	}
	return m
}
func str(m map[string]json.RawMessage, k, def string) string {
	var v string
	if json.Unmarshal(m[k], &v) == nil {
		return v
	}
	return def
}
func num(m map[string]json.RawMessage, k string, def float64) float64 {
	var v float64
	if json.Unmarshal(m[k], &v) == nil && !math.IsNaN(v) && !math.IsInf(v, 0) {
		return v
	}
	return def
}
func boolean(m map[string]json.RawMessage, k string, def bool) bool {
	var v bool
	if json.Unmarshal(m[k], &v) == nil {
		return v
	}
	return def
}
func rawValue(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

var FontSizes = []float64{11, 12, 13, 14, 15, 16, 18, 20}
var SpacingDensities = []int{80, 85, 90, 95, 100, 105, 110, 115, 120}
var Languages = []string{"system", "en", "zh-CN", "ja", "ko", "es", "fr", "de", "pt-BR", "ru", "it"}

type UI struct {
	Language         string  `json:"language"`
	UIFontSize       float64 `json:"ui_font_size"`
	TerminalFontSize float64 `json:"terminal_font_size"`
	EditorFontSize   float64 `json:"editor_font_size"`
	SpacingDensity   int     `json:"spacing_density"`
	ReduceMotion     bool    `json:"reduce_motion"`
	ComposerSendMode string  `json:"composer_send_mode"`
}

func DefaultUI() UI { return UI{"system", 14, 13, 13, 100, false, "follow_up"} }
func (s *Store) UI() UI {
	m := s.raw("ui.json")
	v := DefaultUI()
	v.Language = str(m, "language", v.Language)
	if !slices.Contains(Languages, v.Language) {
		v.Language = "system"
	}
	if _, ok := m["ui_font_size"]; !ok {
		if _, ok = m["interface_scale"]; ok {
			x := num(m, "interface_scale", 100) * 14 / 100
			best := 14.
			for _, n := range FontSizes {
				if math.Abs(n-x) < math.Abs(best-x) {
					best = n
				}
			}
			v.UIFontSize = best
		}
	}
	v.UIFontSize = num(m, "ui_font_size", v.UIFontSize)
	v.TerminalFontSize = num(m, "terminal_font_size", 13)
	v.EditorFontSize = num(m, "editor_font_size", num(m, "code_font_size", 13))
	v.SpacingDensity = int(num(m, "spacing_density", 100))
	v.ReduceMotion = boolean(m, "reduce_motion", false)
	v.ComposerSendMode = str(m, "composer_send_mode", "follow_up")
	return normalizeUI(v)
}
func normalizeUI(v UI) UI {
	if !slices.Contains(Languages, v.Language) {
		v.Language = "system"
	}
	if !slices.Contains(FontSizes, v.UIFontSize) {
		v.UIFontSize = 14
	}
	if !slices.Contains(FontSizes, v.TerminalFontSize) {
		v.TerminalFontSize = 13
	}
	if !slices.Contains(FontSizes, v.EditorFontSize) {
		v.EditorFontSize = 13
	}
	if !slices.Contains(SpacingDensities, v.SpacingDensity) {
		v.SpacingDensity = 100
	}
	if v.ComposerSendMode != "steer" {
		v.ComposerSendMode = "follow_up"
	}
	return v
}
func (s *Store) SaveUI(v UI) error { return s.Write("ui.json", normalizeUI(v)) }

type Fonts struct {
	UIFamily   string `json:"ui_font_family"`
	CodeFamily string `json:"code_font_family"`
}

func (s *Store) Fonts() Fonts {
	m := s.raw("fonts.json")
	v := Fonts{str(m, "ui_font_family", ".ZedSans"), str(m, "code_font_family", ".ZedMono")}
	if v.UIFamily == "" {
		v.UIFamily = ".ZedSans"
	}
	if v.CodeFamily == "" {
		v.CodeFamily = ".ZedMono"
	}
	return v
}
func (s *Store) SaveFonts(v Fonts) error {
	if v.UIFamily == "" || v.CodeFamily == "" {
		return errors.New("font family cannot be empty")
	}
	return s.Write("fonts.json", v)
}

type Appearance struct {
	Mode       string `json:"mode"`
	LightTheme string `json:"light_theme"`
	DarkTheme  string `json:"dark_theme"`
}

var DarkThemes = []string{"orbit", "orbit-darker", "orbit-cobalt", "orbit-azure", "flexoki"}
var LightThemes = []string{"orbit-light", "orbit-azure-light", "orbit-paper", "orbit-contrast", "flexoki-light"}

func DefaultAppearance() Appearance { return Appearance{"system", "orbit-azure-light", "orbit-azure"} }
func (s *Store) Appearance() Appearance {
	s.mu.Lock()
	p, _ := s.path("theme")
	b, e := os.ReadFile(p)
	s.mu.Unlock()
	v := DefaultAppearance()
	if e != nil {
		return v
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(b, &m) != nil {
		theme := themeAlias(strings.TrimSpace(string(b)))
		if theme == "" {
			return v
		}
		if slices.Contains(LightThemes, theme) {
			v.Mode = "light"
			v.LightTheme = theme
		} else {
			v.Mode = "dark"
			v.DarkTheme = theme
		}
		return v
	}
	v.Mode = str(m, "mode", "system")
	l := themeAlias(str(m, "light_theme", v.LightTheme))
	d := themeAlias(str(m, "dark_theme", v.DarkTheme))
	if slices.Contains(LightThemes, l) {
		v.LightTheme = l
	}
	if slices.Contains(DarkThemes, d) {
		v.DarkTheme = d
	}
	if v.Mode != "light" && v.Mode != "dark" {
		v.Mode = "system"
	}
	return v
}
func (v Appearance) Theme(systemDark bool) string {
	if v.Mode == "dark" || (v.Mode == "system" && systemDark) {
		return v.DarkTheme
	}
	return v.LightTheme
}
func (s *Store) SaveAppearance(v Appearance) error {
	if v.Mode != "system" && v.Mode != "dark" && v.Mode != "light" {
		return errors.New("invalid appearance mode")
	}
	if !slices.Contains(LightThemes, v.LightTheme) || !slices.Contains(DarkThemes, v.DarkTheme) {
		return errors.New("palette does not match its appearance")
	}
	return s.Write("theme", v)
}

type Layout struct {
	SidebarWidth   float64 `json:"sidebar_width,omitempty"`
	SidepaneWidth  float64 `json:"sidepane_width,omitempty"`
	ExplorerWidth  float64 `json:"explorer_width,omitempty"`
	TerminalHeight float64 `json:"terminal_height,omitempty"`
}

func normalizeLayout(v Layout) Layout {
	for _, p := range []*float64{&v.SidebarWidth, &v.SidepaneWidth, &v.ExplorerWidth, &v.TerminalHeight} {
		if math.IsNaN(*p) || math.IsInf(*p, 0) || *p < 80 || *p > 4000 {
			*p = 0
		}
	}
	return v
}
func (s *Store) Layout() Layout            { var v Layout; s.Read("layout.json", &v); return normalizeLayout(v) }
func (s *Store) SaveLayout(v Layout) error { return s.Write("layout.json", normalizeLayout(v)) }

type Notifications struct {
	Desktop bool `json:"desktop"`
	Sound   bool `json:"sound"`
	Toasts  bool `json:"toasts"`
}

func (s *Store) Notifications() Notifications {
	m := s.raw("notifications.json")
	return Notifications{boolean(m, "desktop", true), boolean(m, "sound", true), boolean(m, "toasts", true)}
}
func (s *Store) SaveNotifications(v Notifications) error { return s.Write("notifications.json", v) }

var AccessModes = []string{"supervised", "auto-accept-edits", "full-access"}
var WorkflowModes = []string{"plan", "build", "ask"}

func (s *Store) Access() string {
	v := str(s.raw("access.json"), "mode", "full-access")
	if !slices.Contains(AccessModes, v) {
		return "full-access"
	}
	return v
}
func (s *Store) SaveAccess(v string) error {
	if !slices.Contains(AccessModes, v) {
		return errors.New("invalid access mode")
	}
	return s.Write("access.json", map[string]string{"mode": v})
}
func (s *Store) Workflow(session string) string {
	v := str(s.raw("workflow.json"), session, "build")
	if !slices.Contains(WorkflowModes, v) {
		return "build"
	}
	return v
}
func (s *Store) SaveWorkflow(session, mode string) error {
	if session == "" || !slices.Contains(WorkflowModes, mode) {
		return errors.New("invalid session or workflow mode")
	}
	return s.Update("workflow.json", func(m map[string]json.RawMessage) error { m[session] = rawValue(mode); return nil })
}

type Model struct {
	Provider string `json:"provider"`
	ID       string `json:"id"`
}
type SessionDefault struct {
	Provider string `json:"provider,omitempty"`
	ModelID  string `json:"model_id,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

func (s *Store) SessionDefault() SessionDefault {
	m := s.raw("session-defaults.json")
	v := SessionDefault{str(m, "provider", ""), str(m, "model_id", str(m, "modelId", "")), str(m, "thinking", "")}
	if v.ModelID == "" {
		return SessionDefault{}
	}
	return v
}
func (v SessionDefault) Args() []string {
	if v.ModelID == "" {
		return nil
	}
	var a []string
	if v.Provider != "" {
		a = append(a, "--provider", v.Provider, "--model", v.Provider+"/"+v.ModelID)
	} else {
		a = append(a, "--model", v.ModelID)
	}
	if v.Thinking != "" {
		a = append(a, "--thinking", v.Thinking)
	}
	return a
}
func (s *Store) SaveSessionDefault(v SessionDefault) error {
	if v.ModelID == "" {
		v = SessionDefault{}
	}
	return s.Write("session-defaults.json", v)
}

type AutoTitle struct {
	Enabled bool   `json:"enabled"`
	Model   *Model `json:"model"`
}

func (s *Store) AutoTitle() AutoTitle {
	m := s.raw("auto-title.json")
	v := AutoTitle{Enabled: boolean(m, "enabled", true)}
	json.Unmarshal(m["model"], &v.Model)
	if v.Model != nil && (v.Model.ID == "" || v.Model.Provider == "") {
		v.Model = nil
	}
	return v
}
func (s *Store) SaveAutoTitle(v AutoTitle) error {
	if v.Model != nil && (v.Model.ID == "" || v.Model.Provider == "") {
		return errors.New("incomplete title model")
	}
	return s.Write("auto-title.json", v)
}

type Retention struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
	Days    int    `json:"days"`
}

var RetentionDays = []int{7, 14, 30, 60, 90, 180, 365}

func (s *Store) Retention() Retention {
	m := s.raw("session-retention.json")
	v := Retention{boolean(m, "enabled", false), str(m, "mode", "archive"), int(num(m, "days", 30))}
	if v.Mode != "delete" {
		v.Mode = "archive"
	}
	if !slices.Contains(RetentionDays, v.Days) {
		v.Days = 30
	}
	return v
}
func (s *Store) SaveRetention(v Retention) error {
	if (v.Mode != "archive" && v.Mode != "delete") || !slices.Contains(RetentionDays, v.Days) {
		return errors.New("invalid retention policy")
	}
	return s.Write("session-retention.json", v)
}
func (s *Store) SessionSet(archived bool) []string {
	name := "pinned-sessions.json"
	if archived {
		name = "archived-sessions.json"
	}
	var v struct {
		Sessions []string `json:"sessions"`
	}
	s.Read(name, &v)
	return v.Sessions
}
func (s *Store) SetSession(archived bool, path string, enabled bool) error {
	name := "pinned-sessions.json"
	if archived {
		name = "archived-sessions.json"
	}
	return s.Update(name, func(m map[string]json.RawMessage) error {
		var a []string
		json.Unmarshal(m["sessions"], &a)
		a = slices.DeleteFunc(a, func(x string) bool { return x == path })
		if enabled {
			a = append(a, path)
		}
		m["sessions"] = rawValue(a)
		return nil
	})
}
func (s *Store) Models(recent bool) []Model {
	name := "favorites.json"
	if recent {
		name = "recent-models.json"
	}
	var v struct {
		Models []Model `json:"models"`
	}
	s.Read(name, &v)
	return v.Models
}
func (s *Store) SetModel(recent bool, model Model, enabled bool) error {
	if model.ID == "" || model.Provider == "" {
		return errors.New("model id and provider required")
	}
	name := "favorites.json"
	if recent {
		name = "recent-models.json"
	}
	return s.Update(name, func(m map[string]json.RawMessage) error {
		var a []Model
		json.Unmarshal(m["models"], &a)
		a = slices.DeleteFunc(a, func(x Model) bool { return x == model })
		if enabled {
			if recent {
				a = append([]Model{model}, a...)
			} else {
				a = append(a, model)
			}
		}
		if recent && len(a) > 5 {
			a = a[:5]
		}
		m["models"] = rawValue(a)
		return nil
	})
}
