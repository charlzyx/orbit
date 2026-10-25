package mcpui

import (
	"testing"
	"testing/synctest"
	"time"
)

func TestRetainedMCPSectionPollsWithoutRecentRender(t *testing.T) {
	s := pollingState(t)
	// Allows this behavioral regression to run unchanged on the pre-host API.
	if host, ok := any(s).(interface{ SetHostVisible(bool) }); ok {
		host.SetHostVisible(true)
	}
	s.lastViewed = time.Now().Add(-time.Hour)
	s.nextPoll = time.Now().Add(-time.Second)
	s.Tick()
	if !s.Busy {
		t.Fatal("retained MCP section did not schedule due external poll without a recent render")
	}
	waitState(t, s)
}

func pollingState(t *testing.T) *State {
	t.Helper()
	s := stateFixture(t)
	s.service.PiBinary = fakePi(t, `printf '{"servers":[]}\n'`)
	return s
}
func TestHostMCPPollingRetainsSixtySecondBoundary(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := pollingState(t)
		s.SetHostVisible(true)
		s.lastViewed = time.Now().Add(-time.Hour)
		s.nextPoll = time.Now().Add(time.Minute)
		s.Tick()
		if s.Busy {
			t.Fatal("poll ran before due")
		}
		time.Sleep(59 * time.Second)
		s.Tick()
		if s.Busy {
			t.Fatal("poll interval shortened")
		}
		time.Sleep(time.Second)
		s.Tick()
		if !s.Busy {
			t.Fatal("poll absent at sixty-second boundary")
		}
	})
}
func TestHostMCPSectionExitStopsNewPollingButDrainsCompletion(t *testing.T) {
	s := pollingState(t)
	s.SetHostVisible(false)
	s.lastViewed = time.Now() // A previous frame must not override host navigation.
	s.nextPoll = time.Now().Add(-time.Second)
	s.Tick()
	if s.Busy {
		t.Fatal("closed MCP section polled due to recent frame")
	}
	s.SetHostVisible(true)
	s.Tick()
	if !s.Busy {
		t.Fatal("returning to retained MCP section did not poll")
	}
	s.SetHostVisible(false)
	waitState(t, s)
	if !s.loaded || s.Busy {
		t.Fatal("section exit stranded already-running completion")
	}
	s.nextPoll = time.Now().Add(-time.Second)
	s.Tick()
	if s.Busy {
		t.Fatal("completion restarted hidden external polling")
	}
}
func TestHostMCPExternalEditRemainsManualApply(t *testing.T) {
	s := pollingState(t)
	if err := s.service.SaveMCP("global", "", "synthetic", map[string]any{"command": "old-fixture"}); err != nil {
		t.Fatal(err)
	}
	s.adopt(loadSnapshot(s.service))
	modified, applied := 0, 0
	s.Options.OnModified = func() { modified++ }
	s.Options.OnApply = func() { applied++ }
	if err := s.service.SaveMCP("global", "synthetic", "synthetic", map[string]any{"command": "new-fixture"}); err != nil {
		t.Fatal(err)
	}
	s.SetHostVisible(true)
	s.lastViewed = time.Now().Add(-time.Hour)
	s.nextPoll = time.Now().Add(-time.Second)
	s.Tick()
	waitState(t, s)
	if !s.externalChange || len(s.Servers) != 1 || s.Servers[0].Definition["command"] != "new-fixture" {
		t.Fatal("external edit not reflected", s.Servers)
	}
	if modified != 0 || applied != 0 {
		t.Fatal("external polling automatically applied configuration", modified, applied)
	}
	s.apply()
	if applied != 1 {
		t.Fatal("explicit apply lost")
	}
}
