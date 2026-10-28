package appbundle

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
)

func fixture(t *testing.T) (Layout, string) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "App Translocation", "random id", "d", "Orbit Go Test.app")
	exe := filepath.Join(root, "Contents", "MacOS", "Orbit")
	l, ok := Discover(exe, "darwin")
	if !ok {
		t.Fatal("bundle not discovered")
	}
	for _, file := range []struct {
		path, text string
		mode       os.FileMode
	}{
		{exe, "native executable fixture", 0700},
		{filepath.Join(root, "Contents", "Info.plist"), "<?xml version=\"1.0\"?><plist><dict/></plist>", 0600},
		{l.Ghostty, "library fixture", 0600}, {l.Node, "#!/bin/sh\n", 0700}, {l.Pi, "#!/bin/sh\n", 0700}, {l.PiCLI, "#!/usr/bin/env node\n", 0700},
		{filepath.Join(root, "Contents", "Resources", "orbit-bundle.json"), `{"schema":1,"pi_version":"1.4.0","source_commit":"test"}`, 0600},
	} {
		if err := os.MkdirAll(filepath.Dir(file.path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file.path, []byte(file.text), file.mode); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "Contents", "Resources", "pi", "node_modules", ".bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../@earendil-works/pi-coding-agent/dist/bundle/cli.js", filepath.Join(bin, "pi")); err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(base, "User Home")
	if err := os.Mkdir(home, 0700); err != nil {
		t.Fatal(err)
	}
	return l, home
}
func envMap(env []string) map[string]string {
	m := map[string]string{}
	for _, v := range env {
		k, value, _ := strings.Cut(v, "=")
		m[k] = value
	}
	return m
}

func TestDiscoverUsesExecutableNotWorkingDirectory(t *testing.T) {
	l, _ := fixture(t)
	if got, ok := Discover(l.Executable, "darwin"); !ok || got.Root != l.Root {
		t.Fatalf("%+v %v", got, ok)
	}
	for _, path := range []string{"/Applications/Orbit.app/fake/Orbit", "/tmp/orbit", "/tmp/Orbit.app/Contents/Resources/Orbit"} {
		if _, ok := Discover(path, "darwin"); ok {
			t.Fatal(path)
		}
	}
	if _, ok := Discover(l.Executable, "linux"); ok {
		t.Fatal("bundled Linux CLI")
	}
	alias := filepath.Join(t.TempDir(), "orbit-link")
	if err := os.Symlink(l.Executable, alias); err != nil {
		t.Fatal(err)
	}
	if got, ok := Discover(alias, "darwin"); !ok || got.Root != l.Root {
		t.Fatalf("symlink: %+v %v", got, ok)
	}
}
func TestPreparePrivateFirstLaunchAndRepeat(t *testing.T) {
	l, realHome := fixture(t)
	realStore := preferences.New(realHome)
	if err := realStore.Write("auth.json", map[string]string{"credential": "do-not-copy"}); err != nil {
		t.Fatal(err)
	}
	p, err := l.Prepare(realHome, Options{}, []string{"LANG=en_US.UTF-8", "OPENAI_API_KEY=synthetic-inherited-key", "SSH_AUTH_SOCK=/secret", "NODE_OPTIONS=--require=/secret", "HTTP_PROXY=secret", "PI_BIN=/other", "PI_CODING_AGENT_DIR=/normal"})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(realHome, "Library", "Application Support", StateDirectory, "Home")
	if p.Home != want || p.Workspace != filepath.Join(want, "Workspace") {
		t.Fatalf("%+v", p)
	}
	if !strings.HasPrefix(p.PiBinary, filepath.Join(want, "runtime")+string(filepath.Separator)) {
		t.Fatal(p.PiBinary)
	}
	if _, err := os.Stat(filepath.Join(want, ".orbit-pi", "auth.json")); !os.IsNotExist(err) {
		t.Fatal("imported credentials")
	}
	env := envMap(p.Environment)
	if env["HOME"] != realHome || env["PI_CODING_AGENT_DIR"] != "/normal" || env["ORBIT_PREFERENCES_HOME"] != want || env["MYGO_GHOSTTY_VT"] != l.Ghostty || env["PI_BIN"] != p.PiBinary || env["LANG"] != "en_US.UTF-8" {
		t.Fatal(env)
	}
	for _, k := range []string{"OPENAI_API_KEY", "SSH_AUTH_SOCK", "NODE_OPTIONS", "HTTP_PROXY"} {
		if _, ok := env[k]; !ok {
			t.Fatal("normal launch lost environment key", k)
		}
	}
	if env["OPENAI_API_KEY"] != "synthetic-inherited-key" {
		t.Fatal("provider environment not inherited")
	}
	if !strings.HasPrefix(env["PATH"], filepath.Dir(p.PiBinary)+":"+filepath.Dir(l.Node)+":") {
		t.Fatal(env["PATH"])
	}
	info, err := os.Stat(p.Home)
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("private mode %v %v", info, err)
	}
	note := filepath.Join(p.Workspace, "user-edited.txt")
	if err := os.WriteFile(note, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	// Modify only the writable Pi copy: relaunch must not clobber applied patches.
	realPi, err := filepath.EvalSymlinks(p.PiBinary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(realPi, []byte("patched"), 0700); err != nil {
		t.Fatal(err)
	}
	again, err := l.Prepare(realHome, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if again.Home != p.Home || again.Workspace != p.Workspace || again.PiBinary != p.PiBinary {
		t.Fatalf("repeat changed %+v", again)
	}
	for _, item := range []struct{ path, want string }{{note, "keep"}, {realPi, "patched"}, {l.PiCLI, "#!/usr/bin/env node\n"}} {
		data, err := os.ReadFile(item.path)
		if err != nil || string(data) != item.want {
			t.Fatalf("%s changed", item.path)
		}
	}
}
func TestPrepareKeepsSavedWorkspaceAndPiAndExplicitOverrides(t *testing.T) {
	l, home := fixture(t)
	p, err := l.Prepare(home, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	chosen := t.TempDir()
	store := preferences.New(p.Home)
	if err := store.SaveLastWorkspace(chosen); err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(t.TempDir(), "custom-pi")
	if err := os.WriteFile(custom, []byte("#!/bin/sh\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := store.Write("runtime.json", map[string]string{"piBinary": custom}); err != nil {
		t.Fatal(err)
	}
	again, err := l.Prepare(home, Options{}, nil)
	if err != nil || again.Workspace != chosen || again.PiBinary != custom {
		t.Fatalf("%+v %v", again, err)
	}
	explicit := t.TempDir()
	again, err = l.Prepare(home, Options{Workspace: explicit, PiBinary: l.Node}, nil)
	if err != nil || again.Workspace != explicit || again.PiBinary != l.Node {
		t.Fatalf("explicit %+v %v", again, err)
	}
	if store.LastWorkspace() != chosen {
		t.Fatal("explicit override rewrote saved workspace")
	}
	if err := os.Remove(custom); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Prepare(home, Options{}, nil); err == nil {
		t.Fatal("missing saved runtime silently replaced")
	}
}
func TestUnavailableSavedWorkspaceHasNativeRecovery(t *testing.T) {
	l, home := fixture(t)
	p, err := l.Prepare(home, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(home, "removed-project")
	if err := preferences.New(p.Home).SaveLastWorkspace(missing); err != nil {
		t.Fatal(err)
	}
	_, err = l.Prepare(home, Options{}, nil)
	recovery, ok := err.(*WorkspaceError)
	if !ok || recovery.Home != p.Home || strings.Contains(recovery.Error(), "--workspace") {
		t.Fatalf("no native recovery: %v", err)
	}
	var saved struct{ Path string }
	if err := preferences.New(p.Home).Read("last-workspace.json", &saved); err != nil || saved.Path != missing {
		t.Fatal("discarded missing folder selection")
	}
}
func TestMissingResourceDoesNotCreateState(t *testing.T) {
	for _, name := range []string{"node", "ghostty", "pi", "manifest"} {
		t.Run(name, func(t *testing.T) {
			l, home := fixture(t)
			path := map[string]string{"node": l.Node, "ghostty": l.Ghostty, "pi": l.PiCLI, "manifest": filepath.Join(l.Root, "Contents", "Resources", "orbit-bundle.json")}[name]
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if _, err := l.Prepare(home, Options{}, nil); err == nil {
				t.Fatal("missing resource accepted")
			}
			if _, err := os.Stat(filepath.Join(home, "Library")); !os.IsNotExist(err) {
				t.Fatalf("created state before validation: %v", err)
			}
		})
	}
}
func TestRejectBundleWritesAndRedirectedResources(t *testing.T) {
	l, home := fixture(t)
	for _, o := range []Options{{Home: filepath.Join(l.Root, "new-home")}, {Workspace: l.Root}} {
		if _, err := l.Prepare(home, o, nil); err == nil {
			t.Fatal("bundle used for writable state")
		}
	}
	if err := os.Remove(l.Node); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "node")
	if err := os.WriteFile(external, []byte("node"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, l.Node); err != nil {
		t.Fatal(err)
	}
	if err := l.Validate(); err == nil {
		t.Fatal("external resource accepted")
	}
}
func TestRejectDefaultSettingsSymlinkAndRuntimeEscape(t *testing.T) {
	l, home := fixture(t)
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(home, "Library")); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Prepare(home, Options{}, nil); err == nil {
		t.Fatal("followed Library link")
	}
	entries, _ := os.ReadDir(target)
	if len(entries) != 0 {
		t.Fatal("wrote redirected state")
	}
	l, home = fixture(t)
	source := filepath.Join(l.Root, "Contents", "Resources", "pi")
	if err := os.Symlink(target, filepath.Join(source, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Prepare(home, Options{}, nil); err == nil {
		t.Fatal("copied escaping runtime link")
	}
}
func TestReadOnlyBundleAndCorruptRuntime(t *testing.T) {
	l, home := fixture(t)
	if err := filepath.WalkDir(l.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return os.Chmod(path, info.Mode().Perm()&0555)
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = filepath.WalkDir(l.Root, func(path string, d os.DirEntry, err error) error {
			if err == nil && d.Type()&os.ModeSymlink == 0 {
				_ = os.Chmod(path, 0700)
			}
			return nil
		})
	})
	p, err := l.Prepare(home, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(p.PiBinary))))
	if err := os.WriteFile(filepath.Join(root, "ready.json"), []byte("bad"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Prepare(home, Options{}, nil); err == nil {
		t.Fatal("corrupt runtime silently overwritten")
	}
}
func TestFinderArgumentsPreserveCLI(t *testing.T) {
	got := FinderArguments([]string{"-psn_0_1234", "--workspace", "/Folder with spaces", "--version"})
	if strings.Join(got, "|") != "--workspace|/Folder with spaces|--version" {
		t.Fatal(got)
	}
}
func TestSmokeRequiresActualReadinessAndNeverReplacesReport(t *testing.T) {
	l, home := fixture(t)
	p, err := l.Prepare(home, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(t.TempDir(), "launch.json")
	smoke, err := NewSmoke(report, time.Second, l, p)
	if err != nil {
		t.Fatal(err)
	}
	smoke.GhosttyLoaded()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	quit := make(chan struct{})
	smoke.Start(ctx, func() SmokeSnapshot {
		return SmokeSnapshot{Connected: true, Alive: true, PID: 42, SessionID: "synthetic-unit-session", RuntimeID: "live-1", Launcher: p.PiBinary}
	}, func() { close(quit) })
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(report); !os.IsNotExist(err) {
		t.Fatal("passed without real first-frame callback")
	}
	smoke.WindowReady(false)
	smoke.WindowReady(true)
	smoke.Capture(func() ([]byte, error) {
		var out bytes.Buffer
		err := png.Encode(&out, image.NewRGBA(image.Rect(0, 0, 10, 10)))
		return out.Bytes(), err
	})
	select {
	case <-quit:
	case <-time.After(2 * time.Second):
		t.Fatal("smoke did not finish")
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var result SmokeReport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "PASS" || !result.Checks.PiRPCGetState || !result.Checks.FirstFrame || result.PiPID != 42 {
		t.Fatalf("%+v", result)
	}
	if _, err := NewSmoke(report, time.Second, l, p); err == nil {
		t.Fatal("accepted stale report")
	}
}
func TestSmokeTimeoutDoesNotFabricateSuccess(t *testing.T) {
	l, home := fixture(t)
	report := filepath.Join(t.TempDir(), "failed.json")
	smoke, err := NewSmoke(report, 100*time.Millisecond, l, Prepared{Home: home, PiBinary: l.Pi})
	if err != nil {
		t.Fatal(err)
	}
	quit := make(chan struct{})
	smoke.Start(context.Background(), func() SmokeSnapshot { return SmokeSnapshot{} }, func() { close(quit) })
	select {
	case <-quit:
	case <-time.After(time.Second):
		t.Fatal("missing timeout")
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var result SmokeReport
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "FAIL" || result.Checks.FirstFrame || result.Checks.PiRPCGetState {
		t.Fatalf("%+v", result)
	}
}

func TestStarterFilesOnlySeedBrandNewWorkspace(t *testing.T) {
	l, home := fixture(t)
	source := filepath.Join(l.Root, "Contents", "Resources", "sample-workspace")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"README.md", "hello.go"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte("starter"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := l.Prepare(home, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(p.Workspace, "README.md")
	if data, err := os.ReadFile(readme); err != nil || string(data) != "starter" {
		t.Fatalf("starter missing: %s %v", data, err)
	}
	if err := os.WriteFile(readme, []byte("user changes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := l.Prepare(home, Options{}, nil); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(readme); err != nil || string(data) != "user changes" {
		t.Fatal("relaunch changed starter edits")
	}
}
func TestDefaultRelaunchKeepsUserHomeAndPrivateState(t *testing.T) {
	l, home := fixture(t)
	p, err := l.Prepare(home, Options{}, []string{"HOME=" + home, "PI_CODING_AGENT_DIR=~/custom-agent"})
	if err != nil {
		t.Fatal(err)
	}
	again, err := l.Prepare(envMap(p.Environment)["HOME"], Options{}, p.Environment)
	if err != nil {
		t.Fatal(err)
	}
	if p.Home != again.Home || p.UserHome != again.UserHome || p.PiAgentDir != again.PiAgentDir || p.PiBinary != again.PiBinary {
		t.Fatal("relaunch changed configuration identity")
	}
	if envMap(p.Environment)["PATH"] != envMap(again.Environment)["PATH"] {
		t.Fatal("relaunch grew runtime PATH")
	}
	if p.PiAgentDir != filepath.Join(home, "custom-agent") {
		t.Fatal("tilde override not resolved against user home")
	}
}

func TestExplicitHomeKeepsConfigurationIsolated(t *testing.T) {
	l, home := fixture(t)
	isolated := t.TempDir()
	p, err := l.Prepare(home, Options{Home: isolated}, []string{"HOME=" + home, "PI_CODING_AGENT_DIR=~/normal-agent", "XDG_CONFIG_HOME=/normal/config"})
	if err != nil {
		t.Fatal(err)
	}
	env := envMap(p.Environment)
	if p.Home != isolated || p.UserHome != isolated || p.PiAgentDir != filepath.Join(isolated, ".pi", "agent") || env["HOME"] != isolated || env["PI_CODING_AGENT_DIR"] != p.PiAgentDir || env["XDG_CONFIG_HOME"] != filepath.Join(isolated, ".config") {
		t.Fatal("--home failed to isolate paths")
	}
}

func TestPrepareDoesNotCopyOrMigrateSharedPiConfig(t *testing.T) {
	l, home := fixture(t)
	shared := filepath.Join(home, ".pi", "agent")
	if err := os.MkdirAll(shared, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"models.json", "settings.json", "auth.json"} {
		if err := os.WriteFile(filepath.Join(shared, name), []byte(`{"synthetic":"preserve-bootstrap-input"}`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := l.Prepare(home, Options{}, []string{"HOME=" + home})
	if err != nil {
		t.Fatal(err)
	}
	if p.PiAgentDir != shared {
		t.Fatal("normal Pi root not inherited")
	}
	for _, name := range []string{"models.json", "settings.json", "auth.json"} {
		raw, err := os.ReadFile(filepath.Join(shared, name))
		if err != nil || string(raw) != `{"synthetic":"preserve-bootstrap-input"}` {
			t.Fatal("Orbit bootstrap changed shared Pi input", name)
		}
		if _, err := os.Stat(filepath.Join(p.Home, ".pi", "agent", name)); !os.IsNotExist(err) {
			t.Fatal("Orbit copied shared Pi input", name)
		}
	}
	// This assertion covers bootstrap only. Launching stock Pi may perform its
	// ordinary migrations, auth refreshes, model cache and settings writes.
}

func TestSmokeCaptureFailureCannotPass(t *testing.T) {
	l, home := fixture(t)
	path := filepath.Join(t.TempDir(), "capture-failure.json")
	smoke, err := NewSmoke(path, time.Second, l, Prepared{Home: home, PiBinary: l.Pi})
	if err != nil {
		t.Fatal(err)
	}
	smoke.GhosttyLoaded()
	smoke.WindowReady(true)
	smoke.Capture(func() ([]byte, error) { return []byte("not a PNG"), nil })
	done := make(chan struct{})
	smoke.Start(context.Background(), func() SmokeSnapshot {
		return SmokeSnapshot{Connected: true, Alive: true, PID: 42, SessionID: "fixture", Launcher: l.Pi}
	}, func() { close(done) })
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("capture failure did not finish")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report SmokeReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "FAIL" || report.Checks.NativeCapture || report.Error == "" {
		t.Fatalf("false capture PASS: %+v", report)
	}
}

func TestSmokeEnvironmentDropsCredentialsWithoutChangingNormalLaunch(t *testing.T) {
	ambient := []string{"HOME=/synthetic/home", "PATH=/synthetic/bin", "PI_CODING_AGENT_DIR=/synthetic/pi", "OPENAI_API_KEY=synthetic-secret", "SSH_AUTH_SOCK=/synthetic/agent", "NODE_OPTIONS=--require=/synthetic/injection", "AWS_PROFILE=synthetic-profile", "PI_OFFLINE=0"}
	got := envMap(SmokeEnvironment(ambient))
	for _, key := range []string{"OPENAI_API_KEY", "SSH_AUTH_SOCK", "NODE_OPTIONS", "AWS_PROFILE"} {
		if _, ok := got[key]; ok {
			t.Fatal("smoke inherited sensitive ambient key", key)
		}
	}
	if got["HOME"] != "/synthetic/home" || got["PI_CODING_AGENT_DIR"] != "/synthetic/pi" || got["PI_OFFLINE"] != "1" {
		t.Fatal("smoke identity or offline mode lost")
	}
}

func TestInvalidSharedPiConfigFailsBeforePrivateStateCreation(t *testing.T) {
	l, home := fixture(t)
	if _, err := l.Prepare(home, Options{}, []string{"PI_CODING_AGENT_DIR=file:///bad%2Fpath"}); err == nil {
		t.Fatal("invalid shared Pi path accepted")
	}
	if _, err := os.Stat(filepath.Join(home, "Library")); !os.IsNotExist(err) {
		t.Fatal("invalid config created private runtime/state")
	}
}

func TestGoPiRenamedBundleReusesExistingPrivateHome(t *testing.T) {
	old, realHome := fixture(t)
	before, err := old.Prepare(realHome, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	note := filepath.Join(before.Workspace, "existing-user-file.txt")
	if err = os.WriteFile(note, []byte("preserve user data"), 0600); err != nil {
		t.Fatal(err)
	}
	renamed := filepath.Join(filepath.Dir(old.Root), "GoPi.app")
	if err = os.Rename(old.Root, renamed); err != nil {
		t.Fatal(err)
	}
	next, ok := Discover(filepath.Join(renamed, "Contents", "MacOS", "Orbit"), "darwin")
	if !ok {
		t.Fatal("GoPi bundle undiscovered")
	}
	after, err := next.Prepare(realHome, Options{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if after.Home != before.Home || after.Workspace != before.Workspace || after.PiBinary != before.PiBinary {
		t.Fatal("rename changed persistent identity", before.Home, after.Home)
	}
	data, err := os.ReadFile(note)
	if err != nil || string(data) != "preserve user data" {
		t.Fatal("existing data lost", err)
	}
	if _, err = os.Stat(filepath.Join(realHome, "Library", "Application Support", "GoPi")); !os.IsNotExist(err) {
		t.Fatal("unexpected new data directory", err)
	}
}
