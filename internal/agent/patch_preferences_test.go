package agent

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/charlzyx/orbit/internal/preferences"
)

func TestRPCPatchTogglePersistsWithoutRestartOrRevert(t *testing.T) {
	s := preferences.New(t.TempDir())
	active := &liveSession{}
	parked := &liveSession{}
	active.state.snapshot.Extensions.Patches = PatchStatus{Enabled: true, Applied: []string{"auth-rpc"}, PackageRoot: "/fixture/pi"}
	m := &Manager{workflowStore: s, sessions: map[string]*liveSession{"active": active, "parked": parked}}
	if err := m.SetRPCPatchesEnabled(false); err != nil {
		t.Fatal(err)
	}
	if s.RPCPatches().Enabled || active.state.snapshot.Extensions.Patches.Enabled || parked.state.snapshot.Extensions.Patches.Enabled {
		t.Fatal("toggle not persisted/reflected")
	}
	if len(active.state.snapshot.Extensions.Patches.Applied) != 1 || active.state.snapshot.Extensions.Patches.PackageRoot != "/fixture/pi" {
		t.Fatal("toggle invented patch removal")
	}
	options := Options{Home: s.Home, PiBinary: filepath.Join(s.Home, "not-pi"), Environment: []string{}, DisableExtensions: true}
	_, disabled := prepareExtensions(options)
	if disabled.Patches.Enabled || disabled.Patches.Error != "" {
		t.Fatal("next launch ignored saved opt-out", disabled)
	}
	if err := m.SetRPCPatchesEnabled(true); err != nil {
		t.Fatal(err)
	}
	options.Resources = fstest.MapFS{}
	_, enabled := prepareExtensions(options)
	if !enabled.Patches.Enabled || enabled.Patches.Error == "" {
		t.Fatal("next launch did not attempt enabled configuration", enabled)
	}
}

func TestRPCPatchSaveFailureDoesNotChangeVisibleState(t *testing.T) {
	s := preferences.New(t.TempDir())
	p := filepath.Join(s.RootDir, "rpc-patches")
	if err := os.MkdirAll(s.RootDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	entry := &liveSession{}
	entry.state.snapshot.Extensions.Patches.Enabled = true
	m := &Manager{workflowStore: s, sessions: map[string]*liveSession{"active": entry}}
	if err := m.SetRPCPatchesEnabled(false); err == nil {
		t.Fatal("expected save failure")
	}
	if !entry.state.snapshot.Extensions.Patches.Enabled {
		t.Fatal("failed save changed visible switch")
	}
}

func TestRPCPatchEnvironmentRespectsCapturedAndExplicitOverrides(t *testing.T) {
	t.Setenv("ORBIT_NO_RPC_PATCHES", "1")
	for _, test := range []struct {
		o    Options
		want bool
	}{
		{Options{}, true}, {Options{Environment: []string{}}, false},
		{Options{Environment: []string{"ORBIT_NO_RPC_PATCHES=1"}}, true},
		{Options{Environment: []string{"ORBIT_NO_RPC_PATCHES=1"}, EnvOverrides: map[string]string{"ORBIT_NO_RPC_PATCHES": ""}}, false},
		{Options{Environment: []string{}, EnvOverrides: map[string]string{"ORBIT_NO_RPC_PATCHES": "1"}}, true},
	} {
		if got := patchEnvironmentDisabled(test.o); got != test.want {
			t.Fatalf("override policy got %v want %v", got, test.want)
		}
	}
}

func TestRPCPatchLateLaunchDoesNotUndoSavedPolicy(t *testing.T) {
	m := &Manager{patchPolicyEnabled: false, patchPolicyVersion: 1}
	old := ExtensionStatus{Patches: PatchStatus{Enabled: true, Applied: []string{"auth-rpc"}, PackageRoot: "/fixture/pi"}}
	m.adoptPatchStatusLocked(&old, 0)
	if old.Patches.Enabled || m.patchPolicyEnabled || len(old.Patches.Applied) != 1 || old.Patches.PackageRoot != "/fixture/pi" {
		t.Fatal("late launch replaced policy or erased actual patch report")
	}
	next := ExtensionStatus{Patches: PatchStatus{Enabled: true}}
	m.adoptPatchStatusLocked(&next, 1)
	if !m.patchPolicyEnabled {
		t.Fatal("unchanged generation failed to adopt newly loaded configuration")
	}
}
