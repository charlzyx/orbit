package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/egoist/mygo/ui"
	"slices"
	"testing"
)

func TestP20MountedChineseRetryAndEnglishReload(t *testing.T) {
	a := agentAlignmentFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	a.Model.FollowUpMode = ""
	a.Model.AutoCompactionKnown = false
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 1100)
	tt.Frame()
	if !tt.HasText("未知") || !tt.HasText("未知：此 pi 版本不提供查询接口") {
		t.Fatal("English leaked into Chinese unknown state")
	}
	if err := tt.Click("Set retry policy"); err != nil {
		t.Fatal(err)
	}
	if err := tt.ChooseMenuItem("启用"); err != nil {
		t.Fatal(err, tt.Menu())
	}
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "SetAutoRetry" || actions[0].Data["enabled"] != true {
		t.Fatal(actions)
	}
	for _, tc := range []struct {
		state agent.RetrySetting
		text  string
	}{
		{agent.RetrySetting{Known: true, Enabled: true}, "已启用"},
		{agent.RetrySetting{Known: true}, "已禁用"},
		{agent.RetrySetting{Pending: true}, "正在更新…"},
		{agent.RetrySetting{Support: "unsupported"}, "此 pi 运行时不支持"},
	} {
		a.Runtime.Active.AutoRetry = tc.state
		tt.Frame()
		if !tt.HasText(tc.text) {
			t.Fatal("missing Chinese state", tc.text)
		}
	}
	a.Runtime.Active.AutoRetry = agent.RetrySetting{Support: "unknown", Error: "Provider raw error: exact 429"}
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt.Frame()
	if !tt.HasText("Unknown: this pi version has no getter") || !tt.HasText("Provider raw error: exact 429") {
		t.Fatal("reload/raw error changed")
	}
}

func TestP20MountedChineseSessionActionsPreserveIdentity(t *testing.T) {
	a := sourceSidebarFixture(t)
	a.Prefs.Language = "zh-CN"
	a.reloadLocale()
	var actions []Action
	a.Options.OnAction = func(v Action) { actions = append(actions, v) }
	tt := ui.NewTester(a.View, 1470, 922)
	tt.Frame()
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tt.Menu(), "重命名…") {
		t.Fatal(tt.Menu())
	}
	if err := tt.ChooseMenuItem("复制会话 ID"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if len(actions) != 1 || actions[0].Kind != "CopyText" || actions[0].Value != "pin" {
		t.Fatal(actions)
	}
	a.Prefs.Language = "en"
	a.reloadLocale()
	tt.Frame()
	if err := tt.RightClick("pin task"); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(tt.Menu(), "Rename…") || !slices.Contains(tt.Menu(), "Copy session ID") {
		t.Fatal(tt.Menu())
	}
}
