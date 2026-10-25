package resourceui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
)

// This fixture uses real Git commands against a local bare remote. File-only
// transport forbids accidental contact with an external Git service.
func auditGitUpdateFixture(t *testing.T) (*State, string, string, string, func(string, ...string) string) {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip(err)
	}
	t.Setenv("GIT_ALLOW_PROTOCOL", "file")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	store := preferences.New(t.TempDir())
	workspace, base := t.TempDir(), t.TempDir()
	remote, origin := filepath.Join(base, "remote.git"), filepath.Join(base, "source")
	installed := filepath.Join(workspace, ".pi", "git", "example.invalid", "audit", "plugin")
	run := func(dir string, args ...string) string {
		t.Helper()
		cmd := exec.Command(git, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Fixture", "GIT_AUTHOR_EMAIL=fixture@example.invalid", "GIT_COMMITTER_NAME=Fixture", "GIT_COMMITTER_EMAIL=fixture@example.invalid")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run(base, "init", "--bare", remote)
	run(remote, "symbolic-ref", "HEAD", "refs/heads/main")
	run(base, "init", origin)
	put(t, filepath.Join(origin, "file"), "first")
	run(origin, "add", "file")
	run(origin, "commit", "-m", "first")
	run(origin, "remote", "add", "origin", remote)
	run(origin, "push", "origin", "HEAD:refs/heads/main")
	if err := os.MkdirAll(filepath.Dir(installed), 0700); err != nil {
		t.Fatal(err)
	}
	run(base, "clone", remote, installed)
	before := run(installed, "rev-parse", "HEAD")
	put(t, filepath.Join(origin, "file"), "second")
	run(origin, "add", "file")
	run(origin, "commit", "-m", "second")
	run(origin, "push", "origin", "HEAD:refs/heads/main")
	latest := run(origin, "rev-parse", "HEAD")
	put(t, filepath.Join(workspace, ".pi", "settings.json"), `{"packages":["git:example.invalid/audit/plugin"],"unrelated":true}`)
	s := New(store, workspace, Options{Offline: true})
	s.Refresh()
	waitState(t, s)
	s.Options.Offline = false
	s.launchCheckDue = time.Time{}
	t.Cleanup(s.Close)
	return s, installed, before, latest, run
}

func TestAuditNativeLocalGitUpdateOfferSkipAndReset(t *testing.T) {
	s, installed, before, latest, run := auditGitUpdateFixture(t)
	s.CheckUpdates()
	waitState(t, s)
	if s.Error != "" || len(s.UpdateErrors) != 0 || len(s.Updates) != 1 || s.Updates[0].Current != before || s.Updates[0].Latest != latest || s.Updates[0].Scope != "project" || s.CheckedAt.IsZero() {
		t.Fatalf("real local update offer lost: updates=%+v errors=%v / %s", s.Updates, s.UpdateErrors, s.Error)
	}
	source := s.Updates[0].Source
	s.SkipAll()
	waitState(t, s)
	if len(s.Updates) != 0 || s.Skipped[source] != latest {
		t.Fatal("skip did not persist exact offered version", s.Updates, s.Skipped)
	}
	s.CheckUpdates()
	waitState(t, s)
	if len(s.Updates) != 0 {
		t.Fatal("same skipped revision offered again", s.Updates)
	}
	s.SkipUpdate(source, "")
	waitState(t, s)
	s.CheckUpdates()
	waitState(t, s)
	if len(s.Updates) != 1 || s.Updates[0].Latest != latest {
		t.Fatal("cleared skip did not restore actual offer", s.Updates)
	}
	if run(installed, "rev-parse", "HEAD") != before || run(installed, "rev-parse", "origin/main") != before {
		t.Fatal("metadata checks changed installed checkout or fetched remote refs")
	}
}

func TestAuditNativeCheckedWorkspaceRejectsCanceledGitResult(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX bounded process fixture")
	}
	s, _, _, _, _ := auditGitUpdateFixture(t)
	bin := t.TempDir()
	marker := filepath.Join(bin, "started")
	put(t, filepath.Join(bin, "git"), "#!/bin/sh\n: > '"+marker+"'\nsleep 30\n")
	if err := os.Chmod(filepath.Join(bin, "git"), 0700); err != nil {
		t.Fatal(err)
	}
	s.SetPiBinary(filepath.Join(bin, "pi"))
	s.CheckUpdates()
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("real check did not reach fixture process")
		}
		time.Sleep(time.Millisecond)
	}
	next := t.TempDir()
	s.SetWorkspace(next)
	deadline = time.Now().Add(3 * time.Second)
	for len(s.results) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("canceled check did not finish/reap")
		}
		time.Sleep(time.Millisecond)
	}
	s.Tick()
	if s.Workspace != next || len(s.Updates) != 0 || len(s.UpdateErrors) != 0 || !s.CheckedAt.IsZero() || s.Busy || s.Checking || s.Error != "" {
		t.Fatalf("old workspace check leaked into new workspace: %+v", s)
	}
}

func TestAuditNativeDeletedSkillCleanupFailureShowsActualDiskState(t *testing.T) {
	s := fixture(t)
	var path string
	for _, p := range s.Skills {
		if p.Scope == "project" {
			path = p.Path
			break
		}
	}
	if path == "" {
		t.Fatal("missing project fixture")
	}
	changed := 0
	s.Options.OnModified = func() { changed++ }
	settings := filepath.Join(s.Workspace, ".pi", "settings.json")
	put(t, settings, "{malformed")
	s.AskDeleteSkill(path)
	s.ConfirmDeleteSkill()
	waitState(t, s)
	if !strings.Contains(s.Error, "skill deleted, but") || s.skill(path) != nil || s.HasModal() || changed != 1 {
		t.Fatalf("partial delete presented as untouched or retryable: error=%q skills=%v modal=%t changed=%d", s.Error, s.Skills, s.HasModal(), changed)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("skill was not actually deleted", err)
	}
	data, err := os.ReadFile(settings)
	if err != nil || string(data) != "{malformed" {
		t.Fatal("malformed settings overwritten", string(data), err)
	}
	s.ConfirmDeleteSkill()
	waitState(t, s)
	if changed != 1 {
		t.Fatal("repeated confirm reran deletion")
	}
}
