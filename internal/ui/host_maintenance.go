package orbitui

import "time"

// enqueue accepts authoritative host completions without blocking the owner
// thread (Agent.Tick can synchronously emit more than 64 notifications). This
// mailbox is intentionally unbounded; owner drains are bounded for fairness.
// Shutdown closes admission under the same lock and discards queued UI work.
func (x *Controller) enqueue(f func()) {
	if f == nil {
		return
	}
	x.hostMailboxMu.Lock()
	if x.hostMailboxClosed {
		x.hostMailboxMu.Unlock()
		return
	}
	x.hostMailbox = append(x.hostMailbox, f)
	x.hostMailboxMu.Unlock()
	if w := x.window.Load(); w != nil && !w.IsDestroyed() {
		w.Invalidate()
	}
}
func (x *Controller) hostStopped() bool {
	x.hostMailboxMu.Lock()
	closed := x.hostMailboxClosed
	x.hostMailboxMu.Unlock()
	if closed {
		return true
	}
	select {
	case <-x.stop:
		return true
	default:
		return false
	}
}
func (x *Controller) drainHostUpdates() {
	x.hostMailboxMu.Lock()
	n := min(len(x.hostMailbox), 256)
	batch := append([]func(){}, x.hostMailbox[:n]...)
	clear(x.hostMailbox[:n])
	x.hostMailbox = x.hostMailbox[n:]
	if len(x.hostMailbox) == 0 {
		x.hostMailbox = nil
	}
	x.hostMailboxMu.Unlock()
	for _, apply := range batch {
		if x.hostStopped() {
			return
		}
		apply()
	}
	if !x.hostStopped() {
		x.UI.drainUpdatesWhile(func() bool { return !x.hostStopped() })
	}
}

// drainUpdates is owner-thread-only. A snapshot batch ensures callbacks which
// schedule more work cannot monopolize a heartbeat or a rendered frame.
func (a *App) drainUpdates() { a.drainUpdatesWhile(func() bool { return true }) }
func (a *App) drainUpdatesWhile(accept func() bool) {
	n := len(a.updates)
	for range n {
		if !accept() {
			return
		}
		select {
		case apply := <-a.updates:
			if apply != nil {
				apply()
			}
		default:
			return
		}
	}
}
func (x *Controller) tickHost(now time.Time) bool {
	if x.hostStopped() {
		return false
	}
	active := x.Agent.Tick()
	x.maintainHost(now)
	return active
}

// maintainHost is owner-thread-only and performs no rendering. The runtime
// drain remains independently serialized by Manager, as it was before.
func (x *Controller) maintainHost(now time.Time) {
	if x.hostStopped() {
		return
	}
	x.drainHostUpdates()
	if x.hostStopped() {
		return
	}
	x.syncAt(now)
	x.UI.toastStack.Expire(now, x.UI.Prefs.ReduceMotion)
}
