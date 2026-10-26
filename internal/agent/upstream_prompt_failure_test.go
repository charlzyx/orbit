package agent

import (
	"context"
	"errors"
	"github.com/charlzyx/orbit/internal/pirpc"
	"testing"
	"time"
)

func TestUpstreamP1ForeignFailedPromptPreservesComposer(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Call(ctx, pirpc.NewCommand(pirpc.Prompt, map[string]any{"message": "foreign-fail"}))
	var rejected *pirpc.CommandError
	if !errors.As(err, &rejected) {
		t.Fatalf("expected actual failed foreign RPC, got %v", err)
	}
	m.Tick()
	m.mu.Lock()
	awaiting, turn := m.sessions[m.active].state.awaiting, m.sessions[m.active].turnOpen
	m.mu.Unlock()
	if !awaiting || !turn {
		t.Fatalf("foreign failure cancelled composer: awaiting=%v turn=%v", awaiting, turn)
	}
	p1Call(t, m, c, "finish_composer", nil)
	if m.AnyBusy() {
		t.Fatal("matching successful reply did not settle composer")
	}
}
func TestUpstreamP1LateFailedPromptPreservesNewerComposer(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	if err := m.Send(pirpc.Prompt, map[string]any{"message": "composer-pending"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ExtensionStatusText == "composer received 2" })
	p1Call(t, m, c, "fail_first_composer", nil)
	m.mu.Lock()
	awaiting, turn := m.sessions[m.active].state.awaiting, m.sessions[m.active].turnOpen
	m.mu.Unlock()
	if !awaiting || !turn {
		t.Fatalf("old failed request cancelled newer composer: awaiting=%v turn=%v", awaiting, turn)
	}
	p1Call(t, m, c, "fail_current_composer", nil)
	m.mu.Lock()
	open := m.sessions[m.active].turnOpen
	m.mu.Unlock()
	if m.Snapshot().Awaiting || open || m.Snapshot().LatestTurn != 0 {
		t.Fatal("matching failure retained optimistic awaiting")
	}
}

func TestUpstreamP1MissingIDPromptCannotConsumeComposer(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	for _, kind := range []string{"empty_prompt_success", "empty_prompt_failure"} {
		p1Call(t, m, c, kind, nil)
		m.mu.Lock()
		awaiting, turn := m.sessions[m.active].state.awaiting, m.sessions[m.active].turnOpen
		m.mu.Unlock()
		if !awaiting || !turn {
			t.Fatal(kind, "consumed composer")
		}
	}
	p1Call(t, m, c, "finish_composer", nil)
	if m.AnyBusy() {
		t.Fatal("matching response did not settle")
	}
}
func TestUpstreamP1LateSuccessPreservesNewerComposer(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	if err := m.Send(pirpc.Prompt, map[string]any{"message": "composer-pending"}); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ExtensionStatusText == "composer received 2" })
	p1Call(t, m, c, "finish_first_composer", nil)
	if !m.Snapshot().Awaiting {
		t.Fatal("late success consumed newer prompt")
	}
	p1Call(t, m, c, "finish_composer", nil)
	if m.AnyBusy() {
		t.Fatal("matching response did not settle")
	}
}
func TestUpstreamP1FailedResponseSessionABADoesNotCancelNewPrompt(t *testing.T) {
	m := p1Manager(t)
	c := p1Composer(t, m, false)
	row := m.Snapshot().ActiveSession
	if err := m.Abort(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Awaiting && !s.Running && !s.Retrying })
	for _, want := range []string{"p2", "p1"} {
		if err := m.NewSession(row.Workspace); err != nil {
			t.Fatal(err)
		}
		waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == want && !m.AnyBusy() })
	}
	m.mu.Lock()
	sameClient := m.sessions[m.active].client == c
	m.mu.Unlock()
	if !sameClient {
		t.Fatal("fixture failed to reuse actual process")
	}
	if err := m.Submit("composer-pending", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.ExtensionStatusText == "composer received 2" })
	p1Call(t, m, c, "fail_first_composer", nil)
	if !m.Snapshot().Awaiting {
		t.Fatal("old response crossed session ABA")
	}
	p1Call(t, m, c, "finish_composer", nil)
	if m.AnyBusy() {
		t.Fatal("fresh matching response did not settle")
	}
}
