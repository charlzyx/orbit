package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// A capture-only PI_BIN launcher must not suppress Orbit's normal package
// patch discovery. This is a resolver test, not a native capture assertion.
func TestCaptureLauncherInsidePiPackagePreservesDiscovery(t *testing.T) {
	root := t.TempDir()
	packageRoot := filepath.Join(root, "node_modules", "@earendil-works", "pi-coding-agent")
	if err := os.MkdirAll(filepath.Join(packageRoot, "dist", "bundle"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageRoot, "package.json"), []byte(`{"name":"@earendil-works/pi-coding-agent","version":"1.4.0"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{filepath.Join(packageRoot, "dist", "bundle", "cli.js"), filepath.Join(packageRoot, ".orbit-chat-launcher.py")} {
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0700); err != nil {
			t.Fatal(err)
		}
		if got := findPackageRoot(path, filepath.Join(root, "agent")); got != packageRoot {
			t.Fatalf("launcher changed package discovery: %s -> %s", path, got)
		}
	}
	if got := findPackageRoot(filepath.Join(t.TempDir(), "unrelated-fixture-launcher"), filepath.Join(root, "agent")); got != "" {
		t.Fatalf("foreign launcher unexpectedly resolved %s", got)
	}
}
