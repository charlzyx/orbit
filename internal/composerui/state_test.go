package composerui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func waitScan(t *testing.T, c *Controller) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c.Poll()
		if c.Snapshot().Files.State != Loading {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("scan did not finish")
}
func file(t *testing.T, root, path string) {
	t.Helper()
	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestActualWorkspaceEnumerationBoundsAndFaults(t *testing.T) {
	root := t.TempDir()
	for _, p := range []string{"src/main.go", "README.md", ".env", ".hidden/a", "node_modules/pkg/a", "target/debug/a", "dist/a", "build/a", "vendor/a", "__pycache__/a", "visible/.hidden"} {
		file(t, root, p)
	}
	// Source treats a symlink as an entry but does not descend into the target.
	if err := os.Symlink(filepath.Join(root, "src"), filepath.Join(root, "link")); err != nil {
		t.Log("symlink unavailable", err)
	}
	r := ScanWorkspace(context.Background(), root)
	want := []string{"README.md", "link", "src/main.go"}
	if !reflect.DeepEqual(r.Files, want) {
		t.Fatal(r.Files)
	}
	p := ""
	for i := 1; i <= 13; i++ {
		p = filepath.Join(p, fmt.Sprintf("d%d", i))
		file(t, root, filepath.Join(p, "f"))
	}
	r = ScanWorkspace(context.Background(), root)
	for _, p := range r.Files {
		if filepath.Base(p) == "f" && len(splitPath(p)) > 13 {
			t.Fatal("depth exceeded", p)
		}
	}
	r = ScanWorkspace(context.Background(), filepath.Join(root, "missing"))
	if r.Err == nil || r.Skipped != 1 || len(r.Files) != 0 {
		t.Fatal(r)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = ScanWorkspace(ctx, root)
	if !errors.Is(r.Err, context.Canceled) {
		t.Fatal(r)
	}
}
func splitPath(p string) []string {
	out := []string{}
	for p != "." && p != "" {
		out = append(out, filepath.Base(p))
		p = filepath.Dir(p)
	}
	return out
}
func TestActualWorkspaceFileCap(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < MaxFiles+4; i++ {
		file(t, root, fmt.Sprintf("f%04d", i))
	}
	r := ScanWorkspace(context.Background(), root)
	if len(r.Files) != MaxFiles || !r.Capped || r.Err != nil {
		t.Fatal(len(r.Files), r.Capped, r.Err)
	}
}
func TestStateFilteringDismissalKeyboardCommit(t *testing.T) {
	c := New(nil)
	defer c.Close()
	c.SetCommands(raws(`{"name":"review"}`, `{"name":"reset"}`, `{"name":"rename"}`))
	c.Sync("/re", 3, true, false)
	if !c.IsOpen() {
		t.Fatal(c.Snapshot())
	}
	c.Move(-1)
	if c.Snapshot().Highlighted != 2 {
		t.Fatal(c.Snapshot())
	}
	c.Move(1)
	if c.Snapshot().Highlighted != 0 {
		t.Fatal(c.Snapshot())
	}
	c.Move(1)
	c.Sync("/rev", 4, true, false)
	if c.Snapshot().Highlighted != 0 {
		t.Fatal("clamp")
	}
	if !c.Dismiss() {
		t.Fatal("dismiss")
	}
	c.Sync("/rev", 4, true, false)
	if c.IsOpen() {
		t.Fatal("reopened unchanged")
	}
	c.Sync("/re", 3, true, false)
	if !c.IsOpen() {
		t.Fatal("not rearmed")
	}
	c.Sync("/remainder arguments", 3, true, false)
	next, caret, ok := c.Commit(0)
	if !ok || next != "/review  arguments" || caret != 8 {
		t.Fatal(next, caret, ok)
	}
	c.Sync("/re", 3, false, false)
	if _, _, ok := c.Commit(0); ok {
		t.Fatal("unfocused commit")
	}
	c.Sync("/re", 3, true, true)
	if _, _, ok := c.Commit(0); ok {
		t.Fatal("IME commit")
	}
	c.Sync("", 0, true, false)
	if c.Snapshot().Highlighted != 0 || c.IsOpen() {
		t.Fatal(c.Snapshot())
	}
}
func TestFileCompletionPreservesSuffixUnicodeAndCache(t *testing.T) {
	root := t.TempDir()
	file(t, root, "文件.go")
	var scans atomic.Int32
	c := New(func(ctx context.Context, root string) ScanResult { scans.Add(1); return ScanWorkspace(ctx, root) })
	defer c.Close()
	c.SetContext(root, "", "session")
	c.Sync("你好 @文件suffix", 6, true, false)
	waitScan(t, c)
	next, caret, ok := c.Commit(0)
	if !ok || next != "你好 @文件.go suffix" || caret != 10 {
		t.Fatal(next, caret, ok)
	}
	file(t, root, "later.go")
	c.Sync("@", 1, true, false)
	if len(c.Snapshot().Entries) != 1 || scans.Load() != 1 {
		t.Fatal("cache changed")
	}
	c.RefreshFiles()
	waitScan(t, c)
	if len(c.Snapshot().Entries) != 2 || scans.Load() != 2 {
		t.Fatal(c.Snapshot())
	}
}
func TestAsyncCancellationRejectsStaleWorkspaceAndSessionCatalog(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	c := New(func(ctx context.Context, root string) ScanResult {
		started <- root
		if root == "old" {
			<-release
		}
		return ScanResult{Files: []string{root + ".go"}}
	})
	defer c.Close()
	c.SetContext("old", "", "old-session")
	c.SetCommands(raws(`{"name":"old"}`))
	c.Sync("@", 1, true, false)
	<-started
	c.SetContext("new", "", "new-session")
	if c.SetCatalog("old-session", raws(`{"name":"stale"}`)) {
		t.Fatal("accepted stale catalog")
	}
	c.Sync("@", 1, true, false)
	<-started
	waitScan(t, c)
	close(release)
	time.Sleep(3 * time.Millisecond)
	c.Poll()
	if got := c.Snapshot().Entries; len(got) != 1 || got[0].Path != "new.go" {
		t.Fatal(got)
	}
	c.Sync("/", 1, true, false)
	if c.IsOpen() {
		t.Fatal("old commands leaked")
	}
	c.SetCatalog("new-session", raws(`{"name":"current"}`))
	if !c.IsOpen() {
		t.Fatal("fresh command did not update open trigger")
	}
	c.Close()
	if c.IsOpen() || c.SetCatalog("new-session", raws(`{"name":"late"}`)) {
		t.Fatal("closed controller changed")
	}
}
func TestScanErrorPartialEmptyAndExplicitRetry(t *testing.T) {
	var calls atomic.Int32
	c := New(func(context.Context, string) ScanResult {
		if calls.Add(1) == 1 {
			return ScanResult{Err: errors.New("permission fixture"), Files: []string{"readable.go"}, Skipped: 1}
		}
		return ScanResult{Files: []string{}}
	})
	defer c.Close()
	c.SetContext("fixture", "", "s")
	c.Sync("@", 1, true, false)
	waitScan(t, c)
	if s := c.Snapshot(); s.Files.State != Failed || s.Files.Error == "" || !s.Open || s.Files.Skipped != 1 {
		t.Fatal(s)
	}
	c.Sync("@", 1, true, false)
	if calls.Load() != 1 {
		t.Fatal("error was automatically retried every frame")
	}
	c.RefreshFiles()
	waitScan(t, c)
	if s := c.Snapshot(); s.Files.State != Ready || s.Open || len(s.Entries) != 0 {
		t.Fatal(s)
	}
}
