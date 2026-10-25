package mcpui

import (
	"context"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func TestMain(m *testing.M) {
	dir, e := os.MkdirTemp("", "orbit-mcp-ui-cache-")
	if e != nil {
		panic(e)
	}
	os.Setenv("XDG_CACHE_HOME", dir)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
func stateFixture(t *testing.T) *State {
	t.Helper()
	store := preferences.New(t.TempDir())
	s := New(store, "", Options{})
	s.loaded = true
	s.Servers = []workbench.MCPServer{{Name: "test-server", Scope: "global", Source: filepath.Join(store.PiDir, "mcp.json"), Description: "Explicit native rendering fixture", Definition: map[string]any{"command": "node", "args": []any{"server.js"}, "exposure": "codemode", "enabled": true}}}
	s.Runtime["test-server"] = Runtime{State: "connected", Tools: []Tool{{Name: "read_file", Description: "Fixture tool returned by the fake Pi probe", Exposure: "direct"}}, Resources: 2}
	s.checkedAt = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	t.Cleanup(s.Close)
	return s
}
func testerFor(s *State, w, h int) *ui.Tester {
	return ui.NewTester(func(c *ui.Context) {
		theme := ui.DarkTheme()
		theme.FontSize = 14
		c.SetTheme(theme)
		ui.Column(c).Grow(1).Background(theme.Background).Children(func() { s.View(c) })
	}, w, h)
}
func waitState(t *testing.T, s *State) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for s.Busy || s.pendingProbe {
		if time.Now().After(deadline) {
			t.Fatal("state failed to settle", s.activity, s.Error)
		}
		s.Tick()
		time.Sleep(time.Millisecond)
	}
}
func fakePi(t *testing.T, body string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fake CLI")
	}
	p := filepath.Join(t.TempDir(), "fake-pi")
	if e := os.WriteFile(p, []byte("#!/bin/sh\n"+body+"\n"), 0700); e != nil {
		t.Fatal(e)
	}
	return p
}
func TestNativeEditorMinimumHeightAndCancel(t *testing.T) {
	s := stateFixture(t)
	tt := testerFor(s, 800, 560)
	if e := tt.Click("Add server"); e != nil {
		t.Fatal(e)
	}
	if s.Editor == nil || !s.HasModal() {
		t.Fatal("native editor did not open")
	}
	for _, label := range []string{"Name", "Command", "Save", "Cancel"} {
		r, ok := tt.Find(label)
		if !ok || r.X < 0 || r.Y < 0 || r.X+r.W > 800 || r.Y+r.H > 560 {
			t.Fatalf("%s is not reachable: %+v %v", label, r, ok)
		}
	}
	if e := tt.Click("Name"); e != nil {
		t.Fatal(e)
	}
	tt.Type("draft-name")
	if s.Editor.Name != "draft-name" {
		t.Fatal("native input did not bind")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.Editor != nil || s.HasModal() {
		t.Fatal("Escape did not dismiss native editor")
	}
	if _, e := os.Stat(filepath.Join(s.service.Preferences.PiDir, "mcp.json")); !os.IsNotExist(e) {
		t.Fatal("cancel wrote configuration")
	}
}
func TestNativeListSelectionFilterAndRemovalConfirmation(t *testing.T) {
	s := stateFixture(t)
	tt := testerFor(s, 900, 680)
	if e := tt.Click("Server details: test-server"); e != nil {
		t.Fatal(e)
	}
	if s.Selected != "test-server" || !tt.HasText("read_file") {
		t.Fatal("actual tool detail did not expand")
	}
	if e := tt.Click("Delete"); e != nil {
		t.Fatal(e)
	}
	if s.removeTarget != "test-server" || !tt.HasText("Remove") {
		t.Fatal("remove confirmation absent")
	}
	if e := tt.Click("Cancel"); e != nil {
		t.Fatal(e)
	}
	if s.removeTarget != "" || len(s.Servers) != 1 {
		t.Fatal("remove cancellation changed configuration")
	}
	if e := tt.Click("Search MCP servers"); e != nil {
		t.Fatal(e)
	}
	tt.Type("absent")
	if !tt.HasText("No matching servers") {
		t.Fatal("filter did not affect actual list")
	}
}
func TestNativeHTTPFieldsAndPasswordMasking(t *testing.T) {
	s := stateFixture(t)
	s.BeginAdd()
	s.Editor.Transport = "http"
	s.Editor.OAuthOpen = true
	tt := testerFor(s, 900, 900)
	if !tt.HasText("Server URL") || !tt.HasText("Headers") || tt.HasText("Command") {
		t.Fatal("typed HTTP fields not mounted")
	}
	s.Editor.OAuthClientSecret = "never-render-this-secret"
	tt.Frame()
	if strings.Contains(strings.Join(tt.Texts(), "\n"), "never-render-this-secret") {
		t.Fatal("OAuth secret was rendered in text snapshot")
	}
}
func TestSaveLiteralRequiresExplicitConfirmationAndUsesReference(t *testing.T) {
	store := preferences.New(t.TempDir())
	binary := fakePi(t, `printf '%s\n' '{"servers":[{"name":"configured","state":"connected","tools":["read"]}]}'`)
	var invalidations atomic.Int32
	modified := 0
	s := New(store, "", Options{PiBinary: binary, OnChange: func() { invalidations.Add(1) }, OnModified: func() { modified++ }})
	defer s.Close()
	s.BeginAdd()
	s.Editor.Name = "configured"
	s.Editor.Command = "node"
	s.Editor.Env = "TOKEN=literal-value"
	s.Editor.Description = "Description outside Pi"
	s.saveEditor()
	if !s.confirm.open || s.Busy {
		t.Fatal("secret did not wait for explicit confirmation")
	}
	if _, e := os.Stat(filepath.Join(store.RootDir, "mcp-secrets.json")); !os.IsNotExist(e) {
		t.Fatal("secret saved before confirmation")
	}
	s.confirm.accept()
	s.confirm = confirmation{}
	waitState(t, s)
	if s.Error != "" {
		t.Fatal(s.Error)
	}
	if modified != 1 || invalidations.Load() < 2 || s.Editor != nil || len(s.Servers) != 1 || s.Runtime["configured"].State != "connected" {
		t.Fatal("save/probe state incomplete", modified, invalidations.Load(), s.Error)
	}
	config, e := os.ReadFile(filepath.Join(store.PiDir, "mcp.json"))
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(config), "literal-value") || !strings.Contains(string(config), "${MCP_GLOBAL_CONFIGURED_TOKEN}") || strings.Contains(string(config), "Description outside Pi") {
		t.Fatal("config contains a secret or Orbit metadata")
	}
	var secret map[string]string
	if e = store.Read("mcp-secrets.json", &secret); e != nil || secret["MCP_GLOBAL_CONFIGURED_TOKEN"] != "literal-value" {
		t.Fatal("secret did not round trip", e)
	}
}
func TestWorkspaceGenerationRejectsLateResults(t *testing.T) {
	s := New(preferences.New(t.TempDir()), "/old", Options{})
	defer s.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	s.start("old job", func(ctx context.Context) (func(*State), error) {
		close(entered)
		<-release
		return func(s *State) { s.Status = "stale workspace result" }, nil
	})
	<-entered
	s.SetWorkspace("/new")
	close(release)
	deadline := time.Now().Add(time.Second)
	for len(s.results) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	s.Tick()
	if s.Workspace != "/new" || s.Status != "" || s.Busy {
		t.Fatal("stale result crossed workspace generation")
	}
}
func TestCloseRejectsResultsAndRemovesSecretDraft(t *testing.T) {
	s := New(preferences.New(t.TempDir()), "", Options{})
	s.secret = secretPrompt{open: true, value: "secret"}
	generation := s.generation
	s.Close()
	s.results <- completion{generation: generation, apply: func(s *State) { s.Status = "late" }}
	s.Tick()
	if s.Status != "" || s.secret.value != "" || s.HasModal() {
		t.Fatal("closed state retained secret or accepted result")
	}
}
func TestNativeCaptureMCP(t *testing.T) {
	out := os.Getenv("ORBIT_MCP_CAPTURE_DIR")
	if out == "" {
		t.Skip("set ORBIT_MCP_CAPTURE_DIR for explicit native CPU fixture screenshots")
	}
	if e := os.MkdirAll(out, 0755); e != nil {
		t.Fatal(e)
	}
	s := stateFixture(t)
	s.Selected = "test-server"
	tt := testerFor(s, 960, 720)
	save := func(name string) {
		t.Helper()
		tt.Frame()
		f, e := os.Create(filepath.Join(out, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		defer f.Close()
		if e = png.Encode(f, tt.Image()); e != nil {
			t.Fatal(e)
		}
	}
	save("mcp-native-list")
	s.BeginAdd()
	save("mcp-native-editor")
	s.Editor.Transport = "http"
	s.Editor.OAuthOpen = true
	save("mcp-native-http-editor")
}

func TestSecretConfirmationEscapeKeepsEditorAndDoesNotSave(t *testing.T) {
	s := stateFixture(t)
	s.BeginAdd()
	s.Editor.Name = "new-server"
	s.Editor.Command = "node"
	s.Editor.Env = "TOKEN=test-secret"
	s.saveEditor()
	if !s.confirm.open {
		t.Fatal("confirmation absent")
	}
	tt := testerFor(s, 900, 720)
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if s.confirm.open || s.Editor == nil || !s.HasModal() {
		t.Fatal("Escape did not dismiss only the top confirmation")
	}
	if _, e := os.Stat(filepath.Join(s.service.Preferences.RootDir, "mcp-secrets.json")); !os.IsNotExist(e) {
		t.Fatal("Escape stored secret")
	}
	if !s.Escape() || s.HasModal() {
		t.Fatal("host Escape did not dismiss editor")
	}
}
func TestApplyRequestIsNotReportedAsCompleted(t *testing.T) {
	s := stateFixture(t)
	requests := 0
	s.Options.OnApply = func() { requests++ }
	s.applyPending = true
	s.apply()
	if requests != 1 || !s.applyPending || strings.Contains(s.Status, "applied") {
		t.Fatal("restart request treated as completion")
	}
	s.Applied()
	if s.applyPending || !s.pendingProbe {
		t.Fatal("actual apply acknowledgement did not settle state")
	}
}
func TestPiOverrideInvalidatesOldProbeGeneration(t *testing.T) {
	s := stateFixture(t)
	oldService := s.service
	oldGeneration := s.generation
	s.SetPiBinary("/new/pi")
	if s.service == oldService || s.service.PiBinary != "/new/pi" || s.generation == oldGeneration || s.loaded {
		t.Fatal("override reused mutable service or stale generation")
	}
	s.results <- completion{generation: oldGeneration, apply: func(s *State) { s.Error = "stale probe" }}
	s.Tick()
	if s.Error != "" {
		t.Fatal("old binary probe applied")
	}
}
func TestNativeSecretPromptRequiresConfirmationAndMasksValue(t *testing.T) {
	s := stateFixture(t)
	s.openSecret(s.Servers[0])
	s.secret.value = "do-not-display-secret"
	tt := testerFor(s, 800, 560)
	if strings.Contains(strings.Join(tt.Texts(), "\n"), "do-not-display-secret") {
		t.Fatal("secret prompt leaked plaintext")
	}
	if e := tt.Click("Store secret"); e != nil {
		t.Fatal(e)
	}
	if !s.confirm.open || s.Busy {
		t.Fatal("prompt bypassed confirmation")
	}
	if !s.Escape() || s.confirm.open || !s.secret.open {
		t.Fatal("confirmation dismissal destroyed prompt")
	}
	s.Escape()
	if s.secret.value != "" || s.HasModal() {
		t.Fatal("secret draft survived cancellation")
	}
}

func TestSlowProbeDoesNotBlockEditingOrLeakItsResult(t *testing.T) {
	s := stateFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	s.start("probe", func(ctx context.Context) (func(*State), error) {
		close(started)
		<-release
		return func(s *State) { s.Status = "obsolete probe" }, nil
	})
	s.Probing = true
	<-started
	tt := testerFor(s, 900, 720)
	if e := tt.Click("Add server"); e != nil {
		t.Fatal(e)
	}
	if s.Editor == nil || s.Busy || s.Probing {
		t.Fatal("read-only probe blocked native editor")
	}
	close(release)
	deadline := time.Now().Add(time.Second)
	for len(s.results) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	s.Tick()
	if s.Status == "obsolete probe" {
		t.Fatal("interrupted probe result applied")
	}
}
func TestNativeFieldLabelsFocusTheirInputs(t *testing.T) {
	s := stateFixture(t)
	s.BeginAdd()
	tt := testerFor(s, 900, 720)
	if e := tt.Click("Command"); e != nil {
		t.Fatal(e)
	}
	tt.Type("uvx")
	if s.Editor.Command != "uvx" || s.Editor.Name != "" {
		t.Fatal("clicking the field label did not focus its native input")
	}
}

func TestMCPNativeKeyboardNavigationEditAndToggle(t *testing.T) {
	s := stateFixture(t)
	s.service.PiBinary = fakePi(t, `printf '%s\n' '{"servers":[]}'`)
	second := workbench.MCPServer{Name: "z-server", Scope: "global", Definition: map[string]any{"command": "node"}}
	s.Servers = append(s.Servers, second)
	if e := s.service.SaveMCP("global", "", second.Name, second.Definition); e != nil {
		t.Fatal(e)
	}
	tt := testerFor(s, 900, 720)
	if e := tt.Click("Server details: test-server"); e != nil {
		t.Fatal(e)
	}
	tt.Key(0, ui.KeyDown)
	if s.Selected != "z-server" {
		t.Fatal("Down did not select the next server")
	}
	tt.Key(0, ui.KeyEnter)
	if s.Editor == nil || s.Editor.Name != "z-server" {
		t.Fatal("Enter did not open selected server editor")
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if e := tt.Click("Server details: z-server"); e != nil {
		t.Fatal(e)
	}
	// Clicking an already selected row collapses it; choose it again.
	if s.Selected == "" {
		if e := tt.Click("Server details: z-server"); e != nil {
			t.Fatal(e)
		}
	}
	tt.Key(0, ui.KeySpace)
	waitState(t, s)
	server := s.server("z-server")
	if server == nil || enabled(server.Definition) {
		t.Fatal("Space did not persist disabled state")
	}
}
func TestMCPOAuthCLIUsesExactPiCommands(t *testing.T) {
	s := New(preferences.New(t.TempDir()), "", Options{})
	defer s.Close()
	s.service.PiBinary = fakePi(t, `printf '%s\n' "$*" >> "$HOME/auth-args"
if [ "$2" = "list" ]; then printf '%s\n' '{"servers":[]}'; fi`)
	s.auth("remote", true)
	waitState(t, s)
	if s.Error != "" {
		t.Fatal(s.Error)
	}
	s.auth("remote", false)
	waitState(t, s)
	if s.Error != "" {
		t.Fatal(s.Error)
	}
	args, e := os.ReadFile(filepath.Join(s.service.Preferences.Home, "auth-args"))
	if e != nil {
		t.Fatal(e)
	}
	text := string(args)
	if !strings.Contains(text, "mcp login remote --timeout 300\n") || !strings.Contains(text, "mcp logout remote\n") || strings.Contains(text, "--approve") {
		t.Fatal("incorrect OAuth/probe command arguments", text)
	}
}

func TestApplyAcknowledgmentRejectsNewEditWorkspaceAndBinary(t *testing.T) {
	s := stateFixture(t)
	s.modified()
	token := s.ApplyToken()
	s.modified()
	if s.AppliedIfCurrent(token) || !s.applyPending {
		t.Fatal("old reload consumed a newer edit")
	}
	current := s.ApplyToken()
	if !s.AppliedIfCurrent(current) || s.applyPending {
		t.Fatal("matching reload not acknowledged")
	}
	token = s.ApplyToken()
	s.SetWorkspace(t.TempDir())
	s.SetWorkspace(token.workspace)
	if s.AppliedIfCurrent(token) {
		t.Fatal("workspace round trip reused old apply token")
	}
	token = s.ApplyToken()
	s.SetPiBinary("synthetic-other-pi")
	if s.AppliedIfCurrent(token) {
		t.Fatal("old binary reload acknowledged")
	}
	token = s.ApplyToken()
	s.Close()
	if s.AppliedIfCurrent(token) {
		t.Fatal("closed page accepted completion")
	}
}

func TestRefreshAndApplyCapturesFreshConfigBeforeHostRequest(t *testing.T) {
	s := stateFixture(t)
	s.fingerprint = "old-page-snapshot"
	called := 0
	s.Options.OnApply = func() {
		called++
		if s.fingerprint == "old-page-snapshot" {
			t.Error("host scheduled against stale config")
		}
	}
	s.RefreshAndApply()
	if called != 0 {
		t.Fatal("refresh blocked the UI thread or called apply before worker completion")
	}
	waitState(t, s)
	if called != 1 {
		t.Fatal("fresh config did not reach host exactly once", called)
	}
}

func TestRefreshAndApplyInterruptsProbeAndReportsOtherBusy(t *testing.T) {
	s := stateFixture(t)
	called := 0
	s.Options.OnApply = func() { called++ }
	// Hold an actual asynchronous read-only worker until cancellation.
	s.start("probe fixture", func(ctx context.Context) (func(*State), error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})
	s.Probing = true
	s.RefreshAndApply()
	waitState(t, s)
	if called != 1 || s.Error != "" {
		t.Fatal("probe swallowed reconnect", called, s.Error)
	}
	release := make(chan struct{})
	s.start("configuration write fixture", func(ctx context.Context) (func(*State), error) {
		<-release
		return nil, nil
	})
	s.RefreshAndApply()
	if called != 1 || !s.Busy || !strings.Contains(s.Error, "Retry reconnect") {
		t.Fatal("busy write not preserved/reported", called, s.Error)
	}
	close(release)
	waitState(t, s)
	s.RefreshAndApply()
	waitState(t, s)
	if called != 2 {
		t.Fatal("retry after write did not apply", called)
	}
}
