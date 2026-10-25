package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/workbench"
	"github.com/egoist/mygo/ui"
	"strings"
)

type providerState struct {
	catalog                   workbench.ProviderCatalog
	query                     string
	connectedOnly             bool
	editor                    *providerDraft
	secretBusy, secretCookie  bool
	secretError               string
	authTarget, usageProvider string
	dismissedAuthSession      string
}
type providerDraft struct {
	edit                           workbench.ProviderEdit
	originalID, models, key, error string
	busy, inCatalog                bool
}

func (a *App) refreshProviders() {
	a.settingsGeneration++
	generation, root := a.settingsGeneration, a.Model.Workspace
	svc := a.Service()
	a.settingsLoading = true
	a.Async(func() func() {
		v, e := svc.Providers()
		creds, e2 := svc.Credentials()
		catalog, e3 := svc.ProviderCatalog(context.Background())
		return func() {
			if generation != a.settingsGeneration || root != a.Model.Workspace {
				return
			}
			a.providers = v
			a.credentials = creds
			a.providerState.catalog = catalog
			a.settingsLoading = false
			a.settingsErrors = errorStrings([]error{e, e2, e3})
		}
	})
	a.Dispatch(Action{Kind: "RefreshProviders"})
}
func (a *App) isCatalogProvider(id string) bool {
	for _, p := range a.providerState.catalog.Providers {
		if p.ID == id {
			return true
		}
	}
	return false
}
func (a *App) editProvider(p workbench.Provider) {
	inCatalog := a.isCatalogProvider(p.ID)
	a.providerState.editor = &providerDraft{edit: workbench.ProviderEditDefaults(p, inCatalog), originalID: p.ID, models: strings.Join(p.ModelIDs, "\n"), inCatalog: inCatalog}
}

var providerAPIs = []string{"", "openai-completions", "openai-responses", "anthropic-messages", "google-generative-ai", "mistral-conversations", "amazon-bedrock", "azure-openai-responses", "openai-codex-responses", "google-vertex"}

func providerModels(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == '\r' || r == ',' })
}
func (a *App) providerDialog(c *ui.Context) {
	d := a.providerState.editor
	if d == nil {
		return
	}
	open := true
	a.modal(c.Key("provider-editor:"+d.originalID), &open, 520, func() {
		title := "Add provider"
		if d.originalID != "" {
			title = "Edit provider"
		}
		ui.Text(c, title).FontSize(18).FontWeight(500)
		_, height := c.Size()
		ui.Scroll(c).MaxHeight(max(200, height-190)).AlignSelf(ui.Stretch).Gap(8).Children(func() {
			field := func(label string, value *string, placeholder string, readonly, password bool) {
				ui.Text(c, label).TextColor(a.Theme.TextMuted).FontSize(12)
				e := ui.TextInput(c, value).Label(label).Placeholder(placeholder).AlignSelf(ui.Stretch).Disabled(readonly || d.busy)
				if password {
					e.Password()
				}
			}
			field("Provider ID", &d.edit.ID, "my-provider", d.originalID != "", false)
			field("Display name", &d.edit.Name, "Optional display name", false, false)
			field("Base URL", &d.edit.BaseURL, "https://api.example.com/v1", false, false)
			ui.Text(c, "API format").FontSize(12).TextColor(a.Theme.TextMuted)
			label := d.edit.API
			if label == "" {
				label = "Inherit catalog default"
			}
			ui.MenuButton(c, label, func(m *ui.Menu) {
				for _, api := range providerAPIs {
					label := api
					if label == "" {
						label = "Inherit catalog default"
					}
					if m.Item(label).Checked(d.edit.API == api).Chosen() {
						d.edit.API = api
					}
				}
			}).Disabled(d.busy)
			field("API key", &d.key, "Leave blank to keep the existing key", false, true)
			ui.Text(c, "Saving a key stores it locally for future requests to this provider. Existing keys are never shown.").TextColor(a.Theme.TextMuted).FontSize(11)
			ui.Text(c, "Model IDs (one per line)").FontSize(12).TextColor(a.Theme.TextMuted)
			ui.TextArea(c, &d.models).Label("Model IDs").Placeholder("model-id").Height(100).AlignSelf(ui.Stretch).Disabled(d.busy)
			ui.Text(c, "Catalog providers may omit models to keep the built-in catalog. Custom providers need at least one model.").FontSize(11).TextColor(a.Theme.TextMuted)
		})
		if d.error != "" {
			ui.Text(c, d.error).TextColor(a.Theme.Error)
		}
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if a.button(c, "Cancel", "", false).Disabled(d.busy).Clicked() {
				open = false
			}
			if a.button(c, "Save provider", "", true).Disabled(d.busy).Clicked() {
				a.saveProviderDraft(d)
			}
		})
	})
	if !open {
		d.key = ""
		a.providerState.editor = nil
	}
}
func (a *App) saveProviderDraft(d *providerDraft) {
	edit := d.edit
	edit.ModelIDs = providerModels(d.models)
	d.inCatalog = a.isCatalogProvider(edit.ID)
	if e := workbench.ValidateProviderEdit(edit, d.inCatalog); e != nil {
		d.error = e.Error()
		return
	}
	if strings.TrimSpace(d.key) != "" {
		key := d.key
		edit.APIKey = &key
	}
	d.busy = true
	d.error = ""
	root := a.Model.Workspace
	svc := a.Service()
	a.Async(func() func() {
		e := svc.SaveProvider(edit)
		return func() {
			d.busy = false
			if a.providerState.editor != d || root != a.Model.Workspace {
				return
			}
			if e != nil {
				d.error = e.Error()
				return
			}
			d.key = ""
			a.providerState.editor = nil
			a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
			a.refreshProviders()
		}
	})
}
func (a *App) openProviderSecret(id string, cookie bool) {
	a.secretProvider = id
	a.secretValue = ""
	a.secretOpen = true
	a.providerState.secretCookie = cookie
	a.providerState.secretError = ""
}
func (a *App) secretDialog(c *ui.Context) {
	if !a.secretOpen {
		return
	}
	s := &a.providerState
	a.modal(c.Key("secret-input:"+a.secretProvider), &a.secretOpen, 480, func() {
		title := "API key for " + a.secretProvider
		label := "API key"
		if s.secretCookie {
			title = "Ollama web session"
			label = "Cookie header"
		}
		ui.Text(c, title).FontSize(18).Bold()
		ui.Text(c, "Stored in pi's local auth.json for ongoing requests to this provider. This value is never displayed in diagnostics.").TextColor(a.Theme.TextMuted)
		ui.TextInput(c, &a.secretValue).Password().Label(label).Placeholder(label).Disabled(s.secretBusy).AlignSelf(ui.Stretch)
		if s.secretError != "" {
			ui.Text(c, s.secretError).TextColor(a.Theme.Error)
		}
		ui.Row(c).Gap(8).Justify(ui.End).Children(func() {
			if a.button(c, "Cancel", "", false).Disabled(s.secretBusy).Clicked() {
				a.secretOpen = false
				a.secretValue = ""
			}
			if a.button(c, "Save credential", "", true).Disabled(s.secretBusy || strings.TrimSpace(a.secretValue) == "").Clicked() {
				svc := a.Service()
				id, key, cookie := a.secretProvider, a.secretValue, s.secretCookie
				s.secretBusy = true
				a.Async(func() func() {
					var e error
					if cookie {
						e = svc.SaveOllamaSession(key)
					} else {
						e = svc.SaveAPIKey(id, key)
					}
					return func() {
						s.secretBusy = false
						if !a.secretOpen || id != a.secretProvider {
							return
						}
						if e != nil {
							s.secretError = e.Error()
							return
						}
						a.secretValue = ""
						a.secretOpen = false
						a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
						a.refreshProviders()
					}
				})
			}
		})
	})
	if !a.secretOpen {
		a.secretValue = ""
	}
}
