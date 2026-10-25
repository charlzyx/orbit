package native

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/egoist/mygo"
)

// ExternalApp is a resolved, installed folder-capable application. Target is
// an exact native bundle ID or executable; it is never interpolated in a shell.
type ExternalApp struct {
	ID, Label, Target, Path string
	Icon                    []byte
	Arguments               []string
}
type appSpec struct {
	id, label                             string
	bundles, commands, windowsPaths, args []string
}

var appCatalog = []appSpec{
	{id: "vscode", label: "VS Code", bundles: []string{"com.microsoft.VSCode"}, commands: []string{"code", "Code.exe"}, windowsPaths: []string{"%LOCALAPPDATA%\\Programs\\Microsoft VS Code\\Code.exe", "%ProgramFiles%\\Microsoft VS Code\\Code.exe", "%ProgramFiles(x86)%\\Microsoft VS Code\\Code.exe"}},
	{id: "cursor", label: "Cursor", bundles: []string{"com.todesktop.230313mzl4w4u92"}, commands: []string{"cursor", "Cursor.exe"}, windowsPaths: []string{"%LOCALAPPDATA%\\Programs\\cursor\\Cursor.exe", "%LOCALAPPDATA%\\Programs\\Cursor\\Cursor.exe", "%ProgramFiles%\\Cursor\\Cursor.exe"}},
	{id: "zed", label: "Zed", bundles: []string{"dev.zed.Zed", "dev.zed.Zed-Preview"}, commands: []string{"zed", "Zed.exe"}, windowsPaths: []string{"%LOCALAPPDATA%\\Programs\\Zed\\Zed.exe", "%ProgramFiles%\\Zed\\Zed.exe"}},
	{id: "finder", label: "Finder", bundles: []string{"com.apple.finder"}},
	{id: "explorer", label: "File Explorer", commands: []string{"explorer.exe"}, windowsPaths: []string{"%WINDIR%\\explorer.exe"}},
	{id: "files", label: "Files", commands: []string{"nautilus", "dolphin", "thunar", "pcmanfm"}},
	{id: "terminal", label: "Terminal", bundles: []string{"com.apple.Terminal"}, windowsPaths: []string{"%LOCALAPPDATA%\\Microsoft\\WindowsApps\\wt.exe"}, commands: []string{"wt.exe"}},
	{id: "termy", label: "Termy", bundles: []string{"com.lassevestergaard.termy"}},
	{id: "iterm2", label: "iTerm2", bundles: []string{"com.googlecode.iterm2"}},
	{id: "kitty", label: "Kitty", bundles: []string{"net.kovidgoyal.kitty"}, commands: []string{"kitty"}, args: []string{"--directory"}},
	{id: "ghostty", label: "Ghostty", bundles: []string{"com.mitchellh.ghostty"}, commands: []string{"ghostty"}, args: []string{"--working-directory"}},
	{id: "warp", label: "Warp", bundles: []string{"dev.warp.Warp-Stable", "dev.warp.Warp"}},
	{id: "xcode", label: "Xcode", bundles: []string{"com.apple.dt.Xcode"}},
	{id: "rider", label: "Rider", bundles: []string{"com.jetbrains.rider", "com.jetbrains.rider-EAP"}, commands: []string{"rider", "rider64.exe"}, windowsPaths: []string{"%LOCALAPPDATA%\\Programs\\Rider\\bin\\rider64.exe", "%ProgramFiles%\\JetBrains\\JetBrains Rider\\bin\\rider64.exe"}},
	{id: "android-studio", label: "Android Studio", bundles: []string{"com.google.android.studio"}, commands: []string{"android-studio", "studio64.exe"}, windowsPaths: []string{"%ProgramFiles%\\Android\\Android Studio\\bin\\studio64.exe", "%LOCALAPPDATA%\\Programs\\Android Studio\\bin\\studio64.exe"}},
}

// DetectOpenInApps uses the actual desktop's registry/icon service, never a
// decorative list. Call from a goroutine after App.WhenReady.
func DetectOpenInApps() ([]ExternalApp, error) {
	installed, err := mygo.Shell.InstalledApps()
	if err != nil && runtime.GOOS == "darwin" {
		return nil, err
	}
	var apps []ExternalApp
	for _, spec := range appCatalog {
		app := ExternalApp{ID: spec.id, Label: spec.label}
		if runtime.GOOS == "darwin" {
			for _, bundle := range spec.bundles {
				for _, candidate := range installed {
					if candidate.ID == bundle {
						app.Target = bundle
						app.Path = candidate.Path
						break
					}
				}
				if app.Target != "" {
					break
				}
			}
		} else {
			if runtime.GOOS == "windows" {
				for _, raw := range spec.windowsPaths {
					path, ok := expandWindowsEnvironment(raw)
					if !ok {
						continue
					}
					if info, e := os.Stat(path); e == nil && !info.IsDir() {
						app.Target = path
						break
					}
				}
			}
			if app.Target == "" {
				for _, command := range spec.commands {
					if runtime.GOOS != "windows" && strings.HasSuffix(command, ".exe") {
						continue
					}
					if path, e := exec.LookPath(command); e == nil {
						app.Target = path
						break
					}
				}
			}
			app.Path = app.Target
			app.Arguments = append([]string(nil), spec.args...)
			if runtime.GOOS == "windows" && spec.id == "terminal" {
				app.Label = "Windows Terminal"
				app.Arguments = []string{"-d"}
			}
			// Native desktop-entry artwork is preferable to an executable's generic
			// icon when the app is also installed through a Linux desktop menu.
			for _, candidate := range installed {
				if strings.EqualFold(candidate.Name, app.Label) || strings.EqualFold(strings.TrimSuffix(filepath.Base(candidate.ID), ".desktop"), spec.id) {
					app.Path = candidate.Path
					break
				}
			}
		}
		if app.Target == "" {
			continue
		}
		app.Icon, _ = mygo.Shell.FileIcon(app.Path, 36)
		apps = append(apps, app)
	}
	return apps, err
}
func expandWindowsEnvironment(raw string) (string, bool) {
	var b strings.Builder
	for {
		before, rest, ok := strings.Cut(raw, "%")
		b.WriteString(before)
		if !ok {
			return b.String(), true
		}
		name, after, ok := strings.Cut(rest, "%")
		if !ok {
			return "", false
		}
		value, found := os.LookupEnv(name)
		if !found {
			for _, entry := range os.Environ() {
				key, v, _ := strings.Cut(entry, "=")
				if strings.EqualFold(key, name) {
					value = v
					found = true
					break
				}
			}
		}
		if !found {
			return "", false
		}
		b.WriteString(value)
		raw = after
	}
}
func TermyURL(path string) string {
	u := url.URL{Scheme: "termy", Host: "new"}
	q := url.Values{}
	q.Set("dir", path)
	u.RawQuery = q.Encode()
	return u.String()
}

// OpenPathInApp starts the resolved target, preserving spaces and metacharacters
// in paths. It reports launch failure rather than pretending the app opened.
func OpenPathInApp(path string, app ExternalApp) error {
	if app.Target == "" {
		return errors.New("no installed application target")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if _, err = os.Stat(absolute); err != nil {
		return err
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		target := absolute
		if app.ID == "termy" {
			target = TermyURL(absolute)
		}
		cmd = exec.Command("/usr/bin/open", "-b", app.Target, target)
	} else {
		args := append([]string(nil), app.Arguments...)
		args = append(args, absolute)
		cmd = exec.Command(app.Target, args...)
	}
	return launch(cmd)
}
func launch(cmd *exec.Cmd) error {
	hideConsole(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

type OpenInPrefs struct {
	OpenInApp  string            `json:"open_in_app,omitempty"`
	Workspaces map[string]string `json:"workspaces"`
}

func LoadOpenInPrefs(path string) (OpenInPrefs, error) {
	p := OpenInPrefs{Workspaces: map[string]string{}}
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	err = json.Unmarshal(raw, &p)
	if p.Workspaces == nil {
		p.Workspaces = map[string]string{}
	}
	return p, err
}
func (p *OpenInPrefs) Remember(workspace, id string) {
	if p.Workspaces == nil {
		p.Workspaces = map[string]string{}
	}
	p.Workspaces[workspace] = id
}
func (p OpenInPrefs) Preferred(workspace string, apps []ExternalApp) *ExternalApp {
	fallback := "files"
	if runtime.GOOS == "darwin" {
		fallback = "finder"
	}
	if runtime.GOOS == "windows" {
		fallback = "explorer"
	}
	for _, id := range []string{p.Workspaces[workspace], p.OpenInApp, fallback} {
		if id != "" {
			for i := range apps {
				if apps[i].ID == id {
					return &apps[i]
				}
			}
		}
	}
	if len(apps) > 0 {
		return &apps[0]
	}
	return nil
}
func (p OpenInPrefs) Save(path string) error { return atomicJSON(path, p) }
func OpenNotificationSettings() error {
	switch runtime.GOOS {
	case "darwin":
		return mygo.Shell.OpenExternal("x-apple.systempreferences:com.apple.Notifications-Settings.extension")
	case "windows":
		return mygo.Shell.OpenExternal("ms-settings:notifications")
	default:
		for _, candidate := range [][]string{{"gnome-control-center", "notifications"}, {"systemsettings", "kcm_notifications"}} {
			if exe, err := exec.LookPath(candidate[0]); err == nil {
				return launch(exec.Command(exe, candidate[1:]...))
			}
		}
		return errors.New("no notification settings application found")
	}
}

// OpenTerminalProgram is the explicit external-terminal escape hatch for a pi
// authentication flow; normal shell work stays in the integrated terminal.
func OpenTerminalProgram(program string, args ...string) error {
	return OpenTerminalProgramWithEnvironment(program, nil, args...)
}

// OpenTerminalProgramWithEnvironment carries only non-secret Pi identity paths
// across LaunchServices. It never writes inherited credentials to a script.
func OpenTerminalProgramWithEnvironment(program string, environment map[string]string, args ...string) error {
	env, err := terminalPathEnvironment(environment)
	if err != nil {
		return err
	}
	if program == "" {
		return errors.New("terminal program is empty")
	}
	if runtime.GOOS == "darwin" {
		parent, err := EnsureTempDir()
		if err != nil {
			return err
		}
		dir, err := os.MkdirTemp(parent, fmt.Sprintf("update-%d-login-", os.Getpid()))
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "login.command")
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'" }
		command := quote(program)
		if len(env) > 0 {
			command = "/usr/bin/env"
			for _, entry := range env {
				command += " " + quote(entry)
			}
			command += " " + quote(program)
		}
		for _, arg := range args {
			command += " " + quote(arg)
		}
		script := "#!/bin/sh\ntrap 'rm -f -- \"$0\"; rmdir -- \"$(dirname -- \"$0\")\"' EXIT\n" + command + "\n"
		if err = os.WriteFile(path, []byte(script), 0700); err != nil {
			os.RemoveAll(dir)
			return err
		}
		if err = launch(exec.Command("/usr/bin/open", path)); err != nil {
			os.RemoveAll(dir)
			return err
		}
		return nil
	}
	if runtime.GOOS == "windows" {
		return launchExternalTerminal(program, args, env)
	}
	for _, spec := range [][]string{{"x-terminal-emulator", "-e"}, {"gnome-terminal", "--"}, {"konsole", "-e"}, {"xterm", "-e"}} {
		if exe, err := exec.LookPath(spec[0]); err == nil {
			argv := append([]string(nil), spec[1:]...)
			if len(env) > 0 {
				argv = append(argv, "/usr/bin/env")
				argv = append(argv, env...)
			}
			argv = append(argv, program)
			argv = append(argv, args...)
			return launch(exec.Command(exe, argv...))
		}
	}
	return fmt.Errorf("no installed terminal can run %s", program)
}

// Explicit allowlist prevents a future caller accidentally persisting secrets.
func terminalPathEnvironment(values map[string]string) ([]string, error) {
	for key, value := range values {
		switch key {
		case "HOME", "USERPROFILE", "PI_CODING_AGENT_DIR", "ORBIT_PREFERENCES_HOME", "PATH":
		default:
			return nil, errors.New("terminal environment accepts only configuration paths")
		}
		if strings.ContainsRune(value, 0) {
			return nil, errors.New("invalid terminal configuration path")
		}
	}
	var env []string
	for _, key := range []string{"HOME", "USERPROFILE", "PI_CODING_AGENT_DIR", "ORBIT_PREFERENCES_HOME", "PATH"} {
		if value, ok := values[key]; ok {
			env = append(env, key+"="+value)
		}
	}
	return env, nil
}
