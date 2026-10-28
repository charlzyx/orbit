package orbitassets

import (
	"regexp"
	"slices"
	"testing"
)

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

func TestSettingsChineseSupplementPlaceholders(t *testing.T) {
	re := regexp.MustCompile(`%\{[^}]+\}|\$\{[^}]+\}`)
	for key, english := range goEnglish {
		a, b := re.FindAllString(english, -1), re.FindAllString(goChinese[key], -1)
		slices.Sort(a)
		slices.Sort(b)
		if !slices.Equal(a, b) {
			t.Fatal(key, a, b)
		}
	}
}

func TestGoPiDisplayBrandEveryLocale(t *testing.T) {
	for _, language := range Languages {
		catalog, err := Locale(language)
		if err != nil {
			t.Fatal(err)
		}
		if catalog["app.name"] != "GoPi" || catalog["settings.orbit_pi"] != "GoPi" {
			t.Fatal(language, "brand", catalog["app.name"], catalog["settings.orbit_pi"])
		}
	}
	catalog := map[string]string{"view.set_up_orbit": "Set up Orbit", "sidebar.star_orbit": "Star Orbit", "theme.orbit": "Orbit Cobalt", "backend": "Orbit Go Test/Home"}
	supplementGoLocale(catalog, "en")
	if catalog["view.set_up_orbit"] != "Set up GoPi" || catalog["sidebar.star_orbit"] != "Star GoPi" || catalog["theme.orbit"] != "Orbit Cobalt" || catalog["backend"] != "Orbit Go Test/Home" {
		t.Fatal(catalog)
	}
}
