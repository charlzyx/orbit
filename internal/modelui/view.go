package modelui

import (
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"time"

	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
)

type appearance struct {
	Theme
	native *ui.Theme
}

func (s *State) appearance(c *ui.Context) appearance {
	a := appearance{native: c.Theme()}
	if s.Options.Theme != nil {
		a.Theme = s.Options.Theme()
	}
	if a.Px == nil {
		a.Px = func(v float32) float32 { return v * a.native.FontSize / 14 }
	}
	if a.CodePx == nil {
		a.CodePx = a.Px
	}
	if a.Space == nil {
		a.Space = func(v int) float32 { return a.Px(float32(v)) }
	}
	if a.Elevated == (ui.Color{}) {
		a.Elevated = a.native.Surface
	}
	if a.Secondary == (ui.Color{}) {
		a.Secondary = a.native.Text.Alpha(.82)
	}
	if a.Active == (ui.Color{}) {
		a.Active = a.native.Selection
	}
	if a.Hover == (ui.Color{}) {
		a.Hover = a.native.SurfaceHover
	}
	if a.BorderStrong == (ui.Color{}) {
		a.BorderStrong = a.native.Border
	}
	if a.AccentText == (ui.Color{}) {
		a.AccentText = a.native.Accent
	}
	if a.Overlay == (ui.Color{}) {
		a.Overlay = a.native.Text.Alpha(.08)
	}
	return a
}

var iconOnce sync.Once
var icons map[string]*ui.SVG

func icon(c *ui.Context, path string, size float32, color ui.Color) ui.Element {
	iconOnce.Do(func() {
		icons = map[string]*ui.SVG{}
		_ = fs.WalkDir(orbitassets.Assets(), ".", func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(path, ".svg") {
				if data, e := orbitassets.Asset(path); e == nil {
					if svg, e := ui.ParseSVG(data); e == nil {
						icons[path] = svg
					}
				}
			}
			return nil
		})
	})
	if svg := icons[path]; svg != nil {
		return ui.Icon(c, svg).Size(size, size).Shrink(0).TextColor(color)
	}
	return ui.Element{}
}
func providerIcon(id string) string {
	if id == "ollama-cloud" {
		id = "ollama"
	}
	if id == ".manifest" {
		id = "manifest"
	}
	if _, err := fs.Stat(orbitassets.Assets(), "icons/providers/"+id+".svg"); err == nil {
		return "icons/providers/" + id + ".svg"
	}
	return "icons/cloud.svg"
}
func providerColor(provider string, a appearance) ui.Color {
	colors := map[string][2]uint32{"amazon-bedrock": {0xff9900, 0xc77700}, "ant-ling": {0x1677ff, 0x0e5fcc}, "anthropic": {0xd97757, 0xb55035}, "azure-openai-responses": {0x0078d4, 0x005a9e}, "baseten": {0x16a34a, 0x128040}, "cerebras": {0xf15a29, 0xc7461c}, "cloudflare-ai-gateway": {0xf38020, 0xc25e00}, "cloudflare-workers-ai": {0xf38020, 0xc25e00}, "deepseek": {0x4d6bfe, 0x3a52c9}, "fireworks": {0x8b5cf6, 0x6d28d9}, "google": {0x4285f4, 0x1a73e8}, "google-vertex": {0x4285f4, 0x1a73e8}, "groq": {0xf55036, 0xc93a22}, "huggingface": {0xffd21e, 0xa87800}, "kimi-coding": {0x3d5afe, 0x2e45c9}, "moonshotai": {0x3d5afe, 0x2e45c9}, "moonshotai-cn": {0x3d5afe, 0x2e45c9}, "minimax": {0xff4d4f, 0xcc3b3d}, "minimax-cn": {0xff4d4f, 0xcc3b3d}, "mistral": {0xfa520f, 0xc63d00}, "nvidia": {0x76b900, 0x5a8c00}, "openrouter": {0x6467f2, 0x5053c9}, "together": {0x0f6fff, 0x0b57cc}, "zai": {0x3b5bff, 0x2a43cc}, "zai-coding-cn": {0x3b5bff, 0x2a43cc}}
	if strings.HasPrefix(provider, "qwen-token-plan") {
		provider = "qwen-token-plan"
		colors[provider] = [2]uint32{0x615ced, 0x4e49c9}
	}
	if strings.HasPrefix(provider, "xiaomi") {
		provider = "xiaomi"
		colors[provider] = [2]uint32{0xff6900, 0xcc5400}
	}
	values, ok := colors[provider]
	if !ok {
		return a.native.TextMuted
	}
	v := values[1]
	if a.native.Dark {
		v = values[0]
	}
	return ui.RGB(uint8(v>>16), uint8(v>>8), uint8(v))
}
func (s *State) button(c *ui.Context, key, label, glyph string, height float32, active bool) ui.Element {
	a := s.appearance(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(a.Px(height)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(4)).Radius(a.Px(8)).Border(1, a.native.Border).Background(a.Elevated).AlignItems(ui.Center)
	if active {
		b.Background(a.Active)
	} else if b.Hovered() {
		b.Background(a.Hover)
	}
	b.Children(func() {
		if glyph != "" {
			icon(c, glyph, a.Px(16), a.Secondary)
		}
		ui.Text(c, label).FontSize(a.Px(14)).LineHeight(1).FontWeight(500).TextColor(a.Secondary).SingleLine()
	})
	return b
}
func (s *State) notice(c *ui.Context, key, title, detail string, isError bool) {
	a := s.appearance(c)
	color := a.native.TextMuted
	if isError {
		color = a.native.Danger
	}
	ui.Column(c.Key(key)).AlignSelf(ui.Stretch).Padding(a.Space(16)).Gap(a.Space(6)).Radius(a.Px(8)).Border(1, a.native.Border).Background(a.Elevated).Children(func() {
		ui.Text(c, title).FontSize(a.Px(14)).FontWeight(500).TextColor(color)
		if detail != "" {
			ui.Text(c, detail).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		}
	})
}

// Toolbar mounts in the host's fixed Models header. It adds no page title or
// competing scroll container; View supplies the scrolling body separately.
func (s *State) Toolbar(c *ui.Context) ui.Element {
	c = c.Key("models-toolbar")
	a := s.appearance(c)
	return ui.Column(c).AlignSelf(ui.Stretch).Gap(a.Space(12)).Children(func() {
		ui.Row(c.Key("search-row")).AlignSelf(ui.Stretch).Height(a.Px(36)).Padding(0, a.Px(10)).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
			icon(c, "icons/search.svg", a.Px(16), a.native.TextMuted)
			ui.TextInput(c.Key("query"), &s.Query).Border(0, ui.Color{}).Background(ui.Color{}).FocusRing(false).Padding(0).Label("Search model catalog").Placeholder(s.label("app.search_models", "Search models…")).Grow(1).Height(a.Px(36))
		})
		ui.Row(c.Key("status-actions")).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
			ui.Text(c, s.CountLabel()).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Grow(1).MinWidth(0).SingleLine()
			if len(s.Providers()) > 1 {
				s.ProviderControl(c)
			}

			updateLabel := s.label("settings.models_update", "Update model list")
			b := ui.ButtonBase(c.Key("update")).Label(updateLabel).Height(a.Px(32)).Shrink(0).Padding(0, a.Space(8)).Gap(a.Space(4)).Radius(a.Px(8)).Border(1, a.native.Border).Background(a.Elevated).AlignItems(ui.Center)
			if s.Refreshing {
				b.Opacity(.6)
			} else if b.Hovered() {
				b.Background(a.Hover)
			}
			b.Children(func() {
				if s.Refreshing {
					glyph := icon(c.Key("catalog-refresh-glyph"), "icons/refresh.svg", a.Px(16), a.AccentText).Label("Catalog update progress")
					if !a.ReduceMotion && !c.Preferences().ReduceMotion {
						glyph.Rotate(glyph.Loop("catalog-refresh", 700*time.Millisecond, ui.Linear) * 360)
					}
				} else {
					icon(c, "icons/refresh.svg", a.Px(16), a.Secondary)
				}
				ui.Text(c, updateLabel).FontSize(a.Px(14)).LineHeight(1).FontWeight(500).TextColor(a.Secondary).SingleLine()
			})
			if b.Clicked() {
				s.Refresh()
			}

			if s.button(c, "favorites", s.label("settings.favorites", "Favorites"), "icons/star.svg", 32, s.FavoritesOnly).Clicked() {
				s.FavoritesOnly = !s.FavoritesOnly
			}
		})
	})
}
func (s *State) View(c *ui.Context) ui.Element {
	c = c.Key("models-body")
	a := s.appearance(c)
	return ui.Column(c).AlignSelf(ui.Stretch).Gap(a.Space(16)).Children(func() {

		detail := s.Error
		if s.RefreshError != "" {
			detail = s.RefreshError
		}
		if s.Snapshot.Error != "" {
			detail = s.Snapshot.Error
		}
		if detail != "" {
			s.notice(c, "error", s.label("models.error", "Model catalog"), detail, true)
		}
		if len(s.Snapshot.Models) == 0 {
			if s.Snapshot.Loading {
				s.notice(c, "loading", s.label("models.loading", "Loading models…"), "", false)
			} else {
				hint := s.label("settings.no_models_reported_hint", "The running pi has not returned a model catalog yet. Refresh from Settings → Providers, or start a session.")
				if !s.Snapshot.Connected {
					hint = s.label("models.disconnected", "Start or reconnect pi to load its available model catalog.")
				}
				s.notice(c, "empty", s.label("settings.no_models_reported", "No models reported"), hint, false)
			}
			return
		}
		ui.Row(c.Key("active-board")).AlignSelf(ui.Stretch).Padding(a.Space(12)).AlignItems(ui.Center).Gap(a.Space(12)).Background(a.Elevated).Border(1, a.native.Border).Radius(a.Px(8)).Children(func() {
			ui.Column(c).Grow(1).MinWidth(0).Gap(a.Space(4)).Children(func() {
				ui.Text(c, s.label("settings.active_model", "Active model")).FontSize(a.Px(14))
				ui.Text(c, s.label("settings.active_model_description", "The model the current session uses.")).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
			})
			s.ActiveModelControl(c)
		})
		groups := s.Groups()
		for _, g := range groups {
			g := g
			ui.Column(c.Key("provider:" + g.Provider)).AlignSelf(ui.Stretch).Gap(a.Space(8)).Children(func() {
				ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(a.Space(8)).Padding(0, 0, a.Space(6), 0).BorderWidth(0, 0, 1, 0).BorderColor(a.native.Border).Children(func() {
					icon(c, providerIcon(g.Provider), a.Px(14), providerColor(g.Provider, a))
					ui.Text(c, ProviderName(g.Provider)).FontSize(a.Px(12)).FontWeight(500).TextColor(a.Secondary).Grow(1)
					ui.Text(c, s.label("settings.n_models", "%{count} models", "count", fmt.Sprint(len(g.Models)))).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				})
				ui.Grid(c).Columns(3).Gap(a.Space(8)).AlignSelf(ui.Stretch).Children(func() {
					for _, i := range g.Models {
						s.card(c, i)
					}
				})
			})
		}
		if len(groups) == 0 {
			hint := s.label("settings.no_models_match_hint", "Try a different search — the catalog itself is unchanged.")
			if s.FavoritesOnly {
				hint = s.label("settings.no_models_match_favorites_hint", "Nothing favorited matches this search — star a model, or turn the filter off.")
			}
			s.notice(c, "no-matches", s.label("settings.no_models_match", "No models match"), hint, false)
		}
	})
}
func modelLabel(name, id string) string {
	if name != "" {
		return name
	}
	return id
}
func (s *State) card(c *ui.Context, index int) {
	m := s.Snapshot.Models[index]
	a := s.appearance(c)
	c = c.Key("model:" + m.Provider + ":" + m.ID)
	active := s.IsActive(m)
	meta := a.native.TextMuted
	card := ui.ButtonBase(c).Column().AlignItems(ui.Stretch).AlignSelf(ui.Stretch).MinWidth(0).Label("Select model "+m.Provider+" / "+m.ID).Padding(a.Space(12)).Gap(a.Space(8)).Radius(a.Px(12)).Border(1, a.native.Border).Background(a.native.Surface)
	if active {
		card.Background(a.Active).Border(1, a.BorderStrong)
		meta = a.native.Text.Alpha(.72)
	} else if card.Hovered() {
		card.Background(a.Hover).Border(1, a.BorderStrong)
	}
	card.Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
			ui.Text(c, modelLabel(m.Name, m.ID)).FontSize(a.Px(14)).FontWeight(500).TextColor(a.native.Text).Grow(1).MinWidth(0).SingleLine()
			s.star(c, index, false)
		})
		ui.Text(c, m.ID).Font(a.CodeFont).FontSize(a.CodePx(10.5)).TextColor(meta).SingleLine()
		ui.Row(c).AlignSelf(ui.Stretch).AlignItems(ui.Center).Gap(a.Space(6)).Children(func() {
			if m.ContextWindow > 0 {
				s.badge(c, s.label("settings.n_ctx", "%{count} ctx", "count", formatTokens(m.ContextWindow)), meta)
			}
			ui.Spacer(c).Grow(1)
			if active {
				ui.Row(c).Height(a.Space(20)).Padding(0, a.Space(6)).Gap(a.Space(6)).Radius(a.Px(4)).Background(a.Overlay).AlignItems(ui.Center).Children(func() {
					icon(c, "icons/check.svg", a.Px(12), a.native.Text)
					ui.Text(c, s.label("settings.active", "Active")).FontSize(a.Px(10)).FontWeight(500).TextColor(a.native.Text)
				})
			} else {
				label := ui.Text(c, s.label("settings.set_active", "Set active")).Height(a.Space(20)).FontSize(a.Px(10)).FontWeight(500).TextColor(a.native.TextMuted)
				if !card.Hovered() {
					label.Opacity(0)
				}
			}
		})
	})
	if card.Clicked() {
		s.SelectModel(index, false)
	}
}
func (s *State) badge(c *ui.Context, text string, color ui.Color) {
	a := s.appearance(c)
	ui.Row(c).Height(a.Space(20)).Padding(0, a.Space(6)).Radius(a.Px(4)).Background(a.Overlay).AlignItems(ui.Center).Children(func() { ui.Text(c, text).FontSize(a.Px(10)).FontWeight(500).TextColor(color).SingleLine() })
}
func (s *State) star(c *ui.Context, index int, hidden bool) {
	a := s.appearance(c)
	m := s.Snapshot.Models[index]
	favorite := s.IsFavorite(m)
	label := s.label("settings.favorite", "Favorite")
	if favorite {
		label = s.label("settings.unfavorite", "Unfavorite")
	}
	label += " " + m.Provider + " / " + m.ID
	color := a.native.TextMuted
	if favorite {
		color = a.AccentText
	}
	b := ui.ButtonBase(c.Key("favorite:"+m.Provider+":"+m.ID)).Label(label).Tooltip(label).Size(a.Px(22), a.Px(22)).Padding(0).Center().Radius(a.Px(8))
	if hidden && !favorite {
		b.Opacity(0)
	}
	if b.Hovered() {
		b.Background(a.Hover).Opacity(1)
	}
	b.Children(func() { icon(c, "icons/star.svg", a.Px(14), color) })
	if b.Clicked() {
		s.ToggleFavorite(index)
	}
}
func (s *State) PickerView(c *ui.Context) ui.Element {
	if !s.picker.open {
		return ui.Element{}
	}
	c = c.Key("model-selector:" + string(s.picker.kind))
	a := s.appearance(c)
	root := ui.Column(c).Width(s.PickerWidth()).AlignSelf(ui.Stretch).Background(a.Elevated).Border(1, a.BorderStrong).Radius(a.Px(8)).Shadow(0, 8, 16, 0, ui.RGBA(0, 0, 0, .22)).Gap(0).LineHeight(1.618)
	pointerX, pointerY, pointerOver := root.PointerPosition()
	bounds := root.Bounds()
	pointerX += bounds.X
	pointerY += bounds.Y
	s.picker.pointerMoved = pointerOver && (!s.picker.pointerKnown || pointerX != s.picker.pointerX || pointerY != s.picker.pointerY)
	if pointerOver {
		s.picker.pointerKnown = true
		s.picker.pointerX = pointerX
		s.picker.pointerY = pointerY
	}
	root.Children(func() {
		placeholder := s.label("model_selector.search_models", "Search models…")
		if s.picker.kind == ThinkingPicker {
			placeholder = s.label("model_selector.search_levels", "Search levels…")
		}
		ui.Row(c.Key("search-row")).Height(a.Px(36)).Padding(0, a.Px(10)).AlignItems(ui.Center).Gap(a.Space(8)).Children(func() {
			icon(c, "icons/search.svg", a.Px(16), a.native.TextMuted)
			in := ui.TextInput(c.Key("search"), &s.picker.query).Border(0, ui.Color{}).Background(ui.Color{}).FocusRing(false).Padding(0).Label("Model picker search").Placeholder(placeholder).AutoFocus().Grow(1).OnTextKey(func(k ui.TextKey) bool {
				switch k.Key {
				case ui.KeyDown:
					s.StepPicker(1)
					return true
				case ui.KeyUp:
					s.StepPicker(-1)
					return true
				case ui.KeyEnter:
					s.ConfirmPicker()
					return true
				case ui.KeyEscape:
					s.DismissPicker(false)
					return true
				}
				return false
			})
			if in.Changed() && s.picker.query != s.picker.lastQuery {
				query := s.picker.query
				s.picker.query = s.picker.lastQuery
				s.SetPickerQuery(query)
			}
		})
		ui.Box(c).Height(1).Shrink(0).Background(a.native.Border)
		if s.picker.kind == ModelPicker {
			s.scopeRow(c)
		}
		rows := s.PickerRows()
		if len(rows) == 0 {
			title := s.label("model_selector.no_matching_models", "No matching models")
			hint := s.label("model_selector.try_a_provider_model_or_id", "Try a provider, model, or id")
			if s.picker.kind == ThinkingPicker {
				title = s.label("model_selector.no_matching_levels", "No matching levels")
				hint = s.label("model_selector.try_a_reasoning_level", "Try a reasoning level")
			}
			ui.Row(c.Key("no-match")).Height(a.Px(26*1.618+8)+a.Space(4)*2).Padding(a.Space(4), a.Space(10)).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
				icon(c, "icons/search.svg", a.Px(16), a.native.TextMuted)
				ui.Column(c).Grow(1).Gap(0).Children(func() {
					ui.Text(c, title).FontSize(a.Px(14)).FontWeight(500).TextColor(a.native.TextMuted)
					ui.Text(c, hint).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
				})
			})
		} else {
			if s.picker.cursor < 0 || s.picker.cursor >= len(rows) || !rows[s.picker.cursor].Selectable() {
				s.picker.cursor = firstSelectable(rows)
			}
			rowH := a.Px(26*1.618 + 8)
			maxH := 6 * (rowH + a.Space(1))
			reveal := s.picker.scrollPending
			if reveal {
				s.picker.scroll.Y = centeredOffset(rows, s.picker.cursor, rowH, a.Px(20)+a.Space(4), a.Space(1), maxH)
			}
			ui.Scroll(c.Key("list")).KeepFocus().TrackScroll(&s.picker.scroll).MaxHeight(maxH).AlignSelf(ui.Stretch).Children(func() {
				ui.Column(c).AlignSelf(ui.Stretch).Padding(a.Space(4), 0).Gap(a.Space(1)).Children(func() {
					for i, row := range rows {
						s.pickerRow(c, i, row, rowH, reveal && i == s.picker.cursor)
					}
				})
			})
			s.picker.scrollPending = false
		}
		ui.Box(c).Height(1).Shrink(0).Background(a.native.Border)
		count := 0
		for _, row := range rows {
			if row.Selectable() {
				count++
			}
		}
		countKey, countLabel := "n_models", "%{count} models"
		if s.picker.kind == ThinkingPicker {
			countKey, countLabel = "n_levels", "%{count} levels"
		} else if s.picker.scope == Recents {
			countKey, countLabel = "n_recent", "%{count} recent"
		} else if s.picker.scope == Favorites {
			countKey, countLabel = "n_favorites", "%{count} favorites"
		}
		ui.Row(c.Key("footer")).Padding(a.Space(6), a.Px(10)).Gap(a.Space(8)).AlignItems(ui.Center).Children(func() {
			ui.Text(c, s.label("model_selector."+countKey, countLabel, "count", fmt.Sprint(count))).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Grow(1).SingleLine()
			for _, item := range []struct{ key, text string }{{"navigate", "↑↓ Navigate"}, {"select", "⏎ Select"}, {"esc_close", "Esc Close"}} {
				ui.Text(c, s.label("model_selector."+item.key, item.text)).FontSize(a.Px(12)).TextColor(a.native.TextMuted).SingleLine()
			}
		})
	})
	return root
}
func (s *State) scopeRow(c *ui.Context) {
	a := s.appearance(c)
	ui.ScrollHorizontal(c.Key("scopes")).KeepFocus().AlignSelf(ui.Stretch).MaxWidth(s.PickerWidth()).Children(func() {
		ui.Row(c).Padding(a.Space(4), a.Px(10)).Gap(a.Space(6)).Children(func() {
			type entry struct {
				label, icon string
				scope       Scope
			}
			entries := []entry{{s.label("model_selector.all", "All"), "icons/extensions.svg", All}}
			if len(s.recents) > 0 {
				entries = append(entries, entry{s.label("model_selector.recents", "Recents"), "icons/clock.svg", Recents})
			}
			if len(s.favorites) > 0 {
				entries = append(entries, entry{s.label("model_selector.favorites", "Favorites"), "icons/star.svg", Favorites})
			}
			for _, p := range s.Providers() {
				entries = append(entries, entry{ProviderName(p), providerIcon(p), ProviderScope(p)})
			}
			for _, e := range entries {
				if s.button(c, "scope:"+e.scope.Kind+":"+e.scope.Provider, e.label, e.icon, 28, e.scope == s.picker.scope).Clicked() {
					s.SetScope(e.scope)
				}
			}
			if len(s.favorites) == 0 {
				for i, m := range s.Snapshot.Models {
					if IsPickerSelected(m, s.Snapshot.Current) {
						if s.button(c, "favorite-current", s.label("model_selector.favorite_current", "Favorite current"), "icons/star.svg", 28, false).Clicked() {
							s.ToggleFavorite(i)
						}
						break
					}
				}
			}
		})
	})
	ui.Box(c).Height(1).Shrink(0).Background(a.native.Border)
}
func (s *State) pickerRow(c *ui.Context, index int, row Row, height float32, reveal bool) {
	a := s.appearance(c)
	c = c.Key(fmt.Sprintf("row:%d", index))
	if row.Header != "" {
		label := ProviderName(row.Provider)
		glyph := ""
		if row.Header == "recents" {
			label = s.label("model_selector.recents", "Recents")
			glyph = "icons/clock.svg"
		} else if row.Header == "favorites" {
			label = s.label("model_selector.favorites", "Favorites")
			glyph = "icons/star.svg"
		}
		ui.Row(c).Height(a.Px(20)+a.Space(4)).Padding(0, a.Px(8), a.Space(4), a.Px(8)).AlignItems(ui.Center).Gap(a.Space(6)).Children(func() {
			if glyph != "" {
				icon(c, glyph, a.Px(12), a.native.TextMuted)
			}
			ui.Text(c, strings.ToUpper(label)).FontSize(a.Px(12)).TextColor(a.native.TextMuted).Grow(1)
			ui.Text(c, fmt.Sprint(row.Count)).FontSize(a.Px(12)).TextColor(a.native.TextMuted)
		})
		return
	}
	label, detail, glyph := displayLevel(row.Level), s.thinkingHint(row.Level), thinkingIcon(row.Level)
	mono := false
	brand := a.AccentText
	if row.Level == "" {
		m := s.Snapshot.Models[row.ModelIndex]
		label = modelLabel(m.Name, m.ID)
		detail = m.ID
		mono = true
		if m.ContextWindow > 0 {
			detail = s.label("model_selector.context_window", "%{count} context window", "count", formatTokens(m.ContextWindow))
			mono = false
		}
		glyph = providerIcon(m.Provider)
		brand = providerColor(m.Provider, a)
	}
	b := ui.ButtonBase(c).AlignSelf(ui.Stretch).Height(height).Shrink(0).Padding(a.Px(4), a.Space(6)).Margin(0, a.Space(4)).Gap(a.Space(8)).Radius(a.Px(6)).Label("Choose " + label)
	if row.Selected || index == s.picker.cursor {
		b.Background(a.Active)
	} else if b.Hovered() {
		b.Background(a.Hover)
	}
	// Pointer movement is measured once in panel coordinates. A row that
	// scrolls under a stationary pointer cannot steal the keyboard cursor.
	_, _, over := b.PointerPosition()
	if over && s.picker.pointerMoved && time.Now().After(s.picker.suppressHoverUntil) {
		s.picker.cursor = index
	}

	b.Children(func() {
		ui.Box(c).Size(28, 28).Shrink(0).Center().Radius(a.Px(8)).Border(1, a.native.Border).Background(a.Elevated).Children(func() { icon(c, glyph, a.Px(14), brand) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(0).Children(func() {
			ui.Text(c, label).FontSize(a.Px(14)).LineHeight(1.618).SingleLine()
			line := ui.Text(c, detail).FontSize(a.Px(12)).LineHeight(1.618).TextColor(a.native.TextMuted).SingleLine()
			if mono {
				line.Font(a.CodeFont)
			}
		})
		if row.Level == "" {
			s.star(c, row.ModelIndex, !b.Hovered())
		}
		if row.Selected {
			icon(c, "icons/check.svg", a.Px(14), a.AccentText)
		} else {
			ui.Box(c).Size(a.Px(14), a.Px(14)).Shrink(0)
		}
	})
	// The parent centers the selected row in source token coordinates.
	_ = reveal
	if b.Clicked() {
		s.picker.cursor = index
		s.ConfirmPicker()
	}
}
func thinkingIcon(level string) string {
	switch strings.ToLower(level) {
	case "off", "none":
		return "icons/thinking-none.svg"
	case "minimal", "low":
		return "icons/thinking-low.svg"
	case "medium":
		return "icons/thinking-medium.svg"
	case "high":
		return "icons/thinking-high.svg"
	case "xhigh", "ultra":
		return "icons/thinking-xhigh.svg"
	case "max":
		return "icons/thinking-max.svg"
	case "auto":
		return "icons/thinking-auto.svg"
	}
	return "icons/spark.svg"
}

// centeredOffset follows ModelSelector::apply_scroll_to_row, including headers
// and row gaps. It does not depend on potentially stale previous-frame bounds.
func centeredOffset(rows []Row, index int, rowHeight, headerHeight, gap, maxHeight float32) float32 {
	if index < 0 || index >= len(rows) || len(rows) == 0 {
		return 0
	}
	var total, top, height float32
	for i, row := range rows {
		h := rowHeight
		if !row.Selectable() {
			h = headerHeight
		}
		if i < index {
			top += h + gap
		}
		if i == index {
			height = h
		}
		total += h
		if i < len(rows)-1 {
			total += gap
		}
	}
	viewport := min(total, maxHeight)
	return max(float32(0), min(top-(viewport-height)/2, total-viewport))
}
