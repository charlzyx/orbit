package native

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func executableFixture(t *testing.T, path, body string) {
	t.Helper()
	tempWrite(t, path, body)
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatal(err)
	}
}
func managedFixture(t *testing.T, agent string) string {
	t.Helper()
	root := filepath.Join(agent, "install")
	tempWrite(t, filepath.Join(root, "managed-install.json"), `{"kind":"pi-managed-install","schemaVersion":1,"layout":"releases-v1"}`)
	bin := filepath.Join(agent, "bin", "pi")
	executableFixture(t, bin, "#!/bin/sh\necho 1.2.0\n")
	return bin
}
func TestRuntimeInstallHintsAndWindowsSuffixes(t *testing.T) {
	for _, platform := range []string{"darwin", "linux", "windows"} {
		for _, bin := range []string{"pi", "node", "git"} {
			hint := DependencyInstallHint(platform, bin)
			if !strings.Contains(strings.ToLower(hint), bin) {
				t.Fatal(platform, bin, hint)
			}
		}
	}
	dir := t.TempDir()
	tempWrite(t, filepath.Join(dir, "pi"), "posix script")
	tempWrite(t, filepath.Join(dir, "pi.cmd"), "windows shim")
	if got := executableIn(dir, "pi", "windows"); got != filepath.Join(dir, "pi.cmd") {
		t.Fatal(got)
	}
	tempWrite(t, filepath.Join(dir, "node.exe"), "native")
	if executableIn(dir, "node", "windows") == "" {
		t.Fatal("node.exe missed")
	}
}
func TestRuntimeGUIPathsAndOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "")
	t.Setenv("PI_BIN", "")
	node := filepath.Join(home, ".nvm", "versions", "node", "v24.1.0", "bin", "node")
	executableFixture(t, node, "#!/bin/sh\necho v24.1.0\n")
	dirs := runtimeDirectories(home)
	found := false
	for _, dir := range dirs {
		if dir == filepath.Dir(node) {
			found = true
		}
	}
	if !found {
		t.Fatal("nvm directory missed", dirs)
	}
	custom := filepath.Join(t.TempDir(), "pi ; not-a-shell-command")
	executableFixture(t, custom, "#!/bin/sh\necho 9.9.9\n")
	t.Setenv("PI_BIN", custom)
	if got := LocateExecutable("pi"); got != custom {
		t.Fatal(got)
	}
	if version, err := ExecutableVersion(context.Background(), custom); err != nil || version != "9.9.9" {
		t.Fatal(version, err)
	}
	t.Setenv("PI_BIN", filepath.Join(home, "missing"))
	if got := LocateExecutable("pi"); got != "" {
		t.Fatal("invalid explicit override silently selected another pi", got)
	}
}
func TestPiClassificationAndPolicy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	t.Setenv("PI_MANAGED_INSTALL_ROOT", "")
	agent := filepath.Join(t.TempDir(), "agent")
	managed := managedFixture(t, agent)
	if kind, root := ClassifyPiInstallation(managed, agent); kind != PiManaged || root != filepath.Join(agent, "install") {
		t.Fatal(kind, root)
	}
	npm := filepath.Join(t.TempDir(), "node_modules", "@earendil-works", "pi-coding-agent")
	tempWrite(t, filepath.Join(npm, "package.json"), `{"name":"@earendil-works/pi-coding-agent"}`)
	bin := filepath.Join(npm, "dist", "bundle", "cli.js")
	executableFixture(t, bin, "#!/bin/sh\necho 1.0.0\n")
	if kind, root := ClassifyPiInstallation(bin, agent); kind != PiNPM || root != npm {
		t.Fatal(kind, root)
	}
	for kind, want := range map[PiInstallKind]PiUpdatePlan{PiManaged: PiUpdateManaged, PiNPM: PiMigrateNPM, PiOther: PiUnsupportedManager} {
		p := PiInstallation{Bin: bin, Version: "1.0.0", Kind: kind}
		if got := PlanPiUpdate(p, "1.1.0"); got != want {
			t.Fatal(kind, got)
		}
		if PlanPiUpdate(p, "1.0.0") != PiUpToDate {
			t.Fatal("updated current pi")
		}
	}
	if ManagedPiMarkerValid(t.TempDir()) {
		t.Fatal("stray install recognized")
	}
}
func TestManagedPiUpdateUsesExactArgumentsWithoutShell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX fixture")
	}
	agent := filepath.Join(t.TempDir(), "agent with spaces")
	bin := managedFixture(t, agent)
	log := filepath.Join(agent, "invocation")
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	t.Setenv("PI_MANAGED_INSTALL_ROOT", "")
	t.Setenv("PI_BIN", bin)
	t.Setenv("FIXTURE_LOG", log)
	executableFixture(t, bin, "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 1.2.0; exit 0; fi\nprintf '%s\\n' \"$@\" > \"$FIXTURE_LOG\"\n")
	if _, err := RunPiUpdate(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(log)
	if string(raw) != "update\n--no-approve\n" {
		t.Fatal(string(raw))
	}
	if _, err := RefreshPiModelCatalog(context.Background()); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(log)
	if string(raw) != "update\n--models\n--no-approve\n" {
		t.Fatal(string(raw))
	}
}
func TestPiVersionEndpointAndFailureFormatting(t *testing.T) {
	t.Setenv("PI_OFFLINE", "")
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != PiLatestVersionURL {
			t.Fatal(r.URL)
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"version":" 1.3.0 "}`)), Header: http.Header{}}, nil
	})}
	if latest, err := LatestPiVersion(context.Background(), client); err != nil || latest != "1.3.0" {
		t.Fatal(latest, err)
	}
	t.Setenv("PI_OFFLINE", "1")
	if latest, err := LatestPiVersion(context.Background(), nil); err != nil || latest != "" {
		t.Fatal("offline check used network")
	}
	if detail := RuntimeFailureDetail(nil, []byte("\x1b[31mError: npm failed\x1b[39m\nmore details\n")); detail != "npm failed" {
		t.Fatal(detail)
	}
}
func TestConfirmedMigrationFixturePreservesUserData(t *testing.T) {
	if !OfficialPiInstallSupported() {
		t.Skip("official managed installer unavailable")
	}
	base := t.TempDir()
	agent := filepath.Join(base, "agent")
	t.Setenv("PI_CODING_AGENT_DIR", agent)
	t.Setenv("PI_MANAGED_INSTALL_ROOT", "")
	t.Setenv("ORBIT_PI_INSTALL_CMD", "")
	t.Setenv("ORBIT_PI_INSTALL_URL", "")
	t.Setenv("TMPDIR", filepath.Join(base, "tmp"))
	os.MkdirAll(os.Getenv("TMPDIR"), 0700)
	npm := filepath.Join(base, "npm", "node_modules", "@earendil-works", "pi-coding-agent")
	tempWrite(t, filepath.Join(npm, "package.json"), `{"name":"@earendil-works/pi-coding-agent"}`)
	bin := filepath.Join(npm, "dist", "cli.js")
	executableFixture(t, bin, "#!/bin/sh\necho 1.0.0\n")
	t.Setenv("PI_BIN", bin)
	for _, file := range []string{"auth.json", "settings.json", "sessions/s.jsonl", "extensions/keep.js"} {
		tempWrite(t, filepath.Join(agent, filepath.FromSlash(file)), "keep:"+file)
	}
	calls := 0
	script := `#!/bin/sh
set -eu
agent="$PI_CODING_AGENT_DIR"
mkdir -p "$agent/install" "$agent/bin"
printf '%s' '{"kind":"pi-managed-install","schemaVersion":1,"layout":"releases-v1"}' > "$agent/install/managed-install.json"
printf '#!/bin/sh\necho 1.3.0\n' > "$agent/bin/pi"
chmod 755 "$agent/bin/pi"
`
	client := &http.Client{Transport: transport(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(script)), Header: http.Header{}}, nil
	})}
	if _, err := installPi(context.Background(), false, client); err == nil || calls != 0 {
		t.Fatal("unconfirmed migration ran", err, calls)
	}
	if _, err := installPi(context.Background(), true, client); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !ManagedPiMarkerValid(filepath.Join(agent, "install")) {
		t.Fatal("fixture installer failed")
	}
	for _, file := range []string{"auth.json", "settings.json", "sessions/s.jsonl", "extensions/keep.js"} {
		raw, err := os.ReadFile(filepath.Join(agent, filepath.FromSlash(file)))
		if err != nil || string(raw) != "keep:"+file {
			t.Fatal("user data changed", file)
		}
	}
	// This is a fixture-only installer. No real pi/npm/global installation ran.
	t.Log(fmt.Sprintf("fixture migration verified; %d HTTP request served in memory", calls))
	var marker map[string]any
	raw, _ := os.ReadFile(filepath.Join(agent, "install", "managed-install.json"))
	if json.Unmarshal(raw, &marker) != nil {
		t.Fatal("bad marker")
	}
}
