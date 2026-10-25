package agent

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// All clocks/events below are explicit deterministic fixtures, not model usage.
func clockEvent(t *testing.T, s *sessionState, at time.Time, raw string) {
	t.Helper()
	e, err := pirpc.ParseEvent([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	e.ReceivedAt = at
	s.apply(e)
}
func clockDelta(t *testing.T, s *sessionState, at time.Time, kind string, index int, text string) {
	t.Helper()
	raw, _ := json.Marshal(map[string]any{"type": "message_update", "assistantMessageEvent": map[string]any{"type": kind, "contentIndex": index, "delta": text}})
	clockEvent(t, s, at, string(raw))
}
func TestStreamClockUsesReceiveIntervalsAndPreservesSettlement(t *testing.T) {
	s := newState("fixture", "")
	base := time.Unix(1000, 0)
	clockEvent(t, &s, base, `{"type":"message_start","message":{"id":"fixture-a","role":"assistant","content":[]}}`)
	clockDelta(t, &s, base.Add(time.Second), "thinking_delta", 0, "reason")
	clockDelta(t, &s, base.Add(1500*time.Millisecond), "thinking_delta", 0, "ing")
	clockDelta(t, &s, base.Add(2500*time.Millisecond), "text_delta", 1, "answer")       // gap == 1s counts
	clockDelta(t, &s, base.Add(4501*time.Millisecond), "text_delta", 1, " after stall") // gap > 1s excluded
	m := s.snapshot.Messages[0]
	if m.ThinkingDuration == nil || *m.ThinkingDuration != 1500*time.Millisecond || m.GenerationTime == nil || *m.GenerationTime != 1500*time.Millisecond {
		t.Fatalf("clocks %#v", m)
	}
	clockEvent(t, &s, base.Add(5*time.Second), `{"type":"message_end","message":{"role":"assistant","content":[{"type":"thinking","thinking":"reasoning"},{"type":"text","text":"answer after stall"}]}}`)
	m = s.snapshot.Messages[0]
	if *m.ThinkingDuration != 1500*time.Millisecond || *m.GenerationTime != 1500*time.Millisecond {
		t.Fatal("settled replacement lost clocks")
	}
	clockEvent(t, &s, base.Add(time.Hour), `{"type":"agent_settled"}`)
	if *s.snapshot.Messages[0].GenerationTime != 1500*time.Millisecond {
		t.Fatal("settle added tool/wall time")
	}
}
func TestStreamClockToolPauseResetWholeBlocksAndReload(t *testing.T) {
	s := newState("fixture", "")
	base := time.Unix(1000, 0)
	clockEvent(t, &s, base, `{"type":"message_start","message":{"id":"fixture-a","role":"assistant","content":[{"type":"thinking","thinking":"seed"}]}}`)
	clockDelta(t, &s, base.Add(200*time.Millisecond), "thinking_delta", 0, " more")
	clockEvent(t, &s, base.Add(300*time.Millisecond), `{"type":"message_update","assistantMessageEvent":{"type":"toolcall_start","contentIndex":1,"id":"t","toolName":"read"}}`)
	clockDelta(t, &s, base.Add(500*time.Millisecond), "toolcall_delta", 1, `{"path":`)
	if *s.snapshot.Messages[0].GenerationTime != 0 {
		t.Fatal("incomplete argument object published clock")
	}
	clockDelta(t, &s, base.Add(700*time.Millisecond), "toolcall_delta", 1, `"a.go"}`)
	if *s.snapshot.Messages[0].GenerationTime != 500*time.Millisecond || *s.snapshot.Messages[0].ThinkingDuration != 300*time.Millisecond {
		t.Fatal("argument/reasoning clocks differ from source")
	}
	clockEvent(t, &s, base.Add(800*time.Millisecond), `{"type":"tool_execution_start","toolCallId":"t","toolName":"read","args":{"path":"a.go"}}`)
	clockDelta(t, &s, base.Add(900*time.Millisecond), "text_delta", 2, "after tool")
	if *s.snapshot.Messages[0].GenerationTime != 500*time.Millisecond {
		t.Fatal("tool gap counted as decode")
	}
	clockEvent(t, &s, base.Add(time.Second), `{"type":"message_end","message":{"role":"assistant","content":"settled"}}`)
	clockEvent(t, &s, base.Add(2*time.Second), `{"type":"message_start","message":{"id":"fixture-b","role":"assistant","content":[]}}`)
	clockEvent(t, &s, base.Add(3*time.Second), `{"type":"message_update","assistantMessageEvent":{"type":"text_end","contentIndex":0,"content":"whole block"}}`)
	if s.snapshot.Messages[len(s.snapshot.Messages)-1].GenerationTime != nil {
		t.Fatal("whole block invented a decode interval")
	}
	clockEvent(t, &s, base.Add(4*time.Second), `{"type":"response","command":"get_messages","success":true,"data":{"messages":[{"role":"assistant","content":"history","timestamp":12345}]}}`)
	if s.snapshot.Messages[0].ThinkingDuration != nil || s.snapshot.Messages[0].GenerationTime != nil || s.streamClock.messageID != "" {
		t.Fatal("reload invented/preserved live clock")
	}
}
func TestStreamClockMissingFaultCancelRepeatAndClone(t *testing.T) {
	s := newState("fixture", "")
	base := time.Unix(1000, 0)
	apply(t, &s, `{"type":"message_start","message":{"role":"assistant","content":[]}}`)
	apply(t, &s, `{"type":"message_update","assistantMessageEvent":{"type":"thinking_delta","contentIndex":0,"delta":"unobserved"}}`)
	if s.snapshot.Messages[0].ThinkingDuration != nil || s.snapshot.Messages[0].GenerationTime != nil {
		t.Fatal("parsed fixture got fabricated receive clock")
	}
	clockDelta(t, &s, base, "thinking_delta", 0, "real")
	clockDelta(t, &s, base.Add(500*time.Millisecond), "thinking_delta", 0, "clock")
	clockDelta(t, &s, base.Add(500*time.Millisecond), "thinking_delta", 0, "repeat timestamp")
	clockDelta(t, &s, base.Add(250*time.Millisecond), "thinking_delta", 0, "regressed clock")
	if *s.snapshot.Messages[0].GenerationTime != 500*time.Millisecond || *s.snapshot.Messages[0].ThinkingDuration != 500*time.Millisecond {
		t.Fatal("repeat/regression inflated or decreased duration")
	}
	cloned := cloneValue(reflect.ValueOf(s.snapshot)).Interface().(Snapshot)
	*cloned.Messages[0].ThinkingDuration = 99 * time.Hour
	*cloned.Messages[0].GenerationTime = 99 * time.Hour
	if *s.snapshot.Messages[0].ThinkingDuration != 500*time.Millisecond || *s.snapshot.Messages[0].GenerationTime != 500*time.Millisecond {
		t.Fatal("snapshot shares clock pointers")
	}
	clockEvent(t, &s, base.Add(time.Hour), `{"type":"agent_end","willRetry":false}`)
	if *s.snapshot.Messages[0].ThinkingDuration != 500*time.Millisecond || s.streamClock.messageID != "" {
		t.Fatal("cancel without end invented final reasoning interval")
	}
	s.snapshot.ActiveSession.ID = "old"
	clockEvent(t, &s, base.Add(2*time.Hour), `{"type":"response","command":"get_state","success":true,"data":{"sessionId":"new"}}`)
	if len(s.snapshot.Messages) != 0 || s.streamClock.messageID != "" {
		t.Fatal("stale session clock")
	}
}

func TestStreamClockRequestStartsBeforeFirstMessageAndUsageSurvives(t *testing.T) {
	s := newState("fixture", "")
	base := time.Unix(1000, 0)
	clockEvent(t, &s, base, `{"type":"turn_start"}`)
	clockEvent(t, &s, base.Add(3*time.Second), `{"type":"message_start","message":{"id":"prefill","role":"assistant","content":[]}}`)
	if !s.snapshot.Messages[0].RequestStartedAt.Equal(base) {
		t.Fatal("request clock began at first token instead of actual turn_start")
	}
	clockEvent(t, &s, base.Add(4*time.Second), `{"type":"message_update","usage":{"input":4,"output":8},"assistantMessageEvent":{"type":"text_delta","contentIndex":0,"delta":"fixture"}}`)
	clockEvent(t, &s, base.Add(5*time.Second), `{"type":"message_end","message":{"role":"assistant","content":"fixture"}}`)
	if string(s.snapshot.Messages[0].Usage) != `{"input":4,"output":8}` || !s.snapshot.Messages[0].RequestStartedAt.Equal(base) {
		t.Fatal("authoritative end erased supplied usage or receive clock")
	}
	clockEvent(t, &s, base.Add(6*time.Second), `{"type":"turn_start"}`)
	clockEvent(t, &s, base.Add(7*time.Second), `{"type":"message_start","message":{"id":"next-request","role":"assistant","content":[]}}`)
	last := s.snapshot.Messages[len(s.snapshot.Messages)-1]
	if !last.RequestStartedAt.Equal(base.Add(6*time.Second)) || last.GenerationTime != nil {
		t.Fatal("new request inherited previous decode clock")
	}
}
