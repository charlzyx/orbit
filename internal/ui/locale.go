package orbitui

import (
	"github.com/egoist/mygo"
	"os"
	"runtime"
	"strings"
)

// shippedLocale matches pinned i18n.rs::shipped_locale_id, including its
// deliberate exclusion of Traditional Chinese and Portuguese outside Brazil.
func shippedLocale(tag string) string {
	tag = strings.ToLower(strings.ReplaceAll(tag, "_", "-"))
	if tag == "zh-cn" || tag == "zh-sg" || strings.HasPrefix(tag, "zh-hans") {
		return "zh-CN"
	}
	if tag == "pt" || tag == "pt-br" {
		return "pt-BR"
	}
	for _, lang := range []string{"en", "ja", "ko", "es", "fr", "de", "ru", "it"} {
		if tag == lang || strings.HasPrefix(tag, lang+"-") {
			return lang
		}
	}
	return ""
}
func sourceSystemLocale(platform string, lookup func(string) (string, bool), preferred []string) string {
	if platform == "darwin" {
		for _, tag := range preferred {
			if lang := shippedLocale(tag); lang != "" {
				return lang
			}
		}
		return "en"
	}
	// Rust uses Result::or_else, so an explicitly empty higher-priority variable
	// still wins and resolves to English. LANGUAGE is intentionally not consulted.
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value, ok := lookup(key); ok {
			tag, _, _ := strings.Cut(value, ".")
			if lang := shippedLocale(tag); lang != "" {
				return lang
			}
			return "en"
		}
	}
	return "en"
}
func systemLanguage() string {
	var preferred []string
	if runtime.GOOS == "darwin" {
		preferred = mygo.App.PreferredLanguages()
	}
	return sourceSystemLocale(runtime.GOOS, os.LookupEnv, preferred)
}
