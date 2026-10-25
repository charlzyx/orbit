package orbitui

import (
	"github.com/charlzyx/orbit/internal/mcpui"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestQuotaPopoverHasActualMCPPane(t *testing.T) {
	a := fixture(t)
	state := mcpui.New(a.Options.Store, t.TempDir(), mcpui.Options{})
	defer state.Close()
	if host, ok := any(a).(interface{ SetQuotaMCP(*mcpui.State) }); ok {
		host.SetQuotaMCP(state)
	}
	open := true
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, quotaFixtureReports(), "supported", "alpha") }, 800, 700)
	if _, ok := tt.Find("MCP"); !ok {
		t.Fatal("quota popover lacks actual MCP pane")
	}
}

func quotaMCPFixture(t *testing.T, a *App, configured bool) *mcpui.State {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("synthetic shell probe; cross-build separate")
	}
	workspace := t.TempDir()
	bin := filepath.Join(t.TempDir(), "synthetic pi")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\nprintf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"connected\",\"tools\":[\"read_fixture\"]}]}'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if configured {
		if err := workbench.New(a.Options.Store, workspace).SaveMCP("global", "", "fixture", map[string]any{"command": "never-executed-fixture"}); err != nil {
			t.Fatal(err)
		}
	}
	s := mcpui.New(a.Options.Store, workspace, mcpui.Options{PiBinary: bin, Text: a.text})
	t.Cleanup(s.Close)
	a.SetQuotaMCP(s)
	a.Options.RefreshMCP = s.Refresh
	s.Refresh()
	waitQuotaMCP(t, s)
	return s
}
func waitQuotaMCP(t *testing.T, s *mcpui.State) {
	t.Helper()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		s.Tick()
		if !s.Busy && !s.Probing {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("fixture MCP state did not settle", s.Error)
}
func TestQuotaMCPOnlyPillAndSelectedServerSettingsRoute(t *testing.T) {
	a := fixture(t)
	s := quotaMCPFixture(t, a, true)
	// Configuration-only binding creates the MCP entry even without quota RPC.
	if len(s.Servers) != 1 {
		t.Fatal("configured MCP-only fixture unavailable", s.Error)
	}
	open := false
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, nil, "unsupported", "") }, 800, 700)
	chip, show := s.QuotaChip()
	if !show {
		t.Fatal("MCP-only pill hidden")
	}
	if err := tt.Click(chip); err != nil {
		t.Fatal(err)
	}
	if !open || a.quotaMCP.tab != "mcp" {
		t.Fatal("quiet MCP-only entry did not select MCP pane")
	}
	waitQuotaMCP(t, s)
	// Repaint after consuming the real synthetic probe result.
	tt.Advance(time.Millisecond)
	if err := tt.Click("MCP server fixture"); err != nil {
		t.Fatal(err)
	}
	if open || a.page != "settings" || a.settingsPage != "mcp" || s.Selected != "fixture" {
		t.Fatal("settings jump lost selected server", open, a.page, a.settingsPage, s.Selected)
	}
	waitQuotaMCP(t, s)
	if s.Selected != "fixture" {
		t.Fatal("async settings refresh lost existing selected server")
	}
}
func TestQuotaProviderAndMCPPaneSwitchingAndEmptySettingsAction(t *testing.T) {
	a := fixture(t)
	s := quotaMCPFixture(t, a, false)
	open := true
	tt := ui.NewTester(func(c *ui.Context) { a.runtimeQuotaControl(c, &open, quotaFixtureReports(), "supported", "alpha") }, 800, 700)
	if err := tt.Click(a.text("session.tab_mcp", "MCP")); err != nil {
		t.Fatal(err)
	}
	if a.quotaMCP.tab != "mcp" {
		t.Fatal("tab switch did not retain pane state")
	}
	waitQuotaMCP(t, s)
	tt.Advance(time.Millisecond)
	if !tt.HasText(a.text("mcp.empty_title", "No MCP servers")) {
		t.Fatal("actual empty configuration state absent")
	}
	if err := tt.Click(a.text("session.tab_providers", "Providers")); err != nil {
		t.Fatal(err)
	}
	if _, ok := tt.Find(a.text("session.hide_provider", "Hide provider") + " beta"); !ok {
		t.Fatal("provider tab lost functional hidden controls")
	}
	if err := tt.Click(a.text("session.tab_mcp", "MCP")); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click(a.text("mcp.popup_open_settings", "Open MCP settings")); err != nil {
		t.Fatal(err)
	}
	if open || a.page != "settings" || a.settingsPage != "mcp" {
		t.Fatal("empty-state manage action not wired")
	}
}
