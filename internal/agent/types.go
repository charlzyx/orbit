// Package agent owns Pi-backed session state and runtime coordination. UI code
// consumes immutable snapshots; model, auth, quota, and run facts come from Pi.
package agent

import (
	"encoding/json"
	"io/fs"
	"time"
)

// Options are application runtime configuration, not per-prompt credentials.
type Options struct {
	PiBinary           string
	PiPrefixArgs       []string
	PiArgs             []string
	Home               string
	SessionDir         string
	Environment        []string
	EnvOverrides       map[string]string
	ResolveChildEnv    func(workspace string) ([]string, error)
	Resources          fs.FS
	DisableExtensions  bool
	DisablePatches     bool
	DisableCheckpoints bool
	Offline            bool
	NoSession          bool
	MaxParked          int
	ParkedTTL          time.Duration
	CommandTimeout     time.Duration
	OnChange           func()
	OnNotification     func(Notification)
}

type Notification struct {
	SessionID, SessionPath, Title, Message, Kind string
	Active                                       bool
}
type SessionRow struct {
	ID, Path, Name, Workspace, Summary string
	Updated                            time.Time
	Running                            bool
}
type Image struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MIMEType string `json:"mimeType"`
}
type ContentBlock struct {
	Type, Text, Thinking, ToolCallID, ToolName string
	Arguments                                  json.RawMessage
	Image                                      *Image
	Raw                                        json.RawMessage
}
type Message struct {
	RequestStartedAt time.Time `json:"-"`
	// Client-measured live clocks. Nil for reloaded/whole-block-only messages.
	ThinkingDuration, GenerationTime                      *time.Duration `json:"-"`
	ID, Role, Text, Thinking, ToolName, ToolOutput, Error string
	Streaming                                             bool
	Blocks                                                []ContentBlock
	ToolCallID                                            string
	Timestamp                                             time.Time
	Usage                                                 json.RawMessage
	Raw                                                   json.RawMessage
}
type ModelChoice struct {
	ID            string          `json:"id"`
	Provider      string          `json:"provider"`
	Name          string          `json:"name"`
	Reasoning     bool            `json:"reasoning"`
	ContextWindow int             `json:"contextWindow"`
	MaxTokens     int             `json:"maxTokens"`
	Raw           json.RawMessage `json:"-"`
}
type PendingDialog struct {
	ID, Method, Title, Message, Value string
	Options                           []string
	Timeout                           time.Duration
	ReceivedAt                        time.Time
	Phase                             string
	Lines                             []string
	Width, Height                     int
	Cursor                            json.RawMessage
	Raw                               json.RawMessage
}
type Queue struct {
	Steering []string `json:"steering"`
	FollowUp []string `json:"followUp"`
}
type AuthMethod struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type AuthProvider struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Credential    string       `json:"credential"`
	Authenticated bool         `json:"authenticated"`
	Methods       []AuthMethod `json:"methods"`
}
type AuthState struct {
	Support                                                                    string
	Providers                                                                  []AuthProvider
	SessionID, Provider, Method, Status, URL, UserCode, VerificationURI, Error string
	ExpiresAt                                                                  int64
}
type QuotaWindow struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	UsedPercent *float64 `json:"usedPercent"`
	Used        *float64 `json:"used"`
	Limit       *float64 `json:"limit"`
	Unit        string   `json:"unit"`
	ResetsAt    *int64   `json:"resetsAt"`
}
type QuotaBalance struct {
	Label    string  `json:"label"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}
type QuotaReport struct {
	Provider  string         `json:"provider"`
	Kind      string         `json:"kind"`
	Plan      string         `json:"plan"`
	Windows   []QuotaWindow  `json:"windows"`
	Balances  []QuotaBalance `json:"balances"`
	Note      string         `json:"note"`
	FetchedAt *int64         `json:"fetchedAt"`
	Error     string         `json:"error"`
}
type PatchStatus struct {
	Enabled     bool
	PackageRoot string
	Applied     []string
	Error       string
}
type ExtensionStatus struct {
	Guard, Workflow, Title, Quota bool
	Errors                        []string
	Patches                       PatchStatus
}
type Snapshot struct {
	Awaiting bool
	Turn     int
	// LatestTurn is the newest successfully captured completed turn. Turn may
	// already identify a running prompt; Review must never substitute it here.
	LatestTurn                               int
	CheckpointRecovering                     bool
	CheckpointRecoveryError                  string
	Sessions                                 []SessionRow
	Messages                                 []Message
	Models                                   []ModelChoice
	ThinkingLevels                           []string
	ActiveSession                            SessionRow
	Model                                    ModelChoice
	ThinkingLevel, ConnectionError, Status   string
	Connected, Running, Compacting, Retrying bool
	Dialogs                                  []PendingDialog
	Queue                                    Queue
	Auth                                     AuthState
	Quotas                                   []QuotaReport
	QuotaSupport                             string
	Stats                                    json.RawMessage
	Commands                                 []json.RawMessage
	Capabilities                             []string
	StatusEntries                            map[string]string
	ComposerText                             string
	SteeringMode, FollowUpMode               string
	AutoCompactionEnabled                    bool
	AutoCompactionKnown                      bool
	AccessMode, WorkflowMode                 string
	Extensions                               ExtensionStatus
	Diagnostics                              []string
	UnknownEvents                            []json.RawMessage
}

type SendMode string

const (
	SendPrompt   SendMode = "prompt"
	SendSteer    SendMode = "steer"
	SendFollowUp SendMode = "follow_up"
)
