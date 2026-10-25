package agent

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

// Source: transcript.rs reset_generation/note_generation/pause_generation and
// begin_thinking/finish_thinking at pin299a32bc. Only local receive metadata
// drives these clocks; provider message timestamps never become live timers.
const generationGap = time.Second

type streamClock struct {
	messageID                  string
	index                      int
	thinkingStarted, lastDelta time.Time
	generation                 time.Duration
}

func (s *sessionState) clockMessage() *Message {
	c := &s.streamClock
	if c.messageID == "" || c.index < 0 || c.index >= len(s.snapshot.Messages) {
		return nil
	}
	m := &s.snapshot.Messages[c.index]
	if m.ID != c.messageID || m.Role != "assistant" {
		return nil
	}
	return m
}
func (c *streamClock) observeThinking(m *Message, at time.Time) {
	if at.IsZero() {
		return
	}
	if c.thinkingStarted.IsZero() {
		c.thinkingStarted = at
	}
	elapsed := max(time.Duration(0), at.Sub(c.thinkingStarted))
	if m.ThinkingDuration != nil {
		elapsed = max(elapsed, *m.ThinkingDuration)
	}
	m.ThinkingDuration = &elapsed
}
func (c *streamClock) finishThinking(m *Message, at time.Time) {
	if !c.thinkingStarted.IsZero() && !at.IsZero() && m.Thinking != "" {
		c.observeThinking(m, at)
	}
	c.thinkingStarted = time.Time{}
}
func (c *streamClock) observeGeneration(m *Message, at time.Time, publish bool) {
	if at.IsZero() {
		c.lastDelta = time.Time{}
		return
	}
	if !c.lastDelta.IsZero() && at.Before(c.lastDelta) {
		return
	}
	if !c.lastDelta.IsZero() {
		if gap := at.Sub(c.lastDelta); gap <= generationGap {
			c.generation += gap
		}
	}
	c.lastDelta = at
	if publish {
		value := c.generation
		m.GenerationTime = &value
	}
}
func (s *sessionState) beforeStreamClock(event pirpc.Event, at time.Time) {
	if event.Type == "message_end" {
		var value struct{ Role string }
		if event.DecodeField("message", &value) == nil && value.Role == "assistant" {
			if m := s.clockMessage(); m != nil {
				s.streamClock.finishThinking(m, at)
			}
		}
	}
}
func (s *sessionState) afterStreamClock(event pirpc.Event, at time.Time) {
	switch event.Type {
	case "turn_start":
		s.requestStarted = at
	case "message_start":
		var value struct{ Role string }
		if event.DecodeField("message", &value) != nil {
			return
		}
		if value.Role == "toolResult" {
			return
		}
		s.streamClock = streamClock{}
		if value.Role != "assistant" || s.activeMessage < 0 || s.activeMessage >= len(s.snapshot.Messages) {
			return
		}
		m := &s.snapshot.Messages[s.activeMessage]
		m.RequestStartedAt = s.requestStarted
		if m.RequestStartedAt.IsZero() {
			m.RequestStartedAt = at
		}
		s.streamClock = streamClock{messageID: m.ID, index: s.activeMessage}
		// A buffered start can already carry thinking, but no generation interval.
		if m.Thinking != "" && !at.IsZero() {
			s.streamClock.thinkingStarted = at
		}
	case "message_update":
		var delta struct {
			Type         string
			ContentIndex int
		}
		if event.DecodeField("assistantMessageEvent", &delta) != nil || delta.ContentIndex < 0 || delta.ContentIndex > 10000 {
			return
		}
		i := s.activeMessage
		if i < 0 || i >= len(s.snapshot.Messages) || s.snapshot.Messages[i].Role != "assistant" {
			return
		}
		m := &s.snapshot.Messages[i]
		if s.streamClock.messageID != m.ID {
			m.RequestStartedAt = s.requestStarted
			if m.RequestStartedAt.IsZero() {
				m.RequestStartedAt = at
			}
			s.streamClock = streamClock{messageID: m.ID, index: i}
		}
		c := &s.streamClock
		switch delta.Type {
		case "thinking_delta":
			c.observeThinking(m, at)
			c.observeGeneration(m, at, true)
		case "text_delta":
			c.finishThinking(m, at)
			c.observeGeneration(m, at, true)
		case "toolcall_start", "text_end", "thinking_end":
			c.finishThinking(m, at)
		case "toolcall_delta":
			// Rust accrues every argument delta, publishing once the accumulated
			// argument object parses; incomplete JSON is never presented as final.
			publish := false
			if delta.ContentIndex < len(m.Blocks) {
				raw := bytes.TrimSpace(m.Blocks[delta.ContentIndex].Arguments)
				publish = bytes.HasSuffix(raw, []byte("}")) && json.Valid(raw)
			}
			c.observeGeneration(m, at, publish)
		}
	case "tool_execution_start":
		s.streamClock.lastDelta = time.Time{}
	case "message_end":
		var value struct{ Role string }
		if event.DecodeField("message", &value) == nil && value.Role == "assistant" {
			s.streamClock = streamClock{}
		}
	case "agent_settled", "process_exited", "transport_error":
		// A missing message_end cannot invent the unobserved final interval.
		s.streamClock = streamClock{}
		s.requestStarted = time.Time{}
	case "agent_end":
		if !event.Bool("willRetry") {
			if m := s.clockMessage(); m != nil {
				m.RequestStartedAt = time.Time{}
			}
			s.requestStarted = time.Time{}
			s.streamClock = streamClock{}
		}
	case "response":
		if !event.Bool("success") {
			return
		}
		switch event.String("command") {
		case pirpc.GetMessages:
			s.streamClock = streamClock{} // Reload has no local delta history.
			s.requestStarted = time.Time{}
		case pirpc.GetState:
			if s.clockMessage() == nil {
				s.streamClock = streamClock{}
				s.requestStarted = time.Time{}
			}
		}
	}
}
