//go:build !windows

package native

import (
	"context"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRuntimeProbeCancellationDoesNotHangOnDescendantPipes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	started := time.Now()
	out, e := runRuntimeCommand(ctx, "/bin/sh", []string{"-c", "sleep 60 & printf '%s\\n' \"$!\"; wait"}, false)
	if e == nil {
		t.Fatal("cancelled command succeeded")
	}
	if time.Since(started) > 3*time.Second {
		t.Fatal("descendant pipe delayed cancellation")
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(out.Stdout)))
	if pid <= 0 {
		t.Fatalf("missing disposable child PID: %q", out.Stdout)
	}
	// Linux may briefly retain a terminated orphan as a zombie. It is no longer
	// running or holding a pipe. Only the PID created by this fixture is inspected.
	if err := syscall.Kill(pid, 0); err == nil {
		raw, readErr := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
		if readErr == nil && !strings.Contains(string(raw), ") Z ") {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatal("owned descendant survived cancellation")
		}
	}
}
