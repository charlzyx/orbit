package agent

import (
	"encoding/json"
	"github.com/charlzyx/orbit/internal/pirpc"
	"strings"
)

// Ownership is independent of provider text and UI language. The separate
// agent bit cannot collide with an arbitrary extension command's wire name.
type errorSource struct {
	Command string
	Agent   bool
}

func (s *Snapshot) setErrorFrom(source errorSource, message string) {
	s.setConnectionError(message)
	s.ConnectionErrorSource = source
}
func (s *Snapshot) clearErrorFrom(source errorSource) {
	if source != (errorSource{}) && source == s.ConnectionErrorSource {
		s.setConnectionError("")
	}
}

// These responses belong to dedicated capability/auth/cursor channels in
// upstream events.rs, not the ordinary command banner ownership lifecycle.
func dedicatedErrorCommand(command string) bool {
	return strings.HasPrefix(command, "auth.") || command == pirpc.QuotaList || command == pirpc.GetEntries
}
func (s *Snapshot) setCommandError(command, message string) {
	if dedicatedErrorCommand(command) {
		s.setConnectionError(message)
		return
	}
	s.setErrorFrom(errorSource{Command: command}, message)
}

// Exact transcript::message_error contract: any nonempty assistant error,
// except an explicit aborted message, irrespective of other stopReason values.
func assistantMessageError(raw json.RawMessage) string {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil {
		return ""
	}
	if nested, present := fields["message"]; present {
		fields = nil
		if json.Unmarshal(nested, &fields) != nil {
			return ""
		}
	}
	var role, stopReason, message string
	_ = json.Unmarshal(fields["role"], &role)
	_ = json.Unmarshal(fields["stopReason"], &stopReason)
	_ = json.Unmarshal(fields["errorMessage"], &message)
	if role != "assistant" || stopReason == "aborted" {
		return ""
	}
	return strings.TrimSpace(message)
}

// Projection sees the full envelope: missing message falls back to the bare
// event, while explicit null intentionally does not (upstream transcript.rs).
func (s *sessionState) applyMessageError(raw json.RawMessage) {
	if detail := assistantMessageError(raw); detail != "" {
		s.snapshot.setErrorFrom(errorSource{Agent: true}, detail)
	} else {
		s.snapshot.clearErrorFrom(errorSource{Agent: true})
	}
}
