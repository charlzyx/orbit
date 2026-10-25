package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"strings"
	"testing"
	"time"
)

func TestNativeRuntimeUsesRealSnapshotAndStoppedUnknowns(t *testing.T) {
	a := fixture(t)
	a.page = "settings"
	a.settingsPage = "runtime"
	a.nativeSettings.runtimeLoaded = true
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	a.Runtime = agent.RuntimeSnapshot{Transport: "stdio JSONL", ConfiguredExecutable: "/fixture/pi", Active: &agent.RuntimeProcess{PID: 4321, Executable: "/fixture/node", StartedAt: now, Uptime: 2*time.Minute + 3*time.Second, Alive: true, State: "idle", RecentStderr: []string{"explicit fixture diagnostic"}}, Parked: []agent.RuntimeProcess{{PID: 6789, SessionName: "Fixture background", State: "busy", Alive: true}}}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(a.View, 1200, 1000)
	tt.Frame()
	text := strings.Join(tt.Texts(), "\n")
	for _, s := range []string{"4321", "/fixture/node", "2m3s", "Fixture background", "PID 6789 · busy", "explicit fixture diagnostic"} {
		if !strings.Contains(text, s) {
			t.Fatal("missing actual runtime field", s)
		}
	}
	if e := tt.Click("Stop"); e != nil {
		t.Fatal(e)
	}
	if len(actions) != 1 || actions[0].Kind != "StopRuntime" {
		t.Fatal(actions)
	}
	a.Runtime.Active = &agent.RuntimeProcess{State: "stopped", PID: 4321, StartedAt: now, Uptime: 2 * time.Hour}
	tt.Frame()
	if runtimeUptime(a.Runtime.Active) != "—" {
		t.Fatal("stopped process displayed growing uptime")
	}
	if e := tt.Click("Start"); e != nil {
		t.Fatal(e)
	}
	if actions[len(actions)-1].Kind != "StartRuntime" {
		t.Fatal(actions)
	}
}
func TestNativeAgentRetryIsUnknownUntilConfirmed(t *testing.T) {
	a := fixture(t)
	a.Model.Connected = true
	a.Runtime.Active = &agent.RuntimeProcess{AutoRetry: agent.RetrySetting{Support: "unknown"}}
	var actions []Action
	a.Options.OnAction = func(x Action) { actions = append(actions, x) }
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Fill().Children(func() { a.agentRuntimeSettings(c) }) }, 1000, 700)
	if !tt.HasText("Unknown: this pi version has no getter") {
		t.Fatal("unknown retry shown as false/true")
	}
	if e := tt.Click("Set retry policy"); e != nil {
		t.Fatal(e)
	}
	if e := tt.ChooseMenuItem("Enable"); e != nil {
		t.Fatal(e)
	}
	if len(actions) != 1 || actions[0].Kind != "SetAutoRetry" || actions[0].Data["enabled"] != true {
		t.Fatal(actions)
	}
	if a.Runtime.Active.AutoRetry.Known {
		t.Fatal("optimistic ACK state invented")
	}
	a.Runtime.Active.AutoRetry = agent.RetrySetting{Support: "supported", Known: true, Enabled: true}
	tt.Frame()
	if !tt.HasText("Enabled") {
		t.Fatal("confirmed retry absent")
	}
}
