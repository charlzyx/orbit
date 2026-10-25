package native

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

func LoadAutomaticUpdatePreference(path string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return true
	}
	var values map[string]json.RawMessage
	if json.Unmarshal(raw, &values) != nil {
		return true
	}
	var enabled bool
	if value, ok := values["automatic"]; ok && string(value) != "null" && json.Unmarshal(value, &enabled) == nil {
		return enabled
	}
	return true
}
func SaveAutomaticUpdatePreference(path string, enabled bool) error {
	return atomicJSON(path, map[string]bool{"automatic": enabled})
}

// CheckAndStage matches Orbit's launch-time checker: prepare a verified update
// when managed installation is possible, but never install it automatically.
func (u *Updater) CheckAndStage(ctx context.Context) error {
	if err := u.Check(ctx); err != nil {
		return err
	}
	s := u.Snapshot()
	if s.Available != nil && s.CanInstall {
		return u.Stage(ctx)
	}
	return nil
}

// StartAutomatic performs one check per launch, not an invented timer schedule.
// Both checking and preference I/O are off the UI thread.
func (u *Updater) StartAutomatic(ctx context.Context, preferencePath string) {
	go func() {
		if u.Snapshot().Enabled && LoadAutomaticUpdatePreference(preferencePath) {
			_ = u.CheckAndStage(ctx)
		}
	}()
}
func (u *Updater) SetAutomaticChecks(ctx context.Context, preferencePath string, enabled bool) error {
	if err := SaveAutomaticUpdatePreference(preferencePath, enabled); err != nil {
		return err
	}
	if enabled && u.Snapshot().Enabled {
		go func() { _ = u.CheckAndStage(ctx) }()
	}
	return nil
}

// TakeUpdateError surfaces a previous install helper's failure once. The error
// file is bounded, regular and beside this installation, never a followed link.
func TakeUpdateError(layout *InstallLayout) (string, error) {
	if layout == nil {
		return "", nil
	}
	path := filepath.Join(filepath.Dir(layout.Root), "."+filepath.Base(layout.Root)+".update-error")
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<10 {
		return "", nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if err = os.Remove(path); err != nil {
		return "", err
	}
	return string(raw), nil
}
