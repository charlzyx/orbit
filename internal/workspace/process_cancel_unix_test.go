//go:build darwin || linux || freebsd || openbsd || netbsd || dragonfly

package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCancelledCommandStopsDescendantWork(t *testing.T) {
	root := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	_, err := commandOutput(ctx, root, "/bin/sh", nil, []string{"-c", "(sleep 0.3; printf orphan > late-write) & wait"})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	time.Sleep(350 * time.Millisecond)
	if _, err := os.Stat(filepath.Join(root, "late-write")); !os.IsNotExist(err) {
		t.Fatal("cancelled command left a live mutating descendant", err)
	}
}
