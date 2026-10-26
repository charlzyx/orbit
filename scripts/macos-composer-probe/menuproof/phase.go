package menuproof

import (
	"bytes"
	"encoding/json"
	"io"
)

type PhaseToken struct {
	Nonce  string `json:"nonce"`
	PID    int    `json:"pid"`
	Start  uint64 `json:"start_us"`
	Window uint32 `json:"window_number"`
	Phase  string `json:"phase"`
	Status string `json:"status,omitempty"`
}

func ValidatePhaseToken(raw []byte, expected PhaseToken, ack bool) error {
	if len(raw) > 4096 || expected.Nonce == "" || expected.PID < 2 || expected.Start == 0 || expected.Window == 0 {
		return ErrBlocked
	}
	switch expected.Phase {
	case "before", "seed", "final", "undo":
	default:
		return ErrBlocked
	}
	// Reject duplicate keys rather than allowing a later field to bless a token.
	d := json.NewDecoder(bytes.NewReader(raw))
	t, e := d.Token()
	if e != nil || t != json.Delim('{') {
		return ErrBlocked
	}
	seen := map[string]bool{}
	for d.More() {
		k, e := d.Token()
		key, ok := k.(string)
		if e != nil || !ok || seen[key] {
			return ErrBlocked
		}
		seen[key] = true
		var v json.RawMessage
		if d.Decode(&v) != nil {
			return ErrBlocked
		}
	}
	if _, e = d.Token(); e != nil {
		return ErrBlocked
	}
	var v PhaseToken
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || d.Decode(new(any)) != io.EOF {
		return ErrBlocked
	}
	expected.Status = ""
	if ack {
		expected.Status = "PASS"
	}
	if v != expected {
		return ErrBlocked
	}
	return nil
}
