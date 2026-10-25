package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
)

type sessionState struct {
	retrySequence        uint64
	compactionNonce      uint64
	manualCompactPending bool
	uiResponses          []pirpc.UIResponse
	askToolID            string
	askQuestions         []AskQuestion
	askReplay            *askReplay
	noticeSequence       uint64
	requestStarted       time.Time
	streamClock          streamClock
	snapshot             Snapshot
	sequence             uint64
	activeMessage        int
	entriesCursor        string
	awaiting             bool
}

func newState(workspace, path string) sessionState {
	return sessionState{activeMessage: -1, snapshot: Snapshot{ActiveSession: SessionRow{Path: path, Workspace: workspace, Name: "New task"}, ThinkingLevel: "off", SteeringMode: "one-at-a-time", FollowUpMode: "one-at-a-time", AutoCompactionEnabled: true, AccessMode: "full-access", WorkflowMode: "build", StatusEntries: map[string]string{}, Auth: AuthState{Support: "unknown"}, QuotaSupport: "unknown"}}
}
func (s *sessionState) apply(event pirpc.Event) {
	s.applyAt(event, event.ReceivedAt)
}

// applyAt is deterministic for protocol-fixture clocks. Production always uses
// the reader's ReceivedAt, never a drain/heartbeat timestamp.
func (s *sessionState) applyAt(event pirpc.Event, receivedAt time.Time) {
	s.beforeStreamClock(event, receivedAt)
	s.applyRecord(event)
	s.afterStreamClock(event, receivedAt)
}

func (s *sessionState) applyRecord(event pirpc.Event) {
	switch event.Type {
	case "response":
		s.applyResponse(event)
	case "agent_start":
		s.snapshot.Running = true
		s.awaiting = false
		s.snapshot.Status = "Running"
	case "agent_settled":
		s.snapshot.Running = false
		s.awaiting = false
		s.snapshot.Retrying = false
		s.snapshot.RetryDetail = RetryDetail{}
		s.snapshot.Queue = Queue{}
		s.snapshot.Status = "Ready"
		for i := range s.snapshot.Messages {
			s.snapshot.Messages[i].Streaming = false
		}
	case "agent_end":
		if event.Bool("willRetry") {
			s.snapshot.Retrying = true
		}
	case "message_start":
		s.startMessage(event.Fields["message"])
	case "message_end":
		s.endMessage(event.Fields["message"])
	case "message_update":
		s.updateMessage(event)
	case "tool_execution_start", "tool_execution_update", "tool_execution_end":
		s.askToolEvent(event)
		s.toolEvent(event)
	case "queue_update":
		_ = event.Decode(&s.snapshot.Queue)
	case "compaction_start":
		s.snapshot.Compacting = true
	case "compaction_end":
		s.snapshot.Compacting = false
		s.manualCompactPending = false
		if result := event.Fields["result"]; len(result) > 0 && string(result) != "null" && !event.Bool("aborted") {
			// Pi invalidates its context baseline after successful compaction.
			// Keep cumulative totals, but never flash the pre-compaction percent
			// while the asynchronous get_session_stats refresh is in flight.
			var stats map[string]json.RawMessage
			if json.Unmarshal(s.snapshot.Stats, &stats) == nil && stats != nil {
				delete(stats, "contextUsage")
				s.snapshot.Stats, _ = json.Marshal(stats)
			}
		}
	case "auto_retry_start":
		s.snapshot.Retrying = true
		s.retrySequence++
		s.snapshot.RetryDetail = retryDetail(event)
		s.snapshot.RetryDetail.Revision = s.retrySequence
	case "auto_retry_end":
		s.snapshot.Retrying = false
		s.snapshot.RetryDetail = RetryDetail{}
		// Only an explicit false is exhaustion; cancellation/omitted success
		// must not invent a terminal failure (pinned events.rs).
		if string(event.Fields["success"]) == "false" {
			var reported string
			provided := json.Unmarshal(event.Fields["finalError"], &reported) == nil && string(event.Fields["finalError"]) != "null"
			s.snapshot.setRetryFailure(reported, provided)
		}
	case "session_info_changed":
		s.snapshot.ActiveSession.Name = event.String("name")
	case "extension_ui_request":
		s.extensionUI(event)
	case "extension_error":
		s.snapshot.setConnectionError(event.String("error"))
		if s.snapshot.ConnectionError == "" {
			s.snapshot.setConnectionError(event.String("message"))
		}
	case "process_exited", "transport_error":
		s.clearExtensionUI()
		s.snapshot.Connected = false
		s.manualCompactPending = false
		s.snapshot.Running = false
		s.snapshot.Compacting = false
		s.snapshot.Retrying = false
		s.snapshot.RetryDetail = RetryDetail{}
		s.awaiting = false
		s.snapshot.Status = "Disconnected"
		if event.Type == "transport_error" || event.String("error") != pirpc.ErrClosed.Error() {
			s.snapshot.setConnectionError(event.String("error"))
		}
	case "protocol_error":
		s.snapshot.setConnectionError(event.String("error"))
		s.rememberUnknown(event.Raw)
	default:
		if strings.HasPrefix(event.Type, "auth_") {
			s.authEvent(event)
		} else if event.Type != "turn_start" && event.Type != "turn_end" {
			s.rememberUnknown(event.Raw)
		}
	}
	s.snapshot.ActiveSession.Running = s.snapshot.Running
}
func (s *sessionState) rememberUnknown(raw json.RawMessage) {
	if len(s.snapshot.UnknownEvents) == 100 {
		s.snapshot.UnknownEvents = s.snapshot.UnknownEvents[1:]
	}
	s.snapshot.UnknownEvents = append(s.snapshot.UnknownEvents, bytes.Clone(raw))
}
func (s *sessionState) applyResponse(event pirpc.Event) {
	var response pirpc.Response
	if event.Decode(&response) != nil {
		return
	}
	if response.Command == pirpc.Compact {
		s.manualCompactPending = false
		s.snapshot.Compacting = false
	}
	if !response.Success {
		if response.Command == pirpc.AuthList {
			s.snapshot.Auth.Support = "unsupported"
			return
		}
		if response.Command == pirpc.QuotaList {
			s.snapshot.QuotaSupport = "unsupported"
			return
		}
		s.snapshot.setConnectionError(response.Error)
		if response.Command == pirpc.Prompt {
			s.awaiting = false
		}
		return
	}
	data := response.Data
	switch response.Command {
	case pirpc.GetState:
		var state struct {
			Model                               ModelChoice
			ThinkingLevel                       string
			IsStreaming, IsCompacting           bool
			SessionFile, SessionID, SessionName string
			SteeringMode, FollowUpMode          string
			AutoCompactionEnabled               *bool
			Capabilities                        struct {
				ExtensionUI []string `json:"extension_ui"`
			}
		}
		if json.Unmarshal(data, &state) != nil {
			return
		}
		if s.snapshot.ActiveSession.ID != "" && state.SessionID != s.snapshot.ActiveSession.ID {
			s.snapshot.Turn = 0
			s.snapshot.LatestTurn = 0
			s.snapshot.CheckpointRecovering = false
			s.snapshot.CheckpointRecoveryError = ""
			s.snapshot.Messages = nil
			s.snapshot.Queue = Queue{}
			s.cancelExtensionUI()
			s.clearExtensionUI()
			s.entriesCursor = ""
			s.activeMessage = -1
		}
		s.snapshot.ActiveSession.ID = state.SessionID
		if state.SessionFile != "" {
			s.snapshot.ActiveSession.Path = state.SessionFile
		}
		if state.SessionName != "" {
			s.snapshot.ActiveSession.Name = state.SessionName
		}
		s.snapshot.Model = state.Model
		s.snapshot.ThinkingLevel = state.ThinkingLevel
		s.snapshot.Running = state.IsStreaming
		s.snapshot.Compacting = state.IsCompacting
		s.snapshot.Connected = true
		s.snapshot.Capabilities = state.Capabilities.ExtensionUI
		s.snapshot.SteeringMode = state.SteeringMode
		s.snapshot.FollowUpMode = state.FollowUpMode
		s.snapshot.AutoCompactionKnown = state.AutoCompactionEnabled != nil
		if state.AutoCompactionEnabled != nil {
			s.snapshot.AutoCompactionEnabled = *state.AutoCompactionEnabled
		}
	case pirpc.GetMessages:
		var value struct{ Messages []json.RawMessage }
		if json.Unmarshal(data, &value) != nil {
			return
		}
		s.snapshot.Messages = nil
		s.activeMessage = -1
		for _, raw := range value.Messages {
			s.appendMessage(raw, false)
		}
	case pirpc.GetAvailableModels:
		var value struct{ Models []json.RawMessage }
		if json.Unmarshal(data, &value) != nil {
			return
		}
		s.snapshot.Models = nil
		for _, raw := range value.Models {
			var model ModelChoice
			if json.Unmarshal(raw, &model) == nil {
				model.Raw = bytes.Clone(raw)
				s.snapshot.Models = append(s.snapshot.Models, model)
			}
		}
	case pirpc.SetModel:
		var model ModelChoice
		if json.Unmarshal(data, &model) == nil {
			s.snapshot.Model = model
		}
	case pirpc.CycleModel:
		var value struct {
			Model         ModelChoice
			ThinkingLevel string
		}
		if json.Unmarshal(data, &value) == nil {
			s.snapshot.Model = value.Model
			s.snapshot.ThinkingLevel = value.ThinkingLevel
		}
	case pirpc.GetAvailableThinkingLevels:
		var value struct{ Levels []string }
		if json.Unmarshal(data, &value) == nil {
			s.snapshot.ThinkingLevels = value.Levels
		}
	case pirpc.GetSessionStats:
		s.snapshot.Stats = bytes.Clone(data)
	case pirpc.GetCommands:
		var value struct{ Commands []json.RawMessage }
		if json.Unmarshal(data, &value) == nil {
			s.snapshot.Commands = value.Commands
		}
	case pirpc.ClearQueue:
		s.snapshot.Queue = Queue{}
	case pirpc.AuthList:
		var value struct{ Providers []AuthProvider }
		if json.Unmarshal(data, &value) == nil {
			s.snapshot.Auth.Support = "supported"
			s.snapshot.Auth.Providers = value.Providers
		}
	case pirpc.AuthStatus:
		var value AuthProvider
		if json.Unmarshal(data, &value) == nil {
			for i := range s.snapshot.Auth.Providers {
				if s.snapshot.Auth.Providers[i].ID == value.ID {
					s.snapshot.Auth.Providers[i].Credential = value.Credential
					s.snapshot.Auth.Providers[i].Authenticated = value.Authenticated
				}
			}
		}
	case pirpc.QuotaList:
		var value struct{ Providers []QuotaReport }
		if json.Unmarshal(data, &value) == nil {
			s.snapshot.QuotaSupport = "supported"
			s.snapshot.Quotas = value.Providers
		}
	case pirpc.GetEntries:
		s.applyEntries(data)
	case pirpc.Prompt:
		var value struct{ Disposition string }
		_ = json.Unmarshal(data, &value)
		if value.Disposition == "handled" {
			s.awaiting = false
		}
	}
}
func (s *sessionState) appendMessage(raw json.RawMessage, streaming bool) int {
	s.sequence++
	message := decodeMessage(raw, fmt.Sprintf("message-%d", s.sequence))
	message.Streaming = streaming
	s.snapshot.Messages = append(s.snapshot.Messages, message)
	return len(s.snapshot.Messages) - 1
}
func decodeMessage(raw json.RawMessage, fallbackID string) Message {
	var object map[string]json.RawMessage
	_ = json.Unmarshal(raw, &object)
	stringField := func(name string) string { var value string; _ = json.Unmarshal(object[name], &value); return value }
	result := Message{ID: stringField("id"), Role: stringField("role"), ToolCallID: stringField("toolCallId"), ToolName: stringField("toolName"), Error: stringField("errorMessage"), Raw: bytes.Clone(raw), Usage: bytes.Clone(object["usage"])}
	if result.ID == "" {
		result.ID = fallbackID
		if stamp, ok := object["timestamp"]; ok {
			result.ID = "message:" + result.Role + ":" + string(stamp) + ":" + result.ToolCallID
		}
	}
	var timestamp float64
	if json.Unmarshal(object["timestamp"], &timestamp) == nil && timestamp > 0 {
		if timestamp <= 1_000_000_000_000 {
			timestamp *= 1000
		}
		result.Timestamp = time.UnixMilli(int64(timestamp))
	} else {
		var value string
		if json.Unmarshal(object["timestamp"], &value) == nil {
			result.Timestamp, _ = time.Parse(time.RFC3339Nano, value)
		}
	}
	var text string
	if json.Unmarshal(object["content"], &text) == nil {
		result.Blocks = []ContentBlock{{Type: "text", Text: text}}
	} else {
		var blocks []json.RawMessage
		if json.Unmarshal(object["content"], &blocks) == nil {
			for _, rawBlock := range blocks {
				result.Blocks = append(result.Blocks, decodeBlock(rawBlock))
			}
		}
	}
	rebuildMessage(&result)
	if result.Role == "toolResult" {
		result.ToolOutput = result.Text
	}
	return result
}
func decodeBlock(raw json.RawMessage) ContentBlock {
	var value struct {
		Type, Text, Thinking, ID, Name, Data, MIMEType string
		Arguments                                      json.RawMessage
	}
	_ = json.Unmarshal(raw, &value)
	block := ContentBlock{Type: value.Type, Text: value.Text, Thinking: value.Thinking, ToolCallID: value.ID, ToolName: value.Name, Arguments: bytes.Clone(value.Arguments), Raw: bytes.Clone(raw)}
	if block.Type == "image" {
		block.Image = &Image{Type: "image", Data: value.Data, MIMEType: value.MIMEType}
	}
	return block
}
func rebuildMessage(message *Message) {
	var text, thinking strings.Builder
	for _, block := range message.Blocks {
		if block.Type == "text" {
			text.WriteString(block.Text)
		}
		if block.Type == "thinking" {
			thinking.WriteString(block.Thinking)
		}
		if block.Type == "toolCall" && message.ToolName == "" {
			message.ToolName = block.ToolName
		}
	}
	message.Text = text.String()
	message.Thinking = thinking.String()
}
func (s *sessionState) startMessage(raw json.RawMessage) {
	message := decodeMessage(raw, "")
	if message.Role == "toolResult" && message.ToolCallID != "" {
		for i := range s.snapshot.Messages {
			if s.snapshot.Messages[i].ToolCallID == message.ToolCallID {
				message.ID = s.snapshot.Messages[i].ID
				s.snapshot.Messages[i] = message
				s.activeMessage = i
				return
			}
		}
	}
	s.activeMessage = s.appendMessage(raw, message.Role == "assistant")
	if message.Role == "user" && s.snapshot.ActiveSession.Summary == "" {
		s.snapshot.ActiveSession.Summary = message.Text
		if s.snapshot.ActiveSession.Name == "New task" {
			s.snapshot.ActiveSession.Name = truncate(message.Text, 80)
		}
	}
	if !message.Timestamp.IsZero() {
		s.snapshot.ActiveSession.Updated = message.Timestamp
	}
}
func (s *sessionState) endMessage(raw json.RawMessage) {
	message := decodeMessage(raw, "")
	index := s.activeMessage
	if index >= 0 && index < len(s.snapshot.Messages) && s.snapshot.Messages[index].Role == message.Role {
		live := s.snapshot.Messages[index]
		message.ID = live.ID
		message.ThinkingDuration, message.GenerationTime = live.ThinkingDuration, live.GenerationTime
		message.RequestStartedAt = live.RequestStartedAt
		if len(message.Usage) == 0 || string(message.Usage) == "null" {
			message.Usage = bytes.Clone(live.Usage)
		}
		s.snapshot.Messages[index] = message
	} else {
		s.appendMessage(raw, false)
	}
	s.activeMessage = -1
	if !message.Timestamp.IsZero() {
		s.snapshot.ActiveSession.Updated = message.Timestamp
	}
}
func (s *sessionState) updateMessage(event pirpc.Event) {
	var delta struct {
		Type                         string
		ContentIndex                 int
		Delta, Content, ID, ToolName string
		ToolCall                     json.RawMessage
	}
	if event.DecodeField("assistantMessageEvent", &delta) != nil {
		return
	}
	index := s.activeMessage
	if index < 0 || index >= len(s.snapshot.Messages) || s.snapshot.Messages[index].Role != "assistant" {
		index = s.appendMessage(json.RawMessage(`{"role":"assistant","content":[]}`), true)
		s.activeMessage = index
	}
	message := &s.snapshot.Messages[index]
	if delta.ContentIndex < 0 || delta.ContentIndex > 10000 {
		s.snapshot.setConnectionError("Pi content block index exceeds limit")
		return
	}
	for len(message.Blocks) <= delta.ContentIndex {
		message.Blocks = append(message.Blocks, ContentBlock{})
	}
	block := &message.Blocks[delta.ContentIndex]
	switch delta.Type {
	case "text_start":
		block.Type = "text"
	case "text_delta":
		block.Type = "text"
		block.Text += delta.Delta
	case "text_end":
		block.Type = "text"
		block.Text = delta.Content
	case "thinking_start":
		block.Type = "thinking"
	case "thinking_delta":
		block.Type = "thinking"
		block.Thinking += delta.Delta
	case "thinking_end":
		block.Type = "thinking"
		block.Thinking = delta.Content
	case "toolcall_start":
		block.Type = "toolCall"
		block.ToolCallID = delta.ID
		block.ToolName = delta.ToolName
	case "toolcall_delta":
		block.Arguments = append(block.Arguments, delta.Delta...)
	case "toolcall_end":
		*block = decodeBlock(delta.ToolCall)
	default:
		s.rememberUnknown(event.Raw)
	}
	if usage, ok := event.Fields["usage"]; ok {
		message.Usage = bytes.Clone(usage)
	}
	rebuildMessage(message)
}
func (s *sessionState) toolEvent(event pirpc.Event) {
	id := event.String("toolCallId")
	index := -1
	for i := range s.snapshot.Messages {
		if s.snapshot.Messages[i].ToolCallID == id && id != "" {
			index = i
			break
		}
	}
	if index < 0 {
		s.snapshot.Messages = append(s.snapshot.Messages, Message{ID: "tool:" + id, Role: "tool", ToolCallID: id, ToolName: event.String("toolName")})
		index = len(s.snapshot.Messages) - 1
	}
	message := &s.snapshot.Messages[index]
	message.Streaming = event.Type != "tool_execution_end"
	if event.Type == "tool_execution_start" {
		message.Blocks = []ContentBlock{{Type: "toolCall", ToolCallID: id, ToolName: message.ToolName, Arguments: bytes.Clone(event.Fields["args"])}}
	}
	var result struct{ Content []json.RawMessage }
	field := "partialResult"
	if event.Type == "tool_execution_end" {
		field = "result"
	}
	if event.DecodeField(field, &result) == nil {
		var text strings.Builder
		for _, raw := range result.Content {
			block := decodeBlock(raw)
			text.WriteString(block.Text)
		}
		message.ToolOutput = text.String()
	}
	if event.Bool("isError") {
		message.Error = message.ToolOutput
	}
	message.Raw = bytes.Clone(event.Raw)
}
func (s *sessionState) extensionUI(event pirpc.Event) {
	method, id := event.String("method"), event.ID
	switch method {
	case "select", "confirm", "input", "editor", "custom":
		if method != "custom" {
			if s.routeSpecialDialog(event) {
				return
			}
			if s.blockingDialog() {
				s.cancelUI(id)
				return
			}
		}
		var value struct {
			Title, Message, Placeholder, Prefill, Phase string
			Options, Lines                              []string
			Timeout                                     float64
			Width, Height                               int
			Cursor                                      json.RawMessage
		}
		_ = event.Decode(&value)
		value.Options = extensionStringOptions(event.Fields["options"])
		if method == "custom" && value.Phase == "close" {
			s.removeDialog(id)
			return
		}
		dialog := PendingDialog{ID: id, Method: method, Title: value.Title, Message: value.Message, Value: value.Prefill, Placeholder: value.Placeholder, Options: value.Options, Timeout: time.Duration(value.Timeout) * time.Millisecond, ReceivedAt: event.ReceivedAt, Phase: value.Phase, Lines: value.Lines, Width: value.Width, Height: value.Height, Cursor: value.Cursor, Raw: bytes.Clone(event.Raw)}
		for i := range s.snapshot.Dialogs {
			if s.snapshot.Dialogs[i].ID == id {
				s.snapshot.Dialogs[i] = dialog
				return
			}
		}
		if method == "custom" && value.Phase != "open" {
			return
		}
		s.snapshot.Dialogs = append(s.snapshot.Dialogs, dialog)
	case "setStatus":
		var statusText *string
		if event.DecodeField("statusText", &statusText) == nil && statusText != nil {
			s.snapshot.ExtensionStatusText = *statusText
			s.snapshot.ExtensionStatusAt = event.ReceivedAt
		}
		key := event.String("statusKey")
		if value := event.String("statusText"); value != "" {
			s.snapshot.StatusEntries[key] = value
		} else {
			delete(s.snapshot.StatusEntries, key)
		}
	case "set_editor_text":
		var text *string
		if event.DecodeField("text", &text) == nil && text != nil {
			s.setComposerText(*text)
		}
	case "notify":
		var message *string
		if event.DecodeField("message", &message) == nil && message != nil {
			s.noticeSequence++
			s.snapshot.ExtensionNotices = append(s.snapshot.ExtensionNotices, ExtensionNotice{event.ID, s.noticeSequence, *message})
			if len(s.snapshot.ExtensionNotices) > 64 {
				s.snapshot.ExtensionNotices = s.snapshot.ExtensionNotices[1:]
			}
		}
	case "setWidget":
		s.applyWidget(event)
	}
}
func (s *sessionState) removeDialog(id string) {
	if p := s.snapshot.Approval; p != nil && p.ID == id {
		s.snapshot.Approval = nil
		return
	}
	if p := s.snapshot.Ask; p != nil && p.ID == id {
		s.snapshot.Ask = nil
		s.askReplay = nil
		return
	}
	for i := range s.snapshot.Dialogs {
		if s.snapshot.Dialogs[i].ID == id {
			s.snapshot.Dialogs = append(s.snapshot.Dialogs[:i], s.snapshot.Dialogs[i+1:]...)
			return
		}
	}
}
func (s *sessionState) authEvent(event pirpc.Event) {
	auth := &s.snapshot.Auth
	incoming := event.String("sessionId")
	if incoming != "" && auth.SessionID != "" && incoming != auth.SessionID && event.Type != "auth_login_started" {
		return
	}
	if incoming != "" {
		auth.SessionID = incoming
	}
	if provider := event.String("provider"); provider != "" {
		auth.Provider = provider
	}
	switch event.Type {
	case "auth_login_started":
		clearAuthChallenge(auth)
		auth.Status = "pending"
		auth.Method = event.String("method")
		auth.Error = ""
	case "auth_login_url":
		auth.URL = event.String("url")
		auth.Status = "waiting"
	case "auth_device_code":
		auth.UserCode = event.String("userCode")
		auth.VerificationURI = event.String("verificationUri")
		auth.Status = "waiting"
	case "auth_login_waiting":
		auth.Status = "waiting"
	case "auth_login_succeeded":
		auth.Status = "succeeded"
		for i := range auth.Providers {
			if auth.Providers[i].ID == auth.Provider {
				auth.Providers[i].Authenticated = true
				auth.Providers[i].Credential = event.String("credential")
			}
		}
	case "auth_login_failed":
		auth.Status = "failed"
		auth.Error = event.String("message")
	case "auth_login_cancelled":
		auth.Status = "cancelled"
	}
	if raw, ok := event.Fields["expiresAt"]; ok {
		_ = json.Unmarshal(raw, &auth.ExpiresAt)
	}
}
func (s *sessionState) applyEntries(data json.RawMessage) {
	var value struct {
		Entries []struct {
			ID, Type, CustomType string
			Data                 json.RawMessage
		}
		LeafID string
	}
	if json.Unmarshal(data, &value) != nil {
		return
	}
	for _, entry := range value.Entries {
		s.entriesCursor = entry.ID
		if entry.Type != "custom" {
			continue
		}
		switch entry.CustomType {
		case "orbit:quota":
			var payload struct{ Providers []QuotaReport }
			if json.Unmarshal(entry.Data, &payload) == nil && payload.Providers != nil {
				s.snapshot.Quotas = payload.Providers
			}
		case "orbit:workflow":
			var payload struct{ Mode string }
			if json.Unmarshal(entry.Data, &payload) == nil && validWorkflow(payload.Mode) {
				s.snapshot.WorkflowMode = payload.Mode
			}
		}
	}
}
func truncate(value string, length int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > length {
		return string(runes[:length]) + "…"
	}
	return string(runes)
}
func validWorkflow(mode string) bool { return mode == "plan" || mode == "build" || mode == "ask" }
