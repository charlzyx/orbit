package agent

import (
	"path/filepath"
	"testing"
)

func TestManagerConfigSharesResolvedPiIdentity(t *testing.T) {
	private, user := t.TempDir(), t.TempDir()
	for _, tt := range []struct {
		name     string
		o        Options
		home, pi string
	}{
		{"separate roots", Options{Home: private, UserHome: user, PiAgentDir: "~/custom"}, user, filepath.Join(user, "custom")},
		{"complete environment", Options{Environment: []string{"HOME=" + user, "USERPROFILE=" + user, "PI_CODING_AGENT_DIR=~/custom"}}, user, filepath.Join(user, "custom")},
		{"empty override clears ambient PiDir", Options{Home: private, UserHome: user, Environment: []string{"HOME=" + user, "PI_CODING_AGENT_DIR=~/ambient"}, EnvOverrides: map[string]string{"PI_CODING_AGENT_DIR": ""}}, user, filepath.Join(user, ".pi", "agent")},
		{"legacy explicit isolation", Options{Home: private}, private, filepath.Join(private, ".pi", "agent")},
		{"explicit environment override", Options{Home: private, EnvOverrides: map[string]string{"HOME": user, "USERPROFILE": user, "PI_CODING_AGENT_DIR": "~/override"}}, user, filepath.Join(user, "override")},
		{"resolved options win", Options{Home: private, UserHome: user, PiAgentDir: filepath.Join(user, "chosen"), EnvOverrides: map[string]string{"HOME": "wrong", "PI_CODING_AGENT_DIR": "wrong"}}, user, filepath.Join(user, "chosen")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.o)
			defer m.Close()
			env := m.piEnvOverrides()
			if env["HOME"] != tt.home || env["USERPROFILE"] != tt.home || env["PI_CODING_AGENT_DIR"] != tt.pi || m.options.SessionDir != filepath.Join(tt.pi, "sessions") {
				t.Fatal("inconsistent child and session configuration")
			}
			if env["ORBIT_PREFERENCES_HOME"] != m.options.Home {
				t.Fatal("extension prefs not private")
			}
		})
	}
}

func TestManagerRejectsInvalidPiConfigBeforeChildLaunch(t *testing.T) {
	m := NewManager(Options{Home: t.TempDir(), Environment: []string{}, PiAgentDir: "file:///bad%2Fpath", PiBinary: "must-not-launch"})
	defer m.Close()
	if err := m.Open(t.TempDir(), ""); err == nil {
		t.Fatal("invalid Pi config launched a child")
	}
}
