package workbench

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type usageFailAfterReader struct{ io.Reader }

func (r usageFailAfterReader) Read(p []byte) (int, error) {
	n, e := r.Reader.Read(p)
	if e == io.EOF {
		return n, errors.New("fixture read failure")
	}
	return n, e
}

func TestUsageUnreadableSourceHeaderAndReadErrorSemantics(t *testing.T) {
	header := `{"type":"session","id":"s","cwd":"/fixture","timestamp":"2026-10-25T12:00:00Z"}` + "\n"
	for _, tc := range []struct {
		name, body       string
		fail, unreadable bool
		sessions         int
	}{
		{"empty", "", false, true, 0},
		{"valid-array-before-header", "[]\n" + header, false, true, 0},
		{"valid-scalar-before-header", "true\n" + header, false, true, 0},
		{"invalid-utf8-before-header", string([]byte{0xff}) + "\n" + header, false, true, 0},
		{"invalid-utf8-after-header", header + string([]byte{0xff}) + "\n", false, false, 1},
		{"invalid-only", "broken\n", false, true, 0},
		{"wrong-first-valid", `{"type":"message"}` + "\n" + header, false, true, 0},
		{"malformed-before-header", "broken\n" + header, false, false, 1},
		{"empty-id", `{"type":"session","id":""}` + "\n", false, false, 0},
		{"failure-before-header", "", true, true, 0},
		{"failure-after-header", header, true, false, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reader io.Reader = strings.NewReader(tc.body)
			if tc.fail {
				reader = usageFailAfterReader{reader}
			}
			index, err := readUsageStream(context.Background(), "/fixture/session.jsonl", reader)
			if (err != nil) != tc.unreadable || len(index.Sessions) != tc.sessions {
				t.Fatalf("unreadable=%v sessions=%d: %v", err != nil, len(index.Sessions), err)
			}
		})
	}
}
func TestUsageUnreadableCountStableAcrossRepeatRepairRemoval(t *testing.T) {
	root := t.TempDir()
	group := filepath.Join(root, "workspace")
	if err := os.Mkdir(group, 0755); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(group, "bad.jsonl")
	if err := os.WriteFile(bad, []byte("broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	// A directory with a JSONL suffix is enumerable but cannot be read as a file.
	// This exercises an actual OS reader failure without relying on root/chmod.
	unreadableDir := filepath.Join(group, "read-error.jsonl")
	if err := os.Mkdir(unreadableDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(group, "empty-id.jsonl"), []byte(`{"type":"session","id":""}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	scanner := NewUsageScanner(root)
	for i := 0; i < 2; i++ {
		index, err := scanner.Scan(context.Background())
		if err != nil || index.UnreadableFiles != 2 || len(index.Sessions) != 0 {
			t.Fatalf("scan%d: %+v %v", i, index, err)
		}
	}
	if err := os.WriteFile(bad, []byte(`{"type":"session","id":"fixed","cwd":"/fixture"}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	index, err := scanner.Scan(context.Background())
	if err != nil || index.UnreadableFiles != 1 || len(index.Sessions) != 1 {
		t.Fatalf("repair: %+v %v", index, err)
	}
	if err := os.Remove(unreadableDir); err != nil {
		t.Fatal(err)
	}
	index, err = scanner.Scan(context.Background())
	if err != nil || index.UnreadableFiles != 0 {
		t.Fatalf("removal: %+v %v", index, err)
	}
	// Directory enumeration errors/warnings do not claim unreadable sessions.
	file := filepath.Join(t.TempDir(), "not-a-store")
	os.WriteFile(file, []byte("x"), 0600)
	index, err = NewUsageScanner(file).Scan(context.Background())
	if err == nil || index.UnreadableFiles != 0 {
		t.Fatal("root error counted as unreadable file")
	}
}

func TestUsageUnreadableExcludesSkippedSizeAndMetadata(t *testing.T) {
	root := t.TempDir()
	group := filepath.Join(root, "workspace")
	if err := os.Mkdir(group, 0755); err != nil {
		t.Fatal(err)
	}
	large, err := os.Create(filepath.Join(group, "oversized.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate((256 << 20) + 1); err != nil {
		large.Close()
		t.Fatal(err)
	}
	large.Close()
	if err := os.Symlink(filepath.Join(root, "does-not-exist"), filepath.Join(group, "broken.jsonl")); err != nil {
		t.Log("symlink fixture unavailable:", err)
	}
	for attempt := 0; attempt < 2; attempt++ {
		index, err := NewUsageScanner(root).Scan(context.Background())
		if err != nil || index.UnreadableFiles != 0 {
			t.Fatalf("skipped files counted unreadable: %+v %v", index, err)
		}
	}
}
