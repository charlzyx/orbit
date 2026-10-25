//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package workbench

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMetadataProbeCancellationStopsDescendants(t *testing.T) {
	s := testService(t)
	dir := t.TempDir()
	marker := filepath.Join(dir, "late-write")
	started := filepath.Join(dir, "started")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, e := s.resourceProbe(ctx, dir, "sh", "-c", "(sleep 0.3; printf late > '"+marker+"') & printf started > '"+started+"'; wait")
		done <- e
	}()
	deadline := time.After(2 * time.Second)
	for {
		if _, e := os.Stat(started); e == nil {
			break
		}
		select {
		case <-deadline:
			t.Fatal("fixture did not start")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("cancelled probe succeeded")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("probe retained descendant pipes")
	}
	time.Sleep(400 * time.Millisecond)
	if _, e := os.Stat(marker); !os.IsNotExist(e) {
		t.Fatal("descendant continued after cancellation")
	}
}
