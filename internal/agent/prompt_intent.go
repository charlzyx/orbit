package agent

// Upstream #87 session.rs request_abort: an idle Escape cannot consume a
// future retry failure. The intent survives a late agent_start for this run.
func (s *sessionState) requestAbort() {
	if s.awaiting || s.snapshot.Running || s.snapshot.Retrying || s.snapshot.Compacting {
		s.abortRequested = true
	}
}
func (s *sessionState) ownsPrompt(id string) bool {
	return id != "" && id == s.promptInFlight
}
