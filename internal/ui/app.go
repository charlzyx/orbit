// Package orbitui renders Orbit entirely with MyGo's native Go toolkit.
package orbitui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/composerui"
	"github.com/charlzyx/orbit/internal/markdownimage"
	"github.com/charlzyx/orbit/internal/modelui"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/preferences"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/charlzyx/orbit/internal/transcriptui"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/charlzyx/orbit/internal/workspace"
	"github.com/egoist/mygo"
	ghostty "github.com/egoist/mygo/orbit-terminal"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

type SessionRow struct {
	ID, Path, Name, Workspace, Summary string
	Updated                            time.Time
	Running, Pinned, Archived          bool
}
type Message struct {
	SuppressFooter                                        bool
	Logical                                               *transcriptui.Row
	ChangedFiles                                          []transcriptui.ChangedFile
	Images                                                []agent.Image
	ID, Role, Text, Thinking, ToolName, ToolOutput, Error string
	Streaming                                             bool
}
type ModelChoice struct {
	ID, Provider, Name string
	Reasoning          bool
}
type DialogRequest struct {
	Lines                                            []string
	Width, Height                                    int
	Cursor                                           json.RawMessage
	ID, Method, Title, Message, Placeholder, Default string
	Options                                          []string
}
type Usage struct {
	Input, Output, CacheRead, CacheWrite, Total                                              int64
	Cost                                                                                     float64
	ContextPercent                                                                           float64
	Available                                                                                bool
	TokensAvailable, CostAvailable, ContextAvailable                                         bool
	ContextTokens, ContextWindow                                                             int64
	ContextTokensAvailable, SessionAvailable                                                 bool
	InputAvailable, OutputAvailable, CacheReadAvailable, CacheWriteAvailable, TotalAvailable bool
}
type Snapshot struct {
	RetryFailure                                                                                            agent.RetryFailure
	ExtensionStatusText                                                                                     string
	ExtensionStatusAt                                                                                       time.Time
	RuntimeID, ComposerText                                                                                 string
	ComposerRevision                                                                                        uint64
	Approval                                                                                                *agent.ApprovalRequest
	Ask                                                                                                     *agent.AskPrompt
	ExtensionWidgets                                                                                        []agent.ExtensionWidget
	ExtensionNotices                                                                                        []agent.ExtensionNotice
	Auth                                                                                                    agent.AuthState
	Retrying, Compacting                                                                                    bool
	RetryDetail                                                                                             agent.RetryDetail
	ModelCatalog                                                                                            []agent.ModelChoice
	CurrentModel                                                                                            agent.ModelChoice
	LatestTurn                                                                                              int
	RawMessages                                                                                             []agent.Message
	AutoCompactionKnown                                                                                     bool
	Extensions                                                                                              agent.ExtensionStatus
	SteeringMode, FollowUpMode, AccessMode, WorkflowMode                                                    string
	AutoCompactionEnabled                                                                                   bool
	Connected, Running                                                                                      bool
	Turn                                                                                                    int
	Error, Status, SessionID, SessionPath, SessionName, Workspace, ModelID, ModelProvider, Thinking, Branch string
	Messages                                                                                                []Message
	Sessions                                                                                                []SessionRow
	Models                                                                                                  []ModelChoice
	ThinkingLevels, Commands                                                                                []string
	CommandsRaw                                                                                             []json.RawMessage
	Dialog                                                                                                  *DialogRequest
	Queue                                                                                                   agent.Queue
	Queued                                                                                                  []string
	Usage                                                                                                   Usage
	AuthProviders                                                                                           []agent.AuthProvider
	AuthURL, AuthStatus, AuthCode                                                                           string
	Quotas                                                                                                  []agent.QuotaReport
	QuotaSupport                                                                                            string
}
type Action struct {
	Kind, Value, SessionID string
	Data                   map[string]any
}
type Options struct {
	// UI-owned task identity and intent generation for delayed attachment work.
	AttachmentContext     func() (agent.WorkflowTarget, uint64, bool)
	RefreshModels         func() error
	QueryModelCatalog     func()
	BeginRuntimeOperation func() (func(), error)
	Version               string
	Custom                *orbiterm.CustomPanel
	Updater               *native.Updater
	Store                 *preferences.Store
	OnAction              func(Action)
	Terminal              *orbiterm.Panel
	// Optional workbench modules mount their actual native views here.
	Pages            map[string]func(*ui.Context)
	OnPage           func(string)
	WorkIn           func(*ui.Context)
	BugReport        func(*ui.Context)
	MCP              func(*ui.Context)
	Resources        func(*ui.Context, string)
	RefreshResources func(string)
	Overlay          func(*ui.Context)
	ToastBottom      func(*ui.Context) float32
	RefreshMCP       func()
	HasModal         func() bool
	Escape           func() bool
	Workspace        string
	PiBinary         string
}
type draftState struct {
	Text        string
	Attachments []attachment
}
type Notice struct {
	Text  string
	Error bool
	Until time.Time
}
type App struct {
	setup                  onboardingState
	pendingAttachmentPaths []attachmentPathInsertion
	navigation             navigationState
	transcriptScroll       ui.ScrollState
	railHintDismissed      bool
	railHintSaving         bool
	railHintAt             time.Time

	retryStatusAt                                                   time.Time
	retryFeedbackNow                                                func() time.Time
	retryCancellation                                               retryCancellation
	contextOpen                                                     bool
	contextDismissed                                                time.Time
	extensionUI                                                     extensionUIState
	rpcPatchSaving                                                  bool
	Models                                                          *modelui.State
	modelDismissedAt                                                time.Time
	workspacePicker                                                 workspacePickerState
	transcriptHover                                                 string
	transcriptUI                                                    transcriptui.State
	autocomplete                                                    composerui.Controller
	backdrop                                                        backdropState
	Runtime                                                         agent.RuntimeSnapshot
	notificationPrefs                                               preferences.Notifications
	sessionDefaultPrefs                                             preferences.SessionDefault
	autoTitlePrefs                                                  preferences.AutoTitle
	worktreePrefs                                                   workspace.WorktreeConfig
	Options                                                         Options
	Window                                                          *mygo.Window
	window                                                          atomic.Pointer[mygo.Window]
	Model                                                           Snapshot
	Prefs                                                           preferences.UI
	Workspaces                                                      preferences.WorkspaceList
	workspaceActions                                                workspaceMenuState
	OpenApps                                                        []native.ExternalApp
	OpenPrefs                                                       native.OpenInPrefs
	Fonts                                                           preferences.Fonts
	Appearance                                                      preferences.Appearance
	Layout                                                          preferences.Layout
	Theme                                                           Theme
	terminalThemeID                                                 string
	customAppearanceKey                                             string
	catalog, english                                                map[string]string
	page, settingsPage, previousPage                                string
	SidebarVisible, ExplorerVisible, ReviewVisible                  bool
	SidebarWidth, ExplorerWidth, ReviewWidth                        float32
	Draft, Search, paletteQuery, modelQuery                         string
	attachments                                                     []attachment
	messageImages                                                   map[string]ui.ImageSource
	imageLoading                                                    map[string]bool
	imageErrors                                                     map[string]string
	imageOpen                                                       bool
	imagePreview                                                    ui.ImageSource
	nextAttachment                                                  uint64
	submitAccepted                                                  bool
	drafts                                                          map[string]draftState
	pendingNew                                                      bool
	searchOpen, paletteOpen, modelOpen, thinkingOpen, workspaceOpen bool
	commandCursor, searchCursor                                     int
	modelCursor                                                     int
	modelList                                                       ui.ListState
	surfaceFocus                                                    workbenchFocusState
	composer                                                        ui.Handle
	searchInput                                                     ui.Handle
	searchLastQuery                                                 string
	searchLastCount                                                 int
	sessionList, transcript, paletteList                            ui.ListState
	sessionCursor                                                   int
	sidebarState                                                    sidebarState
	collapsedMessages                                               map[string]bool
	md                                                              markdown.Options
	markdownImages                                                  *markdownimage.Cache
	access, workflow                                                string
	dialogValue                                                     string
	dialogID                                                        string
	promptOpen                                                      bool
	promptTitle, promptValue                                        string
	promptCommit                                                    func(string)
	confirmOpen                                                     bool
	confirmTitle, confirmMessage                                    string
	confirmCommit                                                   func()
	toastStack                                                      native.ToastStack
	history                                                         []string
	historyAt                                                       int
	settingsSearch                                                  string
	settingsScroll                                                  map[string]*ui.ScrollState
	skills                                                          []workbench.Skill
	plugins                                                         []workbench.Plugin
	servers                                                         []workbench.MCPServer
	settingsErrors                                                  []string
	settingsLoading                                                 bool
	settingsGeneration                                              uint64
	selectedSkill                                                   int
	jsonOpen                                                        bool
	jsonTitle, jsonName, jsonScope, jsonValue                       string
	jsonCommit                                                      func(string, string, string) error
	updates                                                         chan func()
	files                                                           *filesState
	reviewState                                                     reviewState
	usageState                                                      usageState
	nativeSettings                                                  nativeSettingsState
	secretOpen                                                      bool
	secretProvider, secretValue                                     string
	providerState                                                   providerState
	providers                                                       []workbench.Provider
	credentials                                                     map[string]workbench.CredentialStatus
	retention                                                       preferences.Retention
	retentionBusy                                                   bool
	quotaOpen                                                       bool
	quotaVisibility                                                 quotaVisibilityState
	quotaMCP                                                        quotaMCPState
	modalWasOpen, restoreComposerAfterModal                         bool
}

func New(o Options) *App {
	if o.Store == nil {
		o.Store = preferences.New("")
	}
	a := &App{Options: o, Prefs: o.Store.UI(), Workspaces: o.Store.Workspaces(), Fonts: o.Store.Fonts(), Appearance: o.Store.Appearance(), Layout: o.Store.Layout(), SidebarVisible: true, settingsPage: "general", page: "chat", sessionCursor: -1, collapsedMessages: map[string]bool{}, history: []string{"chat"}, updates: make(chan func(), 64), drafts: map[string]draftState{}, messageImages: map[string]ui.ImageSource{}, imageLoading: map[string]bool{}, imageErrors: map[string]string{}, selectedSkill: -1}
	a.SidebarWidth = float32(a.Layout.SidebarWidth)
	if a.SidebarWidth < SidebarMinWidth {
		a.SidebarWidth = 248
	}
	a.ExplorerWidth = float32(a.Layout.ExplorerWidth)
	a.ReviewWidth = float32(a.Layout.SidepaneWidth)
	a.reviewState.wrap = true
	a.reviewState.tree = true
	a.access = o.Store.Access()
	a.workflow = o.Store.Workflow("")
	a.settingsScroll = map[string]*ui.ScrollState{}
	a.Model.Workspace = o.Workspace
	a.retention = o.Store.Retention()
	a.loadQuotaVisibility()
	a.notificationPrefs = o.Store.Notifications()
	a.sessionDefaultPrefs = o.Store.SessionDefault()
	a.autoTitlePrefs = o.Store.AutoTitle()
	a.worktreePrefs = workspace.LoadWorktreeConfig(filepath.Join(o.Store.RootDir, "worktrees.json"))
	a.english, _ = orbitassets.Locale("en")
	a.reloadLocale()
	a.Theme = NewTheme(a.Appearance.DarkTheme, a.Prefs, a.Fonts)
	var railHints map[string]bool
	_ = o.Store.Read("hints.json", &railHints)
	a.railHintDismissed = railHints["rail_hint_seen"]
	a.transcript.FollowEnd = true
	a.transcript.Key = func(i int) any { return a.Model.Messages[i].ID }
	a.sessionList.Key = func(i int) any { return a.Model.Sessions[i].ID }
	a.md.ShareSelection = true
	a.md.Block = a.markdownBlock
	a.md.LoadImage = markdown.FileImages(o.Workspace)
	a.markdownImages = markdownimage.New()
	a.md.Reveal = &markdown.Reveal{Unit: markdown.AtOnce, Fade: -1}
	a.md.OnLink = func(s string) { a.Dispatch(Action{Kind: "OpenLink", Value: s}) }
	a.loadUsagePrefs()
	a.initBackground()
	a.initModels()
	return a
}
func (a *App) reloadLocale() {
	lang := a.Prefs.Language
	if lang == "system" || lang == "" {
		lang = systemLanguage()
	}
	a.catalog, _ = orbitassets.Locale(lang)
}
func (a *App) Tr(key string) string { return orbitassets.Translate(a.catalog, a.english, key, nil) }
func (a *App) text(key, fallback string) string {
	s := a.Tr(key)
	if s == key {
		return fallback
	}
	return s
}
func (a *App) SetSnapshot(s Snapshot) {
	if s.RuntimeID != a.Model.RuntimeID || s.SessionID != a.Model.SessionID || s.SessionPath != a.Model.SessionPath || s.Workspace != a.Model.Workspace {
		a.resetMarkdownImageOwner()
	}
	a.invalidateWorkspaceMenu(s.Workspace)
	oldID, oldLatest := a.Model.SessionID, a.Model.LatestTurn
	a.transcriptUI.Session(s.SessionID)
	a.projectTranscript(&s)
	a.navigation.turns = nil
	if s.SessionID != a.Model.SessionID || s.RuntimeID != a.Model.RuntimeID {
		a.navigation = navigationState{}
		a.contextOpen = false
		a.retryCancellation = retryCancellation{}
		a.retryStatusAt = time.Time{}
	}
	if s.SessionID != a.Model.SessionID {
		a.transcriptScroll = ui.ScrollState{}
		a.messageImages = map[string]ui.ImageSource{}
		a.imageLoading = map[string]bool{}
		a.imageErrors = map[string]string{}
		a.imageOpen = false
		a.transcript = ui.ListState{FollowEnd: true}
		a.transcript.Key = func(i int) any { return a.Model.Messages[i].ID }
		a.workflow = "build"
		if a.pendingNew {
			a.pendingNew = false
		} else if a.Model.SessionID != "" {
			a.drafts[a.Model.SessionID] = draftState{a.Draft, append([]attachment(nil), a.attachments...)}
			d := a.drafts[s.SessionID]
			a.Draft = d.Text
			a.attachments = d.Attachments
		}
	}
	if s.RetryFailure.Active {
		message := s.RetryFailure.Message
		if !s.RetryFailure.MessageProvided {
			message = a.Tr("events.retries_exhausted")
		}
		s.Error = orbitassets.Translate(a.catalog, a.english, "events.auto_retry_failed", map[string]string{"error": message})
	}
	a.Model = s
	a.pruneAttachmentPaths()
	a.syncExtensionUI(s)
	a.syncModelCatalog()
	if a.reviewState.source == workspace.ReviewLastTurn {
		if s.LatestTurn == 0 {
			a.reviewState.source = workspace.ReviewUncommitted
			a.reviewState.turn = 0
		}
		if a.ReviewVisible && (oldID != s.SessionID || oldLatest != s.LatestTurn) {
			a.RefreshReview()
		}
	}
	if s.AccessMode != "" {
		a.access = s.AccessMode
	}
	if s.WorkflowMode != "" {
		a.workflow = s.WorkflowMode
	}
	if s.Dialog != nil && s.Dialog.ID != a.dialogID {
		a.dialogID = s.Dialog.ID
		a.dialogValue = s.Dialog.Default
	}
}
func (a *App) Dispatch(action Action) {
	if action.Kind == "AbortRetry" && (action.SessionID == "" || action.SessionID == a.Model.SessionID) {
		a.cancelRetryPresentation()
	}
	if action.SessionID == "" {
		action.SessionID = a.Model.SessionID
	}
	if a.Options.OnAction != nil {
		a.Options.OnAction(action)
	} else {
		a.Notify("This action needs a connected runtime: "+action.Kind, true)
	}
}
func (a *App) Notify(text string, isError bool) {
	kind := native.ToastInfo
	if isError {
		kind = native.ToastError
	}
	a.toastStack.Push(kind, "", text)
}

func (a *App) save(err error) {
	if err != nil {
		a.Notify(err.Error(), true)
	}
}
func (a *App) Navigate(page string) {
	a.workspaceOpen = false
	if a.Models != nil {
		a.Models.Escape()
	}
	a.modelOpen = false
	a.thinkingOpen = false
	a.flushFilesOnLeave(page)
	a.previousPage = a.page
	a.page = page
	if a.historyAt < len(a.history)-1 {
		a.history = a.history[:a.historyAt+1]
	}
	if a.history[len(a.history)-1] != page {
		a.history = append(a.history, page)
		a.historyAt = len(a.history) - 1
	}
	if a.Options.OnPage != nil {
		a.Options.OnPage(page)
	}
}
func (a *App) goHistory(delta int) {
	i := a.historyAt + delta
	if i >= 0 && i < len(a.history) {
		a.flushFilesOnLeave(a.history[i])
		a.historyAt = i
		a.page = a.history[i]
		if a.Options.OnPage != nil {
			a.Options.OnPage(a.page)
		}
	}
}
func (a *App) flushFilesOnLeave(next string) {
	if a.page == "files" && next != "files" && a.files != nil {
		a.saveFiles()
	}
}
func (a *App) View(c *ui.Context) {
	a.workspaceActions.rendered = false
	a.drainUpdates()
	a.pruneAttachmentPaths()
	if target := a.workspaceActions.target; a.workspaceActions.open && (target == nil || a.workspaceEntry(target.Path) != target) {
		a.workspaceActions.open = false
	}
	systemDark := c.Theme().Dark
	id := a.Appearance.DarkTheme
	if a.Appearance.Mode == "light" || (a.Appearance.Mode == "system" && !systemDark) {
		id = a.Appearance.LightTheme
	}
	a.Theme = NewTheme(id, a.Prefs, a.Fonts)
	if a.Options.Terminal != nil && a.terminalThemeID != id {
		a.Options.Terminal.SetTheme(a.Theme.Terminal(), a.Theme.Terminal())
		a.terminalThemeID = id
	}
	appearanceKey := fmt.Sprintf("%s|%s|%g|%t", id, a.Fonts.CodeFamily, a.Prefs.EditorFontSize, a.Prefs.ReduceMotion)
	if a.Options.Custom != nil && a.customAppearanceKey != appearanceKey {
		a.Options.Custom.SetAppearance(ghostty.Font{Family: a.Fonts.CodeFamily, Size: float32(a.Prefs.EditorFontSize) * 12 / 13, Features: []string{"-calt", "-liga", "-dlig"}}, a.Theme.Terminal(), a.Theme.Terminal(), a.Prefs.ReduceMotion)
		a.Options.Custom.SetRowHeight(float32(a.Prefs.EditorFontSize) * 18 / 13)
		a.customAppearanceKey = appearanceKey
	}
	c.SetTheme(a.Theme.Native())
	a.md.Theme = a.Theme.Markdown()
	a.md.Theme.CodeColors = a.Theme.Code()
	if a.Prefs.ReduceMotion {
		c.DisableAnimations()
	}
	a.syncReviewLayout(c)
	root := ui.Row(c.Key("root")).Fill().Background(a.Theme.Background).Font(a.Fonts.UIFamily).TextColor(a.Theme.Text).Children(func() {
		if a.page == "settings" {
			a.settings(c)
			return
		}
		if a.setupBlocked() {
			a.onboardingView(c)
			return
		}
		// Keep the workspace's native ancestry stable when the sidebar hides.
		// A conditional Split reparents every editor and discards private undo.
		if a.SidebarVisible {
			a.sidebar(c)
		}
		a.main(c)
		if a.SidebarVisible {
			a.sidebarResize(c)
		}
	})
	if paths := root.DroppedFiles(); len(paths) > 0 && a.page != "settings" {
		a.AddPaths(paths)
	}
	if a.Options.Terminal != nil {
		a.Options.Terminal.SetVisible(a.page == "chat" && !a.reviewFull() && !a.setupBlocked())
	}
	if a.workspaceActions.open && !a.workspaceActions.rendered {
		a.workspaceActions.open = false
	}
	a.shortcuts(c)
	modalNow := a.modalOpen()
	if !modalNow && !a.setupBlocked() && a.Options.Overlay != nil {
		a.Options.Overlay(c)
	}
	if modalNow && !a.modalWasOpen {
		a.restoreComposerAfterModal = a.composer.Focused(c)
	} else if !modalNow && a.modalWasOpen && a.restoreComposerAfterModal && a.page == "chat" {
		a.composer.Focus()
		a.restoreComposerAfterModal = false
	}
	a.modalWasOpen = modalNow
	if a.paletteOpen {
		a.commandPalette(c)
	}
	if a.modelOpen && a.Models == nil {
		a.modelPicker(c)
	}
	if a.Model.Dialog != nil {
		a.extensionDialog(c)
	}
	a.imageDialog(c)
	a.prompt(c)
	a.confirm(c)
	a.jsonDialog(c)
	a.secretDialog(c)
	a.providerDialog(c)
	a.toasts(c)
}
func (a *App) icon(c *ui.Context, name string, size float32, color ui.Color) ui.Element {
	aliases := map[string]string{"icons/external-link.svg": "icons/arrow-up-right.svg", "icons/chart.svg": "icons/chart-analysis.svg", "icons/palette.svg": "icons/contrast.svg", "icons/models.svg": "icons/spark.svg", "icons/save.svg": "icons/check.svg", "icons/download.svg": "icons/arrow-down.svg", "icons/git-branch.svg": "icons/branch.svg", "icons/bug.svg": "icons/info.svg"}
	if v := aliases[name]; v != "" {
		name = v
	}
	if s := icons[name]; s != nil {
		return ui.Icon(c, s).Size(size, size).TextColor(color)
	}
	return ui.Text(c, "·").Size(size, size).TextColor(color).Unselectable()
}
func (a *App) button(c *ui.Context, label, icon string, selected bool) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Justify(ui.Start).AlignItems(ui.Center).Label(label).Height(t.Px(30)).Padding(t.Space(4), t.Space(8)).Radius(t.Px(6)).Gap(t.Space(6))
	if selected {
		b.Background(t.Selection).TextColor(t.SelectionFg)
	} else if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() {
		if icon != "" {
			a.icon(c, icon, t.Px(16), t.TextSecondary)
		}
		if label != "" {
			ui.Text(c, label).FontSize(t.FontSize)
		}
	})
	return b
}
func (a *App) iconButton(c *ui.Context, label, icon string) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c).Label(label).Tooltip(label).Size(t.Px(28), t.Px(28)).Radius(t.Px(6)).Center()
	if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() { a.icon(c, icon, t.Px(16), t.TextSecondary) })
	return b
}
func (a *App) line(c *ui.Context) {
	ui.Box(c).Height(1).Background(a.Theme.Border).AlignSelf(ui.Stretch)
}
func (a *App) sidebar(c *ui.Context) {
	t := a.Theme
	sidebar := ui.Column(c.Key("sessions-sidebar")).Bind(&a.surfaceFocus.emptySidebar).Fill().Width(a.SidebarWidth).Shrink(0).MinWidth(SidebarMinWidth).LineHeight(sourceDefaultLineHeight).Background(t.Surface)
	if len(a.sidebarRows()) == 0 {
		sidebar.Focusable()
	}
	sidebar.Children(func() {
		ui.Row(c).DragWindow().Height(HeaderHeight).PaddingX(12).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Spacer(c)
			if a.iconButton(c, "Toggle sidebar", "icons/layout-left.svg").Clicked() {
				a.SidebarVisible = false
			}
			if a.iconButton(c, "Back", "icons/arrow-left.svg").Clicked() {
				a.goHistory(-1)
			}
			if a.iconButton(c, "Forward", "icons/arrow-right.svg").Clicked() {
				a.goHistory(1)
			}
		})
		ui.Column(c).Padding(2, 12, 12).AlignSelf(ui.Stretch).Center().Children(func() {
			logo := "logo.png"
			if !t.Dark {
				logo = "logo-dark.png"
			}
			if im := logos[logo]; im != nil {
				w, h := im.Size()
				ui.Image(c, im).Width(76).Height(76 * float32(h) / float32(w)).Label("Sidebar wordmark")
			} else {
				ui.Text(c, "ORBIT PI").Bold().FontSize(t.Px(16))
			}
		})
		ui.Column(c).Padding(4, 12, 8).AlignSelf(ui.Stretch).Children(func() {
			b := ui.ButtonBase(c.Key("sidebar-new-task")).Justify(ui.Start).AlignItems(ui.Center).Height(t.Px(32)).AlignSelf(ui.Stretch).PaddingX(t.Space(8)).Gap(t.Space(4)).Radius(t.Px(8)).Background(t.Elevated).Border(1, t.Border).Label(a.text("menu.new_task", "New Task"))
			b.Children(func() {
				a.icon(c, "icons/task-add-01.svg", t.Px(16), t.AccentText)
				ui.Text(c, a.text("menu.new_task", "New Task")).Grow(1).MinWidth(0).MaxLines(1).FontWeight(500).LineHeight(1)
				ui.Text(c, shortcutLabel("secondary-n")).FontSize(t.Px(12)).LineHeight(1).TextColor(t.TextMuted)
			})
			if b.Hovered() {
				b.Background(t.Hover)
			}
			if b.Clicked() {
				a.command("NewSession", "")
			}
		})
		ui.Row(c).AlignSelf(ui.Stretch).Padding(0, 14, 2).AlignItems(ui.Center).Children(func() {
			label := a.text("sidebar.projects", "Projects")
			if a.Workspaces.GroupBy == "one_list" {
				label = a.text("sidebar.all_sessions", "All sessions")
			}
			ui.Text(c, label).Grow(1).MinWidth(0).FontSize(t.Px(12)).FontWeight(500).TextColor(t.TextMuted)
			a.iconButton(c, "Project options", "icons/sort.svg").Size(t.Px(18), t.Px(18)).Menu(func(m *ui.Menu) {
				a.workspaceMenu(m)
				if m.Item("Refresh sessions").Chosen() {
					a.Dispatch(Action{Kind: "RefreshSessions"})
				}
				if m.Item("Browse workspace…").Chosen() {
					a.Dispatch(Action{Kind: "BrowseWorkspace"})
				}
				if m.Item("Worktrees").Chosen() {
					a.Navigate("worktrees")
				}
			})
		})
		rows := a.sidebarRows()
		sticky := a.prepareSidebarSticky(c, rows)
		if len(rows) == 0 {
			a.emptySidebar(c)
		}
		a.sessionList.Key = func(i int) any { return rows[i].key }
		a.sessionList.Label = func(i int) string { return rows[i].label }
		a.sessionList.Selected = &a.sessionCursor
		list := ui.List(c.Key("session-list"), &a.sessionList, len(rows), func(i int) {
			r := rows[i]
			c.Key(r.key)
			if r.group {
				a.workspaceHeader(c, r, false, sticky.index != i)
				return
			}
			if r.more > 0 || r.less {
				a.sidebarOverflowRow(c, r)
				return
			}
			s := r.session
			b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Margin(1, 8).Padding(t.Space(6), t.Space(12)).Radius(t.Px(4)).Label(s.Name)
			if a.sidebarSessionActive(s) {
				b.Background(t.Active)
			} else if b.Hovered() {
				b.Background(t.Hover)
			}
			b.Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
					ui.Text(c, s.Name).MaxLines(1).FontSize(t.FontSize)
					a.workspaceSessionSummary(c, s)
				})
				if s.Running {
					ui.Spinner(c).Size(12, 12)
				}
			})
			if b.Clicked() {
				a.activateSidebarRow(r)
			}
			b.ContextMenu(func(m *ui.Menu) { a.sessionRowMenu(m, s) })
		}).Grow(1).MinHeight(0).Bind(&a.sidebarState.sticky.viewport).Label("Sidebar session viewport").TrackScroll(&a.sidebarState.sticky.scroll).ScrollbarInsets(0, sidebarResizeGripWidth, 0, 0).Children(func() { a.renderSidebarSticky(c, rows, sticky) })
		if a.sidebarState.sticky.scroll.Y != a.sidebarState.sticky.lastY {
			a.sidebarState.sticky.lastY = a.sidebarState.sticky.scroll.Y
			c.AnimationFrame()
		}
		if len(rows) == 0 {
			list.Grow(0).Height(0)
		}
		if list.Submitted() && a.sessionCursor >= 0 && a.sessionCursor < len(rows) {
			a.activateSidebarRow(rows[a.sessionCursor])
		}
		a.line(c)
		ui.Row(c).Height(HeaderHeight).PaddingX(12).AlignItems(ui.Center).Children(func() {
			color := t.TextMuted
			status := "Disconnected"
			if a.Model.Connected {
				color = t.Success
				status = "Connected"
			}
			ui.Box(c).Size(6, 6).Radius(3).Background(color)
			ui.Text(c, status).FontSize(t.Px(12)).TextColor(t.TextMuted).MarginStart(6)
			ui.Spacer(c)
			if a.iconButton(c, "Search commands", "icons/search.svg").Clicked() {
				a.paletteOpen = true
			}
			if a.iconButton(c, "Usage", "icons/chart.svg").Clicked() {
				a.Navigate("usage")
			}
			if a.iconButton(c, "Settings", "icons/settings.svg").Clicked() {
				a.Navigate("settings")
			}
		})
	})
}

func (a *App) main(c *ui.Context) {
	main := ui.Column(c.Key("main")).Fill().MinWidth(0)
	if a.page == "chat" && !a.reviewFull() && len(a.Model.Messages) == 0 {
		// Absolute children paint above flow siblings in MyGo. Paint on the parent
		// before children so the background never overlays the composer/buttons.
		main.Draw(func(p *ui.Painter, r ui.Rect) {
			r.Y += HeaderHeight
			r.H -= HeaderHeight
			p.Clip(r, 0, func() { a.paintBackdrop(p, r, true) })
		})
	}
	main.Children(func() {
		if a.page == "chat" && !a.reviewFull() && len(a.Model.Messages) == 0 {
			ui.Box(c.Key("new-task-backdrop")).Label("New task backdrop").PassThrough().Absolute().Top(HeaderHeight).Left(0).Right(0).Bottom(0)
		}
		a.header(c)
		if a.Model.Error != "" {
			ui.Row(c).Padding(8, 16).Background(a.Theme.Error.Alpha(.1)).Children(func() {
				ui.Text(c, a.Model.Error).TextColor(a.Theme.Error).Grow(1)
				if a.iconButton(c, "Retry runtime", "icons/refresh.svg").Clicked() {
					a.Dispatch(Action{Kind: "RestartRuntime"})
				}
			})
		}
		// Usage owns the whole workspace area. Other feature pages share the
		// same content row with the Review dock, as pinned view.rs does.
		if a.page == "usage" {
			a.usageShell(c)
			return
		}
		ui.Row(c.Key("workspace-content")).Grow(1).MinHeight(0).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Children(func() {
			content := ui.Column(c.Key("chat-content")).Bind(&a.reviewState.layout.content).Grow(1).MinWidth(0).FillHeight()
			if a.reviewFull() {
				// Retain native editor/scroll state without taking layout room,
				// drawing, accessibility, pointer or keyboard input. Unmounting
				// the composer would discard MyGo's private undo/redo history.
				r := a.reviewState.layout.content.Bounds(c)
				content.Absolute().Left(0).Top(0).Width(max(r.W, 300)).Height(max(r.H, 1)).Grow(0).Invisible().Disabled(true)
			}
			content.Children(func() {
				if a.page != "chat" {
					a.renderFeaturePage(c)
					return
				}
				if a.searchOpen {
					a.findBar(c)
				}
				if len(a.Model.Messages) == 0 {
					a.newTask(c)
				} else {
					a.messages(c)
				}
				a.composerView(c)
				if a.Options.Terminal != nil && a.Options.Terminal.IsOpen() && !a.reviewFull() {
					a.terminal(c)
				}
			})
			a.trackReviewFileFocus(c)
			if a.ReviewVisible {
				a.review(c)
			} else if a.ExplorerVisible {
				a.explorer(c)
			}
		})
	})
}
func (a *App) renderFeaturePage(c *ui.Context) {
	if render := a.Options.Pages[a.page]; render != nil {
		render(c)
	} else {
		a.featurePage(c)
	}
}

func (a *App) header(c *ui.Context) {
	t := a.Theme
	// Pinned view.rs main_top_bar uses a fixed20px leading inset.
	left := float32(20)
	if !a.SidebarVisible {
		left = max(left, c.TitleBar().Left)
	}
	header := ui.Row(c.Key("header")).Height(HeaderHeight).Padding(0, max(float32(12), c.TitleBar().Right), 0, left).AlignItems(ui.Center).Gap(8)
	if a.page != "usage" {
		header.DragWindow()
	}
	header.Children(func() {
		if !a.SidebarVisible {
			if a.iconButton(c, "Show sidebar", "icons/layout-left.svg").Clicked() {
				a.SidebarVisible = true
			}
		}
		if a.page == "usage" {
			a.usageHeaderLeading(c)
		} else {
			if a.page != "chat" {
				if a.iconButton(c, "Back to task", "icons/arrow-left.svg").Clicked() {
					a.Navigate("chat")
				}
			}
			title := a.Model.SessionName
			if a.page != "chat" {
				title = strings.ToUpper(a.page[:1]) + a.page[1:]
			} else if title == "" {
				title = a.text("settings.new_task", "New task")
			}
			ui.Text(c, title).TextColor(t.TextSecondary).MaxLines(1).Grow(1)
		}
		if a.Model.Running {
			ui.Spinner(c).Size(14, 14)
			ui.Text(c, "Working").FontSize(t.Px(12)).TextColor(t.TextMuted)
		}
		a.runtimeQuotaControl(c, &a.quotaOpen, a.Model.Quotas, a.Model.QuotaSupport, a.Model.ModelProvider)
		if preferred := a.OpenPrefs.Preferred(a.Model.Workspace, a.OpenApps); preferred != nil {
			if a.button(c, preferred.Label, "icons/arrow-up-right.svg", false).Clicked() {
				a.Dispatch(Action{Kind: "OpenInApp", Value: preferred.ID})
			}
		}
		if len(a.OpenApps) > 0 {
			a.iconButton(c, "Open in…", "icons/external-link.svg").Menu(func(m *ui.Menu) {
				for _, app := range a.OpenApps {
					if m.Item(app.Label).Chosen() {
						a.Dispatch(Action{Kind: "OpenInApp", Value: app.ID})
					}
				}
				for _, entry := range []struct{ label, kind string }{{"File manager", "RevealWorkspace"}, {"Terminal", "ToggleTerminal"}, {"Files", "OpenFiles"}, {"Git", "OpenGit"}} {
					if m.Item(entry.label).Chosen() {
						a.command(entry.kind, "")
					}
				}
			})
		}
		if a.iconButton(c, "Review changes", "icons/panel-right.svg").Background(t.Elevated).Border(1, t.Border).Radius(t.Px(8)).Clicked() {
			a.command("ToggleSidePanel", "")
		}
		a.iconButton(c, "More actions", "icons/more.svg").Background(t.Elevated).Border(1, t.Border).Radius(t.Px(8)).Menu(func(m *ui.Menu) {
			for _, entry := range []struct{ label, kind string }{{"Rename task…", "RenameSession"}, {"Review changes", "ReviewChanges"}, {"Git", "ToggleGit"}, {"Worktrees", "OpenWorktrees"}, {"Usage", "ToggleUsage"}, {"Settings", "OpenSettings"}, {"Export session…", "ExportSession"}} {
				if m.Item(entry.label).Chosen() {
					a.command(entry.kind, "")
				}
			}
		})
	})
}
func (a *App) newTask(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("new-task")).Grow(1).MinHeight(0).Center().Padding(0, 24, 88).LineHeight(sourceDefaultLineHeight).Children(func() {
		width, _ := c.Size()
		if a.SidebarVisible {
			width -= a.SidebarWidth
		}
		fieldWidth := workspaceFieldWidth(width)
		ui.Column(c).Width(fieldWidth).Gap(t.Space(20)).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Size(56, 56).Center().Radius(28).Background(t.Elevated).Border(1, t.Border).Children(func() { a.icon(c, "icons/start-up.svg", 24, t.TextSecondary) })
			ui.Column(c).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.text("view.start_a_task", "Start a task")).FontSize(t.Px(20)).FontWeight(500)
				ui.Text(c, a.text("workspace.pick_workspace_hint", "Pick a workspace, then describe your task below.")).TextColor(t.TextMuted)
			})
			ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(6)).Children(func() {
				ui.Text(c, a.text("view.workspace", "Workspace")).FontSize(t.Px(10)).TextColor(t.TextMuted)
				b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Padding(7, 10, 7, t.Space(8)).Gap(10).Radius(t.Px(8)).Background(t.Elevated).Border(1, t.Border).Label("Choose workspace")
				b.Children(func() {
					ui.Box(c).Size(28, 28).Radius(t.Px(8)).Background(t.Text.Alpha(.05)).Center().Children(func() { a.icon(c, "icons/folder.svg", t.Px(14), t.TextSecondary) })
					ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
						name := filepath.Base(a.Model.Workspace)
						if a.Model.Workspace == "" {
							name = "Choose a folder"
						}
						ui.Text(c, name).FontWeight(500)
						ui.Text(c, a.Model.Workspace).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextMuted)
					})
					a.icon(c, "icons/chevron-down.svg", 14, t.TextMuted)
				})
				if b.Clicked() {
					a.toggleWorkspacePicker(c.Now())
				}
				if a.workspaceOpen {
					b.BorderColor(t.Accent.Alpha(.55))
				}
				a.workspacePickerPopup(c, b, fieldWidth)
			})
		})
	})
}
func (a *App) toasts(c *ui.Context) {
	now := time.Now()
	a.toastStack.Expire(now, a.Prefs.ReduceMotion)
	items := a.toastStack.Items()
	if len(items) == 0 {
		return
	}
	ui.Overlay(c, func() {
		ui.Column(c).Absolute().Right(16).Bottom(a.toastBottom(c)).Width(360).Gap(8).Children(func() {
			for _, n := range items {
				color := a.Theme.Text
				switch n.Kind {
				case native.ToastError:
					color = a.Theme.Error
				case native.ToastWarning:
					color = a.Theme.Warning
				case native.ToastSuccess:
					color = a.Theme.Success
				}
				ui.Row(c.Key(fmt.Sprint(n.ID))).Padding(12).Background(a.Theme.Elevated).Border(1, a.Theme.Border).Radius(8).Gap(8).Opacity(n.Opacity(now, a.Prefs.ReduceMotion)).Children(func() {
					ui.Column(c).Grow(1).MinWidth(0).Children(func() {
						if n.Title != "" {
							ui.Text(c, n.Title).FontWeight(500).TextColor(color)
						}
						ui.Text(c, n.Body).TextColor(color).MaxLines(6)
					})
					if a.iconButton(c, "Dismiss notification", "icons/x.svg").Clicked() {
						a.toastStack.Dismiss(n.ID)
					}
				})
			}
		})
	})
	c.After(40 * time.Millisecond)
}

func (a *App) Empty(c *ui.Context, title, detail string) {
	ui.Column(c).Grow(1).Center().Padding(32).Gap(10).Children(func() {
		ui.Text(c, title).FontSize(a.Theme.Px(20)).FontWeight(500)
		ui.Text(c, detail).TextColor(a.Theme.TextMuted).MaxWidth(600)
	})
}
func (a *App) terminal(c *ui.Context) {
	p := a.Options.Terminal
	t := a.Theme
	s := p.Snapshot()
	var active orbiterm.TabInfo
	hasActive := s.Active >= 0 && s.Active < len(s.Tabs)
	if hasActive {
		active = s.Tabs[s.Active]
	}
	// Rust TerminalPanel::restart and exited_overlay operate only on the active
	// tab. Snapshot state belongs to the host; the Ghostty grid does not render
	// startup failures or the shell-exited recovery affordance for us.
	ui.Column(c.Key("terminal-pane")).Height(p.Height()).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Background(t.Surface).Children(func() {
		handle := ui.Box(c).Height(6).AlignSelf(ui.Stretch).Cursor(ui.CursorResizeNS).Label("Resize terminal")
		if _, dy, ok := handle.Dragged(); ok {
			p.SetHeight(p.Height() - dy)
			a.Layout.TerminalHeight = float64(p.Height())
			a.save(a.Options.Store.SaveLayout(a.Layout))
		}
		ui.Row(c).Height(32).PaddingX(8).Gap(4).AlignItems(ui.Center).Children(func() {
			for i, tab := range s.Tabs {
				if a.button(c.Key(fmt.Sprintf("terminal-tab-%d", tab.ID)), tab.Label, tab.Icon, i == s.Active).Tooltip(tab.Tooltip).Clicked() {
					p.Select(i)
				}
			}
			ui.Spacer(c)
			if a.iconButton(c, "New terminal tab", "icons/plus.svg").Clicked() {
				a.save(p.NewTab())
			}
			if hasActive && active.Exited {
				ui.Text(c, a.text("terminal.exited", "exited")).FontSize(t.Px(11)).TextColor(t.Error).Padding(1, 6).Background(t.Error.Alpha(.15)).Radius(t.Px(4))
			}
			restart := a.iconButton(c, "Restart terminal", "icons/refresh.svg").Tooltip(a.text("terminal.restart", "Restart")).Disabled(!hasActive || active.Loading)
			if s.Restarting {
				restart.Background(t.Hover)
			}
			if restart.Clicked() {
				a.save(p.Restart())
			}
			if a.iconButton(c, "Close terminal tab", "icons/x.svg").Clicked() {
				a.save(p.CloseActive())
			}
			if a.iconButton(c, "Hide terminal", "icons/chevron-down.svg").Clicked() {
				p.Hide()
			}
		})
		ui.Box(c.Key("terminal-content")).Grow(1).MinHeight(0).Children(func() {
			orbiterm.View(c, p).Bind(&a.surfaceFocus.terminal).Fill()
			if hasActive && (active.Exited || active.Error != "") {
				ui.Box(c.Key("terminal-exited-overlay")).Absolute().Top(0).Bottom(0).Left(0).Right(0).Center().Children(func() {
					ui.Column(c).AlignItems(ui.Center).MaxWidth(540).Padding(t.Space(12), t.Space(16)).Gap(t.Space(8)).Radius(t.Px(12)).Background(t.Elevated).Border(1, t.Border).Children(func() {
						title := a.text("terminal.shell_exited", "Shell exited")
						if active.Error != "" {
							title = a.text("terminal.start_failed", "Terminal could not start")
						}
						ui.Text(c, title).FontSize(t.Px(13)).TextColor(t.TextSecondary)
						if active.Error != "" {
							ui.Text(c, active.Error).FontSize(t.Px(12)).TextColor(t.Error).MaxLines(4)
						}
						restart := a.button(c, a.text("terminal.restart", "Restart"), "icons/refresh.svg", true).Bind(&a.surfaceFocus.terminalRestart).Label("Restart shell").Disabled(active.Loading)
						if active.Exited && a.surfaceFocus.terminal.FocusWithin(c) {
							restart.Focus()
						}
						if restart.Clicked() {
							a.save(p.Restart())
						}
					})
				})
			}
		})
	})
	if hasActive && (active.Loading || s.Restarting) {
		c.After(40 * time.Millisecond)
	}
}
func (a *App) String() string { return fmt.Sprintf("Orbit native %s", a.page) }

// Async keeps filesystem/process operations outside the native view thread.
func (a *App) Async(job func() func()) {
	go func() {
		apply := job()
		if apply == nil {
			return
		}
		if w := a.window.Load(); w != nil && !w.IsDestroyed() {
			w.Update(apply)
		} else {
			a.updates <- apply
		}
	}()
}
func (a *App) Service() *workbench.Service {
	s := workbench.New(a.Options.Store, a.Model.Workspace)
	s.PiBinary = a.Options.PiBinary
	return s
}

func (a *App) workspaceMenu(m *ui.Menu) {
	for _, sortMode := range []string{"last_updated", "alphabetical", "alphabetical_desc", "date_added", "session_count", "manual"} {
		if m.Item("Sort: " + strings.ReplaceAll(sortMode, "_", " ")).Checked(a.Workspaces.Sort == sortMode).Chosen() {
			a.Workspaces.Sort = sortMode
			a.save(a.Options.Store.SaveWorkspaces(a.Workspaces))
		}
	}
	for _, filter := range []string{"hide", "show", "only"} {
		if m.Item("Archived: " + filter).Checked(a.Workspaces.ArchivedFilter == filter).Chosen() {
			a.Workspaces.ArchivedFilter = filter
			a.save(a.Options.Store.SaveWorkspaces(a.Workspaces))
		}
	}
	if m.Item("Group by workspace").Checked(a.Workspaces.GroupBy == "workspace").Chosen() {
		if a.Workspaces.GroupBy == "workspace" {
			a.Workspaces.GroupBy = "one_list"
		} else {
			a.Workspaces.GroupBy = "workspace"
		}
		a.save(a.Options.Store.SaveWorkspaces(a.Workspaces))
	}
}

func (a *App) toastBottom(c *ui.Context) float32 {
	if a.Options.ToastBottom != nil {
		return a.Options.ToastBottom(c)
	}
	return 16
}
