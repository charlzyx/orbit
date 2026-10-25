package native

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func tempWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestTempPrivacyScopeAndAllowlist(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, AppTempDirectory)
	tempWrite(t, filepath.Join(app, "bug-report-1", "a.png"), "12345")
	tempWrite(t, filepath.Join(app, "update-4294967295-old"), "up")
	tempWrite(t, filepath.Join(root, "orbit-legacy.log"), "old")
	tempWrite(t, filepath.Join(root, "pi-logo.png"), "foreign")
	tempWrite(t, filepath.Join(root, "other.txt"), "foreign")
	now := time.Now()
	old := now.Add(-2 * time.Hour)
	for _, prefix := range piScratchPrefixes {
		file := filepath.Join(root, prefix+"old")
		tempWrite(t, file, "pi")
		os.Chtimes(file, old, old)
	}
	files := ScanTempFilesIn(root, now)
	if len(files.Entries) != 3 || files.TotalBytes() != 10 || len(files.AIEntries) != len(piScratchPrefixes) {
		t.Fatalf("unexpected scan: %#v", files)
	}
	out := cleanTempFilesAt(files, false, now)
	if out.Removed != 3 || out.Bytes != 10 || len(out.Failed) != 0 {
		t.Fatal(out)
	}
	out = cleanTempFilesAt(files, true, now)
	if out.Removed != len(piScratchPrefixes) || len(out.SkippedActive) != 0 {
		t.Fatal(out)
	}
	for _, name := range []string{"pi-logo.png", "other.txt", AppTempDirectory} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			t.Fatal("removed outside scope:", name)
		}
	}
}
func TestTempPrivacyLivePIDRecentAndStaleSnapshot(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, AppTempDirectory)
	live := filepath.Join(app, fmt.Sprintf("login-%d.command", os.Getpid()))
	legacy := filepath.Join(root, fmt.Sprintf("orbit-update-%d-nonce", os.Getpid()))
	tempWrite(t, live, "live")
	tempWrite(t, legacy, "live")
	oldFile := filepath.Join(root, "pi-output-old")
	recentFile := filepath.Join(root, "pi-bash-recent")
	tempWrite(t, oldFile, "old")
	tempWrite(t, recentFile, "recent")
	now := time.Now()
	old := now.Add(-2 * time.Hour)
	os.Chtimes(oldFile, old, old)
	files := ScanTempFilesIn(root, now)
	out := cleanTempFilesAt(files, false, now)
	if out.Removed != 0 || len(out.SkippedActive) != 2 {
		t.Fatal(out)
	}
	// A file that became active after the page opened must not be removed.
	os.Chtimes(oldFile, now, now)
	out = cleanTempFilesAt(files, true, now)
	if out.Removed != 0 || len(out.SkippedActive) != 2 {
		t.Fatal(out)
	}
}
func TestTempPrivacyNoSymlinkFollowOrForgedPaths(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions")
	}
	root := t.TempDir()
	foreign := t.TempDir()
	secret := filepath.Join(foreign, "keep")
	tempWrite(t, secret, "keep")
	app := filepath.Join(root, AppTempDirectory)
	os.Mkdir(app, 0700)
	link := filepath.Join(app, "link")
	if err := os.Symlink(foreign, link); err != nil {
		t.Fatal(err)
	}
	files := ScanTempFilesIn(root, time.Now())
	files.Entries = append(files.Entries, TempEntry{Path: secret})
	out := CleanTempFiles(files, false)
	if out.Removed != 1 || len(out.Failed) != 1 {
		t.Fatal(out)
	}
	if _, err := os.Stat(secret); err != nil {
		t.Fatal("followed symlink or forged path")
	}
	os.Remove(app)
	if err := os.Symlink(foreign, app); err != nil {
		t.Fatal(err)
	}
	if got := ScanTempFilesIn(root, time.Now()); len(got.Entries) != 0 {
		t.Fatal("followed dedicated directory symlink", got)
	}
}
func TestTempPrivacyVanishedAndNestedRecent(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, AppTempDirectory)
	gone := filepath.Join(app, "gone")
	tempWrite(t, gone, "one")
	now := time.Now()
	files := ScanTempFilesIn(root, now)
	os.Remove(gone)
	if out := cleanTempFilesAt(files, false, now); out.Removed != 1 || out.Bytes != 0 {
		t.Fatal(out)
	}
	dir := filepath.Join(root, "pi-share-old")
	tempWrite(t, filepath.Join(dir, "still-writing"), "new")
	old := now.Add(-2 * time.Hour)
	os.Chtimes(dir, old, old)
	files = ScanTempFilesIn(root, now)
	if len(files.AIEntries) != 1 || !files.AIEntries[0].Active {
		t.Fatal("ignored recent nested activity", files)
	}
}
func TestEmbeddedTempPIDs(t *testing.T) {
	for name, want := range map[string]int{"update-42-9": 42, "login-42.command": 42, "orbit-update-42-9": 42, "orbit-42.command": 42} {
		if got, ok := embeddedPID(name); !ok || got != want {
			t.Fatal(name, got, ok)
		}
	}
	for _, name := range []string{"bug-report-42", "other-42.command", "update-0-x", "update-4294967295-x"} {
		if _, ok := embeddedPID(name); ok {
			t.Fatal(name)
		}
	}
}
