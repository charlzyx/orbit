package resourceui

import (
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "orbit-resource-ui-cache-")
	if err != nil {
		panic(err)
	}
	os.Setenv("XDG_CACHE_HOME", dir)
	fs.WalkDir(orbitassets.Assets(), "fonts", func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".ttf") {
			family := ""
			if strings.Contains(strings.ToLower(path), "plexsans") {
				family = ".ZedSans"
			}
			if strings.Contains(strings.ToLower(path), "lilex") {
				family = ".ZedMono"
			}
			if family != "" {
				data, _ := orbitassets.Asset(path)
				ui.RegisterFont(data, family)
			}
		}
		return nil
	})
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
func put(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) *State {
	t.Helper()
	store := preferences.New(t.TempDir())
	workspace := t.TempDir()
	put(t, filepath.Join(workspace, ".git", "HEAD"), "ref: refs/heads/main\n")
	body := "---\nname: native-review\ndescription: Review native user interfaces with isolated, real file fixtures.\ndisable-model-invocation: true\n---\n# Native Review\n\n## Overview\n\nA real on-disk skill used only by this native UI test.\n\n- Check layout geometry\n- Verify keyboard navigation\n- Preserve settings\n\n## Instructions\n\nUse `go test` with isolated temporary directories.\n\n```go\nfunc verify() bool { return true }\n```\n"
	put(t, filepath.Join(workspace, ".pi", "skills", "native-review", "SKILL.md"), body)
	put(t, filepath.Join(store.PiDir, "skills", "agent-helper", "SKILL.md"), "---\nname: agent-helper\ndescription: Agent-scoped skill fixture.\n---\n# Agent helper\n")
	put(t, filepath.Join(store.Home, ".agents", "skills", "invalid-skill", "SKILL.md"), "# Invalid Skill Body\nNo description frontmatter.\n")
	put(t, filepath.Join(store.PiDir, "settings.json"), `{"packages":["npm:fixture-plugin","./local-tool","npm:missing-plugin"],"unrelated":{"keep":true}}`)
	put(t, filepath.Join(store.PiDir, "npm", "node_modules", "fixture-plugin", "package.json"), `{"name":"fixture-plugin","version":"1.2.3"}`)
	put(t, filepath.Join(store.PiDir, "local-tool", "package.json"), `{"name":"local-tool","version":"0.4.0"}`)
	s := New(store, workspace, Options{Offline: true})
	s.Refresh()
	waitState(t, s)
	t.Cleanup(s.Close)
	return s
}
func waitState(t *testing.T, s *State) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for s.Busy || s.refreshPending || s.checkPending {
		s.Tick()
		if time.Now().After(deadline) {
			t.Fatalf("state did not settle: %s / %s", s.activity, s.Error)
		}
		time.Sleep(time.Millisecond)
	}
}
func fakePi(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX explicit fake-process fixture")
	}
	path := filepath.Join(t.TempDir(), "fixture-pi")
	put(t, path, "#!/bin/sh\n"+body+"\n")
	if err := os.Chmod(path, 0700); err != nil {
		t.Fatal(err)
	}
	return path
}
func tester(s *State, page string, w, h int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme := ui.DarkTheme()
		theme.FontSize = 14
		theme.Font = ".ZedSans"
		theme.Background = ui.Hex("#191919")
		theme.Surface = ui.Hex("#171717")
		theme.Border = ui.Hex("#2b2b2b")
		theme.Text = ui.Hex("#d4d4d4")
		theme.TextMuted = ui.Hex("#858585")
		theme.Accent = ui.Hex("#df7758")
		theme.Selection = ui.Hex("#37271f")
		c.SetTheme(theme)
		s.Options.Theme = func() Theme { return Theme{Elevated: ui.Hex("#242424"), CodeFont: ".ZedMono"} }
		ui.Row(c).Grow(1).Background(theme.Background).Children(func() { s.View(c, page) })
	}, w, h)
}
func click(t *testing.T, tt *ui.Tester, label string) {
	t.Helper()
	if err := tt.Click(label); err != nil {
		t.Fatal(err)
	}
}

func TestSourceSkillsDiscoveryBodySelectionAndFilter(t *testing.T) {
	s := fixture(t)
	if len(s.Skills) != 3 {
		t.Fatal(s.Skills, s.SkillErrors)
	}
	if s.Selected == "" {
		t.Fatal("refresh did not select first skill")
	}
	var target workbench.Skill
	for _, p := range s.Skills {
		if p.Name == "native-review" {
			target = p
		}
	}
	s.SelectSkill(target.Path)
	tt := tester(s, "skills", 1230, 922)
	if !tt.HasText("Native Review") || tt.HasText("disable-model-invocation: true") || !tt.HasText("manual only") {
		t.Fatalf("markdown/frontmatter/manual state wrong: %v", tt.Texts())
	}
	click(t, tt, "Search skills")
	tt.Type("No such skill")
	if !tt.HasText("No skills match") {
		t.Fatal("search did not filter")
	}
	if s.Selected != target.Path {
		t.Fatal("filter changed detail selection")
	}
	s.SkillQuery = "agent-scoped"
	tt.Frame()
	if len(s.visibleSkills()) != 1 {
		t.Fatal("description search")
	}
	s.Refresh()
	waitState(t, s)
	if s.Selected != target.Path {
		t.Fatal("refresh discarded selected path")
	}
}
func TestNativeSkillActionsCallbacksAndCancel(t *testing.T) {
	s := fixture(t)
	p := s.Skills[0]
	s.SelectSkill(p.Path)
	var opened, revealed, copied string
	s.Options.OpenFile = func(v string) { opened = v }
	s.Options.RevealPath = func(v string) { revealed = v }
	s.Options.CopyText = func(v string) { copied = v }
	tt := tester(s, "skills", 1230, 922)
	click(t, tt, "Open SKILL.md")
	click(t, tt, "Show in File Manager")
	click(t, tt, "Copy path")
	if opened != p.Path || revealed != p.Path || copied != s.displayPath(p.Path) {
		t.Fatal(opened, revealed, copied)
	}
	click(t, tt, "Delete")
	if !s.HasModal() {
		t.Fatal("missing inline confirmation")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.HasModal() {
		t.Fatal("Escape did not cancel")
	}
	if _, err := os.Stat(p.Path); err != nil {
		t.Fatal("cancel deleted file", err)
	}
}
func TestNativeSkillTogglePersistsAndDeletionRequiresSecondAction(t *testing.T) {
	s := fixture(t)
	p := s.Skills[0]
	s.SelectSkill(p.Path)
	modified := 0
	s.Options.OnModified = func() { modified++ }
	tt := tester(s, "skills", 1230, 922)
	click(t, tt, "Enable skill")
	waitState(t, s)
	tt.Frame()
	if s.skill(p.Path).Enabled || modified != 1 {
		t.Fatal("toggle did not persist", s.Error, modified)
	}
	reread, _ := s.service.Skills()
	found := false
	for _, v := range reread {
		if v.Path == p.Path {
			found = !v.Enabled
		}
	}
	if !found {
		t.Fatal("toggle was not on disk")
	}
	click(t, tt, "Delete")
	if s.Busy {
		t.Fatal("first delete wrote files")
	}
	click(t, tt, "Delete")
	waitState(t, s)
	tt.Frame()
	if _, err := os.Stat(p.Path); !os.IsNotExist(err) {
		t.Fatal("confirmed delete did not remove real fixture", err, s.Error)
	}
	if s.Selected == p.Path || s.Selected == "" || len(s.Skills) != 2 || modified != 2 {
		t.Fatal("delete did not refresh next selected skill", s.Selected, modified)
	}
}
func TestSkillsReadWriteErrorsRemainVisibleAndNoOptimism(t *testing.T) {
	s := fixture(t)
	var p workbench.Skill
	for _, v := range s.Skills {
		if v.Scope == "global" {
			p = v
		}
	}
	s.SelectSkill(p.Path)
	path := filepath.Join(s.service.Preferences.PiDir, "settings.json")
	put(t, path, "{broken")
	s.SetSkillEnabled(p.Path, false)
	waitState(t, s)
	if s.Error == "" || s.skill(p.Path) == nil || !s.skill(p.Path).Enabled {
		t.Fatal("failed write claimed disabled", s.Error)
	}
	if data, _ := os.ReadFile(path); string(data) != "{broken" {
		t.Fatal("error overwrote malformed settings")
	}
	tt := tester(s, "skills", 1230, 922)
	if !tt.HasText("Action failed") {
		t.Fatal("write failure not visible")
	}
}
func TestPluginScopeSearchVersionsAndRemovalCancel(t *testing.T) {
	s := fixture(t)
	tt := tester(s, "plugins", 1230, 922)
	if !tt.HasText("v1.2.3") || !tt.HasText("Not installed") {
		t.Fatal("real installed version/diagnostics missing")
	}
	click(t, tt, "Project")
	if s.InstallScope != "project" {
		t.Fatal("scope control")
	}
	click(t, tt, "Search installed plugins")
	tt.Type("1.2.3")
	if len(s.visiblePlugins()) != 1 {
		t.Fatal("version search")
	}
	click(t, tt, "Remove")
	if !s.HasModal() || s.Busy {
		t.Fatal("remove started without confirm")
	}
	click(t, tt, "Cancel")
	if s.HasModal() {
		t.Fatal("cancel did not dismiss")
	}
	if data, _ := os.ReadFile(filepath.Join(s.service.Preferences.PiDir, "settings.json")); !strings.Contains(string(data), "fixture-plugin") {
		t.Fatal("cancel mutated packages")
	}
}
func TestPluginInstallRealSettingsFakeProcessLifecycle(t *testing.T) {
	s := fixture(t)
	script := fakePi(t, `printf '%s\n' "$@" >> "$HOME/argv"
mkdir -p "$PWD/.pi"
printf '%s\n' '{"packages":["npm:added-fixture"]}' > "$PWD/.pi/settings.json"
printf '%s\n' 'progress line' 'Installed fixture only'`)
	s.SetPiBinary(script)
	s.InstallScope = "project"
	s.InstallSource = " npm:added-fixture "
	modified := 0
	s.Options.OnModified = func() { modified++ }
	s.Install()
	s.Install()
	waitState(t, s)
	data, err := os.ReadFile(filepath.Join(s.service.Preferences.Home, "argv"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "install\nnpm:added-fixture\n-l\n-a\n" {
		t.Fatalf("wrong/duplicate CLI argv %q", data)
	}
	found := false
	for _, p := range s.Plugins {
		if p.Source == "npm:added-fixture" && p.Scope == "project" {
			found = true
		}
	}
	if !found || s.InstallSource != " npm:added-fixture " || s.Status != "Installed fixture only" || modified != 1 {
		t.Fatal("completion not refreshed", s.Status, s.Error, modified, s.Plugins)
	}
	var m map[string]any
	data, _ = os.ReadFile(filepath.Join(s.service.Preferences.PiDir, "settings.json"))
	json.Unmarshal(data, &m)
	if m["unrelated"] == nil {
		t.Fatal("project install touched global file")
	}
}
func TestPluginFailureCancellationAndRepeatedRetry(t *testing.T) {
	s := fixture(t)
	s.SetPiBinary(fakePi(t, `printf '%s\n' 'fixture installation failed' >&2; exit 9`))
	s.InstallSource = "npm:failure"
	s.Install()
	waitState(t, s)
	if !strings.Contains(s.Error, "fixture installation failed") || s.Busy || s.InstallSource != "npm:failure" {
		t.Fatal(s.Error, s.Busy, s.InstallSource)
	}
	started := filepath.Join(s.service.Preferences.Home, "started")
	s.SetPiBinary(fakePi(t, `echo started > "$HOME/started"; sleep 60`))
	s.Install()
	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(started); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fake process did not start")
		}
		time.Sleep(time.Millisecond)
	}
	s.Cancel()
	if !s.Busy {
		t.Fatal("cancel released duplicate-operation guard before reap")
	}
	waitState(t, s)
	if s.Busy || !strings.Contains(s.Status, "Cancelled") {
		t.Fatal("cancel never settled", s.Error, s.Status)
	}
	s.SetPiBinary(fakePi(t, `echo retry-complete`))
	s.Install()
	waitState(t, s)
	if s.Status != "retry-complete" || s.Error != "" {
		t.Fatal("retry failed", s.Error, s.Status)
	}
}
func TestWorkspaceGenerationAndCloseRejectStaleCompletion(t *testing.T) {
	s := fixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	s.start("fixture pending", "fixture", "", func(context.Context) (func(*State), error) {
		close(started)
		<-release
		return func(s *State) { s.Error = "stale workspace result" }, nil
	})
	<-started
	newWorkspace := t.TempDir()
	s.SetWorkspace(newWorkspace)
	close(release)
	deadline := time.Now().Add(time.Second)
	for len(s.results) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	s.Tick()
	if s.Error != "" || s.Workspace != newWorkspace || s.Busy {
		t.Fatal("old completion applied")
	}
	s.Close()
	if s.start("closed", "fixture", "", func(context.Context) (func(*State), error) { t.Error("closed job ran"); return nil, nil }) {
		t.Fatal("closed state accepted work")
	}
}

func TestNavigationWaitsForCanceledWorkerBeforeAllowingNewOperation(t *testing.T) {
	s := fixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	s.start("fixture operation", "install", "old-workspace", func(context.Context) (func(*State), error) {
		close(started)
		<-release
		return func(s *State) { s.Status = "obsolete" }, nil
	})
	<-started
	s.SetWorkspace(t.TempDir())
	if !s.Busy {
		t.Fatal("navigation released process ownership before reap")
	}
	if s.start("duplicate", "install", "new-workspace", func(context.Context) (func(*State), error) {
		t.Error("second operation ran while canceled first still alive")
		return nil, nil
	}) {
		t.Fatal("duplicate operation accepted")
	}
	close(release)
	waitState(t, s)
	if s.Status == "obsolete" || s.Busy {
		t.Fatal("stale result applied or guard stranded", s.Status, s.Busy)
	}
}
func TestSkipAndResetRealPersistenceAndBatchPartialFailure(t *testing.T) {
	s := fixture(t)
	s.Updates = []workbench.PluginUpdate{{Source: "npm:fixture-plugin", Name: "fixture-plugin", Scope: "global", Current: "1.2.3", Latest: "2.0.0"}}
	s.SkipAll()
	waitState(t, s)
	if len(s.Updates) != 0 || s.Skipped["npm:fixture-plugin"] != "2.0.0" {
		t.Fatal("skip not persisted", s.Error, s.Skipped)
	}
	skips, _ := s.service.SkippedPluginUpdates()
	if skips["npm:fixture-plugin"] != "2.0.0" {
		t.Fatal("skip missing disk")
	}
	s.SkipUpdate("npm:fixture-plugin", "")
	waitState(t, s)
	if len(s.Skipped) != 0 {
		t.Fatal("reset did not persist")
	}
	s.SetPiBinary(fakePi(t, `printf '%s\n' "$@" >> "$HOME/argv"; if [ "$2" = npm:bad-fixture ]; then echo fail >&2; exit 1; fi; echo updated`))
	s.Updates = []workbench.PluginUpdate{{Source: "npm:bad-fixture", Name: "bad-fixture", Scope: "global", Latest: "2"}, {Source: "npm:fixture-plugin", Name: "fixture-plugin", Scope: "project", Latest: "2"}}
	s.UpdateAll()
	s.UpdateAll()
	waitState(t, s)
	if len(s.Updates) != 1 || s.Updates[0].Source != "npm:bad-fixture" || !strings.Contains(s.Error, "bad-fixture") || !strings.Contains(s.Status, "Updated 1") {
		t.Fatal("partial batch state", s.Updates, s.Error, s.Status)
	}
	data, _ := os.ReadFile(filepath.Join(s.service.Preferences.Home, "argv"))
	if string(data) != "update\nnpm:bad-fixture\nupdate\nnpm:fixture-plugin\n" {
		t.Fatalf("batch argv/guard %q", data)
	}
}
func TestOfflineLaunchAndExplicitCheckNeverProbe(t *testing.T) {
	s := fixture(t)
	s.launchCheckDue = time.Now().Add(-time.Second)
	s.Tick()
	if s.Busy || s.Checking {
		t.Fatal("offline launch probed")
	}
	s.CheckUpdates()
	if s.Busy {
		t.Fatal("offline manual check probed")
	}
	s.Options.Offline = false
	t.Setenv("PI_OFFLINE", "1")
	s.CheckUpdates()
	if s.Busy {
		t.Fatal("PI_OFFLINE check probed")
	}
}

func TestSourceFactFormattingAndMutationNotifications(t *testing.T) {
	s := fixture(t)
	now := time.Now()
	for _, test := range []struct {
		bytes int64
		want  string
	}{{512, "512 B"}, {12687, "12.4 KB"}, {2097152, "2.0 MB"}} {
		if got := formatSize(test.bytes); got != test.want {
			t.Fatal(got, test.want)
		}
	}
	if s.skillAge(time.Time{}, now) != "Unknown" || s.skillAge(now.Add(-48*time.Hour), now) != "2d ago" || s.skillAge(now, now) != "now ago" {
		t.Fatal("source time formatting diverged")
	}
	var messages []string
	var errors []bool
	s.Options.Notify = func(text string, isError bool) { messages = append(messages, text); errors = append(errors, isError) }
	s.Refresh()
	waitState(t, s)
	if len(messages) != 0 {
		t.Fatal("routine refresh notified")
	}
	p := s.Skills[0]
	s.SetSkillEnabled(p.Path, false)
	waitState(t, s)
	if len(messages) != 1 || errors[0] || messages[0] != "Disabled "+p.Name {
		t.Fatal("toggle notification", messages, errors)
	}
	var path string
	if p.Scope == "project" {
		path = filepath.Join(s.Workspace, ".pi", "settings.json")
	} else {
		path = filepath.Join(s.service.Preferences.PiDir, "settings.json")
	}
	put(t, path, "{malformed")
	s.SetSkillEnabled(p.Path, true)
	waitState(t, s)
	if len(messages) != 2 || !errors[1] || !strings.Contains(messages[1], "invalid JSON") {
		t.Fatal("error notification", messages, errors)
	}
}
func TestNativeCaptureResources(t *testing.T) {
	out := os.Getenv("ORBIT_RESOURCE_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_RESOURCE_CAPTURE_DIR for explicit native CPU fixture captures")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{1470, 922}, {800, 560}} {
		for _, page := range []string{"skills", "plugins"} {
			s := fixture(t)
			for _, p := range s.Skills {
				if p.Name == "native-review" {
					s.SelectSkill(p.Path)
				}
			}
			s.Options.OpenFile = func(string) {}
			s.Options.RevealPath = func(string) {}
			s.Options.CopyText = func(string) {}
			tt := tester(s, page, size[0], size[1])
			tt.Frame()
			file, err := os.Create(filepath.Join(out, fmtSize(page, size)+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err = png.Encode(file, tt.Image()); err != nil {
				t.Fatal(err)
			}
			file.Close()
		}
	}
}
func fmtSize(page string, size [2]int) string {
	return fmt.Sprintf("%s-native-%dx%d", page, size[0], size[1])
}
