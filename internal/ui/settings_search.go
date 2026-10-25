// Source-derived from pinned settings.rs::SETTINGS_SEARCH_INDEX. Labels are
// translated at query time; synonyms retain the source order and semantics.
package orbitui

import (
	"github.com/egoist/mygo/ui"
	"strings"
)

type settingsSearchEntry struct{ Section, Label, Keywords string }

var settingsSearchIndex = []settingsSearchEntry{
	{"general", "settings.pi_agent", "pi agent process stdio connection child spawn"},
	{"general", "settings.local_by_default", "local data store sessions disk offline"},
	{"general", "settings.workspace", "workspace directory folder path project"},
	{"general", "updater_ui.automatic_updates", "updates update automatic release"},
	{"general", "updater_ui.check_for_updates", "updates update check release version"},
	{"general", "settings.desktop_notifications", "notifications banner desktop system"},
	{"general", "settings.in_app_toasts", "notifications toast in-app"},
	{"general", "settings.notification_sound", "notifications sound alert audio"},
	{"runtime", "settings.status", "runtime process status start stop restart"},
	{"runtime", "settings.process_id", "runtime process id pid"},
	{"runtime", "settings.binary", "runtime binary executable pi"},
	{"runtime", "settings.uptime", "runtime uptime elapsed"},
	{"runtime", "settings.transport", "runtime transport stdio"},
	{"runtime", "settings.last_error", "runtime error stderr last"},
	{"runtime", "settings.parked_processes", "runtime background parked sessions"},
	{"runtime", "settings.recent_stderr", "runtime stderr log output console"},
	{"agent", "settings.enter_while_running", "enter send submit queue steer behavior"},
	{"agent", "settings.follow_up_messages", "follow-up queue messages behavior"},
	{"agent", "settings.auto_compaction", "compaction compact context auto"},
	{"agent", "settings.auto_retry", "retry auto error transient"},
	{"agent", "settings.auto_session_titles", "titles session name auto"},
	{"agent", "settings.title_model", "title model session names"},
	{"agent", "settings.default_model", "default model session new"},
	{"agent", "settings.default_thinking", "thinking level default session reasoning"},
	{"worktrees", "worktree.settings.directory", "worktree worktrees directory location .wt folder path git branch"},
	{"worktrees", "worktree.settings.setup_script", "worktree worktrees setup script install bootstrap hook"},
	{"worktrees", "worktree.settings.run_setup_auto", "worktree worktrees setup script automatic auto run"},
	{"worktrees", "worktree.settings.open_after_create", "worktree worktrees open after create switch workspace"},
	{"appearance", "settings.appearance", "appearance mode light dark system theme"},
	{"appearance", "settings.theme", "theme palette color light dark editor"},
	{"appearance", "settings.background_image", "background image backdrop dither picture"},
	{"appearance", "settings.blur", "background blur backdrop soft"},
	{"appearance", "settings.pixel_size", "background pixel size backdrop dither cell"},
	{"appearance", "settings.bottom_fade", "background fade backdrop gradient"},
	{"appearance", "language.title", "language locale translation interface"},
	{"appearance", "settings.interface_font", "font interface ui typeface"},
	{"appearance", "settings.code_font", "font code mono typeface"},
	{"appearance", "settings.ui_font_size", "font size ui text scale"},
	{"appearance", "settings.terminal_size", "font size terminal"},
	{"appearance", "settings.editor_size", "font size editor"},
	{"appearance", "settings.spacing_density", "spacing density compact comfortable"},
	{"appearance", "settings.show_sidebar", "sidebar show layout navigation"},
	{"appearance", "settings.reduce_motion", "motion animation reduce layout"},
	{"skills", "settings.skills", "skills skill.md instructions agent"},
	{"plugins", "settings.plugins", "plugins packages extensions install npm git"},
	{"mcp", "settings.mcp", "mcp model context protocol servers tools"},
	{"models", "settings.models", "models catalog favorites providers"},
	{"providers", "settings.providers", "providers api key auth credentials"},
	{"shortcuts", "settings.shortcuts", "shortcuts keyboard keys chords"},
	{"privacy", "settings.privacy", "privacy session retention cleanup local data"},
	{"privacy", "settings.temporary_files", "temporary temp cache scratch files cleanup remove folder"},
	{"privacy", "settings.ai_scratch_files", "ai pi agent scratch logs codemode images share editor cleanup temp"},
	{"about", "settings.about", "about version updates source github"},
	{"bug", "settings.report_bug", "bug report issue crash problem broken github feedback"},
	{"bug", "bug_report.heading", "bug report issue title description steps reproduce diagnostics"},
}

func (a *App) settingsSearchHits(query string) []settingsSearchEntry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}
	var out []settingsSearchEntry
	for _, entry := range settingsSearchIndex {
		if strings.Contains(strings.ToLower(a.Tr(entry.Label)), query) || strings.Contains(entry.Keywords, query) {
			out = append(out, entry)
		}
	}
	return out
}
func (a *App) settingsSearchResults(c *ui.Context) {
	t := a.Theme
	hits := a.settingsSearchHits(a.settingsSearch)
	ui.Scroll(c.Key("settings-search-results")).Grow(1).FillHeight().MinWidth(0).Children(func() {
		ui.Column(c).WidthPercent(100).MaxWidth(ContentMaxWidth).AlignSelf(ui.Center).Padding(t.Space(20), t.Space(24), t.Space(24)).Children(func() {
			if len(hits) == 0 {
				a.Empty(c, a.text("settings.no_matches", "No matches"), a.text("settings.try_a_different_search", "Try a different search."))
				return
			}
			ui.Column(c).AlignSelf(ui.Stretch).Background(t.SurfaceSecondary).Border(1, t.Border).Radius(t.Px(12)).Children(func() {
				for i, hit := range hits {
					title, icon := hit.Section, "icons/settings.svg"
					for _, cat := range settingsCategories {
						if cat.ID == hit.Section {
							title = a.settingsTitle(cat.ID, cat.Title)
							icon = cat.Icon
							break
						}
					}
					b := ui.ButtonBase(c.Key(hit.Label)).Label(a.Tr(hit.Label)).AlignSelf(ui.Stretch).Padding(t.Space(12), t.Space(16)).Gap(t.Space(12))
					if b.Hovered() {
						b.Background(t.Hover)
					}
					b.Children(func() {
						a.icon(c, icon, 14, t.TextMuted)
						ui.Text(c, a.Tr(hit.Label)).Grow(1)
						ui.Text(c, title).FontSize(t.Px(12)).TextColor(t.TextMuted)
					})
					if b.Clicked() {
						a.settingsSearch = ""
						a.selectSettingsPage(hit.Section)
					}
					if i < len(hits)-1 {
						ui.Box(c).AlignSelf(ui.Stretch).Height(1).MarginX(t.Space(16)).Background(t.Border)
					}
				}
			})
		})
	})
}
