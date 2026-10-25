package orbitui

import (
	"crypto/ed25519"
	"encoding/base64"
	"path/filepath"
	"testing"

	"github.com/charlzyx/orbit/internal/native"
)

func TestBundledStartupCannotEnableProductionUpdater(t *testing.T) {
	oldFeed, oldKey := native.ReleaseFeedURL, native.ReleasePublicKey
	native.ReleaseFeedURL = "https://example.invalid/production-feed"
	native.ReleasePublicKey = base64.StdEncoding.EncodeToString(make([]byte, ed25519.PublicKeySize))
	t.Cleanup(func() { native.ReleaseFeedURL, native.ReleasePublicKey = oldFeed, oldKey })
	home := t.TempDir()
	userHome, piDir := t.TempDir(), t.TempDir()
	controller, err := NewController(RunOptions{Bundled: true, Home: home, UserHome: userHome, PiAgentDir: piDir, Workspace: t.TempDir(), PiBinary: filepath.Join(home, "missing-test-pi"), Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	defer controller.Close()
	if controller.Store.Home != home || controller.Store.UserHome != userHome || controller.Store.PiDir != piDir {
		t.Fatal("bundled controller collapsed private prefs into shared Pi config")
	}
	snapshot := controller.Updater.Snapshot()
	if snapshot.Enabled || snapshot.CanInstall {
		t.Fatalf("bundled test app enabled production updater: %+v", snapshot)
	}
}
