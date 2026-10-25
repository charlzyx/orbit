package orbitui

import (
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

func (x *Controller) startHeartbeat() {
	x.heartbeatDone = make(chan struct{})
	go func() {
		defer close(x.heartbeatDone)
		runHeartbeat(x.stop, x.tickWindowless)
	}()
}

func (x *Controller) tickWindowless() bool {
	if !x.parkWindowlessTerminal() {
		return false
	}
	// Manager serializes its own state. Do not hold the view handoff lock
	// across Tick: native notification callbacks can wait on the UI thread.
	return x.Agent.Tick()
}

func (x *Controller) parkWindowlessTerminal() bool {
	x.heartbeatMu.Lock()
	defer x.heartbeatMu.Unlock()
	select {
	case <-x.stop:
		return false
	default:
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
