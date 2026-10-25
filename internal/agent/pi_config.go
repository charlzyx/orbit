package agent

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charlzyx/orbit/internal/preferences"
)

func environmentValue(environment []string, key string) string {
	for i := len(environment) - 1; i >= 0; i-- {
		name, value, ok := strings.Cut(environment[i], "=")
		if ok && (name == key || runtime.GOOS == "windows" && strings.EqualFold(name, key)) {
			return value
		}
	}
	return ""
}
func resolvePiOptions(options *Options) error {
	explicitHome := options.Home != ""
	env := options.Environment
	if env == nil {
		env = os.Environ()
	}
	homeKey := "HOME"
	if runtime.GOOS == "windows" {
		homeKey = "USERPROFILE"
	}
	if options.Home == "" {
		options.Home = environmentValue(env, homeKey)
	}
	if options.UserHome == "" {
		options.UserHome, _ = overrideValue(options.EnvOverrides, homeKey)
		if options.UserHome == "" {
			options.UserHome = options.Home
		}
	}
	if options.PiAgentDir == "" {
		var explicitlyOverridden bool
		options.PiAgentDir, explicitlyOverridden = overrideValue(options.EnvOverrides, "PI_CODING_AGENT_DIR")
		// Home alone is an intentional isolated legacy override. A complete child
		// environment remains authoritative when explicitly supplied.
		if !explicitlyOverridden && (!explicitHome || options.Environment != nil) {
			options.PiAgentDir = environmentValue(env, "PI_CODING_AGENT_DIR")
		}
	}
	if err := preferences.ValidatePiDir(options.PiAgentDir); err != nil {
		return err
	}
	options.PiAgentDir = preferences.ResolvePiDir(options.UserHome, options.PiAgentDir)
	if options.SessionDir == "" {
		options.SessionDir = filepath.Join(options.PiAgentDir, "sessions")
	}
	return nil
}
func (m *Manager) piEnvOverrides() map[string]string {
	return preferences.NewWithPi(m.options.Home, m.options.UserHome, m.options.PiAgentDir).PiEnvOverrides()
}

func overrideValue(values map[string]string, key string) (string, bool) {
	if value, ok := values[key]; ok {
		return value, true
	}
	if runtime.GOOS == "windows" {
		for name, value := range values {
			if strings.EqualFold(name, key) {
				return value, true
			}
		}
	}
	return "", false
}
