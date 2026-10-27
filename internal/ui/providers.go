package orbitui

import (
	"context"
	"github.com/charlzyx/orbit/internal/workbench"
	"strings"
)

type providerState struct {
	catalog                   workbench.ProviderCatalog
	query                     string
	connectedOnly             bool
	editor                    *providerDraft
	secretBusy, secretCookie  bool
	secretError               string
	secretGeneration          uint64
	authTarget, usageProvider string
	dismissedAuthSession      string
}
type providerDraft struct {
	edit                           workbench.ProviderEdit
	originalID, models, key, error string
	busy, inCatalog, hadAPIKey     bool
	validationError                bool
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
	a.providerState.editor = &providerDraft{edit: workbench.ProviderEditDefaults(p, inCatalog), originalID: p.ID, models: strings.Join(p.ModelIDs, "\n"), inCatalog: inCatalog, hadAPIKey: p.HasAPIKey}
}

var providerAPIs = []string{"openai-completions", "openai-responses", "anthropic-messages", "google-generative-ai", "mistral-conversations", "amazon-bedrock", "azure-openai-responses", "openai-codex-responses", "google-vertex"}

func providerModels(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == '\r' || r == ',' })
}
func (a *App) saveProviderDraft(d *providerDraft) {
	edit := d.edit
	edit.ModelIDs = providerModels(d.models)
	d.inCatalog = a.isCatalogProvider(edit.ID)
	if e := workbench.ValidateProviderEdit(edit, d.inCatalog); e != nil {
		d.error = e.Error()
		d.validationError = true
		return
	}
	if strings.TrimSpace(d.key) != "" {
		key := d.key
		edit.APIKey = &key
	}
	d.validationError = false
	d.busy = true
	d.error = ""
	root := a.Model.Workspace
	svc := a.Service()
	a.Async(func() func() {
		e := svc.SaveProvider(edit)
		return func() {
			d.busy = false
			if e == nil {
				d.key = ""
			}
			if root != a.Model.Workspace {
				return
			}
			// Dismissing a modal does not undo its completed models.json write.
			// Refresh its successful effect, but never consume a newer draft.
			if e == nil {
				a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
				a.refreshProviders()
			}
			if a.providerState.editor != d {
				return
			}
			if e != nil {
				d.error = e.Error()
				return
			}
			a.providerState.editor = nil
		}
	})
}
func (a *App) openProviderSecret(id string, cookie bool) {
	a.providerState.secretGeneration++
	a.secretProvider = id
	a.secretValue = ""
	a.secretOpen = true
	a.providerState.secretCookie = cookie
	a.providerState.secretError = ""
}
func (a *App) saveProviderSecret() {
	s := &a.providerState
	if s.secretBusy || strings.TrimSpace(a.secretValue) == "" {
		return
	}
	svc := a.Service()
	id, key, cookie := a.secretProvider, a.secretValue, s.secretCookie
	generation := s.secretGeneration
	s.secretBusy = true
	s.secretError = ""
	a.Async(func() func() {
		var err error
		if cookie {
			err = svc.SaveOllamaSession(key)
		} else {
			err = svc.SaveAPIKey(id, key)
		}
		return func() {
			s.secretBusy = false
			// The save already ran even if its modal was dismissed. Refresh the
			// real runtime state, but never consume a subsequently reopened draft.
			if err == nil {
				a.Dispatch(Action{Kind: "RuntimeConfigChanged"})
				a.refreshProviders()
			}
			if !a.secretOpen || id != a.secretProvider || generation != s.secretGeneration {
				if err != nil {
					a.save(err)
				}
				return
			}
			if err != nil {
				s.secretError = err.Error()
				return
			}
			a.secretValue = ""
			a.secretOpen = false
		}
	})
}
