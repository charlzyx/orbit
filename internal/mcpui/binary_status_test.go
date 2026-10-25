package mcpui

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMCPBinaryChangeInvalidatesOldProbeWithoutAutoProbing(t *testing.T) {
	s, log := popupState(t)
	s.Refresh()
	waitState(t, s)
	if s.Runtime["fixture"].State != "connected" {
		t.Fatal("old binary fixture not connected")
	}
	s.Selected = "fixture"
	s.probeErrors = []string{"old probe error"}
	s.probeNote = "old probe note"
	s.Error = "old operation error"
	s.Status = "old operation success"
	next := fakePi(t, "printf '%s\\n' new-binary >> '"+log+"'; printf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"failed\",\"error\":\"new failure\"}]}'")
	s.SetPiBinary(next)
	if s.Runtime["fixture"].State != "" || !s.checkedAt.IsZero() || len(s.probeErrors) > 0 || s.probeNote != "" || s.Error != "" || s.Status != "" {
		t.Fatal("binary switch retained stale probe/operation state", s.Runtime, s.checkedAt, s.Error, s.Status)
	}
	if s.Selected != "fixture" || len(s.Servers) != 1 {
		t.Fatal("binary switch discarded configuration selection")
	}
	s.PrepareQuotaPopup()
	waitState(t, s)
	raw, _ := os.ReadFile(log)
	if strings.Contains(string(raw), "new-binary") {
		t.Fatal("binary change silently transferred probe intent")
	}
	if !strings.Contains(s.QuotaSummary(), "Not checked") {
		t.Fatal("new binary unknown manufactured readiness", s.QuotaSummary())
	}
	s.OpenQuotaPopup(time.Now())
	waitState(t, s)
	raw, _ = os.ReadFile(log)
	if strings.Count(string(raw), "new-binary") != 1 || s.Runtime["fixture"].State != "failed" {
		t.Fatal("fresh explicit intent did not adopt real new-binary result", string(raw), s.Runtime)
	}
}

func TestMCPBinaryNoOpPreservesCurrentProbe(t *testing.T) {
	s, _ := popupState(t)
	s.Refresh()
	waitState(t, s)
	checked := s.checkedAt
	old := s.Runtime["fixture"]
	s.SetPiBinary(s.Options.PiBinary)
	if !s.checkedAt.Equal(checked) || s.Runtime["fixture"].State != old.State {
		t.Fatal("same binary needlessly invalidated current status")
	}
}
func TestMCPBinaryFailureDoesNotRestoreOldReadyAndExplicitRetryWorks(t *testing.T) {
	s, _ := popupState(t)
	s.Refresh()
	waitState(t, s)
	bin := fakePi(t, "echo fixture_failure >&2; exit 2")
	s.SetPiBinary(bin)
	s.Refresh()
	waitState(t, s)
	if s.Error == "" || s.Runtime["fixture"].State != "" || !s.checkedAt.IsZero() {
		t.Fatal("failed replacement reused old ready status", s.Error, s.Runtime)
	}
	if e := os.WriteFile(bin, []byte("#!/bin/sh\nprintf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"connected\",\"tools\":[\"new_tool\"]}]}'\n"), 0700); e != nil {
		t.Fatal(e)
	}
	s.Refresh()
	waitState(t, s)
	if s.Error != "" || s.Runtime["fixture"].State != "connected" || len(s.Runtime["fixture"].Tools) != 1 || s.Runtime["fixture"].Tools[0].Name != "new_tool" {
		t.Fatal("explicit retry did not adopt new binary truth", s.Error, s.Runtime)
	}
}
func TestMCPBinarySupersededOperationCannotRestoreOldStatus(t *testing.T) {
	s, _ := popupState(t)
	s.Refresh()
	waitState(t, s)
	release := make(chan struct{})
	finished := make(chan struct{})
	s.start("held old binary completion", func(ctx context.Context) (func(*State), error) {
		<-release
		defer close(finished)
		return func(s *State) {
			s.Runtime["fixture"] = Runtime{State: "connected", Tools: []Tool{{Name: "stale_tool"}}}
			s.Status = "old success"
		}, errors.New("old failure")
	})
	s.SetPiBinary(fakePi(t, "printf '%s\\n' '{\"servers\":[{\"name\":\"fixture\",\"state\":\"needs-auth\"}]}'"))
	close(release)
	<-finished
	deadline := time.Now().Add(time.Second)
	for len(s.results) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if len(s.results) == 0 {
		t.Fatal("old completion never reached owner mailbox")
	}
	s.Refresh()
	waitState(t, s)
	s.Tick()
	if s.Error != "" || s.Status == "old success" || s.Runtime["fixture"].State != "needs-auth" || len(s.Runtime["fixture"].Tools) > 0 {
		t.Fatal("superseded completion crossed binary generation", s.Error, s.Status, s.Runtime)
	}
}
