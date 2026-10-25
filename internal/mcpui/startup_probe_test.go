package mcpui

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func scheduleInitialFixture(s *State, now time.Time) {
	if scheduler, ok := any(s).(interface{ ScheduleInitialProbe(time.Time) }); ok {
		scheduler.ScheduleInitialProbe(now)
	} else {
		s.PrepareQuotaPopup()
	}
}
func startupTick(s *State, now time.Time) {
	if clock, ok := any(s).(interface{ tickAt(time.Time) bool }); ok {
		clock.tickAt(now)
	} else {
		s.Tick()
	}
}
func TestMCPInitialProbeThreeSecondDeadline(t *testing.T) {
	s, log := popupState(t)
	now := time.Now()
	scheduleInitialFixture(s, now)
	waitState(t, s)
	now = s.initialProbeAt.Add(-3 * time.Second)
	startupTick(s, now.Add(3*time.Second-time.Nanosecond))
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatal("startup probe ran before3s")
	}
	startupTick(s, now.Add(3*time.Second))
	waitState(t, s)
	raw, err := os.ReadFile(log)
	if err != nil || strings.Count(string(raw), "mcp list --json") != 1 {
		t.Fatal("configured startup failed exactly one probe at3s", string(raw), err)
	}
	startupTick(s, now.Add(time.Hour))
	waitState(t, s)
	raw, _ = os.ReadFile(log)
	if strings.Count(string(raw), "mcp list --json") != 1 {
		t.Fatal("startup probe repeated")
	}
}

func TestMCPInitialProbeBusyDefersAndManualProbeCoalesces(t *testing.T) {
	for _, manual := range []bool{false, true} {
		t.Run(map[bool]string{false: "busy", true: "manual"}[manual], func(t *testing.T) {
			s, log := popupState(t)
			now := time.Now()
			s.ScheduleInitialProbe(now)
			waitState(t, s)
			now = s.initialProbeAt.Add(-3 * time.Second)
			if manual {
				s.OpenQuotaPopup(now)
				waitState(t, s)
			} else {
				release := make(chan struct{})
				s.start("held fixture", func(ctx context.Context) (func(*State), error) { <-release; return nil, nil })
				startupTick(s, now.Add(3*time.Second))
				if _, e := os.Stat(log); !os.IsNotExist(e) {
					t.Fatal("probe interrupted busy operation")
				}
				close(release)
				waitState(t, s)
			}
			startupTick(s, now.Add(3*time.Second))
			waitState(t, s)
			raw, e := os.ReadFile(log)
			if e != nil || strings.Count(string(raw), "mcp list --json") != 1 {
				t.Fatal("probe lost or duplicated", string(raw), e)
			}
		})
	}
}
func TestMCPInitialProbeEmptyConfigDoesNotRunPi(t *testing.T) {
	s, log := popupState(t)
	if e := s.service.RemoveMCP("global", "fixture"); e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	s.ScheduleInitialProbe(now)
	waitState(t, s)
	startupTick(s, now.Add(time.Hour))
	waitState(t, s)
	if _, e := os.Stat(log); !os.IsNotExist(e) {
		t.Fatal("empty configuration invoked Pi", e)
	}
}
func TestMCPInitialProbeIntentCancelledByContextChange(t *testing.T) {
	for _, change := range []string{"workspace", "binary", "close"} {
		t.Run(change, func(t *testing.T) {
			s, log := popupState(t)
			now := time.Now()
			s.ScheduleInitialProbe(now)
			waitState(t, s)
			switch change {
			case "workspace":
				s.SetWorkspace(t.TempDir())
			case "binary":
				s.SetPiBinary(fakePi(t, "printf '%s\\n' probe >> '"+log+"'; printf '%s\\n' '{\"servers\":[]}'"))
			case "close":
				s.Close()
			}
			if change != "close" {
				s.PrepareQuotaPopup()
				waitState(t, s)
			}
			startupTick(s, now.Add(time.Hour))
			waitState(t, s)
			if _, e := os.Stat(log); !os.IsNotExist(e) {
				t.Fatal("old startup intent crossed context", change, e)
			}
			if !s.initialProbeAt.IsZero() || s.initialProbeWaiting {
				t.Fatal("cancelled startup deadline retained")
			}
		})
	}
}

func TestMCPInitialProbeSlowLoadWaitsThreeSecondsAfterAdoption(t *testing.T) {
	s, log := popupState(t)
	now := time.Now()
	release := make(chan struct{})
	service := s.service
	s.start("held config load", func(ctx context.Context) (func(*State), error) {
		<-release
		v := loadSnapshot(service)
		return func(s *State) { s.adopt(v) }, nil
	})
	s.ScheduleInitialProbe(now)
	startupTick(s, now.Add(10*time.Second))
	if _, e := os.Stat(log); !os.IsNotExist(e) {
		t.Fatal("probe before config completed")
	}
	close(release)
	deadline := time.Now().Add(time.Second)
	for !s.loaded && time.Now().Before(deadline) {
		startupTick(s, now.Add(10*time.Second))
		time.Sleep(time.Millisecond)
	}
	if !s.loaded {
		t.Fatal("held config never adopted")
	}
	if s.Probing || !s.initialProbeAt.Equal(now.Add(13*time.Second)) {
		t.Fatal("deadline not anchored to successful config adoption", s.Probing, s.initialProbeAt.Sub(now))
	}
	startupTick(s, now.Add(13*time.Second-time.Nanosecond))
	if _, e := os.Stat(log); !os.IsNotExist(e) {
		t.Fatal("probe before post-load3s")
	}
	startupTick(s, now.Add(13*time.Second))
	waitState(t, s)
	raw, e := os.ReadFile(log)
	if e != nil || strings.Count(string(raw), "mcp list --json") != 1 {
		t.Fatal("postload deadline did not probe once", string(raw), e)
	}
}
