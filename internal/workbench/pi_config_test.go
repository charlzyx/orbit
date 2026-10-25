package workbench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/preferences"
)

func TestPiConfigHelperProcess(t *testing.T) {
	if os.Getenv("ORBIT_PI_CONFIG_HELPER") != "1" {
		return
	}
	if os.Getenv("ORBIT_PI_CONFIG_MODE") == "json" {
		fmt.Print(`{"enabled":true,"count":1}`)
		os.Exit(0)
	}
	for _, key := range []string{"HOME", "USERPROFILE", "PI_CODING_AGENT_DIR", "ORBIT_PREFERENCES_HOME"} {
		fmt.Printf("%s=%s\n", key, os.Getenv(key))
	}
	fmt.Printf("secret=%s\n", os.Getenv("SYNTHETIC_API_KEY"))
	os.Exit(0)
}
func TestMCPAndResourceHelpersUseSharedPiAndPrivateOrbit(t *testing.T) {
	private, user, pi := t.TempDir(), t.TempDir(), t.TempDir()
	s := New(preferences.NewWithPi(private, user, pi), t.TempDir())
	s.Environment = []string{"ORBIT_PI_CONFIG_HELPER=1", "HOME=wrong", "USERPROFILE=wrong", "PI_CODING_AGENT_DIR=wrong", "ORBIT_PREFERENCES_HOME=wrong", "SYNTHETIC_API_KEY=synthetic-helper-value"}
	args := []string{"-test.run=^TestPiConfigHelperProcess$", "--"}
	out, err := s.Command(context.Background(), os.Args[0], args...)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range s.Preferences.PiEnvOverrides() {
		if !strings.Contains(out, key+"="+want+"\n") {
			t.Fatal("MCP helper used inconsistent", key)
		}
	}
	if strings.Contains(s.Redact(out), "synthetic-helper-value") {
		t.Fatal("helper returned inherited credential value")
	}
	out, err = s.resourceProbe(context.Background(), s.Workspace, os.Args[0], args...)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range s.Preferences.PiEnvOverrides() {
		if !strings.Contains(out, key+"="+want+"\n") {
			t.Fatal("resource helper used inconsistent", key)
		}
	}
	if strings.Contains(s.Redact(out), "synthetic-helper-value") {
		t.Fatal("resource helper returned inherited credential value")
	}
	// User-scoped skill discovery follows the Pi user home, not private prefs.
	skill := filepath.Join(user, ".agents", "skills", "inherited", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(skill), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(skill, []byte("---\nname: inherited\ndescription: synthetic\n---\nSynthetic skill\n"), 0600); err != nil {
		t.Fatal(err)
	}
	rows, errs := s.Skills()
	if len(errs) != 0 {
		t.Fatal(errs)
	}
	found := false
	for _, row := range rows {
		if row.Path == skill {
			found = true
		}
	}
	if !found {
		t.Fatal("normal user skill directory not inherited")
	}
}

func TestStructuredHelperOutputIsNotRedactedBeforeParsing(t *testing.T) {
	s := New(preferences.New(t.TempDir()), t.TempDir())
	s.Environment = []string{"ORBIT_PI_CONFIG_HELPER=1", "ORBIT_PI_CONFIG_MODE=json", "SYNTHETIC_API_KEY=true", "SYNTHETIC_TOKEN=1"}
	for _, probe := range []bool{false, true} {
		var out string
		var err error
		args := []string{"-test.run=^TestPiConfigHelperProcess$", "--"}
		if probe {
			out, err = s.resourceProbe(context.Background(), s.Workspace, os.Args[0], args...)
		} else {
			out, err = s.Command(context.Background(), os.Args[0], args...)
		}
		if err != nil {
			t.Fatal(err)
		}
		var value struct {
			Enabled bool
			Count   int
		}
		if err := json.Unmarshal([]byte(out), &value); err != nil || !value.Enabled || value.Count != 1 {
			t.Fatal("redaction corrupted structured provider/MCP data", err)
		}
	}
}
