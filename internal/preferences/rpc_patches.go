package preferences

import (
	"encoding/json"
	"path/filepath"
)

// RPCPatches is the pinned Rust rpc_patches::Config contract. Disabling the
// next-launch auto-apply does not undo changes already installed in Pi.
type RPCPatches struct {
	Enabled     bool   `json:"enabled"`
	PackageRoot string `json:"package_root,omitempty"`
}

func (s *Store) RPCPatches() RPCPatches {
	s.mu.Lock()
	defer s.mu.Unlock()
	nested := &Store{RootDir: filepath.Join(s.RootDir, "rpc-patches")}
	value := RPCPatches{Enabled: true}
	if err := nested.Read("config.json", &value); err != nil {
		return RPCPatches{Enabled: true}
	}
	return value
}

func (s *Store) SetRPCPatchesEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	nested := &Store{RootDir: filepath.Join(s.RootDir, "rpc-patches")}
	return nested.Update("config.json", func(m map[string]json.RawMessage) error {
		m["enabled"] = rawValue(enabled)
		return nil
	})
}
