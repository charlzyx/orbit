package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/diagnosticlog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/native"
)

func TestPiUpdateShutdownJoinsActualWorkerNotCancellation(t *testing.T) {
	x := piUpdateTargetFixture(t)
	started, canceled, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	x.startPiLaunchUpdate(func(ctx context.Context, _ native.PiLaunchUpdateOptions) <-chan struct{} {
		go func() { defer close(done); close(started); <-ctx.Done(); close(canceled); <-release }()
		return done
	})
	<-started
	closed := make(chan struct{})
	go func() { x.Close(); close(closed) }()
	<-canceled
	select {
	case <-closed:
		t.Error("Quit returned before automatic worker actual completion")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("Quit waited for stopped owner acknowledgment")
	}
	select {
	case <-done:
	default:
		t.Fatal("Close returned with worker unsettled")
	}
}

func TestPiUpdateShutdownQueuedReceiptDoesNotWaitForOwner(t *testing.T) {
	x := piUpdateTargetFixture(t)
	done := make(chan struct{})
	close(done)
	x.startPiLaunchUpdate(func(_ context.Context, o native.PiLaunchUpdateOptions) <-chan struct{} {
		o.OnEvent(native.PiUpdateEvent{State: "available", Latest: "0.60.2"})
		return done
	})
	x.Close()
	if len(x.UI.toastStack.Items()) != 0 {
		t.Fatal("closed host applied queued worker receipt")
	}
	x.drainHostUpdates()
	if len(x.UI.toastStack.Items()) != 0 {
		t.Fatal("closed owner queue resurrected receipt")
	}
}

func TestPiUpdateShutdownTimeoutIsUncertainAndSharesDeadline(t *testing.T) {
	x := piUpdateTargetFixture(t)
	home := t.TempDir()
	logger, err := diagnosticlog.Open(home, "0.2.12-go-test")
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	x.diagnostics = logger
	setupDone, autoDone := make(chan struct{}), make(chan struct{})
	canceled := false
	x.setupInstall.pending = &setupInstallOperation{runnerSettled: setupDone, cancel: func() { canceled = true }}
	x.piUpdateDone = autoDone
	defer func() { x.setupInstall.pending = nil; x.piUpdateDone = nil }()
	start := time.Now()
	x.waitRuntimeRunners(100 * time.Millisecond)
	elapsed := time.Since(start)
	if !canceled || elapsed > 180*time.Millisecond {
		t.Fatalf("shared bounded join failed: canceled=%v elapsed=%v", canceled, elapsed)
	}
	select {
	case <-autoDone:
		t.Fatal("timeout fabricated automatic completion")
	default:
	}
	select {
	case <-setupDone:
		t.Fatal("timeout fabricated explicit completion")
	default:
	}
	raw, err := os.ReadFile(filepath.Join(home, ".orbit-pi", "logs", "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"event":"runtime_auto_update_shutdown"`) || !strings.Contains(string(raw), `"result":"timeout"`) || strings.Contains(string(raw), `"result":"ok"`) || strings.Contains(string(raw), home) {
		t.Fatalf("untruthful or private shutdown diagnostic: %s", raw)
	}
	close(setupDone)
	close(autoDone)
	// Already closed wins even after the total budget has expired.
	if !x.waitRuntimeWorker(autoDone, diagnosticlog.EventRuntimeAutoUpdateShutdown, -time.Second) || !x.waitRuntimeWorker(nil, diagnosticlog.EventRuntimeAutoUpdateShutdown, 0) {
		t.Fatal("settled/nil worker falsely timed out")
	}
}

func TestPiUpdateShutdownOnlyOneTrackedCheckPerLaunch(t *testing.T) {
	x := piUpdateTargetFixture(t)
	done := make(chan struct{})
	close(done)
	calls := 0
	check := func(context.Context, native.PiLaunchUpdateOptions) <-chan struct{} { calls++; return done }
	x.startPiLaunchUpdate(check)
	x.startPiLaunchUpdate(check)
	if calls != 1 || x.piUpdateDone != done {
		t.Fatal("later check replaced tracked worker")
	}
}
