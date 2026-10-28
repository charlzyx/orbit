// Package appbundle bootstraps the self-contained experimental macOS app.
// Its private preference/runtime home is separate from ordinary shared Pi config.
package appbundle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charlzyx/orbit/internal/preferences"
)

// StateDirectory is a persistent compatibility identity, not the display name.
// Renaming the app must not create an empty preference/runtime home.
const StateDirectory = "Orbit Go Test"

type Layout struct {
	Root       string `json:"bundle_path"`
	Executable string `json:"executable"`
	Ghostty    string `json:"ghostty_path"`
	Node       string `json:"node_path"`
	Pi         string `json:"pi_path"`
	PiCLI      string `json:"pi_cli_path"`
}

// Discover uses the executable, never Finder's arbitrary working directory.
// Resolving the executable also handles aliases/symlinks and App Translocation.
func Discover(executable, goos string) (Layout, bool) {
	if goos != "darwin" {
		return Layout{}, false
	}
	absolute, err := filepath.Abs(executable)
	if err != nil {
		return Layout{}, false
	}
	if real, err := filepath.EvalSymlinks(absolute); err == nil {
		absolute = real
	}
	macos := filepath.Dir(absolute)
	contents := filepath.Dir(macos)
	root := filepath.Dir(contents)
	if filepath.Base(macos) != "MacOS" || filepath.Base(contents) != "Contents" || !strings.HasSuffix(filepath.Base(root), ".app") {
		return Layout{}, false
	}
	resources := filepath.Join(contents, "Resources")
	return Layout{Root: root, Executable: absolute, Ghostty: filepath.Join(contents, "Frameworks", "libghostty-vt.dylib"), Node: filepath.Join(resources, "node", "bin", "node"), Pi: filepath.Join(resources, "bin", "pi"), PiCLI: filepath.Join(resources, "pi", "node_modules", "@earendil-works", "pi-coding-agent", "dist", "bundle", "cli.js")}, true
}

func inside(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && (rel == "." || filepath.IsLocal(rel))
}

func (l Layout) Validate() error {
	for _, item := range []struct {
		name, path string
		executable bool
	}{
		{"app executable", l.Executable, true}, {"Info.plist", filepath.Join(l.Root, "Contents", "Info.plist"), false},
		{"Ghostty terminal library", l.Ghostty, false}, {"Node runtime", l.Node, true}, {"Pi launcher", l.Pi, true}, {"Pi runtime", l.PiCLI, false}, {"runtime manifest", filepath.Join(l.Root, "Contents", "Resources", "orbit-bundle.json"), false},
	} {
		info, err := os.Stat(item.path)
		if err != nil {
			return fmt.Errorf("The app is incomplete: %s is unavailable (%s): %w. Download a fresh copy of Orbit.app; no runtime was installed", item.name, item.path, err)
		}
		if !info.Mode().IsRegular() || item.executable && info.Mode()&0111 == 0 {
			return fmt.Errorf("The bundled %s is not a usable file: %s. Download a fresh copy of Orbit.app", item.name, item.path)
		}
		real, err := filepath.EvalSymlinks(item.path)
		if err != nil || !inside(l.Root, real) {
			return fmt.Errorf("The bundled %s points outside Orbit.app: %s", item.name, item.path)
		}
	}
	return nil
}

type Options struct{ Home, Workspace, PiBinary string }
type Prepared struct {
	Home        string   `json:"home"`      // private Orbit preferences/runtime
	UserHome    string   `json:"user_home"` // actual Pi HOME
	PiAgentDir  string   `json:"pi_agent_dir"`
	Workspace   string   `json:"workspace"`
	PiBinary    string   `json:"-"`
	Environment []string `json:"-"`
}

// Prepare validates first, then creates only private state outside the bundle.
// An explicitly selected workspace or home is never replaced with a demo.
func (l Layout) Prepare(realHome string, o Options, environment []string) (Prepared, error) {
	var p Prepared
	if err := l.Validate(); err != nil {
		return p, err
	}
	home := o.Home
	if home == "" {
		if realHome == "" || !filepath.IsAbs(realHome) {
			return p, errors.New("Cannot locate your home folder for GoPi settings")
		}
		anchor, err := filepath.EvalSymlinks(realHome)
		if err != nil {
			return p, err
		}
		home = filepath.Join(anchor, "Library", "Application Support", StateDirectory, "Home")
		if err := rejectSymlinks(anchor, home); err != nil {
			return p, err
		}
	}
	var err error
	home, err = filepath.Abs(home)
	if err != nil {
		return p, err
	}
	if err = rejectSymlinks(home, home); err != nil {
		return p, err
	}
	if err = l.checkWritableLocation(home); err != nil {
		return p, err
	}
	p.Home, p.PiBinary = home, o.PiBinary
	p.UserHome = realHome
	configuredPiDir := ""
	for _, entry := range environment {
		if key, value, ok := strings.Cut(entry, "="); ok && key == "PI_CODING_AGENT_DIR" {
			configuredPiDir = value
		}
	}
	if o.Home != "" {
		// An explicit --home is an isolated configuration override, not the
		// default private Orbit-preferences directory and not a security sandbox.
		p.UserHome, configuredPiDir = home, ""
	}
	if err := preferences.ValidatePiDir(configuredPiDir); err != nil {
		return p, err
	}
	p.PiAgentDir = preferences.ResolvePiDir(p.UserHome, configuredPiDir)
	if p.PiBinary == "" {
		settings := filepath.Join(home, ".orbit-pi", "runtime.json")
		if err := rejectSymlinks(home, settings); err != nil {
			return p, err
		}
		data, readErr := os.ReadFile(settings)
		if readErr == nil {
			var saved struct {
				PiBinary string `json:"piBinary"`
			}
			if err := json.Unmarshal(data, &saved); err != nil {
				return p, fmt.Errorf("Runtime settings are invalid: %s", settings)
			}
			p.PiBinary = saved.PiBinary
		} else if !errors.Is(readErr, fs.ErrNotExist) {
			return p, readErr
		}
	}
	if p.PiBinary != "" {
		info, err := os.Stat(p.PiBinary)
		if !filepath.IsAbs(p.PiBinary) || err != nil || !info.Mode().IsRegular() || info.Mode()&0111 == 0 {
			if o.PiBinary == "" {
				return p, &RuntimeError{Home: home, Message: "Your selected Pi executable is unavailable: " + p.PiBinary + ". Restore it, or choose Use Bundled Pi to continue with this app’s runtime."}
			}
			return p, fmt.Errorf("The selected Pi executable is unavailable: %s. Restore it or select another executable", p.PiBinary)
		}
	}
	workspace := o.Workspace
	synthetic := false
	if workspace == "" {
		settings := filepath.Join(home, ".orbit-pi", "last-workspace.json")
		if err := rejectSymlinks(home, settings); err != nil {
			return p, err
		}
		data, readErr := os.ReadFile(settings)
		switch {
		case readErr == nil:
			var saved struct {
				Path string `json:"path"`
			}
			if err := json.Unmarshal(data, &saved); err != nil || saved.Path == "" || !filepath.IsAbs(saved.Path) {
				return p, &WorkspaceError{Home: home, Layout: l, Message: "The saved workspace settings could not be read. Choose a folder to continue; your existing projects are not changed."}
			}
			workspace = saved.Path
		case errors.Is(readErr, fs.ErrNotExist):
			workspace = filepath.Join(home, "Workspace")
			synthetic = true
		default:
			return p, fmt.Errorf("Cannot read saved workspace: %w", readErr)
		}
	}
	workspace, err = filepath.Abs(workspace)
	if err != nil {
		return p, err
	}
	if err = l.checkWritableLocation(workspace); err != nil {
		return p, err
	}
	if synthetic {
		if err := rejectSymlinks(home, workspace); err != nil {
			return p, err
		}
	} else if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		if o.Workspace == "" {
			return p, &WorkspaceError{Home: home, Layout: l, Message: "Your last workspace is unavailable: " + workspace + ". Reconnect it and reopen Orbit, or choose another folder."}
		}
		return p, fmt.Errorf("Your selected workspace is unavailable: %s. Reconnect it or select another folder", workspace)
	}
	// Protect only app-owned state. Ordinary Pi configuration may legitimately
	// use symlinks; Orbit neither copies it nor performs startup migrations.
	privatePaths := []string{".orbit-pi"}
	privateDirs := []string{home, filepath.Join(home, ".orbit-pi")}
	if o.Home != "" {
		privatePaths = append(privatePaths, ".pi", ".pi/agent", ".pi/agent/auth.json", ".pi/agent/settings.json", ".pi/agent/models.json", ".config", ".cache", ".local/share", ".local/state")
		for _, sub := range []string{".pi/agent", ".config", ".cache", ".local/share", ".local/state"} {
			privateDirs = append(privateDirs, filepath.Join(home, filepath.FromSlash(sub)))
		}
	}
	for _, sub := range privatePaths {
		dir := filepath.Join(home, filepath.FromSlash(sub))
		if err := rejectSymlinks(home, dir); err != nil {
			return p, err
		}
	}
	for _, dir := range privateDirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return p, fmt.Errorf("Cannot create private app settings: %w", err)
		}
	}
	if synthetic {
		if err := os.Mkdir(workspace, 0700); err == nil {
			if err := l.seedWorkspace(workspace); err != nil {
				return p, err
			}
		} else if !errors.Is(err, fs.ErrExist) {
			return p, fmt.Errorf("Cannot create the first-launch workspace: %w", err)
		} else if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
			return p, fmt.Errorf("The first-launch workspace is not a folder: %s", workspace)
		}
		if err := preferences.New(home).SaveLastWorkspace(workspace); err != nil {
			return p, err
		}
	}
	if p.PiBinary == "" {
		p.PiBinary, err = l.prepareRuntime(home)
		if err != nil {
			return p, err
		}
	}
	p.Workspace = workspace
	p.Environment = l.Environment(p, environment)
	return p, nil
}

func rejectSymlinks(anchor, path string) error {
	rel, err := filepath.Rel(anchor, path)
	if err != nil || rel != "." && !filepath.IsLocal(rel) {
		return fmt.Errorf("Invalid private settings path: %s", path)
	}
	current := anchor
	for _, part := range append([]string{"."}, strings.Split(rel, string(filepath.Separator))...) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("Private app settings must not follow a symbolic link: %s", current)
		}
	}
	return nil
}

// Resolve existing parents, including macOS's /var -> /private/var, before
// checking a new output path. A user cannot select the app as a workspace.
func (l Layout) checkWritableLocation(path string) error {
	current, suffix := path, ""
	for {
		real, err := filepath.EvalSymlinks(current)
		if err == nil {
			if inside(l.Root, filepath.Join(real, suffix)) {
				return fmt.Errorf("Orbit.app is read-only application content, not a home or workspace: %s", path)
			}
			return nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return err
		}
		suffix = filepath.Join(filepath.Base(current), suffix)
		current = parent
	}
}

// Environment keeps ordinary Pi credentials/configuration in their normal
// locations and inherits the launch environment. Only bundled runtime selection
// and private Orbit-extension preferences are overridden. Never log this slice.
func (l Layout) Environment(p Prepared, ambient []string) []string {
	values := map[string]string{}
	for _, entry := range ambient {
		if key, value, ok := strings.Cut(entry, "="); ok {
			values[key] = value
		}
	}
	store := preferences.NewWithPi(p.Home, p.UserHome, p.PiAgentDir)
	for key, value := range store.PiEnvOverrides() {
		values[key] = value
	}
	path := values["PATH"]
	if path == "" {
		path = "/usr/bin:/bin:/usr/sbin:/sbin"
	}
	// Relaunch inherits this environment; do not grow PATH on every relaunch.
	paths := append([]string{filepath.Dir(p.PiBinary), filepath.Dir(l.Node), filepath.Dir(l.Pi)}, filepath.SplitList(path)...)
	seen := map[string]bool{}
	var unique []string
	for _, directory := range paths {
		if !seen[directory] {
			unique = append(unique, directory)
			seen[directory] = true
		}
	}
	values["PATH"] = strings.Join(unique, string(os.PathListSeparator))
	values["PI_BIN"], values["MYGO_GHOSTTY_VT"], values["PI_SKIP_VERSION_CHECK"] = p.PiBinary, l.Ghostty, "1"
	if values["SHELL"] == "" {
		values["SHELL"] = "/bin/zsh"
	}
	// Explicit --home keeps caches and package-manager state inside that override.
	if p.Home == p.UserHome {
		for key, sub := range map[string]string{"XDG_CONFIG_HOME": ".config", "XDG_CACHE_HOME": ".cache", "XDG_DATA_HOME": ".local/share", "XDG_STATE_HOME": ".local/state", "npm_config_cache": ".cache/npm", "npm_config_userconfig": ".npmrc", "npm_config_prefix": ".local"} {
			values[key] = filepath.Join(p.Home, filepath.FromSlash(sub))
		}
	}
	// An old private-HOME marker must never redirect a normal launch.
	delete(values, "ORBIT_BUNDLE_STATE_HOME")
	env := make([]string, 0, len(values))
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	return env
}

// ApplyEnvironment is called once on the main goroutine before any controllers
// or child processes exist. It is deliberately never used by the unbundled CLI.
func ApplyEnvironment(environment []string) error {
	os.Clearenv()
	for _, entry := range environment {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			return errors.New("Invalid bundled environment")
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

func FinderArguments(args []string) []string {
	out := make([]string, 0, len(args))
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-psn_") {
			out = append(out, arg)
		}
	}
	return out
}

func (l Layout) seedWorkspace(workspace string) error {
	for _, name := range []string{"README.md", "hello.go"} {
		path := filepath.Join(l.Root, "Contents", "Resources", "sample-workspace", name)
		real, err := filepath.EvalSymlinks(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil || !inside(l.Root, real) {
			return fmt.Errorf("Invalid starter workspace resource: %s", name)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(filepath.Join(workspace, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		if err != nil {
			return err
		}
		_, err = file.Write(data)
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return err
		}
	}
	return nil
}
