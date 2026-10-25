package orbitassets

// Go-only UI states do not change the pinned Rust catalogs. English is explicit
// fallback for languages whose supplement has not yet been translated.
var goEnglish = map[string]string{
	"go.retry.unknown":     "Unknown: this pi version has no getter",
	"go.retry.unsupported": "Unsupported by this pi runtime",
	"go.state.enabled":     "Enabled",
	"go.state.disabled":    "Disabled",
	"go.action.enable":     "Enable",
	"go.action.disable":    "Disable",
	"go.session.rename":    "Rename…",
}
var goChinese = map[string]string{
	"go.retry.unknown":     "未知：此 pi 版本不提供查询接口",
	"go.retry.unsupported": "此 pi 运行时不支持",
	"go.state.enabled":     "已启用",
	"go.state.disabled":    "已禁用",
	"go.action.enable":     "启用",
	"go.action.disable":    "禁用",
	"go.session.rename":    "重命名…",
}

func supplementGoLocale(catalog map[string]string, language string) {
	for key, fallback := range goEnglish {
		// Source catalog ownership wins, even if future Rust adds this exact key.
		if _, exists := catalog[key]; exists {
			continue
		}
		value := fallback
		if language == "zh-CN" {
			if translated, ok := goChinese[key]; ok {
				value = translated
			}
		}
		catalog[key] = value
	}
}
