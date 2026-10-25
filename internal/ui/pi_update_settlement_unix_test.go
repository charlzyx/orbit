//go:build !windows

package orbitui

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/native"
)

func TestPiUpdateShutdownSettlesOwnedHarmlessProcessGroup(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("descendant assertion uses Linux procfs; native Mac acceptance remains separate")
	}
	x := piUpdateTargetFixture(t)
	marker := filepath.Join(t.TempDir(), "started")
	outputs := make(chan native.CommandOutput, 1)
	admitted := make(chan bool, 1)
	x.startPiLaunchUpdate(func(ctx context.Context, o native.PiLaunchUpdateOptions) <-chan struct{} {
		done := make(chan struct{})
		go func() {
			defer close(done)
			release, ok := o.BeginUpdate()
			admitted <- ok
			if !ok {
				return
			}
			defer release()
			// Fixed harmless local process only: no update program/script/network.
			out, _ := native.RunInstallerInvocation(ctx, native.Invocation{Program: "/bin/sh", Args: []string{"-c", `sleep 60 & printf '%s\n' "$!"; printf ready > "$1"; wait`, "orbit-disposable-auto-fixture", marker}})
			outputs <- out
		}()
		return done
	})
	if !<-admitted {
		t.Fatal("fixture writer was rejected")
	}
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("fixture process never started")
		}
		time.Sleep(time.Millisecond)
	}
	x.Close()
	select {
	case <-x.piUpdateDone:
	default:
		t.Fatal("Close returned before actual process completion")
	}
	if x.maintenance.Load() {
		t.Fatal("settled worker retained maintenance writer")
	}
	out := <-outputs
	pid, err := strconv.Atoi(strings.TrimSpace(string(out.Stdout)))
	if err != nil || pid <= 0 {
		t.Fatalf("missing owned descendant PID: %q", out.Stdout)
	}
	if err = syscall.Kill(pid, 0); err == nil {
		raw, readErr := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
		if readErr == nil && !strings.Contains(string(raw), ") Z ") {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatal("owned descendant survived cancellation")
		}
	}
}
