package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRPCPatchPreferencesPreserveSourcePathAndUnknownKeys(t *testing.T) {
	s := New(t.TempDir())
	if !s.RPCPatches().Enabled {
		t.Fatal("missing config must default enabled")
	}
	path := filepath.Join(s.RootDir, "rpc-patches", "config.json")
	if err := AtomicWrite(path, []byte(`{"enabled":true,"package_root":"/fixture/pi","future":{"keep":1}}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, value := range []bool{false, true, false} {
		if err := s.SetRPCPatchesEnabled(value); err != nil {
			t.Fatal(err)
		}
		if got := New(s.Home).RPCPatches(); got.Enabled != value || got.PackageRoot != "/fixture/pi" {
			t.Fatal(got)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var m map[string]json.RawMessage
		var future map[string]int
		if json.Unmarshal(data, &m) != nil || json.Unmarshal(m["future"], &future) != nil || future["keep"] != 1 {
			t.Fatal("unknown key lost", string(data))
		}
	}
}

func TestRPCPatchMalformedReadDefaultsButSaveDoesNotDestroyBytes(t *testing.T) {
	s := New(t.TempDir())
	p := filepath.Join(s.RootDir, "rpc-patches", "config.json")
	for _, raw := range []string{`not json`, `{"enabled":false,"package_root":17}`, `null`} {
		if err := AtomicWrite(p, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if !s.RPCPatches().Enabled {
			t.Fatal("invalid source config did not default", raw)
		}
		if raw == `null` || raw == `not json` {
			if err := s.SetRPCPatchesEnabled(false); err == nil {
				t.Fatal("malformed config overwritten")
			}
			if got, _ := os.ReadFile(p); string(got) != raw {
				t.Fatal("failed save mutated bytes")
			}
		}
	}
}
