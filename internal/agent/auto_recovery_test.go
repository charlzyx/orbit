package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// A disposable stdio child, never Pi/provider authentication or a tool runner.
func TestAutoRecoveryProcess(t *testing.T) {
	if os.Getenv("ORBIT_RECOVERY_HELPER") != "1" {
		return
	}
	trace := func(kind string, c map[string]any) {
		record := map[string]any{"pid": os.Getpid(), "kind": kind, "at": time.Now().UnixNano(), "command": c}
		b, _ := json.Marshal(record)
		f, err := os.OpenFile(os.Getenv("ORBIT_RECOVERY_TRACE"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			os.Exit(91)
		}
		_, _ = f.Write(append(b, '\n'))
		_ = f.Close()
	}
	write := func(v any) { b, _ := json.Marshal(v); fmt.Println(string(b)) }
	trace("spawn", nil)
	scan := bufio.NewScanner(os.Stdin)
	path := ""
	for scan.Scan() {
		var c map[string]any
		if json.Unmarshal(scan.Bytes(), &c) != nil {
			continue
		}
		kind, _ := c["type"].(string)
		trace(kind, c)
		r := map[string]any{"type": "response", "id": c["id"], "command": kind, "success": true}
		switch kind {
		case pirpc.GetState:
			r["data"] = map[string]any{"sessionId": "recovery-session", "sessionFile": path, "sessionName": "fixture"}
		case pirpc.SwitchSession:
			path, _ = c["sessionPath"].(string)
			mode, _ := os.ReadFile(os.Getenv("ORBIT_RECOVERY_MODE"))
			for string(mode) == "hold-switch" {
				time.Sleep(5 * time.Millisecond)
				mode, _ = os.ReadFile(os.Getenv("ORBIT_RECOVERY_MODE"))
			}
			if string(mode) == "fail-switch" {
				r["success"] = false
				r["error"] = "fixture reattach rejected"
			} else {
				r["data"] = map[string]any{"cancelled": string(mode) == "cancel-switch"}
			}
		case pirpc.Prompt:
			r["data"] = map[string]any{"disposition": "started"}
			write(r)
			write(map[string]any{"type": "agent_start"})
			continue
		case "queue":
			write(map[string]any{"type": "queue_update", "steering": []string{"must not replay"}, "followUp": []string{"must not replay either"}})
		case "crash":
			fmt.Fprintln(os.Stderr, "fixture intentional exit")
			os.Exit(7)
		case "break_stdin":
			_ = os.Stdin.Close()
			write(r)
			time.Sleep(time.Minute)
			os.Exit(8)
		}
		write(r)
	}
	trace("exit", nil)
	os.Exit(0)
}
func recoveryManager(t *testing.T) (*Manager, string, string) {
	t.Helper()
	trace := filepath.Join(t.TempDir(), "trace.jsonl")
	mode := filepath.Join(t.TempDir(), "mode")
	m := testManager(t, func(o *Options) {
		o.PiPrefixArgs = []string{"-test.run=^TestAutoRecoveryProcess$", "--"}
		o.EnvOverrides = map[string]string{"ORBIT_RECOVERY_HELPER": "1", "ORBIT_RECOVERY_TRACE": trace, "ORBIT_RECOVERY_MODE": mode}
		o.DisableCheckpoints = true
	})
	return m, trace, mode
}

type recoveryTrace struct {
	PID     int
	Kind    string
	At      int64
	Command map[string]any
}

func readRecoveryTrace(t *testing.T, path string) []recoveryTrace {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out []recoveryTrace
	scan := bufio.NewScanner(bytes.NewReader(b))
	for scan.Scan() {
		var r recoveryTrace
		if err := json.Unmarshal(scan.Bytes(), &r); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}
func TestAutoRecoveryReattachesWithoutReplayingBusiness(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	before := waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("one explicit fixture prompt", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	if err := m.Send("queue", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return len(s.Queue.FollowUp) == 1 })
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	after := waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != before.RuntimeID })
	if after.ActiveSession.Path != path || after.Running || after.Awaiting || len(after.Queue.FollowUp)+len(after.Queue.Steering) != 0 {
		t.Fatal("reattach/reset mismatch", after)
	}
	counts := map[string]int{}
	for _, r := range readRecoveryTrace(t, trace) {
		counts[r.Kind]++
		if r.Kind == pirpc.SwitchSession && r.Command["sessionPath"] != path {
			t.Fatal("wrong reattach path", r)
		}
	}
	if counts["spawn"] != 2 || counts[pirpc.SwitchSession] != 2 || counts[pirpc.Prompt] != 1 || counts[pirpc.Steer] != 0 || counts[pirpc.FollowUp] != 0 {
		t.Fatal("replay or wrong spawn/reattach count", counts)
	}
}

func TestAutoRecoveryCrashBudgetAndManualRestartPreserveLineage(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	for attempt := 1; attempt <= 4; attempt++ {
		old := m.Snapshot().RuntimeID
		if err := m.Send("crash", nil); err != nil {
			t.Fatal(err)
		}
		if attempt <= 3 {
			waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != old })
		} else {
			waitFor(t, m, func(s Snapshot) bool { return !s.Connected && s.ConnectionError != "" })
		}
	}
	m.mu.Lock()
	e := m.sessions[m.active]
	attempts, due := e.recovery.attempts, e.recovery.due
	m.mu.Unlock()
	if attempts != 3 || !due.IsZero() {
		t.Fatal("crash budget reset/leaked", attempts, due)
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 4 {
		t.Fatal("expected original plus three automatic attempts", count)
	}
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	m.mu.Lock()
	attempts = m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	if attempts != 3 {
		t.Fatal("manual restart refilled unhealthy lineage", attempts)
	}
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Connected && s.ConnectionError != "" })
	m.mu.Lock()
	due = m.sessions[m.active].recovery.due
	m.mu.Unlock()
	if !due.IsZero() {
		t.Fatal("manual restart enabled another automatic loop")
	}
}
func TestAutoRecoveryStopCancelsPendingAttempt(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Status == "Reconnecting" })
	m.mu.Lock()
	e := m.sessions[m.active]
	lease := recoveryLease{m.active, e, e.recovery.generation, e.recovery.due}
	m.mu.Unlock()
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := m.restartWithRecovery("", "", "", false, &lease); err != ErrReloadSuperseded {
		t.Fatal("stale timer admitted", err)
	}
	m.Tick()
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("Stop respawned child", count)
	}
	m.mu.Lock()
	attempts, due := e.recovery.attempts, e.recovery.due
	m.mu.Unlock()
	if attempts != 0 || !due.IsZero() {
		t.Fatal("cancel spent a future attempt", attempts, due)
	}
}
func TestAutoRecoveryNavigationCancelsPendingABA(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Status == "Reconnecting" })
	m.mu.Lock()
	e := m.sessions[m.active]
	lease := recoveryLease{m.active, e, e.recovery.generation, e.recovery.due}
	m.mu.Unlock()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.ActiveSession.Path == path })
	if err := m.restartWithRecovery("", "", "", false, &lease); err != ErrReloadSuperseded {
		t.Fatal("old ABA timer admitted", err)
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 3 {
		t.Fatal("navigation created an automatic extra child", count)
	}
}
func TestAutoRecoveryEmptyPathWithHistoryFailsClosed(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	if err := m.Submit("one explicit prompt", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		return !s.Connected && strings.Contains(s.ConnectionError, "recoverable session path")
	})
	m.mu.Lock()
	due := m.sessions[m.active].recovery.due
	m.mu.Unlock()
	if !due.IsZero() {
		t.Fatal("history without path scheduled fresh empty session")
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("unrecoverable conversation replaced", count)
	}
}
func TestAutoRecoveryHealthUsesActualExitNotDelayedDrain(t *testing.T) {
	start := time.Unix(100, 0)
	ready := start.Add(time.Second)
	if recoveryHealthy(ready, pirpc.ProcessSnapshot{StartedAt: start, ExitedAt: ready.Add(automaticRecoveryHealthy - time.Nanosecond)}) {
		t.Fatal("premature healthy refill")
	}
	if !recoveryHealthy(ready, pirpc.ProcessSnapshot{StartedAt: start, ExitedAt: ready.Add(automaticRecoveryHealthy)}) {
		t.Fatal("healthy boundary not accepted")
	}
	if recoveryHealthy(time.Time{}, pirpc.ProcessSnapshot{StartedAt: start, ExitedAt: start.Add(time.Hour)}) {
		t.Fatal("unready process counted healthy")
	}
}

func recoveryReady(t *testing.T, m *Manager) Snapshot {
	t.Helper()
	return waitFor(t, m, func(s Snapshot) bool { return s.Connected })
}
func exhaustThreeRecoveries(t *testing.T, m *Manager) {
	t.Helper()
	for i := 0; i < 3; i++ {
		old := m.Snapshot().RuntimeID
		if err := m.Send("crash", nil); err != nil {
			t.Fatal(err)
		}
		waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != old })
	}
}
func TestAutoRecoveryRealHealthyUptimeRefillsBudget(t *testing.T) {
	m, _, _ := recoveryManager(t)
	workspace := t.TempDir()
	if err := m.Open(workspace, filepath.Join(workspace, "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	exhaustThreeRecoveries(t, m)
	m.mu.Lock()
	ready := m.sessions[m.active].recovery.readyAt
	m.mu.Unlock()
	timer := time.NewTimer(time.Until(ready.Add(automaticRecoveryHealthy + 20*time.Millisecond)))
	defer timer.Stop()
	<-timer.C
	old := m.Snapshot().RuntimeID
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != old })
	m.mu.Lock()
	attempts := m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	if attempts != 1 {
		t.Fatal("actual healthy child did not refill budget", attempts)
	}
}
func TestAutoRecoveryRealQuickCrashDelayedTickCannotRefill(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	if err := m.Open(workspace, filepath.Join(workspace, "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	exhaustThreeRecoveries(t, m)
	m.mu.Lock()
	e := m.sessions[m.active]
	c := e.client
	ready := e.recovery.readyAt
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Call(ctx, pirpc.NewCommand("crash", nil))
	if err == nil {
		t.Fatal("fixture did not exit")
	}
	<-c.Done()
	process := c.ProcessSnapshot()
	if process.ExitedAt.Sub(ready) >= automaticRecoveryHealthy {
		t.Fatal("fixture was not a quick crash")
	}
	timer := time.NewTimer(time.Until(ready.Add(automaticRecoveryHealthy + 20*time.Millisecond)))
	defer timer.Stop()
	<-timer.C
	m.Tick()
	m.mu.Lock()
	attempts, due := e.recovery.attempts, e.recovery.due
	m.mu.Unlock()
	if attempts != 3 || !due.IsZero() {
		t.Fatal("delayed drain refilled short-lived child", attempts, due)
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 4 {
		t.Fatal("late drain spawned an extra process", count)
	}
}
func TestAutoRecoveryDuplicateTransportAndExitConsumesOnce(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	if err := m.Open(workspace, filepath.Join(workspace, "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	before := recoveryReady(t, m)
	m.mu.Lock()
	c := m.sessions[m.active].client
	m.mu.Unlock()
	p1Call(t, m, c, "break_stdin", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.Call(ctx, pirpc.NewCommand("force_write_failure", nil))
	if err == nil {
		t.Fatal("closed stdin accepted write")
	}
	<-c.Done()
	if !strings.Contains(c.Err().Error(), "write Pi stdin") {
		t.Fatal("fixture did not exercise transport fail", c.Err())
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != before.RuntimeID })
	m.mu.Lock()
	attempts := m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if attempts != 1 || count != 2 {
		t.Fatal("one death consumed multiple attempts", attempts, count)
	}
}
func TestAutoRecoveryParkedExitDoesNotRespawn(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	one := t.TempDir()
	if err := m.Open(one, filepath.Join(one, "one.jsonl")); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	m.mu.Lock()
	c := m.sessions[m.active].client
	m.mu.Unlock()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	current := recoveryReady(t, m)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = c.Call(ctx, pirpc.NewCommand("crash", nil))
	<-c.Done()
	m.Tick()
	if m.Snapshot().RuntimeID != current.RuntimeID {
		t.Fatal("parked death stole selection")
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 2 {
		t.Fatal("parked runtime revived", count)
	}
}
func TestAutoRecoveryLateSpawnStoppedBeforeSamePathRestart(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	var releaseOnce sync.Once
	unlockSpawn := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(unlockSpawn)
	m.mu.Lock()
	m.mcpFingerprint = func(string, string) [32]byte {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		return [32]byte{}
	}
	m.mu.Unlock()
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		m.Tick()
		select {
		case <-entered:
			goto held
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("recovery did not reach held spawn")
		}
		time.Sleep(5 * time.Millisecond)
	}
held:
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := m.Stop(); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	// Freeze only this test's owner lock after Stop. The old worker has passed
	// pre-spawn admission but cannot install its child, so actual child startup
	// is observable before releasing the post-spawn lease check.
	m.mu.Lock()
	locked := true
	defer func() {
		if locked {
			m.mu.Unlock()
		}
	}()
	unlockSpawn()
	latePID := 0
	deadline = time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		records := readRecoveryTrace(t, trace)
		n := 0
		for _, r := range records {
			if r.Kind == "spawn" {
				n++
				if n == 2 {
					latePID = r.PID
				}
			}
		}
		if latePID != 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	m.mu.Unlock()
	locked = false
	if latePID == 0 {
		t.Fatal("actual late child did not start")
	}
	after := recoveryReady(t, m)
	if after.ActiveSession.Path != path {
		t.Fatal("new explicit target changed")
	}
	process, err := os.FindProcess(latePID)
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Signal(syscall.Signal(0)); err == nil {
		t.Fatal("late auto child still alive after same-path explicit start")
	}
	records := readRecoveryTrace(t, trace)
	lateCommands := 0
	spawns := 0
	for _, r := range records {
		if r.Kind == "spawn" {
			spawns++
		}
		if r.PID == latePID && r.Kind != "spawn" && r.Kind != "exit" {
			lateCommands++
		}
	}
	if spawns != 3 || lateCommands != 0 {
		t.Fatal("late child admitted commands or extra writer", spawns, lateCommands)
	}
}

func TestAutoRecoverySwitchAcknowledgmentOwnsReady(t *testing.T) {
	m, trace, mode := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "saved.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	before := recoveryReady(t, m)
	if err := os.WriteFile(mode, []byte("hold-switch"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		count := 0
		for _, r := range readRecoveryTrace(t, trace) {
			if r.Kind == pirpc.SwitchSession {
				count++
			}
		}
		return s.RuntimeID != before.RuntimeID && count == 2
	})
	if m.Snapshot().Connected || m.Snapshot().Status == "Ready" {
		t.Fatal("ready before reattach acknowledgment")
	}
	if err := os.WriteFile(mode, nil, 0600); err != nil {
		t.Fatal(err)
	}
	after := recoveryReady(t, m)
	if after.RuntimeID == before.RuntimeID || after.ActiveSession.Path != path {
		t.Fatal("wrong recovered identity", after.ActiveSession)
	}
}
func TestAutoRecoveryFailedOrCancelledReattachConsumesBudget(t *testing.T) {
	for _, failure := range []string{"fail-switch", "cancel-switch"} {
		t.Run(failure, func(t *testing.T) {
			m, trace, mode := recoveryManager(t)
			workspace := t.TempDir()
			path := filepath.Join(workspace, "saved.jsonl")
			if err := m.Open(workspace, path); err != nil {
				t.Fatal(err)
			}
			before := recoveryReady(t, m)
			if err := os.WriteFile(mode, []byte(failure), 0600); err != nil {
				t.Fatal(err)
			}
			if err := m.Send("crash", nil); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool {
				if s.RuntimeID != before.RuntimeID && s.Connected {
					t.Fatal("failed/cancelled reattach advertised ready")
				}
				m.mu.Lock()
				e := m.sessions[m.active]
				finished := e.recovery.attempts == 3 && e.recovery.due.IsZero() && !e.loading && !e.recovery.launching
				m.mu.Unlock()
				return finished
			})
			counts := map[string]int{}
			var lastPID int
			for _, r := range readRecoveryTrace(t, trace) {
				counts[r.Kind]++
				if r.Kind == "spawn" {
					lastPID = r.PID
				}
			}
			if counts["spawn"] != 4 || counts[pirpc.SwitchSession] != 4 || counts[pirpc.GetState] != 1 {
				t.Fatal("failed reattach escaped budget/bootstrap gate", counts)
			}
			process, _ := os.FindProcess(lastPID)
			if process.Signal(syscall.Signal(0)) == nil {
				t.Fatal("failed bootstrap child was not reaped")
			}
		})
	}
}
func holdNextRecoverySpawn(t *testing.T, m *Manager) (chan struct{}, func()) {
	t.Helper()
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	var once sync.Once
	releaseOnce := func() { once.Do(func() { close(release) }) }
	t.Cleanup(releaseOnce)
	m.mu.Lock()
	m.mcpFingerprint = func(string, string) [32]byte {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		return [32]byte{}
	}
	m.mu.Unlock()
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		m.Tick()
		select {
		case <-entered:
			return entered, releaseOnce
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("recovery spawn barrier not reached")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
func TestAutoRecoveryRetirementDoesNotBlockDifferentPath(t *testing.T) {
	m, _, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	_, release := holdNextRecoverySpawn(t, m)
	other := filepath.Join(workspace, "two.jsonl")
	if err := m.Open(workspace, other); err != nil {
		t.Fatal(err)
	}
	after := recoveryReady(t, m) // old held attempt is deliberately still blocked
	if after.ActiveSession.Path != other {
		t.Fatal("different path blocked or selected incorrectly")
	}
	release()
}
func TestAutoRecoveryRetirementWaitCanStopAndTimeout(t *testing.T) {
	for _, action := range []string{"stop", "timeout"} {
		t.Run(action, func(t *testing.T) {
			m, trace, _ := recoveryManager(t)
			workspace := t.TempDir()
			path := filepath.Join(workspace, "one.jsonl")
			if err := m.Open(workspace, path); err != nil {
				t.Fatal(err)
			}
			recoveryReady(t, m)
			_, release := holdNextRecoverySpawn(t, m)
			if err := m.Open(workspace, path); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool {
				m.mu.Lock()
				waiting := m.sessions[m.active].recoveryWaitCancel != nil
				m.mu.Unlock()
				return waiting
			})
			if action == "stop" {
				if err := m.Stop(); err != nil {
					t.Fatal(err)
				}
				waitFor(t, m, func(s Snapshot) bool { return s.Status == "Stopped" })
			} else {
				waitFor(t, m, func(s Snapshot) bool { return strings.Contains(s.ConnectionError, "still closing") })
			}
			count := 0
			for _, r := range readRecoveryTrace(t, trace) {
				if r.Kind == "spawn" {
					count++
				}
			}
			if count != 1 {
				t.Fatal("wait admitted a second writer before retired spawn completed", count)
			}
			release()
			if action == "stop" {
				if m.Snapshot().Connected {
					t.Fatal("stopped wait became ready")
				}
				if got := m.Snapshot(); got.RuntimeID == "" || got.ActiveSession.Path != path {
					t.Fatal("Stop lost selected task identity", got.ActiveSession)
				}
			}
		})
	}
}
func TestAutoRecoveryCloseCancelsRetirementWait(t *testing.T) {
	m, _, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	_, release := holdNextRecoverySpawn(t, m)
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		m.mu.Lock()
		waiting := m.sessions[m.active].recoveryWaitCancel != nil
		m.mu.Unlock()
		return waiting
	})
	done := make(chan struct{})
	go func() { _ = m.Close(); close(done) }()
	// Close cancels the new wait via manager context, while the deliberately
	// held external spawn callback must still be released to reap its late child.
	select {
	case <-m.ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("Close did not cancel waits")
	}
	release()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Close left automatic worker behind")
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestAutoRecoverySpawnFailureConsumesSameBudget(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	if err := m.Open(workspace, filepath.Join(workspace, "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	m.configuredBinary.Store(filepath.Join(t.TempDir(), "missing-runtime"))
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		m.mu.Lock()
		e := m.sessions[m.active]
		done := e.recovery.attempts == 3 && e.recovery.due.IsZero() && !e.recovery.launching && !e.loading
		m.mu.Unlock()
		return done
	})
	if m.Snapshot().Connected {
		t.Fatal("failed spawn became ready")
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("nonexistent binary spawned helper", count)
	}
	m.configuredBinary.Store(m.options.PiBinary)
	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	m.mu.Lock()
	attempts := m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	if attempts != 3 {
		t.Fatal("manual repair silently replenished budget", attempts)
	}
}
func TestAutoRecoveryExplicitIntentCancelsDue(t *testing.T) {
	for _, action := range []string{"new_session", "manual_start", "close"} {
		t.Run(action, func(t *testing.T) {
			m, trace, _ := recoveryManager(t)
			workspace := t.TempDir()
			if err := m.Open(workspace, filepath.Join(workspace, "saved.jsonl")); err != nil {
				t.Fatal(err)
			}
			recoveryReady(t, m)
			if err := m.Send("crash", nil); err != nil {
				t.Fatal(err)
			}
			waitFor(t, m, func(s Snapshot) bool { return s.Status == "Reconnecting" })
			m.mu.Lock()
			e := m.sessions[m.active]
			lease := recoveryLease{m.active, e, e.recovery.generation, e.recovery.due}
			m.mu.Unlock()
			switch action {
			case "new_session":
				if err := m.NewSession(workspace); err != nil {
					t.Fatal(err)
				}
				recoveryReady(t, m)
			case "manual_start":
				if err := m.Start(); err != nil {
					t.Fatal(err)
				}
				recoveryReady(t, m)
			case "close":
				if err := m.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if err := m.restartWithRecovery("", "", "", false, &lease); err == nil {
				t.Fatal("explicit intent left old auto timer admissible")
			}
			count := 0
			for _, r := range readRecoveryTrace(t, trace) {
				if r.Kind == "spawn" {
					count++
				}
			}
			want := 2
			if action == "close" {
				want = 1
			}
			if count != want {
				t.Fatal("unexpected autonomous child", count, want)
			}
		})
	}
}
func TestAutoRecoveryMCPValidationAndExitHaveSingleReplacement(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	before := recoveryReady(t, m)
	m.mu.Lock()
	old := m.sessions[m.active].client
	m.mu.Unlock()
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	entered, release := make(chan struct{}), make(chan struct{})
	var calls atomic.Int32
	var once sync.Once
	releaseOnce := func() { once.Do(func() { close(release) }) }
	t.Cleanup(releaseOnce)
	m.mu.Lock()
	m.mcpFingerprint = func(string, string) [32]byte {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		return [32]byte{1}
	}
	m.mu.Unlock()
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("MCP validation was not held")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = old.Call(ctx, pirpc.NewCommand("crash", nil))
	<-old.Done()
	m.Tick()
	releaseOnce()
	after := waitFor(t, m, func(s Snapshot) bool {
		return s.Connected && s.ActiveSession.Path == path && s.RuntimeID != before.RuntimeID
	})
	if after.ActiveSession.Workspace != workspace {
		t.Fatal("MCP/recovery race selected wrong workspace")
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 3 {
		t.Fatal("MCP validation and recovery both replaced the runtime", count)
	}
}

func TestAutoRecoveryUnrelatedRetiredPathChangesDoNotBlock(t *testing.T) {
	for _, change := range []string{"removed", "renamed", "symlink_loop"} {
		t.Run(change, func(t *testing.T) {
			m, _, _ := recoveryManager(t)
			workspace := t.TempDir()
			path := filepath.Join(workspace, "one.jsonl")
			header, _ := json.Marshal(map[string]any{"type": "session", "id": "recovery-session", "cwd": workspace})
			if err := os.WriteFile(path, append(header, '\n'), 0600); err != nil {
				t.Fatal(err)
			}
			if err := m.Open(workspace, path); err != nil {
				t.Fatal(err)
			}
			recoveryReady(t, m)
			_, release := holdNextRecoverySpawn(t, m)
			if change == "renamed" {
				if err := os.Rename(path, path+".moved"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
				if change == "symlink_loop" {
					if err := os.Symlink(filepath.Base(path), path); err != nil {
						t.Fatal(err)
					}
				}
			}
			other := filepath.Join(workspace, "other.jsonl")
			if err := m.Open(workspace, other); err != nil {
				t.Fatal(err)
			}
			after := recoveryReady(t, m) // previous attempt deliberately remains held
			if after.ActiveSession.Path != other {
				t.Fatal("unrelated target did not become ready")
			}
			release()
		})
	}
}
func TestAutoRecoveryTerminalPolicyWinsLateCommandFailure(t *testing.T) {
	m, _, _ := recoveryManager(t)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	if err := m.Submit("one explicit prompt", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running })
	m.mu.Lock()
	e := m.sessions[m.active]
	c := e.client
	m.mu.Unlock()
	e.modelMu.Lock()
	locked := true
	defer func() {
		if locked {
			e.modelMu.Unlock()
		}
	}()
	if err := m.Send(pirpc.SetModel, map[string]any{"provider": "fixture", "modelId": "not-a-real-model"}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = c.Call(ctx, pirpc.NewCommand("crash", nil))
	<-c.Done()
	m.Tick()
	if !strings.Contains(m.Snapshot().ConnectionError, "recoverable session path") {
		t.Fatal("terminal policy not established before late completion")
	}
	e.modelMu.Unlock()
	locked = false
	drained := make(chan struct{})
	go func() { m.workers.Wait(); close(drained) }()
	select {
	case <-drained:
	case <-time.After(3 * time.Second):
		t.Fatal("late command worker did not drain")
	}
	if !strings.Contains(m.Snapshot().ConnectionError, "recoverable session path") {
		t.Fatal("late transport completion overwrote terminal policy", m.Snapshot().ConnectionError)
	}
}

func TestAutoRecoveryKnownHardlinkAliasWaitsForRetirement(t *testing.T) {
	m, trace, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	alias := filepath.Join(workspace, "alias.jsonl")
	header, _ := json.Marshal(map[string]any{"type": "session", "id": "recovery-session", "cwd": workspace})
	if err := os.WriteFile(path, append(header, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(path, alias); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	_, release := holdNextRecoverySpawn(t, m)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Base(path), path); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, alias); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool {
		m.mu.Lock()
		waiting := m.sessions[m.active].recoveryWaitCancel != nil
		m.mu.Unlock()
		return waiting
	})
	if m.Snapshot().Connected {
		t.Fatal("hardlink alias escaped retirement barrier")
	}
	count := 0
	for _, r := range readRecoveryTrace(t, trace) {
		if r.Kind == "spawn" {
			count++
		}
	}
	if count != 1 {
		t.Fatal("alias admitted second writer before retirement", count)
	}
	release()
	after := recoveryReady(t, m)
	if after.ActiveSession.Path != alias {
		t.Fatal("explicit alias identity was rewritten")
	}
}
func TestAutoRecoveryTargetDiskErrorIsNotHiddenByRetirement(t *testing.T) {
	m, _, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	_, release := holdNextRecoverySpawn(t, m)
	target := filepath.Join(workspace, "loop.jsonl")
	if err := os.Symlink(filepath.Base(target), target); err != nil {
		t.Fatal(err)
	}
	if err := m.Open(workspace, target); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Connected && s.ConnectionError != "" })
	if m.Snapshot().Status == "Ready" {
		t.Fatal("target disk error hidden as success")
	}
	release()
}
func TestAutoRecoveryExplicitReopenAfterRetirementStartsNewLineage(t *testing.T) {
	m, _, _ := recoveryManager(t)
	workspace := t.TempDir()
	path := filepath.Join(workspace, "one.jsonl")
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	recoveryReady(t, m)
	exhaustThreeRecoveries(t, m)
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Connected && s.ConnectionError != "" })
	m.mu.Lock()
	oldKey := m.active
	old := m.sessions[oldKey]
	attempts := old.recovery.attempts
	m.mu.Unlock()
	if attempts != 3 || old.client.Alive() {
		t.Fatal("old lineage not exhausted/reaped")
	}
	if err := m.Open(workspace, path); err != nil {
		t.Fatal(err)
	}
	after := recoveryReady(t, m)
	m.mu.Lock()
	_, retained := m.sessions[oldKey]
	attempts = m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	if retained || attempts != 0 || after.RuntimeID == oldKey {
		t.Fatal("explicit new lineage did not retire old owner", retained, attempts)
	}
	if err := m.Send("crash", nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.RuntimeID != after.RuntimeID })
	m.mu.Lock()
	attempts = m.sessions[m.active].recovery.attempts
	m.mu.Unlock()
	if attempts != 1 {
		t.Fatal("explicit new lineage did not receive bounded fresh budget", attempts)
	}
}
