package agent

import (
	"context"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func TestAbortDelayedClearRejectsReusedStateButRestoresParkedSession(t *testing.T) {
	for _, reuse := range []bool{true, false} {
		name := "parked"
		if reuse {
			name = "reused"
		}
		t.Run(name, func(t *testing.T) {
			m := testManager(t, func(o *Options) {
				o.CommandTimeout = 15 * time.Second
				o.EnvOverrides["ORBIT_CLEAR_UNTIL_ABORT"] = "1"
				o.EnvOverrides["ORBIT_ABORT_REJECT"] = "1"
			})
			workspace := t.TempDir()
			if err := m.Open(workspace, ""); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" })
			if err := m.Submit("hold", nil, SendPrompt); err != nil {
				t.Fatal(err)
			}
			// Awaiting is locally optimistic; require the fixture agent_start before
			// Abort so the later !Running barrier proves agent_settled was received.
			waitFor(t, m, func(s Snapshot) bool { return s.Running && !s.Awaiting })
			if err := m.Send("emit_abort_queue", nil); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return len(s.Queue.Steering)+len(s.Queue.FollowUp) == 2 })
			m.mu.Lock()
			entry := m.sessions[m.active]
			client, oldID := entry.client, entry.state.snapshot.ActiveSession.ID
			m.mu.Unlock()
			if err := m.Abort(); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return !s.Running })
			if reuse {
				if err := m.NewSession(workspace); err != nil {
					t.Fatal(err)
				}
				waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != oldID && !m.AnyBusy() })
				m.mu.Lock()
				if m.sessions[m.active] != entry || entry.client != client {
					t.Error("fixture did not reuse the same process/state holder")
				}
				entry.state.setComposerText("new task draft")
				m.mu.Unlock()
			} else {
				if err := m.Open(t.TempDir(), ""); err != nil {
					t.Fatal(err)
				}
				waitFor(t, m, func(s Snapshot) bool { return s.Connected && !m.AnyBusy() })
			}
			if err := m.Send("emit_fresh_queue", nil); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return len(s.Queue.FollowUp) == 1 && s.Queue.FollowUp[0] == "new queued text" })
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if _, err := client.Call(ctx, pirpc.NewCommand("release_abort_clear", nil)); err != nil {
				t.Fatal(err)
			}
			// The release reply proves the child sent the old clear response;
			// drain its manager worker before asserting absence of a mutation.
			drained := make(chan struct{})
			go func() { m.workers.Wait(); close(drained) }()
			select {
			case <-drained:
			case <-time.After(5 * time.Second):
				t.Fatal("abort worker did not drain")
			}
			m.Tick()
			m.mu.Lock()
			text, revision := entry.state.snapshot.ComposerText, entry.state.snapshot.ComposerRevision
			currentText := m.sessions[m.active].state.snapshot.ComposerText
			m.mu.Unlock()
			if reuse {
				if text != "new task draft" || revision != 1 {
					t.Fatal("old clear response crossed the session reset", text, revision)
				}
			} else if text != "old steer\n\nold follow" || revision != 1 || currentText != "" {
				t.Fatal("parking lost restoration or contaminated current session", text, revision, currentText)
			}
			active := m.Snapshot()
			if active.ConnectionError != "" || len(active.Queue.FollowUp) != 1 || active.Queue.FollowUp[0] != "new queued text" {
				t.Fatal("old response event crossed into current task", active.ConnectionError, active.Queue)
			}
			if client.PendingCount() != 0 {
				t.Fatal("old abort pipeline leaked pending RPC calls", client.PendingCount())
			}
		})
	}
}

func TestAbortPendingRestorationCancelledByExplicitClear(t *testing.T) {
	m := testManager(t, func(o *Options) { o.CommandTimeout = 15 * time.Second; o.EnvOverrides["ORBIT_CLEAR_UNTIL_ABORT"] = "1" })
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("hold", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	// Awaiting is locally optimistic; require the fixture agent_start before
	// Abort so the later !Running barrier proves agent_settled was received.
	waitFor(t, m, func(s Snapshot) bool { return s.Running && !s.Awaiting })
	if err := m.Send("emit_abort_queue", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return len(s.Queue.Steering)+len(s.Queue.FollowUp) == 2 })
	m.mu.Lock()
	entry := m.sessions[m.active]
	client := entry.client
	entry.state.setComposerText("keep local draft")
	m.mu.Unlock()
	if err := m.Abort(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Running })
	if err := m.ClearQueue(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := client.Call(ctx, pirpc.NewCommand("release_abort_clear", nil)); err != nil {
		t.Fatal(err)
	}
	drained := make(chan struct{})
	go func() { m.workers.Wait(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("abort/clear workers did not drain")
	}
	m.Tick()
	s := m.Snapshot()
	if s.ComposerText != "keep local draft" || s.ComposerRevision != 1 || s.Running || client.PendingCount() != 0 {
		t.Fatal("explicit clear did not cancel pending restoration", s.ComposerText, s.ComposerRevision, s.Running, client.PendingCount())
	}
}

func TestAbortPendingPipelineStopDoesNotRestoreExpectedCloseError(t *testing.T) {
	m := testManager(t, func(o *Options) {
		o.CommandTimeout = 15 * time.Second
		o.EnvOverrides["ORBIT_CLEAR_UNTIL_ABORT"] = "1"
		o.EnvOverrides["ORBIT_ABORT_REJECT"] = "1" // Keep both responses pending until Stop closes the child.
	})
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("hold", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	// Awaiting is locally optimistic; require the fixture agent_start before
	// Abort so the later !Running barrier proves agent_settled was received.
	waitFor(t, m, func(s Snapshot) bool { return s.Running && !s.Awaiting })
	if err := m.Send("emit_abort_queue", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return len(s.Queue.Steering)+len(s.Queue.FollowUp) == 2 })
	if err := m.Abort(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Running })
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	drained := make(chan struct{})
	go func() { m.workers.Wait(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("stopped pipeline did not drain")
	}
	m.Tick()
	if s := m.Snapshot(); s.ConnectionError != "" || s.ComposerRevision != 0 {
		t.Fatal("expected close mutated stopped state", s.ConnectionError, s.ComposerRevision)
	}
}

func TestAbortOriginIDsValidateAndIdentifyResetState(t *testing.T) {
	origin := abortOrigin{sessionID: "session:with/utf8界", reset: 7, nonce: 12}
	for _, kind := range []string{pirpc.Abort, pirpc.ClearQueue} {
		command := origin.command(kind)
		got, ok := parseAbortOrigin(command.ID, kind)
		if !ok || got != origin {
			t.Fatal("origin round trip", command.ID, got, ok)
		}
		if _, ok := parseAbortOrigin(command.ID, pirpc.GetState); ok {
			t.Fatal("wrong response command accepted")
		}
	}
	for _, id := range []string{"orbit-go-1", "orbit-abort-v1:bad", "orbit-abort-v1:-1:1::abort", "orbit-abort-v1:00:1::abort", "orbit-abort-v1:1:0::abort", "orbit-abort-v1:1:2:???::abort", "orbit-abort-v1:1:2::abort:extra"} {
		if _, ok := parseAbortOrigin(id, pirpc.Abort); ok {
			t.Fatal("malformed origin accepted", id)
		}
	}
}
