package menuproof

import (
	"encoding/json"
	"testing"
)

func TestPhaseTokenExactRequestsAndAcknowledgment(t *testing.T) {
	for _, phase := range []string{"before", "seed", "final", "undo"} {
		for _, ack := range []bool{false, true} {
			v := PhaseToken{Nonce: "nonce", PID: 42, Start: 123, Window: 88, Phase: phase}
			if ack {
				v.Status = "PASS"
			}
			raw, _ := json.Marshal(v)
			if ValidatePhaseToken(raw, v, ack) != nil {
				t.Fatal(phase, ack)
			}
		}
	}
}
func TestPhaseTokenForeignMalformedDuplicateAndFailureRejected(t *testing.T) {
	v := PhaseToken{Nonce: "nonce", PID: 42, Start: 123, Window: 88, Phase: "seed"}
	for _, raw := range []string{`{}`, `{"nonce":"wrong","pid":42,"start_us":123,"window_number":88,"phase":"seed"}`, `{"nonce":"nonce","pid":43,"start_us":123,"window_number":88,"phase":"seed"}`, `{"nonce":"nonce","pid":42,"start_us":124,"window_number":88,"phase":"seed"}`, `{"nonce":"nonce","pid":42,"start_us":123,"window_number":89,"phase":"seed"}`, `{"nonce":"nonce","pid":42,"start_us":123,"window_number":88,"phase":"final"}`, `{"nonce":"nonce","pid":42,"pid":42,"start_us":123,"window_number":88,"phase":"seed"}`, `{"nonce":"nonce","pid":42,"start_us":123,"window_number":88,"phase":"seed","status":"BLOCKED"}`, `{"nonce":"nonce","pid":42,"start_us":123,"window_number":88,"phase":"seed","extra":1}`, `null`, `{} {}`} {
		if ValidatePhaseToken([]byte(raw), v, true) == nil {
			t.Fatal(raw)
		}
	}
	raw, _ := json.Marshal(v)
	if ValidatePhaseToken(raw, v, true) == nil {
		t.Fatal("missing PASS ack")
	}
}
