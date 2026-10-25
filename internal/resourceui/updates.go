package resourceui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func lastLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if text := strings.TrimSpace(lines[i]); text != "" {
			return text
		}
	}
	return ""
}

// SkipAll persists the exact versions currently offered, including closing
// the launch notice. A future, newer version is not implicitly skipped.
func (s *State) SkipAll() {
	if len(s.Updates) == 0 || s.Busy || s.closed {
		return
	}
	updates := append([]workbench.PluginUpdate(nil), s.Updates...)
	service := s.service
	s.start(s.label("resources.saving_skip", "Saving update preference…"), "skip-all", "", func(ctx context.Context) (func(*State), error) {
		var failures []error
		saved := map[string]string{}
		for _, u := range updates {
			if ctx.Err() != nil {
				failures = append(failures, ctx.Err())
				break
			}
			if err := service.SkipPluginUpdate(u.Source, u.Latest); err != nil {
				failures = append(failures, err)
			} else {
				saved[u.Source] = u.Latest
			}
		}
		skips, err := service.SkippedPluginUpdates()
		if err != nil {
			failures = append(failures, err)
		}
		return func(s *State) {
			s.Skipped = skips
			var remaining []workbench.PluginUpdate
			for _, u := range s.Updates {
				if saved[u.Source] != u.Latest {
					remaining = append(remaining, u)
				}
			}
			s.Updates = remaining
		}, errors.Join(failures...)
	})
}

// UpdateAll is the source notice's sequential apply action, not an unbounded
// parallel package install. Failures stay visible and don't hide other results.
func (s *State) UpdateAll() {
	if len(s.Updates) == 0 || s.Busy || s.closed {
		return
	}
	updates := append([]workbench.PluginUpdate(nil), s.Updates...)
	service := s.service
	s.start(s.label("resources.updating_plugins", "Updating plugins…"), "update-all", "", func(ctx context.Context) (func(*State), error) {
		var failures []error
		var remaining []workbench.PluginUpdate
		updated := 0
		attempted := false
		for i, u := range updates {
			if ctx.Err() != nil {
				remaining = append(remaining, updates[i:]...)
				failures = append(failures, ctx.Err())
				break
			}
			attempted = true
			if _, err := service.PluginAction(ctx, "update", u.Source, u.Scope); err != nil {
				failures = append(failures, fmt.Errorf("%s: %w", u.Name, err))
				remaining = append(remaining, u)
			} else {
				updated++
				if err := service.SkipPluginUpdate(u.Source, ""); err != nil {
					failures = append(failures, err)
				}
			}
		}
		if ctx.Err() != nil {
			failures = append(failures, ctx.Err())
		}
		result := readSnapshot(service)
		return func(s *State) {
			s.applySnapshot(result)
			s.Updates = remaining
			if attempted {
				s.modified()
			}
			s.Status = s.label("resources.updated_count", "Updated %{count} plugin(s). Pi will reload when idle.", "count", fmt.Sprint(updated))
		}, errors.Join(failures...)
	})
}

// NoticeHeight returns the exact reserved height of the bottom-right notice.
// The host puts normal toasts above it with the usual eight-pixel stack gap.
func (s *State) NoticeHeight(c *ui.Context) float32 {
	if len(s.Updates) == 0 || s.closed || s.operation == "update-all" {
		return 0
	}
	a := s.appearance(c)
	n := min(3, len(s.Updates))
	h := 2*a.Space(8) + a.Px(22) + 2*a.Space(8) + float32(n)*(a.Px(14)+2*a.Space(6)) + float32(n-1) + a.Px(28) + 2
	if len(s.Updates) > 3 {
		h += a.Px(14) + a.Space(8)
	}
	return h
}

// Overlay is the persistent, source-style launch update notice. Mount once at
// the host root, after its ordinary content. It is intentionally not a modal.
func (s *State) Overlay(c *ui.Context) {
	height := s.NoticeHeight(c)
	if height == 0 {
		return
	}
	a := s.appearance(c)
	w, _ := c.Size()
	ui.Overlay(c.Key("plugin-update-notice"), func() {
		ui.Column(c).Absolute().Right(a.Space(16)).Bottom(a.Space(16)).Width(min(360, max(0, w-32))).Height(height).Padding(a.Space(8), a.Space(12)).Gap(a.Space(8)).Background(a.Elevated).Border(1, a.native.Border).Radius(a.Px(8)).Shadow(0, 4, 18, 0, ui.RGBA(0, 0, 0, .25)).Children(func() {
			ui.Row(c.Key("title")).Height(a.Px(22)).Shrink(0).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
				ui.Text(c, s.label("updater_ui.update_available", "Update available")).Grow(1).FontSize(a.Px(14)).FontWeight(500)
				if s.iconButton(c, "close", s.label("resources.close_update_notice", "Close update notice"), "x").Size(a.Px(22), a.Px(22)).Disabled(s.Busy).Clicked() {
					s.SkipAll()
				}
			})
			ui.Column(c.Key("packages")).Shrink(0).Children(func() {
				for i, u := range s.Updates {
					if i == 3 {
						break
					}
					if i > 0 {
						s.separator(c)
					}
					ui.Row(c.Key(updateKey(u))).Height(a.Px(14) + 2*a.Space(6)).Shrink(0).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
						ui.Text(c, u.Name).Grow(1).MinWidth(0).SingleLine().FontSize(a.Px(12))
						ui.Text(c, shortVersion(u.Current)+" → "+shortVersion(u.Latest)).Font(a.CodeFont).FontSize(a.Px(11)).TextColor(a.native.TextMuted).Shrink(0)
					})
				}
			})
			if len(s.Updates) > 3 {
				ui.Text(c, s.label("resources.more_updates", "and %{count} more", "count", fmt.Sprint(len(s.Updates)-3))).Height(a.Px(14)).Shrink(0).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			}
			ui.Row(c.Key("actions")).Height(a.Px(28)).Shrink(0).Gap(a.Space(8)).Justify(ui.End).Children(func() {
				if s.button(c, "skip-all", s.label("git_panel.recover_skip", "Skip"), "", false).Disabled(s.Busy).Clicked() {
					s.SkipAll()
				}
				if s.button(c, "update-all", s.label("updater_ui.update_now", "Update now"), "refresh", true).Disabled(s.Busy).Clicked() {
					s.UpdateAll()
				}
			})
		})
	})
}
func shortVersion(version string) string {
	if len(version) >= 40 {
		return version[:7]
	}
	return version
}
