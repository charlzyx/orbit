package transcriptui

import "time"

type turnClock struct {
	started  time.Time
	elapsed  time.Duration
	finished bool
}

// Observe supplies a client-measured live wall clock and holds the newest
// assistant turn live through tool/round-trip gaps until agent_settled. Call on
// each authoritative session snapshot, after State.Session; never use it for
// stale responses from a previously selected runtime. Historical snapshots do
// not acquire elapsed durations. Pi durationMs remains the rate denominator.
// now is explicit so repeated, cancelled and clock-regression flows are testable.
func (s *State) Observe(rows []Row, running bool, now time.Time) []Row {
	if s.open == nil {
		s.Reset(s.session)
	}
	if s.clocks == nil {
		s.clocks = map[string]turnClock{}
	}
	out := append([]Row(nil), rows...)
	last := -1
	if len(out) > 0 && out[len(out)-1].Role == "assistant" {
		last = len(out) - 1
	}
	for i := range out {
		row := &out[i]
		clock, observed := s.clocks[row.ID]
		active := running && i == last
		// Only an actual streaming row may start a clock; loading an old stopped
		// turn while the next prompt is waiting must never relabel old history.
		if active && row.Streaming && !observed {
			clock.started = now
			observed = true
		}
		if observed {
			if !clock.finished {
				if now.After(clock.started) {
					clock.elapsed = max(clock.elapsed, now.Sub(clock.started))
				}
				if !active {
					clock.finished = true
				}
			}
			duration := clock.elapsed
			row.Elapsed = &duration
			if active && !clock.finished {
				row.Streaming = true
			}
			s.clocks[row.ID] = clock
		}
	}
	return out
}

func (s *State) elapsed(id string, now time.Time) *time.Duration {
	clock, ok := s.clocks[id]
	if !ok {
		return nil
	}
	value := clock.elapsed
	if !clock.finished && now.After(clock.started) {
		value = max(value, now.Sub(clock.started))
	}
	return &value
}
