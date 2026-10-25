package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/workspace"
)

// This child mutates the disposable repo immediately on receiving prompt,
// before reporting agent_start. That catches a baseline captured too late.
func TestCheckpointPromptWriterProcess(t *testing.T) {
	if os.Getenv("ORBIT_CHECKPOINT_WRITER") != "1" {
		return
	}
	scanner := bufio.NewScanner(os.Stdin)
	write := func(value any) {
		raw, _ := json.Marshal(value)
		fmt.Fprintln(os.Stdout, string(raw))
	}
	for scanner.Scan() {
		var command map[string]any
		if json.Unmarshal(scanner.Bytes(), &command) != nil {
			continue
		}
		kind, _ := command["type"].(string)
		response := map[string]any{"type": "response", "id": command["id"], "command": kind, "success": true, "data": map[string]any{}}
		switch kind {
		case "get_state":
			response["data"] = map[string]any{"sessionId": "prompt-order", "isStreaming": false}
		case "auth.list", "quota.list":
			response["success"], response["error"] = false, "Unknown command"
		case "prompt":
			if err := os.WriteFile(os.Getenv("ORBIT_CHECKPOINT_AGENT_FILE"), []byte("immediate tool mutation\n"), 0600); err != nil {
				os.Exit(2)
			}
			response["data"] = map[string]any{"disposition": "started"}
			write(response)
			write(map[string]any{"type": "agent_start"})
			write(map[string]any{"type": "agent_settled"})
			continue
		}
		write(response)
	}
	os.Exit(0)
}

func checkpointRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	g := workspace.Git{Dir: root}
	for _, args := range [][]string{{"init", "-b", "main"}, {"config", "user.name", "Checkpoint test"}, {"config", "user.email", "checkpoint@example.invalid"}, {"config", "commit.gpgsign", "false"}, {"config", "core.hooksPath", filepath.Join(root, "no-hooks")}} {
		if _, err := g.Run(context.Background(), args...); err != nil {
			t.Fatal(err)
		}
	}
	checkpointWrite(t, root, "seed", "seed\n")
	if _, err := g.Run(context.Background(), "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Run(context.Background(), "commit", "-m", "seed"); err != nil {
		t.Fatal(err)
	}
	return root
}

func checkpointWrite(t *testing.T, root, name, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, name), []byte(value), 0600); err != nil {
		t.Fatal(err)
	}
}

func checkpointFixture(t *testing.T, root string) (*Manager, *liveSession) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	entry := &liveSession{state: newState(root, "")}
	entry.state.snapshot.ActiveSession.ID = "fixture"
	m := &Manager{ctx: ctx, cancel: cancel, sessions: map[string]*liveSession{"fixture": entry}, active: "fixture"}
	return m, entry
}

func runCheckpoint(t *testing.T, m *Manager, entry *liveSession, start bool) {
	t.Helper()
	m.mu.Lock()
	work := m.checkpointLocked(entry, start)
	m.mu.Unlock()
	if work == nil {
		t.Fatal("checkpoint unexpectedly disabled")
	}
	work()
}

func TestCheckpointCompletedCounterAdvancesOnlyAfterSuccessfulEnd(t *testing.T) {
	root := checkpointRepository(t)
	m, entry := checkpointFixture(t, root)
	entry.state.snapshot.Turn = 1
	runCheckpoint(t, m, entry, true)
	if entry.state.snapshot.LatestTurn != 0 {
		t.Fatal("start advertised an unfinished turn")
	}
	checkpointWrite(t, root, "first", "first completed turn\n")
	runCheckpoint(t, m, entry, false)
	if entry.state.snapshot.LatestTurn != 1 {
		t.Fatal("successful end not published", entry.state.snapshot.LatestTurn, entry.localDiagnostics)
	}
	entry.state.snapshot.Turn = 2
	runCheckpoint(t, m, entry, true)
	checkpointWrite(t, root, "second", "second in flight\n")
	lock := filepath.Join(root, ".git", filepath.FromSlash(workspace.CheckpointRef("fixture", 2))) + ".lock"
	if err := os.WriteFile(lock, []byte("fixture lock"), 0600); err != nil {
		t.Fatal(err)
	}
	runCheckpoint(t, m, entry, false)
	if entry.state.snapshot.LatestTurn != 1 || len(entry.localDiagnostics) != 1 || !strings.Contains(entry.localDiagnostics[0], "Review checkpoint:") {
		t.Fatal("failed end replaced completed state", entry.state.snapshot.LatestTurn, entry.localDiagnostics)
	}
	if err := os.Remove(lock); err != nil {
		t.Fatal(err)
	}
	runCheckpoint(t, m, entry, false)
	if entry.state.snapshot.LatestTurn != 2 {
		t.Fatal("retry did not publish captured end", entry.state.snapshot.LatestTurn)
	}
	for i := 0; i < 3; i++ {
		runCheckpoint(t, m, entry, false)
		if entry.state.snapshot.LatestTurn != 2 {
			t.Fatal("repeated completion changed numbering")
		}
	}
}

func TestCheckpointCompletionDoesNotLeakAcrossSessionIdentity(t *testing.T) {
	root := checkpointRepository(t)
	m, entry := checkpointFixture(t, root)
	entry.state.snapshot.Turn = 1
	runCheckpoint(t, m, entry, true)
	m.mu.Lock()
	oldCompletion := m.checkpointLocked(entry, false)
	entry.state = newState(root, "")
	entry.state.snapshot.ActiveSession.ID = "replacement"
	m.mu.Unlock()
	oldCompletion()
	if entry.state.snapshot.LatestTurn != 0 || entry.state.snapshot.Turn != 0 {
		t.Fatal("old completion published into replacement session", entry.state.snapshot)
	}
	if _, err := (workspace.Git{Dir: root}).ResolveCommit(context.Background(), workspace.CheckpointRef("fixture", 1)); err != nil {
		t.Fatal("old session ending snapshot was lost", err)
	}
}

func TestCheckpointFirstEndFailureDoesNotEnableLastTurn(t *testing.T) {
	root := checkpointRepository(t)
	m, entry := checkpointFixture(t, root)
	entry.state.snapshot.Turn = 1
	runCheckpoint(t, m, entry, true)
	lock := filepath.Join(root, ".git", filepath.FromSlash(workspace.CheckpointRef("fixture", 1))) + ".lock"
	if err := os.WriteFile(lock, []byte("fixture lock"), 0600); err != nil {
		t.Fatal(err)
	}
	runCheckpoint(t, m, entry, false)
	if entry.state.snapshot.LatestTurn != 0 || len(entry.localDiagnostics) != 1 {
		t.Fatal("failed first end enabled Last Turn", entry.state.snapshot.LatestTurn, entry.localDiagnostics)
	}
}

func TestCheckpointPlainFolderDoesNotAdvertiseMissingDiff(t *testing.T) {
	m, entry := checkpointFixture(t, t.TempDir())
	entry.state.snapshot.Turn = 1
	runCheckpoint(t, m, entry, true)
	runCheckpoint(t, m, entry, false)
	if entry.state.snapshot.LatestTurn != 0 || len(entry.localDiagnostics) != 0 {
		t.Fatal("plain folder advertised a Git checkpoint or false error", entry.state.snapshot.LatestTurn, entry.localDiagnostics)
	}
}

func TestCheckpointIdentityResponseResetsBothTurnCounters(t *testing.T) {
	s := newState(t.TempDir(), "")
	s.snapshot.ActiveSession.ID = "previous"
	s.snapshot.Turn, s.snapshot.LatestTurn = 8, 7
	apply(t, &s, `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"replacement"}}`)
	if s.snapshot.Turn != 0 || s.snapshot.LatestTurn != 0 {
		t.Fatal("session identity retained previous counters", s.snapshot)
	}
}

func TestManagerRestoresCompletedTurnDuringNextPrompt(t *testing.T) {
	root := checkpointRepository(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	if err := g.CaptureTurnStart(ctx, "switched", 3); err != nil {
		t.Fatal(err)
	}
	checkpointWrite(t, root, "previous", "completed before restart\n")
	if err := g.CaptureTurn(ctx, "switched", 3); err != nil {
		t.Fatal(err)
	}
	m := testManager(t, nil)
	if err := m.Open(root, filepath.Join(t.TempDir(), "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && s.Turn == 3 && s.LatestTurn == 3 })
	if err := m.Submit("next prompt", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running && !s.Awaiting && s.Turn == 4 && s.LatestTurn == 3 })
	checkpointWrite(t, root, "next", "next turn change\n")
	if err := m.Send(pirpc.Abort, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return !s.Running && s.LatestTurn == 4 })
	d, err := g.CollectReview(ctx, workspace.ReviewRequest{Source: workspace.ReviewLastTurn, Session: "switched", Turn: 4})
	if err != nil || !strings.Contains(d.Patch, "+next turn change") || strings.Contains(d.Patch, "+completed before restart") {
		start, _ := g.Run(ctx, "show", "--format=fuller", workspace.TurnStartRef("switched", 4))
		end, _ := g.Run(ctx, "show", "--format=fuller", workspace.CheckpointRef("switched", 4))
		t.Fatal("next completed diff contains wrong turn", d.Patch, err, "start", string(start), "end", string(end), "snapshot", m.Snapshot())
	}
}

func TestCheckpointIdentityRecoveryRejectsStaleGenerationAndLiveTurn(t *testing.T) {
	root := checkpointRepository(t)
	g := workspace.Git{Dir: root}
	for id, turn := range map[string]int{"fixture": 3, "replacement": 7} {
		if err := g.CaptureTurn(context.Background(), id, turn); err != nil {
			t.Fatal(err)
		}
	}
	m, entry := checkpointFixture(t, root)
	first := m.checkpointIdentityLocked(entry)
	entry.state.snapshot.ActiveSession.ID = "replacement"
	second := m.checkpointIdentityLocked(entry)
	first()
	if !entry.state.snapshot.CheckpointRecovering || entry.state.snapshot.Turn != 0 {
		t.Fatal("stale recovery released/replaced newer identity", entry.state.snapshot)
	}
	second()
	if entry.state.snapshot.CheckpointRecovering || entry.state.snapshot.Turn != 7 || entry.state.snapshot.LatestTurn != 7 {
		t.Fatal("latest identity not recovered", entry.state.snapshot)
	}
	third := m.checkpointIdentityLocked(entry)
	// Defensive guard even if an unrelated caller bypassed the prompt gate.
	entry.state.snapshot.Turn = 8
	entry.turnOpen = true
	third()
	if entry.state.snapshot.Turn != 8 || entry.state.snapshot.LatestTurn != 7 {
		t.Fatal("late recovery rewound a live counter", entry.state.snapshot)
	}
}

func TestManagerPromptGatePreservesExistingRefDuringCheckpointRecovery(t *testing.T) {
	root := checkpointRepository(t)
	g := workspace.Git{Dir: root}
	ctx := context.Background()
	m := testManager(t, nil)
	if err := m.Open(root, filepath.Join(t.TempDir(), "saved.jsonl")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && !s.CheckpointRecovering })
	if err := g.CaptureTurn(ctx, "switched", 1); err != nil {
		t.Fatal(err)
	}
	before, err := g.ResolveCommit(ctx, workspace.CheckpointRef("switched", 1))
	if err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	recover := m.checkpointIdentityLocked(entry)
	m.mu.Unlock()
	if s := m.Snapshot(); s.Connected || !s.CheckpointRecovering || s.Status != "Recovering turn checkpoints" {
		t.Fatal("pending recovery looks ready", s)
	}
	if err := m.Submit("must remain a draft", nil, SendPrompt); err == nil {
		t.Fatal("prompt accepted before persisted counter recovery")
	}
	if s := m.Snapshot(); s.Turn != 0 || s.Awaiting {
		t.Fatal("rejected prompt reserved a turn", s)
	}
	recover()
	if s := m.Snapshot(); !s.Connected || s.CheckpointRecovering || s.Turn != 1 || s.LatestTurn != 1 {
		t.Fatal("recovered counter was not published before readiness", s)
	}
	if err := m.Submit("safe retry", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Running && !s.Awaiting && s.Turn == 2 })
	after, err := g.ResolveCommit(ctx, workspace.CheckpointRef("switched", 1))
	if err != nil || before != after {
		t.Fatal("previous completed ref overwritten", before, after, err)
	}
}

func TestCheckpointRecoveryFailureStaysGatedUntilSuccessfulRetry(t *testing.T) {
	root := checkpointRepository(t)
	g := workspace.Git{Dir: root}
	if err := g.CaptureTurn(context.Background(), "fixture", 4); err != nil {
		t.Fatal(err)
	}
	m, entry := checkpointFixture(t, root)
	// IsRepo still succeeds, but actual for-each-ref fails rather than yielding
	// a trustworthy zero. Only this disposable repository's metadata is changed.
	packed := filepath.Join(root, ".git", "packed-refs")
	if err := os.WriteFile(packed, []byte("invalid packed refs fixture\n"), 0600); err != nil {
		t.Fatal(err)
	}
	m.checkpointIdentityLocked(entry)()
	if !entry.state.snapshot.CheckpointRecovering || entry.state.snapshot.CheckpointRecoveryError == "" || entry.state.snapshot.Turn != 0 {
		t.Fatal("failed listing claimed a safe zero counter", entry.state.snapshot)
	}
	if err := os.Remove(packed); err != nil {
		t.Fatal(err)
	}
	m.checkpointIdentityLocked(entry)()
	if entry.state.snapshot.CheckpointRecovering || entry.state.snapshot.CheckpointRecoveryError != "" || entry.state.snapshot.Turn != 4 || entry.state.snapshot.LatestTurn != 4 {
		t.Fatal("recovery retry failed", entry.state.snapshot)
	}
}

func TestCheckpointStartPrecedesImmediatePiToolMutation(t *testing.T) {
	root := checkpointRepository(t)
	m := testManager(t, func(options *Options) {
		options.PiPrefixArgs = []string{"-test.run=^TestCheckpointPromptWriterProcess$", "--"}
		options.EnvOverrides["ORBIT_CHECKPOINT_WRITER"] = "1"
		options.EnvOverrides["ORBIT_CHECKPOINT_AGENT_FILE"] = filepath.Join(root, "immediate.txt")
	})
	if err := m.Open(root, ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected && !s.CheckpointRecovering })
	if err := m.Submit("write immediately", nil, SendPrompt); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.LatestTurn == 1 })
	d, err := (workspace.Git{Dir: root}).CollectReview(context.Background(), workspace.ReviewRequest{Source: workspace.ReviewLastTurn, Session: "prompt-order", Turn: 1})
	if err != nil || !strings.Contains(d.Patch, "+immediate tool mutation") {
		t.Fatal("immediate tool write was absorbed into pre-turn baseline", d.Patch, err)
	}
}
