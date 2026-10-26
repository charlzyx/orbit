package agent

import (
	"context"
	"errors"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// Compact reserves the active runtime synchronously, before the asynchronous
// RPC write, so repeated clicks and maintenance cannot race the first event.
func (m *Manager) Compact() error {
	m.mu.Lock()
	key := m.active
	entry := m.sessions[key]
	if m.closed || entry == nil || entry.client == nil || entry.loading || entry.stopped || !entry.client.Alive() || !entry.state.snapshot.Connected || entry.state.snapshot.ActiveSession.ID == "" {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	if entry.state.snapshot.Compacting || entry.state.manualCompactPending {
		m.mu.Unlock()
		return nil
	}
	client := entry.client
	m.sequence++
	origin := abortOrigin{sessionID: entry.state.snapshot.ActiveSession.ID, reset: entry.workflowGeneration, nonce: m.sequence}
	entry.state.compactionNonce = origin.nonce
	entry.state.snapshot.Compacting = true
	entry.state.manualCompactPending = true
	m.mu.Unlock()
	if !m.launch(func() {
		ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
		_, err := client.Call(ctx, origin.command(pirpc.Compact))
		cancel()
		m.mu.Lock()
		if !m.closed && m.sessions[key] == entry && entry.client == client && origin.matches(entry) && entry.state.compactionNonce == origin.nonce {
			var rejected *pirpc.CommandError
			// Timeout does not establish that the child stopped compacting. Retain
			// busy until a terminal event, Stop, or a definite response arrives.
			if err == nil || errors.As(err, &rejected) || errors.Is(err, pirpc.ErrClosed) {
				entry.state.snapshot.Compacting = false
				entry.state.manualCompactPending = false
			}
			if err != nil && !errors.As(err, &rejected) && !entry.observedTerminal(client) {
				entry.state.snapshot.setCommandError(pirpc.Compact, err.Error())
			}
		}
		m.mu.Unlock()
		m.Tick()
		m.changed()
	}) {
		m.mu.Lock()
		if m.sessions[key] == entry && entry.client == client && origin.matches(entry) && entry.state.compactionNonce == origin.nonce {
			entry.state.snapshot.Compacting = false
			entry.state.manualCompactPending = false
		}
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	m.changed()
	return nil
}
