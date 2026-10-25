package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type HostInfo struct{ Label, Arch, Unsupported string }
type HostFact struct {
	Label, LabelKey, Value, ValueKey, Note, NoteKey string
	Alert                                           bool
}

func ProbeHost(ctx context.Context) HostInfo {
	host := HostInfo{Label: runtime.GOOS, Arch: runtime.GOARCH}
	if host.Arch == "amd64" {
		host.Arch = "x86_64"
	}
	if host.Arch == "arm64" {
		host.Arch = "aarch64"
	}
	switch runtime.GOOS {
	case "darwin":
		host.Label = "macOS"
		if raw := hostCommand(ctx, "/usr/bin/sw_vers", "-productVersion"); raw != "" {
			host.Label = "macOS " + raw
			major, _ := strconv.Atoi(strings.Split(raw, ".")[0])
			if major > 0 && major < 13 {
				host.Unsupported = "Requires macOS 13 or newer."
			}
		}
	case "windows":
		host.Label = "Windows"
		raw := hostCommand(ctx, "cmd", "/D", "/C", "ver")
		version := regexp.MustCompile(`\d+\.\d+\.(\d+)`).FindStringSubmatch(raw)
		if len(version) > 1 {
			build, _ := strconv.Atoi(version[1])
			name := "Windows 10"
			if build >= 22000 {
				name = "Windows 11"
			}
			host.Label = name + " (build " + version[1] + ")"
			if build > 0 && build < 17763 {
				host.Unsupported = "Requires Windows 10 version 1809 (build 17763) or newer."
			}
		}
	default:
		if raw, err := os.ReadFile("/etc/os-release"); err == nil {
			if label := parseOSRelease(string(raw)); label != "" {
				host.Label = label
			}
		}
	}
	return host
}
func hostCommand(ctx context.Context, name string, args ...string) string {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	hideConsole(cmd)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
func parseOSRelease(raw string) string {
	values := map[string]string{}
	for _, line := range strings.Split(raw, "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		} else {
			value = strings.Trim(value, "'\"")
		}
		values[key] = value
	}
	if value := values["PRETTY_NAME"]; value != "" {
		return value
	}
	return strings.TrimSpace(values["NAME"] + " " + values["VERSION"])
}
func RuntimeHostFacts(host HostInfo, agentDir, orbitDir string) []HostFact {
	if agentDir == "" {
		agentDir = PiAgentDirectory()
	}
	if orbitDir == "" {
		orbitDir = filepath.Dir(PreferencePath("preferences.json"))
	}
	facts := []HostFact{{Label: "Platform", LabelKey: "onboarding.platform", Value: host.Label + " · " + host.Arch, Note: host.Unsupported, Alert: host.Unsupported != ""}}
	for _, spec := range [][3]string{{"pi sessions", "onboarding.pi_sessions", filepath.Join(agentDir, "sessions")}, {"Orbit config", "onboarding.orbit_config", orbitDir}} {
		f := HostFact{Label: spec[0], LabelKey: spec[1], Value: displayHomePath(spec[2])}
		if info, err := os.Stat(spec[2]); err != nil || !info.IsDir() {
			f.Note = "not created yet"
			f.NoteKey = "onboarding.not_created_yet"
		}
		facts = append(facts, f)
	}
	return facts
}
func PiInstallFacts(p PiInstallation) []HostFact {
	method := "missing"
	key := "onboarding.pi_method_missing"
	if p.Installed() {
		switch p.Kind {
		case PiManaged:
			method = "official (managed)"
			key = "onboarding.pi_method_official"
		case PiNPM:
			method = "npm"
			key = "onboarding.pi_method_npm"
		default:
			method = "other package manager"
			key = "onboarding.pi_method_other"
		}
	}
	path, version := "—", p.Version
	if p.Bin != "" {
		path = displayHomePath(p.Bin)
	}
	if version == "" {
		version = "—"
	}
	return []HostFact{{Label: "Pi install", LabelKey: "onboarding.pi_install", Value: method, ValueKey: key, Alert: !p.Installed()}, {Label: "Pi path", LabelKey: "onboarding.pi_path", Value: path}, {Label: "Pi version", LabelKey: "onboarding.pi_version", Value: version}}
}
func displayHomePath(path string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if path == home {
			path = "~"
		} else if strings.HasPrefix(path, home+string(filepath.Separator)) {
			path = "~" + strings.TrimPrefix(path, home)
		}
	}
	return strings.ReplaceAll(path, "\\", "/")
}
