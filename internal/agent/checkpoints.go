package agent

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/workspace"
)

// checkpointIdentityLocked reserves a read of persisted ending refs for a new
// Pi identity. The caller holds m.mu and launches the returned work after
// releasing it. Prompt dispatch is gated until the counter is recovered, so a
// fast first prompt cannot overwrite a preexisting session's turn-one refs.
func (m *Manager) checkpointIdentityLocked(entry *liveSession) func() {
	if m.options.DisableCheckpoints || entry.state.snapshot.ActiveSession.ID == "" {
		return nil
	}
	row := entry.state.snapshot.ActiveSession
	initialTurn := entry.state.snapshot.Turn
	entry.checkpointGeneration++
	generation := entry.checkpointGeneration
	entry.state.snapshot.CheckpointRecovering = true
	entry.state.snapshot.CheckpointRecoveryError = ""
	previous := entry.checkpointDone
	done := make(chan struct{})
	entry.checkpointDone = done
	return func() {
		var recovered int
		var err error
		defer func() {
			m.mu.Lock()
			current := entry.state.snapshot.ActiveSession
			if entry.checkpointGeneration == generation && current.ID == row.ID && workspace.SamePath(current.Workspace, row.Workspace) {
				if !entry.stopped && err == nil && m.ctx.Err() == nil && entry.state.snapshot.Turn == initialTurn && !entry.turnOpen {
					entry.state.snapshot.Turn = max(initialTurn, recovered)
					entry.state.snapshot.LatestTurn = recovered
				}
				if err != nil && m.ctx.Err() == nil {
					entry.state.snapshot.CheckpointRecoveryError = err.Error()
					entry.localDiagnostics = append(entry.localDiagnostics, "Review checkpoint recovery: "+err.Error())
					if len(entry.localDiagnostics) > 50 {
						entry.localDiagnostics = entry.localDiagnostics[len(entry.localDiagnostics)-50:]
					}
				}
				entry.state.snapshot.CheckpointRecovering = err != nil && m.ctx.Err() == nil
			}
			m.mu.Unlock()
			close(done)
			m.changed()
		}()
		if previous != nil {
			select {
			case <-previous:
			case <-m.ctx.Done():
				err = m.ctx.Err()
				return
			}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
		defer cancel()
		git := workspace.Git{Dir: row.Workspace}
		ok, repoErr := git.IsRepo(ctx)
		if repoErr != nil {
			var commandError *workspace.CommandError
			if errors.As(repoErr, &commandError) && commandError.ExitCode == 128 && strings.Contains(commandError.Stderr, "not a git repository") {
				return // Normal plain-folder session, not a failed ref recovery.
			}
			err = repoErr
			return
		}
		if ok {
			recovered, err = git.LatestTurn(ctx, row.ID)
		}
	}
}

// checkpointLocked reserves ordering while the state mutex is held. The
// returned work runs off-thread and serializes end/start capture across turns.
// Source session.rs keeps its in-flight counter and completed latest_turn
// distinct: only successful completion may advance the latter.
func (m *Manager) checkpointLocked(entry *liveSession, start bool) func() {
	if m.options.DisableCheckpoints {
		return nil
	}
	row, turn := entry.state.snapshot.ActiveSession, entry.state.snapshot.Turn
	if row.ID == "" {
		return nil
	}
	previous := entry.checkpointDone
	done := make(chan struct{})
	entry.checkpointDone = done
	return func() {
		defer close(done)
		if previous != nil {
			select {
			case <-previous:
			case <-m.ctx.Done():
				return
			}
		}
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()
		git := workspace.Git{Dir: row.Workspace}
		var err error
		if start {
			err = git.CaptureTurnStart(ctx, row.ID, turn)
		} else {
			err = git.CaptureTurn(ctx, row.ID, turn)
		}
		// Plain directories legitimately make capture a no-op. Do not advertise
		// a completed Git diff where there is no ending ref to read.
		completed := false
		if err == nil && !start && turn > 0 {
			_, resolveErr := git.ResolveCommit(ctx, workspace.CheckpointRef(row.ID, turn))
			completed = resolveErr == nil
		}
		m.mu.Lock()
		current := entry.state.snapshot.ActiveSession
		if !entry.stopped && current.ID == row.ID && workspace.SamePath(current.Workspace, row.Workspace) {
			if completed && turn > entry.state.snapshot.LatestTurn {
				entry.state.snapshot.LatestTurn = turn
			}
			if err != nil {
				entry.localDiagnostics = append(entry.localDiagnostics, "Review checkpoint: "+err.Error())
				if len(entry.localDiagnostics) > 50 {
					entry.localDiagnostics = entry.localDiagnostics[len(entry.localDiagnostics)-50:]
				}
			}
		}
		m.mu.Unlock()
		m.changed()
	}
}
