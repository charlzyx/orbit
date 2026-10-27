// Package mcpui implements Orbit's MCP settings entirely with native MyGo.
// Disk access and Pi subprocesses run off the UI thread; Tick is the only
// place background results are applied to the retained view state.
package mcpui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

// Theme supplies Orbit-specific geometry without creating an internal/ui cycle.
// All other colors and font settings come from the current MyGo theme.
type Theme struct {
	Elevated, Secondary, Action, ActionText ui.Color
	CodeFont                                string
	Px                                      func(float32) float32
	Space                                   func(int) float32
}
type Options struct {
	// OnChange is a goroutine-safe request to invalidate the host window.
	OnChange func()
	// OnModified runs on the UI thread after a real config/secret change. Hosts
	// should schedule their debounced, busy-gated, session-preserving Pi reload.
	OnModified func()
	// OnApply requests the same safe reload explicitly, never kills a live turn.
	OnApply  func()
	Text     func(key, fallback string) string
	Theme    func() Theme
	PiBinary string
}

type Tool struct {
	Name, Description, Exposure string
	Schema                      json.RawMessage
}
type Runtime struct {
	State, Error                 string
	Tools                        []Tool
	Resources, ResourceTemplates int
	CheckedAt                    time.Time
}
type probe struct {
	Servers map[string]Runtime
	Errors  []string
	Note    string
}
type snapshot struct {
	servers     []workbench.MCPServer
	errors      []string
	secretNames []string
	codemode    *bool
	fingerprint string
}
type completion struct {
	generation uint64
	apply      func(*State)
	err        error
}
type secretPrompt struct {
	server, scope, name, value, error string
	open                              bool
}
type confirmation struct {
	title, detail string
	open          bool
	accept        func()
}

type State struct {
	localBusyError                                             bool
	Options                                                    Options
	Workspace                                                  string
	Servers                                                    []workbench.MCPServer
	Runtime                                                    map[string]Runtime
	Errors                                                     []string
	Error, Status, Query, Selected                             string
	Busy, Probing                                              bool
	Editor                                                     *Draft
	service                                                    *workbench.Service
	results                                                    chan completion
	cancel                                                     context.CancelFunc
	generation                                                 uint64
	applyRevision                                              uint64
	initialProbeWaiting                                        bool
	initialProbeAt                                             time.Time
	quotaProbeRequested                                        bool
	hostVisibilityKnown, hostVisible                           bool
	closed, loaded, pendingProbe, externalChange, applyPending bool
	activity, authTarget                                       string
	fingerprint                                                string
	codemode                                                   *bool
	secretNames                                                []string
	probeErrors                                                []string
	probeNote                                                  string
	checkedAt, nextPoll, lastViewed                            time.Time
	editorOpen                                                 bool
	removeTarget                                               string
	secret                                                     secretPrompt
	confirm                                                    confirmation
}

func New(p *preferences.Store, workspace string, o Options) *State {
	service := workbench.New(p, workspace)
	if o.PiBinary != "" {
		service.PiBinary = o.PiBinary
	}
	return &State{Options: o, Workspace: workspace, service: service, Runtime: map[string]Runtime{}, results: make(chan completion, 16)}
}
func (s *State) label(key, fallback string, values ...string) string {
	text := fallback
	if s.Options.Text != nil {
		text = s.Options.Text(key, fallback)
	}
	for i := 0; i+1 < len(values); i += 2 {
		text = strings.ReplaceAll(text, "%{"+values[i]+"}", values[i+1])
	}
	return text
}
func (s *State) SetWorkspace(workspace string) {
	if workspace == s.Workspace || s.closed {
		return
	}
	s.cancelWork()
	s.applyRevision++
	s.Workspace = workspace
	s.service = workbench.New(s.service.Preferences, workspace)
	if s.Options.PiBinary != "" {
		s.service.PiBinary = s.Options.PiBinary
	}
	s.Servers = nil
	s.Runtime = map[string]Runtime{}
	s.Errors = nil
	s.Error = ""
	s.localBusyError = false
	s.Status = ""
	s.Query = ""
	s.Selected = ""
	s.Editor = nil
	s.secret = secretPrompt{}
	s.confirm = confirmation{}
	s.removeTarget = ""
	s.loaded = false
	s.pendingProbe = false
	s.quotaProbeRequested = false
	s.initialProbeAt = time.Time{}
	s.initialProbeWaiting = false
	s.externalChange = false
	s.applyPending = false
	s.checkedAt = time.Time{}
	s.nextPoll = time.Time{}
}
func (s *State) cancelWork() {
	s.generation++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.Busy = false
	s.Probing = false
	s.authTarget = ""
}
func (s *State) Close() {
	s.initialProbeAt = time.Time{}
	s.initialProbeWaiting = false
	s.cancelWork()
	s.closed = true
	s.Editor = nil
	s.secret = secretPrompt{}
	s.confirm = confirmation{}
}

// Escape dismisses owned transient UI before the host changes settings pages.
func (s *State) Escape() bool {
	if s.confirm.open {
		s.confirm = confirmation{}
		return true
	}
	if s.secret.open {
		s.secret = secretPrompt{}
		return true
	}
	if s.Editor != nil {
		s.Editor = nil
		return true
	}
	if s.removeTarget != "" {
		s.removeTarget = ""
		return true
	}
	return false
}
func (s *State) HasModal() bool { return s.Editor != nil || s.secret.open || s.confirm.open }
func (s *State) start(activity string, job func(context.Context) (func(*State), error)) {
	if s.Busy || s.closed {
		return
	}
	s.Busy = true
	s.activity = activity
	s.Error = ""
	s.localBusyError = false
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	channel, wake := s.results, s.Options.OnChange
	go func() {
		defer cancel()
		apply, err := job(ctx)
		// Canceled work must not strand the UI in Busy; generation guards reject
		// old work after navigation, and a live cancellation gets a visible result.
		result := completion{generation, apply, err}
		select {
		case channel <- result:
			if wake != nil {
				wake()
			}
		case <-ctx.Done():
			select {
			case channel <- result:
				if wake != nil {
					wake()
				}
			default:
			}
		}
	}()
}

// SetHostVisible supplies retained navigation intent, independent of whether a
// native frame was recently rendered. The application owner sets this before
// Tick; standalone embedded views retain their recent-View fallback.
func (s *State) SetHostVisible(visible bool) { s.hostVisibilityKnown, s.hostVisible = true, visible }
func (s *State) pollVisible() bool {
	if s.hostVisibilityKnown {
		return s.hostVisible
	}
	return time.Since(s.lastViewed) < 2*time.Second
}

// Tick drains worker results on the main thread. It never reads files or runs Pi.
func (s *State) Tick() bool { return s.tickAt(time.Now()) }

func (s *State) tickAt(now time.Time) bool {
	changed := false
	for {
		select {
		case r := <-s.results:
			if s.closed || r.generation != s.generation {
				continue
			}
			s.Busy = false
			s.Probing = false
			s.cancel = nil
			s.activity = ""
			s.authTarget = ""
			if r.err != nil {
				s.Error = r.err.Error()
				s.localBusyError = false
			}
			if r.apply != nil {
				r.apply(s)
			}
			changed = true
		default:
			if !s.closed && !s.Busy {
				s.tickInitialProbe(now)
				if s.quotaProbeRequested {
					s.pendingProbe = true
					s.quotaProbeRequested = false
				}
				if s.pendingProbe {
					s.pendingProbe = false
					s.Probe()
					changed = true
				} else if s.loaded && !s.nextPoll.IsZero() && !time.Now().Before(s.nextPoll) && s.pollVisible() {
					s.refresh(false, true)
					changed = true
				}
			}
			return changed
		}
	}
}
func loadSnapshot(service *workbench.Service) snapshot {
	servers, errs := service.MCP()
	out := snapshot{servers: servers}
	for _, e := range errs {
		out.errors = append(out.errors, service.Redact(e.Error()))
	}
	env, e := service.SecretEnv()
	if e != nil {
		out.errors = append(out.errors, e.Error())
	} else {
		for _, entry := range env {
			name, _, _ := strings.Cut(entry, "=")
			out.secretNames = append(out.secretNames, name)
		}
		sort.Strings(out.secretNames)
	}
	out.codemode, e = service.MCPAutoEnableCodemode()
	if e != nil {
		out.errors = append(out.errors, service.Redact(e.Error()))
	}
	out.fingerprint = workbench.MCPFingerprint(servers)
	if out.codemode != nil {
		out.fingerprint += fmt.Sprint(*out.codemode)
	}
	return out
}
func (s *State) adopt(v snapshot) {
	old := s.Runtime
	s.Servers, s.Errors, s.secretNames, s.codemode = v.servers, v.errors, v.secretNames, v.codemode
	s.fingerprint = v.fingerprint
	s.loaded = true
	s.nextPoll = time.Now().Add(time.Minute)
	s.Runtime = map[string]Runtime{}
	for _, server := range s.Servers {
		runtime := old[server.Name]
		if !enabled(server.Definition) {
			runtime = Runtime{State: "disabled"}
		}
		s.Runtime[server.Name] = runtime
	}
	if s.Selected != "" && s.server(s.Selected) == nil {
		s.Selected = ""
	}
}
func (s *State) Refresh() { s.refresh(true, false) }

// RefreshAndApply captures the latest desired fingerprint before the host
// schedules a reload. Scheduling immediately after an async Refresh would bind
// the request to stale page data and could discard the user's reconnect action.
func (s *State) RefreshAndApply() {
	s.interruptProbe()
	if s.closed {
		return
	}
	if s.Busy {
		s.Error = "Configuration operation in progress. Retry reconnect when it finishes."
		s.localBusyError = true
		return
	}
	service := s.service
	s.start(s.label("mcp.refresh", "Refresh"), func(ctx context.Context) (func(*State), error) {
		v := loadSnapshot(service)
		return func(s *State) { s.adopt(v); s.apply() }, nil
	})
}

func (s *State) refresh(withProbe, external bool) {
	service := s.service
	previous := s.fingerprint
	s.start(s.label("mcp.refresh", "Refresh"), func(ctx context.Context) (func(*State), error) {
		v := loadSnapshot(service)
		return func(s *State) {
			s.adopt(v)
			if external && previous != v.fingerprint {
				s.externalChange = true
				s.applyRevision++
			}
			s.pendingProbe = withProbe || (external && previous != v.fingerprint)
		}, nil
	})
}
func (s *State) Probe() {
	if s.Busy || s.closed {
		return
	}
	s.initialProbeAt = time.Time{}
	s.initialProbeWaiting = false
	service := s.service
	s.start(s.label("mcp.checking", "Checking connections…"), func(ctx context.Context) (func(*State), error) {
		ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
		defer cancel()
		raw, e := service.ProbeMCP(ctx)
		if e != nil {
			return nil, e
		}
		p, e := parseProbe(raw)
		if e != nil {
			return nil, e
		}
		return func(s *State) {
			s.checkedAt = time.Now()
			s.probeErrors = p.Errors
			s.probeNote = p.Note
			for _, server := range s.Servers {
				if runtime, ok := p.Servers[server.Name]; ok && enabled(server.Definition) {
					runtime.CheckedAt = s.checkedAt
					s.Runtime[server.Name] = runtime
				}
			}
		}, nil
	})
	s.Probing = true
}
func parseProbe(raw []byte) (probe, error) {
	var payload struct {
		Servers []struct {
			Name, State, Error           string
			Tools                        []json.RawMessage
			ToolExposure                 map[string]string
			Resources, ResourceTemplates int
		}
		Errors []string
		Note   string
	}
	if e := json.Unmarshal(raw, &payload); e != nil {
		return probe{}, e
	}
	if payload.Servers == nil {
		return probe{}, errors.New("Pi MCP probe did not return a servers array")
	}
	out := probe{Servers: map[string]Runtime{}, Errors: payload.Errors, Note: payload.Note}
	for _, entry := range payload.Servers {
		if entry.Name == "" {
			continue
		}
		state := entry.State
		switch state {
		case "disabled", "connecting", "connected", "disconnected", "needs-auth", "failed":
		case "closed":
			state = "disconnected"
		default:
			state = "unknown"
		}
		runtime := Runtime{State: state, Error: entry.Error, Resources: entry.Resources, ResourceTemplates: entry.ResourceTemplates}
		for _, rawTool := range entry.Tools {
			var tool Tool
			if json.Unmarshal(rawTool, &tool.Name) != nil {
				var detail struct {
					Name, Description string
					InputSchema       json.RawMessage
				}
				if json.Unmarshal(rawTool, &detail) != nil {
					continue
				}
				tool.Name, tool.Description, tool.Schema = detail.Name, detail.Description, detail.InputSchema
			}
			if tool.Name == "" {
				continue
			}
			tool.Exposure = entry.ToolExposure[tool.Name]
			runtime.Tools = append(runtime.Tools, tool)
		}
		out.Servers[entry.Name] = runtime
	}
	return out, nil
}
func (s *State) server(name string) *workbench.MCPServer {
	for i := range s.Servers {
		if s.Servers[i].Name == name {
			return &s.Servers[i]
		}
	}
	return nil
}
func (s *State) BeginAdd() {
	s.interruptProbe()
	if !s.Busy {
		s.Editor = newDraft(nil)
		s.editorOpen = true
	}
}
func (s *State) BeginEdit(name string) {
	s.interruptProbe()
	if !s.Busy {
		if server := s.server(name); server != nil {
			s.Editor = newDraft(server)
			s.editorOpen = true
		}
	}
}
func (s *State) saveEditor() {
	if s.Editor == nil || s.Busy {
		return
	}
	prepared, e := s.Editor.prepare(s.Workspace)
	if e != nil {
		s.Editor.Error = e.Error()
		s.Editor.validationError = true
		return
	}
	if len(prepared.secrets) > 0 {
		var names []string
		for _, secret := range prepared.secrets {
			names = append(names, secret.name)
		}
		s.confirm = confirmation{title: s.label("mcp.store_secrets", "Store MCP secrets?"), detail: fmt.Sprintf("%s\n\n%s: %s\n%s", s.label("mcp.store_secrets_detail", "Literal values will be stored in Orbit's private secret store and passed to Pi. The MCP config will contain references, not those values."), prepared.scope, prepared.name, strings.Join(names, "\n")), open: true, accept: func() { s.savePrepared(prepared) }}
		return
	}
	s.savePrepared(prepared)
}
func (s *State) savePrepared(d preparedDraft) {
	service := s.service
	s.start(s.label("mcp.saving", "Saving server…"), func(ctx context.Context) (func(*State), error) {
		if e := ctx.Err(); e != nil {
			return nil, e
		}
		// Refuse malformed metadata before changing either config or secret files.
		if _, errs := service.MCP(); len(errs) > 0 {
			for _, e := range errs {
				if strings.Contains(e.Error(), "metadata") || strings.Contains(e.Error(), "descriptions") {
					return nil, e
				}
			}
		}
		for _, secret := range d.secrets {
			if e := service.SaveSecret(secret.name, secret.value); e != nil {
				return nil, e
			}
		}
		previous := ""
		if d.original != nil && d.original.Scope == d.scope {
			previous = d.original.Name
		}
		err := service.SaveMCP(d.scope, previous, d.name, d.definition)
		modified := err == nil
		if err == nil && d.original != nil && d.original.Scope != d.scope {
			if err = service.RemoveMCP(d.original.Scope, d.original.Name); err != nil {
				err = fmt.Errorf("server saved in %s, but old %s entry could not be removed: %w", d.scope, d.original.Scope, err)
			} else {
				err = service.SetMCPDescription(d.original.Scope, d.original.Name, "")
			}
		}
		if err == nil {
			err = service.SetMCPDescription(d.scope, d.name, d.description)
		}
		v := loadSnapshot(service)
		return func(s *State) {
			s.adopt(v)
			if modified {
				s.modified()
				s.Selected = d.name
			}
			if err == nil {
				s.Editor = nil
				s.Status = s.label("mcp.saved", "Server saved")
			}
			if err != nil && s.Editor != nil {
				s.Editor.Error = err.Error()
				s.Editor.validationError = false
			}
		}, err
	})
}
func (s *State) modified() {
	s.applyRevision++
	s.applyPending = true
	s.externalChange = false
	s.pendingProbe = true
	if s.Options.OnModified != nil {
		s.Options.OnModified()
	}
}
func (s *State) SetEnabled(name string, on bool) {
	s.interruptProbe()
	if server := s.server(name); server != nil {
		scope, service := server.Scope, s.service
		s.start(s.label("mcp.saving", "Saving server…"), func(ctx context.Context) (func(*State), error) {
			e := service.SetMCPEnabled(scope, name, on)
			v := loadSnapshot(service)
			return func(s *State) {
				s.adopt(v)
				if e == nil {
					s.modified()
				}
			}, e
		})
	}
}
func (s *State) remove(name string) {
	s.interruptProbe()
	server := s.server(name)
	if server == nil {
		return
	}
	copy := *server
	service := s.service
	previous := workbench.MCPFingerprint(s.Servers)
	s.removeTarget = ""
	s.start(s.label("mcp.removing", "Removing server…"), func(ctx context.Context) (func(*State), error) {
		e := service.RemoveMCP(copy.Scope, copy.Name)
		v := loadSnapshot(service)
		modified := e == nil || previous != workbench.MCPFingerprint(v.servers)
		return func(s *State) {
			s.adopt(v)
			if modified {
				s.modified()
			}
		}, e
	})
}

// ApplyToken identifies the desired configuration without exposing secrets. It is
// distinct from async operation generation: starting a probe must not invalidate
// a valid runtime reload, but a newer edit or workspace/binary change must.
type ApplyToken struct {
	revision                       uint64
	workspace, fingerprint, binary string
}

func (s *State) ApplyToken() ApplyToken {
	return ApplyToken{s.applyRevision, s.Workspace, s.fingerprint, s.Options.PiBinary}
}
func (s *State) ApplyCurrent(token ApplyToken) bool {
	return !s.closed && s.ApplyToken() == token
}

// AppliedIfCurrent acknowledges a completed reload only while its desired
// configuration request is still current in the settings state.
func (s *State) AppliedIfCurrent(token ApplyToken) bool {
	if !s.ApplyCurrent(token) {
		return false
	}
	s.Applied()
	return true
}

// ApplyFailed leaves pending edits available for an explicit retry. Error text
// is fixed here rather than duplicating possibly sensitive process diagnostics.
func (s *State) ApplyFailed(token ApplyToken) {
	if !s.ApplyCurrent(token) {
		return
	}
	s.applyPending = true
	s.Status = s.label("mcp.apply_failed", "Pi reload failed. Changes remain pending; retry Apply.")
}

// Applied acknowledges a completed, session-preserving host reload. Merely
// clicking Apply never calls it; the host must observe actual runtime success.
func (s *State) Applied() {
	if s.closed {
		return
	}
	s.applyPending = false
	s.externalChange = false
	s.pendingProbe = true
	s.Status = s.label("mcp.applied", "MCP changes applied to Pi")
}
func (s *State) SetPiBinary(binary string) {
	if binary == s.Options.PiBinary || s.closed {
		return
	}
	s.cancelWork()
	s.applyRevision++
	s.Options.PiBinary = binary
	// Probe data belongs to the old executable, not merely the server name.
	s.Runtime = map[string]Runtime{}
	s.checkedAt = time.Time{}
	s.probeErrors = nil
	s.probeNote = ""
	s.Error, s.Status = "", ""
	s.localBusyError = false
	s.service = workbench.New(s.service.Preferences, s.Workspace)
	if binary != "" {
		s.service.PiBinary = binary
	}
	s.loaded = false
	s.pendingProbe = false
	s.quotaProbeRequested = false
	s.initialProbeAt = time.Time{}
	s.initialProbeWaiting = false
}
func (s *State) apply() {
	if s.Options.OnApply != nil {
		s.Status = s.label("mcp.apply_requested", "Reload requested. Any running turn must finish before Pi restarts.")
		s.Options.OnApply()
	}
}
func (s *State) auth(name string, signIn bool) {
	s.interruptProbe()
	if s.Busy {
		return
	}
	service := s.service
	title := s.label("mcp.sign_in", "Sign in")
	if !signIn {
		title = s.label("mcp.sign_out", "Sign out")
	}
	s.start(title, func(ctx context.Context) (func(*State), error) {
		timeout := 120 * time.Second
		args := []string{"mcp", "logout", name}
		if signIn {
			timeout = 330 * time.Second
			args = []string{"mcp", "login", name, "--timeout", "300"}
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		_, e := service.Pi(ctx, args...)
		return func(s *State) {
			s.pendingProbe = true
			if e == nil {
				s.modified()
				s.Status = title + ": " + name
			}
		}, e
	})
	s.authTarget = name
}
func (s *State) openSecret(server workbench.MCPServer) {
	s.interruptProbe()
	reference := generatedSecretName(server.Name, "TOKEN", server.Scope, s.Workspace)
	encoded, _ := json.Marshal(server.Definition)
	if refs := workbench.SecretReferences(string(encoded)); len(refs) > 0 {
		reference = refs[0]
	}
	s.secret = secretPrompt{server: server.Name, scope: server.Scope, name: reference, open: true}
}
func (s *State) saveSecret() {
	if !variableName.MatchString(s.secret.name) {
		s.secret.error = s.label("mcp.error_secret_name", "Enter a valid environment variable name.")
		return
	}
	prompt := s.secret
	if prompt.value == "" {
		s.secret.error = s.label("mcp.error_secret_empty", "Enter the secret value.")
		return
	}
	s.confirm = confirmation{title: s.label("mcp.store_secret", "Store secret?"), detail: fmt.Sprintf("%s\n%s · %s · ${%s}\n\n%s", s.label("mcp.secret_destination", "Save in Orbit's private secret store for this MCP configuration:"), prompt.scope, prompt.server, prompt.name, s.label("mcp.secret_shared_warning", "Pi receives this value as an environment variable. Any configuration using this exact reference will also use the new value.")), open: true, accept: func() {
		service := s.service
		s.start(s.label("mcp.saving_secret", "Saving secret…"), func(ctx context.Context) (func(*State), error) {
			e := service.SaveSecret(prompt.name, prompt.value)
			v := loadSnapshot(service)
			return func(s *State) {
				s.adopt(v)
				if e == nil {
					s.secret = secretPrompt{}
					s.modified()
				}
			}, e
		})
	}}
}

// AddServer is the command-palette entry point for the same typed native form.
func (s *State) AddServer() { s.BeginAdd() }

// A slow read-only probe must never lock the native configuration editor.
// Mutation actions cancel only probes, not another configuration write.
func (s *State) interruptProbe() {
	if s.Probing {
		s.cancelWork()
		s.pendingProbe = false
	}
}
