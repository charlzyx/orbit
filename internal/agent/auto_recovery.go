package agent

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

const automaticRecoveryLimit = 3
const automaticRecoveryDelay = 400 * time.Millisecond
const automaticRecoveryHealthy = 10 * time.Second

type automaticRecovery struct {
	attempts      int
	due, readyAt  time.Time
	generation    uint64
	launching     bool
	done          chan struct{}
	lastTerminal  *pirpc.Client
	identity      string
	identityInfo  os.FileInfo
	identityKnown bool
}
type recoveryRetirement struct {
	done            <-chan struct{}
	workspace, path string
	identity        string
	identityInfo    os.FileInfo
	identityKnown   bool
}

type recoveryLease struct {
	key        string
	entry      *liveSession
	generation uint64
	due        time.Time
}

// Only automatic startup has selection ownership. Ordinary Open may continue
// loading a parked task, as before; an obsolete automatic attempt may not.
func (m *Manager) startCurrentLocked(key string, entry *liveSession) bool {
	return !m.closed && m.sessions[key] == entry && !entry.stopped &&
		(!entry.recovery.launching || m.active == key && entry.recovery.generation == m.recoveryGeneration)
}
func (m *Manager) cancelRecoveryLocked() {
	m.recoveryGeneration++
	for _, entry := range m.sessions {
		entry.recovery.due = time.Time{}
		if !entry.recovery.launching || entry.stopped {
			continue
		}
		// Retire only an auto-owned bootstrap, never an ordinary parked process.
		entry.stopped = true
		entry.loading = false
		entry.state.snapshot.Connected = false
		entry.state.snapshot.Status = "Stopped"
		if entry.recovery.done != nil {
			found := false
			for _, r := range m.recoveryRetiring {
				if r.done == entry.recovery.done {
					found = true
				}
			}
			if !found {
				row := entry.state.snapshot.ActiveSession
				m.recoveryRetiring = append(m.recoveryRetiring, recoveryRetirement{entry.recovery.done, row.Workspace, row.Path, entry.recovery.identity, entry.recovery.identityInfo, entry.recovery.identityKnown})
			}
		}
		if c := entry.client; c != nil {
			m.workers.Add(1)
			go func() { defer m.workers.Done(); _ = c.Close() }()
		}
	}
}

// A cancelled automatic attempt can still be inside spawn. A subsequent start
// waits for its full reap off-thread, so same-path navigation cannot race two
// writers. This barrier never waits with m.mu or the UI thread held.
func (m *Manager) cancelRecoveryWaitLocked(entry *liveSession) {
	if entry != nil && entry.recoveryWaitCancel != nil {
		close(entry.recoveryWaitCancel)
		entry.recoveryWaitCancel = nil
	}
}
func (m *Manager) waitRecoveryRetirements(entry *liveSession) error {
	m.mu.Lock()
	row := entry.state.snapshot.ActiveSession
	var waiting []recoveryRetirement
	kept := m.recoveryRetiring[:0]
	for _, r := range m.recoveryRetiring {
		select {
		case <-r.done:
			continue
		default:
		}
		kept = append(kept, r)
		if r.done != entry.recovery.done {
			waiting = append(waiting, r)
		}
	}
	m.recoveryRetiring = kept
	cancel := make(chan struct{})
	entry.recoveryWaitCancel = cancel
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		if entry.recoveryWaitCancel == cancel {
			entry.recoveryWaitCancel = nil
		}
		m.mu.Unlock()
	}()
	if len(waiting) == 0 {
		return nil
	}
	timer := time.NewTimer(m.options.CommandTimeout)
	defer timer.Stop()
	for _, r := range waiting {
		same := false
		if (row.Path == "") == (r.path == "") {
			target := row.Path
			oldTarget := r.path
			if target == "" {
				target = row.Workspace
				oldTarget = r.workspace
			}
			if r.identityKnown {
				// Only the intended target is read now. Old canonical/inode identity was
				// captured before that automatic attempt could spawn, even for a then-
				// missing leaf. An unrelated old path may now be removed or unresolvable.
				path, info, err := readSessionFileIdentity(target)
				if err != nil {
					return err
				}
				same = sameSessionFile(path, info, r.identity, r.identityInfo)
			} else {
				// Cancellation preceded identity capture, and startCurrentLocked prevents
				// that attempt from ever spawning. Keep the direct same-target barrier;
				// there is no unknown live writer whose aliases need to be guessed.
				same = filepath.Clean(target) == filepath.Clean(oldTarget)
			}
		}
		if !same {
			continue
		}
		select {
		case <-r.done:
		case <-cancel:
			return errors.New("session start cancelled while previous automatic recovery was closing")
		case <-m.ctx.Done():
			return m.ctx.Err()
		case <-timer.C:
			return errors.New("previous automatic recovery is still closing; retry this session when it finishes")
		}
	}
	m.mu.Lock()
	m.retireDeadEntriesLocked(true)
	m.mu.Unlock()
	return nil
}
func recoveryHealthy(readyAt time.Time, process pirpc.ProcessSnapshot) bool {
	// ExitedAt, not drain time: a short-lived process whose event was queued
	// while a window was absent must not receive a fresh budget.
	return !readyAt.IsZero() && !process.ExitedAt.IsZero() &&
		!readyAt.Before(process.StartedAt) && process.ExitedAt.Sub(readyAt) >= automaticRecoveryHealthy
}
func (m *Manager) queueRecoveryLocked(key string, entry *liveSession, now time.Time) bool {
	if m.closed || m.active != key || m.sessions[key] != entry || entry.stopped || !entry.recovery.due.IsZero() {
		return false
	}
	if entry.recovery.launching && entry.recovery.generation != m.recoveryGeneration {
		return false
	}
	row := entry.state.snapshot.ActiveSession
	if row.Path == "" && (len(entry.state.snapshot.Messages) > 0 || entry.state.snapshot.Turn > 0) {
		entry.state.snapshot.setConnectionError("Pi exited without a recoverable session path; reconnect manually. No requests were replayed.")
		return false
	}
	if entry.recovery.attempts >= automaticRecoveryLimit {
		entry.state.snapshot.setConnectionError("Pi exited repeatedly; automatic recovery stopped after 3 attempts. Reconnect manually.")
		return false
	}
	entry.recovery.generation = m.recoveryGeneration
	entry.recovery.due = now.Add(automaticRecoveryDelay)
	entry.state.snapshot.setConnectionError("")
	entry.state.snapshot.Status = "Reconnecting"
	return true
}
func (m *Manager) observeRecoveryExitLocked(key string, entry *liveSession, now time.Time) {
	c := entry.client
	if c == nil || entry.recovery.lastTerminal == c {
		return
	}
	process := c.ProcessSnapshot()
	if process.ExitedAt.IsZero() {
		return
	} // a wire message is not proof of death
	entry.recovery.lastTerminal = c
	if recoveryHealthy(entry.recovery.readyAt, process) {
		entry.recovery.attempts = 0
	}
	m.queueRecoveryLocked(key, entry, now)
}
func (m *Manager) recoveryLeaseValidLocked(lease recoveryLease) bool {
	e := lease.entry
	return !m.closed && m.active == lease.key && m.sessions[lease.key] == e && !e.stopped &&
		m.recoveryGeneration == lease.generation && e.recovery.generation == lease.generation &&
		!e.recovery.due.IsZero() && e.recovery.due.Equal(lease.due) && !e.loading &&
		!e.restartClosing && e.recovery.attempts < automaticRecoveryLimit
}
func (m *Manager) clearDeadRuntimeLocked(entry *liveSession) {
	if entry.promptCancel != nil {
		entry.promptCancel()
	}
	entry.turnOpen = false
	entry.queueRestoreEpoch++
	entry.queueRestorePending = false
	entry.entriesPending = false
	entry.lastEntries = time.Time{}
	entry.state.snapshot.Queue = Queue{}
	for i := range entry.state.snapshot.Messages {
		entry.state.snapshot.Messages[i].Streaming = false
	}
}
func recoveryDiagnostics(entry *liveSession) []string {
	values := append([]string{}, entry.localDiagnostics...)
	if entry.client != nil {
		values = append(values, entry.client.RecentStderr()...)
	}
	if len(values) > 50 {
		values = values[len(values)-50:]
	}
	for i := range values {
		values[i] = redactDiagnostic(values[i], entry.redactions)
	}
	return values
}

// A command worker can complete after Tick has already reconciled this exact
// child's death. Its duplicate transport error must not replace the terminal
// recovery/budget/unsafe-session diagnostic; live-child command failures remain.
func (e *liveSession) observedTerminal(client *pirpc.Client) bool {
	return client != nil && e.client == client && e.recovery.lastTerminal == client
}

// Snapshot only on the startup worker. No filesystem lookup is performed by
// cancellation on the owner thread, and unknown identity can never spawn.
func (m *Manager) captureRecoveryIdentity(key string, entry *liveSession, workspace, path string) error {
	m.mu.Lock()
	automatic := entry.recovery.launching
	current := m.startCurrentLocked(key, entry)
	m.mu.Unlock()
	if !current {
		return errors.New("automatic recovery was superseded")
	}
	if !automatic {
		return nil
	}
	target := path
	if target == "" {
		target = workspace
	}
	canonical, info, err := readSessionFileIdentity(target)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.startCurrentLocked(key, entry) {
		return errors.New("automatic recovery was superseded")
	}
	entry.recovery.identity, entry.recovery.identityInfo, entry.recovery.identityKnown = canonical, info, true
	return nil
}

// Explicit Open is a new lineage only after previously owned children have
// exited and outstanding lifecycle leases are gone. This is not a global or
// historical retry-budget table; automatic replacement never uses this path.
func (m *Manager) retireDeadEntriesLocked(keepSelected bool) {
	for key, e := range m.sessions {
		if keepSelected && key == m.active {
			continue
		}
		if !e.loading && !e.busy() && !e.restartClosing && (e.client == nil || !e.client.Alive()) {
			delete(m.sessions, key)
		}
	}
}
