package mcpui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func popupState(t *testing.T) (*State, string) {
	t.Helper()
	store := preferences.New(t.TempDir())
	log := filepath.Join(t.TempDir(), "probe-args")
	bin := fakePi(t, "printf '%s\\n' \"$*\" >> '"+log+"'\nprintf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"connected\",\"tools\":[\"read_fixture\"]}]}'")
	s := New(store, t.TempDir(), Options{PiBinary: bin})
	t.Cleanup(s.Close)
	if err := s.service.SaveMCP("global", "", "fixture", map[string]any{"command": "never-run-fixture"}); err != nil {
		t.Fatal(err)
	}
	return s, log
}
func TestQuotaPopupPrepareDoesNotProbeAndTwentySecondSelectionRefreshes(t *testing.T) {
	s, log := popupState(t)
	s.PrepareQuotaPopup()
	waitState(t, s)
	if len(s.Servers) != 1 || !s.checkedAt.IsZero() {
		t.Fatal("config preload fabricated probe")
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("config-only preload launched Pi")
	}
	if text, show := s.QuotaChip(); !show || !strings.Contains(text, "Not checked") {
		t.Fatal("unprobed MCP-only entry is not honest", text)
	}
	s.OpenQuotaPopup(time.Now())
	waitState(t, s)
	if s.Runtime["fixture"].State != "connected" || len(s.Runtime["fixture"].Tools) != 1 {
		t.Fatal("actual fixture probe not used")
	}
	checked := s.checkedAt
	s.OpenQuotaPopup(checked.Add(19 * time.Second))
	if s.Busy {
		t.Fatal("fresh20s cache re-probed")
	}
	s.OpenQuotaPopup(checked.Add(20 * time.Second))
	if !s.Busy {
		t.Fatal("stale20s boundary ignored")
	}
	waitState(t, s)
	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(raw), "mcp list --json") != 2 {
		t.Fatal("wrong probe command/count", string(raw))
	}
}
func TestQuotaPopupSelectionQueuedBehindInitialConfigRead(t *testing.T) {
	s, log := popupState(t)
	release := make(chan struct{})
	s.start("held config fixture", func(ctx context.Context) (func(*State), error) {
		<-release
		v := loadSnapshot(s.service)
		return func(s *State) { s.adopt(v) }, nil
	})
	s.OpenQuotaPopup(time.Now())
	if !s.quotaProbeRequested {
		t.Fatal("selection lost while config read busy")
	}
	close(release)
	waitState(t, s)
	if _, err := os.Stat(log); err != nil || s.Runtime["fixture"].State != "connected" {
		t.Fatal("queued probe never reached actual runtime", err)
	}
}
func TestQuotaPopupCountsConfiguredDisabledAndUnknownWithoutInventingState(t *testing.T) {
	s := stateFixture(t)
	s.Servers = []workbench.MCPServer{{Name: "live", Scope: "global", Definition: map[string]any{"command": "fixture"}}, {Name: "disabled", Scope: "project", Definition: map[string]any{"enabled": false}}, {Name: "unprobed", Scope: "global", Definition: map[string]any{"command": "fixture"}}}
	s.Runtime = map[string]Runtime{"live": {State: "connected", Tools: []Tool{{Name: "one"}}}, "disabled": {State: "connected"}}
	if got := s.QuotaSummary(); !strings.Contains(got, "1/3 connected") || !strings.Contains(got, "Not checked") {
		t.Fatal("wrong configured count or lost unknown", got)
	}
	s.Runtime = map[string]Runtime{}
	if got := s.QuotaSummary(); strings.Contains(got, "0/") || !strings.Contains(got, "Not checked") {
		t.Fatal("unknown manufactured disconnected zero", got)
	}
	s.Servers = s.Servers[1:2]
	if _, show := s.QuotaChip(); show {
		t.Fatal("disabled-only config invented enabled MCP pill")
	}
}
func TestQuotaPopupMountedStateActionsAndSelectedServer(t *testing.T) {
	s := stateFixture(t)
	s.Servers = []workbench.MCPServer{{Name: "live", Scope: "global", Definition: map[string]any{"command": "fixture"}}, {Name: "disabled", Scope: "project", Definition: map[string]any{"enabled": false}}, {Name: "unknown", Scope: "global", Definition: map[string]any{"command": "fixture"}}}
	s.Runtime = map[string]Runtime{"live": {State: "connected", Tools: []Tool{{Name: "one"}}}}
	selected := ""
	tt := ui.NewTester(func(c *ui.Context) { s.QuotaPanel(c, func(name string) { selected = name }) }, 600, 600)
	if !tt.HasText("1 tools") {
		t.Fatal("probe tool count absent")
	}
	if !tt.HasText("○ Disabled") || !tt.HasText("○ Not checked") {
		t.Fatal("disabled/unknown state lost")
	}
	if err := tt.Click("MCP server live"); err != nil {
		t.Fatal(err)
	}
	if selected != "live" {
		t.Fatal("server detail routing lost target", selected)
	}
}
func TestQuotaPopupSignInAndCancelUseExistingPiFlow(t *testing.T) {
	for _, cancelLogin := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "cancel"}[cancelLogin], func(t *testing.T) {
			s := stateFixture(t)
			marker := filepath.Join(t.TempDir(), "login-command")
			login := "exit 0"
			if cancelLogin {
				login = "exec sleep 30"
			}
			s.service.PiBinary = fakePi(t, "if [ \"$2\" = login ]; then printf '%s\\n' \"$*\" > '"+marker+"'; "+login+"; fi\nprintf '%s\\n' '{\"servers\":[{\"name\":\"remote\",\"state\":\"needs-auth\"}]}'")
			s.Servers = []workbench.MCPServer{{Name: "remote", Scope: "global", Definition: map[string]any{"url": "https://fixture.invalid"}}}
			s.Runtime = map[string]Runtime{"remote": {State: "needs-auth"}}
			modified := 0
			s.Options.OnModified = func() { modified++ }
			tt := ui.NewTester(func(c *ui.Context) { s.QuotaPanel(c, func(string) {}) }, 600, 400)
			if err := tt.Click("Sign in to MCP remote"); err != nil {
				t.Fatal(err)
			}
			if s.authTarget != "remote" || !s.Busy {
				t.Fatal("mounted sign-in did not invoke real state flow")
			}
			deadline := time.Now().Add(3 * time.Second)
			for time.Now().Before(deadline) {
				if _, err := os.Stat(marker); err == nil {
					break
				}
				time.Sleep(time.Millisecond)
			}
			if raw, err := os.ReadFile(marker); err != nil || strings.TrimSpace(string(raw)) != "mcp login remote --timeout 300" {
				t.Fatal("incorrect synthetic auth command", string(raw), err)
			}
			if cancelLogin {
				if err := tt.Click("Cancel MCP sign-in remote"); err != nil {
					t.Fatal(err)
				}
			}
			waitState(t, s)
			if cancelLogin && modified != 0 {
				t.Fatal("cancelled auth reported applied modification")
			}
			if !cancelLogin && modified != 1 {
				t.Fatal("successful auth not forwarded to existing apply flow", modified)
			}
		})
	}
}

func TestQuotaPopupWorkspaceChangeDropsOldQueuedProbeIntent(t *testing.T) {
	s, log := popupState(t)
	release := make(chan struct{})
	s.start("held old-workspace read", func(ctx context.Context) (func(*State), error) {
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return func(s *State) { s.Status = "stale completion" }, nil
	})
	s.OpenQuotaPopup(time.Now())
	s.SetWorkspace(t.TempDir())
	close(release)
	s.PrepareQuotaPopup()
	waitState(t, s)
	if s.quotaProbeRequested || s.Status == "stale completion" || !s.checkedAt.IsZero() {
		t.Fatal("old workspace completion/request leaked")
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("old selection launched probe in replacement workspace")
	}
}
func TestQuotaPopupProbeFailureRemainsUnknownAndVisible(t *testing.T) {
	s, _ := popupState(t)
	s.PrepareQuotaPopup()
	waitState(t, s)
	s.service.PiBinary = fakePi(t, "printf 'synthetic probe failure\\n' >&2; exit 2")
	s.OpenQuotaPopup(time.Now())
	waitState(t, s)
	if s.Error == "" || !s.checkedAt.IsZero() {
		t.Fatal("failed probe fabricated successful observation")
	}
	if !strings.Contains(s.QuotaSummary(), "Not checked") {
		t.Fatal("failed unprobed state became disconnected count", s.QuotaSummary())
	}
	tt := ui.NewTester(func(c *ui.Context) { s.QuotaPanel(c, func(string) {}) }, 600, 500)
	if !tt.HasText(s.Error) {
		t.Fatal("actual probe failure absent from pane")
	}
}
