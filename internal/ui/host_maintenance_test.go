package orbitui

import (
	"github.com/egoist/mygo"
	"sync/atomic"
	"testing"
	"time"
)

func TestWindowlessHostDrainsAsyncCompletion(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	applied := 0
	x.UI.updates <- func() { applied++ }
	x.tickWindowless()
	if applied != 1 {
		t.Fatal("windowless heartbeat stranded completed host work", applied)
	}
	x.tickWindowless()
	if applied != 1 {
		t.Fatal("completion consumed more than once")
	}
}

// This serial owner dispatcher tests the real production callback boundary, not
// a background goroutine pretending to own widgets. Native RunOnMain wiring is
// exercised by the hosted native gate, not claimed from this headless executor.
func newHostTestOwner(t *testing.T) func(func()) {
	t.Helper()
	jobs := make(chan func())
	done := make(chan struct{})
	go func() {
		defer close(done)
		for f := range jobs {
			f()
		}
	}()
	t.Cleanup(func() { close(jobs); <-done })
	return func(f func()) {
		finished := make(chan struct{})
		jobs <- func() { defer close(finished); f() }
		<-finished
	}
}
func TestWindowlessOwnerRechecksReopenAndShutdownAfterDispatch(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	applied := 0
	x.enqueue(func() { applied++ })
	x.tickWindowlessOnOwner(func(f func()) {
		x.attachWindow(func() *mygo.Window { return &mygo.Window{} })
		f()
	})
	if applied != 0 {
		t.Fatal("queued hidden callback drained after a new view took ownership")
	}
	x.window.Store(nil)
	x.tickWindowlessOnOwner(func(f func()) { x.Close(); f() })
	if applied != 0 {
		t.Fatal("queued callback mutated closed controller")
	}
	x.enqueue(func() { applied++ })
	if len(x.hostMailbox) != 0 {
		t.Fatal("shutdown still accepted completions")
	}
}
func TestWindowlessMailboxBurstAndNestedCallbacksAreBounded(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	owner := newHostTestOwner(t)
	count := 0
	owner(func() {
		for range 700 {
			x.enqueue(func() { count++ })
		}
	})
	owner(func() {
		x.tickWindowless()
		if count != 256 {
			t.Errorf("unbounded or incomplete first snapshot batch: %d", count)
		}
	})
	owner(func() {
		x.tickWindowless()
		if count != 512 {
			t.Errorf("wrong second batch: %d", count)
		}
	})
	owner(func() {
		x.tickWindowless()
		if count != 700 {
			t.Errorf("completion dropped: %d", count)
		}
	})
	owner(func() {
		x.enqueue(func() { count++; x.enqueue(func() { count++ }) })
		x.tickWindowless()
		if count != 701 {
			t.Errorf("nested callback ran in same batch: %d", count)
		}
	})
	owner(func() {
		x.tickWindowless()
		if count != 702 {
			t.Errorf("nested callback lost: %d", count)
		}
	})
}
func TestWindowlessCompletionCanCloseWithoutRunningLaterCallbacks(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	x.enqueue(x.Close)
	x.enqueue(func() { t.Error("mailbox callback ran after Close") })
	x.UI.updates <- func() { t.Error("async callback ran after Close") }
	x.tickWindowless()
}
func TestWindowlessHostAcknowledgesActualMCPReload(t *testing.T) {
	x := newConfigApplyFixture(t)
	before := x.Agent.Snapshot().RuntimeID
	x.MCP.Options.OnApply()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		x.tickWindowless()
		if s := x.Agent.Snapshot(); s.Connected && s.RuntimeID != before && x.MCP.Status == "MCP changes applied to Pi" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("windowless host did not finish/acknowledge actual reload", x.MCP.Status)
}

// Direct fixture calls have a single owner goroutine. Production never uses
// this inline adapter; startHeartbeat always supplies mygo.RunOnMain.
func (x *Controller) tickWindowless() bool { return x.tickWindowlessOnOwner(func(f func()) { f() }) }

func TestWindowlessProductionTimerUsesOwnerDispatcher(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	serial := newHostTestOwner(t)
	var onOwner atomic.Bool
	owner := func(f func()) { serial(func() { onOwner.Store(true); defer onOwner.Store(false); f() }) }
	applied := make(chan bool, 1)
	x.enqueue(func() { applied <- onOwner.Load() })
	x.startHeartbeatOnOwner(owner)
	select {
	case correct := <-applied:
		if !correct {
			t.Fatal("host completion ran on timer worker")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timer failed to drain host completion")
	}
	owner(x.Close)
	select {
	case <-x.heartbeatDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timer failed to stop after owner shutdown")
	}
}
func TestWindowlessHostExpiresToastWithoutRendering(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	defer x.Close()
	x.UI.Notify("synthetic pending notification", false)
	if len(x.UI.toastStack.Items()) == 0 {
		t.Fatal("fixture missing toast")
	}
	x.maintainHost(time.Now().Add(time.Hour))
	if len(x.UI.toastStack.Items()) != 0 {
		t.Fatal("expired toast remained without rendering")
	}
}
