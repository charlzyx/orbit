package orbitui

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/bugreportui"
	"github.com/charlzyx/orbit/internal/gitworkbench"
	"github.com/charlzyx/orbit/internal/highlight"
	"github.com/charlzyx/orbit/internal/mcpui"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/pirpc"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/resourceui"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/charlzyx/orbit/internal/worktreeui"
	"github.com/egoist/mygo"
	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
)

type RunOptions struct {
	Version                            string
	Home, Workspace, Session, PiBinary string
	Offline, NoExtensions              bool
}
type Controller struct {
	catalogRefreshing bool
	UI                *App
	Agent             *agent.Manager
	Terminal          *orbiterm.Panel
	Git               *gitworkbench.State
	Resources         *resourceui.State
	MCP               *mcpui.State
	configApplyAfter  time.Time
	BugReport         *bugreportui.State
	Worktrees         *worktreeui.State
	Store             *preferences.Store
	Notifier          *native.Notifier
	Updater           *native.Updater
	window            atomic.Pointer[mygo.Window]
	close             sync.Once
	stop              chan struct{}
	workspace         string
	watcher           *workspace.Watcher
	pinned, archived  map[string]bool
	rebuildMenu       func()
	sessionGate       sync.RWMutex
	maintenance       atomic.Bool
	background        context.Context
	cancelBackground  context.CancelFunc
}

func NewController(o RunOptions) (*Controller, error) {
	var s *preferences.Store
	var e error
	if o.Home != "" {
		s = preferences.New(o.Home)
	} else {
		s, e = preferences.Default()
		if e != nil {
			return nil, e
		}
	}
	if o.PiBinary == "" {
		var saved struct {
			PiBinary string `json:"piBinary"`
		}
		s.Read("runtime.json", &saved)
		o.PiBinary = saved.PiBinary
	}
	if o.Workspace == "" {
		o.Workspace = s.LastWorkspace()
	}
	if o.Workspace == "" {
		o.Workspace, _ = os.Getwd()
	}
	o.Workspace, e = filepath.Abs(o.Workspace)
	if e != nil {
		return nil, e
	}
	x := &Controller{Store: s, stop: make(chan struct{}), workspace: o.Workspace, pinned: map[string]bool{}, archived: map[string]bool{}}
	x.background, x.cancelBackground = context.WithCancel(context.Background())
	for _, p := range s.SessionSet(false) {
		x.pinned[p] = true
	}
	for _, p := range s.SessionSet(true) {
		x.archived[p] = true
	}
	prefs := s.Notifications()
	x.Notifier = native.NewNotifier(native.NotificationPrefs{Desktop: prefs.Desktop, Sound: prefs.Sound, Toasts: prefs.Toasts})
	invalidate := func() {
		if w := x.window.Load(); w != nil && !w.IsDestroyed() {
			w.Invalidate()
		}
	}
	executable, _ := os.Executable()
	layout, _ := native.DiscoverInstallLayout(executable)
	x.Updater = native.NewUpdater(native.UpdaterOptions{Version: o.Version, Layout: layout, OnChange: invalidate})
	x.Terminal = orbiterm.NewPanel(orbiterm.Options{Workspace: o.Workspace, Height: float32(s.Layout().TerminalHeight), Font: ghostty.Font{Family: s.Fonts().CodeFamily, Size: float32(s.UI().TerminalFontSize), LineHeight: 1, Features: []string{"-calt", "-liga", "-dlig"}}, OnChange: invalidate, OnError: func(e error) { x.enqueue(func() { x.UI.Notify(e.Error(), true) }) }, OpenLink: func(url string) bool { x.enqueue(func() { x.UI.save(native.OpenURL(url)) }); return true }})
	x.UI = New(Options{BeginRuntimeOperation: x.beginRuntimeOperation, Updater: x.Updater, Version: o.Version, Store: s, Workspace: o.Workspace, PiBinary: o.PiBinary, Terminal: x.Terminal, OnAction: x.Action, Pages: map[string]func(*ui.Context){}})
	x.UI.Options.RefreshModels = x.refreshModelCatalog
	x.UI.Options.QueryModelCatalog = func() {
		if !x.Agent.Snapshot().Connected {
			return
		}
		for _, kind := range []string{pirpc.GetAvailableModels, pirpc.GetAvailableThinkingLevels, pirpc.GetCommands, pirpc.GetState} {
			x.UI.save(x.Agent.Send(kind, nil))
		}
	}
	x.Notifier.OnToast = func(kind native.ToastKind, title, body string) {
		x.enqueue(func() { x.UI.toastStack.Push(kind, title, body) })
	}
	x.Agent = agent.NewManager(agent.Options{PiBinary: o.PiBinary, Home: s.Home, SessionDir: filepath.Join(s.PiDir, "sessions"), Resources: orbitassets.Extensions(), ResolveChildEnv: func(dir string) ([]string, error) { return workbench.New(s, dir).SecretEnv() }, Offline: o.Offline, DisableExtensions: o.NoExtensions, OnChange: invalidate, OnNotification: func(n agent.Notification) {
		x.Notifier.Deliver(native.Notification{Session: n.SessionPath, Title: n.Title, Body: n.Message, Kind: native.ToastKind(n.Kind)})
	}})
	x.UI.Options.Custom = orbiterm.NewCustom(orbiterm.CustomOptions{RowHeight: float32(s.UI().EditorFontSize) * 18 / 13, Font: ghostty.Font{Family: s.Fonts().CodeFamily, Size: float32(s.UI().EditorFontSize) * 12 / 13, Features: []string{"-calt", "-liga", "-dlig"}}, Theme: NewTheme(s.Appearance().LightTheme, s.UI(), s.Fonts()).Terminal(), DarkTheme: NewTheme(s.Appearance().DarkTheme, s.UI(), s.Fonts()).Terminal(), ReduceMotion: s.UI().ReduceMotion, OnChange: invalidate, OnInput: x.Agent.CustomUIInput, OnResize: x.Agent.CustomUIResize, OnError: func(e error) { x.enqueue(func() { x.UI.save(e) }) }})
	x.Git = gitworkbench.New(o.Workspace, gitworkbench.Options{TokenColor: func(class highlight.Class) ui.Color { return x.UI.Theme.TokenColor(class) }, AvatarLoader: gitworkbench.LoadAvatar, AvatarHome: s.Home, OnChange: invalidate, OpenFile: x.UI.OpenFile, RevealFile: native.RevealPath, OpenReview: func(source, ref, path string) {
		x.UI.Navigate("chat")
		x.UI.ReviewVisible = true
		x.UI.ExplorerVisible = false
		x.UI.reviewState.source = workspace.ReviewSource(source)
		x.UI.reviewState.ref = ref
		x.UI.reviewState.path = path
		x.UI.RefreshReview()
	}, OpenURL: func(url string) { x.UI.save(native.OpenURL(url)) }, Authenticate: func() { x.UI.Navigate("chat"); x.UI.save(x.Terminal.NewCommandTab([]string{"gh", "auth", "login"})) }})
	var hints map[string]bool
	_ = s.Read("hints.json", &hints)
	x.Worktrees = worktreeui.New(o.Workspace, worktreeui.Options{FocusComposer: func() { x.UI.composer.Focus() }, OnPickerOpen: func() { x.UI.autocomplete.Dismiss(); x.UI.thinkingOpen = false; x.UI.workspaceOpen = false }, Translate: x.UI.Tr, HideHowTo: hints["worktree_howto_dismissed"], OnDismissHowTo: func() {
		x.UI.save(s.Update("hints.json", func(m map[string]json.RawMessage) error {
			m["worktree_howto_dismissed"] = json.RawMessage("true")
			return nil
		}))
	}, OnChange: invalidate, OpenWorkspace: x.SetWorkspace, OpenTerminal: func(path string) { x.Terminal.SetWorkspace(path); x.UI.save(x.Terminal.NewTab()) }, RevealPath: native.RevealPath, ConfigPath: filepath.Join(s.RootDir, "worktrees.json")})
	x.UI.Options.Escape = x.Git.CloseTransient
	x.UI.Options.WorkIn = func(c *ui.Context) {
		worktreeui.WorkInPicker(c, x.Worktrees, func() { x.UI.Navigate("worktrees") }).Height(22)
	}
	x.UI.Options.Pages["git"] = func(c *ui.Context) { gitworkbench.View(c, x.Git).Fill() }
	x.UI.Options.Pages["worktrees"] = func(c *ui.Context) { worktreeui.View(c, x.Worktrees).Fill() }
	x.BugReport = bugreportui.New(bugreportui.Options{OnChange: invalidate, PiBinary: o.PiBinary, Version: o.Version, Home: s.Home, Window: func() *mygo.Window { return x.window.Load() }, Capture: func() ([]byte, error) {
		if w := x.window.Load(); w != nil && !w.IsDestroyed() {
			return w.CapturePage()
		}
		return nil, errors.New("No window to capture")
	}, OpenURL: native.OpenURL, GenerateDraft: func(ctx context.Context, kind native.ReportKind, notes string) (workspace.Draft, error) {
		if !x.sessionGate.TryRLock() {
			return workspace.Draft{}, errors.New("Runtime maintenance is in progress")
		}
		defer x.sessionGate.RUnlock()
		snap := x.Agent.Snapshot()
		env, e := workbench.New(s, o.Workspace).SecretEnv()
		if e != nil {
			return workspace.Draft{}, e
		}
		cfg := workspace.GenerationConfig{Pi: pirpc.Config{Executable: x.Agent.PiBinary(), WorkspaceDir: s.Home, Environment: append(os.Environ(), env...), EnvOverrides: map[string]string{"PI_CODING_AGENT_DIR": s.PiDir}}, Provider: snap.Model.Provider, Model: snap.Model.ID}
		return workspace.GenerateReportDraftRPC(ctx, cfg, string(kind), notes)
	}, Text: x.UI.text})
	x.MCP = mcpui.New(s, o.Workspace, mcpui.Options{OnChange: invalidate, PiBinary: o.PiBinary, Text: x.UI.text, Theme: func() mcpui.Theme {
		return mcpui.Theme{Elevated: x.UI.Theme.Elevated, CodeFont: x.UI.Fonts.CodeFamily, Px: x.UI.Theme.Px, Space: x.UI.Theme.Space}
	}, OnModified: func() { x.configApplyAfter = time.Now().Add(500 * time.Millisecond) }, OnApply: func() { x.configApplyAfter = time.Now() }})
	x.Resources = resourceui.New(s, o.Workspace, resourceui.Options{Notify: x.UI.Notify, Offline: o.Offline, OnChange: invalidate, OnModified: func() { x.configApplyAfter = time.Now().Add(500 * time.Millisecond) }, Text: x.UI.text, PiBinary: o.PiBinary, Theme: func() resourceui.Theme {
		return resourceui.Theme{Secondary: x.UI.Theme.TextSecondary, Action: x.UI.Theme.Action, ActionText: x.UI.Theme.ActionFg, SelectionText: x.UI.Theme.SelectionFg, Elevated: x.UI.Theme.Elevated, CodeFont: x.UI.Fonts.CodeFamily, Px: x.UI.Theme.Px, Space: x.UI.Theme.Space}
	}, OpenFile: func(path string) { x.UI.save(native.OpenPath(path)) }, RevealPath: native.RevealPath, CopyText: func(text string) { x.Action(Action{Kind: "CopyText", Value: text}) }, InvokeSkill: func(command string) { x.UI.Navigate("chat"); x.UI.Draft = command + " "; x.UI.composer.Focus() }})
	x.UI.Options.Overlay = x.Resources.Overlay
	x.UI.Options.ToastBottom = func(c *ui.Context) float32 {
		h := x.Resources.NoticeHeight(c)
		if h > 0 {
			return 16 + h + 8
		}
		return 16
	}
	x.UI.Options.RefreshMCP = x.MCP.Refresh
	x.UI.Options.Resources = func(c *ui.Context, page string) { x.Resources.View(c, page).Fill() }
	x.UI.Options.RefreshResources = func(page string) {
		if page == "plugins" {
			x.Resources.RefreshPlugins()
		} else {
			x.Resources.Refresh()
		}
	}
	x.UI.Options.MCP = func(c *ui.Context) { x.MCP.View(c).Fill() }
	x.UI.Options.HasModal = func() bool {
		return (x.UI.page == "git" && x.Git.HasModal()) || (x.UI.page == "worktrees" && x.Worktrees.HasModal()) || (x.UI.page == "settings" && ((x.UI.settingsPage == "bug" && x.BugReport.HasModal()) || (x.UI.settingsPage == "mcp" && x.MCP.HasModal()) || ((x.UI.settingsPage == "skills" || x.UI.settingsPage == "plugins") && x.Resources.HasModal())))
	}
	x.UI.Options.BugReport = func(c *ui.Context) { bugreportui.View(c, x.BugReport).Fill() }
	x.UI.Options.OnPage = func(page string) {
		if page != "usage" {
			x.UI.closeUsage()
		}
		switch page {
		case "settings":
			x.UI.loadSettings(x.UI.settingsPage)
		case "files":
			x.UI.RefreshFiles()
		case "usage":
			x.UI.openUsage()
		case "git":
			x.Git.Refresh()
		case "worktrees":
			x.Worktrees.Refresh()
		}
	}
	if e = x.Agent.Open(o.Workspace, o.Session); e != nil {
		x.UI.Notify(e.Error(), true)
	}
	x.bindGenerators(o.Workspace)
	x.Git.Refresh()
	x.watcher = workspace.WatchWorkspace(context.Background(), o.Workspace, 500*time.Millisecond)
	return x, nil
}
func (x *Controller) enqueue(f func()) {
	select {
	case x.UI.updates <- f:
	case <-x.stop:
		return
	}
	if w := x.window.Load(); w != nil && !w.IsDestroyed() {
		w.Invalidate()
	}
}
func (x *Controller) sync() {
	s := x.Agent.Snapshot()
	x.UI.Runtime = x.Agent.RuntimeSnapshot()
	x.UI.SetSnapshot(adaptSnapshot(s, x.workspace, x.pinned, x.archived))
	x.UI.syncUsageStore()
	if d := x.UI.Model.Dialog; d != nil && d.Method == "custom" {
		x.UI.Options.Custom.Update(d.ID, d.Lines, d.Width, d.Height, d.Cursor)
	} else {
		x.UI.Options.Custom.Update("", nil, 0, 0, nil)
	}
	if s.ActiveSession.Workspace != "" && s.ActiveSession.Workspace != x.workspace {
		x.changeWorkspaceServices(s.ActiveSession.Workspace)
	}
	if x.watcher != nil {
		select {
		case _, ok := <-x.watcher.Changes:
			if ok {
				x.UI.RefreshFiles()
				x.Worktrees.WatchChanged()
				if x.UI.ReviewVisible {
					x.UI.RefreshReview()
				}
				if x.UI.page == "git" {
					x.Git.Refresh()
				}
				if x.UI.files != nil {
					ed := x.UI.files.editor
					x.UI.Async(func() func() {
						for _, d := range ed.Documents() {
							ed.Reload(d.Content.Path)
						}
						return func() {
							for _, d := range ed.Documents() {
								x.UI.syncDocument(d)
							}
						}
					})
				}
			}
		default:
		}
	}
	x.MCP.Tick()
	x.Resources.Tick()
	if !x.configApplyAfter.IsZero() && time.Now().After(x.configApplyAfter) && !x.Agent.AnyBusy() && x.sessionGate.TryRLock() {
		x.configApplyAfter = time.Time{}
		err := x.Agent.Restart()
		x.sessionGate.RUnlock()
		x.UI.save(err)
		if err == nil {
			x.UI.Notify("Pi configuration reload requested for the current session. Other open sessions keep their running configuration.", false)
		}
	}
	x.BugReport.Tick()
	x.Git.Tick()
	x.UI.Model.Branch = x.Git.Data.Branch
	x.Worktrees.Tick()
}
func (x *Controller) View(c *ui.Context) {
	x.Agent.Tick()
	x.sync()
	x.UI.View(c)
	interval := 90 * time.Millisecond
	if x.UI.Model.Running {
		interval = 16 * time.Millisecond
	}
	c.After(interval)
}
func (x *Controller) SetWorkspace(path string) {
	if path == "" {
		return
	}
	if !x.sessionGate.TryRLock() {
		x.UI.Notify("Runtime maintenance is in progress. Try again when it finishes.", false)
		return
	}
	defer x.sessionGate.RUnlock()
	x.changeWorkspaceServices(path)
	x.UI.save(x.Store.AddWorkspace(path))
	x.UI.save(x.Store.SaveLastWorkspace(path))
	x.UI.Workspaces = x.Store.Workspaces()
	x.UI.save(x.Agent.NewSession(path))
	x.UI.RefreshFiles()
}
func (x *Controller) changeWorkspaceServices(path string) {
	x.workspace = path
	x.UI.Model.Workspace = path
	x.Terminal.SetWorkspace(path)
	x.Git.SetWorkspace(path)
	x.Git.Refresh()
	x.bindGenerators(path)
	x.Worktrees.SetWorkspace(path)
	x.MCP.SetWorkspace(path)
	x.Resources.SetWorkspace(path)
	if x.watcher != nil {
		old := x.watcher
		go old.Close()
	}
	x.watcher = workspace.WatchWorkspace(context.Background(), path, 500*time.Millisecond)
	if x.UI.page == "settings" {
		x.UI.loadSettings(x.UI.settingsPage)
	}
}
func (x *Controller) Close() {
	x.close.Do(func() {
		close(x.stop)
		x.cancelBackground()
		if x.watcher != nil {
			x.watcher.Close()
		}
		x.Git.Close()
		x.BugReport.Close()
		x.MCP.Close()
		x.Resources.Close()
		x.UI.autocomplete.Close()
		x.UI.Models.Close()
		x.UI.closeBackground()
		x.Worktrees.Close()
		x.Agent.Close()
		x.Terminal.Close()
		x.UI.Options.Custom.Close()
		x.Updater.Close()
		if x.UI.files != nil {
			x.UI.files.editor.CloseAll()
		}
	})
}
func adaptSnapshot(s agent.Snapshot, workspacePath string, pinned, archived map[string]bool) Snapshot {
	v := Snapshot{ModelCatalog: s.Models, CurrentModel: s.Model, LatestTurn: s.LatestTurn, CommandsRaw: s.Commands, RawMessages: s.Messages, Turn: s.Turn, Connected: s.Connected, Running: s.Running, Error: s.ConnectionError, Status: s.Status, SessionID: s.ActiveSession.ID, SessionPath: s.ActiveSession.Path, SessionName: s.ActiveSession.Name, Workspace: workspacePath, ModelID: s.Model.ID, ModelProvider: s.Model.Provider, Thinking: s.ThinkingLevel, ThinkingLevels: s.ThinkingLevels, AuthProviders: s.Auth.Providers, AuthURL: s.Auth.URL, AuthStatus: s.Auth.Status, AuthCode: s.Auth.UserCode}
	v.Extensions = s.Extensions
	v.SteeringMode = s.SteeringMode
	v.FollowUpMode = s.FollowUpMode
	v.AutoCompactionEnabled = s.AutoCompactionEnabled
	v.AutoCompactionKnown = s.AutoCompactionKnown
	v.AccessMode = s.AccessMode
	v.WorkflowMode = s.WorkflowMode
	v.Usage, v.Quotas, v.QuotaSupport = runtimeUsage(s), s.Quotas, s.QuotaSupport
	if s.ActiveSession.Workspace != "" {
		v.Workspace = s.ActiveSession.Workspace
	}
	for _, m := range s.Messages {
		im := []agent.Image{}
		for _, b := range m.Blocks {
			if b.Image != nil {
				im = append(im, *b.Image)
			}
		}
		v.Messages = append(v.Messages, Message{ID: m.ID, Role: m.Role, Text: m.Text, Thinking: m.Thinking, ToolName: m.ToolName, ToolOutput: m.ToolOutput, Error: m.Error, Streaming: m.Streaming, Images: im})
	}
	for _, m := range s.Models {
		v.Models = append(v.Models, ModelChoice{m.ID, m.Provider, m.Name, m.Reasoning})
	}
	for _, r := range s.Sessions {
		v.Sessions = append(v.Sessions, SessionRow{r.ID, r.Path, r.Name, r.Workspace, r.Summary, r.Updated, r.Running, pinned[r.Path], archived[r.Path]})
	}
	v.Queued = append(append([]string{}, s.Queue.Steering...), s.Queue.FollowUp...)
	if len(s.Dialogs) > 0 {
		d := s.Dialogs[0]
		v.Dialog = &DialogRequest{ID: d.ID, Method: d.Method, Title: d.Title, Message: d.Message, Default: d.Value, Options: d.Options, Lines: d.Lines, Width: d.Width, Height: d.Height, Cursor: d.Cursor}
	}
	for _, raw := range s.Commands {
		var cmd struct {
			Name string `json:"name"`
		}
		json.Unmarshal(raw, &cmd)
		if cmd.Name != "" {
			v.Commands = append(v.Commands, cmd.Name)
		}
	}
	return v
}
func (x *Controller) Action(a Action) {
	u, m := x.UI, x.Agent
	if sessionMutation(a.Kind) {
		if !x.sessionGate.TryRLock() {
			u.submitAccepted = false
			u.Notify("Runtime maintenance is in progress. Your draft is kept; try again when it finishes.", false)
			return
		}
		defer x.sessionGate.RUnlock()
	}
	var e error
	str := func(k string) string { v, _ := a.Data[k].(string); return v }
	flag := func(k string) bool { v, _ := a.Data[k].(bool); return v }
	switch a.Kind {
	case "NewSession":
		e = m.NewSession(x.workspace)
	case "SwitchSession":
		x.OpenSession(a.Value)
	case "SetWorkspace":
		x.SetWorkspace(a.Value)
	case "BrowseWorkspace":
		p, err := native.BrowseWorkspace(u.Window, x.workspace)
		e = err
		if p != "" {
			x.SetWorkspace(p)
		}
	case "Submit":
		images := make([]agent.Image, 0, len(u.attachments))
		for _, im := range u.attachments {
			images = append(images, im.Image)
		}
		e = m.Submit(a.Value, images, agent.SendMode(str("mode")))
		if e != nil {
			u.Draft = a.Value
			u.submitAccepted = false
		} else {
			u.attachments = nil
		}
	case "AbortRun":
		if u.Model.Running {
			e = m.Abort()
		}
	case "ClearQueue":
		e = m.ClearQueue()
	case "SetModel":
		e = m.SetModel(str("provider"), a.Value)
	case "SetThinking":
		e = m.SetThinkingLevel(a.Value)
	case "SetAccess":
		e = m.SetAccessMode(a.Value)
	case "SetWorkflow":
		target, mode, supported := m.CaptureWorkflowTarget(), a.Value, u.Model.Extensions.Workflow
		u.Async(func() func() {
			var err error
			if !x.sessionGate.TryRLock() {
				err = errors.New("Runtime maintenance is in progress. Workflow was not changed.")
			} else {
				err = m.SetWorkflowModeFor(target, mode)
				x.sessionGate.RUnlock()
			}
			return func() {
				u.save(err)
				if err == nil && !supported {
					u.Notify(strings.ReplaceAll(u.text("workflow.mode_set_unavailable", "%{mode} selected, but the workflow extension is unavailable."), "%{mode}", titleCase(mode)), true)
				}
			}
		})
	case "RenameSession":
		path := str("path")
		if path == "" {
			path = u.Model.SessionPath
		}
		if path == u.Model.SessionPath {
			e = m.RenameSession(a.Value)
		} else {
			e = m.RenameSavedSession(path, a.Value)
		}
	case "CloneSession":
		path := a.Value
		if path == "" {
			path = u.Model.SessionPath
		}
		e = m.CloneSavedSession(path)
	case "PinSession", "ArchiveSession":
		archived := a.Kind == "ArchiveSession"
		values := x.Store.SessionSet(archived)
		found := false
		for _, p := range values {
			if p == a.Value {
				found = true
			}
		}
		e = x.Store.SetSession(archived, a.Value, !found)
		if e == nil {
			set := x.pinned
			if archived {
				set = x.archived
			}
			set[a.Value] = !found
		}
	case "DeleteSession":
		e = m.DeleteSession(a.Value)
	case "RefreshSessions":
		m.ReloadSessions()
		u.RefreshFiles()
		x.Git.Refresh()
		x.Worktrees.Refresh()
	case "RefreshModels":
		e = x.refreshModelCatalog()
	case "RefreshProviders":
		e = m.Send(pirpc.AuthList, nil)
	case "RefreshQuota":
		e = m.RefreshQuota()
		if e == nil {
			e = m.Send(pirpc.GetEntries, nil)
		}
	case "Login":
		e = m.StartLogin(a.Value, str("method"))
	case "Logout":
		e = m.Logout(a.Value)
	case "CancelLogin":
		e = m.CancelLogin()
	case "RestartRuntime":
		e = m.Restart()
	case "StartRuntime":
		if m.RuntimeSnapshot().Active == nil {
			e = m.Open(u.Model.Workspace, u.Model.SessionPath)
		} else {
			e = m.Start()
		}
	case "StopRuntime":
		e = m.Stop()
	case "SetAutoRetry":
		e = m.SetAutoRetry(flag("enabled"))
	case "SetAutoCompaction":
		e = m.Send(pirpc.SetAutoCompaction, map[string]any{"enabled": flag("enabled")})
	case "SetFollowUpMode", "SetSteeringMode":
		command := pirpc.SetFollowUpMode
		if a.Kind == "SetSteeringMode" {
			command = pirpc.SetSteeringMode
		}
		e = m.Send(command, map[string]any{"mode": a.Value})
	case "CheckPiUpdate":
		u.checkPiUpdate()
	case "ChoosePiBinary":
		u.Prompt("Pi executable path", m.PiBinary(), func(p string) {
			if !x.sessionGate.TryRLock() {
				u.Notify("Runtime maintenance is in progress.", false)
				return
			}
			defer x.sessionGate.RUnlock()
			e := m.SetPiBinary(p)
			u.save(e)
			if e == nil {
				u.Options.PiBinary = p
				x.MCP.SetPiBinary(p)
				x.Resources.SetPiBinary(p)
				u.save(x.Store.Write("runtime.json", map[string]any{"piBinary": p}))
			}
		})
	case "RespondDialog":
		var confirmed *bool
		if _, ok := a.Data["confirmed"]; ok {
			v := flag("confirmed")
			confirmed = &v
		}
		e = m.RespondDialog(str("id"), a.Value, confirmed, flag("cancelled"))
	case "CopyText":
		mygo.Clipboard.WriteText(a.Value)
	case "OpenLink":
		if strings.HasPrefix(a.Value, "https://") || strings.HasPrefix(a.Value, "http://") {
			e = native.OpenURL(a.Value)
		} else if !strings.Contains(a.Value, ":") {
			path := a.Value
			if !filepath.IsAbs(path) {
				path = filepath.Join(x.workspace, path)
			}
			u.OpenFile(path)
		} else {
			e = errors.New("Unsupported link scheme")
		}
	case "RevealPath":
		native.RevealPath(a.Value)
	case "OpenInApp":
		for _, app := range u.OpenApps {
			if app.ID == a.Value {
				dir := x.workspace
				u.OpenPrefs.Remember(dir, app.ID)
				prefs := u.OpenPrefs
				u.Async(func() func() {
					e := native.OpenPathInApp(dir, app)
					if e == nil {
						e = prefs.Save(filepath.Join(x.Store.RootDir, "open-in.json"))
					}
					return func() { u.save(e) }
				})
				break
			}
		}
	case "RevealWorkspace":
		native.RevealPath(x.workspace)
	case "RefreshFiles":
		u.RefreshFiles()
	case "RefreshReview":
		u.RefreshReview()
	case "TerminalFontChanged":
		x.Terminal.SetFont(ghostty.Font{Family: u.Fonts.CodeFamily, Size: float32(u.Prefs.TerminalFontSize), Features: []string{"-calt", "-liga", "-dlig"}})
	case "ExportSession":
		p, err := native.SaveFile(u.Window, "Export session", "session.html", []mygo.FileFilter{{Name: "HTML", Extensions: []string{"html"}}})
		e = err
		if p != "" {
			e = m.Send(pirpc.ExportHTML, map[string]any{"outputPath": p})
		}
	case "ExportUsage":
		format := a.Value
		if format != "json" {
			format = "csv"
		}
		p, err := native.SaveFile(u.Window, "Export usage", "orbit-usage."+format, []mygo.FileFilter{{Name: strings.ToUpper(format), Extensions: []string{format}}})
		e = err
		if p != "" {
			idx, ok := a.Data["index"].(workbench.UsageIndex)
			filter, _ := a.Data["filter"].(workbench.UsageFilter)
			if ok {
				u.Async(func() func() {
					f, e := os.Create(p)
					if e == nil {
						if format == "json" {
							options, _ := a.Data["options"].(workbench.UsageExportOptions)
							e = idx.ExportJSON(f, filter, options)
						} else {
							e = idx.ExportCSV(f, filter)
						}
						if ce := f.Close(); e == nil {
							e = ce
						}
					}
					return func() { u.save(e) }
				})
			}
		}
	case "BrowseBackground":
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: u.Window, Title: u.text("settings.choose_background_title", "Choose a background image"), Filters: []mygo.FileFilter{{Name: "Images", Extensions: []string{"png", "jpg", "jpeg", "gif", "webp", "bmp", "tiff"}}}})
		if err != nil {
			u.Notify(err.Error(), true)
		} else if len(paths) > 0 {
			u.chooseBackground(paths[0])
		}
	case "BrowseAttachments":
		paths, err := native.BrowseAttachments(u.Window, x.workspace)
		e = err
		if e == nil {
			u.AddPaths(paths)
		}
	case "PasteImage":
		raw := mygo.Clipboard.ReadImage()
		if len(raw) == 0 {
			e = errors.New("Clipboard does not contain an image")
		} else {
			u.AddImage(raw, "image/png", "Pasted image")
		}
	case "AddMcpServer":
		u.settingsPage = "mcp"
		u.Navigate("settings")
		x.MCP.AddServer()
	case "RefreshMcpServers", "ReconnectMcpServers":
		x.MCP.Refresh()
		if a.Kind == "ReconnectMcpServers" {
			x.configApplyAfter = time.Now()
		}
	case "AutoTitleChanged":
		// The title extension reads its configuration live; no process restart.
	case "RuntimeConfigChanged":
		x.configApplyAfter = time.Now().Add(500 * time.Millisecond)
	case "NotificationsChanged":
		p := x.Store.Notifications()
		x.Notifier.SetPrefs(native.NotificationPrefs{Desktop: p.Desktop, Sound: p.Sound, Toasts: p.Toasts})
	case "RunRetention":
		x.runRetention(true)
	case "LanguageChanged":
		if x.rebuildMenu != nil {
			x.rebuildMenu()
		}
	case "CheckForUpdates":
		u.Async(func() func() { e := x.Updater.CheckAndStage(context.Background()); return func() { u.save(e) } })
	case "RuntimeChanged":
		if !m.AnyBusy() {
			e = m.Restart()
		} else {
			u.Notify("Pi was updated. Running tasks keep their current process; new processes use the update.", false)
		}
	case "Quit":
		mygo.App.Quit()
	default:
		if strings.HasPrefix(a.Kind, "GitTab") {
			ids := []string{"Changes", "History", "Graph", "Stashes", "Issues", "Pulls"}
			for i, id := range ids {
				if a.Kind == "GitTab"+id {
					x.Git.SetTab(i)
				}
			}
		} else {
			e = fmt.Errorf("Action is not available: %s", a.Kind)
		}
	}
	if e != nil {
		u.Notify(e.Error(), true)
	}
}
func Run(o RunOptions) error {
	if e := LoadAssets(); e != nil {
		return e
	}
	x, e := NewController(o)
	if e != nil {
		return e
	}
	defer x.Close()
	mygo.App.SetName("Orbit")
	var lifecycle *native.Lifecycle
	create := func() *mygo.Window {
		display := mygo.Screen.PrimaryDisplay()
		width, height := windowRestoreSize(display.Bounds.Width, display.Bounds.Height)
		options := sourceWindowOptions(width, height)
		options.Content = ui.View(x.View)
		w := mygo.NewWindow(options)
		x.window.Store(w)
		x.UI.Window = w
		x.UI.window.Store(w)
		w.OnReadyToShow(func() { go func() { _ = native.SignalUpdateReady() }() })
		return w
	}
	lifecycle = native.BindLifecycle(native.LifecycleOptions{Reopen: create, OpenSession: func(path string) { x.OpenSession(path) }, OpenFile: x.UI.OpenFile, Quit: x.Close, Notifier: x.Notifier})
	defer lifecycle.Close()
	x.rebuildMenu = func() {
		mygo.App.SetMenu(native.NativeMenu(native.MenuOptions{AppName: "Orbit", Text: x.UI.text, CanCheckUpdates: x.Updater.Snapshot().Enabled, Dispatch: func(command string) {
			if command != "Quit" {
				lifecycle.Show()
			}
			x.UI.nativeCommand(command)
		}}))
	}
	mygo.App.WhenReady(func() {
		x.rebuildMenu()
		lifecycle.SetWindow(create())
		x.UI.Async(func() func() {
			apps, e := native.DetectOpenInApps()
			prefs, e2 := native.LoadOpenInPrefs(filepath.Join(x.Store.RootDir, "open-in.json"))
			return func() { x.UI.OpenApps = apps; x.UI.OpenPrefs = prefs; x.UI.save(e); x.UI.save(e2) }
		})
		x.Updater.StartAutomatic(context.Background(), filepath.Join(x.Store.RootDir, "updates.json"))
		x.runRetention(false)
		if !o.Offline && o.PiBinary == "" && x.Agent.PiBinary() == "" {
			native.CheckPiUpdateOnLaunch(x.background, native.PiLaunchUpdateOptions{Busy: func() bool { return x.maintenance.Load() || x.Agent.AnyBusy() }, BeginUpdate: func() (func(), bool) {
				release, err := x.beginRuntimeOperation()
				if err == nil && x.Agent.PiBinary() != "" {
					release()
					return nil, false
				}
				return release, err == nil
			}, OnEvent: func(event native.PiUpdateEvent) {
				x.enqueue(func() {
					s := &x.UI.nativeSettings
					s.latest, s.installation = event.Latest, event.Installation
					if len(event.Dependencies) > 0 {
						s.dependencies = event.Dependencies
						s.runtimeLoaded = true
					}
					s.output = event.Detail
					switch event.State {
					case "updated":
						x.UI.Notify("Pi updated to "+event.Installation.Version, false)
						x.Action(Action{Kind: "RuntimeChanged"})
					case "failed":
						x.UI.Notify("Pi update failed: "+event.Detail, true)
					case "migration_required":
						x.UI.Notify("A managed Pi installation is available. Review migration in Settings → Runtime.", false)
					}
				})
			}})
		}
	})
	// A closed macOS window still drains parked Pi sessions and timeout dialogs.
	go func() {
		timer := time.NewTicker(90 * time.Millisecond)
		defer timer.Stop()
		for {
			select {
			case <-x.stop:
				return
			case <-timer.C:
				if w := x.window.Load(); w == nil || w.IsDestroyed() {
					x.Terminal.SetVisible(false)
					x.Agent.Tick()
				}
			}
		}
	}()
	return mygo.App.Run()
}

func (x *Controller) bindGenerators(dir string) {
	config := func() (workspace.GenerationConfig, error) {
		snap := x.Agent.Snapshot()
		env, e := workbench.New(x.Store, dir).SecretEnv()
		return workspace.GenerationConfig{Pi: pirpc.Config{Executable: x.Agent.PiBinary(), WorkspaceDir: dir, Environment: append(os.Environ(), env...), EnvOverrides: map[string]string{"PI_CODING_AGENT_DIR": x.Store.PiDir}}, Provider: snap.Model.Provider, Model: snap.Model.ID}, e
	}
	x.Git.Options.GenerateCommitResult = func(ctx context.Context, stagedOnly bool) (workspace.GeneratedText, error) {
		if !x.sessionGate.TryRLock() {
			return workspace.GeneratedText{}, errors.New("Runtime maintenance is in progress")
		}
		defer x.sessionGate.RUnlock()
		cfg, e := config()
		if e != nil {
			return workspace.GeneratedText{}, e
		}
		return workspace.GenerateCommitRPC(ctx, workspace.Git{Dir: dir}, cfg, stagedOnly)
	}
	x.Git.Options.GenerateDraft = func(ctx context.Context, request workspace.DraftRequest) (workspace.Draft, error) {
		if !x.sessionGate.TryRLock() {
			return workspace.Draft{}, errors.New("Runtime maintenance is in progress")
		}
		defer x.sessionGate.RUnlock()
		cfg, e := config()
		if e != nil {
			return workspace.Draft{}, e
		}
		return workspace.GenerateDraftRPC(ctx, workspace.Git{Dir: dir}, cfg, request)
	}
}

func (x *Controller) OpenSession(path string) {
	if !x.sessionGate.TryRLock() {
		x.UI.Notify("Runtime maintenance is in progress. Try opening the task again when it finishes.", false)
		return
	}
	defer x.sessionGate.RUnlock()
	for _, s := range x.Agent.Snapshot().Sessions {
		if s.Path == path {
			dir := s.Workspace
			if dir == "" {
				dir = x.workspace
			}
			x.UI.save(x.Agent.Open(dir, path))
			return
		}
	}
	x.UI.Async(func() func() {
		s, e := agent.ReadSession(path)
		return func() {
			if e != nil {
				x.UI.save(e)
				return
			}
			if !x.sessionGate.TryRLock() {
				x.UI.Notify("Runtime maintenance is in progress. Try opening the task again when it finishes.", false)
				return
			}
			defer x.sessionGate.RUnlock()
			dir := s.Workspace
			if dir == "" {
				dir = x.workspace
			}
			x.UI.save(x.Agent.Open(dir, path))
		}
	})
}

// Source main_window.rs::open sizes restore bounds relative to the primary display.
func windowRestoreSize(displayWidth, displayHeight int) (int, int) {
	if displayWidth <= 0 || displayHeight <= 0 {
		return 1240, 840
	}
	return min(1440, int(float64(displayWidth)*.85)), min(920, int(float64(displayHeight)*.9))
}

// The source has no in-window menu bar on Linux. Auto-hide retains native edit
// roles and accelerators without allocating an unwanted GTK/Windows strip.
func sourceWindowOptions(width, height int) mygo.WindowOptions {
	return mygo.WindowOptions{Title: "Orbit", Width: width, Height: height, MinWidth: 960, MinHeight: 640, Maximized: true, AutoHideMenuBar: true, HideWindowButtons: runtime.GOOS == "linux", TitleBarStyle: mygo.TitleBarHidden, TitleBarHeight: 44, TrafficLightPosition: &mygo.Point{X: 12, Y: 16}}
}
