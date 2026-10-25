package orbitui

import (
	"reflect"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/egoist/mygo"
)

// These exercise the production timer loop with Go's virtual clock, not a
// duration selector or wall-clock tolerance. Source: 299a32b main_window.rs:136.
func TestHeartbeatSourceInitialIdleActiveIdleCadence(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stop, done := make(chan struct{}), make(chan struct{})
		start := time.Now()
		var ticks []time.Duration
		var mu sync.Mutex
		readTicks := func() []time.Duration { mu.Lock(); defer mu.Unlock(); return append([]time.Duration(nil), ticks...) }
		go func() {
			defer close(done)
			runHeartbeat(stop, func() bool {
				mu.Lock()
				defer mu.Unlock()
				ticks = append(ticks, time.Since(start))
				return len(ticks) <= 2
			})
		}()
		synctest.Wait()
		time.Sleep(89 * time.Millisecond)
		synctest.Wait()
		if ticks := readTicks(); len(ticks) != 0 {
			t.Fatal("source starts idle, without an immediate tick", ticks)
		}
		time.Sleep(124 * time.Millisecond)
		synctest.Wait()
		want := []time.Duration{90 * time.Millisecond, 106 * time.Millisecond, 122 * time.Millisecond, 212 * time.Millisecond}
		if ticks := readTicks(); !reflect.DeepEqual(ticks, want) {
			t.Fatalf("actual heartbeat ticks = %v, want %v", ticks, want)
		}
		close(stop)
		<-done
		time.Sleep(time.Second)
		synctest.Wait()
		if ticks := readTicks(); !reflect.DeepEqual(ticks, want) {
			t.Fatal("heartbeat survived shutdown", ticks)
		}
	})
}

func TestHeartbeatSourceWaitsAfterTickAndStopsDuringDrain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stop, done := make(chan struct{}), make(chan struct{})
		start := time.Now()
		var ticks []time.Duration
		go func() {
			defer close(done)
			runHeartbeat(stop, func() bool {
				ticks = append(ticks, time.Since(start))
				if len(ticks) == 2 {
					close(stop)
				}
				time.Sleep(7 * time.Millisecond)
				return true
			})
		}()
		<-done
		want := []time.Duration{90 * time.Millisecond, 113 * time.Millisecond}
		if !reflect.DeepEqual(ticks, want) {
			t.Fatalf("loop used fixed-rate scheduling or ticked after stop: %v", ticks)
		}
	})
}

func TestHeartbeatStoppedBeforeStartDoesNotDrain(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		stop := make(chan struct{})
		close(stop)
		runHeartbeat(stop, func() bool { t.Fatal("drained after stop"); return false })
	})
}

// No native window is created. A zero-value Window is only a safe live-window
// identity (IsDestroyed is atomic); this does not test macOS Dock activation.
func TestHeartbeatWindowOwnershipAndReopenBoundary(t *testing.T) {
	x := &Controller{UI: &App{}, stop: make(chan struct{})}
	first := &mygo.Window{}
	x.attachWindow(func() *mygo.Window {
		// Deliberately nil services: calling either would panic. The opening
		// guard must suppress the old windowless route during first render.
		if x.tickWindowless() {
			t.Fatal("windowless drain ran during native creation")
		}
		return first
	})
	if x.window.Load() != first || x.UI.Window != first || x.UI.window.Load() != first || x.openingWindow {
		t.Fatal("view attachment did not publish one window or release its gate")
	}
	if x.tickWindowless() {
		t.Fatal("live window did not own its heartbeat")
	}
	// Simulate a destroyed view without calling a native OS backend.
	x.window.Store(nil)
	second := &mygo.Window{}
	x.attachWindow(func() *mygo.Window {
		if x.tickWindowless() {
			t.Fatal("reopen allowed the hidden drain to overlap first render")
		}
		return second
	})
	if x.UI.Window != second || x.UI.window.Load() != second || x.window.Load() != second {
		t.Fatal("reopen did not replace only the view")
	}
	close(x.stop)
	x.window.Store(nil)
	if x.tickWindowless() {
		t.Fatal("shutdown allowed another drain")
	}
}

func TestHeartbeatControllerCloseCancelsWorker(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	x.startHeartbeat()
	x.Close()
	x.Close() // Idempotent: no second close or dangling heartbeat.
	select {
	case <-x.heartbeatDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Close did not cancel windowless heartbeat")
	}
}

// An in-flight native notification may need the main thread to return from
// Close. Preserve that existing callback semantics; never join it on main.
func TestHeartbeatCloseDoesNotJoinBlockedCallback(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	entered, release := make(chan struct{}), make(chan struct{})
	x.heartbeatDone = make(chan struct{})
	go func() {
		defer close(x.heartbeatDone)
		runHeartbeat(x.stop, func() bool { close(entered); <-release; return true })
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("heartbeat did not enter callback")
	}
	closed := make(chan struct{})
	go func() { x.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		close(release)
		t.Fatal("Close joined a callback that could be waiting for UI return")
	}
	close(release)
	select {
	case <-x.heartbeatDone:
	case <-time.After(2 * time.Second):
		t.Fatal("worker failed to stop after its in-flight callback returned")
	}
}

func TestHeartbeatFailedWindowCreationReleasesRoutingGate(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	x.attachWindow(func() *mygo.Window { return nil })
	if x.openingWindow || !x.parkWindowlessTerminal() {
		t.Fatal("failed creation stranded the windowless drain")
	}
}

func TestHeartbeatConcurrentWindowlessDrainAndViewHandoff(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	manager := x.Agent
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			x.tickWindowless()
		}
	}()
	for range 1000 {
		x.window.Store(nil)
		x.attachWindow(func() *mygo.Window { return &mygo.Window{} })
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("windowless drain deadlocked with a view handoff")
	}
	if x.Agent != manager {
		t.Fatal("reopening replaced the application-owned runtime")
	}
}
