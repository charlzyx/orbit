package orbitui

import (
	"github.com/charlzyx/orbit/internal/agent"
	"github.com/charlzyx/orbit/internal/workspace"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSettingsAuthAndRetryProjectionKeepsProtocolState(t *testing.T) {
	auth := agent.AuthState{Support: "supported", SessionID: "login-1", Provider: "p", Method: "oauth", Status: "device", URL: "https://example.invalid/login", UserCode: "FIXTURE", VerificationURI: "https://example.invalid/device", Error: "fixture error", ExpiresAt: 123}
	got := adaptSnapshot(agent.Snapshot{Auth: auth, Retrying: true}, "/fixture", nil, nil)
	if !reflect.DeepEqual(got.Auth, auth) || !got.Retrying {
		t.Fatalf("lost source auth/retry state: %+v", got)
	}
	got = adaptSnapshot(agent.Snapshot{}, "/fixture", nil, nil)
	if got.Retrying || got.Auth.Provider != "" {
		t.Fatal("completed state retained old login/retry")
	}
}
func TestSettingsSavedWorktreePolicyReachesNextCreate(t *testing.T) {
	x, err := NewController(RunOptions{Home: t.TempDir(), Workspace: t.TempDir(), Offline: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(x.Close)
	cfg := workspace.DefaultWorktreeConfig()
	cfg.Directory = "../alternate-worktrees"
	cfg.SetupScript = "custom-setup.sh"
	cfg.RunSetupAuto = false
	cfg.OpenAfterCreate = false
	cfg.AllowedSetupRepos = []string{"/fixture/repo"}
	if err := workspace.SaveWorktreeConfig(filepath.Join(x.Store.RootDir, "worktrees.json"), cfg); err != nil {
		t.Fatal(err)
	}
	x.Action(Action{Kind: "WorktreePreferencesChanged"})
	if !reflect.DeepEqual(x.Worktrees.Config, cfg) {
		t.Fatalf("worktree controller kept old settings: %+v", x.Worktrees.Config)
	}
	cfg.Directory = "../changed-outside-page"
	if err := workspace.SaveWorktreeConfig(filepath.Join(x.Store.RootDir, "worktrees.json"), cfg); err != nil {
		t.Fatal(err)
	}
	x.UI.Options.OnPage("worktrees")
	if x.Worktrees.Config.Directory != cfg.Directory {
		t.Fatal("entering Worktrees ignored persisted configuration")
	}
}
func TestSettingsNativeMenusCannotEscapePaintedPopovers(t *testing.T) {
	for _, kind := range []string{"selector", "usage"} {
		t.Run(kind, func(t *testing.T) {
			a := fixture(t)
			a.page = "settings"
			a.settingsPage = "appearance"
			if kind == "selector" {
				a.nativeSettings.selectMenu.id = "font"
			} else {
				a.providerState.usageProvider = "fixture"
			}
			if !a.modalOpen() {
				t.Fatal("popover omitted from modal guard")
			}
			a.nativeCommand("NewSession")
			if a.page != "settings" {
				t.Fatal("native app command escaped active Settings popover")
			}
		})
	}
}
