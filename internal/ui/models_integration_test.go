package orbitui

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
)

func TestModelsMountKeepsAuthoritativeCatalogAndSettingsControls(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{Connected: true, ModelID: "same", ModelProvider: "second", ModelCatalog: []agent.ModelChoice{{ID: "same", Provider: "first", Name: "First Model", ContextWindow: 64000}, {ID: "same", Provider: "second", Name: "Second Model", ContextWindow: 200000, Reasoning: true}}, CurrentModel: agent.ModelChoice{ID: "same", Provider: "second", ContextWindow: 200000}, ThinkingLevels: []string{"off", "low", "high"}})
	a.Navigate("settings")
	a.settingsPage = "models"
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if len(a.Models.Snapshot.Models) != 2 || a.Models.Snapshot.Current.ContextWindow != 200000 {
		t.Fatal("catalog capability data lost")
	}
	for _, label := range []string{"Search model catalog", "Select model first / same", "Select model second / same"} {
		if _, ok := tt.Find(label); !ok {
			t.Fatal("source model surface missing", label)
		}
	}
	tt.Click("Search model catalog")
	tt.Type("no result")
	tt.Frame()
	if len(a.Models.Groups()) != 0 {
		t.Fatal("host did not mount model filter")
	}
}
func TestModelRefreshHostSerializesAndShowsFailureWithoutRestart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture; Windows cross-build is separate")
	}
	x := maintenanceFixture(t)
	x.workspace = t.TempDir()
	bin := filepath.Join(t.TempDir(), "configured pi")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$HOME/catalog-args\"\nprintf '%s\\n' \"$PI_CODING_AGENT_DIR\" > \"$HOME/catalog-home\"\nsleep 0.05\nprintf 'catalog fixture failure\\n' >&2\nexit 4\n"
	if e := os.WriteFile(bin, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	if e := x.Agent.SetPiBinary(bin); e != nil {
		t.Fatal(e)
	}
	x.UI.Options.RefreshModels = x.refreshModelCatalog
	if e := x.refreshModelCatalog(); e != nil {
		t.Fatal(e)
	}
	if !x.maintenance.Load() || !x.UI.Models.Refreshing {
		t.Fatal("host did not expose pending refresh")
	}
	if e := x.refreshModelCatalog(); e == nil {
		t.Fatal("duplicate catalog update allowed")
	}
	x.UI.Draft = "retain me"
	x.Action(Action{Kind: "Submit", Value: x.UI.Draft})
	if x.UI.Draft != "retain me" || x.Agent.AnyBusy() {
		t.Fatal("maintenance admitted a prompt")
	}
	drainRegressionUpdates(t, x.UI, func() bool { return !x.catalogRefreshing })
	if x.maintenance.Load() || x.UI.Models.Refreshing || !strings.Contains(x.UI.Models.RefreshError, "catalog fixture failure") {
		t.Fatal("catalog failure not visible/released", x.UI.Models.RefreshError)
	}
	data, e := os.ReadFile(filepath.Join(x.Store.Home, "catalog-args"))
	if e != nil || strings.TrimSpace(string(data)) != "update --models --no-approve" {
		t.Fatal("wrong configured command", string(data), e)
	}
	data, e = os.ReadFile(filepath.Join(x.Store.Home, "catalog-home"))
	if e != nil || strings.TrimSpace(string(data)) != x.Store.PiDir {
		t.Fatal("wrong isolated Pi home", string(data), e)
	}
	if x.Agent.Snapshot().ActiveSession.Workspace != "" {
		t.Fatal("failed catalog refresh started a process")
	}
	// Cancellation is passed to the real child command, and the gate must release.
	ctx, cancel := context.WithCancel(context.Background())
	x.background = ctx
	if e := x.refreshModelCatalog(); e != nil {
		t.Fatal(e)
	}
	cancel()
	drainRegressionUpdates(t, x.UI, func() bool { return !x.catalogRefreshing })
	if x.maintenance.Load() {
		t.Fatal("cancelled catalog update leaked maintenance lock")
	}
}

// A disposable JSONL peer exercises host update -> process restart -> explicit
// saved-session resume -> authoritative catalog query without any model calls.
func TestModelCatalogHostRPCProcess(t *testing.T) {
	if os.Getenv("ORBIT_CATALOG_HOST_HELPER") != "1" {
		return
	}
	log, err := os.OpenFile(filepath.Join(os.Getenv("HOME"), "rpc-catalog-log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		os.Exit(3)
	}
	defer log.Close()
	fmt.Fprintf(log, "start %d\n", os.Getpid())
	path := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		var command map[string]any
		if json.Unmarshal(scanner.Bytes(), &command) != nil {
			continue
		}
		kind, _ := command["type"].(string)
		fmt.Fprintln(log, kind)
		response := map[string]any{"id": command["id"], "type": "response", "command": kind, "success": true}
		switch kind {
		case "switch_session":
			path, _ = command["sessionPath"].(string)
			response["data"] = map[string]any{"cancelled": false}
		case "get_state":
			response["data"] = map[string]any{"sessionId": "saved-fixture", "sessionFile": path, "model": map[string]any{"id": "fixture", "provider": "fixture"}, "thinkingLevel": "off"}
		case "get_available_models":
			id := "before-update"
			if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), "catalog-updated")); err == nil {
				id = "after-update"
			}
			response["data"] = map[string]any{"models": []any{map[string]any{"id": id, "provider": "fixture", "contextWindow": 64000}}}
		case "get_messages":
			response["data"] = map[string]any{"messages": []any{}}
		case "get_available_thinking_levels":
			response["data"] = map[string]any{"levels": []string{"off"}}
		case "get_commands":
			response["data"] = map[string]any{"commands": []any{}}
		case "get_entries":
			response["data"] = map[string]any{"entries": []any{}}
		case "get_session_stats":
			response["data"] = map[string]any{"tokens": map[string]any{"total": 0}, "cost": 0}
		case "auth.list", "quota.list":
			response["success"] = false
			response["error"] = "Unknown command"
		case "prompt":
			os.Exit(9)
		}
		raw, _ := json.Marshal(response)
		fmt.Fprintln(os.Stdout, string(raw))
	}
	os.Exit(0)
}
func TestModelRefreshHostSuccessRestartsAndResumesSavedSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture; Windows cross-build is separate")
	}
	x := maintenanceFixture(t)
	_ = x.Agent.Close()
	x.workspace = t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
	bin := filepath.Join(t.TempDir(), "configured pi")
	script := "#!/bin/sh\nif [ \"$1\" = update ]; then printf updated > \"$HOME/catalog-updated\"; exit 0; fi\nexport ORBIT_CATALOG_HOST_HELPER=1\nexec " + quote(exe) + " -test.run=^TestModelCatalogHostRPCProcess$ -- \"$@\"\n"
	if err = os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: bin, Home: x.Store.Home, SessionDir: filepath.Join(x.Store.PiDir, "sessions"), Environment: []string{}, Offline: true, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true})
	t.Cleanup(func() { x.Agent.Close() })
	path := filepath.Join(x.Store.Home, "saved.jsonl")
	if err = os.WriteFile(path, []byte(`{"type":"session","version":3,"id":"saved-fixture","cwd":"`+x.workspace+`"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = x.Agent.Open(x.workspace, path); err != nil {
		t.Fatal(err)
	}
	wait := func(condition func(agent.Snapshot) bool) agent.Snapshot {
		t.Helper()
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			x.Agent.Tick()
			s := x.Agent.Snapshot()
			if condition(s) {
				return s
			}
			time.Sleep(time.Millisecond)
		}
		t.Fatal("catalog host process did not settle", x.Agent.Snapshot())
		return agent.Snapshot{}
	}
	wait(func(s agent.Snapshot) bool {
		return s.Connected && len(s.Models) == 1 && s.Models[0].ID == "before-update"
	})
	before := x.Agent.RuntimeSnapshot().Active.PID
	if err = x.refreshModelCatalog(); err != nil {
		t.Fatal(err)
	}
	drainRegressionUpdates(t, x.UI, func() bool { return !x.catalogRefreshing })
	after := wait(func(s agent.Snapshot) bool {
		return s.Connected && len(s.Models) == 1 && s.Models[0].ID == "after-update"
	})
	if after.ActiveSession.Path != path || after.ActiveSession.ID != "saved-fixture" || x.Agent.RuntimeSnapshot().Active.PID == before {
		t.Fatal("successful catalog update lost saved session or did not restart", after.ActiveSession)
	}
	if x.maintenance.Load() || x.UI.Models.RefreshError != "" {
		t.Fatal("successful catalog update left gate/error")
	}
	data, err := os.ReadFile(filepath.Join(x.Store.Home, "rpc-catalog-log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(data), "switch_session") != 2 || strings.Count(string(data), "get_available_models") != 2 || strings.Contains(string(data), "prompt") {
		t.Fatal("incorrect restart/resume/query sequence", string(data))
	}
}

func TestModelPickerOpenQueriesOnceAndUsesSourceAnchor(t *testing.T) {
	a := fixture(t)
	a.SetSnapshot(Snapshot{ModelID: "long-fixture-model", Models: []ModelChoice{{ID: "long-fixture-model", Name: "Fixture", Provider: "fixture"}}, ThinkingLevels: []string{"off", "high"}})
	queries := 0
	a.Options.QueryModelCatalog = func() { queries++ }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	anchor, _ := tt.Find("long-fixture-model")
	if e := tt.Click("long-fixture-model"); e != nil {
		t.Fatal(e)
	}
	tt.Frame()
	popup, ok := tt.Find("Composer model picker")
	if !ok || math.Abs(float64(popup.X-anchor.X)) > 1 || popup.Y+popup.H > anchor.Y {
		t.Fatal("picker not above left of chip", anchor, popup)
	}
	for i := 0; i < 4; i++ {
		tt.Frame()
	}
	if queries != 1 {
		t.Fatal("catalog refreshed every frame", queries)
	}
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if !tt.Focused("Message") {
		t.Fatal("picker Escape did not restore composer")
	}
	a.modelOpen = true
	tt.Frame()
	if queries != 2 {
		t.Fatal("keyboard picker did not refresh catalog once", queries)
	}
}

func TestNativeModelAndWorkspaceHostCaptures(t *testing.T) {
	out := os.Getenv("ORBIT_CAPTURE_DIR")
	if out == "" {
		t.Skip("explicit native CPU fixtures only")
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		t.Fatal(err)
	}
	a := fixture(t)
	a.SetSnapshot(Snapshot{Connected: true, ModelID: "fixture-model", ModelProvider: "fixture", Workspace: "/workspace/example", ModelCatalog: []agent.ModelChoice{{ID: "fixture-model", Name: "Fixture reasoning model", Provider: "fixture", ContextWindow: 200000, Reasoning: true}, {ID: "second-model", Name: "Fixture fast model", Provider: "second", ContextWindow: 128000}}, CurrentModel: agent.ModelChoice{ID: "fixture-model", Name: "Fixture reasoning model", Provider: "fixture", ContextWindow: 200000, Reasoning: true}, ThinkingLevels: []string{"off", "low", "medium", "high"}, Thinking: "medium", Sessions: []SessionRow{{Workspace: "/workspace/second", Updated: time.Unix(1700000000, 0)}}})
	tt := ui.NewTester(a.View, 1470, 922)
	tt.SetDark(true)
	tt.Frame()
	save := func(name string) {
		t.Helper()
		tt.Frame()
		f, e := os.Create(filepath.Join(out, name+".png"))
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, tt.Image())
		_ = f.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
	tt.Click("Choose workspace")
	save("workspace-picker-fixture-dark")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	tt.Click("fixture-model")
	save("model-picker-fixture-dark")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	a.Navigate("settings")
	a.settingsPage = "models"
	save("models-host-fixture-dark")
}

func TestAutomaticTitlePreferencesDoNotRestartRuntime(t *testing.T) {
	x := maintenanceFixture(t)
	x.Action(Action{Kind: "AutoTitleChanged"})
	if !x.configApplyAfter.IsZero() {
		t.Fatal("live-read automatic title preferences scheduled a process restart")
	}
}
