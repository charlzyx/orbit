package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSettingsChineseAllMountedPagesRoundTrip(t *testing.T) {
	home := t.TempDir()
	x, err := NewController(RunOptions{Home: home, Workspace: t.TempDir(), PiBinary: filepath.Join(home, "missing-fixture-pi"), Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	a := x.UI
	a.page = "settings"
	a.nativeSettings.privacyLoaded = true
	if a.Options.MCP == nil || a.Options.Resources == nil || a.Options.BugReport == nil {
		t.Fatal("actual subpages not mounted")
	}
	tt := ui.NewTester(a.View, 1470, 1100)
	for _, cat := range settingsCategories {
		t.Run(cat.ID, func(t *testing.T) {
			a.settingsPage = cat.ID
			for _, lang := range []string{"en", "zh-CN", "en"} {
				a.Prefs.Language = lang
				a.reloadLocale()
				tt.Frame()
				want := a.settingsTitle(cat.ID, cat.Title)
				if !tt.HasText(want) {
					t.Fatalf("%s title %q absent: %v", lang, want, tt.Texts())
				}
				// Real child pages must expose their own localized content, not App's fallback.
				key := map[string]string{"general": "settings.notifications", "appearance": "language.title", "shortcuts": "shortcut.group_composer", "agent": "settings.auto_retry", "worktrees": "worktree.settings.manage", "providers": "settings.add_provider", "skills": "app.search_skills", "plugins": "settings.install", "about": "settings.native_workbench_for_the_pi_coding_agent"}[cat.ID]
				switch cat.ID {
				case "runtime":
					key = "go.runtime.patches"
				case "privacy":
					key = "settings.session_retention"
				case "mcp":
					key = "mcp.search_servers"
				case "models":
					key = "models.disconnected"
				case "bug":
					key = "bug_report.review_intro"
				}
				if key != "" && !tt.HasText(a.Tr(key)) {
					t.Fatalf("%s child key %s missing: %v", lang, key, tt.Texts())
				}
			}
		})
	}
}
func TestSettingsChineseRuntimeActionsPreserveIdentity(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "runtime"
	a.Runtime = agent.RuntimeSnapshot{Active: &agent.RuntimeProcess{State: "idle", Alive: true, PID: 123, Executable: "/fixture/pi", Error: "backend raw: TOKEN=${KEEP}"}}
	var got []Action
	a.Options.OnAction = func(v Action) { got = append(got, v) }
	tt := ui.NewTester(a.View, 1470, 1100)
	for _, lang := range []string{"en", "zh-CN", "en"} {
		a.Prefs.Language = lang
		a.reloadLocale()
		tt.Frame()
		if !tt.HasText(a.Tr("go.runtime.patches")) || !tt.HasText("backend raw: TOKEN=${KEEP}") {
			t.Fatal(tt.Texts())
		}
		if err := tt.Click("Stop"); err != nil {
			t.Fatal(err)
		}
		if got[len(got)-1].Kind != "StopRuntime" {
			t.Fatal(got)
		}
	}
}
func TestSettingsChineseSearchUsesLabelsAndOriginalKeywords(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	tt := ui.NewTester(a.View, 1470, 1100)
	for _, q := range []string{"进程 ID", "pid"} {
		a.settingsSearch = q
		tt.Frame()
		if !tt.HasText(a.Tr("settings.process_id")) {
			t.Fatal(q, tt.Texts())
		}
		if err := tt.Click(a.Tr("settings.process_id")); err != nil {
			t.Fatal(err)
		}
		if a.settingsPage != "runtime" || a.settingsSearch != "" {
			t.Fatal("search changed section identity")
		}
	}
}
func TestSettingsChineseProviderValidationLiveLocaleAndRawDetails(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "providers"
	a.editProvider(workbench.Provider{})
	d := a.providerState.editor
	tt := ui.NewTester(a.View, 1470, 1100)
	tt.Frame()
	if err := tt.Click("Save provider"); err != nil {
		t.Fatal(err)
	}
	if d.error != "provider id is required" {
		t.Fatal("actual validation not reached", d.error)
	}
	for _, lang := range []string{"en", "zh-CN", "en"} {
		a.Prefs.Language = lang
		a.reloadLocale()
		tt.Frame()
		if !tt.HasText(a.Tr("go.providers.validation_0")) {
			t.Fatal(tt.Texts())
		}
		if d.error != "provider id is required" {
			t.Fatal("stored error mutated")
		}
	}
	d.error = "remote raw detail: api_key=fixture-only"
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	tt.Frame()
	if !tt.HasText(d.error) {
		t.Fatal("raw detail altered")
	}
	if err := tt.Click("Provider ID"); err != nil {
		t.Fatal(err)
	}
	tt.Type("custom-provider")
	if d.edit.ID != "custom-provider" {
		t.Fatal(d.edit.ID)
	}
}
func TestSettingsChinesePrivacyConfirmationAndCancel(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "privacy"
	a.nativeSettings.privacyLoaded = true
	a.retention.Mode = "delete"
	a.retention.Days = 30
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 1100)
	tt.Frame()
	if err := tt.Click("Run session cleanup"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("永久删除旧任务？") || !strings.Contains(strings.Join(tt.Texts(), "\n"), "30 天") {
		t.Fatal(tt.Texts())
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	for _, v := range actions {
		if v.Kind == "RunRetention" {
			t.Fatal("cancel ran cleanup")
		}
	}
}
func TestSettingsChineseMountedCaptures(t *testing.T) {
	out := os.Getenv("ORBIT_SETTINGS_ZH_CAPTURE_DIR")
	if out == "" {
		out = t.TempDir()
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	x, err := NewController(RunOptions{Home: home, Workspace: t.TempDir(), PiBinary: filepath.Join(home, "missing-fixture-pi"), Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	a := x.UI
	a.page = "settings"
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	a.nativeSettings.privacyLoaded = true
	tt := ui.NewTester(a.View, 1470, 1100)
	for _, mode := range []string{"light", "dark"} {
		a.Appearance.Mode = mode
		id := a.Appearance.DarkTheme
		if mode == "light" {
			id = a.Appearance.LightTheme
		}
		a.Theme = NewTheme(id, a.Prefs, a.Fonts)
		for _, page := range []string{"runtime", "privacy", "mcp", "providers", "appearance", "about"} {
			a.settingsPage = page
			tt.Frame()
			if a.Theme.Dark != (mode == "dark") {
				t.Fatalf("%s theme did not apply", mode)
			}
			f, e := os.Create(filepath.Join(out, page+"-zh-"+mode+".png"))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, tt.Image())
			f.Close()
			if e != nil {
				t.Fatal(e)
			}
		}
	}
}

func TestSettingsChineseProviderUsageButtonStableAction(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "providers"
	a.providerState.catalog = workbench.ProviderCatalog{Providers: []workbench.ProviderMetadata{{ID: "fixture-provider", Name: "Fixture provider"}}}
	a.providerState.usageProvider = "fixture-provider"
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 1100)
	for _, lang := range []string{"en", "zh-CN", "en"} {
		a.Prefs.Language = lang
		a.reloadLocale()
		tt.Frame()
		if !tt.HasText(a.Tr("go.providers.refresh_usage")) {
			t.Fatal(tt.Texts())
		}
		if e := tt.Click("Refresh usage"); e != nil {
			t.Fatal(e)
		}
	}
	if len(actions) != 3 {
		t.Fatal(actions)
	}
	for _, a := range actions {
		if a.Kind != "RefreshQuota" {
			t.Fatal(a)
		}
	}
}
