package preferences

import (
	"encoding/json"
	"sort"
)

const HiddenQuotaProvidersFile = "hidden-quota-providers.json"

// HiddenQuotaProviders preserves Rust's provider-ID keyed Orbit preference.
// Malformed/unreadable documents mean nothing hidden; mixed arrays retain only
// nonempty strings. This affects presentation, never provider authentication.
func (s *Store) HiddenQuotaProviders() []string {
	var value struct {
		Providers []json.RawMessage `json:"providers"`
	}
	if s.Read(HiddenQuotaProvidersFile, &value) != nil {
		return nil
	}
	var ids []string
	for _, raw := range value.Providers {
		var id string
		if json.Unmarshal(raw, &id) == nil && id != "" {
			ids = append(ids, id)
		}
	}
	return uniqueQuotaProviders(ids)
}
func uniqueQuotaProviders(ids []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, id := range ids {
		if id != "" && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}
func (s *Store) SaveHiddenQuotaProviders(ids []string) error {
	return s.Write(HiddenQuotaProvidersFile, struct {
		Providers []string `json:"providers"`
	}{uniqueQuotaProviders(ids)})
}
