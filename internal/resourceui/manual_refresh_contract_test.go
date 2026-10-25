package resourceui

import (
	"path/filepath"
	"testing"
	"time"
)

// Cmd/Ctrl+R reuses Refresh, not RefreshPlugins. Verify the actual State's
// generation/coalescing and checkPending contract, with offline temp fixtures.
func TestManualSnapshotRefreshDoesNotSchedulePluginUpdateCheck(t *testing.T) {
	s := fixture(t)
	s.launchCheckDue = time.Time{} // exclude independent startup timer
	generation := s.generation
	checked := s.CheckedAt
	s.Refresh()
	s.Refresh()
	if s.generation <= generation || !s.Busy || !s.refreshPending || s.checkPending {
		t.Fatal("read refresh admission/side effects", s.generation, s.Busy, s.refreshPending, s.checkPending)
	}
	waitState(t, s)
	if s.checkPending || s.Checking || !s.CheckedAt.Equal(checked) {
		t.Fatal("snapshot refresh checked upstream metadata")
	}
}
func TestManualSnapshotRefreshKeepsReplacementWorkspaceOwner(t *testing.T) {
	s := fixture(t)
	s.launchCheckDue = time.Time{}
	next := t.TempDir()
	put(t, filepath.Join(next, ".git", "HEAD"), "ref: refs/heads/main\n")
	put(t, filepath.Join(next, ".pi", "skills", "next-owner", "SKILL.md"), "---\nname: next-owner\ndescription: Replacement workspace synthetic skill\n---\n# Next owner\n")
	s.Refresh()
	previous := s.generation
	s.SetWorkspace(next)
	s.Refresh()
	waitState(t, s)
	if s.Workspace != next || s.generation <= previous || s.checkPending || s.Checking {
		t.Fatal("replacement owner/check-state", s.Workspace, s.generation, s.checkPending)
	}
	nextFound := false
	for _, skill := range s.Skills {
		if skill.Name == "native-review" {
			t.Fatal("old workspace result applied")
		}
		if skill.Name == "next-owner" {
			nextFound = true
		}
	}
	if !nextFound {
		t.Fatal("replacement snapshot missing")
	}
}
