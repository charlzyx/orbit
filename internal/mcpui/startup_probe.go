package mcpui

import "time"

// ScheduleInitialProbe mirrors Rust's one deferred launch probe. The controller
// calls it once; configuration loading and Pi execution stay in existing workers.
// Workspace/binary changes cancel this old-context intent, not transfer it.
func (s *State) ScheduleInitialProbe(now time.Time) {
	if s.closed {
		return
	}
	s.initialProbeWaiting = true
	s.initialProbeAt = time.Time{}
	if s.loaded {
		s.tickInitialProbe(now)
	}
	s.PrepareQuotaPopup()
}

// tickInitialProbe is owner-thread maintenance, never a timer goroutine. Empty
// configuration consumes the intent without invoking Pi; busy work defers it.
func (s *State) tickInitialProbe(now time.Time) {
	if !s.loaded {
		return
	}
	if s.initialProbeWaiting {
		s.initialProbeWaiting = false
		s.initialProbeAt = now.Add(3 * time.Second)
	}
	if s.initialProbeAt.IsZero() {
		return
	}
	if len(s.Servers) == 0 {
		s.initialProbeAt = time.Time{}
		return
	}
	if now.Before(s.initialProbeAt) {
		return
	}
	s.initialProbeAt = time.Time{}
	s.pendingProbe = true
}
