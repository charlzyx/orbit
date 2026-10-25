package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestDefaultsDoNotWriteOrChangePi(t *testing.T) {
	s := New(t.TempDir())
	if s.UI() != DefaultUI() || s.Appearance() != DefaultAppearance() || s.Access() != "full-access" || s.Workflow("missing") != "build" || !s.AutoTitle().Enabled || s.Retention().Enabled {
		t.Fatal("wrong defaults")
	}
	if _, e := os.Stat(s.RootDir); !os.IsNotExist(e) {
		t.Fatal("reading created settings")
	}
}
func TestUILegacyMigrationAndValidation(t *testing.T) {
	s := New(t.TempDir())
	s.Write("ui.json", map[string]any{"interface_scale": 110, "code_font_size": 18, "spacing_density": 95, "language": "zh-CN", "reduce_motion": true, "composer_send_mode": "steer"})
	v := s.UI()
	if v.UIFontSize != 15 || v.EditorFontSize != 18 || v.SpacingDensity != 95 || v.Language != "zh-CN" || !v.ReduceMotion || v.ComposerSendMode != "steer" {
		t.Fatalf("%+v", v)
	}
	s.Write("ui.json", map[string]any{"ui_font_size": 999, "editor_font_size": 9, "terminal_font_size": true, "language": "invalid", "spacing_density": 1})
	if s.UI() != DefaultUI() {
		t.Fatalf("%+v", s.UI())
	}
}
func TestAppearanceMigrationAndInactivePalette(t *testing.T) {
	s := New(t.TempDir())
	for _, tc := range []struct{ value, mode, theme string }{{"flexoki", "dark", "flexoki"}, {"one-light", "light", "orbit-light"}, {"not-a-theme", "system", "orbit-azure-light"}} {
		os.MkdirAll(s.RootDir, 0700)
		os.WriteFile(filepath.Join(s.RootDir, "theme"), []byte(tc.value), 0600)
		v := s.Appearance()
		if v.Mode != tc.mode || v.Theme(false) != tc.theme {
			t.Fatalf("%s: %+v", tc.value, v)
		}
	}
	v := Appearance{"system", "flexoki-light", "orbit-cobalt"}
	if e := s.SaveAppearance(v); e != nil {
		t.Fatal(e)
	}
	if s.Appearance() != v || v.Theme(true) != "orbit-cobalt" {
		t.Fatal("lost selection")
	}
	v.DarkTheme = "orbit-light"
	if s.SaveAppearance(v) == nil {
		t.Fatal("accepted light palette as dark")
	}
}
func TestJSONCompatibilityAndModelIsolation(t *testing.T) {
	s := New(t.TempDir())
	want := SessionDefault{"anthropic", "claude", "high"}
	if e := s.SaveSessionDefault(want); e != nil {
		t.Fatal(e)
	}
	if s.SessionDefault() != want || !reflect.DeepEqual(want.Args(), []string{"--provider", "anthropic", "--model", "anthropic/claude", "--thinking", "high"}) {
		t.Fatal("session defaults")
	}
	if _, e := os.Stat(s.PiDir); !os.IsNotExist(e) {
		t.Fatal("modified Pi defaults")
	}
	if e := s.SaveAccess("supervised"); e != nil {
		t.Fatal(e)
	}
	if s.Access() != "supervised" {
		t.Fatal("access")
	}
	if s.SaveAccess("bypass") == nil {
		t.Fatal("invalid access saved")
	}
	s.SaveWorkflow("one", "ask")
	s.SaveWorkflow("two", "plan")
	if s.Workflow("one") != "ask" || s.Workflow("two") != "plan" {
		t.Fatal("lost workflow")
	}
}
func TestConcurrentUpdatePreservesKeys(t *testing.T) {
	s := New(t.TempDir())
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if e := s.Update("keys.json", func(m map[string]json.RawMessage) error { m[string(rune('A'+i))] = rawValue(i); return nil }); e != nil {
				t.Error(e)
			}
		}(i)
	}
	wg.Wait()
	var m map[string]any
	if e := s.Read("keys.json", &m); e != nil || len(m) != 50 {
		t.Fatalf("%d %v", len(m), e)
	}
}
func TestMalformedAndTraversalWritesRejected(t *testing.T) {
	s := New(t.TempDir())
	os.MkdirAll(s.RootDir, 0700)
	p := filepath.Join(s.RootDir, "workflow.json")
	os.WriteFile(p, []byte("{broken"), 0600)
	if s.SaveWorkflow("id", "ask") == nil {
		t.Fatal("clobbered corrupt data")
	}
	if e := s.Write("../outside", map[string]int{}); e == nil {
		t.Fatal("escaped root")
	}
	b, _ := os.ReadFile(p)
	if string(b) != "{broken" {
		t.Fatal("corrupt data changed")
	}
}
func TestSetsAndRecentsSurviveReopen(t *testing.T) {
	s := New(t.TempDir())
	for i := 0; i < 9; i++ {
		s.SetModel(true, Model{"p", string(rune('a' + i))}, true)
	}
	s.SetSession(false, "/a", true)
	s.SetSession(false, "/a", true)
	s.SetSession(true, "/b", true)
	s = New(s.Home)
	if len(s.Models(true)) != 5 || s.Models(true)[0].ID != "i" || len(s.SessionSet(false)) != 1 || len(s.SessionSet(true)) != 1 {
		t.Fatal("set persistence")
	}
	s.SetSession(false, "/a", false)
	if len(s.SessionSet(false)) != 0 {
		t.Fatal("remove")
	}
}
func TestLayoutBoundsAndPermissions(t *testing.T) {
	s := New(t.TempDir())
	if e := s.SaveLayout(Layout{SidebarWidth: 248, SidepaneWidth: 0, ExplorerWidth: 5000, TerminalHeight: 250}); e != nil {
		t.Fatal(e)
	}
	if got := s.Layout(); got.SidebarWidth != 248 || got.ExplorerWidth != 0 || got.TerminalHeight != 250 {
		t.Fatal(got)
	}
	info, e := os.Stat(filepath.Join(s.RootDir, "layout.json"))
	if e != nil {
		t.Fatal(e)
	}
	if info.Mode().Perm()&0077 != 0 {
		t.Fatalf("unsafe settings permissions: %o", info.Mode().Perm())
	}
}
func TestAliasesMatchOriginal(t *testing.T) {
	for in, want := range map[string]string{"zedokai-filter-spectrum": "orbit", "orbit-minimal": "orbit-cobalt", "one-light-pro": "orbit-light", "flexoki-dark": "flexoki", "unknown": ""} {
		if got := themeAlias(in); got != want {
			t.Errorf("%s=%s want%s", in, got, want)
		}
	}
}
