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
	controller, err := NewController(RunOptions{Bundled: true, Home: home, Workspace: t.TempDir(), PiBinary: filepath.Join(home, "missing-test-pi"), Offline: true, NoExtensions: true})
	if err != nil {
		t.Fatal(err)
	}
	defer controller.Close()
	snapshot := controller.Updater.Snapshot()
	if snapshot.Enabled || snapshot.CanInstall {
		t.Fatalf("bundled test app enabled production updater: %+v", snapshot)
	}
}
