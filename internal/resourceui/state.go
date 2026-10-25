// Package resourceui implements Orbit's native Skills and Plugins settings.
// Only Tick applies worker results; rendering never performs file or process I/O.
package resourceui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

type Theme struct {
	Elevated, SurfaceSecondary                   ui.Color
	Secondary, Action, ActionText, SelectionText ui.Color
	CodeFont                                     string
	Px                                           func(float32) float32
	Space                                        func(int) float32
}

type Options struct {
	// OnChange is a goroutine-safe invalidation request. All other callbacks
	// run on the UI thread. OnModified requests a safe, idle-only Pi reload.
	OnChange, OnModified           func()
	Text                           func(key, fallback string) string
	Theme                          func() Theme
	PiBinary                       string
	Offline                        bool
	OpenFile, RevealPath, CopyText func(string)
	InvokeSkill                    func(command string)
	Notify                         func(text string, isError bool)
}

type snapshot struct {
	skills                    []workbench.Skill
	plugins                   []workbench.Plugin
	skips                     map[string]string
	skillErrors, pluginErrors []string
}

type completion struct {
	generation uint64
	apply      func(*State)
	err        error
}

type State struct {
	Options                                                         Options
	Workspace                                                       string
	Skills                                                          []workbench.Skill
	Plugins                                                         []workbench.Plugin
	SkillErrors, PluginErrors, UpdateErrors                         []string
	Updates                                                         []workbench.PluginUpdate
	Skipped                                                         map[string]string
	SkillQuery, PluginQuery, Selected, InstallSource, InstallScope  string
	Error, Status                                                   string
	Busy, Checking                                                  bool
	CheckedAt                                                       time.Time
	launchCheckDue                                                  time.Time
	service                                                         *workbench.Service
	results                                                         chan completion
	retiring                                                        map[uint64]bool
	cancel                                                          context.CancelFunc
	generation                                                      uint64
	closed, loaded, refreshPending, checkPending                    bool
	activity, operation, operationTarget, skillDelete, pluginRemove string
	page                                                            string
	focusSkill                                                      string
	md                                                              markdown.Options
	listScroll, detailScroll, pluginScroll                          ui.ScrollState
}

func New(p *preferences.Store, workspace string, o Options) *State {
	s := &State{Options: o, Workspace: workspace, service: workbench.New(p, workspace), results: make(chan completion, 16), retiring: map[uint64]bool{}, InstallScope: "global", Skipped: map[string]string{}}
	s.launchCheckDue = time.Now().Add(5 * time.Second)
	if o.PiBinary != "" {
		s.service.PiBinary = o.PiBinary
	}
	return s
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

func (s *State) Close() { s.invalidate(); s.closed = true }
func (s *State) invalidate() {
	if s.Busy && s.cancel != nil {
		s.retiring[s.generation] = true
	}
	s.generation++
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	// Navigation invalidates result ownership, not process ownership. Keep the
	// duplicate-operation gate until the canceled worker is actually reaped.
	s.Busy = len(s.retiring) > 0
	if s.Busy {
		s.activity = s.label("resources.cancelling_previous", "Cancelling the previous operation…")
	}
	s.Checking = false
	s.refreshPending = false
	s.checkPending = false
	s.skillDelete = ""
	s.pluginRemove = ""
}
func (s *State) SetWorkspace(workspace string) {
	if workspace == s.Workspace || s.closed {
		return
	}
	s.invalidate()
	s.Workspace = workspace
	if workspace == "" {
		s.InstallScope = "global"
	}
	s.service = workbench.New(s.service.Preferences, workspace)
	if s.Options.PiBinary != "" {
		s.service.PiBinary = s.Options.PiBinary
	}
	s.Skills = nil
	s.Plugins = nil
	s.Updates = nil
	s.Skipped = map[string]string{}
	s.SkillErrors = nil
	s.PluginErrors = nil
	s.UpdateErrors = nil
	s.Error = ""
	s.Status = ""
	s.Selected = ""
	s.SkillQuery = ""
	s.PluginQuery = ""
	s.loaded = false
	s.CheckedAt = time.Time{}
	s.listScroll = ui.ScrollState{}
	s.detailScroll = ui.ScrollState{}
	s.pluginScroll = ui.ScrollState{}
}

func (s *State) SetPiBinary(binary string) {
	if binary == "" {
		binary = "pi"
	}
	if binary == s.service.PiBinary || s.closed {
		return
	}
	// An in-flight command owns its service and binary until it is reaped.
	s.Options.PiBinary = binary
	service := workbench.New(s.service.Preferences, s.Workspace)
	service.PiBinary = binary
	s.service = service
}

func (s *State) HasModal() bool { return s.skillDelete != "" || s.pluginRemove != "" }
func (s *State) Escape() bool {
	if !s.HasModal() {
		return false
	}
	s.skillDelete = ""
	s.pluginRemove = ""
	return true
}

func (s *State) start(activity, operation, target string, job func(context.Context) (func(*State), error)) bool {
	if s.Busy || s.closed {
		return false
	}
	s.Busy = true
	s.activity = activity
	s.operation = operation
	s.operationTarget = target
	s.Error = ""
	s.Status = ""
	s.generation++
	generation := s.generation
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	channel, wake := s.results, s.Options.OnChange
	go func() {
		defer cancel()
		apply, err := job(ctx)
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
	return true
}

// Cancel waits for the worker to finish/reap before enabling actions again.
// A canceled package manager may have changed files, so mutations re-read disk.
func (s *State) Cancel() {
	if s.cancel != nil && s.Busy {
		s.activity = s.label("resources.cancelling", "Cancelling…")
		s.cancel()
	}
}

func (s *State) Tick() {
	if s.closed {
		return
	}
	for {
		select {
		case r := <-s.results:
			if s.closed || r.generation != s.generation {
				if s.retiring[r.generation] {
					delete(s.retiring, r.generation)
					if len(s.retiring) == 0 && s.cancel == nil {
						s.Busy = false
						s.activity = ""
						s.operation = ""
						s.operationTarget = ""
					}
				}
				continue
			}
			s.Busy = false
			s.Checking = false
			s.cancel = nil
			if r.apply != nil {
				r.apply(s)
			}
			if r.err != nil {
				if errors.Is(r.err, context.Canceled) {
					s.Status = s.label("resources.cancelled", "Cancelled. The list reflects the files currently on disk.")
				} else {
					s.Error = r.err.Error()
				}
			}
			if s.operation == "install" || s.operation == "update" || s.operation == "remove" || s.operation == "update-all" || strings.HasPrefix(s.operation, "skill-") {
				if s.Status != "" {
					s.notify(s.Status, false)
				}
				if s.Error != "" {
					s.notify(s.Error, true)
				}
			}
			s.operation = ""
			s.operationTarget = ""
			s.activity = ""
		default:
			if !s.launchCheckDue.IsZero() && time.Now().After(s.launchCheckDue) && !s.Busy {
				s.launchCheckDue = time.Time{}
				if !s.Options.Offline && os.Getenv("PI_OFFLINE") == "" {
					s.checkPending = true
				}
			}
			if s.refreshPending && !s.Busy && !s.closed {
				s.refreshPending = false
				s.Refresh()
			} else if s.checkPending && !s.Busy && !s.closed {
				s.checkPending = false
				s.CheckUpdates()
			}
			return
		}
	}
}

func readSnapshot(service *workbench.Service) snapshot {
	var result snapshot
	var errs []error
	result.skills, errs = service.Skills()
	result.skillErrors = errorStrings(errs)
	result.plugins, errs = service.Plugins()
	result.pluginErrors = errorStrings(errs)
	var err error
	result.skips, err = service.SkippedPluginUpdates()
	if err != nil {
		result.pluginErrors = append(result.pluginErrors, err.Error())
	}
	if result.skips == nil {
		result.skips = map[string]string{}
	}
	return result
}
func errorStrings(errs []error) []string {
	var result []string
	for _, err := range errs {
		if err != nil {
			result = append(result, err.Error())
		}
	}
	return result
}
func (s *State) applySnapshot(result snapshot) {
	s.Skills = result.skills
	s.Plugins = result.plugins
	s.Skipped = result.skips
	s.SkillErrors = result.skillErrors
	s.PluginErrors = result.pluginErrors
	s.loaded = true
	if s.skill(s.Selected) == nil {
		s.Selected = ""
		if len(s.Skills) > 0 {
			s.Selected = s.Skills[0].Path
		}
		s.skillDelete = ""
		s.detailScroll = ui.ScrollState{}
	}
	if s.plugin(s.pluginRemove) == nil {
		s.pluginRemove = ""
	}
}

func (s *State) RefreshPlugins() { s.checkPending = true; s.Refresh() }

func (s *State) Refresh() {
	if s.Busy {
		s.refreshPending = true
		return
	}
	service := s.service
	s.start(s.label("resources.loading", "Loading…"), "refresh", "", func(ctx context.Context) (func(*State), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result := readSnapshot(service)
		return func(s *State) { s.applySnapshot(result) }, ctx.Err()
	})
}

func (s *State) skill(path string) *workbench.Skill {
	for i := range s.Skills {
		if s.Skills[i].Path == path {
			return &s.Skills[i]
		}
	}
	return nil
}
func pluginKey(p workbench.Plugin) string       { return p.Scope + "\x00" + p.Source }
func updateKey(p workbench.PluginUpdate) string { return p.Scope + "\x00" + p.Source }
func (s *State) plugin(key string) *workbench.Plugin {
	for i := range s.Plugins {
		if pluginKey(s.Plugins[i]) == key {
			return &s.Plugins[i]
		}
	}
	return nil
}
func (s *State) visibleSkills() []workbench.Skill {
	q := strings.ToLower(strings.TrimSpace(s.SkillQuery))
	var result []workbench.Skill
	for _, scope := range []string{"project", "global", "user"} {
		for _, p := range s.Skills {
			if p.Scope == scope && (q == "" || strings.Contains(strings.ToLower(p.Name+" "+p.Description), q)) {
				result = append(result, p)
			}
		}
	}
	return result
}
func (s *State) visiblePlugins() []workbench.Plugin {
	q := strings.ToLower(strings.TrimSpace(s.PluginQuery))
	var result []workbench.Plugin
	for _, p := range s.Plugins {
		if q == "" || strings.Contains(strings.ToLower(p.Source+" "+p.Name+" "+p.Version), q) {
			result = append(result, p)
		}
	}
	return result
}
func (s *State) SelectSkill(path string) {
	if s.skill(path) == nil || s.Selected == path {
		return
	}
	s.Selected = path
	s.skillDelete = ""
	s.detailScroll = ui.ScrollState{}
}
func (s *State) moveSkill(delta int) {
	list := s.visibleSkills()
	if len(list) == 0 {
		return
	}
	i := -1
	for n, p := range list {
		if p.Path == s.Selected {
			i = n
			break
		}
	}
	if i < 0 {
		if delta < 0 {
			i = len(list) - 1
		} else {
			i = 0
		}
	} else {
		i = max(0, min(len(list)-1, i+delta))
	}
	s.SelectSkill(list[i].Path)
	s.focusSkill = list[i].Path
}
func (s *State) modified() {
	if s.Options.OnModified != nil {
		s.Options.OnModified()
	}
}
func (s *State) notify(text string, isError bool) {
	if s.Options.Notify != nil {
		s.Options.Notify(text, isError)
	}
}

func (s *State) SetSkillEnabled(path string, enabled bool) {
	p := s.skill(path)
	if p == nil || s.Busy || p.Enabled == enabled {
		return
	}
	skill, service := *p, s.service
	s.start(s.label("skills_ui.saving", "Saving skill…"), "skill-toggle", path, func(ctx context.Context) (func(*State), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		err := service.SetSkillEnabled(skill, enabled)
		result := readSnapshot(service)
		return func(s *State) {
			s.applySnapshot(result)
			if err == nil {
				if enabled {
					s.Status = s.label("skills_ui.enabled_skill", "Enabled %{name}", "name", skill.Name)
				} else {
					s.Status = s.label("skills_ui.disabled_skill", "Disabled %{name}", "name", skill.Name)
				}
				s.modified()
			}
		}, err
	})
}

func (s *State) AskDeleteSkill(path string) {
	if s.skill(path) != nil && !s.Busy {
		s.skillDelete = path
	}
}
func (s *State) ConfirmDeleteSkill() {
	path := s.skillDelete
	if path == "" || s.skill(path) == nil || s.Busy {
		return
	}
	service := s.service
	name := s.skill(path).Name
	s.start(s.label("skills_ui.deleting", "Deleting skill…"), "skill-delete", path, func(ctx context.Context) (func(*State), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		err := service.DeleteSkill(path)
		result := readSnapshot(service)
		return func(s *State) {
			s.skillDelete = ""
			s.applySnapshot(result)
			s.modified()
			if err == nil {
				s.Status = s.label("skills_ui.deleted_skill", "Deleted skill %{name}", "name", name)
			}
		}, err
	})
}

func (s *State) Install() { s.runPlugin("install", strings.TrimSpace(s.InstallSource), s.InstallScope) }
func (s *State) AskRemovePlugin(key string) {
	if s.plugin(key) != nil && !s.Busy {
		s.pluginRemove = key
	}
}
func (s *State) ConfirmRemovePlugin() {
	if p := s.plugin(s.pluginRemove); p != nil {
		s.runPlugin("remove", p.Source, p.Scope)
	}
}
func (s *State) UpdatePlugin(key string) {
	if p := s.plugin(key); p != nil && p.Installed {
		s.runPlugin("update", p.Source, p.Scope)
	}
}
func (s *State) runPlugin(action, source, scope string) {
	if s.Busy || s.closed {
		return
	}
	if source == "" || strings.HasPrefix(source, "-") || strings.ContainsRune(source, 0) {
		s.Error = s.label("resources.invalid_source", "Enter a valid package source.")
		return
	}
	if scope == "project" && s.Workspace == "" {
		s.Error = s.label("resources.workspace_required", "Select a workspace before installing a project plugin.")
		return
	}
	service := s.service
	activity := map[string]string{"install": "Installing", "update": "Updating", "remove": "Removing"}[action]
	s.start(activity+" "+source+"…", action, scope+"\x00"+source, func(ctx context.Context) (func(*State), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		out, err := service.PluginAction(ctx, action, source, scope)
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		if err == nil && action == "update" {
			err = service.SkipPluginUpdate(source, "")
		}
		result := readSnapshot(service)
		return func(s *State) {
			s.applySnapshot(result)
			s.pluginRemove = ""
			s.modified()
			if err == nil {
				s.Status = lastLine(out)
				if s.Status == "" {
					s.Status = s.label("resources.operation_complete", "Plugin operation completed. Pi will reload when idle.")
				}
				var remaining []workbench.PluginUpdate
				for _, u := range s.Updates {
					if u.Source != source {
						remaining = append(remaining, u)
					}
				}
				s.Updates = remaining
			}
		}, err
	})
}

func (s *State) CheckUpdates() {
	if s.Options.Offline || os.Getenv("PI_OFFLINE") != "" {
		s.Status = s.label("resources.offline", "Plugin update checks are disabled by PI_OFFLINE.")
		return
	}
	service := s.service
	if s.start(s.label("resources.checking_updates", "Checking for plugin updates…"), "check", "", func(ctx context.Context) (func(*State), error) {
		updates, errs := service.FindPluginUpdates(ctx)
		checkError := ctx.Err()
		return func(s *State) {
			s.Updates = updates
			s.UpdateErrors = errorStrings(errs)
			s.CheckedAt = time.Now()
			if checkError == nil && len(errs) == 0 {
				s.Status = fmt.Sprintf("%d plugin update(s) available.", len(updates))
			}
		}, ctx.Err()
	}) {
		s.Checking = true
	}
}
func (s *State) SkipUpdate(source, version string) {
	if s.Busy || s.closed {
		return
	}
	service := s.service
	s.start(s.label("resources.saving_skip", "Saving update preference…"), "skip", source, func(ctx context.Context) (func(*State), error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		err := service.SkipPluginUpdate(source, version)
		if err != nil {
			return nil, err
		}
		skips, err := service.SkippedPluginUpdates()
		return func(s *State) {
			s.Skipped = skips
			var remaining []workbench.PluginUpdate
			for _, u := range s.Updates {
				if u.Source != source || u.Latest != version {
					remaining = append(remaining, u)
				}
			}
			s.Updates = remaining
			if version == "" {
				s.Status = s.label("resources.skip_cleared", "Skipped version cleared. Check for updates to see available versions again.")
			}
		}, err
	})
}
