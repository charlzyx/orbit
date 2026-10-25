package runtimepath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplicitEnvironmentNeverInheritsAmbientHomeOrSecrets(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "must-not-leak")
	t.Setenv("PI_BIN", "/ambient/pi")
	t.Setenv("HOME", "/ambient/home")
	env := AugmentEnvironmentFor([]string{"HOME=/clean/home", "PATH=/clean/bin", "MY_FLAG=kept"}, "/selected/bin", "linux")
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "ambient") || strings.Contains(joined, "must-not-leak") || strings.Contains(joined, "OPENAI_API_KEY") || !strings.Contains(joined, "MY_FLAG=kept") {
		t.Fatal(joined)
	}
	if !strings.Contains(joined, "PATH=/selected/bin:/clean/bin:") {
		t.Fatal(joined)
	}
	if got := ResolveIn("pi", []string{"PI_BIN=/definitely/not/a/launcher", "PATH=/clean/bin"}, "linux"); got != "" {
		t.Fatal(got)
	}
}
func TestWindowsSuffixesAndPathCase(t *testing.T) {
	dir := t.TempDir()
	for _, file := range []string{"pi", "pi.cmd", "node.exe"} {
		os.WriteFile(filepath.Join(dir, file), []byte("fixture"), 0600)
	}
	if got := ResolveIn("pi", []string{"Path=" + dir}, "windows"); got != filepath.Join(dir, "pi.cmd") {
		t.Fatal(got)
	}
	if got := ResolveIn("node", []string{"PATH=" + dir}, "windows"); got != filepath.Join(dir, "node.exe") {
		t.Fatal(got)
	}
}
func TestEmptyHomeDoesNotProbeRelativeShims(t *testing.T) {
	for _, dir := range Directories("", "", "linux") {
		if strings.HasPrefix(dir, ".volta") || strings.HasPrefix(dir, ".nvm") {
			t.Fatal(dir)
		}
	}
}
