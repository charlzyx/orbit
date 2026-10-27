package bugreportui

import (
	"errors"
	orbitassets "github.com/charlzyx/orbit"
	"github.com/egoist/mygo/ui"
	"testing"
)

func TestSettingsChineseBugPrivacyLiveSwitchNoConsent(t *testing.T) {
	en, _ := orbitassets.Locale("en")
	lang := "en"
	s := New(Options{Text: func(k, f string) string {
		m, _ := orbitassets.Locale(lang)
		v := orbitassets.Translate(m, en, k, nil)
		if v == k {
			return f
		}
		return v
	}})
	defer s.Close()
	s.optionsOpen = true
	s.Error = localValidationError("Up to 6 screenshots.")
	tt := ui.NewTester(func(c *ui.Context) { View(c, s).Fill() }, 1200, 1100)
	for _, l := range []string{"en", "zh-CN", "en"} {
		lang = l
		tt.Frame()
		if !tt.HasText(s.label("go.bug.copy_10", "Report privacy options")) || !tt.HasText(s.errorMessage(s.Error)) {
			t.Fatal(l, tt.Texts())
		}
		if s.CaptureConsent || s.IncludeIdentity || s.reviewed {
			t.Fatal("translation changed consent")
		}
	}
	lang = "zh-CN"
	s.Error = errors.New("Up to 6 screenshots.")
	tt.Frame()
	if !tt.HasText("Up to 6 screenshots.") {
		t.Fatal("backend error reinterpreted")
	}
}

func TestSettingsChineseBugOwnedStatusPreservesDetails(t *testing.T) {
	en, _ := orbitassets.Locale("en")
	zh, _ := orbitassets.Locale("zh-CN")
	s := &State{Options: Options{Text: func(k, f string) string {
		v := orbitassets.Translate(zh, en, k, nil)
		if v == k {
			return f
		}
		return v
	}}}
	for _, tc := range []struct{ raw, want string }{
		{"Local draft used. Review and edit it before sharing. RAW: denied", "已使用本地草稿。分享前请审核并编辑。 RAW: denied"},
		{"Issue created: https://example.invalid/42 Screenshots: /fixture/PIXELS", "已创建问题： https://example.invalid/42 截图： /fixture/PIXELS"},
	} {
		if got := s.displayMessage(tc.raw); got != tc.want {
			t.Fatalf("%q != %q", got, tc.want)
		}
	}
}
