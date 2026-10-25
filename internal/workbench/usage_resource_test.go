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

func TestUsageSourceLongLineWithinFileBoundAndCachedRepeat(t *testing.T) {
	root := t.TempDir()
	group := filepath.Join(root, "group")
	if err := os.Mkdir(group, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(group, "long.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	header := `{"type":"session","id":"long","cwd":"/fixture","timestamp":"2026-10-01T00:00:00Z"}` + "\n"
	if _, err = f.WriteString(header); err != nil {
		t.Fatal(err)
	}
	// Stream the fixture without holding another 65MiB copy in the test process.
	chunk := strings.Repeat(" ", 1<<20)
	for i := 0; i < 65; i++ {
		if _, err = f.WriteString(chunk); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = f.WriteString(`{"type":"message","timestamp":"2026-10-01T00:00:01Z","message":{"role":"assistant","provider":"p","model":"m","usage":{"input":3}}}` + "\n"); err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Size() <= 64<<20 || stat.Size() > usageMaxFileBytes {
		t.Fatal("not a source-valid formerly truncated line", stat.Size())
	}
	scanner := NewUsageScanner(root)
	for repeat := 0; repeat < 2; repeat++ {
		index, err := scanner.Scan(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(index.Records) != 1 || index.Summary(UsageFilter{}).Tokens.Total != 3 || index.Summary(UsageFilter{}).PricedRequests != 0 || index.UnreadableFiles != 0 {
			t.Fatal("source-valid long line silently truncated", index)
		}
		if repeat == 1 && index.Parsed != 0 {
			t.Fatal("long-line cached repeat reread file", index.Parsed)
		}
	}
}

type usageCountingReader struct {
	reader io.Reader
	reads  int
	cancel context.CancelFunc
}

func (r *usageCountingReader) Read(p []byte) (int, error) {
	r.reads++
	n, err := r.reader.Read(p)
	if r.cancel != nil {
		r.cancel()
	}
	return n, err
}
func TestUsageCancellationStopsBeforeReadAndAfterBufferedRead(t *testing.T) {
	header := `{"type":"session","id":"s"}` + "\n"
	for _, before := range []bool{true, false} {
		ctx, cancel := context.WithCancel(context.Background())
		reader := &usageCountingReader{reader: strings.NewReader(header + strings.Repeat(" ", 128<<10))}
		if before {
			cancel()
		} else {
			reader.cancel = cancel
		}
		index, err := readUsageStream(ctx, "/fixture", reader)
		if !errors.Is(err, context.Canceled) || len(index.Sessions) != 0 {
			t.Fatal("cancellation became successful partial index", before, index, err)
		}
		if (before && reader.reads != 0) || (!before && reader.reads != 1) {
			t.Fatal("canceled stream continued consuming resources", before, reader.reads)
		}
		cancel()
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readUsageFile(ctx, filepath.Join(t.TempDir(), "does-not-exist")); !errors.Is(err, context.Canceled) {
		t.Fatal("pre-canceled file attempted open", err)
	}
	if _, err := NewUsageScanner(filepath.Join(t.TempDir(), "missing")).Scan(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("pre-canceled scanner reported success", err)
	}
}

func TestUsageExportProviderLabelUsesOnlySourceASCIICapitalization(t *testing.T) {
	for _, tc := range []struct{ id, want string }{{"éclair-api", "éclair api"}, {"øpen_ai", "øpen ai"}, {"open_ai", "Open ai"}, {"chatgpt", "ChatGPT"}, {"", "unknown"}} {
		if got := usageProviderLabel(tc.id); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}
