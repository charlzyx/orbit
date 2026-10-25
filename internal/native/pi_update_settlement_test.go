package native

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPiUpdateCompletionSkippedAndCanceledBeforeIO(t *testing.T) {
	for _, env := range []string{"PI_BIN", "PI_OFFLINE", "canceled"} {
		t.Run(env, func(t *testing.T) {
			t.Setenv("PI_BIN", "")
			t.Setenv("PI_OFFLINE", "")
			if env != "canceled" {
				t.Setenv(env, "fixture-skip")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			done := CheckPiUpdateOnLaunch(ctx, PiLaunchUpdateOptions{OnEvent: func(PiUpdateEvent) { t.Error("skipped checker emitted event") }})
			if done == nil {
				t.Fatal("missing completion handle")
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("early return did not settle")
			}
		})
	}
}
func TestPiUpdateWorkerCompletionRequiresActualReturn(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	done := startPiUpdateWorker(func() { close(started); <-release })
	<-started
	select {
	case <-done:
		t.Fatal("worker settled before return")
	default:
	}
	unblock()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker return failed to settle")
	}
	normal := startPiUpdateWorker(func() {})
	select {
	case <-normal:
	case <-time.After(time.Second):
		t.Fatal("normal completion did not settle")
	}
}
