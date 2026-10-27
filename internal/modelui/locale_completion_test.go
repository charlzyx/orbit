package modelui

import (
	"errors"
	orbitassets "github.com/charlzyx/orbit"
	"testing"
)

func TestSettingsChineseModelErrorOrigin(t *testing.T) {
	en, _ := orbitassets.Locale("en")
	zh, _ := orbitassets.Locale("zh-CN")
	s := &State{Options: Options{Text: func(k, f string) string {
		v := orbitassets.Translate(zh, en, k, nil)
		if v == k {
			return f
		}
		return v
	}}}
	var notices []string
	s.Options.Notify = func(v string, _ bool) { notices = append(notices, v) }
	s.failLocal("Model selection is not connected.")
	if len(notices) != 1 || notices[0] != "模型选择尚未连接。" {
		t.Fatal(notices)
	}
	if !s.localError || s.displayValidation(s.Error) != "模型选择尚未连接。" {
		t.Fatal(s.Error)
	}
	s.fail(errors.New("Model selection is not connected."))
	if len(notices) != 2 || notices[1] != "Model selection is not connected." {
		t.Fatal(notices)
	}
	if s.localError {
		t.Fatal("foreign error retained local origin")
	}
}
