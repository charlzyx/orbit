// Package pirpc speaks the pinned Pi 1.4.0 JSONL subprocess protocol.
// It deliberately does not implement model inference, OAuth, or GUI behavior.
package pirpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

const ProtocolVersion = "1.4.0"

const (
	Prompt                     = "prompt"
	Steer                      = "steer"
	FollowUp                   = "follow_up"
	Abort                      = "abort"
	ClearQueue                 = "clear_queue"
	NewSession                 = "new_session"
	SwitchSession              = "switch_session"
	GetState                   = "get_state"
	GetMessages                = "get_messages"
	SetSessionName             = "set_session_name"
	GetSessionStats            = "get_session_stats"
	GetForkMessages            = "get_fork_messages"
	GetEntries                 = "get_entries"
	Fork                       = "fork"
	Clone                      = "clone"
	GetTree                    = "get_tree"
	GetLastAssistantText       = "get_last_assistant_text"
	GetAvailableModels         = "get_available_models"
	SetModel                   = "set_model"
	CycleModel                 = "cycle_model"
	GetAvailableThinkingLevels = "get_available_thinking_levels"
	SetThinkingLevel           = "set_thinking_level"
	CycleThinkingLevel         = "cycle_thinking_level"
	SetSteeringMode            = "set_steering_mode"
	SetFollowUpMode            = "set_follow_up_mode"
	Compact                    = "compact"
	SetAutoCompaction          = "set_auto_compaction"
	SetAutoRetry               = "set_auto_retry"
	AbortRetry                 = "abort_retry"
	Bash                       = "bash"
	AbortBash                  = "abort_bash"
	ExportHTML                 = "export_html"
	GetCommands                = "get_commands"
	AuthList                   = "auth.list"
	AuthStatus                 = "auth.status"
	AuthLogin                  = "auth.login"
	AuthLogout                 = "auth.logout"
	AuthCancel                 = "auth.cancel"
	QuotaList                  = "quota.list"
)

// Command preserves arbitrary Pi-version-specific fields. Params may not replace
// type/id: correlation is owned explicitly by Command.ID and the transport.
type Command struct {
	Type   string
	ID     string
	Params map[string]any
}

func NewCommand(kind string, params map[string]any) Command {
	return Command{Type: kind, Params: params}
}

func (c Command) MarshalJSON() ([]byte, error) {
	if c.Type == "" {
		return nil, errors.New("RPC command type is empty")
	}
	fields := make(map[string]any, len(c.Params)+2)
	for key, value := range c.Params {
		if key == "type" || key == "id" {
			return nil, fmt.Errorf("reserved command field %q", key)
		}
		fields[key] = value
	}
	fields["type"] = c.Type
	if c.ID != "" {
		fields["id"] = c.ID
	}
	return json.Marshal(fields)
}

// Response is command handling status, not completion of a model run.
type Response struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Command string          `json:"command"`
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
	Raw     json.RawMessage `json:"-"`
}

func (r Response) DecodeData(target any) error {
	if len(r.Data) == 0 {
		return errors.New("RPC response has no data")
	}
	return json.Unmarshal(r.Data, target)
}

// CommandError distinguishes a rejected command from a broken transport.
type CommandError struct{ Response Response }

func (e *CommandError) Error() string { return e.Response.Command + ": " + e.Response.Error }

// Event retains its complete original JSON, including unknown fields and event
// families. Synthetic local events use process_exited or transport_error.
type Event struct {
	Type   string
	ID     string
	Raw    json.RawMessage
	Fields map[string]json.RawMessage
}

func ParseEvent(record []byte) (Event, error) {
	e := Event{Raw: bytes.Clone(record)}
	if err := json.Unmarshal(record, &e.Fields); err != nil {
		return e, err
	}
	if e.Fields == nil {
		return e, errors.New("RPC record must be an object")
	}
	if err := json.Unmarshal(e.Fields["type"], &e.Type); err != nil || e.Type == "" {
		return e, errors.New("RPC record has no string type")
	}
	_ = json.Unmarshal(e.Fields["id"], &e.ID)
	return e, nil
}

func (e Event) Decode(target any) error { return json.Unmarshal(e.Raw, target) }
func (e Event) DecodeField(name string, target any) error {
	raw, ok := e.Fields[name]
	if !ok {
		return fmt.Errorf("event %q has no %q", e.Type, name)
	}
	return json.Unmarshal(raw, target)
}
func (e Event) String(name string) string {
	var value string
	_ = e.DecodeField(name, &value)
	return value
}
func (e Event) Bool(name string) bool { var value bool; _ = e.DecodeField(name, &value); return value }

func localEvent(kind, message string) Event {
	raw, _ := json.Marshal(map[string]any{"type": kind, "error": message})
	e, _ := ParseEvent(raw)
	return e
}

// UIResponse answers a dialog using its request ID. A value may be empty; nil
// means absent. Cancelled is top-level, never nested under value.
type UIResponse struct {
	Type      string  `json:"type"`
	ID        string  `json:"id"`
	Value     *string `json:"value,omitempty"`
	Confirmed *bool   `json:"confirmed,omitempty"`
	Cancelled bool    `json:"cancelled,omitempty"`
}
