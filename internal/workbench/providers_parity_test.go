package workbench

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProviderInstalledMetadataIntrospectionAndCounts(t *testing.T) {
	s := testService(t)
	root := t.TempDir()
	dist := filepath.Join(root, "node_modules", "@earendil-works", "pi-ai", "dist")
	cli := filepath.Join(root, "node_modules", "@earendil-works", "pi-coding-agent", "dist", "bundle", "cli.js")
	put(t, cli, "// fixture launcher")
	s.PiBinary = cli
	put(t, filepath.Join(dist, "..", "package.json"), `{"type":"module"}`)
	put(t, filepath.Join(dist, "providers", "all.js"), `export function builtinProviders(){return [{id:"future-provider",name:"Future Provider",auth:{apiKey:{},oauth:{}}},{id:"openai",auth:{apiKey:{}}}]}`)
	put(t, filepath.Join(dist, "env-api-keys.js"), `export function findEnvKeys(id,env){return id==="future-provider"&&env.FUTURE_API_KEY ? ["FUTURE_API_KEY","FUTURE_FALLBACK"] : []}`)
	put(t, filepath.Join(dist, "providers", "data", "future-provider.json"), `{"api-one":{"m":{},"n":{}},"api-two":{"z":{}},"ignored":null}`)
	put(t, filepath.Join(dist, "providers", "data", "broken.json"), `{broken`)
	catalog, e := s.ProviderCatalog(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if !catalog.Dynamic || len(catalog.Providers) != 2 {
		t.Fatal(catalog)
	}
	p := catalog.Providers[0]
	if p.ID != "future-provider" || !p.OAuth || !p.APIKey || len(p.EnvVars) != 2 || p.EnvVars[0] != "FUTURE_API_KEY" || p.CatalogCount != 3 || !p.HasCatalogCount {
		t.Fatal(p)
	}
	if catalog.Providers[1].Name != "openai" || catalog.Providers[1].HasCatalogCount {
		t.Fatal(catalog.Providers[1])
	}
	// Broken dynamic metadata degrades to the exact curated baseline, not an
	// empty settings page or fabricated active/authenticated provider list.
	put(t, filepath.Join(dist, "providers", "all.js"), `export function builtinProviders(){return []}`)
	catalog, e = s.ProviderCatalog(context.Background())
	if e == nil || catalog.Dynamic || len(catalog.Providers) < 35 {
		t.Fatal(catalog, e)
	}
}
func TestProviderFallbackCopiesAndValidation(t *testing.T) {
	s := testService(t)
	s.PiBinary = filepath.Join(t.TempDir(), "missing-pi")
	catalog, e := s.ProviderCatalog(context.Background())
	if e == nil || catalog.Dynamic || len(catalog.Providers) == 0 {
		t.Fatal(catalog, e)
	}
	catalog.Providers[0].ID = "changed"
	catalog.Providers[0].EnvVars[0] = "changed"
	fresh := BuiltinProviders()
	if fresh[0].ID == "changed" || fresh[0].EnvVars[0] == "changed" {
		t.Fatal("fallback returned mutable global data")
	}
	for _, edit := range []ProviderEdit{{ID: "-invalid", BaseURL: "https://example.test", ModelIDs: []string{"a"}}, {ID: "bad space", BaseURL: "https://example.test", ModelIDs: []string{"a"}}, {ID: "valid", BaseURL: "file:///bad", ModelIDs: []string{"a"}}, {ID: "valid", BaseURL: "https://example.test"}, {ID: "valid", ModelIDs: []string{"a"}}} {
		if ValidateProviderEdit(edit, false) == nil {
			t.Fatal(edit)
		}
	}
	if e := ValidateProviderEdit(ProviderEdit{ID: "valid"}, true); e != nil {
		t.Fatal(e)
	}
	if p := ProviderEditDefaults(Provider{}, false); p.API != "openai-completions" {
		t.Fatal(p)
	}
	if p := ProviderEditDefaults(Provider{}, true); p.API != "" {
		t.Fatal(p)
	}
}
func TestProviderLegacyCredentialsOllamaIsolationAndSecretsAbsent(t *testing.T) {
	s := testService(t)
	path := filepath.Join(s.Preferences.PiDir, "auth.json")
	put(t, path, `{"old-oauth":{"refresh":"DO-NOT-EXPOSE"},"old-key":{"key":"DO-NOT-EXPOSE"},"ollama":{"type":"api_key","key":"local-placeholder"},"ollama-cloud-session":{"type":"ollama_cloud_session","session":"cookie=DO-NOT-EXPOSE"},"unknown":{"other":"keep"}}`)
	statuses, e := s.Credentials()
	if e != nil || statuses["old-oauth"].Kind != "oauth" || statuses["old-key"].Kind != "api_key" || statuses["ollama-cloud"].Kind != "ollama_cloud_session" || statuses["ollama"].Kind != "api_key" {
		t.Fatal(statuses, e)
	}
	raw, _ := json.Marshal(statuses)
	if strings.Contains(string(raw), "DO-NOT-EXPOSE") || strings.Contains(string(raw), "local-placeholder") || statuses[ollamaSessionKey].Configured {
		t.Fatal(string(raw))
	}
	if e := s.RemoveOllamaSession(); e != nil {
		t.Fatal(e)
	}
	m, _ := readObject(path)
	if m["ollama"] == nil || m[ollamaSessionKey] != nil || m["unknown"] == nil {
		t.Fatal("Ollama signout removed unrelated credentials")
	}
	put(t, path, `{"ollama":{"type":"ollama_cloud_session","session":"legacy=cookie"},"other":{"key":"keep"}}`)
	if e := s.SaveOllamaSession("new=cookie"); e != nil {
		t.Fatal(e)
	}
	m, _ = readObject(path)
	if m["ollama"] != nil || m[ollamaSessionKey] == nil || m["other"] == nil {
		t.Fatal(m)
	}
	if e := s.SaveOllamaSession("not-a-cookie"); e == nil {
		t.Fatal("accepted key as session cookie")
	}
	if e := s.SaveAPIKey(" trimmed ", "  sample-only  "); e != nil {
		t.Fatal(e)
	}
	m, _ = readObject(path)
	if m["trimmed"] == nil || m[" trimmed "] != nil {
		t.Fatal("provider id not normalized")
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm()&0077 != 0 {
		t.Fatal("credential store not private")
	}
}
func TestProviderPreservesUnknownSchemaAndNoopRemoval(t *testing.T) {
	s := testService(t)
	path := filepath.Join(s.Preferences.PiDir, "models.json")
	if e := s.RemoveProvider("missing"); e != nil {
		t.Fatal(e)
	}
	if _, e := os.Stat(path); !os.IsNotExist(e) {
		t.Fatal("remove created absent file")
	}
	put(t, path, `{"providers":["unknown-schema"],"other":true}`)
	before, _ := os.ReadFile(path)
	if e := s.RemoveProvider("missing"); e != nil {
		t.Fatal(e)
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("remove clobbered unknown schema")
	}
	if e := s.SaveProvider(ProviderEdit{ID: "x", API: "openai-completions"}); e == nil {
		t.Fatal("write clobbered malformed providers")
	}
}
func TestLivePiProviderMetadataWithoutInference(t *testing.T) {
	binary := os.Getenv("ORBIT_PI_METADATA_TEST_BIN")
	if binary == "" {
		t.Skip("set ORBIT_PI_METADATA_TEST_BIN for credential-free installed Pi metadata probe")
	}
	s := testService(t)
	s.PiBinary = binary
	catalog, e := s.ProviderCatalog(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if !catalog.Dynamic || len(catalog.Providers) < 35 {
		t.Fatal("installed Pi provider metadata not loaded")
	}
	known := 0
	for _, p := range catalog.Providers {
		if p.HasCatalogCount {
			known++
		}
	}
	if known == 0 {
		t.Fatal("installed model catalog counts missing")
	}
	t.Logf("Pi metadata: %d providers, %d provider catalogs; no inference or network requests", len(catalog.Providers), known)
}
