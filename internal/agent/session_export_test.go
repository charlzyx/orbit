package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

func TestSessionExportProcess(t *testing.T) {
	if os.Getenv("ORBIT_SESSION_EXPORT_HELPER") != "1" {
		return
	}
	scan := bufio.NewScanner(os.Stdin)
	id := "s-1"
	n := 1
	var held map[string]any
	write := func(v any) { b, _ := json.Marshal(v); fmt.Println(string(b)) }
	complete := func(c map[string]any) {
		kind, _ := c["type"].(string)
		if kind == os.Getenv("ORBIT_SESSION_REJECT") {
			write(map[string]any{"type": "response", "id": c["id"], "command": kind, "success": false, "error": "synthetic authoritative rejection"})
			return
		}
		data := map[string]any{}
		switch kind {
		case "session_test_probe":
			data["held"] = held != nil
		case pirpc.GetState:
			data = map[string]any{"sessionId": id, "isStreaming": false, "thinkingLevel": "off"}
		case pirpc.NewSession, pirpc.Fork, pirpc.Clone, pirpc.SwitchSession:
			if os.Getenv("ORBIT_SESSION_CANCEL") == "1" {
				data["cancelled"] = true
				break
			}
			n++
			id = fmt.Sprintf("s-%d", n)
			data["cancelled"] = false
		case pirpc.SetSessionName:
			if p := os.Getenv("ORBIT_SESSION_RENAME_LOG"); p != "" {
				name, _ := c["name"].(string)
				_ = os.WriteFile(p, []byte(id+":"+name), 0600)
			}
		case pirpc.ExportHTML:
			p, _ := c["outputPath"].(string)
			_ = os.WriteFile(p, []byte(id), 0600)
			data["path"] = p
		}
		replyKind := kind
		if kind == os.Getenv("ORBIT_SESSION_MISMATCH") {
			replyKind = "wrong-command"
		}
		write(map[string]any{"type": "response", "id": c["id"], "command": replyKind, "success": true, "data": data})
	}
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		if kind == os.Getenv("ORBIT_SESSION_HOLD") {
			held = c
			continue
		}
		if kind == "session_test_release" && held != nil {
			complete(held)
			held = nil
		}
		complete(c)
	}
	os.Exit(0)
}
func exportManager(t *testing.T, hold string) (*Manager, string) {
	t.Helper()
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_HOLD"] = hold
		o.CommandTimeout = 100 * time.Millisecond
	})
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-1" })
	return m, w
}
func exportTarget(t *testing.T, m *Manager) SessionExportTarget {
	t.Helper()
	v, e := m.CaptureSessionExportTarget()
	if e != nil {
		t.Fatal(e)
	}
	return v
}
func waitRemoteSessionHeld(t *testing.T, c *pirpc.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		r, err := c.Call(ctx, pirpc.NewCommand("session_test_probe", nil))
		if err != nil {
			t.Fatal(err)
		}
		var state struct{ Held bool }
		_ = r.DecodeData(&state)
		if state.Held {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("worker never reached remote hold")
		default:
		}
		time.Sleep(time.Millisecond)
	}
}
func releaseExportTest(t *testing.T, c *pirpc.Client) {
	t.Helper()
	waitRemoteSessionHeld(t, c)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := c.Call(ctx, pirpc.NewCommand("session_test_release", nil)); err != nil {
		t.Fatal(err)
	}
}
func waitExportLeaseGone(t *testing.T, m *Manager, e *liveSession) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		m.Tick()
		m.mu.Lock()
		gone := e.sessionCommand == nil
		m.mu.Unlock()
		if gone {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("session reservation leaked")
}
func TestSessionExportReservationTimeoutAndFreshTask(t *testing.T) {
	m, w := exportManager(t, pirpc.ExportHTML)
	target := exportTarget(t, m)
	p := filepath.Join(t.TempDir(), "exact.html")
	if err := m.ExportSessionFor(target, p); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	old := m.sessions[m.active]
	m.mu.Unlock()
	waitFor(t, m, func(s Snapshot) bool {
		for _, d := range s.Diagnostics {
			if strings.Contains(d, "timeout does not cancel") {
				return true
			}
		}
		return false
	})
	if err := m.ExportSessionFor(target, p); err == nil {
		t.Fatal("second export admitted")
	}
	for _, kind := range []string{pirpc.Fork, pirpc.Clone, pirpc.SwitchSession, pirpc.NewSession} {
		if err := m.Send(kind, nil); err == nil {
			t.Fatalf("%s bypassed export reservation", kind)
		}
	}
	s := m.Snapshot()
	if err := m.RestartIdle(s.RuntimeID, s.ActiveSession.Path, s.ActiveSession.ID); err != ErrReloadDeferred {
		t.Fatalf("MCP idle restart bypassed export: %v", err)
	}
	if err := m.NewSession(w); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != target.task.key })
	if m.Snapshot().RuntimeID == target.task.key {
		t.Fatal("new task reused exporting child")
	}
	releaseExportTest(t, target.client)
	waitExportLeaseGone(t, m, old)
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "s-1" {
		t.Fatalf("wrong task exported: %q %v", b, err)
	}
	if err := m.ExportSessionFor(target, p); err == nil {
		t.Fatal("parked old target accepted after UI switch")
	}
}
func TestSessionExportRejectsInFlightIdentityAndHydratedNewTask(t *testing.T) {
	for _, kind := range []string{pirpc.NewSession, pirpc.Fork, pirpc.Clone, pirpc.SwitchSession} {
		t.Run(kind, func(t *testing.T) {
			m, _ := exportManager(t, kind)
			target := exportTarget(t, m)
			if err := m.Send(kind, nil); err != nil {
				t.Fatal(err)
			}
			if _, err := m.CaptureSessionExportTarget(); err == nil {
				t.Fatal("capture during transition")
			}
			if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "wrong.html")); err == nil {
				t.Fatal("old target admitted")
			}
			m.mu.Lock()
			entry := m.sessions[m.active]
			m.mu.Unlock()
			releaseExportTest(t, target.client)
			waitExportLeaseGone(t, m, entry)
			waitFor(t, m, func(s Snapshot) bool { return s.ActiveSession.ID == "s-2" })
			if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "wrong.html")); err == nil {
				t.Fatal("old task survived transition")
			}
		})
	}
}
func TestSessionExportStopCloseAndRestartReleaseWaiter(t *testing.T) {
	for _, end := range []string{"stop", "close", "restart"} {
		t.Run(end, func(t *testing.T) {
			m, _ := exportManager(t, pirpc.ExportHTML)
			target := exportTarget(t, m)
			if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "unfinished.html")); err != nil {
				t.Fatal(err)
			}
			m.mu.Lock()
			entry := m.sessions[m.active]
			m.mu.Unlock()
			switch end {
			case "stop":
				if err := m.Stop(); err != nil {
					t.Fatal(err)
				}
			case "close":
				if err := m.Close(); err != nil {
					t.Fatal(err)
				}
			case "restart":
				if err := m.Restart(); err != nil {
					t.Fatal(err)
				}
			}
			waitExportLeaseGone(t, m, entry)
			select {
			case <-target.client.Done():
			case <-time.After(3 * time.Second):
				t.Fatal("child did not terminate")
			}
			if target.client.PendingCount() != 0 {
				t.Fatal("wire waiter leaked")
			}
		})
	}
}

type exportMarshalGate struct {
	entered chan struct{}
	release chan struct{}
}

func (g exportMarshalGate) MarshalJSON() ([]byte, error) {
	close(g.entered)
	<-g.release
	return []byte(`"entry"`), nil
}
func TestSessionExportIdentityIntentCannotCrossSameClientNewTaskDuringMarshal(t *testing.T) {
	m, w := exportManager(t, "")
	old := exportTarget(t, m)
	gate := exportMarshalGate{make(chan struct{}), make(chan struct{})}
	done := make(chan error, 1)
	go func() { done <- m.Send(pirpc.Fork, map[string]any{"entryId": gate}) }()
	select {
	case <-gate.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("marshal not reached")
	}
	if err := m.NewSession(w); err != nil {
		close(gate.release)
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.ID == "s-2" })
	close(gate.release)
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("old fork intent entered newer task")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("fork send stuck")
	}
	if err := m.ExportSessionFor(old, filepath.Join(t.TempDir(), "wrong.html")); err == nil {
		t.Fatal("old export target survived same-client reuse")
	}
}
func TestSessionExportPoolBoundAndOpenSamePath(t *testing.T) {
	m, w := exportManager(t, pirpc.ExportHTML)
	m.mu.Lock()
	m.options.MaxParked = 1
	m.mu.Unlock()
	for i := 0; i < 2; i++ {
		target := exportTarget(t, m)
		if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "held.html")); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "existing-session.jsonl")
		m.mu.Lock()
		m.sessions[m.active].state.snapshot.ActiveSession.Path = path
		m.mu.Unlock()
		if err := m.Open(w, path); err != nil {
			t.Fatal(err)
		}
		if err := m.Send(pirpc.Clone, nil); err == nil {
			t.Fatal("Open same path bypassed reservation")
		}
		if err := m.NewSession(w); err != nil {
			t.Fatal(err)
		}
		waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != target.task.key })
	}
	target := exportTarget(t, m)
	if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "excess.html")); err == nil {
		t.Fatal("pending export pool unbounded")
	}
	if err := m.NewSession(w); err != nil {
		t.Fatalf("full reservation pool blocked a fresh task: %v", err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != target.task.key })
}

func TestSessionExportObservedIdentityChangeButNotHydrationInvalidatesTarget(t *testing.T) {
	m, _ := exportManager(t, pirpc.ExportHTML)
	old := exportTarget(t, m)
	m.mu.Lock()
	entry := m.sessions[m.active]
	entry.state.snapshot.ActiveSession.ID = "extension-changed"
	m.mu.Unlock()
	if err := m.ExportSessionFor(old, filepath.Join(t.TempDir(), "wrong.html")); err == nil {
		t.Fatal("observed extension identity change ignored")
	}
	m.mu.Lock()
	entry.state.snapshot.ActiveSession.ID = ""
	entry.state.snapshot.ActiveSession.Path = ""
	m.mu.Unlock()
	draft := exportTarget(t, m)
	m.mu.Lock()
	entry.state.snapshot.ActiveSession.ID = "s-1"
	entry.state.snapshot.ActiveSession.Path = filepath.Join(t.TempDir(), "hydrated.jsonl")
	m.mu.Unlock()
	if err := m.ExportSessionFor(draft, filepath.Join(t.TempDir(), "hydrated.html")); err != nil {
		t.Fatalf("normal ID hydration discarded target: %v", err)
	}
	releaseExportTest(t, draft.client)
	waitExportLeaseGone(t, m, entry)
}

func TestSessionExportHeldIdentityStopRestartClose(t *testing.T) {
	for _, kind := range []string{"direct-new", pirpc.NewSession, pirpc.Fork, pirpc.Clone, pirpc.SwitchSession} {
		for _, end := range []string{"stop", "restart", "close"} {
			t.Run(kind+"/"+end, func(t *testing.T) {
				hold := kind
				if kind == "direct-new" {
					hold = pirpc.NewSession
				}
				m, w := exportManager(t, hold)
				target := exportTarget(t, m)
				var err error
				if kind == "direct-new" {
					err = m.NewSession(w)
				} else {
					err = m.Send(kind, nil)
				}
				if err != nil {
					t.Fatal(err)
				}
				m.mu.Lock()
				entry := m.sessions[target.task.key]
				m.mu.Unlock()
				// Wait for the worker's notice, proving its actual wire waiter stays held.
				deadline := time.Now().Add(2 * time.Second)
				for {
					m.mu.Lock()
					noticed := entry.sessionCommand != nil && entry.sessionCommand.notified
					m.mu.Unlock()
					if noticed {
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("identity command did not remain pending")
					}
					time.Sleep(time.Millisecond)
				}
				switch end {
				case "stop":
					err = m.Stop()
				case "restart":
					err = m.Restart()
				case "close":
					err = m.Close()
				}
				if err != nil {
					t.Fatal(err)
				}
				waitExportLeaseGone(t, m, entry)
				select {
				case <-target.client.Done():
				case <-time.After(3 * time.Second):
					t.Fatal("old child remained alive")
				}
				if target.client.PendingCount() != 0 {
					t.Fatal("held identity wire waiter leaked")
				}
				m.mu.Lock()
				connected, creating, loading := entry.state.snapshot.Connected, entry.creatingSession, entry.loading
				current := m.sessions[m.active]
				m.mu.Unlock()
				if end != "restart" && (connected || creating || loading) {
					t.Fatalf("identity result revived terminated entry: connected=%v creating=%v loading=%v", connected, creating, loading)
				}
				if end == "restart" {
					if current == entry {
						t.Fatal("restart reused old entry")
					}
					waitFor(t, m, func(s Snapshot) bool {
						return s.Connected && s.RuntimeID != target.task.key && s.ActiveSession.ID == "s-1"
					})
				}
			})
		}
	}
}
func TestSessionExportNewSessionLaunchRejectedAfterClose(t *testing.T) {
	var pointer atomic.Pointer[Manager]
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.OnChange = func() {
			current := pointer.Load()
			if current == nil {
				return
			}
			current.mu.Lock()
			if e := current.sessions[current.active]; e != nil && e.sessionCommand != nil && e.sessionCommand.kind == pirpc.NewSession {
				current.closed = true
				current.cancel()
			}
			current.mu.Unlock()
		}
	})
	pointer.Store(m)
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	// Close admission at the synchronous changed notification, before launch.
	// This models Close's locked admission step; restore only to reap children.
	err := m.NewSession(w)
	if err != pirpc.ErrClosed {
		t.Fatalf("launch unexpectedly accepted: %v", err)
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	bad := entry.sessionCommand != nil || entry.loading || entry.creatingSession || entry.state.snapshot.Connected
	m.closed = false
	m.mu.Unlock()
	if bad {
		t.Fatal("rejected launch retained live/loading task state")
	}
	_ = m.Close()
}

func TestSessionExportCancelledNewSessionCannotRestoreStoppedState(t *testing.T) {
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_HOLD"] = pirpc.NewSession
		o.EnvOverrides["ORBIT_SESSION_CANCEL"] = "1"
	})
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	old := exportTarget(t, m)
	if err := m.NewSession(w); err != nil {
		t.Fatal(err)
	}
	waitRemoteSessionHeld(t, old.client)
	// Pin the result application behind the state lock. Stop has admitted its
	// terminal state, but the child can still finish while asynchronous Close
	// has not reached it. A cancelled reply must not restore the previous draft.
	m.mu.Lock()
	entry := m.sessions[m.active]
	func() {
		defer m.mu.Unlock()
		entry.stopped = true
		entry.loading = false
		entry.state.snapshot.Connected = false
		entry.state.snapshot.Status = "Stopped"
		entry.state.snapshot.ActiveSession.ID = "stopped-sentinel"
		releaseExportTest(t, old.client)
	}()
	waitExportLeaseGone(t, m, entry)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := old.client.Call(ctx, pirpc.NewCommand(pirpc.GetState, nil)); err != nil {
		t.Fatal(err)
	}
	m.Tick() // late hydration must not resurrect a stopped entry either
	m.mu.Lock()
	id, connected, status := entry.state.snapshot.ActiveSession.ID, entry.state.snapshot.Connected, entry.state.snapshot.Status
	m.mu.Unlock()
	if id != "stopped-sentinel" || connected || status != "Stopped" {
		t.Fatalf("cancelled result restored terminated state: %s %v %s", id, connected, status)
	}
}

func TestSessionExportProtocolMismatchRetainsReservationUntilRestart(t *testing.T) {
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_MISMATCH"] = pirpc.ExportHTML
		o.CommandTimeout = 50 * time.Millisecond
	})
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	target := exportTarget(t, m)
	if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "uncertain.html")); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	m.mu.Unlock()
	waitFor(t, m, func(s Snapshot) bool {
		for _, d := range s.Diagnostics {
			if strings.Contains(d, "timeout does not cancel") {
				return true
			}
		}
		return false
	})
	if err := m.ExportSessionFor(target, filepath.Join(t.TempDir(), "must-reject.html")); err == nil {
		t.Fatal("protocol mismatch falsely released operation")
	}
	if err := m.Restart(); err != nil {
		t.Fatal(err)
	}
	waitExportLeaseGone(t, m, entry)
	select {
	case <-target.client.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("uncertain child not reaped")
	}
}

func TestSessionExportMatchingRejectionReleasesReservation(t *testing.T) {
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestSessionExportProcess$", "--"}
		o.EnvOverrides["ORBIT_SESSION_EXPORT_HELPER"] = "1"
		o.EnvOverrides["ORBIT_SESSION_REJECT"] = pirpc.ExportHTML
	})
	w := t.TempDir()
	if err := m.Open(w, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	target := exportTarget(t, m)
	path := filepath.Join(t.TempDir(), "untouched.html")
	if err := m.ExportSessionFor(target, path); err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	m.mu.Unlock()
	waitExportLeaseGone(t, m, entry)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("rejected command wrote output")
	}
	if err := m.ExportSessionFor(target, path); err != nil {
		t.Fatalf("matching rejection quarantined child: %v", err)
	}
	waitExportLeaseGone(t, m, entry)
}
