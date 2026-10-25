package orbitassets

import (
	"io/fs"
	"testing"
)

func TestOriginalCatalogsAndResources(t *testing.T) {
	en, err := Locale("en")
	if err != nil {
		t.Fatal(err)
	}
	if len(en) < 2000 {
		t.Fatalf("source catalog changed: %d keys", len(en))
	}
	for _, lang := range Languages {
		catalog, err := Locale(lang)
		if err != nil {
			t.Fatal(err)
		}
		for key := range en {
			if _, ok := catalog[key]; !ok {
				t.Errorf("%s missing %s", lang, key)
			}
		}
	}
	for _, name := range []string{"app-icon.png", "fonts/zed/IBMPlexSans-Regular.ttf", "fonts/zed/Lilex-Regular.ttf", "icons/lock.svg"} {
		if b, err := Asset(name); err != nil || len(b) == 0 {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := fs.ReadFile(Extensions(), "orbit-guard-extension/index.js"); err != nil {
		t.Fatal(err)
	}
}

func TestTranslationFallbackAndPathGuard(t *testing.T) {
	if _, err := Asset("../../secret"); err == nil {
		t.Fatal("asset path escaped")
	}
	if _, err := Locale("../../secret"); err == nil {
		t.Fatal("locale path escaped")
	}
	got := Translate(nil, map[string]string{"hello": "Hello %{name}"}, "hello", map[string]string{"name": "Ada"})
	if got != "Hello Ada" {
		t.Fatal(got)
	}
}
