package workbench

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMCPTimeoutAcceptsTypedIntegers(t *testing.T) {
	for _, n := range []any{1, int64(20), uint64(300), float64(100), json.Number("900")} {
		if e := ValidateMCP("test", map[string]any{"command": "node", "timeout": n}); e != nil {
			t.Fatalf("%T rejected: %v", n, e)
		}
	}
	for _, n := range []any{0, int64(-1), float64(1.5), "300"} {
		if e := ValidateMCP("test", map[string]any{"command": "node", "timeout": n}); e == nil {
			t.Fatalf("%#v accepted", n)
		}
	}
}
func TestMCPMetadataRoundtripDoesNotPollutePiConfig(t *testing.T) {
	s := testService(t)
	if e := s.SaveMCP("global", "", "one", map[string]any{"command": "node"}); e != nil {
		t.Fatal(e)
	}
	if e := s.SetMCPDescription("global", "one", "Helpful description"); e != nil {
		t.Fatal(e)
	}
	servers, errs := s.MCP()
	if len(errs) > 0 || len(servers) != 1 || servers[0].Description != "Helpful description" {
		t.Fatal(servers, errs)
	}
	raw, _ := os.ReadFile(filepath.Join(s.Preferences.PiDir, "mcp.json"))
	if strings.Contains(string(raw), "Helpful description") {
		t.Fatal("Orbit metadata written to Pi")
	}
	raw, _ = os.ReadFile(filepath.Join(s.Preferences.RootDir, "mcp-metadata.json"))
	if !strings.Contains(string(raw), "global|"+s.Workspace+"|one") {
		t.Fatal("incompatible upstream metadata key", string(raw))
	}
	put(t, filepath.Join(s.Preferences.RootDir, "mcp-metadata.json"), `{"descriptions":false}`)
	if e := s.SetMCPDescription("global", "one", "replacement"); e == nil {
		t.Fatal("malformed metadata overwritten")
	}
	_, errs = s.MCP()
	if len(errs) == 0 {
		t.Fatal("metadata error not visible")
	}
}
func TestMCPCodemodeProjectPrecedence(t *testing.T) {
	s := testService(t)
	initial, e := s.MCPAutoEnableCodemode()
	if e != nil || initial != nil {
		t.Fatal("unset value invented", initial, e)
	}
	if e = s.SetMCPAutoEnableCodemode("global", false); e != nil {
		t.Fatal(e)
	}
	value, e := s.MCPAutoEnableCodemode()
	if e != nil || value == nil || *value {
		t.Fatal(value, e)
	}
	if e = s.SetMCPAutoEnableCodemode("project", true); e != nil {
		t.Fatal(e)
	}
	value, e = s.MCPAutoEnableCodemode()
	if e != nil || value == nil || !*value {
		t.Fatal(value, e)
	}
}
func TestMCPPruningKeepsHiddenGlobalAndOtherNamespaces(t *testing.T) {
	s := testService(t)
	for _, name := range []string{"MCP_GLOBAL_KEEP", "MCP_GLOBAL_GONE", "MCP_POTHER_TOKEN", "USER_TOKEN"} {
		if e := s.SaveSecret(name, "test-value"); e != nil {
			t.Fatal(e)
		}
	}
	put(t, filepath.Join(s.Preferences.PiDir, "mcp.json"), `{"mcpServers":{"same":{"command":"global","env":{"TOKEN":"${MCP_GLOBAL_KEEP}"}}}}`)
	put(t, filepath.Join(s.Workspace, ".pi", "mcp.json"), `{"mcpServers":{"same":{"command":"project"}}}`)
	if e := s.PruneMCPSecrets("MCP_GLOBAL_", []string{"MCP_GLOBAL_KEEP", "MCP_GLOBAL_GONE", "MCP_POTHER_TOKEN", "USER_TOKEN"}); e != nil {
		t.Fatal(e)
	}
	var values map[string]string
	if e := s.Preferences.Read("mcp-secrets.json", &values); e != nil {
		t.Fatal(e)
	}
	if values["MCP_GLOBAL_KEEP"] == "" || values["MCP_GLOBAL_GONE"] != "" || values["MCP_POTHER_TOKEN"] == "" || values["USER_TOKEN"] == "" {
		t.Fatal("pruning removed unrelated secrets")
	}
}
func TestMCPProbeRetainsExitOneServerResults(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fake CLI")
	}
	s := testService(t)
	binary := filepath.Join(t.TempDir(), "fake-pi")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"servers\":[{\"name\":\"one\",\"state\":\"failed\",\"error\":\"server unavailable\"}],\"errors\":[\"connection failed\"]}'\nexit 1\n"
	if e := os.WriteFile(binary, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	s.PiBinary = binary
	raw, e := s.ProbeMCP(context.Background())
	if e != nil || !strings.Contains(string(raw), `"state":"failed"`) {
		t.Fatal(string(raw), e)
	}
	if e = os.WriteFile(binary, []byte("#!/bin/sh\necho 'unusable'; exit 1\n"), 0700); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ProbeMCP(context.Background()); e == nil {
		t.Fatal("invalid exit1 output swallowed")
	}
	if e = os.WriteFile(binary, []byte("#!/bin/sh\necho '{\"servers\":[]}'; exit 2\n"), 0700); e != nil {
		t.Fatal(e)
	}
	if _, e = s.ProbeMCP(context.Background()); e == nil {
		t.Fatal("unrelated command failure swallowed")
	}
}

func TestMCPProbeRedactsEscapedSecretValues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fake CLI")
	}
	s := testService(t)
	secret := "line1\n\"line2\""
	if e := s.SaveSecret("TOKEN", secret); e != nil {
		t.Fatal(e)
	}
	payload, _ := json.Marshal(map[string]any{"servers": []any{map[string]any{"name": "one", "state": "failed", "error": "bad token " + secret}}})
	binary := filepath.Join(t.TempDir(), "fake-pi")
	script := "#!/bin/sh\ncat <<'ORBIT_FAKE_PROBE'\n" + string(payload) + "\nORBIT_FAKE_PROBE\n"
	if e := os.WriteFile(binary, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	s.PiBinary = binary
	raw, e := s.ProbeMCP(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(raw), "line1") || strings.Contains(string(raw), "line2") {
		t.Fatal("JSON escaped secret leaked")
	}
}
func TestMCPRemoveCleansDescriptionAndOrphansWithoutTouchingSharedSecret(t *testing.T) {
	s := testService(t)
	put(t, filepath.Join(s.Preferences.PiDir, "mcp.json"), `{"mcpServers":{"one":{"command":"node","env":{"TOKEN":"${MCP_GLOBAL_ONE_TOKEN}"}},"other":{"command":"node","env":{"TOKEN":"${MCP_GLOBAL_SHARED_TOKEN}"}}}}`)
	for _, name := range []string{"MCP_GLOBAL_ONE_TOKEN", "MCP_GLOBAL_OLD_ORPHAN", "MCP_GLOBAL_SHARED_TOKEN", "MCP_POTHER_SECRET"} {
		if e := s.SaveSecret(name, "value"); e != nil {
			t.Fatal(e)
		}
	}
	if e := s.SetMCPDescription("global", "one", "note"); e != nil {
		t.Fatal(e)
	}
	if e := s.RemoveMCP("global", "one"); e != nil {
		t.Fatal(e)
	}
	descriptions, e := s.mcpDescriptions()
	if e != nil {
		t.Fatal(e)
	}
	if descriptions[s.mcpMetadataKey("global", "one")] != "" {
		t.Fatal("removed metadata retained")
	}
	var values map[string]string
	s.Preferences.Read("mcp-secrets.json", &values)
	if len(values) != 2 || values["MCP_GLOBAL_SHARED_TOKEN"] == "" || values["MCP_POTHER_SECRET"] == "" {
		t.Fatal("incorrect orphan pruning")
	}
}

func TestMCPSecretStoreWrongShapeIsNotSilentlyReplaced(t *testing.T) {
	for _, contents := range []string{`null`, `{"EXISTING":123}`, `{"EXISTING":null}`} {
		t.Run(contents, func(t *testing.T) {
			s := testService(t)
			path := filepath.Join(s.Preferences.RootDir, "mcp-secrets.json")
			put(t, path, contents)
			if e := s.SaveSecret("NEW", "value"); e == nil {
				t.Fatal("invalid secret store modified")
			}
			current, _ := os.ReadFile(path)
			if string(current) != contents {
				t.Fatal("invalid secret store overwritten")
			}
		})
	}
}
