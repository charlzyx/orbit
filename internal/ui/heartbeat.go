package orbitui

import (
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// Source 299a32b main_window.rs::spawn_heartbeat starts idle, then waits
// 16/90 ms AFTER each tick (not a fixed-rate ticker). Manager.Tick deliberately
// includes parked busy sessions and startup/awaiting/recovery work, unlike the
// source events.rs::is_active predicate, which only describes the active view.
func heartbeatDelay(active bool) time.Duration {
	if active {
		return 16 * time.Millisecond
	}
	return 90 * time.Millisecond
}

func runHeartbeat(stop <-chan struct{}, tick func() bool) {
	timer := time.NewTimer(heartbeatDelay(false))
	defer timer.Stop()
	for {
		select {
		case <-stop:
			return
		case <-timer.C:
			// If shutdown and the timer are both ready, shutdown owns the tick.
			select {
			case <-stop:
				return
			default:
			}
			timer.Reset(heartbeatDelay(tick()))
		}
	}
}

func (x *Controller) startHeartbeat() { x.startHeartbeatOnOwner(mygo.RunOnMain) }

// owner is synchronous: production uses MyGo's actual UI-thread dispatcher.
func (x *Controller) startHeartbeatOnOwner(owner func(func())) {
	x.heartbeatDone = make(chan struct{})
	go func() {
		defer close(x.heartbeatDone)
		runHeartbeat(x.stop, func() bool { return x.tickWindowlessOnOwner(owner) })
	}()
}

// The runtime drain stays on the timer worker: native notification delivery
// may wait for the UI loop or invoke the unpackaged macOS notification helper.
// Only thread-safe Manager state is touched there. All host/UI maintenance is
// dispatched on the owner, with ownership rechecked after any slow runtime work.
func (x *Controller) tickWindowlessOnOwner(owner func(func())) bool {
	allowed := false
	owner(func() { allowed = x.parkWindowlessTerminal() })
	if !allowed {
		return false
	}
	active := x.Agent.Tick()
	owner(func() {
		if x.parkWindowlessTerminal() {
			x.maintainHost(time.Now())
		}
	})
	return active
}

func (x *Controller) parkWindowlessTerminal() bool {
	x.heartbeatMu.Lock()
	defer x.heartbeatMu.Unlock()
	if x.hostStopped() {
		return false
	}
	if x.openingWindow {
		return false
	}
	if w := x.window.Load(); w != nil && !w.IsDestroyed() {
		return false // The visible view owns its runtime drain and repaint clock.
	}
	x.Terminal.SetVisible(false)
	return true
}

// Finish windowless terminal hiding before constructing a new view, and stop
// another hide while that first view is rendered. Never hold
// heartbeatMu across native construction: it may synchronously render View.
func (x *Controller) attachWindow(create func() *mygo.Window) *mygo.Window {
	x.heartbeatMu.Lock()
	x.openingWindow = true
	x.heartbeatMu.Unlock()
	defer func() {
		x.heartbeatMu.Lock()
		x.openingWindow = false
		x.heartbeatMu.Unlock()
	}()
	w := create()
	x.window.Store(w)
	x.UI.Window = w
	x.UI.window.Store(w)
	return w
}

// Register before App.Run, but never dispatch owner-thread work before the
// native backend has initialized its callback pointers. On Linux Signal can
// otherwise set its signaled latch while its GLib callback is still nil.
// The one host lifetime starts after the initial window setup, not on reopen.
func registerNativeHostReady(register func(func()), initialize, start func()) {
	var once sync.Once
	register(func() { once.Do(func() { initialize(); start() }) })
}
