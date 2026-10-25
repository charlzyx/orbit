package workbench

import (
	"context"
	"path/filepath"
	"testing"
)

// Manifest exclusion must not drop valid future providers whose data catalog
// exists before builtin metadata catches up. This fixture executes the actual
// credential-free metadata probe and verifies unknown capability semantics.
func TestProviderCatalogOnlyFutureProviderKeepsUnknownCapabilities(t *testing.T) {
	s := testService(t)
	root := t.TempDir()
	dist := filepath.Join(root, "node_modules", "@earendil-works", "pi-ai", "dist")
	cli := filepath.Join(root, "node_modules", "@earendil-works", "pi-coding-agent", "dist", "bundle", "cli.js")
	put(t, cli, "// fixture launcher")
	s.PiBinary = cli
	put(t, filepath.Join(dist, "..", "package.json"), `{"type":"module"}`)
	put(t, filepath.Join(dist, "providers", "all.js"), `export function builtinProviders(){return [{id:"known-provider",auth:{apiKey:{}}}]}`)
	put(t, filepath.Join(dist, "env-api-keys.js"), `export function findEnvKeys(){return []}`)
	data := filepath.Join(dist, "providers", "data")
	put(t, filepath.Join(data, "future-provider.json"), `{"openai-completions":{"future-model":{"id":"future-model","name":"Future model"}}}`)
	put(t, filepath.Join(data, ".manifest.json"), `{"catalogs":{"future-provider":"fixture-version"}}`)
	counts := readProviderCatalogCounts(data)
	if counts["future-provider"] != 1 {
		t.Fatal("valid fixture model catalog was not parsed", counts)
	}
	catalog, err := s.ProviderCatalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, provider := range catalog.Providers {
		if provider.ID == ".manifest" {
			t.Fatal("manifest metadata fabricated a provider")
		}
		if provider.ID == "future-provider" {
			found = true
			if provider.APIKey || provider.OAuth || !provider.HasCatalogCount || provider.CatalogCount != 1 {
				t.Fatal("catalog-only provider invented capability or lost models", provider)
			}
		}
	}
	if !found {
		t.Fatal("valid catalog-only future-provider was omitted by metadata join")
	}
}
