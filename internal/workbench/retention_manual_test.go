package workbench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
)

func TestManualRetentionUsesRequestedPolicyAndKeepsLiveSessions(t *testing.T) {
	s := testService(t)
	old := filepath.Join(s.Preferences.PiDir, "sessions", "fixture", "old.jsonl")
	live := filepath.Join(s.Preferences.PiDir, "sessions", "fixture", "live.jsonl")
	data := strings.ReplaceAll(usageFixture(), "2026-10-25", "2026-01-01")
	put(t, old, data)
	put(t, live, data)
	now := time.Date(2026, 10, 25, 0, 0, 0, 0, time.UTC)
	requested := preferences.Retention{Enabled: false, Mode: "archive", Days: 30}
	if err := s.Preferences.SaveRetention(requested); err != nil {
		t.Fatal(err)
	}
	// A subsequent preference change must not change the already requested pass.
	if err := s.Preferences.SaveRetention(preferences.Retention{Enabled: false, Mode: "delete", Days: 7}); err != nil {
		t.Fatal(err)
	}
	out, err := s.CleanupSessions(now, []string{live}, requested)
	if err != nil || out.Archived != 1 || out.Deleted != 0 || len(out.SkippedLive) != 1 || out.SkippedLive[0] != live {
		t.Fatal(out, err)
	}
	for _, path := range []string{old, live} {
		got, err := os.ReadFile(path)
		if err != nil || string(got) != data {
			t.Fatal("archive changed a session", path, err)
		}
	}
	if s.Preferences.Retention().Enabled {
		t.Fatal("manual cleanup enabled the launch pass")
	}
	out, err = s.CleanupSessions(now, []string{live}, preferences.Retention{Mode: "delete", Days: 30})
	if err != nil || out.Deleted != 1 || len(out.SkippedLive) != 1 {
		t.Fatal(out, err)
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("requested idle file not deleted", err)
	}
	if _, err := os.Stat(live); err != nil {
		t.Fatal("live session removed", err)
	}
	if len(s.Preferences.SessionSet(true)) != 0 {
		t.Fatal("deleted session retained archive entry")
	}
}

func TestManualRetentionRejectsInvalidPolicyWithoutChanges(t *testing.T) {
	s := testService(t)
	path := filepath.Join(s.Preferences.PiDir, "sessions", "fixture", "old.jsonl")
	data := strings.ReplaceAll(usageFixture(), "2026-10-25", "2026-01-01")
	put(t, path, data)
	for _, policy := range []preferences.Retention{{Mode: "delete", Days: 0}, {Mode: "invalid", Days: 30}, {Mode: "delete", Days: -7}} {
		if _, err := s.CleanupSessions(time.Now(), nil, policy); err == nil {
			t.Fatal("invalid policy accepted", policy)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != data || len(s.Preferences.SessionSet(true)) != 0 {
		t.Fatal("invalid policy changed session", err)
	}
}
