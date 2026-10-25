package agent

// setConnectionError changes only presentation ownership metadata. Callers keep
// their existing error text, priority, timing and lifecycle decisions.
func (s *Snapshot) setConnectionError(message string) {
	s.ConnectionError = message
	s.RetryFailure = RetryFailure{}
}
func (s *Snapshot) setRetryFailure(message string, provided bool) {
	fallback := message
	if !provided {
		fallback = "pi exhausted its automatic retries"
	}
	s.setConnectionError("Automatic retry failed: " + fallback)
	s.RetryFailure = RetryFailure{Active: true, MessageProvided: provided, Message: message}
}
