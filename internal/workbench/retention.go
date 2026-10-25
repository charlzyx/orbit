package workbench

import (
	"errors"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/preferences"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type RetentionOutcome struct {
	Archived, Deleted int
	SkippedLive       []string
	Failed            map[string]string
}

// ApplyRetention uses message activity rather than filesystem mtime. live must
// include parked processes as well as the foreground session. This automatic
// entry point is opt-in. Privacy's explicit one-shot action uses CleanupSessions.
func (s *Service) ApplyRetention(now time.Time, live []string) (RetentionOutcome, error) {
	policy := s.Preferences.Retention()
	if !policy.Enabled {
		return RetentionOutcome{Failed: map[string]string{}}, nil
	}
	return s.CleanupSessions(now, live, policy)
}

// CleanupSessions applies an explicitly requested policy without changing the
// automatic-at-launch preference. Match Rust retention_ui.rs: enabled gates the
// launch hook, not "Clean up now". The caller snapshots the policy before moving
// work off the UI thread, and confirms a manual permanent-delete action first.
func (s *Service) CleanupSessions(now time.Time, live []string, policy preferences.Retention) (RetentionOutcome, error) {
	out := RetentionOutcome{Failed: map[string]string{}}
	if (policy.Mode != "archive" && policy.Mode != "delete") || !slices.Contains(preferences.RetentionDays, policy.Days) {
		return out, errors.New("invalid retention policy")
	}
	sessions, e := agent.ScanSessions(filepath.Join(s.Preferences.PiDir, "sessions"))
	if e != nil {
		return out, e
	}
	cutoff := now.Add(-time.Duration(policy.Days) * 24 * time.Hour)
	for _, session := range sessions {
		if session.Updated.IsZero() || !session.Updated.Before(cutoff) {
			continue
		}
		if slices.Contains(live, session.Path) {
			out.SkippedLive = append(out.SkippedLive, session.Path)
			continue
		}
		if policy.Mode == "archive" {
			if e = s.Preferences.SetSession(true, session.Path, true); e != nil {
				out.Failed[session.Path] = e.Error()
			} else {
				out.Archived++
			}
		} else {
			if e = os.Remove(session.Path); e != nil {
				out.Failed[session.Path] = e.Error()
				continue
			}
			out.Deleted++
			if e = s.Preferences.SetSession(false, session.Path, false); e != nil {
				out.Failed[session.Path] = e.Error()
			}
			if e = s.Preferences.SetSession(true, session.Path, false); e != nil {
				out.Failed[session.Path] = e.Error()
			}
		}
	}
	return out, nil
}
