// Package menuproof is an isolated, unwired native-test feasibility helper.
// It is not imported by Orbit, nor enabled by any workflow.
package menuproof

import (
	"errors"
	"math"
	"slices"
	"time"
)

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

func (r Rect) Valid() bool {
	for _, n := range []float64{r.X, r.Y, r.W, r.H, r.X + r.W, r.Y + r.H} {
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return false
		}
	}
	return r.W > 0 && r.H > 0
}
func (r Rect) Contains(s Rect) bool {
	return r.Valid() && s.Valid() && s.X >= r.X && s.Y >= r.Y && s.X+s.W <= r.X+r.W && s.Y+s.H <= r.Y+r.H
}

type Identity struct {
	PID    int     `json:"pid"`
	Start  uint64  `json:"start_us"`
	Window uintptr `json:"window"`
}
type Arm struct {
	Owner           Identity
	At, Deadline    time.Time
	Content, Target Rect
	Matches         int
	Labels          []string
}
type Observation struct {
	Owner        Identity
	Menu         uintptr
	Labels       []string
	At           time.Time
	Key, Visible bool
}
type State struct {
	arm                                   Arm
	armed                                 bool
	menu                                  uintptr
	clickAt                               time.Time
	clickNonce                            uint64
	EscapeRequests, Begins, Ends, Actions int
	Failed                                string
}

var ErrBlocked = errors.New("native menu proof blocked")

func (s *State) Arm(a Arm) error {
	if s.Failed != "" || s.armed || s.Begins != 0 || a.Owner.PID <= 0 || a.Owner.Start == 0 || a.Owner.Window == 0 || a.Matches != 1 || !a.Content.Contains(a.Target) || !a.Deadline.After(a.At) || a.Deadline.Sub(a.At) > 5*time.Second || len(a.Labels) != 4 {
		return ErrBlocked
	}
	for _, v := range a.Labels {
		if v == "" || len(v) > 80 {
			return ErrBlocked
		}
	}
	a.Labels = slices.Clone(a.Labels)
	s.arm = a
	s.armed = true
	return nil
}

// RecordClick binds begin tracking to the one freshly checked Add attachment
// event receipt, not merely a menu with matching labels. Called immediately
// before its down/up pair is queued on the main thread.
func (s *State) RecordClick(owner Identity, target Rect, at time.Time, nonce uint64) error {
	if !s.armed || s.clickNonce != 0 || owner != s.arm.Owner || target != s.arm.Target || nonce == 0 || at.Before(s.arm.At) || !at.Before(s.arm.Deadline) {
		return ErrBlocked
	}
	s.clickAt = at
	s.clickNonce = nonce
	return nil
}
func (s *State) owner(o Observation) bool {
	return s.armed && o.Owner == s.arm.Owner && o.Key && o.Visible && !o.At.Before(s.arm.At) && o.At.Before(s.arm.Deadline)
}

// Begin accepts only the one expected root popup in this process and returns
// permission for one queued Escape. Unrelated notifications do not authorize input.
func (s *State) Begin(o Observation) bool {
	if !s.armed {
		return false
	}
	if s.clickNonce == 0 || o.At.Before(s.clickAt) || !s.owner(o) || o.Menu == 0 || !slices.Equal(o.Labels, s.arm.Labels) {
		s.Failed = "unexpected popup or missing Add click receipt"
		s.armed = false
		return false
	}
	if s.menu != 0 {
		s.Failed = "duplicate menu begin"
		return false
	}
	s.menu = o.Menu
	s.Begins++
	s.EscapeRequests++
	return true
}
func (s *State) Action(menu uintptr) {
	if s.armed && menu == s.menu && menu != 0 {
		s.Actions++
		s.Failed = "unexpected menu action"
	}
}
func (s *State) End(o Observation) {
	if !s.owner(o) || o.Menu != s.menu || s.menu == 0 {
		return
	}
	s.Ends++
	s.armed = false
	s.arm.Labels = nil
}
func (s *State) Result() error {
	if s.Failed != "" || s.armed || s.Begins != 1 || s.Ends != 1 || s.EscapeRequests != 1 || s.Actions != 0 {
		return ErrBlocked
	}
	return nil
}
func (s *State) Close() {
	s.armed = false
	s.arm.Labels = nil
	if s.Ends != 1 {
		s.Failed = "closed without native end observation"
	}
}
