package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
)

type liveSession struct {
	mcpStamp      [32]byte
	mcpStampKnown bool
	mcpChecking   uint64

	queueRestoreEpoch    uint64
	queueRestorePending  bool
	modelMu              sync.Mutex
	modelDefault         birthDefault
	modelGeneration      uint64
	checkpointGeneration uint64
	checkpointID         string
	workflowPending      string
	workflowID           string
	workflowVersion      uint64
	workflowGeneration   uint64
	creatingSession      bool
	autoRetry            RetrySetting
	launcher             string
	stopped              bool
	stopping             bool
	runContext           context.Context
	promptCancel         context.CancelFunc
	redactions           []string
	turnOpen             bool
	checkpointDone       <-chan struct{}
	localDiagnostics     []string
	state                sessionState
	client               *pirpc.Client
	loading              bool
	parkedAt             time.Time
	lastEntries          time.Time
	entriesPending       bool
}

func (s *liveSession) busy() bool {
	return s.stopping || s.loading || s.state.snapshot.CheckpointRecovering || s.state.awaiting || s.state.snapshot.Running || s.state.snapshot.Compacting || s.state.manualCompactPending || s.state.snapshot.Retrying
}

// Manager keeps active and background Pi processes alive independently of any
// window. Tick is nonblocking; Open/Send/ReloadSessions perform I/O in workers.
type Manager struct {
	mcpSelection   uint64
	mcpFingerprint func(string, string) [32]byte

	patchPolicyEnabled bool
	patchPolicyVersion uint64
	patchPolicyMu      sync.Mutex
	workflowStore      *preferences.Store
	workflowMu         sync.Mutex
	sessionIndex       SessionIndex
	configuredBinary   atomic.Value
	options            Options
	configurationErr   error
	mu                 sync.Mutex
	sessions           map[string]*liveSession
	diskRows           []SessionRow
	active             string
	sequence           uint64
	lastError          string
	closed             bool
	scanning           bool
	lastScan           time.Time
	ctx                context.Context
	cancel             context.CancelFunc
	workers            sync.WaitGroup
}

func NewManager(options Options) *Manager {
	options.PiPrefixArgs = append([]string{}, options.PiPrefixArgs...)
	options.PiArgs = append([]string{}, options.PiArgs...)
	if options.Environment != nil {
		options.Environment = append([]string{}, options.Environment...)
	}
	overrides := make(map[string]string, len(options.EnvOverrides))
	for key, value := range options.EnvOverrides {
		overrides[key] = value
	}
	options.EnvOverrides = overrides
	configurationErr := resolvePiOptions(&options)
	if options.MaxParked <= 0 {
		options.MaxParked = 6
	}
	if options.ParkedTTL <= 0 {
		options.ParkedTTL = 5 * time.Minute
	}
	if options.CommandTimeout <= 0 {
		options.CommandTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{options: options, configurationErr: configurationErr, sessions: map[string]*liveSession{}, ctx: ctx, cancel: cancel, workflowStore: preferences.New(options.Home)}
	manager.patchPolicyEnabled = manager.workflowStore.RPCPatches().Enabled && !options.DisablePatches && !patchEnvironmentDisabled(options)
	manager.configuredBinary.Store(options.PiBinary)
	if configurationErr == nil {
		manager.ReloadSessions()
	}
	return manager
}
func (m *Manager) changed() {
	if m.options.OnChange != nil {
		m.options.OnChange()
	}
}
func (m *Manager) launch(task func()) bool {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return false
	}
	m.workers.Add(1)
	m.mu.Unlock()
	go func() { defer m.workers.Done(); task() }()
	return true
}

func (m *Manager) Open(workspace, sessionPath string) error {
	return m.open(workspace, sessionPath, "")
}
func (m *Manager) open(workspace, sessionPath, pendingWorkflow string) error {
	if m.configurationErr != nil {
		return m.configurationErr
	}
	if strings.TrimSpace(workspace) == "" {
		return errors.New("workspace is required")
	}
	absolute, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}
	workspace = absolute
	if sessionPath != "" {
		sessionPath, err = filepath.Abs(sessionPath)
		if err != nil {
			return err
		}
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	m.mcpSelection++
	selection := m.mcpSelection
	if sessionPath != "" {
		for key, session := range m.sessions {
			if !session.stopped && session.state.snapshot.ActiveSession.Path == sessionPath && (session.loading || session.client != nil && session.client.Alive()) {
				if old := m.sessions[m.active]; old != nil {
					old.parkedAt = time.Now()
					if key != m.active {
						old.state.cancelExtensionUI()
					}
				}
				wasActive := m.active == key
				m.active = key
				// A validation in progress is not a busy Pi turn. A second
				// navigation to this target owns a fresh completion token.
				wasChecking := session.mcpChecking != 0
				if wasChecking {
					// Release only our validation-owned loading gate. A running
					// event may have arrived while the previous read was pending.
					session.loading, session.mcpChecking = false, 0
				}
				check := (wasChecking || !wasActive || !session.state.snapshot.Connected) && !session.busy()
				if check {
					session.mcpChecking = selection
					session.loading = true
				}
				m.mu.Unlock()
				m.changed()
				if check {
					m.launch(func() { m.validateParkedMCP(key, session, workspace, sessionPath, selection) })
				}
				return nil
			}
		}
	}
	if previous := m.sessions[m.active]; previous != nil {
		previous.parkedAt = time.Now()
		previous.state.cancelExtensionUI()
	}
	key, entry := m.newEntryLocked(workspace, sessionPath, pendingWorkflow)
	m.mu.Unlock()
	m.changed()
	m.launch(func() { m.start(key, entry, workspace, sessionPath) })
	return nil
}
func (m *Manager) start(key string, entry *liveSession, workspace, path string) {
	if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
		m.startFailed(key, "Workspace is unavailable")
		return
	}
	options := m.options
	options.PiBinary = m.PiBinary()
	m.mu.Lock()
	entry.launcher = options.PiBinary
	patchVersion := m.patchPolicyVersion
	m.mu.Unlock()
	if path != "" {
		row, rowErr := ReadSession(path)
		messages, messageErr := ReadSessionMessages(path)
		m.mu.Lock()
		if current := m.sessions[key]; current == entry {
			if rowErr == nil {
				entry.state.snapshot.ActiveSession = row
			}
			if messageErr == nil {
				for _, raw := range messages {
					entry.state.appendMessage(raw, false)
				}
			}
		}
		m.mu.Unlock()
		m.changed()
	}
	extensions, status := prepareExtensions(options)
	overrides := m.piEnvOverrides()
	var redactions []string
	secretNames := map[string]bool{}
	if m.options.ResolveChildEnv != nil {
		resolved, err := m.options.ResolveChildEnv(workspace)
		if err != nil {
			m.startFailed(key, err.Error())
			return
		}
		for _, variable := range resolved {
			if name, value, ok := strings.Cut(variable, "="); ok && name != "" {
				overrides[name] = value
				secretNames[name] = true
				if value != "" {
					redactions = append(redactions, value)
				}
			}
		}
	}
	for name, value := range m.options.EnvOverrides {
		overrides[name] = value
		if secretNames[name] && value != "" {
			redactions = append(redactions, value)
		}
	}
	baseEnvironment := m.options.Environment
	if baseEnvironment == nil {
		baseEnvironment = os.Environ()
	}
	redactions = append(redactions, sensitiveEnvironmentValues(baseEnvironment)...)
	for name, value := range overrides {
		if sensitiveEnvironmentName(name) && value != "" {
			redactions = append(redactions, value)
		}
	}
	// Keep diagnostics safe even if spawning fails before a client is installed.
	m.mu.Lock()
	entry.redactions = redactions
	m.mu.Unlock()
	// Resolved identity paths win over raw/relative env spellings and MCP secrets.
	for key, value := range m.piEnvOverrides() {
		overrides[key] = value
	}
	config := pirpc.Config{Executable: options.PiBinary, PrefixArgs: m.options.PiPrefixArgs, WorkspaceDir: workspace, SessionDir: m.options.SessionDir, Extensions: extensions, Args: m.options.PiArgs, Environment: m.options.Environment, EnvOverrides: overrides, TrustProject: true, Offline: m.options.Offline, NoSession: m.options.NoSession}
	if path == "" {
		config.Args = append(preferences.New(m.options.Home).SessionDefault().Args(), config.Args...)
	}
	if m.options.NoSession {
		config.SessionDir = ""
	}
	m.mu.Lock()
	cancelled := m.closed || m.sessions[key] != entry || entry.stopped
	m.mu.Unlock()
	if cancelled {
		return
	}
	stamp := m.readMCPStamp(workspace)
	client, err := pirpc.Start(config)
	if err != nil {
		m.startFailed(key, err.Error())
		return
	}
	m.mu.Lock()
	current, exists := m.sessions[key]
	if m.closed || !exists || current != entry || entry.stopped {
		m.mu.Unlock()
		client.Close()
		return
	}
	entry.mcpStamp, entry.mcpStampKnown = stamp, true
	entry.client = client
	m.adoptPatchStatusLocked(&status, patchVersion)
	entry.state.snapshot.Extensions = status
	entry.state.snapshot.AccessMode = readAccess(m.options.Home)
	m.mu.Unlock()
	if path != "" {
		response, err := m.callClient(client, pirpc.NewCommand(pirpc.SwitchSession, map[string]any{"sessionPath": path}))
		if err != nil {
			m.startFailed(key, err.Error())
			return
		}
		var result struct{ Cancelled bool }
		_ = response.DecodeData(&result)
		if result.Cancelled {
			m.startFailed(key, "Pi cancelled the session switch")
			return
		}
	}
	for _, kind := range []string{pirpc.GetState, pirpc.GetMessages, pirpc.GetAvailableModels, pirpc.GetAvailableThinkingLevels, pirpc.GetCommands, pirpc.GetSessionStats, pirpc.GetEntries} {
		if _, err = m.callClient(client, pirpc.NewCommand(kind, nil)); err != nil {
			m.startFailed(key, err.Error())
			return
		}
	}
	m.Tick()
	m.mu.Lock()
	if current := m.sessions[key]; current == entry && !entry.stopped {
		entry.loading = false
		entry.state.snapshot.Status = "Ready"
		entry.state.snapshot.Connected = true
	}
	m.mu.Unlock()
	m.changed()
	// Capability probes are independent of ordinary session bootstrap.
	for _, kind := range []string{pirpc.AuthList, pirpc.QuotaList} {
		_, _ = m.callClient(client, pirpc.NewCommand(kind, nil))
	}
	m.Tick()
}
func (m *Manager) startFailed(key, message string) {
	m.mu.Lock()
	if entry := m.sessions[key]; entry != nil {
		if entry.stopped {
			m.mu.Unlock()
			return
		}
		entry.loading = false
		entry.state.snapshot.Status = "Disconnected"
		entry.state.snapshot.ConnectionError = message
		entry.state.snapshot.Connected = false
	}
	m.mu.Unlock()
	m.changed()
}
func (m *Manager) callClient(client *pirpc.Client, command pirpc.Command) (pirpc.Response, error) {
	ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
	defer cancel()
	return client.Call(ctx, command)
}

// NewSession reuses an idle process in the same workspace. Busy sessions are
// parked, never switched away underneath a running agent.
func (m *Manager) NewSession(workspace string) error {
	absolute, err := filepath.Abs(workspace)
	if err != nil {
		return err
	}
	m.mu.Lock()
	key := m.active
	entry := m.sessions[key]
	reusable := entry != nil && !entry.busy() && entry.client != nil && entry.client.Alive() && entry.state.snapshot.ActiveSession.Workspace == absolute
	var previous sessionState
	var departingUI []pirpc.UIResponse
	var previousWorkflowID string
	var previousDefault birthDefault
	var defaultGeneration uint64
	if reusable {
		entry.state.cancelExtensionUI()
		departingUI = entry.state.takeUIResponses()
		previous = entry.state
		previousDefault = entry.modelDefault
		entry.modelGeneration++
		defaultGeneration = entry.modelGeneration
		entry.modelDefault = birthDefault{armed: true}
		previousWorkflowID = entry.workflowID
		extensions, access := entry.state.snapshot.Extensions, entry.state.snapshot.AccessMode
		entry.state = newState(absolute, "")
		entry.state.snapshot.Extensions, entry.state.snapshot.AccessMode = extensions, access
		entry.state.snapshot.Status = "Connecting"
		entry.workflowPending, entry.workflowID = "", ""
		entry.checkpointID = ""
		entry.checkpointGeneration++
		entry.workflowGeneration++
		entry.workflowVersion++
		entry.creatingSession = true
		entry.loading = true
	}
	m.mu.Unlock()
	if !reusable {
		return m.Open(absolute, "")
	}
	m.changed()
	m.launch(func() {
		for _, reply := range departingUI {
			ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
			err := entry.client.SendUI(ctx, reply)
			cancel()
			if err != nil {
				m.operationError(err)
			}
		}
		value := preferences.New(m.options.Home).SessionDefault()
		m.mu.Lock()
		if entry.modelGeneration == defaultGeneration {
			entry.modelDefault.value = value
			entry.modelDefault.loaded = true
		}
		m.mu.Unlock()
		response, err := m.callClient(entry.client, pirpc.NewCommand(pirpc.NewSession, nil))
		if err != nil {
			m.mu.Lock()
			entry.creatingSession = false
			m.mu.Unlock()
			m.startFailed(key, err.Error())
			return
		}
		var result struct{ Cancelled bool }
		_ = response.DecodeData(&result)
		if result.Cancelled {
			m.mu.Lock()
			entry.state = previous
			entry.modelDefault = previousDefault
			entry.modelGeneration++
			entry.workflowID = previousWorkflowID
			entry.workflowPending = ""
			entry.workflowVersion++
			entry.creatingSession = false
			entry.loading = false
			m.mu.Unlock()
			m.changed()
			return
		}
		// Consume any pre-new_session get_state response while the identity
		// gate is still closed; it must never bind a pending choice to the old ID.
		m.Tick()
		m.mu.Lock()
		entry.creatingSession = false
		m.mu.Unlock()
		for _, kind := range []string{pirpc.GetState, pirpc.GetMessages, pirpc.GetSessionStats, pirpc.GetAvailableModels, pirpc.GetAvailableThinkingLevels, pirpc.GetEntries} {
			_, _ = m.callClient(entry.client, pirpc.NewCommand(kind, nil))
		}
		m.mu.Lock()
		entry.loading = false
		m.mu.Unlock()
		m.Tick()
	})
	return nil
}

// Send exposes Pi's actual command vocabulary without introducing a second API.
// It queues work and returns immediately; rejection appears in Snapshot.
func (m *Manager) Send(kind string, params map[string]any) error {
	if kind == pirpc.SetAutoRetry {
		enabled, ok := params["enabled"].(bool)
		if !ok {
			return errors.New("set_auto_retry requires a boolean enabled value")
		}
		return m.SetAutoRetry(enabled)
	}
	m.mu.Lock()
	key := m.active
	entry := m.sessions[key]
	manualModel := kind == pirpc.SetModel || kind == pirpc.SetThinkingLevel || kind == pirpc.CycleModel || kind == pirpc.CycleThinkingLevel
	if entry != nil && manualModel {
		entry.modelDefault.armed = false
		entry.modelGeneration++
	}
	if m.closed || entry == nil || entry.client == nil || entry.loading || !entry.client.Alive() || !entry.state.snapshot.Connected {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	client := entry.client
	m.mu.Unlock()
	authSession, _ := params["sessionId"].(string)
	command := pirpc.NewCommand(kind, params)
	// Freeze caller-owned maps before the asynchronous worker starts.
	raw, err := json.Marshal(command)
	if err != nil {
		return err
	}
	var frozen map[string]json.RawMessage
	if err = json.Unmarshal(raw, &frozen); err != nil {
		return err
	}
	delete(frozen, "type")
	delete(frozen, "id")
	fields := make(map[string]any, len(frozen))
	for key, value := range frozen {
		fields[key] = value
	}
	command.Params = fields
	m.mu.Lock()
	// Marshaling runs outside the lock and may execute caller-owned code.
	// Revalidate admission before marking a turn busy: an asynchronous MCP
	// reopen, stop or selection change may have replaced the captured runtime.
	if m.closed || m.active != key || m.sessions[key] != entry || entry.stopped || entry.loading || entry.client != client || !client.Alive() || !entry.state.snapshot.Connected {
		m.mu.Unlock()
		return errors.New("Pi session changed or is not ready; your draft is kept")
	}
	var before func()
	commandContext := m.ctx
	if kind == pirpc.Abort && entry.promptCancel != nil {
		entry.promptCancel()
		entry.state.awaiting = false
	}
	if kind == pirpc.Prompt {
		if entry.state.snapshot.CheckpointRecovering {
			m.mu.Unlock()
			return errors.New("Recovering turn checkpoints; your draft is kept")
		}
		if entry.runContext == nil || (!entry.state.snapshot.Running && !entry.state.awaiting) {
			if entry.promptCancel != nil {
				entry.promptCancel()
			}
			entry.runContext, entry.promptCancel = context.WithCancel(m.ctx)
		}
		commandContext = entry.runContext
		if !entry.state.snapshot.Running && !entry.state.awaiting {
			entry.state.snapshot.Turn++
			entry.turnOpen = true
			before = m.checkpointLocked(entry, true)
		}
		entry.state.awaiting = true
		entry.state.snapshot.Status = "Waiting for Pi"
	}
	if (kind == pirpc.Steer || kind == pirpc.FollowUp) && entry.runContext != nil {
		commandContext = entry.runContext
	}
	m.mu.Unlock()
	m.launch(func() {
		if manualModel {
			entry.modelMu.Lock()
			defer entry.modelMu.Unlock()
		}
		if before != nil {
			before()
		}
		ctx, cancel := context.WithTimeout(commandContext, m.options.CommandTimeout)
		_, err := client.Call(ctx, command)
		cancel()
		if err != nil {
			if kind == pirpc.AuthLogin {
				m.mu.Lock()
				if current := m.sessions[key]; current == entry {
					failAuthLogin(&current.state.snapshot.Auth, authSession, err)
				}
				m.mu.Unlock()
			}
			if (kind == pirpc.Prompt || kind == pirpc.Steer || kind == pirpc.FollowUp) && errors.Is(err, context.Canceled) {
				m.mu.Lock()
				entry.state.awaiting = false
				if !entry.state.snapshot.Running {
					entry.state.snapshot.Status = "Ready"
				}
				m.mu.Unlock()
				m.changed()
				return
			}
			var rejected *pirpc.CommandError
			if !errors.As(err, &rejected) {
				m.mu.Lock()
				if current := m.sessions[key]; current == entry {
					current.state.snapshot.ConnectionError = err.Error()
					current.state.awaiting = false
				}
				m.mu.Unlock()
			}
		}
		m.Tick()
		m.changed()
	})
	return nil
}
func (m *Manager) Submit(text string, images []Image, mode SendMode) error {
	text = strings.TrimSpace(text)
	if text == "" && len(images) == 0 {
		return errors.New("prompt is empty")
	}
	if mode != SendPrompt && mode != SendSteer && mode != SendFollowUp {
		return errors.New("unknown send mode")
	}
	m.mu.Lock()
	entry := m.sessions[m.active]
	running := entry != nil && (entry.state.snapshot.Running || entry.state.awaiting)
	m.mu.Unlock()
	kind := pirpc.Prompt
	if running {
		if mode == SendSteer {
			kind = pirpc.Steer
		} else {
			kind = pirpc.FollowUp
		}
	}
	params := map[string]any{"message": text}
	if len(images) > 0 {
		images = append([]Image{}, images...)
		for i := range images {
			images[i].Type = "image"
		}
		params["images"] = images
	}
	if err := m.Send(kind, params); err != nil {
		return err
	}
	m.changed()
	return nil
}

func (m *Manager) ClearQueue() error {
	m.mu.Lock()
	if entry := m.sessions[m.active]; entry != nil {
		entry.queueRestorePending = false
		entry.queueRestoreEpoch++
	}
	m.mu.Unlock()
	return m.Send(pirpc.ClearQueue, nil)
}
func (m *Manager) SetModel(provider, id string) error {
	if provider == "" || id == "" {
		return errors.New("provider and model are required")
	}
	return m.Send(pirpc.SetModel, map[string]any{"provider": provider, "modelId": id})
}
func (m *Manager) SetThinkingLevel(level string) error {
	return m.Send(pirpc.SetThinkingLevel, map[string]any{"level": level})
}
func (m *Manager) RenameSession(name string) error {
	return m.Send(pirpc.SetSessionName, map[string]any{"name": name})
}
func (m *Manager) CloneSession() error { return m.Send(pirpc.Clone, nil) }
func (m *Manager) ForkSession(entryID string) error {
	return m.Send(pirpc.Fork, map[string]any{"entryId": entryID})
}

func (m *Manager) RespondDialog(id, value string, confirmed *bool, cancelled bool) error {
	m.mu.Lock()
	entry := m.sessions[m.active]
	if entry == nil || entry.client == nil {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	var dialog *PendingDialog
	for i := range entry.state.snapshot.Dialogs {
		if entry.state.snapshot.Dialogs[i].ID == id {
			copy := entry.state.snapshot.Dialogs[i]
			dialog = &copy
			break
		}
	}
	if p := entry.state.snapshot.Approval; p != nil && p.ID == id {
		dialog = &PendingDialog{ID: id, Method: "select", Options: p.Options}
	}
	if p := entry.state.snapshot.Ask; p != nil && p.ID == id && !p.Submitted && cancelled {
		dialog = &PendingDialog{ID: id, Method: "input"}
	}
	if dialog == nil {
		m.mu.Unlock()
		return errors.New("dialog request is no longer pending")
	}
	// Claim the request before launching I/O: repeated clicks cannot answer twice.
	entry.state.removeDialog(id)
	client := entry.client
	m.mu.Unlock()
	response := pirpc.UIResponse{ID: id, Cancelled: cancelled, Confirmed: confirmed}
	if !cancelled && dialog.Method != "confirm" {
		response.Value = &value
	}
	m.launch(func() {
		ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
		defer cancel()
		err := client.SendUI(ctx, response)
		m.mu.Lock()
		if err != nil {
			entry.state.snapshot.ConnectionError = err.Error()
		}
		m.mu.Unlock()
		m.changed()
	})
	return nil
}
func (m *Manager) CustomUIInput(id, data string) error {
	return m.customUI(map[string]any{"type": "extension_ui_input", "id": id, "data": data})
}
func (m *Manager) CustomUIResize(id string, width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("invalid custom UI size")
	}
	return m.customUI(map[string]any{"type": "extension_ui_resize", "id": id, "width": width, "height": height})
}
func (m *Manager) customUI(record map[string]any) error {
	m.mu.Lock()
	entry := m.sessions[m.active]
	if entry == nil || entry.client == nil {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	client := entry.client
	m.mu.Unlock()
	m.launch(func() {
		ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
		defer cancel()
		if err := client.SendOneWay(ctx, record); err != nil {
			m.mu.Lock()
			entry.state.snapshot.ConnectionError = err.Error()
			m.mu.Unlock()
			m.changed()
		}
	})
	return nil
}

// Tick drains every active/parked session. Its bool selects the 16 ms active or
// 90 ms idle UI heartbeat. It never waits on a subprocess or the filesystem.
func (m *Manager) Tick() bool {
	type operation struct {
		entry             *liveSession
		command           pirpc.Command
		uiResponse        *pirpc.UIResponse
		defaultCommand    bool
		defaultGeneration uint64
	}
	var operations []operation
	var closeClients []*pirpc.Client
	var checkpoints []func()
	var notifications []Notification
	changed, active := false, false
	now := time.Now()
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return false
	}
	for key, entry := range m.sessions {
		active = active || entry.busy()
		if entry.client == nil {
			continue
		}
		for _, event := range entry.client.DrainEvents() {
			if event.Type == "response" && strings.HasPrefix(event.ID, abortRequestPrefix) {
				origin, valid := parseAbortOrigin(event.ID, event.String("command"))
				if !valid || !origin.matches(entry) || event.String("command") == pirpc.Compact && origin.nonce != entry.state.compactionNonce {
					continue
				}
			}
			if entry.creatingSession && event.Type == "response" && event.String("command") == pirpc.GetState {
				continue
			}
			wasRunning := entry.state.snapshot.Running
			parkedRequest := key != m.active && event.Type == "extension_ui_request"
			if parkedRequest {
				entry.state.cancelUI(event.ID)
			} else {
				entry.state.apply(event)
			}
			if event.Type == "response" && event.String("command") == pirpc.GetState && event.Bool("success") {
				id := entry.state.snapshot.ActiveSession.ID
				if id != "" && id != entry.checkpointID {
					entry.checkpointID = id
					entry.turnOpen = false
					if task := m.checkpointIdentityLocked(entry); task != nil {
						checkpoints = append(checkpoints, task)
					}
				}
				if task := m.workflowIdentityLocked(entry); task != nil {
					checkpoints = append(checkpoints, task)
				}
			}
			if event.Type == "response" && event.Bool("success") {
				command := event.String("command")
				if command == pirpc.Fork || command == pirpc.Clone {
					entry.modelDefault.armed = false
					entry.modelGeneration++
				}
				if command == pirpc.GetState || command == pirpc.GetAvailableModels {
					if next := entry.modelDefault.next(entry.state.snapshot); next != nil {
						operations = append(operations, operation{entry: entry, command: *next, defaultCommand: true, defaultGeneration: entry.modelGeneration})
					}
				}
			}
			changed = true
			if event.Type == "response" && event.String("command") == pirpc.GetEntries {
				entry.entriesPending = false
				entry.lastEntries = now
			}

			if event.Type == "agent_settled" {
				if entry.promptCancel != nil {
					entry.promptCancel()
				}
				if entry.turnOpen {
					entry.turnOpen = false
					checkpoints = append(checkpoints, m.checkpointLocked(entry, false))
				}
				entry.lastEntries = time.Time{}
				operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetSessionStats, nil)})
				if wasRunning {
					row := entry.state.snapshot.ActiveSession
					notifications = append(notifications, Notification{SessionID: row.ID, SessionPath: row.Path, Title: row.Name, Message: "Agent finished", Kind: "completed", Active: key == m.active})
				}
			}
			if event.Type == "compaction_end" {
				operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetSessionStats, nil)})
			}
			if event.Type == "extension_ui_request" {
				method := event.String("method")
				if (method == "select" || method == "confirm" || method == "input" || method == "editor" || (method == "custom" && event.String("phase") == "open")) && key == m.active {
					row := entry.state.snapshot.ActiveSession
					notifications = append(notifications, Notification{SessionID: row.ID, SessionPath: row.Path, Title: row.Name, Message: event.String("title"), Kind: "input-needed", Active: true})
				}
			}
			if event.Type == "response" && event.String("command") == pirpc.Prompt && event.Bool("success") {
				var reply struct{ Disposition string }
				_ = event.DecodeField("data", &reply)
				if reply.Disposition == "handled" && entry.turnOpen {
					entry.turnOpen = false
					checkpoints = append(checkpoints, m.checkpointLocked(entry, false))
					entry.state.snapshot.Status = "Ready"
				}
			}
			if event.Type == "response" && event.String("command") == pirpc.Abort && event.Bool("success") {
				entry.state.awaiting = false
				entry.state.snapshot.Running = false
				entry.state.snapshot.Status = "Ready"
				if entry.turnOpen {
					entry.turnOpen = false
					checkpoints = append(checkpoints, m.checkpointLocked(entry, false))
				}
			}
			if event.Type == "auth_credentials_changed" {
				operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.AuthList, nil)}, operation{entry: entry, command: pirpc.NewCommand(pirpc.QuotaList, nil)}, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetAvailableModels, nil)})
			}
			if event.Type == "response" && event.Bool("success") {
				switch event.String("command") {
				case pirpc.SetModel, pirpc.CycleModel, pirpc.SetThinkingLevel, pirpc.CycleThinkingLevel, pirpc.SetSessionName, pirpc.SetSteeringMode, pirpc.SetFollowUpMode, pirpc.SetAutoCompaction, pirpc.Fork, pirpc.Clone:
					operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetState, nil)}, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetAvailableThinkingLevels, nil)})
					if event.String("command") == pirpc.Fork || event.String("command") == pirpc.Clone {
						operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetMessages, nil)})
					}
				}
			}
		}
		entry.state.snapshot.Diagnostics = append(entry.client.RecentStderr(), entry.localDiagnostics...)
		if entry.stopped {
			entry.state.snapshot.Connected = false
			entry.state.snapshot.Running = false
			entry.state.snapshot.Status = "Stopped"
		}
		auth := &entry.state.snapshot.Auth
		if (auth.Status == "pending" || auth.Status == "waiting") && auth.ExpiresAt > 0 && now.UnixMilli() >= auth.ExpiresAt {
			operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.AuthCancel, map[string]any{"sessionId": auth.SessionID})})
			auth.Status = "timeout"
			auth.Error = "Provider sign-in timed out"
			changed = true
		}
		for _, response := range entry.state.takeUIResponses() {
			response := response
			operations = append(operations, operation{entry: entry, uiResponse: &response})
		}

		if !entry.loading && entry.client.Alive() && !entry.entriesPending && now.Sub(entry.lastEntries) >= time.Minute {
			params := map[string]any{}
			if entry.state.entriesCursor != "" {
				params["since"] = entry.state.entriesCursor
			}
			entry.entriesPending = true
			operations = append(operations, operation{entry: entry, command: pirpc.NewCommand(pirpc.GetEntries, params)})
		}
		busy := entry.busy()
		active = active || busy
		if key != m.active && !entry.loading && !busy && (!entry.client.Alive() || !entry.parkedAt.IsZero() && now.Sub(entry.parkedAt) >= m.options.ParkedTTL) {
			closeClients = append(closeClients, entry.client)
			delete(m.sessions, key)
			changed = true
		}
	}
	// Evict only idle least-recently parked sessions. Running sessions may exceed
	// the warm pool cap and must never be killed merely to satisfy it.
	for len(m.sessions) > m.options.MaxParked+1 {
		oldestKey := ""
		var oldest time.Time
		for key, entry := range m.sessions {
			if key == m.active || entry.busy() {
				continue
			}
			if oldestKey == "" || entry.parkedAt.Before(oldest) {
				oldestKey = key
				oldest = entry.parkedAt
			}
		}
		if oldestKey == "" {
			break
		}
		if client := m.sessions[oldestKey].client; client != nil {
			closeClients = append(closeClients, client)
		}
		delete(m.sessions, oldestKey)
		changed = true
	}
	scan := !m.scanning && now.Sub(m.lastScan) >= 2*time.Second
	m.mu.Unlock()
	for _, checkpoint := range checkpoints {
		if checkpoint != nil {
			m.launch(checkpoint)
		}
	}
	for _, client := range closeClients {
		client := client
		m.launch(func() { client.Close() })
	}
	for _, operation := range operations {
		operation := operation
		m.launch(func() {
			if operation.uiResponse != nil {
				ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
				defer cancel()
				if err := operation.entry.client.SendUI(ctx, *operation.uiResponse); err != nil {
					m.mu.Lock()
					operation.entry.state.snapshot.ConnectionError = err.Error()
					m.mu.Unlock()
					m.changed()
				}

			} else {
				if operation.defaultCommand {
					operation.entry.modelMu.Lock()
					defer operation.entry.modelMu.Unlock()
					m.mu.Lock()
					current := !m.closed && operation.entry.modelGeneration == operation.defaultGeneration && !operation.entry.stopped
					m.mu.Unlock()
					if !current {
						return
					}
				}
				_, err := m.callClient(operation.entry.client, operation.command)
				if err != nil && operation.command.Type == pirpc.GetEntries {
					m.mu.Lock()
					operation.entry.entriesPending = false
					operation.entry.lastEntries = time.Now()
					if operation.entry.state.entriesCursor != "" {
						operation.entry.state.entriesCursor = ""
						operation.entry.lastEntries = time.Time{}
					}
					m.mu.Unlock()
				}
			}
		})
	}
	for _, notification := range notifications {
		if m.options.OnNotification != nil {
			m.options.OnNotification(notification)
		}
	}
	if scan {
		m.ReloadSessions()
	}
	if changed {
		m.changed()
	}
	return active
}

func (m *Manager) ReloadSessions() {
	m.mu.Lock()
	if m.closed || m.scanning {
		m.mu.Unlock()
		return
	}
	m.scanning = true
	m.lastScan = time.Now()
	m.mu.Unlock()
	if !m.launch(func() {
		rows, err := m.sessionIndex.Scan(m.options.SessionDir)
		m.mu.Lock()
		m.scanning = false
		if err == nil {
			m.diskRows = rows
		}
		m.mu.Unlock()
		m.changed()
	}) {
		m.mu.Lock()
		m.scanning = false
		m.mu.Unlock()
	}
}
func (m *Manager) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot := Snapshot{Status: "Disconnected", StatusEntries: map[string]string{}}
	snapshot.Extensions.Patches.Enabled = m.patchPolicyEnabled
	if entry := m.sessions[m.active]; entry != nil {
		snapshot = entry.state.snapshot
		snapshot.RuntimeID = m.active
		snapshot.Awaiting = entry.state.awaiting
		snapshot.Compacting = snapshot.Compacting || entry.state.manualCompactPending
		snapshot.Running = snapshot.Running || entry.state.awaiting
		snapshot.ActiveSession.Running = snapshot.Running
		if entry.loading {
			snapshot.Connected = false
			snapshot.Status = "Connecting"
		}
		if snapshot.CheckpointRecovering {
			snapshot.Connected = false
			snapshot.Status = "Recovering turn checkpoints"
			if snapshot.CheckpointRecoveryError != "" {
				snapshot.Status = "Review checkpoint recovery failed"
				snapshot.ConnectionError = snapshot.CheckpointRecoveryError
			}
		}
	}
	rows := append([]SessionRow{}, m.diskRows...)
	for _, entry := range m.sessions {
		row := entry.state.snapshot.ActiveSession
		found := false
		for i := range rows {
			if row.Path != "" && rows[i].Path == row.Path {
				if row.Name != "" {
					rows[i].Name = row.Name
				}
				rows[i].Running = row.Running
				if row.Updated.After(rows[i].Updated) {
					rows[i].Updated = row.Updated
				}
				found = true
				break
			}
		}
		if !found && row.Summary != "" && (row.ID != "" || row.Path != "") {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Updated.After(rows[j].Updated) })
	snapshot.Sessions = rows
	if snapshot.ConnectionError == "" {
		snapshot.ConnectionError = m.lastError
	}
	var redactions []string
	if entry := m.sessions[m.active]; entry != nil {
		redactions = entry.redactions
	}
	snapshot.ConnectionError = redactDiagnostic(snapshot.ConnectionError, redactions)
	snapshot.Auth.Error = redactDiagnostic(snapshot.Auth.Error, redactions)
	snapshot.Diagnostics = append([]string{}, snapshot.Diagnostics...)
	for i, value := range snapshot.Diagnostics {
		snapshot.Diagnostics[i] = redactDiagnostic(value, redactions)
	}
	return cloneValue(reflect.ValueOf(snapshot)).Interface().(Snapshot)
}

// Deep-copy slices/maps/raw JSON without encoding partially streamed tool args.
func cloneValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for i := 0; i < value.Len(); i++ {
			copy.Index(i).Set(cloneValue(value.Index(i)))
		}
		return copy
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.MakeMapWithSize(value.Type(), value.Len())
		iter := value.MapRange()
		for iter.Next() {
			copy.SetMapIndex(iter.Key(), cloneValue(iter.Value()))
		}
		return copy
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.New(value.Type().Elem())
		copy.Elem().Set(cloneValue(value.Elem()))
		return copy
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return value
		}
		copy := reflect.New(value.Type()).Elem()
		for i := 0; i < value.NumField(); i++ {
			copy.Field(i).Set(cloneValue(value.Field(i)))
		}
		return copy
	default:
		return value
	}
}

// ErrReloadDeferred means the target became busy after the host's idle check.
var ErrReloadDeferred = errors.New("Pi configuration reload deferred until idle")
var ErrReloadSuperseded = errors.New("Pi configuration reload target changed")

func (m *Manager) Restart() error { return m.restart("", "", "", false) }

// RestartIdle validates the exact active runtime and its busy state under the
// same lock that detaches it. A parked busy runtime is never stopped by this call.
func (m *Manager) RestartIdle(expectedRuntime, expectedPath, expectedSession string) error {
	return m.restart(expectedRuntime, expectedPath, expectedSession, true)
}
func (m *Manager) restart(expectedRuntime, expectedPath, expectedSession string, idleOnly bool) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	if idleOnly && m.active != expectedRuntime {
		m.mu.Unlock()
		return ErrReloadSuperseded
	}
	if idleOnly {
		if e := m.sessions[m.active]; e != nil && e.busy() {
			m.mu.Unlock()
			return ErrReloadDeferred
		}
	}
	entry := m.sessions[m.active]
	if entry == nil {
		m.mu.Unlock()
		return errors.New("no active session")
	}
	row := entry.state.snapshot.ActiveSession
	if idleOnly && (row.Path != expectedPath || row.ID != expectedSession) {
		m.mu.Unlock()
		return ErrReloadSuperseded
	}
	pendingWorkflow := entry.workflowPending
	delete(m.sessions, m.active)
	m.active = ""
	m.mu.Unlock()
	if entry.client != nil {
		m.launch(func() { entry.client.Close() })
	}
	return m.open(row.Workspace, row.Path, pendingWorkflow)
}
func (m *Manager) Close() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	m.cancel()
	clients := []*pirpc.Client{}
	for _, entry := range m.sessions {
		if entry.client != nil {
			clients = append(clients, entry.client)
		}
	}
	m.mu.Unlock()
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(client *pirpc.Client) { defer wg.Done(); client.Close() }(client)
	}
	wg.Wait()
	m.workers.Wait()
	return nil
}
func readAccess(home string) string {
	var value struct{ Mode string }
	if raw, err := os.ReadFile(filepath.Join(home, ".orbit-pi", "access.json")); err == nil && json.Unmarshal(raw, &value) == nil {
		switch value.Mode {
		case "supervised", "auto-accept-edits", "full-access":
			return value.Mode
		}
	}
	return "full-access"
}
func (m *Manager) SetAccessMode(mode string) error {
	if mode != "supervised" && mode != "auto-accept-edits" && mode != "full-access" {
		return errors.New("unknown access mode")
	}
	raw, _ := json.Marshal(map[string]string{"mode": mode})
	if err := atomicWrite(filepath.Join(m.options.Home, ".orbit-pi", "access.json"), raw); err != nil {
		return err
	}
	m.mu.Lock()
	for _, entry := range m.sessions {
		entry.state.snapshot.AccessMode = mode
	}
	m.mu.Unlock()
	m.changed()
	return nil
}
func (m *Manager) SetWorkflowMode(mode string) error {
	return m.setWorkflowMode(mode)
}

// StartLogin delegates the advertised method to Pi; no credential value crosses
// this method. The user-facing flow is asynchronous and deadline bounded.
func (m *Manager) StartLogin(provider, method string) error {
	m.mu.Lock()
	entry := m.sessions[m.active]
	if entry == nil || entry.state.snapshot.Auth.Support != "supported" {
		m.mu.Unlock()
		return errors.New("Pi provider authentication is unavailable")
	}
	allowed := false
	for _, p := range entry.state.snapshot.Auth.Providers {
		if p.ID == provider {
			for _, candidate := range p.Methods {
				if candidate.ID == method {
					allowed = true
				}
			}
		}
	}
	if !allowed {
		m.mu.Unlock()
		return errors.New("provider did not advertise this login method")
	}
	m.sequence++
	sessionID := fmt.Sprintf("orbit-go-auth-%d-%d", time.Now().UnixNano(), m.sequence)
	entry.state.snapshot.Auth.SessionID = sessionID
	entry.state.snapshot.Auth.Provider = provider
	entry.state.snapshot.Auth.Method = method
	entry.state.snapshot.Auth.Status = "pending"
	entry.state.snapshot.Auth.Error = ""
	clearAuthChallenge(&entry.state.snapshot.Auth)
	entry.state.snapshot.Auth.ExpiresAt = time.Now().Add(5 * time.Minute).UnixMilli()
	m.mu.Unlock()
	err := m.Send(pirpc.AuthLogin, map[string]any{"provider": provider, "method": method, "sessionId": sessionID})
	if err != nil {
		m.mu.Lock()
		failAuthLogin(&entry.state.snapshot.Auth, sessionID, err)
		m.mu.Unlock()
		m.changed()
	}
	return err
}
func (m *Manager) CancelLogin() error {
	m.mu.Lock()
	entry := m.sessions[m.active]
	if entry == nil || entry.state.snapshot.Auth.SessionID == "" {
		m.mu.Unlock()
		return errors.New("no pending sign-in")
	}
	id := entry.state.snapshot.Auth.SessionID
	m.mu.Unlock()
	return m.Send(pirpc.AuthCancel, map[string]any{"sessionId": id})
}
func (m *Manager) Logout(provider string) error {
	if provider == "" {
		return errors.New("provider is required")
	}
	return m.Send(pirpc.AuthLogout, map[string]any{"provider": provider})
}
func (m *Manager) RefreshQuota() error { return m.Send(pirpc.QuotaList, nil) }

func (m *Manager) PiBinary() string {
	if value := m.configuredBinary.Load(); value != nil {
		return value.(string)
	}
	return ""
}

// LiveSessionPaths returns a sorted, detached set of all session files currently
// owned by the manager, including idle parked and still-loading sessions. Keep
// these files out of retention even if an open session is disconnected.
func (m *Manager) LiveSessionPaths() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	seen := make(map[string]bool, len(m.sessions))
	paths := make([]string, 0, len(m.sessions))
	for _, entry := range m.sessions {
		path := entry.state.snapshot.ActiveSession.Path
		if path != "" && !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

// AnyBusy includes background sessions and startup, pending prompts, compaction,
// and retry work. The result is a point-in-time observation, not an update lock.
func (m *Manager) AnyBusy() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return false
	}
	for _, entry := range m.sessions {
		if entry.busy() {
			return true
		}
	}
	return false
}

// SetPiBinary changes subsequent spawns and restarts only the active session.
// Busy parked sessions continue on their existing runtime without interruption.
func (m *Manager) SetPiBinary(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("Pi executable path is required")
	}
	resolved, err := exec.LookPath(path)
	if err != nil {
		return fmt.Errorf("Pi executable is unavailable: %w", err)
	}
	path = resolved
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return pirpc.ErrClosed
	}
	hasActive := m.sessions[m.active] != nil
	m.mu.Unlock()
	m.configuredBinary.Store(path)
	if hasActive {
		return m.Restart()
	}
	m.changed()
	return nil
}

// Abort implements Orbit's interactive Escape: clear queued messages before
// aborting so they cannot immediately restart the run, and restore their text.
func (m *Manager) Abort() error {
	m.mu.Lock()
	key := m.active
	entry := m.sessions[key]
	if m.closed || entry == nil || entry.client == nil || entry.loading || !entry.client.Alive() {
		m.mu.Unlock()
		return errors.New("Pi session is not ready")
	}
	if entry.promptCancel != nil {
		entry.promptCancel()
	}
	entry.state.awaiting = false
	client := entry.client
	hasQueue := len(entry.state.snapshot.Queue.Steering)+len(entry.state.snapshot.Queue.FollowUp) > 0
	m.sequence++
	origin := abortOrigin{sessionID: entry.state.snapshot.ActiveSession.ID, reset: entry.workflowGeneration, nonce: m.sequence}
	if hasQueue && !entry.queueRestorePending {
		entry.queueRestoreEpoch++
		entry.queueRestorePending = true
	}
	restoreEpoch := entry.queueRestoreEpoch
	m.mu.Unlock()
	if !m.launch(func() {
		commands := []pirpc.Command{}
		if hasQueue {
			commands = append(commands, origin.command(pirpc.ClearQueue))
		}
		commands = append(commands, origin.command(pirpc.Abort))
		ctx, cancel := context.WithTimeout(m.ctx, m.options.CommandTimeout)
		defer cancel()
		// Source enqueues clear_queue then abort immediately. Restoration waits
		// for the clear response; abort transmission must never wait for it.
		client.CallOrdered(ctx, commands, func(i int, response pirpc.Response, err error) {
			if commands[i].Type == pirpc.ClearQueue {
				if err == nil {
					var dropped Queue
					if response.DecodeData(&dropped) == nil {
						parts := append(append([]string{}, dropped.Steering...), dropped.FollowUp...)
						m.mu.Lock()
						// NewSession advances workflowGeneration when replacing the
						// state. A late clear reply belongs to that old state, not
						// a new task reusing its process. Parking alone remains valid.
						if !m.closed && m.sessions[key] == entry && entry.client == client && origin.matches(entry) {
							entry.state.snapshot.Queue = Queue{}
							if entry.queueRestorePending && entry.queueRestoreEpoch == restoreEpoch {
								entry.queueRestorePending = false
								if len(parts) > 0 {
									entry.state.setComposerText(strings.Join(parts, "\n\n"))
								}
							}
						}
						m.mu.Unlock()
						m.changed()
					}
				}
			} else if err != nil {
				m.mu.Lock()
				if !m.closed && m.sessions[key] == entry && entry.client == client && origin.matches(entry) {
					entry.state.snapshot.ConnectionError = err.Error()
				}
				m.mu.Unlock()
			}
		})
		m.Tick()
		m.changed()
	}) {
		return pirpc.ErrClosed
	}
	m.changed()
	return nil
}
