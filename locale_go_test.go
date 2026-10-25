package orbitassets

import "testing"

func TestGoLocaleSupplementFallbackAndOwnership(t *testing.T) {
	en, err := Locale("en")
	if err != nil {
		t.Fatal(err)
	}
	zh, err := Locale("zh-CN")
	if err != nil {
		t.Fatal(err)
	}
	for key, english := range goEnglish {
		if en[key] != english || zh[key] == english || zh[key] == "" {
			t.Fatal("missing locale", key)
		}
		if Translate(nil, en, key, nil) != english {
			t.Fatal("fallback", key)
		}
	}
	for _, lang := range Languages {
		catalog, err := Locale(lang)
		if err != nil {
			t.Fatal(err)
		}
		for key, english := range goEnglish {
			if catalog[key] == "" || catalog[key] == key {
				t.Fatal(lang, key)
			}
			if lang != "zh-CN" && catalog[key] != english {
				t.Fatal("English fallback", lang, key)
			}
		}
	}
	if bad, err := Locale("unsupported"); err == nil || bad != nil {
		t.Fatal("unknown locale was silently accepted")
	}
	original := map[string]string{"go.action.enable": "Future canonical source value", "sidebar.copy_path": "source"}
	supplementGoLocale(original, "zh-CN")
	if original["go.action.enable"] != "Future canonical source value" || original["sidebar.copy_path"] != "source" {
		t.Fatal("source overwritten")
	}
	// Per-load maps must not share writable state across windows or reloads.
	zh["go.action.enable"] = "mutation"
	again, _ := Locale("zh-CN")
	if again["go.action.enable"] != "启用" {
		t.Fatal("shared locale state")
	}
}
