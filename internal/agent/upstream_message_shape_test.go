package agent

import (
	"encoding/json"
	"testing"
)

func TestUpstreamP1MessageShapeSourceContract(t *testing.T) {
	m := p1Manager(t)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	m.mu.Lock()
	c := m.sessions[m.active].client
	m.mu.Unlock()
	emit := func(raw string) {
		t.Helper()
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err != nil {
			t.Fatal(err)
		}
		p1Call(t, m, c, "emit_message_shape", map[string]any{"event": v})
	}
	for _, tc := range []struct{ name, raw, want string }{
		{"missing", `{"type":"message_end"}`, ""},
		{"null", `{"type":"message_end","message":null}`, ""},
		{"array", `{"type":"message_end","message":[]}`, ""},
		{"string", `{"type":"message_end","message":"bad"}`, ""},
		{"number", `{"type":"message_end","message":5}`, ""},
		{"empty_object", `{"type":"message_end","message":{}}`, ""},
		{"user", `{"type":"message_end","message":{"role":"user","errorMessage":"not agent"}}`, ""},
		{"aborted", `{"type":"message_end","message":{"role":"assistant","stopReason":"aborted","errorMessage":"abort"}}`, ""},
		{"bad_error_type", `{"type":"message_end","message":{"role":"assistant","errorMessage":3}}`, ""},
		{"nested_ignores_outer_types", `{"type":"message_end","role":3,"errorMessage":false,"message":{"role":"assistant","errorMessage":"nested error"}}`, "nested error"},
		{"malformed_stop_reason", `{"type":"message_end","message":{"role":"assistant","stopReason":3,"errorMessage":"real error"}}`, "real error"},
		{"bare_envelope", `{"type":"message_end","role":"assistant","errorMessage":" bare error "}`, "bare error"},
		{"explicit_null_no_fallback", `{"type":"message_end","role":"assistant","errorMessage":"outer","message":null}`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			emit(`{"type":"message_end","message":{"role":"assistant","errorMessage":"seed"}}`)
			if m.Snapshot().ConnectionError != "seed" {
				t.Fatal("seed failed")
			}
			emit(tc.raw)
			if got := m.Snapshot().ConnectionError; got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if tc.want == "" {
				emit(`{"type":"response","command":"set_model","success":false,"error":"command-owned"}`)
				emit(tc.raw)
				if m.Snapshot().ConnectionError != "command-owned" {
					t.Fatal("message None cleared unrelated source")
				}
			}
		})
	}
}
func TestUpstreamP1InvalidRecordCannotBecomeSuccessfulMessage(t *testing.T) {
	m := p1Manager(t)
	if err := m.Open(t.TempDir(), ""); err != nil {
		t.Fatal(err)
	}
	waitFor(t, m, func(s Snapshot) bool { return s.Connected })
	m.mu.Lock()
	c := m.sessions[m.active].client
	m.mu.Unlock()
	for _, raw := range []string{`{"type":"message_end"`, `null`, `[]`, `"text"`, `{}`, `{"type":4}`} {
		p1Call(t, m, c, "emit_message_shape", map[string]any{"event": map[string]any{"type": "message_end", "message": map[string]any{"role": "assistant", "errorMessage": "seed"}}})
		p1Call(t, m, c, "emit_message_shape", map[string]any{"raw": raw})
		if got := m.Snapshot().ConnectionError; got == "" {
			t.Fatalf("invalid record %s cleared error", raw)
		}
	}
}
