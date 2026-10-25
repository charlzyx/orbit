package workbench

import (
	"github.com/charlzyx/orbit/internal/agent"
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
// include parked processes as well as the foreground session. No default policy
// removes anything: the user must enable it explicitly in Privacy settings.
func (s *Service) ApplyRetention(now time.Time, live []string) (RetentionOutcome, error) {
	out := RetentionOutcome{Failed: map[string]string{}}
	policy := s.Preferences.Retention()
	if !policy.Enabled {
		return out, nil
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
