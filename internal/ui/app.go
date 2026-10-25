// Package orbitui renders Orbit entirely with MyGo's native Go toolkit.
package orbitui

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/native"
	"github.com/charlzyx/orbit/internal/preferences"
	orbiterm "github.com/charlzyx/orbit/internal/terminal"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo"
	ghostty "github.com/egoist/mygo/plugins/terminal"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

type SessionRow struct {
	ID, Path, Name, Workspace, Summary string
	Updated                            time.Time
	Running, Pinned, Archived          bool
}
type Message struct {
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
	Input, Output, CacheRead, CacheWrite             int64
	Cost                                             float64
	ContextPercent                                   float64
	Available                                        bool
	TokensAvailable, CostAvailable, ContextAvailable bool
	ContextTokens, ContextWindow                     int64
}
type Snapshot struct {
	Connected, Running                                                                                      bool
	Turn                                                                                                    int
	Error, Status, SessionID, SessionPath, SessionName, Workspace, ModelID, ModelProvider, Thinking, Branch string
	Messages                                                                                                []Message
	Sessions                                                                                                []SessionRow
	Models                                                                                                  []ModelChoice
	ThinkingLevels, Commands                                                                                []string
	Dialog                                                                                                  *DialogRequest
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
	BeginRuntimeOperation func() (func(), error)
	Version               string
	Custom                *orbiterm.CustomPanel
	Updater               *native.Updater
	Store                 *preferences.Store
	OnAction              func(Action)
	Terminal              *orbiterm.Panel
	// Optional workbench modules mount their actual native views here.
	Pages     map[string]func(*ui.Context)
	OnPage    func(string)
	WorkIn    func(*ui.Context)
	BugReport func(*ui.Context)
	Escape    func() bool
	Workspace string
	PiBinary  string
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
	Options                                                         Options
	Window                                                          *mygo.Window
	window                                                          atomic.Pointer[mygo.Window]
	Model                                                           Snapshot
	Prefs                                                           preferences.UI
	Workspaces                                                      preferences.WorkspaceList
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
	composer                                                        ui.Handle
	sessionList, transcript, paletteList                            ui.ListState
	sessionCursor                                                   int
	expanded                                                        map[string]bool
	collapsedMessages                                               map[string]bool
	md                                                              markdown.Options
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
	providers                                                       []workbench.Provider
	credentials                                                     map[string]workbench.CredentialStatus
	retention                                                       preferences.Retention
	retentionBusy                                                   bool
	quotaOpen                                                       bool
}

func New(o Options) *App {
	if o.Store == nil {
		o.Store = preferences.New("")
	}
	a := &App{Options: o, Prefs: o.Store.UI(), Workspaces: o.Store.Workspaces(), Fonts: o.Store.Fonts(), Appearance: o.Store.Appearance(), Layout: o.Store.Layout(), SidebarVisible: true, settingsPage: "general", page: "chat", sessionCursor: -1, expanded: map[string]bool{}, collapsedMessages: map[string]bool{}, history: []string{"chat"}, updates: make(chan func(), 64), drafts: map[string]draftState{}, messageImages: map[string]ui.ImageSource{}, imageLoading: map[string]bool{}, imageErrors: map[string]string{}, selectedSkill: -1}
	a.SidebarWidth = float32(a.Layout.SidebarWidth)
	if a.SidebarWidth < SidebarMinWidth {
		a.SidebarWidth = 248
	}
	a.ExplorerWidth = float32(a.Layout.ExplorerWidth)
	a.ReviewWidth = float32(a.Layout.SidepaneWidth)
	a.access = o.Store.Access()
	a.workflow = o.Store.Workflow("")
	a.Model.Workspace = o.Workspace
	a.retention = o.Store.Retention()
	a.english, _ = orbitassets.Locale("en")
	a.reloadLocale()
	a.Theme = NewTheme(a.Appearance.DarkTheme, a.Prefs, a.Fonts)
	a.transcript.FollowEnd = true
	a.transcript.Key = func(i int) any { return a.Model.Messages[i].ID }
	a.sessionList.Key = func(i int) any { return a.Model.Sessions[i].ID }
	a.md.Block = a.markdownBlock
	a.md.LoadImage = markdown.FileImages(o.Workspace)
	a.md.Reveal = &markdown.Reveal{Unit: markdown.AtOnce, Fade: -1}
	a.md.OnLink = func(s string) { a.Dispatch(Action{Kind: "OpenLink", Value: s}) }
	return a
}
func (a *App) reloadLocale() {
	lang := a.Prefs.Language
	if lang == "system" || lang == "" {
		lang = "en"
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
	if s.SessionID != a.Model.SessionID {
		a.messageImages = map[string]ui.ImageSource{}
		a.imageLoading = map[string]bool{}
		a.imageErrors = map[string]string{}
		a.imageOpen = false
		a.transcript = ui.ListState{FollowEnd: true}
		a.transcript.Key = func(i int) any { return a.Model.Messages[i].ID }
		a.workflow = a.Options.Store.Workflow(s.SessionID)
		if a.pendingNew {
			a.pendingNew = false
		} else if a.Model.SessionID != "" {
			a.drafts[a.Model.SessionID] = draftState{a.Draft, append([]attachment(nil), a.attachments...)}
			d := a.drafts[s.SessionID]
			a.Draft = d.Text
			a.attachments = d.Attachments
		}
	}
	a.Model = s
	if s.Dialog != nil && s.Dialog.ID != a.dialogID {
		a.dialogID = s.Dialog.ID
		a.dialogValue = s.Dialog.Default
	}
}
func (a *App) Dispatch(action Action) {
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
		ed := a.files.editor
		a.Async(func() func() { err := ed.SaveAll(); return func() { a.save(err) } })
	}
}
func (a *App) View(c *ui.Context) {
	for {
		select {
		case apply := <-a.updates:
			apply()
		default:
			goto drained
		}
	}
drained:
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
	appearanceKey := fmt.Sprintf("%s|%s|%g|%t", id, a.Fonts.CodeFamily, a.Prefs.TerminalFontSize, a.Prefs.ReduceMotion)
	if a.Options.Custom != nil && a.customAppearanceKey != appearanceKey {
		a.Options.Custom.SetAppearance(ghostty.Font{Family: a.Fonts.CodeFamily, Size: float32(a.Prefs.TerminalFontSize), Features: []string{"-calt", "-liga", "-dlig"}}, a.Theme.Terminal(), a.Theme.Terminal(), a.Prefs.ReduceMotion)
		a.customAppearanceKey = appearanceKey
	}
	c.SetTheme(a.Theme.Native())
	a.md.Theme = a.Theme.Markdown()
	a.md.Theme.CodeColors = a.Theme.Code()
	if a.Prefs.ReduceMotion {
		c.DisableAnimations()
	}
	root := ui.Row(c.Key("root")).Fill().Background(a.Theme.Background).Font(a.Fonts.UIFamily).TextColor(a.Theme.Text).Children(func() {
		if a.page == "settings" {
			a.settings(c)
			return
		}
		if a.SidebarVisible {
			if ui.Split(c, &a.SidebarWidth, func() { a.sidebar(c) }, func() { a.main(c) }).Fill().Changed() {
				a.SidebarWidth = max(SidebarMinWidth, a.SidebarWidth)
				a.Layout.SidebarWidth = float64(a.SidebarWidth)
				a.save(a.Options.Store.SaveLayout(a.Layout))
			}
		} else {
			a.main(c)
		}
	})
	if paths := root.DroppedFiles(); len(paths) > 0 && a.page != "settings" {
		a.AddPaths(paths)
	}
	a.shortcuts(c)
	if a.paletteOpen {
		a.commandPalette(c)
	}
	if a.modelOpen {
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
	ui.Column(c.Key("sessions-sidebar")).Fill().MinWidth(SidebarMinWidth).Background(t.Surface).Children(func() {
		ui.Row(c).DragWindow().Height(HeaderHeight).PaddingX(12).AlignItems(ui.Center).Children(func() {
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
		ui.Column(c).Height(38).Center().Children(func() {
			logo := "logo.png"
			if !t.Dark {
				logo = "logo-dark.png"
			}
			if im := logos[logo]; im != nil {
				ui.Image(c, im).Width(104).Height(25)
			} else {
				ui.Text(c, "ORBIT PI").Bold().FontSize(t.Px(21))
			}
		})
		ui.Column(c).Padding(t.Space(12)).Gap(t.Space(8)).Children(func() {
			b := a.button(c, a.text("sidebar.new_session", "New Task"), "icons/compose.svg", false).Height(t.Px(34)).AlignSelf(ui.Stretch).Background(t.Elevated).Border(1, t.Border)
			if b.Clicked() {
				a.command("NewSession", "")
			}
			ui.Row(c).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.text("sidebar.projects", "Projects")).FontSize(t.Px(12)).TextColor(t.TextMuted)
				ui.Spacer(c)
				a.iconButton(c, "Project options", "icons/more.svg").Menu(func(m *ui.Menu) {
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
		})
		rows := a.sidebarRows()
		a.sessionList.Key = func(i int) any { return rows[i].key }
		a.sessionList.Label = func(i int) string { return rows[i].label }
		a.sessionList.Selected = &a.sessionCursor
		list := ui.List(c.Key("session-list"), &a.sessionList, len(rows), func(i int) {
			r := rows[i]
			c.Key(r.key)
			if r.group {
				b := a.button(c, r.label, "icons/folder.svg", false).AlignSelf(ui.Stretch).MarginX(8)
				if b.Clicked() {
					a.expanded[r.path] = !a.expanded[r.path]
				}
				b.ContextMenu(func(m *ui.Menu) {
					if m.Item("New task here").Chosen() {
						a.Dispatch(Action{Kind: "SetWorkspace", Value: r.path})
						a.Navigate("chat")
					}
					if m.Item("Open in file manager").Chosen() {
						a.Dispatch(Action{Kind: "RevealPath", Value: r.path})
					}
					if m.Item("Worktrees").Chosen() {
						a.Navigate("worktrees")
					}
				})
				return
			}
			s := r.session
			b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Margin(1, 8).Padding(t.Space(6), t.Space(12)).Radius(t.Px(4)).Label(s.Name)
			if s.ID == a.Model.SessionID {
				b.Background(t.Active)
			} else if b.Hovered() {
				b.Background(t.Hover)
			}
			b.Children(func() {
				ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
					ui.Text(c, s.Name).MaxLines(1).FontSize(t.FontSize)
					ui.Text(c, s.Summary).MaxLines(1).FontSize(t.Px(12)).TextColor(t.TextMuted)
				})
				if s.Running {
					ui.Spinner(c).Size(12, 12)
				}
			})
			if b.Clicked() {
				a.Dispatch(Action{Kind: "SwitchSession", Value: s.Path})
				a.Navigate("chat")
			}
			b.ContextMenu(func(m *ui.Menu) {
				for _, cmd := range []struct{ id, label string }{{"RenameSession", "Rename…"}, {"PinSession", "Pin / Unpin"}, {"CloneSession", "Clone"}, {"ArchiveSession", "Archive"}, {"CopySessionId", "Copy session ID"}, {"DeleteSession", "Delete…"}} {
					if m.Item(cmd.label).Chosen() {
						a.commandSession(cmd.id, s)
					}
				}
			})
		}).Grow(1).MinHeight(0)
		if list.Submitted() && a.sessionCursor >= 0 && a.sessionCursor < len(rows) {
			r := rows[a.sessionCursor]
			if r.group {
				a.expanded[r.path] = !a.expanded[r.path]
			} else {
				a.Dispatch(Action{Kind: "SwitchSession", Value: r.session.Path})
				a.Navigate("chat")
			}
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

type sidebarRow struct {
	key, label, path string
	group            bool
	session          SessionRow
}

func (a *App) sidebarRows() []sidebarRow {
	groups := map[string][]SessionRow{}
	for _, s := range a.Model.Sessions {
		if (a.Workspaces.ArchivedFilter == "hide" && s.Archived) || (a.Workspaces.ArchivedFilter == "only" && !s.Archived) {
			continue
		}
		key := s.Workspace
		if a.Workspaces.GroupBy == "one_list" {
			key = ""
		}
		groups[key] = append(groups[key], s)
	}
	if a.Workspaces.GroupBy != "one_list" && a.Model.Workspace != "" && groups[a.Model.Workspace] == nil {
		groups[a.Model.Workspace] = []SessionRow{}
	}
	for _, w := range a.Workspaces.Workspaces {
		if a.Workspaces.GroupBy != "one_list" && groups[w.Path] == nil {
			groups[w.Path] = []SessionRow{}
		}
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		left, right := keys[i], keys[j]
		switch a.Workspaces.Sort {
		case "alphabetical":
			return strings.ToLower(filepath.Base(left)) < strings.ToLower(filepath.Base(right))
		case "alphabetical_desc":
			return strings.ToLower(filepath.Base(left)) > strings.ToLower(filepath.Base(right))
		case "session_count":
			return len(groups[left]) > len(groups[right])
		case "last_updated":
			var a, b time.Time
			for _, s := range groups[left] {
				if s.Updated.After(a) {
					a = s.Updated
				}
			}
			for _, s := range groups[right] {
				if s.Updated.After(b) {
					b = s.Updated
				}
			}
			return a.After(b)
		case "manual", "date_added":
			li, rj := len(a.Workspaces.Workspaces), len(a.Workspaces.Workspaces)
			for k, w := range a.Workspaces.Workspaces {
				if w.Path == left {
					li = k
				}
				if w.Path == right {
					rj = k
				}
			}
			if a.Workspaces.Sort == "date_added" {
				return li > rj
			}
			return li < rj
		}
		return left < right
	})
	var rows []sidebarRow
	for _, k := range keys {
		label := filepath.Base(k)
		if k == "" {
			label = "Other sessions"
		}
		rows = append(rows, sidebarRow{key: "workspace:" + k, label: label, path: k, group: true})
		if a.expanded[k] {
			continue
		}
		sort.SliceStable(groups[k], func(i, j int) bool {
			l, r := groups[k][i], groups[k][j]
			if l.Pinned != r.Pinned {
				return l.Pinned
			}
			return l.Updated.After(r.Updated)
		})
		for _, s := range groups[k] {
			if s.Name == "" {
				s.Name = "Untitled task"
			}
			rows = append(rows, sidebarRow{key: "session:" + s.ID, session: s})
		}
	}
	return rows
}
func (a *App) main(c *ui.Context) {
	ui.Column(c.Key("main")).Fill().MinWidth(0).Children(func() {
		a.header(c)
		if a.Model.Error != "" {
			ui.Row(c).Padding(8, 16).Background(a.Theme.Error.Alpha(.1)).Children(func() {
				ui.Text(c, a.Model.Error).TextColor(a.Theme.Error).Grow(1)
				if a.iconButton(c, "Retry runtime", "icons/refresh.svg").Clicked() {
					a.Dispatch(Action{Kind: "RestartRuntime"})
				}
			})
		}
		if a.page != "chat" {
			if render := a.Options.Pages[a.page]; render != nil {
				render(c)
			} else {
				a.featurePage(c)
			}
			return
		}
		ui.Row(c).Grow(1).MinHeight(0).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).FillHeight().Children(func() {
				if a.searchOpen {
					a.findBar(c)
				}
				if len(a.Model.Messages) == 0 {
					a.newTask(c)
				} else {
					a.messages(c)
				}
				a.composerView(c)
				if a.Options.Terminal != nil && a.Options.Terminal.IsOpen() {
					a.terminal(c)
				}
			})
			if a.ReviewVisible {
				a.review(c)
			} else if a.ExplorerVisible {
				a.explorer(c)
			}
		})
	})
}
func (a *App) header(c *ui.Context) {
	t := a.Theme
	ui.Row(c.Key("header")).DragWindow().Height(HeaderHeight).Padding(0, max(t.Space(16), c.TitleBar().Right), 0, t.Space(16)).AlignItems(ui.Center).Gap(t.Space(6)).Children(func() {
		if !a.SidebarVisible {
			if a.iconButton(c, "Show sidebar", "icons/layout-left.svg").Clicked() {
				a.SidebarVisible = true
			}
		}
		title := a.Model.SessionName
		if a.page != "chat" {
			title = strings.ToUpper(a.page[:1]) + a.page[1:]
		} else if title == "" {
			title = a.text("view.new_task", "New task")
		}
		ui.Text(c, title).TextColor(t.TextSecondary).MaxLines(1).Grow(1)
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
		if a.iconButton(c, "Project files", "icons/panel-right.svg").Clicked() {
			a.ExplorerVisible = !a.ExplorerVisible
			a.ReviewVisible = false
			a.Dispatch(Action{Kind: "RefreshFiles"})
		}
		a.iconButton(c, "More actions", "icons/more.svg").Menu(func(m *ui.Menu) {
			for _, entry := range []struct{ label, kind string }{{"Rename task…", "RenameSession"}, {"Review changes", "ReviewChanges"}, {"Git", "OpenGit"}, {"Worktrees", "OpenWorktrees"}, {"Usage", "ToggleUsage"}, {"Settings", "OpenSettings"}, {"Export session…", "ExportSession"}} {
				if m.Item(entry.label).Chosen() {
					a.command(entry.kind, "")
				}
			}
		})
	})
}
func (a *App) newTask(c *ui.Context) {
	t := a.Theme
	ui.Column(c.Key("new-task")).Grow(1).MinHeight(0).Center().Padding(24, 24, 88).Children(func() {
		ui.Column(c).Width(400).MaxWidth(500).Gap(t.Space(20)).AlignItems(ui.Center).Children(func() {
			ui.Column(c).Size(56, 56).Center().Radius(28).Background(t.Elevated).Border(1, t.Border).Children(func() { a.icon(c, "icons/start-up.svg", 24, t.TextSecondary) })
			ui.Column(c).Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
				ui.Text(c, a.text("view.start_a_task", "Start a task")).FontSize(t.Px(20)).FontWeight(500)
				ui.Text(c, a.text("workspace.pick_workspace_hint", "Pick a workspace, then describe your task below.")).TextColor(t.TextMuted)
			})
			ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(6)).Children(func() {
				ui.Text(c, a.text("view.workspace", "Workspace")).FontSize(t.Px(10)).TextColor(t.TextMuted)
				b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Padding(7, 8).Gap(10).Radius(t.Px(8)).Background(t.Elevated).Border(1, t.Border).Label("Choose workspace")
				b.Children(func() {
					a.icon(c, "icons/folder.svg", 20, t.TextSecondary)
					ui.Column(c).Grow(1).MinWidth(0).Gap(2).Children(func() {
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
					a.Dispatch(Action{Kind: "BrowseWorkspace"})
				}
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
		ui.Column(c).Absolute().Right(16).Bottom(16).Width(360).Gap(8).Children(func() {
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
	ui.Column(c.Key("terminal-pane")).Height(p.Height()).BorderWidth(1, 0, 0, 0).BorderColor(t.Border).Background(t.Surface).Children(func() {
		handle := ui.Box(c).Height(6).AlignSelf(ui.Stretch).Cursor(ui.CursorResizeNS).Label("Resize terminal")
		if _, dy, ok := handle.Dragged(); ok {
			p.SetHeight(p.Height() - dy)
			a.Layout.TerminalHeight = float64(p.Height())
			a.save(a.Options.Store.SaveLayout(a.Layout))
		}
		ui.Row(c).Height(32).PaddingX(8).Gap(4).AlignItems(ui.Center).Children(func() {
			for i, tab := range s.Tabs {
				if a.button(c, tab.Label, tab.Icon, i == s.Active).Clicked() {
					p.Select(i)
				}
			}
			ui.Spacer(c)
			if a.iconButton(c, "New terminal tab", "icons/plus.svg").Clicked() {
				a.save(p.NewTab())
			}
			if a.iconButton(c, "Close terminal tab", "icons/x.svg").Clicked() {
				a.save(p.CloseActive())
			}
			if a.iconButton(c, "Hide terminal", "icons/chevron-down.svg").Clicked() {
				p.Hide()
			}
		})
		orbiterm.View(c, p).Grow(1).MinHeight(0)
	})
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
