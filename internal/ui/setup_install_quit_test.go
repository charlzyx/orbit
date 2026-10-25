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

func TestSetupInstallQuitJoinsActualRunnerNotOnlyCancellation(t *testing.T) {
	x := setupInstallFixture(t)
	started, canceled, release := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	b := fixtureInstallBackend(native.PiInstallation{}, managedSetupFixture(), func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		close(started)
		<-ctx.Done()
		close(canceled)
		<-release
		return native.CommandOutput{}, ctx.Err()
	})
	if err := x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err != nil {
		t.Fatal(err)
	}
	op := x.setupInstall.pending
	<-started
	closed := make(chan struct{})
	go func() { x.Close(); close(closed) }()
	<-canceled
	select {
	case <-closed:
		t.Fatal("Quit returned before actual runner settlement")
	case <-time.After(30 * time.Millisecond):
	}
	unblock()
	select {
	case <-closed:
	case <-time.After(3 * time.Second):
		t.Fatal("Quit waited on stopped owner queue instead of runner")
	}
	select {
	case <-op.done:
	case <-time.After(time.Second):
		t.Fatal("closed-host lease stranded after runner settlement")
	}
}

func TestSetupInstallQuitTimeoutIsDiagnosticNotSettlement(t *testing.T) {
	x := setupInstallFixture(t)
	home := t.TempDir()
	logger, err := diagnosticlog.Open(home, "0.2.12-go-test")
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	x.diagnostics = logger
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	t.Cleanup(unblock)
	b := fixtureInstallBackend(native.PiInstallation{}, managedSetupFixture(), func(ctx context.Context, _ bool) (native.CommandOutput, error) {
		close(started)
		<-release
		return native.CommandOutput{}, ctx.Err()
	})
	if err = x.beginSetupInstall(x.captureSetupInstallIntent(false), b); err != nil {
		t.Fatal(err)
	}
	op := x.setupInstall.pending
	<-started
	if x.waitSetupInstallRunner(10 * time.Millisecond) {
		t.Fatal("cancellation-ignoring runner was reported settled")
	}
	select {
	case <-op.runnerSettled:
		t.Fatal("timeout faked actual runner completion")
	default:
	}
	select {
	case <-op.done:
		t.Fatal("timeout released maintenance lease")
	default:
	}
	raw, err := os.ReadFile(filepath.Join(home, ".orbit-pi", "logs", "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"event":"runtime_install_shutdown"`) || !strings.Contains(string(raw), `"result":"timeout"`) || strings.Contains(string(raw), `"result":"ok"`) {
		t.Fatalf("unknown cleanup was not diagnosed truthfully: %s", raw)
	}
	unblock()
	waitSetupInstallMailbox(t, x)
	x.drainHostUpdates()
	<-op.done
	if x.maintenance.Load() || x.setupInstall.receipt.verified {
		t.Fatal("late canceled return retained lease or reported success")
	}
}
