package workbench

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
)

type Provider struct {
	ID, Name, BaseURL, API string
	HasAPIKey              bool
	ModelIDs               []string
}
type ProviderEdit struct {
	ID, Name, BaseURL, API string
	APIKey                 *string
	ModelIDs               []string
}

func (s *Service) Providers() ([]Provider, error) {
	m, e := readObject(filepath.Join(s.Preferences.PiDir, "models.json"))
	if e != nil {
		return nil, e
	}
	if m["providers"] == nil {
		return nil, nil
	}
	providers, ok := m["providers"].(map[string]any)
	if !ok {
		return nil, errors.New("models.json providers must be an object")
	}
	var out []Provider
	for id, v := range providers {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		p := Provider{ID: id, Name: stringValue(m["name"]), BaseURL: stringValue(m["baseUrl"]), API: stringValue(m["api"]), HasAPIKey: stringValue(m["apiKey"]) != ""}
		models, _ := m["models"].([]any)
		for _, v := range models {
			if model, ok := v.(map[string]any); ok {
				if id := stringValue(model["id"]); id != "" {
					p.ModelIDs = append(p.ModelIDs, id)
				}
			}
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (s *Service) SaveProvider(p ProviderEdit) error {
	p.ID = strings.TrimSpace(p.ID)
	if p.ID == "" {
		return errors.New("provider id is required")
	}
	return s.edit(filepath.Join(s.Preferences.PiDir, "models.json"), func(m map[string]any) error {
		providers, e := object(m, "providers")
		if e != nil {
			return e
		}
		entry, ok := providers[p.ID].(map[string]any)
		if !ok {
			entry = map[string]any{}
		}
		for k, v := range map[string]string{"name": p.Name, "baseUrl": p.BaseURL, "api": p.API} {
			if strings.TrimSpace(v) == "" {
				delete(entry, k)
			} else {
				entry[k] = strings.TrimSpace(v)
			}
		}
		if p.APIKey != nil && strings.TrimSpace(*p.APIKey) != "" {
			entry["apiKey"] = strings.TrimSpace(*p.APIKey)
		}
		if len(p.ModelIDs) > 0 {
			old, _ := entry["models"].([]any)
			models := make([]any, 0, len(p.ModelIDs))
			for _, id := range p.ModelIDs {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				var found any
				for _, raw := range old {
					if model, ok := raw.(map[string]any); ok && stringValue(model["id"]) == id {
						found = model
						break
					}
				}
				if found == nil {
					found = map[string]any{"id": id}
				}
				models = append(models, found)
			}
			entry["models"] = models
		}
		if len(entry) == 0 {
			delete(providers, p.ID)
		} else {
			providers[p.ID] = entry
		}
		return nil
	})
}
func (s *Service) RemoveProvider(id string) error {
	return s.edit(filepath.Join(s.Preferences.PiDir, "models.json"), func(m map[string]any) error {
		p, e := object(m, "providers")
		if e != nil {
			return e
		}
		delete(p, id)
		return nil
	})
}

type CredentialStatus struct {
	Kind       string
	Configured bool
}

// Credentials exposes status only; raw credential values never enter UI snapshots.
func (s *Service) Credentials() (map[string]CredentialStatus, error) {
	m, e := readObject(filepath.Join(s.Preferences.PiDir, "auth.json"))
	if e != nil {
		return nil, e
	}
	out := map[string]CredentialStatus{}
	for id, v := range m {
		if entry, ok := v.(map[string]any); ok {
			kind := stringValue(entry["type"])
			out[id] = CredentialStatus{Kind: kind, Configured: kind == "oauth" || (kind == "api_key" && stringValue(entry["key"]) != "")}
		}
	}
	return out, nil
}
func (s *Service) SaveAPIKey(id, key string) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(key) == "" {
		return errors.New("provider and API key are required")
	}
	return s.edit(filepath.Join(s.Preferences.PiDir, "auth.json"), func(m map[string]any) error {
		m[id] = map[string]any{"type": "api_key", "key": strings.TrimSpace(key)}
		return nil
	})
}
func (s *Service) RemoveCredential(id string) error {
	return s.edit(filepath.Join(s.Preferences.PiDir, "auth.json"), func(m map[string]any) error { delete(m, id); return nil })
}
