package agent

import (
	"context"
	"github.com/charlzyx/orbit/internal/pirpc"
	"strings"
	"testing"
	"time"
)

func compactFixture(t *testing.T, timeout time.Duration) (*Manager, *liveSession) {
	t.Helper()
	m := testManager(t, func(o *Options) {
		o.CommandTimeout = timeout
		o.EnvOverrides["ORBIT_COMPACT_HOLD"] = "1"
		o.EnvOverrides["ORBIT_COMPACT_REJECT"] = "1"
	})
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != "" && !m.AnyBusy() })
	m.mu.Lock()
	entry := m.sessions[m.active]
	m.mu.Unlock()
	return m, entry
}
func compactCall(t *testing.T, client *pirpc.Client, kind string) pirpc.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	response, err := client.Call(ctx, pirpc.NewCommand(kind, nil))
	if err != nil {
		t.Fatal(err)
	}
	return response
}
func waitCompactReceived(t *testing.T, client *pirpc.Client) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		var data struct{ Count int }
		response := compactCall(t, client, "compact_count")
		_ = response.DecodeData(&data)
		if data.Count == 1 {
			return
		}
		if data.Count > 1 {
			t.Fatal("duplicate compact", data.Count)
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("compact never reached fixture")
}
func TestCompactReservesBusyAndDeduplicatesBeforeEvents(t *testing.T) {
	m, entry := compactFixture(t, 2*time.Second)
	if err := m.Compact(); err != nil {
		t.Fatal(err)
	}
	if !m.Snapshot().Compacting || !m.AnyBusy() {
		t.Fatal("reservation not synchronous")
	}
	waitCompactReceived(t, entry.client)
	compactCall(t, entry.client, "emit_stale_compact_state")
	m.Tick()
	if !m.Snapshot().Compacting || !m.AnyBusy() {
		t.Fatal("older get_state erased pending reservation")
	}
	for i := 0; i < 8; i++ {
		if err := m.Compact(); err != nil {
			t.Fatal(err)
		}
	}
	waitCompactReceived(t, entry.client)
	compactCall(t, entry.client, "release_compact")
	waitFor(t, m, func(s Snapshot) bool {
		return !s.Compacting && strings.Contains(s.ConnectionError, "fixture compact rejected")
	})
}
func TestCompactOldResponseCannotMutateReusedSession(t *testing.T) {
	m, entry := compactFixture(t, 2*time.Second)
	oldID := m.Snapshot().ActiveSession.ID
	workspace := m.Snapshot().ActiveSession.Workspace
	if err := m.Compact(); err != nil {
		t.Fatal(err)
	}
	waitCompactReceived(t, entry.client)
	compactCall(t, entry.client, "finish_compact")
	waitFor(t, m, func(s Snapshot) bool { return !s.Compacting })
	if err := m.NewSession(workspace); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID != oldID && !m.AnyBusy() })
	m.mu.Lock()
	if m.sessions[m.active] != entry {
		t.Error("fixture did not reuse holder")
	}
	entry.state.snapshot.ConnectionError = "new task error"
	m.mu.Unlock()
	compactCall(t, entry.client, "release_compact")
	m.workers.Wait()
	m.Tick()
	if s := m.Snapshot(); s.ConnectionError != "new task error" || s.Compacting {
		t.Fatal("old compact crossed reset", s)
	}
}
func TestCompactTimeoutPreservesUncertainBusyUntilStop(t *testing.T) {
	m, entry := compactFixture(t, 100*time.Millisecond)
	if err := m.Compact(); err != nil {
		t.Fatal(err)
	}
	waitCompactReceived(t, entry.client)
	m.workers.Wait()
	m.Tick()
	if s := m.Snapshot(); !s.Compacting || !strings.Contains(s.ConnectionError, "deadline") {
		t.Fatal("timeout misreported stopped compaction", s)
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if s := m.Snapshot(); s.Compacting || s.Connected {
		t.Fatal("Stop retained compaction", s)
	}
}
func TestCompactOriginRoundTrip(t *testing.T) {
	want := abortOrigin{sessionID: "session:界", reset: 4, nonce: 7}
	command := want.command(pirpc.Compact)
	got, ok := parseAbortOrigin(command.ID, pirpc.Compact)
	if !ok || got != want {
		t.Fatal(got, ok)
	}
	if _, ok := parseAbortOrigin(command.ID, pirpc.Abort); ok {
		t.Fatal("compact accepted as abort")
	}
}

func TestCompactRequiresHydratedIdentityWithoutLeavingReservation(t *testing.T) {
	m, entry := compactFixture(t, 2*time.Second)
	m.mu.Lock()
	id := entry.state.snapshot.ActiveSession.ID
	entry.state.snapshot.ActiveSession.ID = ""
	m.mu.Unlock()
	if err := m.Compact(); err == nil {
		t.Fatal("empty identity accepted")
	}
	if m.Snapshot().Compacting || m.AnyBusy() {
		t.Fatal("rejected hydration created busy reservation")
	}
	var data struct{ Count int }
	response := compactCall(t, entry.client, "compact_count")
	_ = response.DecodeData(&data)
	if data.Count != 0 {
		t.Fatal("unready compact sent", data.Count)
	}
	compactCall(t, entry.client, pirpc.GetState)
	waitFor(t, m, func(s Snapshot) bool { return s.ActiveSession.ID == id })
	if err := m.Compact(); err != nil {
		t.Fatal("hydrated retry rejected", err)
	}
	waitCompactReceived(t, entry.client)
	compactCall(t, entry.client, "release_compact")
	waitFor(t, m, func(s Snapshot) bool { return !s.Compacting })
}

// Pinned settings.rs::compact_now only guards is_compacting. The context
// action is offered whenever a client exists, even while a turn/retry is busy.
func TestCompactSourceAllowsOtherBusyPhasesWithoutClearingThem(t *testing.T) {
	for _, phase := range []string{"running", "awaiting", "retrying"} {
		t.Run(phase, func(t *testing.T) {
			m, entry := compactFixture(t, 2*time.Second)
			m.mu.Lock()
			switch phase {
			case "running":
				entry.state.snapshot.Running = true
			case "awaiting":
				entry.state.awaiting = true
			case "retrying":
				entry.state.snapshot.Retrying = true
			}
			m.mu.Unlock()
			if err := m.Compact(); err != nil {
				t.Fatal(err)
			}
			waitCompactReceived(t, entry.client)
			compactCall(t, entry.client, "release_compact")
			waitFor(t, m, func(s Snapshot) bool { return !s.Compacting })
			s := m.Snapshot()
			if phase == "running" && !s.Running || phase == "awaiting" && !s.Awaiting || phase == "retrying" && !s.Retrying {
				t.Fatal("compaction changed another phase", phase, s)
			}
		})
	}
}
