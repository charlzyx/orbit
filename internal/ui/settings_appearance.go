package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/preferences"
	"github.com/egoist/mygo/ui"
	"strconv"
	"strings"
)

var sourceUIFontChoices = []settingsChoice{{"Inter", "Inter"}, {"Fixel Text", "Fixel Text"}, {"Geist", "Geist Sans"}, {"Atkinson Hyperlegible", "Atkinson Hyperlegible"}, {"Source Sans 3", "Source Sans 3"}, {"Roboto", "Roboto"}, {"Noto Sans", "Noto Sans"}, {"DM Sans", "DM Sans"}, {"Manrope", "Manrope"}, {".ZedSans", "IBM Plex Sans"}}
var sourceCodeFontChoices = []settingsChoice{{"JetBrains Mono", "JetBrains Mono"}, {"Fira Code", "Fira Code"}, {"Geist Mono", "Geist Mono"}, {"CommitMono", "Commit Mono"}, {"Source Code Pro", "Source Code Pro"}, {"Cascadia Code", "Cascadia Code"}, {"Roboto Mono", "Roboto Mono"}, {"Iosevka", "Iosevka"}, {"DM Mono", "DM Mono"}, {"IBM Plex Mono", "IBM Plex Mono"}, {"Inconsolata", "Inconsolata"}, {"Noto Sans Mono", "Noto Sans Mono"}, {"Space Mono", "Space Mono"}, {"Anonymous Pro", "Anonymous Pro"}, {"Martian Mono", "Martian Mono"}, {".ZedMono", "Lilex"}}
var sourceLanguageChoices = []settingsChoice{{"system", "System"}, {"en", "English"}, {"zh-CN", "简体中文"}, {"ja", "日本語"}, {"ko", "한국어"}, {"es", "Español"}, {"fr", "Français"}, {"de", "Deutsch"}, {"pt-BR", "Português do Brasil"}, {"ru", "Русский"}, {"it", "Italiano"}}

func sourceThemeChoices(dark bool) []settingsChoice {
	values := preferences.LightThemes
	if dark {
		values = preferences.DarkThemes
	}
	names := map[string]string{"orbit": "Orbit", "orbit-darker": "Orbit Darker", "orbit-cobalt": "Orbit Cobalt", "orbit-azure": "Orbit Azure", "flexoki": "Flexoki", "orbit-light": "Orbit Light", "orbit-azure-light": "Orbit Azure Light", "orbit-paper": "Orbit Paper", "orbit-contrast": "Orbit Contrast", "flexoki-light": "Flexoki Light"}
	out := make([]settingsChoice, 0, len(values))
	for _, v := range values {
		out = append(out, settingsChoice{v, names[v]})
	}
	return out
}
func (a *App) appearanceSettings(c *ui.Context) {
	a.sourceSettingsSection(c, a.Tr("settings.theme_and_background"), "", func() {
		if a.backdrop.config.Source != nil {
			ui.Box(c.Key("background-preview")).AlignSelf(ui.Stretch).Height(84).Radius(a.Theme.Px(12)).Clip().Draw(func(p *ui.Painter, r ui.Rect) { a.paintBackdrop(p, r, false) })
			a.sourceSettingsDivider(c)
		}
		a.sourceSettingsRow(c, a.Tr("settings.appearance"), a.Tr("settings.appearance_mode_description"), "", func() {
			choices := []settingsChoice{{"light", a.Tr("settings.appearance_light")}, {"dark", a.Tr("settings.appearance_dark")}, {"system", a.Tr("settings.appearance_system")}}
			a.sourceSettingsSegments(c, "Appearance mode", a.Appearance.Mode, choices, false, func(v string) { a.Appearance.Mode = v; a.save(a.Options.Store.SaveAppearance(a.Appearance)) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.theme"), "", "", func() {
			value := a.Appearance.LightTheme
			if a.Theme.Dark {
				value = a.Appearance.DarkTheme
			}
			a.sourceSettingsSelect(c, "Theme", value, sourceThemeChoices(a.Theme.Dark), func(v string) {
				if a.Theme.Dark {
					a.Appearance.DarkTheme = v
				} else {
					a.Appearance.LightTheme = v
				}
				a.save(a.Options.Store.SaveAppearance(a.Appearance))
			})
		})
		a.sourceSettingsDivider(c)
		a.sourceBackgroundRow(c)
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("language.title"), a.Tr("language.description"), "", func() {
			a.sourceSettingsSelect(c, "Language", a.Prefs.Language, sourceLanguageChoices, func(v string) {
				a.Prefs.Language = v
				a.reloadLocale()
				a.save(a.Options.Store.SaveUI(a.Prefs))
				a.Dispatch(Action{Kind: "LanguageChanged"})
			})
		})
	})
	if a.backdrop.config.Source != nil {
		a.backgroundSettings(c)
	}
	a.sourceSettingsSection(c, a.Tr("settings.type_and_density"), "", func() {
		a.sourceTypePreview(c)
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.interface_font"), "", "", func() {
			a.sourceSettingsSelect(c, "Interface font", a.Fonts.UIFamily, sourceUIFontChoices, func(v string) { a.Fonts.UIFamily = v; a.save(a.Options.Store.SaveFonts(a.Fonts)) })
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.code_font"), "", "", func() {
			a.sourceSettingsSelect(c, "Code font", a.Fonts.CodeFamily, sourceCodeFontChoices, func(v string) {
				a.Fonts.CodeFamily = v
				a.save(a.Options.Store.SaveFonts(a.Fonts))
				a.Dispatch(Action{Kind: "TerminalFontChanged"})
			})
		})
		for _, r := range []struct {
			key, id string
			value   *float64
		}{{"settings.ui_font_size", "UI font size", &a.Prefs.UIFontSize}, {"settings.terminal_size", "Terminal font size", &a.Prefs.TerminalFontSize}, {"settings.editor_size", "Editor font size", &a.Prefs.EditorFontSize}} {
			a.sourceSettingsDivider(c)
			a.sourceSettingsRow(c, a.Tr(r.key), "", "", func() {
				choices := []settingsChoice{}
				for _, n := range preferences.FontSizes {
					v := fmt.Sprint(n)
					choices = append(choices, settingsChoice{v, v + " px"})
				}
				a.sourceSettingsSelect(c, r.id, fmt.Sprint(*r.value), choices, func(v string) {
					*r.value, _ = strconv.ParseFloat(v, 64)
					a.save(a.Options.Store.SaveUI(a.Prefs))
					if r.id == "Terminal font size" {
						a.Dispatch(Action{Kind: "TerminalFontChanged"})
					}
				})
			})
		}
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.spacing_density"), "", "", func() {
			choices := []settingsChoice{}
			for _, n := range preferences.SpacingDensities {
				v := strconv.Itoa(n)
				choices = append(choices, settingsChoice{v, v + " %"})
			}
			a.sourceSettingsSelect(c, "Spacing density", strconv.Itoa(a.Prefs.SpacingDensity), choices, func(v string) { a.Prefs.SpacingDensity, _ = strconv.Atoi(v); a.save(a.Options.Store.SaveUI(a.Prefs)) })
		})
	})
	a.sourceSettingsSection(c, a.Tr("settings.layout"), "", func() {
		a.sourceSettingsRow(c, a.Tr("settings.show_sidebar"), strings.ReplaceAll(a.Tr("settings.show_the_sessions_sidebar_also_toggleable_from_t"), "%{shortcut}", sourceShortcutLabel("secondary-b")), "", func() {
			visible := a.SidebarVisible
			if a.sourceSettingsToggle(c, &visible, "Show sidebar").Changed() {
				a.command("ToggleSidebar", "")
			}
		})
		a.sourceSettingsDivider(c)
		a.sourceSettingsRow(c, a.Tr("settings.reduce_motion"), a.Tr("settings.stop_looping_animations_spinners_and_the_running"), "", func() {
			if a.sourceSettingsToggle(c, &a.Prefs.ReduceMotion, "Reduce motion").Changed() {
				a.save(a.Options.Store.SaveUI(a.Prefs))
			}
		})
	})
}
func (a *App) sourceTypePreview(c *ui.Context) {
	t := a.Theme
	ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Stretch).Padding(t.Space(12), t.Space(16)).Gap(t.Space(16)).Children(func() {
		for _, terminal := range []bool{false, true} {
			title := a.Tr("settings.interface_preview")
			if terminal {
				title = a.Tr("settings.terminal_preview")
			}
			ui.Column(c).Grow(1).Basis(0).MinWidth(0).Gap(t.Space(6)).Children(func() {
				ui.Text(c, strings.ToUpper(title)).FontSize(t.Px(10)).Height(sourceTextHeight(t.Px(10), sourceDefaultLineHeight)).FontWeight(500).TextColor(t.TextMuted)
				if !terminal {
					ui.Column(c).Label("Interface typography preview").AlignSelf(ui.Stretch).Grow(1).Padding(t.Space(12)).Gap(t.Space(6)).Radius(t.Px(12)).Background(t.Background).Justify(ui.Center).Children(func() {
						ui.Text(c, a.Tr("settings.new_task")).FontSize(t.Px(16)).FontWeight(500)
						ui.Text(c, a.Tr("settings.the_quick_brown_fox_jumps_over_the_lazy_dog")).FontSize(t.Px(12)).TextColor(t.TextSecondary)
					})
					return
				}
				termScale := float32(a.Prefs.TerminalFontSize) / 13
				ui.Column(c).Label("Terminal typography preview").AlignSelf(ui.Stretch).Border(1, t.Border).Radius(t.Px(12)).Background(t.CodeBg).Clip().Children(func() {
					ui.Row(c).AlignSelf(ui.Stretch).Height(26).PaddingX(t.Space(8)).Gap(t.Space(6)).Background(t.Surface).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).AlignItems(ui.Center).Children(func() {
						for _, color := range []ui.Color{t.Error, t.Warning, t.Success} {
							ui.Box(c).Size(7, 7).Radius(3.5).Background(color)
						}
						ui.Text(c, a.Tr("settings.terminal_prompt")).Margin(0, 0, 0, 6).Font(a.Fonts.CodeFamily).FontSize(termScale * 10.5).TextColor(t.TextMuted)
					})
					ui.Column(c).AlignSelf(ui.Stretch).Padding(t.Space(8), t.Space(12)).Gap(t.Space(3)).Font(a.Fonts.CodeFamily).FontSize(termScale * 12.5).Children(func() {
						prompt := func(command string, cursor bool) {
							ui.Row(c).AlignItems(ui.Center).Gap(t.Space(6)).Children(func() {
								ui.Text(c, "$").TextColor(t.Success)
								if command != "" {
									ui.Text(c, command).TextColor(t.CodeText)
								}
								if cursor {
									ui.Box(c).Size(termScale*7, termScale*14).Radius(7).Background(t.TextSecondary)
								}
							})
						}
						prompt("orbit-pi --session", false)
						ui.Text(c, a.Tr("settings.preview_prompt")).TextColor(t.TextSecondary)
						prompt("", true)
					})
				})
			})
		}
	})
}
