package mcpui

import (
	"errors"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsChineseMCPEditorLiveSwitchAndWireValues(t *testing.T) {
	s := stateFixture(t)
	en, _ := orbitassets.Locale("en")
	lang := "en"
	s.Options.Text = func(k, f string) string {
		m, _ := orbitassets.Locale(lang)
		v := orbitassets.Translate(m, en, k, nil)
		if v == k {
			return f
		}
		return v
	}
	s.BeginAdd()
	d := s.Editor
	d.Name = "fixture-server"
	d.Command = "node"
	d.Env = "TOKEN=${EXACT_KEY}"
	d.Scope = "global"
	d.Exposure = "codemode"
	tt := testerFor(s, 1000, 1200)
	for _, l := range []string{"en", "zh-CN", "en"} {
		lang = l
		tt.Frame()
		if !tt.HasText(s.label("mcp.environment", "Environment variables")) || !tt.HasText(s.label("mcp.scope_global_hint", "")) {
			t.Fatal(l, tt.Texts())
		}
		if d.Name != "fixture-server" || d.Command != "node" || d.Env != "TOKEN=${EXACT_KEY}" || d.Scope != "global" || d.Exposure != "codemode" || d.Transport != "stdio" {
			t.Fatal("wire changed", d)
		}
	}
	d.Name = "invalid name"
	s.saveEditor()
	if !d.validationError {
		t.Fatal("validation path not reached")
	}
	lang = "zh-CN"
	tt.Frame()
	if !tt.HasText("服务器名称只能包含字母、数字、连字符或下划线。") {
		t.Fatal(tt.Texts())
	}
	if out := os.Getenv("ORBIT_SETTINGS_ZH_CAPTURE_DIR"); out != "" {
		if e := os.MkdirAll(out, 0755); e != nil {
			t.Fatal(e)
		}
		f, e := os.Create(filepath.Join(out, "mcp-editor-zh-dark.png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, tt.Image())
		f.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
	d.validationError = false
	tt.Frame()
	if !tt.HasText(d.Error) {
		t.Fatal("same-text backend error was translated")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.HasModal() {
		t.Fatal("escape did not cancel")
	}
}

func TestSettingsChineseMCPBusyErrorLiveLocaleAndBackendOrigin(t *testing.T) {
	s := stateFixture(t)
	en, _ := orbitassets.Locale("en")
	lang := "en"
	s.Options.Text = func(k, f string) string {
		m, _ := orbitassets.Locale(lang)
		v := orbitassets.Translate(m, en, k, nil)
		if v == k {
			return f
		}
		return v
	}
	s.Busy = true
	s.RefreshAndApply()
	raw := s.Error
	if !s.localBusyError || raw != "Configuration operation in progress. Retry reconnect when it finishes." {
		t.Fatal("busy path not reached", raw)
	}
	tt := testerFor(s, 1200, 1100)
	for _, l := range []string{"en", "zh-CN", "en"} {
		lang = l
		tt.Frame()
		if !tt.HasText(s.label("go.mcp.busy", raw)) {
			t.Fatal(l, tt.Texts())
		}
		if s.Error != raw {
			t.Fatal("raw error mutated")
		}
	}
	s.results <- completion{generation: s.generation, err: errors.New(raw)}
	s.Tick()
	lang = "zh-CN"
	tt.Frame()
	if s.localBusyError || !tt.HasText(raw) {
		t.Fatal("backend same-text error translated")
	}
}
