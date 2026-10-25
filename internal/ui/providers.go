package orbitui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
)

func (a *App) refreshProviders() {
	a.settingsGeneration++
	generation, root := a.settingsGeneration, a.Model.Workspace
	svc := a.Service()
	a.settingsLoading = true
	a.Async(func() func() {
		v, e := svc.Providers()
		creds, e2 := svc.Credentials()
		return func() {
			if generation != a.settingsGeneration || root != a.Model.Workspace {
				return
			}
			a.providers = v
			a.credentials = creds
			a.settingsLoading = false
			a.settingsErrors = errorStrings([]error{e, e2})
		}
	})
	a.Dispatch(Action{Kind: "RefreshProviders"})
}
func (a *App) providerRows(c *ui.Context) {
	if a.button(c, "Add provider…", "icons/plus.svg", false).Clicked() {
		a.editProvider(workbench.Provider{})
	}
	for _, p := range a.providers {
		p := p
		a.setting(c, p.Name, p.ID+" · "+p.BaseURL, func() {
			a.button(c, "Configure", "icons/settings.svg", false).Menu(func(m *ui.Menu) {
				if m.Item("Edit…").Chosen() {
					a.editProvider(p)
				}
				if m.Item("Remove…").Chosen() {
					a.Confirm("Remove provider?", p.ID, func() {
						svc := a.Service()
						a.Async(func() func() { e := svc.RemoveProvider(p.ID); return func() { a.save(e); a.refreshProviders() } })
					})
				}
			})
		})
	}
	for _, p := range a.Model.AuthProviders {
		p := p
		status := "Not authenticated"
		if p.Authenticated {
			status = "Authenticated"
		}
		a.setting(c, p.Name, status, func() {
			a.button(c, "Authentication", "", false).Menu(func(m *ui.Menu) {
				for _, method := range p.Methods {
					if m.Item(method.Label).Chosen() {
						a.Dispatch(Action{Kind: "Login", Value: p.ID, Data: map[string]any{"method": method.ID}})
					}
				}
				if m.Item("API key…").Chosen() {
					a.secretProvider = p.ID
					a.secretValue = ""
					a.secretOpen = true
				}
				if m.Item("Sign out…").Chosen() {
					a.Confirm("Sign out?", p.ID, func() { a.Dispatch(Action{Kind: "Logout", Value: p.ID}) })
				}
			})
		})
	}
	if a.Model.AuthURL != "" {
		ui.Text(c, a.Model.AuthStatus).TextColor(a.Theme.TextMuted)
		if a.button(c, "Open authentication page", "icons/external-link.svg", false).Clicked() {
			a.Dispatch(Action{Kind: "OpenLink", Value: a.Model.AuthURL})
		}
		if a.Model.AuthCode != "" {
			ui.Text(c, a.Model.AuthCode).Font(a.Fonts.CodeFamily).FontSize(20)
		}
		if a.button(c, "Cancel sign in", "", false).Clicked() {
			a.Dispatch(Action{Kind: "CancelLogin"})
		}
	}
	if len(a.providers) == 0 && len(a.Model.AuthProviders) == 0 && !a.settingsLoading {
		ui.Text(c, "No providers have been reported. Connect pi or add a provider.").TextColor(a.Theme.TextMuted)
	}
}
func (a *App) editProvider(p workbench.Provider) {
	a.jsonTitle = "Provider"
	a.jsonName = p.ID
	a.jsonScope = "user"
	a.jsonValue = fmt.Sprintf("{\n  \"name\": %q,\n  \"baseUrl\": %q,\n  \"api\": %q,\n  \"models\": %q\n}", p.Name, p.BaseURL, p.API, strings.Join(p.ModelIDs, ","))
	a.jsonOpen = true
	svc := a.Service()
	a.jsonCommit = func(id, scope, value string) error {
		var d struct {
			Name    string `json:"name"`
			BaseURL string `json:"baseUrl"`
			API     string `json:"api"`
			Models  string `json:"models"`
		}
		if e := json.Unmarshal([]byte(value), &d); e != nil {
			return e
		}
		return svc.SaveProvider(workbench.ProviderEdit{ID: id, Name: d.Name, BaseURL: d.BaseURL, API: d.API, ModelIDs: strings.Split(d.Models, ",")})
	}
}
func (a *App) secretDialog(c *ui.Context) {
	if !a.secretOpen {
		return
	}
	a.modal(c.Key("secret-input"), &a.secretOpen, 480, func() {
		ui.Text(c, "API key for "+a.secretProvider).FontSize(18).Bold()
		ui.Text(c, "Stored in pi's local auth.json. It will be used for future requests to this provider.").TextColor(a.Theme.TextMuted)
		ui.TextInput(c, &a.secretValue).Password().Label("API key").AlignSelf(ui.Stretch)
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if a.button(c, "Cancel", "", false).Clicked() {
				a.secretOpen = false
				a.secretValue = ""
			}
			if a.button(c, "Save API key", "", true).Clicked() {
				svc := a.Service()
				id, key := a.secretProvider, a.secretValue
				a.secretValue = ""
				a.secretOpen = false
				a.Async(func() func() { e := svc.SaveAPIKey(id, key); return func() { a.save(e); a.refreshProviders() } })
			}
		})
	})
}
