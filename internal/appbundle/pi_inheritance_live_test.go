package appbundle

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
)

// Real stock Pi 1.4 JSONL RPC, but only synthetic local config and catalog/state
// getters. No prompt, inference, login, external provider or remote write occurs.
// Stock Pi is allowed to perform its normal writes inside these disposable roots.
func TestLiveBundledPiInheritsExistingModelsOffline(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_INTEGRATION_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_INTEGRATION_BIN to the pinned Pi 1.4.0 executable")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	for _, configured := range []string{"", "~/custom-agent"} {
		t.Run(map[bool]string{true: "default", false: "tilde-override"}[configured == ""], func(t *testing.T) {
			l, home := fixture(t)
			// The fixture's bundled node shim delegates to the actual test Node only.
			if err := os.WriteFile(l.Node, []byte("#!/bin/sh\nexec '"+node+"' \"$@\"\n"), 0700); err != nil {
				t.Fatal(err)
			}
			piDir := preferences.ResolvePiDir(home, configured)
			if err := os.MkdirAll(piDir, 0700); err != nil {
				t.Fatal(err)
			}
			models := `{"providers":{"orbit-offline-fixture":{"baseUrl":"http://127.0.0.1:1/v1","api":"openai-completions","apiKey":"offline-fixture-only","models":[{"id":"existing-model","name":"Existing local fixture"}]}}}`
			settings := `{"defaultProvider":"orbit-offline-fixture","defaultModel":"existing-model","enableInstallTelemetry":false,"enableAnalytics":false}`
			for name, data := range map[string]string{"models.json": models, "settings.json": settings} {
				if err := os.WriteFile(filepath.Join(piDir, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			// Never borrow os.Environ, API keys, SSH/proxy/cloud state or the real home.
			ambient := []string{"HOME=" + home, "USERPROFILE=" + home, "PATH=" + filepath.Dir(node) + string(os.PathListSeparator) + "/usr/bin:/bin", "PI_CODING_AGENT_DIR=" + configured, "PI_OFFLINE=1", "PI_SKIP_VERSION_CHECK=1"}
			version := exec.Command(binary, "--version")
			version.Env = ambient
			version.Dir = t.TempDir()
			if raw, err := version.Output(); err != nil || string(raw) != "1.4.0\n" {
				t.Fatal("pinned Pi 1.4.0 verification failed", err)
			}
			for launch := 0; launch < 2; launch++ {
				p, err := l.Prepare(home, Options{PiBinary: binary}, ambient)
				if err != nil {
					t.Fatal(err)
				}
				store := preferences.NewWithPi(p.Home, p.UserHome, p.PiAgentDir)
				providers, err := workbench.New(store, p.Workspace).Providers()
				if err != nil || len(providers) != 1 {
					t.Fatal("provider settings do not see existing shared model configuration", err)
				}
				m := agent.NewManager(agent.Options{PiBinary: p.PiBinary, Home: p.Home, UserHome: p.UserHome, PiAgentDir: p.PiAgentDir, Environment: p.Environment, Offline: true, NoSession: true, DisableExtensions: true, DisablePatches: true, DisableCheckpoints: true, CommandTimeout: 15 * time.Second, PiArgs: []string{"--no-extensions", "--no-skills", "--no-prompt-templates", "--no-context-files", "--no-tools", "--no-approve"}})
				if err := m.Open(p.Workspace, ""); err != nil {
					m.Close()
					t.Fatal(err)
				}
				found := false
				deadline := time.Now().Add(20 * time.Second)
				for time.Now().Before(deadline) {
					m.Tick()
					snapshot := m.Snapshot()
					for _, model := range snapshot.Models {
						if model.Provider == "orbit-offline-fixture" && model.ID == "existing-model" {
							found = snapshot.Connected && snapshot.Model.ID == "existing-model"
						}
					}
					if found {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				if err := m.Close(); err != nil {
					t.Fatal(err)
				}
				if !found {
					t.Fatal("real Pi RPC did not inherit the existing synthetic provider/model/default")
				}
				if _, err := os.Stat(filepath.Join(p.Home, ".pi", "agent", "models.json")); !os.IsNotExist(err) {
					t.Fatal("shared Pi models were copied into private Orbit home")
				}
				ambient = p.Environment
			}
		})
	}
}
