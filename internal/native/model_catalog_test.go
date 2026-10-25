//go:build !windows

package native

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestModelCatalogConfiguredExecutableAndIsolatedEnvironment(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(root, "home")
	work := filepath.Join(root, "workspace")
	agent := filepath.Join(root, "pi-data")
	for _, p := range []string{home, work, agent} {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "custom pi ; literal")
	log := filepath.Join(root, "args")
	executableFixture(t, bin, "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$LOG\"\nprintf 'home=%s|profile=%s|agent=%s|cwd=%s|leak=%s|skip=%s\\n' \"$HOME\" \"$USERPROFILE\" \"$PI_CODING_AGENT_DIR\" \"$PWD\" \"$ORBIT_AMBIENT_ONLY\" \"$PI_SKIP_VERSION_CHECK\"\n")
	t.Setenv("ORBIT_AMBIENT_ONLY", "must-not-leak")
	options := ModelCatalogOptions{Binary: bin, Home: home, PiDir: agent, Workspace: work, Environment: []string{}, EnvOverrides: map[string]string{"LOG": log, "HOME": "wrong", "PI_CODING_AGENT_DIR": "wrong"}}
	out, err := RefreshPiModelCatalogWithConfig(context.Background(), options)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(log)
	if string(raw) != "update\n--models\n--no-approve\n" {
		t.Fatal(string(raw))
	}
	want := "home=" + home + "|profile=" + home + "|agent=" + agent + "|cwd=" + work + "|leak=|skip=1\n"
	if string(out.Stdout) != want || out.ExitCode != 0 {
		t.Fatalf("got %q; want %q", out.Stdout, want)
	}
}
func TestModelCatalogConfiguredFailureRedactsEnvironmentSecrets(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "pi")
	executableFixture(t, bin, "#!/bin/sh\nprintf 'Error: failed credential %s\\n' \"$CUSTOM_SECRET\" >&2\nexit 17\n")
	out, err := RefreshPiModelCatalogWithConfig(context.Background(), ModelCatalogOptions{Binary: bin, Environment: []string{"CUSTOM_SECRET=private-test-value"}})
	if err == nil || out.ExitCode != 17 || strings.Contains(err.Error(), "private-test-value") || strings.Contains(string(out.Stderr), "private-test-value") || !strings.Contains(err.Error(), "REDACTED") {
		t.Fatal(out, err)
	}
}
func TestModelCatalogConfiguredTimeoutAndInvalidExecutable(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "pi")
	executableFixture(t, bin, "#!/bin/sh\nsleep 60 & wait\n")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := RefreshPiModelCatalogWithConfig(ctx, ModelCatalogOptions{Binary: bin, Environment: []string{}})
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 3*time.Second {
		t.Fatal(err, time.Since(start))
	}
	_, err = RefreshPiModelCatalogWithConfig(context.Background(), ModelCatalogOptions{Binary: filepath.Join(t.TempDir(), "missing"), Environment: []string{}})
	if err == nil {
		t.Fatal("invalid configured binary fell back")
	}
}
func TestModelCatalogConfiguredOutputIsBounded(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "pi")
	executableFixture(t, bin, "#!/bin/sh\nhead -c 2100000 /dev/zero\nhead -c 2100000 /dev/zero >&2\n")
	out, err := RefreshPiModelCatalogWithConfig(context.Background(), ModelCatalogOptions{Binary: bin, Environment: []string{}})
	if err != nil || len(out.Stdout) > 1<<20 || len(out.Stderr) > 1<<20 || len(out.Stdout) == 0 {
		t.Fatal(len(out.Stdout), len(out.Stderr), err)
	}
}

func TestModelCatalogRedactionCannotExpandPastOutputBound(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "pi")
	executableFixture(t, bin, "#!/bin/sh\nhead -c 1100000 /dev/zero | tr '\\000' A\n")
	out, err := RefreshPiModelCatalogWithConfig(context.Background(), ModelCatalogOptions{Binary: bin, Environment: []string{"FIRST_SECRET=A", "SECOND_SECRET=E", "THIRD_SECRET=D"}})
	if err != nil || len(out.Stdout) != 1<<20 || !strings.HasPrefix(string(out.Stdout), "[REDACTED]") {
		t.Fatal(len(out.Stdout), err)
	}
}
