package agent

import (
	"strings"
	"testing"
)

func TestRetryFailureMetadataOwnershipAndLiteralPreservation(t *testing.T) {
	for _, raw := range []string{`{"type":"auto_retry_end","success":false}`, `{"type":"auto_retry_end","success":false,"finalError":null}`, `{"type":"auto_retry_end","success":false,"finalError":42}`} {
		s := newState("/one", "")
		apply(t, &s, raw)
		if !s.snapshot.RetryFailure.Active || s.snapshot.RetryFailure.MessageProvided || s.snapshot.RetryFailure.Message != "" {
			t.Fatal(raw, s.snapshot)
		}
	}
	s := newState("/one", "")
	apply(t, &s, `{"type":"auto_retry_end","success":false,"finalError":"pi exhausted its automatic retries"}`)
	if !s.snapshot.RetryFailure.MessageProvided || s.snapshot.RetryFailure.Message != "pi exhausted its automatic retries" {
		t.Fatal(s.snapshot)
	}
	same := s.snapshot.ConnectionError
	// A generic error with exactly the same English bytes owns its own text;
	// it must not inherit a retry translation by string-equality guessing.
	apply(t, &s, `{"type":"response","command":"unrelated","success":false,"error":"Automatic retry failed: pi exhausted its automatic retries"}`)
	if s.snapshot.ConnectionError != same || s.snapshot.RetryFailure.Active {
		t.Fatal("generic same-text error inherited metadata", s.snapshot)
	}
	for _, event := range []string{`{"type":"extension_error","message":"extension failed"}`, `{"type":"transport_error","error":"transport failed"}`, `{"type":"protocol_error","error":"protocol failed"}`} {
		s.snapshot.setRetryFailure("old", true)
		apply(t, &s, event)
		if s.snapshot.RetryFailure.Active {
			t.Fatal("superseding owner retained metadata", event, s.snapshot)
		}
	}
	s.snapshot.setRetryFailure("old", true)
	s = newState("/two", "")
	if s.snapshot.RetryFailure.Active {
		t.Fatal("new state retained retry failure")
	}
}

func TestRetryFailureSnapshotRedactionAndCheckpointPrecedence(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !m.AnyBusy() })
	const secret = "synthetic-retry-token-1738"
	m.mu.Lock()
	entry := m.sessions[m.active]
	entry.redactions = append(entry.redactions, secret)
	entry.state.snapshot.setRetryFailure("literal "+secret, true)
	m.mu.Unlock()
	s := m.Snapshot()
	if strings.Contains(s.ConnectionError, secret) || strings.Contains(s.RetryFailure.Message, secret) || !strings.Contains(s.RetryFailure.Message, "[redacted]") || !s.RetryFailure.MessageProvided {
		t.Fatal("typed error bypassed redaction", s.ConnectionError, s.RetryFailure)
	}
	m.mu.Lock()
	entry.state.snapshot.CheckpointRecovering = true
	entry.state.snapshot.CheckpointRecoveryError = "checkpoint failure"
	m.mu.Unlock()
	s = m.Snapshot()
	if s.ConnectionError != "checkpoint failure" || s.RetryFailure.Active {
		t.Fatal("typed error overrode checkpoint priority", s.ConnectionError, s.RetryFailure)
	}
	m.mu.Lock()
	entry.state.snapshot.CheckpointRecovering = false
	m.mu.Unlock()
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if s = m.Snapshot(); s.RetryFailure.Active {
		t.Fatal("Stop retained error owner", s)
	}
}

func TestRetryFailureLateParkedEventDoesNotReachForeground(t *testing.T) {
	m := testManager(t, nil)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !m.AnyBusy() })
	if err := m.Send("emit_retry_start", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Retrying })
	m.mu.Lock()
	oldKey := m.active
	old := m.sessions[oldKey]
	client := old.client
	m.mu.Unlock()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != oldKey })
	compactCall(t, client, "emit_retry_failure")
	m.Tick()
	s := m.Snapshot()
	if s.RetryFailure.Active || s.ConnectionError != "" {
		t.Fatal("old process error reached new foreground", s)
	}
	m.mu.Lock()
	failure := old.state.snapshot.RetryFailure
	m.mu.Unlock()
	if !failure.Active || failure.Message != "parked failure" {
		t.Fatal("old event was dropped instead of owned by old session", failure)
	}
}
