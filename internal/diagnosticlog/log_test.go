package diagnosticlog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func logPath(home string) string { return filepath.Join(home, ".orbit-pi", "logs", currentName) }

func newTestLogger(t *testing.T, limit int64) (*Logger, string) {
	t.Helper()
	home := t.TempDir()
	l, err := open(home, "0.2.12-go-test.20261025.4", limit)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return l, home
}

func readEntries(t *testing.T, path string) []entry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var got []entry
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		var v entry
		decoder := json.NewDecoder(strings.NewReader(scan.Text()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestRecordVisibleBeforeCloseAndPrivateSchema(t *testing.T) {
	l, home := newTestLogger(t, MaxBytes)
	t.Setenv("DIAGNOSTIC_TEST_SECRET", "never-record-auth-header-or-key")
	if !l.Record(EventWorkspaceBrowse, ResultStarted, 0) || !l.Record(EventWorkspaceBrowse, ResultOK, 123*time.Millisecond) {
		t.Fatal("valid events were dropped")
	}
	got := readEntries(t, logPath(home))
	if len(got) != 2 || got[1].DurationMS != 123 || got[1].Event != "workspace_browse" || got[1].Result != "ok" || got[1].Version != "0.2.12-go-test.20261025.4" {
		t.Fatalf("unexpected entries: %+v", got)
	}
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(got[0].RunID) || got[0].RunID != got[1].RunID {
		t.Fatal("missing per-run random identifier")
	}
	if _, err := time.Parse(time.RFC3339Nano, got[0].Timestamp); err != nil || !strings.HasSuffix(got[0].Timestamp, "Z") {
		t.Fatal("timestamp is not UTC RFC3339")
	}
	b, err := os.ReadFile(logPath(home))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{home, "never-record-auth-header-or-key", "DIAGNOSTIC_TEST_SECRET"} {
		if strings.Contains(string(b), secret) {
			t.Fatalf("private information was logged: %q", secret)
		}
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(strings.Split(string(b), "\n")[0]), &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 6 {
		t.Fatalf("unexpected diagnostic fields: %v", fields)
	}
	for _, field := range []string{"timestamp", "run_id", "version", "event", "result", "duration_ms"} {
		if _, ok := fields[field]; !ok {
			t.Fatalf("missing field %s", field)
		}
	}
	if runtime.GOOS != "windows" {
		for _, pair := range []struct {
			path string
			mode os.FileMode
		}{
			{filepath.Join(home, ".orbit-pi"), 0700},
			{filepath.Join(home, ".orbit-pi", "logs"), 0700},
			{logPath(home), 0600},
		} {
			info, err := os.Stat(pair.path)
			if err != nil || info.Mode().Perm() != pair.mode {
				t.Fatalf("nonprivate path mode: %s %v", pair.path, err)
			}
		}
	}
}

func TestInvalidFieldsNeverWritten(t *testing.T) {
	l, home := newTestLogger(t, MaxBytes)
	for _, c := range []struct {
		event    Event
		result   Result
		duration time.Duration
	}{
		{0, ResultOK, 0}, {Event(255), ResultOK, 0},
		{EventAppStartup, 0, 0}, {EventAppStartup, Result(255), 0},
		{EventAppStartup, ResultOK, -1}, {EventAppStartup, ResultOK, 24*time.Hour + 1},
	} {
		if l.Record(c.event, c.result, c.duration) {
			t.Fatalf("accepted invalid fields: %+v", c)
		}
	}
	if got := readEntries(t, logPath(home)); len(got) != 0 {
		t.Fatal("invalid data written")
	}
	if !l.Record(EventAppStartup, ResultOK, 0) {
		t.Fatal("invalid field disabled otherwise usable logger")
	}
}

func TestVersionAllowlist(t *testing.T) {
	for _, version := range []string{"0.2.12-go-dev", "0.2.12-go-test.20261025.3", "0.2.12-go-test.20261025.4", "v1.2.3", "1.2.3-rc.1"} {
		t.Run(version, func(t *testing.T) {
			home := t.TempDir()
			l, err := Open(home, version)
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			if !l.Record(EventAppStartup, ResultOK, 0) {
				t.Fatal("record failed")
			}
			if readEntries(t, logPath(home))[0].Version != version {
				t.Fatal("safe release version discarded")
			}
		})
	}
	for i, version := range []string{"", "/Users/private/workspace", "1.2.3-my-private-project", "1.2.3+secretkey", "auth: Bearer secret", "1.2.3\nsecret", strings.Repeat("1", 1000)} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			home := t.TempDir()
			l, err := Open(home, version)
			if err != nil {
				t.Fatal(err)
			}
			defer l.Close()
			l.Record(EventAppStartup, ResultOK, 0)
			if readEntries(t, logPath(home))[0].Version != "unknown" {
				t.Fatal("untrusted version retained")
			}
		})
	}
}

func TestRotationBoundedAndKeepsNewestEntries(t *testing.T) {
	const limit = 1024
	l, home := newTestLogger(t, limit)
	for i := range 200 {
		if !l.Record(EventWorkspaceSelect, ResultOK, time.Duration(i)*time.Millisecond) {
			t.Fatal("record failed")
		}
	}
	dir := filepath.Dir(logPath(home))
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want bounded current+previous", len(files))
	}
	for _, name := range []string{previousName, currentName} {
		p := filepath.Join(dir, name)
		info, err := os.Stat(p)
		if err != nil || info.Size() > limit {
			t.Fatalf("oversized log %s: %v", name, err)
		}
		if len(readEntries(t, p)) == 0 {
			t.Fatal("empty rotation")
		}
	}
	current := readEntries(t, logPath(home))
	previous := readEntries(t, filepath.Join(dir, previousName))
	if current[len(current)-1].DurationMS != 199 || previous[len(previous)-1].DurationMS+1 != current[0].DurationMS {
		t.Fatal("rotation did not preserve latest continuous stage history")
	}
}

func TestConcurrentWritesAndClose(t *testing.T) {
	l, home := newTestLogger(t, MaxBytes)
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			for range 30 {
				if !l.Record(EventWorkspaceSession, ResultOK, time.Millisecond) {
					t.Error("concurrent record failed")
				}
			}
		})
	}
	wg.Wait()
	if got := readEntries(t, logPath(home)); len(got) != 600 {
		t.Fatalf("lost or interleaved writes: %d", len(got))
	}
	for range 20 {
		wg.Go(func() {
			_ = l.Record(EventAppShutdown, ResultStarted, 0)
			if err := l.Close(); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if l.Record(EventAppShutdown, ResultOK, 0) {
		t.Fatal("record after close succeeded")
	}
}

func TestNilZeroAndFailedLoggerSafe(t *testing.T) {
	var nilLogger *Logger
	var zero Logger
	for _, l := range []*Logger{nilLogger, &zero} {
		if l.Record(EventAppStartup, ResultOK, 0) {
			t.Fatal("disabled logger wrote")
		}
		if err := l.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for _, home := range []string{"", "relative-private-name"} {
		l, err := Open(home, "0.2.12-go-dev")
		if l != nil || !errors.Is(err, ErrHome) || strings.Contains(err.Error(), "relative-private-name") {
			t.Fatal("unsafe home/error")
		}
	}
	missing := filepath.Join(t.TempDir(), "private-missing-home")
	l, err := Open(missing, "0.2.12-go-dev")
	if l != nil || !errors.Is(err, ErrUnavailable) || strings.Contains(err.Error(), missing) {
		t.Fatal("missing home or error leak")
	}
	l, _ = newTestLogger(t, MaxBytes)
	if err := l.file.Close(); err != nil {
		t.Fatal(err)
	}
	if l.Record(EventAppStartup, ResultOK, 0) || l.Record(EventAppStartup, ResultOK, 0) {
		t.Fatal("failed file should disable logging")
	}
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSymlinksRejectedAtEveryOwnedPath(t *testing.T) {
	for _, component := range []string{".orbit-pi", "logs", currentName, previousName} {
		t.Run(component, func(t *testing.T) {
			home := t.TempDir()
			outside := t.TempDir()
			payload := []byte("private external data must stay unchanged")
			target := outside
			link := filepath.Join(home, ".orbit-pi")
			if component != ".orbit-pi" {
				if err := os.Mkdir(link, 0700); err != nil {
					t.Fatal(err)
				}
				link = filepath.Join(link, "logs")
			}
			if component == currentName || component == previousName {
				if err := os.Mkdir(link, 0700); err != nil {
					t.Fatal(err)
				}
				link = filepath.Join(link, component)
				target = filepath.Join(outside, "do-not-modify")
				if err := os.WriteFile(target, payload, 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(target, link); err != nil {
				if runtime.GOOS == "windows" {
					t.Skip("symlinks unavailable")
				}
				t.Fatal(err)
			}
			if l, err := Open(home, "0.2.12-go-dev"); err == nil || l != nil {
				t.Fatal("symlink accepted")
			}
			if component == currentName || component == previousName {
				got, err := os.ReadFile(target)
				if err != nil || string(got) != string(payload) {
					t.Fatal("symlink destination modified")
				}
			} else {
				files, err := os.ReadDir(outside)
				if err != nil || len(files) != 0 {
					t.Fatal("wrote through directory symlink")
				}
			}
		})
	}
}

func TestSymlinkIntroducedAfterOpenDisablesWrites(t *testing.T) {
	l, home := newTestLogger(t, MaxBytes)
	if !l.Record(EventAppStartup, ResultOK, 0) {
		t.Fatal("initial write failed")
	}
	path := logPath(home)
	target := filepath.Join(filepath.Dir(path), "private-data")
	if err := os.WriteFile(target, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("private-data", path); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("symlinks unavailable")
		}
		t.Fatal(err)
	}
	if l.Record(EventWorkspaceBrowse, ResultStarted, 0) {
		t.Fatal("write after replacement succeeded")
	}
	b, err := os.ReadFile(target)
	if err != nil || string(b) != "secret" {
		t.Fatal("symlink target changed")
	}
}

func TestRejectsPublicLogPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix mode bits")
	}
	for _, component := range []string{"logs", currentName, previousName} {
		t.Run(component, func(t *testing.T) {
			l, home := newTestLogger(t, MaxBytes)
			l.Close()
			path := filepath.Join(home, ".orbit-pi", "logs")
			mode := os.FileMode(0755)
			if component != "logs" {
				path = filepath.Join(path, component)
				if err := os.WriteFile(path, nil, 0600); err != nil {
					t.Fatal(err)
				}
				mode = 0644
			}
			if err := os.Chmod(path, mode); err != nil {
				t.Fatal(err)
			}
			if l, err := Open(home, "0.2.12-go-dev"); err == nil || l != nil {
				t.Fatal("public log path accepted")
			}
		})
	}
}

func TestOversizedExistingFilesBecomeBounded(t *testing.T) {
	l, home := newTestLogger(t, 1024)
	l.Close()
	for _, name := range []string{currentName, previousName} {
		if err := os.WriteFile(filepath.Join(filepath.Dir(logPath(home)), name), make([]byte, 2048), 0600); err != nil {
			t.Fatal(err)
		}
	}
	l, err := open(home, "0.2.12-go-dev", 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if !l.Record(EventAppStartup, ResultOK, 0) {
		t.Fatal("record failed")
	}
	files, err := os.ReadDir(filepath.Dir(logPath(home)))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		info, err := f.Info()
		if err != nil || info.Size() > 1024 {
			t.Fatal("oversized file retained")
		}
	}
}

func TestRestartAppendsWithNewRandomRunID(t *testing.T) {
	l, home := newTestLogger(t, MaxBytes)
	l.Record(EventAppStartup, ResultStarted, 0)
	l.Close()
	l, err := Open(home, "0.2.12-go-dev")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	l.Record(EventAppStartup, ResultStarted, 0)
	got := readEntries(t, logPath(home))
	if len(got) != 2 || got[0].RunID == got[1].RunID {
		t.Fatal("restart lost prior stage or reused run identity")
	}
}

func TestExternalGrowthDisablesInsteadOfRotatingOversizedData(t *testing.T) {
	l, home := newTestLogger(t, 1024)
	if err := os.WriteFile(logPath(home), make([]byte, 2048), 0600); err != nil {
		t.Fatal(err)
	}
	if l.Record(EventAppStartup, ResultOK, 0) {
		t.Fatal("external oversized data accepted")
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(logPath(home)), previousName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("oversized data rotated into backup")
	}
}
