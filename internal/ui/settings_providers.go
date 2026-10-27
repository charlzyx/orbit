package orbitui

import (
	"fmt"
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"os"
	"sort"
	"strings"
	"time"
)

type sourceProviderView struct {
	meta                       workbench.ProviderMetadata
	config                     workbench.Provider
	auth                       agent.AuthProvider
	credential                 workbench.CredentialStatus
	name, envName              string
	custom, builtin, connected bool
	liveCount                  int
}

func (a *App) sourceProviderViews() []sourceProviderView {
	metadata := map[string]workbench.ProviderMetadata{}
	configs := map[string]workbench.Provider{}
	auth := map[string]agent.AuthProvider{}
	builtin := map[string]bool{}
	counts := map[string]int{}
	for _, p := range a.providerState.catalog.Providers {
		metadata[p.ID] = p
		builtin[p.ID] = true
	}
	for _, p := range a.providers {
		configs[p.ID] = p
		if _, ok := metadata[p.ID]; !ok {
			metadata[p.ID] = workbench.ProviderMetadata{ID: p.ID, Name: p.Name, APIKey: true}
		}
	}
	for _, p := range a.Model.AuthProviders {
		auth[p.ID] = p
		if _, ok := metadata[p.ID]; !ok {
			metadata[p.ID] = workbench.ProviderMetadata{ID: p.ID, Name: p.Name}
		}
		v := metadata[p.ID]
		for _, method := range p.Methods {
			if method.ID == "api_key" {
				v.APIKey = true
			} else {
				v.OAuth = true
			}
		}
		metadata[p.ID] = v
	}
	for _, m := range a.Model.ModelCatalog {
		counts[m.Provider]++
		if _, ok := metadata[m.Provider]; !ok {
			metadata[m.Provider] = workbench.ProviderMetadata{ID: m.Provider, Name: m.Provider}
		}
	}
	out := []sourceProviderView{}
	for id, meta := range metadata {
		config, custom := configs[id]
		name := meta.Name
		if config.Name != "" {
			name = config.Name
		}
		if name == "" {
			name = id
		}
		v := sourceProviderView{meta: meta, config: config, auth: auth[id], credential: a.credentials[id], name: name, custom: custom, builtin: builtin[id], liveCount: counts[id]}
		for _, env := range meta.EnvVars {
			if _, ok := os.LookupEnv(env); ok {
				v.envName = env
				break
			}
		}
		v.connected = v.credential.Configured || v.auth.Authenticated || config.HasAPIKey || v.envName != ""
		out = append(out, v)
	}
	rank := func(v sourceProviderView) int {
		if v.liveCount > 0 {
			return 0
		}
		if v.connected {
			return 1
		}
		if v.custom {
			return 2
		}
		return 3
	}
	sort.SliceStable(out, func(i, j int) bool {
		if rank(out[i]) != rank(out[j]) {
			return rank(out[i]) < rank(out[j])
		}
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out
}
func (a *App) providerToolbar(c *ui.Context) {
	t := a.Theme
	all := a.sourceProviderViews()
	active, connected := 0, 0
	for _, v := range all {
		if v.liveCount > 0 {
			active++
		}
		if v.connected {
			connected++
		}
	}
	ui.Column(c).AlignSelf(ui.Stretch).Gap(t.Space(12)).Children(func() {
		ui.Row(c).Height(t.Px(36)).PaddingX(t.Space(10)).Gap(t.Px(4)).AlignItems(ui.Center).BorderWidth(0, 0, 1, 0).BorderColor(t.Border).Children(func() {
			a.icon(c, "icons/search.svg", t.Px(14), t.TextMuted)
			ui.TextInputBase(c, &a.providerState.query).Label("Search providers").Placeholder(a.Tr("app.search_providers")).Grow(1)
		})
		ui.Row(c).AlignSelf(ui.Stretch).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
			text := strings.NewReplacer("%{all}", fmt.Sprint(len(all)), "%{active}", fmt.Sprint(active), "%{connected}", fmt.Sprint(connected)).Replace(a.Tr("settings.providers_active_connected"))
			ui.Text(c, text).FontSize(t.Px(12)).TextColor(t.TextMuted).Grow(1)
			if a.sourceProviderButton(c, "Refresh providers", a.Tr("settings.refresh"), "icons/refresh.svg", false).Height(t.Px(32)).Disabled(a.settingsLoading).Clicked() {
				a.refreshProviders()
			}
			if a.sourceProviderButton(c, "Add provider…", a.Tr("settings.add_provider"), "icons/plus.svg", true).Height(t.Px(32)).Clicked() {
				a.editProvider(workbench.Provider{})
			}
		})
	})
}
func (a *App) sourceProviderButton(c *ui.Context, id, label, glyph string, primary bool) ui.Element {
	t := a.Theme
	b := ui.ButtonBase(c.Key(id)).Label(id).Height(t.Px(28)).Padding(0, t.Space(8)).Gap(t.Space(4)).Radius(t.Px(8)).Border(1, t.Border).Background(t.Elevated).TextColor(t.TextSecondary)
	fg := t.TextSecondary
	if primary {
		b.Background(t.Accent).BorderColor(t.Accent).TextColor(t.AccentFg)
		fg = t.AccentFg
	} else if b.Hovered() {
		b.Background(t.Hover)
	}
	b.Children(func() {
		if glyph != "" {
			a.icon(c, glyph, 14, fg)
		}
		ui.Text(c, label).FontSize(t.Px(14)).FontWeight(500).SingleLine()
	})
	return b
}
func (a *App) providersSettings(c *ui.Context) { a.providerRows(c) }
func (a *App) providerRows(c *ui.Context) {
	views := a.sourceProviderViews()
	filtered := []sourceProviderView{}
	q := strings.ToLower(strings.TrimSpace(a.providerState.query))
	for _, v := range views {
		if q == "" || strings.Contains(strings.ToLower(v.name+" "+v.meta.ID+" "+strings.Join(v.meta.EnvVars, " ")), q) {
			filtered = append(filtered, v)
		}
	}
	if len(filtered) == 0 && !a.settingsLoading {
		a.sourceSettingsGroup(c, "No providers", func() {
			ui.Column(c).AlignSelf(ui.Stretch).Padding(a.Theme.Space(40), a.Theme.Space(12)).Gap(a.Theme.Space(8)).AlignItems(ui.Center).Children(func() {
				a.icon(c, "icons/search.svg", 26, a.Theme.TextMuted)
				ui.Text(c, a.Tr("settings.no_providers_match")).FontWeight(500)
				ui.Text(c, a.Tr("settings.try_a_different_search")).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextSecondary)
			})
		})
		return
	}
	ui.Column(c).AlignSelf(ui.Stretch).Gap(a.Theme.Space(12)).Children(func() {
		for start := 0; start < len(filtered); start += 3 {
			ui.Row(c).AlignSelf(ui.Stretch).Gap(a.Theme.Space(12)).AlignItems(ui.Stretch).Children(func() {
				for i := start; i < start+3; i++ {
					if i >= len(filtered) {
						ui.Box(c).Grow(1).Basis(0)
						continue
					}
					a.sourceProviderCard(c.Key(filtered[i].meta.ID), filtered[i])
				}
			})
		}
	})
	if id := a.providerState.usageProvider; id != "" {
		open := true
		a.modal(c.Key("provider-usage"), &open, 520, func() {
			ui.Text(c, a.Tr("session.usage")).FontSize(a.Theme.Px(16)).FontWeight(500)
			for _, r := range a.Model.Quotas {
				if r.Provider == id {
					a.runtimeQuotaCard(c, r, time.Now())
				}
			}
			if a.sourceSettingsButton(c, a.Tr("go.providers.refresh_usage")).Label("Refresh usage").Clicked() {
				a.Dispatch(Action{Kind: "RefreshQuota"})
			}
		})
		if !open {
			a.providerState.usageProvider = ""
		}
	}
}
func (a *App) sourceProviderCard(c *ui.Context, v sourceProviderView) {
	t := a.Theme
	id := v.meta.ID
	ui.Column(c.Key("provider:"+id)).Label("Provider card "+id).Grow(1).Basis(0).MinWidth(0).Padding(t.Space(12)).Gap(t.Space(8)).Border(1, t.Border).Radius(t.Px(12)).Background(t.SurfaceSecondary).Children(func() {
		ui.Row(c).AlignSelf(ui.Stretch).Gap(t.Space(12)).AlignItems(ui.Start).Children(func() {
			ui.Box(c).Size(52, 52).Shrink(0).Radius(t.Px(12)).Border(1, t.Border).Background(t.Elevated).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
				glyph := id
				if glyph == "ollama-cloud" {
					glyph = "ollama"
				}
				path := "icons/providers/" + glyph + ".svg"
				if icons[path] == nil {
					path = "icons/cloud.svg"
				}
				a.icon(c, path, 30, sourceProviderColor(id, t, v.liveCount > 0 || v.connected))
			})
			ui.Column(c).Grow(1).MinWidth(0).Gap(t.Space(4)).Children(func() {
				ui.Text(c, v.name).FontWeight(500).MaxLines(1).Tooltip(v.name)
				ui.Text(c, id).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 10.5 / 13).TextColor(t.TextMuted).MaxLines(1)
			})
		})
		ui.Row(c).Wrap().Gap(t.Space(6)).AlignItems(ui.Center).Children(func() {
			status, color := a.Tr("status.not_configured"), t.TextMuted
			if v.liveCount > 0 {
				status, color = a.Tr("status.active"), t.Success
			} else if v.connected {
				status, color = a.Tr("status.connected"), t.AccentText
			} else if v.custom {
				status, color = a.Tr("status.not_loaded"), t.Warning
			}
			auth := a.Model.Auth
			if auth.Provider == id && auth.SessionID != "" && a.providerState.dismissedAuthSession != auth.SessionID {
				switch auth.Status {
				case "pending":
					status, color = a.Tr("status.connecting"), t.Warning
				case "waiting":
					status, color = a.Tr("status.connecting"), t.AccentText
				case "succeeded":
					status, color = a.Tr("status.connected"), t.Success
				case "failed", "timeout":
					status, color = a.Tr("status.sign_in_failed"), t.Error
				case "cancelled":
					status, color = a.Tr("status.cancelled"), t.TextMuted
				}
			}
			ui.Row(c).Height(t.Px(22)).PaddingX(t.Space(8)).Gap(t.Space(6)).Radius(11).Background(color.Alpha(.12)).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(6, 6).Radius(3).Background(color)
				ui.Text(c, status).FontSize(t.Px(12)).TextColor(color)
			})
			badge := func(label string, accent bool) {
				fg, bg := t.TextMuted, t.Text.Alpha(.08)
				if accent {
					fg, bg = t.AccentText, t.Accent.Alpha(.12)
				}
				ui.Box(c).Height(t.Space(20)).PaddingX(t.Space(6)).Radius(t.Px(4)).Background(bg).AlignItems(ui.Center).Children(func() { ui.Text(c, label).FontSize(t.Px(10)).FontWeight(500).TextColor(fg) })
			}
			if v.meta.OAuth {
				badge("OAuth", true)
			}
			if v.meta.APIKey {
				badge(a.Tr("settings.api_key"), false)
			}
			label := "Provider"
			if v.builtin {
				label = "Built-in"
			}
			if v.custom {
				label = "Custom"
			}
			badge(label, v.custom)
		})
		count := v.liveCount
		key := "settings.n_models"
		if count == 0 && v.meta.HasCatalogCount {
			count = v.meta.CatalogCount
			key = "settings.n_in_catalog"
		} else if count == 0 {
			count = len(v.config.ModelIDs)
		}
		facts := strings.ReplaceAll(a.Tr(key), "%{count}", fmt.Sprint(count))
		if v.credential.Configured {
			facts += " · " + v.credential.Kind
		} else if v.envName != "" {
			facts += " · " + v.envName
		}
		if v.config.HasAPIKey {
			facts += " · " + a.Tr("settings.models_json_key")
		}
		ui.Text(c, facts).FontSize(t.Px(12)).TextColor(t.TextMuted).MaxLines(1)
		endpoint := v.config.BaseURL
		if endpoint == "" {
			endpoint = v.meta.Note
			if endpoint == "" {
				endpoint = a.Tr("settings.pi_default_endpoint")
			}
		} else if v.config.API != "" {
			endpoint += " · " + v.config.API
		}
		ui.Text(c, endpoint).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 10.5 / 13).TextColor(t.TextMuted).MaxLines(1).Tooltip(endpoint)
		ui.Box(c).Height(1).AlignSelf(ui.Stretch).Background(t.Border)
		if !a.sourceProviderAuth(c, v) {
			a.sourceProviderActions(c, v)
		}
	})
}
func (a *App) sourceProviderActions(c *ui.Context, v sourceProviderView) {
	id := v.meta.ID
	ui.Row(c).AlignSelf(ui.Stretch).Wrap().Gap(a.Theme.Space(8)).Children(func() {
		api := false
		for _, method := range v.auth.Methods {
			if method.ID == "api_key" {
				api = true
				continue
			}
			if a.sourceProviderButton(c, "Sign in "+id+" "+method.ID, a.sourceAuthMethodLabel(method, v.connected), "icons/lock.svg", true).Clicked() {
				a.providerState.dismissedAuthSession = ""
				a.providerState.authTarget = id
				a.Dispatch(Action{Kind: "Login", Value: id, Data: map[string]any{"method": method.ID}})
			}
		}
		if api || v.auth.ID == "" && (v.meta.APIKey || v.custom) {
			label := a.Tr("settings.add_api_key")
			if v.credential.Kind == "api_key" || v.config.HasAPIKey {
				label = a.Tr("settings.update_key")
			}
			if a.sourceProviderButton(c, "API key "+id, label, "icons/at-sign.svg", len(v.auth.Methods) == 0 && (!v.meta.OAuth || v.connected)).Clicked() {
				a.openProviderSecret(id, false)
			}
		}
		if v.auth.ID != "" && len(v.auth.Methods) == 0 {
			ui.Text(c, a.Tr("settings.no_sign_in_methods_available")).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
		} else if v.meta.OAuth && v.auth.ID == "" {
			if a.Model.Auth.Support == "unsupported" {
				label := a.Tr("settings.sign_in")
				if v.connected {
					label = a.Tr("settings.reconnect")
				}
				if a.sourceProviderButton(c, "Sign in terminal "+id, label, "icons/lock.svg", true).Clicked() {
					a.Dispatch(Action{Kind: "LoginTerminal", Value: id})
				}
			} else {
				ui.Text(c, a.Tr("go.providers.connect_methods")).FontSize(a.Theme.Px(12)).TextColor(a.Theme.TextMuted)
			}
		}
		if id == "ollama" || id == "ollama-cloud" {
			if a.sourceProviderButton(c, "Ollama session "+id, a.Tr("settings.usage_session"), "icons/clock.svg", false).Clicked() {
				a.openProviderSecret(id, true)
			}
		}
	})
	if v.custom || v.credential.Configured || v.auth.Authenticated || a.providerHasUsage(id) {
		ui.Row(c).Wrap().Gap(a.Theme.Space(6)).Children(func() {
			if v.custom {
				if a.sourceProviderButton(c, "Configure "+id, a.Tr("settings.configure"), "icons/settings.svg", false).Clicked() {
					a.editProvider(v.config)
				}
			}
			if v.auth.Authenticated || v.credential.Configured {
				if a.sourceProviderButton(c, "Disconnect "+id, a.Tr("settings.disconnect"), "icons/stop.svg", false).Clicked() {
					a.confirmProviderDisconnect(v)
				}
			}
			if v.custom {
				if a.sourceProviderButton(c, "Remove "+id, a.Tr("settings.remove"), "icons/trash.svg", false).TextColor(a.Theme.Error).Clicked() {
					a.confirmProviderRemove(v)
				}
			}
			if a.providerHasUsage(id) {
				if a.sourceProviderButton(c, "Usage "+id, a.Tr("session.usage"), "icons/usage-total.svg", false).Clicked() {
					a.providerState.usageProvider = id
				}
			}
		})
	}
}
func (a *App) providerHasUsage(id string) bool {
	for _, r := range a.Model.Quotas {
		if r.Provider == id && (len(r.Windows) > 0 || len(r.Balances) > 0 || r.Error != "" || r.Note != "") {
			return true
		}
	}
	return false
}
func (a *App) confirmProviderDisconnect(v sourceProviderView) {
	if v.auth.Authenticated {
		a.Confirm(a.Tr("go.providers.sign_out"), v.name, func() { a.Dispatch(Action{Kind: "Logout", Value: v.meta.ID}) })
		return
	}
	a.Confirm(a.Tr("go.providers.remove_credential"), strings.ReplaceAll(a.Tr("go.providers.remove_credential_detail"), "%{name}", v.name), func() {
		svc := a.Service()
		a.Async(func() func() {
			var err error
			if v.credential.Kind == "ollama_cloud_session" {
				err = svc.RemoveOllamaSession()
			} else {
				err = svc.RemoveCredential(v.meta.ID)
			}
			return func() {
				a.save(err)
				if err == nil {
					a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
				}
				a.refreshProviders()
			}
		})
	})
}
func (a *App) confirmProviderRemove(v sourceProviderView) {
	a.Confirm(a.Tr("go.providers.remove_config"), strings.ReplaceAll(a.Tr("go.providers.remove_config_detail"), "%{id}", v.meta.ID), func() {
		svc := a.Service()
		a.Async(func() func() {
			err := svc.RemoveProvider(v.meta.ID)
			return func() {
				a.save(err)
				if err == nil {
					a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
				}
				a.refreshProviders()
			}
		})
	})
}

// sourceProviderAuth replaces a card's normal actions only for that provider's
// actual login session. Dismissal is presentation-only; observed RPC state is not
// rewritten to manufacture successful authentication or cancellation.
func (a *App) sourceProviderAuth(c *ui.Context, v sourceProviderView) bool {
	auth := a.Model.Auth
	if auth.Provider != v.meta.ID || auth.SessionID == "" || auth.Status == "" || a.providerState.dismissedAuthSession == auth.SessionID {
		return false
	}
	t := a.Theme
	headline, detail, tint := auth.Status, auth.Error, t.TextMuted
	pending, retry := false, false
	switch auth.Status {
	case "pending":
		headline, detail, tint, pending = a.Tr("status.connecting"), a.Tr("auth.asking_pi"), t.Warning, true
	case "waiting":
		headline, detail, tint, pending = a.Tr("auth.waiting_browser"), a.Tr("auth.finish_in_browser"), t.AccentText, true
		if auth.UserCode != "" {
			headline, detail = a.Tr("auth.enter_device_code"), a.Tr("auth.approve_in_browser")
		}
	case "succeeded":
		headline, detail, tint = a.Tr("status.connected"), a.Tr("auth.credential_saved"), t.Success
	case "failed", "timeout":
		headline, tint, retry = a.Tr("status.sign_in_failed"), t.Error, true
		if detail == "" {
			detail = a.Tr("auth.sign_in_incomplete")
		}
	case "cancelled":
		headline, detail, retry = a.Tr("status.cancelled"), a.Tr("auth.sign_in_cancelled"), true
	}
	ui.Column(c).Label("Authentication status " + v.meta.ID).AlignSelf(ui.Stretch).Gap(t.Space(8)).Children(func() {
		ui.Row(c).Gap(t.Space(8)).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(7, 7).Radius(4).Background(tint)
			ui.Text(c, headline).FontSize(t.Px(12)).FontWeight(500).TextColor(tint)
		})
		if detail != "" {
			ui.Text(c, detail).FontSize(t.Px(12)).TextColor(t.TextSecondary)
		}
		if auth.UserCode != "" {
			ui.Column(c).AlignSelf(ui.Stretch).Padding(t.Space(8), t.Space(12)).Gap(t.Space(4)).Radius(t.Px(8)).Border(1, t.Border).Background(t.Background).Children(func() {
				ui.Text(c, a.Tr("settings.device_code")).FontSize(t.Px(10)).TextColor(t.TextMuted)
				ui.Text(c, auth.UserCode).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 16 / 13).FontWeight(600).Selectable()
				ui.Text(c, auth.VerificationURI).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 10.5 / 13).TextColor(t.TextMuted).MaxLines(1)
			})
		} else if auth.URL != "" {
			ui.Text(c, auth.URL).Font(a.Fonts.CodeFamily).FontSize(t.CodeSize * 10.5 / 13).TextColor(t.TextMuted).MaxLines(1)
		}
		ui.Row(c).Wrap().Gap(t.Space(8)).Children(func() {
			if pending {
				if auth.UserCode != "" {
					if a.sourceProviderButton(c, "Open authentication page "+v.meta.ID, a.Tr("auth.open_page"), "icons/arrow-up-right.svg", true).Disabled(auth.VerificationURI == "").Clicked() {
						a.Dispatch(Action{Kind: "OpenLink", Value: auth.VerificationURI})
					}
					if a.sourceProviderButton(c, "Copy authentication code "+v.meta.ID, a.Tr("auth.copy_code"), "icons/copy.svg", false).Clicked() {
						a.Dispatch(Action{Kind: "CopyText", Value: auth.UserCode})
					}
				}
				if a.sourceProviderButton(c, "Cancel sign in "+v.meta.ID, a.Tr("git_panel.cancel"), "icons/x.svg", false).Clicked() {
					a.Dispatch(Action{Kind: "CancelLogin"})
				}
				return
			}
			if retry {
				available := false
				for _, method := range v.auth.Methods {
					available = available || method.ID == auth.Method
				}
				if a.sourceProviderButton(c, "Retry sign in "+v.meta.ID, a.Tr("view.try_again"), "icons/lock.svg", true).Disabled(!a.Model.Connected || !available).Clicked() {
					a.providerState.dismissedAuthSession = ""
					a.Dispatch(Action{Kind: "Login", Value: auth.Provider, Data: map[string]any{"method": auth.Method}})
				}
			}
			label := "Dismiss"
			if auth.Status == "succeeded" {
				label = a.Tr("common.done")
			}
			if a.sourceProviderButton(c, "Dismiss authentication "+v.meta.ID, label, "icons/check.svg", auth.Status == "succeeded").Clicked() {
				a.providerState.dismissedAuthSession = auth.SessionID
			}
		})
	})
	return true
}
func (a *App) sourceAuthMethodLabel(method agent.AuthMethod, connected bool) string {
	if method.Label != "" && method.Label != method.ID {
		return method.Label
	}
	switch method.ID {
	case "browser", "oauth":
		if connected {
			return a.Tr("settings.reconnect")
		}
		return a.Tr("settings.sign_in")
	case "device_code":
		return a.Tr("settings.use_device_code")
	default:
		label := []rune(strings.NewReplacer("_", " ", "-", " ").Replace(method.ID))
		if len(label) > 0 && label[0] >= 'a' && label[0] <= 'z' {
			label[0] -= 'a' - 'A'
		}
		return string(label)
	}
}
func sourceProviderColor(id string, t Theme, connected bool) ui.Color {
	colors := map[string][2]uint32{"amazon-bedrock": {0xff9900, 0xc77700}, "ant-ling": {0x1677ff, 0x0e5fcc}, "anthropic": {0xd97757, 0xb55035}, "azure-openai-responses": {0x0078d4, 0x005a9e}, "baseten": {0x16a34a, 0x128040}, "cerebras": {0xf15a29, 0xc7461c}, "cloudflare-ai-gateway": {0xf38020, 0xc25e00}, "cloudflare-workers-ai": {0xf38020, 0xc25e00}, "deepseek": {0x4d6bfe, 0x3a52c9}, "fireworks": {0x8b5cf6, 0x6d28d9}, "google": {0x4285f4, 0x1a73e8}, "google-vertex": {0x4285f4, 0x1a73e8}, "groq": {0xf55036, 0xc93a22}, "huggingface": {0xffd21e, 0xa87800}, "kimi-coding": {0x3d5afe, 0x2e45c9}, "moonshotai": {0x3d5afe, 0x2e45c9}, "moonshotai-cn": {0x3d5afe, 0x2e45c9}, "minimax": {0xff4d4f, 0xcc3b3d}, "minimax-cn": {0xff4d4f, 0xcc3b3d}, "mistral": {0xfa520f, 0xc63d00}, "nvidia": {0x76b900, 0x5a8c00}, "openrouter": {0x6467f2, 0x5053c9}, "together": {0x0f6fff, 0x0b57cc}, "zai": {0x3b5bff, 0x2a43cc}, "zai-coding-cn": {0x3b5bff, 0x2a43cc}, "qwen-token-plan": {0x615ced, 0x4e49c9}, "xiaomi": {0xff6900, 0xcc5400}}
	if strings.HasPrefix(id, "qwen-token-plan") {
		id = "qwen-token-plan"
	}
	if strings.HasPrefix(id, "xiaomi") {
		id = "xiaomi"
	}
	pair, ok := colors[id]
	if !ok {
		if connected {
			return t.Text
		}
		return t.TextSecondary
	}
	v := pair[1]
	if t.Dark {
		v = pair[0]
	}
	return ui.RGB(uint8(v>>16), uint8(v>>8), uint8(v))
}
