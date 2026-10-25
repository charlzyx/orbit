package orbitui

import (
	"runtime"
	"testing"
)

func TestSystemLocaleMatchesPinnedSourcePrecedenceAndScripts(t *testing.T) {
	for _, tt := range []struct {
		platform  string
		env       map[string]string
		preferred []string
		want      string
	}{
		{"linux", map[string]string{"LANG": "zh_CN.UTF-8"}, nil, "zh-CN"},
		{"linux", map[string]string{"LC_ALL": "", "LC_MESSAGES": "ja_JP.UTF-8", "LANG": "zh_CN.UTF-8"}, nil, "en"},
		{"linux", map[string]string{"LC_MESSAGES": "ja_JP.UTF-8", "LANG": "zh_CN.UTF-8"}, nil, "ja"},
		{"linux", map[string]string{"LANGUAGE": "zh_CN", "LANG": "de_DE.UTF-8"}, nil, "de"},
		{"windows", map[string]string{"LANG": "ko_KR.UTF-8"}, nil, "ko"},
		{"linux", map[string]string{"LANG": "zh_TW.UTF-8"}, nil, "en"},
		{"linux", map[string]string{"LANG": "pt_PT.UTF-8"}, nil, "en"},
		{"linux", map[string]string{"LANG": "pt_BR.UTF-8"}, nil, "pt-BR"},
		{"linux", map[string]string{}, nil, "en"},
		{"darwin", nil, []string{"hi-IN", "zh-TW", "zh-Hans-CN", "en-GB"}, "zh-CN"},
		{"darwin", nil, []string{"zh-Hant", "pt-PT", "fr-FR"}, "fr"},
		{"darwin", nil, []string{"hi-IN"}, "en"},
	} {
		got := sourceSystemLocale(tt.platform, func(k string) (string, bool) { v, ok := tt.env[k]; return v, ok }, tt.preferred)
		if got != tt.want {
			t.Errorf("%+v got %s", tt, got)
		}
	}
}
func TestSystemLanguageReloadsShippedCatalogWithoutChangingPreference(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("source macOS language uses NSLocale, not environment variables")
	}
	a := fixture(t)
	t.Setenv("LC_ALL", "zh_CN.UTF-8")
	a.Prefs.Language = "system"
	a.reloadLocale()
	if a.Tr("language.system") == a.english["language.system"] || a.Prefs.Language != "system" {
		t.Fatal("system language did not select catalog or changed saved preference")
	}
	a.Prefs.Language = "en"
	a.reloadLocale()
	if a.Tr("language.system") != a.english["language.system"] {
		t.Fatal("explicit language did not override system")
	}
}
