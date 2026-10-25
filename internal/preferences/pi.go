package preferences

import (
	"errors"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// NewWithPi separates Orbit preferences from Pi's ordinary user configuration.
// No files are copied or migrated. New(home) remains an isolated-home shortcut.
func NewWithPi(home, userHome, agentDir string) *Store {
	if absolute, err := filepath.Abs(home); err == nil {
		home = absolute
	}
	if userHome == "" {
		userHome = home
	}
	if absolute, err := filepath.Abs(userHome); err == nil {
		userHome = absolute
	}
	return &Store{Home: home, RootDir: filepath.Join(home, ".orbit-pi"), UserHome: userHome, PiDir: ResolvePiDir(userHome, agentDir)}
}

// ResolvePiDir follows Pi 1.4's getAgentDir/expandTildePath convention. Relative
// overrides are anchored once at startup so workspace switches cannot silently
// move the configuration used by the UI and its different child processes.
func ResolvePiDir(home, configured string) string {
	if err := ValidatePiDir(configured); err != nil {
		return configured
	}
	path := configured
	if runtime.GOOS == "windows" && strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//") && !strings.Contains(path, `\`) {
		// Pi also accepts Git Bash, WSL and Cygwin drive paths on Windows.
		drivePath := strings.TrimPrefix(path, "/")
		drivePath = strings.TrimPrefix(strings.TrimPrefix(drivePath, "mnt/"), "cygdrive/")
		if len(drivePath) >= 1 && ((drivePath[0] >= 'a' && drivePath[0] <= 'z') || (drivePath[0] >= 'A' && drivePath[0] <= 'Z')) && (len(drivePath) == 1 || drivePath[1] == '/') {
			path = strings.ToUpper(drivePath[:1]) + `:\` + strings.ReplaceAll(strings.TrimPrefix(drivePath[1:], "/"), "/", `\`)
		}
	}
	switch {
	case path == "":
		path = filepath.Join(home, ".pi", "agent")
	case path == "~":
		path = home
	case strings.HasPrefix(path, "~/") || runtime.GOOS == "windows" && strings.HasPrefix(path, `~\`):
		path = filepath.Join(home, path[2:])
	case strings.HasPrefix(path, "file://"):
		if u, err := url.Parse(path); err == nil {
			if u.Path == "" {
				u.Path = "/"
			}
			path = filepath.FromSlash(u.Path)
			if runtime.GOOS == "windows" {
				if u.Host != "" && !strings.EqualFold(u.Host, "localhost") {
					path = `\\` + u.Host + path
				} else {
					path = strings.TrimPrefix(path, `\`)
				}
			}
		}
	}
	if absolute, err := filepath.Abs(path); err == nil {
		return absolute
	}
	return filepath.Clean(path)
}

// PiEnvOverrides must be used by every Pi helper as well as the main RPC child.
// ORBIT_PREFERENCES_HOME is consumed only by Orbit extensions, not by stock Pi.
func (s *Store) PiEnvOverrides() map[string]string {
	return map[string]string{"HOME": s.UserHome, "USERPROFILE": s.UserHome, "PI_CODING_AGENT_DIR": s.PiDir, "ORBIT_PREFERENCES_HOME": s.Home}
}
func (s *Store) PiEnvironment(base []string) []string {
	overrides := s.PiEnvOverrides()
	env := make([]string, 0, len(base)+len(overrides))
	for _, entry := range base {
		key, _, _ := strings.Cut(entry, "=")
		if runtime.GOOS == "windows" {
			key = strings.ToUpper(key)
		}
		if _, replaced := overrides[key]; !replaced {
			env = append(env, entry)
		}
	}
	for _, key := range []string{"HOME", "USERPROFILE", "PI_CODING_AGENT_DIR", "ORBIT_PREFERENCES_HOME"} {
		env = append(env, key+"="+overrides[key])
	}
	return env
}

// ValidatePiDir rejects malformed file URLs before any Pi/view code can turn
// them into a different local tree. Pi 1.4 uses Node fileURLToPath, which rejects
// encoded path separators and non-local Unix hosts rather than treating them
// as ordinary relative paths. Do not include potentially private input in errors.
func ValidatePiDir(configured string) error {
	if strings.ContainsRune(configured, 0) {
		return errors.New("Pi configuration directory contains a null character")
	}
	if !strings.HasPrefix(configured, "file://") {
		return nil
	}
	u, err := url.Parse(configured)
	if err != nil || u.Scheme != "file" || u.User != nil || u.Port() != "" || strings.Contains(configured, `\`) {
		return errors.New("Pi configuration directory has an invalid file URL")
	}
	lower := strings.ToLower(u.EscapedPath())
	if strings.Contains(lower, "%2f") || runtime.GOOS == "windows" && strings.Contains(lower, "%5c") {
		return errors.New("Pi configuration file URL contains an encoded path separator")
	}
	if strings.ContainsRune(u.Path, 0) {
		return errors.New("Pi configuration file URL contains a null character")
	}
	if runtime.GOOS != "windows" && u.Host != "" && !strings.EqualFold(u.Host, "localhost") {
		return errors.New("Pi configuration file URL must use a local host")
	}
	if runtime.GOOS == "windows" && (u.Host == "" || strings.EqualFold(u.Host, "localhost")) {
		if len(u.Path) < 3 || u.Path[0] != '/' || u.Path[2] != ':' || !((u.Path[1] >= 'a' && u.Path[1] <= 'z') || (u.Path[1] >= 'A' && u.Path[1] <= 'Z')) {
			return errors.New("Pi configuration file URL must contain an absolute drive path")
		}
	}
	return nil
}
