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

// This invokes only a fixed harmless shell/sleep fixture through the native
// invocation/cancellation API. No installer script or network endpoint is used.
func TestSetupInstallQuitSettlesOwnedNativeProcessGroup(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("descendant verification uses Linux procfs; native macOS acceptance remains separate")
	}
	x := setupInstallFixture(t)
	marker := filepath.Join(t.TempDir(), "started")
	outputs := make(chan native.CommandOutput, 1)
	b := fixtureInstallBackend(native.PiInstallation{}, managedSetupFixture(), func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		out, err := native.RunInstallerInvocation(ctx, native.Invocation{Program: "/bin/sh", Args: []string{"-c", `sleep 60 & printf '%s\n' "$!"; printf ready > "$1"; wait`, "orbit-disposable-fixture", marker}})
		outputs <- out
		return out, err
	})
	if err := x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err != nil {
		t.Fatal(err)
	}
	op := x.setupInstall.pending
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owned fixture never started")
		}
		time.Sleep(time.Millisecond)
	}
	x.Close()
	select {
	case <-op.runnerSettled:
	default:
		t.Fatal("Quit returned with native runner not settled")
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
			t.Fatal("owned descendant survived process-group cancellation")
		}
	}
	select {
	case <-op.done:
	case <-time.After(time.Second):
		t.Fatal("terminated native fixture left maintenance reserved")
	}
	if x.setupInstall.receipt.verified {
		t.Fatal("canceled native fixture reported installation success")
	}
}
