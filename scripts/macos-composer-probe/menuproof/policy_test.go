package menuproof

import (
	"math"
	"testing"
	"time"
)

func armFixture() (Arm, Observation) {
	now := time.Unix(1, 0)
	id := Identity{10, 20, 30}
	labels := []string{"Attach image…", "Attach files…", "Mention file", "Paste image"}
	return Arm{id, now, now.Add(time.Second), Rect{0, 0, 100, 100}, Rect{1, 1, 10, 10}, 1, labels}, Observation{id, 40, labels, now, true, true}
}
func TestMenuExactSingleEscapeAndEnd(t *testing.T) {
	a, o := armFixture()
	s := new(State)
	if armClicked(s, a) != nil || !s.Begin(o) {
		t.Fatal("admission")
	}
	if s.Result() == nil {
		t.Fatal("begin is not cancellation proof")
	}
	s.End(o)
	if s.Result() != nil {
		t.Fatal("end not accepted")
	}
	if s.Begin(o) {
		t.Fatal("repeated input")
	}
}
func TestMenuRejectsAmbiguousInvalidStaleOwnership(t *testing.T) {
	for _, kind := range []string{"duplicate-target", "nan", "inf", "outside", "zero-pid", "zero-start", "zero-window", "expired-arm", "long-deadline"} {
		t.Run(kind, func(t *testing.T) {
			a, _ := armFixture()
			switch kind {
			case "duplicate-target":
				a.Matches = 2
			case "nan":
				a.Target.X = math.NaN()
			case "inf":
				a.Content.W = math.Inf(1)
			case "outside":
				a.Target.X = 200
			case "zero-pid":
				a.Owner.PID = 0
			case "zero-start":
				a.Owner.Start = 0
			case "zero-window":
				a.Owner.Window = 0
			case "expired-arm":
				a.Deadline = a.At
			case "long-deadline":
				a.Deadline = a.At.Add(6 * time.Second)
			}
			if new(State).Arm(a) == nil {
				t.Fatal("unsafe arm")
			}
		})
	}
	for _, kind := range []string{"pid", "start", "window", "labels", "not-key", "hidden", "late", "early", "zero-menu"} {
		t.Run(kind, func(t *testing.T) {
			a, o := armFixture()
			s := new(State)
			if armClicked(s, a) != nil {
				t.Fatal("arm")
			}
			switch kind {
			case "pid":
				o.Owner.PID++
			case "start":
				o.Owner.Start++
			case "window":
				o.Owner.Window++
			case "labels":
				o.Labels = []string{"other"}
			case "not-key":
				o.Key = false
			case "hidden":
				o.Visible = false
			case "late":
				o.At = a.Deadline
			case "early":
				o.At = a.At.Add(-time.Nanosecond)
			case "zero-menu":
				o.Menu = 0
			}
			if s.Begin(o) || s.EscapeRequests != 0 {
				t.Fatal("unsafe event")
			}
		})
	}
}
func TestMenuActionDuplicateWrongEndAndCloseFailClosed(t *testing.T) {
	for _, kind := range []string{"action", "duplicate", "wrong-end", "close"} {
		t.Run(kind, func(t *testing.T) {
			a, o := armFixture()
			s := new(State)
			_ = armClicked(s, a)
			if !s.Begin(o) {
				t.Fatal("begin")
			}
			switch kind {
			case "action":
				s.Action(o.Menu)
				s.End(o)
			case "duplicate":
				if s.Begin(o) {
					t.Fatal("duplicate escape")
				}
				s.End(o)
			case "wrong-end":
				o.Menu++
				s.End(o)
			case "close":
				s.Close()
			}
			if s.Result() == nil {
				t.Fatal("false PASS")
			}
		})
	}
}

func TestClosedBeforeBeginCannotRearmOrAuthorizeEscape(t *testing.T) {
	a, o := armFixture()
	s := new(State)
	if armClicked(s, a) != nil {
		t.Fatal("arm")
	}
	s.Close()
	if armClicked(s, a) == nil || s.Begin(o) || s.EscapeRequests != 0 {
		t.Fatal("closed proof authorized retry input")
	}
	if s.Result() == nil {
		t.Fatal("closed proof passed")
	}
}

func armClicked(s *State, a Arm) error {
	if e := s.Arm(a); e != nil {
		return e
	}
	return s.RecordClick(a.Owner, a.Target, a.At, 1)
}
func TestMenuRequiresClickReceiptAndRejectsOtherPopupFirst(t *testing.T) {
	a, o := armFixture()
	s := new(State)
	if s.Arm(a) != nil {
		t.Fatal("arm")
	}
	if s.Begin(o) || s.EscapeRequests != 0 {
		t.Fatal("no click still authorized input")
	}
	s = new(State)
	if armClicked(s, a) != nil {
		t.Fatal("arm")
	}
	other := o
	other.Labels = []string{"other"}
	if s.Begin(other) || s.Begin(o) || s.EscapeRequests != 0 {
		t.Fatal("other popup did not fail closed")
	}
}
