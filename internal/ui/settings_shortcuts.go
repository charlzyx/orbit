package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
)

// The pinned registry's with_chord/help_only records, rather than palette-only
// actions and alternate binding aliases, supply the shortcut reference.
func sourceAdvertisedChord(id string) string {
	if id == "OpenSessionSlot" {
		return sourceShortcutLabel("secondary-1") + "–" + sourceShortcutLabel("secondary-9")
	}
	for _, b := range Bindings {
		if b.Command == id {
			return sourceShortcutLabel(b.Chord)
		}
	}
	return ""
}
func (a *App) sourceShortcutRow(c *ui.Context, title string, chords ...string) {
	a.sourceSettingsRow(c, title, "", "", func() {
		ui.Row(c).Gap(a.Theme.Space(4)).AlignItems(ui.Center).Children(func() {
			for _, chord := range chords {
				ui.Text(c, chord).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted).Padding(1, a.Theme.Space(6)).Radius(a.Theme.Px(8)).Background(a.Theme.Active)
			}
		})
	})
}
func (a *App) shortcutsSettings(c *ui.Context) {
	for _, category := range []string{"Session", "Agent", "Navigation", "Panels", "Review", "Git", "Application"} {
		a.sourceSettingsSection(c, a.Tr("shortcut.group_"+strings.ToLower(category)), "", func() {
			first := true
			for _, cmd := range Commands {
				if cmd.Category != category {
					continue
				}
				chord := sourceAdvertisedChord(cmd.ID)
				if chord == "" {
					continue
				}
				if !first {
					a.sourceSettingsDivider(c)
				}
				first = false
				a.sourceShortcutRow(c, a.Tr(cmd.TitleKey), chord)
			}
		})
		if category == "Agent" {
			a.sourceComposerShortcuts(c)
		}
	}
}
func (a *App) sourceComposerShortcuts(c *ui.Context) {
	follow, steer := "enter", "alt-enter"
	if a.Prefs.ComposerSendMode == "steer" {
		follow, steer = steer, follow
	}
	a.sourceSettingsSection(c, a.Tr("shortcut.group_composer"), "", func() {
		rows := []struct {
			label string
			keys  []string
		}{
			{a.Tr("shortcut.send_idle"), []string{"enter"}},
			{strings.ReplaceAll(a.Tr("shortcut.send_while_running"), "%{mode}", a.Tr("composer.queue_follow_up")), []string{follow}},
			{strings.ReplaceAll(a.Tr("shortcut.send_while_running"), "%{mode}", a.Tr("composer.steer_task")), []string{steer, "secondary-shift-enter"}},
			{a.Tr("shortcut.newline"), []string{"shift-enter"}},
			{a.Tr("shortcut.accept"), []string{"tab"}},
			{a.Tr("shortcut.stop"), []string{"escape"}},
		}
		for i, row := range rows {
			if i > 0 {
				a.sourceSettingsDivider(c)
			}
			labels := []string{}
			for _, key := range row.keys {
				labels = append(labels, sourceShortcutLabel(key))
			}
			a.sourceShortcutRow(c, row.label, labels...)
		}
	})
}
