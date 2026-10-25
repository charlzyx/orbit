package orbitui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRPCPatchHostSaveIsPendingAndDoesNotLaunchAgent(t *testing.T) {
	x := maintenanceFixture(t)
	x.UI.Model.Extensions.Patches.Enabled = true
	x.Action(Action{Kind: "SetRPCPatches", Data: map[string]any{"enabled": false}})
	if !x.UI.rpcPatchSaving {
		t.Fatal("save did not expose pending state")
	}
	// Duplicate dispatch while the native switch is disabled cannot reorder a save.
	x.Action(Action{Kind: "SetRPCPatches", Data: map[string]any{"enabled": true}})
	drainRegressionUpdates(t, x.UI, func() bool { return !x.UI.rpcPatchSaving })
	if x.Store.RPCPatches().Enabled || x.UI.Model.Extensions.Patches.Enabled || x.Agent.Snapshot().Extensions.Patches.Enabled {
		t.Fatal("policy did not settle to accepted value")
	}
	if x.Agent.RuntimeSnapshot().Active != nil {
		t.Fatal("preference save launched or restarted Pi")
	}
	x.Action(Action{Kind: "SetRPCPatches", Data: map[string]any{"enabled": true}})
	drainRegressionUpdates(t, x.UI, func() bool { return !x.UI.rpcPatchSaving })
	if !x.Store.RPCPatches().Enabled || !x.UI.Model.Extensions.Patches.Enabled {
		t.Fatal("next accepted toggle failed")
	}
}

func TestRPCPatchHostFailureKeepsDisplayedPolicy(t *testing.T) {
	x := maintenanceFixture(t)
	x.UI.Model.Extensions.Patches.Enabled = true
	if err := os.MkdirAll(x.Store.RootDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(x.Store.RootDir, "rpc-patches"), []byte("blocked fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "SetRPCPatches", Data: map[string]any{"enabled": false}})
	drainRegressionUpdates(t, x.UI, func() bool { return !x.UI.rpcPatchSaving })
	if !x.UI.Model.Extensions.Patches.Enabled {
		t.Fatal("failed save changed visible policy")
	}
	if x.Agent.RuntimeSnapshot().Active != nil {
		t.Fatal("save failure touched runtime")
	}
}
